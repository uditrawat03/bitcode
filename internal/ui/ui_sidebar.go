package ui

import (
	"github.com/gdamore/tcell/v2"
	"github.com/uditrawat03/bitcode/internal/layout"
)

type Sidebar struct {
	BaseComponent
	Tree       *TreeView
	Width      int
	Resizing   bool
	OnChange   func()
	onFileOpen func(path string)
}

func (s *Sidebar) SetOnFileOpen(cb func(path string)) {
	s.onFileOpen = cb
}

func NewSidebar(tree *TreeView, initialWidth int, onChange func()) *Sidebar {
	tree.OnChange = onChange
	return &Sidebar{
		Tree:     tree,
		Width:    initialWidth,
		Resizing: false,
		OnChange: onChange,
	}
}

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
	s.Tree.SetOnNodeSelect(s.onFileOpen)

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
