package metadata

import (
	"testing"
)

func TestMergeVariablesForTemplate_SharedDefaults(t *testing.T) {
	bundleMeta := &BundleMetadata{
		SharedVariables: []Variable{
			{Name: "namespace", Default: "default"},
			{Name: "app_name", Default: "myapp"},
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
		SharedVariables: []Variable{
			{Name: "namespace", Default: "default"},
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
		SharedVariables: []Variable{
			{Name: "namespace", Default: "default"},
		},
		TemplateVariables: map[string][]Variable{
			"special.yaml.tmpl": {
				{Name: "namespace", Default: "special-namespace"},
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
		SharedVariables: []Variable{
			{Name: "namespace", Default: "default"},
		},
		TemplateVariables: map[string][]Variable{
			"special.yaml.tmpl": {
				{Name: "namespace", Default: "special-namespace"},
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
		SharedVariables: []Variable{
			{Name: "app_name", Required: true},
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
		SharedVariables: []Variable{},
		TemplateVariables: map[string][]Variable{
			"test.yaml.tmpl": {
				{Name: "replicas", Required: true},
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
		SharedVariables: []Variable{
			{Name: "app_name", Required: true},
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
		SharedVariables: []Variable{
			{Name: "namespace", Default: "default"},
			{Name: "app_name", Required: true},
		},
		TemplateVariables: map[string][]Variable{
			"deployment.yaml.tmpl": {
				{Name: "replicas", Default: "3"},
			},
			"service.yaml.tmpl": {
				{Name: "service_port", Default: "80"},
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
		SharedVariables: []Variable{
			{Name: "namespace", Default: "default"},
		},
		TemplateVariables: map[string][]Variable{
			"deployment.yaml.tmpl": {
				{Name: "namespace", Default: "override"}, // Same name as shared
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
