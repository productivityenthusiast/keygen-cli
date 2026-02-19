package cmd

import (
	"github.com/productivityenthusiast/keygen-cli/internal/api"
	"github.com/productivityenthusiast/keygen-cli/internal/output"
	"github.com/spf13/cobra"
)

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate with Keygen",
	Long: `Authenticate using a token or email/password credentials.

Use --profile to login to a specific profile:
  keygen login token --token mytoken --profile prod
  keygen login password --email admin@example.com --password secret --profile testing`,
}

var loginTokenCmd = &cobra.Command{
	Use:   "token",
	Short: "Login with an API token",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := loadConfig()

		loginToken, _ := cmd.Flags().GetString("token")
		if loginToken != "" {
			cfg.Token = loginToken
		}

		if cfg.Token == "" {
			exitError("token is required (use --token or set KEYGEN_TOKEN)")
		}
		if cfg.AccountID == "" {
			exitError("account ID is required (set KEYGEN_ACCOUNT_ID)")
		}
		if cfg.BaseURL == "" {
			exitError("base URL is required (set KEYGEN_BASE_URL)")
		}

		client := api.NewClient(cfg.BaseURL, cfg.AccountID, cfg.Token)
		_, err := client.ValidateToken()
		if err != nil {
			output.Error("token validation failed: " + err.Error())
			return
		}

		if err := cfg.Save(); err != nil {
			output.Error("failed to save config: " + err.Error())
			return
		}

		output.Success(map[string]interface{}{
			"message":    "Login successful",
			"profile":    cfg.ProfileName,
			"account_id": cfg.AccountID,
			"base_url":   cfg.BaseURL,
		})
	},
}

var loginPasswordCmd = &cobra.Command{
	Use:   "password",
	Short: "Login with email and password",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := loadConfig()

		email, _ := cmd.Flags().GetString("email")
		password, _ := cmd.Flags().GetString("password")

		if email != "" {
			cfg.Email = email
		}
		if password != "" {
			cfg.Password = password
		}

		if cfg.Email == "" || cfg.Password == "" {
			exitError("email and password are required")
		}
		if cfg.AccountID == "" {
			exitError("account ID is required")
		}
		if cfg.BaseURL == "" {
			exitError("base URL is required")
		}

		client := api.NewClient(cfg.BaseURL, cfg.AccountID, "")
		token, err := client.CreateToken(cfg.Email, cfg.Password)
		if err != nil {
			output.Error("login failed: " + err.Error())
			return
		}

		cfg.Token = token.Token
		cfg.TokenExp = token.Expiry
		if err := cfg.Save(); err != nil {
			output.Error("failed to save config: " + err.Error())
			return
		}

		output.Success(map[string]interface{}{
			"message":    "Login successful",
			"profile":    cfg.ProfileName,
			"token_id":   token.ID,
			"expiry":     token.Expiry,
			"account_id": cfg.AccountID,
		})
	},
}

func init() {
	loginTokenCmd.Flags().String("token", "", "API token")
	loginPasswordCmd.Flags().String("email", "", "Account email")
	loginPasswordCmd.Flags().String("password", "", "Account password")

	loginCmd.AddCommand(loginTokenCmd)
	loginCmd.AddCommand(loginPasswordCmd)
	rootCmd.AddCommand(loginCmd)
}
