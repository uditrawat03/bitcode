package ui

import (
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/uditrawat03/bitcode/internal/dialog"
	"github.com/uditrawat03/bitcode/internal/treeview"
)

// OpenDialog displays a modal dialog
func (sm *ScreenManager) OpenDialog(d *dialog.Dialog) {
	sm.dialog = d
	d.SetFocus(true)
}

// CloseDialog closes the modal dialog
func (sm *ScreenManager) CloseDialog() {
	if sm.dialog != nil {
		sm.dialog.SetFocus(false)
		sm.dialog = nil
		sm.screen.HideCursor()
	}
}

func (sm *ScreenManager) IsDialogOpen() bool {
	return sm.dialog != nil
}

// restoreEditorFocus restores focus to the editor
func (sm *ScreenManager) restoreEditorFocus() {
	if sm.editor != nil && sm.editor.GetBuffer() != nil {
		sm.focusOrder[sm.focusedIdx].Blur()
		sm.focusedIdx = 1 // editor index
		sm.editor.Focus()
	}
}

// openNewFileDialog opens a modal dialog to create a new file
func (sm *ScreenManager) openNewFileDialog() {
	folder := "."

	// Determine folder based on focus
	if sm.sidebar.IsFocused() {
		node := sm.sidebar.GetSelectedNode()
		if node != nil {
			if node.IsDir {
				folder = node.Path
			} else if node.Parent != nil {
				folder = node.Parent.Path
			}
		}
	} else if sm.editor.IsFocused() {
		buf := sm.editor.GetBuffer()
		if buf != nil {
			folder = sm.editor.GetBufferParentFolder()
		}
	}

	dialogOpen := dialog.NewDialog(
		"Create New file",
		"",
		40,
		7,
		func(name string) {
			fullPath := filepath.Join(folder, name)

			// Create parent directories if missing
			dir := filepath.Dir(fullPath)
			if err := os.MkdirAll(dir, 0755); err != nil {
				log.Println("Failed to create directories:", err)
				return
			}

			// Create empty file if not exists
			if _, err := os.Stat(fullPath); os.IsNotExist(err) {
				if err := os.WriteFile(fullPath, []byte{}, 0644); err != nil {
					log.Println("Failed to create file:", err)
					return
				}
			}

			// refresh sidebar with new TreeView
			if sm.sidebar != nil {
				node := sm.sidebar.GetSelectedNode()
				sm.sidebar.Tree.ReloadChildren(node)
				sm.sidebar.ScrollY = 0
				sm.sidebar.Selected = 0
			}

			// open in editor
			if sm.editor != nil {
				buf := sm.bufferManager.Open(fullPath)
				sm.editor.SetBuffer(buf)
			}

			sm.CloseDialog()
		},
		func(_ string) {
			sm.CloseDialog()
		},
		func() {
			sm.restoreEditorFocus()
		},
	)

	sm.OpenDialog(dialogOpen)
}

// confirmDeleteNode shows a modal dialog to delete a node
func (sm *ScreenManager) confirmDeleteNode(node *treeview.Node) {
	if node == nil {
		return
	}
	title := "Delete " + node.Name + "?"
	message := "Press Enter to confirm, Esc to cancel."

	dialogWidth := lenLongestLine(message) + 4
	if dialogWidth < 40 {
		dialogWidth = 40
	}
	dialogHeight := countLines(message) + 4

	dialogDelete := dialog.NewDialog(
		title, message, dialogWidth, dialogHeight,
		func(_ string) {
			fullPath := node.Path
			var err error
			if node.IsDir {
				err = os.RemoveAll(fullPath)
			} else {
				err = os.Remove(fullPath)
			}
			if err != nil {
				log.Println("Delete failed:", err)
			} else {
				// 1. Close editor if active file is deleted
				if sm.editor != nil && sm.editor.GetBuffer() != nil {
					if sm.editor.GetBuffer().File == fullPath {
						sm.editor.SetBuffer(nil)
					}
				}

				// 2. Refresh parent folder in TreeView
				if node.Parent != nil && sm.sidebar != nil {
					sm.sidebar.Tree.ReloadChildren(node.Parent)
					sm.sidebar.ScrollY = 0

					// Try to select next available node if possible
					if len(sm.sidebar.Tree.Nodes) > 0 {
						sm.sidebar.Selected = 0
					} else {
						sm.sidebar.Selected = -1
					}
				}
			}

			sm.CloseDialog()
		},
		func(_ string) {
			sm.CloseDialog()
		},
		func() {
			sm.restoreEditorFocus()
		},
	)

	dialogDelete.HasInput = false
	sm.OpenDialog(dialogDelete)
}

// Helpers
func lenLongestLine(s string) int {
	max := 0
	for _, line := range strings.Split(s, "\n") {
		if len(line) > max {
			max = len(line)
		}
	}
	return max
}

func countLines(s string) int {
	return len(strings.Split(s, "\n"))
}
