package config

import (
	"os"
	"path/filepath"
)

func ConfigPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".hookscope", "config.json")
}
