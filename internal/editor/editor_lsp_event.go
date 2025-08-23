package editor

import (
	"strings"
)

// Show hover info
func (ed *Editor) ShowHover() {
	if ed.buffer == nil {
		return
	}

	// Display above the cursor (or adjust x/y for better UX)
	ed.showTooltipFn(
		ed.x+6+ed.buffer.CursorX,
		ed.y+ed.buffer.CursorY-ed.scrollY,
		ed.buffer.HoverInfo,
		nil,
	)
}

// Show completion list
func (ed *Editor) ShowCompletion() {
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

	ed.showTooltipFn(ed.x+6+ed.buffer.CursorX, ed.y+ed.buffer.CursorY-ed.scrollY, "", labels)
}

// Show diagnostics tooltip for current line
func (ed *Editor) ShowDiagnostics() {
	if ed.buffer == nil || len(ed.buffer.Diagnostics) == 0 {
		// ed.hideTooltipFn()
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
		ed.showTooltipFn(
			ed.x+6+ed.buffer.CursorX,
			ed.y+1+ed.buffer.CursorY-ed.scrollY,
			strings.Join(msgs, "\n"),
			nil,
		)
	} else {
		// ed.hideTooltipFn()
	}
}

// Show code actions
func (ed *Editor) ShowCodeActions() {
	if ed.buffer == nil {
		return
	}

	actions := ed.buffer.GetCodeActions()
	if len(actions) == 0 {
		return
	}

	var titles []string
	for _, act := range actions {
		if act.Title != "" {
			titles = append(titles, act.Title)
		}
	}

	ed.showTooltipFn(ed.x+6+ed.buffer.CursorX, ed.y+ed.buffer.CursorY-ed.scrollY, "", titles)
}
