package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const (
	AppName = "get-repo"
	ConfigFileName = "config.json"
	EnvConfigPath = "GET_REPO_CONFIG"
)

// Config holds the application's configuration.
type Config struct {
	CodebasesPath string `json:"codebases_path"`
	ConfigPath    string `json:"-"` // Path where this config was loaded from
}

// Load reads the configuration file and returns a Config struct.
func Load() (Config, error) {
	var cfg Config

	cfgPath, err := GetConfigPath()
	if err != nil {
		return cfg, err
	}

	if cfgPath == "" {
		// No config path found
		return cfg, nil
	}

	if _, err := os.Stat(cfgPath); os.IsNotExist(err) {
		// Config file does not exist, return empty config
		return cfg, nil
	}

	file, err := os.ReadFile(cfgPath)
	if err != nil {
		return cfg, err
	}

	if err := json.Unmarshal(file, &cfg); err != nil {
		return cfg, err
	}

	cfg.ConfigPath = cfgPath
	return cfg, nil
}

// SaveTo writes the configuration to a specific path
func (c Config) SaveTo(path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0750); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	c.ConfigPath = path
	return nil
}

// Save writes the configuration to the config file.
func (c Config) Save() error {
	cfgPath, err := GetConfigPath()
	if err != nil {
		return err
	}
	
	if cfgPath == "" {
		// No config path set, use default
		cfgPath, err = getDefaultConfigPath()
		if err != nil {
			return err
		}
	}

	return c.SaveTo(cfgPath)
}

// GetConfigPath returns the configuration file path
// Priority: 1. Environment variable, 2. Default location
func GetConfigPath() (string, error) {
	// Check environment variable first
	if envPath := os.Getenv(EnvConfigPath); envPath != "" {
		return envPath, nil
	}

	// Check if default config exists
	defaultPath, err := getDefaultConfigPath()
	if err != nil {
		return "", err
	}

	if _, err := os.Stat(defaultPath); err == nil {
		return defaultPath, nil
	}

	// No config found
	return "", nil
}

func getDefaultConfigPath() (string, error) {
	cfgDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(cfgDir, AppName, ConfigFileName), nil
}

// IsFirstRun checks if this is the first run (no config exists)
func IsFirstRun() bool {
	cfgPath, err := GetConfigPath()
	if err != nil || cfgPath == "" {
		return true
	}
	
	_, err = os.Stat(cfgPath)
	return os.IsNotExist(err)
}
