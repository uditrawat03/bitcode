package editor

import (
	"strings"

	"github.com/atotto/clipboard"
	"github.com/gdamore/tcell/v2"
)

// --- Edit tracking ---
type EditType int

const (
	Insert EditType = iota
	Delete
)

type Edit struct {
	Typ        EditType
	PosX, PosY int
	Text       []rune
}

func (ed *Editor) isTypingKey(ev *tcell.EventKey) bool {
	switch ev.Key() {
	case tcell.KeyRune, tcell.KeyTab, tcell.KeyEnter, tcell.KeyBackspace, tcell.KeyBackspace2, tcell.KeyDelete:
		return true
	default:
		return false
	}
}

func (ed *Editor) HandleKey(ev *tcell.EventKey) {
	if !ed.focused || ed.buffer == nil {
		return
	}

	// Cancel Ctrl+A selection on typing keys (except delete/backspace/tab)
	if ed.ctrlASelected && ed.isTypingKey(ev) && ev.Key() != tcell.KeyBackspace && ev.Key() != tcell.KeyDelete && ev.Key() != tcell.KeyTab {
		ed.selecting = false
		ed.ctrlASelected = false
	}

	shift := ev.Modifiers()&tcell.ModShift != 0

	// Start shift-selection
	if shift && !ed.selecting {
		ed.selecting = true
		ed.selStartX = ed.buffer.CursorX
		ed.selStartY = ed.buffer.CursorY
	}

	if ed.tooltip.Visible {
		switch ev.Key() {
		case tcell.KeyUp:
			ed.tooltip.Prev()
			return
		case tcell.KeyDown:
			ed.tooltip.Next()
			return
		case tcell.KeyEnter:
			ed.tooltip.Apply(func(item string) {
				// TODO: hook this to actual action in editor
				ed.buffer.InsertRune('\n')
				for _, r := range item {
					ed.buffer.InsertRune(r)
				}
			})
			return
		case tcell.KeyEsc:
			ed.tooltip.Close()
			return
		default:
			ed.tooltip.Close()
			return
		}
	}

	switch ev.Key() {
	case tcell.KeyUp, tcell.KeyDown, tcell.KeyLeft, tcell.KeyRight:
		ed.handleCursorMovement(ev)
		ed.ShowHover()
		ed.ShowDiagnostics()
	case tcell.KeyHome:
		ed.handleHome()
	case tcell.KeyEnd:
		ed.handleEnd()
	case tcell.KeyPgUp:
		ed.handlePageUp()
	case tcell.KeyPgDn:
		ed.handlePageDown()
	case tcell.KeyTab:
		ed.handleTab()
		ed.ShowCompletion(ed.ctx)
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
	case tcell.KeyCtrlZ:
		ed.buffer.Undo()
	case tcell.KeyCtrlY:
		ed.buffer.Redo()
	case tcell.KeyCtrlL:
		ed.ShowCodeActions()
	default:
		ed.handleRune(ev)
	}

	// Update selection end for shift-selection
	if ed.selecting && !ed.ctrlASelected {
		ed.selEndX = ed.buffer.CursorX
		ed.selEndY = ed.buffer.CursorY
	}

	ed.ensureCursorVisible()
}

// --- Cursor movement ---
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

func (ed *Editor) handleHome() { ed.buffer.CursorX = 0 }
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

// --- Editing ---
func (ed *Editor) handleTab() {
	if ed.buffer == nil {
		return
	}
	const tabSize = 4

	if ed.hasSelection() || ed.ctrlASelected {
		startY, endY := ed.selStartY, ed.selEndY
		if ed.ctrlASelected {
			startY = 0
			endY = len(ed.buffer.Content) - 1
		}
		if startY > endY {
			startY, endY = endY, startY
		}
		for y := startY; y <= endY; y++ {
			line := string(ed.buffer.Content[y])
			ed.buffer.SetLine(y, strings.Repeat(" ", tabSize)+line)
		}
	} else {
		for i := 0; i < tabSize; i++ {
			ed.buffer.InsertRune(' ')
		}
	}
}

func (ed *Editor) handleEnter() { ed.buffer.InsertLine() }

func (ed *Editor) handleBackspace() {
	if ed.ctrlASelected || ed.hasSelection() {
		ed.buffer.DeleteSelectionOrAll(ed.selStartX, ed.selStartY, ed.selEndX, ed.selEndY, ed.ctrlASelected)
		ed.selecting = false
		ed.ctrlASelected = false
		return
	}
	ed.buffer.DeleteRune()
}

func (ed *Editor) handleDelete() {
	if ed.ctrlASelected || ed.hasSelection() {
		ed.buffer.DeleteSelectionOrAll(ed.selStartX, ed.selStartY, ed.selEndX, ed.selEndY, ed.ctrlASelected)
		ed.selecting = false
		ed.ctrlASelected = false
		return
	}
	ed.buffer.DeleteAtCursor(ed.buffer.CursorX, ed.buffer.CursorY, ed.selStartY, ed.selEndY, false)
}

func (ed *Editor) handleRune(ev *tcell.EventKey) {
	if ev.Rune() != 0 {
		if ed.ctrlASelected || ed.hasSelection() {
			ed.buffer.DeleteSelectionOrAll(ed.selStartX, ed.selStartY, ed.selEndX, ed.selEndY, ed.ctrlASelected)
			ed.selecting = false
			ed.ctrlASelected = false
		}
		ed.buffer.InsertRune(ev.Rune())
	}
}

func (ed *Editor) handleSave() { ed.buffer.Save() }

func (ed *Editor) handleCut() {
	if !ed.hasSelection() && !ed.ctrlASelected {
		return
	}
	text := ed.buffer.CopyAllOrSelection(ed.selStartX, ed.selStartY, ed.selEndX, ed.selEndY, ed.ctrlASelected)
	ed.buffer.DeleteSelectionOrAll(ed.selStartX, ed.selStartY, ed.selEndX, ed.selEndY, ed.ctrlASelected)

	ed.clipboard = text
	clipboard.WriteAll(string(text))
	ed.selecting = false
	ed.ctrlASelected = false
}

func (ed *Editor) handleCopy() {
	if !ed.hasSelection() && !ed.ctrlASelected {
		return
	}
	text := ed.buffer.CopyAllOrSelection(ed.selStartX, ed.selStartY, ed.selEndX, ed.selEndY, ed.ctrlASelected)
	ed.clipboard = text
	clipboard.WriteAll(string(text))
}

func (ed *Editor) handlePaste() {
	str, err := clipboard.ReadAll()
	if err != nil || str == "" {
		if len(ed.clipboard) == 0 {
			return
		}
		str = string(ed.clipboard)
	}
	text := []rune(str)

	if ed.ctrlASelected || ed.hasSelection() {
		ed.buffer.DeleteSelectionOrAll(ed.selStartX, ed.selStartY, ed.selEndX, ed.selEndY, ed.ctrlASelected)
		ed.selStartX, ed.selStartY = 0, 0
	}

	ed.buffer.PasteClipboard(text)
	ed.selecting = false
	ed.ctrlASelected = false
}

// --- Selection ---
func (ed *Editor) handleSelectAll() {
	if ed.buffer == nil || len(ed.buffer.Content) == 0 {
		return
	}
	ed.selStartX, ed.selStartY = 0, 0
	lastLine := len(ed.buffer.Content) - 1
	ed.selEndY = lastLine
	ed.selEndX = len(ed.buffer.Content[lastLine])
	ed.selecting = true
	ed.ctrlASelected = true
}

func (ed *Editor) isLineSelected(y int) bool {
	if !ed.selecting && !ed.ctrlASelected {
		return false
	}
	if ed.ctrlASelected {
		return true
	}
	start, end := ed.selStartY, ed.selEndY
	if start > end {
		start, end = end, start
	}
	return y >= start && y <= end
}

func (ed *Editor) hasSelection() bool {
	return ed.selecting && len(ed.buffer.Content) > 0
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
