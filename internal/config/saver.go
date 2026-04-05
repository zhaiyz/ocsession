package config

import (
    "os"
    "path/filepath"
    
    "github.com/pelletier/go-toml/v2"
)

// SaveConfig writes the configuration to a TOML file at the specified path.
// It creates any necessary parent directories.
func SaveConfig(path string, cfg *Config) error {
    dir := filepath.Dir(path)
    if err := os.MkdirAll(dir, 0755); err != nil {
        return err
    }
    
    data, err := toml.Marshal(cfg)
    if err != nil {
        return err
    }
    
    return os.WriteFile(path, data, 0644)
}
