package config

import (
	"path/filepath"

	"github.com/adrg/xdg"
)

const Parent = "readdeck-exporter"

type Directories interface {
	ConfigHome() string
	StateHome() string
}

func ConfigHome() string {
	return filepath.Join(xdg.ConfigHome, Parent)

}

func StateHome() string {
	return filepath.Join(xdg.StateHome, Parent)
}
