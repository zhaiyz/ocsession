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
    SuggestionExpireDays int  `toml:"suggestion_expire_days"`
}

// SessionTags represents tags and notes associated with a session.
type SessionTags struct {
    Tags  []string `toml:"tags"`
    Notes string   `toml:"notes"`
}

// Config represents the main configuration structure for the session manager.
// It contains general settings, command aliases, session tags, and rules.
type Config struct {
    General       GeneralConfig            `toml:"general"`
    Aliases       map[string]string        `toml:"aliases"`
    SessionTags   map[string]SessionTags   `toml:"session_tags"`
    Rules         RulesConfig              `toml:"rules"`
}

// RulesConfig defines rules for session management.
type RulesConfig struct {
    TagKeywords   []string `toml:"tag_keywords"`
    ActiveDays    int      `toml:"active_days"`
    InactiveDays  int      `toml:"inactive_days"`
}

// DefaultConfig returns a new Config instance with sensible default values.
func DefaultConfig() *Config {
    return &Config{
        General: GeneralConfig{
            DefaultSort:        "updated",
            PreviewLines:       10,
            MaxSessionsDisplay: 50,
            Theme:              "default",
            SuggestionExpireDays: 90,
        },
        Aliases:     make(map[string]string),
        SessionTags: make(map[string]SessionTags),
        Rules: RulesConfig{
            TagKeywords: []string{
                "开发: development",
                "实现: implementation",
                "查询: exploration",
                "测试: testing",
                "修复: bugfix",
            },
            ActiveDays:   7,
            InactiveDays: 30,
        },
    }
}