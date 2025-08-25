package ui

import (
	"os"
	"path/filepath"

	"github.com/uditrawat03/bitcode/internal/dialog"
)

// New file dialog
func (sm *ScreenManager) openNewFileDialog() {
	folder := sm.getBufferParentFolder()
	sm.isDialogOpen = true

	var dlg *dialog.InputDialog
	dlg = dialog.NewInputDialog(
		folder,
		"New File",
		"Enter filename:",
		func(input string) {
			if input == "" {
				sm.closeDialog(dlg)
				return
			}

			fullPath := filepath.Join(folder, input)
			if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
				sm.logger.Println("Failed to create directories:", err)
				sm.closeDialog(dlg)
				return
			}

			if _, err := os.Stat(fullPath); os.IsNotExist(err) {
				if err := os.WriteFile(fullPath, []byte{}, 0644); err != nil {
					sm.logger.Println("Failed to create file:", err)
					sm.closeDialog(dlg)
					return
				}
			}

			buf := sm.Bm.Open(fullPath)
			sm.OpenBufferInEditor(buf)

			if sm.selectedNode != nil && sm.refreshTree != nil {
				parent := sm.selectedNode
				if !parent.IsDir && parent.Parent != nil {
					parent = parent.Parent
				}
				sm.refreshTree(parent)
			}

			sm.closeDialog(dlg)
		},
		func() { sm.closeDialog(dlg) },
	)

	sw, sh := sm.screen.Size()
	dlgWidth, dlgHeight := 40, 7
	x := (sw - dlgWidth) / 2
	y := (sh - dlgHeight) / 2
	dlg.SetPosition(x, y)
	dlg.Resize(dlgWidth, dlgHeight)

	sm.AddComponent(dlg, true)
	sm.setFocus(len(sm.focusOrder) - 1)
	sm.RequestRender()
}
