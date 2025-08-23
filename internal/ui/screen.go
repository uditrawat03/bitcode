package ui

import (
	"context"
	"log"

	"github.com/gdamore/tcell/v2"
	"github.com/uditrawat03/bitcode/internal/buffer"
	"github.com/uditrawat03/bitcode/internal/dialog"

	"github.com/uditrawat03/bitcode/internal/editor"
	"github.com/uditrawat03/bitcode/internal/layout"
	lsp "github.com/uditrawat03/bitcode/internal/lsp_client"
	"github.com/uditrawat03/bitcode/internal/sidebar"
	"github.com/uditrawat03/bitcode/internal/statusbar"
	"github.com/uditrawat03/bitcode/internal/tooltip"
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
	ctx           context.Context
	logger        *log.Logger
	layoutManager *layout.LayoutManager
	bufferManager *buffer.BufferManager
	screen        tcell.Screen

	editor    *editor.Editor
	sidebar   *sidebar.Sidebar
	topBar    *topbar.TopBar
	statusBar *statusbar.StatusBar

	dialog  *dialog.Dialog
	tooltip tooltip.Tooltip

	focusOrder []Focusable
	focusedIdx int
	rootPath   string
}

func CreateScreenManager(ctx context.Context, logger *log.Logger, lsp *lsp.Client, rootPath string) *ScreenManager {
	sm := &ScreenManager{
		ctx:           ctx,
		layoutManager: layout.CreateLayoutManager(),
		bufferManager: buffer.NewBufferManager(ctx, lsp),
		logger:        logger,
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
	sm.topBar = topbar.CreateTopBar(sm.ctx, tb.X, tb.Y, tb.Width, tb.Height)

	// Sidebar
	sb := l.GetSidebarArea(screenWidth, screenHeight)
	sm.sidebar = sidebar.CreateSidebar(sm.ctx, sb.X, sb.Y, sb.Width, sb.Height, sm.rootPath)
	sm.sidebar.SetOnFileOpen(func(path string) {
		buf := sm.bufferManager.Open(path)
		sm.editor.SetBuffer(buf)
		// sm.restoreEditorFocus()
	})

	sm.sidebar.SetFocusCallback(func() {
		sm.focusOrder[sm.focusedIdx].Blur()
		sm.focusedIdx = 0 // sidebar index
		sm.focusOrder[sm.focusedIdx].Focus()
	})

	// Editor
	ed := l.GetEditorArea(screenWidth, screenHeight)
	sm.editor = editor.CreateEditor(sm.ctx, sm.logger, ed.X, ed.Y, ed.Width, ed.Height)
	sm.editor.SetFocusCallback(func() {
		sm.focusOrder[sm.focusedIdx].Blur()
		sm.focusedIdx = 1 // editor index in focusOrder
		sm.focusOrder[sm.focusedIdx].Focus()
	})

	sm.editor.SetTooltipHandler(func(x, y int, content string, list []string) {
		sm.ShowTooltip(x, y, content, list)
	})

	// StatusBar
	st := l.GetStatusBarArea(screenWidth, screenHeight)
	sm.statusBar = statusbar.CreateStatusBar(sm.ctx, st.X, st.Y, st.Width, st.Height)

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

	// if sm.tooltip.Visible {
	// 	style := tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.ColorBlack)

	// 	switch sm.tooltip.Type {
	// 	case tooltip.TooltipText:
	// 		for i, r := range sm.tooltip.Content {
	// 			screen.SetContent(sm.tooltip.X+i, sm.tooltip.Y, r, nil, style)
	// 		}
	// 	case tooltip.TooltipList:
	// 		for idx, item := range sm.tooltip.Items {
	// 			y := sm.tooltip.Y + idx
	// 			itemStyle := style
	// 			if idx == sm.tooltip.Selected {
	// 				itemStyle = tcell.StyleDefault.Foreground(tcell.ColorBlack).Background(tcell.ColorWhite)
	// 			}
	// 			for i, r := range item {
	// 				screen.SetContent(sm.tooltip.X+i, y, r, nil, itemStyle)
	// 			}
	// 		}
	// 	}
	// }

	if sm.tooltip.Visible {
		sm.tooltip.Draw(screen)
	}

	sm.screen.Show()
}

func (sm *ScreenManager) IsTooltipVisible() bool {
	return sm.tooltip.Visible
}

func (sm *ScreenManager) IsDialogOpen() bool {
	return sm.dialog != nil
}
