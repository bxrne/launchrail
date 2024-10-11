package main

import (
	"log"

	"fyne.io/fyne/v2/app"

	"github.com/bxrne/launchrail/internal/config"
	"github.com/bxrne/launchrail/internal/logs"
	"github.com/bxrne/launchrail/pkg/gui"
	"github.com/bxrne/launchrail/pkg/gui/views"

	_ "github.com/joho/godotenv/autoload"
)

func main() {
	cfg := config.NewConfig()

	logger, err := logs.NewLogger("main", cfg.GetLogLevel())
	if err != nil {
		log.Fatalf("Error creating logger: %v", err)
	}
	defer logger.Close()

	logger.Info("Starting up...")

	app := app.New()
	window := gui.NewWindow(app, "launchrail", cfg.GetWindowWidth(), cfg.GetWindowHeight())

	viewLogger, err := logs.NewLogger("view", cfg.GetLogLevel())
	if err != nil {
		log.Fatalf("Error creating view logger: %v", err)
	}
	welcomeView := views.WelcomeView(viewLogger, app.Quit)

	window.SetContent(welcomeView)

	logger.Info("Showing window...")
	window.ShowAndRun()

	logger.Info("Exiting...")
}
