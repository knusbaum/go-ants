package main

import (
	"github.com/veandco/go-sdl2/sdl"
)

type point struct {
	x int
	y int
}

type AntScene struct {
	ants []Ant
	grid [WIDTH][HEIGHT]point
}

var _ Scene[GameState] = &AntScene{}

func (as *AntScene) Update(g *Game[GameState], r *sdl.Renderer, s *GameState) error {
	for a := range as.ants {
		if as.ants[a].food > 0 {
			as.grid[as.ants[a].x][as.ants[a].y].x += as.ants[a].marker
		} else {
			as.grid[as.ants[a].x][as.ants[a].y].y += as.ants[a].marker
		}
		// 		nx, ny := as.ants[a].selectDirection(as)
		// 		as.ants[a].x = nx
		// 		as.ants[a].y = ny
		as.ants[a].Move(as)
		nx, ny := as.ants[a].x, as.ants[a].y
		if nx > 500 && nx < 600 && ny > 500 && ny < 600 {
			as.ants[a].food = 100
			as.ants[a].marker = 500
		} else if nx < 100 && ny < 100 {
			as.ants[a].food = 0
			as.ants[a].marker = 500
		}
		// else if as.ants[a].marker > 0 {
		//	as.ants[a].marker -= 1
		//	}
	}
	for x := range as.grid {
		for y := range as.grid[x] {
			if as.grid[x][y].x > 0 {
				as.grid[x][y].x -= 1
			} else if as.grid[x][y].y > 0 {
				as.grid[x][y].y -= 1
			}
		}
	}
	return nil
}

func (as *AntScene) Render(g *Game[GameState], r *sdl.Renderer, s *GameState) error {
	t, err := r.CreateTexture(uint32(sdl.PIXELFORMAT_RGBA8888), sdl.TEXTUREACCESS_STREAMING, WIDTH, HEIGHT)
	if err != nil {
		return err
	}
	defer t.Destroy()

	bs := make([]uint32, WIDTH*HEIGHT+1)
	for y := range as.grid[0] {
		for x := range as.grid {
			vg := uint32(as.grid[x][y].x)
			if vg > 255 {
				vg = 255
			}
			vg = (vg & 0xFF) << 16
			vr := uint32(as.grid[x][y].y)
			if vr > 255 {
				vr = 255
			}
			vr = (vr & 0xFF) << 24
			// 			var (
			// 				vg uint32
			// 				vr uint32
			// 			)
			// 			if as.grid[x][y].x > 0 {
			// 				vg = 0xFF << 16
			// 			}
			// 			if as.grid[x][y].y > 0 {
			// 				vr = 0xFF << 24
			// 			}

			bs[x+y*WIDTH] = 0x000000FF | vg | vr
		}
	}
	for a := range as.ants {
		bs[as.ants[a].x+as.ants[a].y*WIDTH] |= 0xFFFFFFFF
	}
	t.UpdateRGBA(nil, bs, WIDTH)
	r.Copy(t, nil, nil)
	return nil
}

func (as *AntScene) RenderBelow() bool {
	return true
}

func (as *AntScene) Destroy() {
}
