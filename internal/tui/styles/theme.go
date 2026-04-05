package styles

import "github.com/charmbracelet/lipgloss"

var (
	TitleStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("15")).
		Background(lipgloss.Color("62")).
		Padding(0, 1)
	
	ListItemStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("252"))
	
	SelectedItemStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("15")).
		Background(lipgloss.Color("62"))
	
	PreviewStyle = lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("63")).
		Padding(0, 1).  // 修改为 0 padding，避免影响边框
		MarginLeft(1)   // 添加左边距，避免与左侧面板贴在一起
	
	SearchPromptStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("99"))
	
	HelpStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("241"))
	
	// 消息样式
	MessageStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("246"))
)
