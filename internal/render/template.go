package render

import (
	"bytes"
	"fmt"
	"text/template"
)

func RenderTemplate(templateContent string, variables map[string]interface{}) (string, error) {
	tmpl, err := template.New("template").Option("missingkey=error").Parse(templateContent)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	var rendered bytes.Buffer
	if err := tmpl.Execute(&rendered, variables); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return rendered.String(), nil
}
