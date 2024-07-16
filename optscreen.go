package main

import (
	"fmt"
	"image/color"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/examples/resources/fonts"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
)

type OptScene struct {
	as    *AntScene
	blank *ebiten.Image
	font  font.Face
	index int
	opts  []opt
}

func (s *OptScene) Init(g *Game[GameState], st *GameState) error {
	s.blank = ebiten.NewImage(st.width, st.height)
	s.blank.Fill(color.RGBA{A: 0xbf})
	// fmt.Printf("Drawing Black\n")
	// draw.Draw(
	// 	s.blank,
	// 	image.Rect(0, 0, st.width, st.height),
	// 	&image.Uniform{color.RGBA{A: 0xaf}},
	// 	image.Point{},
	// 	draw.Src,
	// )
	// fmt.Printf("Done Drawing Black\n")

	tt, err := opentype.Parse(fonts.MPlus1pRegular_ttf)
	if err != nil {
		return err
	}

	const dpi = 72
	s.font, err = opentype.NewFace(tt, &opentype.FaceOptions{
		Size:    optsceneFontSize,
		DPI:     dpi,
		Hinting: font.HintingVertical,
	})
	if err != nil {
		return err
	}

	s.opts = makeTexts(st)

	return nil
}

func (s *OptScene) DrawUnder(g *Game[GameState], _ *GameState) bool {
	return true
}

type opt struct {
	name  string
	value string
	left  func(duration int)
	right func(duration int)
}

func withProgressiveDuration(f func(i int)) func(int) {
	return func(d int) {
		if d > 600 {
			f(50)
		} else if d > 300 {
			f(10)
		} else if d > 60 {
			f(1)
		} else if d == 1 {
			f(1)
		} else if d%4 == 0 {
			f(1)
		}
	}
}

func makeTexts(st *GameState) []opt {
	texts := []opt{
		{
			name:  "Render Pheromones (P)",
			value: fmt.Sprintf("%t", st.renderPher),
			left:  func(_ int) { st.renderPher = !st.renderPher },
			right: func(_ int) { st.renderPher = !st.renderPher },
		},
		{
			name:  "Render Green (G)",
			value: fmt.Sprintf("%t", st.renderGreen),
			left:  func(_ int) { st.renderGreen = !st.renderGreen },
			right: func(_ int) { st.renderGreen = !st.renderGreen },
		},
		{
			name:  "Render Red (R)",
			value: fmt.Sprintf("%t", st.renderRed),
			left:  func(_ int) { st.renderRed = !st.renderRed },
			right: func(_ int) { st.renderRed = !st.renderRed },
		},
		{
			name:  "Parallel Execution (X)",
			value: fmt.Sprintf("%t", st.parallel),
			left:  func(_ int) { st.parallel = !st.parallel },
			right: func(_ int) { st.parallel = !st.parallel },
		},
		{
			name:  "Follow Walls (W)",
			value: fmt.Sprintf("%t", st.followWalls),
			left:  func(_ int) { st.followWalls = !st.followWalls },
			right: func(_ int) { st.followWalls = !st.followWalls },
		},
		{
			name:  "Antisocial",
			value: fmt.Sprintf("%t", st.antisocial),
			left:  func(_ int) { st.antisocial = !st.antisocial },
			right: func(_ int) { st.antisocial = !st.antisocial },
		},

		{
			name:  "Ant Life (spend 1/tick)",
			value: fmt.Sprintf("%d", st.antlife),
			left:  withProgressiveDuration(func(x int) { st.antlife -= x }),
			right: withProgressiveDuration(func(x int) { st.antlife += x }),
		},
		{
			name:  "Ant Sight Distance",
			value: fmt.Sprintf("%d", st.sight),
			left:  withProgressiveDuration(func(x int) { st.sight -= x }),
			right: withProgressiveDuration(func(x int) { st.sight += x }),
		},
		{
			name:  "Food Life (Life from 1 food)",
			value: fmt.Sprintf("%d", st.foodlife),
			left:  withProgressiveDuration(func(x int) { st.foodlife -= x }),
			right: withProgressiveDuration(func(x int) { st.foodlife += x }),
		},
		{
			name:  "Food Drop (Food Per Pixel)",
			value: fmt.Sprintf("%d", st.foodcount),
			left:  withProgressiveDuration(func(x int) { st.foodcount -= x }),
			right: withProgressiveDuration(func(x int) { st.foodcount += x }),
		},
		{
			name:  "Stockpile Factor (stockpile vs spawn)",
			value: fmt.Sprintf("%d", st.spawnparam),
			left:  withProgressiveDuration(func(x int) { st.spawnparam -= x }),
			right: withProgressiveDuration(func(x int) { st.spawnparam += x }),
		},
		{
			name:  "Max Ant Population",
			value: fmt.Sprintf("%d", st.maxants),
			left:  withProgressiveDuration(func(x int) { st.maxants -= x }),
			right: withProgressiveDuration(func(x int) { st.maxants += x }),
		},
		{
			name:  "Draw Radius",
			value: fmt.Sprintf("%d", st.drawradius),
			left:  withProgressiveDuration(func(x int) { st.drawradius -= x }),
			right: withProgressiveDuration(func(x int) { st.drawradius += x }),
		},
		{
			name:  "Pheromone Resilience",
			value: fmt.Sprintf("%d", st.fadedivisor),
			left:  withProgressiveDuration(func(x int) { st.fadedivisor -= x }),
			right: withProgressiveDuration(func(x int) { st.fadedivisor += x }),
		},
	}
	return texts
}

func maxWidth(o []opt) int {
	width := 0
	for io := range o {
		if l := len(o[io].name); l > width {
			width = l
		}
	}
	return width
}

func (s *OptScene) Draw(g *Game[GameState], st *GameState, screen *ebiten.Image) {
	var dio ebiten.DrawImageOptions
	screen.DrawImage(s.blank, &dio)

	//padding := maxWidth(s.opts) + 10
	//fmtString := fmt.Sprintf("%%-%ds%%v\n", padding)
	y := optsceneFontSpace
	const step = optsceneFontSpace
	text.Draw(screen, "Up/Down - Change option, Left/Right - Change Value", s.font, 10, y, color.White)
	y += step
	text.Draw(screen, "Jake likes the flashing lines", s.font, 10, y, color.White)

	y += step * 3
	var c1 color.Color = color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff}
	var c2 color.Color = color.RGBA{R: 0xaa, G: 0xaa, B: 0xaa, A: 0xff} //color.White
	var c color.Color = c1
	for oi := range s.opts {
		if oi == s.index {
			c = color.RGBA{R: 0x55, G: 0xFF, B: 0xff, A: 0xFF}
			doLine(0, y+2, st.width/2, y+2, func(x, y int) {
				screen.Set(x, y, c)
			})
		}
		text.Draw(screen, fmt.Sprintf("%s", strings.ToUpper(s.opts[oi].name)), s.font, 10, y, c)
		text.Draw(screen, fmt.Sprintf("%v", strings.ToUpper(s.opts[oi].value)), s.font, 450, y, c)
		y += step
		// if oi%2 == 1 {
		// 	c = c1
		// } else {
		// 	c = c2
		// }
		if c == c1 {
			c = c2
		} else {
			c = c1
		}
	}

	texts := []string{
		"Controls:",
		"Left Click - Brush Stroke",
		"Middle Click - Draw Food",
		"Right Click - Erase",
		"P: Toggle Pheromone Rendering",
		"G: Toggle Green Pheromone Rendering",
		"R: Toggle Red Pheromone Rendering",
		"X: Toggle Parallel Execution",
		"W: Toggle Wall Following",
		"A: Reset Ants to (0,0)",
		"S: Save current grid (persists across restarts)",
		"L: Load the saved grid",
		"C: Clear the grid",
		"F: Fill the grid with wall",
		"M: This menu",
		"Space: Pause",
		"Up/Down: Increase and decrease brush radius",
		"Left/Right: Change the current brush",
	}

	y += step
	for oi := range texts {
		text.Draw(screen, strings.ToUpper(fmt.Sprintf("%s", texts[oi])), s.font, 10, y, color.White)
		y += step
	}
}

func (s *OptScene) RenderBelow() bool {
	return false
}

func (s *OptScene) Update(g *Game[GameState], state *GameState) error {
	max := len(s.opts)

	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) || inpututil.IsKeyJustPressed(ebiten.KeyM) {
		g.PopScene()
		s.as.field.UpdateAll()
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyDown) {
		s.index = (s.index + 1) % max
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyUp) {
		s.index = (s.index - 1)
		if s.index < 0 {
			s.index = max - 1
		}
	}
	if d := inpututil.KeyPressDuration(ebiten.KeyLeft); inpututil.IsKeyJustPressed(ebiten.KeyLeft) || d > 20 {
		s.opts[s.index].left(d)
		s.opts = makeTexts(state)
	}
	if d := inpututil.KeyPressDuration(ebiten.KeyRight); inpututil.IsKeyJustPressed(ebiten.KeyRight) || d > 20 {
		s.opts[s.index].right(d)
		s.opts = makeTexts(state)
	}

	// limts
	if state.spawnparam <= 0 {
		state.spawnparam = 1
		s.opts = makeTexts(state)
	}

	if state.fadedivisor <= 0 {
		state.fadedivisor = 1
	}

	return nil
}
