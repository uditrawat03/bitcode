package dialog

import (
	"path/filepath"
	"strings"

	"github.com/gdamore/tcell/v2"
)

type FileSearchDialog struct {
	Dialog *Dialog

	cwd      string
	files    []string
	filtered []string
	selected int
}

// Create a new FileSearchDialog
func NewFileSearchDialog(cwd string, files []string, onSelect func(string), restoreFocus func(), onCancle func()) *FileSearchDialog {
	fs := &FileSearchDialog{
		cwd:      cwd,
		files:    files,
		filtered: files,
		selected: 0,
	}

	d := NewDialog(
		"Search File",
		"",
		60, // width
		20, // height
		true,
		func(_ string) {
			if len(fs.filtered) > 0 {
				onSelect(fs.filtered[fs.selected])
			}
		},
		func() {
			onCancle()
		},
		restoreFocus,
	)

	// Override key handling
	d.HandleKeyFunc = fs.handleKey
	d.CustomDraw = fs.draw
	fs.Dialog = d
	return fs
}

// Filter files based on input
func (fs *FileSearchDialog) updateFilter(input string) {
	fs.filtered = nil
	lower := strings.ToLower(input)
	for _, f := range fs.files {
		if strings.Contains(strings.ToLower(f), lower) {
			fs.filtered = append(fs.filtered, f)
		}
	}
	if fs.selected >= len(fs.filtered) {
		fs.selected = len(fs.filtered) - 1
	}
	if fs.selected < 0 {
		fs.selected = 0
	}
}

// Key handling
func (fs *FileSearchDialog) handleKey(d *Dialog, ev *tcell.EventKey) {
	switch ev.Key() {
	case tcell.KeyUp:
		if fs.selected > 0 {
			fs.selected--
		}
	case tcell.KeyDown:
		if fs.selected < len(fs.filtered)-1 {
			fs.selected++
		}
	case tcell.KeyEnter:
		if len(fs.filtered) > 0 && d.onSubmit != nil {
			d.onSubmit("")
		}
	case tcell.KeyEsc:
		if d.onCancel != nil {
			d.onCancel()
		}
	case tcell.KeyBackspace, tcell.KeyBackspace2:
		if d.cursor > 0 {
			d.input = append(d.input[:d.cursor-1], d.input[d.cursor:]...)
			d.cursor--
		}
	default:
		if ev.Rune() != 0 && d.HasInput {
			d.input = append(d.input[:d.cursor], append([]rune{ev.Rune()}, d.input[d.cursor:]...)...)
			d.cursor++
		}
	}
	fs.updateFilter(string(d.input))
}

// CustomDraw renders input + filtered list
func (fs *FileSearchDialog) draw(d *Dialog, s tcell.Screen) {
	border := tcell.StyleDefault.Background(tcell.NewRGBColor(30, 30, 30)).Foreground(tcell.NewRGBColor(200, 200, 200))
	titleStyle := tcell.StyleDefault.Foreground(tcell.ColorYellow).Background(tcell.NewRGBColor(30, 30, 30))
	inputStyle := tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.NewRGBColor(50, 50, 50))
	itemStyle := tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.NewRGBColor(30, 30, 30))
	selectedStyle := tcell.StyleDefault.Foreground(tcell.ColorBlack).Background(tcell.ColorWhite)
	boldStyle := tcell.StyleDefault.Bold(true).Foreground(tcell.ColorWhite).Background(tcell.NewRGBColor(30, 30, 30))
	selectedBold := tcell.StyleDefault.Bold(true).Foreground(tcell.ColorBlack).Background(tcell.ColorWhite)
	inputBoxStyle := tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.NewRGBColor(50, 50, 50))

	// Draw border
	for y := 0; y < d.Height; y++ {
		for x := 0; x < d.Width; x++ {
			ch := ' '
			switch {
			case y == 0 && x == 0:
				ch = '┌'
			case y == 0 && x == d.Width-1:
				ch = '┐'
			case y == d.Height-1 && x == 0:
				ch = '└'
			case y == d.Height-1 && x == d.Width-1:
				ch = '┘'
			case y == 0 || y == d.Height-1:
				ch = '─'
			case x == 0 || x == d.Width-1:
				ch = '│'
			}
			s.SetContent(d.X+x, d.Y+y, ch, nil, border)
		}
	}

	// Title
	for i, r := range d.title {
		if i >= d.Width-4 {
			break
		}
		s.SetContent(d.X+2+i, d.Y, r, nil, titleStyle)
	}

	// Input box (top, under title)
	inputY := d.Y + 2
	for x := 1; x < d.Width-1; x++ {
		s.SetContent(d.X+x, inputY, ' ', nil, inputBoxStyle)
	}

	// Input line
	inputLine := "> " + string(d.input)
	for i := 0; i < d.Width-2 && i < len(inputLine); i++ {
		s.SetContent(d.X+1+i, d.Y+2, rune(inputLine[i]), nil, inputStyle)
	}

	if d.focused {
		cursorPos := 2 + d.cursor
		if cursorPos < d.Width-2 {
			s.ShowCursor(d.X+1+cursorPos, d.Y+2)
		} else {
			s.HideCursor()
		}
	} else {
		s.HideCursor()
	}

	// Draw filtered list with full line highlight
	startY := d.Y + 4
	maxItems := d.Height - 6
	for i := 0; i < maxItems && i < len(fs.filtered); i++ {
		fullPath := fs.filtered[i]
		name := filepath.Base(fullPath)
		relPath := strings.TrimPrefix(fullPath, fs.cwd+"/")
		relPath = strings.ReplaceAll(relPath, "\\", "/") // normalize

		style := itemStyle
		nameStyle := boldStyle
		if i == fs.selected {
			style = selectedStyle
			nameStyle = selectedBold
		}

		// Fill entire line first
		for x := 1; x < d.Width-1; x++ {
			s.SetContent(d.X+x, startY+i, ' ', nil, style)
		}

		// Truncate if too long
		maxNameLen := d.Width/2 - 2
		maxPathLen := d.Width - maxNameLen - 4
		if len(name) > maxNameLen {
			name = name[:maxNameLen-1] + "…"
		}
		if len(relPath) > maxPathLen {
			relPath = "…" + relPath[len(relPath)-maxPathLen+1:]
		}

		// Print name (left)
		for j, r := range name {
			s.SetContent(d.X+1+j, startY+i, r, nil, nameStyle)
		}

		// Print path (right)
		offset := d.Width - 1 - len(relPath)
		for j, r := range relPath {
			if offset+j > d.Width-2 {
				break
			}
			s.SetContent(d.X+offset+j, startY+i, r, nil, style)
		}
	}
}
