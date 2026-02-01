package list

import (
	"fmt"

	"github.com/spf13/cobra"
)

var ListCmd = &cobra.Command{
	Use:   "list",
	Short: "List the available templates and bundles",
	Long: `List the available templates and bundles.

Example: conjure list templates,
         conjure list bundles,
         conjure list templates -t yaml,
         conjure list bundles -t kubernetes`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("=== Templates ===")
		listTemplates("", false)
		fmt.Println()
		fmt.Println("=== Bundles ===")
		listBundles("", false)
	},
}
