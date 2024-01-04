package main

import (
	"reflect"
	"unsafe"

	"github.com/hajimehoshi/ebiten/v2"
)

type Field[T any] struct {
	vals       []T
	renderbuf  []uint32
	valToColor func(*T) uint32

	width, height int

	//tex *sdl.Texture
}

func NewField[T any](width, height int, toColor func(*T) uint32) (*Field[T], error) {
	// t, err := r.CreateTexture(uint32(sdl.PIXELFORMAT_RGBA8888), sdl.TEXTUREACCESS_STREAMING, int32(width), int32(height))
	// if err != nil {
	// 	return nil, err
	// }
	f := &Field[T]{
		vals:       make([]T, width*height),
		renderbuf:  make([]uint32, width*height),
		valToColor: toColor,
		width:      width,
		height:     height,
		//tex:        t,
	}
	//f.Render(r)
	return f, nil

}

func (f *Field[T]) Clear() {
	f.vals = make([]T, f.width*f.height)
	f.UpdateAll()
}

func (f *Field[T]) Get(x, y int) *T {
	return &f.vals[x+y*f.width]
}

func (f *Field[T]) Update(x, y int) {
	f.renderbuf[x+y*f.width] = f.valToColor(&f.vals[x+y*f.width])
}

func (f *Field[T]) UpdateAll() {
	for i := range f.vals {
		f.renderbuf[i] = f.valToColor(&f.vals[i])
	}
}

func (f *Field[T]) Render(r *ebiten.Image) error {

	var bbs []byte
	sliceHeader := (*reflect.SliceHeader)(unsafe.Pointer(&bbs))
	sliceHeader.Cap = int(len(f.renderbuf) * 4)
	sliceHeader.Len = int(len(f.renderbuf) * 4)
	sliceHeader.Data = uintptr(unsafe.Pointer(&f.renderbuf[0]))
	r.ReplacePixels(bbs)
	return nil
}
