package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const (
	nestDir    = ".nest"
	configFile = "config.json"
)

// UserConfig holds user-level preferences stored at ~/.nest/config.json.
type UserConfig struct {
	Lang string `json:"lang"` // "zh" or "en"
}

func defaultConfig() *UserConfig {
	return &UserConfig{Lang: "zh"}
}

func configPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("get home dir error: %s", err)
	}
	return filepath.Join(home, nestDir, configFile), nil
}

// Load reads ~/.nest/config.json. Falls back to defaults on any error.
func Load() *UserConfig {
	p, err := configPath()
	if err != nil {
		return defaultConfig()
	}
	data, err := os.ReadFile(p)
	if err != nil {
		return defaultConfig()
	}
	cfg := &UserConfig{}
	if err = json.Unmarshal(data, cfg); err != nil {
		return defaultConfig()
	}
	if cfg.Lang != "zh" && cfg.Lang != "en" {
		cfg.Lang = "zh"
	}
	return cfg
}

// Save writes the config to ~/.nest/config.json.
func Save(cfg *UserConfig) error {
	p, err := configPath()
	if err != nil {
		return err
	}
	dir := filepath.Dir(p)
	if err = os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("create config dir error: %s", err)
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal config error: %s", err)
	}
	return os.WriteFile(p, data, 0644)
}
