package dialog

import (
	"github.com/gdamore/tcell/v2"
)

// Dialog is a generic modal dialog
type Dialog struct {
	X, Y, Width, Height int
	title, description  string
	input               []rune
	cursor              int
	focused             bool
	scrollX             int
	HasInput            bool

	onSubmit func(string)
	onCancel func()

	// Optional custom drawing
	CustomDraw    func(d *Dialog, s tcell.Screen)
	HandleKeyFunc func(d *Dialog, ev *tcell.EventKey) // override key handling

	restoreFocus func()
}

// NewDialog creates a new dialog
func NewDialog(title, description string, w, h int, hasInput bool, onSubmit func(string), onCancel func(), restoreFocus func()) *Dialog {
	return &Dialog{
		title:        title,
		description:  description,
		Width:        w,
		Height:       h,
		HasInput:     hasInput,
		onSubmit:     onSubmit,
		onCancel:     onCancel,
		restoreFocus: restoreFocus,
	}
}

// SetFocus sets focus
func (d *Dialog) SetFocus(f bool) {
	d.focused = f
	if !f && d.restoreFocus != nil {
		d.restoreFocus()
	}
}

func (d *Dialog) IsFocused() bool {
	return d.focused
}

// Center positions dialog in the screen
func (d *Dialog) Center(s tcell.Screen) {
	sw, sh := s.Size()
	d.X = (sw - d.Width) / 2
	d.Y = (sh - d.Height) / 2
}

// HandleKey handles keyboard input (or delegates to HandleKeyFunc)
func (d *Dialog) HandleKey(ev *tcell.EventKey) {
	if !d.focused {
		return
	}
	if d.HandleKeyFunc != nil {
		d.HandleKeyFunc(d, ev)
		return
	}

	switch ev.Key() {
	case tcell.KeyEnter:
		if d.onSubmit != nil {
			if d.HasInput {
				d.onSubmit(string(d.input))
			} else {
				d.onSubmit("")
			}
		}
	case tcell.KeyEsc:
		if d.onCancel != nil {
			d.onCancel()
		}
	case tcell.KeyBackspace, tcell.KeyBackspace2:
		if d.cursor > 0 {
			d.input = append(d.input[:d.cursor-1], d.input[d.cursor:]...)
			d.cursor--
			if d.cursor < d.scrollX {
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
		if ev.Rune() != 0 && d.HasInput {
			d.input = append(d.input[:d.cursor], append([]rune{ev.Rune()}, d.input[d.cursor:]...)...)
			d.cursor++
			if d.cursor-d.scrollX >= d.Width-2 {
				d.scrollX++
			}
		}
	}
}

// Draw renders the dialog
func (d *Dialog) Draw(s tcell.Screen) {
	if d.CustomDraw != nil {
		d.CustomDraw(d, s)
		return
	}

	bgStyle := tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.NewRGBColor(30, 30, 30))
	borderStyle := tcell.StyleDefault.Foreground(tcell.NewRGBColor(200, 200, 200)).Background(tcell.NewRGBColor(30, 30, 30))
	titleStyle := tcell.StyleDefault.Foreground(tcell.ColorYellow).Background(tcell.NewRGBColor(30, 30, 30))
	descStyle := tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.NewRGBColor(30, 30, 30))

	// Draw border
	for row := 0; row < d.Height; row++ {
		for col := 0; col < d.Width; col++ {
			ch := ' '
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

	// Title
	for i, r := range d.title {
		if i >= d.Width-4 {
			break
		}
		s.SetContent(d.X+2+i, d.Y, r, nil, titleStyle)
	}

	// Description
	for i, r := range d.description {
		if i >= d.Width-4 {
			break
		}
		s.SetContent(d.X+2+i, d.Y+2, r, nil, descStyle)
	}

	// Input line
	if d.HasInput {
		inputLine := "> " + string(d.input)
		for i, r := range inputLine {
			if i >= d.Width-2 {
				break
			}
			s.SetContent(d.X+1+i, d.Y+d.Height-1, r, nil, bgStyle)
		}
		// Cursor
		if d.focused {
			cursorPos := len("> ") + d.cursor
			if cursorPos < d.Width-2 {
				s.ShowCursor(d.X+1+cursorPos, d.Y+d.Height-1)
			} else {
				s.HideCursor()
			}
		}
	} else {
		s.HideCursor()
	}
}

// HandleMouse can be extended
func (d *Dialog) HandleMouse(ev *tcell.EventMouse) {}
