package cmd

import (
	"fmt"

	"github.com/productivityenthusiast/keygen-cli/internal/output"
	"github.com/spf13/cobra"
)

var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Clear saved authentication for the active profile",
	Long: `Clear saved authentication for the active profile.

Use --profile to logout from a specific profile:
  keygen logout --profile prod`,
	Run: func(cmd *cobra.Command, args []string) {
		cfg := loadConfig()
		if err := cfg.Clear(); err != nil {
			output.Success(map[string]string{
				"message": "No saved config to clear",
				"profile": cfg.ProfileName,
			})
			return
		}
		output.Success(map[string]string{
			"message": fmt.Sprintf("Logged out — profile %q removed", cfg.ProfileName),
			"profile": cfg.ProfileName,
		})
	},
}

func init() {
	rootCmd.AddCommand(logoutCmd)
}
