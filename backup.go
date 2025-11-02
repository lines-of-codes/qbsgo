package main

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/klauspost/compress/zstd"
	"github.com/nrednav/cuid2"
)

func (c *config) backup(targets []string) {
	genCuid, err := cuid2.Init(
		cuid2.WithLength(c.IdLength),
	)

	if err != nil {
		log.Fatal(err.Error())
	}

	fileExt := c.Archive

	if c.Archive == "tar" {
		switch c.Compression {
		case "gzip":
			fileExt += ".gz"
		case "zstd":
			fileExt += ".zst"
		}
	}

	for _, targetName := range targets {
		target := c.Targets[targetName]
		remote := c.Remotes[target.Remote]

		backupId := genCuid()
		date := time.Now()
		fileName := fmt.Sprintf("%s-%d%s%d-%s.%s", targetName, date.Day(), date.Month().String()[0:3], date.Year(), backupId, fileExt)
		outPath := path.Join(c.ArchiveDir, fileName)

		log.Printf("-- Backing up target %s with ID %s\n", targetName, backupId)

		backupStart := time.Now()

		file, err := c.writeToFileFirst(target, outPath)

		log.Printf("Archival took %.2f seconds", time.Since(backupStart).Seconds())

		if err != nil {
			log.Printf("Error in archive creation: %s", err)
			log.Printf("The file %s will be removed.", outPath)
			err = os.Remove(outPath)

			if err != nil {
				log.Printf("Error while deleting backup file: %s", err)
			}
			continue
		}

		defer file.Close()
		var dest string

		switch remote.Type {
		case "copyparty":
			dest, err = c.copypartyUpload(target.Remote, outPath, fileName)
		case "nextcloud":
			dest, err = c.nextcloudUpload(target.Remote, outPath, fileName)
		}

		if err != nil {
			log.Printf("Error while uploading file to %s/%s because:\n%s", target.Remote, fileName, err)
		}

		if c.DeleteAfterUpload {
			log.Printf("Deleting %s...", outPath)
			err = os.Remove(outPath)

			if err != nil {
				log.Printf("Error while deleting backup file: %s", err)
			}
		}

		c.BackupList.append(listEntry{
			Id:       backupId,
			Date:     backupStart.Format(time.RFC3339),
			Remote:   target.Remote,
			FilePath: dest,
		})

		log.Printf("Done with target %s\n", targetName)
	}

	c.BackupList.cleanUp()
}

func (c *config) writeToFileFirst(target target, outPath string) (*os.File, error) {
	log.Printf("Saving backup at %s", outPath)

	file, err := os.Create(outPath)

	if err != nil {
		return nil, fmt.Errorf("Failed to create output file %w", err)
	}

	err = c.createArchive(target.Path, file)

	if err != nil {
		return nil, err
	}

	return file, nil
}

func (c *config) createArchive(sourceDir string, output io.Writer) error {
	switch c.Archive {
	case "tar":
		buff := output

		switch c.Compression {
		case "zstd":
			writer, err := zstd.NewWriter(output, zstd.WithEncoderLevel(zstd.SpeedBestCompression))

			if err != nil {
				return err
			}

			buff = writer
			defer writer.Close()
		case "gzip":
			writer, err := gzip.NewWriterLevel(output, int(c.CompressionLevel))

			if err != nil {
				return err
			}

			buff = writer
			defer writer.Close()
		}

		return createTar(sourceDir, buff)
	case "zip":
		return createZip(sourceDir, output, c.Compression)
	}

	return fmt.Errorf("Unrecognized archive format \"%s\"", c.Archive)
}

func createZip(sourceDir string, output io.Writer, compression string) error {
	zipWriter := zip.NewWriter(output)
	defer zipWriter.Close()

	return filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return fmt.Errorf("Failed to create tar header: %w", err)
		}

		relPath, err := filepath.Rel(sourceDir, path)
		if err != nil {
			return fmt.Errorf("Failed to get relative path: %w", err)
		}
		header.Name = relPath

		if compression == "deflate" {
			header.Method = zip.Deflate
		}

		writer, err := zipWriter.CreateHeader(header)
		if err != nil {
			return fmt.Errorf("Failed to write zip header: %w", err)
		}

		if info.IsDir() {
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			return fmt.Errorf("Failed to open file: %w", err)
		}
		defer file.Close()

		_, err = io.Copy(writer, file)
		return err
	})
}

// Archive with tar
func createTar(sourceDir string, output io.Writer) error {
	tarWriter := tar.NewWriter(output)
	defer tarWriter.Close()

	return filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		header, err := tar.FileInfoHeader(info, info.Name())
		if err != nil {
			return fmt.Errorf("Failed to create tar header: %w", err)
		}

		relPath, err := filepath.Rel(sourceDir, path)
		if err != nil {
			return fmt.Errorf("Failed to get relative path: %w", err)
		}
		header.Name = filepath.ToSlash(relPath)

		if err := tarWriter.WriteHeader(header); err != nil {
			return fmt.Errorf("failed to write tar header: %w", err)
		}

		if info.IsDir() {
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			return fmt.Errorf("failed to open file: %w", err)
		}
		defer file.Close()

		_, err = io.Copy(tarWriter, file)
		return err
	})
}
