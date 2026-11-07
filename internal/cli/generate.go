package cli

import (
	"github.com/spf13/cobra"
)

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate deployment and infrastructure manifests",
	RunE: func(cmd *cobra.Command, args []string) error {
		cmd.Println("Manifest generator will generate declarative configs.")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(generateCmd)
}
