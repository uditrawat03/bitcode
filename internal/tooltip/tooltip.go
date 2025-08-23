package tooltip

// TooltipType defines the type of tooltip to render.
type TooltipType int

const (
	TooltipNone TooltipType = iota
	TooltipText             // show plain text (hover, diagnostics, docs, etc.)
	TooltipList             // selectable list (completion, code actions, menus, etc.)
)

// Tooltip represents a UI popover element for hints, completions, diagnostics, etc.
type Tooltip struct {
	Visible  bool
	Type     TooltipType
	X, Y     int      // placement coordinates (screen or editor grid coords)
	Content  string   // used when Type == TooltipText
	Items    []string // used when Type == TooltipList
	Selected int      // index of the currently highlighted item in list
}

// NewText creates a new plain text tooltip.
func NewText(x, y int, content string) *Tooltip {
	return &Tooltip{
		Visible: true,
		Type:    TooltipText,
		X:       x,
		Y:       y,
		Content: content,
	}
}

// NewList creates a new selectable list tooltip.
func NewList(x, y int, items []string) *Tooltip {
	return &Tooltip{
		Visible:  true,
		Type:     TooltipList,
		X:        x,
		Y:        y,
		Items:    items,
		Selected: 0,
	}
}

// Next moves selection down in a list tooltip.
func (t *Tooltip) Next() {
	if t.Type != TooltipList || len(t.Items) == 0 {
		return
	}
	t.Selected = (t.Selected + 1) % len(t.Items)
}

// Prev moves selection up in a list tooltip.
func (t *Tooltip) Prev() {
	if t.Type != TooltipList || len(t.Items) == 0 {
		return
	}
	t.Selected = (t.Selected - 1 + len(t.Items)) % len(t.Items)
}

// Apply runs the callback with the selected item (if any).
func (t *Tooltip) Apply(applyFn func(item string)) {
	if t.Type == TooltipList && len(t.Items) > 0 && applyFn != nil {
		applyFn(t.Items[t.Selected])
	}
	t.Close()
}

// Close hides and resets the tooltip.
func (t *Tooltip) Close() {
	t.Visible = false
	t.Type = TooltipNone
	t.Content = ""
	t.Items = nil
	t.Selected = 0
}
