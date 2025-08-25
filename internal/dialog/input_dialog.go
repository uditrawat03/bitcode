package dialog

import (
	"os"
	"path/filepath"

	"github.com/gdamore/tcell/v2"
	"github.com/uditrawat03/bitcode/internal/core"
	"github.com/uditrawat03/bitcode/internal/layout"
)

type InputDialog struct {
	core.BaseComponent
	title         string
	prompt        string
	input         []rune
	focused       bool
	cursorPos     int
	onConfirm     func(string)
	onCancel      func()
	ValidationMsg string
	Folder        string
}

// Constructor
func NewInputDialog(folder, title, prompt string, onConfirm func(string), onCancel func()) *InputDialog {
	return &InputDialog{
		title:     title,
		prompt:    prompt,
		input:     []rune{},
		onConfirm: onConfirm,
		onCancel:  onCancel,
		Folder:    folder,
	}
}

// Render the dialog
func (dlg *InputDialog) Render(screen tcell.Screen, lm *layout.LayoutManager) {
	x, y := dlg.Rect.X, dlg.Rect.Y
	w, h := dlg.Rect.Width, dlg.Rect.Height

	titleFg := tcell.ColorWhite
	borderColor := tcell.NewRGBColor(180, 180, 180)
	promptFg := tcell.ColorWhite
	inputFg := tcell.ColorWhite
	inputBg := tcell.StyleDefault.Background(tcell.ColorReset)
	warnStyle := tcell.StyleDefault.Foreground(tcell.ColorRed).Background(tcell.ColorReset)

	// Clear area
	clearStyle := tcell.StyleDefault.Background(tcell.ColorReset)
	for row := y; row < y+h; row++ {
		for col := x; col < x+w; col++ {
			screen.SetContent(col, row, ' ', nil, clearStyle)
		}
	}

	// Draw border
	for col := x; col < x+w; col++ {
		screen.SetContent(col, y, '─', nil, tcell.StyleDefault.Foreground(borderColor))
		screen.SetContent(col, y+h-1, '─', nil, tcell.StyleDefault.Foreground(borderColor))
	}
	for row := y; row < y+h; row++ {
		screen.SetContent(x, row, '│', nil, tcell.StyleDefault.Foreground(borderColor))
		screen.SetContent(x+w-1, row, '│', nil, tcell.StyleDefault.Foreground(borderColor))
	}
	screen.SetContent(x, y, '┌', nil, tcell.StyleDefault.Foreground(borderColor))
	screen.SetContent(x+w-1, y, '┐', nil, tcell.StyleDefault.Foreground(borderColor))
	screen.SetContent(x, y+h-1, '└', nil, tcell.StyleDefault.Foreground(borderColor))
	screen.SetContent(x+w-1, y+h-1, '┘', nil, tcell.StyleDefault.Foreground(borderColor))

	// Title
	titleY := y + 1
	for i, r := range dlg.title {
		if i >= w-2 {
			break
		}
		screen.SetContent(x+1+i, titleY, r, nil, tcell.StyleDefault.Foreground(titleFg))
	}

	// Title bottom border
	for col := x + 1; col < x+w-1; col++ {
		screen.SetContent(col, titleY+1, '─', nil, tcell.StyleDefault.Foreground(borderColor))
	}

	// Prompt
	promptY := titleY + 2
	for i, r := range dlg.prompt {
		if i >= w-2 {
			break
		}
		screen.SetContent(x+1+i, promptY, r, nil, tcell.StyleDefault.Foreground(promptFg))
	}

	// Input field
	inputY := promptY + 1
	for i := 0; i < w-2; i++ {
		screen.SetContent(x+1+i, inputY, ' ', nil, inputBg)
	}
	for i, r := range dlg.input {
		if i >= w-2 {
			break
		}
		screen.SetContent(x+1+i, inputY, r, nil, tcell.StyleDefault.Foreground(inputFg))
	}

	// Cursor
	if dlg.focused {
		cx := x + 1 + dlg.cursorPos
		cy := inputY
		if cx >= x+1 && cx < x+w-1 {
			screen.ShowCursor(cx, cy)
		} else {
			screen.HideCursor()
		}
	} else {
		screen.HideCursor()
	}

	// Validation message (only set when Enter is pressed)
	if dlg.ValidationMsg != "" {
		for i, r := range dlg.ValidationMsg {
			if i >= w-2 {
				break
			}
			screen.SetContent(x+1+i, inputY+1, r, nil, warnStyle)
		}
	}
}

// HandleKey updated: only validate on Enter
func (dlg *InputDialog) HandleKey(ev *tcell.EventKey) {
	if !dlg.focused {
		return
	}

	switch ev.Key() {
	case tcell.KeyEnter:
		fullPath := filepath.Join(dlg.Folder, string(dlg.input))
		// Example validations:
		if len(dlg.input) == 0 {
			dlg.ValidationMsg = "⚠ Input cannot be empty!"
			return
		}
		if _, err := os.Stat(fullPath); !os.IsNotExist(err) {
			dlg.ValidationMsg = "⚠ File already exists!"
			return
		}
		// Validation passed
		dlg.ValidationMsg = ""
		if dlg.onConfirm != nil {
			dlg.onConfirm(fullPath)
		}
	case tcell.KeyEscape:
		if dlg.onCancel != nil {
			dlg.onCancel()
		}
	case tcell.KeyBackspace, tcell.KeyBackspace2:
		if dlg.cursorPos > 0 && len(dlg.input) > 0 {
			dlg.input = append(dlg.input[:dlg.cursorPos-1], dlg.input[dlg.cursorPos:]...)
			dlg.cursorPos--
		}
	case tcell.KeyLeft:
		if dlg.cursorPos > 0 {
			dlg.cursorPos--
		}
	case tcell.KeyRight:
		if dlg.cursorPos < len(dlg.input) {
			dlg.cursorPos++
		}
	default:
		r := ev.Rune()
		if r != 0 {
			before := dlg.input[:dlg.cursorPos]
			after := dlg.input[dlg.cursorPos:]
			dlg.input = append(before, append([]rune{r}, after...)...)
			dlg.cursorPos++
		}
	}
}

func (dlg *InputDialog) Focus()          { dlg.focused = true }
func (dlg *InputDialog) Blur()           { dlg.focused = false }
func (dlg *InputDialog) IsFocused() bool { return dlg.focused }

func (dlg *InputDialog) HandleMouse(ev *tcell.EventMouse) {}
