package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/emirhannsarial/tifybe-cli/pkg/config"
)

func init() {
	rootCmd.AddCommand(logoutCmd)
}

var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Clear your local Tifybe credentials",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := config.ClearConfig(); err != nil {
			return fmt.Errorf("failed to clear credentials: %w", err)
		}
		fmt.Println("✅ Logged out successfully.")
		return nil
	},
}
