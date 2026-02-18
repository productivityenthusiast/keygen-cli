package cmd

import (
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/productivityenthusiast/keygen-cli/internal/auth"
	"github.com/productivityenthusiast/keygen-cli/internal/output"
	"github.com/spf13/cobra"
)

var licensesCmd = &cobra.Command{
	Use:   "licenses",
	Short: "Manage licenses",
}

var licensesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List licenses",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := loadConfig()
		client, err := auth.ResolveClient(cfg)
		if err != nil {
			output.Error(err.Error())
			return
		}

		params := make(map[string]string)
		if v, _ := cmd.Flags().GetString("user"); v != "" {
			params["user"] = v
		}
		if v, _ := cmd.Flags().GetString("product"); v != "" {
			params["product"] = v
		}
		if v, _ := cmd.Flags().GetString("policy"); v != "" {
			params["policy"] = v
		}
		if v, _ := cmd.Flags().GetString("status"); v != "" {
			params["status"] = v
		}
		if v, _ := cmd.Flags().GetInt("limit"); v > 0 {
			params["page[size]"] = fmt.Sprintf("%d", v)
		}
		if v, _ := cmd.Flags().GetInt("page"); v > 0 {
			params["page[number]"] = fmt.Sprintf("%d", v)
		}

		licenses, err := client.ListLicenses(params)
		if err != nil {
			output.Error(err.Error())
			return
		}

		f := getFormat()
		if f == "table" || f == "csv" {
			headers := []string{"ID", "KEY", "NAME", "STATUS", "EXPIRY", "OWNER_ID"}
			rows := make([][]string, len(licenses))
			for i, l := range licenses {
				rows[i] = []string{l.ID, l.Key, l.Name, strings.ToUpper(l.Status), l.Expiry, l.OwnerID}
			}
			output.FormatTable(f, headers, rows)
		} else {
			output.SuccessList(licenses, len(licenses))
		}
	},
}

var licensesShowCmd = &cobra.Command{
	Use:   "show [license-id]",
	Short: "Show license details",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		cfg := loadConfig()
		client, err := auth.ResolveClient(cfg)
		if err != nil {
			output.Error(err.Error())
			return
		}

		license, err := client.GetLicense(args[0])
		if err != nil {
			output.Error(err.Error())
			return
		}

		output.Success(license)
	},
}

var licensesStatusCmd = &cobra.Command{
	Use:   "status [license-id]",
	Short: "Check license status with validation",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		cfg := loadConfig()
		client, err := auth.ResolveClient(cfg)
		if err != nil {
			output.Error(err.Error())
			return
		}

		validation, license, err := client.ValidateLicense(args[0])
		if err != nil {
			output.Error(err.Error())
			return
		}

		machines, _ := client.GetLicenseMachines(args[0])
		machineCount := 0
		componentCount := 0
		if machines != nil {
			machineCount = len(machines)
			for _, m := range machines {
				componentCount += len(m.Components)
			}
		}

		daysRemaining := 0.0
		if license != nil && license.Expiry != "" {
			if t, err := time.Parse(time.RFC3339, license.Expiry); err == nil {
				daysRemaining = math.Max(0, time.Until(t).Hours()/24)
			}
		}

		result := map[string]interface{}{
			"license_id":      args[0],
			"valid":           validation.Valid,
			"status":          "",
			"detail":          validation.Detail,
			"code":            validation.Code,
			"machines":        machineCount,
			"components":      componentCount,
			"days_remaining":  int(daysRemaining),
		}

		if license != nil {
			result["status"] = license.Status
			result["key"] = license.Key
			result["name"] = license.Name
			result["expiry"] = license.Expiry
		}

		f := getFormat()
		if f == "table" || f == "csv" {
			headers := []string{"LICENSE_ID", "VALID", "STATUS", "DAYS_LEFT", "MACHINES", "COMPONENTS"}
			rows := [][]string{{
				args[0],
				fmt.Sprintf("%v", validation.Valid),
				fmt.Sprintf("%v", result["status"]),
				fmt.Sprintf("%d", int(daysRemaining)),
				fmt.Sprintf("%d", machineCount),
				fmt.Sprintf("%d", componentCount),
			}}
			output.FormatTable(f, headers, rows)
		} else {
			output.Success(result)
		}
	},
}

var licensesRenewCmd = &cobra.Command{
	Use:   "renew [license-id]",
	Short: "Renew a license",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		cfg := loadConfig()
		client, err := auth.ResolveClient(cfg)
		if err != nil {
			output.Error(err.Error())
			return
		}

		// Get current license for old expiry
		oldLicense, err := client.GetLicense(args[0])
		if err != nil {
			output.Error("failed to get current license: " + err.Error())
			return
		}

		renewed, err := client.RenewLicense(args[0])
		if err != nil {
			output.Error(err.Error())
			return
		}

		output.Success(map[string]interface{}{
			"license_id": renewed.ID,
			"key":        renewed.Key,
			"name":       renewed.Name,
			"old_expiry": oldLicense.Expiry,
			"new_expiry": renewed.Expiry,
			"status":     renewed.Status,
		})
	},
}

var licensesComponentsCmd = &cobra.Command{
	Use:   "components [license-id]",
	Short: "List all components for a license",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		cfg := loadConfig()
		client, err := auth.ResolveClient(cfg)
		if err != nil {
			output.Error(err.Error())
			return
		}

		machines, err := client.GetLicenseMachines(args[0])
		if err != nil {
			output.Error(err.Error())
			return
		}

		type componentInfo struct {
			ID          string `json:"id"`
			Fingerprint string `json:"fingerprint"`
			Name        string `json:"name"`
			MachineID   string `json:"machine_id"`
			MachineFP   string `json:"machine_fingerprint"`
		}

		var allComponents []componentInfo
		for _, m := range machines {
			for _, c := range m.Components {
				allComponents = append(allComponents, componentInfo{
					ID:          c.ID,
					Fingerprint: c.Fingerprint,
					Name:        c.Name,
					MachineID:   m.ID,
					MachineFP:   m.Fingerprint,
				})
			}
		}

		f := getFormat()
		if f == "table" || f == "csv" {
			headers := []string{"ID", "FINGERPRINT", "NAME", "MACHINE_ID", "MACHINE_FP"}
			rows := make([][]string, len(allComponents))
			for i, c := range allComponents {
				rows[i] = []string{c.ID, c.Fingerprint, c.Name, c.MachineID, c.MachineFP}
			}
			output.FormatTable(f, headers, rows)
		} else {
			output.SuccessList(allComponents, len(allComponents))
		}
	},
}

func init() {
	licensesListCmd.Flags().String("user", "", "Filter by user ID")
	licensesListCmd.Flags().String("product", "", "Filter by product ID")
	licensesListCmd.Flags().String("policy", "", "Filter by policy ID")
	licensesListCmd.Flags().String("status", "", "Filter by status")
	licensesListCmd.Flags().Int("limit", 10, "Results per page")
	licensesListCmd.Flags().Int("page", 1, "Page number")

	licensesCmd.AddCommand(licensesListCmd)
	licensesCmd.AddCommand(licensesShowCmd)
	licensesCmd.AddCommand(licensesStatusCmd)
	licensesCmd.AddCommand(licensesRenewCmd)
	licensesCmd.AddCommand(licensesComponentsCmd)
	rootCmd.AddCommand(licensesCmd)
}
