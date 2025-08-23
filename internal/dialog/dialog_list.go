package dialog

import (
	"strings"

	"github.com/gdamore/tcell/v2"
)

// ListDialog is a Neovim-style file list dialog
type ListDialog struct {
	*Dialog
	Items     []string
	Selected  int
	MaxHeight int
	OnSelect  func(string)
}

// NewListDialog creates a list selection dialog
func NewListDialog(title string, items []string, maxHeight int, onSelect func(string), restoreFocus func()) *ListDialog {
	height := len(items) + 2
	if maxHeight > 0 && height > maxHeight+2 {
		height = maxHeight + 2
	}
	width := 0
	for _, item := range items {
		if len(item) > width {
			width = len(item)
		}
	}
	width += 2

	ld := &ListDialog{
		Dialog:    NewDialog(title, "", width, height, true, nil, nil, restoreFocus),
		Items:     items,
		Selected:  0,
		MaxHeight: maxHeight,
		OnSelect:  onSelect,
	}

	// Override drawing and input
	ld.Dialog.CustomDraw = ld.DrawList
	ld.Dialog.HandleKeyFunc = func(d *Dialog, ev *tcell.EventKey) {
		ld.HandleKey(ev)
	}

	return ld
}

// HandleKey handles navigation in the list
func (ld *ListDialog) HandleKey(ev *tcell.EventKey) {
	switch ev.Key() {
	case tcell.KeyUp:
		ld.Selected--
		if ld.Selected < 0 {
			ld.Selected = len(ld.Items) - 1
		}
	case tcell.KeyDown:
		ld.Selected++
		if ld.Selected >= len(ld.Items) {
			ld.Selected = 0
		}
	case tcell.KeyPgUp:
		ld.Selected -= ld.MaxHeight
		if ld.Selected < 0 {
			ld.Selected = 0
		}
	case tcell.KeyPgDn:
		ld.Selected += ld.MaxHeight
		if ld.Selected >= len(ld.Items) {
			ld.Selected = len(ld.Items) - 1
		}
	case tcell.KeyEnter:
		if ld.OnSelect != nil && len(ld.Items) > 0 {
			ld.OnSelect(ld.Items[ld.Selected])
		}
		if ld.restoreFocus != nil {
			ld.restoreFocus()
		}
	case tcell.KeyEsc:
		if ld.restoreFocus != nil {
			ld.restoreFocus()
		}
	default:
		// handle text input for filtering
		if ev.Rune() != 0 {
			ld.input = append(ld.input[:ld.cursor], append([]rune{ev.Rune()}, ld.input[ld.cursor:]...)...)
			ld.cursor++
		} else if ev.Key() == tcell.KeyBackspace || ev.Key() == tcell.KeyBackspace2 {
			if ld.cursor > 0 {
				ld.input = append(ld.input[:ld.cursor-1], ld.input[ld.cursor:]...)
				ld.cursor--
			}
		}
		// filtering items
		filtered := make([]string, 0)
		query := string(ld.input)
		for _, it := range ld.Items {
			if query == "" || containsIgnoreCase(it, query) {
				filtered = append(filtered, it)
			}
		}
		ld.Items = filtered
		if ld.Selected >= len(ld.Items) {
			ld.Selected = len(ld.Items) - 1
		}
	}
}

// DrawList draws the list dialog
func (ld *ListDialog) DrawList(d *Dialog, s tcell.Screen) {
	bgStyle := tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.NewRGBColor(30, 30, 30))
	borderStyle := tcell.StyleDefault.Foreground(tcell.NewRGBColor(200, 200, 200)).Background(tcell.NewRGBColor(30, 30, 30))
	titleStyle := tcell.StyleDefault.Foreground(tcell.ColorYellow).Background(tcell.NewRGBColor(30, 30, 30))
	selectedStyle := tcell.StyleDefault.Foreground(tcell.ColorBlack).Background(tcell.ColorWhite)

	// Border
	for row := 0; row < ld.Height; row++ {
		for col := 0; col < ld.Width; col++ {
			ch := ' '
			switch {
			case row == 0 && col == 0:
				ch = '┌'
			case row == 0 && col == ld.Width-1:
				ch = '┐'
			case row == ld.Height-1 && col == 0:
				ch = '└'
			case row == ld.Height-1 && col == ld.Width-1:
				ch = '┘'
			case row == 0 || row == ld.Height-1:
				ch = '─'
			case col == 0 || col == ld.Width-1:
				ch = '│'
			}
			s.SetContent(ld.X+col, ld.Y+row, ch, nil, borderStyle)
		}
	}

	// Title
	for i, r := range ld.title {
		if i >= ld.Width-4 {
			break
		}
		s.SetContent(ld.X+2+i, ld.Y, r, nil, titleStyle)
	}

	// Items
	visible := ld.Height - 2
	if ld.MaxHeight > 0 && visible > ld.MaxHeight {
		visible = ld.MaxHeight
	}
	scrollTop := (ld.Selected / visible) * visible
	if scrollTop+visible > len(ld.Items) {
		scrollTop = len(ld.Items) - visible
	}
	if scrollTop < 0 {
		scrollTop = 0
	}

	for i := 0; i < visible && i+scrollTop < len(ld.Items); i++ {
		item := ld.Items[i+scrollTop]
		style := bgStyle
		if i+scrollTop == ld.Selected {
			style = selectedStyle
		}
		for j, r := range item {
			if j+1 >= ld.Width-1 {
				break
			}
			s.SetContent(ld.X+1+j, ld.Y+1+i, r, nil, style)
		}
		for j := len(item); j < ld.Width-2; j++ {
			s.SetContent(ld.X+1+j, ld.Y+1+i, ' ', nil, style)
		}
	}

	// Input line at bottom
	inputLine := "> " + string(ld.input)
	for i, r := range inputLine {
		if i >= ld.Width-2 {
			break
		}
		s.SetContent(ld.X+1+i, ld.Y+ld.Height-1, r, nil, bgStyle)
	}
	if ld.focused {
		cursorPos := len("> ") + ld.cursor
		if cursorPos < ld.Width-2 {
			s.ShowCursor(ld.X+1+cursorPos, ld.Y+ld.Height-1)
		} else {
			s.HideCursor()
		}
	}
}

// helper
func containsIgnoreCase(s, sub string) bool {
	s, sub = strings.ToLower(s), strings.ToLower(sub)
	return strings.Contains(s, sub)
}
