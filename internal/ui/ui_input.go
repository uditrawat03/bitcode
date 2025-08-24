package ui

import "github.com/uditrawat03/bitcode/internal/layout"

type Input struct {
	BaseComponent
	Text     string
	CursorX  int
	Password bool // mask input if true
}

func NewInput(password bool) *Input {
	return &Input{Password: password}
}

func (i *Input) Render(lm *layout.LayoutManager) {
	// For now, place at bottom bar area
	i.Rect = lm.GetLayout().GetStatusBarArea(lm.GetLayout().MinWidth, lm.GetLayout().MinHeight)
	// TODO: Draw input text and cursor
}
