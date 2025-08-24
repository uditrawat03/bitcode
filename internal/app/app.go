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
		ctx:       ctx,
		running:   true,
		lspServer: lsp.NewServer(ctx, logger),
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

	// Create ScreenManager using new UIComponent + Focusable system
	app.ui = ui.NewScreenManager(app.ctx, screen, app.logger, app.lspServer)

	// Add components to ScreenManager
	app.ui.SetupComponents(cwd) // We'll define this inside ScreenManager

	// Start LSP server
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
	// ScreenManager handles event loop and rendering
	app.ui.Run()
	app.logger.Println("Application has exited gracefully.")
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
