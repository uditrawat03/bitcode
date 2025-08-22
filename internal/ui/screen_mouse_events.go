package ui

import "github.com/gdamore/tcell/v2"

func (sm *ScreenManager) HandleMouse(ev *tcell.EventMouse) {
	if sm.dialog != nil {
		sm.dialog.HandleMouse(ev)
		return
	}

	// only delegate
	for _, comp := range sm.focusOrder {
		comp.HandleMouse(ev)
	}
}
