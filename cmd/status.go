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

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show account summary (licenses, users, products, components)",
	Long: `Display a high-level summary of the Keygen account:
total licenses, users, products, and per-license component breakdowns
(devices, printers, servers based on license metadata).`,
	Run: func(cmd *cobra.Command, args []string) {
		cfg := loadConfig()
		client, err := auth.ResolveClient(cfg)
		if err != nil {
			output.Error(err.Error())
			return
		}

		// Fetch licenses, users, products in sequence
		licenses, errL := client.ListLicenses(map[string]string{"page[size]": "100", "page[number]": "1"})
		if errL != nil {
			output.Error("failed to fetch licenses: " + errL.Error())
			return
		}

		users, errU := client.ListUsers(map[string]string{"page[size]": "100", "page[number]": "1"})
		if errU != nil {
			output.Error("failed to fetch users: " + errU.Error())
			return
		}

		products, errP := client.ListProducts()
		if errP != nil {
			output.Error("failed to fetch products: " + errP.Error())
			return
		}

		// Aggregate license statuses
		statusCounts := map[string]int{}
		for _, lic := range licenses {
			s := strings.ToUpper(lic.Status)
			statusCounts[s]++
		}

		// Per-license details with component breakdown
		type licenseDetail struct {
			ID            string `json:"id"`
			Key           string `json:"key"`
			Name          string `json:"name"`
			Status        string `json:"status"`
			Expiry        string `json:"expiry"`
			DaysRemaining int    `json:"days_remaining"`
			OwnerEmail    string `json:"owner_email,omitempty"`
			Machines      int    `json:"machines"`
			MaxDevices    string `json:"max_devices,omitempty"`
			MaxPrinters   string `json:"max_printers,omitempty"`
			MaxServers    string `json:"max_servers,omitempty"`
			Devices       int    `json:"devices"`
			Printers      int    `json:"printers"`
			Servers       int    `json:"servers"`
		}

		var details []licenseDetail
		totalMachines := 0
		totalComponents := 0

		for _, lic := range licenses {
			d := licenseDetail{
				ID:     lic.ID,
				Key:    lic.Key,
				Name:   lic.Name,
				Status: strings.ToUpper(lic.Status),
				Expiry: lic.Expiry,
			}

			// Days remaining
			if lic.Expiry != "" {
				if t, err := time.Parse(time.RFC3339, lic.Expiry); err == nil {
					d.DaysRemaining = int(math.Max(0, time.Until(t).Hours()/24))
				}
			}

			// Metadata limits
			if lic.Metadata != nil {
				if v, ok := lic.Metadata["maxDevices"]; ok {
					d.MaxDevices = fmt.Sprintf("%v", v)
				}
				if v, ok := lic.Metadata["maxPrinters"]; ok {
					d.MaxPrinters = fmt.Sprintf("%v", v)
				}
				if v, ok := lic.Metadata["maxServers"]; ok {
					d.MaxServers = fmt.Sprintf("%v", v)
				}
			}

			// Fetch machines + components for this license
			machines, err := client.GetLicenseMachines(lic.ID)
			if err == nil && machines != nil {
				d.Machines = len(machines)
				totalMachines += len(machines)
				for _, m := range machines {
					totalComponents += len(m.Components)
					for _, comp := range m.Components {
						name := strings.ToLower(comp.Name)
						switch {
						case strings.Contains(name, "printer") || strings.Contains(name, "print"):
							d.Printers++
						case strings.Contains(name, "server") || strings.Contains(name, "srv"):
							d.Servers++
						default:
							d.Devices++
						}
					}
				}
			}

			// Resolve owner email
			if lic.OwnerID != "" {
				if u, err := client.GetUser(lic.OwnerID); err == nil {
					d.OwnerEmail = u.Email
				}
			}

			details = append(details, d)
		}

		result := map[string]interface{}{
			"account_id":       cfg.AccountID,
			"base_url":         cfg.BaseURL,
			"total_licenses":   len(licenses),
			"total_users":      len(users),
			"total_products":   len(products),
			"total_machines":   totalMachines,
			"total_components": totalComponents,
			"license_statuses": statusCounts,
			"licenses":         details,
		}

		f := getFormat()
		if f == "table" || f == "csv" {
			headers := []string{"KEY", "NAME", "STATUS", "DAYS", "OWNER", "MACHINES", "DEVICES", "PRINTERS", "SERVERS"}
			rows := make([][]string, len(details))
			for i, d := range details {
				rows[i] = []string{
					d.Key,
					d.Name,
					d.Status,
					fmt.Sprintf("%d", d.DaysRemaining),
					d.OwnerEmail,
					fmt.Sprintf("%d", d.Machines),
					fmt.Sprintf("%d/%s", d.Devices, d.MaxDevices),
					fmt.Sprintf("%d/%s", d.Printers, d.MaxPrinters),
					fmt.Sprintf("%d/%s", d.Servers, d.MaxServers),
				}
			}
			// Print summary header
			fmt.Printf("Account: %s | Licenses: %d | Users: %d | Products: %d | Machines: %d | Components: %d\n\n",
				cfg.AccountID, len(licenses), len(users), len(products), totalMachines, totalComponents)
			output.FormatTable(f, headers, rows)
		} else {
			output.Success(result)
		}
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
