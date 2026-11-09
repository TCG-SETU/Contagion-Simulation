package main

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

type GameState int

const (
	simulationState GameState = iota
	pauseState
)

const (
	UP    = -1
	DOWN  = 1
	RIGHT = 1
	LEFT  = -1
)

const (
	screenWidth  = 600
	screenHeight = 500
)

type Game struct {
	grid              [][]*Agent
	agents            []*Agent
	agentsCount       int
	infectedCount     int
	infectedCountNow  int
	h, w              int
	infectionChance   float64
	infectionDuration int
	iteration         int
	lastStepTime      time.Time
	stepDelay         int
	deathChance       float64
	gameState         GameState
}

func (g *Game) step() {

	rand.Shuffle(len(g.agents), func(i, j int) { g.agents[i], g.agents[j] = g.agents[j], g.agents[i] })

	for _, agent := range g.agents {
		if agent == nil || agent.state == StateDead {
			continue
		}

		xStep := rand.Intn(3) - 1
		yStep := rand.Intn(3) - 1

		if agent.isInfected() {
			targetX, targetY, found := agent.find(g)
			if found {
				xStep = locationToDirection(targetX, agent.x, RIGHT, LEFT)
				yStep = locationToDirection(targetY, agent.y, DOWN, UP)
			}
		}
		if agent.state == StateHealthy {
			bestMoveX, bestMoveY, found := agent.escape(g)
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
		newX, newY = clampBorders(newX, newY, g.w, g.h)

		// Only move agent
		if g.grid[newX][newY] == nil {
			agent.x = newX
			agent.y = newY
			g.grid[newX][newY] = agent
			g.grid[oldX][oldY] = nil
		}

		if agent.isInfected() {
			for dx := -1; dx <= 1; dx++ {
				for dy := -1; dy <= 1; dy++ {
					if dx != 0 || dy != 0 {
						g.checkDirection(g.grid, agent.x+dx, agent.y+dy)
					}
				}
			}
			agent.handleInfectionTimer(g)
		}
	}
	g.lastStepTime = time.Now()
	g.iteration++
}

func (g *Game) checkDirection(newAgents [][]*Agent, x int, y int) {
	if x < 0 || x >= g.w || y < 0 || y >= g.h {
		return
	}

	target := newAgents[x][y]
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

func (g *Game) Draw(screen *ebiten.Image) {

	screen.Clear()
	//screen.Fill(colornames.White)

	// Draw the agents to the screen
	for _, agent := range g.agents {
		agent.draw(screen)
	}
	ebitenutil.DebugPrint(screen, fmt.Sprintf("TPS: %0.2f\nFPS: %0.2f", ebiten.ActualTPS(), ebiten.ActualFPS()))
	ebitenutil.DebugPrint(screen, fmt.Sprintf("\n\nStep: %d", g.iteration))
	ebitenutil.DebugPrint(screen, fmt.Sprintf("\n\n\nInfected: %d", g.infectedCountNow))

	if g.infectedCountNow == 0 {
		ebitenutil.DebugPrint(screen, "\n\n\n\nEveryone is healthy!")
	}
	if g.infectedCountNow == g.agentsCount {
		ebitenutil.DebugPrint(screen, "\n\n\n\nEveryone is infected!")

	}
}

func (g *Game) Layout(int, int) (int, int) {
	return screenWidth, screenHeight
}

func (g *Game) Update() error {

	if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		if g.gameState == simulationState {
			g.gameState = pauseState
		} else if g.gameState == pauseState {
			g.gameState = simulationState
		}
	}

	if g.infectedCountNow == 0 || g.infectedCountNow == g.agentsCount {
		return nil
	}

	if g.gameState == pauseState {
		if inpututil.IsKeyJustPressed(ebiten.KeyN) {
			g.step()
		}
		return nil
	}
	g.step()

	return nil
}

func newGame(width int, height int, agentCount int, infectedCount int, infectionChance float64, infectionDuration int, deathChance float64) *Game {
	g := Game{}
	grid := make([][]*Agent, width)
	for i := range width {
		grid[i] = make([]*Agent, height)
	}
	agents := make([]*Agent, agentCount)

	g.infectedCount = infectedCount

	for i := range agentCount {
		randX := rng.Intn(width)
		randY := rng.Intn(height)
		for grid[randX][randY] != nil {
			randX = rng.Intn(width)
			randY = rng.Intn(height)
		}
		newAgent := Agent{randX, randY, StateHealthy, 0, 0}

		if infectedCount != 0 {
			newAgent.state = StateInfected
			newAgent.infectionTimer = infectionDuration
			newAgent.timesInfected = 1
			g.infectedCountNow++
			infectedCount--
		}

		grid[randX][randY] = &newAgent
		agents[i] = &newAgent
	}

	g.grid = grid
	g.agents = agents
	g.w = width
	g.h = height
	g.infectionChance = infectionChance
	g.agentsCount = agentCount
	g.infectionDuration = infectionDuration
	g.lastStepTime = time.Now()
	g.stepDelay = 5
	g.deathChance = deathChance
	return &g
}

func (g *Game) String() string {
	var str strings.Builder
	for i := range g.w {
		for j := range g.h {
			if g.grid[i][j] == nil {
				str.WriteString("  ")
			} else {
				if g.grid[i][j].isInfected() {
					str.WriteString(fmt.Sprintf("X "))
				} else {
					str.WriteString(fmt.Sprintf("O "))
				}
			}
		}
		str.WriteString("\n")
	}
	return str.String()
}

func (g *Game) stats() string {

	endAgents := 0
	endInfected := 0
	for i := 0; i < g.w; i++ {
		for j := 0; j < g.h; j++ {
			if g.grid[i][j] != nil {
				endAgents++
				if g.grid[i][j].isInfected() {
					endInfected++
				}
			}
		}
	}

	return fmt.Sprintf(
		"Game Stats: Size: {%d, %d}  Agents: %d  Starting-Infected: %d  Infection Rate: %f  Infection Duration: %d \nEnd-Agents: %d End-Infected: %d",
		g.w, g.h, g.agentsCount, g.infectedCount, g.infectionChance, g.infectionDuration, endAgents, endInfected)

}
