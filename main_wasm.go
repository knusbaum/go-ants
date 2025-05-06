package main

import (
	"log"

	"github.com/hajimehoshi/ebiten/v2"
)

const (
	WIDTH  = 800
	HEIGHT = 450
	nants  = 1000
)

func main() {

	var err error
	ebiten.SetWindowSize(WIDTH, HEIGHT)
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	ebiten.SetWindowTitle("Ants")

	g := NewGame[GameState](WIDTH, HEIGHT, NewGameState(WIDTH, HEIGHT))
	as := &AntScene{homelife: 10 * 3000 * 10000}
	err = g.PushScene(as)
	if err != nil {
		log.Fatal(err)
	}

	// Call ebiten.RunGame to start your game loop.
	if err := ebiten.RunGame(g); err != nil {
		log.Fatal(err)
	}

}
