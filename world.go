package main

type Config struct {
	Width, Height, Agents, InitialInfected int
	InfectionChance, DeathChance           float64
	InfectionDuration                      int
	Seed                                   int64
	HideDead                               bool
}

const contagiousTimer = 100

type World struct {
	grid              [][]*Agent
	agents            []*Agent
	m                 *Map
	agentsCount       int
	initialInfected   int
	agentStatsCount   Counts
	h, w              int
	infectionChance   float64
	infectionDuration int
	iteration         int
	stepDelay         int
	deathChance       float64
	hideDead          bool
	gameState         GameState
	towns             [5]Town
}

// New validates the requested population and builds all world data in one place so the simulation starts with consistent dimensions, agents, towns, and counts.
func New(cfg Config) *World {
	
	if cfg.Agents > cfg.Width*cfg.Height {
		panic("agents count too large")
	}

	if cfg.InitialInfected > cfg.Agents {
		panic("initial infected count too large")
	}

	w := &World{}
	w.w = cfg.Width
	w.h = cfg.Height

	grid := make([][]*Agent, cfg.Width)
	for i := range cfg.Width {
		grid[i] = make([]*Agent, cfg.Height)
	}
	w.grid = grid

	agents := make([]*Agent, cfg.Agents)

	w.m = &Map{
		Image:     "images/Springfield_Island.jpg",
		ImageMask: "images/Springfield_Island_mask.jpg",
	}
	w.m.LoadMask()

	w.towns = [townCount]Town{
		{146 + 44, 367 + 44, 44},
		{409 + 58, 531 + 58, 58},
		{285 + 50, 774 + 50, 50},
		{726 + 44, 642 + 44, 44},
		{684 + 45, 172 + 45, 45}}

	w.initialInfected = cfg.InitialInfected
	w.hideDead = cfg.HideDead
	w.agentStatsCount = Counts{}

	for i := range cfg.Agents {
		townIdx := i % townCount
		x, y := w.spawnPos(w.towns[townIdx])

		newAgent := Agent{x, y, StateHealthy, 0,
			0, 0, townIdx, -1, Local, -1}

		if cfg.InitialInfected != 0 && townIdx == 0 {
			newAgent.state = StateInfected
			newAgent.infectionTimer = cfg.InfectionDuration
			newAgent.contagiousTimer = contagiousTimer
			newAgent.timesInfected = 1
			w.agentStatsCount.Infected++
			cfg.InitialInfected--
		}

		w.grid[x][y] = &newAgent
		agents[i] = &newAgent
	}

	w.agentStatsCount.Healthy = cfg.Agents - w.agentStatsCount.Infected
	w.agents = agents
	w.infectionChance = cfg.InfectionChance
	w.agentsCount = cfg.Agents
	w.infectionDuration = cfg.InfectionDuration
	w.stepDelay = 5
	w.deathChance = cfg.DeathChance
	return w
}

// spawnPos samples around a town centre and accepts only safe, unoccupied cells so agents begin inside their assigned town without overlapping.
func (w *World) spawnPos(t Town) (int, int) {
	sigma := float64(t.radius) / 2
	for {
		x := t.cx + int(rng.NormFloat64()*sigma)
		y := t.cy + int(rng.NormFloat64()*sigma)
		if x >= 0 && x < w.w && y >= 0 && y < w.h && w.grid[x][y] == nil && w.m.IsSafe(x, y) {
			return x, y
		}
	}
}

// Step randomises update order to reduce ordering bias, then moves and updates each living agent before applying population-count changes once per iteration.
func (w *World) Step() {
	rng.Shuffle(len(w.agents), func(i, j int) { w.agents[i], w.agents[j] = w.agents[j], w.agents[i] })
	countsDelta := Counts{}
	for _, agent := range w.agents {
		if agent == nil || agent.state == StateDead {
			continue
		}

		agent.updateTravel(w)
		xStep, yStep := agent.wander(w)

		oldX, oldY := agent.x, agent.y
		newX := agent.x + xStep
		newY := agent.y + yStep

		newX, newY = clampBorders(newX, newY, w.w, w.h)
		newX, newY = clampToMap(w.m, oldX, oldY, newX, newY)

		if w.grid[newX][newY] == nil {
			agent.x = newX
			agent.y = newY
			w.grid[newX][newY] = agent
			w.grid[oldX][oldY] = nil
		}

		if agent.isInfected() || agent.state == StateInfectious {
			agent.attemptInfection(w, &countsDelta)
		}
	}
	w.agentStatsCount.Update(countsDelta)
	w.iteration++

}

// Counts recalculates the population distribution from the agents so the returned statistics reflect their actual states.
func (w *World) Counts() Counts {
	var c Counts

	// counts agents based on their health state.
	for _, agent := range w.agents {
		switch agent.state {
		case StateHealthy:
			c.Healthy++
		case StateInfectious:
			c.Infectious++
		case StateInfected:
			c.Infected++
		case StateDead:
			c.Dead++
		case StateImmune:
			c.Immune++
		default:
			panic("unhandled default case")
		}
	}

	return c
}

// Agents exposes the world’s agent slice so rendering and other callers can iterate over the population.
func (w *World) Agents() []*Agent {
	return w.agents
}

// Done checks the simulation’s terminal conditions so execution can stop once infection has ended or the population has reached the defined outcome.
func (w *World) Done() bool {
	return w.agentStatsCount.Infected+w.agentStatsCount.Infectious == 0 ||
		w.agentStatsCount.Infected+w.agentStatsCount.Dead == w.agentsCount
}

// Helpers
// clamp bounds a value so callers can safely constrain coordinates.
func clamp(a, min, max int) int {
	if a < min {
		a = min
	}
	if a > max {
		a = max
	}
	return a
}

// clampBorders keeps movement inside the grid so agents cannot be placed beyond the world’s edges.
func clampBorders(x, y, w, h int) (int, int) {
	x = clamp(x, 0, w-1)
	y = clamp(y, 0, h-1)
	return x, y
}

// clampToMap tests movement along each axis separately so an agent can move around an obstacle instead of becoming stuck whenever a diagonal step is blocked.
func clampToMap(m *Map, oldX, oldY, newX, newY int) (int, int) {
	if !m.IsSafe(oldX, oldY) {
		panic("old x and y must be safe")
	}

	x := oldX
	y := oldY
	if m.IsSafe(newX, oldY) {
		x = newX
	}
	if m.IsSafe(oldX, newY) {
		y = newY
	}
	if m.IsSafe(x, y) {
		return x, y
	}
	return oldX, oldY
}

// locationToDirection converts a coordinate comparison into a one-cell direction so movement code can steer without duplicating directional logic.
func locationToDirection(a, b, positive, negative int) int {
	val := 0
	if a > b {
		val = positive
	} else if a < b {
		val = negative
	}
	return val
}
