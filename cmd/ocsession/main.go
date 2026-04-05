package main

import (
	"fmt"
	"os"
	
	tea "github.com/charmbracelet/bubbletea"
	
	"github.com/opencode-session-manager/ocsession/internal/config"
	"github.com/opencode-session-manager/ocsession/internal/service"
	"github.com/opencode-session-manager/ocsession/internal/store"
	"github.com/opencode-session-manager/ocsession/internal/tui"
)

func main() {
	// Load config
	cfgPath := config.GetDefaultConfigPath()
	cfg, err := config.LoadOrCreateConfig(cfgPath)
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		os.Exit(1)
	}
	
	// Connect to database
	dbPath := store.DefaultDBPath()
	sqliteStore, err := store.NewSQLiteStore(dbPath)
	if err != nil {
		fmt.Printf("Error connecting to database: %v\n", err)
		os.Exit(1)
	}
	defer sqliteStore.Close()
	
	// Initialize services
	sessionSvc := service.NewSessionService(sqliteStore, cfg)
	err = sessionSvc.LoadSessions()
	if err != nil {
		fmt.Printf("Error loading sessions: %v\n", err)
		os.Exit(1)
	}
	
	// Start TUI
	model := tui.NewModel(sessionSvc)
	p := tea.NewProgram(model)
	
	finalModel, err := p.Run()
	if err != nil {
		fmt.Printf("Error running TUI: %v\n", err)
		os.Exit(1)
	}
	
	// 检查是否需要启动 OpenCode
	if m, ok := finalModel.(tui.Model); ok {
		if m.SessionToStart != nil {
			// 启动 OpenCode
			if err := tui.RunOpenCode(m.SessionToStart); err != nil {
				fmt.Fprintf(os.Stderr, "\n错误: 无法启动会话: %v\n", err)
				fmt.Fprintf(os.Stderr, "请确保 opencode 已正确安装\n")
				os.Exit(1)
			}
		}
	}
}