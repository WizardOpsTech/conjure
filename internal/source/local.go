package source

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/wizardopstech/conjure/internal/metadata"
	"github.com/wizardopstech/conjure/internal/security"
	"github.com/wizardopstech/conjure/internal/version"
)

type LocalSource struct {
	baseDir string
}

func NewLocalSource(baseDir string) (*LocalSource, error) {
	if err := security.ValidatePathSafety(baseDir); err != nil {
		return nil, fmt.Errorf("invalid base directory: %w", err)
	}

	return &LocalSource{
		baseDir: baseDir,
	}, nil
}

func (l *LocalSource) ListTemplates() ([]TemplateInfo, error) {
	templatesDir := filepath.Join(l.baseDir, "templates")

	if _, err := os.Stat(templatesDir); os.IsNotExist(err) {
		return []TemplateInfo{}, nil
	}

	entries, err := os.ReadDir(templatesDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read templates directory: %w", err)
	}

	templates := make([]TemplateInfo, 0)
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		templateName := entry.Name()
		versions, err := l.GetTemplateVersions(templateName)
		if err != nil {
			continue
		}

		if len(versions) == 0 {
			continue
		}

		latestVersion, err := version.FindLatest(versions)
		if err != nil {
			continue
		}

		metadataPath := filepath.Join(templatesDir, templateName, latestVersion, "conjure.json")
		meta, err := metadata.LoadTemplateMetadata(metadataPath)
		if err != nil {
			continue
		}

		templates = append(templates, TemplateInfo{
			Name:        templateName,
			Type:        meta.TemplateType,
			Description: meta.TemplateDescription,
			Versions:    versions,
		})
	}

	return templates, nil
}

func (l *LocalSource) ListBundles() ([]BundleInfo, error) {
	bundlesDir := filepath.Join(l.baseDir, "bundles")

	if _, err := os.Stat(bundlesDir); os.IsNotExist(err) {
		return []BundleInfo{}, nil
	}

	entries, err := os.ReadDir(bundlesDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read bundles directory: %w", err)
	}

	bundles := make([]BundleInfo, 0)
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		bundleName := entry.Name()
		versions, err := l.GetBundleVersions(bundleName)
		if err != nil {
			continue
		}

		if len(versions) == 0 {
			continue
		}

		latestVersion, err := version.FindLatest(versions)
		if err != nil {
			continue
		}

		metadataPath := filepath.Join(bundlesDir, bundleName, latestVersion, "conjure.json")
		meta, err := metadata.LoadBundleMetadata(metadataPath)
		if err != nil {
			continue
		}

		bundles = append(bundles, BundleInfo{
			Name:        bundleName,
			Type:        meta.BundleType,
			Description: meta.BundleDescription,
			Versions:    versions,
		})
	}

	return bundles, nil
}

func (l *LocalSource) GetTemplate(name, requestedVersion string) (*TemplateContent, error) {
	if err := security.ValidateVariableName(name); err != nil {
		return nil, fmt.Errorf("invalid template name: %w", err)
	}

	versions, err := l.GetTemplateVersions(name)
	if err != nil {
		return nil, err
	}

	if len(versions) == 0 {
		return nil, fmt.Errorf("template '%s' not found", name)
	}

	targetVersion := requestedVersion
	if targetVersion == "" {
		targetVersion, err = version.FindLatest(versions)
		if err != nil {
			return nil, fmt.Errorf("failed to find latest version: %w", err)
		}
	} else {
		if err := version.ValidateVersion(targetVersion); err != nil {
			return nil, fmt.Errorf("invalid version: %w", err)
		}
		found := false
		for _, v := range versions {
			if v == targetVersion {
				found = true
				break
			}
		}
		if !found {
			return nil, fmt.Errorf("version '%s' not found for template '%s'", targetVersion, name)
		}
	}

	templateDir := filepath.Join(l.baseDir, "templates", name, targetVersion)
	metadataPath := filepath.Join(templateDir, "conjure.json")
	meta, err := metadata.LoadTemplateMetadata(metadataPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load metadata: %w", err)
	}

	metadataRaw, err := os.ReadFile(metadataPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read metadata file: %w", err)
	}

	templateFile, err := findTemplateFile(templateDir)
	if err != nil {
		return nil, err
	}

	templateRaw, err := os.ReadFile(templateFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read template file: %w", err)
	}

	return &TemplateContent{
		Info: TemplateInfo{
			Name:        name,
			Type:        meta.TemplateType,
			Description: meta.TemplateDescription,
			Versions:    versions,
		},
		Version:     targetVersion,
		MetadataRaw: metadataRaw,
		TemplateRaw: templateRaw,
	}, nil
}

func (l *LocalSource) GetBundle(name, requestedVersion string) (*BundleContent, error) {
	if err := security.ValidateVariableName(name); err != nil {
		return nil, fmt.Errorf("invalid bundle name: %w", err)
	}

	versions, err := l.GetBundleVersions(name)
	if err != nil {
		return nil, err
	}

	if len(versions) == 0 {
		return nil, fmt.Errorf("bundle '%s' not found", name)
	}

	targetVersion := requestedVersion
	if targetVersion == "" {
		targetVersion, err = version.FindLatest(versions)
		if err != nil {
			return nil, fmt.Errorf("failed to find latest version: %w", err)
		}
	} else {
		if err := version.ValidateVersion(targetVersion); err != nil {
			return nil, fmt.Errorf("invalid version: %w", err)
		}
		found := false
		for _, v := range versions {
			if v == targetVersion {
				found = true
				break
			}
		}
		if !found {
			return nil, fmt.Errorf("version '%s' not found for bundle '%s'", targetVersion, name)
		}
	}

	bundleDir := filepath.Join(l.baseDir, "bundles", name, targetVersion)
	metadataPath := filepath.Join(bundleDir, "conjure.json")
	meta, err := metadata.LoadBundleMetadata(metadataPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load metadata: %w", err)
	}

	metadataRaw, err := os.ReadFile(metadataPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read metadata file: %w", err)
	}

	files := make(map[string][]byte)
	entries, err := os.ReadDir(bundleDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read bundle directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		filename := entry.Name()
		if filename == "conjure.json" {
			continue
		}

		filePath := filepath.Join(bundleDir, filename)
		content, err := os.ReadFile(filePath)
		if err != nil {
			return nil, fmt.Errorf("failed to read file '%s': %w", filename, err)
		}

		files[filename] = content
	}

	return &BundleContent{
		Info: BundleInfo{
			Name:        name,
			Type:        meta.BundleType,
			Description: meta.BundleDescription,
			Versions:    versions,
		},
		Version:     targetVersion,
		MetadataRaw: metadataRaw,
		Files:       files,
	}, nil
}

func (l *LocalSource) GetTemplateVersions(name string) ([]string, error) {
	if err := security.ValidateVariableName(name); err != nil {
		return nil, fmt.Errorf("invalid template name: %w", err)
	}

	templateDir := filepath.Join(l.baseDir, "templates", name)

	if _, err := os.Stat(templateDir); os.IsNotExist(err) {
		return []string{}, nil
	}

	entries, err := os.ReadDir(templateDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read template directory: %w", err)
	}

	versions := make([]string, 0)
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		versionStr := entry.Name()
		if err := version.ValidateVersion(versionStr); err != nil {
			continue
		}

		metadataPath := filepath.Join(templateDir, versionStr, "conjure.json")
		if _, err := os.Stat(metadataPath); os.IsNotExist(err) {
			continue
		}

		versions = append(versions, versionStr)
	}

	sort.Strings(versions)

	return versions, nil
}

func (l *LocalSource) GetBundleVersions(name string) ([]string, error) {
	if err := security.ValidateVariableName(name); err != nil {
		return nil, fmt.Errorf("invalid bundle name: %w", err)
	}

	bundleDir := filepath.Join(l.baseDir, "bundles", name)

	if _, err := os.Stat(bundleDir); os.IsNotExist(err) {
		return []string{}, nil
	}

	entries, err := os.ReadDir(bundleDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read bundle directory: %w", err)
	}

	versions := make([]string, 0)
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		versionStr := entry.Name()
		if err := version.ValidateVersion(versionStr); err != nil {
			continue
		}

		metadataPath := filepath.Join(bundleDir, versionStr, "conjure.json")
		if _, err := os.Stat(metadataPath); os.IsNotExist(err) {
			continue
		}

		versions = append(versions, versionStr)
	}

	sort.Strings(versions)

	return versions, nil
}

func findTemplateFile(dir string) (string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return "", fmt.Errorf("failed to read directory: %w", err)
	}

	templateExtensions := []string{".tmpl", ".tpl", ".template"}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		filename := entry.Name()
		if filename == "conjure.json" {
			continue
		}

		for _, ext := range templateExtensions {
			if filepath.Ext(filename) == ext {
				return filepath.Join(dir, filename), nil
			}
		}
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		filename := entry.Name()
		if filename != "conjure.json" {
			return filepath.Join(dir, filename), nil
		}
	}

	return "", fmt.Errorf("no template file found in directory")
}
