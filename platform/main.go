package platform

import (
	_ "embed"
	"runtime"
	"time"

	"github.com/veandco/go-sdl2/sdl"
)

type IdleMode int

const (
	IMUntilPit IdleMode = iota
	IMUntilFrame
)

var VideoWindow *sdl.Window
var VideoSurface *sdl.Surface
var timerTicks int

func TimerTicks() int {
	return timerTicks
}

func MemAvail() int32 {
	// stub
	return 655360
}

func NoSound() {
	// stub
}

func Sound(freq uint16) {
	// stub
}

func SetCBreak(v bool) {
	// stub
}

func Idle(mode IdleMode) {
	// stub
}

func Delay(ms uint32) {
	sdl.Do(func() {
		sdl.Delay(ms)
	})
}

func PlatformMain(mainFunc func()) {
	var err error
	runtime.LockOSThread()

	if err := sdl.Init(sdl.INIT_EVERYTHING); err != nil {
		panic(err)
	}
	defer sdl.Quit()

	VideoWindow, err = sdl.CreateWindow("OpenZoo/Go", sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED,
		640, 350, sdl.WINDOW_SHOWN)
	if err != nil {
		panic(err)
	}
	defer VideoWindow.Destroy()

	// TODO: VideoSurface, err = sdl.CreateRGBSurface(0, 640, 350, 32, 0, 0, 0, 0)
	VideoSurface, err = VideoWindow.GetSurface()
	if err != nil {
		panic(err)
	}

	pitTicker := time.NewTicker(55 * time.Millisecond)
	pitTickerDone := make(chan bool)

	go func() {
		for {
			select {
			case <-pitTicker.C:
				timerTicks++
			case <-pitTickerDone:
				return
			}
		}
	}()

	sdl.Main(mainFunc)

	pitTicker.Stop()
	pitTickerDone <- true
}
