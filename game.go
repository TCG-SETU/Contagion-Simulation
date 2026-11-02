package main

import (
	"fmt"
	"math/rand"
	"strings"
)

const (
	UP    = -1
	DOWN  = 1
	RIGHT = 1
	LEFT  = -1
)

type Game struct {
	agents            [][]*Agent
	agentsCount       int
	infectedCount     int
	infectedCountNow  int
	h, w              int
	infectionChance   float64
	infectionDuration int
}

func (g *Game) step(step int) {
	fmt.Printf("Game %d steps\n", step)

	newAgents := g.agents

	for x := 0; x < g.w; x++ {
		for y := 0; y < g.h; y++ {
			agent := g.agents[x][y]

			if agent == nil {
				continue
			}

			xStep := rand.Intn(3) - 1
			yStep := rand.Intn(3) - 1

			if agent.infected {
				targetX, targetY, found := g.find(agent.x, agent.y)
				if found {
					xStep = locationToDirection(targetX, agent.x, RIGHT, LEFT)
					yStep = locationToDirection(targetY, agent.y, DOWN, UP)
				}
			}
			if !agent.infected {
				bestMoveX, bestMoveY, found := g.escape(agent.x, agent.y)
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

			if agent.infected {
				agent.handleInfectionTimer(g)
				// check up down left right
				g.checkDirection(newAgents, newX+LEFT, newY)
				g.checkDirection(newAgents, newX, newY+UP)
				g.checkDirection(newAgents, newX+RIGHT, newY)
				g.checkDirection(newAgents, newX, newY+DOWN)
				// diagonals
				g.checkDirection(newAgents, newX+LEFT, newY+UP)
				g.checkDirection(newAgents, newX+RIGHT, newY+UP)
				g.checkDirection(newAgents, newX+RIGHT, newY+DOWN)
				g.checkDirection(newAgents, newX+LEFT, newY+DOWN)
			}
		}
	}

	fmt.Printf("%d infected\n", g.infectedCountNow)

	g.agents = newAgents
}

func (g *Game) checkDirection(newAgents [][]*Agent, x int, y int) {
	if x < 0 || x >= g.w || y < 0 || y >= g.h {
		return
	}

	target := newAgents[x][y]
	if target == nil || target.infected {
		return
	}
	shouldInfect := rand.Float64() < g.infectionChance
	if shouldInfect {
		target.infected = true
		g.infectedCountNow++
	}
}

func newGame(width int, height int, agentCount int, infectedCount int, infectionChance float64, infectionDuration int) *Game {
	g := Game{}
	agents := make([][]*Agent, width)
	for i := range width {
		agents[i] = make([]*Agent, height)
	}

	g.infectedCount = infectedCount

	for range agentCount {
		randX := rand.Intn(width)
		randY := rand.Intn(height)
		newAgent := Agent{randX, randY, false, 0}

		if infectedCount != 0 {
			newAgent.infected = true
			newAgent.infectionTimer = infectionDuration
			g.infectedCountNow++
			infectedCount--
		} else {
			newAgent.infected = false
		}
		agents[randX][randY] = &newAgent

	}

	g.agents = agents
	g.w = width
	g.h = height
	g.infectionChance = infectionChance
	g.agentsCount = agentCount
	g.infectionDuration = infectionDuration
	return &g
}

func (g *Game) String() string {
	var str strings.Builder
	for i := range g.w {
		for j := range g.h {
			if g.agents[i][j] == nil {
				str.WriteString("  ")
			} else {
				if g.agents[i][j].infected {
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
				if g.agents[i][j].infected {
					endInfected++
				}
			}
		}
	}

	return fmt.Sprintf(
		"Game Stats: Size: {%d, %d}  Agents: %d  Starting-Infected: %d  Infection Rate: %f  Infection Duration: %d \nEnd-Agents: %d End-Infected: %d",
		g.w, g.h, g.agentsCount, g.infectedCount, g.infectionChance, g.infectionDuration, endAgents, endInfected)

}

func (g *Game) find(x, y int) (int, int, bool) {
	minDist := 99
	targetX, targetY := -1, -1
	found := false

	for dx := -2; dx <= 2; dx++ {
		for dy := -2; dy <= 2; dy++ {
			nx, ny := x+dx, y+dy

			if nx >= 0 && nx < g.w && ny >= 0 && ny < g.h && (dx != 0 || dy != 0) {
				target := g.agents[nx][ny]

				if target != nil && !target.infected {
					dist := dx*dx + dy*dy
					if dist < minDist {
						minDist = dist
						targetX, targetY = nx, ny
						found = true
					}
				}
			}
		}
	}
	return targetX, targetY, found

}
func (g *Game) escape(x, y int) (int, int, bool) {
	closestInfectedDist := 999
	infX, infY := -1, -1
	foundInfected := false

	scanRadius := 3
	for dx := -scanRadius; dx <= scanRadius; dx++ {
		for dy := -scanRadius; dy <= scanRadius; dy++ {
			nx, ny := x+dx, y+dy
			if nx >= 0 && nx < g.w && ny >= 0 && ny < g.h && (dx != 0 || dy != 0) {
				target := g.agents[nx][ny]

				if target != nil && target.infected {
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
		return x, y, false
	}

	bestMoveX, bestMoveY := x, y
	maxDistFromInfected := -1

	for moveDX := -1; moveDX <= 1; moveDX++ {
		for moveDY := -1; moveDY <= 1; moveDY++ {
			nx, ny := x+moveDX, y+moveDY

			if nx >= 0 && nx < g.w && ny >= 0 && ny < g.h {
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
