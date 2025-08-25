package ui

import (
	"context"
	"log"
	"path/filepath"

	"github.com/gdamore/tcell/v2"
	"github.com/uditrawat03/bitcode/internal/buffer"
	"github.com/uditrawat03/bitcode/internal/core"
	"github.com/uditrawat03/bitcode/internal/dialog"
	"github.com/uditrawat03/bitcode/internal/editor"
	"github.com/uditrawat03/bitcode/internal/layout"
	lsp "github.com/uditrawat03/bitcode/internal/lsp_client"
	"github.com/uditrawat03/bitcode/internal/sidebar"
	"github.com/uditrawat03/bitcode/internal/statusbar"
	"github.com/uditrawat03/bitcode/internal/topbar"
	"github.com/uditrawat03/bitcode/internal/treeview"
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
	UIManager *core.UIManager

	focusOrder   []core.UIComponent
	focusedIdx   int
	selectedNode *treeview.TreeNode
	refreshTree  func(node *treeview.TreeNode)

	running       bool
	isDialogOpen  bool
	isTooltipOpen bool
}

func NewScreenManager(ctx context.Context, screen tcell.Screen, logger *log.Logger, lsp *lsp.Client) *ScreenManager {
	lm := layout.CreateLayoutManager()
	uiManager := core.NewUIManager(logger, lm)

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

func (sm *ScreenManager) RequestRender() {
	sm.Render()
}

func (sm *ScreenManager) SetupComponents(rootPath string) {
	sw, sh := sm.screen.Size()
	sm.Lm.UpdateLayout(sw, sh)
	l := sm.Lm.GetLayout()

	// TopBar
	topBar := topbar.NewTopBar("Bitcode IDE")
	rect := l.GetTopBarArea(sw, sh)
	topBar.SetPosition(rect.X, rect.Y)
	topBar.Resize(rect.Width, rect.Height)
	sm.AddComponent(topBar, false)

	// StatusBar
	statusBar := statusbar.NewStatusBar("Ready")
	rect = l.GetStatusBarArea(sw, sh)
	statusBar.SetPosition(rect.X, rect.Y)
	statusBar.Resize(rect.Width, rect.Height)
	sm.AddComponent(statusBar, false)

	// Remove TreeView creation
	sb := sidebar.NewSidebar(rootPath, l.SidebarWidth, sm.RequestRender)

	rect = l.GetSidebarArea(sw, sh)
	sb.SetPosition(rect.X, rect.Y)
	sb.Resize(rect.Width, rect.Height)
	sm.AddComponent(sb, true)

	// Editor
	ed := editor.NewEditor()
	rect = l.GetEditorArea(sw, sh)
	ed.SetPosition(rect.X, rect.Y)
	ed.Resize(rect.Width, rect.Height)
	sm.AddComponent(ed, true)

	// Sidebar callbacks
	sb.SetOnFileOpen(func(path string) {
		buf := sm.Bm.Open(path)
		sm.OpenBufferInEditor(buf)
	})
	sb.SetOnNodeSelect(func(node *treeview.TreeNode) {
		sm.selectedNode = node
		sm.logger.Println("[SetOnNodeSelect]", node)
	})
	sm.SetRefreshTreeView(sb.OnReloadTreeViewChildren)

	// Focus
	sm.focusOrder = []core.UIComponent{sb, ed}
	sm.focusedIdx = 0
	sm.focusCurrent()
}

func (sm *ScreenManager) AddComponent(c core.UIComponent, focusable bool) {
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
	// Ctrl+N opens new file dialog
	if ev.Key() == tcell.KeyCtrlN && !sm.isDialogOpen {
		sm.openNewFileDialog()
		return
	}

	if sm.isDialogOpen || sm.isTooltipOpen {
		sm.focusCurrentKey(ev)
		sm.RequestRender()
		return
	}

	if ev.Key() == tcell.KeyEscape {
		sm.running = false
		return
	}

	sm.focusCurrentKey(ev)
	sm.RequestRender()
}

func (sm *ScreenManager) handleMouse(ev *tcell.EventMouse) {
	x, _ := ev.Position()

	if sm.isDialogOpen || sm.isTooltipOpen {
		sm.focusCurrentMouse(ev)
		sm.RequestRender()
		return
	}

	for idx, c := range sm.focusOrder {
		f, ok := c.(Focusable)
		if !ok {
			continue
		}
		r := c.GetRect()
		if x >= r.X && x < r.X+r.Width {
			sm.setFocus(idx)
			f.HandleMouse(ev)
			sm.RequestRender()
			return
		}
	}

	sm.focusCurrentMouse(ev)
	sm.RequestRender()
}

func (sm *ScreenManager) focusCurrent() {
	if len(sm.focusOrder) == 0 {
		return
	}
	if f, ok := sm.focusOrder[sm.focusedIdx].(Focusable); ok {
		f.Focus()
	}
}

func (sm *ScreenManager) focusCurrentKey(ev *tcell.EventKey) {
	if len(sm.focusOrder) == 0 {
		return
	}
	if f, ok := sm.focusOrder[sm.focusedIdx].(Focusable); ok {
		f.HandleKey(ev)
	}
}

func (sm *ScreenManager) focusCurrentMouse(ev *tcell.EventMouse) {
	if len(sm.focusOrder) == 0 {
		return
	}
	if f, ok := sm.focusOrder[sm.focusedIdx].(Focusable); ok {
		f.HandleMouse(ev)
	}
}

func (sm *ScreenManager) setFocus(idx int) {
	if idx < 0 || idx >= len(sm.focusOrder) {
		return
	}
	if f, ok := sm.focusOrder[sm.focusedIdx].(Focusable); ok {
		f.Blur()
	}
	sm.focusedIdx = idx
	sm.focusCurrent()
}

func (sm *ScreenManager) OpenBufferInEditor(buf *buffer.Buffer) {
	for _, c := range sm.focusOrder {
		if ed, ok := c.(*editor.Editor); ok {
			ed.SetBuffer(buf)
			return
		}
	}
}

func (sm *ScreenManager) RemoveComponent(c core.UIComponent) {
	sm.UIManager.RemoveComponent(c)

	for i, fc := range sm.focusOrder {
		if fc == c {
			sm.focusOrder = append(sm.focusOrder[:i], sm.focusOrder[i+1:]...)
			if sm.focusedIdx >= len(sm.focusOrder) {
				sm.focusedIdx = len(sm.focusOrder) - 1
			}
			break
		}
	}
	sm.focusCurrent()
}

func (sm *ScreenManager) Render() {
	sm.UIManager.RenderAll(sm.screen)
	sm.screen.Show()
}

func (sm *ScreenManager) getBufferParentFolder() string {
	if sm.selectedNode == nil || sm.selectedNode.FullPath == "" {
		return "."
	}
	if sm.selectedNode.IsDir {
		return sm.selectedNode.FullPath
	}
	return filepath.Dir(sm.selectedNode.FullPath)
}

func (sm *ScreenManager) SetRefreshTreeView(cb func(node *treeview.TreeNode)) {
	sm.refreshTree = cb
}

// Dialog helper
func (sm *ScreenManager) closeDialog(dlg *dialog.InputDialog) {
	sm.RemoveComponent(dlg)
	sm.RequestRender()
	sm.isDialogOpen = false
}
