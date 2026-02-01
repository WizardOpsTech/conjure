package source

import (
	"fmt"

	"github.com/wizardopstech/conjure/internal/config"
	"github.com/wizardopstech/conjure/internal/version"
)

type Resolver struct {
	sources       []Source
	sourceNames   []string
	priorityOrder []int
}

func NewResolver(cfg *config.Config, resourceType string) (*Resolver, error) {
	if resourceType != "template" && resourceType != "bundle" {
		return nil, fmt.Errorf("invalid resource type: %s (must be 'template' or 'bundle')", resourceType)
	}

	var sources []Source
	var sourceNames []string
	var priorityOrder []int

	var sourceType config.SourceType
	var localDir, remoteURL string
	var priority config.PriorityOrder

	if resourceType == "template" {
		sourceType = cfg.GetTemplatesSource()
		localDir = cfg.TemplatesLocalDir
		remoteURL = cfg.TemplatesRemoteURL
		priority = cfg.GetTemplatesPriority()
	} else {
		sourceType = cfg.GetBundlesSource()
		localDir = cfg.BundlesLocalDir
		remoteURL = cfg.BundlesRemoteURL
		priority = cfg.GetBundlesPriority()
	}

	switch sourceType {
	case config.SourceTypeLocal:
		localSource, err := NewLocalSource(localDir)
		if err != nil {
			return nil, fmt.Errorf("failed to create local source: %w", err)
		}
		sources = append(sources, localSource)
		sourceNames = append(sourceNames, "local")
		priorityOrder = []int{0}

	case config.SourceTypeRemote:
		remoteSource, err := NewRemoteSource(remoteURL, cfg.CacheDir)
		if err != nil {
			return nil, fmt.Errorf("failed to create remote source: %w", err)
		}
		sources = append(sources, remoteSource)
		sourceNames = append(sourceNames, "remote")
		priorityOrder = []int{0}

	case config.SourceTypeBoth:
		localSource, err := NewLocalSource(localDir)
		if err != nil {
			return nil, fmt.Errorf("failed to create local source: %w", err)
		}

		remoteSource, err := NewRemoteSource(remoteURL, cfg.CacheDir)
		if err != nil {
			return nil, fmt.Errorf("failed to create remote source: %w", err)
		}

		if priority == config.PriorityLocalFirst {
			sources = append(sources, localSource, remoteSource)
			sourceNames = append(sourceNames, "local", "remote")
			priorityOrder = []int{0, 1}
		} else {
			sources = append(sources, remoteSource, localSource)
			sourceNames = append(sourceNames, "remote", "local")
			priorityOrder = []int{0, 1}
		}

	default:
		return nil, fmt.Errorf("invalid source type: %s", sourceType)
	}

	return &Resolver{
		sources:       sources,
		sourceNames:   sourceNames,
		priorityOrder: priorityOrder,
	}, nil
}

func (r *Resolver) GetTemplate(name, requestedVersion string) (*TemplateContent, string, error) {
	var lastErr error

	for _, idx := range r.priorityOrder {
		source := r.sources[idx]
		sourceName := r.sourceNames[idx]

		content, err := source.GetTemplate(name, requestedVersion)
		if err == nil {
			return content, sourceName, nil
		}

		lastErr = err
	}

	if lastErr != nil {
		return nil, "", fmt.Errorf("template not found in any source: %w", lastErr)
	}

	return nil, "", fmt.Errorf("template '%s' not found", name)
}

func (r *Resolver) GetBundle(name, requestedVersion string) (*BundleContent, string, error) {
	var lastErr error

	for _, idx := range r.priorityOrder {
		source := r.sources[idx]
		sourceName := r.sourceNames[idx]

		content, err := source.GetBundle(name, requestedVersion)
		if err == nil {
			return content, sourceName, nil
		}

		lastErr = err
	}

	if lastErr != nil {
		return nil, "", fmt.Errorf("bundle not found in any source: %w", lastErr)
	}

	return nil, "", fmt.Errorf("bundle '%s' not found", name)
}

func (r *Resolver) ListTemplates() ([]TemplateInfo, error) {
	templateMap := make(map[string]TemplateInfo)

	for idx, source := range r.sources {
		sourceName := r.sourceNames[idx]

		templates, err := source.ListTemplates()
		if err != nil {
			fmt.Printf("Warning: failed to list templates from %s: %v\n", sourceName, err)
			continue
		}

		for _, tmpl := range templates {
			if existing, ok := templateMap[tmpl.Name]; ok {
				existing.Versions = mergeVersions(existing.Versions, tmpl.Versions)
				templateMap[tmpl.Name] = existing
			} else {
				templateMap[tmpl.Name] = tmpl
			}
		}
	}

	result := make([]TemplateInfo, 0, len(templateMap))
	for _, tmpl := range templateMap {
		result = append(result, tmpl)
	}

	return result, nil
}

func (r *Resolver) ListBundles() ([]BundleInfo, error) {
	bundleMap := make(map[string]BundleInfo)

	for idx, source := range r.sources {
		sourceName := r.sourceNames[idx]

		bundles, err := source.ListBundles()
		if err != nil {
			fmt.Printf("Warning: failed to list bundles from %s: %v\n", sourceName, err)
			continue
		}

		for _, bundle := range bundles {
			if existing, ok := bundleMap[bundle.Name]; ok {
				existing.Versions = mergeVersions(existing.Versions, bundle.Versions)
				bundleMap[bundle.Name] = existing
			} else {
				bundleMap[bundle.Name] = bundle
			}
		}
	}

	result := make([]BundleInfo, 0, len(bundleMap))
	for _, bundle := range bundleMap {
		result = append(result, bundle)
	}

	return result, nil
}

func (r *Resolver) GetLatestVersion(resourceType, name string) (string, error) {
	var allVersions []string

	if resourceType == "template" {
		for _, source := range r.sources {
			versions, err := source.GetTemplateVersions(name)
			if err != nil {
				continue
			}
			allVersions = append(allVersions, versions...)
		}
	} else if resourceType == "bundle" {
		for _, source := range r.sources {
			versions, err := source.GetBundleVersions(name)
			if err != nil {
				continue
			}
			allVersions = append(allVersions, versions...)
		}
	} else {
		return "", fmt.Errorf("invalid resource type: %s", resourceType)
	}

	if len(allVersions) == 0 {
		return "", fmt.Errorf("%s '%s' not found", resourceType, name)
	}

	allVersions = deduplicateVersions(allVersions)

	return version.FindLatest(allVersions)
}

func (r *Resolver) GetTemplateVersions(name string) ([]string, error) {
	var allVersions []string

	for _, source := range r.sources {
		versions, err := source.GetTemplateVersions(name)
		if err != nil {
			continue
		}
		allVersions = append(allVersions, versions...)
	}

	return deduplicateVersions(allVersions), nil
}

func (r *Resolver) GetBundleVersions(name string) ([]string, error) {
	var allVersions []string

	for _, source := range r.sources {
		versions, err := source.GetBundleVersions(name)
		if err != nil {
			continue
		}
		allVersions = append(allVersions, versions...)
	}

	return deduplicateVersions(allVersions), nil
}

func mergeVersions(a, b []string) []string {
	combined := append([]string{}, a...)
	combined = append(combined, b...)
	return deduplicateVersions(combined)
}

func deduplicateVersions(versions []string) []string {
	seen := make(map[string]bool)
	result := make([]string, 0, len(versions))

	for _, v := range versions {
		if !seen[v] {
			seen[v] = true
			result = append(result, v)
		}
	}

	return result
}
