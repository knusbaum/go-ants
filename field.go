package main

import (
	"reflect"
	"unsafe"

	"github.com/veandco/go-sdl2/sdl"
)

type Field[T any] struct {
	vals       []T
	renderbuf  []uint32
	valToColor func(T) uint32

	width, height int

	tex *sdl.Texture
}

func NewField[T any](r *sdl.Renderer, width, height int, toColor func(T) uint32) (*Field[T], error) {
	t, err := r.CreateTexture(uint32(sdl.PIXELFORMAT_RGBA8888), sdl.TEXTUREACCESS_STREAMING, int32(width), int32(height))
	if err != nil {
		return nil, err
	}
	return &Field[T]{
		vals:       make([]T, width*height),
		renderbuf:  make([]uint32, width*height),
		valToColor: toColor,
		width:      width,
		height:     height,
		tex:        t,
	}, nil

}

func (f *Field[T]) Clear() {
	// TODO
}

//func (f *Field[T]) Set(x, y int, e T) {
//	//fmt.Printf("SETTING %d %d to %x\n", x, y, f.valToColor(e))
//	f.vals[x+y*f.width] = e
//	f.renderbuf[x+y*f.width] = f.valToColor(e)
//}

func (f *Field[T]) Get(x, y int) *T {
	//fmt.Printf("GETTING %d %d\n", x, y)
	return &f.vals[x+y*f.width]
}

func (f *Field[T]) Update(x, y int) {
	f.renderbuf[x+y*f.width] = f.valToColor(f.vals[x+y*f.width])
}

func (f *Field[T]) UpdateAll() {
	for i := range f.vals {
		f.renderbuf[i] = f.valToColor(f.vals[i])
	}
}

func (f *Field[T]) Render(r *sdl.Renderer) error {
	bsb, _, err := f.tex.Lock(nil)
	if err != nil {
		return err
	}
	defer f.tex.Unlock()
	var bs []uint32
	sliceHeader := (*reflect.SliceHeader)(unsafe.Pointer(&bs))
	sliceHeader.Cap = int(len(bsb) / 4)
	sliceHeader.Len = int(len(bsb) / 4)
	sliceHeader.Data = uintptr(unsafe.Pointer(&bsb[0]))
	copy(bs, f.renderbuf)
	r.Copy(f.tex, nil, nil)
	return nil
}
