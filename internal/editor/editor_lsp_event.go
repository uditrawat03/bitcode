package editor

import (
	"context"
	"fmt"
)

// Show hover info
func (ed *Editor) ShowHover() {
	if ed.buffer == nil {
		return
	}

	// ed.buffer.mu.RLock()
	// defer ed.buffer.mu.RUnlock()
	// if ed.buffer.HoverInfo == "" {
	// 	return
	// }

	// Display above the cursor (or adjust x/y for better UX)
	ed.ShowTooltip(
		ed.x+6+ed.buffer.CursorX,
		ed.y+ed.buffer.CursorY-ed.scrollY,
		ed.buffer.HoverInfo,
		nil,
	)
}

// Show completion list
func (ed *Editor) ShowCompletion(ctx context.Context) {
	if ed.buffer == nil || len(ed.buffer.Completions) == 0 {
		return
	}

	items := ed.buffer.Completions
	var labels []string
	for _, it := range items {
		if it.Label != "" {
			labels = append(labels, it.Label)
		}
	}

	ed.ShowTooltip(ed.x+6+ed.buffer.CursorX, ed.y+ed.buffer.CursorY-ed.scrollY, "", labels)
}

// Show diagnostics tooltip for current line
func (ed *Editor) ShowDiagnostics() {
	if ed.buffer == nil || len(ed.buffer.Diagnostics) == 0 {
		return
	}
	line := ed.buffer.CursorY

	var msgs []string
	for _, d := range ed.buffer.Diagnostics {
		if d.Range.Start.Line <= line && line <= d.Range.End.Line {
			msgs = append(msgs, d.Message)
		}
	}

	if len(msgs) > 0 {
		ed.ShowTooltip(ed.x+6+ed.buffer.CursorX, ed.y+1+ed.buffer.CursorY-ed.scrollY, fmt.Sprintf("%s", msgs), nil)
	}
}

// Show code actions
func (ed *Editor) ShowCodeActions() {
	if ed.buffer == nil || len(ed.buffer.CodeActions) == 0 {
		return
	}

	var titles []string
	for _, act := range ed.buffer.CodeActions {
		if act.Title != "" {
			titles = append(titles, act.Title)
		}
	}

	ed.ShowTooltip(ed.x+6+ed.buffer.CursorX, ed.y+ed.buffer.CursorY-ed.scrollY, "", titles)
}
