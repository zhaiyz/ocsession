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
				
				// macOS: 在新Terminal窗口中启动OpenCode
				script := fmt.Sprintf(`tell application "Terminal" to do script "cd %s && opencode -s %s"`,
					selectedSession.Directory,
					selectedSession.ID)
				
				cmd := exec.Command("osascript", "-e", script)
				
				if err := cmd.Run(); err != nil {
					// 如果失败，尝试直接启动
					fmt.Printf("\n无法在新窗口启动，尝试直接启动...\n")
					directCmd := exec.Command("opencode", "-s", selectedSession.ID)
					directCmd.Stdout = os.Stdout
					directCmd.Stderr = os.Stderr
					directCmd.Stdin = os.Stdin
					
					if err := directCmd.Start(); err != nil {
						fmt.Printf("错误: 无法启动会话: %v\n", err)
						fmt.Println("请确保 opencode 已正确安装并在 PATH 中")
						time.Sleep(3 * time.Second)
						return m, nil
					}
				}
				
				// 成功启动，退出TUI
				m.quitting = true
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
		
		// 格式化时间（固定宽度）
		timeStr := formatTime(sess.Updated)
		
		// 截断标题到固定显示宽度
		title := truncate(sess.Title, 50)
		
		// 计算需要的填充空格（考虑中文字符宽度）
		titleWidth := runewidth.StringWidth(title)
		padding := 52 - titleWidth
		if padding < 0 {
			padding = 0
		}
		
		// 构建行（固定时间列位置）
		line := cursor + title + strings.Repeat(" ", padding) + timeStr
		
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
		Width(75).
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
	
	result += fmt.Sprintf("标题: %s\n", truncate(sess.Title, 40))
	result += fmt.Sprintf("ID: %s\n", truncate(sess.ID, 30))
	result += fmt.Sprintf("目录: %s\n", truncate(sess.Directory, 40))
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
		result += fmt.Sprintf("\n备注: %s\n", truncate(sess.Notes, 100))
	}
	
	return result
}

func formatTime(timestamp int64) string {
	if timestamp == 0 {
		return "未知时间"
	}
	// 数据库存储的是毫秒时间戳，转换为秒
	t := time.Unix(timestamp/1000, 0)
	now := time.Now()
	diff := now.Sub(t)
	
	if diff.Hours() < 24 {
		// 今天 - 显示时间
		return t.Format("今天 15:04")
	} else if diff.Hours() < 24*7 {
		// 本周 - 显示星期
		return t.Format("Mon 15:04")
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