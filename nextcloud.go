package main

import (
	"fmt"
	"io"
	"log"
	"net/url"
	"os"
	"path"
	"strconv"

	"github.com/nrednav/cuid2"
	"github.com/studio-b12/gowebdav"
)

const MEBIBYTE = 1024 * 1024
const DEFAULT_CHUNK_SIZE = 50 * MEBIBYTE // 50 MiB
const FILE_MODE = 0644

// See https://docs.nextcloud.com/server/stable/developer_manual/client_apis/WebDAV/chunking.html
// for how Nextcloud does its chunking

func (c *config) nextcloudUpload(remoteName string, inputFile string, fileName string) error {
	remote := c.Remotes[remoteName]

	prefixUrl, err := url.JoinPath(remote.Root, "remote.php/dav")

	if err != nil {
		return fmt.Errorf("Error while joining prefix URL: %w", err)
	}

	client := gowebdav.NewClient(prefixUrl, remote.User, remote.Password)

	destUrl, err := url.JoinPath(prefixUrl, "files", remote.User, remote.DestDir, fileName)

	if err != nil {
		return fmt.Errorf("Error while joining destination URL: %w", err)
	}

	err = client.Connect()

	if err != nil {
		return fmt.Errorf("Error while connecting to Nextcloud server (Remote \"%s\"): %w", remoteName, err)
	}

	file, err := os.Open(inputFile)

	if err != nil {
		return fmt.Errorf("Error while opening input file: %w", err)
	}

	defer file.Close()

	fileStat, err := file.Stat()

	if err != nil {
		return fmt.Errorf("Error while getting file information: %w", err)
	}

	fileSize := fileStat.Size()

	client.SetHeader("OC-Total-Length", strconv.FormatInt(fileSize, 10))

	chunksFolder := fmt.Sprintf("uploads/%s/qbsgo-%s", remote.User, cuid2.Generate())
	err = client.Mkdir(chunksFolder, FILE_MODE)

	if err != nil {
		return fmt.Errorf("Error while creating chunk folder: %w", err)
	}

	var offset int64 = 0
	chunkNum := 1

	for offset < fileSize {
		thisChunkSize := DEFAULT_CHUNK_SIZE
		if offset+int64(DEFAULT_CHUNK_SIZE) > fileSize {
			thisChunkSize = int(fileSize - offset)
		}

		chunk := make([]byte, thisChunkSize)
		bytesRead, err := file.ReadAt(chunk, offset)

		if err != nil && err != io.EOF {
			return fmt.Errorf("Error while reading chunk: %w", err)
		}

		chunkPath := fmt.Sprintf("%s/%05d", chunksFolder, chunkNum)
		err = client.Write(chunkPath, chunk[:bytesRead], FILE_MODE)

		if err != nil {
			return fmt.Errorf("Failed to upload chunk %d: %w", chunkNum, err)
		}

		offset += int64(bytesRead)
		log.Printf("Uploaded chunk %d successfully. (%d/%d MiB, %.2f%%)\n", chunkNum, offset/MEBIBYTE, fileSize/MEBIBYTE, float32(offset)/float32(fileSize)*100)

		chunkNum++
	}

	err = client.Rename(fmt.Sprintf("%s/.file", chunksFolder), path.Join("files", remote.User, remote.DestDir, fileName), true)

	if err != nil {
		return fmt.Errorf("Error while assembling file chunks: %w", err)
	}

	log.Printf("Upload to Nextcloud target \"%s\" completed. File is uploaded to %s\n", remoteName, destUrl)
	return nil
}
