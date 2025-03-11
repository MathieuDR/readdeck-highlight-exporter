package config

import (
  "os"
  "path/filepath"
  "testing"
)

func TestConfigPath(t *testing.T) {
  t.Setenv("XDG_CONFIG_HOME", t.TempDir())
  configPath := os.Getenv("XDG_CONFIG_HOME")

  path := GetConfigPath()

  expected := filepath.Join(configPath, "readdeck-exporter", "config.yaml")
  t.Logf("I'm trying this %s", expected)

  if path != expected {
    t.Errorf("Expected path %s, got %s", expected, path)
  }
}
