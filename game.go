package main // unit: Game

import (
	"bufio"
	"bytes"
	"io"
	"path/filepath"
	"strings"

	"github.com/OpenZoo/openzoo-go/format"
)

const LineChars string = "\xf9\xd0Һ\xb5\xbc\xbb\xb9\xc6\xc8\xc9\xcc\xcd\xca\xcb\xce"

var (
	ProgressAnimColors  [8]byte   = [8]byte{0x14, 0x1C, 0x15, 0x1D, 0x16, 0x1E, 0x17, 0x1F}
	ProgressAnimStrings [8]string = [8]string{"....|", "...*/", "..*.-", ".*..\\", "*...|", "..../", "....-", "....\\"}
	ColorNames          [7]string = [7]string{"Blue", "Green", "Cyan", "Red", "Purple", "Yellow", "White"}
	DiagonalDeltaX      [8]int16  = [8]int16{-1, 0, 1, 1, 1, 0, -1, -1}
	DiagonalDeltaY      [8]int16  = [8]int16{1, 1, 1, 0, -1, -1, -1, 0}
	NeighborDeltaX      [4]int16  = [4]int16{0, 0, -1, 1}
	NeighborDeltaY      [4]int16  = [4]int16{-1, 1, 0, 0}
	TileBorder          TTile     = TTile{Element: E_NORMAL, Color: 0x0E}
	TileBoardEdge       TTile     = TTile{Element: E_BOARD_EDGE, Color: 0x00}
	StatTemplateDefault TStat     = TStat{X: 0, Y: 0, StepX: 0, StepY: 0, Cycle: 0, P1: 0, P2: 0, P3: 0, Follower: -1, Leader: -1}
)

// implementation uses: Dos, Crt, Video, Sounds, Input, Elements, Editor, Oop

func GenerateTransitionTable() {
	var i, ix, iy int16

	i = 0
	TransitionTable = make([]TCoord, BOARD_WIDTH*BOARD_HEIGHT)
	for iy = 1; iy <= BOARD_HEIGHT; iy++ {
		for ix = 1; ix <= BOARD_WIDTH; ix++ {
			TransitionTable[i] = TCoord{X: ix, Y: iy}
			i++
		}
	}
	for ix = 0; ix < int16(len(TransitionTable)); ix++ {
		iy = Random(int16(len(TransitionTable)))
		t := TransitionTable[iy]
		TransitionTable[iy] = TransitionTable[ix]
		TransitionTable[ix] = t
	}
}

func BoardClose() {
	var buf bytes.Buffer
	w := bufio.NewWriter(&buf)
	format.BoardSerialize(&Board, w)
	w.Flush()
	World.BoardData[World.Info.CurrentBoard] = buf.Bytes()
}

func BoardOpen(boardId int16) {
	if int(boardId) >= len(World.BoardData) {
		boardId = World.Info.CurrentBoard
	}
	r := bytes.NewReader(World.BoardData[boardId])
	format.BoardDeserialize(&Board, r)
	World.Info.CurrentBoard = boardId
}

func BoardChange(boardId int16) {
	Board.Tiles.Set(int16(Board.Stats.At(0).X), int16(Board.Stats.At(0).Y), format.TTile{Element: E_PLAYER, Color: ElementDefs[E_PLAYER].Color})
	BoardClose()
	BoardOpen(boardId)
}

func BoardCreate() {
	var ix, iy, i int16
	Board.Tiles = format.NewTileStorage(BOARD_WIDTH, BOARD_HEIGHT)
	Board.Stats = format.NewStatStorage(MAX_STAT)
	Board.Name = ""
	Board.Info.Message = ""
	Board.Info.MaxShots = 255
	Board.Info.IsDark = false
	Board.Info.ReenterWhenZapped = false
	Board.Info.TimeLimitSec = 0
	for i = 0; i <= 3; i++ {
		Board.Info.NeighborBoards[i] = 0
	}
	for ix = 0; ix <= BOARD_WIDTH+1; ix++ {
		Board.Tiles.Set(ix, 0, TileBoardEdge)
		Board.Tiles.Set(ix, BOARD_HEIGHT+1, TileBoardEdge)
	}
	for iy = 0; iy <= BOARD_HEIGHT+1; iy++ {
		Board.Tiles.Set(0, iy, TileBoardEdge)
		Board.Tiles.Set(BOARD_WIDTH+1, iy, TileBoardEdge)
	}
	for ix = 1; ix <= BOARD_WIDTH; ix++ {
		for iy = 1; iy <= BOARD_HEIGHT; iy++ {
			Board.Tiles.Set(ix, iy, TTile{Element: E_EMPTY, Color: 0})
		}
	}
	for ix = 1; ix <= BOARD_WIDTH; ix++ {
		Board.Tiles.Set(ix, 1, TileBorder)
		Board.Tiles.Set(ix, BOARD_HEIGHT, TileBorder)
	}
	for iy = 1; iy <= BOARD_HEIGHT; iy++ {
		Board.Tiles.Set(1, iy, TileBorder)
		Board.Tiles.Set(BOARD_WIDTH, iy, TileBorder)
	}
	Board.Tiles.Set(BOARD_WIDTH/2, BOARD_HEIGHT/2, TTile{Element: E_PLAYER, Color: ElementDefs[E_PLAYER].Color})
	Board.Stats.Count = 0
	Board.Stats.At(0).X = BOARD_WIDTH / 2
	Board.Stats.At(0).Y = BOARD_HEIGHT / 2
	Board.Stats.At(0).Cycle = 1
	Board.Stats.At(0).Under.Element = E_EMPTY
	Board.Stats.At(0).Under.Color = 0
	Board.Stats.At(0).Data = nil
	Board.Stats.At(0).DataLen = 0
}

func WorldCreate() {
	InitElementsGame()
	World.BoardData = make([][]byte, 1)
	World.BoardData[0] = make([]byte, 0)
	InitEditorStatSettings()
	ResetMessageNotShownFlags()
	BoardCreate()
	World.Info.IsSave = false
	World.Info.CurrentBoard = 0
	World.Info.Ammo = 0
	World.Info.Gems = 0
	World.Info.Health = 100
	World.Info.EnergizerTicks = 0
	World.Info.Torches = 0
	World.Info.TorchTicks = 0
	World.Info.Score = 0
	World.Info.BoardTimeSec = 0
	World.Info.BoardTimeHsec = 0
	World.Info.Flags = make([]string, 10)
	for i := 0; i < 7; i++ {
		World.Info.Keys[i] = false
	}
	for i := 0; i < len(World.Info.Flags); i++ {
		World.Info.Flags[i] = ""
	}
	BoardChange(0)
	Board.Name = "Title screen"
	LoadedGameFileName = ""
	World.Info.Name = ""
}

func TransitionDrawToFill(chr byte, color int16) {
	s := Chr(chr)
	for i := 0; i < len(TransitionTable); i++ {
		VideoWriteText(TransitionTable[i].X-1, TransitionTable[i].Y-1, byte(color), s)
	}
}

func BoardDrawTile(x, y int16) {
	if x < 1 || y < 1 || x > 80 || y > 25 {
		return
	}
	var ch byte
	tile := Board.Tiles.Get(x, y)
	if !Board.Info.IsDark || ElementDefs[tile.Element].VisibleInDark || World.Info.TorchTicks > 0 && Sqr(int16(Board.Stats.At(0).X)-x)+Sqr(int16(Board.Stats.At(0).Y)-y)*2 < TORCH_DIST_SQR || ForceDarknessOff {
		if tile.Element == E_EMPTY {
			VideoWriteText(x-1, y-1, 0x0F, " ")
		} else if ElementDefs[tile.Element].HasDrawProc {
			ElementDefs[tile.Element].DrawProc(x, y, &ch)
			VideoWriteText(x-1, y-1, tile.Color, Chr(ch))
		} else if tile.Element < E_TEXT_MIN {
			VideoWriteText(x-1, y-1, tile.Color, Chr(ElementDefs[tile.Element].Character))
		} else {
			if tile.Element == E_TEXT_WHITE {
				VideoWriteText(x-1, y-1, 0x0F, Chr(tile.Color))
			} else if VideoMonochrome {
				VideoWriteText(x-1, y-1, byte((int16(tile.Element-E_TEXT_MIN)+1)*16), Chr(tile.Color))
			} else {
				VideoWriteText(x-1, y-1, byte((int16(tile.Element-E_TEXT_MIN)+1)*16+0x0F), Chr(tile.Color))
			}

		}

	} else {
		VideoWriteText(x-1, y-1, 0x07, "\xb0")
	}
}

func BoardDrawBorder() {
	var ix, iy int16
	for ix = 1; ix <= BOARD_WIDTH; ix++ {
		BoardDrawTile(ix, 1)
		BoardDrawTile(ix, BOARD_HEIGHT)
	}
	for iy = 1; iy <= BOARD_HEIGHT; iy++ {
		BoardDrawTile(1, iy)
		BoardDrawTile(BOARD_WIDTH, iy)
	}
}

func TransitionDrawToBoard() {
	BoardDrawBorder()
	for i := 0; i < len(TransitionTable); i++ {
		BoardDrawTile(TransitionTable[i].X, TransitionTable[i].Y)
	}
}

func PauseOnError() {
	SoundQueue(1, SoundParse("s004x114x9"))
	Delay(2000)
}

func DisplayIOError(e error) bool {
	// stub: appropriately shorten length, etc
	textWindow := NewTextWindowState()
	textWindow.Title = e.Error()
	textWindow.Append("This may be caused by missing")
	textWindow.Append("ZZT files or a bad disk.  If")
	textWindow.Append("you are trying to save a game,")
	textWindow.Append("your disk may be full -- try")
	textWindow.Append("using a blank, formatted disk")
	textWindow.Append("for saving the game!")
	textWindow.DrawOpen()
	textWindow.Select(false, false)
	textWindow.DrawClose()
	return false
}

func WorldUnload() {
	BoardClose()
	for i := 0; i < len(World.BoardData); i++ {
		World.BoardData[i] = nil
	}
}

func WorldLoad(filename, extension string, flags format.WorldDeserializeFlag) bool {
	SidebarAnimateLoading := func(step, _ int) {
		VideoWriteText(69, 5, ProgressAnimColors[step&7], ProgressAnimStrings[step&7])
	}

	SidebarClearLine(4)
	SidebarClearLine(5)
	SidebarClearLine(5)
	VideoWriteText(62, 5, 0x1F, "Loading.....")

	f, err := VfsOpen(filename + extension)
	if err != nil {
		return DisplayIOError(err)
	}
	defer f.Close()

	WorldUnload()

	err = format.WorldDeserialize(f, &World, flags, SidebarAnimateLoading)
	if err == format.ErrWrongZZTVersion {
		VideoWriteText(63, 5, 0x1E, "You need a newer")
		VideoWriteText(63, 6, 0x1E, " version of ZZT!")
		return false
	} else if err != nil {
		return DisplayIOError(err)
	}

	BoardOpen(World.Info.CurrentBoard)
	LoadedGameFileName = filename
	HighScoresLoad()
	SidebarClearLine(5)

	return true
}

func WorldSave(filename, extension string) error {
	BoardClose()
	VideoWriteText(63, 5, 0x1F, "Saving...")
	f, err := VfsCreate(filename + extension)
	if err != nil {
		return err
	}
	defer f.Close()

	err = format.WorldSerialize(f, &World)

	BoardOpen(World.Info.CurrentBoard)
	SidebarClearLine(5)

	return err
}

func GameWorldSave(prompt string, filename *string, extension string) {
	newFilename := *filename
	SidebarPromptString(prompt, extension, &newFilename, PROMPT_ALPHANUM)
	if InputKeyPressed != KEY_ESCAPE && Length(newFilename) != 0 {
		*filename = newFilename
		if extension == ".ZZT" {
			World.Info.Name = *filename
		}
		if err := WorldSave(*filename, extension); err != nil {
			DisplayIOError(err)
		}
	}
}

func GameWorldLoad(extension string) (GameWorldLoad bool) {
	var (
		textWindow TTextWindowState
		entryName  string
	)
	textWindow.Init()
	if extension == ".ZZT" {
		textWindow.Title = "ZZT Worlds"
	} else {
		textWindow.Title = "Saved Games"
	}
	GameWorldLoad = false
	textWindow.Selectable = true
	dirs, err := VfsReadDir(".")
	if err != nil {
		DisplayIOError(err)
		return
	}
	for _, path := range dirs {
		if strings.EqualFold(filepath.Ext(path.Name()), extension) {
			entryName = PathBasenameWithoutExt(path.Name())
			if desc, ok := WorldFileDescs[entryName]; ok {
				entryName = desc
			}
			textWindow.Append(entryName)
		}
	}
	textWindow.Append("Exit")
	textWindow.DrawOpen()
	textWindow.Select(false, false)
	textWindow.DrawClose()
	if textWindow.LinePos < len(textWindow.Lines) && !TextWindowRejected {
		entryName = textWindow.Lines[textWindow.LinePos-1]
		if Pos(' ', entryName) != 0 {
			entryName = Copy(entryName, 1, Pos(' ', entryName)-1)
		}
		GameWorldLoad = WorldLoad(entryName, extension, 0)
		TransitionDrawToFill('\xdb', 0x44)
	}
	return
}

func CopyStatDataToTextWindow(statId int16, state *TTextWindowState) {
	var (
		dataStr strings.Builder
		dataPtr io.Reader
		dataChr byte
		i       int16
	)
	stat := Board.Stats.At(statId)
	state.Init()
	if stat.Data != nil {
		dataPtr = bytes.NewReader(*stat.Data)
		for i = 0; i < stat.DataLen; i++ {
			format.ReadPByte(dataPtr, &dataChr)
			if dataChr == KEY_ENTER {
				state.Append(dataStr.String())
				dataStr.Reset()
			} else {
				dataStr.WriteByte(dataChr)
			}
		}
	}
}

func AddStat(tx, ty int16, element byte, color, tcycle int16, template TStat) {
	if Board.Stats.Count < MAX_STAT {
		Board.Stats.Count++
		*Board.Stats.At(Board.Stats.Count) = template
		stat := Board.Stats.At(Board.Stats.Count)
		stat.X = byte(tx)
		stat.Y = byte(ty)
		stat.Cycle = tcycle
		stat.Under = Board.Tiles.Get(tx, ty)
		stat.DataPos = 0
		if template.Data != nil {
			copiedData := make([]byte, len(*template.Data))
			copy(copiedData, *template.Data)
			Board.Stats.At(Board.Stats.Count).Data = &copiedData
		}
		if ElementDefs[Board.Tiles.Get(tx, ty).Element].PlaceableOnTop {
			Board.Tiles.SetColor(tx, ty, byte(color&0x0F+int16(Board.Tiles.Get(tx, ty).Color)&0x70))
		} else {
			Board.Tiles.SetColor(tx, ty, byte(color))
		}
		Board.Tiles.SetElement(tx, ty, element)
		if ty > 0 {
			BoardDrawTile(tx, ty)
		}
	}
}

func RemoveStat(statId int16) {
	var i int16
	stat := Board.Stats.At(statId)
	if stat.DataLen != 0 {
		for i = 1; i <= Board.Stats.Count; i++ {
			if Board.Stats.At(i).Data == stat.Data && i != statId {
				goto StatDataInUse
			}
		}
		stat.Data = nil
	}
StatDataInUse:
	if statId < CurrentStatTicked {
		CurrentStatTicked--
	}

	Board.Tiles.Set(int16(stat.X), int16(stat.Y), stat.Under)
	if stat.Y > 0 {
		BoardDrawTile(int16(stat.X), int16(stat.Y))
	}
	for i = 1; i <= Board.Stats.Count; i++ {
		if Board.Stats.At(i).Follower >= statId {
			if Board.Stats.At(i).Follower == statId {
				Board.Stats.At(i).Follower = -1
			} else {
				Board.Stats.At(i).Follower--
			}
		}
		if Board.Stats.At(i).Leader >= statId {
			if Board.Stats.At(i).Leader == statId {
				Board.Stats.At(i).Leader = -1
			} else {
				Board.Stats.At(i).Leader--
			}
		}
	}
	for i = statId + 1; i <= Board.Stats.Count; i++ {
		*Board.Stats.At(i - 1) = *Board.Stats.At(i)
	}
	Board.Stats.Count--
}

func GetStatIdAt(x, y int16) (GetStatIdAt int16) {
	i := int16(-1)
	for {
		i++
		if int16(Board.Stats.At(i).X) == x && int16(Board.Stats.At(i).Y) == y || i > Board.Stats.Count {
			break
		}
	}
	if i > Board.Stats.Count {
		GetStatIdAt = -1
	} else {
		GetStatIdAt = i
	}
	return
}

func BoardPrepareTileForPlacement(x, y int16) (BoardPrepareTileForPlacement bool) {
	var (
		statId int16
		result bool
	)
	statId = GetStatIdAt(x, y)
	if statId > 0 {
		RemoveStat(statId)
		result = true
	} else if statId < 0 {
		if !ElementDefs[Board.Tiles.Get(x, y).Element].PlaceableOnTop {
			Board.Tiles.SetElement(x, y, E_EMPTY)
		}
		result = true
	} else {
		result = false
	}

	BoardDrawTile(x, y)
	BoardPrepareTileForPlacement = result
	return
}

func MoveStat(statId int16, newX, newY int16) {
	var (
		iUnder     TTile
		ix, iy     int16
		oldX, oldY int16
	)
	stat := Board.Stats.At(statId)
	iUnder = Board.Stats.At(statId).Under
	Board.Stats.At(statId).Under = Board.Tiles.Get(newX, newY)
	if Board.Tiles.Get(int16(stat.X), int16(stat.Y)).Element == E_PLAYER {
		Board.Tiles.SetColor(newX, newY, Board.Tiles.Get(int16(stat.X), int16(stat.Y)).Color)
	} else if Board.Tiles.Get(newX, newY).Element == E_EMPTY {
		Board.Tiles.SetColor(newX, newY, byte(int16(Board.Tiles.Get(int16(stat.X), int16(stat.Y)).Color)&0x0F))
	} else {
		Board.Tiles.SetColor(newX, newY, byte(int16(Board.Tiles.Get(int16(stat.X), int16(stat.Y)).Color)&0x0F+int16(Board.Tiles.Get(newX, newY).Color)&0x70))
	}

	Board.Tiles.SetElement(newX, newY, Board.Tiles.Get(int16(stat.X), int16(stat.Y)).Element)
	Board.Tiles.Set(int16(stat.X), int16(stat.Y), iUnder)
	oldX = int16(stat.X)
	oldY = int16(stat.Y)
	stat.X = byte(newX)
	stat.Y = byte(newY)
	BoardDrawTile(int16(stat.X), int16(stat.Y))
	BoardDrawTile(oldX, oldY)
	if statId == 0 && Board.Info.IsDark && World.Info.TorchTicks > 0 {
		if Sqr(oldX-int16(stat.X))+Sqr(oldY-int16(stat.Y)) == 1 {
			for ix = int16(stat.X) - TORCH_DX - 3; ix <= int16(stat.X)+TORCH_DX+3; ix++ {
				if ix >= 1 && ix <= BOARD_WIDTH {
					for iy = int16(stat.Y) - TORCH_DY - 3; iy <= int16(stat.Y)+TORCH_DY+3; iy++ {
						if iy >= 1 && iy <= BOARD_HEIGHT {
							if Sqr(ix-oldX)+Sqr(iy-oldY)*2 < TORCH_DIST_SQR != (Sqr(ix-newX)+Sqr(iy-newY)*2 < TORCH_DIST_SQR) {
								BoardDrawTile(ix, iy)
							}
						}
					}
				}
			}
		} else {
			DrawPlayerSurroundings(oldX, oldY, 0)
			DrawPlayerSurroundings(int16(stat.X), int16(stat.Y), 0)
		}
	}
}

func GameUpdateSidebar() {
	var (
		numStr string
		i      int16
	)
	if GameStateElement == E_PLAYER {
		if Board.Info.TimeLimitSec > 0 {
			VideoWriteText(64, 6, 0x1E, "   Time:")
			numStr = Str(Board.Info.TimeLimitSec - World.Info.BoardTimeSec)
			VideoWriteText(72, 6, 0x1E, numStr+" ")
		} else {
			SidebarClearLine(6)
		}
		if World.Info.Health < 0 {
			World.Info.Health = 0
		}
		numStr = Str(World.Info.Health)
		VideoWriteText(72, 7, 0x1E, numStr+" ")
		numStr = Str(World.Info.Ammo)
		VideoWriteText(72, 8, 0x1E, numStr+"  ")
		numStr = Str(World.Info.Torches)
		VideoWriteText(72, 9, 0x1E, numStr+" ")
		numStr = Str(World.Info.Gems)
		VideoWriteText(72, 10, 0x1E, numStr+" ")
		numStr = Str(World.Info.Score)
		VideoWriteText(72, 11, 0x1E, numStr+" ")
		if World.Info.TorchTicks == 0 {
			VideoWriteText(75, 9, 0x16, "    ")
		} else {
			for i = 2; i <= 5; i++ {
				if i <= World.Info.TorchTicks*5/TORCH_DURATION {
					VideoWriteText(73+i, 9, 0x16, "\xb1")
				} else {
					VideoWriteText(73+i, 9, 0x16, "\xb0")
				}
			}
		}
		for i = 1; i <= 7; i++ {
			if World.Info.Keys[i-1] {
				VideoWriteText(71+i, 12, byte(0x18+i), Chr(ElementDefs[E_KEY].Character))
			} else {
				VideoWriteText(71+i, 12, 0x1F, " ")
			}
		}
		if SoundEnabled {
			VideoWriteText(65, 15, 0x1F, " Be quiet")
		} else {
			VideoWriteText(65, 15, 0x1F, " Be noisy")
		}
		if DebugEnabled {
			numStr = Str(MemAvail())
			VideoWriteText(69, 4, 0x1E, "m"+numStr+" ")
		}
	}
}

func DisplayMessage(ticks int16, message string) {
	if GetStatIdAt(0, 0) != -1 {
		RemoveStat(GetStatIdAt(0, 0))
		BoardDrawBorder()
	}
	if Length(message) != 0 {
		AddStat(0, 0, E_MESSAGE_TIMER, 0, 1, StatTemplateDefault)
		Board.Stats.At(Board.Stats.Count).P2 = byte(ticks / (TickTimeDuration + 1))
		Board.Info.Message = message
	}
}

func DamageStat(attackerStatId int16) {
	var oldX, oldY int16
	stat := Board.Stats.At(attackerStatId)
	if attackerStatId == 0 {
		if World.Info.Health > 0 {
			World.Info.Health -= 10
			GameUpdateSidebar()
			DisplayMessage(100, "Ouch!")
			Board.Tiles.SetColor(int16(stat.X), int16(stat.Y), byte(0x70+int16(ElementDefs[E_PLAYER].Color)%0x10))
			if World.Info.Health > 0 {
				World.Info.BoardTimeSec = 0
				if Board.Info.ReenterWhenZapped {
					SoundQueue(4, " \x01#\x01'\x010\x01\x10\x01")
					Board.Tiles.SetElement(int16(stat.X), int16(stat.Y), E_EMPTY)
					BoardDrawTile(int16(stat.X), int16(stat.Y))
					oldX = int16(stat.X)
					oldY = int16(stat.Y)
					stat.X = Board.Info.StartPlayerX
					stat.Y = Board.Info.StartPlayerY
					DrawPlayerSurroundings(oldX, oldY, 0)
					DrawPlayerSurroundings(int16(stat.X), int16(stat.Y), 0)
					GamePaused = true
				}
				SoundQueue(4, "\x10\x01 \x01\x13\x01#\x01")
			} else {
				SoundQueue(5, " \x03#\x03'\x030\x03'\x03*\x032\x037\x035\x038\x03@\x03E\x03\x10\n")
			}
		}
	} else {
		switch Board.Tiles.Get(int16(stat.X), int16(stat.Y)).Element {
		case E_BULLET:
			SoundQueue(3, " \x01")
		case E_OBJECT:
		default:
			SoundQueue(3, "@\x01\x10\x01P\x010\x01")
		}
		RemoveStat(attackerStatId)
	}
}

func BoardDamageTile(x, y int16) {
	var statId int16
	statId = GetStatIdAt(x, y)
	if statId != -1 {
		DamageStat(statId)
	} else {
		Board.Tiles.SetElement(x, y, E_EMPTY)
		BoardDrawTile(x, y)
	}
}

func BoardAttack(attackerStatId int16, x, y int16) {
	if attackerStatId == 0 && World.Info.EnergizerTicks > 0 {
		World.Info.Score = ElementDefs[Board.Tiles.Get(x, y).Element].ScoreValue + World.Info.Score
		GameUpdateSidebar()
	} else {
		DamageStat(attackerStatId)
	}
	if attackerStatId > 0 && attackerStatId <= CurrentStatTicked {
		CurrentStatTicked--
	}
	if Board.Tiles.Get(x, y).Element == E_PLAYER && World.Info.EnergizerTicks > 0 {
		World.Info.Score = ElementDefs[Board.Tiles.Get(int16(Board.Stats.At(attackerStatId).X), int16(Board.Stats.At(attackerStatId).Y)).Element].ScoreValue + World.Info.Score
		GameUpdateSidebar()
	} else {
		BoardDamageTile(x, y)
		SoundQueue(2, "\x10\x01")
	}
}

func BoardShoot(element byte, tx, ty, deltaX, deltaY int16, source int16) (BoardShoot bool) {
	if ElementDefs[Board.Tiles.Get(tx+deltaX, ty+deltaY).Element].Walkable || Board.Tiles.Get(tx+deltaX, ty+deltaY).Element == E_WATER {
		AddStat(tx+deltaX, ty+deltaY, element, int16(ElementDefs[element].Color), 1, StatTemplateDefault)
		stat := Board.Stats.At(Board.Stats.Count)
		stat.P1 = byte(source)
		stat.StepX = deltaX
		stat.StepY = deltaY
		stat.P2 = 100
		BoardShoot = true
	} else if Board.Tiles.Get(tx+deltaX, ty+deltaY).Element == E_BREAKABLE || ElementDefs[Board.Tiles.Get(tx+deltaX, ty+deltaY).Element].Destructible && Board.Tiles.Get(tx+deltaX, ty+deltaY).Element == E_PLAYER == (source != 0) && World.Info.EnergizerTicks <= 0 {
		BoardDamageTile(tx+deltaX, ty+deltaY)
		SoundQueue(2, "\x10\x01")
		BoardShoot = true
	} else {
		BoardShoot = false
	}

	return
}

func CalcDirectionRnd(deltaX, deltaY *int16) {
	*deltaX = Random(3) - 1
	if *deltaX == 0 {
		*deltaY = Random(2)*2 - 1
	} else {
		*deltaY = 0
	}
}

func CalcDirectionSeek(x, y int16, deltaX, deltaY *int16) {
	*deltaX = 0
	*deltaY = 0
	if Random(2) < 1 || int16(Board.Stats.At(0).Y) == y {
		*deltaX = Signum(int16(Board.Stats.At(0).X) - x)
	}
	if *deltaX == 0 {
		*deltaY = Signum(int16(Board.Stats.At(0).Y) - y)
	}
	if World.Info.EnergizerTicks > 0 {
		*deltaX = -*deltaX
		*deltaY = -*deltaY
	}
}

func TransitionDrawBoardChange() {
	TransitionDrawToFill('\xdb', 0x05)
	TransitionDrawToBoard()
}

func BoardEnter() {
	Board.Info.StartPlayerX = Board.Stats.At(0).X
	Board.Info.StartPlayerY = Board.Stats.At(0).Y
	if Board.Info.IsDark && MessageHintTorchNotShown {
		DisplayMessage(200, "Room is dark - you need to light a torch!")
		MessageHintTorchNotShown = false
	}
	World.Info.BoardTimeSec = 0
	GameUpdateSidebar()
}

func BoardPassageTeleport(x, y int16) {
	var (
		// oldBoard   int16
		col        byte
		ix, iy     int16
		newX, newY int16
	)
	col = Board.Tiles.Get(x, y).Color
	// oldBoard = World.Info.CurrentBoard
	BoardChange(int16(Board.Stats.At(GetStatIdAt(x, y)).P3))
	newX = 0
	for ix = 1; ix <= BOARD_WIDTH; ix++ {
		for iy = 1; iy <= BOARD_HEIGHT; iy++ {
			if Board.Tiles.Get(ix, iy).Element == E_PASSAGE && Board.Tiles.Get(ix, iy).Color == col {
				newX = ix
				newY = iy
			}
		}
	}
	Board.Tiles.Set(int16(Board.Stats.At(0).X), int16(Board.Stats.At(0).Y), TTile{Element: E_EMPTY, Color: 0})
	if newX != 0 {
		Board.Stats.At(0).X = byte(newX)
		Board.Stats.At(0).Y = byte(newY)
	}
	GamePaused = true
	SoundQueue(4, "0\x014\x017\x011\x015\x018\x012\x016\x019\x013\x017\x01:\x014\x018\x01@\x01")
	TransitionDrawBoardChange()
	BoardEnter()
}

func GameAboutScreen() {
	TextWindowDisplayFile("ABOUT.HLP", "About ZZT...")
}

func GamePlayLoop(boardChanged bool) {
	var (
		exitLoop   bool
		pauseBlink bool
	)
	GameDrawSidebar := func() {
		SidebarClear()
		SidebarClearLine(0)
		SidebarClearLine(1)
		SidebarClearLine(2)
		VideoWriteText(61, 0, 0x1F, "    - - - - -      ")
		VideoWriteText(62, 1, 0x70, "      ZZT      ")
		VideoWriteText(61, 2, 0x1F, "    - - - - -      ")
		if GameStateElement == E_PLAYER {
			VideoWriteText(64, 7, 0x1E, " Health:")
			VideoWriteText(64, 8, 0x1E, "   Ammo:")
			VideoWriteText(64, 9, 0x1E, "Torches:")
			VideoWriteText(64, 10, 0x1E, "   Gems:")
			VideoWriteText(64, 11, 0x1E, "  Score:")
			VideoWriteText(64, 12, 0x1E, "   Keys:")
			VideoWriteText(62, 7, 0x1F, Chr(ElementDefs[E_PLAYER].Character))
			VideoWriteText(62, 8, 0x1B, Chr(ElementDefs[E_AMMO].Character))
			VideoWriteText(62, 9, 0x16, Chr(ElementDefs[E_TORCH].Character))
			VideoWriteText(62, 10, 0x1B, Chr(ElementDefs[E_GEM].Character))
			VideoWriteText(62, 12, 0x1F, Chr(ElementDefs[E_KEY].Character))
			VideoWriteText(62, 14, 0x70, " T ")
			VideoWriteText(65, 14, 0x1F, " Torch")
			VideoWriteText(62, 15, 0x30, " B ")
			VideoWriteText(62, 16, 0x70, " H ")
			VideoWriteText(65, 16, 0x1F, " Help")
			VideoWriteText(67, 18, 0x30, " \x18\x19\x1a\x1b ")
			VideoWriteText(72, 18, 0x1F, " Move")
			VideoWriteText(61, 19, 0x70, " Shift \x18\x19\x1a\x1b ")
			VideoWriteText(72, 19, 0x1F, " Shoot")
			VideoWriteText(62, 21, 0x70, " S ")
			VideoWriteText(65, 21, 0x1F, " Save game")
			VideoWriteText(62, 22, 0x30, " P ")
			VideoWriteText(65, 22, 0x1F, " Pause")
			VideoWriteText(62, 23, 0x70, " Q ")
			VideoWriteText(65, 23, 0x1F, " Quit")
		} else if GameStateElement == E_MONITOR {
			SidebarPromptSlider(false, 66, 21, "Game speed:;FS", &TickSpeed)
			VideoWriteText(62, 21, 0x70, " S ")
			VideoWriteText(62, 7, 0x30, " W ")
			VideoWriteText(65, 7, 0x1E, " World:")
			if Length(World.Info.Name) != 0 {
				VideoWriteText(69, 8, 0x1F, World.Info.Name)
			} else {
				VideoWriteText(69, 8, 0x1F, "Untitled")
			}
			VideoWriteText(62, 11, 0x70, " P ")
			VideoWriteText(65, 11, 0x1F, " Play")
			VideoWriteText(62, 12, 0x30, " R ")
			VideoWriteText(65, 12, 0x1E, " Restore game")
			VideoWriteText(62, 13, 0x70, " Q ")
			VideoWriteText(65, 13, 0x1E, " Quit")
			VideoWriteText(62, 16, 0x30, " A ")
			VideoWriteText(65, 16, 0x1F, " About ZZT!")
			VideoWriteText(62, 17, 0x70, " H ")
			VideoWriteText(65, 17, 0x1E, " High Scores")
			if EditorEnabled {
				VideoWriteText(62, 18, 0x30, " E ")
				VideoWriteText(65, 18, 0x1E, " Board Editor")
			}
		}

	}

	GameDrawSidebar()
	GameUpdateSidebar()
	if JustStarted {
		GameAboutScreen()
		if Length(StartupWorldFileName) != 0 {
			SidebarClearLine(8)
			VideoWriteText(69, 8, 0x1F, StartupWorldFileName)
			if !WorldLoad(StartupWorldFileName, ".ZZT", format.WorldDeserializeTitleOnly) {
				WorldCreate()
			}
		}
		ReturnBoardId = World.Info.CurrentBoard
		BoardChange(0)
		JustStarted = false
	}
	Board.Tiles.Set(int16(Board.Stats.At(0).X), int16(Board.Stats.At(0).Y), TTile{Element: byte(GameStateElement), Color: ElementDefs[GameStateElement].Color})
	if GameStateElement == E_MONITOR {
		DisplayMessage(0, "")
		VideoWriteText(62, 5, 0x1B, "Pick a command:")
	}
	if boardChanged {
		TransitionDrawBoardChange()
	}
	TickTimeDuration = int16(TickSpeed) * 2
	GamePlayExitRequested = false
	exitLoop = false
	CurrentTick = Random(100)
	CurrentStatTicked = Board.Stats.Count + 1
	for {
		if GamePaused {
			if SoundHasTimeElapsed(&TickTimeCounter, 25) {
				pauseBlink = !pauseBlink
			}
			if pauseBlink {
				VideoWriteText(int16(Board.Stats.At(0).X)-1, int16(Board.Stats.At(0).Y)-1, ElementDefs[E_PLAYER].Color, Chr(ElementDefs[E_PLAYER].Character))
			} else {
				if Board.Tiles.Get(int16(Board.Stats.At(0).X), int16(Board.Stats.At(0).Y)).Element == E_PLAYER {
					VideoWriteText(int16(Board.Stats.At(0).X)-1, int16(Board.Stats.At(0).Y)-1, 0x0F, " ")
				} else {
					BoardDrawTile(int16(Board.Stats.At(0).X), int16(Board.Stats.At(0).Y))
				}
			}
			VideoWriteText(64, 5, 0x1F, "Pausing...")
			Idle(IdleUntilFrame)
			InputUpdate()
			if InputKeyPressed == KEY_ESCAPE {
				GamePromptEndPlay()
			}
			if InputDeltaX != 0 || InputDeltaY != 0 {
				ElementDefs[Board.Tiles.Get(int16(Board.Stats.At(0).X)+InputDeltaX, int16(Board.Stats.At(0).Y)+InputDeltaY).Element].TouchProc(int16(Board.Stats.At(0).X)+InputDeltaX, int16(Board.Stats.At(0).Y)+InputDeltaY, 0, &InputDeltaX, &InputDeltaY)
			}
			if (InputDeltaX != 0 || InputDeltaY != 0) && ElementDefs[Board.Tiles.Get(int16(Board.Stats.At(0).X)+InputDeltaX, int16(Board.Stats.At(0).Y)+InputDeltaY).Element].Walkable {
				if Board.Tiles.Get(int16(Board.Stats.At(0).X), int16(Board.Stats.At(0).Y)).Element == E_PLAYER {
					MoveStat(0, int16(Board.Stats.At(0).X)+InputDeltaX, int16(Board.Stats.At(0).Y)+InputDeltaY)
				} else {
					BoardDrawTile(int16(Board.Stats.At(0).X), int16(Board.Stats.At(0).Y))
					Board.Stats.At(0).X += byte(InputDeltaX)
					Board.Stats.At(0).Y += byte(InputDeltaY)
					Board.Tiles.Set(int16(Board.Stats.At(0).X), int16(Board.Stats.At(0).Y), TTile{Element: E_PLAYER, Color: ElementDefs[E_PLAYER].Color})
					BoardDrawTile(int16(Board.Stats.At(0).X), int16(Board.Stats.At(0).Y))
					DrawPlayerSurroundings(int16(Board.Stats.At(0).X), int16(Board.Stats.At(0).Y), 0)
					DrawPlayerSurroundings(int16(Board.Stats.At(0).X)-InputDeltaX, int16(Board.Stats.At(0).Y)-InputDeltaY, 0)
				}
				GamePaused = false
				SidebarClearLine(5)
				CurrentTick = Random(100)
				CurrentStatTicked = Board.Stats.Count + 1
				World.Info.IsSave = true
			}
		} else {
			if CurrentStatTicked <= Board.Stats.Count {
				stat := Board.Stats.At(CurrentStatTicked)
				if stat.Cycle != 0 && CurrentTick%stat.Cycle == CurrentStatTicked%stat.Cycle {
					ElementDefs[Board.Tiles.Get(int16(stat.X), int16(stat.Y)).Element].TickProc(CurrentStatTicked)
				}
				CurrentStatTicked++
			}
		}
		if CurrentStatTicked > Board.Stats.Count && !GamePlayExitRequested {
			if SoundHasTimeElapsed(&TickTimeCounter, TickTimeDuration) {
				CurrentTick++
				if CurrentTick > 420 {
					CurrentTick = 1
				}
				CurrentStatTicked = 0
				InputUpdate()
				// On platforms like WASM, it is necessary to occasionally yield
				// to not freeze the web browser.
				if TickTimeDuration <= 0 {
					Idle(IdleMinimal)
				}
			} else {
				Idle(IdleUntilPit)
			}
		}
		if (exitLoop || GamePlayExitRequested) && GamePlayExitRequested {
			break
		}
	}
	SoundClearQueue()
	if GameStateElement == E_PLAYER {
		if World.Info.Health <= 0 {
			HighScoresAdd(World.Info.Score)
		}
	} else if GameStateElement == E_MONITOR {
		SidebarClearLine(5)
	}

	Board.Tiles.Set(int16(Board.Stats.At(0).X), int16(Board.Stats.At(0).Y), TTile{Element: E_PLAYER, Color: ElementDefs[E_PLAYER].Color})
	SoundBlockQueueing = false
}

func GameTitleLoop() {
	var (
		boardChanged bool
		startPlay    bool
	)
	GameTitleExitRequested = false
	JustStarted = true
	ReturnBoardId = 0
	boardChanged = true
	for {
		BoardChange(0)
		for {
			GameStateElement = E_MONITOR
			startPlay = false
			GamePaused = false
			GamePlayLoop(boardChanged)
			boardChanged = false
			switch UpCase(InputKeyPressed) {
			case 'W':
				if GameWorldLoad(".ZZT") {
					ReturnBoardId = World.Info.CurrentBoard
					boardChanged = true
				}
			case 'P':
				if World.Info.IsSave && !DebugEnabled {
					startPlay = WorldLoad(World.Info.Name, ".ZZT", 0)
					ReturnBoardId = World.Info.CurrentBoard
				} else {
					startPlay = true
				}
				if startPlay {
					BoardChange(ReturnBoardId)
					BoardEnter()
				}
			case 'A':
				GameAboutScreen()
			case 'E':
				if EditorEnabled {
					EditorLoop()
					ReturnBoardId = World.Info.CurrentBoard
					boardChanged = true
				}
			case 'S':
				SidebarPromptSlider(true, 66, 21, "Game speed:;FS", &TickSpeed)
				InputKeyPressed = '\x00'
			case 'R':
				if GameWorldLoad(".SAV") {
					ReturnBoardId = World.Info.CurrentBoard
					BoardChange(ReturnBoardId)
					startPlay = true
				}
			case 'H':
				HighScoresLoad()
				HighScoresDisplay(1)
			case '|':
				GameDebugPrompt()
			case KEY_ESCAPE, 'Q':
				GameTitleExitRequested = SidebarPromptYesNo("Quit ZZT? ", true)
			}
			if startPlay {
				GameStateElement = E_PLAYER
				GamePaused = true
				GamePlayLoop(true)
				boardChanged = true
			}
			if boardChanged || GameTitleExitRequested {
				break
			}
		}
		if GameTitleExitRequested {
			break
		}
	}
}

func GamePrintRegisterMessage() {
	var (
		s     string
		sLen  byte
		sData []byte
		iy    int16
		color uint8
	)
	SetCBreak(false)
	s = "END" + Chr(byte(49+Random(4))) + ".MSG"
	iy = 0
	color = 0x0F
	for i := 0; i < int(ResourceDataHeader.EntryCount); i++ {
		if ResourceDataHeader.Name[i] == s {
			f, err := VfsOpen(ResourceDataFileName)
			if err != nil {
				return
			}
			defer f.Close()
			f.Seek(int64(ResourceDataHeader.FileOffset[i]), io.SeekStart)
			for err == nil {
				err = format.ReadPByte(f, &sLen)
				if sLen == 0 {
					color--
				} else {
					sData = make([]byte, sLen)
					f.Read(sData)
					if sData[0] != '@' {
						VideoWriteText(0, iy, color, string(sData))
					} else {
						break
					}
				}
				iy++
			}
			VideoWriteText(28, 24, 0x1F, "Press any key to exit...")
			TextColor(LightGray)
			for {
				Idle(IdleUntilFrame)
				if KeyPressed() {
					break
				}
			}
			InputKeyPressed = ReadKey()
			VideoWriteText(28, 24, 0x00, "                        ")
			GotoXY(1, 23)
		}
	}
}
