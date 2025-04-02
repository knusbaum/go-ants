package main

import (
	"fmt"
	"math/rand"
)

// Based on CaptainLuma's maze generation algorithm:
// https://captainluma.github.io/New-Maze-Generating-Algorithm/
type node int

const (
	node_up = iota
	node_down
	node_left
	node_right
	node_none
)

type Maze [][]node

func makeMaze(x, y int) Maze {
	fmt.Printf("Drawing maze X: %d, Y: %d\n", x, y)
	m := make([][]node, x)
	for i := 0; i < x; i++ {
		m[i] = make([]node, y)
		for j := 0; j < y; j++ {
			if i == 0 {
				m[i][j] = node_up
			} else {
				m[i][j] = node_left
			}
		}
	}
	m[0][0] = node_none
	generateMaze(m, 0, 0, x*y*10)
	return m
}

func generateMaze(m Maze, x, y int, iterations int) {
	for iterations > 0 {
		next := node(rand.Intn(node_none))
		switch next {
		case node_up:
			if y == 0 {
				continue
			}
			m[x][y] = next
			y -= 1
		case node_down:
			if y == len(m[0])-1 {
				continue
			}
			m[x][y] = next
			y += 1
		case node_left:
			if x == 0 {
				continue
			}
			m[x][y] = next
			x -= 1
		case node_right:
			if x == len(m)-1 {
				continue
			}
			m[x][y] = next
			x += 1
		default:
			panic("OOF")
		}
		iterations -= 1
	}
	m[x][y] = node_none
}

func drawMaze(as *AntScene, origin point, m Maze) {
	for x := range m {
		for y := range m[x] {
			doNode(as, origin, x, y, m)
		}
	}
}

func drawMazePath(as *AntScene, origin point, m Maze) {
	for x := range m {
		for y := range m[x] {
			doNodePath(as, origin, x, y, m)
		}
	}
}

func doNodePath(as *AntScene, origin point, x, y int, m Maze) {
	const linesize = 40
	startx := origin.x + (x * linesize) + linesize/2
	starty := origin.y + (y * linesize) + linesize/2

	// endx := startx
	// endy := starty
	// switch m[x][y] {
	// case node_up:
	// 	endy = starty - linesize/2
	// case node_down:
	// 	endy = starty + linesize/2
	// case node_left:
	// 	endx = startx - linesize/2
	// case node_right:
	// 	endx = startx + linesize/2
	// }

	if m[x][y] == node_none {
		as.spawnLocation = point{startx, starty}
		doSpot(as, 10, startx, starty, func(x, y int, spot *gridspot) {
			spot.Home = true
			as.field.Update(x, y)
		})
	} else if rand.Intn(10) == 0 {
		doSpot(as, 10, startx, starty, func(x, y int, spot *gridspot) {
			spot.Food = 100
			as.field.Update(x, y)
		})
	}
	// else {
	// 	doLine(startx, starty, endx, endy, func(cx, cy int) {
	// 		doSpot(as, 3, cx, cy, func(x, y int, spot *gridspot) {
	// 			spot.Food = 1
	// 			as.field.Update(x, y)
	// 		})
	// 	})
	// }
}

func doNode(as *AntScene, origin point, x, y int, m Maze) {
	const linesize = 40
	drawtop := true
	drawleft := true
	drawright := true
	drawbottom := true

	switch m[x][y] {
	case node_up:
		drawtop = false
	case node_down:
		drawbottom = false
	case node_left:
		drawleft = false
	case node_right:
		drawright = false
	}
	if y > 0 && m[x][y-1] == node_down {
		drawtop = false
	}
	if y < len(m[0])-1 && m[x][y+1] == node_up {
		drawbottom = false
	}
	if x > 0 && m[x-1][y] == node_right {
		drawleft = false
	}
	if x < len(m)-1 && m[x+1][y] == node_left {
		drawright = false
	}

	startx := origin.x + (x * linesize)
	starty := origin.y + (y * linesize)

	if drawtop {
		doLine(startx, starty, startx+linesize, starty, func(cx, cy int) {
			doSpot(as, 3, cx, cy, func(x, y int, spot *gridspot) {
				spot.Wall = true
				as.field.Update(x, y)
			})
		})
	}
	if drawbottom {
		doLine(startx, starty+linesize, startx+linesize, starty+linesize, func(cx, cy int) {
			doSpot(as, 3, cx, cy, func(x, y int, spot *gridspot) {
				spot.Wall = true
				as.field.Update(x, y)
			})
		})
	}
	if drawleft {
		doLine(startx, starty, startx, starty+linesize, func(cx, cy int) {
			doSpot(as, 3, cx, cy, func(x, y int, spot *gridspot) {
				spot.Wall = true
				as.field.Update(x, y)
			})
		})
	}
	if drawright {
		doLine(startx+linesize, starty, startx+linesize, starty+linesize, func(cx, cy int) {
			doSpot(as, 3, cx, cy, func(x, y int, spot *gridspot) {
				spot.Wall = true
				as.field.Update(x, y)
			})
		})
	}
}
