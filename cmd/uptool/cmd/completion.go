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

const (
	shellBash       = "bash"
	shellZsh        = "zsh"
	shellFish       = "fish"
	shellPowershell = "powershell"
	osWindows       = "windows"
)

// validateFilePath validates that a file path is safe to read/write
func validateFilePath(path string) error {
	// Clean the path to resolve any . or .. components
	cleanPath := filepath.Clean(path)

	// Check for directory traversal attempts
	if strings.Contains(cleanPath, "..") {
		return fmt.Errorf("path contains directory traversal: %s", path)
	}

	return nil
}

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
	ValidArgs:             []string{shellBash, shellZsh, shellFish, shellPowershell},
	Args:                  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
	RunE: func(cmd *cobra.Command, args []string) error {
		switch args[0] {
		case shellBash:
			return rootCmd.GenBashCompletion(os.Stdout)
		case shellZsh:
			return rootCmd.GenZshCompletion(os.Stdout)
		case shellFish:
			return rootCmd.GenFishCompletion(os.Stdout, true)
		case shellPowershell:
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
	ValidArgs: []string{shellBash, shellZsh, shellFish, shellPowershell},
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
	err = os.MkdirAll(dir, 0o750)
	if err != nil {
		return fmt.Errorf("create directory %s: %w", dir, err)
	}

	// Generate completion to file
	// Validate path for security
	err = validateFilePath(installPath)
	if err != nil {
		return fmt.Errorf("invalid install path: %w", err)
	}

	file, err := os.Create(installPath) // #nosec G304 - path is validated above
	if err != nil {
		return fmt.Errorf("create file %s: %w\nTry running with sudo or install manually", installPath, err)
	}
	defer func() {
		if cerr := file.Close(); cerr != nil && err == nil {
			err = cerr
		}
	}()

	// Generate completion
	switch shell {
	case shellBash:
		err = rootCmd.GenBashCompletion(file)
	case shellZsh:
		err = rootCmd.GenZshCompletion(file)
	case shellFish:
		err = rootCmd.GenFishCompletion(file, true)
	case shellPowershell:
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
		case strings.Contains(shellName, shellBash):
			return shellBash, nil
		case strings.Contains(shellName, shellZsh):
			return shellZsh, nil
		case strings.Contains(shellName, shellFish):
			return shellFish, nil
		}
	}

	// On Windows, check for PowerShell
	if runtime.GOOS == osWindows {
		return shellPowershell, nil
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
	if runtime.GOOS == osWindows {
		return ""
	}

	// Validate PID to prevent command injection
	if pid <= 0 || pid > 2147483647 { // Valid PID range
		return ""
	}
	cmd := exec.Command("ps", "-p", fmt.Sprint(pid), "-o", "comm=") // #nosec G204 - PID is validated to be a positive integer
	output, err := cmd.Output()
	if err != nil {
		return ""
	}

	comm := strings.TrimSpace(string(output))
	if strings.Contains(comm, shellBash) {
		return shellBash
	}
	if strings.Contains(comm, shellZsh) {
		return shellZsh
	}
	if strings.Contains(comm, shellFish) {
		return shellFish
	}

	return ""
}

// getCompletionPath returns the appropriate installation path for each shell
func getCompletionPath(shell string) (path, instructions string, err error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", "", fmt.Errorf("get home directory: %w", err)
	}

	switch shell {
	case shellBash:
		// Try system location first, fall back to user location
		if runtime.GOOS == "darwin" {
			// macOS with Homebrew
			if brewPrefix := getBrewPrefix(); brewPrefix != "" {
				path = filepath.Join(brewPrefix, "etc", "bash_completion.d", "uptool")
				instructions = "Restart your terminal or run: source ~/.bash_profile"
				return path, instructions, err
			}
		}
		// Linux or macOS without brew - use user location
		path = filepath.Join(homeDir, ".local", "share", "bash-completion", "completions", "uptool")
		instructions = "Add to ~/.bashrc:\n  [ -f ~/.local/share/bash-completion/completions/uptool ] && source ~/.local/share/bash-completion/completions/uptool\nThen restart your terminal."

	case shellZsh:
		// User's zsh completion directory
		path = filepath.Join(homeDir, ".zsh", "completions", "_uptool")
		instructions = `Add to ~/.zshrc:
  fpath=(~/.zsh/completions $fpath)
  autoload -U compinit && compinit
Then restart your terminal.`

	case shellFish:
		// Fish user completions directory
		configDir := filepath.Join(homeDir, ".config", "fish", "completions")
		path = filepath.Join(configDir, "uptool.fish")
		instructions = "Restart your terminal for completions to take effect."

	case shellPowershell:
		// PowerShell user profile
		if runtime.GOOS == osWindows {
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
