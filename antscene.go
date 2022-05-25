package main

import (
	"github.com/veandco/go-sdl2/sdl"
)

const antTexSize = 8

type AntScene struct {
	ants     []Ant
	grid     [WIDTH][HEIGHT]point
	textures []*sdl.Texture
}

var _ Scene[GameState] = &AntScene{}

func (as *AntScene) Init(g *Game[GameState], r *sdl.Renderer, s *GameState) error {
	as.textures = make([]*sdl.Texture, int(END))
	for i := N; i < END; i++ {
		t, err := r.CreateTexture(uint32(sdl.PIXELFORMAT_RGBA8888), sdl.TEXTUREACCESS_STATIC, antTexSize, antTexSize)
		if err != nil {
			return err
		}
		as.textures[i] = t
		bs := make([]uint32, antTexSize*antTexSize+1)
		for j := 0; j < antTexSize*antTexSize+1; j++ {
			bs[j] = 0xc35b31ff
		}
		as.textures[i].UpdateRGBA(nil, bs, antTexSize)
	}
	return nil
}

func (as *AntScene) Update(g *Game[GameState], r *sdl.Renderer, s *GameState) error {
	for a := range as.ants {
		if as.ants[a].food > 0 {
			as.grid[as.ants[a].pos.x][as.ants[a].pos.y].x += as.ants[a].marker
		} else {
			as.grid[as.ants[a].pos.x][as.ants[a].pos.y].y += as.ants[a].marker
		}

		as.ants[a].Move(as)
		nx, ny := as.ants[a].pos.x, as.ants[a].pos.y
		if nx > 500 && nx < 600 && ny > 500 && ny < 600 {
			as.ants[a].food = 100
			as.ants[a].marker = 5000
		} else if nx < 100 && ny < 100 {
			as.ants[a].food = 0
			as.ants[a].marker = 5000
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
			hasx := as.grid[x][y].x > 700
			hasy := as.grid[x][y].y > 700
			if hasx || hasy {
				pt := point{x, y}
				if pt.Within(1, 1, WIDTH-2, HEIGHT-2) {
					if hasx {
						for d := N; d < END; d++ {
							pt2 := pt.PointAt(d)
							as.grid[pt2.x][pt2.y].x += (as.grid[x][y].x / 9)
						}
						as.grid[x][y].x /= 9
					}
					if hasy {
						for d := N; d < END; d++ {
							pt2 := pt.PointAt(d)
							as.grid[pt2.x][pt2.y].y += (as.grid[x][y].y / 9)
						}
						as.grid[x][y].y /= 9
					}
				}
			}

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
		if as.ants[a].food > 0 {
			bs[as.ants[a].pos.x+as.ants[a].pos.y*WIDTH] |= 0xFFFFFFFF
		} else {
			bs[as.ants[a].pos.x+as.ants[a].pos.y*WIDTH] |= 0x8888FFFF
		}
	}
	t.UpdateRGBA(nil, bs, WIDTH)
	r.Copy(t, nil, nil)
	for a := range as.ants {
		// 		r.SetDrawColor(0xFF, 0x00, 0xFF, 0xFF)
		// 		for i := as.ants[a].dir.Left(1); i <= as.ants[a].dir.Right(1); i++ {
		// 			rct := as.ants[a].OctantRect(i, 30)
		// 			//fmt.Printf("RECT: %#v\n", rct)
		// 			r.FillRect(&rct)
		// 		}
		dst := sdl.Rect{int32(as.ants[a].pos.x - (antTexSize / 2)), int32(as.ants[a].pos.y - (antTexSize / 2)), antTexSize, antTexSize}
		r.Copy(as.textures[as.ants[a].dir], nil, &dst)

	}
	return nil
}

func (as *AntScene) RenderBelow() bool {
	return true
}

func (as *AntScene) Destroy() {
}
