package ui

import (
	"github.com/gdamore/tcell/v2"
	"github.com/uditrawat03/bitcode/internal/layout"
)

type ToolTip struct {
	BaseComponent
	Message string
	focused bool
}

func NewToolTip(msg string) *ToolTip {
	return &ToolTip{Message: msg}
}

func (t *ToolTip) Render(screen tcell.Screen, lm *layout.LayoutManager) {
	t.Rect.Width = len(t.Message) + 2
	t.Rect.Height = 3
	style := tcell.StyleDefault.Background(tcell.ColorWhite).Foreground(tcell.ColorBlack)
	for y := 0; y < t.Rect.Height; y++ {
		for x := 0; x < t.Rect.Width; x++ {
			screen.SetContent(t.Rect.X+x, t.Rect.Y+y, ' ', nil, style)
		}
	}
	for i, r := range t.Message {
		if i < t.Rect.Width {
			screen.SetContent(t.Rect.X+1+i, t.Rect.Y+1, r, nil, style)
		}
	}
}

func (t *ToolTip) Focus()                           { t.focused = true }
func (t *ToolTip) Blur()                            { t.focused = false }
func (t *ToolTip) IsFocused() bool                  { return t.focused }
func (t *ToolTip) HandleKey(ev *tcell.EventKey)     {}
func (t *ToolTip) HandleMouse(ev *tcell.EventMouse) {}
