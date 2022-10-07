//go:build editor

package main // unit: Editor

import (
	"bytes" // interface uses: GameVars, TxtWind
)

// implementation uses: Dos, Crt, Video, Sounds, Input, Elements, Oop, Game

type TDrawMode uint8

const (
	DrawingOff TDrawMode = iota + 1
	DrawingOn
	TextEntry
	EDITOR_COMPILED = true
)

var NeighborBoardStrs [4]string = [4]string{"       Board \x18", "       Board \x19", "       Board \x1b", "       Board \x1a"}

func EditorAppendBoard() {
	if World.BoardCount < MAX_BOARD {
		BoardClose()
		World.BoardCount++
		World.Info.CurrentBoard = World.BoardCount
		World.BoardData[World.BoardCount] = nil
		BoardCreate()
		TransitionDrawToBoard()
		for {
			PopupPromptString("Room's Title:", &Board.Name)
			if Length(Board.Name) != 0 {
				break
			}
		}
		TransitionDrawToBoard()
	}
}

func EditorLoop() {
	var (
		selectedCategory           int16
		elemMenuColor              int16
		wasModified                bool
		editorExitRequested        bool
		drawMode                   TDrawMode
		cursorX, cursorY           int16
		cursorPattern, cursorColor int16
		i, iElem                   int16
		canModify                  bool
		copiedStat                 TStat
		copiedHasStat              bool
		copiedTile                 TTile
		copiedX, copiedY           int16
		cursorBlinker              int16
	)
	EditorDrawSidebar := func() {
		var (
			i         int16
			copiedChr byte
		)
		SidebarClear()
		SidebarClearLine(1)
		VideoWriteText(61, 0, 0x1F, "     - - - -       ")
		VideoWriteText(62, 1, 0x70, "  ZZT Editor   ")
		VideoWriteText(61, 2, 0x1F, "     - - - -       ")
		VideoWriteText(61, 4, 0x70, " L ")
		VideoWriteText(64, 4, 0x1F, " Load")
		VideoWriteText(61, 5, 0x30, " S ")
		VideoWriteText(64, 5, 0x1F, " Save")
		VideoWriteText(70, 4, 0x70, " H ")
		VideoWriteText(73, 4, 0x1E, " Help")
		VideoWriteText(70, 5, 0x30, " Q ")
		VideoWriteText(73, 5, 0x1F, " Quit")
		VideoWriteText(61, 7, 0x70, " B ")
		VideoWriteText(65, 7, 0x1F, " Switch boards")
		VideoWriteText(61, 8, 0x30, " I ")
		VideoWriteText(65, 8, 0x1F, " Board Info")
		VideoWriteText(61, 10, 0x70, "  f1   ")
		VideoWriteText(68, 10, 0x1F, " Item")
		VideoWriteText(61, 11, 0x30, "  f2   ")
		VideoWriteText(68, 11, 0x1F, " Creature")
		VideoWriteText(61, 12, 0x70, "  f3   ")
		VideoWriteText(68, 12, 0x1F, " Terrain")
		VideoWriteText(61, 13, 0x30, "  f4   ")
		VideoWriteText(68, 13, 0x1F, " Enter text")
		VideoWriteText(61, 15, 0x70, " Space ")
		VideoWriteText(68, 15, 0x1F, " Plot")
		VideoWriteText(61, 16, 0x30, "  Tab  ")
		VideoWriteText(68, 16, 0x1F, " Draw mode")
		VideoWriteText(61, 18, 0x70, " P ")
		VideoWriteText(64, 18, 0x1F, " Pattern")
		VideoWriteText(61, 19, 0x30, " C ")
		VideoWriteText(64, 19, 0x1F, " Color:")
		for i = 9; i <= 15; i++ {
			VideoWriteText(61+i, 22, byte(i), "\xdb")
		}
		for i = 1; i <= EditorPatternCount; i++ {
			VideoWriteText(61+i, 22, 0x0F, Chr(ElementDefs[EditorPatterns[i-1]].Character))
		}
		if ElementDefs[copiedTile.Element].HasDrawProc {
			ElementDefs[copiedTile.Element].DrawProc(copiedX, copiedY, &copiedChr)
		} else {
			copiedChr = ElementDefs[copiedTile.Element].Character
		}
		VideoWriteText(62+EditorPatternCount, 22, copiedTile.Color, Chr(copiedChr))
		VideoWriteText(61, 24, 0x1F, " Mode:")
	}

	EditorDrawTileAndNeighborsAt := func(x, y int16) {
		var i, ix, iy int16
		BoardDrawTile(x, y)
		for i = 0; i <= 3; i++ {
			ix = x + NeighborDeltaX[i]
			iy = y + NeighborDeltaY[i]
			if ix >= 1 && ix <= BOARD_WIDTH && iy >= 1 && iy <= BOARD_HEIGHT {
				BoardDrawTile(ix, iy)
			}
		}
	}

	EditorUpdateSidebar := func() {
		if drawMode == DrawingOn {
			VideoWriteText(68, 24, 0x9E, "Drawing on ")
		} else if drawMode == TextEntry {
			VideoWriteText(68, 24, 0x9E, "Text entry ")
		} else if drawMode == DrawingOff {
			VideoWriteText(68, 24, 0x1E, "Drawing off")
		}

		VideoWriteText(72, 19, 0x1E, ColorNames[cursorColor-8-1])
		VideoWriteText(61+cursorPattern, 21, 0x1F, "\x1f")
		VideoWriteText(61+cursorColor, 21, 0x1F, "\x1f")
	}

	EditorDrawRefresh := func() {
		BoardDrawBorder()
		EditorDrawSidebar()
		TransitionDrawToBoard()
		if Length(Board.Name) != 0 {
			VideoWriteText((59-Length(Board.Name))/2, 0, 0x70, " "+Board.Name+" ")
		} else {
			VideoWriteText(26, 0, 0x70, " Untitled ")
		}
	}

	EditorSetAndCopyTile := func(x, y int16, element, color byte) {
		Board.Tiles.Set(x, y, TTile{Element: element, Color: color})
		copiedTile = Board.Tiles.Get(x, y)
		copiedHasStat = false
		copiedX = int16(x)
		copiedY = int16(y)
		EditorDrawTileAndNeighborsAt(int16(x), int16(y))
	}

	EditorAskSaveChanged := func() {
		InputKeyPressed = '\x00'
		if wasModified {
			if SidebarPromptYesNo("Save first? ", true) {
				if InputKeyPressed != KEY_ESCAPE {
					GameWorldSave("Save world", &LoadedGameFileName, ".ZZT")
				}
			}
		}
		World.Info.Name = LoadedGameFileName
	}

	EditorPrepareModifyTile := func(x, y int16) (EditorPrepareModifyTile bool) {
		wasModified = true
		EditorPrepareModifyTile = BoardPrepareTileForPlacement(x, y)
		EditorDrawTileAndNeighborsAt(x, y)
		return
	}

	EditorPrepareModifyStatAtCursor := func() (EditorPrepareModifyStatAtCursor bool) {
		if Board.Stats.Count < MAX_STAT {
			EditorPrepareModifyStatAtCursor = EditorPrepareModifyTile(cursorX, cursorY)
		} else {
			EditorPrepareModifyStatAtCursor = false
		}
		return
	}

	EditorPlaceTile := func(x, y int16) {
		Board.Tiles.With(x, y, func(tile *TTile) {
			if cursorPattern <= EditorPatternCount {
				if EditorPrepareModifyTile(x, y) {
					tile.Element = EditorPatterns[cursorPattern-1]
					tile.Color = byte(cursorColor)
				}
			} else if copiedHasStat {
				if EditorPrepareModifyStatAtCursor() {
					AddStat(x, y, copiedTile.Element, int16(copiedTile.Color), copiedStat.Cycle, copiedStat)
				}
			} else {
				if EditorPrepareModifyTile(x, y) {
					*tile = copiedTile
				}
			}

			EditorDrawTileAndNeighborsAt(x, y)
		})
	}

	EditorEditBoardInfo := func() {
		var (
			state         TTextWindowState
			numStr        string
			exitRequested bool
		)
		BoolToString := func(val bool) (BoolToString string) {
			if val {
				BoolToString = "Yes"
			} else {
				BoolToString = "No "
			}
			return
		}

		state.Title = "Board Information"
		state.DrawOpen()
		state.LinePos = 1
		state.Selectable = true
		exitRequested = false
		for {
			state.Selectable = true
			state.Lines = make([]string, 0)
			titleLine := state.Append("         Title: " + Board.Name)
			canFireLine := state.Append("      Can fire: " + Str(Board.Info.MaxShots) + " shots.")
			isDarkLine := state.Append(" Board is dark: " + BoolToString(Board.Info.IsDark))
			neighborBoardsLine := len(state.Lines) + 1
			for i := 0; i < 4; i++ {
				state.Append(NeighborBoardStrs[i] + ": " + EditorGetBoardName(int16(Board.Info.NeighborBoards[i]), true))
			}
			reEnterWhenZappedLine := state.Append("Re-enter when zapped: " + BoolToString(Board.Info.ReenterWhenZapped))
			timeLimitLine := state.Append("  Time limit, 0=None: " + Str(Board.Info.TimeLimitSec) + " sec.")
			quitLine := state.Append("          Quit!")
			state.Select(false, false)
			if InputKeyPressed == KEY_ENTER && state.LinePos >= 1 && state.LinePos <= 8 {
				wasModified = true
			}
			if InputKeyPressed == KEY_ENTER {
				switch state.LinePos {
				case titleLine:
					PopupPromptString("New title for board:", &Board.Name)
					exitRequested = true
					state.DrawClose()
				case canFireLine:
					numStr = Str(int16(Board.Info.MaxShots))
					SidebarPromptString("Maximum shots?", "", &numStr, PROMPT_NUMERIC)
					if Length(numStr) != 0 {
						Board.Info.MaxShots = byte(Val(numStr))
					}
					EditorDrawSidebar()
				case isDarkLine:
					Board.Info.IsDark = !Board.Info.IsDark
				case neighborBoardsLine, neighborBoardsLine + 1, neighborBoardsLine + 2, neighborBoardsLine + 3:
					Board.Info.NeighborBoards[state.LinePos-neighborBoardsLine] = byte(EditorSelectBoard(NeighborBoardStrs[state.LinePos-neighborBoardsLine], int16(Board.Info.NeighborBoards[state.LinePos-neighborBoardsLine]), true))
					if int16(Board.Info.NeighborBoards[state.LinePos-neighborBoardsLine]) > World.BoardCount {
						EditorAppendBoard()
						exitRequested = true
					}
				case reEnterWhenZappedLine:
					Board.Info.ReenterWhenZapped = !Board.Info.ReenterWhenZapped
				case timeLimitLine:
					numStr = Str(Board.Info.TimeLimitSec)
					SidebarPromptString("Time limit?", " Sec", &numStr, PROMPT_NUMERIC)
					if Length(numStr) != 0 {
						Board.Info.TimeLimitSec = int16(Val(numStr))
					}
					EditorDrawSidebar()
				case quitLine:
					exitRequested = true
					state.DrawClose()
				}
			} else {
				exitRequested = true
				state.DrawClose()
			}
			if exitRequested {
				break
			}
		}
	}

	EditorEditStatText := func(statId int16, prompt string) {
		var state TTextWindowState
		stat := Board.Stats.At(statId)
		state.Title = prompt
		state.DrawOpen()
		state.Selectable = false
		CopyStatDataToTextWindow(statId, &state)
		stat.DataLen = 0
		EditorOpenEditTextWindow(&state)
		data := make([]byte, 0)
		for iLine := 1; iLine <= len(state.Lines); iLine++ {
			data = append(data, state.Lines[iLine-1]...)
			data = append(data, '\r')
		}
		stat.Data = &data
		stat.DataLen = int16(len(data))
		state.DrawClose()
		InputKeyPressed = '\x00'
	}

	EditorEditStat := func(statId int16) {
		var (
			element       byte
			i             int16
			categoryName  string
			selectedBoard byte
			iy            int16
			promptByte    byte
		)
		EditorEditStatSettings := func(selected bool) {
			stat := Board.Stats.At(statId)
			InputKeyPressed = '\x00'
			iy = 9
			if Length(ElementDefs[element].Param1Name) != 0 {
				if Length(ElementDefs[element].ParamTextName) == 0 {
					SidebarPromptSlider(selected, 63, iy, ElementDefs[element].Param1Name, &stat.P1)
				} else {
					if stat.P1 == 0 {
						stat.P1 = World.EditorStatSettings[element].P1
					}
					BoardDrawTile(int16(stat.X), int16(stat.Y))
					SidebarPromptCharacter(selected, 63, iy, ElementDefs[element].Param1Name, &stat.P1)
					BoardDrawTile(int16(stat.X), int16(stat.Y))
				}
				if selected {
					World.EditorStatSettings[element].P1 = stat.P1
				}
				iy += 4
			}
			if InputKeyPressed != KEY_ESCAPE && Length(ElementDefs[element].ParamTextName) != 0 {
				if selected {
					EditorEditStatText(statId, ElementDefs[element].ParamTextName)
				}
			}
			if InputKeyPressed != KEY_ESCAPE && Length(ElementDefs[element].Param2Name) != 0 {
				promptByte = byte(int16(stat.P2) % 0x80)
				SidebarPromptSlider(selected, 63, iy, ElementDefs[element].Param2Name, &promptByte)
				if selected {
					stat.P2 = byte(int16(stat.P2)&0x80 + int16(promptByte))
					World.EditorStatSettings[element].P2 = stat.P2
				}
				iy += 4
			}
			if InputKeyPressed != KEY_ESCAPE && Length(ElementDefs[element].ParamBulletTypeName) != 0 {
				promptByte = byte(int16(stat.P2) / 0x80)
				SidebarPromptChoice(selected, iy, ElementDefs[element].ParamBulletTypeName, "Bullets Stars", &promptByte)
				if selected {
					stat.P2 = byte(int16(stat.P2)%0x80 + int16(promptByte)*0x80)
					World.EditorStatSettings[element].P2 = stat.P2
				}
				iy += 4
			}
			if InputKeyPressed != KEY_ESCAPE && Length(ElementDefs[element].ParamDirName) != 0 {
				SidebarPromptDirection(selected, iy, ElementDefs[element].ParamDirName, &stat.StepX, &stat.StepY)
				if selected {
					World.EditorStatSettings[element].StepX = stat.StepX
					World.EditorStatSettings[element].StepY = stat.StepY
				}
				iy += 4
			}
			if InputKeyPressed != KEY_ESCAPE && Length(ElementDefs[element].ParamBoardName) != 0 {
				if selected {
					selectedBoard = byte(EditorSelectBoard(ElementDefs[element].ParamBoardName, int16(stat.P3), true))
					if selectedBoard != 0 {
						stat.P3 = selectedBoard
						World.EditorStatSettings[element].P3 = byte(World.Info.CurrentBoard)
						if int16(stat.P3) > World.BoardCount {
							EditorAppendBoard()
							copiedHasStat = false
							copiedTile.Element = 0
							copiedTile.Color = 0x0F
						}
						World.EditorStatSettings[element].P3 = stat.P3
					} else {
						InputKeyPressed = KEY_ESCAPE
					}
					iy += 4
				} else {
					VideoWriteText(63, iy, 0x1F, "Room: "+Copy(EditorGetBoardName(int16(stat.P3), true), 1, 10))
				}
			}
		}

		stat := Board.Stats.At(statId)
		SidebarClear()
		element = Board.Tiles.Get(int16(stat.X), int16(stat.Y)).Element
		wasModified = true
		categoryName = ""
		for i = 0; i <= int16(element); i++ {
			if ElementDefs[i].EditorCategory == ElementDefs[element].EditorCategory && Length(ElementDefs[i].CategoryName) != 0 {
				categoryName = ElementDefs[i].CategoryName
			}
		}
		VideoWriteText(64, 6, 0x1E, categoryName)
		VideoWriteText(64, 7, 0x1F, ElementDefs[element].Name)
		EditorEditStatSettings(false)
		EditorEditStatSettings(true)
		if InputKeyPressed != KEY_ESCAPE {
			copiedHasStat = true
			copiedStat = *Board.Stats.At(statId)
			copiedTile = Board.Tiles.Get(int16(stat.X), int16(stat.Y))
			copiedX = int16(stat.X)
			copiedY = int16(stat.Y)
		}
	}

	EditorTransferBoard := func() {
		var i byte
		i = 1
		SidebarPromptChoice(true, 3, "Transfer board:", "Import Export", &i)
		if InputKeyPressed != KEY_ESCAPE {
			if i == 0 {
				SidebarPromptString("Import board", ".BRD", &SavedBoardFileName, PROMPT_ALPHANUM)
				if InputKeyPressed != KEY_ESCAPE && Length(SavedBoardFileName) != 0 {
					f, err := VfsOpen(SavedBoardFileName + ".BRD")
					if err != nil {
						DisplayIOError(err)
						goto TransferEnd
					}
					defer f.Close()
					BoardClose()
					var boardLen uint16
					err = ReadPUShort(f, &boardLen)
					if err != nil && boardLen > 0 {
						data := make([]byte, boardLen)
						_, err = f.Read(data)
						if err != nil {
							DisplayIOError(err)
							goto TransferEnd
						}
						World.BoardData[World.Info.CurrentBoard] = data
					}
					BoardOpen(World.Info.CurrentBoard)
					EditorDrawRefresh()
					for i = 0; i <= 3; i++ {
						Board.Info.NeighborBoards[i] = 0
					}
				}
			} else if i == 1 {
				SidebarPromptString("Export board", ".BRD", &SavedBoardFileName, PROMPT_ALPHANUM)
				if InputKeyPressed != KEY_ESCAPE && Length(SavedBoardFileName) != 0 {
					f, err := VfsCreate(SavedBoardFileName + ".BRD")
					if err != nil {
						DisplayIOError(err)
						goto TransferEnd
					}
					defer f.Close()
					BoardClose()
					WritePShort(f, int16(len(World.BoardData[World.Info.CurrentBoard])))
					_, err = f.Write(World.BoardData[World.Info.CurrentBoard])
					BoardOpen(World.Info.CurrentBoard)
					if err != nil {
						DisplayIOError(err)
					}
				}
			}
		}
	TransferEnd:
		EditorDrawSidebar()
	}

	EditorFloodFill := func(x, y int16, from TTile) {
		var (
			i              int16
			tileAt         TTile
			toFill, filled byte
			xPosition      [256]int16
			yPosition      [256]int16
		)
		toFill = 1
		filled = 0
		for toFill != filled {
			tileAt = Board.Tiles.Get(x, y)
			EditorPlaceTile(x, y)
			if Board.Tiles.Get(x, y).Element != tileAt.Element || Board.Tiles.Get(x, y).Color != tileAt.Color {
				for i = 0; i <= 3; i++ {
					tile := Board.Tiles.Get(x+NeighborDeltaX[i], y+NeighborDeltaY[i])
					if tile.Element == from.Element && (from.Element == 0 || tile.Color == from.Color) {
						xPosition[toFill] = x + NeighborDeltaX[i]
						yPosition[toFill] = y + NeighborDeltaY[i]
						toFill++
					}
				}
			}
			filled++
			x = xPosition[filled]
			y = yPosition[filled]
		}
	}

	if World.Info.IsSave || WorldGetFlagPosition("SECRET") >= 0 {
		WorldUnload()
		WorldCreate()
	}
	InitElementsEditor()
	CurrentTick = 0
	wasModified = false
	cursorX = 30
	cursorY = 12
	drawMode = DrawingOff
	cursorPattern = 1
	cursorColor = 0x0E
	cursorBlinker = 0
	copiedHasStat = false
	copiedTile.Element = 0
	copiedTile.Color = 0x0F
	if World.Info.CurrentBoard != 0 {
		BoardChange(World.Info.CurrentBoard)
	}
	EditorDrawRefresh()
	if World.BoardCount == 0 {
		EditorAppendBoard()
	}
	editorExitRequested = false
	for {
		if drawMode == DrawingOn {
			EditorPlaceTile(cursorX, cursorY)
		}
		Idle(IdleUntilFrame)
		InputUpdate()
		if InputKeyPressed == '\x00' && InputDeltaX == 0 && InputDeltaY == 0 && !InputShiftPressed {
			if SoundHasTimeElapsed(&TickTimeCounter, 15) {
				cursorBlinker = (cursorBlinker + 1) % 3
			}
			if cursorBlinker == 0 {
				BoardDrawTile(cursorX, cursorY)
			} else {
				VideoWriteText(cursorX-1, cursorY-1, 0x0F, "\xc5")
			}
			EditorUpdateSidebar()
		} else {
			BoardDrawTile(cursorX, cursorY)
		}
		if drawMode == TextEntry {
			if InputKeyPressed >= ' ' && InputKeyPressed < '\x80' {
				if EditorPrepareModifyTile(cursorX, cursorY) {
					Board.Tiles.Set(cursorX, cursorY, TTile{Element: byte(cursorColor - 9 + E_TEXT_MIN), Color: byte(InputKeyPressed)})
					EditorDrawTileAndNeighborsAt(cursorX, cursorY)
					InputDeltaX = 1
					InputDeltaY = 0
				}
				InputKeyPressed = '\x00'
			} else if InputKeyPressed == KEY_BACKSPACE && cursorX > 1 && EditorPrepareModifyTile(cursorX-1, cursorY) {
				cursorX--
			} else if InputKeyPressed == KEY_ENTER || InputKeyPressed == KEY_ESCAPE {
				drawMode = DrawingOff
				InputKeyPressed = '\x00'
			}
		}
		tile := Board.Tiles.Pointer(cursorX, cursorY)
		if InputShiftPressed || InputKeyPressed == ' ' {
			InputShiftAccepted = true
			if tile.Element == 0 || ElementDefs[tile.Element].PlaceableOnTop && copiedHasStat && cursorPattern > EditorPatternCount || InputDeltaX != 0 || InputDeltaY != 0 {
				EditorPlaceTile(cursorX, cursorY)
			} else {
				canModify = EditorPrepareModifyTile(cursorX, cursorY)
				if canModify {
					Board.Tiles.SetElement(cursorX, cursorY, E_EMPTY)
				}
			}
		}
		if InputDeltaX != 0 || InputDeltaY != 0 {
			cursorX += InputDeltaX
			if cursorX < 1 {
				cursorX = 1
			}
			if cursorX > BOARD_WIDTH {
				cursorX = BOARD_WIDTH
			}
			cursorY += InputDeltaY
			if cursorY < 1 {
				cursorY = 1
			}
			if cursorY > BOARD_HEIGHT {
				cursorY = BOARD_HEIGHT
			}
			VideoWriteText(cursorX-1, cursorY-1, 0x0F, "\xc5")
			if InputKeyPressed == '\x00' && InputJoystickEnabled {
				Delay(70)
			}
			InputShiftAccepted = false
		}
		switch UpCase(InputKeyPressed) {
		case '`':
			EditorDrawRefresh()
		case 'P':
			VideoWriteText(62, 21, 0x1F, "       ")
			if cursorPattern <= EditorPatternCount {
				cursorPattern++
			} else {
				cursorPattern = 1
			}
		case 'C':
			VideoWriteText(72, 19, 0x1E, "       ")
			VideoWriteText(69, 21, 0x1F, "        ")
			if cursorColor%0x10 != 0x0F {
				cursorColor++
			} else {
				cursorColor = cursorColor/0x10*0x10 + 9
			}
		case 'L':
			EditorAskSaveChanged()
			if InputKeyPressed != KEY_ESCAPE && GameWorldLoad(".ZZT") {
				if World.Info.IsSave || WorldGetFlagPosition("SECRET") >= 0 {
					if !DebugEnabled {
						SidebarClearLine(3)
						SidebarClearLine(4)
						SidebarClearLine(5)
						VideoWriteText(63, 4, 0x1E, "Can not edit")
						if World.Info.IsSave {
							VideoWriteText(63, 5, 0x1E, "a saved game!")
						} else {
							VideoWriteText(63, 5, 0x1E, "  "+World.Info.Name+"!")
						}
						PauseOnError()
						WorldUnload()
						WorldCreate()
					}
				}
				wasModified = false
				EditorDrawRefresh()
			}
			EditorDrawSidebar()
		case 'S':
			GameWorldSave("Save world:", &LoadedGameFileName, ".ZZT")
			if InputKeyPressed != KEY_ESCAPE {
				wasModified = false
			}
			EditorDrawSidebar()
		case 'Z':
			if SidebarPromptYesNo("Clear board? ", false) {
				for i = Board.Stats.Count; i >= 1; i-- {
					RemoveStat(i)
				}
				BoardCreate()
				EditorDrawRefresh()
			} else {
				EditorDrawSidebar()
			}
		case 'N':
			if SidebarPromptYesNo("Make new world? ", false) && InputKeyPressed != KEY_ESCAPE {
				EditorAskSaveChanged()
				if InputKeyPressed != KEY_ESCAPE {
					WorldUnload()
					WorldCreate()
					EditorDrawRefresh()
					wasModified = false
				}
			}
			EditorDrawSidebar()
		case 'Q', KEY_ESCAPE:
			editorExitRequested = true
		case 'B':
			i = EditorSelectBoard("Switch boards", World.Info.CurrentBoard, false)
			if InputKeyPressed != KEY_ESCAPE {
				if i > World.BoardCount {
					if SidebarPromptYesNo("Add new board? ", false) {
						EditorAppendBoard()
					}
				}
				BoardChange(i)
				EditorDrawRefresh()
			}
			EditorDrawSidebar()
		case '?':
			GameDebugPrompt()
			EditorDrawSidebar()
		case KEY_TAB:
			if drawMode == DrawingOff {
				drawMode = DrawingOn
			} else {
				drawMode = DrawingOff
			}
		case KEY_F1, KEY_F2, KEY_F3:
			VideoWriteText(cursorX-1, cursorY-1, 0x0F, "\xc5")
			for i = 3; i <= 20; i++ {
				SidebarClearLine(i)
			}
			switch InputKeyPressed {
			case KEY_F1:
				selectedCategory = CATEGORY_ITEM
			case KEY_F2:
				selectedCategory = CATEGORY_CREATURE
			case KEY_F3:
				selectedCategory = CATEGORY_TERRAIN
			}
			i = 3
			for iElem = 0; iElem <= MAX_ELEMENT; iElem++ {
				if ElementDefs[iElem].EditorCategory == selectedCategory {
					if Length(ElementDefs[iElem].CategoryName) != 0 {
						i++
						VideoWriteText(65, i, 0x1E, ElementDefs[iElem].CategoryName)
						i++
					}
					VideoWriteText(61, i, byte(i%2<<6+0x30), " "+Chr(ElementDefs[iElem].EditorShortcut)+" ")
					VideoWriteText(65, i, 0x1F, ElementDefs[iElem].Name)
					if ElementDefs[iElem].Color == COLOR_CHOICE_ON_BLACK {
						elemMenuColor = cursorColor%0x10 + 0x10
					} else if ElementDefs[iElem].Color == COLOR_WHITE_ON_CHOICE {
						elemMenuColor = cursorColor*0x10 - 0x71
					} else if ElementDefs[iElem].Color == COLOR_CHOICE_ON_CHOICE {
						elemMenuColor = (cursorColor-8)*0x11 + 8
					} else if int16(ElementDefs[iElem].Color)&0x70 == 0x00 {
						elemMenuColor = int16(ElementDefs[iElem].Color)%0x10 + 0x10
					} else {
						elemMenuColor = int16(ElementDefs[iElem].Color)
					}

					VideoWriteText(78, i, byte(elemMenuColor), Chr(ElementDefs[iElem].Character))
					i++
				}
			}
			InputReadWaitKey()
			for iElem = 1; iElem <= MAX_ELEMENT; iElem++ {
				if ElementDefs[iElem].EditorCategory == selectedCategory && ElementDefs[iElem].EditorShortcut == byte(UpCase(InputKeyPressed)) {
					if iElem == E_PLAYER {
						if EditorPrepareModifyTile(cursorX, cursorY) {
							MoveStat(0, cursorX, cursorY)
						}
					} else {
						if ElementDefs[iElem].Color == COLOR_CHOICE_ON_BLACK {
							elemMenuColor = cursorColor
						} else if ElementDefs[iElem].Color == COLOR_WHITE_ON_CHOICE {
							elemMenuColor = cursorColor*0x10 - 0x71
						} else if ElementDefs[iElem].Color == COLOR_CHOICE_ON_CHOICE {
							elemMenuColor = (cursorColor-8)*0x11 + 8
						} else {
							elemMenuColor = int16(ElementDefs[iElem].Color)
						}

						if ElementDefs[iElem].Cycle == -1 {
							if EditorPrepareModifyTile(cursorX, cursorY) {
								EditorSetAndCopyTile(cursorX, cursorY, byte(iElem), byte(elemMenuColor))
							}
						} else {
							if EditorPrepareModifyStatAtCursor() {
								AddStat(cursorX, cursorY, byte(iElem), elemMenuColor, ElementDefs[iElem].Cycle, StatTemplateDefault)
								stat := Board.Stats.At(Board.Stats.Count)
								if Length(ElementDefs[iElem].Param1Name) != 0 {
									stat.P1 = World.EditorStatSettings[iElem].P1
								}
								if Length(ElementDefs[iElem].Param2Name) != 0 {
									stat.P2 = World.EditorStatSettings[iElem].P2
								}
								if Length(ElementDefs[iElem].ParamDirName) != 0 {
									stat.StepX = World.EditorStatSettings[iElem].StepX
									stat.StepY = World.EditorStatSettings[iElem].StepY
								}
								if Length(ElementDefs[iElem].ParamBoardName) != 0 {
									stat.P3 = World.EditorStatSettings[iElem].P3
								}
								EditorEditStat(Board.Stats.Count)
								if InputKeyPressed == KEY_ESCAPE {
									RemoveStat(Board.Stats.Count)
								}
							}
						}
					}
				}
			}
			EditorDrawSidebar()
		case KEY_F4:
			if drawMode != TextEntry {
				drawMode = TextEntry
			} else {
				drawMode = DrawingOff
			}
		case 'H':
			TextWindowDisplayFile("editor.hlp", "World editor help")
		case 'X':
			EditorFloodFill(cursorX, cursorY, Board.Tiles.Get(cursorX, cursorY))
		case '!':
			EditorEditHelpFile()
			EditorDrawSidebar()
		case 'T':
			EditorTransferBoard()
		case KEY_ENTER:
			if GetStatIdAt(cursorX, cursorY) >= 0 {
				EditorEditStat(GetStatIdAt(cursorX, cursorY))
				EditorDrawSidebar()
			} else {
				copiedHasStat = false
				copiedTile = Board.Tiles.Get(cursorX, cursorY)
			}
		case 'I':
			EditorEditBoardInfo()
			TransitionDrawToBoard()
		}
		if editorExitRequested {
			EditorAskSaveChanged()
			if InputKeyPressed == KEY_ESCAPE {
				editorExitRequested = false
				EditorDrawSidebar()
			}
		}
		if editorExitRequested {
			break
		}
	}
	InputKeyPressed = '\x00'
	InitElementsGame()
}

func EditorOpenEditTextWindow(state *TTextWindowState) {
	SidebarClear()
	VideoWriteText(61, 4, 0x30, " Return ")
	VideoWriteText(64, 5, 0x1F, " Insert line")
	VideoWriteText(61, 7, 0x70, " Ctrl-Y ")
	VideoWriteText(64, 8, 0x1F, " Delete line")
	VideoWriteText(61, 10, 0x30, " Cursor keys ")
	VideoWriteText(64, 11, 0x1F, " Move cursor")
	VideoWriteText(61, 13, 0x70, " Insert ")
	VideoWriteText(64, 14, 0x1F, " Insert mode: ")
	VideoWriteText(61, 16, 0x30, " Delete ")
	VideoWriteText(64, 17, 0x1F, " Delete char")
	VideoWriteText(61, 19, 0x70, " Escape ")
	VideoWriteText(64, 20, 0x1F, " Exit editor")
	state.Edit()
}

func EditorEditHelpFile() {
	var (
		textWindow TTextWindowState
		filename   string
	)
	filename = ""
	SidebarPromptString("File to edit", ".HLP", &filename, PROMPT_ALPHANUM)
	if Length(filename) != 0 {
		textWindow.OpenFile("*" + filename + ".HLP")
		textWindow.Title = "Editing " + filename
		textWindow.DrawOpen()
		EditorOpenEditTextWindow(&textWindow)
		textWindow.SaveFile(filename + ".HLP")
		textWindow.DrawClose()
	}
}

func EditorGetBoardName(boardId int16, titleScreenIsNone bool) (EditorGetBoardName string) {
	var (
		copiedName string
	)
	if boardId == 0 && titleScreenIsNone {
		EditorGetBoardName = "None"
	} else if boardId == World.Info.CurrentBoard {
		EditorGetBoardName = Board.Name
	} else {
		boardData := World.BoardData[boardId]
		r := bytes.NewReader(boardData)
		ReadPString(r, &copiedName, BOARD_NAME_LENGTH)
		EditorGetBoardName = copiedName
	}

	return
}

func EditorSelectBoard(title string, currentBoard int16, titleScreenIsNone bool) (EditorSelectBoard int16) {
	var (
		i          int16
		textWindow TTextWindowState
	)
	textWindow.Init()
	textWindow.Title = title
	textWindow.LinePos = int(currentBoard + 1)
	textWindow.Selectable = true
	for i = 0; i <= World.BoardCount; i++ {
		textWindow.Append(EditorGetBoardName(i, titleScreenIsNone))
	}
	textWindow.Append("Add new board")
	textWindow.DrawOpen()
	textWindow.Select(false, false)
	textWindow.DrawClose()
	if InputKeyPressed == KEY_ESCAPE {
		EditorSelectBoard = 0
	} else {
		EditorSelectBoard = int16(textWindow.LinePos - 1)
	}
	return
}
