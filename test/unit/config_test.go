package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/zhaiyz/ocsession/internal/config"
)

func TestDefaultConfig(t *testing.T) {
	cfg := config.DefaultConfig()

	if cfg.General.DefaultSort != "updated" {
		t.Errorf("Expected DefaultSort 'updated', got '%s'", cfg.General.DefaultSort)
	}

	if cfg.General.PreviewLines != 10 {
		t.Errorf("Expected PreviewLines 10, got %d", cfg.General.PreviewLines)
	}

}

func TestSaveConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")

	cfg := config.DefaultConfig()
	cfg.General.Theme = "custom"

	err := config.SaveConfig(configPath, cfg)
	if err != nil {
		t.Fatalf("SaveConfig failed: %v", err)
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("Config file was not created")
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("Failed to read saved config: %v", err)
	}

	if len(data) == 0 {
		t.Error("Config file is empty")
	}
}

func TestLoadConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")

	cfg := config.DefaultConfig()
	cfg.General.Theme = "dark"

	if err := config.SaveConfig(configPath, cfg); err != nil {
		t.Fatalf("Failed to save test config: %v", err)
	}

	loaded, err := config.LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	if loaded.General.Theme != "dark" {
		t.Errorf("Expected Theme 'dark', got '%s'", loaded.General.Theme)
	}
}

func TestLoadOrCreateConfig_CreateNew(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "subdir", "config.toml")

	cfg, err := config.LoadOrCreateConfig(configPath)
	if err != nil {
		t.Fatalf("LoadOrCreateConfig failed: %v", err)
	}

	if cfg == nil {
		t.Fatal("Expected non-nil config")
	}

	if cfg.General.DefaultSort != "updated" {
		t.Errorf("Expected default DefaultSort 'updated', got '%s'", cfg.General.DefaultSort)
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("Config file was not created by LoadOrCreateConfig")
	}
}
