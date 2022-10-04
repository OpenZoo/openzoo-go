package main

import (
	"strings"
)

// Pascal shims - Crt procedures on top of Platform

const (
	Black        uint8 = 0
	Blue         uint8 = 1
	Green        uint8 = 2
	Cyan         uint8 = 3
	Red          uint8 = 4
	Magenta      uint8 = 5
	Brown        uint8 = 6
	LightGray    uint8 = 7
	DarkGray     uint8 = 8
	LightBlue    uint8 = 9
	LightGreen   uint8 = 10
	LightCyan    uint8 = 11
	LightRed     uint8 = 12
	LightMagenta uint8 = 13
	Yellow       uint8 = 14
	White        uint8 = 15
	Blink        uint8 = 128
)

var windowMinX int = 1
var windowMinY int = 1
var windowMaxX int = 80
var windowMaxY int = 25
var cursorX int = 1
var cursorY int = 1
var TextAttr uint8

func Window(x1, y1, x2, y2 int) {
	windowMinX = x1
	windowMinY = y1
	windowMaxX = x2
	windowMaxY = y2

	cursorX = windowMinX
	cursorY = windowMinY
}

func GotoXY(x, y int) {
	cursorX = Clamp(x, windowMinX, windowMaxX)
	cursorY = Clamp(y, windowMinY, windowMaxY)
}

func ClrScr() {
	line := strings.Repeat(" ", windowMaxX-windowMinX+1)
	for iy := windowMinY; iy <= windowMaxY; iy++ {
		IVideoWriteText(int16(windowMinX-1), int16(iy-1), TextAttr, line)
	}
}

func TextBackground(v uint8) {
	TextAttr = (TextAttr & 0x0F) | ((v << 4) & 0xF0)
}

func TextColor(v uint8) {
	TextAttr = (TextAttr & 0xF0) | (v & 0x0F)
}

func Write(s string) {
	for i := 0; i < len(s); i++ {
		switch c := s[i]; c {
		case '\r':
			cursorX = windowMinX
		case '\n':
			if cursorY < windowMaxY {
				cursorY++
			} else {
				// TODO: scroll up
			}
		default:
			IVideoWriteText(int16(cursorX)-1, int16(cursorY)-1, TextAttr, s[i:i+1])
			cursorX++
			if cursorX > windowMaxX {
				Write("\r\n")
			}
		}
	}
}

func WriteLn(s string) {
	Write(s)
	Write("\r\n")
}
