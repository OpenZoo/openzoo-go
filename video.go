package main

import "github.com/OpenZoo/openzoo-go/platform"

var VideoMonochrome bool = false

func VideoInstall(columns int, backgroundColor uint8) {
	platform.VideoInstall(columns)
	platform.VideoClrScr(backgroundColor)
}

func VideoConfigure() bool {
	// stub
	return true
}

func VideoClrScr() {
	platform.VideoClrScr(0)
}

func VideoWriteText(x, y int16, color byte, text string) {
	platform.VideoWriteText(x, y, color, text)
}

func VideoShowCursor() {
	platform.VideoSetCursorVisible(true)
}

func VideoHideCursor() {
	platform.VideoSetCursorVisible(false)
}

func VideoUninstall() {
	platform.VideoUninstall()
}

func VideoMove(x, y, width int16, buffer *[]byte, toVideo bool) {
	platform.VideoMove(x, y, width, buffer, toVideo)
}
