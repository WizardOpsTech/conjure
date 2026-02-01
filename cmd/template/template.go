package template

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/wizardopstech/conjure/internal/config"
	"github.com/wizardopstech/conjure/internal/metadata"
	"github.com/wizardopstech/conjure/internal/prompt"
	"github.com/wizardopstech/conjure/internal/render"
	"github.com/wizardopstech/conjure/internal/source"
	"go.yaml.in/yaml/v3"
)

var TemplateCmd = &cobra.Command{
	Use:   "template <template-name>",
	Short: "Generate a file from a template",
	Long: `Generate a file from a template with variable substitution.

Templates are versioned and can be sourced from local directories or remote repositories.
By default, the latest version is used unless specified with --version.

Examples:
  conjure template deployment -o ./deployment.yaml
  conjure template deployment --version 1.0.0 -o ./deployment.yaml
  conjure template deployment -o ./deployment.yaml -f values.yaml
  conjure template vsphere_vm -o ./main.tf --var vm_name=myvm --var cpu=4`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		templateName := args[0]
		outputPath, _ := cmd.Flags().GetString("output")
		vars, _ := cmd.Flags().GetStringArray("var")
		valuesFile, _ := cmd.Flags().GetString("values")
		templateVersion, _ := cmd.Flags().GetString("version")

		// Determine if interactive mode should be enabled
		// If -i flag was explicitly set, use that value
		// Otherwise, enable interactive if no --var or -f flags provided
		// I need to remove the -i flag its not needed antmore.
		interactive := cmd.Flags().Changed("interactive")
		if interactive {
			interactive, _ = cmd.Flags().GetBool("interactive")
		} else {
			interactive = len(vars) == 0 && valuesFile == ""
		}

		if err := generateTemplate(templateName, templateVersion, outputPath, vars, interactive, valuesFile); err != nil {
			fmt.Printf("Error generating template: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	TemplateCmd.Flags().StringP("output", "o", "", "Output file path (required)")
	_ = TemplateCmd.MarkFlagRequired("output")
	TemplateCmd.Flags().StringArrayP("var", "v", []string{}, "Set template variables (format: key=value)")
	TemplateCmd.Flags().BoolP("interactive", "i", false, "Enable/disable interactive mode (default: auto-enabled when no --var or -f provided)")
	TemplateCmd.Flags().StringP("values", "f", "", "Provide variables using a values.yaml file")
	TemplateCmd.Flags().String("version", "", "Template version (default: latest)")
}

func generateTemplate(templateName, templateVersion, outputPath string, varsList []string, interactive bool, valuesFile string) error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	resolver, err := source.NewResolver(cfg, "template")
	if err != nil {
		return fmt.Errorf("failed to create resolver: %w", err)
	}

	if templateVersion == "" || templateVersion == "latest" {
		resolvedVersion, err := resolver.GetLatestVersion("template", templateName)
		if err != nil {
			return err
		}
		templateVersion = resolvedVersion
	}

	content, sourceName, err := resolver.GetTemplate(templateName, templateVersion)
	if err != nil {
		return err
	}

	fmt.Printf("Using template: %s v%s (source: %s)\n", templateName, content.Version, sourceName)

	userVariables := make(map[string]interface{})

	if valuesFile != "" {
		values, err := parseValues(valuesFile)
		if err != nil {
			return fmt.Errorf("failed to parse values file: %w", err)
		}
		for k, v := range values {
			userVariables[k] = v
		}
	}

	cliVariables, err := render.ParseVariables(varsList)
	if err != nil {
		return fmt.Errorf("failed to parse variables: %w", err)
	}
	for k, v := range cliVariables {
		userVariables[k] = v
	}

	var meta *metadata.TemplateMetadata
	var finalVariables map[string]interface{}

	if content.MetadataRaw != nil {
		var err error
		meta, err = metadata.ParseTemplateMetadata(content.MetadataRaw)
		if err != nil {
			return fmt.Errorf("failed to parse metadata: %w", err)
		}

		fmt.Printf("Using metadata: %s\n\n", meta.TemplateDescription)

		if interactive {
			finalVariables, err = prompt.CollectVariables(meta, userVariables)
			if err != nil {
				return fmt.Errorf("failed to collect variables: %w", err)
			}
		} else {
			finalVariables, err = metadata.ValidateAndMergeVariables(meta, userVariables)
			if err != nil {
				return fmt.Errorf("variable validation failed: %w", err)
			}
		}

		fmt.Println("\nVariables:")
		for _, v := range meta.Variables {
			if val, exists := finalVariables[v.Name]; exists {
				fmt.Printf("  %s = %v\n", v.Name, val)
			}
		}
		fmt.Println()
	} else {
		fmt.Println("Warning: No metadata file found, using variables without validation")
		finalVariables = userVariables
	}

	rendered, err := render.RenderTemplate(string(content.TemplateRaw), finalVariables)
	if err != nil {
		return fmt.Errorf("failed to render template: %w", err)
	}

	outputDir := filepath.Dir(outputPath)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	if err := os.WriteFile(outputPath, []byte(rendered), 0644); err != nil {
		return fmt.Errorf("failed to write output file: %w", err)
	}

	fmt.Printf("✓ Template rendered successfully\n")
	fmt.Printf("✓ Output written to: %s\n", outputPath)

	return nil
}

func parseValues(valuesFile string) (map[string]interface{}, error) {
	var values map[string]interface{}

	data, err := os.ReadFile(valuesFile)
	if err != nil {
		return nil, fmt.Errorf("error reading values file: %w", err)
	}

	if err := yaml.Unmarshal(data, &values); err != nil {
		return nil, fmt.Errorf("error unmarshalling values yaml: %w", err)
	}

	return values, nil
}
