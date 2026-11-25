package main

import (
	"os"

	"gioui.org/app"
	"gioui.org/op"
)

func main() {
	a := newApp()
	go func() {
		w := new(app.Window)
		w.Option(
			app.Title("Capsule Notes"),
			app.Size(960, 640),
			app.MinSize(520, 360),
		)
		var ops op.Ops
		for {
			switch e := w.Event().(type) {
			case app.DestroyEvent:
				a.save()
				os.Exit(0)
			case app.FrameEvent:
				gtx := app.NewContext(&ops, e)
				a.frame(gtx)
				e.Frame(gtx.Ops)
			}
		}
	}()
	app.Main()
}
