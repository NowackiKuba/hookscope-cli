package auth

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/NowackiKuba/hookscope-cli/internal/config"
)

type Config struct {
	Token string `json:"token"`
}

func SaveToken(token string) error {
	path := config.ConfigPath()
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	cfg := Config{Token: token}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0600)
}

func LoadToken() (string, error) {
	path := config.ConfigPath()
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return "", err
	}
	return cfg.Token, nil
}
