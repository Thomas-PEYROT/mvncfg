# mvncfg

Gestionnaire de profils Maven minimaliste. Passe d'un `settings.xml` à l'autre via un symlink dans `~/.m2`.

## Installation

```bash
go install github.com/Thomas-PEYROT/mvncfg/cmd/mvncfg@latest
```

Assure-toi que `$HOME/go/bin` (ou `$GOPATH/bin`) est dans ton `PATH`.

## Utilisation

```bash
mvncfg list                  # liste les profils disponibles
mvncfg current               # affiche le profil actif
mvncfg use <profile>         # active un profil
mvncfg completion <bash|zsh> # affiche le script de completion
```

## Organisation des fichiers

```
~/.m2/
├── profiles/
│   ├── default.xml
│   ├── work.xml
│   └── personal.xml
└── settings.xml -> profiles/work.xml
```

## Completion shell

### Bash

Ajoute dans ton `~/.bashrc` :

```bash
source <(mvncfg completion bash)
```

### Zsh

Ajoute dans ton `~/.zshrc` :

```zsh
source <(mvncfg completion zsh)
```

## Développement

```bash
go test ./...
go build ./cmd/mvncfg
```
