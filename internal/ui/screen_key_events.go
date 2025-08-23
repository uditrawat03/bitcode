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

	if ev.Key() == tcell.KeyCtrlN {
		sm.OpenNewFileDialog()
		return
	}

	if ev.Key() == tcell.KeyCtrlP {
		sm.openFileSearchDialog()
		return
	}

	if sm.sidebar.IsFocused() && ev.Key() == tcell.KeyDelete {
		node := sm.sidebar.GetSelectedNode()
		if node != nil {
			sm.ConfirmDeleteNode(node)
		}
		return
	}

	// Pass to focused component
	if len(sm.focusOrder) == 0 {
		return
	}

	if sm.tooltip.Visible {
		switch ev.Key() {
		case tcell.KeyUp:
			sm.tooltip.Prev()
			return
		case tcell.KeyDown:
			sm.tooltip.Next()
			return
		case tcell.KeyEnter:
			sm.tooltip.Apply(func(item string) {
				// TODO: hook this to actual action in editor
				sm.editor.GetBuffer().InsertRune('\n')
				for _, r := range item {
					sm.editor.GetBuffer().InsertRune(r)
				}
			})
			return
		case tcell.KeyEsc:
			sm.tooltip.Close()
			return
		default:
			sm.tooltip.Close()
			return
		}
	}

	sm.focusOrder[sm.focusedIdx].HandleKey(ev)
}
