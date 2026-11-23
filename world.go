package main

type Counts struct {
	Healthy, Infected, Dead, Immune int
}

type Config struct {
	Width, Height, Agents, InitialInfected int
	InfectionChance, DeathChance           float64
	InfectionDuration                      int
	Seed                                   int64
	HideDead                               bool
}

const (
	townCount         = 5
	travelChance      = .0005
	stayDuration      = 100
	infectedTravelBan = .0005
)

type World struct {
	grid              [][]*Agent
	agents            []*Agent
	agentsCount       int
	initialInfected   int
	infectedCountNow  int
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

	w.towns = [townCount]Town{
		{70, 66, 40},
		{155, 273, 37},
		{406, 299, 55},
		{300, 102, 40},
		{214, 150, 48}}

	w.initialInfected = cfg.InitialInfected
	w.hideDead = cfg.HideDead

	for i := range cfg.Agents {
		townIdx := i % townCount
		x, y := w.spawnPos(w.towns[townIdx])
		newAgent := Agent{x, y, StateHealthy, 0,
			0, townIdx, -1, Local, -1}

		if cfg.InitialInfected != 0 && townIdx == 0 {
			newAgent.state = StateInfected
			newAgent.infectionTimer = cfg.InfectionDuration
			newAgent.timesInfected = 1
			w.infectedCountNow++
			cfg.InitialInfected--
		}

		w.grid[x][y] = &newAgent
		agents[i] = &newAgent
	}

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
		if x >= 0 && x < w.w && y >= 0 && y < w.h && w.grid[x][y] == nil {
			return x, y
		}
	}
}

func (w *World) Step() {
	rng.Shuffle(len(w.agents), func(i, j int) { w.agents[i], w.agents[j] = w.agents[j], w.agents[i] })

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

		// Only move agent
		if w.grid[newX][newY] == nil {
			agent.x = newX
			agent.y = newY
			w.grid[newX][newY] = agent
			w.grid[oldX][oldY] = nil
		}

		if agent.isInfected() {
			agent.attemptInfection(w)
		}
	}
	w.iteration++

}

func (w *World) Counts() Counts {
	var c Counts

	for _, agent := range w.agents {
		switch agent.state {
		case StateHealthy:
			c.Healthy++
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
	return w.infectedCountNow == 0 || w.infectedCountNow == w.agentsCount
}
