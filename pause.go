package main

import (
	"fmt"

	"github.com/flopp/go-findfont"
	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
)

type PauseScene struct {
	pauseText  *sdl.Texture
	pauseSize  sdl.Rect
	quitText   *sdl.Texture
	quitSize   sdl.Rect
	resumeText *sdl.Texture
	resumeSize sdl.Rect
	sel        int
}

func text(r *sdl.Renderer, f *ttf.Font, s string, c sdl.Color) (*sdl.Texture, sdl.Rect, error) {
	surface, err := f.RenderUTF8Blended(s, sdl.Color{0xFF, 0xFF, 0xFF, 0xFF})
	if err != nil {
		return nil, sdl.Rect{}, err
	}
	texture, err := r.CreateTextureFromSurface(surface)
	if err != nil {
		return nil, sdl.Rect{}, err
	}
	w, h, err := f.SizeUTF8(s)
	if err != nil {
		return nil, sdl.Rect{}, err
	}
	rect := sdl.Rect{0, 0, int32(w), int32(h)}
	return texture, rect, nil
}

func NewPauseScene(r *sdl.Renderer) *PauseScene {
	// We should embed whatever font we want.
	fontPath, err := findfont.Find("DejaVu.ttf")
	if err != nil {
		panic(err)
	}
	fmt.Printf("Found 'DejaVu.ttf' in '%s'\n", fontPath)
	bigFont, err := ttf.OpenFont(fontPath, 24)
	if err != nil {
		panic(err)
	}
	defer bigFont.Close()
	smallFont, err := ttf.OpenFont(fontPath, 12)
	if err != nil {
		panic(err)
	}
	defer smallFont.Close()
	pt, pr, err := text(r, bigFont, "Pause", sdl.Color{0xFF, 0xFF, 0xFF, 0xFF})
	if err != nil {
		panic(err)
	}
	qt, qr, err := text(r, smallFont, "Quit", sdl.Color{0xFF, 0xFF, 0xFF, 0xFF})
	if err != nil {
		panic(err)
	}
	rt, rr, err := text(r, smallFont, "Resume", sdl.Color{0xFF, 0xFF, 0xFF, 0xFF})
	if err != nil {
		panic(err)
	}

	return &PauseScene{pt, pr, qt, qr, rt, rr, 0}
}

type number interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~float32 | ~float64
}

func abs[T number](i T) T {
	if i < 0 {
		i = i * -1
	}
	return i
}

func (s *PauseScene) HandleEvent(g *Game[GameState], r *sdl.Renderer, e sdl.Event) error {
	if e.GetType() == sdl.KEYDOWN {
		k := e.(*sdl.KeyboardEvent)
		keyname := sdl.GetKeyName(sdl.GetKeyFromScancode(k.Keysym.Scancode))
		switch keyname {
		case "P":
			fmt.Printf("Popping scene!\n")
			g.PopScene()
		case "Right":
			s.sel = (s.sel + 1) % 2
		case "Left":
			s.sel = abs((s.sel - 1) % 2)
		case "Return":
			if s.sel == 0 {
				return fmt.Errorf("Quit")
			} else if s.sel == 1 {
				g.PopScene()
			}
		default:
			fmt.Printf("KEYNAME: [%s]\n", keyname)
		}
	}
	return nil
}

func (s *PauseScene) Update(g *Game[GameState], r *sdl.Renderer, st *GameState) error {
	return nil
}

func centerxy(container, r *sdl.Rect) {
	r.X = container.X + ((container.W - r.W) / 2)
	r.Y = container.Y + ((container.H - r.H) / 2)
}

func (s *PauseScene) Render(g *Game[GameState], r *sdl.Renderer, st *GameState) error {
	r.SetDrawColor(0xAA, 0xAA, 0xAA, 0xFF)
	x := int32((WIDTH - 600) / 2)
	y := int32((HEIGHT - 300) / 2)
	rect := sdl.Rect{x, y, 600, 300}
	r.FillRect(&rect)

	rect = sdl.Rect{x, y, 600, 150}
	r.DrawRect(&rect)
	centerxy(&rect, &s.pauseSize)
	r.Copy(s.pauseText, nil, &s.pauseSize)

	rect = sdl.Rect{x, y + 150, 300, 150}
	if s.sel == 0 {
		r.SetDrawColor(0xCC, 0x00, 0x00, 0xFF)
	} else {
		r.SetDrawColor(0x66, 0x66, 0x66, 0xFF)
	}
	r.DrawRect(&rect)
	centerxy(&rect, &s.quitSize)
	r.Copy(s.quitText, nil, &s.quitSize)

	rect.X += 300
	if s.sel == 1 {
		r.SetDrawColor(0xCC, 0x00, 0x00, 0xFF)
	} else {
		r.SetDrawColor(0x66, 0x66, 0x66, 0xFF)
	}
	r.DrawRect(&rect)
	centerxy(&rect, &s.resumeSize)
	r.Copy(s.resumeText, nil, &s.resumeSize)
	return nil
}

func (s *PauseScene) RenderBelow() bool {
	return true
}

func (s *PauseScene) Destroy() {}
