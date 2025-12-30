package bundle

import (
	"os"
	"testing"
)

func TestParseVariables_Simple(t *testing.T) {
	varsList := []string{
		"app_name=myapp",
		"namespace=production",
	}

	result, err := parseVariables(varsList)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if result["app_name"] != "myapp" {
		t.Errorf("Expected app_name='myapp', got: %v", result["app_name"])
	}

	if result["namespace"] != "production" {
		t.Errorf("Expected namespace='production', got: %v", result["namespace"])
	}
}

func TestParseVariables_WithSpaces(t *testing.T) {
	varsList := []string{
		"  app_name = myapp  ",
		"namespace=production",
	}

	result, err := parseVariables(varsList)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if result["app_name"] != "myapp" {
		t.Errorf("Expected trimmed app_name='myapp', got: '%v'", result["app_name"])
	}
}

func TestParseVariables_InvalidFormat(t *testing.T) {
	varsList := []string{
		"invalid_no_equals",
	}

	_, err := parseVariables(varsList)
	if err == nil {
		t.Error("Expected error for invalid format, got nil")
	}
}

func TestParseVariables_EmptyKey(t *testing.T) {
	varsList := []string{
		"=value",
	}

	_, err := parseVariables(varsList)
	if err == nil {
		t.Error("Expected error for empty key, got nil")
	}
}

func TestParseVariablesWithTemplateOverrides_SharedOnly(t *testing.T) {
	varsList := []string{
		"app_name=myapp",
		"namespace=production",
	}

	shared, overrides, err := parseVariablesWithTemplateOverrides(varsList)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(shared) != 2 {
		t.Errorf("Expected 2 shared variables, got: %d", len(shared))
	}

	if shared["app_name"] != "myapp" {
		t.Errorf("Expected app_name='myapp', got: %v", shared["app_name"])
	}

	if len(overrides) != 0 {
		t.Errorf("Expected 0 template overrides, got: %d", len(overrides))
	}
}

func TestParseVariablesWithTemplateOverrides_TemplateSpecific(t *testing.T) {
	varsList := []string{
		"namespace=production",
		"ingress.yaml.tmpl:namespace=ingress-nginx",
	}

	shared, overrides, err := parseVariablesWithTemplateOverrides(varsList)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if shared["namespace"] != "production" {
		t.Errorf("Expected shared namespace='production', got: %v", shared["namespace"])
	}

	if len(overrides) != 1 {
		t.Fatalf("Expected 1 template override, got: %d", len(overrides))
	}

	ingressOverrides, exists := overrides["ingress.yaml.tmpl"]
	if !exists {
		t.Fatal("Expected override for 'ingress.yaml.tmpl', not found")
	}

	if ingressOverrides["namespace"] != "ingress-nginx" {
		t.Errorf("Expected ingress namespace='ingress-nginx', got: %v", ingressOverrides["namespace"])
	}
}

func TestParseVariablesWithTemplateOverrides_MultipleTemplates(t *testing.T) {
	varsList := []string{
		"app_name=myapp",
		"ingress.yaml.tmpl:namespace=ingress-nginx",
		"service.yaml.tmpl:service_type=NodePort",
		"ingress.yaml.tmpl:host=example.com",
	}

	shared, overrides, err := parseVariablesWithTemplateOverrides(varsList)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(shared) != 1 {
		t.Errorf("Expected 1 shared variable, got: %d", len(shared))
	}

	if len(overrides) != 2 {
		t.Fatalf("Expected 2 template overrides, got: %d", len(overrides))
	}

	// Check ingress overrides (should have 2 variables)
	ingressOverrides, exists := overrides["ingress.yaml.tmpl"]
	if !exists {
		t.Fatal("Expected override for 'ingress.yaml.tmpl', not found")
	}
	if len(ingressOverrides) != 2 {
		t.Errorf("Expected 2 ingress overrides, got: %d", len(ingressOverrides))
	}

	// Check service overrides
	serviceOverrides, exists := overrides["service.yaml.tmpl"]
	if !exists {
		t.Fatal("Expected override for 'service.yaml.tmpl', not found")
	}
	if serviceOverrides["service_type"] != "NodePort" {
		t.Errorf("Expected service_type='NodePort', got: %v", serviceOverrides["service_type"])
	}
}

func TestParseVariablesWithTemplateOverrides_InvalidFormat(t *testing.T) {
	varsList := []string{
		"template.yaml.tmpl:var:extra=value", // Too many colons
	}

	_, _, err := parseVariablesWithTemplateOverrides(varsList)
	if err == nil {
		t.Error("Expected error for invalid format, got nil")
	}
}

func TestParseVariablesWithTemplateOverrides_EmptyTemplateName(t *testing.T) {
	varsList := []string{
		":namespace=value",
	}

	_, _, err := parseVariablesWithTemplateOverrides(varsList)
	if err == nil {
		t.Error("Expected error for empty template name, got nil")
	}
}

func TestParseVariablesWithTemplateOverrides_EmptyVarName(t *testing.T) {
	varsList := []string{
		"template.yaml.tmpl:=value",
	}

	_, _, err := parseVariablesWithTemplateOverrides(varsList)
	if err == nil {
		t.Error("Expected error for empty variable name, got nil")
	}
}

func TestParseValuesWithOverrides_Simple(t *testing.T) {
	// Create a temporary values file
	content := `
app_name: myapp
namespace: production
replicas: 5
`
	tmpFile, err := createTempFile("values-*.yaml", content)
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer cleanupTempFile(tmpFile)

	shared, overrides, err := parseValuesWithOverrides(tmpFile)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if shared["app_name"] != "myapp" {
		t.Errorf("Expected app_name='myapp', got: %v", shared["app_name"])
	}

	if len(overrides) != 0 {
		t.Errorf("Expected 0 template overrides, got: %d", len(overrides))
	}
}

func TestParseValuesWithOverrides_WithTemplateOverrides(t *testing.T) {
	content := `
app_name: myapp
namespace: production

template_overrides:
  ingress.yaml.tmpl:
    namespace: ingress-nginx
    host: example.com
  service.yaml.tmpl:
    service_type: NodePort
`
	tmpFile, err := createTempFile("values-*.yaml", content)
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer cleanupTempFile(tmpFile)

	shared, overrides, err := parseValuesWithOverrides(tmpFile)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Check shared vars
	if shared["app_name"] != "myapp" {
		t.Errorf("Expected app_name='myapp', got: %v", shared["app_name"])
	}

	if shared["namespace"] != "production" {
		t.Errorf("Expected shared namespace='production', got: %v", shared["namespace"])
	}

	// template_overrides should not be in shared vars
	if _, exists := shared["template_overrides"]; exists {
		t.Error("template_overrides should not be in shared variables")
	}

	// Check template overrides
	if len(overrides) != 2 {
		t.Fatalf("Expected 2 template overrides, got: %d", len(overrides))
	}

	ingressOverrides, exists := overrides["ingress.yaml.tmpl"]
	if !exists {
		t.Fatal("Expected override for 'ingress.yaml.tmpl', not found")
	}

	if ingressOverrides["namespace"] != "ingress-nginx" {
		t.Errorf("Expected ingress namespace='ingress-nginx', got: %v", ingressOverrides["namespace"])
	}

	if ingressOverrides["host"] != "example.com" {
		t.Errorf("Expected ingress host='example.com', got: %v", ingressOverrides["host"])
	}
}

// Helper functions for testing
func createTempFile(pattern, content string) (string, error) {
	tmpFile, err := os.CreateTemp("", pattern)
	if err != nil {
		return "", err
	}
	defer tmpFile.Close()

	_, err = tmpFile.WriteString(content)
	if err != nil {
		return "", err
	}

	return tmpFile.Name(), nil
}

func cleanupTempFile(path string) {
	os.Remove(path)
}
