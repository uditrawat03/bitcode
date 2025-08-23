package editor

import (
	"context"
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/uditrawat03/bitcode/internal/buffer"
	"github.com/uditrawat03/bitcode/internal/tooltip"
)

type Resizable interface {
	Resize(x, y, w, h int)
}

type Editor struct {
	ctx                 context.Context
	x, y, width, height int
	scrollY             int
	focused             bool
	buffer              *buffer.Buffer

	clipboard []rune

	selecting     bool
	ctrlASelected bool
	selStartX     int
	selStartY     int
	selEndX       int
	selEndY       int

	tooltip tooltip.Tooltip

	focusCb func()
}

func (ed *Editor) SetFocusCallback(cb func()) {
	ed.focusCb = cb
}

func CreateEditor(ctx context.Context, x, y, width, height int) *Editor {
	return &Editor{ctx: ctx, x: x, y: y, width: width, height: height}
}

func (ed *Editor) Resize(x, y, w, h int) {
	ed.x, ed.y, ed.width, ed.height = x, y, w, h
}

func (ed *Editor) SetBuffer(buf *buffer.Buffer) {
	ed.buffer = buf
	ed.scrollY = 0
}

// Focusable methods
func (ed *Editor) Focus()          { ed.focused = true }
func (ed *Editor) Blur()           { ed.focused = false }
func (ed *Editor) IsFocused() bool { return ed.focused }

// Draw editor content
func (ed *Editor) Draw(screen tcell.Screen) {
	bg := tcell.ColorBlack

	style := tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(bg)

	// Line highlight style (full width)
	highlightStyle := tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.NewRGBColor(50, 50, 80))

	// background + line highlighting
	for row := 0; row < ed.height; row++ {
		idx := row + ed.scrollY
		currentLineStyle := style
		if ed.buffer != nil && idx == ed.buffer.CursorY {
			currentLineStyle = highlightStyle
		}

		for col := 0; col < ed.width; col++ {
			screen.SetContent(ed.x+col, ed.y+row, ' ', nil, currentLineStyle)
		}
	}

	if ed.buffer == nil {
		return
	}

	// draw buffer lines with line numbers
	for row := 0; row < ed.height; row++ {
		idx := row + ed.scrollY
		if idx >= len(ed.buffer.Content) {
			break
		}
		line := string(ed.buffer.Content[idx])
		lnStr := fmt.Sprintf("%3d ", idx+1)

		// Determine style for this line
		currentLineStyle := style
		if idx == ed.buffer.CursorY {
			currentLineStyle = highlightStyle
		}

		if idx == ed.buffer.CursorY || ed.isLineSelected(idx) {
			currentLineStyle = highlightStyle // full-width highlight
		}

		// Draw line numbers
		for i, r := range lnStr {
			if i >= 4 || i >= ed.width {
				break
			}
			screen.SetContent(ed.x+i, ed.y+row, r, nil, currentLineStyle)
		}

		// Draw text
		for i, r := range line {
			if i+4 >= ed.width {
				break
			}
			screen.SetContent(ed.x+4+i, ed.y+row, r, nil, currentLineStyle)
		}
	}

	if ed.tooltip.Visible {
		switch ed.tooltip.Type {
		case tooltip.TooltipText:
			// draw a simple text box
			for i, r := range ed.tooltip.Content {
				if ed.tooltip.X+i < ed.x+ed.width && ed.tooltip.Y < ed.y+ed.height {
					screen.SetContent(ed.tooltip.X+i, ed.tooltip.Y, r, nil, style)
				}
			}

		case tooltip.TooltipList:
			for idx, item := range ed.tooltip.Items {
				y := ed.tooltip.Y + idx
				if y >= ed.y+ed.height {
					break
				}
				itemStyle := style
				if idx == ed.tooltip.Selected {
					itemStyle = tcell.StyleDefault.Foreground(tcell.ColorBlack).Background(tcell.ColorWhite)
				}
				for i, r := range item {
					if ed.tooltip.X+i < ed.x+ed.width {
						screen.SetContent(ed.tooltip.X+i, y, r, nil, itemStyle)
					}
				}
			}
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

func (ed *Editor) GetBuffer() *buffer.Buffer {
	return ed.buffer
}

func (ed *Editor) GetBufferParentFolder() string {
	return ed.buffer.ParentFolder()
}
