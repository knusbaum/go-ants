package main

import (
	"fmt"
	"log"

	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
)

const (
	WIDTH  = 1000
	HEIGHT = 1000
)

type GameState struct {
	x int
}

type LineScene struct {
	linex1 int32
	liney1 int32
	linex2 int32
	liney2 int32
}

func (s *LineScene) HandleEvent(g *Game[GameState], r *sdl.Renderer, e sdl.Event) error {
	if e.GetType() == sdl.KEYDOWN {
		fmt.Printf("KEYDOWN EVENT: %v\n", e)
		k := e.(*sdl.KeyboardEvent)
		keyname := sdl.GetKeyName(sdl.GetKeyFromScancode(k.Keysym.Scancode))
		fmt.Printf("KEYNAME: %v\n", keyname)
		if keyname == "P" {
			fmt.Printf("Pushing Pause Scene!\n")
			g.PushScene(NewPauseScene(r))
			return nil
		}
	}
	return nil
}

func (s *LineScene) Update(g *Game[GameState], r *sdl.Renderer, st *GameState) error {
	s.linex1 = (s.linex1 + 1) % WIDTH
	s.linex2 = (s.linex2 + 1) % WIDTH
	s.liney1 = (s.liney1 + 1) % HEIGHT
	s.liney2 = (s.liney2 + 1) % HEIGHT
	return nil
}

func (s *LineScene) Render(g *Game[GameState], r *sdl.Renderer, st *GameState) error {
	//log.Printf("Rendering line scene.\n")
	r.SetDrawColor(0xFF, 0x00, 0x00, 0xFF)
	r.DrawLine(s.linex1, s.liney1, s.linex2, s.liney2)
	return nil
}

func (s *LineScene) RenderBelow() bool {
	return true
}

func (s *LineScene) Destroy() {}

func main() {
	if err := sdl.Init(sdl.INIT_EVERYTHING); err != nil {
		panic(err)
	}
	defer sdl.Quit()
	fmt.Printf("INITED\n")

	if err := ttf.Init(); err != nil {
		panic(err)
	}
	defer ttf.Quit()

	g, err := NewGame[GameState](WIDTH, HEIGHT)
	if err != nil {
		log.Fatalf("Failed to create new game: %v", err)
	}
	defer g.Destroy()
	//g.PushScene(&LineScene{0, 0, 100, 100})
	ants := make([]Ant, 1000)
	for i := 0; i < 1000; i++ {
		ants[i] = Ant{x: 0, y: 0, food: 0}
	}
	g.PushScene(&AntScene{ants: ants})
	err = g.Run()
	fmt.Printf("Finished: %v\n", err)

}
