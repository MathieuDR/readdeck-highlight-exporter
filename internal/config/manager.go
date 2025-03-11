package config

import (
  "path/filepath"
)

func ConfigPath(dir ConfigDir) string {
  configHome := dir.ConfigHome()
  return filepath.Join(configHome, "readdeck-exporter", "config.yaml")
}
