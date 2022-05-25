package main

import "testing"

func TestDirection(t *testing.T) {
	d := N

	if d2 := d.Left(1); d2 != NW {
		t.Errorf("1 Left of North should be NW, but is %d", d2)
	}

	if d2 := d.Left(2); d2 != W {
		t.Errorf("2 Left of North should be W, but is %d", d2)
	}

	if d2 := d.Left(9); d2 != NW {
		t.Errorf("9 Left of North should be NW, but is %d", d2)
	}

	if d2 := d.Left(12); d2 != S {
		t.Errorf("12 Left of North should be S, but is %d", d2)
	}

	if d2 := d.Right(1); d2 != NE {
		t.Errorf("1 Right of North should be NE, but is %d", d2)
	}

	if d2 := d.Right(2); d2 != E {
		t.Errorf("2 Right of North should be E, but is %d", d2)
	}

	if d2 := d.Right(9); d2 != NE {
		t.Errorf("9 Right of North should be NE, but is %d", d2)
	}

	if d2 := d.Right(12); d2 != S {
		t.Errorf("12 Right of North should be S, but is %d", d2)
	}

}
