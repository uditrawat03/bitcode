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
}

func CreateEditor(x, y, width, height int) *Editor {
	return &Editor{x: x, y: y, width: width, height: height}
}

func (ed *Editor) SetBuffer(buf *buffer.Buffer) {
	ed.buffer = buf
	ed.scrollY = 0
}

// Focusable methods
func (ed *Editor) Focus()                           { ed.focused = true }
func (ed *Editor) Blur()                            { ed.focused = false }
func (ed *Editor) IsFocused() bool                  { return ed.focused }
func (ed *Editor) HandleMouse(ev *tcell.EventMouse) {}
func (ed *Editor) HandleKey(ev *tcell.EventKey) {
	if !ed.focused {
		return
	}
	switch ev.Key() {
	case tcell.KeyUp:
		ed.Scroll(-1)
	case tcell.KeyDown:
		ed.Scroll(1)
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
		line := ed.buffer.Content[idx]
		lnStr := fmt.Sprintf("%3d ", idx+1)

		for i, r := range lnStr {
			if i >= 4 || i >= ed.width {
				break
			}
			screen.SetContent(ed.x+i, ed.y+row, r, nil, lnStyle)
		}
		for i, r := range line {
			if i+4 >= ed.width {
				break
			}
			screen.SetContent(ed.x+4+i, ed.y+row, r, nil, style)
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
