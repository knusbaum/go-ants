//go:build !wasm

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
	g.antlife = 50000
	//g.foodcount = 20
	//g.foodcount = 20
	g.foodcount = 200
	g.foodlife = 2000
	g.stockpile = 50
	//g.maxants = 4000
	g.maxants = 40000
	//g.maxants = 100000
	g.drawradius = 20
	g.fadedivisor = 1500
	g.sight = 15
	//g.sorttype = none
	return g
}
