package config

import (
    "fmt"
    "os"
    "path/filepath"
    
    "github.com/pelletier/go-toml/v2"
)

func LoadConfig(path string) (*Config, error) {
    if _, err := os.Stat(path); os.IsNotExist(err) {
        return nil, fmt.Errorf("config file not found: %s", path)
    }
    
    data, err := os.ReadFile(path)
    if err != nil {
        return nil, fmt.Errorf("failed to read config file: %w", err)
    }
    
    cfg := DefaultConfig()
    if err := toml.Unmarshal(data, cfg); err != nil {
        return nil, fmt.Errorf("failed to parse TOML: %w", err)
    }
    
    return cfg, nil
}

func LoadOrCreateConfig(path string) (*Config, error) {
    cfg, err := LoadConfig(path)
    if err == nil {
        return cfg, nil
    }
    
    if os.IsNotExist(err) {
        cfg = DefaultConfig()
        if err := SaveConfig(path, cfg); err != nil {
            return nil, fmt.Errorf("failed to create default config: %w", err)
        }
        return cfg, nil
    }
    
    return nil, err
}

func GetDefaultConfigPath() string {
    homeDir, _ := os.UserHomeDir()
    return filepath.Join(homeDir, ".config", "ocsession", "config.toml")
}
