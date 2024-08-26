package main

import (
	"reflect"
	"unsafe"

	"github.com/hajimehoshi/ebiten/v2"
)

type DField struct {
	foodpher []int
	homepher []int
	food     []int
	home     []bool
	wall     []bool

	//vals       []T
	renderbuf []uint32
	//valToColor func(*T) uint32

	width, height int

	//tex *sdl.Texture
}

func (f *DField) renderGridspot(as *AntScene, idx int) uint32 {

	if f.wall[idx] {
		//return 0x333333FF
		return 0xFF333333
	} else if f.food[idx] > 0 {
		//return 0x33FF33FF
		return 0xFF33FF33

	} else if f.home[idx] {
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
			vg = uint32(f.foodpher[idx]) >> pherShift
			// if vg > 255 {
			// 	fmt.Printf("FOOD > 255: %d\n", vg)
			// 	vg = 255
			// }
			vg = (vg & 0xFF) << 8
			//vg = vg << 16
		}
		if as.st.renderRed {
			//vr = uint32(g.HomePher / homedivisor)
			vr = uint32(f.homepher[idx]) >> pherShift
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

func NewDField(width, height int) (*DField, error) {
	f := &DField{
		foodpher:  make([]int, width*height),
		homepher:  make([]int, width*height),
		food:      make([]int, width*height),
		home:      make([]bool, width*height),
		wall:      make([]bool, width*height),
		renderbuf: make([]uint32, width*height),
		//valToColor: toColor,
		width:  width,
		height: height,
	}
	return f, nil
}

func (f *DField) Clear(as *AntScene) {
	//f.vals = make([]T, f.width*f.height)
	f.foodpher = make([]int, f.width*f.height)
	f.homepher = make([]int, f.width*f.height)
	f.food = make([]int, f.width*f.height)
	f.home = make([]bool, f.width*f.height)
	f.wall = make([]bool, f.width*f.height)
	f.UpdateAll(as)
}

// func (f *DField) Get(x, y int) *T {
// 	return &f.vals[x+y*f.width]
// }

func (f *DField) Idx(x, y int) int {
	return x + y*f.width
}

func (f *DField) Update(as *AntScene, x, y int) {
	f.renderbuf[x+y*f.width] = f.renderGridspot(as, x+y*f.width)
}

func (f *DField) UpdateAll(as *AntScene) {
	for i := range f.food {
		f.renderbuf[i] = f.renderGridspot(as, i)
	}
}

func (f *DField) Render(r *ebiten.Image) error {
	var bbs []byte
	sliceHeader := (*reflect.SliceHeader)(unsafe.Pointer(&bbs))
	sliceHeader.Cap = int(len(f.renderbuf) * 4)
	sliceHeader.Len = int(len(f.renderbuf) * 4)
	sliceHeader.Data = uintptr(unsafe.Pointer(&f.renderbuf[0]))
	r.ReplacePixels(bbs)
	return nil
}
