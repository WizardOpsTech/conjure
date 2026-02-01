package indexer

import (
	"os"
	"path/filepath"
	"testing"
)

func TestIndexer_BuildIndex(t *testing.T) {
	testDir := t.TempDir()
	templatesDir := filepath.Join(testDir, "templates")
	bundlesDir := filepath.Join(testDir, "bundles")

	templateDir := filepath.Join(templatesDir, "test-template", "1.0.0")
	if err := os.MkdirAll(templateDir, 0755); err != nil {
		t.Fatal(err)
	}
	templateMetadata := `{
		"schema_version": "v1",
		"template_name": "test-template",
		"template_type": "yaml",
		"template_description": "Test template",
		"version": "1.0.0",
		"variables": []
	}`
	if err := os.WriteFile(filepath.Join(templateDir, "conjure.json"), []byte(templateMetadata), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(templateDir, "template.tmpl"), []byte("test content"), 0600); err != nil {
		t.Fatal(err)
	}

	bundleDir := filepath.Join(bundlesDir, "test-bundle", "1.0.0")
	if err := os.MkdirAll(bundleDir, 0755); err != nil {
		t.Fatal(err)
	}
	bundleMetadata := `{
		"schema_version": "v1",
		"bundle_name": "test-bundle",
		"bundle_type": "kubernetes",
		"bundle_description": "Test bundle",
		"version": "1.0.0",
		"templates": []
	}`
	if err := os.WriteFile(filepath.Join(bundleDir, "conjure.json"), []byte(bundleMetadata), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(bundleDir, "app.yaml"), []byte("test bundle content"), 0600); err != nil {
		t.Fatal(err)
	}

	indexer := NewIndexer()
	index, err := indexer.BuildIndex(templatesDir, bundlesDir)
	if err != nil {
		t.Fatalf("BuildIndex() error: %v", err)
	}

	if index.SchemaVersion != "v1" {
		t.Errorf("SchemaVersion = %s, want v1", index.SchemaVersion)
	}

	if len(index.Templates) != 1 {
		t.Errorf("Templates count = %d, want 1", len(index.Templates))
	}

	if len(index.Bundles) != 1 {
		t.Errorf("Bundles count = %d, want 1", len(index.Bundles))
	}

	if len(index.Templates) > 0 {
		tmpl := index.Templates[0]
		if tmpl.Name != "test-template" {
			t.Errorf("Template name = %s, want test-template", tmpl.Name)
		}
		if len(tmpl.Versions) != 1 {
			t.Errorf("Template versions count = %d, want 1", len(tmpl.Versions))
		}
		if len(tmpl.Versions) > 0 {
			ver := tmpl.Versions[0]
			if ver.Version != "1.0.0" {
				t.Errorf("Template version = %s, want 1.0.0", ver.Version)
			}
			if len(ver.Files) != 2 {
				t.Errorf("Template files count = %d, want 2", len(ver.Files))
			}
			for _, file := range ver.Files {
				if file.SHA256 == "" {
					t.Errorf("File %s missing SHA256 hash", file.Name)
				}
			}
		}
	}

	if len(index.Bundles) > 0 {
		bundle := index.Bundles[0]
		if bundle.Name != "test-bundle" {
			t.Errorf("Bundle name = %s, want test-bundle", bundle.Name)
		}
		if len(bundle.Versions) != 1 {
			t.Errorf("Bundle versions count = %d, want 1", len(bundle.Versions))
		}
		if len(bundle.Versions) > 0 {
			ver := bundle.Versions[0]
			if ver.Version != "1.0.0" {
				t.Errorf("Bundle version = %s, want 1.0.0", ver.Version)
			}
			if len(ver.Files) != 2 {
				t.Errorf("Bundle files count = %d, want 2", len(ver.Files))
			}
			for _, file := range ver.Files {
				if file.SHA256 == "" {
					t.Errorf("File %s missing SHA256 hash", file.Name)
				}
			}
		}
	}
}

func TestIndexer_BuildIndex_TemplatesOnly(t *testing.T) {
	testDir := t.TempDir()
	templatesDir := filepath.Join(testDir, "templates")

	templateDir := filepath.Join(templatesDir, "test-template", "1.0.0")
	if err := os.MkdirAll(templateDir, 0755); err != nil {
		t.Fatal(err)
	}
	templateMetadata := `{
		"schema_version": "v1",
		"template_name": "test-template",
		"template_type": "yaml",
		"template_description": "Test template",
		"version": "1.0.0",
		"variables": []
	}`
	if err := os.WriteFile(filepath.Join(templateDir, "conjure.json"), []byte(templateMetadata), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(templateDir, "template.tmpl"), []byte("test"), 0600); err != nil {
		t.Fatal(err)
	}

	indexer := NewIndexer()
	index, err := indexer.BuildIndex(templatesDir, "")
	if err != nil {
		t.Fatalf("BuildIndex() error: %v", err)
	}

	if len(index.Templates) != 1 {
		t.Errorf("Templates count = %d, want 1", len(index.Templates))
	}

	if len(index.Bundles) != 0 {
		t.Errorf("Bundles count = %d, want 0", len(index.Bundles))
	}
}

func TestIndexer_BuildIndex_BundlesOnly(t *testing.T) {
	testDir := t.TempDir()
	bundlesDir := filepath.Join(testDir, "bundles")

	bundleDir := filepath.Join(bundlesDir, "test-bundle", "1.0.0")
	if err := os.MkdirAll(bundleDir, 0755); err != nil {
		t.Fatal(err)
	}
	bundleMetadata := `{
		"schema_version": "v1",
		"bundle_name": "test-bundle",
		"bundle_type": "kubernetes",
		"bundle_description": "Test bundle",
		"version": "1.0.0",
		"templates": []
	}`
	if err := os.WriteFile(filepath.Join(bundleDir, "conjure.json"), []byte(bundleMetadata), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(bundleDir, "app.yaml"), []byte("test"), 0600); err != nil {
		t.Fatal(err)
	}

	indexer := NewIndexer()
	index, err := indexer.BuildIndex("", bundlesDir)
	if err != nil {
		t.Fatalf("BuildIndex() error: %v", err)
	}

	if len(index.Templates) != 0 {
		t.Errorf("Templates count = %d, want 0", len(index.Templates))
	}

	if len(index.Bundles) != 1 {
		t.Errorf("Bundles count = %d, want 1", len(index.Bundles))
	}
}

func TestIndexer_WriteIndex(t *testing.T) {
	testDir := t.TempDir()
	templatesDir := filepath.Join(testDir, "templates")

	templateDir := filepath.Join(templatesDir, "test-template", "1.0.0")
	if err := os.MkdirAll(templateDir, 0755); err != nil {
		t.Fatal(err)
	}
	templateMetadata := `{
		"schema_version": "v1",
		"template_name": "test-template",
		"template_type": "yaml",
		"template_description": "Test template",
		"version": "1.0.0",
		"variables": []
	}`
	if err := os.WriteFile(filepath.Join(templateDir, "conjure.json"), []byte(templateMetadata), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(templateDir, "template.tmpl"), []byte("test"), 0600); err != nil {
		t.Fatal(err)
	}

	indexer := NewIndexer()
	index, err := indexer.BuildIndex(templatesDir, "")
	if err != nil {
		t.Fatalf("BuildIndex() error: %v", err)
	}

	outputPath := filepath.Join(testDir, "output", "index.json")
	if err := indexer.WriteIndex(index, outputPath); err != nil {
		t.Fatalf("WriteIndex() error: %v", err)
	}

	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Error("WriteIndex() did not create output file")
	}

	data, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	if len(data) == 0 {
		t.Error("WriteIndex() created empty file")
	}
}

func TestIndexer_ValidateStructure(t *testing.T) {
	tests := []struct {
		name         string
		setupFunc    func(string) string
		resourceType string
		wantErr      bool
	}{
		{
			name: "valid templates structure",
			setupFunc: func(testDir string) string {
				templatesDir := filepath.Join(testDir, "templates")
				templateDir := filepath.Join(templatesDir, "test-template", "1.0.0")
				os.MkdirAll(templateDir, 0755)
				metadata := `{
					"schema_version": "v1",
					"template_name": "test-template",
					"template_type": "yaml",
					"description": "Test",
					"version": "1.0.0",
					"variables": []
				}`
				os.WriteFile(filepath.Join(templateDir, "conjure.json"), []byte(metadata), 0600)
				os.WriteFile(filepath.Join(templateDir, "template.tmpl"), []byte("test"), 0600)
				return templatesDir
			},
			resourceType: "templates",
			wantErr:      false,
		},
		{
			name: "missing conjure.json",
			setupFunc: func(testDir string) string {
				templatesDir := filepath.Join(testDir, "templates")
				templateDir := filepath.Join(templatesDir, "test-template", "1.0.0")
				os.MkdirAll(templateDir, 0755)
				os.WriteFile(filepath.Join(templateDir, "template.tmpl"), []byte("test"), 0600)
				return templatesDir
			},
			resourceType: "templates",
			wantErr:      true,
		},
		{
			name: "no versions",
			setupFunc: func(testDir string) string {
				templatesDir := filepath.Join(testDir, "templates")
				os.MkdirAll(filepath.Join(templatesDir, "test-template"), 0755)
				return templatesDir
			},
			resourceType: "templates",
			wantErr:      true,
		},
		{
			name: "invalid resource type",
			setupFunc: func(testDir string) string {
				return testDir
			},
			resourceType: "invalid",
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testDir := t.TempDir()
			dir := tt.setupFunc(testDir)

			indexer := NewIndexer()
			err := indexer.ValidateStructure(dir, tt.resourceType)

			if tt.wantErr && err == nil {
				t.Error("ValidateStructure() expected error, got nil")
			}

			if !tt.wantErr && err != nil {
				t.Errorf("ValidateStructure() unexpected error: %v", err)
			}
		})
	}
}
