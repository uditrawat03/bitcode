package ui

import (
	"os"
	"path/filepath"

	"github.com/gdamore/tcell/v2"
	"github.com/uditrawat03/bitcode/internal/layout"
)

// TreeNode represents a file/folder
type TreeNode struct {
	Name     string
	FullPath string
	Children []*TreeNode
	Expanded bool
	Parent   *TreeNode
	IsDir    bool
}

// TreeView holds flattened nodes
type TreeView struct {
	BaseComponent
	Root   *TreeNode
	Nodes  []*TreeNode
	Levels map[*TreeNode]int

	CursorIdx    int
	ScrollOffset int
	focused      bool
	OnChange     func()
	onNodeSelect func(path string)
}

func (tv *TreeView) SetOnNodeSelect(cb func(path string)) {
	tv.onNodeSelect = cb
}

// --- TreeView constructors ---
func NewTreeView(rootPath string) *TreeView {
	root := &TreeNode{
		Name:     filepath.Base(rootPath),
		FullPath: rootPath,
		IsDir:    true,
		Expanded: true,
	}
	tv := &TreeView{
		Root:   root,
		Levels: make(map[*TreeNode]int),
	}
	tv.loadChildren(root)
	tv.flattenVisible()
	return tv
}

// Lazy load children
func (tv *TreeView) loadChildren(node *TreeNode) {
	if !node.IsDir || node.Children != nil {
		return
	}
	entries, err := os.ReadDir(node.FullPath)
	if err != nil {
		node.Children = nil
		return
	}
	for _, e := range entries {
		child := &TreeNode{
			Name:     e.Name(),
			FullPath: filepath.Join(node.FullPath, e.Name()),
			Parent:   node,
			IsDir:    e.IsDir(),
		}
		node.Children = append(node.Children, child)
	}
}

// Flatten visible nodes
func (tv *TreeView) flattenVisible() {
	tv.Nodes = []*TreeNode{}
	tv.Levels = make(map[*TreeNode]int)
	var walk func(node *TreeNode, level int)
	walk = func(node *TreeNode, level int) {
		tv.Nodes = append(tv.Nodes, node)
		tv.Levels[node] = level
		if node.IsDir && node.Expanded {
			for _, child := range node.Children {
				walk(child, level+1)
			}
		}
	}
	walk(tv.Root, 0)
}

// Toggle expand/collapse
func (tv *TreeView) ToggleNode(node *TreeNode) {
	if node == nil || !node.IsDir {
		return
	}
	if !node.Expanded {
		tv.loadChildren(node)
	}
	node.Expanded = !node.Expanded
	tv.flattenVisible()
}

// --- Scroll helper ---
func (tv *TreeView) Scroll(delta int) {
	tv.CursorIdx += delta
	if tv.CursorIdx < 0 {
		tv.CursorIdx = 0
	} else if tv.CursorIdx >= len(tv.Nodes) {
		tv.CursorIdx = len(tv.Nodes) - 1
	}
	tv.ensureCursorVisible()
	if tv.OnChange != nil {
		tv.OnChange()
	}
}

func (tv *TreeView) ensureCursorVisible() {
	visibleHeight := tv.Rect.Height
	if tv.CursorIdx < tv.ScrollOffset {
		tv.ScrollOffset = tv.CursorIdx
	} else if tv.CursorIdx >= tv.ScrollOffset+visibleHeight {
		tv.ScrollOffset = tv.CursorIdx - visibleHeight + 1
	}
	if tv.ScrollOffset < 0 {
		tv.ScrollOffset = 0
	}
}

func (tv *TreeView) PageUp() {
	tv.Scroll(-tv.Rect.Height)
}

func (tv *TreeView) PageDown() {
	tv.Scroll(tv.Rect.Height)
}

func (tv *TreeView) Home() {
	tv.CursorIdx = 0
	tv.ScrollOffset = 0
	if tv.OnChange != nil {
		tv.OnChange()
	}
}

func (tv *TreeView) End() {
	tv.CursorIdx = len(tv.Nodes) - 1
	tv.ScrollOffset = len(tv.Nodes) - tv.Rect.Height
	if tv.ScrollOffset < 0 {
		tv.ScrollOffset = 0
	}
	if tv.OnChange != nil {
		tv.OnChange()
	}
}

func (tv *TreeView) CurrentNode() *TreeNode {
	if tv.CursorIdx >= 0 && tv.CursorIdx < len(tv.Nodes) {
		return tv.Nodes[tv.CursorIdx]
	}
	return nil
}

// --- Focusable interface ---
func (tv *TreeView) Focus()                       { tv.focused = true }
func (tv *TreeView) Blur()                        { tv.focused = false }
func (tv *TreeView) IsFocused() bool              { return tv.focused }
func (tv *TreeView) HandleKey(ev *tcell.EventKey) { tv.handleKey(ev) }

func (tv *TreeView) handleKey(ev *tcell.EventKey) {
	node := tv.CurrentNode()
	if node == nil {
		return
	}

	switch ev.Key() {
	case tcell.KeyUp:
		tv.Scroll(-1)
	case tcell.KeyDown:
		tv.Scroll(1)
	case tcell.KeyPgUp:
		tv.PageUp()
	case tcell.KeyPgDn:
		tv.PageDown()
	case tcell.KeyHome:
		tv.Home()
	case tcell.KeyEnd:
		tv.End()
	case tcell.KeyEnter:
		if node.IsDir {
			tv.Logger.Println("[TreeView][Enter]", node)
			tv.ToggleNode(node)
		} else {
			tv.onNodeSelect(node.FullPath)
		}
	case tcell.KeyRight:
		if node.IsDir && !node.Expanded {
			tv.ToggleNode(node)
		}
	case tcell.KeyLeft:
		if node.IsDir && node.Expanded {
			tv.ToggleNode(node)
		} else if node.Parent != nil {
			for i, n := range tv.Nodes {
				if n == node.Parent {
					tv.CursorIdx = i
					tv.ensureCursorVisible()
					break
				}
			}
		}
	}

	if tv.OnChange != nil {
		tv.OnChange()
	}
}

// --- Mouse ---
func (tv *TreeView) HandleMouse(ev *tcell.EventMouse) {
	x, y := ev.Position()
	rect := tv.Rect
	if y < rect.Y || y >= rect.Y+rect.Height || x < rect.X || x >= rect.X+rect.Width {
		return
	}

	switch ev.Buttons() {
	case tcell.WheelUp:
		tv.Scroll(-1)
		return
	case tcell.WheelDown:
		tv.Scroll(1)
		return
	}

	// Left click
	if ev.Buttons()&tcell.Button1 == 0 {
		return
	}

	idx := y - rect.Y + tv.ScrollOffset
	if idx < 0 || idx >= len(tv.Nodes) {
		return
	}
	tv.CursorIdx = idx
	node := tv.CurrentNode()
	if node != nil && node.IsDir {
		tv.ToggleNode(node)
	} else if node != nil {
		tv.onNodeSelect(node.FullPath)
	}
	if tv.OnChange != nil {
		tv.OnChange()
	}
}

// Render flattened nodes
func (tv *TreeView) Render(screen tcell.Screen, lm *layout.LayoutManager) {
	rect := tv.Rect
	bg := tcell.StyleDefault.Background(tcell.NewRGBColor(30, 30, 30))
	selectedStyle := tcell.StyleDefault.Background(tcell.NewRGBColor(70, 70, 200)).Foreground(tcell.ColorWhite)
	textStyle := tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.NewRGBColor(30, 30, 30))

	visibleHeight := rect.Height
	tv.ensureCursorVisible() // make sure cursor is visible

	start := tv.ScrollOffset
	for i := 0; i < visibleHeight; i++ {
		screenY := rect.Y + i
		nodeIdx := i + start

		// Clear the line
		for x := 0; x < rect.Width; x++ {
			screen.SetContent(rect.X+x, screenY, ' ', nil, bg)
		}

		if nodeIdx >= len(tv.Nodes) {
			continue
		}

		node := tv.Nodes[nodeIdx]

		// Build line with indentation
		line := ""
		for j := 0; j < tv.Levels[node]; j++ {
			line += "  "
		}
		if node.IsDir {
			if node.Expanded {
				line += "▾ "
			} else {
				line += "▸ "
			}
		} else {
			line += "  "
		}
		line += node.Name

		style := textStyle
		if nodeIdx == tv.CursorIdx {
			style = selectedStyle
		}

		runes := []rune(line)
		for x := 0; x < rect.Width; x++ {
			ch := ' '
			if x < len(runes) {
				ch = runes[x]
			}
			screen.SetContent(rect.X+x, screenY, ch, nil, style)
		}
	}
}
