package cmd

import (
	"github.com/productivityenthusiast/keygen-cli/internal/api"
	"github.com/productivityenthusiast/keygen-cli/internal/output"
	"github.com/spf13/cobra"
)

var whoamiCmd = &cobra.Command{
	Use:   "whoami",
	Short: "Show current auth context",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := loadConfig()

		if cfg.Token == "" {
			output.Error("not logged in (no token configured)")
			return
		}

		client := api.NewClient(cfg.BaseURL, cfg.AccountID, cfg.Token)
		_, err := client.ValidateToken()

		output.Success(map[string]interface{}{
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
