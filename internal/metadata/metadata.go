package metadata

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// TemplateMetadata defines the structure of template metadata files
type TemplateMetadata struct {
	TemplateName string     `json:"template_name"`
	Description  string     `json:"description"`
	Variables    []Variable `json:"variables"`
}

// Variable defines a template variable
type Variable struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Type        string `json:"type"`     // "string", "int", "bool"
	Required    bool   `json:"required"` // Is this variable required?
	Default     string `json:"default"`  // Default value if not provided
}

// LoadTemplateMetadata loads metadata from a .json file
func LoadTemplateMetadata(metadataPath string) (*TemplateMetadata, error) {
	// Check if metadata file exists
	if _, err := os.Stat(metadataPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("metadata file does not exist: %s", metadataPath)
	}

	// Read metadata file
	data, err := os.ReadFile(metadataPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read metadata file: %w", err)
	}

	// Parse JSON
	var metadata TemplateMetadata
	if err := json.Unmarshal(data, &metadata); err != nil {
		return nil, fmt.Errorf("failed to parse metadata JSON: %w", err)
	}

	return &metadata, nil
}

// ValidateAndMergeVariables validates user-provided variables against metadata
// and merges them with defaults. Returns final variable map or error.
func ValidateAndMergeVariables(metadata *TemplateMetadata, userVars map[string]interface{}) (map[string]interface{}, error) {
	finalVars := make(map[string]interface{})

	// Process each variable defined in metadata
	for _, v := range metadata.Variables {
		// Check if user provided this variable
		if userValue, exists := userVars[v.Name]; exists {
			// User provided the value
			finalVars[v.Name] = userValue
		} else if v.Default != "" {
			// Use default value
			finalVars[v.Name] = v.Default
		} else if v.Required {
			// Required but not provided and no default
			return nil, fmt.Errorf("required variable '%s' not provided (%s)", v.Name, v.Description)
		}
		// If not required and no default, skip (variable won't be in final map)
	}

	return finalVars, nil
}

// GetMetadataPath returns the expected metadata file path for a template
func GetMetadataPath(templatesDir, templateName string) string {
	return filepath.Join(templatesDir, "templates", templateName+".json")
}

// BundleMetadata defines the structure of bundle metadata files (conjure.json)
type BundleMetadata struct {
	BundleType        string                       `json:"bundle_type"`
	BundleName        string                       `json:"bundle_name"`
	BundleDescription string                       `json:"bundle_description"`
	SharedVariables   []Variable                   `json:"shared_variables"`
	TemplateVariables map[string][]Variable        `json:"template_variables"`
}

// LoadBundleMetadata loads metadata from a bundle's conjure.json file
func LoadBundleMetadata(metadataPath string) (*BundleMetadata, error) {
	// Check if metadata file exists
	if _, err := os.Stat(metadataPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("bundle metadata file does not exist: %s", metadataPath)
	}

	// Read metadata file
	data, err := os.ReadFile(metadataPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read bundle metadata file: %w", err)
	}

	// Parse JSON
	var metadata BundleMetadata
	if err := json.Unmarshal(data, &metadata); err != nil {
		return nil, fmt.Errorf("failed to parse bundle metadata JSON: %w", err)
	}

	return &metadata, nil
}

// MergeVariablesForTemplate merges variables for a specific template in a bundle
// Precedence: shared defaults < template-specific defaults < user values
func MergeVariablesForTemplate(metadata *BundleMetadata, templateName string, userVars map[string]interface{}) (map[string]interface{}, error) {
	finalVars := make(map[string]interface{})

	// Start with shared variable defaults
	for _, v := range metadata.SharedVariables {
		if v.Default != "" {
			finalVars[v.Name] = v.Default
		}
	}

	// Override with template-specific variable defaults
	if templateVars, exists := metadata.TemplateVariables[templateName]; exists {
		for _, v := range templateVars {
			if v.Default != "" {
				finalVars[v.Name] = v.Default
			}
		}
	}

	// Override with user-provided values
	for k, v := range userVars {
		finalVars[k] = v
	}

	// Validate required variables
	// Check shared variables
	for _, v := range metadata.SharedVariables {
		if v.Required {
			if _, exists := finalVars[v.Name]; !exists {
				return nil, fmt.Errorf("required shared variable '%s' not provided (%s)", v.Name, v.Description)
			}
		}
	}

	// Check template-specific variables
	if templateVars, exists := metadata.TemplateVariables[templateName]; exists {
		for _, v := range templateVars {
			if v.Required {
				if _, exists := finalVars[v.Name]; !exists {
					return nil, fmt.Errorf("required variable '%s' not provided for template '%s' (%s)", v.Name, templateName, v.Description)
				}
			}
		}
	}

	return finalVars, nil
}

// GetBundleMetadataPath returns the expected metadata file path for a bundle
func GetBundleMetadataPath(templatesDir, bundleName string) string {
	return filepath.Join(templatesDir, "bundles", bundleName, "conjure.json")
}

// GetAllVariablesForBundle returns all variables (shared + all template-specific) for interactive prompting
func GetAllVariablesForBundle(metadata *BundleMetadata) []Variable {
	varMap := make(map[string]Variable)

	// Add shared variables
	for _, v := range metadata.SharedVariables {
		varMap[v.Name] = v
	}

	// Add template-specific variables (will override shared if same name)
	for _, templateVars := range metadata.TemplateVariables {
		for _, v := range templateVars {
			// Only add if not already in map, or if it's a different definition
			if _, exists := varMap[v.Name]; !exists {
				varMap[v.Name] = v
			}
		}
	}

	// Convert map to slice
	vars := make([]Variable, 0, len(varMap))
	for _, v := range varMap {
		vars = append(vars, v)
	}

	return vars
}
