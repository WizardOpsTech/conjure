package source

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func createMockServer() *httptest.Server {
	mux := http.NewServeMux()

	mux.HandleFunc("/index.json", func(w http.ResponseWriter, r *http.Request) {
		index := Index{
			SchemaVersion: "v1",
			LastUpdated:   time.Now(),
			Templates: []TemplateIndexEntry{
				{
					Name:        "test-template",
					Type:        "yaml",
					Description: "Test template",
					Versions: []VersionEntry{
						{
							Version: "1.0.0",
							Files: []FileEntry{
								{
									Name:   "conjure.json",
									Path:   "templates/test-template/1.0.0/conjure.json",
									Size:   100,
									SHA256: "abc123",
								},
								{
									Name:   "template.tmpl",
									Path:   "templates/test-template/1.0.0/template.tmpl",
									Size:   50,
									SHA256: "def456",
								},
							},
						},
					},
				},
			},
			Bundles: []BundleIndexEntry{
				{
					Name:        "test-bundle",
					Type:        "kubernetes",
					Description: "Test bundle",
					Versions: []VersionEntry{
						{
							Version: "1.0.0",
							Files: []FileEntry{
								{
									Name:   "conjure.json",
									Path:   "bundles/test-bundle/1.0.0/conjure.json",
									Size:   150,
									SHA256: "ghi789",
								},
								{
									Name:   "app.yaml.tmpl",
									Path:   "bundles/test-bundle/1.0.0/app.yaml.tmpl",
									Size:   75,
									SHA256: "jkl012",
								},
							},
						},
					},
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(index)
	})

	mux.HandleFunc("/templates/test-template/1.0.0/conjure.json", func(w http.ResponseWriter, r *http.Request) {
		metadata := map[string]interface{}{
			"schema_version":       "v1",
			"template_name":        "test-template",
			"template_description": "Test template",
			"version":              "1.0.0",
			"template_type":        "yaml",
			"variables":            []interface{}{},
		}
		json.NewEncoder(w).Encode(metadata)
	})

	mux.HandleFunc("/templates/test-template/1.0.0/template.tmpl", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("template content"))
	})

	mux.HandleFunc("/bundles/test-bundle/1.0.0/conjure.json", func(w http.ResponseWriter, r *http.Request) {
		metadata := map[string]interface{}{
			"schema_version":     "v1",
			"bundle_name":        "test-bundle",
			"bundle_description": "Test bundle",
			"version":            "1.0.0",
			"bundle_type":        "kubernetes",
			"shared_variables":   []interface{}{},
			"template_variables": map[string]interface{}{},
		}
		json.NewEncoder(w).Encode(metadata)
	})

	mux.HandleFunc("/bundles/test-bundle/1.0.0/app.yaml.tmpl", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("app template content"))
	})

	return httptest.NewServer(mux)
}

func TestNewRemoteSource(t *testing.T) {
	tempDir := t.TempDir()
	server := createMockServer()
	defer server.Close()

	tests := []struct {
		name     string
		baseURL  string
		cacheDir string
		wantErr  bool
	}{
		{
			name:     "valid configuration",
			baseURL:  server.URL,
			cacheDir: tempDir,
			wantErr:  false,
		},
		{
			name:     "invalid URL",
			baseURL:  "not-a-url",
			cacheDir: tempDir,
			wantErr:  true,
		},
		{
			name:     "invalid cache dir",
			baseURL:  server.URL,
			cacheDir: "../../../etc",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewRemoteSource(tt.baseURL, tt.cacheDir)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewRemoteSource() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRemoteSource_FetchIndex(t *testing.T) {
	tempDir := t.TempDir()
	server := createMockServer()
	defer server.Close()

	rs, err := NewRemoteSource(server.URL, tempDir)
	if err != nil {
		t.Fatalf("Failed to create remote source: %v", err)
	}

	err = rs.fetchIndex(false)
	if err != nil {
		t.Fatalf("fetchIndex() error = %v", err)
	}

	if rs.indexCache == nil {
		t.Error("Index cache should be populated")
	}

	if rs.indexCache.SchemaVersion != "v1" {
		t.Errorf("Expected schema version 'v1', got '%s'", rs.indexCache.SchemaVersion)
	}

	err = rs.fetchIndex(false)
	if err != nil {
		t.Fatalf("Second fetchIndex() error = %v", err)
	}
}

func TestRemoteSource_ListTemplates(t *testing.T) {
	tempDir := t.TempDir()
	server := createMockServer()
	defer server.Close()

	rs, err := NewRemoteSource(server.URL, tempDir)
	if err != nil {
		t.Fatalf("Failed to create remote source: %v", err)
	}

	templates, err := rs.ListTemplates()
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
			if tmpl.Type != "yaml" {
				t.Errorf("Expected type 'yaml', got '%s'", tmpl.Type)
			}
			if len(tmpl.Versions) != 1 {
				t.Errorf("Expected 1 version, got %d", len(tmpl.Versions))
			}
			break
		}
	}

	if !found {
		t.Error("test-template not found in templates list")
	}
}

func TestRemoteSource_ListBundles(t *testing.T) {
	tempDir := t.TempDir()
	server := createMockServer()
	defer server.Close()

	rs, err := NewRemoteSource(server.URL, tempDir)
	if err != nil {
		t.Fatalf("Failed to create remote source: %v", err)
	}

	bundles, err := rs.ListBundles()
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
			if bundle.Type != "kubernetes" {
				t.Errorf("Expected type 'kubernetes', got '%s'", bundle.Type)
			}
			if len(bundle.Versions) != 1 {
				t.Errorf("Expected 1 version, got %d", len(bundle.Versions))
			}
			break
		}
	}

	if !found {
		t.Error("test-bundle not found in bundles list")
	}
}

func TestRemoteSource_GetTemplateVersions(t *testing.T) {
	tempDir := t.TempDir()
	server := createMockServer()
	defer server.Close()

	rs, err := NewRemoteSource(server.URL, tempDir)
	if err != nil {
		t.Fatalf("Failed to create remote source: %v", err)
	}

	versions, err := rs.GetTemplateVersions("test-template")
	if err != nil {
		t.Fatalf("GetTemplateVersions() error = %v", err)
	}

	if len(versions) != 1 {
		t.Errorf("Expected 1 version, got %d", len(versions))
	}

	if versions[0] != "1.0.0" {
		t.Errorf("Expected version '1.0.0', got '%s'", versions[0])
	}

	versions, err = rs.GetTemplateVersions("non-existent")
	if err != nil {
		t.Fatalf("GetTemplateVersions() for non-existent template error = %v", err)
	}

	if len(versions) != 0 {
		t.Errorf("Expected 0 versions for non-existent template, got %d", len(versions))
	}
}

func TestRemoteSource_GetBundleVersions(t *testing.T) {
	tempDir := t.TempDir()
	server := createMockServer()
	defer server.Close()

	rs, err := NewRemoteSource(server.URL, tempDir)
	if err != nil {
		t.Fatalf("Failed to create remote source: %v", err)
	}

	versions, err := rs.GetBundleVersions("test-bundle")
	if err != nil {
		t.Fatalf("GetBundleVersions() error = %v", err)
	}

	if len(versions) != 1 {
		t.Errorf("Expected 1 version, got %d", len(versions))
	}

	if versions[0] != "1.0.0" {
		t.Errorf("Expected version '1.0.0', got '%s'", versions[0])
	}
}

func TestRemoteSource_RefreshIndex(t *testing.T) {
	tempDir := t.TempDir()
	server := createMockServer()
	defer server.Close()

	rs, err := NewRemoteSource(server.URL, tempDir)
	if err != nil {
		t.Fatalf("Failed to create remote source: %v", err)
	}

	err = rs.fetchIndex(false)
	if err != nil {
		t.Fatalf("fetchIndex() error = %v", err)
	}

	err = rs.RefreshIndex()
	if err != nil {
		t.Fatalf("RefreshIndex() error = %v", err)
	}

	if rs.indexCache == nil {
		t.Error("Index cache should be populated after refresh")
	}
}

func TestRemoteSource_InvalidIndex(t *testing.T) {
	tempDir := t.TempDir()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/index.json" {
			w.Write([]byte("invalid json"))
		}
	}))
	defer server.Close()

	rs, err := NewRemoteSource(server.URL, tempDir)
	if err != nil {
		t.Fatalf("Failed to create remote source: %v", err)
	}

	err = rs.fetchIndex(false)
	if err == nil {
		t.Error("Expected error for invalid index JSON")
	}
}

func TestRemoteSource_UnsupportedSchemaVersion(t *testing.T) {
	tempDir := t.TempDir()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/index.json" {
			index := Index{
				SchemaVersion: "v99",
				Templates:     []TemplateIndexEntry{},
			}
			json.NewEncoder(w).Encode(index)
		}
	}))
	defer server.Close()

	rs, err := NewRemoteSource(server.URL, tempDir)
	if err != nil {
		t.Fatalf("Failed to create remote source: %v", err)
	}

	err = rs.fetchIndex(false)
	if err == nil {
		t.Error("Expected error for unsupported schema version")
	}
}

func TestRemoteSource_HTTPError(t *testing.T) {
	tempDir := t.TempDir()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	rs, err := NewRemoteSource(server.URL, tempDir)
	if err != nil {
		t.Fatalf("Failed to create remote source: %v", err)
	}

	err = rs.fetchIndex(false)
	if err == nil {
		t.Error("Expected error for HTTP 500")
	}
}

func TestRemoteSource_GetTemplateNotFound(t *testing.T) {
	tempDir := t.TempDir()
	server := createMockServer()
	defer server.Close()

	rs, err := NewRemoteSource(server.URL, tempDir)
	if err != nil {
		t.Fatalf("Failed to create remote source: %v", err)
	}

	_, err = rs.GetTemplate("non-existent", "")
	if err == nil {
		t.Error("Expected error for non-existent template")
	}
}

func TestRemoteSource_GetBundleNotFound(t *testing.T) {
	tempDir := t.TempDir()
	server := createMockServer()
	defer server.Close()

	rs, err := NewRemoteSource(server.URL, tempDir)
	if err != nil {
		t.Fatalf("Failed to create remote source: %v", err)
	}

	_, err = rs.GetBundle("non-existent", "")
	if err == nil {
		t.Error("Expected error for non-existent bundle")
	}
}

func TestRemoteSource_InvalidTemplateName(t *testing.T) {
	tempDir := t.TempDir()
	server := createMockServer()
	defer server.Close()

	rs, err := NewRemoteSource(server.URL, tempDir)
	if err != nil {
		t.Fatalf("Failed to create remote source: %v", err)
	}

	_, err = rs.GetTemplate("../invalid", "")
	if err == nil {
		t.Error("Expected error for invalid template name")
	}
}
