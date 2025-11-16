package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
)

var completionCmd = &cobra.Command{
	Use:   "completion [bash|zsh|fish|powershell]",
	Short: "Generate shell completion scripts",
	Long: `Generate shell completion scripts for uptool.

The completion script can be generated for bash, zsh, fish, or powershell.
Scripts are output to stdout, allowing you to pipe them to the appropriate location.

Alternatively, use 'completion install' to automatically install to the correct location.`,
	Example: `  # Output bash completions to stdout
  uptool completion bash

  # Install for current shell automatically
  uptool completion install

  # Install for specific shell
  uptool completion install bash

  # Temporary load (current session only)
  source <(uptool completion bash)`,
	DisableFlagsInUseLine: true,
	ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
	Args:                  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
	RunE: func(cmd *cobra.Command, args []string) error {
		switch args[0] {
		case "bash":
			return rootCmd.GenBashCompletion(os.Stdout)
		case "zsh":
			return rootCmd.GenZshCompletion(os.Stdout)
		case "fish":
			return rootCmd.GenFishCompletion(os.Stdout, true)
		case "powershell":
			return rootCmd.GenPowerShellCompletionWithDesc(os.Stdout)
		}
		return nil
	},
}

var completionInstallCmd = &cobra.Command{
	Use:   "install [bash|zsh|fish|powershell]",
	Short: "Install completion script to the system",
	Long: `Install completion script to the appropriate system location.

If no shell is specified, automatically detects the current shell.
This command requires write permissions to system directories.`,
	Example: `  # Auto-detect and install
  uptool completion install

  # Install for specific shell
  uptool completion install bash`,
	ValidArgs: []string{"bash", "zsh", "fish", "powershell"},
	Args:      cobra.MaximumNArgs(1),
	RunE:      runCompletionInstall,
}

func init() {
	rootCmd.AddCommand(completionCmd)
	completionCmd.AddCommand(completionInstallCmd)
}

func runCompletionInstall(cmd *cobra.Command, args []string) error {
	// Detect shell if not provided
	shell := ""
	if len(args) > 0 {
		shell = args[0]
	} else {
		var err error
		shell, err = detectShell()
		if err != nil {
			return fmt.Errorf("could not detect shell: %w\nPlease specify the shell explicitly: uptool completion install [bash|zsh|fish|powershell]", err)
		}
		fmt.Printf("Detected shell: %s\n", shell)
	}

	// Get install path
	installPath, instructions, err := getCompletionPath(shell)
	if err != nil {
		return err
	}

	// Create parent directory if needed
	dir := filepath.Dir(installPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("create directory %s: %w", dir, err)
	}

	// Generate completion to file
	file, err := os.Create(installPath)
	if err != nil {
		return fmt.Errorf("create file %s: %w\nTry running with sudo or install manually", installPath, err)
	}
	defer file.Close()

	// Generate completion
	switch shell {
	case "bash":
		err = rootCmd.GenBashCompletion(file)
	case "zsh":
		err = rootCmd.GenZshCompletion(file)
	case "fish":
		err = rootCmd.GenFishCompletion(file, true)
	case "powershell":
		err = rootCmd.GenPowerShellCompletionWithDesc(file)
	default:
		return fmt.Errorf("unsupported shell: %s", shell)
	}

	if err != nil {
		return fmt.Errorf("generate completion: %w", err)
	}

	fmt.Printf("✓ Completion installed to: %s\n", installPath)
	if instructions != "" {
		fmt.Printf("\n%s\n", instructions)
	}

	return nil
}

// detectShell attempts to detect the current shell
func detectShell() (string, error) {
	// Check SHELL environment variable
	shellPath := os.Getenv("SHELL")
	if shellPath != "" {
		shellName := filepath.Base(shellPath)
		// Normalize shell names
		switch {
		case strings.Contains(shellName, "bash"):
			return "bash", nil
		case strings.Contains(shellName, "zsh"):
			return "zsh", nil
		case strings.Contains(shellName, "fish"):
			return "fish", nil
		}
	}

	// On Windows, check for PowerShell
	if runtime.GOOS == "windows" {
		return "powershell", nil
	}

	// Try to detect parent process
	if pid := os.Getppid(); pid != 0 {
		// This is a best-effort attempt
		shellName := detectParentShell(pid)
		if shellName != "" {
			return shellName, nil
		}
	}

	return "", fmt.Errorf("unable to detect shell from environment")
}

// detectParentShell tries to detect shell from parent process (best effort)
func detectParentShell(pid int) string {
	// This is a simplified detection, works on most Unix systems
	if runtime.GOOS == "windows" {
		return ""
	}

	cmd := exec.Command("ps", "-p", fmt.Sprint(pid), "-o", "comm=")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}

	comm := strings.TrimSpace(string(output))
	if strings.Contains(comm, "bash") {
		return "bash"
	}
	if strings.Contains(comm, "zsh") {
		return "zsh"
	}
	if strings.Contains(comm, "fish") {
		return "fish"
	}

	return ""
}

// getCompletionPath returns the appropriate installation path for each shell
func getCompletionPath(shell string) (path string, instructions string, err error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", "", fmt.Errorf("get home directory: %w", err)
	}

	switch shell {
	case "bash":
		// Try system location first, fall back to user location
		if runtime.GOOS == "darwin" {
			// macOS with Homebrew
			if brewPrefix := getBrewPrefix(); brewPrefix != "" {
				path = filepath.Join(brewPrefix, "etc", "bash_completion.d", "uptool")
				instructions = "Restart your terminal or run: source ~/.bash_profile"
				return
			}
		}
		// Linux or macOS without brew - use user location
		path = filepath.Join(homeDir, ".local", "share", "bash-completion", "completions", "uptool")
		instructions = "Add to ~/.bashrc:\n  [ -f ~/.local/share/bash-completion/completions/uptool ] && source ~/.local/share/bash-completion/completions/uptool\nThen restart your terminal."

	case "zsh":
		// User's zsh completion directory
		path = filepath.Join(homeDir, ".zsh", "completions", "_uptool")
		instructions = `Add to ~/.zshrc:
  fpath=(~/.zsh/completions $fpath)
  autoload -U compinit && compinit
Then restart your terminal.`

	case "fish":
		// Fish user completions directory
		configDir := filepath.Join(homeDir, ".config", "fish", "completions")
		path = filepath.Join(configDir, "uptool.fish")
		instructions = "Restart your terminal for completions to take effect."

	case "powershell":
		// PowerShell user profile
		if runtime.GOOS == "windows" {
			documentsDir := filepath.Join(homeDir, "Documents")
			path = filepath.Join(documentsDir, "PowerShell", "Scripts", "uptool-completion.ps1")
			instructions = `Add to your PowerShell profile:
  . "$HOME\Documents\PowerShell\Scripts\uptool-completion.ps1"
Then restart PowerShell.`
		} else {
			path = filepath.Join(homeDir, ".config", "powershell", "uptool-completion.ps1")
			instructions = `Add to your PowerShell profile:
  . ~/.config/powershell/uptool-completion.ps1
Then restart PowerShell.`
		}

	default:
		return "", "", fmt.Errorf("unsupported shell: %s", shell)
	}

	return path, instructions, nil
}

// getBrewPrefix returns Homebrew prefix if available
func getBrewPrefix() string {
	cmd := exec.Command("brew", "--prefix")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(output))
}

