package agent

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDetectDBPath(t *testing.T) {
	t.Run("returns opencode.db if exists", func(t *testing.T) {
		tmpDir := t.TempDir()
		shareDir := filepath.Join(tmpDir, ".local", "share", "testagent")
		if err := os.MkdirAll(shareDir, 0755); err != nil {
			t.Fatal(err)
		}

		opencodeDB := filepath.Join(shareDir, "opencode.db")
		if err := os.WriteFile(opencodeDB, []byte(""), 0644); err != nil {
			t.Fatal(err)
		}

		path := detectDBPathInDir("testagent", tmpDir)
		expected := opencodeDB
		if path != expected {
			t.Errorf("detectDBPath() = %v, want %v", path, expected)
		}
	})

	t.Run("returns agent_name.db if opencode.db not exists", func(t *testing.T) {
		tmpDir := t.TempDir()
		shareDir := filepath.Join(tmpDir, ".local", "share", "testagent")
		if err := os.MkdirAll(shareDir, 0755); err != nil {
			t.Fatal(err)
		}

		agentDB := filepath.Join(shareDir, "testagent.db")
		if err := os.WriteFile(agentDB, []byte(""), 0644); err != nil {
			t.Fatal(err)
		}

		path := detectDBPathInDir("testagent", tmpDir)
		expected := agentDB
		if path != expected {
			t.Errorf("detectDBPath() = %v, want %v", path, expected)
		}
	})

	t.Run("returns empty string if both not exist", func(t *testing.T) {
		tmpDir := t.TempDir()

		path := detectDBPathInDir("nonexistent", tmpDir)
		if path != "" {
			t.Errorf("detectDBPath() = %v, want empty string", path)
		}
	})

	t.Run("prefers opencode.db over agent_name.db", func(t *testing.T) {
		tmpDir := t.TempDir()
		shareDir := filepath.Join(tmpDir, ".local", "share", "testagent")
		if err := os.MkdirAll(shareDir, 0755); err != nil {
			t.Fatal(err)
		}

		opencodeDB := filepath.Join(shareDir, "opencode.db")
		agentDB := filepath.Join(shareDir, "testagent.db")

		if err := os.WriteFile(opencodeDB, []byte("opencode"), 0644); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(agentDB, []byte("agent"), 0644); err != nil {
			t.Fatal(err)
		}

		path := detectDBPathInDir("testagent", tmpDir)
		if path != opencodeDB {
			t.Errorf("detectDBPath() = %v, want %v (opencode.db)", path, opencodeDB)
		}
	})
}

func TestDetectAgentDBPath(t *testing.T) {
	t.Run("detects in real home directory", func(t *testing.T) {
		path := DetectAgentDBPath("opencode")
		expectedOpencode := filepath.Join(homeDir, ".local", "share", "opencode", "opencode.db")
		expectedAgent := filepath.Join(homeDir, ".local", "share", "opencode", "opencode.db")

		if _, err := os.Stat(expectedOpencode); err == nil {
			if path != expectedOpencode {
				t.Errorf("path = %v, want %v", path, expectedOpencode)
			}
		} else if _, err := os.Stat(expectedAgent); err == nil {
			if path != expectedAgent {
				t.Errorf("path = %v, want %v", path, expectedAgent)
			}
		} else {
			if path != "" {
				t.Errorf("path = %v, want empty string when no db exists", path)
			}
		}
	})
}

var homeDir string

func init() {
	homeDir, _ = os.UserHomeDir()
}
