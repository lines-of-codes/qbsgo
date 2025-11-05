package main

import (
	"flag"
	"log"
	"os"
	"runtime/debug"
	"strings"
)

var commit = func() string {
	if info, ok := debug.ReadBuildInfo(); ok {
		for _, setting := range info.Settings {
			if setting.Key == "vcs.revision" {
				return setting.Value
			}
		}
	}

	return ""
}()

func main() {
	versionFlag := flag.Bool("version", false, "Prints the version of the program and exit.")
	targetsFlag := flag.String("targets", "", "A comma seperated list of targets. \"all\" can be specified to select every target in the configuration file.")
	backupFlag := flag.Bool("backup", false, "Whether to backup the specified targets or not")
	installFlag := flag.Bool("install", false, "Install the systemd service & timer for the specified target(s).")

	flag.Parse()

	if *versionFlag {
		log.Printf("version 1.0.0 (commit %s)\n", commit)
		os.Exit(0)
	}

	var config config
	loadConfig(&config, *installFlag)

	targets := strings.Split(*targetsFlag, ",")

	if len(targets) == 0 {
		log.Fatalln("No target specified. Please specify them through the -targets flag.")
	}

	if targets[0] == "all" {
		targets = make([]string, len(config.Targets))

		i := 0
		for k := range config.Targets {
			targets[i] = k
			i++
		}
	}

	if *installFlag {
		config.install(targets)
	}

	if *backupFlag {
		config.backup(targets)
	}
}
