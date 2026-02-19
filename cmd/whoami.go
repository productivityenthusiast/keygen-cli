package cmd

import (
	"fmt"

	"github.com/productivityenthusiast/keygen-cli/internal/api"
	"github.com/productivityenthusiast/keygen-cli/internal/output"
	"github.com/spf13/cobra"
)

var whoamiCmd = &cobra.Command{
	Use:   "whoami",
	Short: "Show current auth context and active profile",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := loadConfig()

		if cfg.Token == "" {
			output.Error(fmt.Sprintf("not logged in on profile %q (no token configured)", cfg.ProfileName))
			return
		}

		client := api.NewClient(cfg.BaseURL, cfg.AccountID, cfg.Token)
		_, err := client.ValidateToken()

		output.Success(map[string]interface{}{
			"profile":     cfg.ProfileName,
			"account_id":  cfg.AccountID,
			"base_url":    cfg.BaseURL,
			"auth_method": "token",
			"token_valid": err == nil,
		})
	},
}

func init() {
	rootCmd.AddCommand(whoamiCmd)
}
