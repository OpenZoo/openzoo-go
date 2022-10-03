package main // unit: Keys

import "github.com/OpenZoo/openzoo-go/platform"

var (
	KeysRightShiftHeld bool
	KeysLeftShiftHeld  bool
	KeysShiftHeld      bool
	KeysCtrlHeld       bool
	KeysAltHeld        bool
)

// implementation uses: Dos

func KeysUpdateModifiers() {
	KeysRightShiftHeld = platform.KeysRightShiftHeld
	KeysLeftShiftHeld = platform.KeysLeftShiftHeld
	KeysShiftHeld = platform.KeysShiftHeld
	KeysCtrlHeld = platform.KeysCtrlHeld
	KeysAltHeld = platform.KeysAltHeld
}
