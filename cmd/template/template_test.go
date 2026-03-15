package template

import (
	"os"
	"strings"
	"testing"

	"github.com/wizardopstech/conjure/internal/render"
)

func TestParseVariables_Simple(t *testing.T) {
	varsList := []string{
		"vm_name=test-vm",
		"cpu=4",
		"memory=8192",
	}

	result, err := render.ParseVariables(varsList)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if result["vm_name"] != "test-vm" {
		t.Errorf("Expected vm_name='test-vm', got: %v", result["vm_name"])
	}

	if result["cpu"] != "4" {
		t.Errorf("Expected cpu='4', got: %v", result["cpu"])
	}

	if result["memory"] != "8192" {
		t.Errorf("Expected memory='8192', got: %v", result["memory"])
	}
}

func TestParseVariables_WithSpaces(t *testing.T) {
	varsList := []string{
		"  vm_name = test-vm  ",
		"cpu=4",
	}

	result, err := render.ParseVariables(varsList)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if result["vm_name"] != "test-vm" {
		t.Errorf("Expected trimmed vm_name='test-vm', got: '%v'", result["vm_name"])
	}
}

func TestParseVariables_InvalidFormat(t *testing.T) {
	varsList := []string{
		"invalid_no_equals",
	}

	_, err := render.ParseVariables(varsList)
	if err == nil {
		t.Error("Expected error for invalid format, got nil")
	}
}

func TestParseVariables_EmptyKey(t *testing.T) {
	varsList := []string{
		"=value",
	}

	_, err := render.ParseVariables(varsList)
	if err == nil {
		t.Error("Expected error for empty key, got nil")
	}
}

func TestRenderTemplate_Simple(t *testing.T) {
	templateContent := "Hello {{ .name }}!"
	variables := map[string]interface{}{
		"name": "World",
	}

	result, err := render.RenderTemplate(templateContent, variables)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if result != "Hello World!" {
		t.Errorf("Expected 'Hello World!', got: %s", result)
	}
}

func TestRenderTemplate_MultipleVariables(t *testing.T) {
	templateContent := "VM: {{ .vm_name }}, CPU: {{ .cpu }}, Memory: {{ .memory }}"
	variables := map[string]interface{}{
		"vm_name": "test-vm",
		"cpu":     "4",
		"memory":  "8192",
	}

	result, err := render.RenderTemplate(templateContent, variables)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	expected := "VM: test-vm, CPU: 4, Memory: 8192"
	if result != expected {
		t.Errorf("Expected '%s', got: %s", expected, result)
	}
}

func TestRenderTemplate_MissingVariable(t *testing.T) {
	templateContent := "Hello {{ .missing }}!"
	variables := map[string]interface{}{}

	_, err := render.RenderTemplate(templateContent, variables)
	if err == nil {
		t.Error("Expected error for missing variable, got nil")
	}
}

func TestRenderTemplate_InvalidSyntax(t *testing.T) {
	templateContent := "Hello {{ .name !"
	variables := map[string]interface{}{
		"name": "World",
	}

	_, err := render.RenderTemplate(templateContent, variables)
	if err == nil {
		t.Error("Expected error for invalid template syntax, got nil")
	}
}

func TestRenderTemplate_ConditionalLogic(t *testing.T) {
	templateContent := `{{if .enabled}}Enabled{{else}}Disabled{{end}}`

	variables := map[string]interface{}{
		"enabled": true,
	}

	result, err := render.RenderTemplate(templateContent, variables)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if result != "Enabled" {
		t.Errorf("Expected 'Enabled', got: %s", result)
	}

	// Test with false
	variables["enabled"] = false
	result, err = render.RenderTemplate(templateContent, variables)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if result != "Disabled" {
		t.Errorf("Expected 'Disabled', got: %s", result)
	}
}

func TestRenderTemplate_Loops(t *testing.T) {
	templateContent := `{{range .items}}{{ . }}
{{end}}`

	variables := map[string]interface{}{
		"items": []string{"one", "two", "three"},
	}

	result, err := render.RenderTemplate(templateContent, variables)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(result), "\n")
	if len(lines) != 3 {
		t.Errorf("Expected 3 lines, got: %d", len(lines))
	}

	if lines[0] != "one" {
		t.Errorf("Expected first line 'one', got: %s", lines[0])
	}
}

func TestParseValues_Simple(t *testing.T) {
	content := `
vm_name: test-vm-01
num_cpus: 2
memory: 4096
`
	tmpFile, err := createTempFile("values-*.yaml", content)
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer cleanupTempFile(tmpFile)

	result, err := parseValues(tmpFile)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if result["vm_name"] != "test-vm-01" {
		t.Errorf("Expected vm_name='test-vm-01', got: %v", result["vm_name"])
	}

	// YAML numbers are parsed as int
	if result["num_cpus"] != 2 {
		t.Errorf("Expected num_cpus=2, got: %v", result["num_cpus"])
	}
}

func TestParseValues_InvalidYAML(t *testing.T) {
	content := `
invalid yaml content
  bad indentation:
`
	tmpFile, err := createTempFile("values-*.yaml", content)
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer cleanupTempFile(tmpFile)

	_, err = parseValues(tmpFile)
	if err == nil {
		t.Error("Expected error for invalid YAML, got nil")
	}
}

func TestParseValues_NonexistentFile(t *testing.T) {
	_, err := parseValues("/nonexistent/file.yaml")
	if err == nil {
		t.Error("Expected error for nonexistent file, got nil")
	}
}

func createTempFile(pattern, content string) (string, error) {
	tmpFile, err := os.CreateTemp("", pattern)
	if err != nil {
		return "", err
	}
	defer func() { _ = tmpFile.Close() }()

	_, err = tmpFile.WriteString(content)
	if err != nil {
		return "", err
	}

	return tmpFile.Name(), nil
}

func cleanupTempFile(path string) {
	_ = os.Remove(path)
}
