package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNoteTitle(t *testing.T) {
	if got := noteTitle(""); got != "New Note" {
		t.Errorf("empty note: got %q, want %q", got, "New Note")
	}
	if got := noteTitle("  \n \n\t\n"); got != "New Note" {
		t.Errorf("blank lines: got %q, want %q", got, "New Note")
	}
	if got := noteTitle("\n\n  first line\nsecond line"); got != "first line" {
		t.Errorf("first non-empty line: got %q, want %q", got, "first line")
	}
	long := strings.Repeat("x", 100)
	got := noteTitle(long)
	if !strings.HasSuffix(got, "…") || len([]rune(got)) != maxTitleRunes+1 {
		t.Errorf("long title not truncated to %d runes + ellipsis: %q", maxTitleRunes, got)
	}
}

func TestStoreRoundTrip(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", dir)

	a := &appState{dirty: true}
	n1 := &note{}
	n1.editor.SetText("hello capsuleer\nsecond line")
	a.notes = append(a.notes, n1)
	n2 := &note{}
	n2.editor.SetText("planetary industry routes")
	a.notes = append(a.notes, n2)
	a.active = 1
	a.save()

	notesDir := filepath.Join(dir, "capsule-notes", "notes")
	for i, want := range []string{"hello capsuleer\nsecond line", "planetary industry routes"} {
		data, err := os.ReadFile(filepath.Join(notesDir, noteFileName(i)))
		if err != nil {
			t.Fatalf("note file %d not written: %v", i, err)
		}
		if string(data) != want {
			t.Errorf("note %d: got %q, want %q", i, data, want)
		}
	}
	session, err := os.ReadFile(filepath.Join(dir, "capsule-notes", "session"))
	if err != nil {
		t.Fatalf("session file not written: %v", err)
	}
	if strings.TrimSpace(string(session)) != "1" {
		t.Errorf("session file: got %q, want %q", session, "1")
	}
	if a.dirty {
		t.Error("dirty flag not cleared after save")
	}

	b := &appState{}
	b.load()
	if len(b.notes) != 2 {
		t.Fatalf("loaded %d notes, want 2", len(b.notes))
	}
	if got := b.notes[0].editor.Text(); got != "hello capsuleer\nsecond line" {
		t.Errorf("note 1 text: got %q", got)
	}
	if got := b.notes[1].editor.Text(); got != "planetary industry routes" {
		t.Errorf("note 2 text: got %q", got)
	}
	if b.active != 1 {
		t.Errorf("active tab: got %d, want 1", b.active)
	}
}

func TestSaveRemovesStaleFiles(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", dir)

	nd := filepath.Join(dir, "capsule-notes", "notes")
	if err := os.MkdirAll(nd, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(nd, "0003.txt"), []byte("old note"), 0o644); err != nil {
		t.Fatal(err)
	}

	a := &appState{dirty: true}
	n := &note{}
	n.editor.SetText("only note")
	a.notes = append(a.notes, n)
	a.save()

	if _, err := os.Stat(filepath.Join(nd, "0003.txt")); !os.IsNotExist(err) {
		t.Errorf("stale 0003.txt still exists after saving 1 note")
	}
	if _, err := os.Stat(filepath.Join(nd, "0001.txt")); err != nil {
		t.Errorf("0001.txt missing: %v", err)
	}
}

func TestLoadMissingDir(t *testing.T) {
	t.Setenv("XDG_DATA_HOME", t.TempDir())
	a := &appState{}
	a.load()
	if len(a.notes) != 0 {
		t.Errorf("expected no notes without a data dir, got %d", len(a.notes))
	}
}
