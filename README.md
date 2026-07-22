# mvncfg

Gestionnaire de profils Maven minimaliste. Passe d'un `settings.xml` à l'autre via un symlink dans `~/.m2`.

## Installation

### Depuis un clone local (recommandé)

Le script `install.sh` compile le binaire, le copie dans `~/.local/bin` et configure la completion pour ton shell :

```bash
cd /home/tpe/Bureau/Perso/mvncfg
./install.sh
```

Par défaut, le script détecte ton shell (`$SHELL`). Pour forcer un shell particulier :

```bash
./install.sh zsh
# ou
./install.sh bash
```

Sinon, tu peux aussi compiler manuellement :

```bash
cd /home/tpe/Bureau/Perso/mvncfg
go build -o ~/.local/bin/mvncfg ./cmd/mvncfg
```

Ou via `go install` :

```bash
cd /home/tpe/Bureau/Perso/mvncfg
go install ./cmd/mvncfg
```

Le binaire sera alors dans `$HOME/go/bin` (ou `$GOPATH/bin`).

### Une fois le repo public

```bash
go install github.com/Thomas-PEYROT/mvncfg/cmd/mvncfg@latest
```

## Configuration de l'autocomplétion

Après avoir installé le binaire :

```bash
mvncfg install-completion
```

Cette commande détecte ton shell (`bash` ou `zsh`) et configure la completion automatiquement. Il te suffit ensuite de recharger ton shell :

```bash
source ~/.bashrc   # ou source ~/.zshrc
```

## Utilisation

```bash
mvncfg list                  # liste les profils disponibles
mvncfg current               # affiche le profil actif
mvncfg use <profile>         # active un profil
mvncfg install-completion    # installe la completion pour le shell courant
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

## Développement

```bash
go test ./...
go build ./cmd/mvncfg
```
