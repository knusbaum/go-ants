package main

import (
	"encoding/gob"
	"fmt"
	"os"
	"reflect"
	"sync"
	"unsafe"

	"github.com/veandco/go-sdl2/sdl"
)

type gridspot struct {
	FoodPher int
	HomePher int
	Food     int
	Home     bool
	Wall     bool
}

const antTexSize = 3
const foodCount = 20
const pheromoneMax = 6000
const marker = 3000
const fadeDivisor = 600 // bigger number, slower pheromone fade.
const pheromoneExtend = 0

type AntScene struct {
	ants               []Ant
	grid               [WIDTH][HEIGHT]gridspot
	textures           []*sdl.Texture
	sceneTex           *sdl.Texture
	renderPher         bool
	renderGreen        bool
	renderRed          bool
	propPher           bool
	oldPropPher        bool
	pherOverload       bool
	pherOverloadFactor int
	parallel           bool
	pause              bool
	mousePX            int32
	mousePY            int32
	gridWorkChan       chan gridwork
	antWorkChan        chan gridwork
	renderWorkChan     chan renderwork
}

type gridwork struct {
	wg    *sync.WaitGroup
	start int
	end   int
}

type renderwork struct {
	wg    *sync.WaitGroup
	start int
	end   int
	bs    []uint32
}

var _ Scene[GameState] = &AntScene{}

func (as *AntScene) SaveGrid() error {
	f, err := os.Create("ants.grid")
	if err != nil {
		return err
	}
	defer f.Close()

	enc := gob.NewEncoder(f)
	return enc.Encode(as.grid)
}

func (as *AntScene) LoadGrid() error {
	f, err := os.Open("ants.grid")
	if err != nil {
		return err
	}
	defer f.Close()

	enc := gob.NewDecoder(f)
	g := [WIDTH][HEIGHT]gridspot{}
	err = enc.Decode(&g)
	if err != nil {
		return err
	}
	as.grid = g
	return nil
}

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
			fmt.Printf("New Pheromone Propagation: %t\n", as.propPher)
		case "F":
			as.oldPropPher = !as.oldPropPher
			fmt.Printf("Old Pheromone Propagation: %t\n", as.oldPropPher)
		case "G":
			as.renderGreen = !as.renderGreen
		case "R":
			as.renderRed = !as.renderRed
		case "S":
			err := as.SaveGrid()
			if err != nil {
				fmt.Printf("Failed to save grid: %v\n", err)
			}
		case "L":
			err := as.LoadGrid()
			if err != nil {
				fmt.Printf("Failed to Load grid: %v\n", err)
			}
		case "A":
			as.ants = make([]Ant, len(as.ants))
		case "C":
			as.grid = [WIDTH][HEIGHT]gridspot{}
			for x := 0; x < 100; x++ {
				for y := 0; y < 100; y++ {
					as.grid[x][y].Home = true
				}
			}
			as.ants = make([]Ant, len(as.ants))
		case "O":
			as.pherOverload = !as.pherOverload
			fmt.Printf("Pheromone Overloading Enabled: %t, Factor: %v\n", as.pherOverload, as.pherOverloadFactor)
		case "X":
			as.parallel = !as.parallel
			fmt.Printf("Parallel update: %t\n", as.parallel)
		case "Space":
			as.pause = !as.pause
		case "Up":
			as.pherOverloadFactor += 100
			fmt.Printf("Pheromone Overloading Enabled: %t, Factor: %v\n", as.pherOverload, as.pherOverloadFactor)
		case "Down":
			as.pherOverloadFactor -= 100
			fmt.Printf("Pheromone Overloading Enabled: %t, Factor: %v\n", as.pherOverload, as.pherOverloadFactor)
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
						as.grid[i][j].Wall = true
						as.grid[i][j].Home = false
						as.grid[i][j].Food = 0
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
						as.grid[i][j].Wall = false
						as.grid[i][j].Home = false
						as.grid[i][j].Food = 0
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
						as.grid[i][j].Wall = false
						as.grid[i][j].Home = false
						as.grid[i][j].Food = foodCount
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
	as.renderPher = false
	as.renderGreen = true
	as.renderRed = true
	as.propPher = false
	as.pause = true
	as.parallel = true
	as.pherOverloadFactor = 500
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

	for x := 0; x < 100; x++ {
		for y := 0; y < 100; y++ {
			as.grid[x][y].Home = true
		}
	}

	// Experimental
	as.gridWorkChan = make(chan gridwork, 10)
	for i := 0; i < 16; i++ {
		go func() {
			for {
				gw, ok := <-as.gridWorkChan
				if !ok {
					return
				}
				for x := gw.start; x < gw.end; x++ {
					for y := range as.grid[x] {
						if as.grid[x][y].FoodPher > 0 {
							as.grid[x][y].FoodPher -= (as.grid[x][y].FoodPher / fadeDivisor) + 1
						}
						if as.grid[x][y].HomePher > 0 {
							as.grid[x][y].HomePher -= (as.grid[x][y].HomePher / fadeDivisor) + 1
						}
					}
				}
				gw.wg.Done()
			}
		}()
	}

	as.antWorkChan = make(chan gridwork, 10)
	for i := 0; i < 16; i++ {
		go func() {
			for {
				gw, ok := <-as.antWorkChan
				if !ok {
					return
				}
				//fmt.Printf("GW.start: %d, gw.end: %d\n", gw.start, gw.end)
				for a := gw.start; a < gw.end; a++ {
					//fmt.Printf("Moving ant %d\n")
					as.ants[a].Move(as)
					if as.grid[as.ants[a].pos.x][as.ants[a].pos.y].Home {
						as.ants[a].food = 0
						as.ants[a].marker = marker
					}
					if as.grid[as.ants[a].pos.x][as.ants[a].pos.y].Food > 0 {
						if as.ants[a].food == 0 {
							as.ants[a].dir = as.ants[a].dir.Right(4)
							as.grid[as.ants[a].pos.x][as.ants[a].pos.y].Food -= 10
							as.ants[a].food = 10
						}
						as.ants[a].marker = marker
					}

					if as.ants[a].food > 0 {
						if as.grid[as.ants[a].pos.x][as.ants[a].pos.y].FoodPher < pheromoneMax {
							as.grid[as.ants[a].pos.x][as.ants[a].pos.y].FoodPher += as.ants[a].marker
							if as.ants[a].marker > 0 {
								as.ants[a].marker -= 1
							}
						} else if as.ants[a].marker > pheromoneExtend {
							as.ants[a].marker -= 1
						}
					} else {
						if as.grid[as.ants[a].pos.x][as.ants[a].pos.y].HomePher < pheromoneMax {
							as.grid[as.ants[a].pos.x][as.ants[a].pos.y].HomePher += as.ants[a].marker
							if as.ants[a].marker > 0 {
								as.ants[a].marker -= 1
							}
						} else if as.ants[a].marker > pheromoneExtend {
							as.ants[a].marker -= 1
						}
					}
				}
				gw.wg.Done()
			}
		}()
	}

	as.renderWorkChan = make(chan renderwork, 10)
	for i := 0; i < 16; i++ {
		go func() {
			for {
				rw, ok := <-as.renderWorkChan
				if !ok {
					return
				}
				//fmt.Printf("GW.start: %d, gw.end: %d\n", gw.start, gw.end)
				divisor := pheromoneMax / 255
				for y := rw.start; y < rw.end; y++ {
					for x := range as.grid {
						if as.grid[x][y].Wall {
							rw.bs[x+y*WIDTH] = 0x333333FF
							continue
						} else if as.grid[x][y].Food > 0 {
							rw.bs[x+y*WIDTH] = 0x33FF33FF
							continue
						} else if as.grid[x][y].Home {
							rw.bs[x+y*WIDTH] = 0xFF3333FF
							continue
						}
						if as.renderPher {
							var (
								vg uint32
								vr uint32
							)
							if as.renderGreen {
								vg = uint32(as.grid[x][y].FoodPher / divisor) //uint32((float32(as.grid[x][y].foodPher) / 500.0) * 255.0)
								if vg > 255 {
									vg = 255
								}
								vg = vg << 16
							}
							if as.renderRed {
								vr = uint32(as.grid[x][y].HomePher / divisor) //uint32((float32(as.grid[x][y].homePher) / 500.0) * 255.0)
								if vr > 255 {
									vr = 255
								}
								vr = vr << 24
							}
							rw.bs[x+y*WIDTH] = 0x000000FF | vg | vr
						} else {
							rw.bs[x+y*WIDTH] = 0
						}
					}
				}
				rw.wg.Done()
			}
		}()

	}

	return nil
}

func (as *AntScene) Update(g *Game[GameState], r *sdl.Renderer, s *GameState) error {
	if as.pause {
		return nil
	}
	if as.parallel {
		return as.ParallelUpdate(g, r, s)
	}
	for a := range as.ants {
		as.ants[a].Move(as)
		if as.grid[as.ants[a].pos.x][as.ants[a].pos.y].Home {
			as.ants[a].food = 0
			as.ants[a].marker = marker
		}
		if as.grid[as.ants[a].pos.x][as.ants[a].pos.y].Food > 0 {
			if as.ants[a].food == 0 {
				as.ants[a].dir = as.ants[a].dir.Right(4)
				as.grid[as.ants[a].pos.x][as.ants[a].pos.y].Food -= 10
				as.ants[a].food = 10
			}
			as.ants[a].marker = marker
		}

		if as.ants[a].food > 0 {
			if as.grid[as.ants[a].pos.x][as.ants[a].pos.y].FoodPher < pheromoneMax {
				as.grid[as.ants[a].pos.x][as.ants[a].pos.y].FoodPher += as.ants[a].marker
				if as.ants[a].marker > 0 {
					as.ants[a].marker -= 1
				}
			} else if as.ants[a].marker > pheromoneExtend {
				as.ants[a].marker -= 1
			}
		} else {
			if as.grid[as.ants[a].pos.x][as.ants[a].pos.y].HomePher < pheromoneMax {
				as.grid[as.ants[a].pos.x][as.ants[a].pos.y].HomePher += as.ants[a].marker
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
			if as.grid[x][y].FoodPher > 0 {
				as.grid[x][y].FoodPher -= (as.grid[x][y].FoodPher / fadeDivisor) + 1
			}
			if as.grid[x][y].HomePher > 0 {
				as.grid[x][y].HomePher -= (as.grid[x][y].HomePher / fadeDivisor) + 1
			}
			if as.propPher {
				hasFood := as.grid[x][y].FoodPher > marker
				hasHome := as.grid[x][y].HomePher > marker
				if hasFood || hasHome {
					pt := point{x, y}
					if pt.Within(1, 1, WIDTH-2, HEIGHT-2) {
						if hasFood {
							as.grid[x][y].FoodPher /= 2
							for d := N; d < END; d++ {
								pt2 := pt.PointAt(d)
								if as.grid[pt2.x][pt2.y].FoodPher < pheromoneMax {
									as.grid[pt2.x][pt2.y].FoodPher += (as.grid[x][y].FoodPher / 9)
								}
							}
						}
						if hasHome {
							as.grid[x][y].HomePher /= 2
							for d := N; d < END; d++ {
								pt2 := pt.PointAt(d)
								if as.grid[pt2.x][pt2.y].HomePher < pheromoneMax {
									as.grid[pt2.x][pt2.y].HomePher += (as.grid[x][y].HomePher / 9)
								}
							}
						}
					}
				}
			} else if as.oldPropPher {
				hasFood := as.grid[x][y].FoodPher > 100
				hasHome := as.grid[x][y].HomePher > 100
				if hasFood || hasHome {
					pt := point{x, y}
					if pt.Within(1, 1, WIDTH-2, HEIGHT-2) {
						if hasFood {
							for d := N; d < END; d++ {
								pt2 := pt.PointAt(d)
								as.grid[pt2.x][pt2.y].FoodPher += (as.grid[x][y].FoodPher / 9)
							}
							as.grid[x][y].FoodPher /= 9
						}
						if hasHome {
							for d := N; d < END; d++ {
								pt2 := pt.PointAt(d)
								as.grid[pt2.x][pt2.y].HomePher += (as.grid[x][y].HomePher / 9)
							}
							as.grid[x][y].HomePher /= 9
						}
					}
				}
			}
		}
	}
	return nil
}

func (as *AntScene) ParallelUpdate(g *Game[GameState], r *sdl.Renderer, s *GameState) error {
	var wg sync.WaitGroup
	a := 0
	for a < len(as.ants) {
		end := a + len(as.ants)/16
		if end > len(as.ants) {
			end = len(as.ants)
		}
		//fmt.Printf("%d -> %d\n", a, end)
		as.antWorkChan <- gridwork{start: a, end: end, wg: &wg}
		a = end
		wg.Add(1)
	}
	wg.Wait()
	//fmt.Printf("DONE\n")

	x := 0
	for x < len(as.grid) {
		end := x + len(as.grid)/16
		if end > len(as.grid) {
			end = len(as.grid)
		}
		//fmt.Printf("%d -> %d\n", x, end)
		as.gridWorkChan <- gridwork{start: x, end: end, wg: &wg}
		x = end
		wg.Add(1)
	}
	wg.Wait()
	return nil
}

func (as *AntScene) Render(g *Game[GameState], r *sdl.Renderer, s *GameState) error {
	if as.parallel {
		return as.ParallelRender(g, r, s)
	}
	bsb, _, err := as.sceneTex.Lock(nil)
	if err != nil {
		return err
	}
	var bs []uint32
	sliceHeader := (*reflect.SliceHeader)(unsafe.Pointer(&bs))
	sliceHeader.Cap = int(len(bsb) / 4)
	sliceHeader.Len = int(len(bsb) / 4)
	sliceHeader.Data = uintptr(unsafe.Pointer(&bsb[0]))
	divisor := pheromoneMax / 255
	for y := range as.grid[0] {
		for x := range as.grid {
			if as.grid[x][y].Wall {
				bs[x+y*WIDTH] = 0x333333FF
				continue
			} else if as.grid[x][y].Food > 0 {
				bs[x+y*WIDTH] = 0x33FF33FF
				continue
			} else if as.grid[x][y].Home {
				bs[x+y*WIDTH] = 0xFF3333FF
				continue
			}
			if as.renderPher {
				var (
					vg uint32
					vr uint32
				)
				if as.renderGreen {
					vg = uint32(as.grid[x][y].FoodPher / divisor) //uint32((float32(as.grid[x][y].foodPher) / 500.0) * 255.0)
					if vg > 255 {
						vg = 255
					}
					vg = vg << 16
				}
				if as.renderRed {
					vr = uint32(as.grid[x][y].HomePher / divisor) //uint32((float32(as.grid[x][y].homePher) / 500.0) * 255.0)
					if vr > 255 {
						vr = 255
					}
					vr = vr << 24
				}
				bs[x+y*WIDTH] = 0x000000FF | vg | vr
			} else {
				bs[x+y*WIDTH] = 0
			}
		}
	}
	as.sceneTex.Unlock()
	r.Copy(as.sceneTex, nil, nil)

	for a := range as.ants {
		dst := sdl.Rect{int32(as.ants[a].pos.x - (antTexSize / 2)), int32(as.ants[a].pos.y - (antTexSize / 2)), antTexSize, antTexSize}
		r.Copy(as.textures[as.ants[a].dir], nil, &dst)
	}
	return nil
}

func (as *AntScene) ParallelRender(g *Game[GameState], r *sdl.Renderer, s *GameState) error {
	bsb, _, err := as.sceneTex.Lock(nil)
	if err != nil {
		return err
	}
	var bs []uint32
	sliceHeader := (*reflect.SliceHeader)(unsafe.Pointer(&bs))
	sliceHeader.Cap = int(len(bsb) / 4)
	sliceHeader.Len = int(len(bsb) / 4)
	sliceHeader.Data = uintptr(unsafe.Pointer(&bsb[0]))
	var wg sync.WaitGroup
	a := 0
	for a < len(as.grid[0]) {
		end := a + len(as.grid[0])/16
		if end > len(as.grid[0]) {
			end = len(as.grid[0])
		}
		//fmt.Printf("%d -> %d\n", a, end)
		as.renderWorkChan <- renderwork{start: a, end: end, wg: &wg, bs: bs}
		a = end
		wg.Add(1)
	}
	wg.Wait()
	as.sceneTex.Unlock()
	r.Copy(as.sceneTex, nil, nil)

	for a := range as.ants {
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
