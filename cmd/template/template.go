/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package template

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/spf13/cobra"
	"github.com/thesudoYT/conjure/internal/config"
	"github.com/thesudoYT/conjure/internal/metadata"
	"github.com/thesudoYT/conjure/internal/prompt"
	"go.yaml.in/yaml/v3"
)

// TemplateCmd represents the template command
var TemplateCmd = &cobra.Command{
	Use:   "template <template-name>",
	Short: "Generate a file from a template",
	Long: `Generate a file from a template with variable substitution.

Templates are named with the pattern: <name>.<type>.tmpl (e.g., deployment.yaml.tmpl)
The template name should not include the .tmpl extension when invoking.

Examples:
  conjure template deployment.yaml -o ./deployment.yaml
  conjure template deployment.yaml -o ./deployment.yaml -f values.yaml
  conjure template vsphere_vm.tf -o ./main.tf --var vm_name=myvm --var cpu=4`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		templateName := args[0]
		outputPath, _ := cmd.Flags().GetString("output")
		vars, _ := cmd.Flags().GetStringArray("var")
		valuesFile, _ := cmd.Flags().GetString("values")

		// Determine if interactive mode should be enabled
		// If -i flag was explicitly set, use that value
		// Otherwise, enable interactive if no --var or -f flags provided
		interactive := cmd.Flags().Changed("interactive")
		if interactive {
			interactive, _ = cmd.Flags().GetBool("interactive")
		} else {
			// Auto-enable interactive if no variables provided
			interactive = len(vars) == 0 && valuesFile == ""
		}

		if err := generateTemplate(templateName, outputPath, vars, interactive, valuesFile); err != nil {
			fmt.Printf("Error generating template: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	// Add flags
	TemplateCmd.Flags().StringP("output", "o", "", "Output file path (required)")
	_ = TemplateCmd.MarkFlagRequired("output")
	TemplateCmd.Flags().StringArrayP("var", "v", []string{}, "Set template variables (format: key=value)")
	TemplateCmd.Flags().BoolP("interactive", "i", false, "Enable/disable interactive mode (default: auto-enabled when no --var or -f provided)")
	TemplateCmd.Flags().StringP("values", "f", "", "Provide variables using a values.yaml file")
}

func generateTemplate(templateName, outputPath string, varsList []string, interactive bool, valuesFile string) error {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Build path to template file
	templatePath := filepath.Join(cfg.TemplatesDir, "templates", templateName+".tmpl")

	// Check if template exists
	if _, err := os.Stat(templatePath); os.IsNotExist(err) {
		return fmt.Errorf("template file does not exist: %s", templatePath)
	}

	// Read template file
	templateContent, err := os.ReadFile(templatePath)
	if err != nil {
		return fmt.Errorf("failed to read template file: %w", err)
	}

	// Start with an empty variables map
	userVariables := make(map[string]interface{})

	// First, load values from values.yaml file if provided
	if valuesFile != "" {
		values, err := parseValues(valuesFile)
		if err != nil {
			return fmt.Errorf("failed to parse values file: %w", err)
		}
		// Copy values from YAML into userVariables
		for k, v := range values {
			userVariables[k] = v
		}
	}

	// Then, parse variables from --var flags (these override values.yaml)
	cliVariables, err := parseVariables(varsList)
	if err != nil {
		return fmt.Errorf("failed to parse variables: %w", err)
	}
	// Merge CLI variables, overriding any values from YAML
	for k, v := range cliVariables {
		userVariables[k] = v
	}

	// Try to load metadata file
	metadataPath := metadata.GetMetadataPath(cfg.TemplatesDir, templateName)
	var finalVariables map[string]interface{}

	if _, err := os.Stat(metadataPath); err == nil {
		// Metadata file exists - load and validate
		meta, err := metadata.LoadTemplateMetadata(metadataPath)
		if err != nil {
			return fmt.Errorf("failed to load metadata: %w", err)
		}

		fmt.Printf("Using metadata: %s\n\n", meta.Description)

		// If interactive mode is enabled, prompt for variables
		if interactive {
			// Use interactive prompts to collect variables
			finalVariables, err = prompt.CollectVariables(meta, userVariables)
			if err != nil {
				return fmt.Errorf("failed to collect variables: %w", err)
			}
		} else {
			// Non-interactive mode - validate and merge with defaults
			finalVariables, err = metadata.ValidateAndMergeVariables(meta, userVariables)
			if err != nil {
				return fmt.Errorf("variable validation failed: %w", err)
			}
		}

		// Show what variables are being used
		fmt.Println("\nVariables:")
		for _, v := range meta.Variables {
			if val, exists := finalVariables[v.Name]; exists {
				fmt.Printf("  %s = %v\n", v.Name, val)
			}
		}
		fmt.Println()
	} else {
		// No metadata file - use variables as-is (legacy mode)
		fmt.Println("Warning: No metadata file found, using variables without validation")
		finalVariables = userVariables
	}

	// Render template
	rendered, err := renderTemplate(string(templateContent), finalVariables)
	if err != nil {
		return fmt.Errorf("failed to render template: %w", err)
	}

	// Create output directory if it doesn't exist
	outputDir := filepath.Dir(outputPath)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Write rendered content to output file
	if err := os.WriteFile(outputPath, []byte(rendered), 0644); err != nil {
		return fmt.Errorf("failed to write output file: %w", err)
	}

	fmt.Printf("✓ Template rendered successfully\n")
	fmt.Printf("✓ Output written to: %s\n", outputPath)

	return nil
}

// parseVariables parses --var flags in the format key=value
func parseVariables(varsList []string) (map[string]interface{}, error) {
	variables := make(map[string]interface{})

	for _, v := range varsList {
		parts := strings.SplitN(v, "=", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid variable format '%s', expected key=value", v)
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		if key == "" {
			return nil, fmt.Errorf("variable key cannot be empty")
		}

		variables[key] = value
	}

	return variables, nil
}

// renderTemplate renders a Go template with the provided variables
func renderTemplate(templateContent string, variables map[string]interface{}) (string, error) {
	// Create a new template with strict error checking for missing keys
	tmpl, err := template.New("template").Option("missingkey=error").Parse(templateContent)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	// Execute template with variables
	var rendered bytes.Buffer
	if err := tmpl.Execute(&rendered, variables); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return rendered.String(), nil
}

// parseValues parses a values.yaml file
func parseValues(valuesFile string) (map[string]interface{}, error) {
	var values map[string]interface{}

	// Read the values.yaml file provided
	data, err := os.ReadFile(valuesFile)
	if err != nil {
		return nil, fmt.Errorf("error reading values file: %w", err)
	}

	// Unmarshal the data
	if err := yaml.Unmarshal(data, &values); err != nil {
		return nil, fmt.Errorf("error unmarshalling values yaml: %w", err)
	}

	return values, nil
}
