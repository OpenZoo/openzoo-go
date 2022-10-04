//go:build sdl2

package main

import (
	_ "embed"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/veandco/go-sdl2/sdl"
)

type IdleMode int

const (
	IMUntilPit IdleMode = iota
	IMUntilFrame
)

var mainTaskQueue = make(chan func())
var FrameTickCond = sync.NewCond(&sync.Mutex{})
var PitTickCond = sync.NewCond(&sync.Mutex{})
var VideoWindow *sdl.Window
var VideoSurface *sdl.Surface
var VideoUpdateRequested atomic.Bool
var timerTicks int

func MainThreadAsync(f func()) {
	mainTaskQueue <- f
}

func MainThreadSync(f func()) {
	done := make(chan bool, 1)
	mainTaskQueue <- func() {
		f()
		done <- true
	}
	<-done
}

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
	switch mode {
	case IMUntilFrame:
		FrameTickCond.L.Lock()
		FrameTickCond.Wait()
		FrameTickCond.L.Unlock()
	case IMUntilPit:
		PitTickCond.L.Lock()
		PitTickCond.Wait()
		PitTickCond.L.Unlock()
	}
}

func Delay(ms uint32) {
	// sdl.Delay(ms)
	time.Sleep(time.Duration(ms) * time.Millisecond)
}

func updateSdlEvents() {
	for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
		switch e := event.(type) {
		case *sdl.KeyboardEvent:
			ParseSDLKeyboardEvent(e)
			break
		}
	}
}

func main() {
	var err error
	runtime.LockOSThread()
	if runtime.NumCPU() > 2 {
		runtime.GOMAXPROCS(2)
	}

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

	/* file, _ := os.Create("./cpu.pprof")
	pprof.StartCPUProfile(file)
	defer pprof.StopCPUProfile() */

	frameTicker := time.NewTicker(16667 * time.Microsecond)
	pitTicker := time.NewTicker(55 * time.Millisecond)
	tickerDone := make(chan bool)

	go func() {
		for {
			select {
			case <-frameTicker.C:
				MainThreadSync(updateSdlEvents)
				if VideoUpdateRequested.Swap(false) {
					MainThreadSync(func() {
						VideoWindow.UpdateSurface()
					})
				}
				FrameTickCond.Broadcast()
			case <-pitTicker.C:
				timerTicks++
				PitTickCond.Broadcast()
			case <-tickerDone:
				return
			}
		}
	}()

	go func() {
		ZZTMain()
		close(mainTaskQueue)
	}()

	for f := range mainTaskQueue {
		f()
	}

	frameTicker.Stop()
	pitTicker.Stop()
	tickerDone <- true
}
