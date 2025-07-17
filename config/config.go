package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

const (
	AppName = "get-repo"
	ConfigFileName = "config.json"
)

// Config holds the application's configuration.
type Config struct {
	CodebasesPath string `json:"codebases_path"`
}

// Load reads the configuration file and returns a Config struct.
func Load() (Config, error) {
	var cfg Config

	cfgPath, err := getConfigPath()
	if err != nil {
		return cfg, err
	}

	if _, err := os.Stat(cfgPath); os.IsNotExist(err) {
		// Config file does not exist, return default config.
		return defaultConfig(), nil
	}

	file, err := os.ReadFile(cfgPath)
	if err != nil {
		return cfg, err
	}

	if err := json.Unmarshal(file, &cfg); err != nil {
		return cfg, err
	}

	return cfg, nil
}

// Save writes the configuration to the config file.
func (c Config) Save() error {
	cfgPath, err := getConfigPath()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(cfgPath), 0750); err != nil {
		return err
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(cfgPath, data, 0644)
}

func defaultConfig() Config {
	return Config{}
}

func getConfigPath() (string, error) {
	cfgDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(cfgDir, AppName, ConfigFileName), nil
}
