package dialog

import (
	"path/filepath"

	"github.com/gdamore/tcell/v2"
)

type NewFileDialog struct {
	*Dialog
	Folder string
}

func NewNewFileDialog(folder string, onSubmit func(string), onCancel func()) *NewFileDialog {
	d := &NewFileDialog{
		Folder: folder,
	}

	dialog := NewDialog(
		"Create New File",
		"",
		40,
		7,
		true,
		func(_ string) {
			fullPath := filepath.Join(folder, string(d.input))
			onSubmit(fullPath)
		},
		onCancel,
		nil,
	)

	// Custom key handling to enable typing
	dialog.HandleKeyFunc = func(dlg *Dialog, ev *tcell.EventKey) {
		switch ev.Key() {
		case tcell.KeyEnter:
			if onSubmit != nil {
				onSubmit(string(dlg.input))
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
	}

	d.Dialog = dialog
	return d
}
