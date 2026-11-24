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
	StateInfectious
	StateInfected
	StateImmune
	StateDead
)

type TravelMode int

const (
	Local     TravelMode = iota // wandering around home
	Outbound                    // heading to another town
	Visiting                    // wandering around the destination
	Returning                   // heading home
)

type Agent struct {
	x, y            int
	state           HealthState
	timesInfected   int
	contagiousTimer int
	infectionTimer  int

	home      int // index into g.towns
	dest      int
	mode      TravelMode
	stayTimer int
}

func (a *Agent) anchor(w *World) Town {
	if a.mode == Outbound || a.mode == Visiting {
		return w.towns[a.dest]
	}
	return w.towns[a.home]
}

func (a *Agent) inTown(t Town) bool {
	dx, dy := a.x-t.cx, a.y-t.cy
	return dx*dx+dy*dy <= int(t.radius*t.radius)
}

func (a *Agent) updateTravel(w *World) {
	switch a.mode {
	case Local:
		if rng.Float64() >= travelChance {
			return
		}
		// symptomatic agents are less likely to travel
		if a.isInfected() && rng.Float64() < infectedTravelBan {
			return
		}
		a.dest = rng.Intn(len(w.towns) - 1)
		if a.dest >= a.home { // pick any town except home
			a.dest++
		}
		a.mode = Outbound
	case Outbound:
		if a.inTown(w.towns[a.dest]) {
			a.mode = Visiting
			a.stayTimer = stayDuration/2 + rng.Intn(stayDuration)
		}
	case Visiting:
		a.stayTimer--
		if a.stayTimer <= 0 {
			a.mode = Returning
		}
	case Returning:
		if a.inTown(w.towns[a.home]) {
			a.mode = Local
		}
	}
}

func (a *Agent) flowStep(w *World, town int) (int, int) {
	field := w.towns[town].flowField
	current := field[a.y*w.m.mw+a.x]
	// at the target or unreachable cell, skip
	if current < 0 {
		return 0, 0
	}
	bestDX, bestDY := 0, 0
	best := current // best direction

	for _, d := range directions {
		if !w.m.canStep(a.x, a.y, d[0], d[1]) {
			continue
		}
		nx, ny := a.x+d[0], a.y+d[1]
		nd := field[ny*w.m.mw+nx]

		// invalid / occupied square
		if nd < 0 || w.grid[nx][ny] != nil {
			continue
		}
		// compare with previous best dir
		if nd < best || (nd == best && best < current && rng.Intn(2) == 0) {
			best, bestDX, bestDY = nd, d[0], d[1]
		}
	}
	return bestDX, bestDY
}

func (a *Agent) wander(w *World) (int, int) {

	dx, dy := rng.Intn(3)-1, rng.Intn(3)-1

	switch a.mode {
	case Outbound, Returning:
		if rng.Float64() < 0.8 {
			target := a.home
			if a.mode == Outbound {
				target = a.dest
			}
			if fx, fy := a.flowStep(w, target); fx != 0 || fy != 0 {
				return fx, fy
			}
		}
	default: // Local, Visiting
		townIdx := a.anchor(w)
		if !a.inTown(townIdx) {
			town := a.home
			if a.mode == Visiting {
				town = a.dest
			}
			if fx, fy := a.flowStep(w, town); fx != 0 || fy != 0 {
				return fx, fy
			}
		}
	}
	return dx, dy
}

func (a *Agent) attemptInfection(w *World, c *Counts) {
	for dx := -1; dx <= 1; dx++ {
		for dy := -1; dy <= 1; dy++ {
			if dx != 0 || dy != 0 {
				checkDirection(w, a.x+dx, a.y+dy, c)
			}
		}
	}
	a.handleInfectionTimer(w, c)
}

func (a *Agent) handleInfectionTimer(w *World, c *Counts) {
	if a.state == StateInfectious {
		a.contagiousTimer--
		if a.contagiousTimer <= 0 {
			a.contagiousTimer = 0
			a.state = StateInfected
			c.Infected++
			c.Infectious--
		}
		return
	}

	// Role death check
	chance := rng.Float64()
	if chance < w.deathChance {
		a.state = StateDead
		c.Dead++
		c.Infected--
		return
	}
	a.infectionTimer--
	if a.infectionTimer <= 0 {
		a.timesInfected++
		if a.timesInfected >= 2 {
			a.state = StateImmune
			c.Immune++
		} else {
			a.state = StateHealthy
			c.Healthy++
		}
		c.Infected--
	}
}

func (a *Agent) Draw(screen *ebiten.Image, camera *Camera) {
	if !camera.Visible(float32(a.x), float32(a.y), 1) {
		return
	}
	colour := color.RGBA{}
	if a.state == StateInfected {
		colour = colornames.Red
	} else if a.state == StateDead {
		colour = colornames.Darkgrey
	} else if a.state == StateImmune {
		colour = colornames.Cyan
	} else if a.state == StateInfectious {
		colour = colornames.Orange
	} else {
		colour = colornames.Lightgreen
	}

	sx, sy := camera.WorldToScreen(float32(a.x), float32(a.y))
	size := max(camera.Zoom, 0.5) // never smaller than 1 pixel
	vector.StrokeRect(screen, sx, sy, size, size, size, colour, false)

}

func checkDirection(w *World, x int, y int, c *Counts) {
	if x < 0 || x >= w.w || y < 0 || y >= w.h {
		return
	}

	target := w.grid[x][y]
	if target == nil || target.state != StateHealthy {
		return
	}
	shouldInfect := rng.Float64() < w.infectionChance
	if shouldInfect {
		target.state = StateInfectious
		target.infectionTimer = w.infectionDuration
		target.contagiousTimer = contagiousTimer
		c.Infectious++
		c.Healthy--
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
