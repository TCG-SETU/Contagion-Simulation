package main

type Counts struct {
	Healthy, Infectious, Infected, Dead, Immune int
}

// Total provides a single population total so callers do not need to repeat the health-state calculation.
func (c *Counts) Total() int {
	return c.Healthy + c.Infected + c.Infectious + c.Dead + c.Immune
}

// Update applies population changes together so counts stay consistent after an iteration.
func (c *Counts) Update(delta Counts) {
	c.Healthy += delta.Healthy
	c.Infected += delta.Infected
	c.Infectious += delta.Infectious
	c.Dead += delta.Dead
	c.Immune += delta.Immune
}
