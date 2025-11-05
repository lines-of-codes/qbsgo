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

## Flags

- `-targets targetA,targetB`: A comma seperated list of targets. "all" can be
  used to select every target in the configuration file.
- `-backup`: Triggers a backup for the specified targets
- `-install`: Install systemd Timers to trigger backups periodically.
- `-version`: Prints the version of the program and exit.

## Systemd Timers

Systemd timers can be installed by running QBSGo with the install flag. For
example:

```bash
qbsgo -targets all -install
```

Running QBSGo as root will install the unit files onto the system at
`/etc/systemd/system` and running it as other users will install it
at `/home/USER/.config/systemd/user`

To uninstall those unit files, You can stop & disable them manually then delete
them at the directory specified above, or you can run the install command again,
instruct the program to clean up old unit files, and press Ctrl+C when the save
options pops up. The process would generally go as follows:

```
❯ ./qbsgo -targets all -install
2025/11/02 16:59:59 Validating target "/var/lib/qsm-web/servers/PaperTest/"
  Original form: weekly
Normalized form: Mon *-*-* 00:00:00
    Next elapse: Mon 2025-11-03 00:00:00 +07
       (in UTC): Sun 2025-11-02 17:00:00 UTC
       From now: 7h left
Unit files will be installed to /home/linesofcodes/.config/systemd/user
Do you wish to clean up existing QBS unit files? (if there is any) [Y/n] Y
The following units will be disabled:
qbsgo-generated-weekly.timer
Do you wish to continue? [Y/n] Y
Running: systemctl --user disable --now qbsgo-generated-weekly.timer
The following files will be deleted:
/home/linesofcodes/.config/systemd/user/qbsgo-generated-weekly.service
/home/linesofcodes/.config/systemd/user/qbsgo-generated-weekly.timer
Do you wish to delete the unit files? [Y/n] Y
Deleting: /home/linesofcodes/.config/systemd/user/qbsgo-generated-weekly.service
Deleting: /home/linesofcodes/.config/systemd/user/qbsgo-generated-weekly.timer
Done deleting.
Using vim as the text editor. Set the EDITOR environment variable to use something else.

Interval: weekly
Target(s): PaperTest
This will generate the following files:
/home/linesofcodes/.config/systemd/user/qbsgo-generated-weekly.service
/home/linesofcodes/.config/systemd/user/qbsgo-generated-weekly.timer
Please choose an action:
[r]eview/[e]dit/[s]ave/save [a]ll ^C
```

**Note:** `^C` represents when Ctrl+C was pressed.

## Backup List file

QBSGo can store a list of backups. It is disabled by default and can be enabled
through the configuration file in the `backupList` section.

The backup list file is named `backuplist.json` and is stored in the same
directory as the configuration file

The backup list is not guaranteed to be accurate, as the backup file could
be deleted or renamed on the remote file storage system and QBS wouldn't know.

## Configuration

QBSGo uses a TOML configuration file named `qbsgo.toml`. It expects the file
to be at `/etc/qbsgo/qbsgo.toml` or in the same directory as the binary

An example configuration file can be found at [`qbsgo.example.toml`](./qbsgo.example.toml).

### Global Options

`archive`

A string value of either `tar` or `zip`

`archiveDir`

The directory to save the archive to. In most cases, `/tmp` is fine unless you
disabled `deleteAfterUpload` and keep local copies.

`compression`

- Valid values for Zip archives: `none` or `deflate`
- Valid values for Tar archives: `none`, `gzip`, or `zstd`

`compressionLevel`

A value from 1 to 9, Only applicable for gzip.

Zstandard always use the "best compression" setting, roughly equivalent to
level 11. It is a limitation of the used Zstandard library that it cannot go
up to level 22 like the original implementation.

`deleteAfterUpload`

If `true`, After the backup archive has been uploaded, The local archive will
be deleted.

### `backupList`

```toml
[backupList]
# Whether the feature is enabled
enabled = true

# Whether to clean up older entries
cleanEntries = true

# When `cleanEntries` is enabled, If an entry is older than the specified
# duration, It is forgotten. The following number suffixes are supported:
# y, m, w, d, which are year, month, week, and day respectively.
# To specify something like 1 year 1 month, You can do "1y 1m". Numbers
# are seperated by space, so "1y1m" is invalid.
olderThan = "1m"
```

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

#### password

The password field can optionally refer to a file instead of directly
specifying the password in the config file.

To refer to a file, use a `file:` prefix. For example:

```toml
# Refers to a file within the same working directory as QBS
password = "file:password.txt"
# Absolute paths also work!
password = "file:/home/user/Documents/password.txt"
```

Note that this isn't the regular URI style, So it's not `file://`

### Targets

An example target:
```toml
[targets.PaperTest]
path = "/var/lib/qsm-web/servers/PaperTest/"
remote = "copyparty"
interval = "weekly"
```

`remote` is the name of a remote you named.

`interval` is any valid value for systemd timers' `OnCalendar` value.
Most commonly, you'll be using magic values such as `daily`, `weekly`, or
`monthly`. See
[systemd.time(7)](https://man.archlinux.org/man/systemd.time.7#CALENDAR_EVENTS)
for more information.

## Building from source

1. Install [Go](https://go.dev/)
2. Clone this repository
3. Run `go build` at the project's root

## License

This project is licensed under GNU GPLv3.
