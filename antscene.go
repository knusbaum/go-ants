package main

import (
	"github.com/veandco/go-sdl2/sdl"
)

type gridspot struct {
	foodPher int
	homePher int
	food     int
	home     bool
}

const antTexSize = 8

type AntScene struct {
	ants     []Ant
	grid     [WIDTH][HEIGHT]gridspot
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
			as.grid[as.ants[a].pos.x][as.ants[a].pos.y].foodPher += as.ants[a].marker
		} else {
			as.grid[as.ants[a].pos.x][as.ants[a].pos.y].homePher += as.ants[a].marker
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
			if as.grid[x][y].foodPher > 0 {
				as.grid[x][y].foodPher -= 1
			} else if as.grid[x][y].homePher > 0 {
				as.grid[x][y].homePher -= 1
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
			hasFood := as.grid[x][y].foodPher > 1000
			hasHome := as.grid[x][y].homePher > 1000
			if hasFood || hasHome {
				pt := point{x, y}
				if pt.Within(1, 1, WIDTH-2, HEIGHT-2) {
					if hasFood {
						for d := N; d < END; d++ {
							pt2 := pt.PointAt(d)
							as.grid[pt2.x][pt2.y].foodPher += (as.grid[x][y].foodPher / 9)
						}
						as.grid[x][y].foodPher /= 9
					}
					if hasHome {
						for d := N; d < END; d++ {
							pt2 := pt.PointAt(d)
							as.grid[pt2.x][pt2.y].homePher += (as.grid[x][y].homePher / 9)
						}
						as.grid[x][y].homePher /= 9
					}
				}
			}

			vg := uint32((float32(as.grid[x][y].foodPher) / 500.0) * 255.0)
			if vg >= 255 {
				vg = 254
			}
			vg = (vg & 0xFF) << 16
			vr := uint32((float32(as.grid[x][y].homePher) / 500.0) * 255.0)
			if vr >= 255 {
				vr = 254
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
