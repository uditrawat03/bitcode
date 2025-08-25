package dialog

import (
	"github.com/gdamore/tcell/v2"
	"github.com/uditrawat03/bitcode/internal/core"
)

type Dialog struct {
	core.BaseComponent
	title    string
	message  string
	options  []string
	selected int
	focused  bool
	onSelect func(idx int)
	onCancel func()
}

func NewDialog(title, message string, options []string, onSelect func(int), onCancel func()) *Dialog {
	return &Dialog{
		title:    title,
		message:  message,
		options:  options,
		selected: 0,
		onSelect: onSelect,
		onCancel: onCancel,
	}
}

func (d *Dialog) Focus()          { d.focused = true }
func (d *Dialog) Blur()           { d.focused = false }
func (d *Dialog) IsFocused() bool { return d.focused }

func (d *Dialog) Render(screen tcell.Screen, _ interface{}) {
	bg := tcell.ColorDarkGray
	style := tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(bg)
	titleStyle := tcell.StyleDefault.Foreground(tcell.ColorYellow).Background(bg).Bold(true)
	selectedStyle := tcell.StyleDefault.Foreground(tcell.ColorBlack).Background(tcell.ColorWhite).Bold(true)

	// Draw dialog box border
	for y := d.Rect.Y; y < d.Rect.Y+d.Rect.Height; y++ {
		for x := d.Rect.X; x < d.Rect.X+d.Rect.Width; x++ {
			screen.SetContent(x, y, ' ', nil, style)
		}
	}

	// Title
	for i, r := range d.title {
		if i >= d.Rect.Width {
			break
		}
		screen.SetContent(d.Rect.X+i+1, d.Rect.Y, r, nil, titleStyle)
	}

	// Message
	for i, r := range d.message {
		if i >= d.Rect.Width-2 {
			break
		}
		screen.SetContent(d.Rect.X+1+i, d.Rect.Y+2, r, nil, style)
	}

	// Options
	for idx, opt := range d.options {
		optStyle := style
		if idx == d.selected && d.focused {
			optStyle = selectedStyle
		}
		for i, r := range opt {
			if i >= d.Rect.Width-2 {
				break
			}
			screen.SetContent(d.Rect.X+1+i, d.Rect.Y+4+idx, r, nil, optStyle)
		}
	}
}

func (d *Dialog) HandleKey(ev *tcell.EventKey) {
	switch ev.Key() {
	case tcell.KeyUp:
		if d.selected > 0 {
			d.selected--
		}
	case tcell.KeyDown:
		if d.selected < len(d.options)-1 {
			d.selected++
		}
	case tcell.KeyEnter:
		if d.onSelect != nil {
			d.onSelect(d.selected)
		}
	case tcell.KeyEscape:
		if d.onCancel != nil {
			d.onCancel()
		}
	}
}

func (t *Dialog) HandleMouse(ev *tcell.EventMouse) {}
