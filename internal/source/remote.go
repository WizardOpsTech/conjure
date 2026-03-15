package source

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/wizardopstech/conjure/internal/security"
	"github.com/wizardopstech/conjure/internal/version"
)

const (
	DefaultHTTPTimeout = 30 * time.Second
)

type RemoteSource struct {
	baseURL      string
	client       *http.Client
	cache        *CacheManager
	verifier     *Verifier
	indexCache   *Index
	indexFetched bool
}

func NewRemoteSource(baseURL string, cacheDir string) (*RemoteSource, error) {
	if err := security.ValidateURL(baseURL); err != nil {
		return nil, fmt.Errorf("invalid base URL: %w", err)
	}

	cache, err := NewCacheManager(cacheDir)
	if err != nil {
		return nil, fmt.Errorf("failed to create cache manager: %w", err)
	}

	client := &http.Client{
		Timeout: DefaultHTTPTimeout,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				MinVersion: tls.VersionTLS12,
			},
		},
	}

	return &RemoteSource{
		baseURL:  strings.TrimRight(baseURL, "/"),
		client:   client,
		cache:    cache,
		verifier: NewVerifier(),
	}, nil
}

func (r *RemoteSource) fetchIndex(skipCache bool) error {
	if r.indexFetched && r.indexCache != nil {
		return nil
	}

	if !skipCache {
		cachedIndex, err := r.cache.LoadIndex(r.baseURL)
		if err == nil {
			r.indexCache = cachedIndex
			r.indexFetched = true
			return nil
		}
	}

	indexURL := r.baseURL + "/index.json"
	resp, err := r.client.Get(indexURL)
	if err != nil {
		return fmt.Errorf("failed to fetch index: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to fetch index: HTTP %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read index response: %w", err)
	}

	var index Index
	if err := json.Unmarshal(data, &index); err != nil {
		return fmt.Errorf("failed to parse index JSON: %w", err)
	}

	if index.SchemaVersion != "v1" {
		return fmt.Errorf("unsupported index schema version: %s", index.SchemaVersion)
	}

	if err := r.cache.SaveIndex(r.baseURL, &index); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to cache index: %v\n", err)
	}

	r.indexCache = &index
	r.indexFetched = true

	return nil
}

func (r *RemoteSource) ListTemplates() ([]TemplateInfo, error) {
	if err := r.fetchIndex(true); err != nil {
		return nil, err
	}

	templates := make([]TemplateInfo, 0, len(r.indexCache.Templates))
	for _, entry := range r.indexCache.Templates {
		versions := make([]string, 0, len(entry.Versions))
		for _, v := range entry.Versions {
			versions = append(versions, v.Version)
		}

		templates = append(templates, TemplateInfo{
			Name:        entry.Name,
			Type:        entry.Type,
			Description: entry.Description,
			Versions:    versions,
		})
	}

	return templates, nil
}

func (r *RemoteSource) ListBundles() ([]BundleInfo, error) {
	if err := r.fetchIndex(true); err != nil {
		return nil, err
	}

	bundles := make([]BundleInfo, 0, len(r.indexCache.Bundles))
	for _, entry := range r.indexCache.Bundles {
		versions := make([]string, 0, len(entry.Versions))
		for _, v := range entry.Versions {
			versions = append(versions, v.Version)
		}

		bundles = append(bundles, BundleInfo{
			Name:        entry.Name,
			Type:        entry.Type,
			Description: entry.Description,
			Versions:    versions,
		})
	}

	return bundles, nil
}

func (r *RemoteSource) GetTemplate(name, requestedVersion string) (*TemplateContent, error) {
	if err := security.ValidateVariableName(name); err != nil {
		return nil, fmt.Errorf("invalid template name: %w", err)
	}

	if err := r.fetchIndex(false); err != nil {
		return nil, err
	}

	var templateEntry *TemplateIndexEntry
	for i := range r.indexCache.Templates {
		if r.indexCache.Templates[i].Name == name {
			templateEntry = &r.indexCache.Templates[i]
			break
		}
	}

	if templateEntry == nil {
		return nil, fmt.Errorf("template '%s' not found", name)
	}

	targetVersion := requestedVersion
	if targetVersion == "" {
		versions := make([]string, 0, len(templateEntry.Versions))
		for _, v := range templateEntry.Versions {
			versions = append(versions, v.Version)
		}
		var err error
		targetVersion, err = version.FindLatest(versions)
		if err != nil {
			return nil, fmt.Errorf("failed to find latest version: %w", err)
		}
	} else {
		if err := version.ValidateVersion(targetVersion); err != nil {
			return nil, fmt.Errorf("invalid version: %w", err)
		}
	}

	var versionEntry *VersionEntry
	for i := range templateEntry.Versions {
		if templateEntry.Versions[i].Version == targetVersion {
			versionEntry = &templateEntry.Versions[i]
			break
		}
	}

	if versionEntry == nil {
		return nil, fmt.Errorf("version '%s' not found for template '%s'", targetVersion, name)
	}

	if r.cache.HasTemplate(name, targetVersion) {
		if r.verifyCachedTemplate(name, targetVersion, versionEntry) {
			localSource, err := NewLocalSource(r.cache.GetCacheDir())
			if err != nil {
				return nil, fmt.Errorf("failed to create local source for cache: %w", err)
			}
			return localSource.GetTemplate(name, targetVersion)
		}
	}

	var metadataRaw []byte
	var templateRaw []byte

	for _, file := range versionEntry.Files {
		fileURL := r.baseURL + "/" + file.Path

		tempFile, err := r.downloadFile(fileURL, file.Size, file.SHA256)
		if err != nil {
			return nil, fmt.Errorf("failed to download file '%s': %w", file.Name, err)
		}
		defer func(f string) { _ = os.Remove(f) }(tempFile)

		content, err := os.ReadFile(tempFile)
		if err != nil {
			return nil, fmt.Errorf("failed to read downloaded file: %w", err)
		}

		if file.Name == "conjure.json" {
			metadataRaw = content
		} else {
			templateRaw = content
		}
	}

	if metadataRaw == nil {
		return nil, fmt.Errorf("metadata file not found in template")
	}

	if templateRaw == nil {
		return nil, fmt.Errorf("template file not found in template")
	}

	if err := r.cache.SaveTemplate(name, targetVersion, metadataRaw, templateRaw); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to cache template: %v\n", err)
	}

	return &TemplateContent{
		Info: TemplateInfo{
			Name:        name,
			Type:        templateEntry.Type,
			Description: templateEntry.Description,
			Versions:    []string{targetVersion},
		},
		Version:     targetVersion,
		MetadataRaw: metadataRaw,
		TemplateRaw: templateRaw,
	}, nil
}

func (r *RemoteSource) GetBundle(name, requestedVersion string) (*BundleContent, error) {
	if err := security.ValidateVariableName(name); err != nil {
		return nil, fmt.Errorf("invalid bundle name: %w", err)
	}

	if err := r.fetchIndex(false); err != nil {
		return nil, fmt.Errorf("failed to fetch bundle index: %w", err)
	}

	var bundleEntry *BundleIndexEntry
	for i := range r.indexCache.Bundles {
		if r.indexCache.Bundles[i].Name == name {
			bundleEntry = &r.indexCache.Bundles[i]
			break
		}
	}

	if bundleEntry == nil {
		return nil, fmt.Errorf("bundle '%s' not found", name)
	}

	targetVersion := requestedVersion
	if targetVersion == "" {
		versions := make([]string, 0, len(bundleEntry.Versions))
		for _, v := range bundleEntry.Versions {
			versions = append(versions, v.Version)
		}
		var err error
		targetVersion, err = version.FindLatest(versions)
		if err != nil {
			return nil, fmt.Errorf("failed to find latest version: %w", err)
		}
	} else {
		if err := version.ValidateVersion(targetVersion); err != nil {
			return nil, fmt.Errorf("invalid version: %w", err)
		}
	}

	var versionEntry *VersionEntry
	for i := range bundleEntry.Versions {
		if bundleEntry.Versions[i].Version == targetVersion {
			versionEntry = &bundleEntry.Versions[i]
			break
		}
	}

	if versionEntry == nil {
		return nil, fmt.Errorf("version '%s' not found for bundle '%s'", targetVersion, name)
	}

	if r.cache.HasBundle(name, targetVersion) {
		if r.verifyCachedBundle(name, targetVersion, versionEntry) {
			localSource, err := NewLocalSource(r.cache.GetCacheDir())
			if err != nil {
				return nil, fmt.Errorf("failed to create local source for cache: %w", err)
			}
			return localSource.GetBundle(name, targetVersion)
		}
	}

	var metadataRaw []byte
	files := make(map[string][]byte)

	for _, file := range versionEntry.Files {
		fileURL := r.baseURL + "/" + file.Path

		tempFile, err := r.downloadFile(fileURL, file.Size, file.SHA256)
		if err != nil {
			return nil, fmt.Errorf("failed to download file '%s': %w", file.Name, err)
		}
		defer func(f string) { _ = os.Remove(f) }(tempFile)

		content, err := os.ReadFile(tempFile)
		if err != nil {
			return nil, fmt.Errorf("failed to read downloaded file: %w", err)
		}

		if file.Name == "conjure.json" {
			metadataRaw = content
		} else {
			files[file.Name] = content
		}
	}

	if metadataRaw == nil {
		return nil, fmt.Errorf("metadata file not found in bundle")
	}

	if err := r.cache.SaveBundle(name, targetVersion, metadataRaw, files); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to cache bundle: %v\n", err)
	}

	return &BundleContent{
		Info: BundleInfo{
			Name:        name,
			Type:        bundleEntry.Type,
			Description: bundleEntry.Description,
			Versions:    []string{targetVersion},
		},
		Version:     targetVersion,
		MetadataRaw: metadataRaw,
		Files:       files,
	}, nil
}

func (r *RemoteSource) GetTemplateVersions(name string) ([]string, error) {
	if err := security.ValidateVariableName(name); err != nil {
		return nil, fmt.Errorf("invalid template name: %w", err)
	}

	if err := r.fetchIndex(false); err != nil {
		return nil, err
	}

	for _, entry := range r.indexCache.Templates {
		if entry.Name == name {
			versions := make([]string, 0, len(entry.Versions))
			for _, v := range entry.Versions {
				versions = append(versions, v.Version)
			}
			return versions, nil
		}
	}

	return []string{}, nil
}

func (r *RemoteSource) GetBundleVersions(name string) ([]string, error) {
	if err := security.ValidateVariableName(name); err != nil {
		return nil, fmt.Errorf("invalid bundle name: %w", err)
	}

	if err := r.fetchIndex(false); err != nil {
		return nil, err
	}

	for _, entry := range r.indexCache.Bundles {
		if entry.Name == name {
			versions := make([]string, 0, len(entry.Versions))
			for _, v := range entry.Versions {
				versions = append(versions, v.Version)
			}
			return versions, nil
		}
	}

	return []string{}, nil
}

func (r *RemoteSource) downloadFile(url string, expectedSize int64, expectedHash string) (string, error) {
	if err := r.verifier.ValidateFileSize(expectedSize); err != nil {
		return "", fmt.Errorf("file size validation failed: %w", err)
	}

	resp, err := r.client.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to download: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to download: HTTP %d", resp.StatusCode)
	}

	tempFile, err := os.CreateTemp("", "conjure-download-*")
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}
	tempPath := tempFile.Name()

	written, err := io.Copy(tempFile, resp.Body)
	_ = tempFile.Close()

	if err != nil {
		_ = os.Remove(tempPath)
		return "", fmt.Errorf("failed to write file: %w", err)
	}

	if written != expectedSize {
		_ = os.Remove(tempPath)
		return "", fmt.Errorf("file size mismatch: expected %d bytes, got %d bytes. This likely means index.json was generated with different line endings (Windows CRLF vs Linux LF) than the files on the remote repository", expectedSize, written)
	}

	if expectedHash != "" {
		if err := r.verifier.VerifySHA256(tempPath, expectedHash); err != nil {
			_ = os.Remove(tempPath)
			return "", fmt.Errorf("%w. This means the file content on the remote repository doesn't match what was used to generate index.json", err)
		}
	}

	if err := os.Chmod(tempPath, cacheFilePerm); err != nil {
		_ = os.Remove(tempPath)
		return "", fmt.Errorf("failed to set file permissions: %w", err)
	}

	return tempPath, nil
}

func (r *RemoteSource) verifyCachedTemplate(name, version string, versionEntry *VersionEntry) bool {
	for _, file := range versionEntry.Files {
		cachedPath := r.cache.GetTemplatePath(name, version, file.Name)

		if _, err := os.Stat(cachedPath); err != nil {
			return false
		}

		if file.SHA256 != "" {
			if err := r.verifier.VerifySHA256(cachedPath, file.SHA256); err != nil {
				return false
			}
		}
	}
	return true
}

func (r *RemoteSource) verifyCachedBundle(name, version string, versionEntry *VersionEntry) bool {
	for _, file := range versionEntry.Files {
		cachedPath := r.cache.GetBundlePath(name, version, file.Name)

		if _, err := os.Stat(cachedPath); err != nil {
			return false
		}

		if file.SHA256 != "" {
			if err := r.verifier.VerifySHA256(cachedPath, file.SHA256); err != nil {
				return false
			}
		}
	}
	return true
}

func (r *RemoteSource) RefreshIndex() error {
	r.indexFetched = false
	r.indexCache = nil
	return r.fetchIndex(true)
}
