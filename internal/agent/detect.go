package agent

import (
	"os"
	"path/filepath"
)

func DetectAgentDBPath(agentName string) string {
	home, _ := os.UserHomeDir()
	return detectDBPathInDir(agentName, home)
}

func detectDBPathInDir(agentName, homeDir string) string {
	shareDir := filepath.Join(homeDir, ".local", "share", agentName)

	opencodeDB := filepath.Join(shareDir, "opencode.db")
	if _, err := os.Stat(opencodeDB); err == nil {
		return opencodeDB
	}

	agentDB := filepath.Join(shareDir, agentName+".db")
	if _, err := os.Stat(agentDB); err == nil {
		return agentDB
	}

	return ""
}
