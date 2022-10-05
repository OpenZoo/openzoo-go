//go:build sdl2

package main

// typedef unsigned char Uint8;
// void ZooSdlAudioCallback(void *userdata, Uint8 *stream, int len);
import "C"
import (
	_ "embed"
	"reflect"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"

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
var VideoRenderer *sdl.Renderer
var VideoZTexture *sdl.Texture
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
		case *sdl.TextInputEvent:
			ParseSDLTextInputEvent(e)
		}
	}
}

//export ZooSdlAudioCallback
func ZooSdlAudioCallback(userdata unsafe.Pointer, stream *C.Uint8, length C.int) {
	n := int(length)
	hdr := reflect.SliceHeader{Data: uintptr(unsafe.Pointer(stream)), Len: n, Cap: n}
	buf := *(*[]byte)(unsafe.Pointer(&hdr))

	CurrentAudioSimulator.Simulate(buf)
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

	VideoRenderer, err = sdl.CreateRenderer(VideoWindow, -1, 0)
	if err != nil {
		panic(err)
	}
	defer VideoRenderer.Destroy()

	VideoZTexture, err = VideoRenderer.CreateTexture(uint32(sdl.PIXELFORMAT_ABGR32), sdl.TEXTUREACCESS_STREAMING, 640, 350)
	if err != nil {
		panic(err)
	}
	defer VideoZTexture.Destroy()

	/* file, _ := os.Create("./cpu.pprof")
	pprof.StartCPUProfile(file)
	defer pprof.StopCPUProfile() */

	frameTicker := time.NewTicker(16666667 * time.Nanosecond)
	pitTicker := time.NewTicker(55 * time.Millisecond)
	blinkTicker := time.NewTicker(266666667 * time.Nanosecond)
	tickerDone := make(chan bool)

	CurrentAudioSimulator = NewAudioSimulatorNearest(48000, byte(32))
	audioSpec := sdl.AudioSpec{
		Freq:     48000,
		Format:   sdl.AUDIO_U8,
		Channels: 1,
		Samples:  2048,
		Callback: sdl.AudioCallback(C.ZooSdlAudioCallback),
	}
	err = sdl.OpenAudio(&audioSpec, nil)
	if err != nil {
		panic(err)
	}
	defer sdl.CloseAudio()
	sdl.PauseAudio(false)

	sdl.StartTextInput()
	defer sdl.StopTextInput()

	go func() {
		for {
			select {
			case <-blinkTicker.C:
				MainThreadAsync(ZooSdlToggleBlinkChars)
			case <-frameTicker.C:
				MainThreadAsync(updateSdlEvents)
				MainThreadAsync(func() {
					VideoRenderer.Copy(VideoZTexture, nil, nil)
					VideoRenderer.Present()
				})
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

	go func() {
		ZZTMain()

		frameTicker.Stop()
		pitTicker.Stop()
		blinkTicker.Stop()
		tickerDone <- true

		close(mainTaskQueue)
	}()

	for f := range mainTaskQueue {
		f()
	}
}
