package main

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

var townColor = color.RGBA{R: 100, G: 100, B: 200, A: 60}

// Town Constants
const (
	// townCount represents the number of towns generated in the world.
	townCount = 5
	// travelChance determines the probability of an agent starting a journey.
	travelChance = .0005
	// stayDuration determines the base duration of an agent's visits to another town.
	stayDuration = 100
	// infectionTravelBan determines the probability of an infected agent being prevented from travelling.
	infectedTravelBan = .05
)

// Town represents a location within the simulation.
// Each town has a centre position and a radius defining its area.
type Town struct {
	cx, cy int
	radius float32
}

// Draw renders the town on the simulation screen.
// It converts the town's world coordinates into screen coordinates
// and draws a circle scaled according to the camera's zoom level.
func (t *Town) Draw(screen *ebiten.Image, camera *Camera) {
	x, y := camera.WorldToScreen(float32(t.cx), float32(t.cy))
	vector.FillCircle(screen, x, y, t.radius*camera.Zoom, townColor, false)
}
