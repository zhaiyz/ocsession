package agent

import (
	"os"
	"path/filepath"
	"testing"
)

func TestAgentConfig_GetCommand(t *testing.T) {
	tests := []struct {
		name     string
		config   AgentConfig
		expected string
	}{
		{
			name: "command equals agent_name",
			config: AgentConfig{
				AgentName: "codewiz",
			},
			expected: "codewiz",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.config.GetCommand(); got != tt.expected {
				t.Errorf("GetCommand() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestAgentConfig_GetDBPath(t *testing.T) {
	homeDir, _ := os.UserHomeDir()

	tests := []struct {
		name     string
		config   AgentConfig
		expected string
	}{
		{
			name: "empty db_path uses default",
			config: AgentConfig{
				AgentName: "codewiz",
			},
			expected: filepath.Join(homeDir, ".local", "share", "codewiz", "codewiz.db"),
		},
		{
			name: "custom db_path with tilde expansion",
			config: AgentConfig{
				AgentName: "codewiz",
				DBPath:    "~/.data/codewiz/db.db",
			},
			expected: filepath.Join(homeDir, ".data", "codewiz", "db.db"),
		},
		{
			name: "custom db_path absolute path",
			config: AgentConfig{
				AgentName: "codewiz",
				DBPath:    "/custom/path/db.db",
			},
			expected: "/custom/path/db.db",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.config.GetDBPath(); got != tt.expected {
				t.Errorf("GetDBPath() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestAgentConfig_GetDisplayName(t *testing.T) {
	tests := []struct {
		name     string
		config   AgentConfig
		expected string
	}{
		{
			name: "lowercase agent_name gets capitalized",
			config: AgentConfig{
				AgentName: "codewiz",
			},
			expected: "Codewiz",
		},
		{
			name: "uppercase agent_name stays uppercase",
			config: AgentConfig{
				AgentName: "OPCODE",
			},
			expected: "Opcode",
		},
		{
			name: "mixed case agent_name gets normalized",
			config: AgentConfig{
				AgentName: "CodeWiz",
			},
			expected: "Codewiz",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.config.GetDisplayName(); got != tt.expected {
				t.Errorf("GetDisplayName() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestLoadAgentConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "agent.toml")

	t.Run("file not exists returns error", func(t *testing.T) {
		_, err := LoadAgentConfig(configPath)
		if err == nil {
			t.Error("expected error for non-existent file")
		}
	})

	t.Run("valid file loads successfully", func(t *testing.T) {
		content := `
agent_name = "codewiz"
db_path = "~/.local/share/codewiz/codewiz.db"
`
		if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}

		cfg, err := LoadAgentConfig(configPath)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if cfg.AgentName != "codewiz" {
			t.Errorf("AgentName = %v, want codewiz", cfg.AgentName)
		}
		if cfg.DBPath != "~/.local/share/codewiz/codewiz.db" {
			t.Errorf("DBPath = %v, want ~/.local/share/codewiz/codewiz.db", cfg.DBPath)
		}
	})

	t.Run("minimal config with only agent_name", func(t *testing.T) {
		minimalPath := filepath.Join(tmpDir, "minimal.toml")
		content := `agent_name = "opencode"`
		if err := os.WriteFile(minimalPath, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}

		cfg, err := LoadAgentConfig(minimalPath)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if cfg.AgentName != "opencode" {
			t.Errorf("AgentName = %v, want opencode", cfg.AgentName)
		}
		if cfg.DBPath != "" {
			t.Errorf("DBPath should be empty for default, got %v", cfg.DBPath)
		}
	})
}

func TestSaveAgentConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")

	cfg := &AgentConfig{
		AgentName: "codewiz",
		DBPath:    "~/.local/share/codewiz/codewiz.db",
	}

	if err := SaveAgentConfig(configPath, cfg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("config file was not created")
	}

	loaded, err := LoadAgentConfig(configPath)
	if err != nil {
		t.Fatalf("unexpected error loading saved config: %v", err)
	}

	if loaded.AgentName != cfg.AgentName {
		t.Errorf("loaded AgentName = %v, want %v", loaded.AgentName, cfg.AgentName)
	}
	if loaded.DBPath != cfg.DBPath {
		t.Errorf("loaded DBPath = %v, want %v", loaded.DBPath, cfg.DBPath)
	}
}
