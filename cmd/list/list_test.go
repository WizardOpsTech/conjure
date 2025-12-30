package list

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/viper"
)

// setupTestEnvironment creates a temporary directory with conjure config
func setupTestEnvironment(t *testing.T) (baseDir string, cleanup func()) {
	tmpDir := t.TempDir()

	// Create directory structures
	templatesDir := filepath.Join(tmpDir, "templates")
	bundlesDir := filepath.Join(tmpDir, "bundles")

	if err := os.MkdirAll(templatesDir, 0755); err != nil {
		t.Fatalf("Failed to create templates dir: %v", err)
	}
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

// captureOutput captures stdout output from a function
func captureOutput(f func()) string {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	f()

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	return buf.String()
}

// =============================================================================
// Tests for isValidBundleType
// =============================================================================

func TestIsValidBundleType_Kubernetes(t *testing.T) {
	if !isValidBundleType("kubernetes") {
		t.Error("Expected 'kubernetes' to be valid")
	}
}

func TestIsValidBundleType_Terraform(t *testing.T) {
	if !isValidBundleType("terraform") {
		t.Error("Expected 'terraform' to be valid")
	}
}

func TestIsValidBundleType_Invalid(t *testing.T) {
	if isValidBundleType("invalid") {
		t.Error("Expected 'invalid' to be invalid")
	}
	if isValidBundleType("docker") {
		t.Error("Expected 'docker' to be invalid")
	}
}

// =============================================================================
// Tests for isValidTemplateType
// =============================================================================

func TestIsValidTemplateType_YAML(t *testing.T) {
	if !isValidTemplateType(".yaml") {
		t.Error("Expected '.yaml' to be valid")
	}
	if !isValidTemplateType("yaml") {
		t.Error("Expected 'yaml' (without dot) to be valid")
	}
}

func TestIsValidTemplateType_TF(t *testing.T) {
	if !isValidTemplateType(".tf") {
		t.Error("Expected '.tf' to be valid")
	}
	if !isValidTemplateType("tf") {
		t.Error("Expected 'tf' (without dot) to be valid")
	}
}

func TestIsValidTemplateType_JSON(t *testing.T) {
	if !isValidTemplateType(".json") {
		t.Error("Expected '.json' to be valid")
	}
}

func TestIsValidTemplateType_Invalid(t *testing.T) {
	if isValidTemplateType(".txt") {
		t.Error("Expected '.txt' to be invalid")
	}
	if isValidTemplateType(".xml") {
		t.Error("Expected '.xml' to be invalid")
	}
}

// =============================================================================
// Tests for listTemplates
// =============================================================================

func TestListTemplates_EmptyDirectory(t *testing.T) {
	_, cleanup := setupTestEnvironment(t)
	defer cleanup()

	// No templates created - directory is empty

	output := captureOutput(func() {
		listTemplates("")
	})

	if !strings.Contains(output, "No templates found") {
		t.Errorf("Expected 'No templates found', got:\n%s", output)
	}
}

func TestListTemplates_WithTemplates(t *testing.T) {
	baseDir, cleanup := setupTestEnvironment(t)
	defer cleanup()

	// Create template files
	templatesPath := filepath.Join(baseDir, "templates")
	templates := []string{
		"deployment.yaml.tmpl",
		"service.yaml.tmpl",
		"vsphere_vm.tf.tmpl",
		"config.json.tmpl",
	}

	for _, tmpl := range templates {
		path := filepath.Join(templatesPath, tmpl)
		if err := os.WriteFile(path, []byte("test content"), 0644); err != nil {
			t.Fatalf("Failed to create template: %v", err)
		}
	}

	output := captureOutput(func() {
		listTemplates("")
	})

	// Verify all templates are listed (without .tmpl extension)
	if !strings.Contains(output, "deployment.yaml") {
		t.Errorf("Output should contain 'deployment.yaml', got:\n%s", output)
	}
	if !strings.Contains(output, "service.yaml") {
		t.Errorf("Output should contain 'service.yaml', got:\n%s", output)
	}
	if !strings.Contains(output, "vsphere_vm.tf") {
		t.Errorf("Output should contain 'vsphere_vm.tf', got:\n%s", output)
	}
	if !strings.Contains(output, "config.json") {
		t.Errorf("Output should contain 'config.json', got:\n%s", output)
	}

	// Verify .tmpl extension is NOT shown
	if strings.Contains(output, ".tmpl") {
		t.Errorf("Output should not contain '.tmpl' extension, got:\n%s", output)
	}
}

func TestListTemplates_WithTypeFilter_YAML(t *testing.T) {
	baseDir, cleanup := setupTestEnvironment(t)
	defer cleanup()

	// Create mixed template files
	templatesPath := filepath.Join(baseDir, "templates")
	templates := map[string]bool{
		"deployment.yaml.tmpl":  true,  // should match
		"service.yaml.tmpl":     true,  // should match
		"vsphere_vm.tf.tmpl":    false, // should NOT match
		"config.json.tmpl":      false, // should NOT match
	}

	for tmpl := range templates {
		path := filepath.Join(templatesPath, tmpl)
		if err := os.WriteFile(path, []byte("test"), 0644); err != nil {
			t.Fatalf("Failed to create template: %v", err)
		}
	}

	output := captureOutput(func() {
		listTemplates(".yaml")
	})

	// Verify only YAML templates are listed
	if !strings.Contains(output, "deployment.yaml") {
		t.Errorf("Output should contain 'deployment.yaml', got:\n%s", output)
	}
	if !strings.Contains(output, "service.yaml") {
		t.Errorf("Output should contain 'service.yaml', got:\n%s", output)
	}

	// Verify non-YAML templates are NOT listed
	if strings.Contains(output, "vsphere_vm.tf") {
		t.Errorf("Output should NOT contain 'vsphere_vm.tf', got:\n%s", output)
	}
	if strings.Contains(output, "config.json") {
		t.Errorf("Output should NOT contain 'config.json', got:\n%s", output)
	}
}

func TestListTemplates_WithTypeFilter_TF(t *testing.T) {
	baseDir, cleanup := setupTestEnvironment(t)
	defer cleanup()

	templatesPath := filepath.Join(baseDir, "templates")
	_ = os.WriteFile(filepath.Join(templatesPath, "main.tf.tmpl"), []byte("test"), 0644)
	_ = os.WriteFile(filepath.Join(templatesPath, "deployment.yaml.tmpl"), []byte("test"), 0644)

	output := captureOutput(func() {
		listTemplates("tf") // without dot prefix
	})

	// Should show .tf templates
	if !strings.Contains(output, "main.tf") {
		t.Errorf("Output should contain 'main.tf', got:\n%s", output)
	}

	// Should NOT show .yaml templates
	if strings.Contains(output, "deployment.yaml") {
		t.Errorf("Output should NOT contain 'deployment.yaml', got:\n%s", output)
	}
}

func TestListTemplates_NoMatchesForFilter(t *testing.T) {
	baseDir, cleanup := setupTestEnvironment(t)
	defer cleanup()

	// Create only YAML templates
	templatesPath := filepath.Join(baseDir, "templates")
	_ = os.WriteFile(filepath.Join(templatesPath, "deployment.yaml.tmpl"), []byte("test"), 0644)

	output := captureOutput(func() {
		listTemplates(".tf") // Filter for TF, but none exist
	})

	if !strings.Contains(output, "No templates found with extension '.tf'") {
		t.Errorf("Expected message about no templates found, got:\n%s", output)
	}
}

func TestListTemplates_IgnoresNonTmplFiles(t *testing.T) {
	baseDir, cleanup := setupTestEnvironment(t)
	defer cleanup()

	templatesPath := filepath.Join(baseDir, "templates")

	// Create .tmpl file and non-.tmpl file
	_ = os.WriteFile(filepath.Join(templatesPath, "valid.yaml.tmpl"), []byte("test"), 0644)
	_ = os.WriteFile(filepath.Join(templatesPath, "invalid.yaml"), []byte("test"), 0644)
	_ = os.WriteFile(filepath.Join(templatesPath, "README.md"), []byte("test"), 0644)

	output := captureOutput(func() {
		listTemplates("")
	})

	// Should show .tmpl file
	if !strings.Contains(output, "valid.yaml") {
		t.Errorf("Output should contain 'valid.yaml', got:\n%s", output)
	}

	// Should NOT show non-.tmpl files
	if strings.Contains(output, "invalid.yaml") {
		t.Errorf("Output should NOT contain 'invalid.yaml' (not a .tmpl file), got:\n%s", output)
	}
	if strings.Contains(output, "README.md") {
		t.Errorf("Output should NOT contain 'README.md', got:\n%s", output)
	}
}

// =============================================================================
// Tests for listBundles
// =============================================================================

func TestListBundles_EmptyDirectory(t *testing.T) {
	_, cleanup := setupTestEnvironment(t)
	defer cleanup()

	// No bundles created - directory is empty

	output := captureOutput(func() {
		listBundles("")
	})

	if !strings.Contains(output, "No bundles found") {
		t.Errorf("Expected 'No bundles found', got:\n%s", output)
	}
}

func TestListBundles_WithBundles(t *testing.T) {
	baseDir, cleanup := setupTestEnvironment(t)
	defer cleanup()

	// Create bundle directories with metadata
	bundlesPath := filepath.Join(baseDir, "bundles")

	// Bundle 1: Kubernetes
	bundle1Path := filepath.Join(bundlesPath, "k8s-app")
	_ = os.MkdirAll(bundle1Path, 0755)
	metadata1 := BundleMetadata{
		BundleType:        "kubernetes",
		BundleName:        "k8s-app",
		BundleDescription: "Kubernetes application bundle",
	}
	data1, _ := json.Marshal(metadata1)
	_ = os.WriteFile(filepath.Join(bundle1Path, "conjure.json"), data1, 0644)

	// Bundle 2: Terraform
	bundle2Path := filepath.Join(bundlesPath, "tf-infra")
	_ = os.MkdirAll(bundle2Path, 0755)
	metadata2 := BundleMetadata{
		BundleType:        "terraform",
		BundleName:        "tf-infra",
		BundleDescription: "Terraform infrastructure bundle",
	}
	data2, _ := json.Marshal(metadata2)
	_ = os.WriteFile(filepath.Join(bundle2Path, "conjure.json"), data2, 0644)

	output := captureOutput(func() {
		listBundles("")
	})

	// Verify both bundles are listed
	if !strings.Contains(output, "k8s-app") {
		t.Errorf("Output should contain 'k8s-app', got:\n%s", output)
	}
	if !strings.Contains(output, "tf-infra") {
		t.Errorf("Output should contain 'tf-infra', got:\n%s", output)
	}
	if !strings.Contains(output, "Kubernetes application bundle") {
		t.Errorf("Output should contain bundle description, got:\n%s", output)
	}
}

func TestListBundles_WithTypeFilter_Kubernetes(t *testing.T) {
	baseDir, cleanup := setupTestEnvironment(t)
	defer cleanup()

	bundlesPath := filepath.Join(baseDir, "bundles")

	// Create Kubernetes bundle
	k8sPath := filepath.Join(bundlesPath, "k8s-app")
	_ = os.MkdirAll(k8sPath, 0755)
	k8sMeta := BundleMetadata{
		BundleType:        "kubernetes",
		BundleName:        "k8s-app",
		BundleDescription: "K8s bundle",
	}
	k8sData, _ := json.Marshal(k8sMeta)
	_ = os.WriteFile(filepath.Join(k8sPath, "conjure.json"), k8sData, 0644)

	// Create Terraform bundle
	tfPath := filepath.Join(bundlesPath, "tf-infra")
	_ = os.MkdirAll(tfPath, 0755)
	tfMeta := BundleMetadata{
		BundleType:        "terraform",
		BundleName:        "tf-infra",
		BundleDescription: "TF bundle",
	}
	tfData, _ := json.Marshal(tfMeta)
	_ = os.WriteFile(filepath.Join(tfPath, "conjure.json"), tfData, 0644)

	output := captureOutput(func() {
		listBundles("kubernetes")
	})

	// Should show Kubernetes bundle
	if !strings.Contains(output, "k8s-app") {
		t.Errorf("Output should contain 'k8s-app', got:\n%s", output)
	}

	// Should NOT show Terraform bundle
	if strings.Contains(output, "tf-infra") {
		t.Errorf("Output should NOT contain 'tf-infra', got:\n%s", output)
	}
}

func TestListBundles_NoMatchesForFilter(t *testing.T) {
	baseDir, cleanup := setupTestEnvironment(t)
	defer cleanup()

	bundlesPath := filepath.Join(baseDir, "bundles")

	// Create only Kubernetes bundle
	k8sPath := filepath.Join(bundlesPath, "k8s-app")
	_ = os.MkdirAll(k8sPath, 0755)
	metadata := BundleMetadata{
		BundleType:        "kubernetes",
		BundleName:        "k8s-app",
		BundleDescription: "K8s bundle",
	}
	data, _ := json.Marshal(metadata)
	_ = os.WriteFile(filepath.Join(k8sPath, "conjure.json"), data, 0644)

	output := captureOutput(func() {
		listBundles("terraform") // Filter for terraform, but none exist
	})

	if !strings.Contains(output, "No bundles found with type 'terraform'") {
		t.Errorf("Expected message about no bundles found, got:\n%s", output)
	}
}

func TestListBundles_SkipsDirectoriesWithoutMetadata(t *testing.T) {
	baseDir, cleanup := setupTestEnvironment(t)
	defer cleanup()

	bundlesPath := filepath.Join(baseDir, "bundles")

	// Create directory without conjure.json
	invalidPath := filepath.Join(bundlesPath, "invalid-bundle")
	_ = os.MkdirAll(invalidPath, 0755)

	// Create valid bundle
	validPath := filepath.Join(bundlesPath, "valid-bundle")
	_ = os.MkdirAll(validPath, 0755)
	metadata := BundleMetadata{
		BundleType:        "kubernetes",
		BundleName:        "valid-bundle",
		BundleDescription: "Valid",
	}
	data, _ := json.Marshal(metadata)
	_ = os.WriteFile(filepath.Join(validPath, "conjure.json"), data, 0644)

	output := captureOutput(func() {
		listBundles("")
	})

	// Should show valid bundle
	if !strings.Contains(output, "valid-bundle") {
		t.Errorf("Output should contain 'valid-bundle', got:\n%s", output)
	}

	// Should mention invalid bundle
	if !strings.Contains(output, "invalid-bundle") {
		t.Errorf("Output should mention 'invalid-bundle' as having no metadata, got:\n%s", output)
	}
}

func TestListBundles_SkipsFiles(t *testing.T) {
	baseDir, cleanup := setupTestEnvironment(t)
	defer cleanup()

	bundlesPath := filepath.Join(baseDir, "bundles")

	// Create a file (not a directory) in bundles directory
	_ = os.WriteFile(filepath.Join(bundlesPath, "README.md"), []byte("test"), 0644)

	// Create a valid bundle
	validPath := filepath.Join(bundlesPath, "valid-bundle")
	_ = os.MkdirAll(validPath, 0755)
	metadata := BundleMetadata{
		BundleType:        "kubernetes",
		BundleName:        "valid-bundle",
		BundleDescription: "Valid",
	}
	data, _ := json.Marshal(metadata)
	_ = os.WriteFile(filepath.Join(validPath, "conjure.json"), data, 0644)

	output := captureOutput(func() {
		listBundles("")
	})

	// Should show valid bundle
	if !strings.Contains(output, "valid-bundle") {
		t.Errorf("Output should contain 'valid-bundle', got:\n%s", output)
	}

	// Should NOT show file
	if strings.Contains(output, "README.md") {
		t.Errorf("Output should NOT contain 'README.md' (it's a file, not a bundle), got:\n%s", output)
	}
}
