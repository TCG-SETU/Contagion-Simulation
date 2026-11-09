package main

import (
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

	const (
		w                 = 500
		h                 = 400
		agentCount        = 10000
		infectedCount     = 300
		infectionChance   = 0.8
		infectionDuration = 50
		deathChance       = 0.001
		randomSeed        = 42
	)
	source := rand.NewSource(randomSeed)
	rng = rand.New(source)

	game := newGame(w, h, agentCount, infectedCount, infectionChance, infectionDuration, deathChance)

	ebiten.SetWindowSize(screenWidth*2, screenHeight*2)
	ebiten.SetWindowTitle("Contagion")

	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}

	fmt.Println(game.stats())
}
