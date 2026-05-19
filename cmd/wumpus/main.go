package main

import (
	"flag"
	"fmt"
	"time"
	"wumpusworld/pkg/agent"
	"wumpusworld/pkg/world"
)

func main() {
	var maxSteps int
	var delayMs int
	var width int
	var height int
	var human bool

	flag.IntVar(&maxSteps, "steps", 200, "maximum steps to run")
	flag.IntVar(&delayMs, "delay", 200, "milliseconds between steps (0 for no delay)")
	flag.IntVar(&width, "width", 4, "world width")
	flag.IntVar(&height, "height", 4, "world height")
	flag.BoolVar(&human, "human", false, "use human-controlled agent (interactive)")
	flag.Parse()

	w := world.New(width, height)
	var ctrl agent.Controller
	if human {
		ctrl = agent.NewHuman(w.AgentX, w.AgentY, w.AgentDir)
	} else {
		ctrl = agent.New(w.AgentX, w.AgentY, w.AgentDir)
	}

	if human {
		fmt.Println("Interactive human-controlled session. Use -human=false to run AI agent.")
	}

	// initial percept
	percept := w.Sense()

	for step := 0; step < maxSteps; step++ {
		// capture pre-action render for comparison
		before := w.Render()

		if !human {
			// show current map for AI runs
			fmt.Println(before)
			fmt.Printf("Percept: stench=%v breeze=%v glitter=%v bump=%v scream=%v\n", percept.Stench, percept.Breeze, percept.Glitter, percept.Bump, percept.Scream)
		}

		act := ctrl.Next(percept, w)
		fmt.Printf("Step %d: Agent at (%d,%d) facing %s -> action: %s\n", step, w.AgentX, w.AgentY, world.DirName(w.AgentDir), act.String())
		percept = w.Step(act)
		ctrl.Notify(act, percept, w)

		// show percept after the action (map is printed at start of next loop)
		fmt.Printf("Percept: stench=%v breeze=%v glitter=%v bump=%v scream=%v\n", percept.Stench, percept.Breeze, percept.Glitter, percept.Bump, percept.Scream)

		if !w.AgentAlive {
			fmt.Println("Agent died. Simulation ends.")
			break
		}
		if w.AgentHasGold {
			fmt.Println("Agent grabbed the gold!")
			break
		}
		if delayMs > 0 {
			time.Sleep(time.Duration(delayMs) * time.Millisecond)
		}
	}
}
