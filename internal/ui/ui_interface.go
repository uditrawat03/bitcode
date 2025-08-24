package ui

import (
	"log"

	"github.com/gdamore/tcell/v2"
	"github.com/uditrawat03/bitcode/internal/layout"
)

type UIComponent interface {
	Render(screen tcell.Screen, lm *layout.LayoutManager)
	Resize(width, height int)
	SetPosition(x, y int)
	GetRect() layout.Rect
	SetLogger(logger *log.Logger)
}

type BaseComponent struct {
	Rect   layout.Rect
	Name   string
	Logger *log.Logger
}

func (b *BaseComponent) SetPosition(x, y int)         { b.Rect.X = x; b.Rect.Y = y }
func (b *BaseComponent) Resize(width, height int)     { b.Rect.Width = width; b.Rect.Height = height }
func (b *BaseComponent) GetRect() layout.Rect         { return b.Rect }
func (b *BaseComponent) SetLogger(logger *log.Logger) { b.Logger = logger }
