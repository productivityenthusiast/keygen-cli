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
(devices, printers, servers based on license metadata).

Use --user to scope to a single user's licenses.
Use --fields to choose which columns to display (comma-separated).

Available fields:
  key, name, status, days, owner, machines, devices, printers, servers

Examples:
  keygen status
  keygen status --user admin@example.com
  keygen status --fields status,devices,printers,servers --format table
  keygen status --user admin@example.com --fields key,status,days`,
	Run: func(cmd *cobra.Command, args []string) {
		cfg := loadConfig()
		client, err := auth.ResolveClient(cfg)
		if err != nil {
			output.Error(err.Error())
			return
		}

		// Resolve --user flag
		userFilter, _ := cmd.Flags().GetString("user")
		var filterUserID, filterUserEmail string

		if userFilter != "" {
			if strings.Contains(userFilter, "@") {
				u, err := client.FindUserByEmail(userFilter)
				if err != nil {
					output.Error("user not found: " + err.Error())
					return
				}
				filterUserID = u.ID
				filterUserEmail = u.Email
			} else {
				u, err := client.GetUser(userFilter)
				if err != nil {
					output.Error("user not found: " + err.Error())
					return
				}
				filterUserID = u.ID
				filterUserEmail = u.Email
			}
		}

		// Fetch licenses — scoped to user if filter provided
		var licenseParams = map[string]string{"page[size]": "100", "page[number]": "1"}
		if filterUserID != "" {
			licenseParams["user"] = filterUserID
		}
		licenses, errL := client.ListLicenses(licenseParams)
		if errL != nil {
			output.Error("failed to fetch licenses: " + errL.Error())
			return
		}

		// Only fetch account-wide counts when not filtering by user
		userCount := 0
		productCount := 0
		if filterUserID == "" {
			users, errU := client.ListUsers(map[string]string{"page[size]": "100", "page[number]": "1"})
			if errU == nil {
				userCount = len(users)
			}
			products, errP := client.ListProducts()
			if errP == nil {
				productCount = len(products)
			}
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

			// Resolve owner email (skip if we already know from --user)
			if lic.OwnerID != "" {
				if filterUserEmail != "" && lic.OwnerID == filterUserID {
					d.OwnerEmail = filterUserEmail
				} else if u, err := client.GetUser(lic.OwnerID); err == nil {
					d.OwnerEmail = u.Email
				}
			}

			details = append(details, d)
		}

		// Parse --fields flag
		fieldsFlag, _ := cmd.Flags().GetString("fields")
		allFields := []string{"key", "name", "status", "days", "owner", "machines", "devices", "printers", "servers"}
		selectedFields := allFields // default: all

		if fieldsFlag != "" {
			selectedFields = nil
			for _, f := range strings.Split(fieldsFlag, ",") {
				f = strings.TrimSpace(strings.ToLower(f))
				if f != "" {
					selectedFields = append(selectedFields, f)
				}
			}
		}

		// Build a set for quick lookup
		fieldSet := map[string]bool{}
		for _, f := range selectedFields {
			fieldSet[f] = true
		}

		// Build JSON result — filter fields for JSON output too
		result := map[string]interface{}{}

		if filterUserID != "" {
			result["user_id"] = filterUserID
			result["user_email"] = filterUserEmail
		} else {
			result["account_id"] = cfg.AccountID
			result["base_url"] = cfg.BaseURL
			result["total_users"] = userCount
			result["total_products"] = productCount
		}
		result["total_licenses"] = len(licenses)
		result["total_machines"] = totalMachines
		result["total_components"] = totalComponents
		result["license_statuses"] = statusCounts

		// Filter license detail fields for JSON
		filteredDetails := make([]map[string]interface{}, len(details))
		for i, d := range details {
			m := map[string]interface{}{"id": d.ID}
			if fieldSet["key"] {
				m["key"] = d.Key
			}
			if fieldSet["name"] {
				m["name"] = d.Name
			}
			if fieldSet["status"] {
				m["status"] = d.Status
			}
			if fieldSet["days"] {
				m["days_remaining"] = d.DaysRemaining
				m["expiry"] = d.Expiry
			}
			if fieldSet["owner"] {
				m["owner_email"] = d.OwnerEmail
			}
			if fieldSet["machines"] {
				m["machines"] = d.Machines
			}
			if fieldSet["devices"] {
				m["devices"] = d.Devices
				m["max_devices"] = d.MaxDevices
			}
			if fieldSet["printers"] {
				m["printers"] = d.Printers
				m["max_printers"] = d.MaxPrinters
			}
			if fieldSet["servers"] {
				m["servers"] = d.Servers
				m["max_servers"] = d.MaxServers
			}
			filteredDetails[i] = m
		}
		result["licenses"] = filteredDetails

		f := getFormat()
		if f == "table" || f == "csv" {
			// Map field names to headers and row builders
			type colDef struct {
				header string
				value  func(d licenseDetail) string
			}
			colMap := map[string]colDef{
				"key":      {"KEY", func(d licenseDetail) string { return d.Key }},
				"name":     {"NAME", func(d licenseDetail) string { return d.Name }},
				"status":   {"STATUS", func(d licenseDetail) string { return d.Status }},
				"days":     {"DAYS", func(d licenseDetail) string { return fmt.Sprintf("%d", d.DaysRemaining) }},
				"owner":    {"OWNER", func(d licenseDetail) string { return d.OwnerEmail }},
				"machines": {"MACHINES", func(d licenseDetail) string { return fmt.Sprintf("%d", d.Machines) }},
				"devices":  {"DEVICES", func(d licenseDetail) string { return fmt.Sprintf("%d/%s", d.Devices, d.MaxDevices) }},
				"printers": {"PRINTERS", func(d licenseDetail) string { return fmt.Sprintf("%d/%s", d.Printers, d.MaxPrinters) }},
				"servers":  {"SERVERS", func(d licenseDetail) string { return fmt.Sprintf("%d/%s", d.Servers, d.MaxServers) }},
			}

			var headers []string
			var builders []func(d licenseDetail) string
			for _, field := range selectedFields {
				if c, ok := colMap[field]; ok {
					headers = append(headers, c.header)
					builders = append(builders, c.value)
				}
			}

			rows := make([][]string, len(details))
			for i, d := range details {
				row := make([]string, len(builders))
				for j, fn := range builders {
					row[j] = fn(d)
				}
				rows[i] = row
			}

			// Print summary header
			if filterUserID != "" {
				fmt.Printf("User: %s | Licenses: %d | Machines: %d | Components: %d\n\n",
					filterUserEmail, len(licenses), totalMachines, totalComponents)
			} else {
				fmt.Printf("Account: %s | Licenses: %d | Users: %d | Products: %d | Machines: %d | Components: %d\n\n",
					cfg.AccountID, len(licenses), userCount, productCount, totalMachines, totalComponents)
			}
			output.FormatTable(f, headers, rows)
		} else {
			output.Success(result)
		}
	},
}

func init() {
	statusCmd.Flags().String("user", "", "Filter by user ID or email")
	statusCmd.Flags().String("fields", "", "Comma-separated fields to show: key,name,status,days,owner,machines,devices,printers,servers")
	rootCmd.AddCommand(statusCmd)
}
