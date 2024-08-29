package main

import (
	"fmt"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
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
	pos point
	//idx    int // index based on position == pos.x + pos.y*height
	dir    direction
	food   int
	marker int
	life   int
	tex    *ebiten.Image
}

func (a *Ant) WallAt(an *AntScene, d direction) bool {
	np := a.pos.PointAt(d)
	if np.Within(0, 0, an.st.width, an.st.height) {
		return an.field.GetHomeWall(np.x, np.y) == e_wall
		//return an.field.wall[np.x+np.y*an.field.width]
		//return an.field.vals[np.x+np.y*an.field.width].Wall
		//return an.field.vals[a.idx].Wall
		//return an.field.Get(np.x, np.y).Wall
	}
	return true
}

// func (a *Ant) GridAt(an *AntScene, d direction) (*gridspot, bool) {
// 	np := a.pos.PointAt(d)
// 	if np.Within(0, 0, an.st.width, an.st.height) {
// 		//return an.grid[np.x][np.y], true
// 		return an.field.Get(np.x, np.y), true
// 	}
// 	return nil, false
// }

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

// func (a *Ant) OctantRect(d direction, size int) sdl.Rect {
// 	var (
// 		start point
// 		end   point
// 	)
// 	switch d {
// 	case N:
// 		start.x = a.pos.x - (size / 2)
// 		start.y = a.pos.y - size
// 		end.x = a.pos.x + (size / 2)
// 		end.y = a.pos.y
// 	case NE:
// 		start.x = a.pos.x
// 		start.y = a.pos.y - size
// 		end.x = a.pos.x + size
// 		end.y = a.pos.y
// 	case E:
// 		start.x = a.pos.x
// 		start.y = a.pos.y - (size / 2)
// 		end.x = a.pos.x + size
// 		end.y = a.pos.y + (size / 2)
// 	case SE:
// 		start.x = a.pos.x
// 		start.y = a.pos.y
// 		end.x = a.pos.x + size
// 		end.y = a.pos.y + size
// 	case S:
// 		start.x = a.pos.x - (size / 2)
// 		start.y = a.pos.y
// 		end.x = a.pos.x + (size / 2)
// 		end.y = a.pos.y + size
// 	case SW:
// 		start.x = a.pos.x - size
// 		start.y = a.pos.y
// 		end.x = a.pos.x
// 		end.y = a.pos.y + size
// 	case W:
// 		start.x = a.pos.x - size
// 		start.y = a.pos.y - (size / 2)
// 		end.x = a.pos.x
// 		end.y = a.pos.y + (size / 2)
// 	case NW:
// 		start.x = a.pos.x - size
// 		start.y = a.pos.y - size
// 		end.x = a.pos.x
// 		end.y = a.pos.y
// 	}
// 	return sdl.Rect{int32(start.x), int32(start.y), int32(end.x - start.x), int32(end.y - start.y)}
// }

func (a *Ant) Line(an *AntScene, d direction, size int) gridspot {
	var pt gridspot
	//addspot := func(g *gridspot) bool {
	addspot := func(x, y int) bool {
		idx := x + y*an.field.width
		//if an.field.wall[idx] {
		if an.field.GetHomeWall(x, y) == e_wall {
			pt.Wall = true
			return false
		}
		// if g.Food < 0 {
		// 	panic("g.FOOD < 0 \n")
		// }
		//pt.FoodPher += int(an.field.foodpher[idx]) + (an.field.food[idx] << 8) //g.Food*pheromoneMax*2
		pt.FoodPher += an.field.GetFoodPher(x, y) + int(an.field.food[idx]<<8)
		//pt.HomePher += int(an.field.homepher[idx])
		pt.HomePher += int(an.field.GetHomePher(x, y))
		//if an.field.home[idx] {
		if an.field.GetHomeWall(x, y) == e_home {
			pt.HomePher += pheromoneMax << 1 //* 2
		}
		return true
	}
	switch d {
	case N:
		for y := a.pos.y; y > a.pos.y-size; y-- {
			if y < 0 {
				pt.Wall = true
				break
			}
			//if !addspot(an.field.Get(a.pos.x, y)) {
			if !addspot(a.pos.x, y) {
				break
			}
		}
	case NE:
		for i := 0; i < size; i++ {
			y := a.pos.y - i
			x := a.pos.x + i
			if y < 0 || x >= an.field.width {
				pt.Wall = true
				break
			}
			if !addspot(x, y) {
				break
			}
		}
	case E:
		for x := a.pos.x; x < a.pos.x+size; x++ {
			if x >= an.field.width {
				pt.Wall = true
				break
			}
			if !addspot(x, a.pos.y) {
				break
			}
		}
	case SE:
		for i := 0; i < size; i++ {
			y := a.pos.y + i
			x := a.pos.x + i
			if y >= an.field.height || x >= an.field.width {
				pt.Wall = true
				break
			}
			if !addspot(x, y) {
				break
			}
		}
	case S:
		for y := a.pos.y; y < a.pos.y+size; y++ {
			if y >= an.field.height {
				pt.Wall = true
				break
			}
			if !addspot(a.pos.x, y) {
				break
			}
		}
	case SW:
		for i := 0; i < size; i++ {
			y := a.pos.y + i
			x := a.pos.x - i
			if y >= an.field.height || x < 0 {
				pt.Wall = true
				break
			}
			if !addspot(x, y) {
				break
			}
		}
	case W:
		for x := a.pos.x; x > a.pos.x-size; x-- {
			if x < 0 {
				pt.Wall = true
				break
			}
			if !addspot(x, a.pos.y) {
				break
			}
		}
	case NW:
		for i := 0; i < size; i++ {
			y := a.pos.y - i
			x := a.pos.x - i
			if y < 0 || x < 0 {
				pt.Wall = true
				break
			}
			if !addspot(x, y) {
				break
			}
		}
	default:
		panic("NO SUCH DIRECTION")
	}
	return pt
}

func (a *Ant) SumOctant(g *GameState, an *AntScene, d direction, size int) gridspot {
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
			if p.Within(0, 0, an.st.width, an.st.height) {
				//spot := an.field.Get(x, y)
				//if !an.field.wall[x+y*an.field.width] {
				if an.field.GetHomeWall(x, y) != e_wall {
					//pt.FoodPher += int(an.field.foodpher[x+y*an.field.width]) + an.field.food[x+y*an.field.width]*100000 // - (an.grid[x][y].homePher / 4)
					//pt.HomePher += int(an.field.homepher[x+y*an.field.width])                                            // - (an.grid[x][y].foodPher / 4)

					pt.FoodPher += an.field.GetFoodPher(x, y) + int(an.field.food[x+y*an.field.width]*100000)
					pt.HomePher += int(an.field.GetHomePher(x, y))
					//if an.field.home[x+y*an.field.width] {
					if an.field.GetHomeWall(x, y) == e_home {
						pt.HomePher += 100000
					}
				}
			}
		}
	}
	// if an.pherOverload {
	// 	if pt.FoodPher > pheromoneMax*an.pherOverloadFactor {
	// 		pt.FoodPher = pheromoneMax * an.pherOverloadFactor
	// 	}
	// 	if pt.HomePher > pheromoneMax*an.pherOverloadFactor {
	// 		pt.HomePher = pheromoneMax * an.pherOverloadFactor
	// 	}
	// }
	return pt
}

func antFadeDivisor(d int) int {
	d -= 100
	if d <= 0 {
		d = 1
	}
	return d
}

// Returns whether or not the ant is alive
func (a *Ant) Update(as *AntScene) bool {
	a.Move(as)
	if a.life <= 0 {
		//as.field.Get(a.pos.x, a.pos.y).Food = 1
		// if a.food > 0 {
		// 	as.field.Get(a.pos.x, a.pos.y).Food += a.food
		// }
		return false
	}
	aidx := as.field.Idx(a.pos.x, a.pos.y)
	//if as.field.vals[a.idx].Home {
	//if as.field.home[aidx] {
	if as.field.GetHomeWall(a.pos.x, a.pos.y) == e_home {
		//if as.field.Get(a.pos.x, a.pos.y).Home {
		if a.food > 0 {
			as.homelife += int64(a.food) * int64(as.st.foodlife)
			a.food = 0
		}
		// need := int64(antlife - a.life)
		// if need > as.homefood {
		// 	need = as.homefood
		// }
		// as.homefood -= need
		// a.life += int(need)
		a.marker = marker
	}
	//if spot := as.field.vals[a.idx]; spot.Food > 0 {
	//if spot := as.field.Get(a.pos.x, a.pos.y); spot.Food > 0 {
	if food := as.field.food[aidx]; food > 0 {
		if a.food == 0 {
			//a.dir = a.dir.Right(4)
			if food > 10 {
				as.field.food[aidx] -= 10
				as.field.Update(as, a.pos.x, a.pos.y)
				a.food = 10
			} else {
				a.food = int(food)
				as.field.food[aidx] = 0
				as.field.Update(as, a.pos.x, a.pos.y)
			}
		}
		a.marker = marker
	}

	if a.food > 0 {
		a.tex = as.fullTextures[a.dir]
		//spot := as.field.vals[a.idx]
		//spot := as.field.Get(a.pos.x, a.pos.y)
		//fp := int(as.field.foodpher[aidx])
		fp := int(as.field.GetFoodPher(a.pos.x, a.pos.y))
		if fp > a.marker {
			a.marker = fp
			a.marker -= (a.marker / antFadeDivisor(as.st.fadedivisor)) + 1
		} else {
			//as.field.foodpher[aidx] = uint32(a.marker)
			as.field.SetFoodPher(a.pos.x, a.pos.y, a.marker)
			a.marker -= (a.marker / antFadeDivisor(as.st.fadedivisor)) + 1
			if as.st.renderPher {
				as.field.Update(as, a.pos.x, a.pos.y)
			}
		}
	} else {
		a.tex = as.textures[a.dir]
		//spot := as.field.vals[a.idx]
		//spot := as.field.Get(a.pos.x, a.pos.y)
		//hp := int(as.field.homepher[aidx])
		hp := as.field.GetHomePher(a.pos.x, a.pos.y)
		if hp > a.marker {
			a.marker = hp
			a.marker -= (a.marker / antFadeDivisor(as.st.fadedivisor)) + 1
		} else {
			//as.field.homepher[aidx] = uint32(a.marker)
			as.field.SetHomePher(a.pos.x, a.pos.y, a.marker)
			a.marker -= (a.marker / antFadeDivisor(as.st.fadedivisor)) + 1
			if as.st.renderPher {
				as.field.Update(as, a.pos.x, a.pos.y)
			}
		}
	}
	return true
}

func (a *Ant) Move(an *AntScene) {
	a.life -= 1
	if a.life <= 0 {
		return
		// if a.food > 0 {
		// 	a.food--
		// 	a.life += foodlife
		// } else {
		// 	return
		// }
	}

	// We need ants to not always follow exactly the right path, or else they
	// get stuck following very tight lines, and never explore.
	//fmt.Printf("Dizziness: %d\n", a.dizziness)
	if n := rand.Intn(10); n == 0 {
		// straight := a.SumOctant(an, a.dir, 50)
		// left := a.SumOctant(an, a.dir.Left(1), 50)
		// right := a.SumOctant(an, a.dir.Right(1), 50)

		//const sight = 50
		//const sight = 10
		straight := a.Line(an, a.dir, an.st.sight)
		left := a.Line(an, a.dir.Left(1), an.st.sight)
		lleft := a.Line(an, a.dir.Left(2), an.st.sight)
		right := a.Line(an, a.dir.Right(1), an.st.sight)
		rright := a.Line(an, a.dir.Right(2), an.st.sight)

		if straight.FoodPher < 0 || right.FoodPher < 0 || left.FoodPher < 0 {
			panic(fmt.Sprintf("Ant(%d,%d,%d): Less that zero: straight: %#v, left: %#v, right: %#v, lleft: %#v, rright: %#v",
				a.pos.x, a.pos.y, a.dir, straight, left, right, lleft, rright))
		}

		// Directions include weighted values of their left and right directions
		straight.FoodPher += left.FoodPher/2 + right.FoodPher/2
		straight.HomePher += left.HomePher/2 + right.HomePher/2
		left.FoodPher += lleft.FoodPher/2 + straight.FoodPher/2
		left.HomePher += lleft.HomePher/2 + straight.HomePher/2
		right.FoodPher += rright.FoodPher/2 + straight.FoodPher/2
		right.HomePher += rright.HomePher/2 + straight.HomePher/2

		followingPher := false
		if a.food > 0 { //|| a.life < antlife/2 { // go home if we have food or we need food
			// if rightPower > straightPower && rightPower > leftPower {
			// 	a.dir = a.dir.Right(1)
			// 	followingPher = true
			// } else if leftPower > straightPower && leftPower > rightPower {
			// 	a.dir = a.dir.Left(1)
			// 	followingPher = true
			// }
			if right.HomePher > straight.HomePher && right.HomePher > left.HomePher {
				a.dir = a.dir.Right(1)
				followingPher = true
			} else if left.HomePher > straight.HomePher && left.HomePher > right.HomePher {
				a.dir = a.dir.Left(1)
				followingPher = true
			} else if straight.HomePher > left.HomePher && straight.HomePher > right.HomePher {
				followingPher = true
			}
		} else {
			if an.st.antisocial {
				if straight.Wall {
					straight.HomePher += pheromoneMax * an.st.sight
				}
				if left.Wall {
					left.HomePher += pheromoneMax * an.st.sight
				}
				if right.Wall {
					right.HomePher += pheromoneMax * an.st.sight
				}
				straightPower := straight.HomePher - (straight.FoodPher * 2)
				leftPower := left.HomePher - (left.FoodPher * 2)
				rightPower := right.HomePher - (right.FoodPher * 2)

				if rightPower < straightPower && rightPower < leftPower {
					a.dir = a.dir.Right(1)
					//followingPher = true
				} else if leftPower < straightPower && leftPower < rightPower {
					a.dir = a.dir.Left(1)
					//followingPher = true
				}
			} else {
				if right.FoodPher > straight.FoodPher && right.FoodPher > left.FoodPher {
					a.dir = a.dir.Right(1)
					followingPher = true
				} else if left.FoodPher > straight.FoodPher && left.FoodPher > right.FoodPher {
					a.dir = a.dir.Left(1)
					followingPher = true
				} else if straight.FoodPher > left.FoodPher && straight.FoodPher > right.FoodPher {
					followingPher = true
				}
			}
		}

		if an.st.followWalls {
			if !followingPher {
				if lleft.Wall {
					if left.Wall {
						if straight.Wall {
							a.dir = a.dir.Right(1)
						}
					} else {
						a.dir = a.dir.Left(1)
					}
				}
				if rright.Wall {
					if right.Wall {
						if straight.Wall {
							a.dir = a.dir.Left(1)
						}
					} else {
						a.dir = a.dir.Right(1)
					}
				}
			}
		}

		// Take a random turn every once in a while
		n := rand.Intn(10)
		if n == 0 {
			a.dir = a.dir.Left(1)
		} else if n == 1 {
			a.dir = a.dir.Right(1)
		}
	}

	// g, ok := a.GridAt(an, a.dir)

	// w := !ok || g.Wall

	// if w {
	if a.WallAt(an, a.dir) {
		a.dir = a.dir.Right((rand.Intn(3) - 1) * 2)
		//g, ok := a.GridAt(an, a.dir)
		i := 0
		for a.WallAt(an, a.dir) {
			a.dir = a.dir.Right((rand.Intn(3) - 1) * 2)
			//a.dir = a.dir.Right(1)
			i++
			if i >= 64 {
				a.pos.x = 1
				a.pos.y = 1
				return
			}
		}
	}

	a.pos = a.pos.PointAt(a.dir)
	//a.idx = a.pos.x + a.pos.y*an.field.width
}
