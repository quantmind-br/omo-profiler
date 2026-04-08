package profile

import (
	"reflect"
	"testing"
)

func TestNewBlankSelectionHasNothingSelected(t *testing.T) {
	selection := NewBlankSelection()

	for _, path := range allFieldPaths {
		if selection.IsSelected(path) {
			t.Fatalf("expected %q to be unselected", path)
		}
	}
}

func TestNewSelectionFromPresenceSeedsFromExistingProfile(t *testing.T) {
	selection := NewSelectionFromPresence(map[string]bool{
		"disabled_hooks": true,
		"agents":         true,
	})

	if !selection.IsSelected("disabled_hooks") {
		t.Fatal("expected disabled_hooks to be selected")
	}

	if !selection.IsSelected("agents.*.model") {
		t.Fatal("expected agents.*.model to be selected from top-level presence")
	}

	if !selection.IsSelected("agents.builder.model") {
		t.Fatal("expected agents.builder.model to inherit wildcard selection")
	}

	if selection.IsSelected("categories.*.model") {
		t.Fatal("expected categories.*.model to remain unselected")
	}
}

func TestToggleDoesNotAffectValues(t *testing.T) {
	selection := NewBlankSelection()
	path := "agents.*.model"
	value := "gpt-5"

	selection.SetSelected(path, true)
	selection.Toggle(path)
	if selection.IsSelected(path) {
		t.Fatal("expected path to be unselected after toggle")
	}
	if value != "gpt-5" {
		t.Fatal("selection toggle unexpectedly changed stored value")
	}

	selection.Toggle(path)
	if !selection.IsSelected(path) {
		t.Fatal("expected path to be selected after second toggle")
	}
	if value != "gpt-5" {
		t.Fatal("selection toggle unexpectedly changed stored value")
	}
}

func TestWildcardMatching(t *testing.T) {
	selection := NewBlankSelection()
	selection.SetSelected("agents.*.model", true)

	if !selection.IsSelected("agents.build.model") {
		t.Fatal("expected wildcard agent selection to match concrete agent path")
	}
}

func TestCloneIndependence(t *testing.T) {
	original := NewBlankSelection()
	original.SetSelected("disabled_hooks", true)

	clone := original.Clone()
	clone.SetSelected("disabled_hooks", false)
	clone.SetSelected("agents.*.model", true)

	if !original.IsSelected("disabled_hooks") {
		t.Fatal("expected original selection to remain unchanged")
	}
	if original.IsSelected("agents.*.model") {
		t.Fatal("expected clone mutation not to affect original")
	}
}

func TestSelectedPathsReturnsSorted(t *testing.T) {
	selection := NewBlankSelection()
	selection.SetSelected("tmux.layout", true)
	selection.SetSelected("agents.*.model", true)
	selection.SetSelected("disabled_hooks", true)

	want := []string{"agents.*.model", "disabled_hooks", "tmux.layout"}
	if got := selection.SelectedPaths(); !reflect.DeepEqual(got, want) {
		t.Fatalf("SelectedPaths() = %v, want %v", got, want)
	}
}
