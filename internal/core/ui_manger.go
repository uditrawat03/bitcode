package core

import (
	"log"

	"github.com/gdamore/tcell/v2"
	"github.com/uditrawat03/bitcode/internal/layout"
)

type UIManager struct {
	Components []UIComponent
	Lm         *layout.LayoutManager
	Logger     *log.Logger
}

func NewUIManager(logger *log.Logger, lm *layout.LayoutManager) *UIManager {
	return &UIManager{Lm: lm, Logger: logger}
}

func (u *UIManager) AddComponent(c UIComponent) {
	c.SetLogger(u.Logger)
	u.Components = append(u.Components, c)
}

func (u *UIManager) RenderAll(screen tcell.Screen) {
	for _, comp := range u.Components {
		comp.Render(screen, u.Lm)
	}
}

func (ui *UIManager) RemoveComponent(c UIComponent) {
	for i, comp := range ui.Components {
		if comp == c {
			ui.Components = append(ui.Components[:i], ui.Components[i+1:]...)
			break
		}
	}
}
