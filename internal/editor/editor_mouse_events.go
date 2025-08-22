package editor

import "github.com/gdamore/tcell/v2"

func (ed *Editor) HandleMouse(ev *tcell.EventMouse) {
	x, y := ev.Position()
	if ev.Buttons()&tcell.Button1 != 0 { // left click
		if x >= ed.x && x < ed.x+ed.width && y >= ed.y && y < ed.y+ed.height {
			if ed.focusCb != nil {
				ed.focusCb() // tell ScreenManager to focus editor
			}

			// update cursor to click position
			if ed.buffer != nil {
				col := x - ed.x - 4 // account for line number gutter
				if col < 0 {
					col = 0
				}
				if col > len(ed.buffer.Content[ed.buffer.CursorY]) {
					col = len(ed.buffer.Content[ed.buffer.CursorY])
				}
				ed.buffer.CursorX = col

				row := y - ed.y + ed.scrollY
				if row >= 0 && row < len(ed.buffer.Content) {
					ed.buffer.CursorY = row
				}
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
	if ed.scrollY > len(ed.buffer.Content)-ed.height {
		ed.scrollY = max(0, len(ed.buffer.Content)-ed.height)
	}
}
