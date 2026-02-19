package cmd

import (
	"fmt"
	"strings"

	"github.com/productivityenthusiast/keygen-cli/internal/config"
	"github.com/productivityenthusiast/keygen-cli/internal/output"
	"github.com/spf13/cobra"
)

var profileCmd = &cobra.Command{
	Use:   "profile",
	Short: "Manage named profiles (like AWS CLI profiles)",
	Long: `Manage named profiles for different Keygen environments.

Each profile stores its own server URL, account ID, and credentials.
Use --profile <name> on any command to target a specific environment.

Examples:
  keygen profile add prod --account-id ABC --base-url https://lm.example.com
  keygen profile add testing --account-id XYZ --base-url https://lm-test.example.com
  keygen profile list
  keygen profile use prod
  keygen login token --token mytoken --profile prod
  keygen licenses list --profile testing`,
}

var profileListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all profiles",
	Run: func(cmd *cobra.Command, args []string) {
		names, defaultName := config.ListProfiles()

		if len(names) == 0 {
			output.Success(map[string]interface{}{
				"profiles": []string{},
				"message":  "No profiles configured. Run 'keygen profile add <name>' to create one.",
			})
			return
		}

		f := getFormat()
		if f == "table" || f == "csv" {
			headers := []string{"PROFILE", "DEFAULT", "ACCOUNT_ID", "BASE_URL"}
			rows := make([][]string, len(names))
			for i, name := range names {
				isDefault := ""
				if name == defaultName {
					isDefault = "*"
				}
				p, _ := config.GetProfile(name)
				acct := ""
				base := ""
				if p != nil {
					acct = p.AccountID
					base = p.BaseURL
				}
				rows[i] = []string{name, isDefault, acct, base}
			}
			output.FormatTable(f, headers, rows)
		} else {
			type profileInfo struct {
				Name      string `json:"name"`
				Default   bool   `json:"default"`
				AccountID string `json:"account_id"`
				BaseURL   string `json:"base_url"`
			}
			items := make([]profileInfo, len(names))
			for i, name := range names {
				p, _ := config.GetProfile(name)
				acct := ""
				base := ""
				if p != nil {
					acct = p.AccountID
					base = p.BaseURL
				}
				items[i] = profileInfo{
					Name:      name,
					Default:   name == defaultName,
					AccountID: acct,
					BaseURL:   base,
				}
			}
			output.SuccessList(items, len(items))
		}
	},
}

var profileAddCmd = &cobra.Command{
	Use:   "add [name]",
	Short: "Add a new profile",
	Long: `Add a new named profile with connection details.

Examples:
  keygen profile add prod --account-id ABC --base-url https://lm.example.com
  keygen profile add testing --account-id XYZ --base-url https://lm-test.example.com --email admin@test.com`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]

		// Check if profile already exists
		existing, _ := config.GetProfile(name)
		if existing != nil {
			output.Error(fmt.Sprintf("profile %q already exists. Use 'keygen profile edit %s' to modify it.", name, name))
			return
		}

		cfg := &config.Config{}

		if v, _ := cmd.Flags().GetString("account-id"); v != "" {
			cfg.AccountID = v
		}
		if v, _ := cmd.Flags().GetString("base-url"); v != "" {
			cfg.BaseURL = v
		}
		if v, _ := cmd.Flags().GetString("token"); v != "" {
			cfg.Token = v
		}
		if v, _ := cmd.Flags().GetString("email"); v != "" {
			cfg.Email = v
		}
		if v, _ := cmd.Flags().GetString("password"); v != "" {
			cfg.Password = v
		}

		if cfg.AccountID == "" || cfg.BaseURL == "" {
			output.Error("--account-id and --base-url are required when adding a profile")
			return
		}

		if err := config.SaveProfile(name, cfg); err != nil {
			output.Error("failed to save profile: " + err.Error())
			return
		}

		// If this is the first profile, make it the default
		names, _ := config.ListProfiles()
		if len(names) == 1 {
			_ = config.SetDefaultProfile(name)
		}

		maskedToken := maskToken(cfg.Token)

		output.Success(map[string]interface{}{
			"message":    fmt.Sprintf("Profile %q created", name),
			"profile":    name,
			"account_id": cfg.AccountID,
			"base_url":   cfg.BaseURL,
			"token":      maskedToken,
			"email":      cfg.Email,
		})
	},
}

var profileEditCmd = &cobra.Command{
	Use:   "edit [name]",
	Short: "Edit an existing profile",
	Long: `Edit an existing profile's connection details.

Only the flags you provide will be updated; other fields remain unchanged.

Examples:
  keygen profile edit prod --base-url https://new-url.example.com
  keygen profile edit testing --token newtoken123`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]

		cfg, err := config.GetProfile(name)
		if err != nil {
			output.Error(err.Error())
			return
		}

		changed := false
		if cmd.Flags().Changed("account-id") {
			v, _ := cmd.Flags().GetString("account-id")
			cfg.AccountID = v
			changed = true
		}
		if cmd.Flags().Changed("base-url") {
			v, _ := cmd.Flags().GetString("base-url")
			cfg.BaseURL = v
			changed = true
		}
		if cmd.Flags().Changed("token") {
			v, _ := cmd.Flags().GetString("token")
			cfg.Token = v
			changed = true
		}
		if cmd.Flags().Changed("email") {
			v, _ := cmd.Flags().GetString("email")
			cfg.Email = v
			changed = true
		}
		if cmd.Flags().Changed("password") {
			v, _ := cmd.Flags().GetString("password")
			cfg.Password = v
			changed = true
		}

		if !changed {
			output.Error("no update flags provided. Use --account-id, --base-url, --token, --email, or --password")
			return
		}

		if err := config.SaveProfile(name, cfg); err != nil {
			output.Error("failed to save profile: " + err.Error())
			return
		}

		output.Success(map[string]interface{}{
			"message":    fmt.Sprintf("Profile %q updated", name),
			"profile":    name,
			"account_id": cfg.AccountID,
			"base_url":   cfg.BaseURL,
		})
	},
}

var profileDeleteCmd = &cobra.Command{
	Use:   "delete [name]",
	Short: "Delete a profile",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]

		if err := config.DeleteProfile(name); err != nil {
			output.Error(err.Error())
			return
		}

		output.Success(map[string]interface{}{
			"message": fmt.Sprintf("Profile %q deleted", name),
			"profile": name,
		})
	},
}

var profileShowCmd = &cobra.Command{
	Use:   "show [name]",
	Short: "Show profile details",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]

		cfg, err := config.GetProfile(name)
		if err != nil {
			output.Error(err.Error())
			return
		}

		_, defaultName := config.ListProfiles()

		output.Success(map[string]interface{}{
			"profile":      name,
			"default":      name == defaultName,
			"account_id":   cfg.AccountID,
			"base_url":     cfg.BaseURL,
			"token":        maskToken(cfg.Token),
			"email":        cfg.Email,
			"has_password":  cfg.Password != "",
			"token_expiry": cfg.TokenExp,
		})
	},
}

var profileUseCmd = &cobra.Command{
	Use:   "use [name]",
	Short: "Set the default profile",
	Long: `Set which profile is used when --profile is not specified.

Examples:
  keygen profile use prod
  keygen profile use testing`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]

		if err := config.SetDefaultProfile(name); err != nil {
			output.Error(err.Error())
			return
		}

		output.Success(map[string]interface{}{
			"message": fmt.Sprintf("Default profile set to %q", name),
			"profile": name,
		})
	},
}

var profileRenameCmd = &cobra.Command{
	Use:   "rename [old-name] [new-name]",
	Short: "Rename a profile",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		oldName := args[0]
		newName := args[1]

		if err := config.RenameProfile(oldName, newName); err != nil {
			output.Error(err.Error())
			return
		}

		output.Success(map[string]interface{}{
			"message":  fmt.Sprintf("Profile %q renamed to %q", oldName, newName),
			"old_name": oldName,
			"new_name": newName,
		})
	},
}

func maskToken(t string) string {
	if t == "" {
		return ""
	}
	if len(t) > 8 {
		return t[:4] + strings.Repeat("*", len(t)-8) + t[len(t)-4:]
	}
	return "****"
}

func init() {
	profileAddCmd.Flags().String("account-id", "", "Keygen account ID")
	profileAddCmd.Flags().String("base-url", "", "Keygen API base URL")
	profileAddCmd.Flags().String("token", "", "API token")
	profileAddCmd.Flags().String("email", "", "Account email (for token refresh)")
	profileAddCmd.Flags().String("password", "", "Account password (for token refresh)")

	profileEditCmd.Flags().String("account-id", "", "Keygen account ID")
	profileEditCmd.Flags().String("base-url", "", "Keygen API base URL")
	profileEditCmd.Flags().String("token", "", "API token")
	profileEditCmd.Flags().String("email", "", "Account email")
	profileEditCmd.Flags().String("password", "", "Account password")

	profileCmd.AddCommand(profileListCmd)
	profileCmd.AddCommand(profileAddCmd)
	profileCmd.AddCommand(profileEditCmd)
	profileCmd.AddCommand(profileDeleteCmd)
	profileCmd.AddCommand(profileShowCmd)
	profileCmd.AddCommand(profileUseCmd)
	profileCmd.AddCommand(profileRenameCmd)
	rootCmd.AddCommand(profileCmd)
}
