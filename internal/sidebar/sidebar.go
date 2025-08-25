package sidebar

import (
	"github.com/gdamore/tcell/v2"
	"github.com/uditrawat03/bitcode/internal/core"
	"github.com/uditrawat03/bitcode/internal/layout"
	"github.com/uditrawat03/bitcode/internal/treeview"
)

type Sidebar struct {
	core.BaseComponent
	Tree       *treeview.TreeView
	Width      int
	Resizing   bool
	OnChange   func()
	onFileOpen func(path string)

	onNodeSelect func(node *treeview.TreeNode)
}

func (s *Sidebar) SetOnFileOpen(cb func(node string))               { s.onFileOpen = cb }
func (s *Sidebar) SetOnNodeSelect(cb func(node *treeview.TreeNode)) { s.onNodeSelect = cb }

func (s *Sidebar) OnReloadTreeViewChildren(node *treeview.TreeNode) {
	s.Tree.LoadChildren(node)
	if s.OnChange != nil {
		s.OnChange()
	}
}

func NewSidebar(rootPath string, initialWidth int, onChange func()) *Sidebar {
	tv := treeview.NewTreeView(rootPath)
	tv.OnChange = onChange

	sb := &Sidebar{
		Tree:     tv,
		Width:    initialWidth,
		Resizing: false,
		OnChange: onChange,
	}

	// Wrap TreeView callbacks
	tv.SetOnNodeSelect(func(path string) {
		if sb.onFileOpen != nil {
			sb.onFileOpen(path)
		}
		if sb.OnChange != nil {
			sb.OnChange()
		}
	})
	tv.SetOnNodeItemSelect(func(node *treeview.TreeNode) {
		if sb.onNodeSelect != nil {
			sb.onNodeSelect(node)
		}
		if sb.OnChange != nil {
			sb.OnChange()
		}
	})

	return sb
}

func (s *Sidebar) ReloadNode(node *treeview.TreeNode) {
	s.Tree.LoadChildren(node)
	s.Tree.FlattenVisible()
	if s.OnChange != nil {
		s.OnChange()
	}
}

func (s *Sidebar) SetPosition(x, y int)     { s.Rect.X = x; s.Rect.Y = y }
func (s *Sidebar) Resize(width, height int) { s.Rect.Width = width; s.Rect.Height = height }

func (s *Sidebar) Render(screen tcell.Screen, lm *layout.LayoutManager) {
	bg := tcell.NewRGBColor(30, 30, 30)
	style := tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(bg)
	_, screenHeight := screen.Size()
	layout := lm.GetLayout()

	topBarHeight := layout.TopBarHeight
	statusBarHeight := layout.StatusBarHeight

	// Position sidebar
	x := 0
	y := topBarHeight
	s.Rect.X = x
	s.Rect.Y = y
	s.Rect.Width = s.Width
	s.Rect.Height = screenHeight - topBarHeight - statusBarHeight

	// Draw right border
	for row := y; row < y+s.Rect.Height; row++ {
		screen.SetContent(x+s.Width, row, '│', nil, style)
	}

	// Render TreeView
	s.Tree.SetLogger(s.Logger)
	s.Tree.SetPosition(x, y)
	s.Tree.Resize(s.Width, s.Rect.Height)
	s.Tree.Render(screen, lm)
}

// Focusable
func (s *Sidebar) Focus()                       { s.Tree.Focus() }
func (s *Sidebar) Blur()                        { s.Tree.Blur() }
func (s *Sidebar) IsFocused() bool              { return s.Tree.IsFocused() }
func (s *Sidebar) HandleKey(ev *tcell.EventKey) { s.Tree.HandleKey(ev) }

func (s *Sidebar) HandleMouse(ev *tcell.EventMouse) {
	x, _ := ev.Position()

	// Drag start on right border
	if ev.Buttons()&tcell.Button1 != 0 && x == s.Rect.X+s.Width {
		s.Resizing = true
		return
	}
	if ev.Buttons() == 0 {
		s.Resizing = false
		return
	}
	if s.Resizing {
		newWidth := x - s.Rect.X
		if newWidth < 10 {
			newWidth = 10
		}
		s.Width = newWidth
		if s.OnChange != nil {
			s.OnChange()
		}
		return
	}

	// Pass mouse wheel and clicks to tree
	s.Tree.HandleMouse(ev)
}
