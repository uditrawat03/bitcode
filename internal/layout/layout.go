package layout

type EditorPadding struct {
	Top    int
	Bottom int
	Left   int
	Right  int
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

func NewUILayout() *UILayout {
	return &UILayout{
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
}

func (l *UILayout) CalculateDimensions(screenWidth, screenHeight int) (int, int) {
	width := screenWidth
	height := screenHeight

	if width < l.MinWidth {
		width = l.MinWidth
	}

	if height < l.MinHeight {
		height = l.MinHeight
	}

	return width, height
}

func (l *UILayout) GetEditorArea(screenWidth, screenHeight int) (int, int, int, int) {
	width, height := l.CalculateDimensions(screenWidth, screenHeight)

	editorX := l.SidebarWidth + l.EditorPadding.Left
	editorY := l.TopBarHeight + l.EditorPadding.Top
	editorWidth := width - l.SidebarWidth - l.EditorPadding.Left - l.EditorPadding.Right
	editorHeight := height - l.TopBarHeight - l.StatusBarHeight - l.EditorPadding.Top - l.EditorPadding.Bottom

	return editorX, editorY, editorWidth, editorHeight
}

func (l *UILayout) GetSidebarArea(screenWidth, screenHeight int) (int, int, int, int) {
	_, height := l.CalculateDimensions(screenWidth, screenHeight)

	sidebarX := 0
	sidebarY := l.TopBarHeight
	sidebarWidth := l.SidebarWidth
	sidebarHeight := height - l.TopBarHeight - l.StatusBarHeight

	return sidebarX, sidebarY, sidebarWidth, sidebarHeight
}

func (l *UILayout) GetStatusBarArea(screenWidth, screenHeight int) (int, int, int, int) {
	width, height := l.CalculateDimensions(screenWidth, screenHeight)
	return 0, height - l.StatusBarHeight, width, l.StatusBarHeight
}

func (l *UILayout) GetTopBarArea(screenWidth, screenHeight int) (int, int, int, int) {
	width, _ := l.CalculateDimensions(screenWidth, screenHeight)
	return 0, 0, width, l.TopBarHeight
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
