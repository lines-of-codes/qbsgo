package main

import (
	"log"
	"os"
	"os/exec"
	"strings"

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

		BackupList backupList

		IdLength int
		Remotes  map[string]remote
		Targets  map[string]target

		// For internal reference
		configPath string
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

	backupList struct {
		Enabled      bool
		CleanEntries bool
		OlderThan    string
	}
)

const DEFAULT_CUID_LENGTH = 8
const FILE_PREFIX = "file:"

var AppFileDir = "/etc/qbsgo/"

func loadConfig(config *config, validate bool) {
	config.configPath = "/etc/qbsgo/qbsgo.toml"

	if _, err := os.Stat(config.configPath); os.IsNotExist(err) {
		AppFileDir = "./"
		config.configPath = "./qbsgo.toml"
	}

	_, err := toml.DecodeFile(config.configPath, config)

	if err != nil {
		log.Fatal(err.Error())
	}

	if config.IdLength == 0 {
		config.IdLength = DEFAULT_CUID_LENGTH
	}

	for remoteName, remote := range config.Remotes {
		if !strings.HasPrefix(remote.Password, FILE_PREFIX) {
			continue
		}

		filePath := remote.Password[len(FILE_PREFIX):]
		contents, err := os.ReadFile(filePath)

		if err != nil {
			log.Fatalf("Unable to read password file for remote \"%s\"", remoteName)
		}

		remote.Password = strings.TrimSpace(string(contents))
		config.Remotes[remoteName] = remote
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
