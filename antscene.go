package main

import (
	"encoding/gob"
	"fmt"
	"image/color"
	"log"
	"os"
	"runtime"
	"sync"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"

	"github.com/hajimehoshi/ebiten/v2/examples/resources/fonts"
)

var workers = runtime.GOMAXPROCS(0)
var mplusNormalFont font.Face

type gridspot struct {
	FoodPher int
	HomePher int
	Food     int
	Home     bool
	Wall     bool
}

const antTexSize = 3
const foodCount = 10
const pheromoneMax = 8191
const pherShift = 5 //(2^13 = 8192), meaning 8192 is within 13 bits range, We want to shift that to 8 bits, so shift 5 out.
const marker = 5000
const fadeDivisor = 700 // bigger number, slower pheromone fade.
const antFadeDivisor = 600

// const pheromoneExtend = 200
const antlife = 5000 // an ant spends 1 life per frame
const foodlife = 100 // amount of life 1 food gives an ant
const spawnParam = 10
const maxants = 5000

type AntScene struct {
	ants               []Ant
	field              *Field[gridspot]
	textures           []*ebiten.Image
	fullTextures       []*ebiten.Image
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
	followWalls        bool
	homefood           int64

	antwg            sync.WaitGroup
	antworkerTrigger []chan struct{}

	pherwg            sync.WaitGroup
	pherworkerTrigger []chan struct{}
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

//var _ Scene[GameState] = &AntScene{}

func (as *AntScene) SaveGrid() error {
	f, err := os.Create("ants.grid")
	if err != nil {
		return err
	}
	defer f.Close()

	enc := gob.NewEncoder(f)
	return enc.Encode(as.field.vals)
}

func (as *AntScene) LoadGrid() error {
	f, err := os.Open("ants.grid")
	if err != nil {
		return err
	}
	defer f.Close()

	enc := gob.NewDecoder(f)
	g := []gridspot{}
	err = enc.Decode(&g)
	if err != nil {
		return err
	}
	as.field.vals = g
	as.field.UpdateAll()
	return nil
}

// func (as *AntScene) HandleEvent(g *Game[GameState], r *sdl.Renderer, e sdl.Event) error {
func (as *AntScene) HandleInput() error {
	if inpututil.IsKeyJustPressed(ebiten.KeyP) {
		as.renderPher = !as.renderPher
		fmt.Printf("RENDER PHEROMONES: %t\n", as.renderPher)
		as.field.UpdateAll()
	} else if inpututil.IsKeyJustPressed(ebiten.KeyY) {
		as.propPher = !as.propPher
		fmt.Printf("New Pheromone Propagation: %t\n", as.propPher)
	} else if inpututil.IsKeyJustPressed(ebiten.KeyF) {
		as.oldPropPher = !as.oldPropPher
		fmt.Printf("Old Pheromone Propagation: %t\n", as.oldPropPher)
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
		//as.ants = make([]Ant, len(as.ants))
		as.relocateAnts()
	} else if inpututil.IsKeyJustPressed(ebiten.KeyC) {
		as.field.Clear()
		for x := 0; x < 100; x++ {
			for y := 0; y < 100; y++ {
				as.field.Get(x, y).Home = true
				as.field.Update(x, y)
			}
		}
		//as.ants = make([]Ant, len(as.ants))
		as.relocateAnts()
	} else if inpututil.IsKeyJustPressed(ebiten.KeyO) {
		as.pherOverload = !as.pherOverload
		fmt.Printf("Pheromone Overloading Enabled: %t, Factor: %v\n", as.pherOverload, as.pherOverloadFactor)
	} else if inpututil.IsKeyJustPressed(ebiten.KeyX) {
		as.parallel = !as.parallel
		fmt.Printf("Parallel update: %t\n", as.parallel)
	} else if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		as.pause = !as.pause
	} else if inpututil.IsKeyJustPressed(ebiten.KeyUp) {
		as.pherOverloadFactor += 100
		fmt.Printf("Pheromone Overloading Enabled: %t, Factor: %v\n", as.pherOverload, as.pherOverloadFactor)
	} else if inpututil.IsKeyJustPressed(ebiten.KeyDown) {
		as.pherOverloadFactor -= 100
		fmt.Printf("Pheromone Overloading Enabled: %t, Factor: %v\n", as.pherOverload, as.pherOverloadFactor)
	} else if inpututil.IsKeyJustPressed(ebiten.KeyW) {
		as.followWalls = !as.followWalls
		fmt.Printf("Wall Following: %t\n", as.followWalls)
	}

	doSpot := func(x, y int, f func(x, y int, g *gridspot)) {
		for i := x - 5; i < x+5; i++ {
			if i < 0 || i >= WIDTH {
				continue
			}
			for j := y - 5; j < y+5; j++ {
				if j < 0 || j >= HEIGHT {
					continue
				}
				spot := as.field.Get(int(i), int(j))
				f(int(i), int(j), spot)
			}
		}
	}

	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		mx, my := ebiten.CursorPosition()
		if mx != as.mousePX || my != as.mousePY {
			doLine(mx, my, as.mousePX, as.mousePY, func(cx, cy int) {
				doSpot(cx, cy, func(x, y int, spot *gridspot) {
					spot.Wall = true
					spot.Home = false
					spot.Food = 0
					as.field.Update(x, y)
				})
			})
		}
	} else if ebiten.IsMouseButtonPressed(ebiten.MouseButtonRight) {
		mx, my := ebiten.CursorPosition()
		if mx != as.mousePX || my != as.mousePY {
			doLine(mx, my, as.mousePX, as.mousePY, func(cx, cy int) {
				doSpot(cx, cy, func(x, y int, spot *gridspot) {
					spot.Wall = false
					spot.Home = false
					spot.Food = 0
					as.field.Update(x, y)
				})
			})
		}
	} else if ebiten.IsMouseButtonPressed(ebiten.MouseButtonMiddle) {
		mx, my := ebiten.CursorPosition()
		if mx != as.mousePX || my != as.mousePY {
			doLine(mx, my, as.mousePX, as.mousePY, func(cx, cy int) {
				doSpot(cx, cy, func(x, y int, spot *gridspot) {
					spot.Wall = false
					spot.Home = false
					spot.Food = foodCount
					as.field.Update(x, y)
				})
			})
		}
	}
	mx, my := ebiten.CursorPosition()
	as.mousePX = mx
	as.mousePY = my
	return nil
}

//var homePherMaxPresent = 1
//var foodPherMaxPresent = 1

func (as *AntScene) renderGridspot(g *gridspot) uint32 {
	//homedivisor := (homePherMaxPresent / 255) + 1
	//fooddivisor := (foodPherMaxPresent / 255) + 1
	//homedivisor := (homePherMaxPresent >> 8) + 1
	//fooddivisor := (foodPherMaxPresent >> 8) + 1
	if g.Wall {
		//return 0x333333FF
		return 0xFF333333
	} else if g.Food > 0 {
		//return 0x33FF33FF
		return 0xFF33FF33

	} else if g.Home {
		//return 0xFF3333FF
		return 0xFF3333FF
	}
	if as.renderPher {
		var (
			vg uint32
			vr uint32
		)
		if as.renderGreen {
			//vg = uint32(g.FoodPher / fooddivisor)
			if g.FoodPher > pheromoneMax {
				panic("TOO BIG")
			}
			vg = uint32(g.FoodPher) >> pherShift
			// if vg > 255 {
			// 	fmt.Printf("FOOD > 255: %d\n", vg)
			// 	vg = 255
			// }
			vg = (vg & 0xFF) << 8
			//vg = vg << 16
		}
		if as.renderRed {
			if g.HomePher > pheromoneMax {
				panic("TOO BIG")
			}
			//vr = uint32(g.HomePher / homedivisor)
			vr = uint32(g.HomePher) >> pherShift
			// if vr > 255 {
			// 	fmt.Printf("HOME > 255: %d\n", vg)
			// 	vr = 255
			// }
			vr = (vr & 0xFF)
			//vr = vr << 24
		}
		//return 0x000000FF | vg | vr
		return 0xFF000000 | vg | vr
	} else {
		return 0
	}
}

// func (as *AntScene) Init(g *Game[GameState], r *sdl.Renderer, s *GameState) error {
func (as *AntScene) Init() error {
	as.parallel = true
	as.renderPher = false
	as.renderGreen = true
	as.renderRed = true
	as.propPher = false
	as.pause = true
	as.pherOverloadFactor = 500
	f, err := NewField[gridspot](WIDTH, HEIGHT, as.renderGridspot)
	if err != nil {
		return err
	}
	as.field = f

	//as.textures = make([]*sdl.Texture, int(END))
	as.textures = make([]*ebiten.Image, int(END))
	as.fullTextures = make([]*ebiten.Image, int(END))
	for i := N; i < END; i++ {
		// t, err := r.CreateTexture(uint32(sdl.PIXELFORMAT_RGBA8888), sdl.TEXTUREACCESS_STATIC, antTexSize, antTexSize)
		// if err != nil {
		// 	return err
		// }
		// as.textures[i] = t
		// bs := make([]uint32, antTexSize*antTexSize+1)
		// for j := 0; j < antTexSize*antTexSize+1; j++ {
		// 	bs[j] = 0xc35b31ff
		// }
		// as.textures[i].UpdateRGBA(nil, bs, antTexSize)
		as.textures[i] = ebiten.NewImage(antTexSize, antTexSize)
		as.textures[i].Fill(color.RGBA{R: 0xc3, G: 0x5b, B: 0x31, A: 0xff})
		as.fullTextures[i] = ebiten.NewImage(antTexSize, antTexSize)
		as.fullTextures[i].Fill(color.RGBA{R: 0xc3, G: 0x5b, B: 0xff, A: 0xff})
	}

	for y := 0; y < HEIGHT; y++ {
		for x := 0; x < WIDTH; x++ {
			as.field.Get(x, y).Wall = true
			as.field.Update(x, y)
		}
	}

	for x := 0; x < 100; x++ {
		for y := 0; y < 100; y++ {
			as.field.Get(x, y).Wall = false
			as.field.Get(x, y).Home = true
			as.field.Update(x, y)
		}
	}

	for a := range as.ants {
		as.ants[a].life = antlife
	}
	if as.homefood == 0 {
		as.homefood = int64(len(as.ants)) * antlife * spawnParam
	}

	// TTF
	tt, err := opentype.Parse(fonts.MPlus1pRegular_ttf)
	if err != nil {
		log.Fatal(err)
	}

	const dpi = 72
	mplusNormalFont, err = opentype.NewFace(tt, &opentype.FaceOptions{
		Size:    24,
		DPI:     dpi,
		Hinting: font.HintingVertical,
	})

	for i := 0; i < workers; i++ {
		as.antworkerTrigger = append(as.antworkerTrigger, make(chan struct{}))
		go func(i int) {
			for range as.antworkerTrigger[i] {
				partsize := (len(as.ants) / workers) + 1
				as.UpdateAntPartial((partsize * i), (partsize*i)+partsize)
				as.antwg.Done()
			}
		}(i)

		as.pherworkerTrigger = append(as.pherworkerTrigger, make(chan struct{}))
		go func(i int) {
			partsize := (as.field.height / workers) + 1
			for range as.pherworkerTrigger[i] {
				as.UpdatePherPartial((partsize * i), (partsize*i)+partsize)
				as.pherwg.Done()
			}
		}(i)

	}

	return nil
}

func (as *AntScene) relocateAnts() {
	for a := range as.ants {
		as.ants[a].pos.x = 0
		as.ants[a].pos.y = 0
	}
}

var frame uint64

func (as *AntScene) UpdateAntPartial(start, end int) {
	if start >= len(as.ants) {
		return
	}
	if end > len(as.ants) {
		end = len(as.ants)
	}
	//fmt.Printf("UPDATING %d - %d\n", start, end)
	for a := range as.ants[start:end] {
		as.ants[start+a].Update(as)
	}
}

//var newfoodPherMaxPresent int32
//var newhomePherMaxPresent int32

func (as *AntScene) UpdatePherPartial(start, end int) {
	if start >= as.field.height {
		return
	}
	if end > as.field.height {
		end = as.field.height
	}

	// localNewfoodPherMaxPresent := 1
	// localNewhomePherMaxPresent := 1

	for y := start; y < end; y++ {
		for x := 0; x < as.field.width; x++ {
			update := false
			spot := as.field.Get(x, y)
			// TODO: Atomics or something. These are data races
			// if spot.FoodPher > localNewfoodPherMaxPresent {
			// 	localNewfoodPherMaxPresent = spot.FoodPher
			// }
			// if spot.HomePher > localNewhomePherMaxPresent {
			// 	localNewhomePherMaxPresent = spot.HomePher
			// }
			if spot.FoodPher > 0 {
				spot.FoodPher -= (spot.FoodPher / fadeDivisor) + 1
				update = true
			}
			if spot.HomePher > 0 {
				spot.HomePher -= (spot.HomePher / fadeDivisor) + 1
				update = true
			}

			if update && as.renderPher {
				as.field.Update(x, y)
			}
		}
	}
	// for {
	// 	if v := atomic.LoadInt32(&newfoodPherMaxPresent); int32(localNewfoodPherMaxPresent) > v {
	// 		if atomic.CompareAndSwapInt32(&newfoodPherMaxPresent, v, int32(localNewfoodPherMaxPresent)) {
	// 			break
	// 		}
	// 	} else {
	// 		break
	// 	}
	// }
	// for {
	// 	if v := atomic.LoadInt32(&newhomePherMaxPresent); int32(localNewhomePherMaxPresent) > v {
	// 		if atomic.CompareAndSwapInt32(&newhomePherMaxPresent, v, int32(localNewhomePherMaxPresent)) {
	// 			break
	// 		}
	// 	} else {
	// 		break
	// 	}
	// }
}

// func (as *AntScene) Update(g *Game[GameState], r *sdl.Renderer, s *GameState) error {
func (as *AntScene) Update() error {
	if err := as.HandleInput(); err != nil {
		return err
	}
	if as.pause {
		return nil
	}
	frame++

	n := maxants / antlife
	if n == 0 {
		n = 1
	}
	for i := 0; i < n; i++ {
		if as.homefood/(antlife*spawnParam) > int64(len(as.ants)) {
			//fmt.Printf("GREATER!\n")
			as.homefood -= antlife
			as.ants = append(as.ants, Ant{life: antlife})
		}
	}

	if frame%10 == 0 {
		fmt.Printf("homefood: %d, ants: %d, ratio: %d / %d \n", as.homefood, len(as.ants), as.homefood/(antlife*spawnParam), len(as.ants))
	}

	// partsize := (len(as.ants) / workers) + 1
	// for i := 0; i < workers; i++ {
	// 	as.UpdateAntPartial((partsize * i), (partsize*i)+partsize)
	// }

	if as.parallel {
		as.antwg.Add(workers)
		for i := 0; i < workers; i++ {
			as.antworkerTrigger[i] <- struct{}{}
		}
		as.antwg.Wait()
	} else {
		as.UpdateAntPartial(0, len(as.ants))
	}

	var k int
	for a := range as.ants {
		if as.ants[a].life < 0 {
			continue
		}
		as.ants[k] = as.ants[a]
		k++
	}
	as.ants = as.ants[:k]

	// newfoodPherMaxPresent = 1
	// newhomePherMaxPresent = 1

	if as.parallel {
		as.pherwg.Add(workers)
		for i := 0; i < workers; i++ {
			as.pherworkerTrigger[i] <- struct{}{}
		}
		as.pherwg.Wait()
	} else {
		as.UpdatePherPartial(0, as.field.height)
	}

	// foodPherMaxPresent = int(newfoodPherMaxPresent)
	// homePherMaxPresent = int(newhomePherMaxPresent)
	return nil
}

// Pheromone propagation from ants update loop.
// Doesn't work well, but may be interesting in the future.
// if as.propPher {
// 	hasFood := spot.FoodPher > marker
// 	hasHome := spot.HomePher > marker
// 	if hasFood || hasHome {
// 		pt := point{x, y}
// 		if pt.Within(1, 1, WIDTH-2, HEIGHT-2) {
// 			if hasFood {
// 				spot.FoodPher /= 2
// 				update = true
// 				for d := N; d < END; d++ {
// 					pt2 := pt.PointAt(d)
// 					spot2 := as.field.Get(pt2.x, pt2.y)
// 					if spot2.FoodPher < pheromoneMax {
// 						spot2.FoodPher += (spot.FoodPher / 9)
// 						if as.renderPher {
// 							as.field.Update(pt2.x, pt2.y)
// 						}
// 					}
// 				}
// 			}
// 			if hasHome {
// 				spot.HomePher /= 2
// 				update = true
// 				for d := N; d < END; d++ {
// 					pt2 := pt.PointAt(d)
// 					spot2 := as.field.Get(pt2.x, pt2.y)
// 					if spot2.HomePher < pheromoneMax {
// 						spot2.HomePher += (spot.HomePher / 9)
// 						if as.renderPher {
// 							as.field.Update(pt2.x, pt2.y)
// 						}
// 					}
// 				}
// 			}
// 		}
// 	}
// } else if as.oldPropPher {
// 	hasFood := spot.FoodPher > 100
// 	hasHome := spot.HomePher > 100
// 	if hasFood || hasHome {
// 		pt := point{x, y}
// 		if pt.Within(1, 1, WIDTH-2, HEIGHT-2) {
// 			if hasFood {
// 				for d := N; d < END; d++ {
// 					pt2 := pt.PointAt(d)
// 					spot2 := as.field.Get(pt2.x, pt2.y)
// 					spot2.FoodPher += (spot.FoodPher / 9)
// 					if as.renderPher {
// 						as.field.Update(pt2.x, pt2.y)
// 					}
// 				}
// 				spot.FoodPher /= 9
// 				update = true
// 			}
// 			if hasHome {
// 				for d := N; d < END; d++ {
// 					pt2 := pt.PointAt(d)
// 					spot2 := as.field.Get(pt2.x, pt2.y)
// 					spot2.HomePher += (spot.HomePher / 9)
// 					if as.renderPher {
// 						as.field.Update(pt2.x, pt2.y)
// 					}
// 				}
// 				spot.HomePher /= 9
// 				update = true
// 			}
// 		}
// 	}
// }

// func (as *AntScene) Render(g *Game[GameState], r *sdl.Renderer, s *GameState) error {
func (as *AntScene) Draw(screen *ebiten.Image) {
	err := as.field.Render(screen)
	if err != nil {
		panic(err)
	}

	var dio ebiten.DrawImageOptions
	for a := range as.ants {
		if as.ants[a].food > 0 {
			im := as.fullTextures[as.ants[a].dir]
			dio.GeoM = ebiten.GeoM{}
			dio.GeoM.Translate(float64(as.ants[a].pos.x-(antTexSize/2)), float64(as.ants[a].pos.y-(antTexSize/2)))
			screen.DrawImage(im, &dio)
		} else {
			im := as.textures[as.ants[a].dir]
			dio.GeoM = ebiten.GeoM{}
			dio.GeoM.Translate(float64(as.ants[a].pos.x-(antTexSize/2)), float64(as.ants[a].pos.y-(antTexSize/2)))
			screen.DrawImage(im, &dio)
		}
	}
	msg := fmt.Sprintf("FPS: %02.f, Ticks/Sec: %0.2f, Hive Food: %d, Ants: %d", ebiten.ActualFPS(), ebiten.ActualTPS(), as.homefood, len(as.ants))
	text.Draw(screen, msg, mplusNormalFont, 10, 40, color.White)
	return
}

func (as *AntScene) RenderBelow() bool {
	return true
}

func (as *AntScene) Destroy() {
}

// Layout takes the outside size (e.g., the window size) and returns the (logical) screen size.
// If you don't have to adjust the screen size with the outside size, just return a fixed size.
func (as *AntScene) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return WIDTH, HEIGHT
}

func absi(i int) int {
	if i < 0 {
		return -i
	}
	return i
}

// Bresenham's line algorithm
func doLine(x0, y0, x1, y1 int, f func(x, y int)) {
	dx := absi(x1 - x0)
	var sx int
	if x0 < x1 {
		sx = 1
	} else {
		sx = -1
	}
	dy := -absi(y1 - y0)
	var sy int
	if y0 < y1 {
		sy = 1
	} else {
		sy = -1
	}
	error := dx + dy

	for {
		f(x0, y0)
		if x0 == x1 && y0 == y1 {
			break
		}
		e2 := 2 * error
		if e2 >= dy {
			if x0 == x1 {
				break
			}
			error = error + dy
			x0 = x0 + sx
		}
		if e2 <= dx {
			if y0 == y1 {
				break
			}
			error = error + dx
			y0 = y0 + sy
		}
	}
}
