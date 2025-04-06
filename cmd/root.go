package cmd

import (
	"fmt"
	"os"

	"github.com/mathieudr/readdeck-highlight-exporter/internal/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "readdeck-highlight-exporter",
	Short: "Export Readdeck highlights to Zettelkasten notes",
	Long: `Readdeck Highlight Exporter is a CLI tool that exports highlights 
from Readdeck (a read-it-later service) to your Zettelkasten note-taking system.

The tool reads from Readdeck without modifying it and tracks exported 
highlights through the generated notes themselves, ensuring idempotent operation.

To get started, run the 'config' command to set up your configuration:
  readdeck-highlight-exporter config --help`,
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
	cobra.OnInitialize(initConfig)

	// Config file flag
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $XDG_CONFIG_HOME/readdeck-exporter/settings.yaml)")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	// Set up defaults from the Settings struct
	defaults := config.DefaultSettings()
	viper.SetDefault("readdeck.bookmarks_per_page", defaults.Readdeck.BookmarksPerPage)
	viper.SetDefault("readdeck.request_timeout", defaults.Readdeck.RequestTimeout)

	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find the standard config location.
		configDir := config.ConfigHome()

		viper.AddConfigPath(configDir)
		viper.SetConfigType("yaml")
		viper.SetConfigName("settings")
	}

	// Read environment variables with prefix READDECK_EXPORTER
	viper.SetEnvPrefix("readdeck_exporter")
	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}
