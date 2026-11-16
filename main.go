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
		Width:             500,
		Height:            400,
		Agents:            20000,
		InitialInfected:   300,
		InfectionChance:   0.9,
		DeathChance:       0.01,
		InfectionDuration: 500,
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
