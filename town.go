package main

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

var townColor = color.RGBA{R: 40, G: 40, B: 80, A: 10}

// Town Constants
const (
	townCount         = 8
	travelChance      = .0005
	stayDuration      = 100
	infectedTravelBan = .05
)

type Town struct {
	cx, cy    int
	radius    float32
	flowField []int32
}

func (t *Town) Draw(screen *ebiten.Image, camera *Camera) {
	x, y := camera.WorldToScreen(float32(t.cx), float32(t.cy))
	vector.FillCircle(screen, x, y, t.radius*camera.Zoom, townColor, false)
}
