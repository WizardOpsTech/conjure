package source

import (
	"time"
)

// contains information about a template
type TemplateInfo struct {
	Name        string
	Type        string
	Description string
	Versions    []string
}

// contains information about a bundle
type BundleInfo struct {
	Name        string
	Type        string
	Description string
	Versions    []string
}

// contains the full content of a template at a specific version
type TemplateContent struct {
	Info        TemplateInfo
	Version     string
	MetadataRaw []byte
	TemplateRaw []byte
}

// contains the full content of a bundle at a specific version
type BundleContent struct {
	Info        BundleInfo
	Version     string
	MetadataRaw []byte
	Files       map[string][]byte // filename -> content
}

// Index represents the structure of an index.json file
type Index struct {
	SchemaVersion string               `json:"schema_version"`
	LastUpdated   time.Time            `json:"last_updated"`
	Templates     []TemplateIndexEntry `json:"templates"`
	Bundles       []BundleIndexEntry   `json:"bundles,omitempty"`
}

// represents a template in the index
type TemplateIndexEntry struct {
	Name        string         `json:"name"`
	Type        string         `json:"type"`
	Description string         `json:"description"`
	Versions    []VersionEntry `json:"versions"`
}

// represents a bundle in the index
type BundleIndexEntry struct {
	Name        string         `json:"name"`
	Type        string         `json:"type"`
	Description string         `json:"description"`
	Versions    []VersionEntry `json:"versions"`
}

// represents a specific version of a template or bundle
type VersionEntry struct {
	Version string      `json:"version"`
	Files   []FileEntry `json:"files"`
}

// represents a file in a version
type FileEntry struct {
	Name   string `json:"name"`
	Path   string `json:"path"`
	Size   int64  `json:"size"`
	SHA256 string `json:"sha256"`
}
