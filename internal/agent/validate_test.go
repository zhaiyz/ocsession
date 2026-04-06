package agent

import (
	"os"
	"path/filepath"
	"testing"
)

func TestValidateAgent(t *testing.T) {
	t.Run("invalid command returns error", func(t *testing.T) {
		cfg := &AgentConfig{
			AgentName: "nonexistent-command-xyz",
		}

		err := ValidateAgent(cfg)
		if err == nil {
			t.Error("expected error for nonexistent command")
		}
		if !containsString(err.Error(), "命令") {
			t.Errorf("error should mention command, got: %v", err)
		}
	})

	t.Run("missing database returns error", func(t *testing.T) {
		cfg := &AgentConfig{
			AgentName: "go",
			DBPath:    "/nonexistent/path/db.db",
		}

		err := ValidateAgent(cfg)
		if err == nil {
			t.Error("expected error for missing database")
		}
		if !containsString(err.Error(), "数据库") {
			t.Errorf("error should mention database, got: %v", err)
		}
	})

	t.Run("valid config with existing command and database", func(t *testing.T) {
		tmpDir := t.TempDir()
		dbPath := filepath.Join(tmpDir, "test.db")

		if err := os.WriteFile(dbPath, []byte(""), 0644); err != nil {
			t.Fatal(err)
		}

		cfg := &AgentConfig{
			AgentName: "go",
			DBPath:    dbPath,
		}

		err := ValidateAgent(cfg)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
}

func TestValidateCommand(t *testing.T) {
	t.Run("existing command succeeds", func(t *testing.T) {
		err := validateCommand("go")
		if err != nil {
			t.Errorf("unexpected error for 'go' command: %v", err)
		}
	})

	t.Run("nonexistent command fails", func(t *testing.T) {
		err := validateCommand("nonexistent-xyz")
		if err == nil {
			t.Error("expected error for nonexistent command")
		}
	})
}

func TestValidateDatabase(t *testing.T) {
	t.Run("existing database succeeds", func(t *testing.T) {
		tmpDir := t.TempDir()
		dbPath := filepath.Join(tmpDir, "test.db")

		if err := os.WriteFile(dbPath, []byte(""), 0644); err != nil {
			t.Fatal(err)
		}

		err := validateDatabase(dbPath)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("nonexistent database fails", func(t *testing.T) {
		err := validateDatabase("/nonexistent/path/db.db")
		if err == nil {
			t.Error("expected error for nonexistent database")
		}
	})
}

func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsSubstring(s, substr))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
