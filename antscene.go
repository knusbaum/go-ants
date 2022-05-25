package main

import (
	"reflect"
	"unsafe"

	"github.com/veandco/go-sdl2/sdl"
)

type gridspot struct {
	foodPher int
	homePher int
	food     int
	home     bool
}

const antTexSize = 3

type AntScene struct {
	ants       []Ant
	grid       [WIDTH][HEIGHT]gridspot
	textures   []*sdl.Texture
	sceneTex   *sdl.Texture
	renderPher bool
	propPher   bool
}

var _ Scene[GameState] = &AntScene{}

func (as *AntScene) HandleEvent(g *Game[GameState], r *sdl.Renderer, e sdl.Event) error {
	if e.GetType() == sdl.KEYDOWN {
		k := e.(*sdl.KeyboardEvent)
		keyname := sdl.GetKeyName(sdl.GetKeyFromScancode(k.Keysym.Scancode))
		switch keyname {
		case "P":
			as.renderPher = !as.renderPher
			t, err := r.CreateTexture(uint32(sdl.PIXELFORMAT_RGBA8888), sdl.TEXTUREACCESS_STREAMING, WIDTH, HEIGHT)
			if err != nil {
				return err
			}
			as.sceneTex.Destroy()
			as.sceneTex = t
		case "Y":
			as.propPher = !as.propPher
		}
	}
	return nil
}

func (as *AntScene) Init(g *Game[GameState], r *sdl.Renderer, s *GameState) error {
	as.renderPher = true
	as.propPher = false
	t, err := r.CreateTexture(uint32(sdl.PIXELFORMAT_RGBA8888), sdl.TEXTUREACCESS_STREAMING, WIDTH, HEIGHT)
	if err != nil {
		return err
	}
	as.sceneTex = t

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

	for x := 500; x < 520; x++ {
		for y := 500; y < 520; y++ {
			as.grid[x][y].food = 10
		}
	}

	for x := 800; x < 820; x++ {
		for y := 500; y < 520; y++ {
			as.grid[x][y].food = 10
		}
	}

	for x := 500; x < 520; x++ {
		for y := 800; y < 820; y++ {
			as.grid[x][y].food = 10
		}
	}

	for x := 0; x < 100; x++ {
		for y := 0; y < 100; y++ {
			as.grid[x][y].home = true
		}
	}

	return nil
}

func (as *AntScene) Update(g *Game[GameState], r *sdl.Renderer, s *GameState) error {
	for a := range as.ants {
		as.ants[a].Move(as)
		if as.grid[as.ants[a].pos.x][as.ants[a].pos.y].home {
			as.ants[a].food = 0
			as.ants[a].marker = 5000
		}
		if as.grid[as.ants[a].pos.x][as.ants[a].pos.y].food > 0 {
			if as.ants[a].food == 0 {
				as.ants[a].dir = as.ants[a].dir.Right(4)
				as.grid[as.ants[a].pos.x][as.ants[a].pos.y].food -= 10
				as.ants[a].food = 10
			}
			as.ants[a].marker = 5000
		}

		if as.ants[a].food > 0 {
			if as.grid[as.ants[a].pos.x][as.ants[a].pos.y].foodPher < 5000 {
				as.grid[as.ants[a].pos.x][as.ants[a].pos.y].foodPher += as.ants[a].marker
			}
		} else {
			if as.grid[as.ants[a].pos.x][as.ants[a].pos.y].homePher < 5000 {
				as.grid[as.ants[a].pos.x][as.ants[a].pos.y].homePher += as.ants[a].marker
			}
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
	bsb, _, err := as.sceneTex.Lock(nil)
	if err != nil {
		return err
	}
	var bs []uint32
	sliceHeader := (*reflect.SliceHeader)(unsafe.Pointer(&bs))
	sliceHeader.Cap = int(len(bsb) / 4)
	sliceHeader.Len = int(len(bsb) / 4)
	sliceHeader.Data = uintptr(unsafe.Pointer(&bsb[0]))
	for y := range as.grid[0] {
		for x := range as.grid {
			if as.propPher {
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
			}

			if as.grid[x][y].food > 0 {
				bs[x+y*WIDTH] = 0x33FF33FF
				continue
			} else if as.grid[x][y].home {
				bs[x+y*WIDTH] = 0xFF3333FF
				continue
			}
			if as.renderPher {
				vg := uint32(as.grid[x][y].foodPher / 2) //uint32((float32(as.grid[x][y].foodPher) / 500.0) * 255.0)
				if vg > 255 {
					vg = 255
				}
				vg = vg << 16
				vr := uint32(as.grid[x][y].homePher / 2) //uint32((float32(as.grid[x][y].homePher) / 500.0) * 255.0)
				if vr > 255 {
					vr = 255
				}
				vr = vr << 24
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
			} else {
				bs[x+y*WIDTH] = 0
			}
		}
	}
	// 	for a := range as.ants {
	// 		if as.ants[a].food > 0 {
	// 			bs[as.ants[a].pos.x+as.ants[a].pos.y*WIDTH] |= 0xFFFFFFFF
	// 		} else {
	// 			bs[as.ants[a].pos.x+as.ants[a].pos.y*WIDTH] |= 0x8888FFFF
	// 		}
	// 	}
	as.sceneTex.Unlock()
	r.Copy(as.sceneTex, nil, nil)

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
