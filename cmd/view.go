package cmd

import (
	"fmt"
	"strings"

	"github.com/mathieudr/readdeck-highlight-exporter/internal/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// Common function to show configuration
func showConfig() {
	// Get default settings for comparison
	defaults := config.DefaultSettings()

	fmt.Println("Current Configuration:")
	fmt.Println("======================")

	// Readdeck settings
	fmt.Println("\nReaddeck:")
	fmt.Printf("  Base URL:           %s\n", viper.GetString("readdeck.base_url"))

	// Don't show full token for security, just a masked version
	token := viper.GetString("readdeck.token")
	if token != "" {
		masked := token
		if len(token) > 8 {
			masked = token[:4] + strings.Repeat("*", len(token)-8) + token[len(token)-4:]
		} else if len(token) > 0 {
			masked = strings.Repeat("*", len(token))
		}
		fmt.Printf("  Token:              %s\n", masked)
	} else {
		fmt.Printf("  Token:              <not set>\n")
	}

	// For values with defaults, indicate if using default
	bpp := viper.GetInt("readdeck.bookmarks_per_page")
	defaultIndicator := ""
	if bpp == defaults.Readdeck.BookmarksPerPage {
		defaultIndicator = " (default)"
	}
	fmt.Printf("  Bookmarks per page: %d%s\n", bpp, defaultIndicator)

	timeout := viper.GetDuration("readdeck.request_timeout")
	defaultIndicator = ""
	if timeout == defaults.Readdeck.RequestTimeout {
		defaultIndicator = " (default)"
	}
	fmt.Printf("  Request timeout:    %s%s\n", timeout, defaultIndicator)

	// Export settings
	fmt.Println("\nExport:")
	fmt.Printf("  Fleeting path:      %s\n", viper.GetString("export.fleeting_path"))

	// Display config file location
	fmt.Printf("\nConfiguration file: %s\n", viper.ConfigFileUsed())
}

// viewCmd represents the view command
var viewCmd = &cobra.Command{
	Use:   "view",
	Short: "View current configuration",
	Long: `Display the current configuration settings for readdeck-highlight-exporter.
	
This will show all configured values and their defaults.`,
	Run: func(cmd *cobra.Command, args []string) {
		showConfig()
	},
}

func init() {
	configCmd.AddCommand(viewCmd)
}
