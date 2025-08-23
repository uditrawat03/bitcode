package dialog

import (
	"log"
	"os"

	"github.com/gdamore/tcell/v2"
)

type DeleteNodeDialog struct {
	*Dialog
	Path     string
	OnDelete func(path string)
}

func NewDeleteNodeDialog(path string, onDelete func(path string), restoreFocus func()) *DeleteNodeDialog {
	d := &DeleteNodeDialog{
		Dialog: NewDialog(
			"Delete "+path+"?",
			"Press Enter to confirm, Esc to cancel",
			40, 5,
			false, // no input
			nil,
			func() {},
			restoreFocus,
		),
		Path:     path,
		OnDelete: onDelete,
	}

	d.Dialog.HandleKeyFunc = func(dlg *Dialog, ev *tcell.EventKey) {
		switch ev.Key() {
		case tcell.KeyEnter:
			var err error
			if info, e := os.Stat(path); e == nil && info.IsDir() {
				err = os.RemoveAll(path)
			} else {
				err = os.Remove(path)
			}
			if err != nil {
				log.Println("Delete failed:", err)
			} else if d.OnDelete != nil {
				d.OnDelete(path)
			}
		case tcell.KeyEsc:
			if dlg.onCancel != nil {
				dlg.onCancel()
			}
		}
	}

	return d
}
