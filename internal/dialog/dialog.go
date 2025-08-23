package dialog

import (
	"github.com/gdamore/tcell/v2"
)

type Dialog struct {
	X, Y, Width, Height int
	title               string
	description         string
	input               []rune
	cursor              int
	focused             bool
	onSubmit            func(name string)
	onCancel            func(string)

	// Scroll offset for input longer than width
	scrollX int

	// Focus restoration
	restoreFocus func()

	HasInput bool
}

// NewDialog creates a dialog
func NewDialog(title string, description string, w, h int, onSubmit, onCancel func(string), restoreFocus func()) *Dialog {
	return &Dialog{
		Width:        w,
		Height:       h,
		title:        title,
		description:  description,
		HasInput:     true,
		onSubmit:     onSubmit,
		onCancel:     onCancel,
		restoreFocus: restoreFocus,
	}
}

func (d *Dialog) SetFocus(f bool) {
	d.focused = f
	if !f && d.restoreFocus != nil {
		d.restoreFocus()
	}
}

func (d *Dialog) IsFocused() bool {
	return d.focused
}

// Center the dialog dynamically on screen
func (d *Dialog) Center(screen tcell.Screen) {
	sw, sh := screen.Size()
	d.X = (sw - d.Width) / 2
	d.Y = (sh - d.Height) / 2
}

// Handle keyboard input
func (d *Dialog) HandleKey(ev *tcell.EventKey) {
	if !d.focused {
		return
	}

	switch ev.Key() {
	case tcell.KeyEnter:
		if d.onSubmit != nil {
			d.onSubmit(string(d.input))
		}
	case tcell.KeyEsc:
		if d.onCancel != nil {
			d.onCancel("")
		}
	case tcell.KeyBackspace, tcell.KeyBackspace2:
		if d.cursor > 0 {
			d.input = append(d.input[:d.cursor-1], d.input[d.cursor:]...)
			d.cursor--
			if d.scrollX > 0 && d.cursor < d.scrollX {
				d.scrollX--
			}
		}
	case tcell.KeyLeft:
		if d.cursor > 0 {
			d.cursor--
			if d.cursor < d.scrollX {
				d.scrollX--
			}
		}
	case tcell.KeyRight:
		if d.cursor < len(d.input) {
			d.cursor++
			if d.cursor-d.scrollX >= d.Width-2 {
				d.scrollX++
			}
		}
	default:
		if ev.Rune() != 0 {
			d.input = append(d.input[:d.cursor], append([]rune{ev.Rune()}, d.input[d.cursor:]...)...)
			d.cursor++
			if d.cursor-d.scrollX >= d.Width-2 {
				d.scrollX++
			}
		}
	}
}

// Draw dialog with nice border, background, title, and scrolling input
func (d *Dialog) Draw(s tcell.Screen) {
	bgStyle := tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.NewRGBColor(30, 30, 30))
	borderStyle := tcell.StyleDefault.Foreground(tcell.NewRGBColor(200, 200, 200)).Background(tcell.NewRGBColor(30, 30, 30))
	titleStyle := tcell.StyleDefault.Foreground(tcell.ColorYellow).Background(tcell.NewRGBColor(30, 30, 30))
	descStyle := tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.NewRGBColor(30, 30, 30))

	// Draw border with corners
	for row := 0; row < d.Height; row++ {
		for col := 0; col < d.Width; col++ {
			var ch rune = ' '
			switch {
			case row == 0 && col == 0:
				ch = '┌'
			case row == 0 && col == d.Width-1:
				ch = '┐'
			case row == d.Height-1 && col == 0:
				ch = '└'
			case row == d.Height-1 && col == d.Width-1:
				ch = '┘'
			case row == 0 || row == d.Height-1:
				ch = '─'
			case col == 0 || col == d.Width-1:
				ch = '│'
			}
			s.SetContent(d.X+col, d.Y+row, ch, nil, borderStyle)
		}
	}

	// Draw title
	for i, r := range d.title {
		if i >= d.Width-4 { // leave space for corners
			break
		}
		s.SetContent(d.X+2+i, d.Y, r, nil, titleStyle)
	}

	// Draw description
	for i, r := range d.description {
		if i >= d.Width-4 {
			break
		}
		s.SetContent(d.X+2+i, d.Y+2, r, nil, descStyle)
	}

	// Draw input area
	if d.HasInput {
		for i := 0; i < d.Width-2 && d.scrollX+i < len(d.input); i++ {
			s.SetContent(d.X+1+i, d.Y+4, d.input[d.scrollX+i], nil, bgStyle)
		}

		if d.focused {
			cursorPos := d.cursor - d.scrollX
			if cursorPos >= 0 && cursorPos < d.Width-2 {
				s.ShowCursor(d.X+1+cursorPos, d.Y+4)
			} else {
				s.HideCursor()
			}
		} else {
			s.HideCursor()
		}
	} else {
		s.HideCursor()
	}
}

func (d *Dialog) HandleMouse(ev *tcell.EventMouse) {
	// Ignore mouse events for now
}
