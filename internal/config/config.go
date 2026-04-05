package config

type GeneralConfig struct {
    DefaultSort        string `toml:"default_sort"`
    PreviewLines       int    `toml:"preview_lines"`
    MaxSessionsDisplay int    `toml:"max_sessions_display"`
    Theme              string `toml:"theme"`
    SuggestionExpireDays int  `toml:"suggestion_expire_days"`
}

type SessionTags struct {
    Tags  []string `toml:"tags"`
    Notes string   `toml:"notes"`
}

type Config struct {
    General       GeneralConfig            `toml:"general"`
    Aliases       map[string]string        `toml:"aliases"`
    SessionTags   map[string]SessionTags   `toml:"session_tags"`
    Rules         RulesConfig              `toml:"rules"`
}

type RulesConfig struct {
    TagKeywords   []string `toml:"tag_keywords"`
    ActiveDays    int      `toml:"active_days"`
    InactiveDays  int      `toml:"inactive_days"`
}

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
