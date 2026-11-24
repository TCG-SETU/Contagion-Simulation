package main

// Counts stores the number of agents in each health state within the simulation.
type Counts struct {
	Healthy, Infectious, Infected, Dead, Immune int
}

// Total calculates and returns the total number of agents across all of the health states
func (c *Counts) Total() int {
	return c.Healthy + c.Infected + c.Infectious + c.Dead + c.Immune
}

// Update adds the health state changes from delta to the current counts.
// It updates each healthy category independently.
func (c *Counts) Update(delta Counts) {
	c.Healthy += delta.Healthy
	c.Infected += delta.Infected
	c.Infectious += delta.Infectious
	c.Dead += delta.Dead
	c.Immune += delta.Immune
}
