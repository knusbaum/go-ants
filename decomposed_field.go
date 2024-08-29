package main

import (
	_ "embed"
	"fmt"
	"os"
	"reflect"
	"unsafe"

	"github.com/hajimehoshi/ebiten/v2"
)

const (
	e_none = iota
	e_home
	e_wall
)

type DField struct {
	foodpher []uint32
	homepher []uint32
	food     []uint32
	//home     []bool
	//wall     []bool
	home_wall []uint32

	//vals       []T
	renderbuf []uint32
	//valToColor func(*T) uint32

	width, height int

	//tex *sdl.Texture
	shader *ebiten.Shader
	fader  *ebiten.Shader
	fpim   *ebiten.Image
	hpim   *ebiten.Image
	fim    *ebiten.Image
	hwim   *ebiten.Image
	//frw    sync.RWMutex
	//hrw    sync.RWMutex

	intermediateDST *ebiten.Image
	intermediateSRC *ebiten.Image
}

func (f *DField) GetHomeWall(x, y int) uint8 {
	return uint8(f.home_wall[x+y*f.width])
}

func (f *DField) SetHomeWall(x, y int, v uint8) {
	f.home_wall[x+y*f.width] = uint32(v)
}

func (f *DField) SetFoodPher(x, y, v int) {
	f.foodpher[x+y*f.width] = uint32(v)
	// f.frw.Lock()
	// f.fpim.Set(x, y, color.RGBA{
	// 	R: uint8(v >> 24),
	// 	G: uint8((v >> 16) & 0xFF),
	// 	B: uint8((v >> 8) & 0xFF),
	// 	A: uint8(v & 0xFF)})
	// f.frw.Unlock()
}

func (f *DField) GetFoodPher(x, y int) int {
	return int(f.foodpher[x+y*f.width])
	// f.frw.RLock()
	// c := f.fpim.At(x, y).(color.RGBA)
	// f.frw.RUnlock()
	// return int(uint32(c.A) | (uint32(c.B) << 8) | (uint32(c.G) << 16) | (uint32(c.R) << 24))
}

func (f *DField) SetHomePher(x, y, v int) {
	f.homepher[x+y*f.width] = uint32(v)
	// f.hrw.Lock()
	// f.hpim.Set(x, y, color.RGBA{
	// 	R: uint8(v >> 24),
	// 	G: uint8((v >> 16) & 0xFF),
	// 	B: uint8((v >> 8) & 0xFF),
	// 	A: uint8(v & 0xFF)})
	// f.hrw.Unlock()
}

func (f *DField) GetHomePher(x, y int) int {
	return int(f.homepher[x+y*f.width])
	// f.hrw.RLock()
	// c := f.hpim.At(x, y).(color.RGBA)
	// f.hrw.RUnlock()
	// return int(uint32(c.A) | (uint32(c.B) << 8) | (uint32(c.G) << 16) | (uint32(c.R) << 24))
}

func runShaderOn(st *GameState, s *ebiten.Shader, arr []uint32, width int) {
	src := ebiten.NewImage(width, len(arr)/width)
	dst := ebiten.NewImage(width, len(arr)/width)

	{
		var bbs []byte
		sliceHeader := (*reflect.SliceHeader)(unsafe.Pointer(&bbs))
		sliceHeader.Cap = int(len(arr) * 4)
		sliceHeader.Len = int(len(arr) * 4)
		sliceHeader.Data = uintptr(unsafe.Pointer(&arr[0]))
		src.ReplacePixels(bbs)
	}

	var vertices [4]ebiten.Vertex
	// map the vertices to the target image
	bounds := dst.Bounds()
	vertices[0].DstX = float32(bounds.Min.X) // top-left
	vertices[0].DstY = float32(bounds.Min.Y) // top-left
	vertices[1].DstX = float32(bounds.Max.X) // top-right
	vertices[1].DstY = float32(bounds.Min.Y) // top-right
	vertices[2].DstX = float32(bounds.Min.X) // bottom-left
	vertices[2].DstY = float32(bounds.Max.Y) // bottom-left
	vertices[3].DstX = float32(bounds.Max.X) // bottom-right
	vertices[3].DstY = float32(bounds.Max.Y) // bottom-right

	// set the source image sampling coordinates
	srcBounds := src.Bounds()
	vertices[0].SrcX = float32(srcBounds.Min.X) // top-left
	vertices[0].SrcY = float32(srcBounds.Min.Y) // top-left
	vertices[1].SrcX = float32(srcBounds.Max.X) // top-right
	vertices[1].SrcY = float32(srcBounds.Min.Y) // top-right
	vertices[2].SrcX = float32(srcBounds.Min.X) // bottom-left
	vertices[2].SrcY = float32(srcBounds.Max.Y) // bottom-left
	vertices[3].SrcX = float32(srcBounds.Max.X) // bottom-right
	vertices[3].SrcY = float32(srcBounds.Max.Y) // bottom-right

	// triangle shader options
	var shaderOpts ebiten.DrawTrianglesShaderOptions
	shaderOpts.Uniforms = make(map[string]any)
	shaderOpts.Uniforms["FadeDivisor"] = st.fadedivisor
	shaderOpts.Images[0] = src

	// draw shader
	indices := []uint16{0, 1, 2, 2, 1, 3} // map vertices to triangles
	dst.DrawTrianglesShader(vertices[:], indices, s, &shaderOpts)

	{
		var bbs []byte
		sliceHeader := (*reflect.SliceHeader)(unsafe.Pointer(&bbs))
		sliceHeader.Cap = int(len(arr) * 4)
		sliceHeader.Len = int(len(arr) * 4)
		sliceHeader.Data = uintptr(unsafe.Pointer(&arr[0]))
		//f.fpim.ReplacePixels(bbs)
		dst.ReadPixels(bbs)
	}
	// var bbs = make([]byte, 4*len(arr))
	// dst.ReadPixels(bbs)
	// return bbs
	//return nil
}

func (f *DField) renderGridspot(as *AntScene, idx int) uint32 {
	if f.home_wall[idx] == e_wall {
		//return 0x333333FF
		return 0xFF333333
	} else if f.food[idx] > 0 {
		//return 0x33FF33FF
		return 0xFF33FF33

	} else if f.home_wall[idx] == e_home {
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
	//return 0
}

//go:embed shader.kage
var shaderProgram []byte

//go:embed fader.kage
var faderProgram []byte

func NewDField(width, height int) (*DField, error) {
	shader, err := ebiten.NewShader(shaderProgram)
	if err != nil {
		fmt.Printf("Fatal, failed to compile shader: %v\n", err)
		os.Exit(1)
	}

	fader, err := ebiten.NewShader(faderProgram)
	if err != nil {
		fmt.Printf("Fatal, failed to compile shader: %v\n", err)
		os.Exit(1)
	}

	f := &DField{
		foodpher: make([]uint32, width*height),
		homepher: make([]uint32, width*height),
		food:     make([]uint32, width*height),
		//home:      make([]bool, width*height),
		//wall:      make([]bool, width*height),
		home_wall: make([]uint32, width*height),
		renderbuf: make([]uint32, width*height),
		//valToColor: toColor,
		width:           width,
		height:          height,
		shader:          shader,
		fader:           fader,
		fpim:            ebiten.NewImage(WIDTH, HEIGHT),
		hpim:            ebiten.NewImage(WIDTH, HEIGHT),
		fim:             ebiten.NewImage(WIDTH, HEIGHT),
		hwim:            ebiten.NewImage(WIDTH, HEIGHT),
		intermediateSRC: ebiten.NewImage(WIDTH, HEIGHT),
	}
	return f, nil
}

func (f *DField) Clear(as *AntScene) {

	f.foodpher = make([]uint32, f.width*f.height)
	f.homepher = make([]uint32, f.width*f.height)
	f.food = make([]uint32, f.width*f.height)
	//f.home = make([]bool, f.width*f.height)
	//f.wall = make([]bool, f.width*f.height)
	f.home_wall = make([]uint32, f.width*f.height)
	f.UpdateAll(as)
}

// func (f *DField) Get(x, y int) *T {
// 	return &f.vals[x+y*f.width]
// }

func (f *DField) Idx(x, y int) int {
	return x + y*f.width
}

func (f *DField) Update(as *AntScene, x, y int) {
	//	f.renderbuf[x+y*f.width] = f.renderGridspot(as, x+y*f.width)
}

func (f *DField) UpdateAll(as *AntScene) {
	// for i := range f.food {
	// 	f.renderbuf[i] = f.renderGridspot(as, i)
	// }
}

func (f *DField) Render(as *AntScene, r *ebiten.Image) error {
	for i := range f.food {
		f.renderbuf[i] = f.renderGridspot(as, i)
	}

	var bbs []byte
	sliceHeader := (*reflect.SliceHeader)(unsafe.Pointer(&bbs))
	sliceHeader.Cap = int(len(f.renderbuf) * 4)
	sliceHeader.Len = int(len(f.renderbuf) * 4)
	sliceHeader.Data = uintptr(unsafe.Pointer(&f.renderbuf[0]))
	r.ReplacePixels(bbs)
	return nil
}

func (f *DField) fade(st *GameState, data []uint32, dsti *ebiten.Image) {
	//start := f.homepher[200+200*f.width]
	src := f.intermediateSRC
	dst := dsti
	dst.Clear()
	{

		var bbs []byte
		sliceHeader := (*reflect.SliceHeader)(unsafe.Pointer(&bbs))
		sliceHeader.Cap = int(len(data) * 4)
		sliceHeader.Len = int(len(data) * 4)
		sliceHeader.Data = uintptr(unsafe.Pointer(&data[0]))
		src.WritePixels(bbs)
	}

	var vertices [4]ebiten.Vertex
	// map the vertices to the target image
	bounds := dst.Bounds()
	vertices[0].DstX = float32(bounds.Min.X) // top-left
	vertices[0].DstY = float32(bounds.Min.Y) // top-left
	vertices[1].DstX = float32(bounds.Max.X) // top-right
	vertices[1].DstY = float32(bounds.Min.Y) // top-right
	vertices[2].DstX = float32(bounds.Min.X) // bottom-left
	vertices[2].DstY = float32(bounds.Max.Y) // bottom-left
	vertices[3].DstX = float32(bounds.Max.X) // bottom-right
	vertices[3].DstY = float32(bounds.Max.Y) // bottom-right

	// set the source image sampling coordinates
	srcBounds := src.Bounds()
	vertices[0].SrcX = float32(srcBounds.Min.X) // top-left
	vertices[0].SrcY = float32(srcBounds.Min.Y) // top-left
	vertices[1].SrcX = float32(srcBounds.Max.X) // top-right
	vertices[1].SrcY = float32(srcBounds.Min.Y) // top-right
	vertices[2].SrcX = float32(srcBounds.Min.X) // bottom-left
	vertices[2].SrcY = float32(srcBounds.Max.Y) // bottom-left
	vertices[3].SrcX = float32(srcBounds.Max.X) // bottom-right
	vertices[3].SrcY = float32(srcBounds.Max.Y) // bottom-right

	// triangle shader options
	var shaderOpts ebiten.DrawTrianglesShaderOptions
	shaderOpts.Uniforms = make(map[string]any)
	shaderOpts.Uniforms["FadeDivisor"] = st.fadedivisor
	shaderOpts.Images[0] = src

	// draw shader
	indices := []uint16{0, 1, 2, 2, 1, 3} // map vertices to triangles
	dst.DrawTrianglesShader(vertices[:], indices, f.fader, &shaderOpts)
	{
		var bbs []byte
		sliceHeader := (*reflect.SliceHeader)(unsafe.Pointer(&bbs))
		sliceHeader.Cap = int(len(data) * 4)
		sliceHeader.Len = int(len(data) * 4)
		sliceHeader.Data = uintptr(unsafe.Pointer(&data[0]))
		dst.ReadPixels(bbs)
	}

	//fmt.Printf("homepher: %X(%d) -> %X(%d)\n", start, start, f.homepher[200+200*f.width], f.homepher[200+200*f.width])
}

func (f *DField) UpdatePher(st *GameState) {
	f.fade(st, f.homepher, f.hpim)
	f.fade(st, f.foodpher, f.fpim)
	// start := f.homepher[200+200*f.width]
	// //runShaderOn(st, f.fader, f.homepher, f.width)
	// //src := ebiten.NewImage(f.width, f.height)
	// //dst := ebiten.NewImage(f.width, f.height)
	// src := f.intermediateSRC
	// dst := f.
	// 	//src.Clear()
	// 	dst.Clear()
	// {

	// 	var bbs []byte
	// 	sliceHeader := (*reflect.SliceHeader)(unsafe.Pointer(&bbs))
	// 	sliceHeader.Cap = int(len(f.homepher) * 4)
	// 	sliceHeader.Len = int(len(f.homepher) * 4)
	// 	sliceHeader.Data = uintptr(unsafe.Pointer(&f.homepher[0]))
	// 	//f.hpim.ReplacePixels(bbs)
	// 	src.WritePixels(bbs)
	// }
	// // {
	// // 	var bbs []byte
	// // 	sliceHeader := (*reflect.SliceHeader)(unsafe.Pointer(&bbs))
	// // 	sliceHeader.Cap = int(len(f.homepher) * 4)
	// // 	sliceHeader.Len = int(len(f.homepher) * 4)
	// // 	sliceHeader.Data = uintptr(unsafe.Pointer(&f.homepher[0]))
	// // 	f.hpim.ReplacePixels(bbs)
	// // }

	// var vertices [4]ebiten.Vertex
	// // map the vertices to the target image
	// bounds := dst.Bounds()
	// vertices[0].DstX = float32(bounds.Min.X) // top-left
	// vertices[0].DstY = float32(bounds.Min.Y) // top-left
	// vertices[1].DstX = float32(bounds.Max.X) // top-right
	// vertices[1].DstY = float32(bounds.Min.Y) // top-right
	// vertices[2].DstX = float32(bounds.Min.X) // bottom-left
	// vertices[2].DstY = float32(bounds.Max.Y) // bottom-left
	// vertices[3].DstX = float32(bounds.Max.X) // bottom-right
	// vertices[3].DstY = float32(bounds.Max.Y) // bottom-right

	// // set the source image sampling coordinates
	// srcBounds := src.Bounds()
	// vertices[0].SrcX = float32(srcBounds.Min.X) // top-left
	// vertices[0].SrcY = float32(srcBounds.Min.Y) // top-left
	// vertices[1].SrcX = float32(srcBounds.Max.X) // top-right
	// vertices[1].SrcY = float32(srcBounds.Min.Y) // top-right
	// vertices[2].SrcX = float32(srcBounds.Min.X) // bottom-left
	// vertices[2].SrcY = float32(srcBounds.Max.Y) // bottom-left
	// vertices[3].SrcX = float32(srcBounds.Max.X) // bottom-right
	// vertices[3].SrcY = float32(srcBounds.Max.Y) // bottom-right

	// // triangle shader options
	// var shaderOpts ebiten.DrawTrianglesShaderOptions
	// shaderOpts.Uniforms = make(map[string]any)
	// shaderOpts.Uniforms["FadeDivisor"] = st.fadedivisor
	// shaderOpts.Images[0] = src

	// // draw shader
	// indices := []uint16{0, 1, 2, 2, 1, 3} // map vertices to triangles
	// dst.DrawTrianglesShader(vertices[:], indices, f.fader, &shaderOpts)

	// {
	// 	var bbs []byte
	// 	sliceHeader := (*reflect.SliceHeader)(unsafe.Pointer(&bbs))
	// 	sliceHeader.Cap = int(len(f.homepher) * 4)
	// 	sliceHeader.Len = int(len(f.homepher) * 4)
	// 	sliceHeader.Data = uintptr(unsafe.Pointer(&f.homepher[0]))
	// 	dst.ReadPixels(bbs)
	// }

	// fmt.Printf("homepher: %X(%d) -> %X(%d)\n", start, start, f.homepher[200+200*f.width], f.homepher[200+200*f.width])
	// //}

	// //f.hpim.DrawTrianglesShader(vertices[:], indices, f.fader, &shaderOpts)
}

func (f *DField) GPURender(as *AntScene, r *ebiten.Image) error {
	// We don't need to copy fpim and hpim because
	// they are updated by the UpdatePher method.
	{
		var bbs []byte
		sliceHeader := (*reflect.SliceHeader)(unsafe.Pointer(&bbs))
		sliceHeader.Cap = int(len(f.foodpher) * 4)
		sliceHeader.Len = int(len(f.foodpher) * 4)
		sliceHeader.Data = uintptr(unsafe.Pointer(&f.foodpher[0]))
		f.fpim.ReplacePixels(bbs)
	}
	{
		var bbs []byte
		sliceHeader := (*reflect.SliceHeader)(unsafe.Pointer(&bbs))
		sliceHeader.Cap = int(len(f.homepher) * 4)
		sliceHeader.Len = int(len(f.homepher) * 4)
		sliceHeader.Data = uintptr(unsafe.Pointer(&f.homepher[0]))
		f.hpim.ReplacePixels(bbs)
	}
	{
		var bbs []byte
		sliceHeader := (*reflect.SliceHeader)(unsafe.Pointer(&bbs))
		sliceHeader.Cap = int(len(f.food) * 4)
		sliceHeader.Len = int(len(f.food) * 4)
		sliceHeader.Data = uintptr(unsafe.Pointer(&f.food[0]))
		f.fim.ReplacePixels(bbs)
	}
	{
		var bbs []byte
		sliceHeader := (*reflect.SliceHeader)(unsafe.Pointer(&bbs))
		sliceHeader.Cap = int(len(f.home_wall) * 4)
		sliceHeader.Len = int(len(f.home_wall) * 4)
		sliceHeader.Data = uintptr(unsafe.Pointer(&f.home_wall[0]))
		f.hwim.ReplacePixels(bbs)
	}

	var vertices [4]ebiten.Vertex
	// map the vertices to the target image
	bounds := r.Bounds()
	vertices[0].DstX = float32(bounds.Min.X) // top-left
	vertices[0].DstY = float32(bounds.Min.Y) // top-left
	vertices[1].DstX = float32(bounds.Max.X) // top-right
	vertices[1].DstY = float32(bounds.Min.Y) // top-right
	vertices[2].DstX = float32(bounds.Min.X) // bottom-left
	vertices[2].DstY = float32(bounds.Max.Y) // bottom-left
	vertices[3].DstX = float32(bounds.Max.X) // bottom-right
	vertices[3].DstY = float32(bounds.Max.Y) // bottom-right

	// set the source image sampling coordinates
	srcBounds := f.fpim.Bounds()
	vertices[0].SrcX = float32(srcBounds.Min.X) // top-left
	vertices[0].SrcY = float32(srcBounds.Min.Y) // top-left
	vertices[1].SrcX = float32(srcBounds.Max.X) // top-right
	vertices[1].SrcY = float32(srcBounds.Min.Y) // top-right
	vertices[2].SrcX = float32(srcBounds.Min.X) // bottom-left
	vertices[2].SrcY = float32(srcBounds.Max.Y) // bottom-left
	vertices[3].SrcX = float32(srcBounds.Max.X) // bottom-right
	vertices[3].SrcY = float32(srcBounds.Max.Y) // bottom-right

	// triangle shader options
	var shaderOpts ebiten.DrawTrianglesShaderOptions
	shaderOpts.Images[0] = f.fpim
	shaderOpts.Images[1] = f.hpim
	shaderOpts.Images[2] = f.fim
	shaderOpts.Images[3] = f.hwim

	// draw shader
	indices := []uint16{0, 1, 2, 2, 1, 3} // map vertices to triangles
	r.DrawTrianglesShader(vertices[:], indices, f.shader, &shaderOpts)
	return nil
}
