package source

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"

	"github.com/wizardopstech/conjure/internal/security"
)

const (
	// MaxFileSize limits the maximum size of downloadable files. Make a config option later.
	MaxFileSize = 100 * 1024 * 1024 // 100 MB
)

type Verifier struct{}

func NewVerifier() *Verifier {
	return &Verifier{}
}

func (v *Verifier) ComputeSHA256(filePath string) (string, error) {
	if err := security.ValidatePathSafety(filePath); err != nil {
		return "", fmt.Errorf("invalid file path: %w", err)
	}

	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", fmt.Errorf("failed to compute hash: %w", err)
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}

func (v *Verifier) ComputeSHA256Bytes(data []byte) string {
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

func (v *Verifier) VerifySHA256(filePath, expectedHash string) error {
	actualHash, err := v.ComputeSHA256(filePath)
	if err != nil {
		return err
	}

	if !security.CompareHashes(actualHash, expectedHash) {
		return fmt.Errorf("SHA256 mismatch: expected %s, got %s", expectedHash, actualHash)
	}

	return nil
}

func (v *Verifier) VerifySHA256Bytes(data []byte, expectedHash string) error {
	actualHash := v.ComputeSHA256Bytes(data)

	if !security.CompareHashes(actualHash, expectedHash) {
		return fmt.Errorf("SHA256 mismatch: expected %s, got %s", expectedHash, actualHash)
	}

	return nil
}

func (v *Verifier) ValidateFileSize(size int64) error {
	return security.ValidateFileSize(size, MaxFileSize)
}

func (v *Verifier) VerifyFile(filePath string, expectedSize int64, expectedHash string) error {
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return fmt.Errorf("failed to stat file: %w", err)
	}

	if fileInfo.Size() != expectedSize {
		return fmt.Errorf("file size mismatch: expected %d bytes, got %d bytes", expectedSize, fileInfo.Size())
	}

	if err := v.ValidateFileSize(fileInfo.Size()); err != nil {
		return err
	}

	return v.VerifySHA256(filePath, expectedHash)
}
