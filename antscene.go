package main

import (
	"encoding/gob"
	"fmt"
	"image/color"
	"log"
	"os"
	"sync"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"

	"github.com/hajimehoshi/ebiten/v2/examples/resources/fonts"
)

var workers = 6
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
const pheromoneMax = 8000
const marker = 2500
const fadeDivisor = 700 // bigger number, slower pheromone fade.
const antFadeDivisor = 600
const pheromoneExtend = 200
const antlife = 10000 // an ant spends 1 life per frame
const foodlife = 2000 // amount of life 1 food gives an ant

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
					spot := as.field.Get(int(i), int(j))
					spot.Wall = true
					spot.Home = false
					spot.Food = 0
					as.field.Update(int(i), int(j))
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
					spot := as.field.Get(int(i), int(j))
					spot.Wall = false
					spot.Home = false
					spot.Food = 0
					as.field.Update(int(i), int(j))
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
					spot := as.field.Get(int(i), int(j))
					spot.Wall = false
					spot.Home = false
					spot.Food = foodCount
					as.field.Update(int(i), int(j))
				}
			}
		}
	}
	mx, my := ebiten.CursorPosition()
	as.mousePX = mx
	as.mousePY = my
	return nil
}

var homePherMaxPresent = 1
var foodPherMaxPresent = 1

func (as *AntScene) renderGridspot(g *gridspot) uint32 {
	homedivisor := (homePherMaxPresent / 255) + 1
	fooddivisor := (foodPherMaxPresent / 255) + 1
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
			vg = uint32(g.FoodPher / fooddivisor)
			if vg > 255 {
				vg = 255
			}
			//vg = vg << 16
			vg = vg << 8
		}
		if as.renderRed {
			vr = uint32(g.HomePher / homedivisor)
			if vr > 255 {
				vr = 255
			}
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
		as.homefood = int64(len(as.ants)) * antlife * 10
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

	return nil
}

func (as *AntScene) relocateAnts() {
	for a := range as.ants {
		as.ants[a].pos.x = 0
		as.ants[a].pos.y = 0
	}
}

var frame uint64

// func (as *AntScene) Update(g *Game[GameState], r *sdl.Renderer, s *GameState) error {
func (as *AntScene) Update() error {
	frame++

	if as.homefood/(antlife*10) > int64(len(as.ants)) {
		as.homefood -= antlife
		as.ants = append(as.ants, Ant{life: antlife})
	}

	if frame%120 == 0 {
		fmt.Printf("homefood: %d, ants: %d\n", as.homefood, len(as.ants))
	}
	if err := as.HandleInput(); err != nil {
		return err
	}
	if as.pause {
		return nil
	}
	var k int
	for a := range as.ants {
		as.ants[a].Move(as)
		if as.ants[a].life <= 0 {
			//as.field.Get(as.ants[a].pos.x, as.ants[a].pos.y).Food = 1
			// if as.ants[a].food > 0 {
			// 	as.field.Get(as.ants[a].pos.x, as.ants[a].pos.y).Food += as.ants[a].food
			// }
			continue
		}
		if as.field.Get(as.ants[a].pos.x, as.ants[a].pos.y).Home {
			if as.ants[a].food > 0 {
				as.homefood += int64(as.ants[a].food) * foodlife
				as.ants[a].food = 0
			}
			// need := int64(antlife - as.ants[a].life)
			// if need > as.homefood {
			// 	need = as.homefood
			// }
			// as.homefood -= need
			// as.ants[a].life += int(need)
			as.ants[a].marker = marker
		}
		if spot := as.field.Get(as.ants[a].pos.x, as.ants[a].pos.y); spot.Food > 0 {
			if as.ants[a].food == 0 {
				as.ants[a].dir = as.ants[a].dir.Right(4)
				if spot.Food > 10 {
					spot.Food -= 10
					as.field.Update(as.ants[a].pos.x, as.ants[a].pos.y)
					as.ants[a].food = 10
				} else {
					as.ants[a].food = spot.Food
					spot.Food = 0
					as.field.Update(as.ants[a].pos.x, as.ants[a].pos.y)
				}
			}
			as.ants[a].marker = marker
		}

		if as.ants[a].food > 0 {
			spot := as.field.Get(as.ants[a].pos.x, as.ants[a].pos.y)
			if spot.FoodPher > as.ants[a].marker {
				as.ants[a].marker = spot.FoodPher
			} else {
				spot.FoodPher = as.ants[a].marker
				as.ants[a].marker -= (as.ants[a].marker / antFadeDivisor) + 1
				if as.renderPher {
					as.field.Update(as.ants[a].pos.x, as.ants[a].pos.y)
				}
			}
		} else {
			spot := as.field.Get(as.ants[a].pos.x, as.ants[a].pos.y)
			if spot.HomePher > as.ants[a].marker {
				as.ants[a].marker = spot.HomePher
			} else {
				spot.HomePher = as.ants[a].marker
				as.ants[a].marker -= (as.ants[a].marker / antFadeDivisor) + 1
				if as.renderPher {
					as.field.Update(as.ants[a].pos.x, as.ants[a].pos.y)
				}
			}
		}
		as.ants[k] = as.ants[a]
		k++
	}
	//fmt.Printf("Copyng ants(%d) to ants[:%d]\n", len(as.ants), k)
	as.ants = as.ants[:k]
	newfoodPherMaxPresent := 1
	newhomePherMaxPresent := 1

	for y := 0; y < as.field.height; y++ {
		for x := 0; x < as.field.width; x++ {
			update := false
			spot := as.field.Get(x, y)
			if spot.FoodPher > newfoodPherMaxPresent {
				newfoodPherMaxPresent = spot.FoodPher
			}
			if spot.HomePher > newhomePherMaxPresent {
				newhomePherMaxPresent = spot.HomePher
			}
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

	foodPherMaxPresent = newfoodPherMaxPresent
	homePherMaxPresent = newhomePherMaxPresent
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
	msg := fmt.Sprintf("Ticks/Sec: %0.2f, Hive Food: %d, Ants: %d", ebiten.ActualTPS(), as.homefood, len(as.ants))
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
