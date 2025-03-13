package config

import (
  "path/filepath"
  "testing"
)

func TestXDGConfigHome (t *testing.T) {
  xgdDir := XDGDir{}
  path := xgdDir.ConfigHome()

  if path == "" {
    t.Error("Path should not be empty, bet received empty string")
  }

  // Would only work on UNIX without customization, so lets skip it.
  // baseDir := filepath.Base(path)
  // if baseDir != ".config" {
  //   t.Errorf("Base directory should be %s, got %s", ".config", baseDir)
  // }
}
