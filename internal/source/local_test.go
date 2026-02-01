package source

import (
	"testing"
)

func TestNewLocalSource(t *testing.T) {
	tests := []struct {
		name    string
		baseDir string
		wantErr bool
	}{
		{
			name:    "valid base directory",
			baseDir: "testdata/local",
			wantErr: false,
		},
		{
			name:    "invalid path with traversal",
			baseDir: "../../../etc",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewLocalSource(tt.baseDir)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewLocalSource() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestLocalSource_ListTemplates(t *testing.T) {
	source, err := NewLocalSource("testdata/local")
	if err != nil {
		t.Fatalf("Failed to create local source: %v", err)
	}

	templates, err := source.ListTemplates()
	if err != nil {
		t.Fatalf("ListTemplates() error = %v", err)
	}

	if len(templates) == 0 {
		t.Error("Expected at least one template")
	}

	found := false
	for _, tmpl := range templates {
		if tmpl.Name == "test-template" {
			found = true
			if len(tmpl.Versions) != 2 {
				t.Errorf("Expected 2 versions, got %d", len(tmpl.Versions))
			}
			if tmpl.Type != "yaml" {
				t.Errorf("Expected type 'yaml', got '%s'", tmpl.Type)
			}
			break
		}
	}

	if !found {
		t.Error("test-template not found in templates list")
	}
}

func TestLocalSource_ListBundles(t *testing.T) {
	source, err := NewLocalSource("testdata/local")
	if err != nil {
		t.Fatalf("Failed to create local source: %v", err)
	}

	bundles, err := source.ListBundles()
	if err != nil {
		t.Fatalf("ListBundles() error = %v", err)
	}

	if len(bundles) == 0 {
		t.Error("Expected at least one bundle")
	}

	found := false
	for _, bundle := range bundles {
		if bundle.Name == "test-bundle" {
			found = true
			if len(bundle.Versions) != 2 {
				t.Errorf("Expected 2 versions, got %d", len(bundle.Versions))
			}
			if bundle.Type != "kubernetes" {
				t.Errorf("Expected type 'kubernetes', got '%s'", bundle.Type)
			}
			break
		}
	}

	if !found {
		t.Error("test-bundle not found in bundles list")
	}
}

func TestLocalSource_GetTemplateVersions(t *testing.T) {
	source, err := NewLocalSource("testdata/local")
	if err != nil {
		t.Fatalf("Failed to create local source: %v", err)
	}

	versions, err := source.GetTemplateVersions("test-template")
	if err != nil {
		t.Fatalf("GetTemplateVersions() error = %v", err)
	}

	if len(versions) != 2 {
		t.Errorf("Expected 2 versions, got %d", len(versions))
	}

	if versions[0] != "1.0.0" || versions[1] != "2.0.0" {
		t.Errorf("Versions not properly sorted: %v", versions)
	}
}

func TestLocalSource_GetTemplate(t *testing.T) {
	source, err := NewLocalSource("testdata/local")
	if err != nil {
		t.Fatalf("Failed to create local source: %v", err)
	}

	tests := []struct {
		name     string
		tmplName string
		version  string
		wantErr  bool
	}{
		{
			name:     "get latest version",
			tmplName: "test-template",
			version:  "",
			wantErr:  false,
		},
		{
			name:     "get specific version",
			tmplName: "test-template",
			version:  "1.0.0",
			wantErr:  false,
		},
		{
			name:     "get non-existent template",
			tmplName: "non-existent",
			version:  "",
			wantErr:  true,
		},
		{
			name:     "get non-existent version",
			tmplName: "test-template",
			version:  "99.0.0",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			content, err := source.GetTemplate(tt.tmplName, tt.version)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetTemplate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if content == nil {
					t.Error("Expected non-nil content")
					return
				}
				if len(content.MetadataRaw) == 0 {
					t.Error("Expected metadata content")
				}
				if len(content.TemplateRaw) == 0 {
					t.Error("Expected template content")
				}
			}
		})
	}
}

func TestLocalSource_GetBundle(t *testing.T) {
	source, err := NewLocalSource("testdata/local")
	if err != nil {
		t.Fatalf("Failed to create local source: %v", err)
	}

	tests := []struct {
		name       string
		bundleName string
		version    string
		wantErr    bool
		wantFiles  int
	}{
		{
			name:       "get latest version",
			bundleName: "test-bundle",
			version:    "",
			wantErr:    false,
			wantFiles:  1,
		},
		{
			name:       "get specific version with multiple files",
			bundleName: "test-bundle",
			version:    "1.0.0",
			wantErr:    false,
			wantFiles:  2,
		},
		{
			name:       "get non-existent bundle",
			bundleName: "non-existent",
			version:    "",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			content, err := source.GetBundle(tt.bundleName, tt.version)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetBundle() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if content == nil {
					t.Error("Expected non-nil content")
					return
				}
				if len(content.MetadataRaw) == 0 {
					t.Error("Expected metadata content")
				}
				if len(content.Files) != tt.wantFiles {
					t.Errorf("Expected %d files, got %d", tt.wantFiles, len(content.Files))
				}
			}
		})
	}
}

func TestFindTemplateFile(t *testing.T) {
	tests := []struct {
		name    string
		dir     string
		wantErr bool
	}{
		{
			name:    "valid template directory",
			dir:     "testdata/local/templates/test-template/1.0.0",
			wantErr: false,
		},
		{
			name:    "non-existent directory",
			dir:     "testdata/local/templates/non-existent/1.0.0",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			file, err := findTemplateFile(tt.dir)
			if (err != nil) != tt.wantErr {
				t.Errorf("findTemplateFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && file == "" {
				t.Error("Expected non-empty file path")
			}
		})
	}
}
