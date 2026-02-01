package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
	"github.com/wizardopstech/conjure/internal/security"
)

const (
	DefaultCacheDir = ".conjure"
)

type SourceType string

const (
	SourceTypeLocal  SourceType = "local"
	SourceTypeRemote SourceType = "remote"
	SourceTypeBoth   SourceType = "both"
)

type PriorityOrder string

const (
	PriorityLocalFirst  PriorityOrder = "local-first"
	PriorityRemoteFirst PriorityOrder = "remote-first"
)

type Config struct {
	TemplatesSource    string `mapstructure:"templates_source"`
	BundlesSource      string `mapstructure:"bundles_source"`
	TemplatesLocalDir  string `mapstructure:"templates_local_dir"`
	TemplatesRemoteURL string `mapstructure:"templates_remote_url"`
	BundlesLocalDir    string `mapstructure:"bundles_local_dir"`
	BundlesRemoteURL   string `mapstructure:"bundles_remote_url"`
	CacheDir           string `mapstructure:"cache_dir"`
	TemplatesPriority  string `mapstructure:"templates_priority"`
	BundlesPriority    string `mapstructure:"bundles_priority"`
}

func LoadConfig() (*Config, error) {
	var cfg Config

	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	cfg.applyDefaults()

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func (c *Config) applyDefaults() {
	if c.TemplatesSource == "" {
		if c.TemplatesRemoteURL != "" {
			c.TemplatesSource = string(SourceTypeBoth)
		} else {
			c.TemplatesSource = string(SourceTypeLocal)
		}
	}

	if c.BundlesSource == "" {
		if c.BundlesRemoteURL != "" {
			c.BundlesSource = string(SourceTypeBoth)
		} else {
			c.BundlesSource = string(SourceTypeLocal)
		}
	}

	if c.CacheDir == "" {
		c.CacheDir = DefaultCacheDir
	}

	if strings.HasPrefix(c.CacheDir, "~/") {
		homeDir, err := os.UserHomeDir()
		if err == nil {
			c.CacheDir = filepath.Join(homeDir, c.CacheDir[2:])
		}
	}

	if !filepath.IsAbs(c.CacheDir) {
		absPath, err := filepath.Abs(c.CacheDir)
		if err == nil {
			c.CacheDir = absPath
		}
	}

	if c.TemplatesPriority == "" {
		c.TemplatesPriority = string(PriorityLocalFirst)
	}
	if c.BundlesPriority == "" {
		c.BundlesPriority = string(PriorityLocalFirst)
	}
}

func (c *Config) Validate() error {
	if err := c.validateSource(c.TemplatesSource, "templates"); err != nil {
		return err
	}

	if err := c.validateSource(c.BundlesSource, "bundles"); err != nil {
		return err
	}

	if c.GetTemplatesSource() == SourceTypeLocal || c.GetTemplatesSource() == SourceTypeBoth {
		if c.TemplatesLocalDir == "" {
			return fmt.Errorf("templates_local_dir is required when templates_source is 'local' or 'both'")
		}
		if err := security.ValidatePathSafety(c.TemplatesLocalDir); err != nil {
			return fmt.Errorf("invalid templates_local_dir: %w", err)
		}
	}

	if c.GetTemplatesSource() == SourceTypeRemote || c.GetTemplatesSource() == SourceTypeBoth {
		if c.TemplatesRemoteURL == "" {
			return fmt.Errorf("templates_remote_url is required when templates_source is 'remote' or 'both'")
		}
		if err := security.ValidateURL(c.TemplatesRemoteURL); err != nil {
			return fmt.Errorf("invalid templates_remote_url: %w", err)
		}
	}

	if c.GetBundlesSource() == SourceTypeLocal || c.GetBundlesSource() == SourceTypeBoth {
		if c.BundlesLocalDir == "" {
			return fmt.Errorf("bundles_local_dir is required when bundles_source is 'local' or 'both'")
		}
		if err := security.ValidatePathSafety(c.BundlesLocalDir); err != nil {
			return fmt.Errorf("invalid bundles_local_dir: %w", err)
		}
	}

	if c.GetBundlesSource() == SourceTypeRemote || c.GetBundlesSource() == SourceTypeBoth {
		if c.BundlesRemoteURL == "" {
			return fmt.Errorf("bundles_remote_url is required when bundles_source is 'remote' or 'both'")
		}
		if err := security.ValidateURL(c.BundlesRemoteURL); err != nil {
			return fmt.Errorf("invalid bundles_remote_url: %w", err)
		}
	}

	if err := security.ValidatePathSafety(c.CacheDir); err != nil {
		return fmt.Errorf("invalid cache_dir: %w", err)
	}

	if err := c.validatePriority(c.TemplatesPriority, "templates_priority"); err != nil {
		return err
	}
	if err := c.validatePriority(c.BundlesPriority, "bundles_priority"); err != nil {
		return err
	}

	return nil
}

func (c *Config) validateSource(source, name string) error {
	validSources := []string{string(SourceTypeLocal), string(SourceTypeRemote), string(SourceTypeBoth)}
	for _, valid := range validSources {
		if source == valid {
			return nil
		}
	}
	return fmt.Errorf("invalid %s_source: %s (must be 'local', 'remote', or 'both')", name, source)
}

func (c *Config) validatePriority(priority, name string) error {
	validPriorities := []string{string(PriorityLocalFirst), string(PriorityRemoteFirst)}
	for _, valid := range validPriorities {
		if priority == valid {
			return nil
		}
	}
	return fmt.Errorf("invalid %s: %s (must be 'local-first' or 'remote-first')", name, priority)
}

func (c *Config) GetTemplatesSource() SourceType {
	return SourceType(c.TemplatesSource)
}

func (c *Config) GetBundlesSource() SourceType {
	return SourceType(c.BundlesSource)
}

func (c *Config) GetTemplatesPriority() PriorityOrder {
	return PriorityOrder(c.TemplatesPriority)
}

func (c *Config) GetBundlesPriority() PriorityOrder {
	return PriorityOrder(c.BundlesPriority)
}
