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
		//return an.grid[np.x][np.y], true
		return *an.field.Get(np.x, np.y), true
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

func (a *Ant) Line(an *AntScene, d direction, size int) gridspot {
	var pt gridspot
	addspot := func(g *gridspot) bool {
		if g.Wall {
			return false
		}
		pt.FoodPher += g.FoodPher + g.Food*100000
		pt.HomePher += g.HomePher
		if g.Home {
			pt.HomePher += 100000
		}
		return true
	}
	switch d {
	case N:
		for y := a.pos.y; y > a.pos.y-size && y > 0; y-- {
			if !addspot(an.field.Get(a.pos.x, y)) {
				break
			}
		}
	case NE:
		for i := 0; i < size; i++ {
			y := a.pos.y - i
			x := a.pos.x + i
			if y < 0 || x >= an.field.height {
				break
			}
			if !addspot(an.field.Get(x, y)) {
				break
			}
		}
	case E:
		for x := a.pos.x; x < a.pos.x+size && x < an.field.height; x++ {
			if !addspot(an.field.Get(x, a.pos.y)) {
				break
			}
		}
	case SE:
		for i := 0; i < size; i++ {
			y := a.pos.y + i
			x := a.pos.x + i
			if y >= an.field.height || x >= an.field.width {
				break
			}
			if !addspot(an.field.Get(x, y)) {
				break
			}
		}
	case S:
		for y := a.pos.y; y < a.pos.y+size && y < an.field.height; y++ {
			if !addspot(an.field.Get(a.pos.x, y)) {
				break
			}
		}
	case SW:
		for i := 0; i < size; i++ {
			y := a.pos.y + i
			x := a.pos.x - i
			if y >= an.field.height || x < 0 {
				break
			}
			if !addspot(an.field.Get(x, y)) {
				break
			}
		}
	case W:
		for x := a.pos.x; x > a.pos.x-size && x > 0; x-- {
			if !addspot(an.field.Get(x, a.pos.y)) {
				break
			}
		}
	case NW:
		for i := 0; i < size; i++ {
			y := a.pos.y - i
			x := a.pos.x - i
			if y < 0 || x < 0 {
				break
			}
			if !addspot(an.field.Get(x, y)) {
				break
			}
		}
	default:
		panic("NO SUCH DIRECTION")
	}
	return pt
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
	for y := start.y; y < end.y; y++ {
		for x := start.x; x < end.x; x++ {

			p := point{x, y}
			//if p.Within(0, 0, WIDTH, HEIGHT) && !an.grid[x][y].Wall
			if p.Within(0, 0, WIDTH, HEIGHT) {
				//spot := an.field.Get(x, y)
				if !an.field.vals[x+y*an.field.width].Wall {
					pt.FoodPher += an.field.vals[x+y*an.field.width].FoodPher + an.field.vals[x+y*an.field.width].Food*100000 // - (an.grid[x][y].homePher / 4)
					pt.HomePher += an.field.vals[x+y*an.field.width].HomePher                                                 // - (an.grid[x][y].foodPher / 4)
					if an.field.vals[x+y*an.field.width].Home {
						pt.HomePher += 100000
					}
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
		// straight := a.SumOctant(an, a.dir, 50)
		// left := a.SumOctant(an, a.dir.Left(1), 50)
		// right := a.SumOctant(an, a.dir.Right(1), 50)

		straight := a.Line(an, a.dir, 50)
		left := a.Line(an, a.dir.Left(1), 50)
		lleft := a.Line(an, a.dir.Left(2), 50)
		right := a.Line(an, a.dir.Right(1), 50)
		rright := a.Line(an, a.dir.Right(2), 50)

		// Directions include weighted values of their left and right directions
		straight.FoodPher += left.FoodPher/2 + right.FoodPher/2
		straight.HomePher += left.HomePher/2 + right.HomePher/2
		left.FoodPher += lleft.FoodPher/2 + straight.FoodPher/2
		left.HomePher += lleft.HomePher/2 + straight.HomePher/2
		right.FoodPher += rright.FoodPher/2 + straight.FoodPher/2
		right.HomePher += rright.HomePher/2 + straight.HomePher/2

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
