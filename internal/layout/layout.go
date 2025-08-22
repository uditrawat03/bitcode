package layout

type EditorPadding struct {
	Top    int
	Bottom int
	Left   int
	Right  int
}

type Rect struct {
	X, Y, Width, Height int
}

type UILayout struct {
	SidebarWidth    int
	EditorWidth     int
	StatusBarHeight int
	TopBarHeight    int
	EditorPadding   EditorPadding
	MinWidth        int
	MinHeight       int
}

func NewUILayout(opts ...func(*UILayout)) *UILayout {
	l := &UILayout{
		SidebarWidth:    20,
		EditorWidth:     80,
		StatusBarHeight: 1,
		TopBarHeight:    1,
		EditorPadding: EditorPadding{
			Top:    1,
			Bottom: 1,
			Left:   2,
			Right:  2,
		},
		MinWidth:  100,
		MinHeight: 30,
	}
	for _, opt := range opts {
		opt(l)
	}
	return l
}

func (l *UILayout) CalculateDimensions(sw, sh int) (int, int) {
	if sw < l.MinWidth {
		sw = l.MinWidth
	}
	if sh < l.MinHeight {
		sh = l.MinHeight
	}
	return sw, sh
}

func (l *UILayout) GetEditorArea(sw, sh int) Rect {
	width, height := l.CalculateDimensions(sw, sh)
	return Rect{
		X:      l.SidebarWidth + l.EditorPadding.Left,
		Y:      l.TopBarHeight + l.EditorPadding.Top,
		Width:  width - l.SidebarWidth - l.EditorPadding.Left - l.EditorPadding.Right,
		Height: height - l.TopBarHeight - l.StatusBarHeight - l.EditorPadding.Top - l.EditorPadding.Bottom,
	}
}

func (l *UILayout) GetSidebarArea(sw, sh int) Rect {
	_, height := l.CalculateDimensions(sw, sh)
	return Rect{
		X:      0,
		Y:      l.TopBarHeight,
		Width:  l.SidebarWidth,
		Height: height - l.TopBarHeight - l.StatusBarHeight,
	}
}

func (l *UILayout) GetStatusBarArea(sw, sh int) Rect {
	width, height := l.CalculateDimensions(sw, sh)
	return Rect{
		X:      0,
		Y:      height - l.StatusBarHeight,
		Width:  width,
		Height: l.StatusBarHeight,
	}
}

func (l *UILayout) GetTopBarArea(sw, _ int) Rect {
	width, _ := l.CalculateDimensions(sw, 0)
	return Rect{X: 0, Y: 0, Width: width, Height: l.TopBarHeight}
}

func (l *UILayout) GetEditorPadding() EditorPadding {
	return l.EditorPadding
}

func ResponsiveLayout(screenWidth, screenHeight int) *UILayout {
	layout := NewUILayout()
	width, height := layout.CalculateDimensions(screenWidth, screenHeight)

	if width < layout.MinWidth || height < layout.MinHeight {
		return layout
	}

	return layout
}
