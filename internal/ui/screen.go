package ui

import (
	"context"
	"log"

	"github.com/gdamore/tcell/v2"
	"github.com/uditrawat03/bitcode/internal/buffer"
	"github.com/uditrawat03/bitcode/internal/layout"
	lsp "github.com/uditrawat03/bitcode/internal/lsp_client"
)

type Focusable interface {
	Focus()
	Blur()
	IsFocused() bool
	HandleKey(*tcell.EventKey)
	HandleMouse(*tcell.EventMouse)
}

type ScreenManager struct {
	ctx       context.Context
	screen    tcell.Screen
	logger    *log.Logger
	Lm        *layout.LayoutManager
	Bm        *buffer.BufferManager
	UIManager *UIManager

	focusOrder []UIComponent
	focusedIdx int
	running    bool
}

func NewScreenManager(ctx context.Context, screen tcell.Screen, logger *log.Logger, lsp *lsp.Client) *ScreenManager {
	lm := layout.CreateLayoutManager()
	uiManager := NewUIManager(logger, lm)

	return &ScreenManager{
		ctx:       ctx,
		screen:    screen,
		logger:    logger,
		Lm:        lm,
		UIManager: uiManager,
		Bm:        buffer.NewBufferManager(ctx, lsp),
		running:   true,
	}
}

// RequestRender triggers a redraw
func (sm *ScreenManager) RequestRender() {
	sm.Render()
}

func (sm *ScreenManager) SetupComponents(rootPath string) {
	sw, sh := sm.screen.Size()
	sm.Lm.UpdateLayout(sw, sh)
	l := sm.Lm.GetLayout()

	// TopBar
	topBar := NewTopBar("Bitcode IDE")
	rect := l.GetTopBarArea(sw, sh)
	topBar.SetPosition(rect.X, rect.Y)
	topBar.Resize(rect.Width, rect.Height)
	sm.AddComponent(topBar, false)

	// StatusBar
	statusBar := NewBottomBar("Ready")
	rect = l.GetStatusBarArea(sw, sh)
	statusBar.SetPosition(rect.X, rect.Y)
	statusBar.Resize(rect.Width, rect.Height)
	sm.AddComponent(statusBar, false)

	// Sidebar + TreeView
	treeView := NewTreeView(rootPath)
	sidebar := NewSidebar(treeView, sm.Lm.GetLayout().SidebarWidth, sm.RequestRender)

	rect = l.GetSidebarArea(sw, sh)
	sidebar.SetPosition(rect.X, rect.Y)
	sidebar.Resize(rect.Width, rect.Height)
	sm.AddComponent(sidebar, true)

	// Editor
	editor := NewCodeArea()
	rect = l.GetEditorArea(sw, sh)
	editor.SetPosition(rect.X, rect.Y)
	editor.Resize(rect.Width, rect.Height)

	sm.AddComponent(editor, true)

	sm.focusOrder = []UIComponent{sidebar, editor}
	sm.focusedIdx = 0
	if f, ok := sm.focusOrder[sm.focusedIdx].(Focusable); ok {
		f.Focus()
	}
}

func (sm *ScreenManager) AddComponent(c UIComponent, focusable bool) {
	sm.UIManager.AddComponent(c)
	if focusable {
		sm.focusOrder = append(sm.focusOrder, c)
	}
}

func (sm *ScreenManager) Run() {
	for sm.running {
		ev := sm.screen.PollEvent()
		if ev == nil {
			continue
		}

		switch e := ev.(type) {
		case *tcell.EventKey:
			sm.handleKey(e)
		case *tcell.EventResize:
			sm.screen.Clear()
			sm.Render()
		case *tcell.EventMouse:
			sm.handleMouse(e)
		}
	}
}

func (sm *ScreenManager) handleKey(ev *tcell.EventKey) {
	if ev.Key() == tcell.KeyEscape {
		sm.running = false
		return
	}
	if len(sm.focusOrder) > 0 {
		if f, ok := sm.focusOrder[sm.focusedIdx].(Focusable); ok {
			f.HandleKey(ev)
		}
	}
}

func (sm *ScreenManager) handleMouse(ev *tcell.EventMouse) {
	if len(sm.focusOrder) > 0 {
		if f, ok := sm.focusOrder[sm.focusedIdx].(Focusable); ok {
			f.HandleMouse(ev)
		}
	}
}

func (sm *ScreenManager) Render() {
	sm.UIManager.RenderAll(sm.screen)
	sm.screen.Show()
}
