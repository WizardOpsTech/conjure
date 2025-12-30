/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/thesudoYT/conjure/cmd/bundle"
	"github.com/thesudoYT/conjure/cmd/list"
	"github.com/thesudoYT/conjure/cmd/template"
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands
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
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig) // load the config into the config struct at runtime

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.conjure.yaml)")

	rootCmd.AddCommand(list.ListCmd)
	rootCmd.AddCommand(template.TemplateCmd)
	rootCmd.AddCommand(bundle.BundleCmd)

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".conjure" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".conjure")
	}

	viper.AutomaticEnv()           // read in environment variables that match
	viper.SetEnvPrefix("CONJURE")  // prefix CONJURE_ before env vars e.g. CONJURE_TEMPLATES_DIR
	viper.BindEnv("TEMPLATES_DIR") // path to a directory containing templates.

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}
