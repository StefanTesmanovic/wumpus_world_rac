# wumpus-world (Go)

This repository contains a Go implementation of a Wumpus World simulator and a simple knowledge-based agent. The project focuses on the agent logic and a terminal-only visualization (ASCII) of the world — no graphical UI.

Features:
- 4x4 sample world matching a compact ASCII view
- Simple agent that explores safe cells, grabs gold, and uses an arrow to handle the Wumpus heuristically
- Terminal visualization showing the grid, agent, hazards and gold

Build and run
1. Install Go (1.20+ recommended).
2. From the repository root, build or run the simulator:

```bash
go run ./cmd/wumpus
# or build a binary
go build -o wumpus ./cmd/wumpus
./wumpus
```

Runtime flags:
- `-steps` (default 200): maximum steps to run
- `-delay` (default 200): milliseconds between steps (set 0 for no delay)
- `-width` (default 4): world width (columns)
- `-height` (default 4): world height (rows)
- `-human` (default false): run with a human-controlled agent (interactive)

Example output (trimmed):

	╔════════════╗
	║ E  .  H  . ║ 3
	║ .  .  .  . ║ 2
	║ .  W  G  . ║ 1
	║ .  .  H  . ║ 0
	╚════════════╗
		0  1  2  3

Project layout
- `cmd/wumpus` - command runner
- `pkg/world`  - world simulation and ASCII renderer
- `pkg/agent`  - the knowledge-based agent

Documentation
- Architecture overview: [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md)
- Technical deep-dive: [docs/TECHNICAL.md](docs/TECHNICAL.md)

Customization
- Adjust the sample world in `[pkg/world/world.go](pkg/world/world.go)` (pits, wumpus, gold, and start position).
- Improve or replace the agent in `[pkg/agent/agent.go](pkg/agent/agent.go)`.

If you'd like, I can:
- Add unit tests and a CI job
- Improve the agent's logic (probabilistic inference / single-arrow reasoning)
- Provide a small CLI to load world configurations from JSON

