package app

import (
	"context"
	"log"
	"os"

	"github.com/gdamore/tcell/v2"
	lsp "github.com/uditrawat03/bitcode/internal/lsp_client"
	"github.com/uditrawat03/bitcode/internal/ui"
)

type App struct {
	ctx       context.Context
	screen    tcell.Screen
	ui        *ui.ScreenManager
	lspServer *lsp.Client
	logger    *log.Logger
	running   bool
}

func CreateApp(ctx context.Context, logger *log.Logger) *App {
	return &App{
		running:   true,
		lspServer: lsp.NewServer(ctx, logger),
		logger:    logger,
		ctx:       ctx,
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
	app.ui = ui.CreateScreenManager(app.ctx, app.logger, app.lspServer, cwd)

	width, height := screen.Size()
	app.ui.InitComponents(width, height)

	// app.lspServer.Start("/var/www/html/MindFreak/GO/lsp-server/main", "--stdio")

	app.lspServer.Start("phpactor", "language-server")

	rootUri := "file://" + cwd
	_, err = app.lspServer.Initialize(rootUri)
	if err != nil {
		return err
	}

	app.lspServer.Initialized()

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
			if ev.Key() == tcell.KeyEscape && !app.ui.IsTooltipVisible() && !app.ui.IsDialogOpen() {
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
