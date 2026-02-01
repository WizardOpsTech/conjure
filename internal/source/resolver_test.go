package source

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/wizardopstech/conjure/internal/config"
)

func TestNewResolver(t *testing.T) {
	tests := []struct {
		name         string
		cfg          *config.Config
		resourceType string
		wantErr      bool
		wantSources  int
	}{
		{
			name: "local only templates",
			cfg: &config.Config{
				TemplatesSource:   "local",
				TemplatesLocalDir: t.TempDir(),
				TemplatesPriority: "local-first",
				CacheDir:          t.TempDir(),
			},
			resourceType: "template",
			wantErr:      false,
			wantSources:  1,
		},
		{
			name: "local only bundles",
			cfg: &config.Config{
				BundlesSource:   "local",
				BundlesLocalDir: t.TempDir(),
				BundlesPriority: "local-first",
				CacheDir:        t.TempDir(),
			},
			resourceType: "bundle",
			wantErr:      false,
			wantSources:  1,
		},
		{
			name: "both sources - local first",
			cfg: &config.Config{
				TemplatesSource:    "both",
				TemplatesLocalDir:  t.TempDir(),
				TemplatesRemoteURL: "https://example.com/templates",
				TemplatesPriority:  "local-first",
				CacheDir:           t.TempDir(),
			},
			resourceType: "template",
			wantErr:      false,
			wantSources:  2,
		},
		{
			name: "both sources - remote first",
			cfg: &config.Config{
				TemplatesSource:    "both",
				TemplatesLocalDir:  t.TempDir(),
				TemplatesRemoteURL: "https://example.com/templates",
				TemplatesPriority:  "remote-first",
				CacheDir:           t.TempDir(),
			},
			resourceType: "template",
			wantErr:      false,
			wantSources:  2,
		},
		{
			name:         "invalid resource type",
			cfg:          &config.Config{},
			resourceType: "invalid",
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resolver, err := NewResolver(tt.cfg, tt.resourceType)

			if tt.wantErr {
				if err == nil {
					t.Error("NewResolver() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("NewResolver() unexpected error: %v", err)
			}

			if len(resolver.sources) != tt.wantSources {
				t.Errorf("NewResolver() got %d sources, want %d", len(resolver.sources), tt.wantSources)
			}

			if len(resolver.sourceNames) != tt.wantSources {
				t.Errorf("NewResolver() got %d source names, want %d", len(resolver.sourceNames), tt.wantSources)
			}

			if len(resolver.priorityOrder) != tt.wantSources {
				t.Errorf("NewResolver() got %d priority entries, want %d", len(resolver.priorityOrder), tt.wantSources)
			}
		})
	}
}

func TestResolver_GetTemplate(t *testing.T) {
	testDir := t.TempDir()
	templateDir := filepath.Join(testDir, "templates", "test-template", "1.0.0")
	if err := os.MkdirAll(templateDir, 0755); err != nil {
		t.Fatal(err)
	}

	metadataContent := `{
		"schema_version": "v1",
		"template_name": "test-template",
		"template_type": "yaml",
		"template_description": "Test template",
		"version": "1.0.0",
		"variables": []
	}`
	if err := os.WriteFile(filepath.Join(templateDir, "conjure.json"), []byte(metadataContent), 0600); err != nil {
		t.Fatal(err)
	}

	templateContent := "test: {{ .value }}"
	if err := os.WriteFile(filepath.Join(templateDir, "template.tmpl"), []byte(templateContent), 0600); err != nil {
		t.Fatal(err)
	}

	cfg := &config.Config{
		TemplatesSource:   "local",
		TemplatesLocalDir: testDir,
		TemplatesPriority: "local-first",
		CacheDir:          t.TempDir(),
	}

	resolver, err := NewResolver(cfg, "template")
	if err != nil {
		t.Fatalf("NewResolver() error: %v", err)
	}

	content, sourceName, err := resolver.GetTemplate("test-template", "1.0.0")
	if err != nil {
		t.Fatalf("GetTemplate() error: %v", err)
	}

	if sourceName != "local" {
		t.Errorf("GetTemplate() got source %s, want local", sourceName)
	}

	if content.Version != "1.0.0" {
		t.Errorf("GetTemplate() got version %s, want 1.0.0", content.Version)
	}

	_, _, err = resolver.GetTemplate("non-existent", "1.0.0")
	if err == nil {
		t.Error("GetTemplate() expected error for non-existent template, got nil")
	}
}

func TestResolver_GetBundle(t *testing.T) {
	testDir := t.TempDir()
	bundleDir := filepath.Join(testDir, "bundles", "test-bundle", "1.0.0")
	if err := os.MkdirAll(bundleDir, 0755); err != nil {
		t.Fatal(err)
	}

	metadataContent := `{
		"schema_version": "v1",
		"bundle_name": "test-bundle",
		"bundle_type": "kubernetes",
		"bundle_description": "Test bundle",
		"version": "1.0.0",
		"templates": []
	}`
	if err := os.WriteFile(filepath.Join(bundleDir, "conjure.json"), []byte(metadataContent), 0600); err != nil {
		t.Fatal(err)
	}

	templateContent := "test: value"
	if err := os.WriteFile(filepath.Join(bundleDir, "test.yaml"), []byte(templateContent), 0600); err != nil {
		t.Fatal(err)
	}

	cfg := &config.Config{
		BundlesSource:   "local",
		BundlesLocalDir: testDir,
		BundlesPriority: "local-first",
		CacheDir:        t.TempDir(),
	}

	resolver, err := NewResolver(cfg, "bundle")
	if err != nil {
		t.Fatalf("NewResolver() error: %v", err)
	}

	content, sourceName, err := resolver.GetBundle("test-bundle", "1.0.0")
	if err != nil {
		t.Fatalf("GetBundle() error: %v", err)
	}

	if sourceName != "local" {
		t.Errorf("GetBundle() got source %s, want local", sourceName)
	}

	if content.Version != "1.0.0" {
		t.Errorf("GetBundle() got version %s, want 1.0.0", content.Version)
	}

	_, _, err = resolver.GetBundle("non-existent", "1.0.0")
	if err == nil {
		t.Error("GetBundle() expected error for non-existent bundle, got nil")
	}
}

func TestResolver_ListTemplates(t *testing.T) {
	testDir := t.TempDir()

	templateDir1 := filepath.Join(testDir, "templates", "template-1", "1.0.0")
	if err := os.MkdirAll(templateDir1, 0755); err != nil {
		t.Fatal(err)
	}
	metadata1 := `{
		"schema_version": "v1",
		"template_name": "template-1",
		"template_type": "yaml",
		"template_description": "Template 1",
		"version": "1.0.0",
		"variables": []
	}`
	if err := os.WriteFile(filepath.Join(templateDir1, "conjure.json"), []byte(metadata1), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(templateDir1, "template.tmpl"), []byte("test1"), 0600); err != nil {
		t.Fatal(err)
	}

	templateDir2 := filepath.Join(testDir, "templates", "template-2", "2.0.0")
	if err := os.MkdirAll(templateDir2, 0755); err != nil {
		t.Fatal(err)
	}
	metadata2 := `{
		"schema_version": "v1",
		"template_name": "template-2",
		"template_type": "json",
		"template_description": "Template 2",
		"version": "2.0.0",
		"variables": []
	}`
	if err := os.WriteFile(filepath.Join(templateDir2, "conjure.json"), []byte(metadata2), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(templateDir2, "template.tmpl"), []byte("test2"), 0600); err != nil {
		t.Fatal(err)
	}

	cfg := &config.Config{
		TemplatesSource:   "local",
		TemplatesLocalDir: testDir,
		TemplatesPriority: "local-first",
		CacheDir:          t.TempDir(),
	}

	resolver, err := NewResolver(cfg, "template")
	if err != nil {
		t.Fatalf("NewResolver() error: %v", err)
	}

	templates, err := resolver.ListTemplates()
	if err != nil {
		t.Fatalf("ListTemplates() error: %v", err)
	}

	if len(templates) != 2 {
		t.Errorf("ListTemplates() got %d templates, want 2", len(templates))
	}
}

func TestResolver_ListBundles(t *testing.T) {
	testDir := t.TempDir()

	bundleDir1 := filepath.Join(testDir, "bundles", "bundle-1", "1.0.0")
	if err := os.MkdirAll(bundleDir1, 0755); err != nil {
		t.Fatal(err)
	}
	metadata1 := `{
		"schema_version": "v1",
		"bundle_name": "bundle-1",
		"bundle_type": "kubernetes",
		"bundle_description": "Bundle 1",
		"version": "1.0.0",
		"templates": []
	}`
	if err := os.WriteFile(filepath.Join(bundleDir1, "conjure.json"), []byte(metadata1), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(bundleDir1, "test.yaml"), []byte("test1"), 0600); err != nil {
		t.Fatal(err)
	}

	bundleDir2 := filepath.Join(testDir, "bundles", "bundle-2", "2.0.0")
	if err := os.MkdirAll(bundleDir2, 0755); err != nil {
		t.Fatal(err)
	}
	metadata2 := `{
		"schema_version": "v1",
		"bundle_name": "bundle-2",
		"bundle_type": "terraform",
		"bundle_description": "Bundle 2",
		"version": "2.0.0",
		"templates": []
	}`
	if err := os.WriteFile(filepath.Join(bundleDir2, "conjure.json"), []byte(metadata2), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(bundleDir2, "main.tf"), []byte("test2"), 0600); err != nil {
		t.Fatal(err)
	}

	cfg := &config.Config{
		BundlesSource:   "local",
		BundlesLocalDir: testDir,
		BundlesPriority: "local-first",
		CacheDir:        t.TempDir(),
	}

	resolver, err := NewResolver(cfg, "bundle")
	if err != nil {
		t.Fatalf("NewResolver() error: %v", err)
	}

	bundles, err := resolver.ListBundles()
	if err != nil {
		t.Fatalf("ListBundles() error: %v", err)
	}

	if len(bundles) != 2 {
		t.Errorf("ListBundles() got %d bundles, want 2", len(bundles))
	}
}

func TestResolver_GetLatestVersion(t *testing.T) {
	testDir := t.TempDir()

	versions := []string{"1.0.0", "1.1.0", "2.0.0"}
	for _, ver := range versions {
		templateDir := filepath.Join(testDir, "templates", "test-template", ver)
		if err := os.MkdirAll(templateDir, 0755); err != nil {
			t.Fatal(err)
		}

		metadata := `{
			"schema_version": "v1",
			"template_name": "test-template",
			"template_type": "yaml",
			"template_description": "Test template",
			"version": "` + ver + `",
			"variables": []
		}`
		if err := os.WriteFile(filepath.Join(templateDir, "conjure.json"), []byte(metadata), 0600); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(templateDir, "template.tmpl"), []byte("test"), 0600); err != nil {
			t.Fatal(err)
		}
	}

	cfg := &config.Config{
		TemplatesSource:   "local",
		TemplatesLocalDir: testDir,
		TemplatesPriority: "local-first",
		CacheDir:          t.TempDir(),
	}

	resolver, err := NewResolver(cfg, "template")
	if err != nil {
		t.Fatalf("NewResolver() error: %v", err)
	}

	latest, err := resolver.GetLatestVersion("template", "test-template")
	if err != nil {
		t.Fatalf("GetLatestVersion() error: %v", err)
	}

	if latest != "2.0.0" {
		t.Errorf("GetLatestVersion() got %s, want 2.0.0", latest)
	}

	_, err = resolver.GetLatestVersion("template", "non-existent")
	if err == nil {
		t.Error("GetLatestVersion() expected error for non-existent template, got nil")
	}
}

func TestDeduplicateVersions(t *testing.T) {
	tests := []struct {
		name     string
		versions []string
		want     []string
	}{
		{
			name:     "no duplicates",
			versions: []string{"1.0.0", "1.1.0", "2.0.0"},
			want:     []string{"1.0.0", "1.1.0", "2.0.0"},
		},
		{
			name:     "with duplicates",
			versions: []string{"1.0.0", "1.1.0", "1.0.0", "2.0.0", "1.1.0"},
			want:     []string{"1.0.0", "1.1.0", "2.0.0"},
		},
		{
			name:     "empty",
			versions: []string{},
			want:     []string{},
		},
		{
			name:     "all duplicates",
			versions: []string{"1.0.0", "1.0.0", "1.0.0"},
			want:     []string{"1.0.0"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := deduplicateVersions(tt.versions)

			if len(got) != len(tt.want) {
				t.Errorf("deduplicateVersions() length = %d, want %d", len(got), len(tt.want))
				return
			}

			gotMap := make(map[string]bool)
			for _, v := range got {
				gotMap[v] = true
			}

			for _, v := range tt.want {
				if !gotMap[v] {
					t.Errorf("deduplicateVersions() missing version %s", v)
				}
			}
		})
	}
}
