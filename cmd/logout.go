package cmd

import (
	"github.com/productivityenthusiast/keygen-cli/internal/config"
	"github.com/productivityenthusiast/keygen-cli/internal/output"
	"github.com/spf13/cobra"
)

var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Clear saved authentication",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.Load()
		if err := cfg.Clear(); err != nil {
			output.Success(map[string]string{
				"message": "No saved config to clear",
			})
			return
		}
		output.Success(map[string]string{
			"message": "Logged out — saved token removed",
		})
	},
}

func init() {
	rootCmd.AddCommand(logoutCmd)
}
