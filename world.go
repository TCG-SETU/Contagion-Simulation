package main

type Counts struct {
	Healthy, Infected, Dead, Immune int
}

type Config struct {
	Width, Height, Agents, InitialInfected int
	InfectionChance, DeathChance           float64
	InfectionDuration                      int
	Seed                                   int64
}

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
	gameState         GameState
}

func New(cfg Config) *World {

	if cfg.Agents > cfg.Width*cfg.Height {
		panic("agents count too large")
	}
	if cfg.InitialInfected > cfg.Agents {
		panic("initial infected count too large")
	}

	w := &World{}
	grid := make([][]*Agent, cfg.Width)
	for i := range cfg.Width {
		grid[i] = make([]*Agent, cfg.Height)
	}
	agents := make([]*Agent, cfg.Agents)

	w.initialInfected = cfg.InitialInfected

	for i := range cfg.Agents {
		randX := rng.Intn(cfg.Width)
		randY := rng.Intn(cfg.Height)
		for grid[randX][randY] != nil {
			randX = rng.Intn(cfg.Width)
			randY = rng.Intn(cfg.Height)
		}
		newAgent := Agent{randX, randY, StateHealthy, 0, 0}

		if cfg.InitialInfected != 0 {
			newAgent.state = StateInfected
			newAgent.infectionTimer = cfg.InfectionDuration
			newAgent.timesInfected = 1
			w.infectedCountNow++
			cfg.InitialInfected--
		}

		grid[randX][randY] = &newAgent
		agents[i] = &newAgent
	}

	w.grid = grid
	w.agents = agents
	w.w = cfg.Width
	w.h = cfg.Height
	w.infectionChance = cfg.InfectionChance
	w.agentsCount = cfg.Agents
	w.infectionDuration = cfg.InfectionDuration
	w.stepDelay = 5
	w.deathChance = cfg.DeathChance
	return w
}

func (w *World) Step() {
	rng.Shuffle(len(w.agents), func(i, j int) { w.agents[i], w.agents[j] = w.agents[j], w.agents[i] })

	for _, agent := range w.agents {
		if agent == nil || agent.state == StateDead {
			continue
		}

		xStep := rng.Intn(3) - 1
		yStep := rng.Intn(3) - 1

		if agent.isInfected() {
			targetX, targetY, found := agent.find(w)
			if found {
				xStep = locationToDirection(targetX, agent.x, RIGHT, LEFT)
				yStep = locationToDirection(targetY, agent.y, DOWN, UP)
			}
		}
		if agent.state == StateHealthy {
			bestMoveX, bestMoveY, found := agent.escape(w)
			if found {
				if bestMoveX > agent.x {
					xStep = RIGHT
				} else if bestMoveX < agent.x {
					xStep = LEFT
				} else {
					xStep = 0
				}

				if bestMoveY > agent.y {
					yStep = DOWN
				} else if bestMoveY < agent.y {
					yStep = UP
				} else {
					yStep = 0
				}
			}
		}

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
			for dx := -1; dx <= 1; dx++ {
				for dy := -1; dy <= 1; dy++ {
					if dx != 0 || dy != 0 {
						agent.checkDirection(w, agent.x+dx, agent.y+dy)
					}
				}
			}
			agent.handleInfectionTimer(w)
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
