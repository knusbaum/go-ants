package main

import (
	"log"

	"github.com/hajimehoshi/ebiten/v2"
)

const (
	WIDTH  = 1500
	HEIGHT = 1000
)

func main() {

	game := &EGame{}
	game.ants = make([]Ant, 1000)
	game.Init()
	// Specify the window size as you like. Here, a doubled size is specified.
	ebiten.SetWindowSize(WIDTH, HEIGHT)
	ebiten.SetWindowTitle("Your game's title")
	// Call ebiten.RunGame to start your game loop.
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}

}
