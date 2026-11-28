package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
)

var rng *rand.Rand

// main centralises configuration and startup so command-line options can select either graphical or headless execution.
func main() {

	// Defines the command line flags.
	headless := flag.Bool("headless", false, "run without a window")
	maxSteps := flag.Int("max-steps", 10000, "step limit in headless mode")
	seed := flag.Int64("seed", 42, "RNG seed")
	hideDead := flag.Bool("hide-dead", false, "hide dead window")

	flag.Parse()

	cfg := Config{
		Width:             1000,
		Height:            1000,
		Agents:            1000,
		InitialInfected:   20,
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

	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Contagion")

	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}

	fmt.Println(game.stats())
}
