package tui

import (
	"testing"

	"github.com/charmbracelet/bubbletea"
)

func TestSearchQueryDeleteChinese(t *testing.T) {
	m := Model{
		searchMode:  true,
		searchQuery: "测试文本",
	}

	m.searchMode = true
	m.searchQuery = "测试文本"

	backspaceMsg := tea.KeyMsg{Type: tea.KeyBackspace}
	updatedModel, _ := m.Update(backspaceMsg)
	m = updatedModel.(Model)

	expected := "测试文"
	if m.searchQuery != expected {
		t.Errorf("After deleting one Chinese character, expected '%s', got '%s'", expected, m.searchQuery)
	}

	if len([]rune(m.searchQuery)) != len([]rune(expected)) {
		t.Errorf("Rune length mismatch: expected %d runes, got %d runes", len([]rune(expected)), len([]rune(m.searchQuery)))
	}
}

func TestSearchQueryDeleteMixedCharacters(t *testing.T) {
	m := Model{
		searchMode:  true,
		searchQuery: "test测试123",
	}

	for i := 0; i < 3; i++ {
		backspaceMsg := tea.KeyMsg{Type: tea.KeyBackspace}
		updatedModel, _ := m.Update(backspaceMsg)
		m = updatedModel.(Model)
	}

	expected := "test测试"
	if m.searchQuery != expected {
		t.Errorf("After deleting 3 characters, expected '%s', got '%s'", expected, m.searchQuery)
	}

	for i := 0; i < 2; i++ {
		backspaceMsg := tea.KeyMsg{Type: tea.KeyBackspace}
		updatedModel, _ := m.Update(backspaceMsg)
		m = updatedModel.(Model)
	}

	expected2 := "test"
	if m.searchQuery != expected2 {
		t.Errorf("After deleting 2 Chinese characters, expected '%s', got '%s'", expected2, m.searchQuery)
	}
}

func TestSearchQueryDeleteEmpty(t *testing.T) {
	m := Model{
		searchMode:  true,
		searchQuery: "",
	}

	backspaceMsg := tea.KeyMsg{Type: tea.KeyBackspace}
	updatedModel, _ := m.Update(backspaceMsg)
	m = updatedModel.(Model)

	if m.searchQuery != "" {
		t.Errorf("Deleting from empty query should keep it empty, got '%s'", m.searchQuery)
	}
}

func TestSearchQueryDeleteSingleChinese(t *testing.T) {
	m := Model{
		searchMode:  true,
		searchQuery: "中",
	}

	backspaceMsg := tea.KeyMsg{Type: tea.KeyBackspace}
	updatedModel, _ := m.Update(backspaceMsg)
	m = updatedModel.(Model)

	if m.searchQuery != "" {
		t.Errorf("Deleting single Chinese character should result in empty string, got '%s'", m.searchQuery)
	}
}

func TestSearchQueryDeleteEmoji(t *testing.T) {
	m := Model{
		searchMode:  true,
		searchQuery: "测试😀🎉",
	}

	backspaceMsg := tea.KeyMsg{Type: tea.KeyBackspace}
	updatedModel, _ := m.Update(backspaceMsg)
	m = updatedModel.(Model)

	expected := "测试😀"
	if m.searchQuery != expected {
		t.Errorf("After deleting emoji, expected '%s', got '%s'", expected, m.searchQuery)
	}
}
