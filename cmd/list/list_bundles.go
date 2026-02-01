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
	listBundlesCmd = &cobra.Command{
		Use:   "bundles",
		Short: "List bundles",
		Long: `List available bundles with version information.

Bundles are sourced from local directories or remote repositories based on
your configuration. By default, only the latest version is shown unless the
--versions flag is specified.

Examples:
  conjure list bundles
  conjure list bundles --versions
  conjure list bundles --type kubernetes`,
		Run: func(cmd *cobra.Command, args []string) {
			bundleType, _ := cmd.Flags().GetString("type")
			showAllVersions, _ := cmd.Flags().GetBool("versions")
			listBundles(bundleType, showAllVersions)
		},
	}
)

func init() {
	ListCmd.AddCommand(listBundlesCmd)

	listBundlesCmd.Flags().StringP("type", "t", "", "Filter bundles by type (kubernetes, terraform, etc.)")
	listBundlesCmd.Flags().Bool("versions", false, "Show all versions (default: latest only)")
}

func listBundles(filterType string, showAllVersions bool) {
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		os.Exit(1)
	}

	resolver, err := source.NewResolver(cfg, "bundle")
	if err != nil {
		fmt.Printf("Error creating resolver: %v\n", err)
		os.Exit(1)
	}

	bundles, err := resolver.ListBundles()
	if err != nil {
		fmt.Printf("Error listing bundles: %v\n", err)
		os.Exit(1)
	}

	if len(bundles) == 0 {
		fmt.Println("No bundles found")
		return
	}

	if filterType != "" {
		filtered := make([]source.BundleInfo, 0)
		for _, bundle := range bundles {
			if strings.EqualFold(bundle.Type, filterType) {
				filtered = append(filtered, bundle)
			}
		}
		bundles = filtered

		if len(bundles) == 0 {
			fmt.Printf("No bundles found with type '%s'\n", filterType)
			return
		}
	}

	sort.Slice(bundles, func(i, j int) bool {
		return bundles[i].Name < bundles[j].Name
	})

	if filterType != "" {
		fmt.Printf("Available Bundles (type: %s):\n\n", filterType)
	} else {
		fmt.Println("Available Bundles:")
		fmt.Println()
	}

	for _, bundle := range bundles {
		fmt.Printf("  %s\n", bundle.Name)
		if bundle.Description != "" {
			fmt.Printf("    Description: %s\n", bundle.Description)
		}
		fmt.Printf("    Type: %s\n", bundle.Type)

		if len(bundle.Versions) > 0 {
			latestVersion, err := version.FindLatest(bundle.Versions)
			if err == nil {
				fmt.Printf("    Latest: %s\n", latestVersion)
			}

			if showAllVersions {
				sortedVersions := make([]string, len(bundle.Versions))
				copy(sortedVersions, bundle.Versions)
				sort.Strings(sortedVersions)

				versionsStr := strings.Join(sortedVersions, ", ")
				fmt.Printf("    All Versions: %s\n", versionsStr)
			}
		}

		fmt.Println()
	}

	fmt.Printf("Total: %d bundle", len(bundles))
	if len(bundles) != 1 {
		fmt.Printf("s")
	}
	fmt.Println()
}
