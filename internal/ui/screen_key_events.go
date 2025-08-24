package ui

// // func (sm *ScreenManager) HandleKey(ev *tcell.EventKey) {
// // 	if sm.dialog != nil {
// // 		sm.dialog.HandleKey(ev)
// // 		return
// // 	}
// // 	if len(sm.focusOrder) == 0 {
// // 		return
// // 	}
// // 	sm.focusOrder[sm.focusedIdx].HandleKey(ev)
// // }

// func (sm *ScreenManager) HandleKey(ev *tcell.EventKey) {
// 	// Dialog active
// 	if sm.dialog != nil {
// 		sm.dialog.HandleKey(ev)
// 		return
// 	}

// 	if ev.Key() == tcell.KeyCtrlN {
// 		sm.OpenNewFileDialog()
// 		return
// 	}

// 	if ev.Key() == tcell.KeyCtrlP {
// 		sm.openFileSearchDialog()
// 		return
// 	}

// 	if sm.sidebar.IsFocused() && ev.Key() == tcell.KeyDelete {
// 		node := sm.sidebar.GetSelectedNode()
// 		if node != nil {
// 			sm.ConfirmDeleteNode(node)
// 		}
// 		return
// 	}

// 	// Pass to focused component
// 	if len(sm.focusOrder) == 0 {
// 		return
// 	}

// 	if sm.tooltip.Visible {
// 		switch ev.Key() {
// 		case tcell.KeyUp:
// 			sm.tooltip.Prev()
// 			return
// 		case tcell.KeyDown:
// 			sm.tooltip.Next()
// 			return
// 		case tcell.KeyEnter:
// 			sm.tooltip.Apply(func(item string) {
// 				buf := sm.editor.GetBuffer()
// 				if buf == nil || buf.CursorY < 0 || buf.CursorY >= len(buf.Content) {
// 					return
// 				}

// 				// Get current line
// 				line := buf.Content[buf.CursorY]
// 				cursorX := buf.CursorX
// 				if cursorX > len(line) {
// 					cursorX = len(line)
// 				}

// 				// Find start of current word
// 				start := cursorX
// 				for start > 0 {
// 					ch := line[start-1]
// 					if ch == ' ' || ch == '\t' || ch == '\n' {
// 						break
// 					}
// 					start--
// 				}

// 				// Delete existing word safely
// 				buf.DeleteSelection(start, buf.CursorY, cursorX, buf.CursorY)

// 				// Insert selected completion
// 				for _, r := range item {
// 					buf.InsertRune(r)
// 				}
// 			})
// 			return

// 		case tcell.KeyEsc:
// 			sm.tooltip.Close()
// 			return
// 		default:
// 			sm.tooltip.Close()
// 			return
// 		}
// 	}

// 	sm.focusOrder[sm.focusedIdx].HandleKey(ev)
// }
