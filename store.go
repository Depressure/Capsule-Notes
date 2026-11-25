package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

// dataDir returns the XDG data directory for the app.
func dataDir() (string, error) {
	if d := os.Getenv("XDG_DATA_HOME"); d != "" {
		return filepath.Join(d, "capsule-notes"), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".local", "share", "capsule-notes"), nil
}

func notesDir() (string, error) {
	dir, err := dataDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "notes"), nil
}

func sessionPath() (string, error) {
	dir, err := dataDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "session"), nil
}

func noteFileName(i int) string {
	return fmt.Sprintf("%04d.txt", i+1)
}

func parseNoteFile(name string) (int, bool) {
	if !strings.HasSuffix(name, ".txt") {
		return 0, false
	}
	num, err := strconv.Atoi(strings.TrimSuffix(name, ".txt"))
	if err != nil || num < 1 {
		return 0, false
	}
	return num, true
}

func writeFileAtomic(path, content string) {
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, []byte(content), 0o644); err != nil {
		return
	}
	os.Rename(tmp, path)
}

// save writes every changed note to its own text file, removes files of
// closed tabs, and stores the active tab index in the session file.
func (a *appState) save() {
	dir, err := notesDir()
	if err != nil {
		return
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return
	}
	for i, n := range a.notes {
		text := n.editor.Text()
		if n.persisted && text == n.saved {
			continue
		}
		writeFileAtomic(filepath.Join(dir, noteFileName(i)), text)
		n.saved = text
		n.persisted = true
	}
	entries, err := os.ReadDir(dir)
	if err == nil {
		for _, e := range entries {
			if num, ok := parseNoteFile(e.Name()); ok && num > len(a.notes) {
				os.Remove(filepath.Join(dir, e.Name()))
			}
		}
	}
	if sp, err := sessionPath(); err == nil {
		writeFileAtomic(sp, strconv.Itoa(a.active)+"\n")
	}
	a.dirty = false
	a.lastSave = time.Now()
}

// load restores all note files (in filename order) and the active tab.
func (a *appState) load() {
	dir, err := notesDir()
	if err != nil {
		return
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}
	var files []string
	for _, e := range entries {
		if !e.IsDir() {
			if _, ok := parseNoteFile(e.Name()); ok {
				files = append(files, e.Name())
			}
		}
	}
	sort.Strings(files)
	for _, f := range files {
		data, err := os.ReadFile(filepath.Join(dir, f))
		if err != nil {
			continue
		}
		n := &note{saved: string(data), persisted: true}
		n.editor.SetText(string(data))
		a.notes = append(a.notes, n)
	}
	if sp, err := sessionPath(); err == nil {
		if data, err := os.ReadFile(sp); err == nil {
			if idx, err := strconv.Atoi(strings.TrimSpace(string(data))); err == nil && idx >= 0 && idx < len(a.notes) {
				a.active = idx
			}
		}
	}
}
