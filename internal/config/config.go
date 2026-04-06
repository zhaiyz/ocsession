// Package config provides configuration management for the session manager.
// It supports TOML-based configuration files with default values and
// automatic configuration file creation.
package config

// GeneralConfig holds general application settings.
type GeneralConfig struct {
	DefaultSort        string `toml:"default_sort"`
	PreviewLines       int    `toml:"preview_lines"`
	MaxSessionsDisplay int    `toml:"max_sessions_display"`
	Theme              string `toml:"theme"`
}

// Config represents the main configuration structure for the session manager.
type Config struct {
	General GeneralConfig `toml:"general"`
}

// DefaultConfig returns a new Config instance with sensible default values.
func DefaultConfig() *Config {
	return &Config{
		General: GeneralConfig{
			DefaultSort:        "updated",
			PreviewLines:       10,
			MaxSessionsDisplay: 50,
			Theme:              "default",
		},
	}
}
