# Capsule Notes

A desktop notepad for Arch Linux inspired by the in-game Notepad in EVE Online:
multiple tabs of plain text, a dark sci-fi UI, and automatic saving. Written in
pure Go with [Gio](https://gioui.org) (immediate-mode GUI, works on X11 and
Wayland).

## Features

- Tabbed notes — click a tab to switch, `×` or `Ctrl+W` to close, `+` or
  `Ctrl+T` to add a new note. The tab label is the first line of the note.
- Plain-text editing with selection, clipboard (`Ctrl+C/X/V`), undo/redo.
- Autosave every 2 seconds; `Ctrl+S` saves immediately, and notes are saved on
  exit. Each note is stored as its own plain text file, so you can edit or
  grep them with any editor.
- Session restore: tabs and contents are reloaded on start.
- Status bar with word/character count and save state.
- `Ctrl+Tab` / `Ctrl+Shift+Tab` cycle tabs, `Ctrl+Q` quits.

## Build dependencies (Arch Linux)

Gio uses cgo for keyboard handling (xkbcommon), Wayland/X11, and EGL/Vulkan:

```sh
sudo pacman -S base-devel libxkbcommon wayland libegl libglvnd vulkan-devel vulkan-icd-loader
```

At runtime you need a Vulkan ICD (e.g. your GPU driver, or `mesa` with lavapipe
for software rendering).

## Build and run

```sh
go build -o capsule-notes .
./capsule-notes
```

or directly:

```sh
go run .
```

## Test

```sh
go test ./...
```

## Data

Each note is a plain text file in `~/.local/share/capsule-notes/notes/`
(`0001.txt`, `0002.txt`, …, in tab order; override the location with
`XDG_DATA_HOME`). The active tab index is kept in
`~/.local/share/capsule-notes/session`. Delete the directory to start fresh.
