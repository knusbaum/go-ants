package main

type clickmode int

const (
	wall clickmode = iota
	food
	erase
	end
)

func (m clickmode) String() string {
	switch m {
	case wall:
		return "Wall"
	case food:
		return "Food"
	case erase:
		return "Erase"
	default:
		return "Error"
	}
}

type sorttype int

const (
	none sorttype = iota
	qsort
	slicesort
	sortend
	mergesort
	stdsort
)

func (t sorttype) String() string {
	switch t {
	case none:
		return "None"
	case qsort:
		return "Quicksort"
	case mergesort:
		return "Mergesort"
	case stdsort:
		return "sort.Sort"
	case slicesort:
		return "slices.SortFunc"
	default:
		return "Error (UNKNOWN SORT)"
	}
}

type GameState struct {
	width, height int

	renderPher  bool
	renderGreen bool
	renderRed   bool
	renderAnts  bool
	parallel    bool
	followWalls bool
	antisocial  bool
	antlife     int // an ant spends 1 life per frame
	foodcount   int // Amount of food to drop on a pixel while painting
	foodlife    int // amount of life that 1 food gives
	spawnparam  int // SpawnParam determines how much food the colony stockpiles before spawning more ants as a function of population.
	maxants     int // Crude limit to the number of ants spawned
	drawradius  int //Radius of the cursor paintbrush
	fadedivisor int // pheromone -= pheromone / fadedivisor // bigger number, slower fade
	sight       int
	leftmode    clickmode
	sorttype    sorttype
}
