package agent

import (
	"fmt"
	"os"
	"os/exec"
)

func ValidateAgent(cfg *AgentConfig) error {
	if err := validateCommand(cfg.GetCommand()); err != nil {
		return err
	}

	if err := validateDatabase(cfg.GetDBPath()); err != nil {
		return err
	}

	return nil
}

func validateCommand(cmd string) error {
	_, err := exec.LookPath(cmd)
	if err != nil {
		return fmt.Errorf("命令 '%s' 不存在或不可执行", cmd)
	}
	return nil
}

func validateDatabase(dbPath string) error {
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		return fmt.Errorf("数据库 '%s' 不存在", dbPath)
	}
	return nil
}
