package main

import (
	"encoding/gob"
	"fmt"
	"math/rand"
	"os"
	"reflect"
	"runtime"
	"sync"
	"time"
	"unsafe"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

const antTexSize = 3
const foodCount = 20
const pheromoneMax = 6000
const marker = 6000
const fadeDivisor = 800 // bigger number, slower pheromone fade.
const pheromoneExtend = 0

type gridspot struct {
	FoodPher int
	HomePher int
	Food     int
	Home     bool
	Wall     bool
}

type EGame struct {
	ants               []Ant
	grid               [WIDTH][HEIGHT]gridspot
	renderPher         bool
	renderGreen        bool
	renderRed          bool
	propPher           bool
	oldPropPher        bool
	pherOverload       bool
	pherOverloadFactor int
	parallel           bool
	pause              bool
	mousePX            int
	mousePY            int
	gridWorkChan       chan gridwork
	antWorkChan        chan gridwork
	renderWorkChan     chan renderwork
	newAnts            <-chan time.Time
	food               int64
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

func (as *EGame) SaveGrid() error {
	f, err := os.Create("ants.grid")
	if err != nil {
		return err
	}
	defer f.Close()

	enc := gob.NewEncoder(f)
	return enc.Encode(as.grid)
}

func (as *EGame) LoadGrid() error {
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

func (as *EGame) Init() error {
	as.renderPher = false
	as.renderGreen = true
	as.renderRed = true
	as.propPher = false
	as.pause = true
	as.parallel = true
	as.pherOverload = false
	as.pherOverloadFactor = 1500
	as.food = int64(len(as.ants)) * 80
	ebiten.SetMaxTPS(100)
	for x := 0; x < 100; x++ {
		for y := 0; y < 100; y++ {
			as.grid[x][y].Home = true
		}
	}

	for a := range as.ants {
		as.ants[a].dir = as.ants[a].dir.Left(rand.Intn(9))
	}

	// Experimental

	as.newAnts = time.Tick(time.Second)

	as.gridWorkChan = make(chan gridwork, 10)
	for i := 0; i < runtime.GOMAXPROCS(0); i++ {
		go func() {
			for {
				gw, ok := <-as.gridWorkChan
				if !ok {
					return
				}
				as.updateGrid(gw.start, gw.end)
				gw.wg.Done()
			}
		}()
	}

	as.antWorkChan = make(chan gridwork, 10)
	for i := 0; i < runtime.GOMAXPROCS(0); i++ {
		go func() {
			for {
				gw, ok := <-as.antWorkChan
				if !ok {
					return
				}
				as.updateAnts(gw.start, gw.end)
				gw.wg.Done()
			}
		}()
	}

	as.renderWorkChan = make(chan renderwork, 10)
	for i := 0; i < runtime.GOMAXPROCS(0); i++ {
		go func() {
			for {
				rw, ok := <-as.renderWorkChan
				if !ok {
					return
				}
				//fmt.Printf("GW.start: %d, gw.end: %d\n", gw.start, gw.end)
				as.renderPartial(rw.start, rw.end, rw.bs)
				rw.wg.Done()
			}
		}()
	}

	return nil
}

func (as *EGame) updateGrid(start, end int) {
	for x := start; x < end; x++ {
		for y := range as.grid[x] {
			if as.grid[x][y].FoodPher > 0 {
				as.grid[x][y].FoodPher -= (as.grid[x][y].FoodPher / fadeDivisor) + 1
			}
			if as.grid[x][y].HomePher > 0 {
				as.grid[x][y].HomePher -= (as.grid[x][y].HomePher / fadeDivisor) + 1
			}
		}
	}
}

func (as *EGame) updateAnts(start, end int) {
	for a := start; a < end; a++ {
		//fmt.Printf("Moving ant %d\n")
		as.ants[a].Move(as)
		if as.grid[as.ants[a].pos.x][as.ants[a].pos.y].Home {
			as.food += int64(as.ants[a].food)
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
				if as.grid[as.ants[a].pos.x][as.ants[a].pos.y].FoodPher > pheromoneMax {
					as.grid[as.ants[a].pos.x][as.ants[a].pos.y].FoodPher = pheromoneMax
				}

				if as.ants[a].marker > 0 {
					as.ants[a].marker -= 2
				}
			} else if as.ants[a].marker > pheromoneExtend {
				as.ants[a].marker -= 2
			}
		} else {
			if as.grid[as.ants[a].pos.x][as.ants[a].pos.y].HomePher < pheromoneMax {
				as.grid[as.ants[a].pos.x][as.ants[a].pos.y].HomePher += as.ants[a].marker
				if as.grid[as.ants[a].pos.x][as.ants[a].pos.y].HomePher > pheromoneMax {
					as.grid[as.ants[a].pos.x][as.ants[a].pos.y].HomePher = pheromoneMax
				}
				if as.ants[a].marker > 0 {
					as.ants[a].marker -= 2
				}
			} else if as.ants[a].marker > pheromoneExtend {
				as.ants[a].marker -= 2
			}
		}
	}
}

func (as *EGame) renderPartial(start, end int, bs []uint32) {
	divisor := pheromoneMax / 255
	for y := start; y < end; y++ {
		for x := range as.grid {
			if as.grid[x][y].Wall {
				bs[x+y*WIDTH] = 0xFF333333 //0x333333FF
				continue
			} else if as.grid[x][y].Food > 0 {
				bs[x+y*WIDTH] = 0xFF33FF33 //0x33FF33FF
				continue
			} else if as.grid[x][y].Home {
				bs[x+y*WIDTH] = 0xFF3333FF //0xFF3333FF
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
					vg = vg << 8
				}
				if as.renderRed {
					vr = uint32(as.grid[x][y].HomePher / divisor) //uint32((float32(as.grid[x][y].homePher) / 500.0) * 255.0)
					if vr > 255 {
						vr = 255
					}
				}
				bs[x+y*WIDTH] = 0xFF000000 | vg | vr
			} else {
				bs[x+y*WIDTH] = 0xFF000000
			}
		}
	}
}

func (as *EGame) HandleInput() error {
	if inpututil.IsKeyJustPressed(ebiten.KeyP) {
		as.renderPher = !as.renderPher
		fmt.Printf("Pheromone Rendering: %t\n", as.renderPher)
	} else if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		as.pause = !as.pause
		fmt.Printf("Pause: %t\n", as.pause)
	} else if inpututil.IsKeyJustPressed(ebiten.KeyG) {
		as.renderGreen = !as.renderGreen
	} else if inpututil.IsKeyJustPressed(ebiten.KeyR) {
		as.renderRed = !as.renderRed
	} else if inpututil.IsKeyJustPressed(ebiten.KeyS) {
		err := as.SaveGrid()
		if err != nil {
			fmt.Printf("Failed to save grid: %v\n", err)
		}
	} else if inpututil.IsKeyJustPressed(ebiten.KeyL) {
		err := as.LoadGrid()
		if err != nil {
			fmt.Printf("Failed to Load grid: %v\n", err)
		}
	} else if inpututil.IsKeyJustPressed(ebiten.KeyA) {
		as.ants = make([]Ant, len(as.ants))
		for a := range as.ants {
			as.ants[a].dir = as.ants[a].dir.Left(rand.Intn(9))
		}
	} else if inpututil.IsKeyJustPressed(ebiten.KeyC) {
		as.grid = [WIDTH][HEIGHT]gridspot{}
		for x := 0; x < 100; x++ {
			for y := 0; y < 100; y++ {
				as.grid[x][y].Home = true
			}
		}
		as.ants = make([]Ant, len(as.ants))
		for a := range as.ants {
			as.ants[a].dir = as.ants[a].dir.Left(rand.Intn(9))
		}
	} else if inpututil.IsKeyJustPressed(ebiten.KeyO) {
		as.pherOverload = !as.pherOverload
		fmt.Printf("Pheromone Overloading Enabled: %t, Factor: %v\n", as.pherOverload, as.pherOverloadFactor)
	} else if inpututil.IsKeyJustPressed(ebiten.KeyX) {
		as.parallel = !as.parallel
		fmt.Printf("Parallel update: %t\n", as.parallel)
	} else if inpututil.IsKeyJustPressed(ebiten.KeyDown) {
		as.pherOverloadFactor -= 100
		fmt.Printf("Pheromone Overloading Enabled: %t, Factor: %v\n", as.pherOverload, as.pherOverloadFactor)
	} else if inpututil.IsKeyJustPressed(ebiten.KeyUp) {
		as.pherOverloadFactor += 100
		fmt.Printf("Pheromone Overloading Enabled: %t, Factor: %v\n", as.pherOverload, as.pherOverloadFactor)
	}

	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		mx, my := ebiten.CursorPosition()
		if mx != as.mousePX || my != as.mousePY {
			for i := mx - 5; i < mx+5; i++ {
				if i < 0 || i >= WIDTH {
					continue
				}
				for j := my - 5; j < my+5; j++ {
					if j < 0 || j >= HEIGHT {
						continue
					}
					as.grid[i][j].Wall = true
					as.grid[i][j].Home = false
					as.grid[i][j].Food = 0
				}
			}
		}
	} else if ebiten.IsMouseButtonPressed(ebiten.MouseButtonRight) {
		mx, my := ebiten.CursorPosition()
		if mx != as.mousePX || my != as.mousePY {
			for i := mx - 5; i < mx+5; i++ {
				if i < 0 || i >= WIDTH {
					continue
				}
				for j := my - 5; j < my+5; j++ {
					if j < 0 || j >= HEIGHT {
						continue
					}
					as.grid[i][j].Wall = false
					as.grid[i][j].Home = false
					as.grid[i][j].Food = 0
				}
			}
		}
	} else if ebiten.IsMouseButtonPressed(ebiten.MouseButtonMiddle) {
		mx, my := ebiten.CursorPosition()
		if mx != as.mousePX || my != as.mousePY {
			for i := mx - 5; i < mx+5; i++ {
				if i < 0 || i >= WIDTH {
					continue
				}
				for j := my - 5; j < my+5; j++ {
					if j < 0 || j >= HEIGHT {
						continue
					}
					as.grid[i][j].Wall = false
					as.grid[i][j].Home = false
					as.grid[i][j].Food = foodCount
				}
			}
		}
	}
	mx, my := ebiten.CursorPosition()
	as.mousePX = mx
	as.mousePY = my
	return nil
}

// Update proceeds the game state.
// Update is called every tick (1/60 [s] by default).
func (as *EGame) Update() error {
	if err := as.HandleInput(); err != nil {
		return err
	}
	if as.pause {
		return nil
	}
	// 	select {
	// 	case <-as.newAnts:
	// 		l := int64(len(as.ants))
	// 		if as.food > 0 {
	// 			as.food -= l / 10
	// 		}
	// 		min := l * 40
	// 		ratio := as.food / min
	// 		if ratio >= 1 {
	// 			bs := make([]Ant, ratio)
	// 			as.ants = append(as.ants, bs...)
	// 		} else if ratio < 1 {
	// 			keep := (len(as.ants) * int(as.food)) / int(min)
	// 			delete := len(as.ants) - keep
	// 			as.ants = as.ants[:len(as.ants)-delete]
	// 		}
	// 		fmt.Printf("Food: %d, Ants: %d, Ratio: %d, keep: (%d)\n", as.food, len(as.ants), ratio, (len(as.ants)*int(as.food))/int(min))
	//
	// 	default:
	// 	}
	if as.parallel {
		return as.ParallelUpdate()
	}
	as.updateAnts(0, len(as.ants))
	as.updateGrid(0, WIDTH)
	return nil
}

func (as *EGame) ParallelUpdate() error {
	var wg sync.WaitGroup
	a := 0
	for a < len(as.ants) {
		end := a + len(as.ants)/16
		if end > len(as.ants) {
			end = len(as.ants)
		}
		wg.Add(1)
		as.antWorkChan <- gridwork{start: a, end: end, wg: &wg}
		a = end

	}
	wg.Wait()
	//fmt.Printf("DONE\n")

	x := 0
	for x < len(as.grid) {
		end := x + len(as.grid)/16
		if end > len(as.grid) {
			end = len(as.grid)
		}
		wg.Add(1)
		as.gridWorkChan <- gridwork{start: x, end: end, wg: &wg}
		x = end

	}
	wg.Wait()
	return nil
}

// Draw draws the game screen.
// Draw is called every frame (typically 1/60[s] for 60Hz display).
func (as *EGame) Draw(screen *ebiten.Image) {
	if as.parallel {
		as.ParallelRender(screen)
		return
	}
	bs := make([]uint32, WIDTH*HEIGHT)
	divisor := pheromoneMax / 255
	for y := range as.grid[0] {
		for x := range as.grid {
			if as.grid[x][y].Wall {
				bs[x+y*WIDTH] = 0xFF333333 //0x333333FF
				continue
			} else if as.grid[x][y].Food > 0 {
				bs[x+y*WIDTH] = 0xFF33FF33 //0x33FF33FF
				continue
			} else if as.grid[x][y].Home {
				bs[x+y*WIDTH] = 0xFF3333FF //0xFF3333FF
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
					vg = vg << 8
				}
				if as.renderRed {
					vr = uint32(as.grid[x][y].HomePher / divisor) //uint32((float32(as.grid[x][y].homePher) / 500.0) * 255.0)
					if vr > 255 {
						vr = 255
					}
				}
				bs[x+y*WIDTH] = 0xFF000000 | vg | vr
			} else {
				bs[x+y*WIDTH] = 0xFF000000
			}
		}
	}
	for a := range as.ants {
		for x := as.ants[a].pos.x - (antTexSize / 2); x < as.ants[a].pos.x+(antTexSize/2); x++ {
			if x < 0 || x >= WIDTH {
				continue
			}
			for y := as.ants[a].pos.y - (antTexSize / 2); y < as.ants[a].pos.y+(antTexSize/2); y++ {
				if y < 0 || y >= HEIGHT {
					continue
				}
				bs[x+y*WIDTH] = 0xFFFFFFFF
			}
		}
	}

	var bbs []byte
	sliceHeader := (*reflect.SliceHeader)(unsafe.Pointer(&bbs))
	sliceHeader.Cap = int(len(bs) * 4)
	sliceHeader.Len = int(len(bs) * 4)
	sliceHeader.Data = uintptr(unsafe.Pointer(&bs[0]))
	screen.ReplacePixels(bbs)

	return
}

func (as *EGame) ParallelRender(screen *ebiten.Image) {
	bs := make([]uint32, WIDTH*HEIGHT)
	var wg sync.WaitGroup
	a := 0
	for a < len(as.grid[0]) {
		end := a + len(as.grid[0])/16
		if end > len(as.grid[0]) {
			end = len(as.grid[0])
		}
		//fmt.Printf("%d -> %d\n", a, end)
		wg.Add(1)
		as.renderWorkChan <- renderwork{start: a, end: end, wg: &wg, bs: bs}
		a = end

	}
	wg.Wait()
	for a := range as.ants {
		for x := as.ants[a].pos.x - (antTexSize / 2); x < as.ants[a].pos.x+(antTexSize/2); x++ {
			if x < 0 || x >= WIDTH {
				continue
			}
			for y := as.ants[a].pos.y - (antTexSize / 2); y < as.ants[a].pos.y+(antTexSize/2); y++ {
				if y < 0 || y >= HEIGHT {
					continue
				}
				bs[x+y*WIDTH] = 0xFFFFFFFF
			}
		}
	}
	var bbs []byte
	sliceHeader := (*reflect.SliceHeader)(unsafe.Pointer(&bbs))
	sliceHeader.Cap = int(len(bs) * 4)
	sliceHeader.Len = int(len(bs) * 4)
	sliceHeader.Data = uintptr(unsafe.Pointer(&bs[0]))
	screen.ReplacePixels(bbs)
	return
}

// Layout takes the outside size (e.g., the window size) and returns the (logical) screen size.
// If you don't have to adjust the screen size with the outside size, just return a fixed size.
func (as *EGame) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return WIDTH, HEIGHT
}
