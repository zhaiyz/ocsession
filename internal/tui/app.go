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
	allSessions    []store.Session  // 保存所有会话用于搜索过滤
	selectedIndex  int
	searchQuery    string
	searchMode     bool
	sessionService *service.SessionService
	quitting       bool
	
	// 用于启动 OpenCode 的会话信息
	SessionToStart *store.Session
}

func NewModel(svc *service.SessionService) Model {
	sessions := svc.GetAllSessions()
	return Model{
		sessions:       sessions,
		allSessions:    sessions,
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
		// 在搜索模式下，处理字符输入
		if m.searchMode {
			switch msg.Type {
			case tea.KeyEsc:
				// ESC 退出搜索模式
				m.searchMode = false
				m.searchQuery = ""
				m.sessions = m.allSessions
				m.selectedIndex = 0
				
			case tea.KeyEnter:
				// Enter 退出搜索模式，保持过滤结果
				m.searchMode = false
				
			case tea.KeyBackspace:
				// 退格删除字符
				if len(m.searchQuery) > 0 {
					m.searchQuery = m.searchQuery[:len(m.searchQuery)-1]
					m.filterSessions()
				}
				
			case tea.KeyRunes:
				// 输入字符
				m.searchQuery += string(msg.Runes)
				m.filterSessions()
				
			default:
				// 其他按键（如方向键）在搜索模式下也有效
				switch msg.String() {
				case "up", "k":
					if m.selectedIndex > 0 {
						m.selectedIndex--
					}
				case "down", "j":
					if m.selectedIndex < len(m.sessions)-1 {
						m.selectedIndex++
					}
				}
			}
		} else {
			// 非搜索模式的按键处理
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
			
			case "enter":
				if len(m.sessions) > 0 {
					// 保存要启动的会话
					session := m.sessions[m.selectedIndex]
					m.SessionToStart = &session
					m.quitting = true
					return m, tea.Quit
				}
			
			case "r":
				_ = m.sessionService.LoadSessions()
				m.allSessions = m.sessionService.GetAllSessions()
				m.sessions = m.allSessions
			}
		}
	}
	
	return m, nil
}

// filterSessions 根据搜索查询过滤会话
func (m *Model) filterSessions() {
	if m.searchQuery == "" {
		m.sessions = m.allSessions
		m.selectedIndex = 0
		return
	}
	
	query := strings.ToLower(m.searchQuery)
	var filtered []store.Session
	
	for _, sess := range m.allSessions {
		title := strings.ToLower(sess.Title)
		dir := strings.ToLower(sess.Directory)
		
		// 标题或目录包含搜索词
		if strings.Contains(title, query) || strings.Contains(dir, query) {
			filtered = append(filtered, sess)
		}
	}
	
	m.sessions = filtered
	if m.selectedIndex >= len(m.sessions) {
		m.selectedIndex = len(m.sessions) - 1
	}
	if m.selectedIndex < 0 {
		m.selectedIndex = 0
	}
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
		title := truncate(sess.Title, 60)  // 增加标题宽度到60
		
		// 计算填充
		titleWidth := runewidth.StringWidth(title)
		padding := 65 - titleWidth  // 增加基础padding到65
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
	
	// 如果没有搜索结果，显示提示
	if len(m.sessions) == 0 && m.searchMode {
		list = styles.HelpStyle.Render("  没有匹配的会话")
	}

	// 预览面板
	preview := ""
	if len(m.sessions) > 0 && m.selectedIndex < len(m.sessions) {
		sess := m.sessions[m.selectedIndex]
		preview = renderPreview(sess)
	}

	// 固定布局 - 避免选中行影响
	leftPanel := lipgloss.NewStyle().
		Width(85).  // 增加左侧面板宽度到85
		Height(18).
		Render(list)
	
	rightPanel := styles.PreviewStyle.
		Width(50).  // 右侧面板保持50
		Height(18).
		Render(preview)

	mainContent := lipgloss.JoinHorizontal(lipgloss.Top, leftPanel, rightPanel)

	if m.searchMode {
		searchBox := styles.SearchPromptStyle.Render("搜索: ") + m.searchQuery + "█"
		helpText := styles.HelpStyle.Render("  [Esc:取消] [Enter:确认] [Backspace:删除]")
		return lipgloss.JoinVertical(lipgloss.Left,
			header,
			searchBox + helpText,
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

// RunOpenCode 启动 OpenCode 会话
func RunOpenCode(session *store.Session) error {
	if session == nil {
		return fmt.Errorf("session is nil")
	}
	
	cmd := exec.Command("opencode", "-s", session.ID)
	cmd.Dir = session.Directory
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	
	return cmd.Run()
}