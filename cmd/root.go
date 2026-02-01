/*
Copyright © 2025 WizardOps LLC headwizard@wizardops.dev
*/
package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/wizardopstech/conjure/cmd/bundle"
	"github.com/wizardopstech/conjure/cmd/list"
	"github.com/wizardopstech/conjure/cmd/repo"
	"github.com/wizardopstech/conjure/cmd/template"
)

var (
	cfgFile     string
	showVersion bool
	versionInfo string
	commitInfo  string
	dateInfo    string
)

var rootCmd = &cobra.Command{
	Use:   "conjure",
	Short: "Template and configuration management tool",
	Long: `Conjure is a CLI tool for managing and generating files from templates.

Conjure allows you to create reusable templates and bundles with variable
substitution, metadata, and interactive prompts for consistent configuration
management across your infrastructure.

Examples:
  conjure template deployment.yaml -o ./deployment.yaml
  conjure bundle kubernetes_deployment -o ./k8s/
  conjure list templates
  conjure list bundles`,
	Run: func(cmd *cobra.Command, args []string) {
		if showVersion {
			fmt.Printf("conjure version %s\n", versionInfo)
			fmt.Printf("commit: %s\n", commitInfo)
			fmt.Printf("built: %s\n", dateInfo)
			return
		}
		cmd.Help()
	},
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if showVersion {
			fmt.Printf("conjure version %s\n", versionInfo)
			fmt.Printf("commit: %s\n", commitInfo)
			fmt.Printf("built: %s\n", dateInfo)
			os.Exit(0)
		}
	},
}

func Execute(version, commit, date string) {
	versionInfo = version
	commitInfo = commit
	dateInfo = date

	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfigIfNeeded)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.conjure.yaml)")
	rootCmd.PersistentFlags().BoolVarP(&showVersion, "version", "v", false, "display version information")

	rootCmd.AddCommand(list.ListCmd)
	rootCmd.AddCommand(template.TemplateCmd)
	rootCmd.AddCommand(bundle.BundleCmd)
	rootCmd.AddCommand(repo.RepoCmd)
}

func initConfigIfNeeded() {
	if !showVersion {
		initConfig()
	}
}

func initConfig() {
	if cfgFile != "" {
		resolvedPath, err := resolveConfigPath(cfgFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error resolving config path: %v\n", err)
			os.Exit(1)
		}

		viper.SetConfigFile(resolvedPath)
	} else {
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".conjure")
	}

	viper.AutomaticEnv()
	viper.SetEnvPrefix("CONJURE") // prefix CONJURE_ before env vars e.g. CONJURE_TEMPLATES_LOCAL_DIR

	// Bind environment variables for config fields
	_ = viper.BindEnv("TEMPLATES_LOCAL_DIR")
	_ = viper.BindEnv("BUNDLES_LOCAL_DIR")
	_ = viper.BindEnv("TEMPLATES_REMOTE_URL")
	_ = viper.BindEnv("BUNDLES_REMOTE_URL")
	_ = viper.BindEnv("TEMPLATES_SOURCE")
	_ = viper.BindEnv("BUNDLES_SOURCE")
	_ = viper.BindEnv("CACHE_DIR")
	_ = viper.BindEnv("TEMPLATES_PRIORITY")
	_ = viper.BindEnv("BUNDLES_PRIORITY")

	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}

func resolveConfigPath(path string) (string, error) {
	if strings.HasPrefix(path, "~") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("failed to get user home directory: %w", err)
		}
		if path == "~" {
			path = home
		} else if strings.HasPrefix(path, "~/") || strings.HasPrefix(path, "~\\") {
			path = filepath.Join(home, path[2:])
		}
	}

	if !filepath.IsAbs(path) {
		cwd, err := os.Getwd()
		if err != nil {
			return "", fmt.Errorf("failed to get current working directory: %w", err)
		}
		path = filepath.Join(cwd, path)
	}

	path = filepath.Clean(path)

	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("config file does not exist: %s", path)
		}
		return "", fmt.Errorf("failed to access config file: %w", err)
	}

	return path, nil
}
