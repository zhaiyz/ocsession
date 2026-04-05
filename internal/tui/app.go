package tui

import (
    "github.com/charmbracelet/bubbletea"
    "github.com/charmbracelet/lipgloss"
    
    "github.com/opencode-session-manager/ocsession/internal/service"
    "github.com/opencode-session-manager/ocsession/internal/store"
    "github.com/opencode-session-manager/ocsession/internal/tui/styles"
)

// Model represents the TUI application state
type Model struct {
    sessions       []store.Session
    selectedIndex  int
    searchQuery    string
    searchMode     bool
    sessionService *service.SessionService
}

// NewModel creates a new TUI model
func NewModel(svc *service.SessionService) Model {
    return Model{
        sessions:       svc.GetAllSessions(),
        sessionService: svc,
        selectedIndex:  0,
        searchMode:     false,
    }
}

// Init initializes the TUI
func (m Model) Init() tea.Cmd {
    return nil
}

// Update handles TUI updates
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch msg.String() {
        case "q", "ctrl+c":
            return m, tea.Quit
        case "up", "k":
            if m.selectedIndex > 0 {
                m.selectedIndex--
            }
        case "down", "j":
            if m.selectedIndex < len(m.sessions)-1 {
                m.selectedIndex++
            }
        case "/":
            m.searchMode = true
            m.searchQuery = ""
        case "enter":
            if m.searchMode {
                m.searchMode = false
                // Apply search filter
            } else {
                // Continue session: opencode -s <session-id>
                return m, tea.Quit
            }
        case "esc":
            if m.searchMode {
                m.searchMode = false
                m.searchQuery = ""
            }
        }
    }
    
    return m, nil
}

// View renders the TUI
func (m Model) View() string {
    // Render header
    header := " OpenCode Session Manager [q:退出] "
    
    // Render session list
    list := ""
    for i, sess := range m.sessions {
        style := styles.ListItemStyle
        if i == m.selectedIndex {
            style = styles.SelectedItemStyle
        }
        list += style.Render(sess.Title) + "\n"
    }
    
    // Render preview
    preview := "Preview panel"
    
    // Layout
    leftPanel := lipgloss.NewStyle().Width(40).Render(list)
    rightPanel := styles.PreviewStyle.Width(60).Render(preview)
    
    mainContent := lipgloss.JoinHorizontal(lipgloss.Top, leftPanel, rightPanel)
    
    return lipgloss.JoinVertical(lipgloss.Left, 
        styles.TitleStyle.Render(header),
        mainContent,
        styles.HelpStyle.Render("[Enter:继续] [/:搜索]"),
    )
}
