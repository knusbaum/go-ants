package main

import (
	"fmt"
	"os"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

func TestGPGPU(t *testing.T) {

	fader, err := ebiten.NewShader(shaderProgram)
	if err != nil {
		fmt.Printf("Fatal, failed to compile shader: %v\n", err)
		os.Exit(1)
	}

	v := make([]uint32, 4)
	v[0] = 0x11223344
	v[1] = 0x55667788
	v[2] = 0x99aabbcc
	v[3] = 0xddeeff00

	runShaderOn(fader, v, 2)

	if v[0] != 0x11223344 {
		t.Fatalf("EXPECTED 0x11223344 but got %8X", v[0])
	}

}
