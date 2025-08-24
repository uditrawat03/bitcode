package dialog

import (
	"os"
	"path/filepath"

	"github.com/gdamore/tcell/v2"
)

type NewFileDialog struct {
	*Dialog
	Folder        string
	ValidationMsg string
}

// NewNewFileDialog creates a new dialog for creating files
func NewNewFileDialog(folder string, onSubmit func(string), onCancel func()) *NewFileDialog {
	d := &NewFileDialog{
		Folder: folder,
	}

	dialog := NewDialog(
		"Create New File",
		"",
		50, // width
		7,  // height (enough for title + input + validation)
		true,
		func(_ string) {
			fullPath := filepath.Join(folder, string(d.input))
			onSubmit(fullPath)
		},
		onCancel,
		nil,
	)

	// Key handling to enable typing + real-time validation
	dialog.HandleKeyFunc = func(dlg *Dialog, ev *tcell.EventKey) {
		switch ev.Key() {
		case tcell.KeyEnter:
			fullPath := filepath.Join(folder, string(dlg.input))
			if _, err := os.Stat(fullPath); os.IsNotExist(err) && len(dlg.input) > 0 {
				if onSubmit != nil {
					onSubmit(fullPath)
				}
			}
		case tcell.KeyEsc:
			if onCancel != nil {
				onCancel()
			}
		case tcell.KeyBackspace, tcell.KeyBackspace2:
			if dlg.cursor > 0 {
				dlg.input = append(dlg.input[:dlg.cursor-1], dlg.input[dlg.cursor:]...)
				dlg.cursor--
				if dlg.cursor < dlg.scrollX {
					dlg.scrollX--
				}
			}
		case tcell.KeyLeft:
			if dlg.cursor > 0 {
				dlg.cursor--
				if dlg.cursor < dlg.scrollX {
					dlg.scrollX--
				}
			}
		case tcell.KeyRight:
			if dlg.cursor < len(dlg.input) {
				dlg.cursor++
				if dlg.cursor-dlg.scrollX >= dlg.Width-2 {
					dlg.scrollX++
				}
			}
		default:
			if ev.Rune() != 0 {
				dlg.input = append(dlg.input[:dlg.cursor], append([]rune{ev.Rune()}, dlg.input[dlg.cursor:]...)...)
				dlg.cursor++
				if dlg.cursor-dlg.scrollX >= dlg.Width-2 {
					dlg.scrollX++
				}
			}
		}

		// Update validation message
		fullPath := filepath.Join(folder, string(dlg.input))
		if _, err := os.Stat(fullPath); !os.IsNotExist(err) && len(dlg.input) > 0 {
			d.ValidationMsg = "⚠ File already exists!"
		} else {
			d.ValidationMsg = ""
		}
	}

	// Use CustomDraw for proper top input and validation
	dialog.CustomDraw = d.draw
	d.Dialog = dialog
	return d
}

// draw renders the NewFileDialog with top input and validation
func (d *NewFileDialog) draw(dialog *Dialog, s tcell.Screen) {
	borderStyle := tcell.StyleDefault.Foreground(tcell.NewRGBColor(200, 200, 200)).Background(tcell.NewRGBColor(30, 30, 30))
	titleStyle := tcell.StyleDefault.Foreground(tcell.ColorYellow).Background(tcell.NewRGBColor(30, 30, 30))
	inputBoxStyle := tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.NewRGBColor(50, 50, 50))
	warnStyle := tcell.StyleDefault.Foreground(tcell.ColorRed).Background(tcell.NewRGBColor(30, 30, 30))

	// Draw border
	for y := 0; y < dialog.Height; y++ {
		for x := 0; x < dialog.Width; x++ {
			ch := ' '
			switch {
			case y == 0 && x == 0:
				ch = '┌'
			case y == 0 && x == dialog.Width-1:
				ch = '┐'
			case y == dialog.Height-1 && x == 0:
				ch = '└'
			case y == dialog.Height-1 && x == dialog.Width-1:
				ch = '┘'
			case y == 0 || y == dialog.Height-1:
				ch = '─'
			case x == 0 || x == dialog.Width-1:
				ch = '│'
			}
			s.SetContent(dialog.X+x, dialog.Y+y, ch, nil, borderStyle)
		}
	}

	// Title
	for i, r := range dialog.title {
		if i >= dialog.Width-4 {
			break
		}
		s.SetContent(dialog.X+2+i, dialog.Y, r, nil, titleStyle)
	}

	// Input box (top, under title)
	inputY := dialog.Y + 2
	for x := 1; x < dialog.Width-1; x++ {
		s.SetContent(dialog.X+x, inputY, ' ', nil, inputBoxStyle)
	}

	// Typed content
	inputLine := string(dialog.input)
	for i := 0; i < len(inputLine) && i < dialog.Width-2; i++ {
		s.SetContent(dialog.X+1+i, inputY, rune(inputLine[i]), nil, inputBoxStyle)
	}

	// Show cursor
	if dialog.focused {
		cursorPos := dialog.cursor
		if cursorPos >= dialog.Width-2 {
			cursorPos = dialog.Width - 3
		}
		s.ShowCursor(dialog.X+1+cursorPos, inputY)
	} else {
		s.HideCursor()
	}

	// Validation message below input
	if d.ValidationMsg != "" {
		for i, r := range d.ValidationMsg {
			if i >= dialog.Width-2 {
				break
			}
			s.SetContent(dialog.X+1+i, inputY+1, r, nil, warnStyle)
		}
	}
}
