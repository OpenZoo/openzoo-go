package main

var VideoMonochrome bool = false

func VideoInstall(columns int, backgroundColor uint8) {
	if VideoMonochrome {
		backgroundColor = 0
	}
	IVideoSetMode(columns)
	IVideoClrScr(backgroundColor)
}

func colorToBw(color byte) byte {
	// This is inspired on the ZZT 2.0/3.0 algorithm, not the ZZT 3.1/3.2 algorithm.
	// TODO: Super ZZT prefers the ZZT 3.1/3.2 algorithm.

	// FIX: Special handling of blinking solids
	if (color & 0x80) == 0x80 {
		if ((color >> 4) & 0x07) == (color & 0x0F) {
			color = color & 0x7F
		}
	}

	if (color & 0x09) == 0x09 {
		color = (color & 0xF0) | 0x0F
	} else if (color & 0x07) != 0 {
		color = (color & 0xF0) | 0x07
	}

	if (color & 0x0F) == 0x00 {
		if (color & 0x70) == 0x00 {
			color = (color & 0x8F)
		} else {
			color = (color & 0x8F) | 0x70
		}
	} else if (color & 0x70) != 0x70 {
		color = (color & 0x8F)
	}

	return color
}

func VideoConfigure() bool {
	WriteLn("")
	Write("  Video mode:  C)olor,  M)onochrome?  ")
	for {
		for {
			if KeyPressed() {
				break
			}
			Idle(IdleUntilFrame)
		}
		charTyped := UpCase(ReadKey())
		switch charTyped {
		case 'C':
			VideoMonochrome = false
			return true
		case 'M':
			VideoMonochrome = true
			return true
		case KEY_ESCAPE:
			return false
		}
	}
}

func VideoClrScr() {
	IVideoClrScr(0)
}

func VideoWriteText(x, y int16, color byte, text string) {
	if VideoMonochrome {
		color = colorToBw(color)
	}
	IVideoWriteText(x, y, color, text)
}

func VideoShowCursor() {
	IVideoSetCursorVisible(true)
}

func VideoHideCursor() {
	IVideoSetCursorVisible(false)
}
