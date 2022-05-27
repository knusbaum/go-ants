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

func (a *Ant) GridAt(an *AntScene, d direction) (gridspot, bool) {
	np := a.pos.PointAt(d)
	if np.Within(0, 0, WIDTH, HEIGHT) {
		return an.grid[np.x][np.y], true
	}
	return gridspot{}, false
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
	if d < 0 {
		d += END
	}
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

func (a *Ant) SumOctant(an *AntScene, d direction, size int) gridspot {
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
	var pt gridspot
	for x := start.x; x < end.x; x++ {
		for y := start.y; y < end.y; y++ {
			p := point{x, y}
			if p.Within(0, 0, WIDTH, HEIGHT) && !an.grid[x][y].Wall {
				pt.FoodPher += an.grid[x][y].FoodPher + an.grid[x][y].Food*1000000 // - (an.grid[x][y].homePher / 4)
				pt.HomePher += an.grid[x][y].HomePher                              // - (an.grid[x][y].foodPher / 4)
				if an.grid[x][y].Home {
					pt.HomePher += 1000000
				}
			}
		}
	}
	if an.pherOverload {
		if pt.FoodPher > pheromoneMax*an.pherOverloadFactor {
			pt.FoodPher = pheromoneMax * an.pherOverloadFactor
		}
		if pt.HomePher > pheromoneMax*an.pherOverloadFactor {
			pt.HomePher = pheromoneMax * an.pherOverloadFactor
		}
	}
	return pt
}

func (a *Ant) Move(an *AntScene) {
	if n := rand.Intn(10); n == 0 {
		straight := a.SumOctant(an, a.dir, 50)
		left := a.SumOctant(an, a.dir.Left(1), 50)
		right := a.SumOctant(an, a.dir.Right(1), 50)

		if a.food > 0 {
			if right.HomePher > straight.HomePher && right.HomePher > left.HomePher {
				a.dir = a.dir.Right(1)
			} else if left.HomePher > straight.HomePher && left.HomePher > right.HomePher {
				a.dir = a.dir.Left(1)
			}
		} else {
			if right.FoodPher > straight.FoodPher && right.FoodPher > left.FoodPher {
				a.dir = a.dir.Right(1)
			} else if left.FoodPher > straight.FoodPher && left.FoodPher > right.FoodPher {
				a.dir = a.dir.Left(1)
			}
		}

		n := rand.Intn(10)
		if n == 0 {
			a.dir = a.dir.Left(1)
		} else if n == 1 {
			a.dir = a.dir.Right(1)
		}
	}

	if g, ok := a.GridAt(an, a.dir); !ok || g.Wall {
		a.dir = a.dir.Right((rand.Intn(3) - 1) * 2)
		g, ok := a.GridAt(an, a.dir)
		i := 0
		for ; !ok || g.Wall; g, ok = a.GridAt(an, a.dir) {
			//fmt.Printf("SPIN\n")
			a.dir = a.dir.Right(1)
			i++
			if i >= 8 {
				a.pos.x = 1
				a.pos.y = 1
				return
			}
		}
	}

	a.pos = a.pos.PointAt(a.dir)

	// 	if a.marker > 0 {
	// 		a.marker -= 8
	// 	}
}
