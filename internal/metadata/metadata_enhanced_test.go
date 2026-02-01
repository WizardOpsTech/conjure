package metadata

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestTemplateMetadataWithVersion(t *testing.T) {
	tests := []struct {
		name     string
		metadata TemplateMetadata
		wantErr  bool
	}{
		{
			name: "valid template with version",
			metadata: TemplateMetadata{
				SchemaVersion:       "v1",
				TemplateName:        "test-template",
				TemplateDescription: "Test template",
				Version:             "1.2.3",
				TemplateType:        "terraform",
				Variables: []Variable{
					{
						Name:        "test_var",
						Description: "Test variable",
						Type:        "string",
						Default:     "default",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "valid template without version",
			metadata: TemplateMetadata{
				SchemaVersion:       "v1",
				TemplateName:        "test-template",
				TemplateDescription: "Test template",
				TemplateType:        "yaml",
				Variables: []Variable{
					{
						Name:        "test_var",
						Description: "Test variable",
						Type:        "string",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "invalid version format",
			metadata: TemplateMetadata{
				SchemaVersion:       "v1",
				TemplateName:        "test-template",
				TemplateDescription: "Test template",
				TemplateType:        "yaml",
				Version:             "v1.2.3",
				Variables:           []Variable{},
			},
			wantErr: true,
		},
		{
			name: "invalid version with v prefix",
			metadata: TemplateMetadata{
				SchemaVersion:       "v1",
				TemplateName:        "test-template",
				TemplateDescription: "Test template",
				TemplateType:        "yaml",
				Version:             "1.2",
				Variables:           []Variable{},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateTemplateMetadata(&tt.metadata)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateTemplateMetadata() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestBundleMetadataWithVersion(t *testing.T) {
	tests := []struct {
		name     string
		metadata BundleMetadata
		wantErr  bool
	}{
		{
			name: "valid bundle with version",
			metadata: BundleMetadata{
				SchemaVersion:     "v1",
				BundleType:        "terraform",
				BundleName:        "test-bundle",
				BundleDescription: "Test bundle",
				Version:           "2.0.0",
				SharedVariables: []Variable{
					{
						Name:        "shared_var",
						Description: "Shared variable",
						Type:        "string",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "valid bundle without version",
			metadata: BundleMetadata{
				SchemaVersion:     "v1",
				BundleType:        "kubernetes",
				BundleName:        "test-bundle",
				BundleDescription: "Test bundle",
				SharedVariables:   []Variable{},
			},
			wantErr: false,
		},
		{
			name: "invalid bundle version",
			metadata: BundleMetadata{
				SchemaVersion:     "v1",
				BundleType:        "terraform",
				BundleName:        "test-bundle",
				BundleDescription: "Test bundle",
				Version:           "invalid",
				SharedVariables:   []Variable{},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateBundleMetadata(&tt.metadata)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateBundleMetadata() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestVariableNameValidation(t *testing.T) {
	tests := []struct {
		name     string
		variable Variable
		wantErr  bool
	}{
		{
			name: "valid variable with underscore",
			variable: Variable{
				Name:        "my_variable",
				Description: "Test variable",
				Type:        "string",
			},
			wantErr: false,
		},
		{
			name: "valid variable with hyphen",
			variable: Variable{
				Name:        "my-variable",
				Description: "Test variable",
				Type:        "string",
			},
			wantErr: false,
		},
		{
			name: "valid variable starting with number",
			variable: Variable{
				Name:        "123variable",
				Description: "Test variable",
				Type:        "string",
			},
			wantErr: false,
		},
		{
			name: "invalid variable with space",
			variable: Variable{
				Name:        "my variable",
				Description: "Test variable",
				Type:        "string",
			},
			wantErr: true,
		},
		{
			name: "invalid variable with special char",
			variable: Variable{
				Name:        "my$variable",
				Description: "Test variable",
				Type:        "string",
			},
			wantErr: true,
		},
		{
			name: "empty variable name",
			variable: Variable{
				Name:        "",
				Description: "Test variable",
				Type:        "string",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateVariable(&tt.variable)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateVariable() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMaxVariablesLimit(t *testing.T) {
	// Create a template with too many variables
	variables := make([]Variable, MaxVariablesPerTemplate+1)
	for i := 0; i < MaxVariablesPerTemplate+1; i++ {
		variables[i] = Variable{
			Name:        "var_" + string(rune(i)),
			Description: "Test variable",
			Type:        "string",
		}
	}

	metadata := TemplateMetadata{
		SchemaVersion:       "v1",
		TemplateName:        "test-template",
		TemplateDescription: "Test template",
		TemplateType:        "yaml",
		Variables:           variables,
	}

	err := validateTemplateMetadata(&metadata)
	if err == nil {
		t.Error("validateTemplateMetadata() should fail with too many variables")
	}
}

func TestLoadTemplateMetadataFileSize(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a metadata file that's too large
	largeContent := strings.Repeat("x", MaxMetadataFileSize+1)
	largePath := filepath.Join(tmpDir, "large.json")
	err := os.WriteFile(largePath, []byte(largeContent), 0600)
	if err != nil {
		t.Fatalf("Failed to create large file: %v", err)
	}

	_, err = LoadTemplateMetadata(largePath)
	if err == nil {
		t.Error("LoadTemplateMetadata() should fail with oversized file")
	}
}

func TestLoadBundleMetadataFileSize(t *testing.T) {
	tmpDir := t.TempDir()

	largeContent := strings.Repeat("x", MaxMetadataFileSize+1)
	largePath := filepath.Join(tmpDir, "large.json")
	err := os.WriteFile(largePath, []byte(largeContent), 0600)
	if err != nil {
		t.Fatalf("Failed to create large file: %v", err)
	}

	_, err = LoadBundleMetadata(largePath)
	if err == nil {
		t.Error("LoadBundleMetadata() should fail with oversized file")
	}
}

func TestLoadTemplateMetadataWithNewFields(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a valid metadata file with new fields
	metadata := TemplateMetadata{
		SchemaVersion:       "v1",
		TemplateName:        "test-template",
		TemplateDescription: "Test template",
		Version:             "1.0.0",
		TemplateType:        "terraform",
		Variables: []Variable{
			{
				Name:        "test_var",
				Description: "Test variable",
				Type:        "string",
				Default:     "test",
			},
		},
	}

	data, err := json.Marshal(metadata)
	if err != nil {
		t.Fatalf("Failed to marshal metadata: %v", err)
	}

	metadataPath := filepath.Join(tmpDir, "test.json")
	err = os.WriteFile(metadataPath, data, 0600)
	if err != nil {
		t.Fatalf("Failed to write metadata file: %v", err)
	}

	loaded, err := LoadTemplateMetadata(metadataPath)
	if err != nil {
		t.Fatalf("LoadTemplateMetadata() error = %v", err)
	}

	if loaded.Version != "1.0.0" {
		t.Errorf("Version = %v, want %v", loaded.Version, "1.0.0")
	}

	if loaded.TemplateType != "terraform" {
		t.Errorf("TemplateType = %v, want %v", loaded.TemplateType, "terraform")
	}
}

func TestLoadBundleMetadataWithVersion(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a valid bundle metadata file with version
	metadata := BundleMetadata{
		SchemaVersion:     "v1",
		BundleType:        "terraform",
		BundleName:        "test-bundle",
		BundleDescription: "Test bundle",
		Version:           "2.1.0",
		SharedVariables: []Variable{
			{
				Name:        "env",
				Description: "Environment",
				Type:        "string",
				Default:     "dev",
			},
		},
	}

	data, err := json.Marshal(metadata)
	if err != nil {
		t.Fatalf("Failed to marshal metadata: %v", err)
	}

	metadataPath := filepath.Join(tmpDir, "conjure.json")
	err = os.WriteFile(metadataPath, data, 0600)
	if err != nil {
		t.Fatalf("Failed to write metadata file: %v", err)
	}

	loaded, err := LoadBundleMetadata(metadataPath)
	if err != nil {
		t.Fatalf("LoadBundleMetadata() error = %v", err)
	}

	if loaded.Version != "2.1.0" {
		t.Errorf("Version = %v, want %v", loaded.Version, "2.1.0")
	}
}
