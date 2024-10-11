package gui

import (
	"fyne.io/fyne/v2"
)

type Window struct {
	App    fyne.App
	Window fyne.Window
}

func NewWindow(app fyne.App, title string, width float32, height float32) *Window {
	win := app.NewWindow(title)
	win.Resize(fyne.NewSize(width, height))
	win.CenterOnScreen()

	return &Window{
		App:    app,
		Window: win,
	}
}

func (w *Window) SetContent(view *View) {
	w.Window.SetContent(view.Content)
	w.Window.SetTitle(view.Title)
	w.Window.SetPadded(false)
}

func (w *Window) ShowAndRun() {
	w.Window.ShowAndRun()
}
