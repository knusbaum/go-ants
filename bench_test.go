package main

import (
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

func BenchmarkRender(b *testing.B) {
	im := ebiten.NewImage(WIDTH, HEIGHT)
	f, err := NewDField(WIDTH, HEIGHT)
	if err != nil {
		b.Fatalf("Failed to instantiate field: %v", err)
	}

	as := &AntScene{homelife: 3000 * 10000 * 100}
	as.ants = make([]Ant, 200000)
	gs := NewGameState(WIDTH, HEIGHT)
	as.st = &gs
	as.field = f

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		f.Render(as, im)
	}
}

func BenchmarkGPURender(b *testing.B) {
	im := ebiten.NewImage(WIDTH, HEIGHT)
	f, err := NewDField(WIDTH, HEIGHT)
	if err != nil {
		b.Fatalf("Failed to instantiate field: %v", err)
	}

	as := &AntScene{homelife: 3000 * 10000 * 100}
	as.ants = make([]Ant, 200000)
	gs := NewGameState(WIDTH, HEIGHT)
	as.st = &gs
	as.field = f

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		f.GPURender(as, im)
	}
}
