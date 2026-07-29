package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/emirhannsarial/tifybe-cli/pkg/config"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

func init() {
	rootCmd.AddCommand(loginCmd)
}

// readAPIKey takes the key without echoing it.
//
// The credentials file is written 0600, but typing the key in the clear
// leaves it in terminal scrollback and in whatever records that session —
// screen shares and asciinema recordings included. Piped or redirected input
// has no terminal to hide, so that path falls back to a plain read.
func readAPIKey() (string, error) {
	fd := int(os.Stdin.Fd())
	if !term.IsTerminal(fd) {
		key, err := bufio.NewReader(os.Stdin).ReadString('\n')
		if err != nil {
			return "", err
		}
		return key, nil
	}

	raw, err := term.ReadPassword(fd)
	// ReadPassword swallows the newline the user typed, so the next line of
	// output would otherwise run onto the prompt.
	fmt.Println()
	if err != nil {
		return "", err
	}
	return string(raw), nil
}

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate the CLI with your Tifybe API key",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("To get your API key, visit: https://tifybe.com/dashboard/settings")
		fmt.Print("Enter your Tifybe API key (tfy_...): ")

		apiKey, err := readAPIKey()
		if err != nil {
			return fmt.Errorf("failed to read input: %w", err)
		}

		apiKey = strings.TrimSpace(apiKey)
		if !strings.HasPrefix(apiKey, "tfy_") || len(apiKey) < 16 {
			return fmt.Errorf("invalid API key format. Must start with 'tfy_'")
		}

		cfg := &config.Config{
			APIKey: apiKey,
		}

		if err := config.SaveConfig(cfg); err != nil {
			return fmt.Errorf("failed to save credentials: %w", err)
		}

		fmt.Println("\nAuthenticated. Credentials saved to ~/.tifybe/credentials.json (0600).")
		fmt.Println("You can now claim a persistent URL: tifybe listen 8080 --subdomain=my-startup")
		return nil
	},
}
