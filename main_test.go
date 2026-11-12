package main

import (
	"strings"
	"testing"
)

// Test that the game creates the correct number of agents.
func TestNewGameAgentCount(t *testing.T) {
	game := newGame(10, 10, 70, 10, 0.1, 7)

	count := 0

	for x := 0; x < game.w; x++ {
		for y := 0; y < game.h; y++ {
			if game.agents[x][y] != nil {
				count++
			}
		}
	}

	if count != 70 {
		t.Errorf("Expected 70 agents, got %d", count)
	}
}

// Test that the game creates 10 infected agents.
func TestNewGameInfectedCount(t *testing.T) {
	game := newGame(10, 10, 70, 10, 0.1, 7)

	count := 0

	for x := 0; x < game.w; x++ {
		for y := 0; y < game.h; y++ {
			agent := game.agents[x][y]

			if agent != nil && agent.infected {
				count++
			}
		}
	}

	if count != 10 {
		t.Errorf("Expected 10 infected agents, got %d", count)
	}
}

// Test that a healthy agent becomes infected.
func TestCheckDirection(t *testing.T) {
	game := newGame(5, 5, 0, 0, 1.0, 7)

	game.agents[0][0] = &Agent{infected: true}
	game.agents[1][0] = &Agent{infected: false}

	game.checkDirection(game.agents, 1, 0)

	if game.agents[1][0].infected == false {
		t.Error("Expected agent to become infected")
	}
}

// Test that an agent stays healthy when infection chance is zero.
func TestInfectionChanceZero(t *testing.T) {
	game := newGame(5, 5, 0, 0, 0.0, 7)

	game.agents[0][0] = &Agent{infected: true}
	game.agents[1][0] = &Agent{infected: false}

	game.checkDirection(game.agents, 1, 0)

	if game.agents[1][0].infected == true {
		t.Error("Expected agent to stay healthy")
	}
}

// Test that the number of agents stays the same after a step.
func TestAgentCountAfterStep(t *testing.T) {
	game := newGame(10, 10, 70, 10, 0.1, 7)

	game.step(1)

	count := 0

	for x := 0; x < game.w; x++ {
		for y := 0; y < game.h; y++ {
			if game.agents[x][y] != nil {
				count++
			}
		}
	}

	if count != 70 {
		t.Errorf("Expected 70 agents, got %d", count)
	}
}

// Test that agents stay inside the grid.
func TestClampBorders(t *testing.T) {
	x, y := clampBorders(-1, 12, 10, 10)

	if x != 0 || y != 9 {
		t.Errorf("Expected 0 and 9, got %d and %d", x, y)
	}
}

// Test that an infected agent can find a healthy agent.
func TestFindHealthyAgent(t *testing.T) {
	game := newGame(5, 5, 0, 0, 0.1, 7)

	game.agents[2][2] = &Agent{infected: true}
	game.agents[3][2] = &Agent{infected: false}

	x, y, found := game.find(2, 2)

	if found == false {
		t.Error("Expected to find a healthy agent")
	}

	if x != 3 || y != 2 {
		t.Errorf("Expected position 3, 2, got %d, %d", x, y)
	}
}

// Test that a healthy agent moves away from an infected agent.
func TestEscapeFromInfectedAgent(t *testing.T) {
	game := newGame(5, 5, 0, 0, 0.1, 7)

	game.agents[2][2] = &Agent{infected: false}
	game.agents[3][2] = &Agent{infected: true}

	x, y, found := game.escape(2, 2)

	if found == false {
		t.Error("Expected to find an infected agent")
	}

	distanceBefore := 1
	distanceAfter := (x-3)*(x-3) + (y-2)*(y-2)

	if distanceAfter <= distanceBefore {
		t.Error("Expected agent to move farther away")
	}
}

// Test that the game statistics show the correct numbers.
func TestGameStats(t *testing.T) {
	game := newGame(2, 2, 0, 0, 0.1, 7)

	game.agents[0][0] = &Agent{infected: true}
	game.agents[1][1] = &Agent{infected: false}

	result := game.stats()

	if !strings.Contains(result, "End-Agents: 2") {
		t.Error("Expected 2 agents in stats")
	}

	if !strings.Contains(result, "End-Infected: 1") {
		t.Error("Expected 1 infected agent in stats")
	}
}
