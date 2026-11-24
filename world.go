package main

// Config contains the settings used to initialise a simulation.
type Config struct {
	Width, Height, Agents, InitialInfected int
	InfectionChance, DeathChance           float64
	InfectionDuration                      int
	Seed                                   int64
	HideDead                               bool
}

// ContagiousTimer says how long an infected agents will remain contagious.
const contagiousTimer = 100

// World represents the simulation environment and manages the agents,
// map, towns, infection stats and simulation state.
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

// New creates and initialises a world using a supplied config
//
// It creates the agent grid, loads the app, sets up the towns,
// spawns the requested number of agents.
func New(cfg Config) *World {

	// checks that the number of requested agents will fit inside the world.
	if cfg.Agents > cfg.Width*cfg.Height {
		panic("agents count too large")
	}

	// Makes sure the initial infected agents count does not exceed the agent count.
	if cfg.InitialInfected > cfg.Agents {
		panic("initial infected count too large")
	}

	w := &World{}
	w.w = cfg.Width
	w.h = cfg.Height

	// Creates a grid to keep track of agent positions.
	grid := make([][]*Agent, cfg.Width)
	for i := range cfg.Width {
		grid[i] = make([]*Agent, cfg.Height)
	}
	w.grid = grid

	// Creates the slice that stores all of the agents.
	agents := make([]*Agent, cfg.Agents)

	// Loads the map image and its safety mask.
	w.m = &Map{
		Image:     "images/Springfield_Island.jpg",
		ImageMask: "images/Springfield_Island_mask.jpg",
	}
	w.m.LoadMask()

	// Defines the locations and sizes of the towns.
	w.towns = [townCount]Town{
		{146 + 44, 367 + 44, 44},
		{409 + 58, 531 + 58, 58},
		{285 + 50, 774 + 50, 50},
		{726 + 44, 642 + 44, 44},
		{684 + 45, 172 + 45, 45}}

	w.initialInfected = cfg.InitialInfected
	w.hideDead = cfg.HideDead
	w.agentStatsCount = Counts{}

	// Spawns each agent in one of the towns.
	for i := range cfg.Agents {
		townIdx := i % townCount
		x, y := w.spawnPos(w.towns[townIdx])

		// Creates a new healthy agent at the selected position.
		newAgent := Agent{x, y, StateHealthy, 0,
			0, 0, townIdx, -1, Local, -1}

		// Infect the requested number of agents in the first town.
		if cfg.InitialInfected != 0 && townIdx == 0 {
			newAgent.state = StateInfected
			newAgent.infectionTimer = cfg.InfectionDuration
			newAgent.contagiousTimer = contagiousTimer
			newAgent.timesInfected = 1
			w.agentStatsCount.Infected++
			cfg.InitialInfected--
		}

		// Stores the agent in both the grid and agent slice
		w.grid[x][y] = &newAgent
		agents[i] = &newAgent
	}

	// Calculates the number of agents that are initially healthy.
	w.agentStatsCount.Healthy = cfg.Agents - w.agentStatsCount.Infected
	w.agents = agents
	w.infectionChance = cfg.InfectionChance
	w.agentsCount = cfg.Agents
	w.infectionDuration = cfg.InfectionDuration
	w.stepDelay = 5
	w.deathChance = cfg.DeathChance
	return w
}

// spawnPos finds an available safe positions inside the specified town.
//
// It uses a normal distributions to keep spawned agents close to the centre of town.
func (w *World) spawnPos(t Town) (int, int) {
	sigma := float64(t.radius) / 2
	for {
		// Generate a random position around the town centre.
		x := t.cx + int(rng.NormFloat64()*sigma)
		y := t.cy + int(rng.NormFloat64()*sigma)
		// return the position if it is inside the world, unoccupied, and safe.
		if x >= 0 && x < w.w && y >= 0 && y < w.h && w.grid[x][y] == nil && w.m.IsSafe(x, y) {
			return x, y
		}
	}
}

// Step advances the simulation by one iteration.
//
// Agents are shuffled before being updated to randomise their update order.
// Each living agent moves, updates it's travel state and attempts to infect
// nearby agents when infectious.
func (w *World) Step() {
	// randomise the order in which agents are updated
	rng.Shuffle(len(w.agents), func(i, j int) { w.agents[i], w.agents[j] = w.agents[j], w.agents[i] })
	countsDelta := Counts{}
	for _, agent := range w.agents {
		// Skip empty agent entries and dead agents.
		if agent == nil || agent.state == StateDead {
			continue
		}

		// Update the agent's travel behaviour and determine its movement.
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

// Counts calculates and returns the current number of agents in each state.
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

// Agents return the slice containing all agents in the world.
func (w *World) Agents() []*Agent {
	return w.agents
}

// Done checks whether the simulation has reached an ending condition.
//
// the simulation ends when there are no infected or infectious agents,
// or when  all agents are either infected or dead.
func (w *World) Done() bool {
	return w.agentStatsCount.Infected+w.agentStatsCount.Infectious == 0 ||
		w.agentStatsCount.Infected+w.agentStatsCount.Dead == w.agentsCount
}

// Helpers
// clamp restricts a value to the specified minium and maximum
func clamp(a, min, max int) int {
	if a < min {
		a = min
	}
	if a > max {
		a = max
	}
	return a
}

// clampBorders keeps the given coordinates inside the world boundaries.
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

// locationToDirection determines the directions from one coordinate to another.
func locationToDirection(a, b, positive, negative int) int {
	val := 0
	if a > b {
		val = positive
	} else if a < b {
		val = negative
	}
	return val
}
