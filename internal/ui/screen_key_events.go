package ui

import (
	"github.com/gdamore/tcell/v2"
)

// func (sm *ScreenManager) HandleKey(ev *tcell.EventKey) {
// 	if sm.dialog != nil {
// 		sm.dialog.HandleKey(ev)
// 		return
// 	}
// 	if len(sm.focusOrder) == 0 {
// 		return
// 	}
// 	sm.focusOrder[sm.focusedIdx].HandleKey(ev)
// }

func (sm *ScreenManager) HandleKey(ev *tcell.EventKey) {
	// Dialog active
	if sm.dialog != nil {
		sm.dialog.HandleKey(ev)
		return
	}

	// Ctrl+N → New File Dialog
	if ev.Key() == tcell.KeyCtrlN {
		sm.openNewFileDialog()
		return
	}

	if sm.sidebar.IsFocused() && ev.Key() == tcell.KeyDelete {
		node := sm.sidebar.GetSelectedNode()
		if node != nil {
			sm.confirmDeleteNode(node)
		}
		return
	}

	// Pass to focused component
	if len(sm.focusOrder) == 0 {
		return
	}
	sm.focusOrder[sm.focusedIdx].HandleKey(ev)
}
