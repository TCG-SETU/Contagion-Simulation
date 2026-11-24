package main

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// GameState represents the current state of the simulation.
type GameState int

const (
	// simulationState means the simulation is running normally.
	simulationState GameState = iota

	// pauseState means the simulation is paused.
	pauseState
)

// Movement direction constants
const (
	UP    = -1
	DOWN  = 1
	RIGHT = 1
	LEFT  = -1
)

// Screen dimensions constants
const (
	screenWidth  = 1080
	screenHeight = 1080
)

// Game constants
type Game struct {
	world *World
	cam   *Camera
}

// Draw renders the simulation and its statistic onto the screen.
func (g *Game) Draw(screen *ebiten.Image) {

	// Clear the screen with the background colour.
	screen.Fill(color.RGBA{R: 46, G: 7, B: 29, A: 0xFF})

	// Converts the top-left world coordinate into screen coordinates.
	x0, y0 := g.cam.WorldToScreen(0, 0)

	// Draws the boundary of the simulation world
	vector.StrokeRect(screen, x0, y0,
		float32(g.world.w)*g.cam.Zoom, float32(g.world.h)*g.cam.Zoom,
		1, color.RGBA{R: 80, G: 80, B: 80, A: 255}, false)

	// Draws the mapp and towns before drawing the agents.
	g.world.m.Draw(screen, g.cam, g.world.w, g.world.h)

	for _, t := range g.world.towns {
		t.Draw(screen, g.cam)
	}

	// Gets the current agent stats
	var c = g.world.agentStatsCount

	// Draws each agent and optionally hides those that are dead.
	for _, agent := range g.world.Agents() {
		if !(g.world.hideDead && agent.state == StateDead) {
			agent.Draw(screen, g.cam)
		}
	}

	// Displays a message when the simulation reaches either end condition.
	status := ""
	switch {
	case c.Infected == 0 && c.Infectious == 0:
		status = "\nEveryone is healthy!"
	case c.Healthy == 0 && c.Immune == 0:
		status = "\nEveryone is infected!"
	}

	// Displays a pause indicator when the simulation is paused
	paused := ""
	if g.world.gameState == pauseState {
		paused = "  [PAUSED - N to step]"
	}

	// Displays the simulation stats and performance information.
	ebitenutil.DebugPrint(screen, fmt.Sprintf(
		"TPS: %0.1f  FPS: %0.1f%s\n"+
			"Iteration: %d\n"+
			"Healthy:  %d\n"+
			"Infected: %d\n"+
			"Infectious: %d\n"+
			"Immune:   %d\n"+
			"Dead:     %d\n"+
			"Total:    %d%s",
		ebiten.ActualTPS(), ebiten.ActualFPS(), paused,
		g.world.iteration, c.Healthy, c.Infected, c.Infectious, c.Immune, c.Dead,
		c.Total(), status,
	))
}

// Layout defines the dimensions of the game screen.
func (g *Game) Layout(int, int) (int, int) {
	return screenWidth, screenHeight
}

// Update handles the camera based on the dimensions of the world.
func (g *Game) Update() error {

	// updates the camera based on the dimensions of the world.
	g.cam.Update(g.world.w, g.world.h)

	// Toggle between running and paused states when space is pressed
	if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		if g.world.gameState == simulationState {
			g.world.gameState = pauseState
		} else if g.world.gameState == pauseState {
			g.world.gameState = simulationState
		}
	}

	// Stop updating when the simulation is finished.
	if g.world.Done() {
		return nil
	}

	// When paused it allows the user to advance one stop when they press n.
	if g.world.gameState == pauseState {
		if inpututil.IsKeyJustPressed(ebiten.KeyN) {
			g.world.Step()
		}
		return nil
	}
	// Advance the simulation by one step while it is running
	g.world.Step()

	return nil
}

// newGame creates a new game using supplied configuration.
func newGame(cfg Config) *Game {
	g := Game{}
	g.world = New(cfg)
	g.cam = NewCamera(g.world.w, g.world.h)
	return &g
}

// Stats returns a formatted string containing the simulation's final stats.
func (g *Game) stats() string {

	c := g.world.Counts()

	return fmt.Sprintf(
		"Game Stats: Size: {%d, %d}  Agents: %d  Starting-Infected: %d  Infection Rate: %f  Infection Duration: %d \nEnd-Agents: %d End-Infected: %d",
		g.world.w, g.world.h, g.world.agentsCount, g.world.initialInfected, g.world.infectionChance, g.world.infectionDuration, g.world.agentsCount, c.Infected)

}
