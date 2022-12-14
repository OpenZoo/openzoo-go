package main // unit: Oop

import "strings" // interface uses: GameVars

// implementation uses: Sounds, TxtWind, Game, Elements

func OopError(statId int16, message string) {
	stat := Board.Stats.At(statId)
	DisplayMessage(200, "ERR: "+message)
	SoundQueue(5, "P\n")
	stat.DataPos = -1
}

func OopReadChar(statId int16, position *int16) {
	stat := Board.Stats.At(statId)
	if *position >= 0 && *position < stat.DataLen {
		OopChar = (*stat.Data)[*position]
		*position++
	} else {
		OopChar = '\x00'
	}
}

func OopReadWord(statId int16, position *int16) {
	var word strings.Builder
	for {
		OopReadChar(statId, position)
		if OopChar != ' ' {
			break
		}
	}
	OopChar = UpCase(OopChar)
	if OopChar < '0' || OopChar > '9' {
		for OopChar >= 'A' && OopChar <= 'Z' || OopChar == ':' || OopChar >= '0' && OopChar <= '9' || OopChar == '_' {
			word.WriteByte(OopChar)
			OopReadChar(statId, position)
			OopChar = UpCase(OopChar)
		}
	}
	OopWord = word.String()
	if *position > 0 {
		*position--
	}
}

func OopReadValue(statId int16, position *int16) {
	var sb strings.Builder
	for {
		OopReadChar(statId, position)
		if OopChar != ' ' {
			break
		}
	}
	OopChar = UpCase(OopChar)
	for OopChar >= '0' && OopChar <= '9' {
		sb.WriteByte(OopChar)
		OopReadChar(statId, position)
		OopChar = UpCase(OopChar)
	}
	s := sb.String()
	if *position > 0 {
		*position--
	}
	if Length(s) != 0 {
		OopValue = int16(Val(s))
	} else {
		OopValue = -1
	}
}

func OopSkipLine(statId int16, position *int16) {
	for {
		OopReadChar(statId, position)
		if OopChar == '\x00' || OopChar == '\r' {
			break
		}
	}
}

func OopParseDirection(statId int16, position *int16, dx, dy *int16) (result bool) {
	stat := Board.Stats.At(statId)
	result = true
	if OopWord == "N" || OopWord == "NORTH" {
		*dx = 0
		*dy = -1
	} else if OopWord == "S" || OopWord == "SOUTH" {
		*dx = 0
		*dy = 1
	} else if OopWord == "E" || OopWord == "EAST" {
		*dx = 1
		*dy = 0
	} else if OopWord == "W" || OopWord == "WEST" {
		*dx = -1
		*dy = 0
	} else if OopWord == "I" || OopWord == "IDLE" {
		*dx = 0
		*dy = 0
	} else if OopWord == "SEEK" {
		CalcDirectionSeek(int16(stat.X), int16(stat.Y), dx, dy)
	} else if OopWord == "FLOW" {
		*dx = stat.StepX
		*dy = stat.StepY
	} else if OopWord == "RND" {
		CalcDirectionRnd(dx, dy)
	} else if OopWord == "RNDNS" {
		*dx = 0
		*dy = Random(2)*2 - 1
	} else if OopWord == "RNDNE" {
		*dx = Random(2)
		if *dx == 0 {
			*dy = -1
		} else {
			*dy = 0
		}
	} else if OopWord == "CW" {
		OopReadWord(statId, position)
		result = OopParseDirection(statId, position, dy, dx)
		*dx = -*dx
	} else if OopWord == "CCW" {
		OopReadWord(statId, position)
		result = OopParseDirection(statId, position, dy, dx)
		*dy = -*dy
	} else if OopWord == "RNDP" {
		OopReadWord(statId, position)
		result = OopParseDirection(statId, position, dy, dx)
		if Random(2) == 0 {
			*dx = -*dx
		} else {
			*dy = -*dy
		}
	} else if OopWord == "OPP" {
		OopReadWord(statId, position)
		result = OopParseDirection(statId, position, dx, dy)
		*dx = -*dx
		*dy = -*dy
	} else {
		*dx = 0
		*dy = 0
		result = false
	}

	return
}

func OopReadDirection(statId int16, position *int16, dx, dy *int16) {
	OopReadWord(statId, position)
	if !OopParseDirection(statId, position, dx, dy) {
		OopError(statId, "Bad direction")
	}
}

func OopFindString(statId int16, s string, startPos int16) (OopFindString int16) {
	var wordPos, cmpPos int16
	stat := Board.Stats.At(statId)
	pos := startPos
	for pos <= stat.DataLen {
		wordPos = 1
		cmpPos = pos
		for {
			OopReadChar(statId, &cmpPos)
			if UpCase(s[wordPos-1]) != UpCase(OopChar) {
				goto NoMatch
			}
			wordPos++
			if wordPos > Length(s) {
				break
			}
		}
		OopReadChar(statId, &cmpPos)
		OopChar = UpCase(OopChar)
		if OopChar >= 'A' && OopChar <= 'Z' || OopChar == '_' {
		} else {
			OopFindString = pos
			return
		}
	NoMatch:
		pos++

	}
	OopFindString = -1
	return
}

func OopIterateStat(statId int16, iStat *int16, lookup string) (OopIterateStat bool) {
	var (
		pos   int16
		found bool
	)
	*iStat++
	found = false
	if lookup == "ALL" {
		if *iStat <= Board.Stats.Count {
			found = true
		}
	} else if lookup == "OTHERS" {
		if *iStat <= Board.Stats.Count {
			if *iStat != statId {
				found = true
			} else {
				*iStat++
				found = *iStat <= Board.Stats.Count
			}
		}
	} else if lookup == "SELF" {
		if statId > 0 && *iStat <= statId {
			*iStat = statId
			found = true
		}
	} else {
		for *iStat <= Board.Stats.Count && !found {
			if Board.Stats.At(*iStat).Data != nil {
				pos = 0
				OopReadChar(*iStat, &pos)
				if OopChar == '@' {
					OopReadWord(*iStat, &pos)
					if OopWord == lookup {
						found = true
					}
				}
			}
			if !found {
				*iStat++
			}
		}
	}

	OopIterateStat = found
	return
}

func OopFindLabel(statId int16, sendLabel string, iStat, iDataPos *int16, labelPrefix string) (OopFindLabel bool) {
	var (
		targetSplitPos int16
		targetLookup   string
		objectMessage  string
		foundStat      bool
	)
	foundStat = false
	targetSplitPos = Pos(':', sendLabel)
	if targetSplitPos <= 0 {
		if *iStat < statId {
			objectMessage = sendLabel
			*iStat = statId
			targetSplitPos = 0
			foundStat = true
		}
	} else {
		targetLookup = Copy(sendLabel, 1, targetSplitPos-1)
		objectMessage = Copy(sendLabel, targetSplitPos+1, Length(sendLabel)-targetSplitPos)
	}
FindNextStat:
	if targetSplitPos > 0 {
		foundStat = OopIterateStat(statId, iStat, targetLookup)
	}
	if foundStat {
		if objectMessage == "RESTART" {
			*iDataPos = 0
		} else {
			*iDataPos = OopFindString(*iStat, labelPrefix+objectMessage, 0)
			if *iDataPos < 0 && targetSplitPos > 0 {
				goto FindNextStat
			}
		}
		foundStat = *iDataPos >= 0
	}
	OopFindLabel = foundStat
	return
}

func WorldGetFlagPosition(name string) int {
	for i := 0; i < len(World.Info.Flags); i++ {
		if World.Info.Flags[i] == name {
			return i
		}
	}
	return -1
}

func WorldSetFlag(name string) {
	if WorldGetFlagPosition(name) < 0 {
		i := 0
		for i < (len(World.Info.Flags)-1) && len(World.Info.Flags[i]) != 0 {
			i++
		}
		World.Info.Flags[i] = name
	}
}

func WorldClearFlag(name string) {
	if WorldGetFlagPosition(name) >= 0 {
		World.Info.Flags[WorldGetFlagPosition(name)] = ""
	}
}

func OopStringToWord(input string) (OopStringToWord string) {
	var (
		output strings.Builder
		i      int16
	)
	for i = 1; i <= Length(input); i++ {
		if input[i-1] >= 'A' && input[i-1] <= 'Z' || input[i-1] >= '0' && input[i-1] <= '9' {
			output.WriteByte(input[i-1])
		} else if input[i-1] >= 'a' && input[i-1] <= 'z' {
			output.WriteByte(input[i-1] - 0x20)
		}

	}
	OopStringToWord = output.String()
	return
}

func OopParseTile(statId, position *int16, tile *TTile) (OopParseTile bool) {
	var i int16
	OopParseTile = false
	tile.Color = 0
	OopReadWord(*statId, position)
	for i = 1; i <= 7; i++ {
		if OopWord == OopStringToWord(ColorNames[i-1]) {
			tile.Color = byte(i + 0x08)
			OopReadWord(*statId, position)
			goto ColorFound
		}
	}
ColorFound:
	for i = 0; i <= MAX_ELEMENT; i++ {
		if OopWord == OopStringToWord(ElementDefs[i].Name) {
			OopParseTile = true
			tile.Element = byte(i)
			return
		}
	}

	return
}

func GetColorForTileMatch(tile TTile) (GetColorForTileMatch byte) {
	if ElementDefs[tile.Element].Color < COLOR_SPECIAL_MIN {
		GetColorForTileMatch = byte(int16(ElementDefs[tile.Element].Color) & 0x07)
	} else if ElementDefs[tile.Element].Color == COLOR_WHITE_ON_CHOICE {
		GetColorForTileMatch = byte(int16(tile.Color)>>4&0x0F + 8)
	} else {
		GetColorForTileMatch = byte(int16(tile.Color) & 0x0F)
	}

	return
}

func FindTileOnBoard(x, y *int16, tile TTile) (FindTileOnBoard bool) {
	FindTileOnBoard = false
	for true {
		*x++
		if *x > BOARD_WIDTH {
			*x = 1
			*y++
			if *y > BOARD_HEIGHT {
				return
			}
		}
		if Board.Tiles.Get(*x, *y).Element == tile.Element {
			if tile.Color == 0 || GetColorForTileMatch(Board.Tiles.Get(*x, *y)) == tile.Color {
				FindTileOnBoard = true
				return
			}
		}
	}
	return
}

func OopPlaceTile(x, y int16, tile *TTile) {
	var color byte
	if Board.Tiles.Get(x, y).Element != E_PLAYER {
		color = tile.Color
		if ElementDefs[tile.Element].Color < COLOR_SPECIAL_MIN {
			color = ElementDefs[tile.Element].Color
		} else {
			if color == 0 {
				color = Board.Tiles.Get(x, y).Color
			}
			if color == 0 {
				color = 0x0F
			}
			if ElementDefs[tile.Element].Color == COLOR_WHITE_ON_CHOICE {
				color = byte((int16(color)-8)*0x10 + 0x0F)
			}
		}
		if Board.Tiles.Get(x, y).Element == tile.Element {
			Board.Tiles.SetColor(x, y, color)
		} else {
			BoardDamageTile(x, y)
			if ElementDefs[tile.Element].Cycle >= 0 {
				AddStat(x, y, tile.Element, int16(color), ElementDefs[tile.Element].Cycle, StatTemplateDefault)
			} else {
				Board.Tiles.Set(x, y, TTile{Element: tile.Element, Color: color})
			}
		}
		BoardDrawTile(x, y)
	}
}

func OopCheckCondition(statId int16, position *int16) (result bool) {
	var (
		deltaX, deltaY int16
		tile           TTile
		ix, iy         int16
	)
	stat := Board.Stats.At(statId)
	if OopWord == "NOT" {
		OopReadWord(statId, position)
		result = !OopCheckCondition(statId, position)
	} else if OopWord == "ALLIGNED" {
		result = stat.X == Board.Stats.At(0).X || stat.Y == Board.Stats.At(0).Y
	} else if OopWord == "CONTACT" {
		result = Sqr(int16(stat.X)-int16(Board.Stats.At(0).X))+Sqr(int16(stat.Y)-int16(Board.Stats.At(0).Y)) == 1
	} else if OopWord == "BLOCKED" {
		OopReadDirection(statId, position, &deltaX, &deltaY)
		result = !ElementDefs[Board.Tiles.Get(int16(stat.X)+deltaX, int16(stat.Y)+deltaY).Element].Walkable
	} else if OopWord == "ENERGIZED" {
		result = World.Info.EnergizerTicks > 0
	} else if OopWord == "ANY" {
		if !OopParseTile(&statId, position, &tile) {
			OopError(statId, "Bad object kind")
		}
		ix = 0
		iy = 1
		result = FindTileOnBoard(&ix, &iy, tile)
	} else {
		result = WorldGetFlagPosition(OopWord) >= 0
	}

	return
}

func OopReadLineToEnd(statId int16, position *int16) string {
	var s strings.Builder
	OopReadChar(statId, position)
	for OopChar != '\x00' && OopChar != '\r' {
		s.WriteByte(OopChar)
		OopReadChar(statId, position)
	}
	return s.String()
}

func OopSend(statId int16, sendLabel string, ignoreLock bool) (OopSend bool) {
	var (
		iDataPos, iStat int16
		respectSelfLock bool
	)
	if statId < 0 {
		statId = -statId
		respectSelfLock = true
	} else {
		respectSelfLock = false
	}
	OopSend = false
	iStat = 0
	for OopFindLabel(statId, sendLabel, &iStat, &iDataPos, "\r:") {
		if Board.Stats.At(iStat).P2 == 0 || ignoreLock || statId == iStat && !respectSelfLock {
			if iStat == statId {
				OopSend = true
			}
			Board.Stats.At(iStat).DataPos = iDataPos
		}
	}
	return
}

func OopExecute(statId int16, position *int16, name string) {
	var (
		textWindow        *TTextWindowState
		textLine          string
		deltaX, deltaY    int16
		ix, iy            int16
		stopRunning       bool
		replaceStat       bool
		endOfProgram      bool
		replaceTile       TTile
		namePosition      int16
		lastPosition      int16
		repeatInsNextTick bool
		lineFinished      bool
		labelDataPos      int16
		labelStatId       int16
		counterPtr        *int16
		counterSubtract   bool
		bindStatId        int16
		insCount          int16
		argTile           TTile
		argTile2          TTile
	)
	stat := Board.Stats.At(statId)
StartParsing:
	stopRunning = false
	repeatInsNextTick = false
	replaceStat = false
	endOfProgram = false
	insCount = 0
	for {
	ReadInstruction:
		lineFinished = true

		lastPosition = *position
		OopReadChar(statId, position)
		for OopChar == ':' {
			for {
				OopReadChar(statId, position)
				if OopChar == '\x00' || OopChar == '\r' {
					break
				}
			}
			OopReadChar(statId, position)
		}
		if OopChar == '\'' {
			OopSkipLine(statId, position)
		} else if OopChar == '@' {
			OopSkipLine(statId, position)
		} else if OopChar == '/' || OopChar == '?' {
			if OopChar == '/' {
				repeatInsNextTick = true
			}
			OopReadWord(statId, position)
			if OopParseDirection(statId, position, &deltaX, &deltaY) {
				if deltaX != 0 || deltaY != 0 {
					if !ElementDefs[Board.Tiles.Get(int16(stat.X)+deltaX, int16(stat.Y)+deltaY).Element].Walkable {
						ElementPushablePush(int16(stat.X)+deltaX, int16(stat.Y)+deltaY, deltaX, deltaY)
					}
					if ElementDefs[Board.Tiles.Get(int16(stat.X)+deltaX, int16(stat.Y)+deltaY).Element].Walkable {
						MoveStat(statId, int16(stat.X)+deltaX, int16(stat.Y)+deltaY)
						repeatInsNextTick = false
					}
				} else {
					repeatInsNextTick = false
				}
				OopReadChar(statId, position)
				if OopChar != '\r' {
					*position--
				}
				stopRunning = true
			} else {
				OopError(statId, "Bad direction")
			}
		} else if OopChar == '#' {
		ReadCommand:
			OopReadWord(statId, position)

			if OopWord == "THEN" {
				OopReadWord(statId, position)
			}
			if Length(OopWord) == 0 {
				goto ReadInstruction
			}
			insCount++
			if Length(OopWord) != 0 {
				if OopWord == "GO" {
					OopReadDirection(statId, position, &deltaX, &deltaY)
					if !ElementDefs[Board.Tiles.Get(int16(stat.X)+deltaX, int16(stat.Y)+deltaY).Element].Walkable {
						ElementPushablePush(int16(stat.X)+deltaX, int16(stat.Y)+deltaY, deltaX, deltaY)
					}
					if ElementDefs[Board.Tiles.Get(int16(stat.X)+deltaX, int16(stat.Y)+deltaY).Element].Walkable {
						MoveStat(statId, int16(stat.X)+deltaX, int16(stat.Y)+deltaY)
					} else {
						repeatInsNextTick = true
					}
					stopRunning = true
				} else if OopWord == "TRY" {
					OopReadDirection(statId, position, &deltaX, &deltaY)
					if !ElementDefs[Board.Tiles.Get(int16(stat.X)+deltaX, int16(stat.Y)+deltaY).Element].Walkable {
						ElementPushablePush(int16(stat.X)+deltaX, int16(stat.Y)+deltaY, deltaX, deltaY)
					}
					if ElementDefs[Board.Tiles.Get(int16(stat.X)+deltaX, int16(stat.Y)+deltaY).Element].Walkable {
						MoveStat(statId, int16(stat.X)+deltaX, int16(stat.Y)+deltaY)
						stopRunning = true
					} else {
						goto ReadCommand
					}
				} else if OopWord == "WALK" {
					OopReadDirection(statId, position, &deltaX, &deltaY)
					stat.StepX = deltaX
					stat.StepY = deltaY
				} else if OopWord == "SET" {
					OopReadWord(statId, position)
					WorldSetFlag(OopWord)
				} else if OopWord == "CLEAR" {
					OopReadWord(statId, position)
					WorldClearFlag(OopWord)
				} else if OopWord == "IF" {
					OopReadWord(statId, position)
					if OopCheckCondition(statId, position) {
						goto ReadCommand
					}
				} else if OopWord == "SHOOT" {
					OopReadDirection(statId, position, &deltaX, &deltaY)
					if BoardShoot(E_BULLET, int16(stat.X), int16(stat.Y), deltaX, deltaY, SHOT_SOURCE_ENEMY) {
						SoundQueue(2, "0\x01&\x01")
					}
					stopRunning = true
				} else if OopWord == "THROWSTAR" {
					OopReadDirection(statId, position, &deltaX, &deltaY)
					if BoardShoot(E_STAR, int16(stat.X), int16(stat.Y), deltaX, deltaY, SHOT_SOURCE_ENEMY) {
					}
					stopRunning = true
				} else if OopWord == "GIVE" || OopWord == "TAKE" {
					if OopWord == "TAKE" {
						counterSubtract = true
					} else {
						counterSubtract = false
					}
					OopReadWord(statId, position)
					if OopWord == "HEALTH" {
						counterPtr = &World.Info.Health
					} else if OopWord == "AMMO" {
						counterPtr = &World.Info.Ammo
					} else if OopWord == "GEMS" {
						counterPtr = &World.Info.Gems
					} else if OopWord == "TORCHES" {
						counterPtr = &World.Info.Torches
					} else if OopWord == "SCORE" {
						counterPtr = &World.Info.Score
					} else if OopWord == "TIME" {
						counterPtr = &World.Info.BoardTimeSec
					} else {
						counterPtr = nil
					}

					if counterPtr != nil {
						OopReadValue(statId, position)
						if OopValue > 0 {
							if counterSubtract {
								OopValue = -OopValue
							}
							if *counterPtr+OopValue >= 0 {
								*counterPtr += OopValue
							} else {
								goto ReadCommand
							}
						}
					}
					GameUpdateSidebar()
				} else if OopWord == "END" {
					*position = -1
					OopChar = '\x00'
				} else if OopWord == "ENDGAME" {
					World.Info.Health = 0
				} else if OopWord == "IDLE" {
					stopRunning = true
				} else if OopWord == "RESTART" {
					*position = 0
					lineFinished = false
				} else if OopWord == "ZAP" {
					OopReadWord(statId, position)
					labelStatId = 0
					for OopFindLabel(statId, OopWord, &labelStatId, &labelDataPos, "\r:") {
						(*(Board.Stats.At(labelStatId).Data))[labelDataPos+1] = '\''
					}
				} else if OopWord == "RESTORE" {
					OopReadWord(statId, position)
					labelStatId = 0
					for OopFindLabel(statId, OopWord, &labelStatId, &labelDataPos, "\r'") {
						for {
							(*(Board.Stats.At(labelStatId).Data))[labelDataPos+1] = ':'
							labelDataPos = OopFindString(labelStatId, "\r'"+OopWord+"\r", labelDataPos)
							if labelDataPos <= 0 {
								break
							}
						}
					}
				} else if OopWord == "LOCK" {
					stat.P2 = 1
				} else if OopWord == "UNLOCK" {
					stat.P2 = 0
				} else if OopWord == "SEND" {
					OopReadWord(statId, position)
					if OopSend(statId, OopWord, false) {
						lineFinished = false
					}
				} else if OopWord == "BECOME" {
					if OopParseTile(&statId, position, &argTile) {
						replaceStat = true
						replaceTile.Element = argTile.Element
						replaceTile.Color = argTile.Color
					} else {
						OopError(statId, "Bad #BECOME")
					}
				} else if OopWord == "PUT" {
					OopReadDirection(statId, position, &deltaX, &deltaY)
					if deltaX == 0 && deltaY == 0 {
						OopError(statId, "Bad #PUT")
					} else if !OopParseTile(&statId, position, &argTile) {
						OopError(statId, "Bad #PUT")
					} else if int16(stat.X)+deltaX > 0 && int16(stat.X)+deltaX <= BOARD_WIDTH && int16(stat.Y)+deltaY > 0 && int16(stat.Y)+deltaY < BOARD_HEIGHT {
						if !ElementDefs[Board.Tiles.Get(int16(stat.X)+deltaX, int16(stat.Y)+deltaY).Element].Walkable {
							ElementPushablePush(int16(stat.X)+deltaX, int16(stat.Y)+deltaY, deltaX, deltaY)
						}
						OopPlaceTile(int16(stat.X)+deltaX, int16(stat.Y)+deltaY, &argTile)
					}

				} else if OopWord == "CHANGE" {
					if !OopParseTile(&statId, position, &argTile) {
						OopError(statId, "Bad #CHANGE")
					}
					if !OopParseTile(&statId, position, &argTile2) {
						OopError(statId, "Bad #CHANGE")
					}
					ix = 0
					iy = 1
					if argTile2.Color == 0 && ElementDefs[argTile2.Element].Color < COLOR_SPECIAL_MIN {
						argTile2.Color = ElementDefs[argTile2.Element].Color
					}
					for FindTileOnBoard(&ix, &iy, argTile) {
						OopPlaceTile(ix, iy, &argTile2)
					}
				} else if OopWord == "PLAY" {
					textLine = SoundParse(OopReadLineToEnd(statId, position))
					if Length(textLine) != 0 {
						SoundQueue(-1, textLine)
					}
					lineFinished = false
				} else if OopWord == "CYCLE" {
					OopReadValue(statId, position)
					if OopValue > 0 {
						stat.Cycle = OopValue
					}
				} else if OopWord == "CHAR" {
					OopReadValue(statId, position)
					if OopValue > 0 && OopValue <= 255 {
						stat.P1 = byte(OopValue)
						BoardDrawTile(int16(stat.X), int16(stat.Y))
					}
				} else if OopWord == "DIE" {
					replaceStat = true
					replaceTile.Element = E_EMPTY
					replaceTile.Color = 0x0F
				} else if OopWord == "BIND" {
					OopReadWord(statId, position)
					bindStatId = 0
					if OopIterateStat(statId, &bindStatId, OopWord) {
						stat.Data = Board.Stats.At(bindStatId).Data
						stat.DataLen = Board.Stats.At(bindStatId).DataLen
						*position = 0
					}
				} else {
					textLine = OopWord
					if OopSend(statId, OopWord, false) {
						lineFinished = false
					} else {
						if Pos(':', textLine) <= 0 {
							OopError(statId, "Bad command "+textLine)
						}
					}
				}

			}
			if lineFinished {
				OopSkipLine(statId, position)
			}
		} else if OopChar == '\r' {
			if textWindow != nil && len(textWindow.Lines) > 0 {
				textWindow.Append("")
			}
		} else if OopChar == '\x00' {
			endOfProgram = true
		} else {
			textLine = Chr(OopChar)
			textLine += OopReadLineToEnd(statId, position)
			if textWindow == nil {
				textWindow = NewTextWindowState()
				textWindow.Selectable = false
			}
			textWindow.Append(textLine)
		}

		if endOfProgram || stopRunning || repeatInsNextTick || replaceStat || insCount > 32 {
			break
		}
	}
	if repeatInsNextTick {
		*position = lastPosition
	}
	if OopChar == '\x00' {
		*position = -1
	}
	if textWindow != nil {
		if len(textWindow.Lines) > 1 {
			namePosition = 0
			OopReadChar(statId, &namePosition)
			if OopChar == '@' {
				name = OopReadLineToEnd(statId, &namePosition)
			}
			if Length(name) == 0 {
				name = "Interaction"
			}
			textWindow.Title = name
			textWindow.DrawOpen()
			textWindow.Select(true, false)
			textWindow.DrawClose()
			if Length(textWindow.Hyperlink) != 0 {
				if OopSend(statId, textWindow.Hyperlink, false) {
					textWindow = nil
					goto StartParsing
				}
			}
		} else if len(textWindow.Lines) == 1 {
			DisplayMessage(200, textWindow.Lines[0])
		}
	}

	if replaceStat {
		ix = int16(stat.X)
		iy = int16(stat.Y)
		DamageStat(statId)
		OopPlaceTile(ix, iy, &replaceTile)
	}
}
