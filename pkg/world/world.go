package world

import (
	"fmt"
	"math/rand"
	"strings"
	"time"
)

type Coord struct{ X, Y int }

type Cell struct {
	Pit    bool
	Wumpus bool
	Gold   bool
}

type Percept struct {
	Stench  bool
	Breeze  bool
	Glitter bool
	Bump    bool
	Scream  bool
}

type Action int

const (
	Forward Action = iota
	TurnLeft
	TurnRight
	Grab
	Shoot
	Climb
)

func (a Action) String() string {
	switch a {
	case Forward:
		return "Forward"
	case TurnLeft:
		return "TurnLeft"
	case TurnRight:
		return "TurnRight"
	case Grab:
		return "Grab"
	case Shoot:
		return "Shoot"
	case Climb:
		return "Climb"
	default:
		return "Unknown"
	}
}

const (
	North = iota
	East
	South
	West
)

type World struct {
	Width, Height int
	Cells         [][]Cell

	AgentX, AgentY int
	AgentDir       int
	AgentAlive     bool
	AgentHasGold   bool

	WumpusAlive bool
	lastScream  bool
}

// New constructs a world of the given size. If width or height are invalid
// they default to 4. When creating a 4x4 world the classic sample layout
// is used; otherwise a small default placement is used so the world is
// playable at other sizes.
func New(width, height int) *World {
	if width <= 0 || height <= 0 {
		width = 4
		height = 4
	}
	w := &World{Width: width, Height: height}
	w.Cells = make([][]Cell, w.Height)
	for y := 0; y < w.Height; y++ {
		w.Cells[y] = make([]Cell, w.Width)
	}

	// Seed RNG for pit/wumpus/gold placement
	rand.Seed(time.Now().UnixNano())

	// Starting tile is bottom-left (0, height-1)
	sx := 0
	sy := w.Height - 1

	// Mark excluded cells: start and tiles adjacent to start
	excluded := make(map[Coord]bool)
	excluded[Coord{sx, sy}] = true
	for _, n := range w.neighbors(sx, sy) {
		excluded[n] = true
	}

	// Place pits with ~15% probability on every non-excluded tile
	for y := 0; y < w.Height; y++ {
		for x := 0; x < w.Width; x++ {
			c := Coord{x, y}
			if excluded[c] {
				continue
			}
			if rand.Float64() < 0.15 {
				w.Cells[y][x].Pit = true
			}
		}
	}

	// Build list of candidate cells for Wumpus/Gold: non-start, non-pit
	candidates := []Coord{}
	for y := 0; y < w.Height; y++ {
		for x := 0; x < w.Width; x++ {
			if x == sx && y == sy {
				continue
			}
			if w.Cells[y][x].Pit {
				continue
			}
			candidates = append(candidates, Coord{x, y})
		}
	}

	// Ensure at least one candidate exists
	if len(candidates) == 0 {
		for y := 0; y < w.Height; y++ {
			for x := 0; x < w.Width; x++ {
				if x == sx && y == sy {
					continue
				}
				// clear a pit here and add as candidate
				w.Cells[y][x].Pit = false
				candidates = append(candidates, Coord{x, y})
				break
			}
			if len(candidates) > 0 {
				break
			}
		}
	}

	// Choose Wumpus
	if len(candidates) > 0 {
		wi := rand.Intn(len(candidates))
		wump := candidates[wi]
		w.Cells[wump.Y][wump.X].Wumpus = true
		// remove chosen from candidates
		candidates = append(candidates[:wi], candidates[wi+1:]...)
	}

	// Choose Gold from remaining candidates (if any)
	if len(candidates) > 0 {
		gi := rand.Intn(len(candidates))
		g := candidates[gi]
		w.Cells[g.Y][g.X].Gold = true
	} else {
		// Fallback: place gold in any non-start cell (even if it's the wumpus cell)
		for y := 0; y < w.Height; y++ {
			for x := 0; x < w.Width; x++ {
				if x == sx && y == sy {
					continue
				}
				w.Cells[y][x].Gold = true
				goto placedGold
			}
		}
	}
placedGold:

	// Agent starts at bottom-left (0, height-1) facing East
	w.AgentX = sx
	w.AgentY = sy
	w.AgentDir = East
	w.AgentAlive = true
	w.WumpusAlive = true
	return w
}

// NewSampleWorld constructs a 4x4 sample world matching the ASCII example.
func NewSampleWorld() *World {
	w := &World{Width: 4, Height: 4}
	w.Cells = make([][]Cell, w.Height)
	for y := 0; y < w.Height; y++ {
		w.Cells[y] = make([]Cell, w.Width)
	}

	// sample layout (x,y):
	// Pits at (2,3) and (2,0)
	w.Cells[3][2].Pit = true
	w.Cells[0][2].Pit = true

	// Wumpus at (1,1)
	w.Cells[1][1].Wumpus = true

	// Gold at (2,1)
	w.Cells[1][2].Gold = true

	// Agent start at (0,3) facing East
	w.AgentX = 0
	w.AgentY = 3
	w.AgentDir = East
	w.AgentAlive = true
	w.WumpusAlive = true
	return w
}

func (w *World) inBounds(x, y int) bool {
	return x >= 0 && x < w.Width && y >= 0 && y < w.Height
}

func (w *World) neighbors(x, y int) []Coord {
	out := []Coord{}
	cand := []Coord{{x + 1, y}, {x - 1, y}, {x, y + 1}, {x, y - 1}}
	for _, c := range cand {
		if w.inBounds(c.X, c.Y) {
			out = append(out, c)
		}
	}
	return out
}

// Sense returns the current percept at the agent's location.
func (w *World) Sense() Percept {
	p := Percept{}
	x := w.AgentX
	y := w.AgentY
	// Glitter if gold in same cell
	if w.Cells[y][x].Gold {
		p.Glitter = true
	}
	// Breeze if any adjacent pit
	for _, n := range w.neighbors(x, y) {
		if w.Cells[n.Y][n.X].Pit {
			p.Breeze = true
			break
		}
	}
	// Stench if any adjacent live wumpus
	if w.WumpusAlive {
		for _, n := range w.neighbors(x, y) {
			if w.Cells[n.Y][n.X].Wumpus {
				p.Stench = true
				break
			}
		}
	}
	// Bump handled during Step
	if w.lastScream {
		p.Scream = true
		w.lastScream = false
	}
	return p
}

func (w *World) Step(a Action) Percept {
	p := Percept{}
	if !w.AgentAlive {
		return p
	}

	switch a {
	case TurnLeft:
		w.AgentDir = (w.AgentDir + 3) % 4
	case TurnRight:
		w.AgentDir = (w.AgentDir + 1) % 4
	case Forward:
		nx, ny := w.AgentX, w.AgentY
		switch w.AgentDir {
		case North:
			ny++
		case East:
			nx++
		case South:
			ny--
		case West:
			nx--
		}
		if !w.inBounds(nx, ny) {
			p.Bump = true
			// agent stays in place
		} else {
			w.AgentX = nx
			w.AgentY = ny
			// check death conditions
			c := w.Cells[w.AgentY][w.AgentX]
			if c.Pit {
				w.AgentAlive = false
			}
			if c.Wumpus && w.WumpusAlive {
				w.AgentAlive = false
			}
		}
	case Grab:
		if w.Cells[w.AgentY][w.AgentX].Gold {
			w.AgentHasGold = true
			w.Cells[w.AgentY][w.AgentX].Gold = false
		}
	case Shoot:
		// fire arrow along facing direction until hit or edge
		ax, ay := w.AgentX, w.AgentY
		for {
			switch w.AgentDir {
			case North:
				ay++
			case East:
				ax++
			case South:
				ay--
			case West:
				ax--
			}
			if !w.inBounds(ax, ay) {
				break
			}
			if w.Cells[ay][ax].Wumpus && w.WumpusAlive {
				w.WumpusAlive = false
				w.lastScream = true
				break
			}
		}
	case Climb:
		// climbing only meaningful in GUI; treated as no-op here
	}

	// After action, compute percept relative to new state
	sp := w.Sense()
	// If we had a bump or scream from action, preserve them
	sp.Bump = sp.Bump || p.Bump
	sp.Scream = sp.Scream || p.Scream || w.lastScream
	return sp
}

func (w *World) symbolAt(x, y int) string {
	if w.AgentX == x && w.AgentY == y && w.AgentAlive {
		return "P"
	}
	c := w.Cells[y][x]
	if c.Wumpus && w.WumpusAlive {
		return "W"
	}
	if c.Gold {
		return "G"
	}
	if c.Pit {
		return "H"
	}
	return "."
}

func (w *World) Render() string {
	var b strings.Builder
	cellW := 3 // Every cell will be exactly 3 characters wide
	innerW := w.Width * cellW

	// 1. Top border
	fmt.Fprintf(&b, "╔%s╗\n", strings.Repeat("═", innerW))

	// 2. Map Rows (Printed from top/max-Y down to 0)
	for y := w.Height - 1; y >= 0; y-- {
		b.WriteString("║")
		for x := 0; x < w.Width; x++ {
			// Get symbol and pad/center it evenly to match cellW
			tok := w.symbolAt(x, y)
			b.WriteString(centerToken(tok, cellW))
		}
		// Right border and Y coordinate label
		fmt.Fprintf(&b, "║  %d\n", y)
	}

	// 3. Bottom border (Fixed the trailing '╝' corner)
	fmt.Fprintf(&b, "╚%s╝\n", strings.Repeat("═", innerW))

	// 4. Bottom X coordinates
	// Add 1 space to skip past the '╚' character, then center every index number
	b.WriteString(" ")
	for x := 0; x < w.Width; x++ {
		strX := fmt.Sprintf("%d", x)
		b.WriteString(centerToken(strX, cellW))
	}
	b.WriteByte('\n')

	return b.String()
}

// Helper function to keep text padding perfectly uniform
func centerToken(tok string, width int) string {
	l := len(tok)
	if l >= width {
		return tok[:width]
	}
	left := (width - l) / 2
	right := width - l - left
	return strings.Repeat(" ", left) + tok + strings.Repeat(" ", right)
}
