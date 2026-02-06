package views

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/diogenes/omo-profiler/internal/diff"
)

func TestNewDiff(t *testing.T) {
	d := NewDiff()

	if d.focused != focusLeft {
		t.Errorf("expected focused to be focusLeft, got %v", d.focused)
	}

	if d.ready {
		t.Error("expected ready to be false initially")
	}

	if d.selectingLeft {
		t.Error("expected selectingLeft to be false initially")
	}

	if d.selectingRight {
		t.Error("expected selectingRight to be false initially")
	}
}

func TestDiffInit(t *testing.T) {
	d := NewDiff()
	cmd := d.Init()

	if cmd == nil {
		t.Fatal("expected non-nil command from Init")
	}

	msg := cmd()
	if _, ok := msg.(diffProfilesLoadedMsg); !ok {
		t.Errorf("expected diffProfilesLoadedMsg, got %T", msg)
	}
}

func TestDiffLoadProfiles(t *testing.T) {
	d := NewDiff()

	msg := d.loadProfiles()
	if _, ok := msg.(diffProfilesLoadedMsg); !ok {
		t.Errorf("expected diffProfilesLoadedMsg, got %T", msg)
	}
}

func TestDiffComputeDiff(t *testing.T) {
	d := NewDiff()

	// No profiles selected
	msg := d.computeDiff()
	result, ok := msg.(diffComputedMsg)
	if !ok {
		t.Errorf("expected diffComputedMsg, got %T", msg)
	}

	if result.err != nil {
		t.Errorf("expected no error when no profiles selected, got %v", result.err)
	}

	if result.result != nil {
		t.Error("expected nil result when no profiles selected")
	}
}

func TestDiffSetSize(t *testing.T) {
	d := NewDiff()

	d.SetSize(100, 50)

	if d.width != 100 {
		t.Errorf("expected width 100, got %d", d.width)
	}

	if d.height != 50 {
		t.Errorf("expected height 50, got %d", d.height)
	}

	if !d.ready {
		t.Error("expected ready to be true after SetSize")
	}

	// Call SetSize again to test the ready=true path
	d.SetSize(80, 40)

	if d.width != 80 {
		t.Errorf("expected width 80, got %d", d.width)
	}

	if d.height != 40 {
		t.Errorf("expected height 40, got %d", d.height)
	}
}

func TestDiffSetSizeZeroDimensions(t *testing.T) {
	d := NewDiff()

	// Set size with zero dimensions should not crash
	d.SetSize(0, 0)

	// Viewports are not initialized with zero dimensions
	// but ready becomes true after SetSize is called
	_ = d.ready
}

func TestDiffUpdateWindowSizeMsg(t *testing.T) {
	d := NewDiff()

	msg := tea.WindowSizeMsg{Width: 80, Height: 24}
	updated, cmd := d.Update(msg)

	if cmd != nil {
		t.Error("expected nil command for WindowSizeMsg")
	}

	if updated.width != 80 {
		t.Errorf("expected width 80, got %d", updated.width)
	}

	if updated.height != 24 {
		t.Errorf("expected height 24, got %d", updated.height)
	}
}

func TestDiffUpdateDiffProfilesLoadedMsg(t *testing.T) {
	d := NewDiff()

	// With 2 or more profiles
	msg := diffProfilesLoadedMsg{
		profiles: []string{"profile1", "profile2", "profile3"},
		err:      nil,
	}

	updated, cmd := d.Update(msg)

	if updated.profiles == nil {
		t.Error("expected profiles to be set")
	}

	if len(updated.profiles) != 3 {
		t.Errorf("expected 3 profiles, got %d", len(updated.profiles))
	}

	if updated.leftProfile != "profile1" {
		t.Errorf("expected left profile 'profile1', got %q", updated.leftProfile)
	}

	if updated.rightProfile != "profile2" {
		t.Errorf("expected right profile 'profile2', got %q", updated.rightProfile)
	}

	if updated.leftIdx != 0 {
		t.Errorf("expected left index 0, got %d", updated.leftIdx)
	}

	if updated.rightIdx != 1 {
		t.Errorf("expected right index 1, got %d", updated.rightIdx)
	}

	// Should return a command to compute diff
	if cmd == nil {
		t.Error("expected non-nil command to compute diff")
	}
}

func TestDiffUpdateDiffProfilesLoadedMsgLessThanTwo(t *testing.T) {
	d := NewDiff()

	// With only 1 profile
	msg := diffProfilesLoadedMsg{
		profiles: []string{"profile1"},
		err:      nil,
	}

	updated, cmd := d.Update(msg)

	if cmd != nil {
		t.Error("expected nil command when less than 2 profiles")
	}

	if updated.leftProfile != "" {
		t.Errorf("expected left profile to be empty, got %q", updated.leftProfile)
	}

	if updated.rightProfile != "" {
		t.Errorf("expected right profile to be empty, got %q", updated.rightProfile)
	}
}

func TestDiffUpdateDiffProfilesLoadedMsgWithError(t *testing.T) {
	d := NewDiff()

	msg := diffProfilesLoadedMsg{
		profiles: nil,
		err:      &testError{},
	}

	updated, cmd := d.Update(msg)

	if cmd != nil {
		t.Error("expected nil command when there's an error")
	}

	if updated.err == nil {
		t.Error("expected error to be set")
	}
}

func TestDiffUpdateDiffComputedMsg(t *testing.T) {
	d := NewDiff()
	d.SetSize(80, 24)

	result := &diff.DiffResult{
		Left:  []diff.DiffLine{{Type: diff.DiffEqual, Text: "line1"}},
		Right: []diff.DiffLine{{Type: diff.DiffEqual, Text: "line1"}},
	}

	msg := diffComputedMsg{
		result: result,
		err:    nil,
	}

	updated, cmd := d.Update(msg)

	if cmd != nil {
		t.Error("expected nil command for diffComputedMsg")
	}

	if updated.diffResult != result {
		t.Error("expected diffResult to be set")
	}
}

func TestDiffHandleNavigationKeysUp(t *testing.T) {
	d := NewDiff()
	d.SetSize(80, 24)

	msg := tea.KeyMsg{Type: tea.KeyUp}
	updated, _ := d.Update(msg)

	// Just verify no panic - scroll behavior is viewport-dependent
	_ = updated.leftViewport.YOffset
	_ = updated.rightViewport.YOffset
}

func TestDiffHandleNavigationKeysDown(t *testing.T) {
	d := NewDiff()
	d.SetSize(80, 24)

	msg := tea.KeyMsg{Type: tea.KeyDown}
	updated, _ := d.Update(msg)

	// Just verify no panic - scroll behavior is viewport-dependent
	_ = updated.leftViewport.YOffset
}

func TestDiffHandleNavigationKeysTab(t *testing.T) {
	d := NewDiff()

	msg := tea.KeyMsg{Type: tea.KeyTab}
	updated, _ := d.Update(msg)

	// After tab, focus should switch to right
	if updated.focused == focusLeft {
		t.Error("expected focused to switch from focusLeft")
	}
}

func TestDiffHandleNavigationKeysEnter(t *testing.T) {
	d := NewDiff()
	d.profiles = []string{"profile1", "profile2"}
	d.leftProfile = "profile1"
	d.rightProfile = "profile2"

	msg := tea.KeyMsg{Type: tea.KeyEnter}
	updated, _ := d.Update(msg)

	if !updated.selectingLeft {
		t.Error("expected selectingLeft to be true after Enter on left focus")
	}

	if updated.selectingRight {
		t.Error("expected selectingRight to remain false")
	}
}

func TestDiffHandleNavigationKeysEnterRightFocus(t *testing.T) {
	d := NewDiff()
	d.focused = focusRight
	d.profiles = []string{"profile1", "profile2"}
	d.leftProfile = "profile1"
	d.rightProfile = "profile2"

	msg := tea.KeyMsg{Type: tea.KeyEnter}
	updated, _ := d.Update(msg)

	if !updated.selectingRight {
		t.Error("expected selectingRight to be true after Enter on right focus")
	}

	if updated.selectingLeft {
		t.Error("expected selectingLeft to remain false")
	}
}

func TestDiffHandleSelectionKeysUp(t *testing.T) {
	d := NewDiff()
	d.profiles = []string{"profile1", "profile2", "profile3"}
	d.selectingLeft = true
	d.leftIdx = 2

	msg := tea.KeyMsg{Type: tea.KeyUp}
	updated, _ := d.Update(msg)

	// Should have decremented the index
	if updated.leftIdx >= 2 {
		t.Errorf("expected leftIdx to be less than 2, got %d", updated.leftIdx)
	}

	// selectingLeft is preserved during selection
	if !updated.selectingLeft {
		t.Error("expected selectingLeft to remain true during selection")
	}
}

func TestDiffHandleSelectionKeysDown(t *testing.T) {
	d := NewDiff()
	d.profiles = []string{"profile1", "profile2", "profile3"}
	d.selectingLeft = true
	d.leftIdx = 0

	msg := tea.KeyMsg{Type: tea.KeyDown}
	updated, _ := d.Update(msg)

	if updated.leftIdx != 1 {
		t.Errorf("expected leftIdx to be 1, got %d", updated.leftIdx)
	}
}

func TestDiffHandleSelectionKeysEnter(t *testing.T) {
	d := NewDiff()
	d.profiles = []string{"profile1", "profile2", "profile3"}
	d.selectingLeft = true
	d.leftIdx = 1

	msg := tea.KeyMsg{Type: tea.KeyEnter}
	updated, cmd := d.Update(msg)

	if updated.selectingLeft {
		t.Error("expected selectingLeft to be false after Enter")
	}

	if updated.leftProfile != "profile2" {
		t.Errorf("expected left profile 'profile2', got %q", updated.leftProfile)
	}

	// Should return a command to compute diff
	if cmd == nil {
		t.Error("expected non-nil command to compute diff")
	}
}

func TestDiffHandleSelectionKeysEsc(t *testing.T) {
	d := NewDiff()
	d.selectingLeft = true

	msg := tea.KeyMsg{Type: tea.KeyEsc}
	updated, cmd := d.Update(msg)

	if cmd != nil {
		t.Error("expected nil command for Esc key")
	}

	if updated.selectingLeft {
		t.Error("expected selectingLeft to be false after Esc")
	}

	if updated.selectingRight {
		t.Error("expected selectingRight to be false")
	}
}

func TestDiffScrollBoth(t *testing.T) {
	d := NewDiff()
	d.SetSize(80, 24)

	// Test that scrollBoth doesn't panic when ready
	d.scrollBoth(5)
	d.scrollBoth(-5)

	// Just verify no panic and state is valid
	if !d.ready {
		t.Error("expected ready to be true after SetSize")
	}
}

func TestDiffScrollBothClampsToZero(t *testing.T) {
	d := NewDiff()
	d.initViewports()

	// Try to scroll up from 0
	d.scrollBoth(-1)

	if d.leftViewport.YOffset != 0 {
		t.Errorf("expected YOffset to remain 0, got %d", d.leftViewport.YOffset)
	}
}

func TestDiffScrollBothWhenNotReady(t *testing.T) {
	d := NewDiff()
	// Don't call initViewports, so ready is false

	// Should not panic
	d.scrollBoth(5)

	// Viewport should not be initialized
	if d.ready {
		t.Error("expected ready to be false")
	}
}

func TestDiffBorderColor(t *testing.T) {
	d := NewDiff()
	d.focused = focusLeft

	color := d.borderColor(focusLeft)
	if color != diffPurple {
		t.Errorf("expected diffPurple for focused pane, got %v", color)
	}

	color = d.borderColor(focusRight)
	if color != diffGray {
		t.Errorf("expected diffGray for unfocused pane, got %v", color)
	}

	d.focused = focusRight
	color = d.borderColor(focusRight)
	if color != diffPurple {
		t.Errorf("expected diffPurple for focused pane, got %v", color)
	}

	color = d.borderColor(focusLeft)
	if color != diffGray {
		t.Errorf("expected diffGray for unfocused pane, got %v", color)
	}
}

func TestDiffShouldReturn(t *testing.T) {
	d := NewDiff()

	if d.ShouldReturn() {
		t.Error("expected ShouldReturn to be false")
	}
}

func TestDiffViewWithError(t *testing.T) {
	d := NewDiff()
	d.err = &testError{}

	view := d.View()

	if view == "" {
		t.Error("expected non-empty view")
	}

	if !contains(view, "Error:") {
		t.Error("expected 'Error:' in view")
	}
}

func TestDiffViewNoProfiles(t *testing.T) {
	d := NewDiff()
	d.profiles = []string{}

	view := d.View()

	if view == "" {
		t.Error("expected non-empty view")
	}

	if !contains(view, "No profiles available") {
		t.Error("expected 'No profiles available' in view")
	}
}

func TestDiffViewLessThanTwoProfiles(t *testing.T) {
	d := NewDiff()
	d.profiles = []string{"profile1"}

	view := d.View()

	if view == "" {
		t.Error("expected non-empty view")
	}

	if !contains(view, "Need at least 2 profiles") {
		t.Error("expected 'Need at least 2 profiles' in view")
	}
}

func TestDiffViewBasic(t *testing.T) {
	d := NewDiff()
	d.SetSize(80, 24)
	d.profiles = []string{"profile1", "profile2"}
	d.updateViewportContent()

	view := d.View()

	if view == "" {
		t.Error("expected non-empty view")
	}

	if !contains(view, "Profile Diff") {
		t.Error("expected 'Profile Diff' in view")
	}

	// The selectors should be rendered
	// but without updateViewportContent, the viewports don't show the profiles
	// Let's just verify the view is not empty
	if len(view) < 10 {
		t.Errorf("expected longer view, got length %d", len(view))
	}
}

func TestDiffViewComputingDiff(t *testing.T) {
	d := NewDiff()
	d.SetSize(80, 24)
	d.profiles = []string{"profile1", "profile2"}
	d.leftProfile = "profile1"
	d.rightProfile = "profile2"
	d.diffResult = nil // Computing

	view := d.View()

	if view == "" {
		t.Error("expected non-empty view")
	}

	if !contains(view, "Computing diff...") {
		t.Error("expected 'Computing diff...' in view")
	}
}

func TestDiffViewWithDiffResult(t *testing.T) {
	d := NewDiff()
	d.SetSize(80, 24)
	d.profiles = []string{"profile1", "profile2"}
	d.leftProfile = "profile1"
	d.rightProfile = "profile2"
	d.diffResult = &diff.DiffResult{
		Left:  []diff.DiffLine{{Type: diff.DiffEqual, Text: "same line"}},
		Right: []diff.DiffLine{{Type: diff.DiffEqual, Text: "same line"}},
	}

	d.updateViewportContent()

	view := d.View()

	if view == "" {
		t.Error("expected non-empty view")
	}

	// Should contain the diff panes
	if !contains(view, "same line") {
		t.Error("expected diff content in view")
	}
}

func TestDiffRenderDiffPane(t *testing.T) {
	d := NewDiff()

	lines := []diff.DiffLine{
		{Type: diff.DiffEqual, Text: "same line"},
		{Type: diff.DiffAdded, Text: "added line"},
		{Type: diff.DiffRemoved, Text: "removed line"},
	}

	// Test left pane
	result := d.renderDiffPane(lines, true)

	if result == "" {
		t.Error("expected non-empty result")
	}

	if !contains(result, "- removed line") {
		t.Error("expected '- removed line' in left pane")
	}

	// Test right pane
	result = d.renderDiffPane(lines, false)

	if !contains(result, "+ added line") {
		t.Error("expected '+ added line' in right pane")
	}

	if !contains(result, "same line") {
		t.Error("expected 'same line' in both panes")
	}
}

func TestDiffRenderSelectorNotSelecting(t *testing.T) {
	d := NewDiff()
	d.SetSize(80, 24)
	d.profiles = []string{"profile1", "profile2"}

	result := d.renderSelector("profile1", false, 0, true)

	if result == "" {
		t.Error("expected non-empty result")
	}

	if !contains(result, "profile1") {
		t.Error("expected profile name in result")
	}
}

func TestDiffRenderSelectorSelecting(t *testing.T) {
	d := NewDiff()
	d.SetSize(80, 24)
	d.profiles = []string{"profile1", "profile2", "profile3"}

	result := d.renderSelector("", true, 0, true)

	if result == "" {
		t.Error("expected non-empty result")
	}

	// Should show all profiles
	if !contains(result, "profile1") {
		t.Error("expected 'profile1' in result")
	}

	if !contains(result, "> ") {
		t.Error("expected cursor marker in result")
	}

	if !contains(result, "profile2") {
		t.Error("expected 'profile2' in result")
	}

	if !contains(result, "profile3") {
		t.Error("expected 'profile3' in result")
	}
}

// Helper type for testing
type testError struct{}

func (e *testError) Error() string {
	return "test error"
}
