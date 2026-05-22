package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tifybe/tifybe-cli/pkg/config"
)

func init() {
	rootCmd.AddCommand(loginCmd)
}

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate the CLI with your Tifybe API key",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("To get your API key, visit: https://tifybe.com/dashboard/settings")
		fmt.Print("Enter your Tifybe API key (tfy_...): ")

		reader := bufio.NewReader(os.Stdin)
		apiKey, err := reader.ReadString('\n')
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

		fmt.Println("\n✅ Successfully authenticated!")
		fmt.Println("You can now use premium features like `tifybe listen <port> --subdomain=my-startup`")
		return nil
	},
}
