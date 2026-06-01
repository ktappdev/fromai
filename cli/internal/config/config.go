package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

type Config struct {
	APIKey  string `toml:"api_key"`
	BaseURL string `toml:"base_url"`
}

func Path() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("cannot find home directory: %w", err)
	}
	return filepath.Join(home, ".config", "fromai", "config.toml"), nil
}

func Load() (*Config, error) {
	path, err := Path()
	if err != nil {
		return nil, err
	}
	var cfg Config
	if _, err := toml.DecodeFile(path, &cfg); err != nil {
		if os.IsNotExist(err) {
			return &Config{}, nil
		}
		return nil, fmt.Errorf("failed to read config: %w", err)
	}
	return &cfg, nil
}

func Save(cfg *Config) error {
	path, err := Path()
	if err != nil {
		return err
	}
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("failed to create config file: %w", err)
	}
	defer f.Close()
	if err := toml.NewEncoder(f).Encode(cfg); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}
	// Ensure permissions are 0600 even if file existed with different perms
	if err := os.Chmod(path, 0600); err != nil {
		return fmt.Errorf("failed to set config permissions: %w", err)
	}
	return nil
}

func Delete() error {
	path, err := Path()
	if err != nil {
		return err
	}
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete config: %w", err)
	}
	return nil
}
