package security

import (
	"strings"
	"testing"
)

func TestValidatePathSafety(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{
			name:    "valid simple path",
			path:    "templates/mytemplate",
			wantErr: false,
		},
		{
			name:    "valid path with subdirectory",
			path:    "bundles/mybundle/templates",
			wantErr: false,
		},
		{
			name:    "empty path",
			path:    "",
			wantErr: true,
		},
		{
			name:    "path with parent directory reference",
			path:    "../../../etc/passwd",
			wantErr: true,
		},
		{
			name:    "path with embedded parent directory",
			path:    "templates/../../../etc/passwd",
			wantErr: true,
		},
		{
			name:    "path with null byte",
			path:    "templates\x00/mytemplate",
			wantErr: true,
		},
		{
			name:    "path with suspicious pattern /../",
			path:    "templates/../config",
			wantErr: true,
		},
		{
			name:    "path with suspicious pattern /./",
			path:    "templates/./config",
			wantErr: true,
		},
		{
			name:    "Windows path with parent directory",
			path:    "templates\\..\\..\\windows\\system32",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePathSafety(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidatePathSafety() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateURL(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		wantErr bool
	}{
		{
			name:    "valid https URL",
			url:     "https://github.com/org/repo",
			wantErr: false,
		},
		{
			name:    "valid http URL",
			url:     "http://example.com/templates",
			wantErr: false,
		},
		{
			name:    "valid localhost URL",
			url:     "http://localhost:8080/templates",
			wantErr: false,
		},
		{
			name:    "valid private IP URL",
			url:     "http://192.168.1.100/templates",
			wantErr: false,
		},
		{
			name:    "empty URL",
			url:     "",
			wantErr: true,
		},
		{
			name:    "invalid scheme ftp",
			url:     "ftp://example.com/file",
			wantErr: true,
		},
		{
			name:    "invalid scheme file",
			url:     "file:///etc/passwd",
			wantErr: true,
		},
		{
			name:    "URL without hostname",
			url:     "https://",
			wantErr: true,
		},
		{
			name:    "malformed URL",
			url:     "not a url",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateURL(tt.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateURL() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateVariableName(t *testing.T) {
	tests := []struct {
		name    string
		varName string
		wantErr bool
	}{
		{
			name:    "valid simple variable",
			varName: "my_variable",
			wantErr: false,
		},
		{
			name:    "valid uppercase variable",
			varName: "MY_VARIABLE",
			wantErr: false,
		},
		{
			name:    "valid mixed case variable",
			varName: "MyVariable",
			wantErr: false,
		},
		{
			name:    "valid variable with numbers",
			varName: "variable123",
			wantErr: false,
		},
		{
			name:    "valid variable with underscores",
			varName: "my_var_123",
			wantErr: false,
		},
		{
			name:    "valid variable starting with number",
			varName: "123variable",
			wantErr: false,
		},
		{
			name:    "valid variable with hyphen",
			varName: "my-variable",
			wantErr: false,
		},
		{
			name:    "valid variable with mixed separators",
			varName: "my-var_123",
			wantErr: false,
		},
		{
			name:    "empty variable name",
			varName: "",
			wantErr: true,
		},
		{
			name:    "variable with space",
			varName: "my variable",
			wantErr: true,
		},
		{
			name:    "variable with special characters",
			varName: "my$variable",
			wantErr: true,
		},
		{
			name:    "variable with dot",
			varName: "my.variable",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateVariableName(tt.varName)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateVariableName() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCompareHashes(t *testing.T) {
	tests := []struct {
		name  string
		hash1 string
		hash2 string
		want  bool
	}{
		{
			name:  "identical hashes",
			hash1: "abc123def456",
			hash2: "abc123def456",
			want:  true,
		},
		{
			name:  "different hashes same length",
			hash1: "abc123def456",
			hash2: "abc123def457",
			want:  false,
		},
		{
			name:  "different length hashes",
			hash1: "abc123",
			hash2: "abc123def",
			want:  false,
		},
		{
			name:  "empty hashes",
			hash1: "",
			hash2: "",
			want:  true,
		},
		{
			name:  "one empty hash",
			hash1: "abc123",
			hash2: "",
			want:  false,
		},
		{
			name:  "long identical hashes",
			hash1: strings.Repeat("a", 64),
			hash2: strings.Repeat("a", 64),
			want:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CompareHashes(tt.hash1, tt.hash2); got != tt.want {
				t.Errorf("CompareHashes() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestValidateFileSize(t *testing.T) {
	tests := []struct {
		name    string
		size    int64
		maxSize int64
		wantErr bool
	}{
		{
			name:    "size within limit",
			size:    1024,
			maxSize: 2048,
			wantErr: false,
		},
		{
			name:    "size equal to limit",
			size:    2048,
			maxSize: 2048,
			wantErr: false,
		},
		{
			name:    "size exceeds limit",
			size:    3048,
			maxSize: 2048,
			wantErr: true,
		},
		{
			name:    "zero size",
			size:    0,
			maxSize: 1024,
			wantErr: false,
		},
		{
			name:    "negative size",
			size:    -1,
			maxSize: 1024,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateFileSize(tt.size, tt.maxSize)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateFileSize() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
