package ui

import "github.com/uditrawat03/bitcode/internal/layout"

type TextView struct {
	BaseComponent
	Content []string
}

func NewTextView(content []string) *TextView {
	return &TextView{Content: content}
}

func (t *TextView) Render(lm *layout.LayoutManager) {
	// For now, occupy editor area
	t.Rect = lm.GetLayout().GetEditorArea(lm.GetLayout().MinWidth, lm.GetLayout().MinHeight)
	// TODO: Draw each line in terminal
}
