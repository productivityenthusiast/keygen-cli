package cmd

import (
	"fmt"
	"strings"

	"github.com/productivityenthusiast/keygen-cli/internal/auth"
	"github.com/productivityenthusiast/keygen-cli/internal/output"
	"github.com/spf13/cobra"
)

var usersCmd = &cobra.Command{
	Use:   "users",
	Short: "Manage users",
}

var usersShowCmd = &cobra.Command{
	Use:   "show [user-id-or-email]",
	Short: "Show user details",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		cfg := loadConfig()
		client, err := auth.ResolveClient(cfg)
		if err != nil {
			output.Error(err.Error())
			return
		}

		identifier := args[0]
		var user *struct {
			ID        string                 `json:"id"`
			Email     string                 `json:"email"`
			FirstName string                 `json:"first_name"`
			LastName  string                 `json:"last_name"`
			Role      string                 `json:"role"`
			Status    string                 `json:"status"`
			Created   string                 `json:"created"`
			Updated   string                 `json:"updated"`
			Metadata  map[string]interface{} `json:"metadata,omitempty"`
			Licenses  int                    `json:"license_count"`
		}

		// Try as email first if it contains @
		if strings.Contains(identifier, "@") {
			u, err := client.FindUserByEmail(identifier)
			if err != nil {
				output.Error(err.Error())
				return
			}
			licenses, _ := client.GetUserLicenses(u.ID)
			user = &struct {
				ID        string                 `json:"id"`
				Email     string                 `json:"email"`
				FirstName string                 `json:"first_name"`
				LastName  string                 `json:"last_name"`
				Role      string                 `json:"role"`
				Status    string                 `json:"status"`
				Created   string                 `json:"created"`
				Updated   string                 `json:"updated"`
				Metadata  map[string]interface{} `json:"metadata,omitempty"`
				Licenses  int                    `json:"license_count"`
			}{u.ID, u.Email, u.FirstName, u.LastName, u.Role, u.Status, u.Created, u.Updated, u.Metadata, len(licenses)}
		} else {
			u, err := client.GetUser(identifier)
			if err != nil {
				output.Error(err.Error())
				return
			}
			licenses, _ := client.GetUserLicenses(u.ID)
			user = &struct {
				ID        string                 `json:"id"`
				Email     string                 `json:"email"`
				FirstName string                 `json:"first_name"`
				LastName  string                 `json:"last_name"`
				Role      string                 `json:"role"`
				Status    string                 `json:"status"`
				Created   string                 `json:"created"`
				Updated   string                 `json:"updated"`
				Metadata  map[string]interface{} `json:"metadata,omitempty"`
				Licenses  int                    `json:"license_count"`
			}{u.ID, u.Email, u.FirstName, u.LastName, u.Role, u.Status, u.Created, u.Updated, u.Metadata, len(licenses)}
		}

		output.Success(user)
	},
}

var usersStatusCmd = &cobra.Command{
	Use:   "status [user-id-or-email]",
	Short: "Show user status summary",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		cfg := loadConfig()
		client, err := auth.ResolveClient(cfg)
		if err != nil {
			output.Error(err.Error())
			return
		}

		identifier := args[0]
		var userID, userEmail string

		if strings.Contains(identifier, "@") {
			u, err := client.FindUserByEmail(identifier)
			if err != nil {
				output.Error(err.Error())
				return
			}
			userID = u.ID
			userEmail = u.Email
		} else {
			u, err := client.GetUser(identifier)
			if err != nil {
				output.Error(err.Error())
				return
			}
			userID = u.ID
			userEmail = u.Email
		}

		licenses, err := client.GetUserLicenses(userID)
		if err != nil {
			output.Error(err.Error())
			return
		}

		active := 0
		expiring := 0
		expired := 0
		suspended := 0
		totalMachines := 0
		totalComponents := 0

		for _, lic := range licenses {
			switch strings.ToUpper(lic.Status) {
			case "ACTIVE":
				active++
			case "EXPIRING":
				expiring++
			case "EXPIRED":
				expired++
			case "SUSPENDED":
				suspended++
			}

			machines, _ := client.GetLicenseMachines(lic.ID)
			if machines != nil {
				totalMachines += len(machines)
				for _, m := range machines {
					totalComponents += len(m.Components)
				}
			}
		}

		result := map[string]interface{}{
			"user_id":          userID,
			"email":            userEmail,
			"total_licenses":   len(licenses),
			"active":           active,
			"expiring":         expiring,
			"expired":          expired,
			"suspended":        suspended,
			"total_machines":   totalMachines,
			"total_components": totalComponents,
		}

		f := getFormat()
		if f == "table" || f == "csv" {
			headers := []string{"USER_ID", "EMAIL", "LICENSES", "ACTIVE", "EXPIRING", "EXPIRED", "MACHINES", "COMPONENTS"}
			rows := [][]string{{
				userID, userEmail,
				fmt.Sprintf("%d", len(licenses)),
				fmt.Sprintf("%d", active),
				fmt.Sprintf("%d", expiring),
				fmt.Sprintf("%d", expired),
				fmt.Sprintf("%d", totalMachines),
				fmt.Sprintf("%d", totalComponents),
			}}
			output.FormatTable(f, headers, rows)
		} else {
			output.Success(result)
		}
	},
}

func init() {
	usersCmd.AddCommand(usersShowCmd)
	usersCmd.AddCommand(usersStatusCmd)
	rootCmd.AddCommand(usersCmd)
}
