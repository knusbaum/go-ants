package main

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/veandco/go-sdl2/sdl"
)

type Game[T any] struct {
	window *sdl.Window
	r      *sdl.Renderer
	state  T
	scenes []Scene[T]
	render []Scene[T]
}

func NewGame[T any](w, h int32) (*Game[T], error) {
	rand.Seed(time.Now().UnixNano())
	g := Game[T]{}
	window, err := sdl.CreateWindow("test", 0, 0,
		w, h, sdl.WINDOW_SHOWN)
	if err != nil {
		return nil, err
	}
	g.window = window
	r, err := sdl.CreateRenderer(g.window, -1, sdl.RENDERER_ACCELERATED)
	if err != nil {
		return nil, err
	}
	g.r = r

	g.scenes = make([]Scene[T], 0)
	g.setRenderStack()
	return &g, nil
}

func (g *Game[T]) Destroy() {
	if g.r != nil {
		g.r.Destroy()
	}
	g.window.Destroy()
}

func (g *Game[T]) Run() error {
	for {
		// 		start := sdl.GetTicks64()
		h, ok := g.scenes[len(g.scenes)-1].(EventHandler[T])

		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			if event.GetType() == sdl.QUIT {
				println("Quit")
				return nil
			}
			if ok {
				err := h.HandleEvent(g, g.r, event)
				if err != nil {
					return err
				}
			}
		}
		err := g.scenes[len(g.scenes)-1].Update(g, g.r, &g.state)
		if err != nil {
			fmt.Printf("ERROR UPDATING: %v\n", err)
			return err
		}
		g.r.SetDrawColor(0x00, 0x00, 0x00, 0xFF)
		g.r.Clear()
		for _, scene := range g.render {
			err := scene.Render(g, g.r, &g.state)
			if err != nil {
				return err
			}
		}
		g.r.Present()
		// 		end := sdl.GetTicks64()
		// 		fmt.Printf("%d FPS\n", 1000/(end-start))
		// 		next := start + 20
		// 		fmt.Printf("next: %d, end: %d\n", next, end)
		// 		if next > end {
		// 			fmt.Printf("DELAY: %d\n", next-end)
		// 			//sdl.Delay(uint32(next - end))
		// 		}
		//sdl.Delay(1)
	}
	return nil
}

func (g *Game[T]) setRenderStack() {

	var i int
	for i = len(g.scenes) - 1; i >= 0; i-- {
		if !g.scenes[i].RenderBelow() {
			i--
			break
		}
	}
	fmt.Printf("g.render = g.scenes[%d:]\n", i+1)
	g.render = g.scenes[i+1:]
}

func (g *Game[T]) PushScene(s Scene[T]) {
	if i, ok := s.(Initer[T]); ok {
		i.Init(g, g.r, &g.state)
	}
	g.scenes = append(g.scenes, s)
	g.setRenderStack()
}

func (g *Game[T]) PopScene() {
	if len(g.scenes) == 0 {
		return
	}
	popped := g.scenes[len(g.scenes)-1]
	popped.Destroy()
	g.scenes = g.scenes[:len(g.scenes)-1]
	g.setRenderStack()
}
