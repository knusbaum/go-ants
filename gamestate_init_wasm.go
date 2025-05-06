package main

func NewGameState(width, height int) GameState {
	g := GameState{}
	g.width = width
	g.height = height
	g.parallel = true
	g.renderPher = false
	g.renderGreen = true
	g.renderRed = true
	g.renderAnts = true
	g.antlife = 40000
	g.followWalls = true
	g.foodcount = 200
	g.foodlife = 2000
	g.stockpile = 100
	g.maxants = 5000
	g.drawradius = 20
	g.fadedivisor = 3000
	g.sight = 15
	g.adaptiveNavigation = true
	return g
}
