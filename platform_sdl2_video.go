//go:build sdl2

package main

import (
	"C"
	_ "embed"
	"image/color"
)

//go:embed ascii.chr
var charsetData []byte
var textBuffer [25][160]byte
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

	ch := int(textBuffer[iy][ix*2]) * 14
	co := textBuffer[iy][ix*2+1]
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

func IVideoInstall(columns int) {
	MainThreadSync(func() {
		textColumns = columns
	})
}

func IVideoClrScr(backgroundColor uint8) {
	MainThreadAsync(func() {
		for iy := 0; iy < 25; iy++ {
			for ix := 0; ix < textColumns; ix++ {
				textBuffer[iy][ix*2] = ' '
				textBuffer[iy][ix*2+1] = backgroundColor << 4
				redrawChar(ix, iy)
			}
		}
		VideoUpdateRequested.Store(true)
	})
}

func IVideoWriteText(x, y int16, color byte, text string) {
	MainThreadAsync(func() {
		for i := 0; i < len(text); i++ {
			textBuffer[y][x*2] = text[i]
			textBuffer[y][x*2+1] = color
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
		VideoUpdateRequested.Store(true)
	})
}

func IVideoSetCursorVisible(v bool) {
	MainThreadSync(func() {
		// stub
	})
}

func VideoUninstall() {
	MainThreadSync(func() {
		// stub
	})
}

func VideoMove(x, y, width int16, buffer *[]byte, toVideo bool) {
	if toVideo {
		MainThreadAsync(func() {
			if buffer != nil {
				if width > int16(len(*buffer)>>1) {
					width = int16(len(*buffer) >> 1)
				}
				for i := 0; i < int(width)*2; i++ {
					textBuffer[y][int(x)*2+i] = (*buffer)[i]
				}
				for i := 0; i < int(width); i++ {
					redrawChar(int(x)+i, int(y))
				}
			}

			VideoUpdateRequested.Store(true)
		})
	} else {
		// wait for all async video writes to end
		MainThreadSync(func() {

		})
		*buffer = make([]byte, width*2)
		copy(*buffer, textBuffer[y][x*2:(x+width)*2])
	}
}
