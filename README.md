# Contagion-Simulation

This project is a contagion simulation written in Go. It simulates a population of agents moving around a set grid, where infected agents are able to spread their infection to nearby healthy agents. 

Its displays the grid inside of the terminal and updates over multiple steps


## Installation

* Install Go Land on your computer

Installs Go's Standard Library

```bash
go mod tidy
```
* Install Git to clone the repository

```bash
git clone <repository-url>
```



## Example of simulation running

```Go
Game 49 steps
100 infected
X X X X X X X X X X 
X X X X X X X X X X 
X X X X X X X X X X 
X X X X X X X X X X 
X X X X X X X X X X 
X X X X X X X X X X 
X X X X X X X X X X 
X X X X X X X X X X 
X X X X X X X X X X 
X X X X X X X X X X 
Game Stats: Size: {10, 10}  Agents: 100  Starting-Infected: 10  Infection Rate: 0.100000  Infection Duration: 7 
End-Agents: 100 End-Infected: 100
```

## Simulation Settings
The simulation can be configured by changing the values in the Main function

| Setting | Description |
| --- | --- |
| h | Height of the grid |
| w | Width of the grid |
| agent Count | Total number of agents |
| infected Count | Initial amount of infected agents |
| infection Chance | Chance of the infection spreading to healthy agent |
| infection Duration | Number of steps an agent remains infected |


## Contributing

Contributions are welcome. Please open an issue or a pull request.

## Licence

GPL-3.0
