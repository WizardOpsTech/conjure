package config

import (
	"testing"
)

func TestConfigApplyDefaults(t *testing.T) {
	tests := []struct {
		name     string
		input    Config
		expected Config
	}{
		{
			name: "remote URL triggers both source",
			input: Config{
				TemplatesRemoteURL: "https://example.com/templates",
				BundlesRemoteURL:   "https://example.com/bundles",
			},
			expected: Config{
				TemplatesRemoteURL: "https://example.com/templates",
				BundlesRemoteURL:   "https://example.com/bundles",
				TemplatesSource:    "both",
				BundlesSource:      "both",
				TemplatesPriority:  "local-first",
				BundlesPriority:    "local-first",
			},
		},
		{
			name: "explicit source types preserved",
			input: Config{
				TemplatesSource: "remote",
				BundlesSource:   "local",
			},
			expected: Config{
				TemplatesSource:   "remote",
				BundlesSource:     "local",
				TemplatesPriority: "local-first",
				BundlesPriority:   "local-first",
			},
		},
		{
			name: "custom priority preserved",
			input: Config{
				TemplatesPriority: "remote-first",
				BundlesPriority:   "remote-first",
			},
			expected: Config{
				TemplatesSource:   "local",
				BundlesSource:     "local",
				TemplatesPriority: "remote-first",
				BundlesPriority:   "remote-first",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := tt.input
			cfg.applyDefaults()

			if cfg.TemplatesSource != tt.expected.TemplatesSource {
				t.Errorf("TemplatesSource = %v, want %v", cfg.TemplatesSource, tt.expected.TemplatesSource)
			}
			if cfg.BundlesSource != tt.expected.BundlesSource {
				t.Errorf("BundlesSource = %v, want %v", cfg.BundlesSource, tt.expected.BundlesSource)
			}
			if cfg.TemplatesPriority != tt.expected.TemplatesPriority {
				t.Errorf("TemplatesPriority = %v, want %v", cfg.TemplatesPriority, tt.expected.TemplatesPriority)
			}
			if cfg.BundlesPriority != tt.expected.BundlesPriority {
				t.Errorf("BundlesPriority = %v, want %v", cfg.BundlesPriority, tt.expected.BundlesPriority)
			}
		})
	}
}

func TestConfigValidate(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name: "valid local only config",
			config: Config{
				TemplatesSource:   "local",
				BundlesSource:     "local",
				TemplatesLocalDir: "templates",
				BundlesLocalDir:   "bundles",
				CacheDir:          ".cache",
				TemplatesPriority: "local-first",
				BundlesPriority:   "local-first",
			},
			wantErr: false,
		},
		{
			name: "valid remote only config",
			config: Config{
				TemplatesSource:    "remote",
				BundlesSource:      "remote",
				TemplatesRemoteURL: "https://example.com/templates",
				BundlesRemoteURL:   "https://example.com/bundles",
				CacheDir:           ".cache",
				TemplatesPriority:  "remote-first",
				BundlesPriority:    "remote-first",
			},
			wantErr: false,
		},
		{
			name: "valid both sources config",
			config: Config{
				TemplatesSource:    "both",
				BundlesSource:      "both",
				TemplatesLocalDir:  "templates",
				TemplatesRemoteURL: "https://example.com/templates",
				BundlesLocalDir:    "bundles",
				BundlesRemoteURL:   "https://example.com/bundles",
				CacheDir:           ".cache",
				TemplatesPriority:  "local-first",
				BundlesPriority:    "remote-first",
			},
			wantErr: false,
		},
		{
			name: "invalid templates source",
			config: Config{
				TemplatesSource: "invalid",
				BundlesSource:   "local",
				BundlesLocalDir: "bundles",
				CacheDir:        ".cache",
			},
			wantErr: true,
		},
		{
			name: "invalid bundles source",
			config: Config{
				TemplatesSource:   "local",
				TemplatesLocalDir: "templates",
				BundlesSource:     "invalid",
				CacheDir:          ".cache",
			},
			wantErr: true,
		},
		{
			name: "missing local dir for local source",
			config: Config{
				TemplatesSource:   "local",
				BundlesSource:     "local",
				BundlesLocalDir:   "bundles",
				CacheDir:          ".cache",
				TemplatesPriority: "local-first",
				BundlesPriority:   "local-first",
			},
			wantErr: true,
		},
		{
			name: "missing remote URL for remote source",
			config: Config{
				TemplatesSource:   "remote",
				BundlesSource:     "remote",
				BundlesRemoteURL:  "https://example.com/bundles",
				CacheDir:          ".cache",
				TemplatesPriority: "remote-first",
				BundlesPriority:   "remote-first",
			},
			wantErr: true,
		},
		{
			name: "invalid templates remote URL",
			config: Config{
				TemplatesSource:    "remote",
				BundlesSource:      "remote",
				TemplatesRemoteURL: "ftp://example.com/templates",
				BundlesRemoteURL:   "https://example.com/bundles",
				CacheDir:           ".cache",
			},
			wantErr: true,
		},
		{
			name: "path traversal in local dir",
			config: Config{
				TemplatesSource:   "local",
				BundlesSource:     "local",
				TemplatesLocalDir: "../../../etc/passwd",
				BundlesLocalDir:   "bundles",
				CacheDir:          ".cache",
			},
			wantErr: true,
		},
		{
			name: "invalid templates priority",
			config: Config{
				TemplatesSource:   "local",
				BundlesSource:     "local",
				TemplatesLocalDir: "templates",
				BundlesLocalDir:   "bundles",
				CacheDir:          ".cache",
				TemplatesPriority: "invalid",
				BundlesPriority:   "local-first",
			},
			wantErr: true,
		},
		{
			name: "invalid bundles priority",
			config: Config{
				TemplatesSource:   "local",
				BundlesSource:     "local",
				TemplatesLocalDir: "templates",
				BundlesLocalDir:   "bundles",
				CacheDir:          ".cache",
				TemplatesPriority: "local-first",
				BundlesPriority:   "invalid",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Config.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestConfigGetters(t *testing.T) {
	cfg := Config{
		TemplatesSource:   "both",
		BundlesSource:     "remote",
		TemplatesPriority: "remote-first",
		BundlesPriority:   "local-first",
	}

	if got := cfg.GetTemplatesSource(); got != SourceTypeBoth {
		t.Errorf("GetTemplatesSource() = %v, want %v", got, SourceTypeBoth)
	}

	if got := cfg.GetBundlesSource(); got != SourceTypeRemote {
		t.Errorf("GetBundlesSource() = %v, want %v", got, SourceTypeRemote)
	}

	if got := cfg.GetTemplatesPriority(); got != PriorityRemoteFirst {
		t.Errorf("GetTemplatesPriority() = %v, want %v", got, PriorityRemoteFirst)
	}

	if got := cfg.GetBundlesPriority(); got != PriorityLocalFirst {
		t.Errorf("GetBundlesPriority() = %v, want %v", got, PriorityLocalFirst)
	}
}
