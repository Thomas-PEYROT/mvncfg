package main

import (
	"fmt"
	"os"

	"github.com/Thomas-PEYROT/mvncfg/internal/completion"
	"github.com/Thomas-PEYROT/mvncfg/internal/config"
	"github.com/Thomas-PEYROT/mvncfg/internal/profile"
)

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
	case "help", "--help", "-h":
		printUsage()
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

func printUsage() {
	fmt.Print(usageText())
}

func usageText() string {
	return `Commands:
  mvncfg list
  mvncfg current
  mvncfg use <profile>
  mvncfg install-completion
`
}
