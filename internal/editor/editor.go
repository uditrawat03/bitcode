package editor

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/uditrawat03/bitcode/internal/buffer"
)

type Editor struct {
	x, y, width, height int
	scrollY             int
	focused             bool
	buffer              *buffer.Buffer

	focusCb func()
}

func (ed *Editor) SetFocusCallback(cb func()) {
	ed.focusCb = cb
}

func CreateEditor(x, y, width, height int) *Editor {
	return &Editor{x: x, y: y, width: width, height: height}
}

func (ed *Editor) SetBuffer(buf *buffer.Buffer) {
	ed.buffer = buf
	ed.scrollY = 0
}

// Focusable methods
func (ed *Editor) Focus()          { ed.focused = true }
func (ed *Editor) Blur()           { ed.focused = false }
func (ed *Editor) IsFocused() bool { return ed.focused }

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

func (ed *Editor) HandleKey(ev *tcell.EventKey) {
	if !ed.focused || ed.buffer == nil {
		return
	}

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
	case tcell.KeyEnter:
		ed.buffer.InsertLine()
	case tcell.KeyBackspace, tcell.KeyBackspace2:
		ed.buffer.DeleteRune()
	case tcell.KeyCtrlS:
		ed.buffer.Save()
	default:
		if ev.Rune() != 0 {
			ed.buffer.InsertRune(ev.Rune())
		}
	}

	ed.ensureCursorVisible()
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

// Draw editor content
func (ed *Editor) Draw(screen tcell.Screen) {
	bg := tcell.ColorBlack
	if ed.focused {
		bg = tcell.ColorDarkBlue
	}
	style := tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(bg)
	lnStyle := tcell.StyleDefault.Foreground(tcell.ColorYellow).Background(bg)

	// background
	for row := 0; row < ed.height; row++ {
		for col := 0; col < ed.width; col++ {
			screen.SetContent(ed.x+col, ed.y+row, ' ', nil, style)
		}
	}

	if ed.buffer == nil {
		return
	}

	// draw buffer lines
	for row := 0; row < ed.height; row++ {
		idx := row + ed.scrollY
		if idx >= len(ed.buffer.Content) {
			break
		}
		line := string(ed.buffer.Content[idx])
		lnStr := fmt.Sprintf("%3d ", idx+1)

		// line numbers
		for i, r := range lnStr {
			if i >= 4 || i >= ed.width {
				break
			}
			screen.SetContent(ed.x+i, ed.y+row, r, nil, lnStyle)
		}
		// text
		for i, r := range line {
			if i+4 >= ed.width {
				break
			}
			screen.SetContent(ed.x+4+i, ed.y+row, r, nil, style)
		}
	}

	// draw cursor
	if ed.focused {
		cx := ed.x + 4 + ed.buffer.CursorX
		cy := ed.y + ed.buffer.CursorY - ed.scrollY
		if cy >= 0 && cy < ed.height && cx >= ed.x && cx < ed.x+ed.width {
			screen.ShowCursor(cx, cy)
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

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
