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

Initialize a config file:

```bash
classreach config init --base-url https://providencewilmington.classreach.com
```

Config path:

```text
$XDG_CONFIG_HOME/classreach/config.yaml
```

Environment variables:

| Variable | Purpose |
| --- | --- |
| `CLASSREACH_BASE_URL` | API base URL |
| `CLASSREACH_TOKEN` | API token |
| `CLASSREACH_AUTH_HEADER` | Auth header name |
| `CLASSREACH_AUTH_SCHEME` | Auth scheme, for example `Bearer` |

Precedence:

```text
flags > environment > config file > defaults
```

Tokens are intentionally not accepted as command-line flags by default. Use the
environment, a `0600` config file, or stdin-backed setup.

## Usage

```bash
classreach --help
classreach --version
classreach version
classreach doctor
classreach students list
classreach students get 123
classreach raw GET /v1/me --json
printf '%s\n' "$TOKEN" | classreach config init --token-stdin --force
classreach completion zsh > ~/.zfunc/_classreach
```

## Global Flags

| Flag | Description |
| --- | --- |
| `--config` | Config file path |
| `--base-url` | API base URL override |
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

Tag a semver release:

```bash
git tag v0.1.0
git push origin main
git push origin v0.1.0
```

The release workflow uses GoReleaser to publish archives and checksums.
Set `HOMEBREW_TAP_TOKEN` before the first tagged release so GoReleaser can
update `jwmoss/homebrew-tap`.
