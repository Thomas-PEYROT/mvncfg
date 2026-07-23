package completion

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Install detects the current shell and installs the appropriate completion.
func Install() error {
	shell := detectShell()
	switch shell {
	case "zsh":
		return installZsh()
	case "bash":
		return installBash()
	case "":
		return fmt.Errorf("cannot detect shell; please set $SHELL or run 'mvncfg completion <bash|zsh>' manually")
	default:
		return fmt.Errorf("unsupported shell: %s (supported: bash, zsh)", shell)
	}
}

func detectShell() string {
	s := os.Getenv("SHELL")
	if s == "" {
		return ""
	}
	return filepath.Base(s)
}

func installZsh() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("cannot determine home directory: %w", err)
	}

	completionDir := filepath.Join(home, ".config", "zsh", "completions")
	if err := os.MkdirAll(completionDir, 0o755); err != nil {
		return fmt.Errorf("cannot create completion directory: %w", err)
	}

	scriptPath := filepath.Join(completionDir, "_mvncfg")
	if err := os.WriteFile(scriptPath, []byte(Zsh()), 0o644); err != nil {
		return fmt.Errorf("cannot write completion script: %w", err)
	}

	rcPath := filepath.Join(home, ".zshrc")
	lines := []string{
		"# mvncfg completion",
		"fpath+=" + completionDir,
		"autoload -Uz compinit && compinit",
	}
	added, err := appendMissingLines(rcPath, lines)
	if err != nil {
		return fmt.Errorf("cannot update %s: %w", rcPath, err)
	}

	fmt.Printf("Installed zsh completion to %s\n", scriptPath)
	if len(added) > 0 {
		fmt.Printf("Added to %s:\n%s\n", rcPath, strings.Join(added, "\n"))
	} else {
		fmt.Printf("No changes needed in %s\n", rcPath)
	}
	fmt.Println("Reload your shell or run: source ~/.zshrc")
	return nil
}

func installBash() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("cannot determine home directory: %w", err)
	}

	rcPath := filepath.Join(home, ".bashrc")
	lines := []string{
		"# mvncfg completion",
		"source <(mvncfg completion bash)",
	}
	added, err := appendMissingLines(rcPath, lines)
	if err != nil {
		return fmt.Errorf("cannot update %s: %w", rcPath, err)
	}

	if len(added) > 0 {
		fmt.Printf("Added to %s:\n%s\n", rcPath, strings.Join(added, "\n"))
	} else {
		fmt.Printf("No changes needed in %s\n", rcPath)
	}
	fmt.Println("Reload your shell or run: source ~/.bashrc")
	return nil
}

// appendMissingLines appends only the lines that are not already present in the file.
// A line is considered present only if it matches an entire existing line (ignoring leading/trailing whitespace).
// If the file already exists and is not empty, a blank separator line is inserted before the new lines.
func appendMissingLines(path string, lines []string) ([]string, error) {
	content, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	existing := string(content)

	existingLines := make(map[string]struct{})
	for _, line := range strings.Split(existing, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			existingLines[trimmed] = struct{}{}
		}
	}

	var missing []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		if _, ok := existingLines[trimmed]; !ok {
			missing = append(missing, line)
		}
	}
	if len(missing) == 0 {
		return nil, nil
	}

	flag := os.O_APPEND | os.O_CREATE | os.O_WRONLY
	f, err := os.OpenFile(path, flag, 0o644)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	if len(existing) > 0 {
		if !strings.HasSuffix(existing, "\n") {
			if _, err := f.WriteString("\n"); err != nil {
				return nil, err
			}
		}
		if _, err := f.WriteString("\n"); err != nil {
			return nil, err
		}
	}

	if _, err := f.WriteString(strings.Join(missing, "\n") + "\n"); err != nil {
		return nil, err
	}

	return missing, nil
}
