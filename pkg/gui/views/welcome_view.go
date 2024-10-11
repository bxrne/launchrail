package views

import (
	"image/color"

	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"github.com/bxrne/launchrail/internal/logs"
	"github.com/bxrne/launchrail/pkg/gui"
)

func WelcomeView(logger *logs.Logger, onQuit func()) *gui.View {
	bg := canvas.NewRectangle(color.RGBA{R: 30, G: 30, B: 30, A: 255})

	headerSection := gui.NewVBoxContainer(
		gui.NewLabel("Welcome to Launchrail!"),
		gui.NewLabel("Your ultimate project management tool"),
		gui.NewImage("logo_title_128.png", float32(192), float32(192)),
	)

	heroSection := gui.NewVBoxContainer(
		gui.NewLabel("🚀 Launchrail"),
		gui.NewLabel("Manage your projects efficiently"),
	)

	contentSection := gui.NewVBoxContainer(
		gui.NewLabel("Recent Projects"),
		gui.NewButton("Create New Project", func() {
			logger.Debug("Create New Project button clicked")
		}),
		gui.NewButton("Open Existing Project", func() {
			logger.Debug("Open Existing Project button clicked")
		}),
	)

	footerSection := gui.NewVBoxContainer(
		gui.NewButton("Settings", func() {
			logger.Debug("Settings button clicked")
		}),
		gui.NewButton("Help", func() {
			logger.Debug("Help button clicked")
		}),
		gui.NewButton("Quit", onQuit),
	)

	content := gui.NewVBoxContainer(headerSection, heroSection, contentSection, footerSection)
	mainContainer := container.NewStack(bg, content)

	return gui.NewView("Welcome", mainContainer)
}
