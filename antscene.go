package main

import (
	"fmt"
	"reflect"
	"unsafe"

	"github.com/veandco/go-sdl2/sdl"
)

type gridspot struct {
	foodPher int
	homePher int
	food     int
	home     bool
	wall     bool
}

const antTexSize = 3
const foodCount = 1
const pheromoneMax = 1000
const marker = 2000
const fadeDivisor = 500 // bigger number, slower pheromone fade.
const pheromoneExtend = 0

type AntScene struct {
	ants        []Ant
	grid        [WIDTH][HEIGHT]gridspot
	textures    []*sdl.Texture
	sceneTex    *sdl.Texture
	renderPher  bool
	renderGreen bool
	renderRed   bool
	propPher    bool
	oldPropPher bool
	pause       bool
	mousePX     int32
	mousePY     int32
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
		case "F":
			as.oldPropPher = !as.oldPropPher
		case "G":
			as.renderGreen = !as.renderGreen
		case "R":
			as.renderRed = !as.renderRed
		case "Space":
			as.pause = !as.pause
		}
	} else if e.GetType() == sdl.MOUSEMOTION {
		m := e.(*sdl.MouseMotionEvent)
		if m.X >= 0 && m.Y >= 0 && m.X < WIDTH && m.Y < HEIGHT {
			if m.State&sdl.ButtonLMask() > 0 {
				fmt.Printf("Left Mouse Drag (%d, %d)\n", m.X, m.Y)
				for i := m.X - 5; i < m.X+5; i++ {
					if i < 0 || i >= WIDTH {
						fmt.Printf("Continue i\n")
						continue
					}
					for j := m.Y - 5; j < m.Y+5; j++ {
						if j < 0 || j >= HEIGHT {
							fmt.Printf("Continue j\n")
							continue
						}
						fmt.Printf("WALL AT %d, %d\n", i, j)
						as.grid[i][j].wall = true
						as.grid[i][j].home = false
						as.grid[i][j].food = 0
					}
				}
			} else if m.State&sdl.ButtonRMask() > 0 {
				fmt.Printf("Right Mouse Drag (%d, %d)\n", m.X, m.Y)
				for i := m.X - 5; i < m.X+5; i++ {
					if i < 0 || i >= WIDTH {
						fmt.Printf("Continue i\n")
						continue
					}
					for j := m.Y - 5; j < m.Y+5; j++ {
						if j < 0 || j >= HEIGHT {
							fmt.Printf("Continue j\n")
							continue
						}
						fmt.Printf("ERASE AT %d, %d\n", i, j)
						as.grid[i][j].wall = false
						as.grid[i][j].home = false
						as.grid[i][j].food = 0
					}
				}
			} else if m.State&sdl.ButtonMMask() > 0 {
				fmt.Printf("Middle Mouse Drag (%d, %d)\n", m.X, m.Y)
				for i := m.X - 5; i < m.X+5; i++ {
					if i < 0 || i >= WIDTH {
						fmt.Printf("Continue i\n")
						continue
					}
					for j := m.Y - 5; j < m.Y+5; j++ {
						if j < 0 || j >= HEIGHT {
							fmt.Printf("Continue j\n")
							continue
						}
						fmt.Printf("FOOD AT %d, %d\n", i, j)
						as.grid[i][j].wall = false
						as.grid[i][j].home = false
						as.grid[i][j].food = foodCount
					}
				}
			}
			as.mousePX = m.X
			as.mousePY = m.Y
		}
	}
	return nil
}

func (as *AntScene) Init(g *Game[GameState], r *sdl.Renderer, s *GameState) error {
	as.renderPher = true
	as.renderGreen = true
	as.renderRed = true
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
			as.grid[x][y].food = foodCount
		}
	}

	for x := 800; x < 820; x++ {
		for y := 500; y < 520; y++ {
			as.grid[x][y].food = foodCount
		}
	}

	for x := 500; x < 520; x++ {
		for y := 800; y < 820; y++ {
			as.grid[x][y].food = foodCount
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
	if as.pause {
		return nil
	}
	for a := range as.ants {
		as.ants[a].Move(as)
		if as.grid[as.ants[a].pos.x][as.ants[a].pos.y].home {
			as.ants[a].food = 0
			as.ants[a].marker = marker
		}
		if as.grid[as.ants[a].pos.x][as.ants[a].pos.y].food > 0 {
			if as.ants[a].food == 0 {
				as.ants[a].dir = as.ants[a].dir.Right(4)
				as.grid[as.ants[a].pos.x][as.ants[a].pos.y].food -= 10
				as.ants[a].food = 10
			}
			as.ants[a].marker = marker
		}

		if as.ants[a].food > 0 {
			if as.grid[as.ants[a].pos.x][as.ants[a].pos.y].foodPher < pheromoneMax {
				as.grid[as.ants[a].pos.x][as.ants[a].pos.y].foodPher += as.ants[a].marker
				if as.ants[a].marker > 0 {
					as.ants[a].marker -= 1
				}
			} else if as.ants[a].marker > pheromoneExtend {
				as.ants[a].marker -= 1
			}
		} else {
			if as.grid[as.ants[a].pos.x][as.ants[a].pos.y].homePher < pheromoneMax {
				as.grid[as.ants[a].pos.x][as.ants[a].pos.y].homePher += as.ants[a].marker
				if as.ants[a].marker > 0 {
					as.ants[a].marker -= 1
				}
			} else if as.ants[a].marker > pheromoneExtend {
				as.ants[a].marker -= 1
			}
		}
	}
	for x := range as.grid {
		for y := range as.grid[x] {
			if as.grid[x][y].foodPher > 0 {
				as.grid[x][y].foodPher -= (as.grid[x][y].foodPher / fadeDivisor) + 1
			}
			if as.grid[x][y].homePher > 0 {
				as.grid[x][y].homePher -= (as.grid[x][y].homePher / fadeDivisor) + 1
			}
			if as.propPher {
				hasFood := as.grid[x][y].foodPher > 5000
				hasHome := as.grid[x][y].homePher > 5000
				if hasFood || hasHome {
					pt := point{x, y}
					if pt.Within(1, 1, WIDTH-2, HEIGHT-2) {
						if hasFood {
							as.grid[x][y].foodPher /= 2
							for d := N; d < END; d++ {
								pt2 := pt.PointAt(d)
								if as.grid[pt2.x][pt2.y].foodPher < pheromoneMax {
									as.grid[pt2.x][pt2.y].foodPher += (as.grid[x][y].foodPher / 9)
								}
							}
						}
						if hasHome {
							as.grid[x][y].homePher /= 2
							for d := N; d < END; d++ {
								pt2 := pt.PointAt(d)
								if as.grid[pt2.x][pt2.y].homePher < pheromoneMax {
									as.grid[pt2.x][pt2.y].homePher += (as.grid[x][y].homePher / 9)
								}
							}
						}
					}
				}
			} else if as.oldPropPher {
				hasFood := as.grid[x][y].foodPher > 100
				hasHome := as.grid[x][y].homePher > 100
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
	maxFoodPher := 0
	maxHomePher := 0
	for y := range as.grid[0] {
		for x := range as.grid {
			if as.grid[x][y].homePher > maxHomePher {
				maxHomePher = as.grid[x][y].homePher
			}
			if as.grid[x][y].foodPher > maxFoodPher {
				maxFoodPher = as.grid[x][y].foodPher
			}
			if as.grid[x][y].wall {
				bs[x+y*WIDTH] = 0x333333FF
				continue
			} else if as.grid[x][y].food > 0 {
				bs[x+y*WIDTH] = 0x33FF33FF
				continue
			} else if as.grid[x][y].home {
				bs[x+y*WIDTH] = 0xFF3333FF
				continue
			}
			if as.renderPher {
				var (
					vg uint32
					vr uint32
				)
				if as.renderGreen {
					vg = uint32(as.grid[x][y].foodPher / 2) //uint32((float32(as.grid[x][y].foodPher) / 500.0) * 255.0)
					if vg > 255 {
						vg = 255
					}
					vg = vg << 16
				}
				if as.renderRed {
					vr = uint32(as.grid[x][y].homePher / 2) //uint32((float32(as.grid[x][y].homePher) / 500.0) * 255.0)
					if vr > 255 {
						vr = 255
					}
					vr = vr << 24
				}
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
	fmt.Printf("Max Food: %d, Max Home: %d\n", maxFoodPher, maxHomePher)
	return nil
}

func (as *AntScene) RenderBelow() bool {
	return true
}

func (as *AntScene) Destroy() {
}
