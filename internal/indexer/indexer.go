package indexer

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/wizardopstech/conjure/internal/source"
)

type Indexer struct {
	verifier *source.Verifier
}

func NewIndexer() *Indexer {
	return &Indexer{
		verifier: source.NewVerifier(),
	}
}

func (i *Indexer) BuildIndex(templatesDir, bundlesDir string) (*source.Index, error) {
	index := &source.Index{
		SchemaVersion: "v1",
		LastUpdated:   time.Now(),
		Templates:     make([]source.TemplateIndexEntry, 0),
		Bundles:       make([]source.BundleIndexEntry, 0),
	}

	if templatesDir != "" {
		if _, err := os.Stat(templatesDir); err == nil {
			templates, err := i.buildTemplatesIndex(templatesDir)
			if err != nil {
				return nil, fmt.Errorf("failed to build templates index: %w", err)
			}
			index.Templates = templates
		}
	}

	if bundlesDir != "" {
		if _, err := os.Stat(bundlesDir); err == nil {
			bundles, err := i.buildBundlesIndex(bundlesDir)
			if err != nil {
				return nil, fmt.Errorf("failed to build bundles index: %w", err)
			}
			index.Bundles = bundles
		}
	}

	return index, nil
}

func (i *Indexer) buildTemplatesIndex(templatesDir string) ([]source.TemplateIndexEntry, error) {
	localSource, err := source.NewLocalSource(filepath.Dir(templatesDir))
	if err != nil {
		return nil, fmt.Errorf("failed to create local source: %w", err)
	}

	templates, err := localSource.ListTemplates()
	if err != nil {
		return nil, fmt.Errorf("failed to list templates: %w", err)
	}

	entries := make([]source.TemplateIndexEntry, 0, len(templates))
	for _, tmpl := range templates {
		entry := source.TemplateIndexEntry{
			Name:        tmpl.Name,
			Type:        tmpl.Type,
			Description: tmpl.Description,
			Versions:    make([]source.VersionEntry, 0, len(tmpl.Versions)),
		}

		for _, ver := range tmpl.Versions {
			versionEntry, err := i.buildVersionEntry(templatesDir, tmpl.Name, ver)
			if err != nil {
				return nil, fmt.Errorf("failed to build version entry for template %s version %s: %w", tmpl.Name, ver, err)
			}
			entry.Versions = append(entry.Versions, versionEntry)
		}

		entries = append(entries, entry)
	}

	return entries, nil
}

func (i *Indexer) buildBundlesIndex(bundlesDir string) ([]source.BundleIndexEntry, error) {
	localSource, err := source.NewLocalSource(filepath.Dir(bundlesDir))
	if err != nil {
		return nil, fmt.Errorf("failed to create local source: %w", err)
	}

	bundles, err := localSource.ListBundles()
	if err != nil {
		return nil, fmt.Errorf("failed to list bundles: %w", err)
	}

	entries := make([]source.BundleIndexEntry, 0, len(bundles))
	for _, bundle := range bundles {
		entry := source.BundleIndexEntry{
			Name:        bundle.Name,
			Type:        bundle.Type,
			Description: bundle.Description,
			Versions:    make([]source.VersionEntry, 0, len(bundle.Versions)),
		}

		for _, ver := range bundle.Versions {
			versionEntry, err := i.buildVersionEntry(bundlesDir, bundle.Name, ver)
			if err != nil {
				return nil, fmt.Errorf("failed to build version entry for bundle %s version %s: %w", bundle.Name, ver, err)
			}
			entry.Versions = append(entry.Versions, versionEntry)
		}

		entries = append(entries, entry)
	}

	return entries, nil
}

func (i *Indexer) buildVersionEntry(baseDir, name, version string) (source.VersionEntry, error) {
	versionDir := filepath.Join(baseDir, name, version)

	entries, err := os.ReadDir(versionDir)
	if err != nil {
		return source.VersionEntry{}, fmt.Errorf("failed to read version directory: %w", err)
	}

	files := make([]source.FileEntry, 0)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		filePath := filepath.Join(versionDir, entry.Name())
		fileInfo, err := os.Stat(filePath)
		if err != nil {
			return source.VersionEntry{}, fmt.Errorf("failed to stat file %s: %w", entry.Name(), err)
		}

		hash, err := i.verifier.ComputeSHA256(filePath)
		if err != nil {
			return source.VersionEntry{}, fmt.Errorf("failed to compute hash for file %s: %w", entry.Name(), err)
		}

		relativePath := filepath.Join(filepath.Base(baseDir), name, version, entry.Name())

		files = append(files, source.FileEntry{
			Name:   entry.Name(),
			Path:   filepath.ToSlash(relativePath),
			Size:   fileInfo.Size(),
			SHA256: hash,
		})
	}

	return source.VersionEntry{
		Version: version,
		Files:   files,
	}, nil
}

func (i *Indexer) WriteIndex(index *source.Index, outputPath string) error {
	outputDir := filepath.Dir(outputPath)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	data, err := json.MarshalIndent(index, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal index: %w", err)
	}

	if err := os.WriteFile(outputPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write index file: %w", err)
	}

	return nil
}

func (i *Indexer) ValidateStructure(dir, resourceType string) error {
	if resourceType != "templates" && resourceType != "bundles" {
		return fmt.Errorf("invalid resource type: %s (must be 'templates' or 'bundles')", resourceType)
	}

	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return fmt.Errorf("directory does not exist: %s", dir)
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("failed to read directory: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		resourceName := entry.Name()
		resourceDir := filepath.Join(dir, resourceName)

		versionEntries, err := os.ReadDir(resourceDir)
		if err != nil {
			return fmt.Errorf("failed to read resource directory %s: %w", resourceName, err)
		}

		hasValidVersion := false
		for _, versionEntry := range versionEntries {
			if !versionEntry.IsDir() {
				continue
			}

			versionDir := filepath.Join(resourceDir, versionEntry.Name())

			metadataPath := filepath.Join(versionDir, "conjure.json")
			if _, err := os.Stat(metadataPath); os.IsNotExist(err) {
				return fmt.Errorf("missing conjure.json in %s/%s/%s", resourceType, resourceName, versionEntry.Name())
			}

			hasValidVersion = true
		}

		if !hasValidVersion {
			return fmt.Errorf("no valid versions found for %s/%s", resourceType, resourceName)
		}
	}

	return nil
}
