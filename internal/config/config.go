package config

import (
	"github.com/spf13/viper"
)

type Config struct {
	TemplatesDir string `mapstructure:"templates_dir"`
	BundlesDir   string `mapstructure:"bundles_dir"`
}

func LoadConfig() (*Config, error) {
	var cfg Config

	// Unmarshal viper config into struct
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
