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
	Local     TravelMode = iota 
	Outbound                   
	Visiting             
	Returning               
)


type Agent struct {
	x, y            int
	state           HealthState
	timesInfected   int
	contagiousTimer int
	infectionTimer  int

	home      int 
	dest      int
	mode      TravelMode
	stayTimer int 
}

// anchor selects the town used as the agent’s movement reference so travelling agents
// head toward their destination while others remain tied to home.
func (a *Agent) anchor(w *World) Town {
	if a.mode == Outbound || a.mode == Visiting {
		return w.towns[a.dest]
	}
	return w.towns[a.home]
}

// inTown uses squared distance to avoid a square-root calculation when checking whether an agent has reached a town.
func (a *Agent) inTown(t Town) bool {
	dx, dy := a.x-t.cx, a.y-t.cy
	return dx*dx+dy*dy <= int(t.radius*t.radius)
}

// updateTravel controls travel transitions so agents follow a consistent journey from home to a destination and back.
func (a *Agent) updateTravel(w *World) {
	switch a.mode {

	case Local:
		if rng.Float64() >= travelChance {
			return
		}
		if a.isInfected() && rng.Float64() < infectedTravelBan {
			return
		}
		a.dest = rng.Intn(len(w.towns) - 1)
		if a.dest >= a.home {
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

// wander combines random movement with directional guidance so agents explore towns without losing their route between them.
func (a *Agent) wander(w *World) (int, int) {
	t := a.anchor(w)
	dx, dy := rng.Intn(3)-1, rng.Intn(3)-1

	switch a.mode {
	case Outbound, Returning:
		if rng.Float64() < 0.8 {
			dx = locationToDirection(t.cx, a.x, RIGHT, LEFT)
			dy = locationToDirection(t.cy, a.y, DOWN, UP)
		}
	default:
		if !a.inTown(t) {
			dx = locationToDirection(t.cx, a.x, RIGHT, LEFT)
			dy = locationToDirection(t.cy, a.y, DOWN, UP)
		}
	}
	return dx, dy
}

// attemptInfection checks the eight surrounding gird positions for possible infectious.
// It also updates the agent's infection timers and health states.
func (a *Agent) attemptInfection(w *World, c *Counts) {
	// checks all neighboring positions.
	for dx := -1; dx <= 1; dx++ {
		for dy := -1; dy <= 1; dy++ {
			if dx != 0 || dy != 0 {
				checkDirection(w, a.x+dx, a.y+dy, c)
			}
		}
	}
	a.handleInfectionTimer(w, c)
}

// attemptInfection checks neighbouring cells because infection can only spread through nearby agents, then advances the current agent’s infection timeline.
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

// Draw skips agents outside the camera view to avoid spending time rendering objects the player cannot see.
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
		colour = colornames.Green
	}
	sx, sy := camera.WorldToScreen(float32(a.x), float32(a.y))
	size := max(camera.Zoom, 0.5)
	vector.StrokeRect(screen, sx, sy, size, size, size, colour, false)

}

// checkDirection rejects invalid or unsuitable grid cells before applying the infection probability, 
// preventing infection from spreading through empty or non-healthy positions.
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

// isInfected provides one consistent check for the infected state so travel decisions and other behaviour use the same definition.
func (a *Agent) isInfected() bool {
	return a.state == StateInfected
}

// escape moves an agent away from the nearest infected neighbour so the simulation can model avoidance behaviour.
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

// find locates the nearest healthy neighbour within a limited area so an agent can choose a nearby target without scanning the entire world.
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
