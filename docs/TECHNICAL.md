# Technical Implementation & Full Details

This file documents the implementation details, algorithms, invariants and semantics used throughout the Go port.

## Data model
- `type Coord struct { X, Y int }` — integer coordinates (0-based), origin is bottom-left when rendering.
- `type Cell struct { Pit bool; Wumpus bool; Gold bool }` — boolean flags for hazards and treasure.
- `type World struct { Width, Height int; Cells [][]Cell; AgentX, AgentY int; AgentDir int; AgentAlive bool; AgentHasGold bool; WumpusAlive bool; lastScream bool }`.
- `type Percept struct { Stench, Breeze, Glitter, Bump, Scream bool }`.
- Directions: `North`, `East`, `South`, `West` are `0..3`. Movement semantics:
  - `North` increments `Y`.
  - `East` increments `X`.
  - `South` decrements `Y`.
  - `West` decrements `X`.

## World construction (`New(width,height)`)
- The constructor seeds the RNG with `time.Now().UnixNano()` and builds an empty grid of `width x height` cells.
- Starting tile is fixed at `(0, height-1)` (bottom-left) and is excluded from pit placement.
- Any tile adjacent to the starting tile (4-neighborhood) is also excluded from pit placement so the player has safe initial moves.
- Pit placement rule: for every tile that is not excluded, set `Pit = true` with probability exactly 20% (independent draws). Implementation uses `rand.Float64() < 0.20`.
- After pits are placed, Wumpus and Gold are placed randomly among candidate tiles (non-start, non-pit). The Wumpus is chosen first, then Gold is chosen from the remaining candidates. If there are no candidates, the implementation falls back to placing a gold at the first available non-start tile.

## Percepts (`Sense()`)
- `Glitter` is true if the agent's current cell contains `Gold`.
- `Breeze` is true if any adjacent cell (4-neighborhood) contains a `Pit`.
- `Stench` is true if any adjacent cell contains a *live* `Wumpus`.
- `Scream` is true for one turn after a `Shoot` kills the Wumpus (the `lastScream` flag is set when `Shoot` hits the Wumpus and cleared when a percept is returned containing `Scream`).

## Actions and `Step(action)` semantics
- `Forward`: move one tile in the facing direction. If the destination is out-of-bounds, set `Percept.Bump = true` and stay in place.
- If the `Forward` results in entering a tile with `Pit` or a live `Wumpus`, `AgentAlive` becomes `false`.
- `TurnLeft` / `TurnRight` update `AgentDir` modulo 4.
- `Grab` picks up gold at the agent's tile (`AgentHasGold = true` and the cell's `Gold` cleared).
- `Shoot` fires along the facing direction until it hits a Wumpus or leaves the map; if it hits a live Wumpus, `WumpusAlive` becomes `false` and `lastScream` is set.
- `Climb` is a no-op for the terminal simulator (reserved for adapters that implement exits).

## Rendering (`Render()`)
- The renderer uses a fixed cell width (3 chars) and Unicode box-drawing characters to produce output like the example:

```
╔════════════╗
║ E  .  H  . ║ 3
║ .  .  .  . ║ 2
║ .  W  G  . ║ 1
║ .  .  H  . ║ 0
╚════════════╝
  0  1  2  3
```
- `symbolAt(x,y)` maps: `E` (agent), `W` (wumpus if alive), `G` (gold), `H` (pit), `.` otherwise.

## AI agent details (`pkg/agent/agent.go`)
- The AI agent maintains two maps: `KnownSafe map[Coord]bool` and `Visited map[Coord]bool` plus a `Plan []Action` (a queue of planned low-level actions). It starts knowing its start tile is safe.
- Decision flow in `Next(percept, world)` (high level):
  1. If a `Plan` exists, pop and execute the next low-level action.
  2. Mark current tile visited.
  3. If `Glitter` -> return `Grab` immediately.
  4. If neither `Breeze` nor `Stench`, mark all 4-neighbors as `KnownSafe`.
  5. Prefer immediate unvisited `KnownSafe` neighbors: plan rotate(s) + `Forward` to move there.
  6. Otherwise, BFS (`findNearestKnownUnvisited`) over `KnownSafe` cells to find a path, convert path to rotations+forwards and execute.
  7. If `Stench` and `ArrowCount>0`, heuristically attempt `Shoot` in plausible directions.
  8. If nothing left — `Climb` (end).
- Path planning notes: the BFS uses a queue + `prev` map to reconstruct the path. Complexity is O(n) where n = width*height for a single BFS.

## Human controller (`pkg/agent/human.go`)
- Prompts the player every turn with the current `Render()` and the `Percept` values. Provides hints (`Glitter`, `Breeze`, `Stench`) and lists possible commands: `f`/`l`/`r`/`g`/`s`/`c`/`q`.
- Reads input from `stdin` and maps commands to `Action`s. The human controller implements the same `Controller` contract as the AI agent.

## Path rotation helper (`rotationActions`) and orientation
- `rotationActions(curDir, targetDir)` computes the minimal rotation sequence to aim at `targetDir` and appends `Forward`. For example, to rotate from `East` to `South` it will return `[TurnRight, Forward]`.
- Calling code may remove the trailing `Forward` when converting a rotation into a `Shoot` action sequence.

## Randomness & reproducibility
- The constructor seeds the global `math/rand` RNG using `time.Now().UnixNano()`. To obtain reproducible maps you can replace the seeding call with a fixed seed or accept a `seed` parameter to `New(...)`.

## Complexity & limits
- All operations are single-threaded. BFS and other scans are O(n) with small constants; rendering is O(n).
- The grid is stored as `[][]Cell`. For very large maps you may prefer a sparse representation.

## Extension & testing ideas
- Add a `--seed` CLI flag to make world generation reproducible.
- Load/save world JSON files and add `cmd/wumpus load --file world.json`.
- Add unit tests for `Sense()`, `Step()` and `Render()` using deterministic seeds.
- Improve the AI to perform probabilistic inference (e.g., maintain probability that each tile contains a pit or Wumpus).

## Where to look in the code (quick links)
- World model and generation: [pkg/world/world.go](pkg/world/world.go)
- AI agent: [pkg/agent/agent.go](pkg/agent/agent.go)
- Human controller: [pkg/agent/human.go](pkg/agent/human.go)
- CLI runner: [cmd/wumpus/main.go](cmd/wumpus/main.go)
