package config

import (
  "github.com/adrg/xdg"
)

type ConfigDir interface {
  ConfigHome() string
}

type XDGDir struct{}

func (p XDGDir) ConfigHome() string {
  return xdg.ConfigHome
}
