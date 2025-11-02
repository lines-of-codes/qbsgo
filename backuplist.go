package main

import (
	"encoding/json"
	"errors"
	"io/fs"
	"log"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/gofrs/flock"
)

type listEntry struct {
	Id       string
	Remote   string
	FilePath string
	Date     string
}

const LIST_FILE_NAME = "backuplist.json"

// Appends a new backup to the backup list.
// Blocking function, Exits immediately if it encounters an error.
func (b *backupList) append(newBackup listEntry) {
	if !b.Enabled {
		return
	}

	listFile := path.Join(AppFileDir, LIST_FILE_NAME)
	fileLock := flock.New(listFile + ".lock")

	log.Println("Locking the list file. This is a blocking operation.")

	err := fileLock.Lock()

	if err != nil {
		log.Fatalf("Unable to obtain list file lock: %s", err)
	}

	log.Println("File locked.")
	defer fileLock.Unlock()

	content, err := os.ReadFile(listFile)

	var listEntries []listEntry
	if !errors.Is(err, fs.ErrNotExist) {
		if err != nil {
			log.Fatalf("Unable to read list file: %s", err)
		}

		err = json.Unmarshal(content, &listEntries)

		if err != nil {
			log.Fatalf("Unable to parse JSON: %s", err)
		}
	}

	listEntries = append(listEntries, newBackup)

	newContent, err := json.Marshal(listEntries)

	if err != nil {
		log.Fatalf("Unable to encode to JSON: %s", err)
	}

	err = os.WriteFile(listFile, newContent, 0644)

	if err != nil {
		log.Fatalf("Unable to write to list file: %s", err)
	}
}

func (b *backupList) cleanUp() {
	if !b.Enabled || !b.CleanEntries {
		return
	}

	listFile := path.Join(AppFileDir, LIST_FILE_NAME)
	fileLock := flock.New(listFile + ".lock")

	log.Println("Locking the list file. This is a blocking operation.")

	err := fileLock.Lock()

	if err != nil {
		log.Fatalf("Unable to obtain list file lock: %s", err)
	}

	log.Println("File locked.")
	defer fileLock.Unlock()

	content, err := os.ReadFile(listFile)

	if err != nil {
		log.Fatalf("Unable to read list file: %s", err)
	}

	var listEntries []listEntry
	err = json.Unmarshal(content, &listEntries)

	if err != nil {
		log.Fatalf("Unable to parse JSON: %s", err)
	}

	newContent, err := json.Marshal(b.cleanList(listEntries))

	if err != nil {
		log.Fatalf("Unable to encode to JSON: %s", err)
	}

	err = os.WriteFile(listFile, newContent, 0644)

	if err != nil {
		log.Fatalf("Unable to write to list file: %s", err)
	}
}

// Not meant for public usage
func (b *backupList) cleanList(entries []listEntry) []listEntry {
	var newList []listEntry
	oldDate := time.Now()

	olderThanRaw := strings.SplitSeq(b.OlderThan, " ")

	for olderThanSect := range olderThanRaw {
		num, err := strconv.Atoi(olderThanSect[:len(olderThanSect)-1])

		if err != nil {
			log.Fatalf("Unable to parse integer \"%s\": %s\n", olderThanSect[:len(olderThanSect)-1], err)
		}

		switch olderThanSect[len(olderThanSect)-1:] {
		case "y":
			oldDate = oldDate.AddDate(-num, 0, 0)
		case "m":
			oldDate = oldDate.AddDate(0, -num, 0)
		case "w":
			oldDate = oldDate.AddDate(0, 0, -(num * 7))
		case "d":
			oldDate = oldDate.AddDate(0, 0, -num)
		}
	}

	log.Printf("Backups older than %s will be forgotten\n", oldDate.Format(time.DateTime))

	for _, entry := range entries {
		backupDate, err := time.Parse(time.RFC3339, entry.Date)

		if err != nil {
			log.Printf("Unable to parse date \"%s\", Skipping entry: %s\n", entry.Date, err)
			continue
		}

		if backupDate.After(oldDate) {
			newList = append(newList, entry)
		}
	}

	return newList
}
