package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/zhaiyz/ocsession/internal/agent"
	"github.com/zhaiyz/ocsession/internal/config"
	"github.com/zhaiyz/ocsession/internal/service"
	"github.com/zhaiyz/ocsession/internal/store"
	"github.com/zhaiyz/ocsession/internal/tui"
)

func main() {
	agentCfgPath := getAgentConfigPath()

	agentCfg, err := loadOrDetectAgentConfig(agentCfgPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "错误: %v\n", err)
		os.Exit(1)
	}

	cfgPath := config.GetDefaultConfigPath()
	cfg, err := config.LoadOrCreateConfig(cfgPath)
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		os.Exit(1)
	}

	dbPath := agentCfg.GetDBPath()
	sqliteStore, err := store.NewSQLiteStore(dbPath)
	if err != nil {
		fmt.Printf("Error connecting to database: %v\n", err)
		os.Exit(1)
	}
	defer sqliteStore.Close()

	sessionSvc := service.NewSessionService(sqliteStore, cfg)
	err = sessionSvc.LoadSessions()
	if err != nil {
		fmt.Printf("Error loading sessions: %v\n", err)
		os.Exit(1)
	}

	model := tui.NewModel(sessionSvc, agentCfg)
	p := tea.NewProgram(model)

	finalModel, err := p.Run()
	if err != nil {
		fmt.Printf("Error running TUI: %v\n", err)
		os.Exit(1)
	}

	if m, ok := finalModel.(tui.Model); ok {
		if m.SessionToStart != nil {
			if err := tui.RunAgent(m.SessionToStart, agentCfg); err != nil {
				fmt.Fprintf(os.Stderr, "\n错误: 无法启动会话: %v\n", err)
				fmt.Fprintf(os.Stderr, "请确保 %s 已正确安装\n", agentCfg.GetCommand())
				os.Exit(1)
			}
		}
	}
}

func getAgentConfigPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "ocsession", "agent.toml")
}

func loadOrDetectAgentConfig(path string) (*agent.AgentConfig, error) {
	cfg, err := agent.LoadAgentConfig(path)
	if err == nil {
		if err := agent.ValidateAgent(cfg); err == nil {
			return cfg, nil
		}
		fmt.Printf("配置无效: %v\n", err)
	}

	return detectOrPromptAgentConfig(path)
}

func detectOrPromptAgentConfig(path string) (*agent.AgentConfig, error) {
	defaultCfg := &agent.AgentConfig{AgentName: "opencode"}

	dbPath := agent.DetectAgentDBPath("opencode")
	if dbPath != "" {
		defaultCfg.DBPath = dbPath
	}

	if err := agent.ValidateAgent(defaultCfg); err == nil {
		fmt.Printf("使用默认配置: opencode\n")
		if err := agent.SaveAgentConfig(path, defaultCfg); err != nil {
			return nil, fmt.Errorf("保存配置失败: %w", err)
		}
		return defaultCfg, nil
	}

	fmt.Println("未找到有效配置，请输入 AI Agent 名称（如 opencode、codewiz）:")
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("> ")
		input, err := reader.ReadString('\n')
		if err != nil {
			return nil, fmt.Errorf("读取输入失败: %w", err)
		}

		agentName := strings.TrimSpace(input)
		if agentName == "" {
			fmt.Println("Agent 名称不能为空，请重新输入")
			continue
		}

		cfg := &agent.AgentConfig{AgentName: agentName}

		dbPath := agent.DetectAgentDBPath(agentName)
		if dbPath == "" {
			fmt.Printf("未找到 %s 的数据库，请输入数据库完整路径:\n", agentName)
			fmt.Printf("> ")
			dbInput, err := reader.ReadString('\n')
			if err != nil {
				return nil, fmt.Errorf("读取输入失败: %w", err)
			}
			cfg.DBPath = strings.TrimSpace(dbInput)
		} else {
			cfg.DBPath = dbPath
		}

		if err := agent.ValidateAgent(cfg); err != nil {
			fmt.Printf("验证失败: %v\n请重新输入 Agent 名称:\n", err)
			continue
		}

		if err := agent.SaveAgentConfig(path, cfg); err != nil {
			return nil, fmt.Errorf("保存配置失败: %w", err)
		}

		fmt.Printf("配置已保存，使用 %s\n", agentName)
		return cfg, nil
	}
}
