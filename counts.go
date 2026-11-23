package main

type Counts struct {
	Healthy, Infectious, Infected, Dead, Immune int
}

func (c *Counts) Total() int {
	return c.Healthy + c.Infected + c.Infectious + c.Dead + c.Immune
}

func (c *Counts) Update(delta Counts) {
	c.Healthy += delta.Healthy
	c.Infected += delta.Infected
	c.Infectious += delta.Infectious
	c.Dead += delta.Dead
	c.Immune += delta.Immune
}
