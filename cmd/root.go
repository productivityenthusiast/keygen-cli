package cmd

import (
	"fmt"
	"os"

	"github.com/productivityenthusiast/keygen-cli/internal/config"
	"github.com/spf13/cobra"
)

var (
	cfgFile     string
	format      string
	quiet       bool
	verbose     bool
	envFile     string
	accountID   string
	baseURL     string
	token       string
	profileName string

	Version = "dev"
)

var rootCmd = &cobra.Command{
	Use:   "keygen",
	Short: "Keygen CLI - License management from the command line",
	Long: `keygen-cli is a command line tool for managing Keygen licenses,
machines, components, and users. Designed for use in scripts and by LLMs.

Supports multiple profiles (like AWS CLI). The --profile flag is required
for all authenticated commands. Use 'keygen profile --help' to manage profiles.

Examples:
  keygen profile add prod --account-id ABC --base-url https://lm.example.com
  keygen login token --token mytoken --profile prod
  keygen licenses list --profile prod
  keygen status --profile prod --format table

All output is JSON by default. Use --format table or --format csv for alternatives.`,
	Version: Version,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func SetVersion(v string) {
	Version = v
	rootCmd.Version = v
}

func init() {
	rootCmd.PersistentFlags().StringVar(&profileName, "profile", "", "Named profile to use (required for all authenticated commands)")
	rootCmd.PersistentFlags().StringVar(&format, "format", "json", "Output format: json, table, csv")
	rootCmd.PersistentFlags().BoolVar(&quiet, "quiet", false, "Suppress non-essential output")
	rootCmd.PersistentFlags().BoolVar(&verbose, "verbose", false, "Show verbose output")
	rootCmd.PersistentFlags().StringVar(&envFile, "env-file", "", "Path to .env file")
	rootCmd.PersistentFlags().StringVar(&accountID, "account-id", "", "Keygen account ID")
	rootCmd.PersistentFlags().StringVar(&baseURL, "base-url", "", "Keygen API base URL")
	rootCmd.PersistentFlags().StringVar(&token, "token", "", "Keygen API token")
}

// loadConfig loads the active profile. Requires --profile to be set.
// If --profile is not provided, prints an error with available profiles and exits.
func loadConfig() *config.Config {
	if profileName == "" {
		names, defaultName := config.ListProfiles()
		if len(names) == 0 {
			fmt.Fprintln(os.Stderr, "Error: --profile is required. No profiles configured yet.")
			fmt.Fprintln(os.Stderr, "  Run: keygen profile add <name> --account-id <id> --base-url <url>")
		} else {
			fmt.Fprintln(os.Stderr, "Error: --profile is required. Available profiles:")
			for _, n := range names {
				marker := "  "
				if n == defaultName {
					marker = "* "
				}
				fmt.Fprintf(os.Stderr, "  %s%s\n", marker, n)
			}
			fmt.Fprintf(os.Stderr, "\n  Example: keygen <command> --profile %s\n", names[0])
		}
		os.Exit(1)
	}

	cfg := config.LoadProfile(profileName)

	// CLI flags override everything
	if accountID != "" {
		cfg.AccountID = accountID
	}
	if baseURL != "" {
		cfg.BaseURL = baseURL
	}
	if token != "" {
		cfg.Token = token
	}

	return cfg
}

func getFormat() string {
	return format
}

func exitError(msg string) {
	fmt.Fprintf(os.Stderr, "Error: %s\n", msg)
	os.Exit(1)
}
