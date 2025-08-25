package editor

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/uditrawat03/bitcode/internal/buffer"
	"github.com/uditrawat03/bitcode/internal/core"
	"github.com/uditrawat03/bitcode/internal/layout"
)

type Editor struct {
	core.BaseComponent
	focused bool

	scrollY int

	buffer *buffer.Buffer

	Content []string

	selecting     bool
	ctrlASelected bool
	selStartX     int
	selStartY     int
	selEndX       int
	selEndY       int

	clipboard []rune
}

func NewEditor() *Editor {
	return &Editor{
		Content: []string{"// Your code goes here"},
	}
}

func (ed *Editor) Render(screen tcell.Screen, lm *layout.LayoutManager) {
	bg := tcell.ColorBlack
	style := tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(bg)
	highlightStyle := tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.NewRGBColor(50, 50, 80))

	// Fill background
	for row := 0; row < ed.Rect.Height; row++ {
		idx := row + ed.scrollY
		currentLineStyle := style
		if ed.buffer != nil && idx == ed.buffer.CursorY {
			currentLineStyle = highlightStyle
		}
		for col := 0; col < ed.Rect.Width; col++ {
			screen.SetContent(ed.Rect.X+col, ed.Rect.Y+row, ' ', nil, currentLineStyle)
		}
	}

	if ed.buffer == nil {
		return
	}

	// Draw buffer lines
	for row := 0; row < ed.Rect.Height; row++ {
		idx := row + ed.scrollY
		if idx >= len(ed.buffer.Content) {
			break
		}
		line := string(ed.buffer.Content[idx])
		lnStr := fmt.Sprintf("%3d ", idx+1)

		// Style for this line
		currentLineStyle := style
		if idx == ed.buffer.CursorY || ed.isLineSelected(idx) {
			currentLineStyle = highlightStyle
		}

		// Line numbers
		for i, r := range lnStr {
			if i >= 4 || i >= ed.Rect.Width {
				break
			}
			screen.SetContent(ed.Rect.X+i, ed.Rect.Y+row, r, nil, currentLineStyle)
		}

		// Text (syntax highlight)
		tokens := highlightLine(line)
		for i, t := range tokens {
			if i+4 >= ed.Rect.Width {
				break
			}
			screen.SetContent(ed.Rect.X+4+i, ed.Rect.Y+row, t.ch, nil, t.style)
		}
	}

	// Cursor
	if ed.IsFocused() {
		cx := ed.Rect.X + 4 + ed.buffer.CursorX
		cy := ed.Rect.Y + (ed.buffer.CursorY - ed.scrollY)

		if cy >= ed.Rect.Y && cy < ed.Rect.Y+ed.Rect.Height &&
			cx >= ed.Rect.X && cx < ed.Rect.X+ed.Rect.Width {

			screen.ShowCursor(cx, cy)

			mainCh, comb, st, _ := screen.GetContent(cx, cy)
			if mainCh == ' ' {
				// draw visible block if empty
				screen.SetContent(cx, cy, '▉', nil, st)
			} else {
				// invert existing character
				screen.SetContent(cx, cy, mainCh, comb, st.Reverse(true))
			}
		} else {
			screen.HideCursor()
		}
	}
}

func (ed *Editor) Focus() {
	ed.focused = true
	ed.ensureCursorVisible()
}
func (ed *Editor) Blur()           { ed.focused = false }
func (ed *Editor) IsFocused() bool { return ed.focused }

func (ed *Editor) SetBuffer(buf *buffer.Buffer) {
	ed.buffer = buf
	ed.scrollY = 0
}

func (ed *Editor) GetBufferParentFolder() string {
	return ed.buffer.ParentFolder()
}
