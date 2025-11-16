package main

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/nrednav/cuid2"
)

const UNIT_NAME_PREFIX = "qbsgo-generated-"
const SEPERATOR = "----------"

var unitFilesLocation = "/etc/systemd/system"
var filePerms = os.FileMode(0644)
var operationMode = "--system"

func (c *config) install(targets []string, dontAsk bool) {
	currentUser, err := user.Current()

	if err != nil {
		log.Fatalf("Error while getting current user: %s", err)
	}

	if currentUser.Username != "root" {
		unitFilesLocation = fmt.Sprintf("/home/%s/.config/systemd/user", currentUser.Username)
		operationMode = "--user"
		err := os.MkdirAll(unitFilesLocation, filePerms)

		if err != nil {
			log.Fatalf("Unable to create the user system units folder: %s", err)
		}
	}

	fmt.Printf("Unit files will be installed to %s\n", unitFilesLocation)
	fmt.Print("Do you wish to clean up existing QBS unit files? (if there is any) [Y/n] ")

	answer := askOrFallback("y", dontAsk)

	if strings.ToLower(answer) == "y" {
		cleanUnits(dontAsk)
	} else {
		fmt.Println("Note: Existing QBS unit files may be overwritten.")
	}

	textEditor, found := os.LookupEnv("EDITOR")

	if !found {
		textEditor = "vim"
	}

	fmt.Printf("Using %s as the text editor. Set the EDITOR environment variable to use something else.\n\n", textEditor)

	username := ""

	if operationMode == "--system" {
		fmt.Println("In the system service files, Do you want the backup to run as a specific user?")
		fmt.Println("Enter the wanted username or enter nothing to run as root.")
		fmt.Println("Note: QBS will assume the user has a group of the same name and the service will run with that group.")
		fmt.Print("> ")
		username = askOrFallback("", dontAsk)

		fmt.Println()
	}

	intervals := make(map[string][]string)

	for _, targetName := range targets {
		target := c.Targets[targetName]

		intervals[target.Interval] = append(intervals[target.Interval], targetName)
	}

	saveAll := false
	for interval, targetList := range intervals {
		fmt.Printf("Interval: %s\nTarget(s): %s\n", interval, strings.Join(targetList, ", "))

		fmt.Println("This will generate the following files:")

		fileName := intervalOrServerNames(interval, targetList)

		serviceFile := fmt.Sprintf("%s/%s%s.service", unitFilesLocation, UNIT_NAME_PREFIX, fileName)
		fmt.Println(serviceFile)

		timerFile := fmt.Sprintf("%s/%s%s.timer", unitFilesLocation, UNIT_NAME_PREFIX, fileName)
		fmt.Println(timerFile)

		serviceUnit, err := c.genService(targetList, username)

		if err != nil {
			log.Fatalf("Error while generating service unit for interval \"%s\": %s", interval, err)
		}

		timerUnit := genTimer(targetList, interval)

		if saveAll {
			saveTimerFiles(serviceFile, serviceUnit, timerFile, timerUnit)
			continue
		}

		finish := false

		for !finish {
			fmt.Println("Please choose an action:")
			fmt.Print("[r]eview/[e]dit/[s]ave/save [a]ll ")

			answer := strings.ToLower(askOrFallback("a", dontAsk))

			switch answer {
			case "r":
				fmt.Println(SEPERATOR)
				fmt.Println(serviceFile)
				fmt.Println(SEPERATOR)
				fmt.Println(serviceUnit)
				fmt.Println(SEPERATOR)
				fmt.Println(timerFile)
				fmt.Println(SEPERATOR)
				fmt.Println(timerUnit)
				fmt.Println(SEPERATOR)
			case "e":
				serviceUnit, timerUnit = editUnitFiles(serviceUnit, timerUnit, textEditor)
			case "s":
				saveTimerFiles(serviceFile, serviceUnit, timerFile, timerUnit)
				finish = true
			case "a":
				saveTimerFiles(serviceFile, serviceUnit, timerFile, timerUnit)
				finish = true
				saveAll = true
			}
		}
	}

	fmt.Printf("Running \"systemctl %s daemon-reload\"...\n", operationMode)
	reloadCmd := exec.Command("systemctl", operationMode, "daemon-reload")

	if err := reloadCmd.Run(); err != nil {
		log.Printf("Command \"systemctl daemon-reload\" failed: %s", err)
	}

	fmt.Print("Do you wish to enable and start these timers right away? [Y/n] ")
	answer = strings.ToLower(askOrFallback("y", dontAsk))

	if answer != "y" {
		return
	}

	for interval, _ := range intervals {
		timerName := fmt.Sprintf("%s%s.timer", UNIT_NAME_PREFIX, interval)
		fmt.Printf("Running: systemctl %s enable --now %s\n", operationMode, timerName)
		cmd := exec.Command("systemctl", operationMode, "enable", "--now", timerName)

		if err := cmd.Run(); err != nil {
			fmt.Printf("Enabling and starting unit failed for unit \"%s\". Because: %s\n", timerName, err)
		}
	}
}

func askOrFallback(fallback string, dontAsk bool) string {
	if dontAsk {
		fmt.Println(fallback)
		return fallback
	}

	var answer string
	_, err := fmt.Scanln(&answer)

	if err != nil {
		log.Fatalf("Error while reading user input: %s", err)
	}

	return answer
}

func intervalOrServerNames(interval string, names []string) string {
	if strings.ContainsAny(interval, " *:/") {
		return strings.Join(names, "")
	}

	return interval
}

func saveTimerFiles(servicePath, service, timerPath, timer string) {
	os.WriteFile(servicePath, []byte(service), filePerms)
	os.WriteFile(timerPath, []byte(timer), filePerms)
}

func editUnitFiles(service string, timer string, editor string) (newService string, newTimer string) {
	fmt.Println("Would you like to edit the service file or the timer file?")
	fmt.Print("[s]ervice/[t]imer/[c]ancel ")

	var answer string
	fmt.Scanln(&answer)
	answer = strings.ToLower(answer)

	id := cuid2.Generate()
	switch answer {
	case "c":
		return service, timer
	case "s":
		fileName := fmt.Sprintf("/tmp/%s.service", id)
		return editFile(editor, fileName, service), timer
	case "t":
		fileName := fmt.Sprintf("/tmp/%s.timer", id)
		return service, editFile(editor, fileName, timer)
	default:
		fmt.Printf("Warning: Invalid input, Expected c, s, or t. Received: %s\n", answer)
		return service, timer
	}
}

func editFile(editor, filePath, original string) string {
	err := os.WriteFile(filePath, []byte(original), filePerms)

	if err != nil {
		log.Fatalf("Cannot create temporary file for editing: %s", err)
	}

	fmt.Printf("Running command: %s %s\n", editor, filePath)
	cmd := exec.Command(editor, filePath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if err := cmd.Run(); err != nil {
		fmt.Printf("Warning: Running text editor command resulted in an error: %s", err)
	}

	contents, err := os.ReadFile(filePath)

	if err != nil {
		log.Printf("Error while reading temporary file, Changes are not saved. Cause: %s", err)
		return original
	}

	os.Remove(filePath)

	return string(contents)
}

func cleanUnits(dontAsk bool) {
	var fileList []string
	toBeDisabled := make(map[string]struct{})

	fmt.Println("The following units will be disabled:")

	filepath.WalkDir(unitFilesLocation, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		name := entry.Name()

		if strings.HasPrefix(name, UNIT_NAME_PREFIX) {
			fileList = append(fileList, path)
		}

		_, planned := toBeDisabled[name]
		if !planned && strings.HasSuffix(name, ".timer") {
			toBeDisabled[name] = struct{}{}
			fmt.Println(name)
		}

		return nil
	})

	if len(fileList) == 0 {
		fmt.Println("No QBSGo unit files found. Continuing.")
		return
	}

	if len(toBeDisabled) == 0 {
		fmt.Println("No timers found.")
	} else {
		fmt.Print("Do you wish to continue? [Y/n] ")

		answer := askOrFallback("y", dontAsk)

		if strings.ToLower(answer) != "y" {
			return
		}

		for unit := range toBeDisabled {
			fmt.Printf("Running: systemctl %s disable --now %s\n", operationMode, unit)
			cmd := exec.Command("systemctl", operationMode, "disable", "--now", unit)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			cmd.Stdin = os.Stdin

			if err := cmd.Run(); err != nil {
				fmt.Printf("Skipping unit, Unable to disable unit %s: %s\n", unit, err)
			}
		}
	}

	fmt.Println("The following files will be deleted:")

	var newFileList []string
	filepath.WalkDir(unitFilesLocation, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		name := entry.Name()

		if strings.HasPrefix(name, UNIT_NAME_PREFIX) {
			newFileList = append(newFileList, path)
			fmt.Println(path)
		}

		return nil
	})

	fmt.Print("Do you wish to delete the unit files? [Y/n] ")

	answer := askOrFallback("y", dontAsk)

	if strings.ToLower(answer) != "y" {
		return
	}

	for _, file := range newFileList {
		fmt.Printf("Deleting: %s\n", file)
		os.Remove(file)
	}

	fmt.Println("Done deleting.")
}

func (c *config) genService(names []string, user string) (string, error) {
	exe, err := os.Executable()

	if err != nil {
		return "", err
	}

	name := strings.Join(names, ",")

	additionalInfo := ""

	if user != "" {
		additionalInfo = fmt.Sprintf("\nUser=%s\nGroup=%s", user, user)
	}

	if c.configPath == "./qbsgo.toml" {
		exe, err := os.Executable()

		if err != nil {
			log.Fatalf("Unable to get the executable's path: %s", err)
		}

		additionalInfo += fmt.Sprintf("\nWorkingDirectory=%s", filepath.Dir(exe))
	}

	return fmt.Sprintf(`[Unit]
Description=Backups %s through QBS
Wants=network-online.target
After=network-online.target

[Service]%s
ExecStart=%s -targets %s -backup`, name, additionalInfo, exe, name), nil
}

func genTimer(names []string, interval string) string {
	name := strings.Join(names, ", ")

	return fmt.Sprintf(`[Unit]
Description=Trigger backups for target(s): %s

[Timer]
OnCalendar=%s
Persistent=true

[Install]
WantedBy=timers.target`, name, interval)
}
