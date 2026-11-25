package main

import (
	"strings"

	"gioui.org/widget"
)

const maxTitleRunes = 22

// note is a single tab in the notepad. saved is the last persisted text;
// persisted reports whether the note exists on disk yet.
type note struct {
	editor    widget.Editor
	tab       widget.Clickable
	close     widget.Clickable
	saved     string
	persisted bool
}

// noteTitle derives a tab label from the first non-empty line of the note.
func noteTitle(text string) string {
	for _, line := range strings.Split(text, "\n") {
		if t := strings.TrimSpace(line); t != "" {
			r := []rune(t)
			if len(r) > maxTitleRunes {
				return string(r[:maxTitleRunes]) + "…"
			}
			return t
		}
	}
	return "New Note"
}
