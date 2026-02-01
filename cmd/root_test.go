package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolveConfigPath(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current working directory: %v", err)
	}

	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("failed to get user home directory: %v", err)
	}

	tempDir := t.TempDir()
	testConfigFile := filepath.Join(tempDir, "test-config.yaml")
	if err := os.WriteFile(testConfigFile, []byte("test: config"), 0644); err != nil {
		t.Fatalf("failed to create test config file: %v", err)
	}

	tests := []struct {
		name        string
		inputPath   string
		expectError bool
		setup       func() string
	}{
		{
			name:        "absolute path to existing file",
			inputPath:   testConfigFile,
			expectError: false,
			setup: func() string {
				return testConfigFile
			},
		},
		{
			name:        "relative path from cwd",
			inputPath:   "root_test.go",
			expectError: false,
			setup: func() string {
				return filepath.Join(cwd, "root_test.go")
			},
		},
		{
			name:        "non-existent file",
			inputPath:   filepath.Join(tempDir, "nonexistent.yaml"),
			expectError: true,
			setup: func() string {
				return ""
			},
		},
		{
			name:        "home directory expansion with existing file",
			inputPath:   "~",
			expectError: false,
			setup: func() string {
				// just verify home is expanded correctly
				return home
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expectedPath := tt.setup()

			resolvedPath, err := resolveConfigPath(tt.inputPath)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			resolvedPath = filepath.Clean(resolvedPath)
			expectedPath = filepath.Clean(expectedPath)

			if resolvedPath != expectedPath {
				t.Errorf("expected path %q, got %q", expectedPath, resolvedPath)
			}
		})
	}
}

func TestResolveConfigPathRelative(t *testing.T) {
	tempDir := t.TempDir()
	subDir := filepath.Join(tempDir, "subdir")
	if err := os.Mkdir(subDir, 0755); err != nil {
		t.Fatalf("failed to create subdirectory: %v", err)
	}

	configFile := filepath.Join(tempDir, "config.yaml")
	if err := os.WriteFile(configFile, []byte("test: config"), 0644); err != nil {
		t.Fatalf("failed to create config file: %v", err)
	}

	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get original working directory: %v", err)
	}
	defer os.Chdir(originalWd)

	if err := os.Chdir(subDir); err != nil {
		t.Fatalf("failed to change to subdirectory: %v", err)
	}

	resolvedPath, err := resolveConfigPath("../config.yaml")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}

	expectedPath := filepath.Clean(configFile)
	resolvedPath = filepath.Clean(resolvedPath)

	if resolvedPath != expectedPath {
		t.Errorf("expected path %q, got %q", expectedPath, resolvedPath)
	}
}

func TestResolveConfigPathHomeDirExpansion(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("failed to get user home directory: %v", err)
	}

	tests := []struct {
		name         string
		createFile   string
		inputPath    string
		shouldExpand bool
	}{
		{
			name:         "tilde with forward slash",
			createFile:   "test-conjure-config.yaml",
			inputPath:    "~/test-conjure-config.yaml",
			shouldExpand: true,
		},
		{
			name:         "tilde with backslash",
			createFile:   "test-conjure-config2.yaml",
			inputPath:    "~\\test-conjure-config2.yaml",
			shouldExpand: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testFile := filepath.Join(home, tt.createFile)
			if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
				t.Skipf("failed to create test file in home directory: %v", err)
				return
			}
			defer os.Remove(testFile)

			resolvedPath, err := resolveConfigPath(tt.inputPath)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			expectedPath := filepath.Clean(testFile)
			resolvedPath = filepath.Clean(resolvedPath)

			if resolvedPath != expectedPath {
				t.Errorf("expected path %q, got %q", expectedPath, resolvedPath)
			}

			if tt.shouldExpand && !filepath.IsAbs(resolvedPath) {
				t.Errorf("expected absolute path, got %q", resolvedPath)
			}
		})
	}
}
