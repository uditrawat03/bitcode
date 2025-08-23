package ui

import "github.com/uditrawat03/bitcode/internal/tooltip"

func (sm *ScreenManager) ShowTooltip(x, y int, text string, items []string) {
	if len(items) > 0 {
		sm.tooltip = tooltip.Tooltip{
			Visible:  true,
			Type:     tooltip.TooltipList,
			X:        x,
			Y:        y,
			Items:    items,
			Selected: 0,
		}
	} else if text != "" {
		sm.tooltip = tooltip.Tooltip{
			Visible: true,
			Type:    tooltip.TooltipText,
			X:       x,
			Y:       y,
			Content: text,
		}
	} else {
		sm.HideTooltip()
	}
}

func (ed *ScreenManager) HideTooltip() {
	ed.tooltip.Close()
}
