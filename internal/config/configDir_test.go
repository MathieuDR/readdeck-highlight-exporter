package config

import (
	"testing"
)

func TestXDGConfigHome(t *testing.T) {
	xgdDir := XDGDir{}
	path := xgdDir.ConfigHome()

	if path == "" {
		t.Error("Path should not be empty, bet received empty string")
	}
}
