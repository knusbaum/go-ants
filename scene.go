package main

import (
	"github.com/veandco/go-sdl2/sdl"
)

type Scene[T any] interface {
	Update(*Game[T], *sdl.Renderer, *T) error
	Render(*Game[T], *sdl.Renderer, *T) error
	RenderBelow() bool
	Destroy()
}

type EventHandler[T any] interface {
	HandleEvent(*Game[T], *sdl.Renderer, sdl.Event) error
}

type Initer[T any] interface {
	Init(*Game[T], *sdl.Renderer, *T) error
}
