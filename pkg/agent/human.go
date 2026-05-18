package agent

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"wumpusworld/pkg/world"
)

type Human struct {
	Pos        world.Coord
	Dir        int
	Reader     *bufio.Reader
	ArrowCount int
}

func NewHuman(x, y, dir int) *Human {
	return &Human{
		Pos:        world.Coord{X: x, Y: y},
		Dir:        dir,
		Reader:     bufio.NewReader(os.Stdin),
		ArrowCount: 1,
	}
}

func (h *Human) Next(p world.Percept, w *world.World) world.Action {
	fmt.Println("--- Player Turn ---")
	// show the current map and percepts to the human player
	fmt.Println(w.Render())
	fmt.Printf("Percept: stench=%v breeze=%v glitter=%v bump=%v scream=%v\n", p.Stench, p.Breeze, p.Glitter, p.Bump, p.Scream)
	if p.Glitter {
		fmt.Println("Hint: Gold is here — consider 'g' to grab.")
	}
	if p.Breeze {
		fmt.Println("Hint: Breeze detected — a pit may be adjacent.")
	}
	if p.Stench {
		fmt.Println("Hint: Stench detected — the Wumpus may be nearby.")
	}

	fmt.Println("Possible moves:")
	fmt.Println("  f  Forward")
	fmt.Println("  l  TurnLeft")
	fmt.Println("  r  TurnRight")
	fmt.Println("  g  Grab")
	fmt.Println("  s  Shoot (if you have an arrow)")
	fmt.Println("  c  Climb")
	fmt.Println("  q  Quit")

	for {
		fmt.Print("Enter move (f/l/r/g/s/c/q): ")
		line, err := h.Reader.ReadString('\n')
		if err != nil {
			fmt.Println("input error:", err)
			continue
		}
		s := strings.TrimSpace(strings.ToLower(line))
		switch s {
		case "f", "forward":
			return world.Forward
		case "l", "left", "turnleft":
			return world.TurnLeft
		case "r", "right", "turnright":
			return world.TurnRight
		case "g", "grab":
			return world.Grab
		case "s", "shoot":
			if h.ArrowCount > 0 {
				return world.Shoot
			}
			fmt.Println("No arrows left.")
		case "c", "climb":
			return world.Climb
		case "q", "quit":
			fmt.Println("Quitting.")
			os.Exit(0)
		default:
			fmt.Println("Unknown command. Try again.")
		}
	}
}

func (h *Human) Notify(act world.Action, p world.Percept, w *world.World) {
	h.Pos = world.Coord{X: w.AgentX, Y: w.AgentY}
	h.Dir = w.AgentDir
	if act == world.Shoot && h.ArrowCount > 0 {
		h.ArrowCount--
	}
}
