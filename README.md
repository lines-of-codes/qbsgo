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

## Configuration

QBSGo uses a TOML configuration file named `qbsgo.toml`. It expects the file
to be at `/etc/qbsgo.toml` or in the same directory as the binary

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

`interval` is any valid value for systemd timers' `OnCalendar` value.
Most commonly, you'll be using magic values such as `daily`, `weekly`, or
`monthly`. See
[systemd.time(7)](https://man.archlinux.org/man/systemd.time.7#CALENDAR_EVENTS)
for more information.

## Building from source

1. Install [Go](https://go.dev/)
2. Clone this repository
3. Run `go build` at the project's root
