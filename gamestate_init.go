//go:build !wasm

package main

func NewGameState(width, height int) GameState {
	g := GameState{}
	g.width = width
	g.height = height
	//g.parallel = true
	g.parallel = false
	g.renderPher = false
	g.renderGreen = true
	g.renderRed = true
	g.renderAnts = true
	g.antlife = 10000
	//g.foodcount = 20
	//g.foodcount = 20
	g.foodcount = 200
	g.foodlife = 2000
	g.spawnparam = 1
	//g.maxants = 4000
	//g.maxants = 40000
	g.maxants = 400000
	g.drawradius = 20
	g.fadedivisor = 700
	g.sight = 10
	g.sorttype = none
	return g
}
