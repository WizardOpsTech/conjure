package list

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/thesudoYT/conjure/internal/config"
)

// AllowedTemplateTypes defines the permitted template file extensions
var AllowedTemplateTypes = []string{".yaml", ".tf", ".json"}

// isValidTemplateType checks if a file extension is in the allowed list
func isValidTemplateType(fileType string) bool {
	// Ensure the extension starts with a dot
	if !strings.HasPrefix(fileType, ".") {
		fileType = "." + fileType
	}

	for _, allowed := range AllowedTemplateTypes {
		if fileType == allowed {
			return true
		}
	}
	return false
}

var (
	// `conjure list templates` command
	listTemplatesCmd = &cobra.Command{
		Use:   "templates",
		Short: "List templates",
		Long: `List available templates in the templates_dir/templates directory.

Templates should be named with the pattern: <name>.<type>.tmpl
Examples: vsphere_vm.tf.tmpl, deployment.yaml.tmpl, config.json.tmpl

The .tmpl extension is hidden in the output for clarity.`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			// Validate --type flag if provided
			templateType, _ := cmd.Flags().GetString("type")
			if templateType != "" && !isValidTemplateType(templateType) {
				return fmt.Errorf("invalid template type '%s'. Allowed types: %v", templateType, AllowedTemplateTypes)
			}
			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			templateType, _ := cmd.Flags().GetString("type")
			listTemplates(templateType)
		},
		PostRun: func(cmd *cobra.Command, args []string) {
		},
	}
)

func init() {
	// Add commands
	ListCmd.AddCommand(listTemplatesCmd)

	// Add flags
	listTemplatesCmd.Flags().StringP("type", "t", "", "Filter templates by file extension (.yaml, .tf, .json)")
}

func listTemplates(filterType string) {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		os.Exit(1)
	}

	// Build the path to templates directory
	templatesPath := filepath.Join(cfg.TemplatesDir, "templates")

	// Check if the directory exists
	if _, err := os.Stat(templatesPath); os.IsNotExist(err) {
		fmt.Printf("Error: templates directory does not exist at %s\n", templatesPath)
		fmt.Printf("Please ensure %s/templates exists\n", cfg.TemplatesDir)
		os.Exit(1)
	}

	// Read directory contents
	entries, err := os.ReadDir(templatesPath)
	if err != nil {
		fmt.Printf("Error reading templates directory: %v\n", err)
		os.Exit(1)
	}

	// Print templates
	if len(entries) == 0 {
		fmt.Println("No templates found")
		return
	}

	// Normalize filterType to include dot prefix
	if filterType != "" && !strings.HasPrefix(filterType, ".") {
		filterType = "." + filterType
	}

	if filterType != "" {
		fmt.Printf("Available templates (type: %s):\n", filterType)
	} else {
		fmt.Println("Available templates:")
	}

	templatesFound := 0
	for _, entry := range entries {
		// Skip directories
		if entry.IsDir() {
			continue
		}

		// Only show files ending in .tmpl
		if !strings.HasSuffix(entry.Name(), ".tmpl") {
			continue
		}

		// Get the name without .tmpl extension
		nameWithoutTmpl := strings.TrimSuffix(entry.Name(), ".tmpl")

		// Filter by file type if specified
		// For a template named "vsphere_vm.tf.tmpl", we check if it ends with ".tf"
		if filterType != "" {
			if !strings.HasSuffix(nameWithoutTmpl, filterType) {
				continue
			}
		}

		// Display the template name without .tmpl (e.g., "vsphere_vm.tf" instead of "vsphere_vm.tf.tmpl")
		fmt.Printf("  %s\n", nameWithoutTmpl)
		templatesFound++
	}

	if templatesFound == 0 && filterType != "" {
		fmt.Printf("\nNo templates found with extension '%s'\n", filterType)
	}
}
