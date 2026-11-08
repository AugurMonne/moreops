package cli

import (
	"path/filepath"
	"sort"

	"github.com/spf13/cobra"
	"moreops/internal/core/template"
)

var templateCmd = &cobra.Command{
	Use:   "template",
	Short: "Manage and customize MoreOps project templates",
	Long: `Inspect, list, and eject built-in templates into your project.

MoreOps allows full customization for open-source teams. When you eject a template
into .moreops/templates/, MoreOps automatically uses your customized version during
code generation and manifest management.`,
}

var templateListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all customizable templates",
	Run: func(cmd *cobra.Command, args []string) {
		catalog := template.BuiltinTemplates()

		cmd.Println("\nAvailable Customizable Templates in MoreOps:")
		cmd.Println("===========================================")

		keys := make([]string, 0, len(catalog))
		for k := range catalog {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		for _, k := range keys {
			info := catalog[k]
			cmd.Printf("  • %-24s [%-10s] %s\n", info.ID, info.Category, info.Description)
		}

		cmd.Println("\nTip: Run 'moreops template eject <id>' or 'moreops template eject all' to copy into .moreops/templates/")
		cmd.Println()
	},
}

var templateEjectDir string

var templateEjectCmd = &cobra.Command{
	Use:   "eject [template-id | all]",
	Short: "Eject built-in templates into .moreops/templates/ for custom editing",
	RunE: func(cmd *cobra.Command, args []string) error {
		targetDir := templateEjectDir
		if targetDir == "" {
			targetDir = filepath.Join(".moreops", "templates")
		}

		ejected, err := template.Eject(targetDir, args...)
		if err != nil {
			return err
		}

		cmd.Println("\n✓ Successfully ejected templates for customization:")
		for _, f := range ejected {
			cmd.Printf("  + %s\n", f)
		}
		cmd.Printf("\nYou can now customize these templates. MoreOps will prioritize files in '%s'!\n\n", targetDir)
		return nil
	},
}

func init() {
	templateEjectCmd.Flags().StringVarP(&templateEjectDir, "dir", "d", "", "Target directory for ejected templates (default: .moreops/templates)")
	templateCmd.AddCommand(templateListCmd)
	templateCmd.AddCommand(templateEjectCmd)
	rootCmd.AddCommand(templateCmd)
}
