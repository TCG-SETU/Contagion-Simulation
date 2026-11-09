package main

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
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
	agents            [][]*Agent
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
}

func (g *Game) step() {

	newAgents := make([][]*Agent, g.w)
	for i := range newAgents {
		newAgents[i] = make([]*Agent, g.w)
	}
	for x := 0; x < g.w; x++ {
		for y := 0; y < g.h; y++ {
			if g.agents[x][y] != nil {
				newAgents[x][y] = g.agents[x][y]
			}
		}
	}
	for x := 0; x < g.w; x++ {
		for y := 0; y < g.h; y++ {
			agent := g.agents[x][y]
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

			newX := agent.x + xStep
			newY := agent.y + yStep

			// Keep agent inside bounds
			newX, newY = clampBorders(newX, newY, g.w, g.h)

			// Only move agent
			if newAgents[newX][newY] == nil {
				agent.x = newX
				agent.y = newY
				newAgents[newX][newY] = agent
				newAgents[x][y] = nil
			}

			if agent.isInfected() {
				for dx := -1; dx <= 1; dx++ {
					for dy := -1; dy <= 1; dy++ {
						if dx != 0 || dy != 0 {
							g.checkDirection(g.agents, agent.x+dx, agent.y+dy)
						}
					}
				}
				agent.handleInfectionTimer(g)
			}
		}
	}

	g.agents = newAgents
}

func (g *Game) checkDirection(newAgents [][]*Agent, x int, y int) {
	if x < 0 || x >= g.w || y < 0 || y >= g.h {
		return
	}

	target := newAgents[x][y]
	if target == nil || target.state != StateHealthy {
		return
	}
	shouldInfect := rand.Float64() < g.infectionChance
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
		for _, agentState := range agent {
			if agentState != nil {
				agentState.draw(screen)
			}
		}
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
	if g.infectedCountNow == 0 || g.infectedCountNow == g.agentsCount {
		return nil
	}
	g.stepDelay--
	if g.stepDelay == 0 {
		g.step()
		g.lastStepTime = time.Now()
		g.iteration++
		g.stepDelay = 1
	}
	return nil
}

func newGame(width int, height int, agentCount int, infectedCount int, infectionChance float64, infectionDuration int, deathChance float64) *Game {
	g := Game{}
	agents := make([][]*Agent, width)
	for i := range width {
		agents[i] = make([]*Agent, height)
	}

	g.infectedCount = infectedCount

	for range agentCount {
		randX := rand.Intn(width)
		randY := rand.Intn(height)
		for agents[randX][randY] != nil {
			randX = rand.Intn(width)
			randY = rand.Intn(height)
		}
		newAgent := Agent{randX, randY, StateHealthy, 0, 0}

		if infectedCount != 0 {
			newAgent.state = StateInfected
			newAgent.infectionTimer = infectionDuration
			newAgent.timesInfected = 1
			g.infectedCountNow++
			infectedCount--
		}

		agents[randX][randY] = &newAgent

	}

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
			if g.agents[i][j] == nil {
				str.WriteString("  ")
			} else {
				if g.agents[i][j].isInfected() {
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
			if g.agents[i][j] != nil {
				endAgents++
				if g.agents[i][j].isInfected() {
					endInfected++
				}
			}
		}
	}

	return fmt.Sprintf(
		"Game Stats: Size: {%d, %d}  Agents: %d  Starting-Infected: %d  Infection Rate: %f  Infection Duration: %d \nEnd-Agents: %d End-Infected: %d",
		g.w, g.h, g.agentsCount, g.infectedCount, g.infectionChance, g.infectionDuration, endAgents, endInfected)

}
