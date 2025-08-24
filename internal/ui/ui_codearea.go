package ui

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/uditrawat03/bitcode/internal/buffer"
	"github.com/uditrawat03/bitcode/internal/editor"
	"github.com/uditrawat03/bitcode/internal/layout"
)

type CodeArea struct {
	BaseComponent
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
}

func NewCodeArea() *CodeArea {
	return &CodeArea{
		Content: []string{"// Your code goes here"},
	}
}

func (c *CodeArea) Render(screen tcell.Screen, lm *layout.LayoutManager) {
	bg := tcell.ColorBlack

	style := tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(bg)

	// Line highlight style (full width)
	highlightStyle := tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.NewRGBColor(50, 50, 80))

	for row := 0; row < c.Rect.Height; row++ {
		idx := row + c.scrollY
		currentLineStyle := style
		if c.buffer != nil && idx == c.buffer.CursorY {
			currentLineStyle = highlightStyle
		}

		for col := 0; col < c.Rect.Width; col++ {
			screen.SetContent(c.Rect.X+col, c.Rect.Y+row, ' ', nil, currentLineStyle)
		}
	}

	if c.buffer == nil {
		return
	}

	// draw buffer lines with line numbers
	for row := 0; row < c.Rect.Height; row++ {
		idx := row + c.scrollY
		if idx >= len(c.buffer.Content) {
			break
		}
		line := string(c.buffer.Content[idx])
		lnStr := fmt.Sprintf("%3d ", idx+1)

		// Determine style for this line
		currentLineStyle := style
		if idx == c.buffer.CursorY {
			currentLineStyle = highlightStyle
		}

		if idx == c.buffer.CursorY || c.isLineSelected(idx) {
			currentLineStyle = highlightStyle // full-width highlight
		}

		// Draw line numbers
		for i, r := range lnStr {
			if i >= 4 || i >= c.Rect.Width {
				break
			}
			screen.SetContent(c.Rect.X+i, c.Rect.Y+row, r, nil, currentLineStyle)
		}

		// Draw text
		// for i, r := range line {
		// 	if i+4 >= c.width {
		// 		break
		// 	}
		// 	screen.SetContent(c.x+4+i, c.y+row, r, nil, currentLineStyle)
		// }

		tokens := editor.GetHighlightLine(line)
		for i, t := range tokens {
			if i+4 >= c.Rect.Width {
				break
			}
			screen.SetContent(c.Rect.X+4+i, c.Rect.Y+row, t.Ch, nil, t.Style)
		}
	}

	// draw cursor
	if c.focused {
		cx := c.Rect.X + 4 + c.buffer.CursorX
		cy := c.Rect.Y + c.buffer.CursorY - c.scrollY
		if cy >= 0 && cy < c.Rect.Height && cx >= c.Rect.X && cx < c.Rect.Y+c.Rect.Width {
			screen.ShowCursor(cx, cy)
		}
	}
}

func (c *CodeArea) Focus()                           { c.focused = true }
func (c *CodeArea) Blur()                            { c.focused = false }
func (c *CodeArea) IsFocused() bool                  { return c.focused }
func (c *CodeArea) HandleKey(ev *tcell.EventKey)     {}
func (c *CodeArea) HandleMouse(ev *tcell.EventMouse) {}

func (c *CodeArea) SetBuffer(buf *buffer.Buffer) {
	c.buffer = buf
	c.scrollY = 0
}

func (c *CodeArea) isLineSelected(y int) bool {
	if !c.selecting && !c.ctrlASelected {
		return false
	}
	if c.ctrlASelected {
		return true
	}
	start, end := c.selStartY, c.selEndY
	if start > end {
		start, end = end, start
	}
	return y >= start && y <= end
}
