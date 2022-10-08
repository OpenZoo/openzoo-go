package main

import "strings"

func GameDebugPrompt() {
	var (
		input  string
		i      int16
		toggle bool
	)
	input = ""
	SidebarClearLine(4)
	SidebarClearLine(5)
	PromptString(63, 5, 0x1E, 0x0F, 11, PROMPT_ANY, &input)
	input = strings.ToUpper(input)
	toggle = true
	if input[0] == '+' || input[0] == '-' {
		if input[0] == '-' {
			toggle = false
		}
		input = Copy(input, 2, Length(input)-1)
		if toggle {
			WorldSetFlag(input)
		} else {
			WorldClearFlag(input)
		}
	}
	DebugEnabled = WorldGetFlagPosition("DEBUG") >= 0
	if input == "HEALTH" {
		World.Info.Health += 50
	} else if input == "AMMO" {
		World.Info.Ammo += 5
	} else if input == "KEYS" {
		for i = 1; i <= 7; i++ {
			World.Info.Keys[i-1] = true
		}
	} else if input == "TORCHES" {
		World.Info.Torches += 3
	} else if input == "TIME" {
		World.Info.BoardTimeSec -= 30
	} else if input == "GEMS" {
		World.Info.Gems += 5
	} else if input == "DARK" {
		Board.Info.IsDark = toggle
		TransitionDrawToBoard()
	} else if input == "ZAP" {
		for i = 0; i <= 3; i++ {
			BoardDamageTile(int16(Board.Stats.At(0).X)+NeighborDeltaX[i], int16(Board.Stats.At(0).Y)+NeighborDeltaY[i])
			Board.Tiles.SetElement(int16(Board.Stats.At(0).X)+NeighborDeltaX[i], int16(Board.Stats.At(0).Y)+NeighborDeltaY[i], E_EMPTY)
			BoardDrawTile(int16(Board.Stats.At(0).X)+NeighborDeltaX[i], int16(Board.Stats.At(0).Y)+NeighborDeltaY[i])
		}
	}

	SoundQueue(10, "'\x04")
	SidebarClearLine(4)
	SidebarClearLine(5)
	GameUpdateSidebar()
}
