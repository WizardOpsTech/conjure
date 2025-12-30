package list

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/thesudoYT/conjure/internal/config"
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
		// Create AppConfig from the LoadConfig unmarshal function. Refer to AppConfig anywhere in the code.
		cfg, err := config.LoadConfig()
		if err != nil {
			fmt.Printf("Error loading config: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("list called")
		fmt.Println(cfg.TemplatesDir)
		fmt.Println(cfg.BundlesDir)
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
