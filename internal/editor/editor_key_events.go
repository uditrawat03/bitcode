package editor

import (
	"github.com/atotto/clipboard"
	"github.com/gdamore/tcell/v2"
)

func (ed *Editor) HandleKey(ev *tcell.EventKey) {
	if !ed.focused || ed.buffer == nil {
		return
	}

	shift := ev.Modifiers()&tcell.ModShift != 0

	// Start selection if Shift pressed
	if shift && !ed.selecting {
		ed.selecting = true
		ed.selStartX = ed.buffer.CursorX
		ed.selStartY = ed.buffer.CursorY
	}

	switch ev.Key() {
	case tcell.KeyUp, tcell.KeyDown, tcell.KeyLeft, tcell.KeyRight:
		ed.handleCursorMovement(ev)
	case tcell.KeyHome:
		ed.handleHome()
	case tcell.KeyEnd:
		ed.handleEnd()
	case tcell.KeyPgUp:
		ed.handlePageUp()
	case tcell.KeyPgDn:
		ed.handlePageDown()
	case tcell.KeyDelete:
		ed.handleDelete()
	case tcell.KeyEnter:
		ed.handleEnter()
	case tcell.KeyBackspace, tcell.KeyBackspace2:
		ed.handleBackspace()
	case tcell.KeyCtrlX:
		ed.handleCut()
	case tcell.KeyCtrlC:
		ed.handleCopy()
	case tcell.KeyCtrlV:
		ed.handlePaste()
	case tcell.KeyCtrlS:
		ed.handleSave()
	case tcell.KeyCtrlA:
		ed.handleSelectAll()
	default:
		ed.handleRune(ev)
	}

	// Update selection end
	if ed.selecting {
		ed.selEndX = ed.buffer.CursorX
		ed.selEndY = ed.buffer.CursorY
	}

	ed.ensureCursorVisible()
}

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

func (ed *Editor) handleHome() {
	ed.buffer.CursorX = 0
}

func (ed *Editor) handleEnd() {
	if ed.buffer.CursorY < len(ed.buffer.Content) {
		ed.buffer.CursorX = len(ed.buffer.Content[ed.buffer.CursorY])
	}
}

func (ed *Editor) handlePageUp() {
	ed.buffer.CursorY -= ed.height
	if ed.buffer.CursorY < 0 {
		ed.buffer.CursorY = 0
	}
}

func (ed *Editor) handlePageDown() {
	ed.buffer.CursorY += ed.height
	if ed.buffer.CursorY >= len(ed.buffer.Content) {
		ed.buffer.CursorY = len(ed.buffer.Content) - 1
	}
}

func (ed *Editor) handleEnter() {
	ed.buffer.InsertLine()
}

func (ed *Editor) handleBackspace() {
	ed.buffer.DeleteRune()
}

func (ed *Editor) handleDelete() {
	ed.buffer.DeleteAtCursor(ed.buffer.CursorX, ed.buffer.CursorY,
		ed.selStartY, ed.selEndY, ed.selecting)
	ed.selecting = false
}

func (ed *Editor) handleRune(ev *tcell.EventKey) {
	if ev.Rune() != 0 {
		ed.buffer.InsertRune(ev.Rune())
	}
}

func (ed *Editor) handleSave() {
	ed.buffer.Save()
}

// Cut selected text and also write to system clipboard
func (ed *Editor) handleCut() {
	if !ed.selecting {
		return
	}
	text := ed.buffer.CutSelection(ed.selStartX, ed.selStartY, ed.selEndX, ed.selEndY)
	ed.selecting = false

	ed.clipboard = text
	clipboard.WriteAll(string(text))
}

// Copy selected text and also write to system clipboard
func (ed *Editor) handleCopy() {
	if !ed.selecting {
		return
	}
	text := ed.buffer.CopySelection(ed.selStartX, ed.selStartY, ed.selEndX, ed.selEndY)
	ed.clipboard = text
	clipboard.WriteAll(string(text))
}

// Paste from system clipboard (or internal fallback)
func (ed *Editor) handlePaste() {
	str, err := clipboard.ReadAll()
	if err != nil || str == "" {
		if len(ed.clipboard) == 0 {
			return
		}
		str = string(ed.clipboard)
	}

	text := []rune(str)
	ed.buffer.PasteClipboard(text)
}

// Ctrl+A select all
func (ed *Editor) handleSelectAll() {
	ed.selStartX, ed.selStartY = 0, 0
	lastLine := len(ed.buffer.Content) - 1
	ed.selEndY = lastLine
	ed.selEndX = len(ed.buffer.Content[lastLine])
	ed.selecting = true
}

// Selection check
func (ed *Editor) isLineSelected(y int) bool {
	if !ed.selecting {
		return false
	}
	start, end := ed.selStartY, ed.selEndY
	if start > end {
		start, end = end, start
	}
	return y >= start && y <= end
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
