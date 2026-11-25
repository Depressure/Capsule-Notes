package main

import (
	"fmt"
	"image"
	"image/color"
	"os"
	"strings"
	"time"
	"unicode/utf8"

	"gioui.org/font/gofont"
	"gioui.org/io/event"
	"gioui.org/io/key"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
)

const saveInterval = 2 * time.Second

type appState struct {
	th *material.Theme

	notes  []*note
	active int

	newTab widget.Clickable

	dirty      bool
	lastEdit   time.Time
	lastSave   time.Time
	lastSynced string
}

func newApp() *appState {
	th := material.NewTheme()
	th.Shaper = text.NewShaper(text.WithCollection(gofont.Collection()))
	th.TextSize = unit.Sp(15)
	a := &appState{th: th, lastSave: time.Now()}
	a.load()
	if len(a.notes) == 0 {
		a.notes = append(a.notes, &note{})
	}
	a.lastSynced = a.notes[a.active].editor.Text()
	return a
}

func (a *appState) frame(gtx layout.Context) {
	event.Op(gtx.Ops, a)
	a.handleKeys(gtx)
	a.syncText()
	if a.dirty && time.Since(a.lastEdit) >= saveInterval {
		a.save()
	}
	a.layout(gtx)
	a.handleClicks(gtx)
	gtx.Execute(op.InvalidateCmd{At: gtx.Now.Add(saveInterval)})
}

func (a *appState) handleKeys(gtx layout.Context) {
	filters := []event.Filter{
		key.Filter{Name: "T", Required: key.ModCtrl},
		key.Filter{Name: "W", Required: key.ModCtrl},
		key.Filter{Name: "S", Required: key.ModCtrl},
		key.Filter{Name: "Q", Required: key.ModCtrl},
		key.Filter{Name: key.NameTab, Required: key.ModCtrl, Optional: key.ModShift},
	}
	for {
		e, ok := gtx.Event(filters...)
		if !ok {
			break
		}
		ke, ok := e.(key.Event)
		if !ok || ke.State != key.Press {
			continue
		}
		switch ke.Name {
		case "T":
			a.addNote(gtx, "")
		case "W":
			a.closeTab(gtx, a.active)
		case "S":
			a.save()
		case "Q":
			a.save()
			os.Exit(0)
		case key.NameTab:
			if ke.Modifiers.Contain(key.ModShift) {
				a.selectTab(gtx, (a.active-1+len(a.notes))%len(a.notes))
			} else {
				a.selectTab(gtx, (a.active+1)%len(a.notes))
			}
		}
	}
}

func (a *appState) handleClicks(gtx layout.Context) {
	if a.newTab.Clicked(gtx) {
		a.addNote(gtx, "")
		return
	}
	for i, n := range a.notes {
		if n.close.Clicked(gtx) {
			a.closeTab(gtx, i)
			return
		}
		if n.tab.Clicked(gtx) {
			a.selectTab(gtx, i)
		}
	}
}

func (a *appState) syncText() {
	t := a.notes[a.active].editor.Text()
	if t != a.lastSynced {
		a.lastSynced = t
		a.dirty = true
		a.lastEdit = time.Now()
	}
}

func (a *appState) selectTab(gtx layout.Context, i int) {
	if i < 0 || i >= len(a.notes) {
		return
	}
	a.active = i
	a.lastSynced = a.notes[i].editor.Text()
	gtx.Execute(key.FocusCmd{Tag: &a.notes[i].editor})
}

func (a *appState) addNote(gtx layout.Context, text string) {
	n := &note{}
	n.editor.SetText(text)
	a.notes = append(a.notes, n)
	a.selectTab(gtx, len(a.notes)-1)
	a.dirty = true
	a.lastEdit = time.Now()
}

func (a *appState) closeTab(gtx layout.Context, i int) {
	if i < 0 || i >= len(a.notes) {
		return
	}
	if len(a.notes) == 1 {
		n := a.notes[0]
		n.editor.SetText("")
		a.lastSynced = ""
		a.dirty = true
		a.lastEdit = time.Now()
		return
	}
	a.notes = append(a.notes[:i], a.notes[i+1:]...)
	if a.active > i {
		a.active--
	}
	if a.active >= len(a.notes) {
		a.active = len(a.notes) - 1
	}
	a.lastSynced = a.notes[a.active].editor.Text()
	gtx.Execute(key.FocusCmd{Tag: &a.notes[a.active].editor})
	a.dirty = true
	a.lastEdit = time.Now()
}

func (a *appState) layout(gtx layout.Context) layout.Dimensions {
	paint.Fill(gtx.Ops, palette.bg)
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		layout.Rigid(a.layoutTabBar),
		layout.Flexed(1, a.layoutEditor),
		layout.Rigid(a.layoutStatusBar),
	)
}

func (a *appState) layoutTabBar(gtx layout.Context) layout.Dimensions {
	m := op.Record(gtx.Ops)
	dims := layout.Inset{Top: unit.Dp(8), Left: unit.Dp(10), Right: unit.Dp(10)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		tabList := layout.List{Axis: layout.Horizontal}
		return tabList.Layout(gtx, len(a.notes)+1, func(gtx layout.Context, i int) layout.Dimensions {
			if i < len(a.notes) {
				return a.layoutTab(gtx, i, a.notes[i])
			}
			return a.layoutNewTab(gtx)
		})
	})
	c := m.Stop()
	paint.FillShape(gtx.Ops, palette.panel, clip.Rect(image.Rectangle{Max: dims.Size}).Op())
	paint.FillShape(gtx.Ops, palette.line, clip.Rect(image.Rectangle{
		Min: image.Point{Y: dims.Size.Y - 1},
		Max: dims.Size,
	}).Op())
	c.Add(gtx.Ops)
	return dims
}

func (a *appState) layoutTab(gtx layout.Context, i int, n *note) layout.Dimensions {
	active := i == a.active
	return n.tab.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		m := op.Record(gtx.Ops)
		dims := layout.Inset{Top: unit.Dp(6), Bottom: unit.Dp(7), Left: unit.Dp(12), Right: unit.Dp(8)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return a.label(gtx, noteTitle(n.editor.Text()), tabTextColor(active, n.tab.Hovered()))
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return n.close.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						c := palette.dim
						if n.close.Hovered() {
							c = palette.fg
						}
						return layout.Inset{Left: unit.Dp(6), Top: unit.Dp(2), Bottom: unit.Dp(2), Right: unit.Dp(2)}.Layout(gtx,
							func(gtx layout.Context) layout.Dimensions {
								return a.label(gtx, "×", c)
							})
					})
				}),
			)
		})
		c := m.Stop()
		bg := palette.tab
		switch {
		case active:
			bg = palette.tabActive
		case n.tab.Hovered():
			bg = palette.tabHover
		}
		paint.FillShape(gtx.Ops, bg, clip.Rect(image.Rectangle{Max: dims.Size}).Op())
		if active {
			paint.FillShape(gtx.Ops, palette.accent, clip.Rect(image.Rectangle{
				Min: image.Point{Y: dims.Size.Y - 2},
				Max: dims.Size,
			}).Op())
		}
		c.Add(gtx.Ops)
		return dims
	})
}

func (a *appState) layoutNewTab(gtx layout.Context) layout.Dimensions {
	return a.newTab.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		m := op.Record(gtx.Ops)
		dims := layout.Inset{Top: unit.Dp(6), Bottom: unit.Dp(7), Left: unit.Dp(12), Right: unit.Dp(12)}.Layout(gtx,
			func(gtx layout.Context) layout.Dimensions {
				return a.label(gtx, "+", palette.dim)
			})
		c := m.Stop()
		bg := palette.tab
		if a.newTab.Hovered() {
			bg = palette.tabHover
		}
		paint.FillShape(gtx.Ops, bg, clip.Rect(image.Rectangle{Max: dims.Size}).Op())
		c.Add(gtx.Ops)
		return dims
	})
}

func (a *appState) layoutEditor(gtx layout.Context) layout.Dimensions {
	n := a.notes[a.active]
	return layout.Inset{Top: unit.Dp(10), Bottom: unit.Dp(10), Left: unit.Dp(16), Right: unit.Dp(16)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		ed := material.Editor(a.th, &n.editor, "Type your note here…")
		ed.Color = palette.fg
		ed.HintColor = palette.hint
		ed.SelectionColor = palette.selection
		ed.TextSize = unit.Sp(16)
		return ed.Layout(gtx)
	})
}

func (a *appState) layoutStatusBar(gtx layout.Context) layout.Dimensions {
	n := a.notes[a.active]
	left := fmt.Sprintf("%s — %d words, %d chars", noteTitle(n.editor.Text()),
		len(strings.Fields(n.editor.Text())), utf8.RuneCountInString(n.editor.Text()))
	right := "saved " + a.lastSave.Format("15:04:05")
	rightColor := palette.dim
	if a.dirty {
		right = "● unsaved changes"
		rightColor = palette.gold
	}
	hint := "Ctrl+T new · Ctrl+W close · Ctrl+S save · Ctrl+Tab cycle"
	m := op.Record(gtx.Ops)
	dims := layout.Inset{Top: unit.Dp(5), Bottom: unit.Dp(5), Left: unit.Dp(12), Right: unit.Dp(12)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx,
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return a.label(gtx, left, palette.dim)
			}),
			layout.Flexed(1, layout.Spacer{}.Layout),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return a.label(gtx, hint, palette.hint)
			}),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layout.Inset{Left: unit.Dp(16)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					return a.label(gtx, right, rightColor)
				})
			}),
		)
	})
	c := m.Stop()
	paint.FillShape(gtx.Ops, palette.panel, clip.Rect(image.Rectangle{Max: dims.Size}).Op())
	paint.FillShape(gtx.Ops, palette.line, clip.Rect(image.Rectangle{
		Max: image.Point{X: dims.Size.X, Y: 1},
	}).Op())
	c.Add(gtx.Ops)
	return dims
}

func (a *appState) label(gtx layout.Context, txt string, c color.NRGBA) layout.Dimensions {
	l := material.Label(a.th, unit.Sp(12), txt)
	l.Color = c
	l.MaxLines = 1
	return l.Layout(gtx)
}

func tabTextColor(active, hovered bool) color.NRGBA {
	switch {
	case active:
		return palette.fg
	case hovered:
		return palette.dim
	default:
		return palette.dim
	}
}
