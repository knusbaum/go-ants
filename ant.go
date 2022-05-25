package main

import (
	"math/rand"
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
		np.x -= 1
	case E:
		np.x -= 1
	case SE:
		np.y += 1
		np.x -= 1
	case S:
		np.y += 1
	case SW:
		np.y += 1
		np.x += 1
	case W:
		np.x += 1
	case NW:
		np.y -= 1
		np.x += 1
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
	// 	nx, ny := a.x, a.y
	// 	switch d {
	// 	case N:
	// 		ny -= 1
	// 	case NE:
	// 		ny -= 1
	// 		nx -= 1
	// 	case E:
	// 		nx -= 1
	// 	case SE:
	// 		ny += 1
	// 		nx -= 1
	// 	case S:
	// 		ny += 1
	// 	case SW:
	// 		ny += 1
	// 		nx += 1
	// 	case W:
	// 		nx += 1
	// 	case NW:
	// 		ny -= 1
	// 		nx += 1
	// 	}
	// 	if nx < 0 || nx >= WIDTH || ny < 0 || ny >= HEIGHT {
	// 		return point{}, false
	// 	}
	// 	return an.grid[nx][ny], true
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

// func (a *Ant) canMove() bool {
// 	nx, ny := a.x, a.y
// 	switch d {
// 	case N:
// 		ny -= 1
// 	case NE:
// 		ny -= 1
// 		nx -= 1
// 	case E:
// 		nx -= 1
// 	case SE:
// 		ny += 1
// 		nx -= 1
// 	case S:
// 		ny += 1
// 	case SW:
// 		ny += 1
// 		nx += 1
// 	case W:
// 		nx += 1
// 	case NW:
// 		ny -= 1
// 		nx += 1
// 	}
// 	if nx < 0 || nx >= WIDTH || ny < 0 || ny >= HEIGHT {
//
// }

func (a *Ant) Move(an *AntScene) {
	//fmt.Printf("MOVE\n")
	n := rand.Intn(400)
	if n == 0 {
		a.dir = a.dir.Left(1)
	} else if n == 1 {
		a.dir = a.dir.Right(1)
	}
	left, right := point{}, point{}
	if p, ok := a.GridAt(an, a.dir.Left(1)); ok {
		left.x += p.x
		left.y += p.y
	}
	if p, ok := a.GridAt(an, a.dir.Left(2)); ok {
		left.x += p.x
		left.y += p.y
	}
	if p, ok := a.GridAt(an, a.dir.Right(1)); ok {
		right.x += p.x
		right.y += p.y
	}
	if p, ok := a.GridAt(an, a.dir.Right(2)); ok {
		right.x += p.x
		right.y += p.y
	}
	if a.food > 0 {
		if right.y-right.x > left.y-left.x {
			a.dir = a.dir.Right(1)
		} else if left.y-left.x > right.y-right.x {
			a.dir = a.dir.Left(1)
		}
	} else {
		if right.x-right.y > left.x-left.y {
			a.dir = a.dir.Right(1)
		} else if left.x-left.y > right.x-right.y {
			a.dir = a.dir.Left(1)
		}
	}

	//fmt.Printf("CHECK POINT %d\n", a.dir)
	if _, ok := a.GridAt(an, a.dir); !ok {
		a.dir = a.dir.Right(4)
		_, ok := a.GridAt(an, a.dir)
		for ; !ok; _, ok = a.GridAt(an, a.dir) {
			a.dir = a.dir.Right(1)
		}
		//fmt.Printf("FLIPPED POINT %d\n", a.dir)
	}
	//nx, ny := a.x, a.y
	//fmt.Printf("AX: %d, Y: %d\n", a.x, a.y)
	np := a.pos.PointAt(a.dir)
	a.pos = np
	// 	switch a.dir {
	// 	case N:
	// 		a.y -= 1
	// 	case NE:
	// 		a.y -= 1
	// 		a.x -= 1
	// 	case E:
	// 		a.x -= 1
	// 	case SE:
	// 		a.y += 1
	// 		a.x -= 1
	// 	case S:
	// 		a.y += 1
	// 	case SW:
	// 		a.y += 1
	// 		a.x += 1
	// 	case W:
	// 		a.x += 1
	// 	case NW:
	// 		a.y -= 1
	// 		a.x += 1
	// 	}
	if a.marker > 0 {
		a.marker -= 1
	}

	//if a.marker < 10 && a.food > 0 {
	//if a.food > 0 && an.grid[a.x][a.y].x > 0 {
	//	a.marker = 10
	//}
	// 		else if an.grid[a.x][a.y].y > 0 {
	// 			a.marker = 1
	// 		}
	//	}

	//fmt.Printf("BX: %d, Y: %d\n", a.x, a.y)
}

// func (a *Ant) checkg(an *AntScene, x, y int, val int) bool {
// 	if x > 0 && x < WIDTH && y > 0 && y < HEIGHT {
// 		return an.grid[x][y].x > val
// 	}
// 	return false
// }
//
// func (a *Ant) checkl(an *AntScene, x, y int, val int) bool {
// 	if x > 0 && x < WIDTH && y > 0 && y < HEIGHT {
// 		return an.grid[x][y].y > val
// 	}
// 	return false
// }
//
// func (a *Ant) selectDirection(as *AntScene) (int, int) {
// 	return 0, 0
// 	//
// 	// 	val := int(0)
// 	// 	nx := a.x
// 	// 	ny := a.y
// 	//
// 	// 	for i := a.x - 1; i < a.x+2; i++ {
// 	// 		for j := a.y - 1; j < a.y+2; j++ {
// 	// 			if a.food > 0 {
// 	// 				if a.checkl(as, i, j, val) {
// 	// 					nx = i
// 	// 					ny = j
// 	// 					val = as.grid[i][j].y
// 	// 				}
// 	// 			} else {
// 	// 				if a.checkg(as, i, j, val) {
// 	// 					nx = i
// 	// 					ny = j
// 	// 					val = as.grid[i][j].x - as.grid[i][j].y
// 	// 				}
// 	// 			}
// 	// 		}
// 	// 	}
// 	// 	if nx == a.x && ny == a.y {
// 	// 		nx := a.x + a.dirx
// 	// 		ny := a.y + a.diry
// 	// 		r := rand.Intn(50)
// 	// 		for r == 0 || nx < 0 || nx >= WIDTH || ny < 0 || ny >= HEIGHT {
// 	// 			a.dirx = rand.Intn(3) - 1
// 	// 			a.diry = rand.Intn(3) - 1
// 	// 			nx = a.x + a.dirx
// 	// 			ny = a.y + a.diry
// 	// 			r = rand.Intn(100)
// 	// 		}
// 	//
// 	// 		// 		bs := make([]point, 0, 8)
// 	// 		// 		for i := a.x - 1; i < a.x+2; i++ {
// 	// 		// 			if i == a.x {
// 	// 		// 				continue
// 	// 		// 			}
// 	// 		// 			for j := a.y - 1; j < a.y+2; j++ {
// 	// 		// 				if j == a.y {
// 	// 		// 					continue
// 	// 		// 				}
// 	// 		// 				if i > 0 && i < WIDTH && j > 0 && j < HEIGHT {
// 	// 		// 					bs = append(bs, point{i, j})
// 	// 		// 				}
// 	// 		// 			}
// 	// 		// 		}
// 	// 		// 		pt := bs[rand.Intn(len(bs))]
// 	// 		// 		return pt.x, pt.y
// 	// 		return nx, ny
// 	// 	}
// 	// 	return nx, ny
// }
