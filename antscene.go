package main

import (
	"math/rand"

	"github.com/veandco/go-sdl2/sdl"
)

type Ant struct {
	x      int
	y      int
	dirx   int
	diry   int
	food   int
	marker int
}

type AntScene struct {
	ants []Ant
	grid [WIDTH][HEIGHT]point
}

var _ Scene[GameState] = &AntScene{}

type point struct {
	x int
	y int
}

func (a *Ant) checkg(an *AntScene, x, y int, val int) bool {
	if x > 0 && x < WIDTH && y > 0 && y < HEIGHT {
		return an.grid[x][y].x > val
	}
	return false
}

func (a *Ant) checkl(an *AntScene, x, y int, val int) bool {
	if x > 0 && x < WIDTH && y > 0 && y < HEIGHT {
		return an.grid[x][y].y > val
	}
	return false
}

func (a *Ant) selectDirection(as *AntScene) (int, int) {

	val := int(0)
	nx := a.x
	ny := a.y

	for i := a.x - 1; i < a.x+2; i++ {
		for j := a.y - 1; j < a.y+2; j++ {
			if a.food > 0 {
				if a.checkl(as, i, j, val) {
					nx = i
					ny = j
					val = as.grid[i][j].y
				}
			} else {
				if a.checkg(as, i, j, val) {
					nx = i
					ny = j
					val = as.grid[i][j].x - as.grid[i][j].y
				}
			}
		}
	}
	if nx == a.x && ny == a.y {
		nx := a.x + a.dirx
		ny := a.y + a.diry
		r := rand.Intn(50)
		for r == 0 || nx < 0 || nx >= WIDTH || ny < 0 || ny >= HEIGHT {
			a.dirx = rand.Intn(3) - 1
			a.diry = rand.Intn(3) - 1
			nx = a.x + a.dirx
			ny = a.y + a.diry
			r = rand.Intn(100)
		}

		// 		bs := make([]point, 0, 8)
		// 		for i := a.x - 1; i < a.x+2; i++ {
		// 			if i == a.x {
		// 				continue
		// 			}
		// 			for j := a.y - 1; j < a.y+2; j++ {
		// 				if j == a.y {
		// 					continue
		// 				}
		// 				if i > 0 && i < WIDTH && j > 0 && j < HEIGHT {
		// 					bs = append(bs, point{i, j})
		// 				}
		// 			}
		// 		}
		// 		pt := bs[rand.Intn(len(bs))]
		// 		return pt.x, pt.y
		return nx, ny
	}
	return nx, ny
}

func (as *AntScene) Update(g *Game[GameState], r *sdl.Renderer, s *GameState) error {
	for a := range as.ants {
		if as.ants[a].food > 0 {
			as.grid[as.ants[a].x][as.ants[a].y].x += as.ants[a].marker
		} else {
			as.grid[as.ants[a].x][as.ants[a].y].y += as.ants[a].marker
		}
		nx, ny := as.ants[a].selectDirection(as)
		as.ants[a].x = nx
		as.ants[a].y = ny
		if nx > 500 && nx < 600 && ny > 500 && ny < 600 {
			as.ants[a].food = 100
			as.ants[a].marker = 5000
		} else if nx < 100 && ny < 100 {
			as.ants[a].food = 0
			as.ants[a].marker = 5000
		} else if as.ants[a].marker > 0 {
			as.ants[a].marker -= 1
		}
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
			// 			vg := uint32(as.grid[x][y].x / 2)
			// 			vg = (vg & 0xFF) << 16
			// 			vr := uint32(as.grid[x][y].y / 2)
			// 			vr = (vr & 0xFF) << 24
			var (
				vg uint32
				vr uint32
			)
			if as.grid[x][y].x > 0 {
				vg = 0xFF << 16
			}
			if as.grid[x][y].y > 0 {
				vr = 0xFF << 24
			}

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
