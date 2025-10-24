# QBSGo

Quick Backup System is a simple backup system written in Go, designed for use
with the Quick Server Manager (QSM) software. (But it'll also work for other use
cases)

Although is possible, QBS is not meant for creating local archives. You can
simply trigger `tar` periodically for that.

## Supported Configurations

The following configuration is best supported and other software setup may
require workarounds

- Linux: The program assumes Linux paths by default, configure new paths accordingly in the config file.
- systemd: Required for the `-install` flag, easiest workaround is to trigger backups through cron.
- Only Nextcloud and copyparty is supported as a backup upload destination.

## Triggering backups manually

(also for people who prefers cron)

You can use the command in the format of:
```
qbsgo -targets targetA,targetB -backup
```

## Configuration

QBSGo uses a TOML configuration file named `qbsgo.toml`. It expects the file
to be at `/etc/qbsgo.toml` or in the same directory as the binary

### Remotes

Remotes are backup upload destinations.

QBSGo currently supports Nextcloud and copyparty as an upload destination.

The example below shows how to create a Nextcloud remote
```toml
[remotes.nextcloud]
type = "nextcloud"
root = "https://nextcloud.example.com"
destDir = "Backups" # (optional)
user = "johndoe"
password = "AppPassword"
```

#### copyparty

copyparty can be used with the options available above but with the type value
changed to `copyparty`

Notes:

- Do not provide the `user` value if your copyparty server does not use the
  `--usernames` flag
- QBS relies on the `u2c` script. Please make them available in PATH or provide
  them with the `script` key

### Targets

An example target:
```toml
[targets.PaperTest]
path = "/var/lib/qsm-web/servers/PaperTest/"
remote = "copyparty"
interval = "weekly"
```

`remote` is the name of a remote you named.

`interval` is any valid value for systemd timers' `OnCalendar` value. See
[systemd.time(7)](https://man.archlinux.org/man/systemd.time.7#CALENDAR_EVENTS)
for more information.
