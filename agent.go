package main

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"golang.org/x/image/colornames"
)

// HealthState represents the current health condition of an agent
type HealthState int

const (
	// StateHealthy represents an agent that has not been infected.
	StateHealthy HealthState = iota
	// StateInfectious represents an agent that is currently contagious.
	StateInfectious
	// StateInfected represents an agent that is infected but not contagious.
	StateInfected
	// StateImmune represents an agent that has been recovered from multiple infections.
	StateImmune
	// StateDead represents an agent that has died from the infection.
	StateDead
)

// TravelMode represents the travel behavior of the agent.
type TravelMode int

const (
	Local     TravelMode = iota // wandering around home.
	Outbound                    // heading to another town.
	Visiting                    // wandering around the destination.
	Returning                   // heading home.
)

// Agent represents an individual within the simulation
// Each Agent has a position, health state, infection timer and travel information.
type Agent struct {
	x, y            int
	state           HealthState
	timesInfected   int
	contagiousTimer int
	infectionTimer  int

	home      int // Index of the agent's hometown in the world.
	dest      int // Index of the agent's destination town.
	mode      TravelMode
	stayTimer int // remaining time before the agent leaves its destination.
}

// Anchor returns the town that an agent should currently move around or travel towards.
// Agents that are traveling outbound or visiting use their destination town.
// All other agents use their hometown.
func (a *Agent) anchor(w *World) Town {
	if a.mode == Outbound || a.mode == Visiting {
		return w.towns[a.dest]
	}
	return w.towns[a.home]
}

// inTown checks whether an agent is currently within the radius of a specified town.
// It uses squared radius to avoid an unnecessary square root calculation.
func (a *Agent) inTown(t Town) bool {
	dx, dy := a.x-t.cx, a.y-t.cy
	return dx*dx+dy*dy <= int(t.radius*t.radius)
}

// updateTravel controls an agents movement between town's
// it handles travel decision, visiting durations and returning home.
func (a *Agent) updateTravel(w *World) {
	switch a.mode {

	case Local:
		// checks whether the agent decides to travel
		if rng.Float64() >= travelChance {
			return
		}
		// symptomatic agents are less likely to travel.
		if a.isInfected() && rng.Float64() < infectedTravelBan {
			return
		}
		// selects a random town that is not the agent's hometown.
		a.dest = rng.Intn(len(w.towns) - 1)
		if a.dest >= a.home { // pick any town except home.
			a.dest++
		}
		a.mode = Outbound
	case Outbound:
		// checks whether the agent has reached their destination.
		if a.inTown(w.towns[a.dest]) {
			a.mode = Visiting
			// sets a random visiting duration
			a.stayTimer = stayDuration/2 + rng.Intn(stayDuration)
		}
	case Visiting:
		// decreases the remaining visiting time.
		a.stayTimer--
		// starts traveling home when the visiting time expires.
		if a.stayTimer <= 0 {
			a.mode = Returning
		}
	case Returning:
		// Return to local movement state once home
		if a.inTown(w.towns[a.home]) {
			a.mode = Local
		}
	}
}

// wander calculates the direction that an agent should move.
// Agents traveling between towns generally move towns their destination.
// Local and visiting agents wander randomly but move towards their town if outside its radius
func (a *Agent) wander(w *World) (int, int) {
	t := a.anchor(w)

	// Generates a random movement direction.
	dx, dy := rng.Intn(3)-1, rng.Intn(3)-1

	switch a.mode {
	case Outbound, Returning:
		// Usually moves towards the destination town.
		if rng.Float64() < 0.8 {
			dx = locationToDirection(t.cx, a.x, RIGHT, LEFT)
			dy = locationToDirection(t.cy, a.y, DOWN, UP)
		}
	default: // Local, Visiting
		// move towards the town if outside its radius.
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
	// updates the agent's infection timer.
	a.handleInfectionTimer(w, c)
}

// handleInfectionTimer manages the progression of an infection.
// it handles the contagious period, death chance, recovery and immunity.
func (a *Agent) handleInfectionTimer(w *World, c *Counts) {
	// checks whether the agent is still contagious.
	if a.state == StateInfectious {
		a.contagiousTimer--
		// End the contagious period when the timer expires.
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
	// Decreases the remaining infection duration.
	a.infectionTimer--
	// handles recovery when the infection timer expires.
	if a.infectionTimer <= 0 {
		a.timesInfected++
		// handles recovery after their second infection.
		if a.timesInfected >= 2 {
			a.state = StateImmune
			c.Immune++
		} else {
			// Otherwise the agent will return to a healthy state
			a.state = StateHealthy
			c.Healthy++
		}
		c.Infected--
	}
}

// Draw renders an agent on the screen using its current health state.
// Agents outside the Camara's visible area are not drawn.
func (a *Agent) Draw(screen *ebiten.Image, camera *Camera) {
	// skips rendering if the agent is outside the camera's view
	if !camera.Visible(float32(a.x), float32(a.y), 1) {
		return
	}
	// assigns a colour based on the agent's health state
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

	// convert the world's coordinates into screen coordinates.
	sx, sy := camera.WorldToScreen(float32(a.x), float32(a.y))
	// calculates the agent's size ensuring it remains visible when zooming out.
	size := max(camera.Zoom, 0.5) // never smaller than 1 pixel.
	// Draws the agent as a coloured based retangle.
	vector.StrokeRect(screen, sx, sy, size, size, size, colour, false)

}

// checkDirection checks whether an agent at a given gird position can become infected.
// Infection only occurs if the target position contains a healthy agent.
// The Infection chance is determined by the world's infection probabilit.
func checkDirection(w *World, x int, y int, c *Counts) {

	// Ignores positions outside the world's boundaries.
	if x < 0 || x >= w.w || y < 0 || y >= w.h {
		return
	}

	// Get an agent's occupying the target grid position.
	target := w.grid[x][y]

	// Ignores empty positions and agents that are not health.
	if target == nil || target.state != StateHealthy {
		return
	}

	// randomly determines whether the infection is successful.
	shouldInfect := rng.Float64() < w.infectionChance
	if shouldInfect {
		// changes the target agent's health state
		target.state = StateInfectious

		// Initialise the infection timer and contagious timer.
		target.infectionTimer = w.infectionDuration
		target.contagiousTimer = contagiousTimer

		// Updates the population counts.
		c.Infectious++
		c.Healthy--
	}
}

// isInfected checks whether an agent is currently infected health state.
func (a *Agent) isInfected() bool {
	return a.state == StateInfected
}

// escape searches for nearby infection agents and calculates a movement direction.
// that increases the distance between the current agent and the closest infected agent.
// It returns the new X and Y coordinates and whether an infected agent was found.
func (a *Agent) escape(w *World) (int, int, bool) {
	// Initialise the closest infected agent search
	closestInfectedDist := 999
	infX, infY := -1, -1
	foundInfected := false

	// searches the surrounding area within in the radius of three grid positions.
	scanRadius := 3
	for dx := -scanRadius; dx <= scanRadius; dx++ {
		for dy := -scanRadius; dy <= scanRadius; dy++ {
			nx, ny := a.x+dx, a.y+dy

			// checks that the position is inside the world boundaries.
			if nx >= 0 && nx < w.w && ny >= 0 && ny < w.h && (dx != 0 || dy != 0) {
				target := w.grid[nx][ny]

				// checks whether the position contains an infected agent.
				if target != nil && target.state == StateInfected {
					dist := dx*dx + dy*dy

					// keeps tracks of the closest infected agent.
					if dist < closestInfectedDist {
						closestInfectedDist = dist
						infX, infY = nx, ny
						foundInfected = true
					}
				}
			}
		}
	}

	// No Infected agents were found nearby.
	if !foundInfected {
		return a.x, a.y, false
	}

	// Initialise the best movement position.
	bestMoveX, bestMoveY := a.x, a.y
	maxDistFromInfected := -1

	// checks all possible movements around the agent.
	for moveDX := -1; moveDX <= 1; moveDX++ {
		for moveDY := -1; moveDY <= 1; moveDY++ {
			nx, ny := a.x+moveDX, a.y+moveDY

			// Ignores positions outside the world boundaries.
			if nx >= 0 && nx < w.w && ny >= 0 && ny < w.h {

				// calculates the distance from the closest infected agent.
				distToInfected := (nx-infX)*(nx-infX) + (ny-infY)*(ny-infY)

				// selects the position that maximizes the distance.
				if distToInfected > maxDistFromInfected {
					maxDistFromInfected = distToInfected
					bestMoveX, bestMoveY = nx, ny
				}
			}
		}
	}

	return bestMoveX, bestMoveY, true
}

// find searches the surrounding five by five grid area for the healthy agent.
// It returns the coordinates of the closest healthy agent and whether one was found.
func (a *Agent) find(w *World) (int, int, bool) {

	// Initialise the closest distance.
	bestDist := math.MaxInt
	tx, ty, found := 0, 0, false

	// searches positions within the two grid spaces of the current agent
	for dx := -2; dx <= 2; dx++ {
		for dy := -2; dy <= 2; dy++ {
			nx, ny := a.x+dx, a.y+dy

			// skips  the current position and positions outside the world.
			if (dx == 0 && dy == 0) || nx < 0 || nx >= w.w || ny < 0 || ny >= w.h {
				continue
			}

			//  checks whether the position contains a healthy agent
			if t := w.grid[nx][ny]; t != nil && t.state == StateHealthy {

				// calculates the distance and keep the closest healthy agent.
				if d := dx*dx + dy*dy; d < bestDist {
					bestDist, tx, ty, found = d, nx, ny, true
				}
			}
		}
	}
	return tx, ty, found
}
