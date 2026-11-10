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

func (a *Agent) handleInfectionTimer(g *Game) {
	// Role death check
	chance := rng.Float64()
	if chance < g.deathChance {
		a.state = StateDead
		g.infectedCountNow--
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
		g.infectedCountNow--
	}
}

func (a *Agent) draw(screen *ebiten.Image) {
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

	vector.StrokeRect(screen, float32(a.x+gridOffset), float32(a.y+gridOffset), 0.5, 0.5, 0.5, colour, false)

}

func (a *Agent) checkDirection(g *Game, x int, y int) {
	if x < 0 || x >= g.w || y < 0 || y >= g.h {
		return
	}

	target := g.grid[x][y]
	if target == nil || target.state != StateHealthy {
		return
	}
	shouldInfect := rng.Float64() < g.infectionChance
	if shouldInfect {
		target.state = StateInfected
		target.infectionTimer = g.infectionDuration
		g.infectedCountNow++
	}
}

func (a *Agent) isInfected() bool {
	return a.state == StateInfected
}

func (a *Agent) escape(g *Game) (int, int, bool) {
	closestInfectedDist := 999
	infX, infY := -1, -1
	foundInfected := false

	scanRadius := 3
	for dx := -scanRadius; dx <= scanRadius; dx++ {
		for dy := -scanRadius; dy <= scanRadius; dy++ {
			nx, ny := a.x+dx, a.y+dy
			if nx >= 0 && nx < g.w && ny >= 0 && ny < g.h && (dx != 0 || dy != 0) {
				target := g.grid[nx][ny]

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

			if nx >= 0 && nx < g.w && ny >= 0 && ny < g.h {
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

func (a *Agent) find(g *Game) (int, int, bool) {
	bestDist := math.MaxInt
	tx, ty, found := 0, 0, false
	for dx := -2; dx <= 2; dx++ {
		for dy := -2; dy <= 2; dy++ {
			nx, ny := a.x+dx, a.y+dy
			if (dx == 0 && dy == 0) || nx < 0 || nx >= g.w || ny < 0 || ny >= g.h {
				continue
			}
			if t := g.grid[nx][ny]; t != nil && t.state == StateHealthy {
				if d := dx*dx + dy*dy; d < bestDist {
					bestDist, tx, ty, found = d, nx, ny, true
				}
			}
		}
	}
	return tx, ty, found
}
