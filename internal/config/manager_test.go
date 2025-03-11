package config

import (
  "path/filepath"
  "testing"
)

type MockDir struct {
  TempDir string
}

func (m MockDir) ConfigHome() string {
  return m.TempDir
}

func NewMockDir(t *testing.T) ConfigDir {
  dir := t.TempDir()

  return MockDir{
    TempDir: dir,
  }
}

func TestConfigHome(t *testing.T) {
  provider := NewMockDir(t)
  expectedBaseDir := provider.ConfigHome()

  path := ConfigPath(provider)

  _, file := filepath.Split(path)
  if file != "config.yaml" {
    t.Errorf("Expected filename to be config.yaml, got %s", file)
  }

  parent := filepath.Base(filepath.Dir(path))
  if parent != "readdeck-exporter" {
    t.Errorf("Config should be in readdeck-exporter directory, got %s", parent)
  }
  
  baseDir := filepath.Dir(filepath.Dir(path))
  if baseDir != expectedBaseDir {
    t.Errorf("Base directory should be %s, got %s", expectedBaseDir, baseDir)
  }
}
