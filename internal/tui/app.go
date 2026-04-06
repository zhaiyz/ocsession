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

	"github.com/zhaiyz/ocsession/internal/agent"
	"github.com/zhaiyz/ocsession/internal/service"
	"github.com/zhaiyz/ocsession/internal/store"
	"github.com/zhaiyz/ocsession/internal/tui/styles"
)

type Model struct {
	sessions       []store.Session
	allSessions    []store.Session // 保存所有会话用于搜索过滤
	selectedIndex  int
	searchQuery    string
	searchMode     bool
	sessionService *service.SessionService
	quitting       bool
	agentConfig    *agent.AgentConfig

	SessionToStart *store.Session

	currentDetail *store.SessionDetail
}

func NewModel(svc *service.SessionService, agentCfg *agent.AgentConfig) Model {
	sessions := svc.GetAllSessions()
	return Model{
		sessions:       sessions,
		allSessions:    sessions,
		sessionService: svc,
		selectedIndex:  0,
		searchMode:     false,
		agentConfig:    agentCfg,
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
				if len(m.searchQuery) > 0 {
					runes := []rune(m.searchQuery)
					m.searchQuery = string(runes[:len(runes)-1])
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

	header := styles.TitleStyle.Render(" "+m.agentConfig.GetDisplayName()+" Session Manager ") +
		styles.HelpStyle.Render("  [q:退出] [j/k:导航] [/:搜索] [r:刷新] [Enter:继续]")

	// 会话列表 - 固定宽度
	listLines := make([]string, 0, len(m.sessions))
	for i, sess := range m.sessions {
		cursor := "  "
		if i == m.selectedIndex {
			cursor = "→ "
		}

		timeStr := formatTime(sess.Updated)

		// 提取文件夹名
		folderName := extractFolderName(sess.Directory)
		folder := truncate(folderName, 15)

		// 标题截断到35字符
		title := truncate(sess.Title, 35)

		// 计算标题填充
		titleWidth := runewidth.StringWidth(title)
		titlePadding := 38 - titleWidth
		if titlePadding < 1 {
			titlePadding = 1
		}

		// 构建行：文件夹名(15) + 标题(35) + 时间
		line := cursor + folder + strings.Repeat(" ", 16-runewidth.StringWidth(folder)) + title + strings.Repeat(" ", titlePadding) + timeStr

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

		// 加载详细信息（包括消息）
		if m.currentDetail == nil || m.currentDetail.Session.ID != sess.ID {
			m.currentDetail, _ = m.sessionService.GetSessionDetail(sess.ID)
		}

		if m.currentDetail != nil {
			preview = renderPreview(*m.currentDetail)
		} else {
			preview = renderPreview(store.SessionDetail{Session: sess})
		}
	}

	// 固定布局 - 确保边框完整
	leftPanel := lipgloss.NewStyle().
		Width(85).
		Height(22).
		Render(list)

	rightPanel := styles.PreviewStyle.
		Width(58).
		Height(22).
		Render(preview)

	mainContent := lipgloss.JoinHorizontal(lipgloss.Top, leftPanel, rightPanel)

	if m.searchMode {
		searchBox := styles.SearchPromptStyle.Render("搜索: ") + m.searchQuery + "█"
		helpText := styles.HelpStyle.Render("  [Esc:取消] [Enter:确认] [Backspace:删除]")
		return lipgloss.JoinVertical(lipgloss.Left,
			header,
			searchBox+helpText,
			mainContent,
		)
	}

	return lipgloss.JoinVertical(lipgloss.Left,
		header,
		mainContent,
	)
}

func renderPreview(detail store.SessionDetail) string {
	var result strings.Builder

	sess := detail.Session

	// 标题
	result.WriteString(styles.TitleStyle.Render("会话详情") + "\n\n")

	// 基本信息 - 标题完整显示
	title := sess.Title
	if runewidth.StringWidth(title) > 50 {
		// 多行显示长标题
		lines := wrapText(title, 50)
		result.WriteString(fmt.Sprintf("标题: %s\n", lines[0]))
		for i := 1; i < len(lines); i++ {
			result.WriteString(fmt.Sprintf("      %s\n", lines[i]))
		}
	} else {
		result.WriteString(fmt.Sprintf("标题: %s\n", title))
	}

	result.WriteString(fmt.Sprintf("ID: %s\n", truncate(sess.ID, 25)))

	// 项目路径（完整路径）
	if sess.Directory != "" {
		result.WriteString(fmt.Sprintf("路径: %s\n", truncate(sess.Directory, 45)))
	}

	// 时间信息
	result.WriteString(fmt.Sprintf("更新: %s\n", formatTime(sess.Updated)))
	result.WriteString(fmt.Sprintf("创建: %s\n", formatTime(sess.Created)))

	// 会话时长
	if sess.Updated > 0 && sess.Created > 0 {
		duration := (sess.Updated - sess.Created) / 1000
		result.WriteString(fmt.Sprintf("时长: %s\n", formatDuration(duration)))
	}

	// 统计信息
	if detail.Stats.MessageCount > 0 {
		result.WriteString(fmt.Sprintf("消息: %d 条\n", detail.Stats.MessageCount))
	}

	// 标签
	if len(sess.Tags) > 0 {
		tagsStr := strings.Join(sess.Tags, " ")
		result.WriteString(fmt.Sprintf("\n标签: %s\n", truncate(tagsStr, 40)))
	}

	// 别名
	if sess.Alias != "" {
		result.WriteString(fmt.Sprintf("\n别名: %s\n", sess.Alias))
	}

	// 备注
	if sess.Notes != "" {
		result.WriteString(fmt.Sprintf("\n备注: %s\n", truncate(sess.Notes, 45)))
	}

	// 对话内容
	if len(detail.LastMessages) > 0 {
		result.WriteString("\n" + styles.HelpStyle.Render("─ 对话记录 ─") + "\n")

		msgs := detail.LastMessages
		totalCount := len(msgs)

		if totalCount <= 10 {
			// 少于等于10条，全部显示
			for i, msg := range msgs {
				cleanMsg := strings.TrimSpace(msg.Content)
				cleanMsg = strings.ReplaceAll(cleanMsg, "\n", " ")
				cleanMsg = strings.ReplaceAll(cleanMsg, "\r", " ")
				truncated := truncate(cleanMsg, 50)
				result.WriteString(fmt.Sprintf("%d. %s\n", i+1, truncated))
			}
		} else {
			// 超过10条，显示前5条和后5条
			// 前5条
			for i := 0; i < 5; i++ {
				cleanMsg := strings.TrimSpace(msgs[i].Content)
				cleanMsg = strings.ReplaceAll(cleanMsg, "\n", " ")
				cleanMsg = strings.ReplaceAll(cleanMsg, "\r", " ")
				truncated := truncate(cleanMsg, 50)
				result.WriteString(fmt.Sprintf("%d. %s\n", i+1, truncated))
			}

			// 省略提示
			skipped := totalCount - 10
			result.WriteString(fmt.Sprintf("... 省略 %d 条消息 ...\n", skipped))

			// 后5条
			for i := 0; i < 5; i++ {
				idx := totalCount - 5 + i
				cleanMsg := strings.TrimSpace(msgs[idx].Content)
				cleanMsg = strings.ReplaceAll(cleanMsg, "\n", " ")
				cleanMsg = strings.ReplaceAll(cleanMsg, "\r", " ")
				truncated := truncate(cleanMsg, 50)
				result.WriteString(fmt.Sprintf("%d. %s\n", idx+1, truncated))
			}
		}
	}

	return result.String()
}

// extractFolderName 从目录路径提取文件夹名
func extractFolderName(directory string) string {
	if directory == "" {
		return "-"
	}
	parts := strings.TrimRight(directory, "/")
	folders := strings.Split(parts, "/")
	if len(folders) > 0 {
		return folders[len(folders)-1]
	}
	return directory
}

// wrapText 文本换行
func wrapText(text string, maxLen int) []string {
	var lines []string
	words := strings.Fields(text)
	if len(words) == 0 {
		return []string{text}
	}

	currentLine := words[0]
	for _, word := range words[1:] {
		if runewidth.StringWidth(currentLine+" "+word) <= maxLen {
			currentLine += " " + word
		} else {
			lines = append(lines, currentLine)
			currentLine = word
		}
	}
	lines = append(lines, currentLine)
	return lines
}

// formatMessage 格式化单条消息
func formatMessage(content string, index int) string {
	// 简化消息内容
	content = strings.TrimSpace(content)
	if len(content) > 80 {
		content = content[:77] + "..."
	}
	// 移除换行符
	content = strings.ReplaceAll(content, "\n", " ")
	return styles.MessageStyle.Render(fmt.Sprintf("%d. %s\n", index, content))
}

// extractProjectName 从目录路径提取项目名
func extractProjectName(directory string) string {
	if directory == "" {
		return "-"
	}
	parts := strings.Split(strings.TrimRight(directory, "/"), "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return directory
}

// formatDuration 格式化持续时间
func formatDuration(seconds int64) string {
	if seconds < 60 {
		return fmt.Sprintf("%d秒", seconds)
	} else if seconds < 3600 {
		minutes := seconds / 60
		secs := seconds % 60
		if secs > 0 {
			return fmt.Sprintf("%d分%d秒", minutes, secs)
		}
		return fmt.Sprintf("%d分钟", minutes)
	} else {
		hours := seconds / 3600
		minutes := (seconds % 3600) / 60
		if minutes > 0 {
			return fmt.Sprintf("%d小时%d分", hours, minutes)
		}
		return fmt.Sprintf("%d小时", hours)
	}
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
func RunAgent(session *store.Session, agentCfg *agent.AgentConfig) error {
	if session == nil {
		return fmt.Errorf("session is nil")
	}

	cmd := exec.Command(agentCfg.GetCommand(), "-s", session.ID)
	cmd.Dir = session.Directory
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	return cmd.Run()
}
