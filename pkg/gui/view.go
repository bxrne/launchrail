package gui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

type View struct {
	Title   string
	Content fyne.CanvasObject
}

func NewView(title string, content fyne.CanvasObject) *View {
	return &View{
		Title:   title,
		Content: content,
	}
}

func NewButton(label string, action func()) *widget.Button {
	return widget.NewButton(label, action)
}

func NewLabel(text string) *widget.Label {
	return widget.NewLabel(text)
}

func NewVBoxContainer(objects ...fyne.CanvasObject) *fyne.Container {
	return container.NewVBox(objects...)
}

func NewImage(name string, width float32, height float32) *canvas.Image {
	image := &canvas.Image{}
	image.File = "./assets/" + name
	image.FillMode = canvas.ImageFillContain
	image.SetMinSize(fyne.NewSize(width, height))
	return image
}
