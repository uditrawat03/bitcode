package sidebar

import (
	"context"

	"github.com/gdamore/tcell/v2"
	"github.com/uditrawat03/bitcode/internal/treeview"
)

type Resizable interface {
	Resize(x, y, w, h int)
}

type Sidebar struct {
	ctx                 context.Context
	X, Y, Width, Height int
	Focused             bool
	ScrollY             int
	Selected            int
	Hovered             int
	resizing            bool

	Tree *treeview.TreeView

	onFileOpen func(path string)
	focusCb    func()
}

func CreateSidebar(ctx context.Context, x, y, width, height int, cwd string) *Sidebar {
	return NewSidebar(ctx, x, y, width, height, cwd)
}

func (sb *Sidebar) Resize(x, y, w, h int) {
	sb.X, sb.Y, sb.Width, sb.Height = x, y, w, h
}

// Create a new sidebar
func NewSidebar(ctx context.Context, x, y, width, height int, rootPath string) *Sidebar {
	return &Sidebar{
		ctx:    ctx,
		X:      x,
		Y:      y,
		Width:  width,
		Height: height,
		Tree:   treeview.NewTreeView(rootPath),
	}
}

func (sb *Sidebar) SetFocusCallback(cb func()) { sb.focusCb = cb }
func (sb *Sidebar) SetOnFileOpen(cb func(path string)) {
	sb.onFileOpen = cb
}

// Focusable
func (sb *Sidebar) Focus()          { sb.Focused = true }
func (sb *Sidebar) Blur()           { sb.Focused = false }
func (sb *Sidebar) IsFocused() bool { return sb.Focused }

// Draw the sidebar
func (sb *Sidebar) Draw(s tcell.Screen) {
	bg := tcell.NewRGBColor(30, 30, 30)
	style := tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(bg)
	selectedStyle := tcell.StyleDefault.Foreground(tcell.ColorBlack).Background(tcell.NewRGBColor(100, 100, 255))
	hoverStyle := tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.NewRGBColor(60, 60, 60))

	// Fill background with right border
	for row := 0; row < sb.Height; row++ {
		for col := 0; col < sb.Width; col++ {
			ch := ' '
			if col == sb.Width-1 {
				ch = '│'
			}
			s.SetContent(sb.X+col, sb.Y+row, ch, nil, style)
		}
	}

	// Draw nodes
	for i := 0; i < sb.Height; i++ {
		idx := i + sb.ScrollY
		if idx >= len(sb.Tree.Nodes) {
			break
		}
		node := sb.Tree.Nodes[idx]
		level := sb.Tree.Level(node)

		line := ""
		for j := 0; j < level; j++ {
			line += "  "
		}
		if node.IsDir {
			if node.Expanded {
				line += "▼ "
			} else {
				line += "▶ "
			}
		} else {
			line += "  "
		}
		line += node.Name

		runes := []rune(line)
		if len(runes) > sb.Width-1 {
			runes = runes[:sb.Width-2]
			runes = append(runes, '…')
		}

		currentStyle := style
		if idx == sb.Selected {
			currentStyle = selectedStyle
		} else if idx == sb.Hovered {
			currentStyle = hoverStyle
		}

		for col := 0; col < sb.Width-1; col++ {
			ch := ' '
			if col < len(runes) {
				ch = runes[col]
			}
			s.SetContent(sb.X+col, sb.Y+i, ch, nil, currentStyle)
		}
	}

	// Draw scrollbar
	if len(sb.Tree.Nodes) > sb.Height {
		scrollbarHeight := sb.Height * sb.Height / len(sb.Tree.Nodes)
		if scrollbarHeight < 1 {
			scrollbarHeight = 1
		}
		scrollbarY := sb.ScrollY * sb.Height / len(sb.Tree.Nodes)

		for i := 0; i < scrollbarHeight && scrollbarY+i < sb.Height; i++ {
			s.SetContent(sb.X+sb.Width-1, sb.Y+scrollbarY+i, '█', nil,
				tcell.StyleDefault.Foreground(tcell.ColorGray))
		}
	}
}

// Get selected node
func (sb *Sidebar) GetSelectedNode() *treeview.Node {
	if sb.Selected >= 0 && sb.Selected < len(sb.Tree.Nodes) {
		return sb.Tree.Nodes[sb.Selected]
	}
	return nil
}
