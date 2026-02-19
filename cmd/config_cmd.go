package cmd

import (
	"fmt"
	"strings"

	"github.com/productivityenthusiast/keygen-cli/internal/output"
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage CLI configuration",
}

var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show current configuration for the active profile",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := loadConfig()

		maskedToken := ""
		if cfg.Token != "" {
			if len(cfg.Token) > 8 {
				maskedToken = cfg.Token[:4] + strings.Repeat("*", len(cfg.Token)-8) + cfg.Token[len(cfg.Token)-4:]
			} else {
				maskedToken = "****"
			}
		}

		output.Success(map[string]interface{}{
			"profile":      cfg.ProfileName,
			"account_id":   cfg.AccountID,
			"base_url":     cfg.BaseURL,
			"token":        maskedToken,
			"email":        cfg.Email,
			"has_password": cfg.Password != "",
			"token_expiry": cfg.TokenExp,
		})
	},
}

var configClearCmd = &cobra.Command{
	Use:   "clear",
	Short: "Clear saved configuration for the active profile",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := loadConfig()
		if err := cfg.Clear(); err != nil {
			fmt.Printf("Note: %v\n", err)
		}
		output.Success(map[string]string{
			"message": fmt.Sprintf("Configuration for profile %q cleared", cfg.ProfileName),
			"profile": cfg.ProfileName,
		})
	},
}

func init() {
	configCmd.AddCommand(configShowCmd)
	configCmd.AddCommand(configClearCmd)
	rootCmd.AddCommand(configCmd)
}
