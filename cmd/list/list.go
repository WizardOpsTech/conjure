package list

import (
	"fmt"

	"github.com/spf13/cobra"
)

// listCmd represents the list command
var ListCmd = &cobra.Command{
	Use:   "list",
	Short: "List the available templates and bundles",
	Long: `List the available templates and bundles.

Example: conjure list templates,
         conjure list bundles,
         conjure list templates -t yaml,
         conjure list bundles -t kubernetes`,
	Run: func(cmd *cobra.Command, args []string) {
		// When no subcommand is specified, list both templates and bundles
		fmt.Println("=== Templates ===")
		listTemplates("")
		fmt.Println()
		fmt.Println("=== Bundles ===")
		listBundles("")
	},
}

func init() {
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// listCmd.PersistentFlags().String("foo", "", "A help for foo")
	ListCmd.PersistentFlags().String("templates", "", "list single file templates such as secret.yaml or s3.tf")
	ListCmd.PersistentFlags().String("bundles", "", "list compiled bundles that consist of multiple templates.")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// listCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
