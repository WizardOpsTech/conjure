package metadata

import (
	"os"
	"path/filepath"
	"testing"
)

func TestMergeVariablesForTemplate_SharedDefaults(t *testing.T) {
	bundleMeta := &BundleMetadata{
		SchemaVersion: "v1",
		SharedVariables: []Variable{
			{Name: "namespace", Description: "Namespace", Type: "string", Default: "default"},
			{Name: "app_name", Description: "App name", Type: "string", Default: "myapp"},
		},
		TemplateVariables: map[string][]Variable{},
	}

	userVars := make(map[string]interface{})

	result, err := MergeVariablesForTemplate(bundleMeta, "test.yaml.tmpl", userVars)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if result["namespace"] != "default" {
		t.Errorf("Expected namespace='default', got: %v", result["namespace"])
	}

	if result["app_name"] != "myapp" {
		t.Errorf("Expected app_name='myapp', got: %v", result["app_name"])
	}
}

func TestMergeVariablesForTemplate_UserVarsOverrideDefaults(t *testing.T) {
	bundleMeta := &BundleMetadata{
		SchemaVersion: "v1",
		SharedVariables: []Variable{
			{Name: "namespace", Description: "Namespace", Type: "string", Default: "default"},
		},
		TemplateVariables: map[string][]Variable{},
	}

	userVars := map[string]interface{}{
		"namespace": "production",
	}

	result, err := MergeVariablesForTemplate(bundleMeta, "test.yaml.tmpl", userVars)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if result["namespace"] != "production" {
		t.Errorf("Expected namespace='production', got: %v", result["namespace"])
	}
}

func TestMergeVariablesForTemplate_TemplateSpecificOverride(t *testing.T) {
	bundleMeta := &BundleMetadata{
		SchemaVersion: "v1",
		SharedVariables: []Variable{
			{Name: "namespace", Description: "Namespace", Type: "string", Default: "default"},
		},
		TemplateVariables: map[string][]Variable{
			"special.yaml.tmpl": {
				{Name: "namespace", Description: "Special namespace", Type: "string", Default: "special-namespace"},
			},
		},
	}

	userVars := make(map[string]interface{})

	// For regular template, should get shared default
	result, err := MergeVariablesForTemplate(bundleMeta, "regular.yaml.tmpl", userVars)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if result["namespace"] != "default" {
		t.Errorf("Expected namespace='default' for regular template, got: %v", result["namespace"])
	}

	// For special template, should get template-specific default
	result, err = MergeVariablesForTemplate(bundleMeta, "special.yaml.tmpl", userVars)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if result["namespace"] != "special-namespace" {
		t.Errorf("Expected namespace='special-namespace' for special template, got: %v", result["namespace"])
	}
}

func TestMergeVariablesForTemplate_UserVarsOverrideTemplateDefaults(t *testing.T) {
	bundleMeta := &BundleMetadata{
		SchemaVersion: "v1",
		SharedVariables: []Variable{
			{Name: "namespace", Description: "Namespace", Type: "string", Default: "default"},
		},
		TemplateVariables: map[string][]Variable{
			"special.yaml.tmpl": {
				{Name: "namespace", Description: "Special namespace", Type: "string", Default: "special-namespace"},
			},
		},
	}

	userVars := map[string]interface{}{
		"namespace": "user-override",
	}

	// User vars should override both shared and template-specific defaults
	result, err := MergeVariablesForTemplate(bundleMeta, "special.yaml.tmpl", userVars)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if result["namespace"] != "user-override" {
		t.Errorf("Expected namespace='user-override', got: %v", result["namespace"])
	}
}

func TestMergeVariablesForTemplate_RequiredSharedVariable(t *testing.T) {
	bundleMeta := &BundleMetadata{
		SchemaVersion: "v1",
		SharedVariables: []Variable{
			{Name: "app_name", Description: "App name", Type: "string", Default: ""},
		},
		TemplateVariables: map[string][]Variable{},
	}

	userVars := make(map[string]interface{})

	_, err := MergeVariablesForTemplate(bundleMeta, "test.yaml.tmpl", userVars)
	if err == nil {
		t.Error("Expected error for missing required variable, got nil")
	}
}

func TestMergeVariablesForTemplate_RequiredTemplateVariable(t *testing.T) {
	bundleMeta := &BundleMetadata{
		SchemaVersion:   "v1",
		SharedVariables: []Variable{},
		TemplateVariables: map[string][]Variable{
			"test.yaml.tmpl": {
				{Name: "replicas", Description: "Replica count", Type: "int", Default: ""},
			},
		},
	}

	userVars := make(map[string]interface{})

	_, err := MergeVariablesForTemplate(bundleMeta, "test.yaml.tmpl", userVars)
	if err == nil {
		t.Error("Expected error for missing required template variable, got nil")
	}
}

func TestMergeVariablesForTemplate_RequiredVariableProvided(t *testing.T) {
	bundleMeta := &BundleMetadata{
		SchemaVersion: "v1",
		SharedVariables: []Variable{
			{Name: "app_name", Description: "App name", Type: "string", Default: ""},
		},
		TemplateVariables: map[string][]Variable{},
	}

	userVars := map[string]interface{}{
		"app_name": "testapp",
	}

	result, err := MergeVariablesForTemplate(bundleMeta, "test.yaml.tmpl", userVars)
	if err != nil {
		t.Fatalf("Expected no error when required variable provided, got: %v", err)
	}

	if result["app_name"] != "testapp" {
		t.Errorf("Expected app_name='testapp', got: %v", result["app_name"])
	}
}

func TestGetAllVariablesForBundle(t *testing.T) {
	bundleMeta := &BundleMetadata{
		SchemaVersion: "v1",
		SharedVariables: []Variable{
			{Name: "namespace", Description: "Namespace", Type: "string", Default: "default"},
			{Name: "app_name", Description: "App name", Type: "string", Default: ""},
		},
		TemplateVariables: map[string][]Variable{
			"deployment.yaml.tmpl": {
				{Name: "replicas", Description: "Replicas", Type: "int", Default: "3"},
			},
			"service.yaml.tmpl": {
				{Name: "service_port", Description: "Service port", Type: "int", Default: "80"},
			},
		},
	}

	allVars := GetAllVariablesForBundle(bundleMeta)

	// Should have 4 unique variables
	if len(allVars) != 4 {
		t.Errorf("Expected 4 variables, got: %d", len(allVars))
	}

	// Check that all expected variables are present
	varNames := make(map[string]bool)
	for _, v := range allVars {
		varNames[v.Name] = true
	}

	expectedVars := []string{"namespace", "app_name", "replicas", "service_port"}
	for _, expected := range expectedVars {
		if !varNames[expected] {
			t.Errorf("Expected variable '%s' not found in result", expected)
		}
	}
}

func TestGetAllVariablesForBundle_NoDuplicates(t *testing.T) {
	bundleMeta := &BundleMetadata{
		SchemaVersion: "v1",
		SharedVariables: []Variable{
			{Name: "namespace", Description: "Namespace", Type: "string", Default: "default"},
		},
		TemplateVariables: map[string][]Variable{
			"deployment.yaml.tmpl": {
				{Name: "namespace", Description: "Override namespace", Type: "string", Default: "override"}, // Same name as shared
			},
		},
	}

	allVars := GetAllVariablesForBundle(bundleMeta)

	// Should only have 1 variable (namespace), no duplicates
	if len(allVars) != 1 {
		t.Errorf("Expected 1 unique variable, got: %d", len(allVars))
	}

	if allVars[0].Name != "namespace" {
		t.Errorf("Expected variable 'namespace', got: %s", allVars[0].Name)
	}
}

// Tests for schema_version validation
func TestValidateTemplateMetadata_MissingVersion(t *testing.T) {
	metadata := &TemplateMetadata{
		SchemaVersion:       "",
		TemplateName:        "test",
		TemplateDescription: "Test template",
		TemplateType:        "yaml",
		Variables:           []Variable{},
	}

	err := validateTemplateMetadata(metadata)
	if err == nil {
		t.Error("Expected error for missing schema_version, got nil")
	}
}

func TestValidateTemplateMetadata_UnsupportedVersion(t *testing.T) {
	metadata := &TemplateMetadata{
		SchemaVersion:       "v2",
		TemplateName:        "test",
		TemplateDescription: "Test template",
		TemplateType:        "yaml",
		Variables:           []Variable{},
	}

	err := validateTemplateMetadata(metadata)
	if err == nil {
		t.Error("Expected error for unsupported schema_version, got nil")
	}
}

func TestValidateTemplateMetadata_ValidVersion(t *testing.T) {
	metadata := &TemplateMetadata{
		SchemaVersion:       "v1",
		TemplateName:        "test",
		TemplateDescription: "Test template",
		TemplateType:        "yaml",
		Variables:           []Variable{},
	}

	err := validateTemplateMetadata(metadata)
	if err != nil {
		t.Errorf("Expected no error for valid schema_version, got: %v", err)
	}
}

// Tests for template_type validation
func TestValidateTemplateMetadata_MissingTemplateType(t *testing.T) {
	metadata := &TemplateMetadata{
		SchemaVersion:       "v1",
		TemplateName:        "test",
		TemplateDescription: "Test template",
		TemplateType:        "",
		Variables:           []Variable{},
	}

	err := validateTemplateMetadata(metadata)
	if err == nil {
		t.Error("Expected error for missing template_type, got nil")
	}
}

// Tests for template_name validation
func TestValidateTemplateMetadata_MissingTemplateName(t *testing.T) {
	metadata := &TemplateMetadata{
		SchemaVersion:       "v1",
		TemplateName:        "",
		TemplateDescription: "Test template",
		TemplateType:        "yaml",
		Variables:           []Variable{},
	}

	err := validateTemplateMetadata(metadata)
	if err == nil {
		t.Error("Expected error for missing template_name, got nil")
	}
}

func TestValidateTemplateMetadata_ValidTemplateName(t *testing.T) {
	metadata := &TemplateMetadata{
		SchemaVersion:       "v1",
		TemplateName:        "my-template",
		TemplateDescription: "Test template",
		TemplateType:        "yaml",
		Variables:           []Variable{},
	}

	err := validateTemplateMetadata(metadata)
	if err != nil {
		t.Errorf("Expected no error for valid template_name, got: %v", err)
	}
}

// Tests for template_description validation
func TestValidateTemplateMetadata_MissingTemplateDescription(t *testing.T) {
	metadata := &TemplateMetadata{
		SchemaVersion:       "v1",
		TemplateName:        "test",
		TemplateDescription: "",
		TemplateType:        "yaml",
		Variables:           []Variable{},
	}

	err := validateTemplateMetadata(metadata)
	if err == nil {
		t.Error("Expected error for missing template_description, got nil")
	}
}

func TestValidateTemplateMetadata_ValidTemplateDescription(t *testing.T) {
	metadata := &TemplateMetadata{
		SchemaVersion:       "v1",
		TemplateName:        "test",
		TemplateDescription: "This is a test template",
		TemplateType:        "yaml",
		Variables:           []Variable{},
	}

	err := validateTemplateMetadata(metadata)
	if err != nil {
		t.Errorf("Expected no error for valid template_description, got: %v", err)
	}
}

// Tests for variables field - empty array should be valid
func TestValidateTemplateMetadata_EmptyVariablesArray(t *testing.T) {
	metadata := &TemplateMetadata{
		SchemaVersion:       "v1",
		TemplateName:        "test",
		TemplateDescription: "Test template",
		TemplateType:        "yaml",
		Variables:           []Variable{},
	}

	err := validateTemplateMetadata(metadata)
	if err != nil {
		t.Errorf("Expected no error for empty variables array, got: %v", err)
	}
}

func TestValidateTemplateMetadata_NilVariablesArray(t *testing.T) {
	metadata := &TemplateMetadata{
		SchemaVersion:       "v1",
		TemplateName:        "test",
		TemplateDescription: "Test template",
		TemplateType:        "yaml",
		Variables:           nil,
	}

	err := validateTemplateMetadata(metadata)
	if err != nil {
		t.Errorf("Expected no error for nil variables array, got: %v", err)
	}
}

// Tests for description requirement
func TestValidateVariable_MissingDescription(t *testing.T) {
	variable := &Variable{
		Name:        "test_var",
		Description: "",
		Type:        "string",
		Default:     "default",
	}

	err := validateVariable(variable)
	if err == nil {
		t.Error("Expected error for missing description, got nil")
	}
}

func TestValidateVariable_WithDescription(t *testing.T) {
	variable := &Variable{
		Name:        "test_var",
		Description: "A test variable",
		Type:        "string",
		Default:     "default",
	}

	err := validateVariable(variable)
	if err != nil {
		t.Errorf("Expected no error for valid variable, got: %v", err)
	}
}

// Tests for type requirement and validation
func TestValidateVariable_MissingType(t *testing.T) {
	variable := &Variable{
		Name:        "test_var",
		Description: "A test variable",
		Type:        "",
		Default:     "default",
	}

	err := validateVariable(variable)
	if err == nil {
		t.Error("Expected error for missing type, got nil")
	}
}

func TestValidateVariable_InvalidType(t *testing.T) {
	variable := &Variable{
		Name:        "test_var",
		Description: "A test variable",
		Type:        "float",
		Default:     "1.5",
	}

	err := validateVariable(variable)
	if err == nil {
		t.Error("Expected error for invalid type, got nil")
	}
}

// Tests for type validation
func TestValidateValueType_String(t *testing.T) {
	tests := []struct {
		value    interface{}
		expected error
	}{
		{"hello", nil},
		{"123", nil},
		{"true", nil},
		{"", nil},
	}

	for _, test := range tests {
		err := validateValueType(test.value, "string")
		if (err == nil) != (test.expected == nil) {
			t.Errorf("validateValueType(%v, 'string') returned %v, expected %v", test.value, err, test.expected)
		}
	}
}

func TestValidateValueType_Int(t *testing.T) {
	tests := []struct {
		value       interface{}
		shouldError bool
	}{
		{"123", false},
		{"0", false},
		{"-456", false},
		{"abc", true},
		{"12.5", true},
		{"true", true},
	}

	for _, test := range tests {
		err := validateValueType(test.value, "int")
		if (err != nil) != test.shouldError {
			t.Errorf("validateValueType(%v, 'int') error = %v, shouldError = %v", test.value, err, test.shouldError)
		}
	}
}

func TestValidateValueType_Bool(t *testing.T) {
	tests := []struct {
		value       interface{}
		shouldError bool
	}{
		{"true", false},
		{"false", false},
		{"True", true},
		{"FALSE", true},
		{"1", true},
		{"0", true},
		{"yes", true},
	}

	for _, test := range tests {
		err := validateValueType(test.value, "bool")
		if (err != nil) != test.shouldError {
			t.Errorf("validateValueType(%v, 'bool') error = %v, shouldError = %v", test.value, err, test.shouldError)
		}
	}
}

// Tests for default value validation
func TestValidateVariable_InvalidDefaultType(t *testing.T) {
	variable := &Variable{
		Name:        "test_var",
		Description: "A test variable",
		Type:        "int",
		Default:     "not_a_number",
	}

	err := validateVariable(variable)
	if err == nil {
		t.Error("Expected error for invalid default value type, got nil")
	}
}

func TestValidateVariable_ValidDefaultType(t *testing.T) {
	tests := []struct {
		varType     string
		defaultVal  string
		shouldError bool
	}{
		{"string", "hello", false},
		{"int", "123", false},
		{"int", "abc", true},
		{"bool", "true", false},
		{"bool", "yes", true},
	}

	for _, test := range tests {
		variable := &Variable{
			Name:        "test_var",
			Description: "A test variable",
			Type:        test.varType,
			Default:     test.defaultVal,
		}

		err := validateVariable(variable)
		if (err != nil) != test.shouldError {
			t.Errorf("validateVariable with type=%s, default=%s: error = %v, shouldError = %v",
				test.varType, test.defaultVal, err, test.shouldError)
		}
	}
}

// Tests for ValidateAndMergeVariables with type checking
func TestValidateAndMergeVariables_TypeValidation(t *testing.T) {
	metadata := &TemplateMetadata{
		SchemaVersion:       "v1",
		TemplateName:        "test",
		TemplateDescription: "Test template",
		TemplateType:        "yaml",
		Variables: []Variable{
			{
				Name:        "count",
				Description: "Number of items",
				Type:        "int",
				Default:     "",
			},
		},
	}

	// Valid integer value
	userVars := map[string]interface{}{
		"count": "42",
	}
	_, err := ValidateAndMergeVariables(metadata, userVars)
	if err != nil {
		t.Errorf("Expected no error for valid integer, got: %v", err)
	}

	// Invalid integer value
	userVars = map[string]interface{}{
		"count": "not_a_number",
	}
	_, err = ValidateAndMergeVariables(metadata, userVars)
	if err == nil {
		t.Error("Expected error for invalid integer, got nil")
	}
}

// Tests for bundle metadata validation
func TestValidateBundleMetadata_MissingVersion(t *testing.T) {
	metadata := &BundleMetadata{
		SchemaVersion:     "",
		BundleType:        "k8s",
		BundleName:        "test",
		BundleDescription: "Test bundle",
		SharedVariables:   []Variable{},
	}

	err := validateBundleMetadata(metadata)
	if err == nil {
		t.Error("Expected error for missing schema_version, got nil")
	}
}

// Tests for bundle_name validation
func TestValidateBundleMetadata_MissingBundleName(t *testing.T) {
	metadata := &BundleMetadata{
		SchemaVersion:     "v1",
		BundleType:        "k8s",
		BundleName:        "",
		BundleDescription: "Test bundle",
		SharedVariables:   []Variable{},
	}

	err := validateBundleMetadata(metadata)
	if err == nil {
		t.Error("Expected error for missing bundle_name, got nil")
	}
}

func TestValidateBundleMetadata_ValidBundleName(t *testing.T) {
	metadata := &BundleMetadata{
		SchemaVersion:     "v1",
		BundleType:        "k8s",
		BundleName:        "my-bundle",
		BundleDescription: "Test bundle",
		SharedVariables:   []Variable{},
	}

	err := validateBundleMetadata(metadata)
	if err != nil {
		t.Errorf("Expected no error for valid bundle_name, got: %v", err)
	}
}

// Tests for bundle_type validation
func TestValidateBundleMetadata_MissingBundleType(t *testing.T) {
	metadata := &BundleMetadata{
		SchemaVersion:     "v1",
		BundleType:        "",
		BundleName:        "test",
		BundleDescription: "Test bundle",
		SharedVariables:   []Variable{},
	}

	err := validateBundleMetadata(metadata)
	if err == nil {
		t.Error("Expected error for missing bundle_type, got nil")
	}
}

func TestValidateBundleMetadata_ValidBundleType(t *testing.T) {
	metadata := &BundleMetadata{
		SchemaVersion:     "v1",
		BundleType:        "terraform",
		BundleName:        "test",
		BundleDescription: "Test bundle",
		SharedVariables:   []Variable{},
	}

	err := validateBundleMetadata(metadata)
	if err != nil {
		t.Errorf("Expected no error for valid bundle_type, got: %v", err)
	}
}

// Tests for bundle_description validation
func TestValidateBundleMetadata_MissingBundleDescription(t *testing.T) {
	metadata := &BundleMetadata{
		SchemaVersion:     "v1",
		BundleType:        "k8s",
		BundleName:        "test",
		BundleDescription: "",
		SharedVariables:   []Variable{},
	}

	err := validateBundleMetadata(metadata)
	if err == nil {
		t.Error("Expected error for missing bundle_description, got nil")
	}
}

func TestValidateBundleMetadata_ValidBundleDescription(t *testing.T) {
	metadata := &BundleMetadata{
		SchemaVersion:     "v1",
		BundleType:        "k8s",
		BundleName:        "test",
		BundleDescription: "This is a test bundle",
		SharedVariables:   []Variable{},
	}

	err := validateBundleMetadata(metadata)
	if err != nil {
		t.Errorf("Expected no error for valid bundle_description, got: %v", err)
	}
}

// Tests for bundle variables - various empty/nil combinations should be valid
func TestValidateBundleMetadata_EmptySharedVariablesWithTemplateVariables(t *testing.T) {
	metadata := &BundleMetadata{
		SchemaVersion:     "v1",
		BundleType:        "k8s",
		BundleName:        "test",
		BundleDescription: "Test bundle",
		SharedVariables:   []Variable{},
		TemplateVariables: map[string][]Variable{
			"test.yaml.tmpl": {
				{
					Name:        "test_var",
					Description: "Test variable",
					Type:        "string",
					Default:     "default",
				},
			},
		},
	}

	err := validateBundleMetadata(metadata)
	if err != nil {
		t.Errorf("Expected no error for empty shared_variables with template_variables, got: %v", err)
	}
}

func TestValidateBundleMetadata_SharedVariablesWithEmptyTemplateVariables(t *testing.T) {
	metadata := &BundleMetadata{
		SchemaVersion:     "v1",
		BundleType:        "k8s",
		BundleName:        "test",
		BundleDescription: "Test bundle",
		SharedVariables: []Variable{
			{
				Name:        "shared_var",
				Description: "Shared variable",
				Type:        "string",
				Default:     "default",
			},
		},
		TemplateVariables: map[string][]Variable{},
	}

	err := validateBundleMetadata(metadata)
	if err != nil {
		t.Errorf("Expected no error for shared_variables with empty template_variables, got: %v", err)
	}
}

func TestValidateBundleMetadata_BothVariablesEmpty(t *testing.T) {
	metadata := &BundleMetadata{
		SchemaVersion:     "v1",
		BundleType:        "k8s",
		BundleName:        "test",
		BundleDescription: "Test bundle",
		SharedVariables:   []Variable{},
		TemplateVariables: map[string][]Variable{},
	}

	err := validateBundleMetadata(metadata)
	if err != nil {
		t.Errorf("Expected no error for both variables empty, got: %v", err)
	}
}

func TestValidateBundleMetadata_NilSharedVariables(t *testing.T) {
	metadata := &BundleMetadata{
		SchemaVersion:     "v1",
		BundleType:        "k8s",
		BundleName:        "test",
		BundleDescription: "Test bundle",
		SharedVariables:   nil,
		TemplateVariables: map[string][]Variable{},
	}

	err := validateBundleMetadata(metadata)
	if err != nil {
		t.Errorf("Expected no error for nil shared_variables, got: %v", err)
	}
}

func TestValidateBundleMetadata_NilTemplateVariables(t *testing.T) {
	metadata := &BundleMetadata{
		SchemaVersion:     "v1",
		BundleType:        "k8s",
		BundleName:        "test",
		BundleDescription: "Test bundle",
		SharedVariables:   []Variable{},
		TemplateVariables: nil,
	}

	err := validateBundleMetadata(metadata)
	if err != nil {
		t.Errorf("Expected no error for nil template_variables, got: %v", err)
	}
}

func TestValidateBundleMetadata_InvalidSharedVariable(t *testing.T) {
	metadata := &BundleMetadata{
		SchemaVersion:     "v1",
		BundleType:        "k8s",
		BundleName:        "test",
		BundleDescription: "Test bundle",
		SharedVariables: []Variable{
			{
				Name:        "invalid_var",
				Description: "",
				Type:        "string",
			},
		},
	}

	err := validateBundleMetadata(metadata)
	if err == nil {
		t.Error("Expected error for invalid shared variable, got nil")
	}
}

func TestValidateBundleMetadata_InvalidTemplateVariable(t *testing.T) {
	metadata := &BundleMetadata{
		SchemaVersion:     "v1",
		BundleType:        "k8s",
		BundleName:        "test",
		BundleDescription: "Test bundle",
		SharedVariables:   []Variable{},
		TemplateVariables: map[string][]Variable{
			"test.yaml.tmpl": {
				{
					Name:        "invalid_var",
					Description: "Valid description",
					Type:        "invalid_type",
				},
			},
		},
	}

	err := validateBundleMetadata(metadata)
	if err == nil {
		t.Error("Expected error for invalid template variable, got nil")
	}
}

// Integration tests for LoadTemplateMetadata
func TestLoadTemplateMetadata_WithValidation(t *testing.T) {
	// Create a temporary directory for test files
	tmpDir := t.TempDir()

	// Test case 1: Valid metadata
	validMetadata := `{
		"schema_version": "v1",
		"template_name": "test",
		"template_description": "Test template",
		"template_type": "yaml",
		"variables": [
			{
				"name": "app_name",
				"description": "Application name",
				"type": "string",
				"default": "myapp"
			}
		]
	}`
	validPath := filepath.Join(tmpDir, "valid.json")
	if err := os.WriteFile(validPath, []byte(validMetadata), 0644); err != nil {
		t.Fatal(err)
	}

	metadata, err := LoadTemplateMetadata(validPath)
	if err != nil {
		t.Errorf("Expected no error for valid metadata, got: %v", err)
	}
	if metadata == nil {
		t.Error("Expected metadata to be loaded")
	}

	// Test case 2: Missing schema_version
	invalidMetadata := `{
		"template_name": "test",
		"template_description": "Test template",
		"variables": []
	}`
	invalidPath := filepath.Join(tmpDir, "invalid.json")
	if err := os.WriteFile(invalidPath, []byte(invalidMetadata), 0644); err != nil {
		t.Fatal(err)
	}

	_, err = LoadTemplateMetadata(invalidPath)
	if err == nil {
		t.Error("Expected error for missing schema_version, got nil")
	}

	// Test case 3: Invalid variable (missing type)
	invalidVarMetadata := `{
		"schema_version": "v1",
		"template_name": "test",
		"template_description": "Test template",
		"template_type": "yaml",
		"variables": [
			{
				"name": "app_name",
				"description": "Application name",
				"default": "myapp"
			}
		]
	}`
	invalidVarPath := filepath.Join(tmpDir, "invalid_var.json")
	if err := os.WriteFile(invalidVarPath, []byte(invalidVarMetadata), 0644); err != nil {
		t.Fatal(err)
	}

	_, err = LoadTemplateMetadata(invalidVarPath)
	if err == nil {
		t.Error("Expected error for missing variable type, got nil")
	}
}

// Integration tests for LoadBundleMetadata
func TestLoadBundleMetadata_WithValidation(t *testing.T) {
	tmpDir := t.TempDir()

	// Valid bundle metadata
	validMetadata := `{
		"schema_version": "v1",
		"bundle_type": "k8s",
		"bundle_name": "test-bundle",
		"bundle_description": "Test bundle",
		"shared_variables": [
			{
				"name": "namespace",
				"description": "Kubernetes namespace",
				"type": "string",
				"default": "default"
			}
		],
		"template_variables": {}
	}`
	validPath := filepath.Join(tmpDir, "valid_bundle.json")
	if err := os.WriteFile(validPath, []byte(validMetadata), 0644); err != nil {
		t.Fatal(err)
	}

	metadata, err := LoadBundleMetadata(validPath)
	if err != nil {
		t.Errorf("Expected no error for valid bundle metadata, got: %v", err)
	}
	if metadata == nil {
		t.Error("Expected bundle metadata to be loaded")
	}

	// Invalid bundle metadata (missing schema_version)
	invalidMetadata := `{
		"bundle_type": "k8s",
		"bundle_name": "test-bundle",
		"bundle_description": "Test bundle",
		"shared_variables": [],
		"template_variables": {}
	}`
	invalidPath := filepath.Join(tmpDir, "invalid_bundle.json")
	if err := os.WriteFile(invalidPath, []byte(invalidMetadata), 0644); err != nil {
		t.Fatal(err)
	}

	_, err = LoadBundleMetadata(invalidPath)
	if err == nil {
		t.Error("Expected error for missing schema_version in bundle, got nil")
	}
}

// Test MergeVariablesForTemplate with type validation
func TestMergeVariablesForTemplate_TypeValidation(t *testing.T) {
	bundleMeta := &BundleMetadata{
		SchemaVersion: "v1",
		SharedVariables: []Variable{
			{
				Name:        "replicas",
				Description: "Number of replicas",
				Type:        "int",
				Default:     "3",
			},
		},
		TemplateVariables: map[string][]Variable{},
	}

	// Valid integer value
	userVars := map[string]interface{}{
		"replicas": "5",
	}
	result, err := MergeVariablesForTemplate(bundleMeta, "test.yaml.tmpl", userVars)
	if err != nil {
		t.Errorf("Expected no error for valid integer, got: %v", err)
	}
	if result["replicas"] != "5" {
		t.Errorf("Expected replicas='5', got: %v", result["replicas"])
	}

	// Invalid integer value
	userVars = map[string]interface{}{
		"replicas": "not_a_number",
	}
	_, err = MergeVariablesForTemplate(bundleMeta, "test.yaml.tmpl", userVars)
	if err == nil {
		t.Error("Expected error for invalid integer, got nil")
	}
}
