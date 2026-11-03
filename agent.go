package main

type Agent struct {
	x, y           int
	infected       bool
	infectionTimer int
}

func (a *Agent) handleInfectionTimer(g *Game) {
	a.infectionTimer--
	if a.infectionTimer <= 0 {
		a.infected = false
		g.infectedCountNow--
	}
}
