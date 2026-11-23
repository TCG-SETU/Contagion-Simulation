package main

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
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
	screenWidth  = 400
	screenHeight = 400
)

type Game struct {
	world *World
	cam   *Camera
}

func (g *Game) Draw(screen *ebiten.Image) {

	screen.Fill(color.RGBA{R: 46, G: 7, B: 29, A: 0xFF})

	x0, y0 := g.cam.WorldToScreen(0, 0)
	vector.StrokeRect(screen, x0, y0,
		float32(g.world.w)*g.cam.Zoom, float32(g.world.h)*g.cam.Zoom,
		1, color.RGBA{R: 80, G: 80, B: 80, A: 255}, false)

	g.world.m.Draw(screen, g.cam, g.world.w, g.world.h)

	var c = Counts{}

	for _, agent := range g.world.Agents() {
		if !(g.world.hideDead && agent.state == StateDead) {
			agent.Draw(screen, g.cam)
		}

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

	status := ""
	switch {
	case c.Infected == 0:
		status = "\nEveryone is healthy!"
	case c.Healthy == 0 && c.Dead == 0:
		status = "\nEveryone is infected!"
	}

	paused := ""
	if g.world.gameState == pauseState {
		paused = "  [PAUSED - N to step]"
	}

	ebitenutil.DebugPrint(screen, fmt.Sprintf(
		"TPS: %0.1f  FPS: %0.1f%s\n"+
			"Iteration: %d\n"+
			"Healthy:  %d\n"+
			"Infected: %d\n"+
			"Immune:   %d\n"+
			"Dead:     %d\n"+
			"Total:    %d%s",
		ebiten.ActualTPS(), ebiten.ActualFPS(), paused,
		g.world.iteration, c.Healthy, c.Infected, c.Immune, c.Dead,
		c.Healthy+c.Infected+c.Dead+c.Immune, status,
	))
}

func (g *Game) Layout(int, int) (int, int) {
	return screenWidth, screenHeight
}

func (g *Game) Update() error {

	g.cam.Update(g.world.w, g.world.h)

	if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		if g.world.gameState == simulationState {
			g.world.gameState = pauseState
		} else if g.world.gameState == pauseState {
			g.world.gameState = simulationState
		}
	}

	if g.world.Done() {
		return nil
	}

	if g.world.gameState == pauseState {
		if inpututil.IsKeyJustPressed(ebiten.KeyN) {
			g.world.Step()
		}
		return nil
	}
	g.world.Step()

	return nil
}

func newGame(cfg Config) *Game {
	g := Game{}
	g.world = New(cfg)
	g.cam = NewCamera(g.world.w, g.world.h)
	return &g
}

func (g *Game) stats() string {

	c := g.world.Counts()

	return fmt.Sprintf(
		"Game Stats: Size: {%d, %d}  Agents: %d  Starting-Infected: %d  Infection Rate: %f  Infection Duration: %d \nEnd-Agents: %d End-Infected: %d",
		g.world.w, g.world.h, g.world.agentsCount, g.world.initialInfected, g.world.infectionChance, g.world.infectionDuration, g.world.agentsCount, c.Infected)

}
