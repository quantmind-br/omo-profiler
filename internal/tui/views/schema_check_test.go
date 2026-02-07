package views

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/diogenes/omo-profiler/internal/schema"
)

func TestNewSchemaCheck(t *testing.T) {
	sc := NewSchemaCheck()

	if sc.state != stateSchemaCheckLoading {
		t.Errorf("expected stateSchemaCheckLoading, got %v", sc.state)
	}

	if sc.spinner.Spinner.Frames[0] == "" {
		t.Error("expected spinner to be initialized")
	}

	if sc.keys.Esc.Help().Key == "" {
		t.Error("expected Esc key to be initialized")
	}
}

func TestSchemaCheck_Init(t *testing.T) {
	sc := NewSchemaCheck()
	cmd := sc.Init()

	if cmd == nil {
		t.Fatal("expected non-nil command from Init")
	}
}

func TestSchemaCheck_Update_ResultMsg_Identical(t *testing.T) {
	sc := NewSchemaCheck()
	result := &schema.CompareResult{
		Identical: true,
	}

	msg := schemaCheckResultMsg{
		result: result,
		err:    nil,
	}

	updated, cmd := sc.Update(msg)

	if cmd != nil {
		t.Error("expected nil command for result msg")
	}

	if updated.state != stateSchemaCheckResult {
		t.Errorf("expected stateSchemaCheckResult, got %v", updated.state)
	}

	if updated.result != result {
		t.Error("expected result to be set")
	}
}

func TestSchemaCheck_Update_ResultMsg_Different(t *testing.T) {
	sc := NewSchemaCheck()
	result := &schema.CompareResult{
		Identical: false,
		Diff:      "--- embedded\n+++ upstream\n+ new field",
	}

	msg := schemaCheckResultMsg{
		result: result,
		err:    nil,
	}

	updated, _ := sc.Update(msg)

	if updated.state != stateSchemaCheckResult {
		t.Errorf("expected stateSchemaCheckResult, got %v", updated.state)
	}

	if updated.result.Identical {
		t.Error("expected Identical to be false")
	}
}

func TestSchemaCheck_Update_ResultMsg_ErrorReal(t *testing.T) {
	sc := NewSchemaCheck()
	msg := schemaCheckResultMsg{
		result: nil,
		err:    &testError{},
	}

	updated, _ := sc.Update(msg)

	if updated.state != stateSchemaCheckError {
		t.Errorf("expected stateSchemaCheckError, got %v", updated.state)
	}

	if updated.errorMsg != "test error" {
		t.Errorf("expected errorMsg 'test error', got %q", updated.errorMsg)
	}
}

func TestSchemaCheck_Update_Esc(t *testing.T) {
	sc := NewSchemaCheck()
	msg := tea.KeyMsg{Type: tea.KeyEsc}

	_, cmd := sc.Update(msg)

	if cmd == nil {
		t.Fatal("expected non-nil command for Esc")
	}

	res := cmd()
	if _, ok := res.(SchemaCheckBackMsg); !ok {
		t.Errorf("expected SchemaCheckBackMsg, got %T", res)
	}
}

func TestDashboard_HasSchemaCheckMenuItem(t *testing.T) {
	d := NewDashboard()
	_ = d
	found := false
	for _, item := range menuItems {
		if item == "Check Schema Updates" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected 'Check Schema Updates' in dashboard menu items")
	}
}
