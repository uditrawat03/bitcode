package topbar

import "github.com/gdamore/tcell/v2"

type Resizable interface {
	Resize(x, y, w, h int)
}

type TopBar struct {
	x, y, width, height int
	focused             bool
}

func CreateTopBar(x, y, width, height int) *TopBar {
	return &TopBar{x: x, y: y, width: width, height: height}
}

func (tb *TopBar) Resize(x, y, w, h int) {
	tb.x, tb.y, tb.width, tb.height = x, y, w, h
}

// Focusable
func (tb *TopBar) Focus()                           { tb.focused = true }
func (tb *TopBar) Blur()                            { tb.focused = false }
func (tb *TopBar) IsFocused() bool                  { return tb.focused }
func (tb *TopBar) HandleKey(ev *tcell.EventKey)     {}
func (tb *TopBar) HandleMouse(ev *tcell.EventMouse) {}

// Draw
func (tb *TopBar) Draw(s tcell.Screen) {
	style := tcell.StyleDefault.Foreground(tcell.ColorBlack).Background(tcell.ColorGreen)
	if tb.focused {
		style = style.Reverse(true)
	}

	content := " Top Bar Content "
	for row := 0; row < tb.height; row++ {
		for col := 0; col < tb.width; col++ {
			s.SetContent(tb.x+col, tb.y+row, ' ', nil, style)
		}
	}

	for i, r := range content {
		if i >= tb.width {
			break
		}
		s.SetContent(tb.x+i, tb.y, r, nil, style)
	}
}
