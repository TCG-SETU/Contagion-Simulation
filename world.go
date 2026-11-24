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
	towns             [townCount]Town
}

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
		{cx: 255 + 53, cy: 165 + 53, radius: 53},
		{cx: 146 + 44, cy: 367 + 44, radius: 44},
		{cx: 409 + 58, cy: 531 + 58, radius: 58},
		{cx: 285 + 50, cy: 774 + 50, radius: 50},
		{cx: 726 + 44, cy: 642 + 44, radius: 44},
		{cx: 684 + 45, cy: 172 + 45, radius: 45},
		{cx: 505 + 36, cy: 287 + 36, radius: 36},
		{cx: 84 + 40, cy: 656 + 40, radius: 40},
	}

	// Build each town's Flow field
	for i, t := range w.towns {
		w.towns[i].flowField = w.m.BuildFlowField(t.cx, t.cy)
	}

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

		// Keep agent inside bounds
		newX, newY = clampBorders(newX, newY, w.w, w.h)
		newX, newY = clampToMap(w.m, oldX, oldY, newX, newY)

		// Only move agent
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

func (w *World) Counts() Counts {
	var c Counts

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

func (w *World) Agents() []*Agent {
	return w.agents
}

func (w *World) Done() bool {
	return w.agentStatsCount.Infected+w.agentStatsCount.Infectious == 0 ||
		w.agentStatsCount.Infected+w.agentStatsCount.Dead == w.agentsCount
}

// Helpers
func clamp(a, min, max int) int {
	if a < min {
		a = min
	}
	if a > max {
		a = max
	}
	return a
}

func clampBorders(x, y, w, h int) (int, int) {
	x = clamp(x, 0, w-1)
	y = clamp(y, 0, h-1)
	return x, y
}

// clamps x and y to safe locations on the map
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

func locationToDirection(a, b, positive, negative int) int {
	val := 0
	if a > b {
		val = positive
	} else if a < b {
		val = negative
	}
	return val
}
