package main

import (
	"image/color"

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

	vector.StrokeRect(screen, float32(a.x+gridOffset), float32(a.y+gridOffset), 1, 1, 0.5, colour, true)

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
				target := g.agents[nx][ny]

				if target != nil && target.state == StateHealthy {
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
	minDist := 99
	targetX, targetY := -1, -1
	found := false

	foundList := make([]struct{ x, y int }, 0)
	for dx := -2; dx <= 2; dx++ {
		for dy := -2; dy <= 2; dy++ {
			nx, ny := a.x+dx, a.y+dy

			if nx >= 0 && nx < g.w && ny >= 0 && ny < g.h && (dx != 0 || dy != 0) {
				target := g.agents[nx][ny]

				if target != nil && target.state == StateHealthy {
					foundList = append(foundList, struct{ x, y int }{nx, ny})
				}
			}
		}
	}

	if len(foundList) > 0 {
		index := rng.Intn(len(foundList))
		t := foundList[index]
		dist := t.x*t.x + t.y*t.y
		if dist < minDist {
			minDist = dist
			targetX, targetY = a.x+t.x, a.y+t.y
			found = true
		}

	}
	return targetX, targetY, found

}
