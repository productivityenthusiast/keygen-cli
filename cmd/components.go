package cmd

import (
	"fmt"

	"github.com/productivityenthusiast/keygen-cli/internal/auth"
	"github.com/productivityenthusiast/keygen-cli/internal/output"
	"github.com/spf13/cobra"
)

var componentsCmd = &cobra.Command{
	Use:   "components",
	Short: "Manage machine components",
}

var componentsCheckCmd = &cobra.Command{
	Use:   "check [fingerprint]",
	Short: "Check if a component fingerprint is registered",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		cfg := loadConfig()
		client, err := auth.ResolveClient(cfg)
		if err != nil {
			output.Error(err.Error())
			return
		}

		fingerprint := args[0]
		comp, err := client.FindComponentByFingerprint(fingerprint)
		if err != nil {
			output.Error(err.Error())
			return
		}

		if comp != nil {
			output.Success(map[string]interface{}{
				"found":       true,
				"fingerprint": comp.Fingerprint,
				"id":          comp.ID,
				"name":        comp.Name,
				"machine_id":  comp.MachineID,
			})
		} else {
			output.Success(map[string]interface{}{
				"found":       false,
				"fingerprint": fingerprint,
			})
		}
	},
}

var componentsDeleteCmd = &cobra.Command{
	Use:   "delete [fingerprint]",
	Short: "Delete a component by fingerprint",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		cfg := loadConfig()
		client, err := auth.ResolveClient(cfg)
		if err != nil {
			output.Error(err.Error())
			return
		}

		fingerprint := args[0]
		force, _ := cmd.Flags().GetBool("force")

		comp, err := client.FindComponentByFingerprint(fingerprint)
		if err != nil {
			output.Error(err.Error())
			return
		}

		if comp == nil {
			output.Error(fmt.Sprintf("component not found: %s", fingerprint))
			return
		}

		if !force {
			output.Success(map[string]interface{}{
				"action":      "delete",
				"fingerprint": comp.Fingerprint,
				"id":          comp.ID,
				"name":        comp.Name,
				"machine_id":  comp.MachineID,
				"confirm":     "use --force to confirm deletion",
			})
			return
		}

		if err := client.DeleteComponent(comp.ID); err != nil {
			output.Error("delete failed: " + err.Error())
			return
		}

		output.Success(map[string]interface{}{
			"deleted":     true,
			"fingerprint": comp.Fingerprint,
			"id":          comp.ID,
			"name":        comp.Name,
		})
	},
}

func init() {
	componentsDeleteCmd.Flags().Bool("force", false, "Skip confirmation")

	componentsCmd.AddCommand(componentsCheckCmd)
	componentsCmd.AddCommand(componentsDeleteCmd)
	rootCmd.AddCommand(componentsCmd)
}
