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
# or run with flags
go run ./cmd/wumpus -steps 20 -delay 0 -human=false 
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


