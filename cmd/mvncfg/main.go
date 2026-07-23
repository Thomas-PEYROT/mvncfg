package main

import (
	"fmt"
	"os"
	"runtime/debug"
	"strings"

	"github.com/Thomas-PEYROT/mvncfg/internal/completion"
	"github.com/Thomas-PEYROT/mvncfg/internal/config"
	"github.com/Thomas-PEYROT/mvncfg/internal/profile"
)

type commandInfo struct {
	name        string
	usage       string
	description string
	example     string
}

// version is set at build time via -ldflags. If not set, go module version is used as a fallback.
var version = "dev"

func getVersion() string {
	if version != "dev" {
		return version
	}
	if info, ok := debug.ReadBuildInfo(); ok {
		if info.Main.Version != "" && info.Main.Version != "(devel)" {
			return info.Main.Version
		}
	}
	return version
}

var publicCommands = []commandInfo{
	{
		name:        "init",
		usage:       "mvncfg init",
		description: "Initialize the ~/.m2/profiles directory and create a default profile.",
		example:     "mvncfg init",
	},
	{
		name:        "list",
		usage:       "mvncfg list",
		description: "List all available Maven profiles in ~/.m2/profiles.",
		example:     "mvncfg list",
	},
	{
		name:        "current",
		usage:       "mvncfg current",
		description: "Show the profile currently active via the ~/.m2/settings.xml symlink.",
		example:     "mvncfg current",
	},
	{
		name:        "use",
		usage:       "mvncfg use <profile>",
		description: "Activate a Maven profile by symlinking ~/.m2/settings.xml to it.",
		example:     "mvncfg use work",
	},
	{
		name:        "create",
		usage:       "mvncfg create <profile>",
		description: "Create a new Maven profile from a default settings.xml template.",
		example:     "mvncfg create work",
	},
	{
		name:        "delete",
		usage:       "mvncfg delete <profile>",
		description: "Delete a Maven profile. The active profile cannot be deleted.",
		example:     "mvncfg delete work",
	},
	{
		name:        "rename",
		usage:       "mvncfg rename <old> <new>",
		description: "Rename a Maven profile. If it is the active profile, the symlink is updated.",
		example:     "mvncfg rename work personal",
	},
	{
		name:        "install-completion",
		usage:       "mvncfg install-completion",
		description: "Install shell completion for bash or zsh.",
		example:     "mvncfg install-completion",
	},
	{
		name:        "help",
		usage:       "mvncfg help [command]",
		description: "Show this help message or detailed help for a command.",
		example:     "mvncfg help use",
	},
	{
		name:        "version",
		usage:       "mvncfg version",
		description: "Show the version of mvncfg.",
		example:     "mvncfg version",
	},
}

// commandHandler executes a command given the resolved config and remaining arguments.
type commandHandler func(cfg *config.M2Config, args []string) error

var commands = map[string]commandHandler{
	"init":               cmdInit,
	"list":               cmdList,
	"current":            cmdCurrent,
	"use":                cmdUse,
	"create":             cmdCreate,
	"delete":             cmdDelete,
	"rename":             cmdRename,
	"install-completion": cmdInstallCompletion,
}

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) == 0 {
		printUsage()
		return nil
	}

	switch args[0] {
	case "version", "--version", "-v":
		fmt.Println(getVersion())
		return nil
	case "help", "--help", "-h":
		printHelp(args[1:])
		return nil
	case "completion":
		if len(args) < 2 {
			return fmt.Errorf("usage: mvncfg completion <bash|zsh>")
		}
		return cmdCompletion(args[1])
	}

	handler, ok := commands[args[0]]
	if !ok {
		return fmt.Errorf("unknown command: %s\n\n%s", args[0], usageText())
	}

	cfg, err := config.New()
	if err != nil {
		return err
	}

	return handler(cfg, args[1:])
}

func cmdList(cfg *config.M2Config, _ []string) error {
	profiles, err := profile.List(cfg)
	if err != nil {
		return err
	}
	for _, p := range profiles {
		fmt.Println(p)
	}
	return nil
}

func cmdCurrent(cfg *config.M2Config, _ []string) error {
	current, err := profile.Current(cfg)
	if err != nil {
		return err
	}
	fmt.Println(current)
	return nil
}

func cmdUse(cfg *config.M2Config, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: mvncfg use <profile>")
	}
	if err := profile.Use(cfg, args[0]); err != nil {
		return err
	}
	fmt.Printf("Switched to %s\n", args[0])
	return nil
}

func cmdCreate(cfg *config.M2Config, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: mvncfg create <profile>")
	}
	if err := profile.Create(cfg, args[0]); err != nil {
		return err
	}
	fmt.Printf("Created profile %s\n", args[0])
	return nil
}

func cmdDelete(cfg *config.M2Config, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: mvncfg delete <profile>")
	}
	if err := profile.Delete(cfg, args[0]); err != nil {
		return err
	}
	fmt.Printf("Deleted profile %s\n", args[0])
	return nil
}

func cmdRename(cfg *config.M2Config, args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: mvncfg rename <old> <new>")
	}
	if err := profile.Rename(cfg, args[0], args[1]); err != nil {
		return err
	}
	fmt.Printf("Renamed profile %s to %s\n", args[0], args[1])
	return nil
}

func cmdInit(cfg *config.M2Config, _ []string) error {
	if err := profile.Init(cfg); err != nil {
		return err
	}
	fmt.Println("Initialized mvncfg with a default profile")
	return nil
}

func cmdInstallCompletion(_ *config.M2Config, _ []string) error {
	return completion.Install()
}

func cmdCompletion(shell string) error {
	switch shell {
	case "bash":
		fmt.Print(completion.Bash())
	case "zsh":
		fmt.Print(completion.Zsh())
	default:
		return fmt.Errorf("unsupported shell: %s (supported: bash, zsh)", shell)
	}
	return nil
}

func printUsage() {
	fmt.Print(usageText())
}

func usageText() string {
	var b strings.Builder
	b.WriteString("mvncfg — switch between Maven settings.xml profiles\n\n")
	b.WriteString("Usage:\n  mvncfg <command> [args]\n\n")
	b.WriteString("Commands:\n")
	for _, cmd := range publicCommands {
		b.WriteString(fmt.Sprintf("  %-20s %s\n", cmd.name, cmd.description))
	}
	b.WriteString("\nRun 'mvncfg help <command>' for more information on a command.\n")
	return b.String()
}

func printHelp(args []string) {
	if len(args) == 0 {
		fmt.Print(usageText())
		return
	}

	name := args[0]
	for _, cmd := range publicCommands {
		if cmd.name == name {
			fmt.Printf("%s\n\n", cmd.usage)
			fmt.Printf("Description:\n  %s\n\n", cmd.description)
			fmt.Printf("Example:\n  %s\n", cmd.example)
			return
		}
	}

	fmt.Printf("Unknown command: %s\n\n%s", name, usageText())
}
