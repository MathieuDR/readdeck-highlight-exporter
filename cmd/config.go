package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/mathieudr/readdeck-highlight-exporter/internal/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	baseURL          string
	token            string
	bookmarksPerPage int
	timeout          time.Duration
	fleetingPath     string
)

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Configure readdeck-highlight-exporter",
	Long: `Configure readdeck-highlight-exporter settings.
	
Without arguments, this command shows the current configuration.
With flags, it creates or updates the configuration settings.

Required fields for initial setup: readdeck.base_url, readdeck.token, and export.fleeting_path.

Examples:
  # Show current configuration
  readdeck-highlight-exporter config
  
  # Create initial configuration
  readdeck-highlight-exporter config --base-url=https://read.example.com --token=yourtoken --fleeting-path=/path/to/notes
  
  # Update timeout
  readdeck-highlight-exporter config --timeout=45s
  
  # Revert to default bookmarks per page
  readdeck-highlight-exporter config --unset=readdeck.bookmarks_per_page
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// If no flags are set, show the current configuration (like 'view')
		if !cmd.Flags().Changed("base-url") &&
			!cmd.Flags().Changed("token") &&
			!cmd.Flags().Changed("bookmarks-per-page") &&
			!cmd.Flags().Changed("timeout") &&
			!cmd.Flags().Changed("fleeting-path") {
			showConfig()
			return nil
		}

		// Initialize defaults if config doesn't exist
		if !configExists() {
			defaults := config.DefaultSettings()
			viper.SetDefault("readdeck.bookmarks_per_page", defaults.Readdeck.BookmarksPerPage)
			viper.SetDefault("readdeck.request_timeout", defaults.Readdeck.RequestTimeout)
		}

		// Set new values from flags
		if cmd.Flags().Changed("base-url") {
			viper.Set("readdeck.base_url", baseURL)
		}
		if cmd.Flags().Changed("token") {
			viper.Set("readdeck.token", token)
		}
		if cmd.Flags().Changed("bookmarks-per-page") {
			if bookmarksPerPage < 10 {
				return fmt.Errorf("bookmarks-per-page must be at least 10")
			}
			viper.Set("readdeck.bookmarks_per_page", bookmarksPerPage)
		}
		if cmd.Flags().Changed("timeout") {
			viper.Set("readdeck.request_timeout", timeout)
		}
		if cmd.Flags().Changed("fleeting-path") {
			viper.Set("export.fleeting_path", fleetingPath)
		}

		// Validate required fields for a new configuration
		if !configExists() {
			if viper.GetString("readdeck.base_url") == "" {
				return fmt.Errorf("readdeck.base_url is required for initial setup")
			}
			if viper.GetString("readdeck.token") == "" {
				return fmt.Errorf("readdeck.token is required for initial setup")
			}
			if viper.GetString("export.fleeting_path") == "" {
				return fmt.Errorf("export.fleeting_path is required for initial setup")
			}
		} else {
			// For existing configs, warn about empty required fields but don't error
			if cmd.Flags().Changed("base-url") && baseURL == "" {
				fmt.Println("Warning: Setting empty base-url")
			}
			if cmd.Flags().Changed("token") && token == "" {
				fmt.Println("Warning: Setting empty token")
			}
			if cmd.Flags().Changed("fleeting-path") && fleetingPath == "" {
				fmt.Println("Warning: Setting empty fleeting-path")
			}
		}

		// Create config directory if it doesn't exist
		configDir := config.ConfigHome()
		if err := os.MkdirAll(configDir, 0755); err != nil {
			return fmt.Errorf("failed to create config directory: %w", err)
		}

		// Save configuration
		configFile := configDir + "/settings.yaml"
		if err := viper.WriteConfigAs(configFile); err != nil {
			return fmt.Errorf("failed to save configuration: %w", err)
		}

		fmt.Printf("Configuration saved to %s\n", configFile)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(configCmd)

	// Define flags
	configCmd.Flags().StringVar(&baseURL, "base-url", "", "Readdeck base URL")
	configCmd.Flags().StringVar(&token, "token", "", "Readdeck API token")
	configCmd.Flags().IntVar(&bookmarksPerPage, "bookmarks-per-page", 100, "Number of bookmarks to fetch per request (min 10)")
	configCmd.Flags().DurationVar(&timeout, "timeout", 30*time.Second, "HTTP request timeout")
	configCmd.Flags().StringVar(&fleetingPath, "fleeting-path", "", "Path to fleeting notes directory")
}

// configExists checks if a config file exists
func configExists() bool {
	configFile := viper.ConfigFileUsed()
	if configFile == "" {
		return false
	}
	_, err := os.Stat(configFile)
	return err == nil
}

// GetConfig loads configuration into Settings struct
func GetConfig() (config.Settings, error) {
	var settings config.Settings
	if err := viper.Unmarshal(&settings); err != nil {
		return config.Settings{}, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Merge with defaults for any unset values
	defaults := config.DefaultSettings()
	if settings.Readdeck.BookmarksPerPage == 0 {
		settings.Readdeck.BookmarksPerPage = defaults.Readdeck.BookmarksPerPage
	}
	if settings.Readdeck.RequestTimeout == 0 {
		settings.Readdeck.RequestTimeout = defaults.Readdeck.RequestTimeout
	}

	return settings, nil
}
