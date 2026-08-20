# classreach

[![CI](https://github.com/jwmoss/classreach/actions/workflows/ci.yml/badge.svg)](https://github.com/jwmoss/classreach/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/v/release/jwmoss/classreach)](https://github.com/jwmoss/classreach/releases/latest)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

Command-line client for the ClassReach private API.

## Install

### Go

```bash
go install github.com/jwmoss/classreach/cmd/classreach@latest
```


### Homebrew

```bash
brew tap jwmoss/tap
brew install --cask jwmoss/tap/classreach
```


### Source

```bash
git clone https://github.com/jwmoss/classreach.git
cd classreach
make build
./bin/classreach version
```

## Configuration

Initialize a config file without putting the password in shell history:

```bash
printf '%s\n' "$CLASSREACH_PASSWORD" | classreach config init \
  --username guardian@example.com \
  --password-stdin
```

Show the platform-specific config path:

```bash
classreach config show
```

The generated file uses mode `0600`:

```yaml
base_url: https://providencewilmington.classreach.com
origin_host: classreach.azurewebsites.net
username: guardian@example.com
password: your-password
```

Environment variables:

| Variable | Purpose |
| --- | --- |
| `CLASSREACH_BASE_URL` | ClassReach tenant URL |
| `CLASSREACH_ORIGIN_HOST` | Azure origin host |

Precedence:

```text
flags > environment > config file > defaults
```

The CLI connects to the Azure origin while it preserves the configured tenant hostname. This
allows direct API access without browser automation. `config show` redacts the credentials.

## Usage

```bash
classreach --help
classreach --version
classreach version
classreach login
classreach doctor --json
classreach overview --json
classreach agenda download --week 2026-08-17 --output agenda
classreach students list --json
classreach courses list --json
classreach assignments list --student <student-id> --section <section-id> --json
classreach grades list --student <student-id> --json
classreach attendance list --student <student-id> --section <section-id> --json
classreach messages list --json
classreach messages get <thread-id> --json
classreach messages download <thread-id> <file-id> --output attachment.pdf
classreach documents list --json
classreach announcements list --json
classreach calendar list --start 2026-08-20 --end 2026-09-20 --json
classreach directory list --json
classreach directory families --search NAME --json
classreach raw get /Home/GetQuickView --query weekDate=2026-08-17T00:00:00 --json
classreach completion zsh > ~/.zfunc/_classreach
```

`messages get` marks an unread thread as read in ClassReach. Document downloads require an
explicit output path and use mode `0600`.

## Agent skill

The repository includes `.agents/skills/classreach/SKILL.md`. It follows the SkyCLI skill pattern
and directs OpenClaw to use stable JSON and read-only commands.

Use this repository as the agent workspace or install that skill directory through the host's
normal skill installation method.

## Global Flags

| Flag | Description |
| --- | --- |
| `--config` | Config file path |
| `--base-url` | ClassReach tenant URL override |
| `--origin-host` | ClassReach Azure origin host override |
| `--version` | Print version information |
| `--json` | Emit JSON to stdout |
| `--plain` | Emit stable plain text where available |
| `--quiet`, `-q` | Suppress non-essential output |
| `--no-color` | Disable color |
| `--timeout` | HTTP timeout |
| `--trace-http` | Log HTTP method/path/status to stderr |
| `--dry-run` | Refuse non-GET HTTP requests |
| `--no-input` | Disable interactive prompts |

## Exit Codes

| Code | Meaning |
| --- | --- |
| 0 | Success |
| 1 | Runtime error |
| 2 | Invalid usage |

## Development

```bash
make check
```

## Release

After the release PR merges, update local `main` and tag the merge commit:

```bash
git switch main
git pull --ff-only origin main
git tag v0.1.0
git push origin v0.1.0
```

The release workflow uses GoReleaser to publish archives and checksums.
Set `HOMEBREW_TAP_TOKEN` before the first tagged release so GoReleaser can
update `jwmoss/homebrew-tap`.
