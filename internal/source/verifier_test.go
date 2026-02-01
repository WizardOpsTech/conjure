package source

import (
	"os"
	"path/filepath"
	"testing"
)

func TestVerifier_ComputeSHA256(t *testing.T) {
	verifier := NewVerifier()

	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.txt")
	testContent := []byte("Hello, World!")

	err := os.WriteFile(testFile, testContent, 0600)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	hash, err := verifier.ComputeSHA256(testFile)
	if err != nil {
		t.Fatalf("ComputeSHA256() error = %v", err)
	}

	// Known SHA256 of "Hello, World!"
	expectedHash := "dffd6021bb2bd5b0af676290809ec3a53191dd81c7f70a4b28688a362182986f"

	if hash != expectedHash {
		t.Errorf("Expected hash %s, got %s", expectedHash, hash)
	}
}

func TestVerifier_ComputeSHA256Bytes(t *testing.T) {
	verifier := NewVerifier()

	testContent := []byte("Hello, World!")
	hash := verifier.ComputeSHA256Bytes(testContent)

	// Known SHA256 of "Hello, World!"
	expectedHash := "dffd6021bb2bd5b0af676290809ec3a53191dd81c7f70a4b28688a362182986f"

	if hash != expectedHash {
		t.Errorf("Expected hash %s, got %s", expectedHash, hash)
	}
}

func TestVerifier_VerifySHA256(t *testing.T) {
	verifier := NewVerifier()

	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.txt")
	testContent := []byte("Hello, World!")

	err := os.WriteFile(testFile, testContent, 0600)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	correctHash := "dffd6021bb2bd5b0af676290809ec3a53191dd81c7f70a4b28688a362182986f"
	incorrectHash := "0000000000000000000000000000000000000000000000000000000000000000"

	tests := []struct {
		name         string
		expectedHash string
		wantErr      bool
	}{
		{
			name:         "correct hash",
			expectedHash: correctHash,
			wantErr:      false,
		},
		{
			name:         "incorrect hash",
			expectedHash: incorrectHash,
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := verifier.VerifySHA256(testFile, tt.expectedHash)
			if (err != nil) != tt.wantErr {
				t.Errorf("VerifySHA256() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestVerifier_VerifySHA256Bytes(t *testing.T) {
	verifier := NewVerifier()

	testContent := []byte("Hello, World!")
	correctHash := "dffd6021bb2bd5b0af676290809ec3a53191dd81c7f70a4b28688a362182986f"
	incorrectHash := "0000000000000000000000000000000000000000000000000000000000000000"

	tests := []struct {
		name         string
		expectedHash string
		wantErr      bool
	}{
		{
			name:         "correct hash",
			expectedHash: correctHash,
			wantErr:      false,
		},
		{
			name:         "incorrect hash",
			expectedHash: incorrectHash,
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := verifier.VerifySHA256Bytes(testContent, tt.expectedHash)
			if (err != nil) != tt.wantErr {
				t.Errorf("VerifySHA256Bytes() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestVerifier_ValidateFileSize(t *testing.T) {
	verifier := NewVerifier()

	tests := []struct {
		name    string
		size    int64
		wantErr bool
	}{
		{
			name:    "valid size",
			size:    1024,
			wantErr: false,
		},
		{
			name:    "zero size",
			size:    0,
			wantErr: false,
		},
		{
			name:    "max size",
			size:    MaxFileSize,
			wantErr: false,
		},
		{
			name:    "exceeds max size",
			size:    MaxFileSize + 1,
			wantErr: true,
		},
		{
			name:    "negative size",
			size:    -1,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := verifier.ValidateFileSize(tt.size)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateFileSize() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestVerifier_VerifyFile(t *testing.T) {
	verifier := NewVerifier()

	// Create test file
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.txt")
	testContent := []byte("Hello, World!")

	err := os.WriteFile(testFile, testContent, 0600)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	correctHash := "dffd6021bb2bd5b0af676290809ec3a53191dd81c7f70a4b28688a362182986f"
	correctSize := int64(len(testContent))

	tests := []struct {
		name         string
		expectedSize int64
		expectedHash string
		wantErr      bool
	}{
		{
			name:         "correct size and hash",
			expectedSize: correctSize,
			expectedHash: correctHash,
			wantErr:      false,
		},
		{
			name:         "incorrect size",
			expectedSize: 999,
			expectedHash: correctHash,
			wantErr:      true,
		},
		{
			name:         "incorrect hash",
			expectedSize: correctSize,
			expectedHash: "0000000000000000000000000000000000000000000000000000000000000000",
			wantErr:      true,
		},
		{
			name:         "size exceeds max",
			expectedSize: MaxFileSize + 1,
			expectedHash: correctHash,
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := verifier.VerifyFile(testFile, tt.expectedSize, tt.expectedHash)
			if (err != nil) != tt.wantErr {
				t.Errorf("VerifyFile() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestVerifier_ComputeSHA256InvalidPath(t *testing.T) {
	verifier := NewVerifier()

	_, err := verifier.ComputeSHA256("../../../etc/passwd")
	if err == nil {
		t.Error("Expected error for path traversal")
	}

	_, err = verifier.ComputeSHA256("testdata/nonexistent.txt")
	if err == nil {
		t.Error("Expected error for non-existent file")
	}
}
