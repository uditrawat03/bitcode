package app

import (
	"log"
	"os"

	"github.com/gdamore/tcell/v2"
	lsp "github.com/uditrawat03/bitcode/internal/lsp_client"
	"github.com/uditrawat03/bitcode/internal/ui"
)

type App struct {
	screen    tcell.Screen
	ui        *ui.ScreenManager
	lspServer *lsp.Client
	logger    *log.Logger
	running   bool
}

func CreateApp(logger *log.Logger) *App {
	return &App{
		running:   true,
		lspServer: lsp.NewServer(logger),
		logger:    logger,
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

	cwd, err := os.Getwd()
	if err != nil {
		cwd = "." // fallback
	}

	screen.EnableMouse()
	screen.EnablePaste()

	app.screen = screen
	app.ui = ui.CreateScreenManager(app.lspServer, cwd)

	width, height := screen.Size()
	app.ui.InitComponents(width, height)

	app.lspServer.Start("/var/www/html/MindFreak/GO/lsp-server/main", "--stdio")

	app.lspServer.Initialize(cwd)

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
			if ev.Key() == tcell.KeyEscape && !app.ui.IsDialogOpen() {
				app.running = false
			} else {
				app.ui.HandleKey(ev)
			}
		case *tcell.EventResize:
			app.draw()
		case *tcell.EventMouse:
			app.ui.HandleMouse(ev)
		}

		app.draw()
	}

	app.logger.Println("Application has exited gracefully.")
}

func (app *App) draw() {
	app.screen.Clear()
	app.ui.Draw(app.screen)
	app.screen.Show()
}

func (app *App) Shutdown() {
	if app.lspServer != nil {
		app.lspServer.Stop()
		app.lspServer = nil
	}

	if app.screen != nil {
		app.screen.Fini()
		app.screen = nil
	}

	app.logger.Println("Application shut down.")
}
