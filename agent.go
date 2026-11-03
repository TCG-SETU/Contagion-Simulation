package main

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"golang.org/x/image/colornames"
)

type Agent struct {
	x, y           int
	infected       bool
	infectionTimer int
}

func (a *Agent) handleInfectionTimer(g *Game) {
	a.infectionTimer--
	if a.infectionTimer <= 0 {
		a.infected = false
		g.infectedCountNow--
	}
}

func (a *Agent) draw(screen *ebiten.Image) {
	gridOffset := 50

	colour := color.RGBA{}
	if a.infected {
		colour = colornames.Red
	} else {
		colour = colornames.Green
	}

	vector.StrokeRect(screen, float32(a.x*2+gridOffset), float32(a.y*2+gridOffset), 1, 1, 2, colour, false)

}
