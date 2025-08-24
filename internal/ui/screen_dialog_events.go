package ui

// import (
// 	"os"
// 	"path/filepath"

// 	"github.com/uditrawat03/bitcode/internal/dialog"
// 	"github.com/uditrawat03/bitcode/internal/treeview"
// )

// // OpenDialog opens a modal dialog
// func (sm *ScreenManager) OpenDialog(d *dialog.Dialog) {
// 	sm.dialog = d
// 	d.SetFocus(true)
// }

// // CloseDialog closes the modal dialog
// func (sm *ScreenManager) CloseDialog() {
// 	if sm.dialog != nil {
// 		sm.dialog.SetFocus(false)
// 		sm.dialog = nil
// 		sm.screen.HideCursor()
// 	}
// }

// func (sm *ScreenManager) restoreEditorFocus() {
// 	if sm.editor != nil && sm.editor.GetBuffer() != nil {
// 		sm.focusOrder[sm.focusedIdx].Blur()
// 		sm.focusedIdx = 1
// 		sm.editor.Focus()
// 	}
// }

// func (sm *ScreenManager) OpenNewFileDialog() {
// 	folder := "."
// 	if sm.sidebar.IsFocused() {
// 		node := sm.sidebar.GetSelectedNode()
// 		if node != nil {
// 			if node.IsDir {
// 				folder = node.Path
// 			} else if node.Parent != nil {
// 				folder = node.Parent.Path
// 			}
// 		}
// 	} else if sm.editor.IsFocused() {
// 		buf := sm.editor.GetBuffer()
// 		if buf != nil {
// 			folder = sm.editor.GetBufferParentFolder()
// 		}
// 	}

// 	dialog := dialog.NewNewFileDialog(folder, func(fullPath string) {
// 		sm.logger.Println("[New file created]:", fullPath)

// 		// Ensure parent directory exists
// 		dir := filepath.Dir(fullPath)
// 		if err := os.MkdirAll(dir, 0755); err != nil {
// 			sm.logger.Println("Failed to create directories:", err)
// 			sm.CloseDialog()
// 			return
// 		}

// 		// Create empty file if not exists
// 		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
// 			if err := os.WriteFile(fullPath, []byte{}, 0644); err != nil {
// 				sm.logger.Println("Failed to create file:", err)
// 				sm.CloseDialog()
// 				return
// 			}
// 		}

// 		// Refresh sidebar
// 		if sm.sidebar != nil {
// 			node := sm.sidebar.GetSelectedNode()
// 			if node != nil {
// 				if node.IsDir {
// 					sm.sidebar.Tree.ReloadChildren(node)
// 				} else if node.Parent != nil {
// 					sm.sidebar.Tree.ReloadChildren(node.Parent)
// 				}
// 			}
// 			sm.sidebar.ScrollY = 0
// 			sm.sidebar.Selected = 0
// 		}

// 		// Open in editor
// 		if sm.editor != nil {
// 			buf := sm.bufferManager.Open(fullPath)
// 			sm.editor.SetBuffer(buf)
// 		}

// 		sm.CloseDialog()
// 	}, sm.CloseDialog)

// 	sm.OpenDialog(dialog.Dialog)
// }

// func (sm *ScreenManager) ConfirmDeleteNode(node *treeview.Node) {
// 	if node == nil {
// 		return
// 	}

// 	cancelFunc := func() {
// 		sm.CloseDialog()
// 		sm.restoreEditorFocus()
// 	}

// 	dialogDel := dialog.NewDeleteNodeDialog(node.Path, func(path string) {
// 		sm.logger.Println("[Deleted]:", path)

// 		// Clear editor if deleted
// 		if sm.editor != nil && sm.editor.GetBuffer() != nil {
// 			if sm.editor.GetBuffer().File == path {
// 				sm.editor.SetBuffer(nil)
// 			}
// 		}

// 		// Refresh parent folder
// 		if node.Parent != nil && sm.sidebar != nil {
// 			sm.sidebar.Tree.ReloadChildren(node.Parent)
// 			sm.sidebar.ScrollY = 0
// 			if len(sm.sidebar.Tree.Nodes) > 0 {
// 				sm.sidebar.Selected = 0
// 			} else {
// 				sm.sidebar.Selected = -1
// 			}
// 		}

// 		sm.CloseDialog()
// 	}, sm.restoreEditorFocus, cancelFunc)

// 	sm.OpenDialog(dialogDel.Dialog)
// }

// func (sm *ScreenManager) openFileSearchDialog() {
// 	files := sm.bufferManager.ListFilesInCwd()
// 	fsDialog := dialog.NewFileSearchDialog(sm.rootPath, files, func(selected string) {
// 		buf := sm.bufferManager.Open(selected)
// 		sm.editor.SetBuffer(buf)
// 		sm.CloseDialog()
// 	}, sm.restoreEditorFocus, sm.CloseDialog)

// 	sm.OpenDialog(fsDialog.Dialog)
// }
