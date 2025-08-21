package layout

type LayoutManager struct {
	layout *UILayout
}

func CreateLayoutManager() *LayoutManager {
	return &LayoutManager{
		layout: NewUILayout(),
	}
}

func (lm *LayoutManager) UpdateLayout(width, height int) {
	lm.layout = ResponsiveLayout(width, height)
}

// GetLayout returns the current layout
func (lm *LayoutManager) GetLayout() *UILayout {
	return lm.layout
}
