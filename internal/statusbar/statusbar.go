package statusbar

import (
	"github.com/gdamore/tcell/v2"
	"github.com/uditrawat03/bitcode/internal/core"
	"github.com/uditrawat03/bitcode/internal/layout"
)

type BottomBar struct {
	core.BaseComponent
	Message string
}

func NewStatusBar(msg string) *BottomBar {
	return &BottomBar{Message: msg}
}

func (b *BottomBar) Render(screen tcell.Screen, lm *layout.LayoutManager) {
	style := tcell.StyleDefault.Background(tcell.ColorGray).Foreground(tcell.ColorBlack)
	for x := 0; x < b.Rect.Width; x++ {
		screen.SetContent(b.Rect.X+x, b.Rect.Y, ' ', nil, style)
	}
	for i, r := range b.Message {
		if i < b.Rect.Width {
			screen.SetContent(b.Rect.X+i, b.Rect.Y, r, nil, style)
		}
	}
}

func (b *BottomBar) Focus()                           {}
func (b *BottomBar) Blur()                            {}
func (b *BottomBar) IsFocused() bool                  { return false }
func (b *BottomBar) HandleKey(ev *tcell.EventKey)     {}
func (b *BottomBar) HandleMouse(ev *tcell.EventMouse) {}
