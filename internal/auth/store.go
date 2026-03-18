package auth

import (
	"encoding/json"
	"errors"
	"os"

	"github.com/NowackiKuba/hookscope-cli/internal/config"
)

type Credentials struct {
	Token  string `json:"token"`
	APIURL string `json:"api_url"`
}

func Save(creds Credentials) error {
	if creds.APIURL == "" {
		creds.APIURL = config.DefaultAPIURL
	}
	if err := config.EnsureConfigDir(); err != nil {
		return err
	}
	data, err := json.MarshalIndent(creds, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(config.ConfigPath(), data, 0o600)
}

func Load() (Credentials, error) {
	data, err := os.ReadFile(config.ConfigPath())
	if err != nil {
		if os.IsNotExist(err) {
			return Credentials{}, os.ErrNotExist
		}
		return Credentials{}, err
	}
	var creds Credentials
	if err := json.Unmarshal(data, &creds); err != nil {
		return Credentials{}, err
	}
	if creds.APIURL == "" {
		creds.APIURL = config.DefaultAPIURL
	}
	if creds.Token == "" {
		return Credentials{}, errors.New("missing token in config")
	}
	return creds, nil
}

func Clear() error {
	err := os.Remove(config.ConfigPath())
	if err != nil && os.IsNotExist(err) {
		return nil
	}
	return err
}
