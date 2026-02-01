package repo

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/wizardopstech/conjure/internal/indexer"
)

var (
	templatesDir string
	bundlesDir   string
	outputDir    string
)

var indexCmd = &cobra.Command{
	Use:   "index",
	Short: "Generate repository index files",
	Long: `Generate index files for a template and bundle repository.

This command scans your templates and bundles directories, validates their
structure, computes file hashes, and generates index.json files that can be
used to host a remote repository.

Example:
  conjure repo index --templates ./templates --bundles ./bundles --out ./repo

The command will create index.json in the output directory with metadata
about all templates and bundles, including version information and file hashes.`,
	RunE: runIndex,
}

func init() {
	indexCmd.Flags().StringVar(&templatesDir, "templates", "", "Path to templates directory")
	indexCmd.Flags().StringVar(&bundlesDir, "bundles", "", "Path to bundles directory")
	indexCmd.Flags().StringVarP(&outputDir, "out", "o", ".", "Output directory for index file")
}

func runIndex(cmd *cobra.Command, args []string) error {
	if templatesDir == "" && bundlesDir == "" {
		return fmt.Errorf("at least one of --templates or --bundles must be specified")
	}

	fmt.Println("Indexing repository...")
	fmt.Println()

	idx := indexer.NewIndexer()

	if templatesDir != "" {
		fmt.Printf("Validating templates directory: %s\n", templatesDir)
		if err := idx.ValidateStructure(templatesDir, "templates"); err != nil {
			return fmt.Errorf("templates validation failed: %w", err)
		}
	}

	if bundlesDir != "" {
		fmt.Printf("Validating bundles directory: %s\n", bundlesDir)
		if err := idx.ValidateStructure(bundlesDir, "bundles"); err != nil {
			return fmt.Errorf("bundles validation failed: %w", err)
		}
	}

	fmt.Println()

	// Build index
	index, err := idx.BuildIndex(templatesDir, bundlesDir)
	if err != nil {
		return fmt.Errorf("failed to build index: %w", err)
	}

	fmt.Println("Index Summary:")
	fmt.Println()

	if len(index.Templates) > 0 {
		fmt.Println("Templates:")
		totalTemplateVersions := 0
		for _, tmpl := range index.Templates {
			versionCount := len(tmpl.Versions)
			totalTemplateVersions += versionCount

			versions := ""
			for i, v := range tmpl.Versions {
				if i > 0 {
					versions += ", "
				}
				versions += v.Version
			}

			fmt.Printf("  ✓ %s (%d version", tmpl.Name, versionCount)
			if versionCount != 1 {
				fmt.Printf("s")
			}
			fmt.Printf(": %s)\n", versions)
		}
		fmt.Println()
	}

	if len(index.Bundles) > 0 {
		fmt.Println("Bundles:")
		totalBundleVersions := 0
		for _, bundle := range index.Bundles {
			versionCount := len(bundle.Versions)
			totalBundleVersions += versionCount

			versions := ""
			for i, v := range bundle.Versions {
				if i > 0 {
					versions += ", "
				}
				versions += v.Version
			}

			fmt.Printf("  ✓ %s (%d version", bundle.Name, versionCount)
			if versionCount != 1 {
				fmt.Printf("s")
			}
			fmt.Printf(": %s)\n", versions)
		}
		fmt.Println()
	}

	outputPath := filepath.Join(outputDir, "index.json")
	if err := idx.WriteIndex(index, outputPath); err != nil {
		return fmt.Errorf("failed to write index: %w", err)
	}

	fmt.Println("Index file created:")
	absPath, err := filepath.Abs(outputPath)
	if err != nil {
		absPath = outputPath
	}
	fmt.Printf("  %s\n", absPath)

	fileInfo, err := os.Stat(outputPath)
	if err == nil {
		fmt.Printf("  Size: %d bytes\n", fileInfo.Size())
	}

	totalResources := len(index.Templates) + len(index.Bundles)
	totalVersions := 0
	for _, tmpl := range index.Templates {
		totalVersions += len(tmpl.Versions)
	}
	for _, bundle := range index.Bundles {
		totalVersions += len(bundle.Versions)
	}

	fmt.Printf("  Resources: %d (%d template", totalResources, len(index.Templates))
	if len(index.Templates) != 1 {
		fmt.Printf("s")
	}
	fmt.Printf(", %d bundle", len(index.Bundles))
	if len(index.Bundles) != 1 {
		fmt.Printf("s")
	}
	fmt.Printf(")\n")
	fmt.Printf("  Versions: %d\n", totalVersions)

	fmt.Println()
	fmt.Println("✓ Repository index created successfully")

	return nil
}
