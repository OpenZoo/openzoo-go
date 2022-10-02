package platform

import (
	"C"
	_ "embed"
	"image/color"
)
import "github.com/veandco/go-sdl2/sdl"

//go:embed ascii.chr
var charsetData []byte
var textBuffer [25][80][2]byte
var textColumns int = 80

var palette = []color.RGBA{
	{0x00, 0x00, 0x00, 255},
	{0x00, 0x00, 0xAA, 255},
	{0x00, 0xAA, 0x00, 255},
	{0x00, 0xAA, 0xAA, 255},
	{0xAA, 0x00, 0x00, 255},
	{0xAA, 0x00, 0xAA, 255},
	{0xAA, 0x55, 0x00, 255},
	{0xAA, 0xAA, 0xAA, 255},
	{0x55, 0x55, 0x55, 255},
	{0x55, 0x55, 0xFF, 255},
	{0x55, 0xFF, 0x55, 255},
	{0x55, 0xFF, 0xFF, 255},
	{0xFF, 0x55, 0x55, 255},
	{0xFF, 0x55, 0xFF, 255},
	{0xFF, 0xFF, 0x55, 255},
	{0xFF, 0xFF, 0xFF, 255},
}

func redrawChar(ix, iy int) {
	// TODO: This is *super* slow.

	px := ix * 8
	py := iy * 14
	VideoSurface.Lock()

	ch := int(textBuffer[iy][ix][0]) * 14
	co := textBuffer[iy][ix][1]
	coBg := palette[co>>4]
	coFg := palette[co&0xF]

	for ly := 0; ly < 14; ly++ {
		c := charsetData[ch+ly]
		for lx := 0; lx < 8; lx++ {
			if (c & 0x80) != 0 {
				VideoSurface.Set(px+lx, py+ly, coFg)
			} else {
				VideoSurface.Set(px+lx, py+ly, coBg)
			}
			c <<= 1
		}
	}

	VideoSurface.Unlock()
}

func VideoInstall(columns int) {
	sdl.Do(func() {
		textColumns = columns
	})
}

func VideoClrScr(backgroundColor uint8) {
	sdl.Do(func() {
		for iy := 0; iy < 25; iy++ {
			for ix := 0; ix < textColumns; ix++ {
				textBuffer[iy][ix][0] = ' '
				textBuffer[iy][ix][1] = backgroundColor << 4
				redrawChar(ix, iy)
			}
		}
		VideoWindow.UpdateSurface()
	})
}

func VideoWriteText(x, y int16, color byte, text string) {
	sdl.Do(func() {
		for i := 0; i < len(text); i++ {
			textBuffer[y][x][0] = text[i]
			textBuffer[y][x][1] = color
			redrawChar(int(x), int(y))
			x++
			if x >= int16(textColumns) {
				x = 0
				y++
				if y >= 25 {
					break
				}
			}
		}
		VideoWindow.UpdateSurface()
	})
}

func VideoSetCursorVisible(v bool) {
	sdl.Do(func() {
		// stub
	})
}

func VideoUninstall() {
	sdl.Do(func() {
		// stub
	})
}

func VideoMove(x, y, width int16, buffer *[]byte, toVideo bool) {
	sdl.Do(func() {
		// stub
		VideoWindow.UpdateSurface()
	})
}
