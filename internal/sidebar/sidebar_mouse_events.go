package sidebar

import "github.com/gdamore/tcell/v2"

// HandleMouse handles clicks, scroll, and resize
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

	// ignore clicks outside sidebar
	if x < sb.X || x >= sb.X+sb.Width || y < sb.Y || y >= sb.Y+sb.Height {
		return
	}

	// scroll wheel
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
	if idx < 0 || idx >= len(sb.Tree.Nodes) {
		return
	}

	// trigger focus first
	if ev.Buttons()&tcell.Button1 != 0 && sb.focusCb != nil {
		sb.focusCb()
	}

	sb.Hovered = idx

	if ev.Buttons()&tcell.Button1 != 0 {
		sb.Selected = idx
		node := sb.Tree.Nodes[idx]
		if node.IsDir {
			sb.Tree.Toggle(node)
		} else if sb.onFileOpen != nil {
			sb.onFileOpen(node.Path)
		}
	}
}

// maxScroll returns maximum scroll offset
func (sb *Sidebar) maxScroll() int {
	if len(sb.Tree.Nodes) > sb.Height {
		return len(sb.Tree.Nodes) - sb.Height
	}
	return 0
}

// Scroll by delta
func (sb *Sidebar) Scroll(delta int) {
	sb.Selected += delta
	if sb.Selected < 0 {
		sb.Selected = 0
	} else if sb.Selected >= len(sb.Tree.Nodes) {
		sb.Selected = len(sb.Tree.Nodes) - 1
	}

	// Keep selected in visible window
	if sb.Selected < sb.ScrollY {
		sb.ScrollY = sb.Selected
	} else if sb.Selected >= sb.ScrollY+sb.Height {
		sb.ScrollY = sb.Selected - sb.Height + 1
	}
}
