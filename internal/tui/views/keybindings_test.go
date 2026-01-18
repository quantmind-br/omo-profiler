package views

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// keyMsg creates a tea.KeyMsg for regular character keys (letters, numbers, symbols)
// Example: keyMsg("a"), keyMsg("1"), keyMsg(" ")
func keyMsg(k string) tea.KeyMsg {
	if len(k) == 0 {
		return tea.KeyMsg{}
	}
	return tea.KeyMsg{
		Type:  tea.KeyRunes,
		Runes: []rune(k),
	}
}

// keyMsgSpecial creates a tea.KeyMsg for special keys (arrows, enter, etc.)
// Example: keyMsgSpecial(tea.KeyRight), keyMsgSpecial(tea.KeyEnter)
func keyMsgSpecial(k tea.KeyType) tea.KeyMsg {
	return tea.KeyMsg{
		Type: k,
	}
}

// TestKeybindingsTestSetup validates that the test infrastructure compiles and runs
func TestKeybindingsTestSetup(t *testing.T) {
	// Verify keyMsg helper works
	msg := keyMsg("a")
	if msg.Type != tea.KeyRunes {
		t.Errorf("expected KeyRunes type, got %v", msg.Type)
	}
	if len(msg.Runes) != 1 || msg.Runes[0] != 'a' {
		t.Errorf("expected rune 'a', got %v", msg.Runes)
	}

	// Verify keyMsgSpecial helper works
	specialMsg := keyMsgSpecial(tea.KeyRight)
	if specialMsg.Type != tea.KeyRight {
		t.Errorf("expected KeyRight type, got %v", specialMsg.Type)
	}

	// Test infrastructure is ready
	t.Log("Test infrastructure setup complete")
}
