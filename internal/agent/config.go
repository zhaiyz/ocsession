package agent

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/pelletier/go-toml/v2"
)

type AgentConfig struct {
	AgentName string `toml:"agent_name"`
	DBPath    string `toml:"db_path"`
}

func (c *AgentConfig) GetCommand() string {
	return c.AgentName
}

func (c *AgentConfig) GetDBPath() string {
	if c.DBPath != "" {
		return expandHome(c.DBPath)
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".local", "share", c.AgentName, c.AgentName+".db")
}

func (c *AgentConfig) GetDisplayName() string {
	if len(c.AgentName) == 0 {
		return ""
	}
	return strings.ToUpper(c.AgentName[:1]) + strings.ToLower(c.AgentName[1:])
}

func LoadAgentConfig(path string) (*AgentConfig, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg AgentConfig
	if err := toml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func SaveAgentConfig(path string, cfg *AgentConfig) error {
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

func expandHome(path string) string {
	if strings.HasPrefix(path, "~") {
		home, _ := os.UserHomeDir()
		return filepath.Join(home, path[1:])
	}
	return path
}
