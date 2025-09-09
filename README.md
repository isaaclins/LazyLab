# lazylab

Go CLI for running and managing malware analysis Docker containers with safe defaults, profiles, and conveniences.

## Install

Build from source:

```bash
./build.sh
./bin/lazylab --version
```

## Quick start

```bash
# Run with read-only FS, writable /work, no network, 1 CPU, 1G RAM
./bin/lazylab --no-net --read-only --writable /work --cpus 1 --memory 1g

# Copy a file to container home and start fish (auto-installs when online)
./bin/lazylab -c ./example/main.py --shell fish

# Save and reuse profiles
./bin/lazylab --no-net --read-only -m "$PWD":/work profile save offline
./bin/lazylab profile run offline --cpus 2
```

## Flags (highlights)

- `-c, --copy`: copy host files/dirs to container `$HOME` (made read-only)
- `-m, --mount`: bind mount host paths (default dest `$HOME/<basename>`)
- `-p, --packages`: install Homebrew packages (idempotent, cached)
- `--cache-packages` (default true): persist Homebrew cache between runs
- `--purge-cache`: remove the lazylab cache volume on exit
- `--read-only` + `--writable <path>`: read-only root with tmpfs exceptions (mkdir ensured)
- `--no-net`: no network (skips installs and fish auto-install)
- `--image`: container image (default `homebrew/brew:latest`)
- `--shell`: preferred shell (default `fish`; fallbacks bash → zsh → sh)
- `--user`: run as specific user (e.g., `1000:1000`)
- `--cap-drop-all`, `--no-new-privs`: security hardening
- `--graceful`: graceful stop via `docker stop --timeout 10` (default is fast kill)
- `--prefix`, `-n`: name generation and custom name (validated and uniqueness-checked)
- `--profile <name>`: load a saved profile (`~/.lazylab/profiles/*.yaml|*.json`)
- `--verbose`: print docker commands
- `--version`: print version/build info

## Profiles

Profiles are stored under `~/.lazylab/profiles/` as YAML or JSON. They capture all flags. Use:

```bash
./bin/lazylab profile save offline
./bin/lazylab profile list
./bin/lazylab profile run offline --cpus 2
./bin/lazylab profile edit offline
```

## Testing

```bash
# Unit and integration tests
./test.sh

# With verbose
VERBOSE=1 ./test.sh

# Run specific package
PKG=./profiles ./test.sh
```

CI runs the same `go test` tasks. End-to-end interactive checks are opt-in locally.
