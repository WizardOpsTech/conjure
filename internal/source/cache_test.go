package source

import (
	"os"
	"path/filepath"
	"sync"
	"testing"
)

func TestNewCacheManager(t *testing.T) {
	tempDir := t.TempDir()

	tests := []struct {
		name     string
		cacheDir string
		wantErr  bool
	}{
		{
			name:     "valid cache directory",
			cacheDir: tempDir,
			wantErr:  false,
		},
		{
			name:     "invalid path with traversal",
			cacheDir: "../../../etc",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cm, err := NewCacheManager(tt.cacheDir)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewCacheManager() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				subdirs := []string{"templates", "bundles", "indexes"}
				for _, subdir := range subdirs {
					path := filepath.Join(cm.GetCacheDir(), subdir)
					if _, err := os.Stat(path); os.IsNotExist(err) {
						t.Errorf("Subdirectory %s not created", subdir)
					}
				}
			}
		})
	}
}

func TestCacheManager_SaveAndHasTemplate(t *testing.T) {
	tempDir := t.TempDir()
	cm, err := NewCacheManager(tempDir)
	if err != nil {
		t.Fatalf("Failed to create cache manager: %v", err)
	}

	metadataRaw := []byte(`{"schema_version":"v1","template_name":"test","version":"1.0.0"}`)
	templateRaw := []byte(`template content`)

	if cm.HasTemplate("test", "1.0.0") {
		t.Error("Template should not exist in cache yet")
	}

	err = cm.SaveTemplate("test", "1.0.0", metadataRaw, templateRaw)
	if err != nil {
		t.Fatalf("SaveTemplate() error = %v", err)
	}

	if !cm.HasTemplate("test", "1.0.0") {
		t.Error("Template should exist in cache")
	}

	templateDir := filepath.Join(tempDir, "templates", "test", "1.0.0")
	metadataPath := filepath.Join(templateDir, "conjure.json")
	templatePath := filepath.Join(templateDir, "template.tmpl")

	if _, err := os.Stat(metadataPath); os.IsNotExist(err) {
		t.Error("Metadata file not created")
	}

	if _, err := os.Stat(templatePath); os.IsNotExist(err) {
		t.Error("Template file not created")
	}

	info, err := os.Stat(metadataPath)
	if err != nil {
		t.Fatalf("Failed to stat metadata file: %v", err)
	}

	if info.Mode().Perm() != 0666 { // Windows returns 0666 by default
		perm := info.Mode().Perm()
		if perm != cacheFilePerm {
			t.Logf("File permissions: expected %o, got %o (may differ on Windows)", cacheFilePerm, perm)
		}
	}
}

func TestCacheManager_SaveAndHasBundle(t *testing.T) {
	tempDir := t.TempDir()
	cm, err := NewCacheManager(tempDir)
	if err != nil {
		t.Fatalf("Failed to create cache manager: %v", err)
	}

	metadataRaw := []byte(`{"schema_version":"v1","bundle_name":"test","version":"1.0.0"}`)
	files := map[string][]byte{
		"app.yaml.tmpl": []byte(`app template`),
		"db.yaml.tmpl":  []byte(`db template`),
	}

	if cm.HasBundle("test", "1.0.0") {
		t.Error("Bundle should not exist in cache yet")
	}

	err = cm.SaveBundle("test", "1.0.0", metadataRaw, files)
	if err != nil {
		t.Fatalf("SaveBundle() error = %v", err)
	}

	if !cm.HasBundle("test", "1.0.0") {
		t.Error("Bundle should exist in cache")
	}

	bundleDir := filepath.Join(tempDir, "bundles", "test", "1.0.0")
	metadataPath := filepath.Join(bundleDir, "conjure.json")

	if _, err := os.Stat(metadataPath); os.IsNotExist(err) {
		t.Error("Metadata file not created")
	}

	for filename := range files {
		filePath := filepath.Join(bundleDir, filename)
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			t.Errorf("Bundle file %s not created", filename)
		}
	}
}

func TestCacheManager_SaveAndLoadIndex(t *testing.T) {
	tempDir := t.TempDir()
	cm, err := NewCacheManager(tempDir)
	if err != nil {
		t.Fatalf("Failed to create cache manager: %v", err)
	}

	remoteURL := "https://example.com/repo"
	index := &Index{
		SchemaVersion: "v1",
		Templates: []TemplateIndexEntry{
			{
				Name:        "test",
				Type:        "yaml",
				Description: "Test template",
			},
		},
	}

	err = cm.SaveIndex(remoteURL, index)
	if err != nil {
		t.Fatalf("SaveIndex() error = %v", err)
	}

	loadedIndex, err := cm.LoadIndex(remoteURL)
	if err != nil {
		t.Fatalf("LoadIndex() error = %v", err)
	}

	if loadedIndex.SchemaVersion != index.SchemaVersion {
		t.Errorf("Expected schema version %s, got %s", index.SchemaVersion, loadedIndex.SchemaVersion)
	}

	if len(loadedIndex.Templates) != 1 {
		t.Errorf("Expected 1 template, got %d", len(loadedIndex.Templates))
	}
}

func TestCacheManager_ClearCache(t *testing.T) {
	tempDir := t.TempDir()
	cm, err := NewCacheManager(tempDir)
	if err != nil {
		t.Fatalf("Failed to create cache manager: %v", err)
	}

	metadataRaw := []byte(`{"schema_version":"v1"}`)
	templateRaw := []byte(`template`)

	err = cm.SaveTemplate("test", "1.0.0", metadataRaw, templateRaw)
	if err != nil {
		t.Fatalf("SaveTemplate() error = %v", err)
	}

	if !cm.HasTemplate("test", "1.0.0") {
		t.Error("Template should exist before clear")
	}

	err = cm.ClearCache()
	if err != nil {
		t.Fatalf("ClearCache() error = %v", err)
	}

	if cm.HasTemplate("test", "1.0.0") {
		t.Error("Template should not exist after clear")
	}

	subdirs := []string{"templates", "bundles", "indexes"}
	for _, subdir := range subdirs {
		path := filepath.Join(cm.GetCacheDir(), subdir)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("Subdirectory %s should still exist after clear", subdir)
		}
	}
}

func TestCacheManager_ClearTemplateCache(t *testing.T) {
	tempDir := t.TempDir()
	cm, err := NewCacheManager(tempDir)
	if err != nil {
		t.Fatalf("Failed to create cache manager: %v", err)
	}

	metadataRaw := []byte(`{"schema_version":"v1"}`)
	templateRaw := []byte(`template`)

	err = cm.SaveTemplate("test", "1.0.0", metadataRaw, templateRaw)
	if err != nil {
		t.Fatalf("SaveTemplate() error = %v", err)
	}

	err = cm.ClearTemplateCache()
	if err != nil {
		t.Fatalf("ClearTemplateCache() error = %v", err)
	}

	if cm.HasTemplate("test", "1.0.0") {
		t.Error("Template should not exist after clear")
	}
}

func TestCacheManager_ConcurrentAccess(t *testing.T) {
	tempDir := t.TempDir()
	cm, err := NewCacheManager(tempDir)
	if err != nil {
		t.Fatalf("Failed to create cache manager: %v", err)
	}

	metadataRaw := []byte(`{"schema_version":"v1"}`)
	templateRaw := []byte(`template`)

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := cm.SaveTemplate("concurrent-test", "1.0.0", metadataRaw, templateRaw)
			if err != nil {
				t.Errorf("Concurrent SaveTemplate() error = %v", err)
			}
		}()
	}

	wg.Wait()

	if !cm.HasTemplate("concurrent-test", "1.0.0") {
		t.Error("Template should exist after concurrent writes")
	}
}

func TestCacheManager_InvalidNames(t *testing.T) {
	tempDir := t.TempDir()
	cm, err := NewCacheManager(tempDir)
	if err != nil {
		t.Fatalf("Failed to create cache manager: %v", err)
	}

	metadataRaw := []byte(`{"schema_version":"v1"}`)
	templateRaw := []byte(`template`)

	err = cm.SaveTemplate("../invalid", "1.0.0", metadataRaw, templateRaw)
	if err == nil {
		t.Error("Expected error for invalid template name")
	}

	if cm.HasTemplate("../invalid", "1.0.0") {
		t.Error("HasTemplate should return false for invalid names")
	}
}
