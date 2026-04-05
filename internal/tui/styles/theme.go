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
		Padding(1, 2)
    
    SearchPromptStyle = lipgloss.NewStyle().
        Foreground(lipgloss.Color("99"))
    
    HelpStyle = lipgloss.NewStyle().
        Foreground(lipgloss.Color("241"))
)
