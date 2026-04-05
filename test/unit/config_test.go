package config

import (
	"testing"

	"github.com/opencode-session-manager/ocsession/internal/config"
)

func TestDefaultConfig(t *testing.T) {
	cfg := config.DefaultConfig()

	if cfg.General.DefaultSort != "updated" {
		t.Errorf("Expected DefaultSort 'updated', got '%s'", cfg.General.DefaultSort)
	}

	if cfg.General.PreviewLines != 10 {
		t.Errorf("Expected PreviewLines 10, got %d", cfg.General.PreviewLines)
	}

	if len(cfg.Rules.TagKeywords) == 0 {
		t.Error("Expected non-empty TagKeywords")
	}
}