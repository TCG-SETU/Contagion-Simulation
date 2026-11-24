package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
)

// rng is the random number generator shared across the simulation
var rng *rand.Rand

// main configures and starts the simulation.
// It reads command line flags, creates the simulation configuration.
// initialises the random number generator and starts either headless.
// or graphical simulation mode
func main() {

	// Defines the command line flags.
	headless := flag.Bool("headless", false, "run without a window")
	maxSteps := flag.Int("max-steps", 10000, "step limit in headless mode")
	seed := flag.Int64("seed", 42, "RNG seed")
	hideDead := flag.Bool("hide-dead", false, "hide dead window")

	// parses the command line arguments.
	flag.Parse()

	// Configures the simulation settings.
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
	// Initialises the random number generator using the configuration seed.
	rng = rand.New(rand.NewSource(cfg.Seed))

	// Run the simulation without any graphical rendering if headless mode is enabled.
	if *headless {
		runHeadless(New(cfg), *maxSteps)
		return
	}

	// Creates a new graphical simulation
	game := newGame(cfg)

	// Configures the game window
	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Contagion")

	// Starts the graphical simulation and handles errors.
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}

	// prints the simulation settings when the graphical simulation exits.
	fmt.Println(game.stats())
}
