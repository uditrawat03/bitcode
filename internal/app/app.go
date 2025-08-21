package app

import (
	"log"

	"github.com/gdamore/tcell/v2"
	"github.com/uditrawat03/bitcode/internal/ui"
)

type App struct {
	screen  tcell.Screen
	ui      *ui.ScreenManager
	running bool
}

func CreateApp() *App {
	return &App{
		running: true,
	}
}

func (app *App) Initialize() error {
	screen, err := tcell.NewScreen()
	if err != nil {
		return err
	}

	if err := screen.Init(); err != nil {
		return err
	}

	// Enable mouse events
	screen.EnableMouse()
	screen.EnablePaste()

	app.screen = screen

	app.ui = ui.CreateScreenManager()

	// Initialize all UI components with screen dimensions
	screenWidth, screenHeight := screen.Size()
	app.ui.InitComponents(screenWidth, screenHeight)

	return nil
}

func (app *App) Run() {
	app.draw()

	for app.running {
		event := app.screen.PollEvent()
		if event == nil {
			continue
		}

		switch ev := event.(type) {
		case *tcell.EventKey:
			if ev.Key() == tcell.KeyEscape {
				app.running = false
			} else {
				app.ui.HandleKey(ev)
			}
		case *tcell.EventMouse:
			app.ui.HandleMouse(ev)
		}

		app.draw()
	}

	log.Println("Application has exited gracefully.")
}

func (app *App) draw() {
	app.screen.Clear()
	app.ui.Draw(app.screen)
	app.screen.Show()
}

func (app *App) Shutdown() {
	if app.screen != nil {
		app.screen.Fini()
		app.screen = nil
	}
	log.Println("Application shut down.")
}
