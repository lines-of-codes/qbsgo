package main

import (
	"fmt"
	"os"
)

func genService(name string) (string, error) {
	exe, err := os.Executable()

	if err != nil {
		return "", err
	}

	return fmt.Sprintf(`[Unit]
Description=Backups %s through QBS
Wants=network-online.target
After=network-online.target

[Service]
User=qbs
Group=qbs
ExecStart=%s -targets %s -backup`, name, exe, name), nil
}

func genTimer(name string, target target) string {
	return fmt.Sprintf(`[Unit]
Description=Triggers a backup of target %s through QBS

[Timer]
OnCalendar=%s
Persistent=true

[Install]
WantedBy=timers.target`, name, target.Interval)
}
