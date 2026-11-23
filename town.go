package main

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

var townColor = color.RGBA{R: 100, G: 100, B: 200, A: 100}

type Town struct {
	cx, cy int
	radius float32
}

func (t *Town) Draw(screen *ebiten.Image) {
	vector.FillCircle(screen, float32(t.cx+gridOffset), float32(t.cy+gridOffset), t.radius, townColor, false)
}
