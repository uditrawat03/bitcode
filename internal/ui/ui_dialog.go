package ui

import (
	"github.com/gdamore/tcell/v2"
	"github.com/uditrawat03/bitcode/internal/layout"
)

type Dialog struct {
	BaseComponent
	Title   string
	Content []string
	focused bool
	Width   int
	Height  int
}

func NewDialog(title string, content []string, width, height int) *Dialog {
	return &Dialog{Title: title, Content: content, Width: width, Height: height}
}

func (d *Dialog) Render(screen tcell.Screen, lm *layout.LayoutManager) {
	sw, sh := lm.GetLayout().CalculateDimensions(lm.GetLayout().MinWidth, lm.GetLayout().MinHeight)
	d.Rect.X = (sw - d.Width) / 2
	d.Rect.Y = (sh - d.Height) / 2
	d.Rect.Width = d.Width
	d.Rect.Height = d.Height

	style := tcell.StyleDefault.Background(tcell.ColorBlack).Foreground(tcell.ColorWhite)
	for y := 0; y < d.Height; y++ {
		for x := 0; x < d.Width; x++ {
			screen.SetContent(d.Rect.X+x, d.Rect.Y+y, ' ', nil, style)
		}
	}
	for i, r := range d.Title {
		if i < d.Width {
			screen.SetContent(d.Rect.X+i, d.Rect.Y, r, nil, style)
		}
	}
}

func (d *Dialog) Focus()                           { d.focused = true }
func (d *Dialog) Blur()                            { d.focused = false }
func (d *Dialog) IsFocused() bool                  { return d.focused }
func (d *Dialog) HandleKey(ev *tcell.EventKey)     {}
func (d *Dialog) HandleMouse(ev *tcell.EventMouse) {}
