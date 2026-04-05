package bundle

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/wizardopstech/conjure/internal/config"
	"github.com/wizardopstech/conjure/internal/metadata"
	"github.com/wizardopstech/conjure/internal/prompt"
	"github.com/wizardopstech/conjure/internal/render"
	"github.com/wizardopstech/conjure/internal/source"
	"go.yaml.in/yaml/v3"
)

var BundleCmd = &cobra.Command{
	Use:   "bundle <bundle-name>",
	Short: "Generate files from a bundle of templates",
	Long: `Generate multiple files from a bundle of templates with variable substitution.

Bundles are versioned collections of related templates that are rendered together.
Each bundle has a conjure.json metadata file that defines shared and template-specific variables.
By default, the latest version is used unless specified with --version.

Variable Overrides:
  - Shared variables apply to all templates
  - Template-specific overrides using: template.tmpl:key=value
  - Use template_overrides section in values.yaml
  - Interactive mode prompts for overrides after collecting main variables

Examples:
  conjure bundle kubernetes-deployment -o ./k8s/
  conjure bundle kubernetes-deployment --version 1.0.0 -o ./k8s/
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
		bundleVersion, _ := cmd.Flags().GetString("version")

		interactive := cmd.Flags().Changed("interactive")
		if interactive {
			interactive, _ = cmd.Flags().GetBool("interactive")
		} else {
			interactive = len(vars) == 0 && valuesFile == ""
		}

		if err := generateBundle(bundleName, bundleVersion, outputPath, vars, interactive, valuesFile); err != nil {
			fmt.Printf("Error generating bundle: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	BundleCmd.Flags().StringP("output", "o", "", "Output directory path (required)")
	_ = BundleCmd.MarkFlagRequired("output")
	BundleCmd.Flags().StringArrayP("var", "v", []string{}, "Set template variables (format: key=value or template.tmpl:key=value)")
	BundleCmd.Flags().BoolP("interactive", "i", false, "Enable/disable interactive mode (default: auto-enabled when no --var or -f provided)")
	BundleCmd.Flags().StringP("values", "f", "", "Provide variables using a values.yaml file")
	BundleCmd.Flags().String("version", "", "Bundle version (default: latest)")
}

func generateBundle(bundleName, bundleVersion, outputPath string, varsList []string, interactive bool, valuesFile string) error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	resolver, err := source.NewResolver(cfg, "bundle")
	if err != nil {
		return fmt.Errorf("failed to create resolver: %w", err)
	}

	// Resolve version if not specified or if "latest" is explicitly requested
	if bundleVersion == "" || bundleVersion == "latest" {
		resolvedVersion, err := resolver.GetLatestVersion("bundle", bundleName)
		if err != nil {
			versions, verErr := resolver.GetBundleVersions(bundleName)
			if verErr == nil && len(versions) > 0 {
				versionsStr := strings.Join(versions, ", ")
				return fmt.Errorf("bundle '%s' not found or has no valid versions. Available versions: %s", bundleName, versionsStr)
			}
			return fmt.Errorf("bundle '%s' not found: %w", bundleName, err)
		}
		bundleVersion = resolvedVersion
	}

	// Get bundle from resolver
	content, sourceName, err := resolver.GetBundle(bundleName, bundleVersion)
	if err != nil {
		return err
	}

	// Parse metadata
	bundleMeta, err := metadata.ParseBundleMetadata(content.MetadataRaw)
	if err != nil {
		return fmt.Errorf("failed to parse bundle metadata: %w", err)
	}

	fmt.Printf("Using bundle: %s v%s (source: %s)\n", bundleName, content.Version, sourceName)
	fmt.Printf("Description: %s\n", bundleMeta.BundleDescription)
	fmt.Printf("Type: %s\n\n", bundleMeta.BundleType)

	userVariables := make(map[string]interface{})
	templateOverrides := make(map[string]map[string]interface{})

	if valuesFile != "" {
		values, overrides, err := parseValuesWithOverrides(valuesFile)
		if err != nil {
			return fmt.Errorf("failed to parse values file: %w", err)
		}
		for k, v := range values {
			userVariables[k] = v
		}
		templateOverrides = overrides
	}

	cliVariables, cliTemplateOverrides, err := parseVariablesWithTemplateOverrides(varsList)
	if err != nil {
		return fmt.Errorf("failed to parse variables: %w", err)
	}
	for k, v := range cliVariables {
		userVariables[k] = v
	}

	for templateName, overrides := range cliTemplateOverrides {
		if templateOverrides[templateName] == nil {
			templateOverrides[templateName] = make(map[string]interface{})
		}
		for k, v := range overrides {
			templateOverrides[templateName][k] = v
		}
	}

	if interactive {
		var interactiveOverrides map[string]map[string]interface{}
		userVariables, interactiveOverrides, err = prompt.CollectBundleVariables(bundleMeta, userVariables, cfg.ColorTheme)
		if err != nil {
			return fmt.Errorf("failed to collect variables: %w", err)
		}
		for templateName, overrides := range interactiveOverrides {
			if templateOverrides[templateName] == nil {
				templateOverrides[templateName] = make(map[string]interface{})
			}
			for k, v := range overrides {
				templateOverrides[templateName][k] = v
			}
		}
	}

	if err := os.MkdirAll(outputPath, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	templatesRendered := 0
	for templateFileName, templateContentBytes := range content.Files {
		if !strings.HasSuffix(templateFileName, ".tmpl") {
			continue
		}

		varsForTemplate := make(map[string]interface{})
		for k, v := range userVariables {
			varsForTemplate[k] = v
		}

		if overrides, exists := templateOverrides[templateFileName]; exists {
			for k, v := range overrides {
				varsForTemplate[k] = v
			}
		}

		templateVars, err := metadata.MergeVariablesForTemplate(bundleMeta, templateFileName, varsForTemplate)
		if err != nil {
			return fmt.Errorf("failed to merge variables for %s: %w", templateFileName, err)
		}

		rendered, err := render.RenderTemplate(string(templateContentBytes), templateVars)
		if err != nil {
			return fmt.Errorf("failed to render template %s: %w", templateFileName, err)
		}

		outputFileName := strings.TrimSuffix(templateFileName, ".tmpl")
		outputFilePath := filepath.Join(outputPath, outputFileName)

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

		if strings.Contains(keyPart, ":") {
			templateParts := strings.SplitN(keyPart, ":", 2)
			if len(templateParts) != 2 {
				return nil, nil, fmt.Errorf("invalid template override format '%s', expected template.tmpl:key=value", v)
			}

			templateName := strings.TrimSpace(templateParts[0])
			varName := strings.TrimSpace(templateParts[1])

			if templateName == "" || varName == "" {
				return nil, nil, fmt.Errorf("template name and variable name cannot be empty in '%s'", v)
			}

			if strings.Contains(varName, ":") {
				return nil, nil, fmt.Errorf("invalid template override format '%s', too many colons (expected template.tmpl:key=value)", v)
			}

			if templateOverrides[templateName] == nil {
				templateOverrides[templateName] = make(map[string]interface{})
			}

			templateOverrides[templateName][varName] = value
		} else {
			sharedVars[keyPart] = value
		}
	}

	return sharedVars, templateOverrides, nil
}

// parseValuesWithOverrides parses a values.yaml file and extracts template-specific overrides
func parseValuesWithOverrides(valuesFile string) (map[string]interface{}, map[string]map[string]interface{}, error) {
	var allValues map[string]interface{}

	data, err := os.ReadFile(valuesFile)
	if err != nil {
		return nil, nil, fmt.Errorf("error reading values file: %w", err)
	}

	if err := yaml.Unmarshal(data, &allValues); err != nil {
		return nil, nil, fmt.Errorf("error unmarshalling values yaml: %w", err)
	}

	sharedVars := make(map[string]interface{})
	templateOverrides := make(map[string]map[string]interface{})

	for k, v := range allValues {
		if k == "template_overrides" {
			if overridesMap, ok := v.(map[string]interface{}); ok {
				for templateName, templateVars := range overridesMap {
					if varsMap, ok := templateVars.(map[string]interface{}); ok {
						templateOverrides[templateName] = varsMap
					}
				}
			}
		} else {
			sharedVars[k] = v
		}
	}

	return sharedVars, templateOverrides, nil
}
