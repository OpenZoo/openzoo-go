//go:build wasm

package main

import (
	_ "embed"
	"sync"
	"sync/atomic"
	"syscall/js"
	"time"
)

var FrameTickCond = sync.NewCond(&sync.Mutex{})
var PitTickCond = sync.NewCond(&sync.Mutex{})
var VideoUpdateRequested atomic.Bool

var timerTicks int = 0

func TimerTicks() int {
	return timerTicks
}

func MemAvail() int32 {
	// no-op
	return 655360
}

func SetCBreak(v bool) {
	// no-op
}

func Idle(mode IdleMode) {
	switch mode {
	case IdleUntilFrame:
		FrameTickCond.L.Lock()
		FrameTickCond.Wait()
		FrameTickCond.L.Unlock()
	case IdleUntilPit:
		PitTickCond.L.Lock()
		PitTickCond.Wait()
		PitTickCond.L.Unlock()
	case IdleMinimal:
		time.Sleep(1 * time.Nanosecond)
	}
}

func Delay(ms uint32) {
	time.Sleep(time.Duration(ms) * time.Millisecond)
}

func KeysUpdateModifiers() {
	v := js.Global().Get("ozg_keymod").Invoke().Int()
	KeysRightShiftHeld = false
	KeysLeftShiftHeld = (v & 0x01) != 0
	KeysShiftHeld = (v & 0x01) != 0
	KeysCtrlHeld = (v & 0x04) != 0
	KeysAltHeld = (v & 0x08) != 0
}

func KeyPressed() bool {
	return js.Global().Get("ozg_key").Invoke(false).Int() >= 0
}

func ReadKey() byte {
	return byte(js.Global().Get("ozg_key").Invoke(true).Int())
}

var frameTicker = make(chan bool)
var tickerDone = make(chan bool)
var audioSlice = make([]byte, 2048)

func main() {
	IVideoSetMode(80)

	js.Global().Set("ozg_videoRenderCommunicate", js.FuncOf(func(this js.Value, args []js.Value) any {
		frameTicker <- true

		data := args[0]
		force := args[1].Truthy()

		requested := VideoUpdateRequested.Swap(false)
		if force || requested {
			js.CopyBytesToJS(data, textBuffer)
			return true
		} else {
			return false
		}
	}))

	CurrentAudioSimulator = NewAudioSimulatorNearest(48000, 32)

	js.Global().Set("ozg_audioCallback", js.FuncOf(func(this js.Value, args []js.Value) any {
		CurrentAudioSimulator.Simulate(audioSlice)
		js.CopyBytesToJS(args[0], audioSlice)
		return nil
	}))

	js.Global().Get("ozg_init").Invoke()

	Delay(33)

	pitTicker := time.NewTicker(55 * time.Millisecond)

	go func() {
		for {
			select {
			case <-frameTicker:
				FrameTickCond.Broadcast()
			case <-pitTicker.C:
				SoundTimerHandler()
				timerTicks++
				PitTickCond.Broadcast()
			case <-tickerDone:
				return
			}
		}
	}()

	ZZTMain()

	pitTicker.Stop()
	tickerDone <- true
}

//go:embed ascii.chr
var charsetData []byte
var textBuffer = make([]byte, 4000)
var textColumns int = 80
var blinkState bool = false

var palette = []int{
	0x000000,
	0x0000AA,
	0x00AA00,
	0x00AAAA,
	0xAA0000,
	0xAA00AA,
	0xAA5500,
	0xAAAAAA,
	0x555555,
	0x5555FF,
	0x55FF55,
	0x55FFFF,
	0xFF5555,
	0xFF55FF,
	0xFFFF55,
	0xFFFFFF,
}

func createBytesJS(data []byte) js.Value {
	dst := js.Global().Get("Uint8Array").New(len(data))
	js.CopyBytesToJS(dst, data)
	return dst
}

func IVideoSetMode(columns int) {
	js.Global().Get("ozg_setCharset").Invoke(8, 14, createBytesJS(charsetData))
	// js.Global().Get("ozg_setPalette").Invoke(palette)
	textColumns = columns
}

func IVideoClrScr(backgroundColor uint8) {
	for iy := 0; iy < 25; iy++ {
		for ix := 0; ix < textColumns; ix++ {
			textBuffer[iy*160+ix*2] = ' '
			textBuffer[iy*160+ix*2+1] = backgroundColor << 4
		}
	}
	VideoUpdateRequested.Store(true)
}

func IVideoWriteText(x, y int16, color byte, text string) {
	for i := 0; i < len(text); i++ {
		textBuffer[y*160+x*2] = text[i]
		textBuffer[y*160+x*2+1] = color
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
}

func IVideoSetCursorVisible(v bool) {
	// stub
}

func VideoMove(x, y, width int16, buffer *[]byte, toVideo bool) {
	if toVideo {
		if buffer != nil {
			if width > int16(len(*buffer)>>1) {
				width = int16(len(*buffer) >> 1)
			}
			for i := 0; i < int(width)*2; i++ {
				textBuffer[int(y)*160+int(x)*2+i] = (*buffer)[i]
			}
		}
		VideoUpdateRequested.Store(true)
	} else {
		*buffer = make([]byte, width*2)
		copy(*buffer, textBuffer[int(y)*160+int(x)*2:int(y)*160+(int(x)+int(width))*2])
	}
}
