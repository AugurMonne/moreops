package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show MoreOps version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("MoreOps version %s\n", Version)
	},
}
