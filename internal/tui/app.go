package tui

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-runewidth"

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
				
				// 直接启动 OpenCode
				cmd := exec.Command("opencode", "-s", selectedSession.ID)
				cmd.Dir = selectedSession.Directory
				
				// 设置标准输入输出
				cmd.Stdout = os.Stdout
				cmd.Stderr = os.Stderr
				cmd.Stdin = os.Stdin
				
				// 退出TUI，让OpenCode接管终端
				m.quitting = true
				return m, tea.Sequence(
					tea.Quit,
					func() tea.Msg {
						if err := cmd.Run(); err != nil {
							fmt.Fprintf(os.Stderr, "\n错误: 无法启动会话: %v\n", err)
						}
						return nil
					},
				)
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

	// 会话列表 - 固定宽度
	listLines := make([]string, 0, len(m.sessions))
	for i, sess := range m.sessions {
		cursor := "  "
		if i == m.selectedIndex {
			cursor = "→ "
		}
		
		timeStr := formatTime(sess.Updated)
		title := truncate(sess.Title, 48)
		
		// 计算填充
		titleWidth := runewidth.StringWidth(title)
		padding := 50 - titleWidth
		if padding < 1 {
			padding = 1
		}
		
		// 构建行
		line := cursor + title + strings.Repeat(" ", padding) + timeStr
		
		if i == m.selectedIndex {
			line = styles.SelectedItemStyle.Render(line)
		} else {
			line = styles.ListItemStyle.Render(line)
		}
		
		listLines = append(listLines, line)
	}
	
	list := strings.Join(listLines, "\n")

	// 预览面板
	preview := ""
	if len(m.sessions) > 0 && m.selectedIndex < len(m.sessions) {
		sess := m.sessions[m.selectedIndex]
		preview = renderPreview(sess)
	}

	// 固定布局 - 避免选中行影响
	leftPanel := lipgloss.NewStyle().
		Width(72).
		Height(18).
		Render(list)
	
	rightPanel := styles.PreviewStyle.
		Width(52).
		Height(18).
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
	
	result += fmt.Sprintf("标题: %s\n", truncate(sess.Title, 40))
	result += fmt.Sprintf("ID: %s\n", truncate(sess.ID, 30))
	result += fmt.Sprintf("目录: %s\n", truncate(sess.Directory, 40))
	result += fmt.Sprintf("更新: %s\n", formatTime(sess.Updated))
	result += fmt.Sprintf("创建: %s\n", formatTime(sess.Created))
	
	if len(sess.Tags) > 0 {
		tagsStr := strings.Join(sess.Tags, " ")
		result += fmt.Sprintf("\n标签: %s\n", tagsStr)
	}
	
	if sess.Alias != "" {
		result += fmt.Sprintf("\n别名: %s\n", sess.Alias)
	}
	
	if sess.Notes != "" {
		result += fmt.Sprintf("\n备注: %s\n", truncate(sess.Notes, 100))
	}
	
	return result
}

func formatTime(timestamp int64) string {
	if timestamp == 0 {
		return "未知时间"
	}
	
	// 毫秒时间戳转换为秒
	t := time.Unix(timestamp/1000, 0)
	now := time.Now()
	diff := now.Sub(t)
	
	// 中文星期映射
	weekdays := map[string]string{
		"Monday":    "周一",
		"Tuesday":   "周二",
		"Wednesday": "周三",
		"Thursday":  "周四",
		"Friday":    "周五",
		"Saturday":  "周六",
		"Sunday":    "周日",
	}
	
	if diff.Hours() < 24 {
		// 今天 - 显示"今天 + 时间"
		return "今天 " + t.Format("15:04")
	} else if diff.Hours() < 24*7 {
		// 本周 - 显示中文星期 + 时间
		weekday := t.Format("Monday")
		cnWeekday := weekdays[weekday]
		return cnWeekday + " " + t.Format("15:04")
	} else {
		// 更早 - 显示日期
		return t.Format("2006-01-02")
	}
}

func truncate(s string, maxLen int) string {
	width := runewidth.StringWidth(s)
	if width <= maxLen {
		return s
	}
	return runewidth.Truncate(s, maxLen-3, "...")
}