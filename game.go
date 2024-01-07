package main

import (
	"sync"

	"github.com/hajimehoshi/ebiten/v2"
)

type Scene[T any] interface {
	Draw(g *Game[T], state *T, screen *ebiten.Image)
	Update(g *Game[T], state *T) error
	Init(g *Game[T], state *T) error
	DrawUnder(g *Game[T], state *T) bool
}

type Game[T any] struct {
	width, height int
	state         T
	sceneStack    []Scene[T]
	lock          sync.Mutex
}

func NewGame[T any](width, height int, init T) *Game[T] {
	return &Game[T]{
		width:  width,
		height: height,
		state:  init,
	}
}

func (g *Game[T]) Draw(screen *ebiten.Image) {
	for i := 0; i < len(g.sceneStack)-1; i++ {
		if g.sceneStack[i].DrawUnder(g, &g.state) {
			g.sceneStack[i].Draw(g, &g.state, screen)
		}
	}
	if len(g.sceneStack) > 0 {
		g.sceneStack[len(g.sceneStack)-1].Draw(g, &g.state, screen)
	}
}

func (g *Game[T]) Update() error {
	if len(g.sceneStack) > 0 {
		return g.sceneStack[len(g.sceneStack)-1].Update(g, &g.state)
	}
	return nil
}

// Layout takes the outside size (e.g., the window size) and returns the (logical) screen size.
// If you don't have to adjust the screen size with the outside size, just return a fixed size.
func (g *Game[T]) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return g.width, g.height
	//return WIDTH, HEIGHT
	//return outsideWidth, outsideHeight
}

func (g *Game[T]) PushScene(s Scene[T]) error {
	g.lock.Lock()
	defer g.lock.Unlock()
	err := s.Init(g, &g.state)
	if err != nil {
		return err
	}
	g.sceneStack = append(g.sceneStack, s)
	return nil
}

func (g *Game[T]) PopScene() {
	g.lock.Lock()
	defer g.lock.Unlock()
	g.sceneStack = g.sceneStack[:len(g.sceneStack)-1]
}
