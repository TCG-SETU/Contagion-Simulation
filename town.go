package main

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

var townColor = color.RGBA{R: 100, G: 100, B: 200, A: 60}

// Town Constants
const (
	townCount         = 5
	travelChance      = .0005
	stayDuration      = 100
	infectedTravelBan = .05
)

type Town struct {
	cx, cy int
	radius float32
}

func (t *Town) Draw(screen *ebiten.Image, camera *Camera) {
	x, y := camera.WorldToScreen(float32(t.cx), float32(t.cy))
	vector.FillCircle(screen, x, y, t.radius*camera.Zoom, townColor, false)
}
