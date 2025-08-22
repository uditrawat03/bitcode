package editor

import "github.com/gdamore/tcell/v2"

func (ed *Editor) HandleKey(ev *tcell.EventKey) {
	if !ed.focused || ed.buffer == nil {
		return
	}

	switch ev.Key() {
	case tcell.KeyUp, tcell.KeyDown, tcell.KeyLeft, tcell.KeyRight:
		ed.handleCursorMovement(ev)
	case tcell.KeyEnter:
		ed.handleEnter()
	case tcell.KeyBackspace, tcell.KeyBackspace2:
		ed.handleBackspace()
	case tcell.KeyCtrlS:
		ed.handleSave()
	default:
		ed.handleRune(ev)
	}

	ed.ensureCursorVisible()
}

// Cursor movement
func (ed *Editor) handleCursorMovement(ev *tcell.EventKey) {
	switch ev.Key() {
	case tcell.KeyUp:
		if ed.buffer.CursorY > 0 {
			ed.buffer.CursorY--
			if ed.buffer.CursorX > len(ed.buffer.Content[ed.buffer.CursorY]) {
				ed.buffer.CursorX = len(ed.buffer.Content[ed.buffer.CursorY])
			}
		}
	case tcell.KeyDown:
		if ed.buffer.CursorY < len(ed.buffer.Content)-1 {
			ed.buffer.CursorY++
			if ed.buffer.CursorX > len(ed.buffer.Content[ed.buffer.CursorY]) {
				ed.buffer.CursorX = len(ed.buffer.Content[ed.buffer.CursorY])
			}
		}
	case tcell.KeyLeft:
		if ed.buffer.CursorX > 0 {
			ed.buffer.CursorX--
		} else if ed.buffer.CursorY > 0 {
			ed.buffer.CursorY--
			ed.buffer.CursorX = len(ed.buffer.Content[ed.buffer.CursorY])
		}
	case tcell.KeyRight:
		if ed.buffer.CursorX < len(ed.buffer.Content[ed.buffer.CursorY]) {
			ed.buffer.CursorX++
		} else if ed.buffer.CursorY < len(ed.buffer.Content)-1 {
			ed.buffer.CursorY++
			ed.buffer.CursorX = 0
		}
	}
}

// Enter key inserts a new line
func (ed *Editor) handleEnter() {
	ed.buffer.InsertLine()
}

// Backspace deletes a character
func (ed *Editor) handleBackspace() {
	ed.buffer.DeleteRune()
}

// Save buffer
func (ed *Editor) handleSave() {
	ed.buffer.Save()
}

// Insert typed rune
func (ed *Editor) handleRune(ev *tcell.EventKey) {
	if ev.Rune() != 0 {
		ed.buffer.InsertRune(ev.Rune())
	}
}

func (ed *Editor) ensureCursorVisible() {
	if ed.buffer == nil {
		return
	}
	if ed.buffer.CursorY < ed.scrollY {
		ed.scrollY = ed.buffer.CursorY
	}
	if ed.buffer.CursorY >= ed.scrollY+ed.height {
		ed.scrollY = ed.buffer.CursorY - ed.height + 1
	}
}
