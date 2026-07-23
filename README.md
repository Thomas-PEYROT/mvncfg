# mvncfg

[![CI](https://github.com/Thomas-PEYROT/mvncfg/actions/workflows/ci.yml/badge.svg)](https://github.com/Thomas-PEYROT/mvncfg/actions/workflows/ci.yml)
[![Tests](https://github.com/Thomas-PEYROT/mvncfg/actions/workflows/tests.yml/badge.svg)](https://github.com/Thomas-PEYROT/mvncfg/actions/workflows/tests.yml)

A tiny CLI to switch between Maven `settings.xml` profiles using symlinks.

## Features

- `init` — initialize the `~/.m2/profiles` directory and a default profile.
- `list` — list available Maven profiles.
- `current` — show the active profile.
- `use <profile>` — switch to another profile.
- `create <profile>` — create a new profile from a default `settings.xml` template.
- `delete <profile>` — delete a profile (cannot delete the active one).
- `rename <old> <new>` — rename a profile (updates the symlink if active).
- `install-completion` — set up shell completion for bash or zsh.
- `version` / `--version` — show the installed version.

## Requirements

- Go 1.26+ (only for `go install` or building from source).
- Linux or macOS. Windows is not supported yet because `mvncfg` relies on symlinks.

## Installation

### Quick install

Clone the repository and run the install script:

```bash
git clone https://github.com/Thomas-PEYROT/mvncfg.git && cd mvncfg && ./install.sh
```

### Using `go install`

```bash
go install github.com/Thomas-PEYROT/mvncfg/cmd/mvncfg@latest
```

Make sure `$HOME/go/bin` (or `$GOPATH/bin`) is in your `PATH`.

### From a GitHub release

Download the binary for your platform from the [releases page](https://github.com/Thomas-PEYROT/mvncfg/releases), make it executable, and place it in a directory in your `PATH`.

Example for Linux amd64:

```bash
curl -sL -o ~/.local/bin/mvncfg https://github.com/Thomas-PEYROT/mvncfg/releases/latest/download/mvncfg-linux-amd64
chmod +x ~/.local/bin/mvncfg
```

### From source

```bash
git clone https://github.com/Thomas-PEYROT/mvncfg.git
cd mvncfg
go build -o ~/.local/bin/mvncfg ./cmd/mvncfg
```

Or use the provided install script from the clone:

```bash
git clone https://github.com/Thomas-PEYROT/mvncfg.git
cd mvncfg
./install.sh
```

By default, `./install.sh` installs to `~/.local/bin` and configures completion for your current shell. You can force a specific shell:

```bash
./install.sh zsh
./install.sh bash
```

## Shell completion

After installing the binary:

```bash
mvncfg install-completion
```

Then reload your shell:

```bash
source ~/.bashrc   # or ~/.zshrc
```

## Usage

```bash
mvncfg init                  # initialize profiles directory and default profile
mvncfg list                  # list available profiles
mvncfg current               # show the active profile
mvncfg use <profile>         # activate a profile
mvncfg create <profile>      # create a new profile from a default template
mvncfg delete <profile>      # delete a profile (not the active one)
mvncfg rename <old> <new>    # rename a profile
mvncfg install-completion    # install shell completion
mvncfg help [command]        # show help
mvncfg version               # show version
mvncfg --version             # show version
```

## File layout

`mvncfg` expects Maven profiles to live in `~/.m2/profiles/` and manages the active configuration via a symlink at `~/.m2/settings.xml`.

```
~/.m2/
├── profiles/
│   ├── default.xml
│   ├── work.xml
│   └── personal.xml
└── settings.xml -> profiles/work.xml
```

## Development

```bash
go test ./...
go build ./cmd/mvncfg
```

## Platform support

| Platform | Architecture | Status |
|----------|--------------|--------|
| Linux    | amd64, arm64 | ✅ supported |
| macOS    | amd64, arm64 | ✅ supported |
| Windows  | —            | ❌ not supported |
