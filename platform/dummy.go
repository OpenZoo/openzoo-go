//go:build dummy

package platform

var KeysLeftShiftHeld = false
var KeysRightShiftHeld = false
var KeysShiftHeld = false
var KeysCtrlHeld = false
var KeysAltHeld = false

type IdleMode int

const (
	IMUntilPit IdleMode = iota
	IMUntilFrame
)

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

func NoSound() {
	// no-op
}

func Sound(freq uint16) {
	// no-op
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

func VideoInstall(columns int) {
	// no-op
}

func VideoClrScr(backgroundColor uint8) {
	// no-op
}

func VideoWriteText(x, y int16, color byte, text string) {
	// no-op
}

func VideoSetCursorVisible(v bool) {
	// no-op
}

func VideoUninstall() {
	// no-op
}

func VideoMove(x, y, width int16, buffer *[]byte, toVideo bool) {
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

func PlatformMain(mainFunc func()) {
	mainFunc()
}
