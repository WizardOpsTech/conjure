package source

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/wizardopstech/conjure/internal/security"
)

const (
	cacheDirPerm  = 0700
	cacheFilePerm = 0600
)

type CacheManager struct {
	cacheDir string
	locks    map[string]*sync.Mutex
	locksMux sync.Mutex
}

func NewCacheManager(cacheDir string) (*CacheManager, error) {
	if err := security.ValidatePathSafety(cacheDir); err != nil {
		return nil, fmt.Errorf("invalid cache directory: %w", err)
	}

	if len(cacheDir) > 0 && cacheDir[0] == '~' {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get home directory: %w", err)
		}
		cacheDir = filepath.Join(homeDir, cacheDir[1:])
	}

	if !filepath.IsAbs(cacheDir) {
		absPath, err := filepath.Abs(cacheDir)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve absolute path: %w", err)
		}
		cacheDir = absPath
	}

	cm := &CacheManager{
		cacheDir: cacheDir,
		locks:    make(map[string]*sync.Mutex),
	}

	if err := cm.initCacheDir(); err != nil {
		return nil, err
	}

	return cm, nil
}

func (c *CacheManager) initCacheDir() error {
	if err := os.MkdirAll(c.cacheDir, cacheDirPerm); err != nil {
		return fmt.Errorf("failed to create cache directory: %w", err)
	}

	dirs := []string{
		filepath.Join(c.cacheDir, "templates"),
		filepath.Join(c.cacheDir, "bundles"),
		filepath.Join(c.cacheDir, "indexes"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, cacheDirPerm); err != nil {
			return fmt.Errorf("failed to create cache subdirectory: %w", err)
		}
	}

	return nil
}

func (c *CacheManager) getLock(path string) *sync.Mutex {
	c.locksMux.Lock()
	defer c.locksMux.Unlock()

	if lock, exists := c.locks[path]; exists {
		return lock
	}

	lock := &sync.Mutex{}
	c.locks[path] = lock
	return lock
}

func (c *CacheManager) SaveTemplate(name, version string, metadataRaw, templateRaw []byte) error {
	if err := security.ValidateVariableName(name); err != nil {
		return fmt.Errorf("invalid template name: %w", err)
	}

	templateDir := filepath.Join(c.cacheDir, "templates", name, version)
	lock := c.getLock(templateDir)
	lock.Lock()
	defer lock.Unlock()

	if err := os.MkdirAll(templateDir, cacheDirPerm); err != nil {
		return fmt.Errorf("failed to create template directory: %w", err)
	}

	metadataPath := filepath.Join(templateDir, "conjure.json")
	if err := os.WriteFile(metadataPath, metadataRaw, cacheFilePerm); err != nil {
		return fmt.Errorf("failed to write metadata file: %w", err)
	}

	templatePath := filepath.Join(templateDir, "template.tmpl")
	if err := os.WriteFile(templatePath, templateRaw, cacheFilePerm); err != nil {
		return fmt.Errorf("failed to write template file: %w", err)
	}

	return nil
}

func (c *CacheManager) SaveBundle(name, version string, metadataRaw []byte, files map[string][]byte) error {
	if err := security.ValidateVariableName(name); err != nil {
		return fmt.Errorf("invalid bundle name: %w", err)
	}

	bundleDir := filepath.Join(c.cacheDir, "bundles", name, version)
	lock := c.getLock(bundleDir)
	lock.Lock()
	defer lock.Unlock()

	if err := os.MkdirAll(bundleDir, cacheDirPerm); err != nil {
		return fmt.Errorf("failed to create bundle directory: %w", err)
	}

	metadataPath := filepath.Join(bundleDir, "conjure.json")
	if err := os.WriteFile(metadataPath, metadataRaw, cacheFilePerm); err != nil {
		return fmt.Errorf("failed to write metadata file: %w", err)
	}

	for filename, content := range files {
		filePath := filepath.Join(bundleDir, filename)
		if err := os.WriteFile(filePath, content, cacheFilePerm); err != nil {
			return fmt.Errorf("failed to write file '%s': %w", filename, err)
		}
	}

	return nil
}

func (c *CacheManager) SaveIndex(remoteURL string, index *Index) error {
	if err := security.ValidateURL(remoteURL); err != nil {
		return fmt.Errorf("invalid remote URL: %w", err)
	}

	indexPath := c.getIndexPath(remoteURL)
	lock := c.getLock(indexPath)
	lock.Lock()
	defer lock.Unlock()

	data, err := json.MarshalIndent(index, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal index: %w", err)
	}

	if err := os.WriteFile(indexPath, data, cacheFilePerm); err != nil {
		return fmt.Errorf("failed to write index file: %w", err)
	}

	return nil
}

func (c *CacheManager) LoadIndex(remoteURL string) (*Index, error) {
	if err := security.ValidateURL(remoteURL); err != nil {
		return nil, fmt.Errorf("invalid remote URL: %w", err)
	}

	indexPath := c.getIndexPath(remoteURL)
	lock := c.getLock(indexPath)
	lock.Lock()
	defer lock.Unlock()

	data, err := os.ReadFile(indexPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("index not found in cache")
		}
		return nil, fmt.Errorf("failed to read index file: %w", err)
	}

	var index Index
	if err := json.Unmarshal(data, &index); err != nil {
		return nil, fmt.Errorf("failed to unmarshal index: %w", err)
	}

	return &index, nil
}

func (c *CacheManager) HasTemplate(name, version string) bool {
	if err := security.ValidateVariableName(name); err != nil {
		return false
	}

	templateDir := filepath.Join(c.cacheDir, "templates", name, version)
	metadataPath := filepath.Join(templateDir, "conjure.json")

	_, err := os.Stat(metadataPath)
	return err == nil
}

func (c *CacheManager) HasBundle(name, version string) bool {
	if err := security.ValidateVariableName(name); err != nil {
		return false
	}

	bundleDir := filepath.Join(c.cacheDir, "bundles", name, version)
	metadataPath := filepath.Join(bundleDir, "conjure.json")

	_, err := os.Stat(metadataPath)
	return err == nil
}

func (c *CacheManager) GetTemplatePath(name, version, filename string) string {
	return filepath.Join(c.cacheDir, "templates", name, version, filename)
}

func (c *CacheManager) GetBundlePath(name, version, filename string) string {
	return filepath.Join(c.cacheDir, "bundles", name, version, filename)
}

func (c *CacheManager) GetCacheDir() string {
	return c.cacheDir
}

func (c *CacheManager) ClearCache() error {
	templatesDir := filepath.Join(c.cacheDir, "templates")
	if err := os.RemoveAll(templatesDir); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove templates cache: %w", err)
	}

	bundlesDir := filepath.Join(c.cacheDir, "bundles")
	if err := os.RemoveAll(bundlesDir); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove bundles cache: %w", err)
	}

	indexesDir := filepath.Join(c.cacheDir, "indexes")
	if err := os.RemoveAll(indexesDir); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove indexes cache: %w", err)
	}

	return c.initCacheDir()
}

func (c *CacheManager) ClearTemplateCache() error {
	templatesDir := filepath.Join(c.cacheDir, "templates")
	if err := os.RemoveAll(templatesDir); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove templates cache: %w", err)
	}

	return os.MkdirAll(templatesDir, cacheDirPerm)
}

func (c *CacheManager) ClearBundleCache() error {
	bundlesDir := filepath.Join(c.cacheDir, "bundles")
	if err := os.RemoveAll(bundlesDir); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove bundles cache: %w", err)
	}

	return os.MkdirAll(bundlesDir, cacheDirPerm)
}

func (c *CacheManager) getIndexPath(remoteURL string) string {
	safeName := filepath.Base(remoteURL)
	if safeName == "" || safeName == "." || safeName == "/" {
		safeName = "default"
	}

	return filepath.Join(c.cacheDir, "indexes", safeName+".json")
}
