package main

import (
	"reflect"
	"unsafe"

	"github.com/veandco/go-sdl2/sdl"
)

type Field[T any] struct {
	vals       []T
	renderbuf  []uint32
	valToColor func(*T) uint32

	width, height int

	tex *sdl.Texture
}

func NewField[T any](r *sdl.Renderer, width, height int, toColor func(*T) uint32) (*Field[T], error) {
	t, err := r.CreateTexture(uint32(sdl.PIXELFORMAT_RGBA8888), sdl.TEXTUREACCESS_STREAMING, int32(width), int32(height))
	if err != nil {
		return nil, err
	}
	f := &Field[T]{
		vals:       make([]T, width*height),
		renderbuf:  make([]uint32, width*height),
		valToColor: toColor,
		width:      width,
		height:     height,
		tex:        t,
	}
	f.Render(r)
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

func (f *Field[T]) Render(r *sdl.Renderer) error {
	f.tex.Unlock()
	r.Copy(f.tex, nil, nil)
	bsb, _, err := f.tex.Lock(nil)
	if err != nil {
		return err
	}

	var bs []uint32
	sliceHeader := (*reflect.SliceHeader)(unsafe.Pointer(&bs))
	sliceHeader.Cap = int(len(bsb) / 4)
	sliceHeader.Len = int(len(bsb) / 4)
	sliceHeader.Data = uintptr(unsafe.Pointer(&bsb[0]))
	f.renderbuf = bs
	return nil

	// bsb, _, err := f.tex.Lock(nil)
	// if err != nil {
	// 	return err
	// }

	// var bs []uint32
	// sliceHeader := (*reflect.SliceHeader)(unsafe.Pointer(&bs))
	// sliceHeader.Cap = int(len(bsb) / 4)
	// sliceHeader.Len = int(len(bsb) / 4)
	// sliceHeader.Data = uintptr(unsafe.Pointer(&bsb[0]))
	// f.tex.Unlock()
	// copy(bs, f.renderbuf)
	// r.Copy(f.tex, nil, nil)
	// return nil
}
