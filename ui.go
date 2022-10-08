package main

const (
	PROMPT_NUMERIC  = 0
	PROMPT_ALPHANUM = 1
	PROMPT_ANY      = 2
)

func SidebarClearLine(y int16) {
	VideoWriteText(60, y, 0x11, "\xb3                   ")
}

func SidebarClear() {
	var i int16
	for i = 3; i <= 24; i++ {
		SidebarClearLine(i)
	}
}

func SidebarPromptCharacter(editable bool, x, y int16, prompt string, value *byte) {
	var i, newValue int16
	SidebarClearLine(y)
	VideoWriteText(x, y, byte(BoolToInt(editable)+0x1E), prompt)
	SidebarClearLine(y + 1)
	VideoWriteText(x+5, y+1, 0x9F, "\x1f")
	SidebarClearLine(y + 2)
	for {
		for i = int16(*value) - 4; i <= int16(*value)+4; i++ {
			VideoWriteText(x+i-int16(*value)+5, y+2, 0x1E, Chr(byte((i+0x100)%0x100)))
		}
		if editable {
			Delay(25)
			InputUpdate()
			if InputKeyPressed == KEY_TAB {
				InputDeltaX = 9
			}
			newValue = int16(*value) + InputDeltaX
			if int16(*value) != newValue {
				*value = byte((newValue + 0x100) % 0x100)
				SidebarClearLine(y + 2)
			}
		}
		if InputKeyPressed == KEY_ENTER || InputKeyPressed == KEY_ESCAPE || !editable || InputShiftPressed {
			break
		}
	}
	VideoWriteText(x+5, y+1, 0x1F, "\x1f")
}

func SidebarPromptSlider(editable bool, x, y int16, prompt string, value *byte) {
	var (
		newValue           int16
		startChar, endChar byte
	)
	if prompt[Length(prompt)-2-1] == ';' {
		startChar = prompt[Length(prompt)-1-1]
		endChar = prompt[Length(prompt)-1]
		prompt = Copy(prompt, 1, Length(prompt)-3)
	} else {
		startChar = '1'
		endChar = '9'
	}
	SidebarClearLine(y)
	VideoWriteText(x, y, byte(BoolToInt(editable)+0x1E), prompt)
	SidebarClearLine(y + 1)
	SidebarClearLine(y + 2)
	VideoWriteText(x, y+2, 0x1E, Chr(startChar)+"....:...."+Chr(endChar))
	for {
		if editable {
			if InputJoystickMoved {
				Delay(45)
			}
			VideoWriteText(x+int16(*value)+1, y+1, 0x9F, "\x1f")
			Idle(IdleUntilFrame)
			InputUpdate()
			if InputKeyPressed >= '1' && InputKeyPressed <= '9' {
				*value = byte(InputKeyPressed) - 49
				SidebarClearLine(y + 1)
			} else {
				newValue = int16(*value) + InputDeltaX
				if int16(*value) != newValue && newValue >= 0 && newValue <= 8 {
					*value = byte(newValue)
					SidebarClearLine(y + 1)
				}
			}
		}
		if InputKeyPressed == KEY_ENTER || InputKeyPressed == KEY_ESCAPE || !editable || InputShiftPressed {
			break
		}
	}
	VideoWriteText(x+int16(*value)+1, y+1, 0x1F, "\x1f")
}

func SidebarPromptChoice(editable bool, y int16, prompt, choiceStr string, result *byte) {
	var (
		i, j, choiceCount int16
		newResult         int16
	)
	SidebarClearLine(y)
	SidebarClearLine(y + 1)
	SidebarClearLine(y + 2)
	VideoWriteText(63, y, byte(BoolToInt(editable)+0x1E), prompt)
	VideoWriteText(63, y+2, 0x1E, choiceStr)
	choiceCount = 1
	for i = 1; i <= Length(choiceStr); i++ {
		if choiceStr[i-1] == ' ' {
			choiceCount++
		}
	}
	for {
		j = 0
		i = 1
		for j < int16(*result) && i < Length(choiceStr) {
			if choiceStr[i-1] == ' ' {
				j++
			}
			i++
		}
		if editable {
			VideoWriteText(62+i, y+1, 0x9F, "\x1f")
			Delay(35)
			InputUpdate()
			newResult = int16(*result) + InputDeltaX
			if int16(*result) != newResult && newResult >= 0 && newResult <= choiceCount-1 {
				*result = byte(newResult)
				SidebarClearLine(y + 1)
			}
		}
		if InputKeyPressed == KEY_ENTER || InputKeyPressed == KEY_ESCAPE || !editable || InputShiftPressed {
			break
		}
	}
	VideoWriteText(62+i, y+1, 0x1F, "\x1f")
}

func SidebarPromptDirection(editable bool, y int16, prompt string, deltaX, deltaY *int16) {
	var choice byte
	if *deltaY == -1 {
		choice = 0
	} else if *deltaY == 1 {
		choice = 1
	} else if *deltaX == -1 {
		choice = 2
	} else {
		choice = 3
	}

	SidebarPromptChoice(editable, y, prompt, "\x18 \x19 \x1b \x1a", &choice)
	*deltaX = NeighborDeltaX[choice]
	*deltaY = NeighborDeltaY[choice]
}

func PromptString(x, y, arrowColor, color, width int16, mode byte, buffer *string) {
	var (
		i             int16
		oldBuffer     string
		firstKeyPress bool
	)
	oldBuffer = *buffer
	firstKeyPress = true
	for {
		for i = 0; i <= width-1; i++ {
			VideoWriteText(x+i, y, byte(color), " ")
			VideoWriteText(x+i, y-1, byte(arrowColor), " ")
		}
		VideoWriteText(x+width, y-1, byte(arrowColor), " ")
		VideoWriteText(x+Length(*buffer), y-1, byte(arrowColor/0x10*16+0x0F), "\x1f")
		VideoWriteText(x, y, byte(color), *buffer)
		InputReadWaitKey()
		if Length(*buffer) < width && InputKeyPressed >= ' ' && InputKeyPressed < '\x80' {
			if firstKeyPress {
				*buffer = ""
			}
			switch mode {
			case PROMPT_NUMERIC:
				if InputKeyPressed >= '0' && InputKeyPressed <= '9' {
					*buffer += string([]byte{byte(InputKeyPressed)})
				}
			case PROMPT_ANY:
				*buffer += string([]byte{byte(InputKeyPressed)})
			case PROMPT_ALPHANUM:
				if UpCase(InputKeyPressed) >= 'A' && UpCase(InputKeyPressed) <= 'Z' || InputKeyPressed >= '0' && InputKeyPressed <= '9' || InputKeyPressed == '-' {
					*buffer += string([]byte{byte(UpCase(InputKeyPressed))})
				}
			}
		} else if InputKeyPressed == KEY_LEFT || InputKeyPressed == KEY_BACKSPACE {
			*buffer = Copy(*buffer, 1, Length(*buffer)-1)
		}

		firstKeyPress = false
		if InputKeyPressed == KEY_ENTER || InputKeyPressed == KEY_ESCAPE {
			break
		}
	}
	if InputKeyPressed == KEY_ESCAPE {
		*buffer = oldBuffer
	}
}

func SidebarPromptYesNo(message string, defaultReturn bool) (SidebarPromptYesNo bool) {
	SidebarClearLine(3)
	SidebarClearLine(4)
	SidebarClearLine(5)
	VideoWriteText(63, 5, 0x1F, message)
	VideoWriteText(63+Length(message), 5, 0x9E, "_")
	for {
		InputReadWaitKey()
		if UpCase(InputKeyPressed) == KEY_ESCAPE || UpCase(InputKeyPressed) == 'N' || UpCase(InputKeyPressed) == 'Y' {
			break
		}
	}
	if UpCase(InputKeyPressed) == 'Y' {
		defaultReturn = true
	} else {
		defaultReturn = false
	}
	SidebarClearLine(5)
	SidebarPromptYesNo = defaultReturn
	return
}

func SidebarPromptString(prompt string, extension string, filename *string, promptMode byte) {
	SidebarClearLine(3)
	SidebarClearLine(4)
	SidebarClearLine(5)
	VideoWriteText(75-Length(prompt), 3, 0x1F, prompt)
	VideoWriteText(63, 5, 0x0F, "        "+extension)
	PromptString(63, 5, 0x1E, 0x0F, 8, promptMode, filename)
	SidebarClearLine(3)
	SidebarClearLine(4)
	SidebarClearLine(5)
}

func PopupPromptString(question string, buffer *string) {
	var x, y int16
	VideoWriteText(3, 18, 0x4F, TextWindowStrTop)
	VideoWriteText(3, 19, 0x4F, TextWindowStrText)
	VideoWriteText(3, 20, 0x4F, TextWindowStrSep)
	VideoWriteText(3, 21, 0x4F, TextWindowStrText)
	VideoWriteText(3, 22, 0x4F, TextWindowStrText)
	VideoWriteText(3, 23, 0x4F, TextWindowStrBottom)
	VideoWriteText(4+(TextWindowWidth-Length(question))/2, 19, 0x4F, question)
	*buffer = ""
	PromptString(10, 22, 0x4F, 0x4E, TextWindowWidth-16, PROMPT_ANY, buffer)
	for y = 18; y <= 23; y++ {
		for x = 3; x <= TextWindowWidth+3; x++ {
			BoardDrawTile(x+1, y+1)
		}
	}
}
