# mvncfg

Minimalist Maven profile manager. Switches between `settings.xml` files using a symlink in `~/.m2`.

## Installation

### From a local clone (recommended)

The `install.sh` script builds the binary, copies it to `~/.local/bin`, and sets up shell completion:

```bash
cd /home/tpe/Bureau/Perso/mvncfg
./install.sh
```

By default, it detects your shell (`$SHELL`). To force a specific shell:

```bash
./install.sh zsh
# or
./install.sh bash
```

Alternatively, you can build manually:

```bash
cd /home/tpe/Bureau/Perso/mvncfg
go build -o ~/.local/bin/mvncfg ./cmd/mvncfg
```

Or use `go install`:

```bash
cd /home/tpe/Bureau/Perso/mvncfg
go install ./cmd/mvncfg
```

The binary will be placed in `$HOME/go/bin` (or `$GOPATH/bin`).

### From a GitHub release

Precompiled binaries are published for Linux and macOS (amd64 and arm64). Windows is not supported yet.

Example for Linux amd64:

```bash
curl -sL -o ~/.local/bin/mvncfg https://github.com/Thomas-PEYROT/mvncfg/releases/latest/download/mvncfg-linux-amd64
chmod +x ~/.local/bin/mvncfg
mvncfg install-completion
```

### Via `go install`

```bash
go install github.com/Thomas-PEYROT/mvncfg/cmd/mvncfg@latest
mvncfg install-completion
```

## Shell completion setup

After installing the binary:

```bash
mvncfg install-completion
```

This detects your shell (`bash` or `zsh`) and configures completion automatically. Then reload your shell:

```bash
source ~/.bashrc   # or ~/.zshrc
```

## Usage

```bash
mvncfg list                  # list available profiles
mvncfg current               # show the active profile
mvncfg use <profile>         # activate a profile
mvncfg install-completion    # install completion for the current shell
```

## File layout

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
