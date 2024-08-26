package main

import (
	"fmt"
	"image/color"
	"math"
	"math/rand"
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

const antTexSize = 5
const pheromoneMax = 8191
const pherShift = 5 //(2^13 = 8192), meaning 8192 is within 13 bits range, We want to shift that to 8 bits, so shift 5 out.
const marker = 5000

// DrawData is passed to draw workers, which will draw ants to screen.
type DrawData struct {
	screen *ebiten.Image
	ants   []Ant
}

type AntScene struct {
	st             *GameState
	ants           []Ant
	field          *DField //*Field[gridspot]
	textures       []*ebiten.Image
	fullTextures   []*ebiten.Image
	pause          bool
	mousePX        int
	mousePY        int
	gridWorkChan   chan gridwork
	antWorkChan    chan gridwork
	renderWorkChan chan renderwork
	homelife       int64

	antwg            sync.WaitGroup
	antworkerTrigger []chan struct{}

	pherwg            sync.WaitGroup
	pherworkerTrigger []chan struct{}

	drawwg         sync.WaitGroup
	drawworkerData []chan DrawData
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

	//enc := gob.NewEncoder(f)
	//return enc.Encode(as.field.vals)
	panic("NOT IMPLEMENTED FOR DFIELD")
	return nil
}

func (as *AntScene) LoadGrid() error {
	panic("NOT IMPLEMENTED FOR DFIELD")
	// f, err := os.Open("ants.grid")
	// if err != nil {
	// 	return err
	// }
	// defer f.Close()

	// enc := gob.NewDecoder(f)
	// g := []gridspot{}
	// err = enc.Decode(&g)
	// if err != nil {
	// 	return err
	// }
	// as.field.vals = g
	// as.field.UpdateAll()

	return nil
}

// func (as *AntScene) HandleEvent(g *Game[GameState], r *sdl.Renderer, e sdl.Event) error {
func (as *AntScene) HandleInput(g *Game[GameState]) error {
	if inpututil.IsKeyJustPressed(ebiten.KeyP) {
		as.st.renderPher = !as.st.renderPher
		fmt.Printf("RENDER PHEROMONES: %t\n", as.st.renderPher)
		as.field.UpdateAll(as)
	} else if inpututil.IsKeyJustPressed(ebiten.KeyM) {
		o := &OptScene{as: as}
		fmt.Printf("PushingScene\n")
		err := g.PushScene(o)
		fmt.Printf("Done Pushing Scene\n")
		if err != nil {
			return err
		}
	} else if inpututil.IsKeyJustPressed(ebiten.KeyG) {
		as.st.renderGreen = !as.st.renderGreen
	} else if inpututil.IsKeyJustPressed(ebiten.KeyR) {
		as.st.renderRed = !as.st.renderRed
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
		as.relocateAnts()
	} else if inpututil.IsKeyJustPressed(ebiten.KeyC) {
		as.field.Clear(as)
		for x := 0; x < 100; x++ {
			for y := 0; y < 100; y++ {
				//as.field.Get(x, y).Home = true
				as.field.home[as.field.Idx(x, y)] = true
				as.field.Update(as, x, y)
			}
		}
		as.relocateAnts()
	} else if inpututil.IsKeyJustPressed(ebiten.KeyF) {
		as.field.Clear(as)
		for y := 0; y < g.height; y++ {
			for x := 0; x < g.width; x++ {
				i := as.field.Idx(x, y)
				as.field.wall[i] = true
				as.field.Update(as, x, y)
			}
		}

		for x := 0; x < 100; x++ {
			for y := 0; y < 100; y++ {
				as.field.home[as.field.Idx(x, y)] = true
				as.field.Update(as, x, y)
			}
		}
	} else if inpututil.IsKeyJustPressed(ebiten.KeyX) {
		as.st.parallel = !as.st.parallel
		fmt.Printf("Parallel update: %t\n", as.st.parallel)
	} else if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		as.pause = !as.pause
	} else if inpututil.IsKeyJustPressed(ebiten.KeyW) {
		as.st.followWalls = !as.st.followWalls
		fmt.Printf("Wall Following: %t\n", as.st.followWalls)
	} else if inpututil.IsKeyJustPressed(ebiten.KeyUp) {
		as.st.drawradius++
	} else if inpututil.IsKeyJustPressed(ebiten.KeyDown) {
		as.st.drawradius--
	} else if inpututil.IsKeyJustPressed(ebiten.KeyLeft) {
		g.state.leftmode -= 1
		if g.state.leftmode < 0 {
			g.state.leftmode = end - 1
		}
	} else if inpututil.IsKeyJustPressed(ebiten.KeyRight) {
		g.state.leftmode = (g.state.leftmode + 1) % end
	} else if inpututil.IsKeyJustPressed(ebiten.KeyZ) {
		g.state.renderAnts = !g.state.renderAnts
	} else if inpututil.IsKeyJustPressed(ebiten.KeyQ) {
		g.state.cluster = !g.state.cluster
	}

	distance := func(x0, y0, x1, y1 int) int {
		dx := x0 - x1
		dy := y0 - y1
		return int(math.Sqrt(float64(dx*dx) + float64(dy*dy)))
	}

	//radius := 15
	doSpot := func(x, y int, f func(x, y, idx int)) {
		for i := x - as.st.drawradius; i < x+as.st.drawradius; i++ {
			if i < 0 || i >= g.width {
				continue
			}
			for j := y - as.st.drawradius; j < y+as.st.drawradius; j++ {
				if j < 0 || j >= g.height {
					continue
				}
				//fmt.Printf("x0: %d, y0: %d, x1: %d, y1: %d, Dist: %d\n", i, j, x, y, distance(i, j, x, y))
				if distance(i, j, x, y) > as.st.drawradius {
					continue
				}
				//spot := as.field.Get(int(i), int(j))
				f(int(i), int(j), as.field.Idx(i, j))
			}
		}
	}

	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) && g.state.leftmode == wall {
		mx, my := ebiten.CursorPosition()
		//if mx != as.mousePX || my != as.mousePY {
		doLine(mx, my, as.mousePX, as.mousePY, func(cx, cy int) {
			doSpot(cx, cy, func(x, y, idx int) {
				if as.field.home[idx] {
					return
				}
				as.field.wall[idx] = true
				//spot.Home = false
				as.field.food[idx] = 0
				as.field.Update(as, x, y)
			})
		})
		//}
	} else if ebiten.IsMouseButtonPressed(ebiten.MouseButtonRight) ||
		(ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) && g.state.leftmode == erase) {
		mx, my := ebiten.CursorPosition()
		//if mx != as.mousePX || my != as.mousePY {
		doLine(mx, my, as.mousePX, as.mousePY, func(cx, cy int) {
			doSpot(cx, cy, func(x, y, idx int) {
				as.field.wall[idx] = false
				//spot.Home = false
				as.field.food[idx] = 0
				as.field.Update(as, x, y)
			})
		})
		//}
	} else if ebiten.IsMouseButtonPressed(ebiten.MouseButtonMiddle) ||
		(ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) && g.state.leftmode == food) {
		mx, my := ebiten.CursorPosition()
		//if mx != as.mousePX || my != as.mousePY {
		doLine(mx, my, as.mousePX, as.mousePY, func(cx, cy int) {
			doSpot(cx, cy, func(x, y, idx int) {
				as.field.wall[idx] = false
				//spot.Home = false
				as.field.food[idx] = as.st.foodcount
				as.field.Update(as, x, y)
			})
		})
		//}
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
	if as.st.renderPher {
		var (
			vg uint32
			vr uint32
		)
		if as.st.renderGreen {
			//vg = uint32(g.FoodPher / fooddivisor)
			vg = uint32(g.FoodPher) >> pherShift
			// if vg > 255 {
			// 	fmt.Printf("FOOD > 255: %d\n", vg)
			// 	vg = 255
			// }
			vg = (vg & 0xFF) << 8
			//vg = vg << 16
		}
		if as.st.renderRed {
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

func drawAntTextures(c color.Color) []*ebiten.Image {
	textures := make([]*ebiten.Image, int(END))
	setColor := func(im *ebiten.Image, c color.Color) func(x, y int) {
		return func(x, y int) {
			im.Set(x, y, c)
		}
	}
	//N
	textures[N] = ebiten.NewImage(antTexSize, antTexSize)
	doLine(antTexSize/2, 0, antTexSize/2, antTexSize, setColor(textures[N], c))

	//NE
	textures[NE] = ebiten.NewImage(antTexSize, antTexSize)
	doLine(0, antTexSize, antTexSize, 0, setColor(textures[NE], c))

	textures[E] = ebiten.NewImage(antTexSize, antTexSize)
	doLine(0, antTexSize/2, antTexSize, antTexSize/2, setColor(textures[E], c))

	textures[SE] = ebiten.NewImage(antTexSize, antTexSize)
	doLine(0, 0, antTexSize, antTexSize, setColor(textures[SE], c))

	textures[S] = ebiten.NewImage(antTexSize, antTexSize)
	doLine(antTexSize/2, 0, antTexSize/2, antTexSize, setColor(textures[S], c))

	textures[SW] = ebiten.NewImage(antTexSize, antTexSize)
	doLine(0, antTexSize, antTexSize, 0, setColor(textures[SW], c))

	textures[W] = ebiten.NewImage(antTexSize, antTexSize)
	doLine(0, antTexSize/2, antTexSize, antTexSize/2, setColor(textures[W], c))

	textures[NW] = ebiten.NewImage(antTexSize, antTexSize)
	doLine(0, 0, antTexSize, antTexSize, setColor(textures[NW], c))

	return textures
}

func (as *AntScene) Init(g *Game[GameState], st *GameState) error {
	as.st = st
	as.pause = true
	f, err := NewDField(g.width, g.height)
	if err != nil {
		return err
	}
	as.field = f

	as.textures = make([]*ebiten.Image, int(END))
	as.fullTextures = make([]*ebiten.Image, int(END))
	//for i := N; i < END; i++ {
	// as.textures[i] = ebiten.NewImage(antTexSize, antTexSize)
	// as.textures[i].Fill(color.RGBA{R: 0xc3, G: 0x5b, B: 0x31, A: 0xff})
	// as.fullTextures[i] = ebiten.NewImage(antTexSize, antTexSize)
	// as.fullTextures[i].Fill(color.RGBA{R: 0xc3, G: 0x5b, B: 0xff, A: 0xff})
	as.textures = drawAntTextures(color.RGBA{R: 0xc3, G: 0x5b, B: 0x31, A: 0xff})
	as.fullTextures = drawAntTextures(color.RGBA{R: 0xc3, G: 0x5b, B: 0xff, A: 0xff})
	gentiles(as)
	//}

	for y := 0; y < g.height; y++ {
		for x := 0; x < g.width; x++ {
			//as.field.Get(x, y).Wall = true
			as.field.wall[as.field.Idx(x, y)] = true
			as.field.Update(as, x, y)
		}
	}

	for x := 0; x < 100; x++ {
		for y := 0; y < 100; y++ {
			idx := as.field.Idx(x, y)
			as.field.wall[idx] = false
			as.field.home[idx] = true
			as.field.Update(as, x, y)
		}
	}

	for a := range as.ants {
		as.ants[a].life = as.st.antlife
		as.ants[a].tex = as.textures[N]
	}
	if as.homelife == 0 {
		as.homelife = int64(len(as.ants)) * int64(as.st.antlife) * int64(as.st.stockpile)
	}

	// TTF
	tt, err := opentype.Parse(fonts.MPlus1pRegular_ttf)
	if err != nil {
		return err
	}

	const dpi = 72
	mplusNormalFont, err = opentype.NewFace(tt, &opentype.FaceOptions{
		Size:    antsceneFontSize,
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

		dc := make(chan DrawData)
		as.drawworkerData = append(as.drawworkerData, dc)
		go func() {
			for d := range dc {
				var dio ebiten.DrawImageOptions
				var lastx, lasty int
				for a := range d.ants {
					if as.ants[a].pos.x == lastx && as.ants[a].pos.y == lasty {
						continue
					}
					lastx, lasty = as.ants[a].pos.x, as.ants[a].pos.y
					if as.ants[a].food > 0 {
						im := as.fullTextures[as.ants[a].dir]
						dio.GeoM = ebiten.GeoM{}
						dio.GeoM.Translate(float64(as.ants[a].pos.x-(antTexSize/2)), float64(as.ants[a].pos.y-(antTexSize/2)))
						d.screen.DrawImage(im, &dio)
					} else {
						im := as.textures[as.ants[a].dir]
						dio.GeoM = ebiten.GeoM{}
						dio.GeoM.Translate(float64(as.ants[a].pos.x-(antTexSize/2)), float64(as.ants[a].pos.y-(antTexSize/2)))
						d.screen.DrawImage(im, &dio)
					}
				}
				as.drawwg.Done()
			}
		}()

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
	for a := range as.ants[start:end] {
		as.ants[start+a].Update(as)
	}
}

func (as *AntScene) UpdatePherPartial(start, end int) {
	if start >= as.field.height {
		return
	}
	if end > as.field.height {
		end = as.field.height
	}

	for y := start; y < end; y++ {
		for x := 0; x < as.field.width; x++ {
			update := false
			//spot := as.field.Get(x, y)
			idx := as.field.Idx(x, y)
			if as.field.foodpher[idx] > 0 { //spot.FoodPher > 0 {
				//spot.FoodPher -= (spot.FoodPher / as.st.fadedivisor) + 1
				as.field.foodpher[idx] -= (as.field.foodpher[idx] / as.st.fadedivisor) + 1
				update = true
			}
			if as.field.homepher[idx] > 0 {
				//spot.HomePher -= (spot.HomePher / as.st.fadedivisor) + 1
				as.field.homepher[idx] -= (as.field.homepher[idx] / as.st.fadedivisor) + 1
				update = true
			}

			if update && as.st.renderPher {
				as.field.Update(as, x, y)
			}
		}
	}
}

// func (as *AntScene) Update(g *Game[GameState], r *sdl.Renderer, s *GameState) error {
func (as *AntScene) Update(g *Game[GameState], st *GameState) error {
	as.st = st
	if err := as.HandleInput(g); err != nil {
		return err
	}
	if as.pause {
		return nil
	}
	frame++

	n := as.st.maxants / as.st.antlife
	if n == 0 {
		n = 1
	}
	for i := 0; i < n; i++ {
		if len(as.ants) < st.maxants && as.homelife/(int64(st.antlife)*int64(st.stockpile)) > int64(len(as.ants)) {
			as.homelife -= int64(st.antlife)
			as.ants = append(as.ants, Ant{life: as.st.antlife, tex: as.textures[N]})
		}
	}

	if frame%10 == 0 {
		fmt.Printf("n: %d, homefood: %d, ants: %d, ratio: %d / %d \n",
			n, as.homelife, len(as.ants), as.homelife/(int64(st.antlife)*int64(st.stockpile)), len(as.ants))
	}

	// partsize := (len(as.ants) / workers) + 1
	// for i := 0; i < workers; i++ {
	// 	as.UpdateAntPartial((partsize * i), (partsize*i)+partsize)
	// }

	if st.parallel {
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

	if st.parallel {
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

func (as *AntScene) DrawUnder(g *Game[GameState], _ *GameState) bool {
	return false
	//return true
}

type sortYX []Ant

func (s sortYX) Len() int {
	return len([]Ant(s))
}
func (s sortYX) Less(i, j int) bool {
	if []Ant(s)[i].pos.y < []Ant(s)[j].pos.y {
		return true
	}
	return []Ant(s)[i].pos.x < []Ant(s)[j].pos.x
}
func (s sortYX) Swap(i, j int) {
	sl := []Ant(s)
	sl[i], sl[j] = sl[j], sl[i]
}

func qsortAnts(a []Ant) {
	if len(a) < 2 {
		return
	}
	left, right := 0, len(a)-1
	pivot := len(a) / 2
	a[pivot], a[right] = a[right], a[pivot]
	for i := range a {
		if a[i].pos.y < a[right].pos.y || (a[i].pos.y == a[right].pos.y && a[i].pos.x < a[right].pos.x) {
			a[i], a[left] = a[left], a[i]
			left++
		}
	}
	a[left], a[right] = a[right], a[left]
	qsortAnts(a[:left])
	qsortAnts(a[left+1:])
}

func mergeSortAnts(a []Ant) []Ant {
	if len(a) <= 1 {
		return a
	}
	left := mergeSortAnts(a[:len(a)/2])
	right := mergeSortAnts(a[(len(a) / 2):])
	return merge(left, right)
}

func merge(as, bs []Ant) []Ant {
	//fmt.Printf("Merging as[%d] bs[%d]\n", len(as), len(bs))
	rs := make([]Ant, 0, len(as)+len(bs))
	// if len(space) != len(as)+len(bs) {
	// 	panic(fmt.Sprintf("BAD SPACE A: %d, B: %d, SPACE: %d", len(as), len(bs), len(space)))
	// }
	// space = space[:0]
	for len(as) > 0 && len(bs) > 0 {
		if as[0].pos.y < bs[0].pos.y || (as[0].pos.y == bs[0].pos.y && as[0].pos.x < bs[0].pos.x) {
			rs = append(rs, as[0])
			//space = append(space, as[0])
			as = as[1:]
		} else {
			rs = append(rs, bs[0])
			//space = append(space, bs[0])
			bs = bs[1:]
		}
	}
	if len(as) > 0 {
		rs = append(rs, as...)
		//space = append(space, as...)
	} else if len(bs) > 0 {
		//space = append(space, bs...)
		rs = append(rs, bs...)
	}
	//fmt.Printf("Merged into len(%d)\n", len(rs))
	return rs
}

func cmpAnt(a, b Ant) int {
	//ay := a.pos.y << 8
	//by := b.pos.y << 8

	return ((a.pos.y - b.pos.y) << 32) + (a.pos.x - b.pos.x)

	// if a.pos.y < b.pos.y {
	// 	return -1
	// } else if a.pos.y == b.pos.y {
	// 	if a.pos.x < b.pos.x {
	// 		return -1
	// 	} else if a.pos.x == b.pos.x {
	// 		return 0
	// 	}
	// }
	// return 1
}

// var chunksize = 16
// var maxperchunk uint8 = 1
var chunksize = 1
var maxperchunk uint8 = 32
var cycle = 0

type space struct {
	brown uint
	blue  uint
}

var markers []uint8 = make([]uint8, WIDTH*HEIGHT)

func memsetRepeat(a []uint8, v uint8) {
	if len(a) == 0 {
		return
	}
	a[0] = v
	for bp := 1; bp < len(a); bp *= 2 {
		copy(a[bp:], a[:bp])
	}
}

// func (as *AntScene) Render(g *Game[GameState], r *sdl.Renderer, s *GameState) error {
func (as *AntScene) Draw(g *Game[GameState], st *GameState, screen *ebiten.Image) {
	err := as.field.Render(screen)
	if err != nil {
		panic(err)
	}

	if st.renderAnts {
		if st.cluster {
			if !as.pause {
				afps := ebiten.ActualFPS()
				if afps < 45 {
					if maxperchunk == 1 && chunksize < 8 {
						chunksize *= 2
						maxperchunk *= 4
					}
					if maxperchunk > 1 {
						cycle -= 1
						if cycle <= -30 || afps < 15 {
							cycle = 0
							maxperchunk -= 1
						}
					}
				} else if afps > 55 {
					if maxperchunk == 32 {
						if chunksize > 1 {
							//if chunksize > 64 {
							chunksize = chunksize / 2
							maxperchunk = maxperchunk / 4
						}
					} else if maxperchunk < 32 {
						cycle += 1
						if cycle == 30 {
							cycle = 0
							maxperchunk++
						}
					}
				}
			}
			memsetRepeat(markers, 0)
			mask := ^(chunksize - 1)
			var dio ebiten.DrawImageOptions
			for a := range as.ants {
				// if st.cluster {
				loc := ((as.ants[a].pos.y & mask) * WIDTH) + (as.ants[a].pos.x & mask)
				//if as.ants[a].food > 0 {
				markers[loc] += 1
				//} else {
				//	markers[loc].brown += 1
				//}
				if markers[loc] > maxperchunk {
					//if markers[loc].brown > maxperchunk {
					continue
				}
				markers[loc] += 1
				// }

				// if as.ants[a].food > 0 {
				// 	im := as.fullTextures[as.ants[a].dir]
				// 	dio.GeoM = ebiten.GeoM{}
				// 	dio.GeoM.Translate(float64(as.ants[a].pos.x-(antTexSize/2)), float64(as.ants[a].pos.y-(antTexSize/2)))
				// 	screen.DrawImage(im, &dio)
				// } else {
				// 	im := as.textures[as.ants[a].dir]
				// 	dio.GeoM = ebiten.GeoM{}
				// 	dio.GeoM.Translate(float64(as.ants[a].pos.x-(antTexSize/2)), float64(as.ants[a].pos.y-(antTexSize/2)))
				// 	screen.DrawImage(im, &dio)
				// }
				im := as.ants[a].tex
				dio.GeoM = ebiten.GeoM{}
				dio.GeoM.Translate(float64(as.ants[a].pos.x-(antTexSize/2)), float64(as.ants[a].pos.y-(antTexSize/2)))
				screen.DrawImage(im, &dio)
			}

			// if st.cluster {
			// 	var dio ebiten.DrawImageOptions
			// 	for y := 0; y < HEIGHT/chunksize; y++ {
			// 		for x := 0; x < WIDTH/chunksize; x++ {
			// 			loc := ((y * chunksize * WIDTH) + (x * chunksize))
			// 			if markers[loc].brown > maxperchunk {
			// 				ac := markers[loc].brown - maxperchunk
			// 				//ac &= 0xff
			// 				if ac > 255 {
			// 					ac = 255
			// 				}
			// 				im := tiles[chunksize][ac]
			// 				dio.GeoM = ebiten.GeoM{}
			// 				dio.GeoM.Translate(float64(x*chunksize)-3, float64(y*chunksize)-3)
			// 				screen.DrawImage(im, &dio)
			// 			}
			// 			if markers[loc].blue > maxperchunk {
			// 				ac := markers[loc].blue - maxperchunk
			// 				//ac &= 0xff
			// 				if ac > 255 {
			// 					ac = 255
			// 				}
			// 				im := bluetiles[chunksize][ac]
			// 				dio.GeoM = ebiten.GeoM{}
			// 				dio.GeoM.Translate(float64(x*chunksize)-3, float64(y*chunksize)-3)
			// 				screen.DrawImage(im, &dio)
			// 			}
			// 			markers[loc] = space{}
			// 		}
			// 	}
			// }
		} else {
			var dio ebiten.DrawImageOptions
			for a := range as.ants {
				//if as.ants[a].food > 0 {
				im := as.ants[a].tex
				dio.GeoM = ebiten.GeoM{}
				dio.GeoM.Translate(float64(as.ants[a].pos.x-(antTexSize/2)), float64(as.ants[a].pos.y-(antTexSize/2)))
				screen.DrawImage(im, &dio)
				//} else {
				// im := as.textures[as.ants[a].dir]
				// dio.GeoM = ebiten.GeoM{}
				// dio.GeoM.Translate(float64(as.ants[a].pos.x-(antTexSize/2)), float64(as.ants[a].pos.y-(antTexSize/2)))
				// screen.DrawImage(im, &dio)
				//}
			}
		}
	}
	msg := fmt.Sprintf("FPS: %02.f, Ticks/Sec: %0.2f, Draw Radius: %d, Hive Life: %d, Ants: %d, Brush: %s, Cluster: %v",
		ebiten.ActualFPS(), ebiten.ActualTPS(), st.drawradius, as.homelife, len(as.ants), as.st.leftmode, st.cluster)
	start := antsceneFontSize * 2
	text.Draw(screen, msg, mplusNormalFont, 10, start, color.White)
	text.Draw(screen, "(M) menu", mplusNormalFont, 10, start+antsceneFontSpace, color.White)
	text.Draw(screen, fmt.Sprintf("maxperchunk: %d, chunksize: %d", maxperchunk, chunksize), mplusNormalFont, 10, start+(antsceneFontSpace*2), color.White)
	return
}

func (as *AntScene) RenderBelow() bool {
	return true
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

// experimental tiling
var tiles [][]*ebiten.Image
var bluetiles [][]*ebiten.Image
var timg []*ebiten.Image

func gentiles(as *AntScene) {
	timg = make([]*ebiten.Image, 33)
	tiles = make([][]*ebiten.Image, 33)
	bluetiles = make([][]*ebiten.Image, 33)
	for i := 1; i < 64; i *= 2 {
		tiles[i] = make([]*ebiten.Image, 256)
		bluetiles[i] = make([]*ebiten.Image, 256)
		timg[i] = ebiten.NewImage(i, i)
		timg[i].Fill(color.RGBA{R: 0xc3, G: 0x5b, B: 0x31, A: 0xff})
		for n := 0; n < 256; n++ {
			im := ebiten.NewImage(i+6, i+6)
			blim := ebiten.NewImage(i+6, i+6)
			for k := 0; k < n; k++ {
				var dio ebiten.DrawImageOptions
				dio.GeoM.Translate(float64(rand.Intn(i)+3), float64(rand.Intn(i)+3))
				im.DrawImage(as.textures[rand.Intn(int(END))], &dio)
				dio.GeoM.Reset()
				dio.GeoM.Translate(float64(rand.Intn(i)+3), float64(rand.Intn(i)+3))
				blim.DrawImage(as.fullTextures[rand.Intn(int(END))], &dio)
			}
			tiles[i][n] = im
			bluetiles[i][n] = blim
		}
	}
}
