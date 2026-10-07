package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
)

func key(code rune) tea.KeyPressMsg {
	return tea.KeyPressMsg(tea.Key{Code: code, Text: string(code)})
}

func TestFolderEntryAndBrowsingRunCleanup(t *testing.T) {
	root := t.TempDir()
	jpg := filepath.Join(root, "jpg")
	raw := filepath.Join(root, "raw")
	for _, dir := range []string{jpg, raw} {
		if err := os.Mkdir(dir, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(raw, "Drop.CR3"), []byte("photo"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(jpg, "Keep.JPG"), []byte("jpg"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(raw, "Keep.CR3"), []byte("raw"), 0o644); err != nil {
		t.Fatal(err)
	}
	a := newApp()
	model, _ := a.Update(key('e'))
	a = model.(app)
	if !a.editing {
		t.Fatal("path entry did not open")
	}
	a.input.SetValue(filepath.Join(root, "missing"))
	model, _ = a.Update(tea.KeyPressMsg(tea.Key{Code: tea.KeyEnter}))
	a = model.(app)
	if a.err == nil || a.stage != 0 {
		t.Fatalf("invalid path not reported: stage=%d, err=%v", a.stage, a.err)
	}
	a.input.SetValue(jpg)
	model, _ = a.Update(tea.KeyPressMsg(tea.Key{Code: tea.KeyEnter}))
	a = model.(app)
	if a.stage != 1 || a.editing || a.err != nil {
		t.Fatalf("JPEG selection failed: %+v", a)
	}
	a.picker.CurrentDirectory = raw
	model, cmd := a.Update(key('s'))
	a = model.(app)
	if a.stage != 2 || cmd == nil {
		t.Fatalf("RAW browser selection did not start cleanup: %+v", a)
	}
	model, _ = a.Update(cmd())
	a = model.(app)
	if a.stage != 3 || a.err != nil || a.result.Kept != 1 || a.result.Moved != 1 || a.result.Failed != 0 {
		t.Fatalf("cleanup result = %+v", a)
	}
	view := a.View().Content
	if !strings.Contains(view, "Kept: 1 · Moved: 1 · Failed: 0") || !strings.Contains(view, "Drop.CR3") || !strings.Contains(view, "tmp-delete") {
		t.Fatalf("result screen missing run details: %s", view)
	}
}

func TestResultScreenCanScrollThroughFiles(t *testing.T) {
	a := newApp()
	a.stage = 3
	a.height = 6
	a.details = []string{"first", "second", "third"}
	if view := a.View().Content; !strings.Contains(view, "first") || strings.Contains(view, "third") {
		t.Fatalf("unexpected first page: %s", view)
	}
	model, _ := a.Update(key('j'))
	a = model.(app)
	if view := a.View().Content; !strings.Contains(view, "second") || strings.Contains(view, "first") {
		t.Fatalf("unexpected scrolled page: %s", view)
	}
}

func TestResultScreenWrapsLongPathsWithinTerminal(t *testing.T) {
	a := newApp()
	a.stage = 3
	a.height = 12
	a.width = 40
	a.result.Moved = 1
	a.details = []string{
		"Moved: " + strings.Repeat("A", 100),
		"Failed: " + strings.Repeat("B", 100) + " END",
	}
	view := a.View().Content
	for _, line := range strings.Split(view, "\n") {
		if ansi.StringWidth(line) > a.width {
			t.Fatalf("line exceeds %d columns: %q", a.width, line)
		}
	}
	if lines := len(strings.Split(strings.TrimSuffix(view, "\n"), "\n")); lines > a.height {
		t.Fatalf("result uses %d rows in a %d-row terminal", lines, a.height)
	}
	for range 20 {
		model, _ := a.Update(key('j'))
		a = model.(app)
	}
	if !strings.Contains(a.View().Content, "END") {
		t.Fatal("cannot scroll to the end of a wrapped failure")
	}
}
