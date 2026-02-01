package list

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"github.com/wizardopstech/conjure/internal/config"
	"github.com/wizardopstech/conjure/internal/source"
	"github.com/wizardopstech/conjure/internal/version"
)

var (
	listTemplatesCmd = &cobra.Command{
		Use:   "templates",
		Short: "List templates",
		Long: `List available templates with version information.

Templates are sourced from local directories or remote repositories based on
your configuration. By default, only the latest version is shown unless the
--versions flag is specified.

Examples:
  conjure list templates
  conjure list templates --versions
  conjure list templates --type yaml`,
		Run: func(cmd *cobra.Command, args []string) {
			templateType, _ := cmd.Flags().GetString("type")
			showAllVersions, _ := cmd.Flags().GetBool("versions")
			listTemplates(templateType, showAllVersions)
		},
	}
)

func init() {
	ListCmd.AddCommand(listTemplatesCmd)

	listTemplatesCmd.Flags().StringP("type", "t", "", "Filter templates by type (yaml, json, etc.)")
	listTemplatesCmd.Flags().Bool("versions", false, "Show all versions (default: latest only)")
}

func listTemplates(filterType string, showAllVersions bool) {
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		os.Exit(1)
	}

	resolver, err := source.NewResolver(cfg, "template")
	if err != nil {
		fmt.Printf("Error creating resolver: %v\n", err)
		os.Exit(1)
	}

	templates, err := resolver.ListTemplates()
	if err != nil {
		fmt.Printf("Error listing templates: %v\n", err)
		os.Exit(1)
	}

	if len(templates) == 0 {
		fmt.Println("No templates found")
		return
	}

	if filterType != "" {
		filtered := make([]source.TemplateInfo, 0)
		for _, tmpl := range templates {
			if strings.EqualFold(tmpl.Type, filterType) {
				filtered = append(filtered, tmpl)
			}
		}
		templates = filtered

		if len(templates) == 0 {
			fmt.Printf("No templates found with type '%s'\n", filterType)
			return
		}
	}

	sort.Slice(templates, func(i, j int) bool {
		return templates[i].Name < templates[j].Name
	})

	if filterType != "" {
		fmt.Printf("Available Templates (type: %s):\n\n", filterType)
	} else {
		fmt.Println("Available Templates:")
		fmt.Println()
	}

	for _, tmpl := range templates {
		fmt.Printf("  %s\n", tmpl.Name)
		if tmpl.Description != "" {
			fmt.Printf("    Description: %s\n", tmpl.Description)
		}
		fmt.Printf("    Type: %s\n", tmpl.Type)

		if len(tmpl.Versions) > 0 {
			latestVersion, err := version.FindLatest(tmpl.Versions)
			if err == nil {
				fmt.Printf("    Latest: %s\n", latestVersion)
			}

			if showAllVersions {
				sortedVersions := make([]string, len(tmpl.Versions))
				copy(sortedVersions, tmpl.Versions)
				sort.Strings(sortedVersions)

				versionsStr := strings.Join(sortedVersions, ", ")
				fmt.Printf("    All Versions: %s\n", versionsStr)
			}
		}

		fmt.Println()
	}

	fmt.Printf("Total: %d template", len(templates))
	if len(templates) != 1 {
		fmt.Printf("s")
	}
	fmt.Println()
}
