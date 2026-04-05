package main

import (
    "fmt"
    
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
        return
    }
    
    // Connect to database
    dbPath := store.DefaultDBPath()
    sqliteStore, err := store.NewSQLiteStore(dbPath)
    if err != nil {
        fmt.Printf("Error connecting to database: %v\n", err)
        return
    }
    defer sqliteStore.Close()
    
    // Initialize services
    sessionSvc := service.NewSessionService(sqliteStore, cfg)
    err = sessionSvc.LoadSessions()
    if err != nil {
        fmt.Printf("Error loading sessions: %v\n", err)
        return
    }
    
    // Start TUI
    model := tui.NewModel(sessionSvc)
    p := tea.NewProgram(model)
    
    if _, err := p.Run(); err != nil {
        fmt.Printf("Error running TUI: %v\n", err)
        return
    }
}
