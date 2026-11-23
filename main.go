package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
)

var rng *rand.Rand

func clamp(a, min, max int) int {
	if a < min {
		a = min
	}
	if a > max {
		a = max
	}
	return a
}

func clampBorders(x, y, w, h int) (int, int) {
	x = clamp(x, 0, w-1)
	y = clamp(y, 0, h-1)
	return x, y
}

// clamps x and y to safe locations on the map
func clampToMap(m *Map, oldX, oldY, newX, newY int) (int, int) {
	if !m.IsSafe(oldX, oldY) {
		panic(fmt.Sprintf("old x and y must be safe, %d, %d", oldX, oldY))
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

func locationToDirection(a, b, positive, negative int) int {
	val := 0
	if a > b {
		val = positive
	} else if a < b {
		val = negative
	}
	return val
}

func main() {
	headless := flag.Bool("headless", false, "run without a window")
	maxSteps := flag.Int("max-steps", 10000, "step limit in headless mode")
	seed := flag.Int64("seed", 42, "RNG seed")
	hideDead := flag.Bool("hide-dead", false, "hide dead window")
	flag.Parse()

	cfg := Config{
		Width:             1000,
		Height:            1000,
		Agents:            10000,
		InitialInfected:   300,
		InfectionChance:   0.3,
		DeathChance:       0.00001,
		InfectionDuration: 200,
		Seed:              *seed,
		HideDead:          *hideDead,
	}
	rng = rand.New(rand.NewSource(cfg.Seed))

	if *headless {
		runHeadless(New(cfg), *maxSteps)
		return
	}

	game := newGame(cfg)

	ebiten.SetWindowSize(screenWidth*2, screenHeight*2)
	ebiten.SetWindowTitle("Contagion")

	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}

	fmt.Println(game.stats())
}
