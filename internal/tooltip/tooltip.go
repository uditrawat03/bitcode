package tooltip

import "github.com/gdamore/tcell/v2"

// TooltipType defines the type of tooltip to render
type TooltipType int

const (
	TooltipNone TooltipType = iota
	TooltipText
	TooltipList
)

// Alignment for floating windows
type Align int

const (
	AlignTop Align = iota
	AlignBottom
	AlignLeft
	AlignRight
)

// Tooltip represents a Neovim-style floating popup
type Tooltip struct {
	Visible   bool
	Type      TooltipType
	X, Y      int      // placement coordinates
	Width     int      // 0 = auto-size
	Height    int      // 0 = auto-size
	Align     Align    // optional alignment
	Content   string   // for TooltipText
	Items     []string // for TooltipList
	Selected  int      // selected index in list
	MaxHeight int      // max height for scrolling
}

// --- Constructors ---

func NewText(x, y int, content string) *Tooltip {
	return &Tooltip{
		Visible: true,
		Type:    TooltipText,
		X:       x,
		Y:       y,
		Content: content,
	}
}

func NewList(x, y int, items []string, maxHeight int) *Tooltip {
	return &Tooltip{
		Visible:   true,
		Type:      TooltipList,
		X:         x,
		Y:         y,
		Items:     items,
		Selected:  0,
		MaxHeight: maxHeight,
	}
}

// --- Navigation ---

func (t *Tooltip) Next() {
	if t.Type != TooltipList || len(t.Items) == 0 {
		return
	}
	if t.Selected < len(t.Items)-1 {
		t.Selected++
	}
}

func (t *Tooltip) Prev() {
	if t.Type != TooltipList || len(t.Items) == 0 {
		return
	}
	if t.Selected > 0 {
		t.Selected--
	}
}

func (t *Tooltip) PageDown() {
	if t.Type != TooltipList || len(t.Items) == 0 {
		return
	}
	t.Selected += t.MaxHeight
	if t.Selected >= len(t.Items) {
		t.Selected = len(t.Items) - 1
	}
}

func (t *Tooltip) PageUp() {
	if t.Type != TooltipList || len(t.Items) == 0 {
		return
	}
	t.Selected -= t.MaxHeight
	if t.Selected < 0 {
		t.Selected = 0
	}
}

// Apply selection
func (t *Tooltip) Apply(applyFn func(item string)) {
	if t.Type == TooltipList && len(t.Items) > 0 && applyFn != nil {
		applyFn(t.Items[t.Selected])
	}
	t.Close()
}

// Close tooltip
func (t *Tooltip) Close() {
	t.Visible = false
	t.Type = TooltipNone
	t.Content = ""
	t.Items = nil
	t.Selected = 0
}

// --- Rendering ---

func (t *Tooltip) Draw(screen tcell.Screen) {
	if !t.Visible {
		return
	}

	style := tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.ColorBlack)
	borderStyle := tcell.StyleDefault.Foreground(tcell.ColorGray).Background(tcell.ColorBlack)

	w, h := t.Width, t.Height

	switch t.Type {
	case TooltipText:
		lines := splitLines(t.Content)
		if w == 0 {
			maxLen := 0
			for _, line := range lines {
				if len(line) > maxLen {
					maxLen = len(line)
				}
			}
			w = maxLen + 2
		}
		if h == 0 {
			h = len(lines) + 2
		}

	case TooltipList:
		if len(t.Items) == 0 {
			return
		}
		if w == 0 {
			maxItemLen := 0
			for _, item := range t.Items {
				if len(item) > maxItemLen {
					maxItemLen = len(item)
				}
			}
			w = maxItemLen + 2
		}
		if h == 0 {
			h = len(t.Items) + 2
			if t.MaxHeight > 0 && h > t.MaxHeight+2 {
				h = t.MaxHeight + 2
			}
		}
	}

	// --- Draw border ---
	for dx := 0; dx < w; dx++ {
		screen.SetContent(t.X+dx, t.Y, '─', nil, borderStyle)
		screen.SetContent(t.X+dx, t.Y+h-1, '─', nil, borderStyle)
	}
	for dy := 0; dy < h; dy++ {
		screen.SetContent(t.X, t.Y+dy, '│', nil, borderStyle)
		screen.SetContent(t.X+w-1, t.Y+dy, '│', nil, borderStyle)
	}
	screen.SetContent(t.X, t.Y, '┌', nil, borderStyle)
	screen.SetContent(t.X+w-1, t.Y, '┐', nil, borderStyle)
	screen.SetContent(t.X, t.Y+h-1, '└', nil, borderStyle)
	screen.SetContent(t.X+w-1, t.Y+h-1, '┘', nil, borderStyle)

	// --- Draw content ---
	switch t.Type {
	case TooltipText:
		lines := splitLines(t.Content)
		for i, line := range lines {
			if i+1 >= h-1 {
				break
			}
			for j, r := range line {
				if j+1 >= w-1 {
					break
				}
				screen.SetContent(t.X+1+j, t.Y+1+i, r, nil, style)
			}
		}

	case TooltipList:
		visibleCount := h - 2
		if visibleCount <= 0 {
			visibleCount = 1
		}

		// Center selected item
		scrollTop := t.Selected - visibleCount/2
		if scrollTop < 0 {
			scrollTop = 0
		}
		if scrollTop+visibleCount > len(t.Items) {
			scrollTop = len(t.Items) - visibleCount
		}
		if scrollTop < 0 {
			scrollTop = 0
		}

		for i := 0; i < visibleCount && i+scrollTop < len(t.Items); i++ {
			item := t.Items[i+scrollTop]
			itemStyle := style
			if i+scrollTop == t.Selected {
				itemStyle = tcell.StyleDefault.Foreground(tcell.ColorBlack).Background(tcell.ColorWhite)
			}
			for j, r := range item {
				if j+1 >= w-1 {
					break
				}
				screen.SetContent(t.X+1+j, t.Y+1+i, r, nil, itemStyle)
			}
			// fill remaining space
			for j := len(item); j < w-2; j++ {
				screen.SetContent(t.X+1+j, t.Y+1+i, ' ', nil, itemStyle)
			}
		}
	}
}

// --- Helpers ---

func splitLines(s string) []string {
	if s == "" {
		return nil
	}
	lines := []string{}
	start := 0
	for i, r := range s {
		if r == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		}
	}
	lines = append(lines, s[start:])
	return lines
}
