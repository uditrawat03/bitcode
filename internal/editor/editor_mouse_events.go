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
		if x >= ed.Rect.X && x < ed.Rect.X+ed.Rect.Width && y >= ed.Rect.Y && y < ed.Rect.Y+ed.Rect.Height {
			// if ed.focusCb != nil {
			// 	ed.focusCb()
			// }

			ed.Focus()

			clickedRow := y - ed.Rect.Y + ed.scrollY
			if clickedRow < 0 {
				clickedRow = 0
			}

			// Clamp to existing buffer (no new lines added)
			if clickedRow >= len(ed.buffer.Content) {
				clickedRow = len(ed.buffer.Content) - 1
			}

			ed.buffer.CursorY = clickedRow

			// Compute column
			col := x - ed.Rect.X - 4
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
			} else if ed.buffer.CursorY >= ed.scrollY+ed.Rect.Height {
				ed.scrollY = ed.buffer.CursorY - ed.Rect.Height + 1
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

	maxScroll := len(ed.buffer.Content) - ed.Rect.Height
	if maxScroll < 0 {
		maxScroll = 0
	}
	if ed.scrollY > maxScroll {
		ed.scrollY = maxScroll
	}
}
