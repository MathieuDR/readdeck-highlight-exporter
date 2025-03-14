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

	parent := filepath.Base(path)
	if parent != "readdeck-exporter" {
		t.Errorf("Config should be in readdeck-exporter directory, got %s in %s", parent, path)
	}

	baseDir := filepath.Dir(path)
	if baseDir != expectedBaseDir {
		t.Errorf("Base directory should be %s, got %s in %s", expectedBaseDir, baseDir, path)
	}
}
