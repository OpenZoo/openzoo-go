//go:build dummy

package main

var timerTicks int = 0

func TimerTicks() int {
	// no-op
	timerTicks += 1
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
	// no-op
}

func Delay(ms uint32) {
	// no-op
}

func IVideoSetMode(columns int) {
	// no-op
}

func IVideoClrScr(backgroundColor uint8) {
	// no-op
}

func IVideoWriteText(x, y int16, color byte, text string) {
	// no-op
}

func IVideoSetCursorVisible(v bool) {
	// no-op
}

func VideoMove(x, y, width int16, buffer *[]byte, toVideo bool) {
	// no-op
}

func KeysUpdateModifiers() {
	// no-op
}

func KeyPressed() bool {
	// no-op
	return false
}

func ReadKey() byte {
	// no-op
	return 0
}

func main() {
	ZZTMain()
}
