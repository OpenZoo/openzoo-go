package main // unit: TxtWind

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
) // interface uses: Video

const (
	MAX_TEXT_WINDOW_LINES   = 1024
	MAX_RESOURCE_DATA_FILES = 24
)

type (
	TTextWindowLine  string
	TTextWindowState struct {
		Selectable     bool
		LineCount      int16
		LinePos        int16
		Lines          [MAX_TEXT_WINDOW_LINES]string
		Hyperlink      string
		Title          string
		LoadedFilename string
		ScreenCopy     [25][]byte
	}
	TResourceDataHeader struct {
		EntryCount int16
		Name       [MAX_RESOURCE_DATA_FILES]string
		FileOffset [MAX_RESOURCE_DATA_FILES]int32
	}
)

var (
	TextWindowX, TextWindowY          int16
	TextWindowWidth, TextWindowHeight int16
	TextWindowStrInnerEmpty           string
	TextWindowStrText                 string
	TextWindowStrInnerLine            string
	TextWindowStrTop                  string
	TextWindowStrBottom               string
	TextWindowStrSep                  string
	TextWindowStrInnerSep             string
	TextWindowStrInnerArrows          string
	TextWindowRejected                bool
	ResourceDataFileName              string
	ResourceDataHeader                TResourceDataHeader
	OrderPrintId                      *string
)

func ReadResourceDataHeader(r io.Reader, b *TResourceDataHeader) error {
	if err := ReadPShort(r, &b.EntryCount); err != nil {
		return err
	}
	for i := 0; i < MAX_RESOURCE_DATA_FILES; i++ {
		if err := ReadPString(r, &b.Name[i], 50); err != nil {
			return err
		}
	}
	for i := 0; i < MAX_RESOURCE_DATA_FILES; i++ {
		if err := ReadPLongint(r, &b.FileOffset[i]); err != nil {
			return err
		}
	}
	return nil
}

// implementation uses: Crt, Input, Printer

const TEXT_WINDOW_ANIM_SPEED = 25

func NewTextWindowState() *TTextWindowState {
	var state TTextWindowState
	TextWindowInitState(&state)
	return &state
}

func TextWindowInitState(state *TTextWindowState) {
	state.LineCount = 0
	state.LinePos = 1
	state.LoadedFilename = ""
}

func TextWindowDrawTitle(color int16, title string) {
	VideoWriteText(TextWindowX+2, TextWindowY+1, byte(color), TextWindowStrInnerEmpty)
	VideoWriteText(TextWindowX+(TextWindowWidth-Length(title))/2, TextWindowY+1, byte(color), title)
}

func TextWindowDrawOpen(state *TTextWindowState) {
	var iy int16
	for iy = 1; iy <= TextWindowHeight+1; iy++ {
		VideoMove(TextWindowX, iy+TextWindowY-1, TextWindowWidth, &state.ScreenCopy[iy-1], false)
	}
	for iy = TextWindowHeight / 2; iy >= 0; iy-- {
		VideoWriteText(TextWindowX, TextWindowY+iy+1, 0x0F, TextWindowStrText)
		VideoWriteText(TextWindowX, TextWindowY+TextWindowHeight-iy-1, 0x0F, TextWindowStrText)
		VideoWriteText(TextWindowX, TextWindowY+iy, 0x0F, TextWindowStrTop)
		VideoWriteText(TextWindowX, TextWindowY+TextWindowHeight-iy, 0x0F, TextWindowStrBottom)
		Delay(TEXT_WINDOW_ANIM_SPEED)
	}
	VideoWriteText(TextWindowX, TextWindowY+2, 0x0F, TextWindowStrSep)
	TextWindowDrawTitle(0x1E, state.Title)
}

func TextWindowDrawClose(state *TTextWindowState) {
	var iy int16
	for iy = 0; iy <= TextWindowHeight/2; iy++ {
		VideoWriteText(TextWindowX, TextWindowY+iy, 0x0F, TextWindowStrTop)
		VideoWriteText(TextWindowX, TextWindowY+TextWindowHeight-iy, 0x0F, TextWindowStrBottom)
		Delay(TEXT_WINDOW_ANIM_SPEED * 3 / 4)
		VideoMove(TextWindowX, TextWindowY+iy, TextWindowWidth, &state.ScreenCopy[iy+1-1], true)
		VideoMove(TextWindowX, TextWindowY+TextWindowHeight-iy, TextWindowWidth, &state.ScreenCopy[TextWindowHeight-iy+1-1], true)
	}
}

func TextWindowDrawLine(state *TTextWindowState, lpos int16, withoutFormatting, viewingFile bool) {
	var (
		lineY                        int16
		textOffset, textColor, textX int16
	)
	lineY = TextWindowY + lpos - state.LinePos + TextWindowHeight/2 + 1
	if lpos == state.LinePos {
		VideoWriteText(TextWindowX+2, lineY, 0x1C, TextWindowStrInnerArrows)
	} else {
		VideoWriteText(TextWindowX+2, lineY, 0x1E, TextWindowStrInnerEmpty)
	}
	if lpos > 0 && lpos <= state.LineCount {
		if withoutFormatting {
			VideoWriteText(TextWindowX+4, lineY, 0x1E, state.Lines[lpos-1])
		} else {
			textOffset = 1
			textColor = 0x1E
			textX = TextWindowX + 4
			if Length(state.Lines[lpos-1]) > 0 {
				switch state.Lines[lpos-1][0] {
				case '!':
					textOffset = Pos(';', state.Lines[lpos-1]) + 1
					VideoWriteText(textX+2, lineY, 0x1D, "\x10")
					textX += 5
					textColor = 0x1F
				case ':':
					textOffset = Pos(';', state.Lines[lpos-1]) + 1
					textColor = 0x1F
				case '$':
					textOffset = 2
					textColor = 0x1F
					textX = textX - 4 + (TextWindowWidth-Length(state.Lines[lpos-1]))/2
				}
			}
			if textOffset > 0 {
				VideoWriteText(textX, lineY, byte(textColor), Copy(state.Lines[lpos-1], textOffset, Length(state.Lines[lpos-1])-textOffset+1))
			}
		}
	} else if lpos == 0 || lpos == state.LineCount+1 {
		VideoWriteText(TextWindowX+2, lineY, 0x1E, TextWindowStrInnerSep)
	} else if lpos == -4 && viewingFile {
		VideoWriteText(TextWindowX+2, lineY, 0x1A, "   Use            to view text,")
		VideoWriteText(TextWindowX+2+7, lineY, 0x1F, "\x18 \x19, Enter")
	} else if lpos == -3 && viewingFile {
		VideoWriteText(TextWindowX+2+1, lineY, 0x1A, "                 to print.")
		VideoWriteText(TextWindowX+2+12, lineY, 0x1F, "Alt-P")
	}

}

func TextWindowDraw(state *TTextWindowState, withoutFormatting, viewingFile bool) {
	var i int16
	for i = 0; i <= TextWindowHeight-4; i++ {
		TextWindowDrawLine(state, state.LinePos-TextWindowHeight/2+i+2, withoutFormatting, viewingFile)
	}
	TextWindowDrawTitle(0x1E, state.Title)
}

func TextWindowAppend(state *TTextWindowState, line string) {
	state.LineCount++
	state.Lines[state.LineCount-1] = line
}

func TextWindowFree(state *TTextWindowState) {
	for state.LineCount > 0 {
		state.Lines[state.LineCount-1] = ""
		state.LineCount--
	}
	state.LoadedFilename = ""
}

func TextWindowPrint(state *TTextWindowState) {
	// stub
	/* var (
		iLine, iChar int16
		line         string
	)
	Rewrite(Lst)
	for iLine = 1; iLine <= state.LineCount; iLine++ {
		line = state.Lines[iLine-1]
		if Length(line) > 0 {
			switch line[0] {
			case '$':
				line = Delete(line, 1, 1)
				for iChar = (80 - Length(line)) / 2; iChar >= 1; iChar-- {
					line = " " + line
				}
			case '!', ':':
				iChar = Pos(';', line)
				if iChar > 0 {
					line = Delete(line, 1, iChar)
				} else {
					line = ""
				}
			default:
				line = "          " + line
			}
		}
		WriteLn(Lst, line)
		if IOResult() != 0 {
			Close(Lst)
			return
		}
	}
	if state.LoadedFilename == "ORDER.HLP" {
		WriteLn(Lst, *OrderPrintId)
	}
	Write(Lst, Chr(12))
	Close(Lst) */
}

func TextWindowSelect(state *TTextWindowState, hyperlinkAsSelect, viewingFile bool) {
	var (
		newLinePos   int16
		iLine, iChar int16
		pointerStr   string
	)
	TextWindowRejected = false
	state.Hyperlink = ""
	TextWindowDraw(state, false, viewingFile)
	for {
		Idle(IMUntilFrame)
		InputUpdate()
		newLinePos = state.LinePos
		if InputDeltaY != 0 {
			newLinePos += InputDeltaY
		} else if InputShiftPressed || InputKeyPressed == KEY_ENTER {
			InputShiftAccepted = true
			if len(state.Lines[state.LinePos-1]) > 0 && state.Lines[state.LinePos-1][0] == '!' {
				pointerStr = Copy(state.Lines[state.LinePos-1], 2, Length(state.Lines[state.LinePos-1])-1)
				if Pos(';', pointerStr) > 0 {
					pointerStr = Copy(pointerStr, 1, Pos(';', pointerStr)-1)
				}
				if len(pointerStr) > 0 && pointerStr[0] == '-' {
					pointerStr = pointerStr[1:]
					TextWindowFree(state)
					TextWindowOpenFile(pointerStr, state)
					if state.LineCount == 0 {
						return
					} else {
						viewingFile = true
						newLinePos = state.LinePos
						TextWindowDraw(state, false, viewingFile)
						InputKeyPressed = '\x00'
						InputShiftPressed = false
					}
				} else {
					if hyperlinkAsSelect {
						state.Hyperlink = pointerStr
					} else {
						pointerStr = ":" + pointerStr
						for iLine = 1; iLine <= state.LineCount; iLine++ {
							if Length(pointerStr) > Length(state.Lines[iLine-1]) {
							} else {
								for iChar = 1; iChar <= Length(pointerStr); iChar++ {
									if UpCase(pointerStr[iChar-1]) != UpCase(byte(state.Lines[iLine-1][iChar-1])) {
										goto LabelNotMatched
									}
								}
								newLinePos = iLine
								InputKeyPressed = '\x00'
								InputShiftPressed = false
								goto LabelMatched
							LabelNotMatched:
							}
						}
					}
				}
			}
		} else {
			if InputKeyPressed == KEY_PAGE_UP {
				newLinePos = state.LinePos - TextWindowHeight + 4
			} else if InputKeyPressed == KEY_PAGE_DOWN {
				newLinePos = state.LinePos + TextWindowHeight - 4
			} else if InputKeyPressed == KEY_ALT_P {
				TextWindowPrint(state)
			}

		}

	LabelMatched:
		if newLinePos < 1 {
			newLinePos = 1
		} else if newLinePos > state.LineCount {
			newLinePos = state.LineCount
		}

		if newLinePos != state.LinePos {
			state.LinePos = newLinePos
			TextWindowDraw(state, false, viewingFile)
			if len(state.Lines[state.LinePos-1]) > 0 && state.Lines[state.LinePos-1][0] == '!' {
				if hyperlinkAsSelect {
					TextWindowDrawTitle(0x1E, "\xaePress ENTER to select this\xaf")
				} else {
					TextWindowDrawTitle(0x1E, "\xaePress ENTER for more info\xaf")
				}
			}
		}
		if InputJoystickMoved {
			Delay(35)
		}
		if InputKeyPressed == KEY_ESCAPE || InputKeyPressed == KEY_ENTER || InputShiftPressed {
			break
		}
	}
	if InputKeyPressed == KEY_ESCAPE {
		InputKeyPressed = '\x00'
		TextWindowRejected = true
	}
}

func TextWindowEdit(state *TTextWindowState) {
	var (
		newLinePos int16
		insertMode bool
		charPos    int16
		i          int16
	)
	DeleteCurrLine := func() {
		var i int16
		if state.LineCount > 1 {
			state.Lines[state.LinePos-1] = ""
			for i = state.LinePos + 1; i <= state.LineCount; i++ {
				state.Lines[i-1-1] = state.Lines[i-1]
			}
			state.LineCount--
			if state.LinePos > state.LineCount {
				newLinePos = state.LineCount
			} else {
				TextWindowDraw(state, true, false)
			}
		} else {
			state.Lines[0] = ""
		}
	}

	if state.LineCount == 0 {
		TextWindowAppend(state, "")
	}
	insertMode = true
	state.LinePos = 1
	charPos = 1
	TextWindowDraw(state, true, false)
	for {
		if insertMode {
			VideoWriteText(77, 14, 0x1E, "on ")
		} else {
			VideoWriteText(77, 14, 0x1E, "off")
		}
		if charPos >= Length(state.Lines[state.LinePos-1])+1 {
			charPos = Length(state.Lines[state.LinePos-1]) + 1
			VideoWriteText(charPos+TextWindowX+3, TextWindowY+TextWindowHeight/2+1, 0x70, " ")
		} else {
			VideoWriteText(charPos+TextWindowX+3, TextWindowY+TextWindowHeight/2+1, 0x70, string([]byte{state.Lines[state.LinePos-1][charPos-1]}))
		}
		InputReadWaitKey()
		newLinePos = state.LinePos
		switch InputKeyPressed {
		case KEY_UP:
			newLinePos = state.LinePos - 1
		case KEY_DOWN:
			newLinePos = state.LinePos + 1
		case KEY_PAGE_UP:
			newLinePos = state.LinePos - TextWindowHeight + 4
		case KEY_PAGE_DOWN:
			newLinePos = state.LinePos + TextWindowHeight - 4
		case KEY_RIGHT:
			charPos++
			if charPos > Length(state.Lines[state.LinePos-1])+1 {
				charPos = 1
				newLinePos = state.LinePos + 1
			}
		case KEY_LEFT:
			charPos--
			if charPos < 1 {
				charPos = TextWindowWidth
				newLinePos = state.LinePos - 1
			}
		case KEY_ENTER:
			if state.LineCount < MAX_TEXT_WINDOW_LINES {
				for i = state.LineCount; i >= state.LinePos+1; i-- {
					state.Lines[i+1-1] = state.Lines[i-1]
				}
				state.Lines[state.LinePos+1-1] = Copy(state.Lines[state.LinePos-1], charPos, Length(state.Lines[state.LinePos-1])-charPos+1)
				state.Lines[state.LinePos-1] = Copy(state.Lines[state.LinePos-1], 1, charPos-1)
				newLinePos = state.LinePos + 1
				charPos = 1
				state.LineCount++
			}
		case KEY_BACKSPACE:
			if charPos > 1 {
				state.Lines[state.LinePos-1] = Copy(state.Lines[state.LinePos-1], 1, charPos-2) + Copy(state.Lines[state.LinePos-1], charPos, Length(state.Lines[state.LinePos-1])-charPos+1)
				charPos--
			} else if Length(state.Lines[state.LinePos-1]) == 0 {
				DeleteCurrLine()
				newLinePos = state.LinePos - 1
				charPos = TextWindowWidth
			}

		case KEY_INSERT:
			insertMode = !insertMode
		case KEY_DELETE:
			state.Lines[state.LinePos-1] = Copy(state.Lines[state.LinePos-1], 1, charPos-1) + Copy(state.Lines[state.LinePos-1], charPos+1, Length(state.Lines[state.LinePos-1])-charPos)
		case KEY_CTRL_Y:
			DeleteCurrLine()
		default:
			if InputKeyPressed >= ' ' && charPos < TextWindowWidth-7 {
				if !insertMode {
					state.Lines[state.LinePos-1] = Copy(state.Lines[state.LinePos-1], 1, charPos-1) + string(InputKeyPressed) + Copy(state.Lines[state.LinePos-1], charPos+1, Length(state.Lines[state.LinePos-1])-charPos)
					charPos++
				} else {
					if Length(state.Lines[state.LinePos-1]) < TextWindowWidth-8 {
						state.Lines[state.LinePos-1] = Copy(state.Lines[state.LinePos-1], 1, charPos-1) + string(InputKeyPressed) + Copy(state.Lines[state.LinePos-1], charPos, Length(state.Lines[state.LinePos-1])-charPos+1)
						charPos++
					}
				}
			}
		}
		if newLinePos < 1 {
			newLinePos = 1
		} else if newLinePos > state.LineCount {
			newLinePos = state.LineCount
		}

		if newLinePos != state.LinePos {
			state.LinePos = newLinePos
			TextWindowDraw(state, true, false)
		} else {
			TextWindowDrawLine(state, state.LinePos, true, false)
		}
		if InputKeyPressed == KEY_ESCAPE {
			break
		}
	}
	if Length(state.Lines[state.LineCount-1]) == 0 {
		state.LineCount--
	}
}

func TextWindowOpenFile(filename string, state *TTextWindowState) error {
	var (
		i        int16
		entryPos int16
		retVal   bool
	)
	retVal = true
	for i = 1; i <= Length(filename); i++ {
		retVal = retVal && filename[i-1] != '.'
	}
	if retVal {
		filename += ".HLP"
	}
	if len(filename) > 0 && filename[0] == '*' {
		filename = Copy(filename, 2, Length(filename)-1)
		entryPos = -1
	} else {
		entryPos = 0
	}
	TextWindowInitState(state)
	state.LoadedFilename = strings.ToUpper(filename)
	if ResourceDataHeader.EntryCount == 0 {
		f, err := os.Open(PathFindCaseInsensitiveFile(ResourceDataFileName))
		ResourceDataHeader.EntryCount = -1
		if err == nil {
			err := ReadResourceDataHeader(f, &ResourceDataHeader)
			if err != nil {
				ResourceDataHeader.EntryCount = -1
			}
		}
		f.Close()
	}
	if entryPos == 0 {
		for i = 1; i <= ResourceDataHeader.EntryCount; i++ {
			if strings.EqualFold(ResourceDataHeader.Name[i-1], filename) {
				entryPos = i
			}
		}
	}
	if entryPos <= 0 {
		f, err := os.Open(PathFindCaseInsensitiveFile(filename))
		if err != nil {
			return err
		}
		defer f.Close()
		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			for _, l := range strings.Split(scanner.Text(), "\r") {
				state.LineCount++
				state.Lines[state.LineCount-1] = l
			}
		}
		return scanner.Err()
	} else {
		f, err := os.Open(PathFindCaseInsensitiveFile(ResourceDataFileName))
		if err != nil {
			return err
		}
		defer f.Close()

		_, err = f.Seek(int64(ResourceDataHeader.FileOffset[entryPos-1]), os.SEEK_SET)
		if err != nil {
			return err
		}

		for {
			state.LineCount++
			var lineLen byte
			if err := ReadPByte(f, &lineLen); err != nil {
				return err
			}
			if lineLen == 0 {
				state.Lines[state.LineCount-1] = ""
			} else {
				strB := make([]byte, lineLen)
				if _, err := f.Read(strB); err != nil {
					return err
				}
				state.Lines[state.LineCount-1] = string(strB)
			}
			if state.Lines[state.LineCount-1] == "@" {
				state.Lines[state.LineCount-1] = ""
				break
			}
		}
	}
	return nil
}

func TextWindowSaveFile(filename string, state *TTextWindowState) error {
	var i int16
	f, err := os.Create(PathFindCaseInsensitiveFile(filename))
	if err != nil {
		return err
	}
	defer f.Close()

	for i = 1; i <= state.LineCount; i++ {
		if _, err := fmt.Fprintln(f, state.Lines[i-1]); err != nil {
			return err
		}
	}

	return nil
}

func TextWindowDisplayFile(filename, title string) {
	var state TTextWindowState
	state.Title = title
	TextWindowOpenFile(filename, &state)
	state.Selectable = false
	if state.LineCount > 0 {
		TextWindowDrawOpen(&state)
		TextWindowSelect(&state, false, true)
		TextWindowDrawClose(&state)
	}
	TextWindowFree(&state)
}

func TextWindowInit(x, y, width, height int16) {
	var i int16
	TextWindowX = x
	TextWindowWidth = width
	TextWindowY = y
	TextWindowHeight = height
	TextWindowStrInnerEmpty = ""
	TextWindowStrInnerLine = ""
	for i = 1; i <= TextWindowWidth-5; i++ {
		TextWindowStrInnerEmpty += " "
		TextWindowStrInnerLine += "\xcd"
	}
	TextWindowStrTop = "\xc6\xd1" + TextWindowStrInnerLine + "\xd1" + "\xb5"
	TextWindowStrBottom = "\xc6\xcf" + TextWindowStrInnerLine + "\xcf" + "\xb5"
	TextWindowStrSep = " \xc6" + TextWindowStrInnerLine + "\xb5" + " "
	TextWindowStrText = " \xb3" + TextWindowStrInnerEmpty + "\xb3" + " "
	TextWindowStrInnerArrows = "\xaf" + TextWindowStrInnerEmpty[1:len(TextWindowStrInnerEmpty)-1] + "\xae"
	TextWindowStrInnerSep = TextWindowStrInnerEmpty
	for i = 1; i < TextWindowWidth/5; i++ {
		splicePos := i*5 + TextWindowWidth%5/2 - 1
		TextWindowStrInnerSep = TextWindowStrInnerSep[:splicePos] + "\x07" + TextWindowStrInnerSep[splicePos+1:]
	}
}

func init() {
	ResourceDataFileName = ""
	ResourceDataHeader.EntryCount = 0
}
