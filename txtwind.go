package main // unit: TxtWind

import (
	"bufio"
	"io"
	"strings"

	"github.com/OpenZoo/openzoo-go/format"
) // interface uses: Video

const (
	MAX_RESOURCE_DATA_FILES = 24
)

type (
	TTextWindowLine  string
	TTextWindowState struct {
		Selectable     bool
		LinePos        int
		Lines          []string
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
	if err := format.ReadPShort(r, &b.EntryCount); err != nil {
		return err
	}
	for i := 0; i < MAX_RESOURCE_DATA_FILES; i++ {
		if err := format.ReadPString(r, &b.Name[i], 50); err != nil {
			return err
		}
	}
	for i := 0; i < MAX_RESOURCE_DATA_FILES; i++ {
		if err := format.ReadPLongint(r, &b.FileOffset[i]); err != nil {
			return err
		}
	}
	return nil
}

// implementation uses: Crt, Input, Printer

const TEXT_WINDOW_ANIM_SPEED = 25

func NewTextWindowState() *TTextWindowState {
	var state TTextWindowState
	state.Init()
	return &state
}

func (state *TTextWindowState) Init() {
	state.Lines = make([]string, 0)
	state.LinePos = 1
	state.LoadedFilename = ""
}

func (state *TTextWindowState) drawTitle(color int16, title string) {
	VideoWriteText(TextWindowX+2, TextWindowY+1, byte(color), TextWindowStrInnerEmpty)
	VideoWriteText(TextWindowX+(TextWindowWidth-Length(title))/2, TextWindowY+1, byte(color), title)
}

func (state *TTextWindowState) DrawOpen() {
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
	state.drawTitle(0x1E, state.Title)
}

func (state *TTextWindowState) DrawClose() {
	var iy int16
	for iy = 0; iy <= TextWindowHeight/2; iy++ {
		VideoWriteText(TextWindowX, TextWindowY+iy, 0x0F, TextWindowStrTop)
		VideoWriteText(TextWindowX, TextWindowY+TextWindowHeight-iy, 0x0F, TextWindowStrBottom)
		Delay(TEXT_WINDOW_ANIM_SPEED * 3 / 4)
		VideoMove(TextWindowX, TextWindowY+iy, TextWindowWidth, &state.ScreenCopy[iy+1-1], true)
		VideoMove(TextWindowX, TextWindowY+TextWindowHeight-iy, TextWindowWidth, &state.ScreenCopy[TextWindowHeight-iy+1-1], true)
	}
}

func (state *TTextWindowState) drawLine(lpos int16, withoutFormatting, viewingFile bool) {
	var (
		lineY                        int16
		textOffset, textColor, textX int16
	)
	lineY = TextWindowY + lpos - int16(state.LinePos) + TextWindowHeight/2 + 1
	if int(lpos) == state.LinePos {
		VideoWriteText(TextWindowX+2, lineY, 0x1C, TextWindowStrInnerArrows)
	} else {
		VideoWriteText(TextWindowX+2, lineY, 0x1E, TextWindowStrInnerEmpty)
	}
	if lpos > 0 && int(lpos) <= len(state.Lines) {
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
	} else if lpos == 0 || int(lpos) == len(state.Lines)+1 {
		VideoWriteText(TextWindowX+2, lineY, 0x1E, TextWindowStrInnerSep)
	} else if lpos == -4 && viewingFile {
		VideoWriteText(TextWindowX+2, lineY, 0x1A, "   Use            to view text,")
		VideoWriteText(TextWindowX+2+7, lineY, 0x1F, "\x18 \x19, Enter")
	} else if lpos == -3 && viewingFile {
		VideoWriteText(TextWindowX+2+1, lineY, 0x1A, "                 to print.")
		VideoWriteText(TextWindowX+2+12, lineY, 0x1F, "Alt-P")
	}

}

func (state *TTextWindowState) Draw(withoutFormatting, viewingFile bool) {
	var i int16
	for i = 0; i <= TextWindowHeight-4; i++ {
		state.drawLine(int16(state.LinePos)-TextWindowHeight/2+i+2, withoutFormatting, viewingFile)
	}
	state.drawTitle(0x1E, state.Title)
}

func (state *TTextWindowState) Append(line string) (newLinePos int) {
	state.Lines = append(state.Lines, line)
	newLinePos = len(state.Lines)
	return
}

func (state *TTextWindowState) Print() {
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

func (state *TTextWindowState) Select(hyperlinkAsSelect, viewingFile bool) {
	var (
		newLinePos int
		iChar      int16
		pointerStr string
	)
	TextWindowRejected = false
	state.Hyperlink = ""
	state.Draw(false, viewingFile)
	for {
		Idle(IdleUntilFrame)
		InputUpdate()
		newLinePos = state.LinePos
		if InputDeltaY != 0 {
			newLinePos += int(InputDeltaY)
		} else if InputShiftPressed || InputKeyPressed == KEY_ENTER {
			InputShiftAccepted = true
			if len(state.Lines[state.LinePos-1]) > 0 && state.Lines[state.LinePos-1][0] == '!' {
				pointerStr = Copy(state.Lines[state.LinePos-1], 2, Length(state.Lines[state.LinePos-1])-1)
				if Pos(';', pointerStr) > 0 {
					pointerStr = Copy(pointerStr, 1, Pos(';', pointerStr)-1)
				}
				if len(pointerStr) > 0 && pointerStr[0] == '-' {
					pointerStr = pointerStr[1:]
					state.OpenFile(pointerStr)
					if len(state.Lines) == 0 {
						return
					} else {
						viewingFile = true
						newLinePos = state.LinePos
						state.Draw(false, viewingFile)
						InputKeyPressed = '\x00'
						InputShiftPressed = false
					}
				} else {
					if hyperlinkAsSelect {
						state.Hyperlink = pointerStr
					} else {
						pointerStr = ":" + pointerStr
						for iLine := 1; iLine <= len(state.Lines); iLine++ {
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
				newLinePos = state.LinePos - int(TextWindowHeight) + 4
			} else if InputKeyPressed == KEY_PAGE_DOWN {
				newLinePos = state.LinePos + int(TextWindowHeight) - 4
			} else if InputKeyPressed == KEY_ALT_P {
				state.Print()
			}

		}

	LabelMatched:
		if newLinePos < 1 {
			newLinePos = 1
		} else if newLinePos > len(state.Lines) {
			newLinePos = len(state.Lines)
		}

		if newLinePos != state.LinePos {
			state.LinePos = newLinePos
			state.Draw(false, viewingFile)
			if len(state.Lines[state.LinePos-1]) > 0 && state.Lines[state.LinePos-1][0] == '!' {
				if hyperlinkAsSelect {
					state.drawTitle(0x1E, "\xaePress ENTER to select this\xaf")
				} else {
					state.drawTitle(0x1E, "\xaePress ENTER for more info\xaf")
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

func (state *TTextWindowState) Edit() {
	var (
		newLinePos int
		insertMode bool
		charPos    int16
	)
	DeleteCurrLine := func() {
		if len(state.Lines) > 1 {
			state.Lines[state.LinePos-1] = ""
			state.Lines = append(state.Lines[:state.LinePos-1], state.Lines[state.LinePos:]...)
			if state.LinePos > len(state.Lines) {
				newLinePos = len(state.Lines)
			} else {
				state.Draw(true, false)
			}
		} else {
			state.Lines[0] = ""
		}
	}

	if len(state.Lines) == 0 {
		state.Append("")
	}
	insertMode = true
	state.LinePos = 1
	charPos = 1
	state.Draw(true, false)
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
			newLinePos = state.LinePos - int(TextWindowHeight) + 4
		case KEY_PAGE_DOWN:
			newLinePos = state.LinePos + int(TextWindowHeight) - 4
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
			state.Lines = append(state.Lines, "")
			for i := len(state.Lines) - 1; i >= state.LinePos+1; i-- {
				state.Lines[i+1-1] = state.Lines[i-1]
			}
			state.Lines[state.LinePos+1-1] = Copy(state.Lines[state.LinePos-1], charPos, Length(state.Lines[state.LinePos-1])-charPos+1)
			state.Lines[state.LinePos-1] = Copy(state.Lines[state.LinePos-1], 1, charPos-1)
			newLinePos = state.LinePos + 1
			charPos = 1
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
		} else if newLinePos > len(state.Lines) {
			newLinePos = len(state.Lines)
		}

		if newLinePos != state.LinePos {
			state.LinePos = newLinePos
			state.Draw(true, false)
		} else {
			state.drawLine(int16(state.LinePos), true, false)
		}
		if InputKeyPressed == KEY_ESCAPE {
			break
		}
	}
	if Length(state.Lines[len(state.Lines)-1]) == 0 {
		state.Lines = state.Lines[:len(state.Lines)-1]
	}
}

func (state *TTextWindowState) OpenFile(filename string) error {
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
	state.Init()
	state.LoadedFilename = strings.ToUpper(filename)
	if ResourceDataHeader.EntryCount == 0 {
		f, err := VfsOpen(ResourceDataFileName)
		ResourceDataHeader.EntryCount = -1
		if err == nil {
			err := ReadResourceDataHeader(f, &ResourceDataHeader)
			if err != nil {
				ResourceDataHeader.EntryCount = -1
			}
			f.Close()
		}
	}
	if entryPos == 0 {
		for i = 1; i <= ResourceDataHeader.EntryCount; i++ {
			if strings.EqualFold(ResourceDataHeader.Name[i-1], filename) {
				entryPos = i
			}
		}
	}
	if entryPos <= 0 {
		f, err := VfsOpen(filename)
		if err != nil {
			return err
		}
		defer f.Close()
		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			for _, l := range strings.Split(scanner.Text(), "\r") {
				state.Lines = append(state.Lines, l)
			}
		}
		return scanner.Err()
	} else {
		f, err := VfsOpen(ResourceDataFileName)
		if err != nil {
			return err
		}
		defer f.Close()

		_, err = f.Seek(int64(ResourceDataHeader.FileOffset[entryPos-1]), io.SeekStart)
		if err != nil {
			return err
		}

		for {
			var lineLen byte
			if err := format.ReadPByte(f, &lineLen); err != nil {
				return err
			}
			if lineLen == 0 {
				state.Lines = append(state.Lines, "")
			} else {
				strB := make([]byte, lineLen)
				if _, err := f.Read(strB); err != nil {
					return err
				}
				strS := string(strB)
				if strS == "@" {
					state.Lines = append(state.Lines, "")
					break
				} else {
					state.Lines = append(state.Lines, string(strB))
				}
			}
		}
	}
	return nil
}

func (state *TTextWindowState) SaveFile(filename string) error {
	f, err := VfsCreate(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	w := bufio.NewWriter(f)
	defer w.Flush()

	for i := 0; i < len(state.Lines); i++ {
		if _, err := w.WriteString(state.Lines[i] + "\n"); err != nil {
			return err
		}
	}

	return nil
}

func TextWindowDisplayFile(filename, title string) {
	var state TTextWindowState
	state.Title = title
	state.OpenFile(filename)
	state.Selectable = false
	if len(state.Lines) > 0 {
		state.DrawOpen()
		state.Select(false, true)
		state.DrawClose()
	}
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
