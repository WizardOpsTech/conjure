package bundle

import (
	"bytes"
	"encoding/json"
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

// BundleCmd represents the bundle command
var BundleCmd = &cobra.Command{
	Use:   "bundle <bundle-name>",
	Short: "Generate files from a bundle of templates",
	Long: `Generate multiple files from a bundle of templates with variable substitution.

Bundles are collections of related templates that are rendered together.
Each bundle has a conjure.json metadata file that defines shared and template-specific variables.

Variable Overrides:
  - Shared variables apply to all templates
  - Template-specific overrides using: template.tmpl:key=value
  - Use template_overrides section in values.yaml
  - Interactive mode prompts for overrides after collecting main variables

Examples:
  conjure bundle kubernetes-deployment -o ./k8s/
  conjure bundle kubernetes-deployment -o ./k8s/ -f values.yaml
  conjure bundle kubernetes-deployment -o ./k8s/ --var app_name=myapp --var replicas=5
  conjure bundle kubernetes-deployment -o ./k8s/ --var namespace=prod --var ingress.yaml.tmpl:namespace=ingress-nginx
  conjure bundle kubernetes-deployment -o ./k8s/ -i`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		bundleName := args[0]
		outputPath, _ := cmd.Flags().GetString("output")
		vars, _ := cmd.Flags().GetStringArray("var")
		valuesFile, _ := cmd.Flags().GetString("values")

		// Determine if interactive mode should be enabled
		interactive := cmd.Flags().Changed("interactive")
		if interactive {
			interactive, _ = cmd.Flags().GetBool("interactive")
		} else {
			// Auto-enable interactive if no variables provided
			interactive = len(vars) == 0 && valuesFile == ""
		}

		if err := generateBundle(bundleName, outputPath, vars, interactive, valuesFile); err != nil {
			fmt.Printf("Error generating bundle: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	// Add flags
	BundleCmd.Flags().StringP("output", "o", "", "Output directory path (required)")
	_ = BundleCmd.MarkFlagRequired("output")
	BundleCmd.Flags().StringArrayP("var", "v", []string{}, "Set template variables (format: key=value or template.tmpl:key=value)")
	BundleCmd.Flags().BoolP("interactive", "i", false, "Enable/disable interactive mode (default: auto-enabled when no --var or -f provided)")
	BundleCmd.Flags().StringP("values", "f", "", "Provide variables using a values.yaml file")
}

func generateBundle(bundleName, outputPath string, varsList []string, interactive bool, valuesFile string) error {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Resolve bundle name to directory path
	bundlePath, err := resolveBundleDirectory(cfg.TemplatesDir, bundleName)
	if err != nil {
		return err
	}

	// Load bundle metadata
	metadataPath := filepath.Join(bundlePath, "conjure.json")
	bundleMeta, err := metadata.LoadBundleMetadata(metadataPath)
	if err != nil {
		return fmt.Errorf("failed to load bundle metadata: %w", err)
	}

	fmt.Printf("Bundle: %s\n", bundleMeta.BundleName)
	fmt.Printf("Description: %s\n", bundleMeta.BundleDescription)
	fmt.Printf("Type: %s\n\n", bundleMeta.BundleType)

	// Start with an empty variables map
	userVariables := make(map[string]interface{})
	templateOverrides := make(map[string]map[string]interface{})

	// First, load values from values.yaml file if provided
	if valuesFile != "" {
		values, overrides, err := parseValuesWithOverrides(valuesFile)
		if err != nil {
			return fmt.Errorf("failed to parse values file: %w", err)
		}
		// Copy values from YAML into userVariables
		for k, v := range values {
			userVariables[k] = v
		}
		// Store template-specific overrides
		templateOverrides = overrides
	}

	// Then, parse variables from --var flags (these override values.yaml)
	cliVariables, cliTemplateOverrides, err := parseVariablesWithTemplateOverrides(varsList)
	if err != nil {
		return fmt.Errorf("failed to parse variables: %w", err)
	}
	// Merge CLI variables, overriding any values from YAML
	for k, v := range cliVariables {
		userVariables[k] = v
	}
	// Merge CLI template overrides (these override values.yaml template_overrides)
	for templateName, overrides := range cliTemplateOverrides {
		if templateOverrides[templateName] == nil {
			templateOverrides[templateName] = make(map[string]interface{})
		}
		for k, v := range overrides {
			templateOverrides[templateName][k] = v
		}
	}

	// If interactive mode, collect all variables and overrides
	if interactive {
		var interactiveOverrides map[string]map[string]interface{}
		userVariables, interactiveOverrides, err = prompt.CollectBundleVariables(bundleMeta, userVariables)
		if err != nil {
			return fmt.Errorf("failed to collect variables: %w", err)
		}
		// Merge interactive overrides with existing template overrides
		for templateName, overrides := range interactiveOverrides {
			if templateOverrides[templateName] == nil {
				templateOverrides[templateName] = make(map[string]interface{})
			}
			for k, v := range overrides {
				templateOverrides[templateName][k] = v
			}
		}
	}

	// Create output directory if it doesn't exist
	if err := os.MkdirAll(outputPath, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Read all template files in the bundle directory
	entries, err := os.ReadDir(bundlePath)
	if err != nil {
		return fmt.Errorf("failed to read bundle directory: %w", err)
	}

	templatesRendered := 0
	for _, entry := range entries {
		// Skip directories and non-.tmpl files
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".tmpl") {
			continue
		}

		templateFileName := entry.Name()
		templatePath := filepath.Join(bundlePath, templateFileName)

		// Read template file
		templateContent, err := os.ReadFile(templatePath)
		if err != nil {
			return fmt.Errorf("failed to read template file %s: %w", templateFileName, err)
		}

		// Start with user variables
		varsForTemplate := make(map[string]interface{})
		for k, v := range userVariables {
			varsForTemplate[k] = v
		}

		// Apply template-specific overrides from values.yaml
		if overrides, exists := templateOverrides[templateFileName]; exists {
			for k, v := range overrides {
				varsForTemplate[k] = v
			}
		}

		// Merge variables for this specific template
		templateVars, err := metadata.MergeVariablesForTemplate(bundleMeta, templateFileName, varsForTemplate)
		if err != nil {
			return fmt.Errorf("failed to merge variables for %s: %w", templateFileName, err)
		}

		// Render template
		rendered, err := renderTemplate(string(templateContent), templateVars)
		if err != nil {
			return fmt.Errorf("failed to render template %s: %w", templateFileName, err)
		}

		// Output file name (remove .tmpl extension)
		outputFileName := strings.TrimSuffix(templateFileName, ".tmpl")
		outputFilePath := filepath.Join(outputPath, outputFileName)

		// Write rendered content to output file
		if err := os.WriteFile(outputFilePath, []byte(rendered), 0644); err != nil {
			return fmt.Errorf("failed to write output file %s: %w", outputFileName, err)
		}

		fmt.Printf("  ✓ Rendered: %s\n", outputFileName)
		templatesRendered++
	}

	if templatesRendered == 0 {
		return fmt.Errorf("no templates found in bundle")
	}

	fmt.Printf("\n✓ Bundle rendered successfully\n")
	fmt.Printf("✓ %d files written to: %s\n", templatesRendered, outputPath)

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

// parseVariablesWithTemplateOverrides parses --var flags supporting both:
// - key=value (shared variable)
// - template.yaml.tmpl:key=value (template-specific override)
func parseVariablesWithTemplateOverrides(varsList []string) (map[string]interface{}, map[string]map[string]interface{}, error) {
	sharedVars := make(map[string]interface{})
	templateOverrides := make(map[string]map[string]interface{})

	for _, v := range varsList {
		// Split on '=' to get key=value
		parts := strings.SplitN(v, "=", 2)
		if len(parts) != 2 {
			return nil, nil, fmt.Errorf("invalid variable format '%s', expected key=value or template.tmpl:key=value", v)
		}

		keyPart := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		if keyPart == "" {
			return nil, nil, fmt.Errorf("variable key cannot be empty")
		}

		// Check if this is a template-specific override (contains ':')
		if strings.Contains(keyPart, ":") {
			// Format: template.yaml.tmpl:key
			templateParts := strings.SplitN(keyPart, ":", 2)
			if len(templateParts) != 2 {
				return nil, nil, fmt.Errorf("invalid template override format '%s', expected template.tmpl:key=value", v)
			}

			templateName := strings.TrimSpace(templateParts[0])
			varName := strings.TrimSpace(templateParts[1])

			if templateName == "" || varName == "" {
				return nil, nil, fmt.Errorf("template name and variable name cannot be empty in '%s'", v)
			}

			// Validate that varName doesn't contain additional colons
			if strings.Contains(varName, ":") {
				return nil, nil, fmt.Errorf("invalid template override format '%s', too many colons (expected template.tmpl:key=value)", v)
			}

			// Initialize map for this template if needed
			if templateOverrides[templateName] == nil {
				templateOverrides[templateName] = make(map[string]interface{})
			}

			templateOverrides[templateName][varName] = value
		} else {
			// Regular shared variable
			sharedVars[keyPart] = value
		}
	}

	return sharedVars, templateOverrides, nil
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

// parseValuesWithOverrides parses a values.yaml file and extracts template-specific overrides
// Returns: (shared variables, template overrides, error)
func parseValuesWithOverrides(valuesFile string) (map[string]interface{}, map[string]map[string]interface{}, error) {
	var allValues map[string]interface{}

	// Read the values.yaml file
	data, err := os.ReadFile(valuesFile)
	if err != nil {
		return nil, nil, fmt.Errorf("error reading values file: %w", err)
	}

	// Unmarshal the data
	if err := yaml.Unmarshal(data, &allValues); err != nil {
		return nil, nil, fmt.Errorf("error unmarshalling values yaml: %w", err)
	}

	// Separate shared variables from template overrides
	sharedVars := make(map[string]interface{})
	templateOverrides := make(map[string]map[string]interface{})

	for k, v := range allValues {
		// Check if this is the template_overrides section
		if k == "template_overrides" {
			// Parse template-specific overrides
			if overridesMap, ok := v.(map[string]interface{}); ok {
				for templateName, templateVars := range overridesMap {
					if varsMap, ok := templateVars.(map[string]interface{}); ok {
						templateOverrides[templateName] = varsMap
					}
				}
			}
		} else {
			// Regular shared variable
			sharedVars[k] = v
		}
	}

	return sharedVars, templateOverrides, nil
}

// resolveBundleDirectory resolves a bundle name to its directory path
// It scans all bundle directories and matches by bundle_name in metadata
func resolveBundleDirectory(templatesDir, bundleName string) (string, error) {
	bundlesPath := filepath.Join(templatesDir, "bundles")

	// Read all bundle directories
	entries, err := os.ReadDir(bundlesPath)
	if err != nil {
		return "", fmt.Errorf("failed to read bundles directory: %w", err)
	}

	// Scan each directory for matching bundle_name in metadata
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		// Try to load metadata for this bundle directory
		dirPath := filepath.Join(bundlesPath, entry.Name())
		metadataPath := filepath.Join(dirPath, "conjure.json")

		// Check if metadata exists
		data, err := os.ReadFile(metadataPath)
		if err != nil {
			continue // Skip directories without metadata
		}

		// Parse metadata (conjure.json is JSON format)
		var meta metadata.BundleMetadata
		if err := json.Unmarshal(data, &meta); err != nil {
			continue // Skip if can't parse
		}

		// Check if bundle_name matches
		if meta.BundleName == bundleName {
			return dirPath, nil
		}
	}

	return "", fmt.Errorf("bundle '%s' not found (searched by bundle_name in metadata)", bundleName)
}
