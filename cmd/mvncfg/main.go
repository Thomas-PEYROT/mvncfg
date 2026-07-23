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

func allCommands() []commandInfo {
	return publicCommands
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
	case "list":
		cfg, err := config.New()
		if err != nil {
			return err
		}
		return cmdList(cfg)
	case "current":
		cfg, err := config.New()
		if err != nil {
			return err
		}
		return cmdCurrent(cfg)
	case "use":
		if len(args) < 2 {
			return fmt.Errorf("usage: mvncfg use <profile>")
		}
		cfg, err := config.New()
		if err != nil {
			return err
		}
		return cmdUse(cfg, args[1])
	case "completion":
		if len(args) < 2 {
			return fmt.Errorf("usage: mvncfg completion <bash|zsh>")
		}
		return cmdCompletion(args[1])
	case "install-completion":
		return completion.Install()
	case "init":
		cfg, err := config.New()
		if err != nil {
			return err
		}
		return cmdInit(cfg)
	case "create":
		if len(args) < 2 {
			return fmt.Errorf("usage: mvncfg create <profile>")
		}
		cfg, err := config.New()
		if err != nil {
			return err
		}
		return cmdCreate(cfg, args[1])
	case "delete":
		if len(args) < 2 {
			return fmt.Errorf("usage: mvncfg delete <profile>")
		}
		cfg, err := config.New()
		if err != nil {
			return err
		}
		return cmdDelete(cfg, args[1])
	case "rename":
		if len(args) < 3 {
			return fmt.Errorf("usage: mvncfg rename <old> <new>")
		}
		cfg, err := config.New()
		if err != nil {
			return err
		}
		return cmdRename(cfg, args[1], args[2])
	case "version", "--version", "-v":
		fmt.Println(getVersion())
		return nil
	case "help", "--help", "-h":
		printHelp(args[1:])
		return nil
	default:
		return fmt.Errorf("unknown command: %s\n\n%s", args[0], usageText())
	}
}

func cmdList(cfg *config.M2Config) error {
	profiles, err := profile.List(cfg)
	if err != nil {
		return err
	}
	for _, p := range profiles {
		fmt.Println(p)
	}
	return nil
}

func cmdCurrent(cfg *config.M2Config) error {
	current, err := profile.Current(cfg)
	if err != nil {
		return err
	}
	fmt.Println(current)
	return nil
}

func cmdUse(cfg *config.M2Config, name string) error {
	if err := profile.Use(cfg, name); err != nil {
		return err
	}
	fmt.Printf("Switched to %s\n", name)
	return nil
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

func cmdInit(cfg *config.M2Config) error {
	if err := profile.Init(cfg); err != nil {
		return err
	}
	fmt.Println("Initialized mvncfg with a default profile")
	return nil
}

func cmdCreate(cfg *config.M2Config, name string) error {
	if err := profile.Create(cfg, name); err != nil {
		return err
	}
	fmt.Printf("Created profile %s\n", name)
	return nil
}

func cmdDelete(cfg *config.M2Config, name string) error {
	if err := profile.Delete(cfg, name); err != nil {
		return err
	}
	fmt.Printf("Deleted profile %s\n", name)
	return nil
}

func cmdRename(cfg *config.M2Config, oldName, newName string) error {
	if err := profile.Rename(cfg, oldName, newName); err != nil {
		return err
	}
	fmt.Printf("Renamed profile %s to %s\n", oldName, newName)
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
	for _, cmd := range allCommands() {
		if cmd.name == name {
			fmt.Printf("%s\n\n", cmd.usage)
			fmt.Printf("Description:\n  %s\n\n", cmd.description)
			fmt.Printf("Example:\n  %s\n", cmd.example)
			return
		}
	}

	fmt.Printf("Unknown command: %s\n\n%s", name, usageText())
}
