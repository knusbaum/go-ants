package main

func NewGameState(width, height int) GameState {
	g := GameState{}
	g.width = width
	g.height = height
	g.parallel = false
	g.renderPher = false
	g.renderGreen = true
	g.renderRed = true
	g.antlife = 10000
	g.foodcount = 20
	g.foodlife = 2000
	g.spawnparam = 1
	g.maxants = 1000
	g.drawradius = 20
	g.fadedivisor = 500
	return g
}
