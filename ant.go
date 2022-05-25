package main

import (
	"math/rand"

	"github.com/veandco/go-sdl2/sdl"
)

type direction int

const (
	N direction = iota
	NE
	E
	SE
	S
	SW
	W
	NW
	END // Used for modular arithmetic
)

func (d direction) String() string {
	switch d {
	case N:
		return "N"
	case NE:
		return "NE"
	case E:
		return "E"
	case SE:
		return "SE"
	case S:
		return "S"
	case SW:
		return "SW"
	case W:
		return "W"
	case NW:
		return "NW"
	}
	return "UNKNOWN"
}

type point struct {
	x int
	y int
}

func (p point) PointAt(d direction) point {
	np := point{p.x, p.y}
	switch d {
	case N:
		np.y -= 1
	case NE:
		np.y -= 1
		np.x += 1
	case E:
		np.x += 1
	case SE:
		np.y += 1
		np.x += 1
	case S:
		np.y += 1
	case SW:
		np.y += 1
		np.x -= 1
	case W:
		np.x -= 1
	case NW:
		np.y -= 1
		np.x -= 1
	}
	return np
}

func (p point) Within(x, y, w, h int) bool {
	return p.x >= x && p.x < x+w && p.y >= y && p.y < y+h
}

type Ant struct {
	pos    point
	dir    direction
	food   int
	marker int
}

func (a *Ant) GridAt(an *AntScene, d direction) (point, bool) {
	np := a.pos.PointAt(d)
	if np.Within(0, 0, WIDTH, HEIGHT) {
		return an.grid[np.x][np.y], true
	}
	return point{}, false
}

func (d direction) Left(n int) direction {
	d -= direction(n)
	d = d % END
	if d < 0 {
		d += END
	}
	return d
}

func (d direction) Right(n int) direction {
	d += direction(n)
	d = d % END
	return d
}

func (a *Ant) OctantRect(d direction, size int) sdl.Rect {
	var (
		start point
		end   point
	)
	switch d {
	case N:
		start.x = a.pos.x - (size / 2)
		start.y = a.pos.y - size
		end.x = a.pos.x + (size / 2)
		end.y = a.pos.y
	case NE:
		start.x = a.pos.x
		start.y = a.pos.y - size
		end.x = a.pos.x + size
		end.y = a.pos.y
	case E:
		start.x = a.pos.x
		start.y = a.pos.y - (size / 2)
		end.x = a.pos.x + size
		end.y = a.pos.y + (size / 2)
	case SE:
		start.x = a.pos.x
		start.y = a.pos.y
		end.x = a.pos.x + size
		end.y = a.pos.y + size
	case S:
		start.x = a.pos.x - (size / 2)
		start.y = a.pos.y
		end.x = a.pos.x + (size / 2)
		end.y = a.pos.y + size
	case SW:
		start.x = a.pos.x - size
		start.y = a.pos.y
		end.x = a.pos.x
		end.y = a.pos.y + size
	case W:
		start.x = a.pos.x - size
		start.y = a.pos.y - (size / 2)
		end.x = a.pos.x
		end.y = a.pos.y + (size / 2)
	case NW:
		start.x = a.pos.x - size
		start.y = a.pos.y - size
		end.x = a.pos.x
		end.y = a.pos.y
	}
	return sdl.Rect{int32(start.x), int32(start.y), int32(end.x - start.x), int32(end.y - start.y)}
}

func (a *Ant) SumOctant(an *AntScene, d direction, size int) point {
	var (
		start point
		end   point
	)
	switch d {
	case N:
		start.x = a.pos.x - (size / 2)
		start.y = a.pos.y - size
		end.x = a.pos.x + (size / 2)
		end.y = a.pos.y
	case NE:
		start.x = a.pos.x
		start.y = a.pos.y - size
		end.x = a.pos.x + size
		end.y = a.pos.y
	case E:
		start.x = a.pos.x
		start.y = a.pos.y - (size / 2)
		end.x = a.pos.x + size
		end.y = a.pos.y + (size / 2)
	case SE:
		start.x = a.pos.x
		start.y = a.pos.y
		end.x = a.pos.x + size
		end.y = a.pos.y + size
	case S:
		start.x = a.pos.x - (size / 2)
		start.y = a.pos.y
		end.x = a.pos.x + (size / 2)
		end.y = a.pos.y + size
	case SW:
		start.x = a.pos.x - size
		start.y = a.pos.y
		end.x = a.pos.x
		end.y = a.pos.y + size
	case W:
		start.x = a.pos.x - size
		start.y = a.pos.y - (size / 2)
		end.x = a.pos.x
		end.y = a.pos.y + (size / 2)
	case NW:
		start.x = a.pos.x - size
		start.y = a.pos.y - size
		end.x = a.pos.x
		end.y = a.pos.y
	}
	var pt point
	for x := start.x; x < end.x; x++ {
		for y := start.y; y < end.y; y++ {
			p := point{x, y}
			if p.Within(0, 0, WIDTH, HEIGHT) {
				pt.x += an.grid[x][y].x
				pt.y += an.grid[x][y].y
			}
		}
	}
	return pt
}

func (a *Ant) Move(an *AntScene) {
	straight := a.SumOctant(an, a.dir, 50)
	left := a.SumOctant(an, a.dir.Left(1), 50)
	right := a.SumOctant(an, a.dir.Right(1), 50)

	if a.food > 0 {
		if right.y > straight.y && right.y > left.y {
			a.dir = a.dir.Right(1)
		} else if left.y > straight.y && left.y > right.y {
			a.dir = a.dir.Left(1)
		}
	} else {
		if right.x > straight.x && right.x > left.x {
			a.dir = a.dir.Right(1)
		} else if left.x > straight.x && left.x > right.x {
			a.dir = a.dir.Left(1)
		}
	}

	n := rand.Intn(40)
	if n == 0 {
		a.dir = a.dir.Left(1)
	} else if n == 1 {
		a.dir = a.dir.Right(1)
	}

	if _, ok := a.GridAt(an, a.dir); !ok {
		a.dir = a.dir.Right(4)
		_, ok := a.GridAt(an, a.dir)
		for ; !ok; _, ok = a.GridAt(an, a.dir) {
			a.dir = a.dir.Right(1)
		}
	}

	a.pos = a.pos.PointAt(a.dir)

	if a.marker > 0 {
		a.marker -= 5
	}
}
