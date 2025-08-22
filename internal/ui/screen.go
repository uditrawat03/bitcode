package ui

import (
	"github.com/gdamore/tcell/v2"
	"github.com/uditrawat03/bitcode/internal/buffer"
	"github.com/uditrawat03/bitcode/internal/editor"
	"github.com/uditrawat03/bitcode/internal/layout"
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

	editor    *editor.Editor
	sidebar   *sidebar.Sidebar
	topBar    *topbar.TopBar
	statusBar *statusbar.StatusBar

	focusOrder []Focusable
	focusedIdx int
}

func CreateScreenManager() *ScreenManager {
	sm := &ScreenManager{
		layoutManager: layout.CreateLayoutManager(),
		bufferManager: buffer.NewBufferManager(),
	}
	return sm
}

// Initialize components and focus order
func (sm *ScreenManager) InitComponents(screenWidth, screenHeight int) {
	// Update layout
	sm.layoutManager.UpdateLayout(screenWidth, screenHeight)
	l := sm.layoutManager.GetLayout()

	// TopBar
	tbX, tbY, tbW, tbH := l.GetTopBarArea(screenWidth, screenHeight)
	sm.topBar = topbar.CreateTopBar(tbX, tbY, tbW, tbH)

	// Sidebar
	sbX, sbY, sbW, sbH := l.GetSidebarArea(screenWidth, screenHeight)
	sm.sidebar = sidebar.CreateSidebar(sbX, sbY, sbW, sbH)
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
	edX, edY, edW, edH := l.GetEditorArea(screenWidth, screenHeight)
	sm.editor = editor.CreateEditor(edX, edY, edW, edH)
	sm.editor.SetFocusCallback(func() {
		sm.focusOrder[sm.focusedIdx].Blur()
		sm.focusedIdx = 1 // editor index in focusOrder
		sm.focusOrder[sm.focusedIdx].Focus()
	})

	// StatusBar
	stX, stY, stW, stH := l.GetStatusBarArea(screenWidth, screenHeight)
	sm.statusBar = statusbar.CreateStatusBar(stX, stY, stW, stH)

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
	sm.focusOrder[sm.focusedIdx].Blur()
	sm.focusedIdx = (sm.focusedIdx + 1) % len(sm.focusOrder)
	sm.focusOrder[sm.focusedIdx].Focus()
}

func (sm *ScreenManager) HandleMouse(ev *tcell.EventMouse) {
	// only delegate
	for _, comp := range sm.focusOrder {
		comp.HandleMouse(ev)
	}
}

// screen_manager.go
func (sm *ScreenManager) HandleKey(ev *tcell.EventKey) {
	if len(sm.focusOrder) == 0 {
		return
	}
	sm.focusOrder[sm.focusedIdx].HandleKey(ev)
}

// Draw all components
func (sm *ScreenManager) Draw(screen tcell.Screen) {
	screenWidth, screenHeight := screen.Size()
	sm.layoutManager.UpdateLayout(screenWidth, screenHeight)

	// Redraw components
	sm.topBar.Draw(screen)
	sm.sidebar.Draw(screen)
	sm.editor.Draw(screen)
	sm.statusBar.Draw(screen)

	screen.Show()
}
