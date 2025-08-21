package sidebar

import (
	"os"
	"path/filepath"

	"github.com/gdamore/tcell/v2"
)

type Node struct {
	Name     string
	Path     string
	IsDir    bool
	Children []*Node
	Expanded bool
	Parent   *Node
}

type Sidebar struct {
	X, Y, Width, Height int
	Focused             bool
	root                *Node
	nodes               []*Node
	nodeLevels          map[*Node]int
	ScrollY             int
	Selected            int
	Hovered             int

	// resizing state
	resizing bool

	onFileOpen func(path string)
}

func (sb *Sidebar) SetOnFileOpen(cb func(path string)) {
	sb.onFileOpen = cb
}

// CreateSidebar initializes sidebar with current folder
func CreateSidebar(x, y, width, height int) *Sidebar {
	sb := &Sidebar{
		X:          x,
		Y:          y,
		Width:      width,
		Height:     height,
		nodeLevels: map[*Node]int{},
	}
	cwd, err := os.Getwd()
	if err != nil {
		cwd = "."
	}
	sb.LoadDirectory(cwd)
	return sb
}

// LoadDirectory loads folder recursively
func (sb *Sidebar) LoadDirectory(rootPath string) {
	sb.root = &Node{
		Name:     filepath.Base(rootPath),
		Path:     rootPath,
		IsDir:    true,
		Expanded: true,
	}
	sb.buildTree(sb.root)
	sb.flattenVisible()
}

// Recursively build tree
func (sb *Sidebar) buildTree(node *Node) {
	entries, err := os.ReadDir(node.Path)
	if err != nil {
		return
	}
	for _, entry := range entries {
		child := &Node{
			Name:   entry.Name(),
			Path:   filepath.Join(node.Path, entry.Name()),
			IsDir:  entry.IsDir(),
			Parent: node,
		}
		node.Children = append(node.Children, child)
		if entry.IsDir() {
			child.Expanded = false
			sb.buildTree(child)
		}
	}
}

// Flatten tree to visible nodes
func (sb *Sidebar) flattenVisible() {
	sb.nodes = []*Node{}
	sb.nodeLevels = map[*Node]int{}
	sb.flattenNode(sb.root, 0)
}

func (sb *Sidebar) flattenNode(node *Node, level int) {
	sb.nodes = append(sb.nodes, node)
	sb.nodeLevels[node] = level
	if node.IsDir && node.Expanded {
		for _, child := range node.Children {
			sb.flattenNode(child, level+1)
		}
	}
}

// Draw sidebar with right border and scrollbar
func (sb *Sidebar) Draw(s tcell.Screen) {
	bg := tcell.NewRGBColor(30, 30, 30)
	style := tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(bg)
	selectedStyle := tcell.StyleDefault.Foreground(tcell.ColorBlack).Background(tcell.NewRGBColor(100, 100, 255))
	hoverStyle := tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.NewRGBColor(60, 60, 60))

	// Fill background
	for row := 0; row < sb.Height; row++ {
		for col := 0; col < sb.Width; col++ {
			ch := ' '
			// right border
			if col == sb.Width-1 {
				ch = '│'
			}
			s.SetContent(sb.X+col, sb.Y+row, ch, nil, style)
		}
	}

	// Draw visible nodes
	for i := 0; i < sb.Height; i++ {
		idx := i + sb.ScrollY
		if idx >= len(sb.nodes) {
			break
		}
		node := sb.nodes[idx]
		level := sb.nodeLevels[node]

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

		// Truncate safely (keep border free)
		runes := []rune(line)
		if len(runes) > sb.Width-1 {
			runes = runes[:sb.Width-2]
			runes = append(runes, '…')
		}

		// Choose style
		currentStyle := style
		if idx == sb.Selected {
			currentStyle = selectedStyle
		} else if idx == sb.Hovered {
			currentStyle = hoverStyle
		}

		// Draw text (exclude last col = border)
		for col := 0; col < sb.Width-1; col++ {
			ch := ' '
			if col < len(runes) {
				ch = runes[col]
			}
			s.SetContent(sb.X+col, sb.Y+i, ch, nil, currentStyle)
		}
	}

	// Draw scrollbar if needed
	if len(sb.nodes) > sb.Height {
		scrollbarHeight := sb.Height * sb.Height / len(sb.nodes)
		if scrollbarHeight < 1 {
			scrollbarHeight = 1
		}
		scrollbarY := sb.ScrollY * sb.Height / len(sb.nodes)

		for i := 0; i < scrollbarHeight && scrollbarY+i < sb.Height; i++ {
			s.SetContent(sb.X+sb.Width-1, sb.Y+scrollbarY+i, '█', nil,
				tcell.StyleDefault.Foreground(tcell.ColorGray))
		}
	}
}

// Scroll
func (sb *Sidebar) Scroll(dy int) {
	sb.Selected += dy
	if sb.Selected < 0 {
		sb.Selected = 0
	}
	if sb.Selected >= len(sb.nodes) {
		sb.Selected = len(sb.nodes) - 1
	}

	if sb.Selected < sb.ScrollY {
		sb.ScrollY = sb.Selected
	} else if sb.Selected >= sb.ScrollY+sb.Height {
		sb.ScrollY = sb.Selected - sb.Height + 1
	}
}

// Toggle directory expand/collapse
func (sb *Sidebar) Toggle() {
	node := sb.nodes[sb.Selected]
	if node.IsDir {
		node.Expanded = !node.Expanded
		sb.flattenVisible()
	}
}

// Focusable
func (sb *Sidebar) Focus()          { sb.Focused = true }
func (sb *Sidebar) Blur()           { sb.Focused = false }
func (sb *Sidebar) IsFocused() bool { return sb.Focused }

// Handle key events
func (sb *Sidebar) HandleKey(ev *tcell.EventKey) {
	switch ev.Key() {
	case tcell.KeyUp:
		sb.Scroll(-1)
	case tcell.KeyDown:
		sb.Scroll(1)
	case tcell.KeyEnter, tcell.KeyRight:
		node := sb.nodes[sb.Selected]
		if node.IsDir && !node.Expanded {
			node.Expanded = true
			sb.flattenVisible()
		} else {
			if sb.onFileOpen != nil {
				sb.onFileOpen(node.Path)
			}
		}
	case tcell.KeyLeft:
		node := sb.nodes[sb.Selected]
		if node.IsDir && node.Expanded {
			node.Expanded = false
			sb.flattenVisible()
		} else if node.Parent != nil {
			for i, n := range sb.nodes {
				if n == node.Parent {
					sb.Selected = i
					sb.Scroll(0)
					break
				}
			}
		}
	}
}

// Expose node count safely
func (s *Sidebar) NodeCount() int {
	return len(s.nodes)
}

// Expose "is directory" check
func (s *Sidebar) IsDir(index int) bool {
	if index >= 0 && index < len(s.nodes) {
		return s.nodes[index].IsDir
	}
	return false
}

// Toggle expand/collapse at given index
func (s *Sidebar) ToggleAt(index int) {
	if index >= 0 && index < len(s.nodes) {
		s.nodes[index].Expanded = !s.nodes[index].Expanded
	}
}

// Handle mouse events (click, scroll, resize)
func (sb *Sidebar) HandleMouse(ev *tcell.EventMouse) {
	x, y := ev.Position()

	// resizing mode
	if sb.resizing {
		if ev.Buttons() == 0 {
			sb.resizing = false
		} else {
			newW := x - sb.X + 1
			if newW > 10 {
				sb.Width = newW
			}
		}
		return
	}

	// drag start if on right border
	if ev.Buttons()&tcell.Button1 != 0 && x == sb.X+sb.Width-1 {
		sb.resizing = true
		return
	}

	// Ignore clicks outside sidebar
	if x < sb.X || x >= sb.X+sb.Width || y < sb.Y || y >= sb.Y+sb.Height {
		return
	}

	// Scroll wheel
	switch ev.Buttons() {
	case tcell.WheelUp:
		if sb.ScrollY > 0 {
			sb.ScrollY--
		}
		return
	case tcell.WheelDown:
		if sb.ScrollY < sb.maxScroll() {
			sb.ScrollY++
		}
		return
	}

	// map mouse position to node index
	idx := y - sb.Y + sb.ScrollY
	if idx < 0 || idx >= len(sb.nodes) {
		return
	}

	sb.Hovered = idx

	if ev.Buttons()&tcell.Button1 != 0 {
		sb.Selected = idx
		node := sb.nodes[idx]
		if node.IsDir {
			node.Expanded = !node.Expanded
			sb.flattenVisible()
		} else if sb.onFileOpen != nil {
			sb.onFileOpen(node.Path)
		}
	}
}

func (sb *Sidebar) maxScroll() int {
	if len(sb.nodes) > sb.Height {
		return len(sb.nodes) - sb.Height
	}
	return 0
}
