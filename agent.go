package main

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"golang.org/x/image/colornames"
)

type HealthState int

const (
	StateHealthy HealthState = iota
	StateInfected
	StateImmune
	StateDead
)

type Agent struct {
	x, y           int
	state          HealthState
	timesInfected  int
	infectionTimer int
}

func (a *Agent) handleInfectionTimer(w *World) {
	// Role death check
	chance := rng.Float64()
	if chance < w.deathChance {
		a.state = StateDead
		w.infectedCountNow--
		return
	}
	a.infectionTimer--
	if a.infectionTimer <= 0 {
		a.timesInfected++
		if a.timesInfected >= 5 {
			a.state = StateImmune
		} else {
			a.state = StateHealthy
		}
		w.infectedCountNow--
	}
}

func (a *Agent) Draw(screen *ebiten.Image) {
	gridOffset := 50

	colour := color.RGBA{}
	if a.state == StateInfected {
		colour = colornames.Red
	} else if a.state == StateDead {
		colour = colornames.Darkgrey
	} else if a.state == StateImmune {
		colour = colornames.Cyan
	} else {
		colour = colornames.Green
	}

	vector.StrokeRect(screen, float32(a.x+gridOffset), float32(a.y+gridOffset), 1, 1, 1, colour, false)

}

func (a *Agent) checkDirection(w *World, x int, y int) {
	if x < 0 || x >= w.w || y < 0 || y >= w.h {
		return
	}

	target := w.grid[x][y]
	if target == nil || target.state != StateHealthy {
		return
	}
	shouldInfect := rng.Float64() < w.infectionChance
	if shouldInfect {
		target.state = StateInfected
		target.infectionTimer = w.infectionDuration
		w.infectedCountNow++
	}
}

func (a *Agent) isInfected() bool {
	return a.state == StateInfected
}

func (a *Agent) escape(w *World) (int, int, bool) {
	closestInfectedDist := 999
	infX, infY := -1, -1
	foundInfected := false

	scanRadius := 3
	for dx := -scanRadius; dx <= scanRadius; dx++ {
		for dy := -scanRadius; dy <= scanRadius; dy++ {
			nx, ny := a.x+dx, a.y+dy
			if nx >= 0 && nx < w.w && ny >= 0 && ny < w.h && (dx != 0 || dy != 0) {
				target := w.grid[nx][ny]

				if target != nil && target.state == StateInfected {
					dist := dx*dx + dy*dy
					if dist < closestInfectedDist {
						closestInfectedDist = dist
						infX, infY = nx, ny
						foundInfected = true
					}
				}
			}
		}
	}
	if !foundInfected {
		return a.x, a.y, false
	}

	bestMoveX, bestMoveY := a.x, a.y
	maxDistFromInfected := -1

	for moveDX := -1; moveDX <= 1; moveDX++ {
		for moveDY := -1; moveDY <= 1; moveDY++ {
			nx, ny := a.x+moveDX, a.y+moveDY

			if nx >= 0 && nx < w.w && ny >= 0 && ny < w.h {
				distToInfected := (nx-infX)*(nx-infX) + (ny-infY)*(ny-infY)

				if distToInfected > maxDistFromInfected {
					maxDistFromInfected = distToInfected
					bestMoveX, bestMoveY = nx, ny
				}
			}
		}
	}

	return bestMoveX, bestMoveY, true
}

func (a *Agent) find(w *World) (int, int, bool) {
	bestDist := math.MaxInt
	tx, ty, found := 0, 0, false
	for dx := -2; dx <= 2; dx++ {
		for dy := -2; dy <= 2; dy++ {
			nx, ny := a.x+dx, a.y+dy
			if (dx == 0 && dy == 0) || nx < 0 || nx >= w.w || ny < 0 || ny >= w.h {
				continue
			}
			if t := w.grid[nx][ny]; t != nil && t.state == StateHealthy {
				if d := dx*dx + dy*dy; d < bestDist {
					bestDist, tx, ty, found = d, nx, ny, true
				}
			}
		}
	}
	return tx, ty, found
}
