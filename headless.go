package main

import "fmt"

// runHeadless runs the simulation without graphical rendering.
// It continues stepping the world until the simulation finishes.
// or maximum number of steps is reached.
// Once finished, it prints the simulation settings and final population counts.
func runHeadless(w *World, maxSteps int) {
	for !w.Done() && w.iteration < maxSteps {
		w.Step()
	}

	// retrieves the final population count.
	c := w.Counts()

	// Prints the total number of iteration steps.
	fmt.Printf("Finished after %d steps\n", w.iteration)

	// Prints the simulation config.
	fmt.Printf("Size: {%d, %d}  Agents: %d  Starting-Infected: %d  Infection Rate: %f  Infection Duration: %d\n",
		w.w, w.h, w.agentsCount, w.initialInfected, w.infectionChance, w.infectionDuration)

	// Prints the final population distribution.
	fmt.Printf("Healthy: %d  Infected: %d  Immune: %d  Dead: %d\n",
		c.Healthy, c.Infected, c.Immune, c.Dead)
}
