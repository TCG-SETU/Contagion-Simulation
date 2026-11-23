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
	screenWidth  = 600
	screenHeight = 500
)

type Game struct {
	world      *World
	cleanupBus *cleanupBus
}

const gridOffset = 50

// cleanupBus is a small blue bus that drives across the grid. Its front edge
// is the cleanup line: every dead agent behind it is removed from the world.
type cleanupBus struct {
	x int
}

func (b *cleanupBus) update(w *World) bool {
	b.x += 8
	w.RemoveDeadThrough(b.x - gridOffset)
	return b.x > gridOffset+w.w+60
}

func (b *cleanupBus) draw(screen *ebiten.Image, w *World) {
	const (
		busWidth  = 56
		busHeight = 24
	)
	y := screenHeight/2 - busHeight/2

	// A subtle vertical sweep shows that the bus is collecting agents across
	// the whole simulation grid, rather than just those on the road.
	vector.DrawFilledRect(screen, float32(gridOffset), float32(gridOffset), float32(b.x-gridOffset), float32(w.h), color.RGBA{R: 45, G: 135, B: 220, A: 24}, false)
	vector.DrawFilledRect(screen, float32(b.x-busWidth), float32(y), busWidth, busHeight, color.RGBA{R: 31, G: 112, B: 201, A: 0xFF}, false)
	vector.DrawFilledRect(screen, float32(b.x-busWidth+8), float32(y+5), 12, 8, color.RGBA{R: 180, G: 225, B: 255, A: 0xFF}, false)
	vector.DrawFilledRect(screen, float32(b.x-busWidth+24), float32(y+5), 12, 8, color.RGBA{R: 180, G: 225, B: 255, A: 0xFF}, false)
	vector.DrawFilledRect(screen, float32(b.x-busWidth+41), float32(y+5), 8, 8, color.RGBA{R: 180, G: 225, B: 255, A: 0xFF}, false)
	vector.DrawFilledCircle(screen, float32(b.x-busWidth+12), float32(y+busHeight), 4, color.RGBA{R: 20, G: 25, B: 35, A: 0xFF}, false)
	vector.DrawFilledCircle(screen, float32(b.x-12), float32(y+busHeight), 4, color.RGBA{R: 20, G: 25, B: 35, A: 0xFF}, false)
}

func (g *Game) Draw(screen *ebiten.Image) {

	screen.Fill(color.RGBA{R: 46, G: 7, B: 29, A: 0xFF})

	var c = Counts{}

	for _, agent := range g.world.Agents() {
		if !(g.world.hideDead && agent.state == StateDead) {
			agent.Draw(screen)
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
	if g.cleanupBus != nil {
		g.cleanupBus.draw(screen, g.world)
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
		"TPS: %0.1f  FPS: %0.1f%s  [B: cleanup bus]\n"+
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

	if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		if g.world.gameState == simulationState {
			g.world.gameState = pauseState
		} else if g.world.gameState == pauseState {
			g.world.gameState = simulationState
		}
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyB) && g.cleanupBus == nil && g.world.Counts().Dead > 0 {
		g.cleanupBus = &cleanupBus{x: gridOffset - 1}
	}

	if g.cleanupBus != nil {
		if g.cleanupBus.update(g.world) {
			g.cleanupBus = nil
		}
		return nil
	}

	if g.world.Done() {
		if g.world.Counts().Dead > 0 {
			g.cleanupBus = &cleanupBus{x: gridOffset - 1}
		}
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
	return &g
}

func (g *Game) stats() string {

	c := g.world.Counts()

	return fmt.Sprintf(
		"Game Stats: Size: {%d, %d}  Agents: %d  Starting-Infected: %d  Infection Rate: %f  Infection Duration: %d \nEnd-Agents: %d End-Infected: %d",
		g.world.w, g.world.h, g.world.agentsCount, g.world.initialInfected, g.world.infectionChance, g.world.infectionDuration, g.world.agentsCount, c.Infected)

}
