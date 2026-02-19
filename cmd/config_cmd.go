package cmd

import (
	"fmt"
	"strings"

	"github.com/productivityenthusiast/keygen-cli/internal/config"
	"github.com/productivityenthusiast/keygen-cli/internal/output"
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage CLI configuration",
}

var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show current configuration",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.Load()

		maskedToken := ""
		if cfg.Token != "" {
			if len(cfg.Token) > 8 {
				maskedToken = cfg.Token[:4] + strings.Repeat("*", len(cfg.Token)-8) + cfg.Token[len(cfg.Token)-4:]
			} else {
				maskedToken = "****"
			}
		}

		output.Success(map[string]interface{}{
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
	Short: "Clear saved configuration",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.Load()
		if err := cfg.Clear(); err != nil {
			fmt.Printf("Note: %v\n", err)
		}
		output.Success(map[string]string{
			"message": "Configuration cleared",
		})
	},
}

func init() {
	configCmd.AddCommand(configShowCmd)
	configCmd.AddCommand(configClearCmd)
	rootCmd.AddCommand(configCmd)
}
