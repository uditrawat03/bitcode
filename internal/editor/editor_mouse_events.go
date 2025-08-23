package editor

import "github.com/gdamore/tcell/v2"

func (ed *Editor) HandleMouse(ev *tcell.EventMouse) {
	x, y := ev.Position()
	btns := ev.Buttons()

	if ed.buffer == nil {
		return
	}

	// Mouse wheel scrolling
	if btns&tcell.WheelUp != 0 {
		ed.Scroll(-3)
		return
	}
	if btns&tcell.WheelDown != 0 {
		ed.Scroll(3)
		return
	}

	// Left click
	if btns&tcell.Button1 != 0 {
		if x >= ed.x && x < ed.x+ed.width && y >= ed.y && y < ed.y+ed.height {
			if ed.focusCb != nil {
				ed.focusCb()
			}

			clickedRow := y - ed.y + ed.scrollY
			if clickedRow < 0 {
				clickedRow = 0
			}

			// Clamp to existing buffer (no new lines added)
			if clickedRow >= len(ed.buffer.Content) {
				clickedRow = len(ed.buffer.Content) - 1
			}

			ed.buffer.CursorY = clickedRow

			// Compute column
			col := x - ed.x - 4
			if col < 0 {
				col = 0
			}
			lineLen := len(ed.buffer.Content[ed.buffer.CursorY])
			if col > lineLen {
				col = lineLen
			}
			ed.buffer.CursorX = col

			// Scroll adjustment
			if ed.buffer.CursorY < ed.scrollY {
				ed.scrollY = ed.buffer.CursorY
			} else if ed.buffer.CursorY >= ed.scrollY+ed.height {
				ed.scrollY = ed.buffer.CursorY - ed.height + 1
			}
		}
	}
}

func (ed *Editor) Scroll(dy int) {
	if ed.buffer == nil {
		return
	}

	ed.scrollY += dy
	if ed.scrollY < 0 {
		ed.scrollY = 0
	}

	maxScroll := len(ed.buffer.Content) - ed.height
	if maxScroll < 0 {
		maxScroll = 0
	}
	if ed.scrollY > maxScroll {
		ed.scrollY = maxScroll
	}
}
