package tui

import (
	"fmt"
	"os/exec"
	"time"

	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/opencode-session-manager/ocsession/internal/service"
	"github.com/opencode-session-manager/ocsession/internal/store"
	"github.com/opencode-session-manager/ocsession/internal/tui/styles"
)

type Model struct {
	sessions       []store.Session
	selectedIndex  int
	searchQuery    string
	searchMode     bool
	sessionService *service.SessionService
	quitting       bool
}

func NewModel(svc *service.SessionService) Model {
	return Model{
		sessions:       svc.GetAllSessions(),
		sessionService: svc,
		selectedIndex:  0,
		searchMode:     false,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			m.quitting = true
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
		
		case "esc":
			if m.searchMode {
				m.searchMode = false
				m.searchQuery = ""
			}
		
		case "enter":
			if m.searchMode {
				m.searchMode = false
			} else if len(m.sessions) > 0 {
				selectedSession := m.sessions[m.selectedIndex]
				cmd := exec.Command("opencode", "-s", selectedSession.ID)
				cmd.Start()
				return m, tea.Quit
			}
		
		case "r":
			_ = m.sessionService.LoadSessions()
			m.sessions = m.sessionService.GetAllSessions()
		}
	}
	
	return m, nil
}

func (m Model) View() string {
	if m.quitting {
		return ""
	}

	header := styles.TitleStyle.Render(" OpenCode Session Manager ") + 
		styles.HelpStyle.Render("  [q:退出] [j/k:导航] [/:搜索] [r:刷新] [Enter:继续]")

	list := ""
	for i, sess := range m.sessions {
		cursor := "  "
		if i == m.selectedIndex {
			cursor = "→ "
		}
		
		timeStr := formatTime(sess.Updated)
		
		line := fmt.Sprintf("%s %-40s %s", 
			cursor, 
			truncate(sess.Title, 40), 
			styles.HelpStyle.Render(timeStr),
		)
		
		if i == m.selectedIndex {
			line = styles.SelectedItemStyle.Render(line)
		} else {
			line = styles.ListItemStyle.Render(line)
		}
		
		list += line + "\n"
	}

	preview := ""
	if len(m.sessions) > 0 && m.selectedIndex < len(m.sessions) {
		sess := m.sessions[m.selectedIndex]
		preview = renderPreview(sess)
	}

	leftPanel := lipgloss.NewStyle().
		Width(70).
		Height(20).
		Render(list)
	
	rightPanel := styles.PreviewStyle.
		Width(50).
		Height(20).
		Render(preview)

	mainContent := lipgloss.JoinHorizontal(lipgloss.Top, leftPanel, rightPanel)

	if m.searchMode {
		searchBox := styles.SearchPromptStyle.Render("搜索: ") + m.searchQuery + "█"
		return lipgloss.JoinVertical(lipgloss.Left,
			header,
			searchBox,
			mainContent,
		)
	}

	return lipgloss.JoinVertical(lipgloss.Left,
		header,
		mainContent,
	)
}

func renderPreview(sess store.Session) string {
	result := styles.TitleStyle.Render("会话详情") + "\n\n"
	
	result += fmt.Sprintf("标题: %s\n", sess.Title)
	result += fmt.Sprintf("ID: %s\n", truncate(sess.ID, 30))
	result += fmt.Sprintf("目录: %s\n", sess.Directory)
	result += fmt.Sprintf("更新: %s\n", formatTime(sess.Updated))
	result += fmt.Sprintf("创建: %s\n", formatTime(sess.Created))
	
	if len(sess.Tags) > 0 {
		tagsStr := ""
		for _, tag := range sess.Tags {
			tagsStr += tag + " "
		}
		result += fmt.Sprintf("\n标签: %s\n", tagsStr)
	}
	
	if sess.Alias != "" {
		result += fmt.Sprintf("\n别名: %s\n", sess.Alias)
	}
	
	if sess.Notes != "" {
		result += fmt.Sprintf("\n备注: %s\n", sess.Notes)
	}
	
	return result
}

func formatTime(timestamp int64) string {
	if timestamp == 0 {
		return "-"
	}
	t := time.Unix(timestamp, 0)
	now := time.Now()
	diff := now.Sub(t)
	
	if diff.Hours() < 24 {
		return t.Format("15:04")
	} else if diff.Hours() < 24*7 {
		return t.Format("周一 15:04")
	} else {
		return t.Format("2006-01-02")
	}
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}