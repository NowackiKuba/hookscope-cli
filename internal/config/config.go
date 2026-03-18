package config

import (
	"os"
	"path/filepath"
)

const DefaultAPIURL = "https://api.hookscope.dev"
const ConfigDir = ".hookscope"
const ConfigFile = "config.json"

func ConfigPath() string {
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		// Best-effort fallback; callers will error on use if unusable.
		return filepath.Join(string(os.PathSeparator), ConfigDir, ConfigFile)
	}
	return filepath.Join(home, ConfigDir, ConfigFile)
}

func EnsureConfigDir() error {
	dir := filepath.Dir(ConfigPath())
	return os.MkdirAll(dir, 0o700)
}
