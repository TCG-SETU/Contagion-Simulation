package main

import "fmt"

// runHeadless steps the world until it finishes or maxSteps is reached.
func runHeadless(w *World, maxSteps int) {
	for !w.Done() && w.iteration < maxSteps {
		w.Step()
	}

	c := w.Counts()
	fmt.Printf("Finished after %d steps\n", w.iteration)
	fmt.Printf("Size: {%d, %d}  Agents: %d  Starting-Infected: %d  Infection Rate: %f  Infection Duration: %d\n",
		w.w, w.h, w.agentsCount, w.initialInfected, w.infectionChance, w.infectionDuration)
	fmt.Printf("Healthy: %d  Infected: %d  Immune: %d  Dead: %d\n",
		c.Healthy, c.Infected, c.Immune, c.Dead)
}
