package metadata

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/wizardopstech/conjure/internal/security"
	"github.com/wizardopstech/conjure/internal/version"
)

const (
	// Going to make a GH issue to make these adjustable per config
	// limit the size of metadata files to prevent abuse
	MaxMetadataFileSize = 512 * 1024 // 512 KB
	// limit number of variables per template
	MaxVariablesPerTemplate = 100
)

type TemplateMetadata struct {
	SchemaVersion       string     `json:"schema_version"`
	TemplateName        string     `json:"template_name"`
	TemplateDescription string     `json:"template_description"`
	Version             string     `json:"version,omitempty"`
	TemplateType        string     `json:"template_type"`
	Variables           []Variable `json:"variables"`
}

type Variable struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Type        string `json:"type"`
	Default     string `json:"default"`
}

func LoadTemplateMetadata(metadataPath string) (*TemplateMetadata, error) {
	fileInfo, err := os.Stat(metadataPath)
	if os.IsNotExist(err) {
		return nil, fmt.Errorf("metadata file does not exist: %s", metadataPath)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to stat metadata file: %w", err)
	}

	if err := security.ValidateFileSize(fileInfo.Size(), MaxMetadataFileSize); err != nil {
		return nil, fmt.Errorf("metadata file size validation failed: %w", err)
	}

	data, err := os.ReadFile(metadataPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read metadata file: %w", err)
	}

	var metadata TemplateMetadata
	if err := json.Unmarshal(data, &metadata); err != nil {
		return nil, fmt.Errorf("failed to parse metadata JSON: %w", err)
	}

	if err := validateTemplateMetadata(&metadata); err != nil {
		return nil, err
	}

	return &metadata, nil
}

func ParseTemplateMetadata(data []byte) (*TemplateMetadata, error) {
	if err := security.ValidateFileSize(int64(len(data)), MaxMetadataFileSize); err != nil {
		return nil, fmt.Errorf("metadata size validation failed: %w", err)
	}

	var metadata TemplateMetadata
	if err := json.Unmarshal(data, &metadata); err != nil {
		return nil, fmt.Errorf("failed to parse metadata JSON: %w", err)
	}

	if err := validateTemplateMetadata(&metadata); err != nil {
		return nil, err
	}

	return &metadata, nil
}

func validateTemplateMetadata(metadata *TemplateMetadata) error {
	if metadata.SchemaVersion == "" {
		return fmt.Errorf("schema_version field is required (see documentation for migration from metadata_version)")
	}
	if metadata.SchemaVersion != "v1" {
		return fmt.Errorf("unsupported schema_version: %s (only 'v1' is supported)", metadata.SchemaVersion)
	}

	if metadata.TemplateType == "" {
		return fmt.Errorf("template_type field is required")
	}

	if metadata.TemplateName == "" {
		return fmt.Errorf("template_name field is required")
	}

	if metadata.TemplateDescription == "" {
		return fmt.Errorf("template_description field is required")
	}

	if metadata.Version != "" {
		if err := version.ValidateVersion(metadata.Version); err != nil {
			return fmt.Errorf("invalid version: %w", err)
		}
	}

	if len(metadata.Variables) > MaxVariablesPerTemplate {
		return fmt.Errorf("template has %d variables, exceeding maximum of %d", len(metadata.Variables), MaxVariablesPerTemplate)
	}

	for i, v := range metadata.Variables {
		if err := validateVariable(&v); err != nil {
			return fmt.Errorf("variable %d (%s): %w", i, v.Name, err)
		}
	}

	return nil
}

func validateVariable(v *Variable) error {
	if err := security.ValidateVariableName(v.Name); err != nil {
		return fmt.Errorf("invalid variable name: %w", err)
	}

	if v.Type == "" {
		return fmt.Errorf("type field is required")
	}

	validTypes := map[string]bool{
		"string": true,
		"int":    true,
		"bool":   true,
	}
	if !validTypes[v.Type] {
		return fmt.Errorf("invalid type '%s' (must be 'string', 'int', or 'bool')", v.Type)
	}

	if v.Description == "" {
		return fmt.Errorf("description field is required")
	}

	if v.Default != "" {
		if err := validateValueType(v.Default, v.Type); err != nil {
			return fmt.Errorf("default value validation failed: %w", err)
		}
	}

	return nil
}

func validateValueType(value interface{}, expectedType string) error {
	strValue := fmt.Sprintf("%v", value)

	switch expectedType {
	case "string":
		return nil

	case "int":
		_, err := strconv.Atoi(strValue)
		if err != nil {
			return fmt.Errorf("value '%s' is not a valid integer", strValue)
		}
		return nil

	case "bool":
		if strValue != "true" && strValue != "false" {
			return fmt.Errorf("value '%s' is not a valid boolean (must be 'true' or 'false')", strValue)
		}
		return nil

	default:
		return fmt.Errorf("unknown type: %s", expectedType)
	}
}

func ValidateAndMergeVariables(metadata *TemplateMetadata, userVars map[string]interface{}) (map[string]interface{}, error) {
	finalVars := make(map[string]interface{})

	for _, v := range metadata.Variables {
		if userValue, exists := userVars[v.Name]; exists {
			if err := validateValueType(userValue, v.Type); err != nil {
				return nil, fmt.Errorf("variable '%s': %w", v.Name, err)
			}
			finalVars[v.Name] = userValue
		} else if v.Default != "" {
			finalVars[v.Name] = v.Default
		} else {
			return nil, fmt.Errorf("required variable '%s' not provided (%s)", v.Name, v.Description)
		}
	}

	return finalVars, nil
}

type BundleMetadata struct {
	SchemaVersion     string                `json:"schema_version"`
	BundleType        string                `json:"bundle_type"`
	BundleName        string                `json:"bundle_name"`
	BundleDescription string                `json:"bundle_description"`
	Version           string                `json:"version,omitempty"` // Semantic version (e.g., "1.2.3")
	SharedVariables   []Variable            `json:"shared_variables"`
	TemplateVariables map[string][]Variable `json:"template_variables"`
}

func LoadBundleMetadata(metadataPath string) (*BundleMetadata, error) {
	fileInfo, err := os.Stat(metadataPath)
	if os.IsNotExist(err) {
		return nil, fmt.Errorf("bundle metadata file does not exist: %s", metadataPath)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to stat bundle metadata file: %w", err)
	}

	if err := security.ValidateFileSize(fileInfo.Size(), MaxMetadataFileSize); err != nil {
		return nil, fmt.Errorf("bundle metadata file size validation failed: %w", err)
	}

	data, err := os.ReadFile(metadataPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read bundle metadata file: %w", err)
	}

	var metadata BundleMetadata
	if err := json.Unmarshal(data, &metadata); err != nil {
		return nil, fmt.Errorf("failed to parse bundle metadata JSON: %w", err)
	}

	if err := validateBundleMetadata(&metadata); err != nil {
		return nil, err
	}

	return &metadata, nil
}

func ParseBundleMetadata(data []byte) (*BundleMetadata, error) {
	if err := security.ValidateFileSize(int64(len(data)), MaxMetadataFileSize); err != nil {
		return nil, fmt.Errorf("metadata size validation failed: %w", err)
	}

	var metadata BundleMetadata
	if err := json.Unmarshal(data, &metadata); err != nil {
		return nil, fmt.Errorf("failed to parse bundle metadata JSON: %w", err)
	}

	if err := validateBundleMetadata(&metadata); err != nil {
		return nil, err
	}

	return &metadata, nil
}

func validateBundleMetadata(metadata *BundleMetadata) error {
	if metadata.SchemaVersion == "" {
		return fmt.Errorf("schema_version field is required (see documentation for migration from metadata_version)")
	}
	if metadata.SchemaVersion != "v1" {
		return fmt.Errorf("unsupported schema_version: %s (only 'v1' is supported)", metadata.SchemaVersion)
	}

	if metadata.BundleName == "" {
		return fmt.Errorf("bundle_name field is required")
	}

	if metadata.BundleType == "" {
		return fmt.Errorf("bundle_type field is required")
	}

	if metadata.BundleDescription == "" {
		return fmt.Errorf("bundle_description field is required")
	}

	if metadata.Version != "" {
		if err := version.ValidateVersion(metadata.Version); err != nil {
			return fmt.Errorf("invalid version: %w", err)
		}
	}

	totalVars := len(metadata.SharedVariables)
	for _, templateVars := range metadata.TemplateVariables {
		totalVars += len(templateVars)
	}
	if totalVars > MaxVariablesPerTemplate {
		return fmt.Errorf("bundle has %d total variables, exceeding maximum of %d", totalVars, MaxVariablesPerTemplate)
	}

	for i, v := range metadata.SharedVariables {
		if err := validateVariable(&v); err != nil {
			return fmt.Errorf("shared variable %d (%s): %w", i, v.Name, err)
		}
	}

	for templateName, templateVars := range metadata.TemplateVariables {
		for i, v := range templateVars {
			if err := validateVariable(&v); err != nil {
				return fmt.Errorf("template '%s' variable %d (%s): %w", templateName, i, v.Name, err)
			}
		}
	}

	return nil
}

func MergeVariablesForTemplate(metadata *BundleMetadata, templateName string, userVars map[string]interface{}) (map[string]interface{}, error) {
	finalVars := make(map[string]interface{})

	varTypes := make(map[string]string)

	for _, v := range metadata.SharedVariables {
		varTypes[v.Name] = v.Type
		if v.Default != "" {
			finalVars[v.Name] = v.Default
		}
	}

	if templateVars, exists := metadata.TemplateVariables[templateName]; exists {
		for _, v := range templateVars {
			varTypes[v.Name] = v.Type
			if v.Default != "" {
				finalVars[v.Name] = v.Default
			}
		}
	}

	for k, v := range userVars {
		if expectedType, exists := varTypes[k]; exists {
			if err := validateValueType(v, expectedType); err != nil {
				return nil, fmt.Errorf("variable '%s': %w", k, err)
			}
		}
		finalVars[k] = v
	}

	for _, v := range metadata.SharedVariables {
		if v.Default == "" {
			if _, exists := finalVars[v.Name]; !exists {
				return nil, fmt.Errorf("required shared variable '%s' not provided (%s)", v.Name, v.Description)
			}
		}
	}

	if templateVars, exists := metadata.TemplateVariables[templateName]; exists {
		for _, v := range templateVars {
			if v.Default == "" {
				if _, exists := finalVars[v.Name]; !exists {
					return nil, fmt.Errorf("required variable '%s' not provided for template '%s' (%s)", v.Name, templateName, v.Description)
				}
			}
		}
	}

	return finalVars, nil
}

func GetAllVariablesForBundle(metadata *BundleMetadata) []Variable {
	varMap := make(map[string]Variable)

	for _, v := range metadata.SharedVariables {
		varMap[v.Name] = v
	}

	for _, templateVars := range metadata.TemplateVariables {
		for _, v := range templateVars {
			if _, exists := varMap[v.Name]; !exists {
				varMap[v.Name] = v
			}
		}
	}

	vars := make([]Variable, 0, len(varMap))
	for _, v := range varMap {
		vars = append(vars, v)
	}

	return vars
}
