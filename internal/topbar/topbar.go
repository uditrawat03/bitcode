package topbar

import (
	"github.com/gdamore/tcell/v2"
	"github.com/uditrawat03/bitcode/internal/core"
	"github.com/uditrawat03/bitcode/internal/layout"
)

type TopBar struct {
	core.BaseComponent
	Title string
}

func NewTopBar(title string) *TopBar {
	return &TopBar{Title: title}
}

func (t *TopBar) Render(screen tcell.Screen, lm *layout.LayoutManager) {
	style := tcell.StyleDefault.Background(tcell.ColorBlue).Foreground(tcell.ColorWhite)
	for x := 0; x < t.Rect.Width; x++ {
		screen.SetContent(t.Rect.X+x, t.Rect.Y, ' ', nil, style)
	}
	// Draw title
	for i, r := range t.Title {
		if i < t.Rect.Width {
			screen.SetContent(t.Rect.X+i, t.Rect.Y, r, nil, style)
		}
	}
}

func (t *TopBar) Focus()                           {}
func (t *TopBar) Blur()                            {}
func (t *TopBar) IsFocused() bool                  { return false }
func (t *TopBar) HandleKey(ev *tcell.EventKey)     {}
func (t *TopBar) HandleMouse(ev *tcell.EventMouse) {}
