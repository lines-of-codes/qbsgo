package main

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"os/exec"
)

// Returns: Destination URL, Error
func (c *config) copypartyUpload(remoteName string, inputFile string, fileName string) (string, error) {
	remote := c.Remotes[remoteName]
	script := remote.Script

	if script == "" {
		script = "u2c"
	}

	password := ""

	if remote.User != "" {
		remote.User += ":"
	}

	if remote.Password != "" {
		password = fmt.Sprintf("-a %s%s", remote.User, remote.Password)
	}

	dest, err := url.JoinPath(remote.Root, remote.DestDir)

	if err != nil {
		return "", fmt.Errorf("Error while URL is being joined: %w", err)
	}

	destWithFile, err := url.JoinPath(dest, fileName)

	if err != nil {
		return "", fmt.Errorf("Error while URL is being joined: %w", err)
	}

	cmd := exec.Command(script, password, dest, inputFile)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	log.Printf("Running command: %s", cmd.String())

	return destWithFile, cmd.Run()
}
