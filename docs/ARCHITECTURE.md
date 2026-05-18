# Architecture & Code Map

This document explains how the project is organized, where core logic lives, and how the runtime loop flows.

**Overview**
- The project is small and intentionally modular:
  - `pkg/world` — world model, percepts, actions and ASCII renderer ([pkg/world/world.go](pkg/world/world.go)).
  - `pkg/agent` — agents and controllers: the AI agent and the human controller ([pkg/agent/agent.go](pkg/agent/agent.go), [pkg/agent/human.go](pkg/agent/human.go)).
  - `cmd/wumpus` — command-line runner and flags ([cmd/wumpus/main.go](cmd/wumpus/main.go)).

**Key types & places**
- `World` (`pkg/world/world.go`): the grid and game state. Fields include `Width`, `Height`, `Cells`, `AgentX/AgentY`, `AgentDir`, `AgentAlive`, `AgentHasGold`, `WumpusAlive`.
- `Cell` (`pkg/world/world.go`): `Pit`, `Wumpus`, `Gold` booleans.
- `Percept` (`pkg/world/world.go`): `Stench`, `Breeze`, `Glitter`, `Bump`, `Scream` — produced by `World.Sense()`.
- `Action` (`pkg/world/world.go`): enum of `Forward`, `TurnLeft`, `TurnRight`, `Grab`, `Shoot`, `Climb`.
- `Agent` (`pkg/agent/agent.go`): the AI agent with `KnownSafe`, `Visited`, `Plan` and planning helpers.
- `Human` (`pkg/agent/human.go`): a console-driven controller that prompts the player and returns `Action`s.

**Core functions**
- `New(width,height)` in [pkg/world/world.go](pkg/world/world.go): constructs the world, generates pits, places the Wumpus and Gold, and sets the starting tile.
- `Sense()` in [pkg/world/world.go](pkg/world/world.go): computes the `Percept` for the agent's current cell.
- `Step(action)` in [pkg/world/world.go](pkg/world/world.go): executes an `Action`, updates the `World` and returns the resulting `Percept`.
- `Render()` in [pkg/world/world.go](pkg/world/world.go): returns an ASCII box with contents and coordinates.
- `Next(percept, world)` in [pkg/agent/agent.go](pkg/agent/agent.go): AI decision logic. The `Human` controller implements the same controller interface and provides interactive input instead.

**Simulation loop (where to look)**
The main loop is in [cmd/wumpus/main.go](cmd/wumpus/main.go):

1. Build the world: `w := world.New(width,height)`.
2. Select controller: `ctrl := agent.New(...)` or `agent.NewHuman(...)` depending on `-human` flag.
3. `percept := w.Sense()` then repeatedly:
   - `act := ctrl.Next(percept, w)`
   - `percept = w.Step(act)`
   - `ctrl.Notify(act, percept, w)`
   - display `w.Render()` and check `w.AgentAlive` / `w.AgentHasGold` for termination.

**Where to change behavior**
- World layout and generation: edit `New(width,height)` in [pkg/world/world.go](pkg/world/world.go).
- Pit probability and placement rules: see the pit-generation section inside `New(...)`.
- Wumpus/Gold placement: same `New(...)` function — randomness is seeded in that constructor.
- Agent decision logic: modify `Next(...)` in [pkg/agent/agent.go](pkg/agent/agent.go).
- Human UI prompts: edit [pkg/agent/human.go](pkg/agent/human.go).

**Run examples**
- Non-interactive (AI):
  ```bash
  go run ./cmd/wumpus -human=false -width 4 -height 4
  ```
- Interactive (human player):
  ```bash
  go run ./cmd/wumpus -human=true -width 4 -height 4
  ```

See also the technical deep-dive at [docs/TECHNICAL.md](docs/TECHNICAL.md) for details on the algorithms, data structures and exact semantics.
