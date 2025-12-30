package bundle

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/viper"
)

// setupTestEnvironment creates a temporary directory with conjure config and bundle structure
func setupTestEnvironment(t *testing.T) (baseDir string, cleanup func()) {
	tmpDir := t.TempDir()

	// Create bundles directory
	bundlesDir := filepath.Join(tmpDir, "bundles")
	if err := os.MkdirAll(bundlesDir, 0755); err != nil {
		t.Fatalf("Failed to create bundles dir: %v", err)
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

// TestGenerateBundle_FullWorkflow tests complete bundle generation
func TestGenerateBundle_FullWorkflow(t *testing.T) {
	baseDir, cleanup := setupTestEnvironment(t)
	defer cleanup()

	// Create a bundle directory structure
	bundlePath := filepath.Join(baseDir, "bundles", "test-bundle")
	if err := os.MkdirAll(bundlePath, 0755); err != nil {
		t.Fatalf("Failed to create bundle dir: %v", err)
	}

	// Create bundle metadata file
	metadataPath := filepath.Join(bundlePath, "conjure.json")
	metadataContent := `{
  "bundle_type": "kubernetes",
  "bundle_name": "test-bundle",
  "bundle_description": "Test bundle for integration tests",
  "shared_variables": [
    {
      "name": "app_name",
      "description": "Application name",
      "required": true
    },
    {
      "name": "namespace",
      "description": "Kubernetes namespace",
      "default": "default"
    }
  ]
}`

	if err := os.WriteFile(metadataPath, []byte(metadataContent), 0644); err != nil {
		t.Fatalf("Failed to create metadata file: %v", err)
	}

	// Create template files in the bundle
	deploymentTemplate := filepath.Join(bundlePath, "deployment.yaml.tmpl")
	deploymentContent := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{.app_name}}-deployment
  namespace: {{.namespace}}`

	if err := os.WriteFile(deploymentTemplate, []byte(deploymentContent), 0644); err != nil {
		t.Fatalf("Failed to create deployment template: %v", err)
	}

	serviceTemplate := filepath.Join(bundlePath, "service.yaml.tmpl")
	serviceContent := `apiVersion: v1
kind: Service
metadata:
  name: {{.app_name}}-service
  namespace: {{.namespace}}`

	if err := os.WriteFile(serviceTemplate, []byte(serviceContent), 0644); err != nil {
		t.Fatalf("Failed to create service template: %v", err)
	}

	// Create output directory
	outputDir := filepath.Join(baseDir, "output")
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		t.Fatalf("Failed to create output dir: %v", err)
	}

	// Execute bundle generation
	varsList := []string{"app_name=myapp"}
	err := generateBundle("test-bundle", outputDir, varsList, false, "")
	if err != nil {
		t.Fatalf("generateBundle() failed: %v", err)
	}

	// Verify deployment.yaml was created
	deploymentOutput := filepath.Join(outputDir, "deployment.yaml")
	if _, err := os.Stat(deploymentOutput); os.IsNotExist(err) {
		t.Fatalf("deployment.yaml was not created")
	}

	// Verify deployment content
	deploymentBytes, err := os.ReadFile(deploymentOutput)
	if err != nil {
		t.Fatalf("Failed to read deployment.yaml: %v", err)
	}

	deploymentStr := string(deploymentBytes)
	if !strings.Contains(deploymentStr, "name: myapp-deployment") {
		t.Errorf("deployment.yaml should contain 'name: myapp-deployment', got:\n%s", deploymentStr)
	}
	if !strings.Contains(deploymentStr, "namespace: default") {
		t.Errorf("deployment.yaml should use default namespace, got:\n%s", deploymentStr)
	}

	// Verify service.yaml was created
	serviceOutput := filepath.Join(outputDir, "service.yaml")
	if _, err := os.Stat(serviceOutput); os.IsNotExist(err) {
		t.Fatalf("service.yaml was not created")
	}

	// Verify service content
	serviceBytes, err := os.ReadFile(serviceOutput)
	if err != nil {
		t.Fatalf("Failed to read service.yaml: %v", err)
	}

	serviceStr := string(serviceBytes)
	if !strings.Contains(serviceStr, "name: myapp-service") {
		t.Errorf("service.yaml should contain 'name: myapp-service', got:\n%s", serviceStr)
	}
}

// TestGenerateBundle_WithTemplateOverrides tests bundle generation with template-specific overrides
func TestGenerateBundle_WithTemplateOverrides(t *testing.T) {
	baseDir, cleanup := setupTestEnvironment(t)
	defer cleanup()

	// Create bundle
	bundlePath := filepath.Join(baseDir, "bundles", "override-test")
	if err := os.MkdirAll(bundlePath, 0755); err != nil {
		t.Fatalf("Failed to create bundle dir: %v", err)
	}

	// Simple metadata
	metadataPath := filepath.Join(bundlePath, "conjure.json")
	metadataContent := `{
  "bundle_type": "test",
  "bundle_name": "override-test",
  "bundle_description": "Test template overrides",
  "shared_variables": [
    {
      "name": "namespace",
      "default": "default"
    }
  ]
}`

	if err := os.WriteFile(metadataPath, []byte(metadataContent), 0644); err != nil {
		t.Fatalf("Failed to create metadata: %v", err)
	}

	// Create two templates
	template1 := filepath.Join(bundlePath, "app1.yaml.tmpl")
	template1Content := `namespace: {{.namespace}}`

	if err := os.WriteFile(template1, []byte(template1Content), 0644); err != nil {
		t.Fatalf("Failed to create template1: %v", err)
	}

	template2 := filepath.Join(bundlePath, "app2.yaml.tmpl")
	template2Content := `namespace: {{.namespace}}`

	if err := os.WriteFile(template2, []byte(template2Content), 0644); err != nil {
		t.Fatalf("Failed to create template2: %v", err)
	}

	outputDir := filepath.Join(baseDir, "output2")
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		t.Fatalf("Failed to create output dir: %v", err)
	}

	// Execute with template-specific overrides
	varsList := []string{
		"namespace=production",               // shared
		"app1.yaml.tmpl:namespace=app1-ns",  // override for app1
		"app2.yaml.tmpl:namespace=app2-ns",  // override for app2
	}
	err := generateBundle("override-test", outputDir, varsList, false, "")
	if err != nil {
		t.Fatalf("generateBundle() failed: %v", err)
	}

	// Verify app1 uses its override
	app1Output := filepath.Join(outputDir, "app1.yaml")
	app1Bytes, err := os.ReadFile(app1Output)
	if err != nil {
		t.Fatalf("Failed to read app1.yaml: %v", err)
	}

	if !strings.Contains(string(app1Bytes), "namespace: app1-ns") {
		t.Errorf("app1.yaml should use override 'app1-ns', got:\n%s", app1Bytes)
	}

	// Verify app2 uses its override
	app2Output := filepath.Join(outputDir, "app2.yaml")
	app2Bytes, err := os.ReadFile(app2Output)
	if err != nil {
		t.Fatalf("Failed to read app2.yaml: %v", err)
	}

	if !strings.Contains(string(app2Bytes), "namespace: app2-ns") {
		t.Errorf("app2.yaml should use override 'app2-ns', got:\n%s", app2Bytes)
	}
}

// TestGenerateBundle_WithValuesFile tests bundle generation with a values file
func TestGenerateBundle_WithValuesFile(t *testing.T) {
	baseDir, cleanup := setupTestEnvironment(t)
	defer cleanup()

	// Create bundle
	bundlePath := filepath.Join(baseDir, "bundles", "values-test")
	if err := os.MkdirAll(bundlePath, 0755); err != nil {
		t.Fatalf("Failed to create bundle dir: %v", err)
	}

	// Metadata
	metadataPath := filepath.Join(bundlePath, "conjure.json")
	metadataContent := `{
  "bundle_type": "test",
  "bundle_name": "values-test",
  "bundle_description": "Test values file",
  "shared_variables": []
}`

	if err := os.WriteFile(metadataPath, []byte(metadataContent), 0644); err != nil {
		t.Fatalf("Failed to create metadata: %v", err)
	}

	// Template
	template := filepath.Join(bundlePath, "config.yaml.tmpl")
	templateContent := `app: {{.app_name}}
env: {{.environment}}`

	if err := os.WriteFile(template, []byte(templateContent), 0644); err != nil {
		t.Fatalf("Failed to create template: %v", err)
	}

	// Create values file
	valuesPath := filepath.Join(baseDir, "bundle-values.yaml")
	valuesContent := `app_name: myapp
environment: staging`

	if err := os.WriteFile(valuesPath, []byte(valuesContent), 0644); err != nil {
		t.Fatalf("Failed to create values file: %v", err)
	}

	outputDir := filepath.Join(baseDir, "output3")
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		t.Fatalf("Failed to create output dir: %v", err)
	}

	// Execute with values file
	err := generateBundle("values-test", outputDir, []string{}, false, valuesPath)
	if err != nil {
		t.Fatalf("generateBundle() failed: %v", err)
	}

	// Verify output
	configOutput := filepath.Join(outputDir, "config.yaml")
	configBytes, err := os.ReadFile(configOutput)
	if err != nil {
		t.Fatalf("Failed to read config.yaml: %v", err)
	}

	configStr := string(configBytes)
	if !strings.Contains(configStr, "app: myapp") {
		t.Errorf("config.yaml should contain 'app: myapp', got:\n%s", configStr)
	}
	if !strings.Contains(configStr, "env: staging") {
		t.Errorf("config.yaml should contain 'env: staging', got:\n%s", configStr)
	}
}

// TestGenerateBundle_NonexistentBundle tests error handling for missing bundle
func TestGenerateBundle_NonexistentBundle(t *testing.T) {
	baseDir, cleanup := setupTestEnvironment(t)
	defer cleanup()

	outputDir := filepath.Join(baseDir, "output4")
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		t.Fatalf("Failed to create output dir: %v", err)
	}

	// Execute with nonexistent bundle
	err := generateBundle("nonexistent-bundle", outputDir, []string{}, false, "")
	if err == nil {
		t.Error("Expected error for nonexistent bundle, got nil")
	}

	// Verify error mentions bundle not found
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("Error should mention bundle not found, got: %v", err)
	}
}
