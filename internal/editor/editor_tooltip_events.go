package editor

import "github.com/uditrawat03/bitcode/internal/tooltip"

func (ed *Editor) ShowTooltip(x, y int, text string, items []string) {
	if len(items) > 0 {
		ed.tooltip = tooltip.Tooltip{
			Visible:  true,
			Type:     tooltip.TooltipList,
			X:        x,
			Y:        y,
			Items:    items,
			Selected: 0,
		}
	} else if text != "" {
		ed.tooltip = tooltip.Tooltip{
			Visible: true,
			Type:    tooltip.TooltipText,
			X:       x,
			Y:       y,
			Content: text,
		}
	} else {
		ed.HideTooltip()
	}
}

func (ed *Editor) HideTooltip() {
	ed.tooltip.Close()
}
