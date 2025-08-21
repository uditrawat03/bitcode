package statusbar

import "github.com/gdamore/tcell/v2"

type StatusBar struct {
	x, y, width, height int
	focused             bool
}

func CreateStatusBar(x, y, width, height int) *StatusBar {
	return &StatusBar{x: x, y: y, width: width, height: height}
}

// Focusable
func (sb *StatusBar) Focus()                           { sb.focused = true }
func (sb *StatusBar) Blur()                            { sb.focused = false }
func (sb *StatusBar) IsFocused() bool                  { return sb.focused }
func (sb *StatusBar) HandleKey(ev *tcell.EventKey)     {}
func (sb *StatusBar) HandleMouse(ev *tcell.EventMouse) {}

// Draw
func (sb *StatusBar) Draw(s tcell.Screen) {
	style := tcell.StyleDefault.Foreground(tcell.ColorBlack).Background(tcell.ColorYellow)
	if sb.focused {
		style = style.Reverse(true)
	}
	content := " Status: Ready "
	for row := 0; row < sb.height; row++ {
		for col := 0; col < sb.width; col++ {
			s.SetContent(sb.x+col, sb.y+row, ' ', nil, style)
		}
	}
	for i, r := range content {
		if i >= sb.width {
			break
		}
		s.SetContent(sb.x+i, sb.y, r, nil, style)
	}
}
