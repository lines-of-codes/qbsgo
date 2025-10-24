package main

import (
	"flag"
	"log"
	"os"
	"strings"
)

func (c *config) install(targets []string) {
	for _, target := range targets {
		log.Println(genService(target))
		log.Println(genTimer(target, c.Targets[target]))
	}
}

func main() {
	versionFlag := flag.Bool("version", false, "Prints the version of the program and exit.")
	targetsFlag := flag.String("targets", "", "A comma seperated list of targets")
	backupFlag := flag.Bool("backup", false, "Whether to backup the specified targets or not")
	installFlag := flag.Bool("install", false, "Install the systemd service & timer for the specified target(s).")

	flag.Parse()

	if *versionFlag {
		log.Println("version 0.0.1")
		os.Exit(0)
	}

	var config config
	loadConfig(&config, *installFlag)

	targets := strings.Split(*targetsFlag, ",")

	if len(targets) == 0 {
		log.Fatalln("No target specified. Please specify them through the -targets flag.")
	}

	if *installFlag {
		config.install(targets)
	}

	if *backupFlag {
		config.backup(targets)
	}
}
