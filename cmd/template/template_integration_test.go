package template

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/viper"
)

// setupTestEnvironment creates a temporary directory with conjure config and templates
func setupTestEnvironment(t *testing.T) (baseDir string, cleanup func()) {
	tmpDir := t.TempDir()

	// Create directory structure: tmpDir/templates/
	templatesDir := filepath.Join(tmpDir, "templates")
	if err := os.MkdirAll(templatesDir, 0755); err != nil {
		t.Fatalf("Failed to create templates dir: %v", err)
	}

	// Store old viper config values
	oldTemplatesDir := viper.GetString("templates_dir")
	oldBundlesDir := viper.GetString("bundles_dir")

	// Set viper config directly for tests
	viper.Set("templates_dir", tmpDir)
	viper.Set("bundles_dir", tmpDir)

	cleanup = func() {
		// Restore old viper config
		if oldTemplatesDir != "" {
			viper.Set("templates_dir", oldTemplatesDir)
		}
		if oldBundlesDir != "" {
			viper.Set("bundles_dir", oldBundlesDir)
		}
	}

	return tmpDir, cleanup
}

// TestGenerateTemplate_FullWorkflow tests the complete template generation workflow
// from reading template + values files to writing output
func TestGenerateTemplate_FullWorkflow(t *testing.T) {
	// Setup test environment
	baseDir, cleanup := setupTestEnvironment(t)
	defer cleanup()

	// Create a test template file in templates/
	templatePath := filepath.Join(baseDir, "templates", "deployment.yaml.tmpl")
	templateContent := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{.app_name}}-deployment
  namespace: {{.namespace}}
spec:
  replicas: {{.replicas}}`

	if err := os.WriteFile(templatePath, []byte(templateContent), 0644); err != nil {
		t.Fatalf("Failed to create template file: %v", err)
	}

	// Create a values file
	valuesPath := filepath.Join(baseDir, "values.yaml")
	valuesContent := `app_name: myapp
namespace: production
replicas: 3`

	if err := os.WriteFile(valuesPath, []byte(valuesContent), 0644); err != nil {
		t.Fatalf("Failed to create values file: %v", err)
	}

	// Set output path
	outputPath := filepath.Join(baseDir, "deployment.yaml")

	// Execute the full generateTemplate workflow
	err := generateTemplate("deployment.yaml", outputPath, []string{}, false, valuesPath)
	if err != nil {
		t.Fatalf("generateTemplate() failed: %v", err)
	}

	// Verify: Check that output file was created
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Fatalf("Output file was not created: %s", outputPath)
	}

	// Read and verify output content
	output, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	outputStr := string(output)

	// Verify template variables were replaced correctly
	if !strings.Contains(outputStr, "name: myapp-deployment") {
		t.Errorf("Output should contain 'name: myapp-deployment', got:\n%s", outputStr)
	}
	if !strings.Contains(outputStr, "namespace: production") {
		t.Errorf("Output should contain 'namespace: production', got:\n%s", outputStr)
	}
	if !strings.Contains(outputStr, "replicas: 3") {
		t.Errorf("Output should contain 'replicas: 3', got:\n%s", outputStr)
	}

	// Verify no template syntax remains
	if strings.Contains(outputStr, "{{") || strings.Contains(outputStr, "}}") {
		t.Errorf("Output should not contain template syntax, got:\n%s", outputStr)
	}
}

// TestGenerateTemplate_WithCLIOverrides tests that CLI --var flags override values file
func TestGenerateTemplate_WithCLIOverrides(t *testing.T) {
	baseDir, cleanup := setupTestEnvironment(t)
	defer cleanup()

	// Template with multiple variables
	templatePath := filepath.Join(baseDir, "templates", "config.yaml.tmpl")
	templateContent := `app: {{.app_name}}
env: {{.environment}}
debug: {{.debug_mode}}`

	if err := os.WriteFile(templatePath, []byte(templateContent), 0644); err != nil {
		t.Fatalf("Failed to create template file: %v", err)
	}

	// Values file with defaults
	valuesPath := filepath.Join(baseDir, "values.yaml")
	valuesContent := `app_name: default-app
environment: development
debug_mode: false`

	if err := os.WriteFile(valuesPath, []byte(valuesContent), 0644); err != nil {
		t.Fatalf("Failed to create values file: %v", err)
	}

	outputPath := filepath.Join(baseDir, "config.yaml")

	// Execute with CLI overrides
	varsList := []string{
		"app_name=override-app",
		"environment=production",
	}
	err := generateTemplate("config.yaml", outputPath, varsList, false, valuesPath)
	if err != nil {
		t.Fatalf("generateTemplate() failed: %v", err)
	}

	// Verify output
	output, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	outputStr := string(output)

	// CLI overrides should take precedence
	if !strings.Contains(outputStr, "app: override-app") {
		t.Errorf("CLI override should set app to 'override-app', got:\n%s", outputStr)
	}
	if !strings.Contains(outputStr, "env: production") {
		t.Errorf("CLI override should set env to 'production', got:\n%s", outputStr)
	}
	// Non-overridden value from file should still be used
	if !strings.Contains(outputStr, "debug: false") {
		t.Errorf("Values file should provide debug_mode, got:\n%s", outputStr)
	}
}

// TestGenerateTemplate_MissingRequiredVariable tests error handling for missing variables
func TestGenerateTemplate_MissingRequiredVariable(t *testing.T) {
	baseDir, cleanup := setupTestEnvironment(t)
	defer cleanup()

	// Template requiring a variable
	templatePath := filepath.Join(baseDir, "templates", "test.yaml.tmpl")
	templateContent := `name: {{.required_var}}`

	if err := os.WriteFile(templatePath, []byte(templateContent), 0644); err != nil {
		t.Fatalf("Failed to create template file: %v", err)
	}

	outputPath := filepath.Join(baseDir, "output.yaml")

	// Execute WITHOUT providing the required variable - should fail
	err := generateTemplate("test.yaml", outputPath, []string{}, false, "")
	if err == nil {
		t.Error("Expected error for missing required variable, got nil")
	}

	// Verify error message mentions the issue
	if !strings.Contains(err.Error(), "execute") {
		t.Errorf("Error should mention template execution failure, got: %v", err)
	}
}

// TestGenerateTemplate_WithMetadata tests template generation with metadata file
func TestGenerateTemplate_WithMetadata(t *testing.T) {
	baseDir, cleanup := setupTestEnvironment(t)
	defer cleanup()

	// Template file
	templatePath := filepath.Join(baseDir, "templates", "service.yaml.tmpl")
	templateContent := `service: {{.service_name}}
port: {{.port}}`

	if err := os.WriteFile(templatePath, []byte(templateContent), 0644); err != nil {
		t.Fatalf("Failed to create template file: %v", err)
	}

	// Create metadata file with defaults (JSON format)
	metadataPath := filepath.Join(baseDir, "templates", "service.yaml.json")
	metadataContent := `{
  "description": "Service configuration template",
  "variables": [
    {
      "name": "service_name",
      "description": "Service name",
      "required": true
    },
    {
      "name": "port",
      "description": "Service port",
      "default": "8080"
    }
  ]
}`

	if err := os.WriteFile(metadataPath, []byte(metadataContent), 0644); err != nil {
		t.Fatalf("Failed to create metadata file: %v", err)
	}

	outputPath := filepath.Join(baseDir, "service.yaml")

	// Execute with only service_name (port should use default from metadata)
	varsList := []string{"service_name=myservice"}
	err := generateTemplate("service.yaml", outputPath, varsList, false, "")
	if err != nil {
		t.Fatalf("generateTemplate() failed: %v", err)
	}

	// Verify content
	output, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to read output: %v", err)
	}

	outputStr := string(output)

	// Should use provided variable and metadata default
	if !strings.Contains(outputStr, "service: myservice") {
		t.Errorf("Output should contain 'service: myservice', got:\n%s", outputStr)
	}
	if !strings.Contains(outputStr, "port: 8080") {
		t.Errorf("Output should use default port 8080 from metadata, got:\n%s", outputStr)
	}
}

// TestGenerateTemplate_NonexistentTemplate tests error handling for missing template file
func TestGenerateTemplate_NonexistentTemplate(t *testing.T) {
	baseDir, cleanup := setupTestEnvironment(t)
	defer cleanup()

	outputPath := filepath.Join(baseDir, "output.yaml")

	// Execute with nonexistent template - should fail
	err := generateTemplate("nonexistent.yaml", outputPath, []string{}, false, "")
	if err == nil {
		t.Error("Expected error for nonexistent template file, got nil")
	}

	// Verify error mentions the template doesn't exist
	if !strings.Contains(err.Error(), "does not exist") {
		t.Errorf("Error should mention template doesn't exist, got: %v", err)
	}
}

// TestGenerateTemplate_InvalidValuesFile tests error handling for invalid YAML
func TestGenerateTemplate_InvalidValuesFile(t *testing.T) {
	baseDir, cleanup := setupTestEnvironment(t)
	defer cleanup()

	// Valid template
	templatePath := filepath.Join(baseDir, "templates", "test.yaml.tmpl")
	templateContent := `name: {{.name}}`

	if err := os.WriteFile(templatePath, []byte(templateContent), 0644); err != nil {
		t.Fatalf("Failed to create template file: %v", err)
	}

	// Invalid YAML values file
	valuesPath := filepath.Join(baseDir, "bad-values.yaml")
	invalidYAML := `invalid: yaml: content:
  bad:: indentation`

	if err := os.WriteFile(valuesPath, []byte(invalidYAML), 0644); err != nil {
		t.Fatalf("Failed to create values file: %v", err)
	}

	outputPath := filepath.Join(baseDir, "output.yaml")

	// Execute with invalid values file - should fail
	err := generateTemplate("test.yaml", outputPath, []string{}, false, valuesPath)
	if err == nil {
		t.Error("Expected error for invalid YAML values file, got nil")
	}

	// Verify error mentions parsing failure
	if !strings.Contains(err.Error(), "parse") && !strings.Contains(err.Error(), "yaml") {
		t.Errorf("Error should mention YAML parsing failure, got: %v", err)
	}
}
