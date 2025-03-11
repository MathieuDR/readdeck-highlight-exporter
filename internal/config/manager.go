package config

import (
  "path/filepath"
  "github.com/adrg/xdg"
)

func GetConfigPath() string {
  config := xdg.ConfigHome
  return filepath.Join(config, "readdeck-exporter", "config.yaml")
}
