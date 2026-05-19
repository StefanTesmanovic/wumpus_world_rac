package agent

import (
	"fmt"
	"wumpusworld/pkg/world"
)

type Agent struct {
	Pos world.Coord
	Dir int

	KnownSafe map[world.Coord]bool
	Visited   map[world.Coord]bool
	Plan      []world.Action

	ArrowCount int
	Start      world.Coord
}

// Controller is the interface implemented by both AI and human agents.
type Controller interface {
	Next(world.Percept, *world.World) world.Action
	Notify(world.Action, world.Percept, *world.World)
}

func New(x, y, dir int) *Agent {
	a := &Agent{
		Pos:        world.Coord{X: x, Y: y},
		Dir:        dir,
		KnownSafe:  make(map[world.Coord]bool),
		Visited:    make(map[world.Coord]bool),
		Plan:       []world.Action{},
		ArrowCount: 1,
	}
	a.KnownSafe[a.Pos] = true
	a.Start = a.Pos
	return a
}

func (a *Agent) hasPlanned() bool {
	return len(a.Plan) > 0
}

func (a *Agent) popPlan() world.Action {
	if len(a.Plan) == 0 {
		return world.Action(-1)
	}
	act := a.Plan[0]
	a.Plan = a.Plan[1:]
	return act
}

func (a *Agent) neighborCoords(w *world.World) []world.Coord {
	// order: East, South, West, North
	offs := []world.Coord{{1, 0}, {0, -1}, {-1, 0}, {0, 1}}
	out := []world.Coord{}
	for _, o := range offs {
		nx := a.Pos.X + o.X
		ny := a.Pos.Y + o.Y
		if nx >= 0 && nx < w.Width && ny >= 0 && ny < w.Height {
			out = append(out, world.Coord{nx, ny})
		}
	}
	return out
}

func dirForDelta(dx, dy int) int {
	switch {
	case dx == 1 && dy == 0:
		return world.East
	case dx == -1 && dy == 0:
		return world.West
	case dx == 0 && dy == 1:
		return world.North
	case dx == 0 && dy == -1:
		return world.South
	}
	return -1
}

func rotationActions(curDir, targetDir int) []world.Action {
	diff := (targetDir - curDir + 4) % 4
	switch diff {
	case 0:
		return []world.Action{world.Forward}
	case 1:
		return []world.Action{world.TurnRight, world.Forward}
	case 2:
		return []world.Action{world.TurnRight, world.TurnRight, world.Forward}
	case 3:
		return []world.Action{world.TurnLeft, world.Forward}
	}
	return []world.Action{world.Forward}
}

// Next decides the next action given the last percept and the world reference.
func (a *Agent) Next(p world.Percept, w *world.World) world.Action {
	if a.hasPlanned() {
		return a.popPlan()
	}

	// mark current as visited
	a.Visited[a.Pos] = true

	// if glitter: grab immediately
	if p.Glitter {
		return world.Grab
	}

	// if no breeze and no stench, mark neighbors safe
	if !p.Breeze && !p.Stench {
		for _, n := range a.neighborCoords(w) {
			a.KnownSafe[n] = true
		}
	}

	// try immediate safe unvisited neighbors
	for _, n := range a.neighborCoords(w) {
		if a.KnownSafe[n] && !a.Visited[n] {
			// plan rotation + move
			dx := n.X - a.Pos.X
			dy := n.Y - a.Pos.Y
			targetDir := dirForDelta(dx, dy)
			acts := rotationActions(a.Dir, targetDir)
			a.Plan = append(a.Plan, acts...) // ... unpacks slice into elements
			return a.popPlan()
		}
	}

	// search for any known safe but unvisited cell (BFS)
	target, path := a.findNearestKnownUnvisited(w)
	if target != nil {
		// convert path to actions
		dir := a.Dir
		for i := 1; i < len(path); i++ {
			cur := path[i-1]
			nxt := path[i]
			dx := nxt.X - cur.X
			dy := nxt.Y - cur.Y
			td := dirForDelta(dx, dy)
			acts := rotationActions(dir, td)
			a.Plan = append(a.Plan, acts...)
			// update dir assuming actions will be executed
			dir = td
		}
		if a.hasPlanned() {
			return a.popPlan()
		}
	}

	// If we smell a wumpus and have arrows, try shooting in directions (simple heuristic)
	if p.Stench && a.ArrowCount > 0 {
		// try all directions: Right, Down, Left, Up
		tries := []int{world.East, world.South, world.West, world.North}
		for _, td := range tries {
			acts := rotationActions(a.Dir, td)
			// rotationActions appends a Forward; remove final forward and replace with Shoot
			if len(acts) > 0 && acts[len(acts)-1] == world.Forward {
				acts = acts[:len(acts)-1]
			}
			acts = append(acts, world.Shoot)
			a.Plan = append(a.Plan, acts...)
			return a.popPlan()
		}
	}

	// nothing else to do: finish
	return world.Climb
}

// findNearestKnownUnvisited returns the nearest known safe cell that is not yet visited and the path to it.
func (a *Agent) findNearestKnownUnvisited(w *world.World) (*world.Coord, []world.Coord) {
	type node struct{ C world.Coord }
	q := []world.Coord{a.Pos}
	prev := make(map[world.Coord]*world.Coord)
	seen := make(map[world.Coord]bool)
	seen[a.Pos] = true

	var found *world.Coord
	for len(q) > 0 {
		cur := q[0]
		q = q[1:]
		// neighbor offsets
		offs := []world.Coord{{1, 0}, {0, -1}, {-1, 0}, {0, 1}}
		for _, o := range offs {
			nx := cur.X + o.X
			ny := cur.Y + o.Y
			nc := world.Coord{X: nx, Y: ny}
			if nx < 0 || nx >= w.Width || ny < 0 || ny >= w.Height {
				continue
			}
			if seen[nc] {
				continue
			}
			// only traverse known safe cells
			if !a.KnownSafe[nc] {
				continue
			}
			seen[nc] = true
			// set prev
			ccopy := cur
			prev[nc] = &ccopy
			// if this is a known safe unvisited cell, return path
			if !a.Visited[nc] {
				found = &nc
				// build path
				path := []world.Coord{nc}
				p := prev[nc]
				for p != nil {
					path = append([]world.Coord{*p}, path...)
					p = prev[*p]
				}
				return found, path
			}
			q = append(q, nc)
		}
	}
	return nil, nil
}

// Notify updates the agent's internal state after an action was executed and the world responded.
func (a *Agent) Notify(act world.Action, p world.Percept, w *world.World) {
	// update position and direction from world (ground truth)
	a.Pos = world.Coord{X: w.AgentX, Y: w.AgentY}
	a.Dir = w.AgentDir
	if p.Scream {
		// if wumpus died, mark all cells as possibly safe later when observed
		// For simplicity do nothing here besides logging.
		fmt.Println("Agent heard a scream (Wumpus may be dead)")
	}
	if act == world.Shoot {
		if a.ArrowCount > 0 {
			a.ArrowCount--
		}
	}
	// mark current as safe and visited if alive
	a.KnownSafe[a.Pos] = true
	a.Visited[a.Pos] = true
}
