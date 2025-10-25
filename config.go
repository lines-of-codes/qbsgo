package main

import (
	"log"
	"os"
	"os/exec"

	"github.com/BurntSushi/toml"
)

type (
	config struct {
		// A value of either "tar" or "zip"
		Archive string

		// The directory to save archives to.
		ArchiveDir string

		// Values valid for tar: none, gzip, and zstd
		// Values valid for zip: none, deflate
		Compression string

		// Compression level for gzip
		// Should be in the range of 1 to 9
		CompressionLevel uint8

		// Whether to delete the backup archive after it is uploaded or not.
		DeleteAfterUpload bool

		// WriteToFileFirst bool

		IdLength int
		Remotes  map[string]remote
		Targets  map[string]target
	}

	remote struct {
		Type     string
		Root     string
		User     string
		Password string
		Script   string
		DestDir  string
	}

	target struct {
		Path     string
		Remote   string
		Interval string
	}
)

const DEFAULT_CUID_LENGTH = 8

func loadConfig(config *config, validate bool) {
	configPath := "/etc/qbsgo.toml"

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		configPath = "./qbsgo.toml"
	}

	_, err := toml.DecodeFile(configPath, config)

	if err != nil {
		log.Fatal(err.Error())
	}

	if config.IdLength == 0 {
		config.IdLength = DEFAULT_CUID_LENGTH
	}

	if !validate {
		return
	}

	for _, target := range config.Targets {
		log.Printf("Validating target \"%s\"\n", target.Path)
		cmd := exec.Command("systemd-analyze", "calendar", target.Interval)
		cmd.Stdout = os.Stdout
		if err := cmd.Run(); err != nil {
			log.Printf("Interval value check failed: %s", err)
			log.Fatalf("Invalid interval value \"%s\" for target \"%s\"\n", target.Interval, target.Path)
		}
	}
}
