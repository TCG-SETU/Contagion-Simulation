package main

import (
	"fmt"
	"os"
	"os/exec"
	"time"
)

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

	h := 50
	w := 50
	agentCount := 300
	infectedCount := 50
	infectionChance := 0.3
	infectionDuration := 3

	game := newGame(h, w, agentCount, infectedCount, infectionChance, infectionDuration)

	for i := range 50 {
		cmd := exec.Command("clear") //Linux example, its tested
		cmd.Stdout = os.Stdout
		err := cmd.Run()
		if err != nil {
			// ignore
		}
		game.step(i)
		fmt.Print(game.String()) // Clear screen and print field.
		time.Sleep(100 * time.Millisecond)
	}

	fmt.Println(game.stats())
}
