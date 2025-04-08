package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

type Settings struct {
	Readdeck ReaddeckSettings `mapstructure:"readdeck"`
	Export   ExportSettings   `mapstructure:"export"`
}

type ReaddeckSettings struct {
	BaseURL          string        `mapstructure:"base_url"`
	Token            string        `mapstructure:"token"`
	BookmarksPerPage int           `mapstructure:"bookmarks_per_page"`
	RequestTimeout   time.Duration `mapstructure:"request_timeout"`
}

type ExportSettings struct {
	FleetingPath string `mapstructure:"fleeting_path"`
}

func DefaultSettings() Settings {
	return Settings{
		Readdeck: ReaddeckSettings{
			BookmarksPerPage: 100,
			RequestTimeout:   time.Second * 30,
		},
	}
}

func LoadAndValidate() (Settings, error) {
	var settings Settings
	if err := viper.Unmarshal(&settings); err != nil {
		return Settings{}, err
	}

	// Apply defaults for any unset values
	defaults := DefaultSettings()
	if settings.Readdeck.BookmarksPerPage == 0 {
		settings.Readdeck.BookmarksPerPage = defaults.Readdeck.BookmarksPerPage
	} else if settings.Readdeck.BookmarksPerPage < 10 {
		return Settings{}, fmt.Errorf("bookmarks_per_page must be at least 10")
	}

	if settings.Readdeck.RequestTimeout == 0 {
		settings.Readdeck.RequestTimeout = defaults.Readdeck.RequestTimeout
	}

	// Validate required fields
	if settings.Readdeck.BaseURL == "" {
		return Settings{}, fmt.Errorf("readdeck.base_url is required")
	}

	if settings.Readdeck.Token == "" {
		return Settings{}, fmt.Errorf("readdeck.token is required")
	}

	if settings.Export.FleetingPath == "" {
		return Settings{}, fmt.Errorf("export.fleeting_path is required")
	}

	return settings, nil
}
