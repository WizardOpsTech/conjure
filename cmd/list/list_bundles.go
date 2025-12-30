package list

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/thesudoYT/conjure/internal/config"
)

type BundleMetadata struct {
	BundleType        string `json:"bundle_type"`
	BundleName        string `json:"bundle_name"`
	BundleDescription string `json:"bundle_description"`
}

// AllowedBundleTypes defines the permitted bundle types
var AllowedBundleTypes = []string{"kubernetes", "terraform"}

// isValidBundleType checks if a type is in the allowed list
func isValidBundleType(bundleType string) bool {
	for _, allowed := range AllowedBundleTypes {
		if bundleType == allowed {
			return true
		}
	}
	return false
}

var (
	// `conjure list bundles` command
	listBundlesCmd = &cobra.Command{
		Use:   "bundles",
		Short: "List bundles",
		Long: `List available bundles in the bundles_dir/bundles directory.
		
Bundles should be contained in a sub-directory dedicated to that bundle and contain a conjure.json metadata
file along with all of the templates required for that bundle.`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			// Validate --type flag if provided
			bundleType, _ := cmd.Flags().GetString("type")
			if bundleType != "" && !isValidBundleType(bundleType) {
				return fmt.Errorf("invalid bundle type '%s'. Allowed types: %v", bundleType, AllowedBundleTypes)
			}
			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			bundleType, _ := cmd.Flags().GetString("type")
			listBundles(bundleType)
		},
		PostRun: func(cmd *cobra.Command, args []string) {
		},
	}
)

func init() {
	// Add commands
	ListCmd.AddCommand(listBundlesCmd)

	// Add flags
	listBundlesCmd.Flags().StringP("type", "t", "", "Filter bundles by type (kubernetes, terraform)")
}

func listBundles(filterType string) {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		os.Exit(1)
	}

	// Build the path to bundles directory
	bundlesPath := filepath.Join(cfg.BundlesDir, "bundles")

	// Check if the directory exists
	if _, err := os.Stat(bundlesPath); os.IsNotExist(err) {
		fmt.Printf("Error: bundles directory does not exist at %s\n", bundlesPath)
		fmt.Printf("Please ensure %s/bundles exists\n", cfg.BundlesDir)
		os.Exit(1)
	}

	// Read directory contents
	entries, err := os.ReadDir(bundlesPath)
	if err != nil {
		fmt.Printf("Error reading bundles directory: %v\n", err)
		os.Exit(1)
	}

	// Print bundles
	if len(entries) == 0 {
		fmt.Println("No bundles found")
		return
	}

	if filterType != "" {
		fmt.Printf("Available bundles (type: %s):\n", filterType)
	} else {
		fmt.Println("Available bundles:")
	}
	fmt.Println()

	bundlesFound := 0
	for _, entry := range entries {
		// Only process directories
		if !entry.IsDir() {
			continue
		}

		// Build path to conjure.json
		bundleDir := filepath.Join(bundlesPath, entry.Name())
		metadataPath := filepath.Join(bundleDir, "conjure.json")

		// Read and parse conjure.json
		data, err := os.ReadFile(metadataPath)
		if err != nil {
			fmt.Printf("  [%s] - No metadata file found (conjure.json missing)\n", entry.Name())
			continue
		}

		var metadata BundleMetadata
		if err := json.Unmarshal(data, &metadata); err != nil {
			fmt.Printf("  [%s] - Invalid metadata file (failed to parse conjure.json)\n", entry.Name())
			continue
		}

		// Filter by type if specified
		if filterType != "" && metadata.BundleType != filterType {
			continue
		}

		// Display bundle metadata
		fmt.Printf("  Name: %s\n", metadata.BundleName)
		fmt.Printf("  Type: %s\n", metadata.BundleType)
		fmt.Printf("  Description: %s\n", metadata.BundleDescription)
		fmt.Println()
		bundlesFound++
	}

	if bundlesFound == 0 && filterType != "" {
		fmt.Printf("No bundles found with type '%s'\n", filterType)
	}
}
