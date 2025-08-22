package ui

import (
	"github.com/gdamore/tcell/v2"
	dialog "github.com/uditrawat03/bitcode/internal/Dialog"
	"github.com/uditrawat03/bitcode/internal/buffer"
	"github.com/uditrawat03/bitcode/internal/editor"
	"github.com/uditrawat03/bitcode/internal/layout"
	lsp "github.com/uditrawat03/bitcode/internal/lsp_client"
	"github.com/uditrawat03/bitcode/internal/sidebar"
	"github.com/uditrawat03/bitcode/internal/statusbar"
	"github.com/uditrawat03/bitcode/internal/topbar"
)

type Focusable interface {
	Focus()
	Blur()
	IsFocused() bool
	HandleKey(*tcell.EventKey)
	HandleMouse(*tcell.EventMouse)
	Draw(tcell.Screen)
}

type ScreenManager struct {
	layoutManager *layout.LayoutManager
	bufferManager *buffer.BufferManager
	screen        tcell.Screen

	editor    *editor.Editor
	sidebar   *sidebar.Sidebar
	topBar    *topbar.TopBar
	statusBar *statusbar.StatusBar

	dialog *dialog.Dialog

	focusOrder []Focusable
	focusedIdx int
	rootPath   string
}

func CreateScreenManager(lsp *lsp.Client, rootPath string) *ScreenManager {
	sm := &ScreenManager{
		layoutManager: layout.CreateLayoutManager(),
		bufferManager: buffer.NewBufferManager(lsp),
		rootPath:      rootPath,
	}
	return sm
}

// Initialize components and focus order
func (sm *ScreenManager) InitComponents(screenWidth, screenHeight int) {
	// Update layout
	sm.layoutManager.UpdateLayout(screenWidth, screenHeight)
	l := sm.layoutManager.GetLayout()

	// TopBar
	tb := l.GetTopBarArea(screenWidth, screenHeight)
	sm.topBar = topbar.CreateTopBar(tb.X, tb.Y, tb.Width, tb.Height)

	// Sidebar
	sb := l.GetSidebarArea(screenWidth, screenHeight)
	sm.sidebar = sidebar.CreateSidebar(sb.X, sb.Y, sb.Width, sb.Height, sm.rootPath)
	sm.sidebar.SetOnFileOpen(func(path string) {
		buf := sm.bufferManager.Open(path)
		sm.editor.SetBuffer(buf)
	})

	sm.sidebar.SetFocusCallback(func() {
		sm.focusOrder[sm.focusedIdx].Blur()
		sm.focusedIdx = 0 // sidebar index
		sm.focusOrder[sm.focusedIdx].Focus()
	})

	// Editor
	ed := l.GetEditorArea(screenWidth, screenHeight)
	sm.editor = editor.CreateEditor(ed.X, ed.Y, ed.Width, ed.Height)
	sm.editor.SetFocusCallback(func() {
		sm.focusOrder[sm.focusedIdx].Blur()
		sm.focusedIdx = 1 // editor index in focusOrder
		sm.focusOrder[sm.focusedIdx].Focus()
	})

	// StatusBar
	st := l.GetStatusBarArea(screenWidth, screenHeight)
	sm.statusBar = statusbar.CreateStatusBar(st.X, st.Y, st.Width, st.Height)

	// Set focus order
	sm.focusOrder = []Focusable{
		sm.sidebar,
		sm.editor,
		sm.topBar,
		sm.statusBar,
	}

	sm.focusedIdx = 0
	sm.focusOrder[sm.focusedIdx].Focus()
}

// Switch focus to next component
func (sm *ScreenManager) FocusNext() {
	if sm.dialog != nil {
		// dialog keeps focus if open
		return
	}
	sm.focusOrder[sm.focusedIdx].Blur()
	sm.focusedIdx = (sm.focusedIdx + 1) % len(sm.focusOrder)
	sm.focusOrder[sm.focusedIdx].Focus()
}

// Draw all components
func (sm *ScreenManager) Draw(screen tcell.Screen) {
	sm.screen = screen
	screenWidth, screenHeight := sm.screen.Size()
	sm.layoutManager.UpdateLayout(screenWidth, screenHeight)

	l := sm.layoutManager.GetLayout()
	// Update positions & sizes
	tb := l.GetTopBarArea(screenWidth, screenHeight)
	sm.topBar.Resize(tb.X, tb.Y, tb.Width, tb.Height)

	sb := l.GetSidebarArea(screenWidth, screenHeight)
	sm.sidebar.Resize(sb.X, sb.Y, sb.Width, sb.Height)

	ed := l.GetEditorArea(screenWidth, screenHeight)
	sm.editor.Resize(ed.X, ed.Y, ed.Width, ed.Height)

	st := l.GetStatusBarArea(screenWidth, screenHeight)
	sm.statusBar.Resize(st.X, st.Y, st.Width, st.Height)

	// Redraw components
	sm.topBar.Draw(screen)
	sm.sidebar.Draw(screen)
	sm.editor.Draw(screen)
	sm.statusBar.Draw(screen)

	// Draw dialog on top
	if sm.dialog != nil {
		sm.dialog.Center(screen)
		sm.dialog.Draw(screen)
	}

	sm.screen.Show()
}
