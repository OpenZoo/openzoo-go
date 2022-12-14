package main

// Pascal shims - System/misc. procedures

import (
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"golang.org/x/exp/constraints"
)

type PasChar interface {
	byte | rune
}

var RandSeed uint32 = 1

func PathBasenameWithoutExt(s string) string {
	b := filepath.Base(s)
	bDot := strings.Index(b, ".")
	if bDot > 0 {
		return b[0:bDot]
	} else {
		return b
	}

}

func Signum(val int16) (Signum int16) {
	if val > 0 {
		Signum = 1
	} else if val < 0 {
		Signum = -1
	} else {
		Signum = 0
	}

	return
}

func Difference(a, b int16) (Difference int16) {
	if a-b >= 0 {
		Difference = a - b
	} else {
		Difference = b - a
	}
	return
}

func Randomize() {
	RandSeed = uint32(time.Now().UnixMilli())
}

func Random[T constraints.Integer](max T) int16 {
	RandSeed = (RandSeed * 134775813) + 1
	return int16((RandSeed >> 16) % uint32(max))
}

func UpCase(s byte) byte {
	if s >= 'a' && s <= 'z' {
		return s - 0x20
	} else {
		return s
	}
}

func Length(s string) int16 {
	return int16(len(s))
}

func Pos(needleChr byte, s string) int16 {
	return int16(strings.IndexByte(s, needleChr) + 1)
}

func Chr(v byte) string {
	return string([]byte{v})
}

func Ord(v string) byte {
	return byte(v[0])
}

func Val(s string) int {
	i, err := strconv.Atoi(s)
	if err != nil {
		return 0
	} else {
		return i
	}
}

func BoolToInt(v bool) int {
	if v {
		return 1
	} else {
		return 0
	}
}

func Sqr[T constraints.Integer](a T) T {
	return a * a
}

func Str[T constraints.Integer](v T) string {
	return strconv.Itoa(int(v))
}

func StrWidth[T constraints.Integer](v T, minLength int) string {
	s := strconv.Itoa(int(v))
	if len(s) >= minLength {
		return s
	} else {
		return strings.Repeat(" ", minLength-len(s)) + s
	}
}

func Copy(s string, start int16, length int16) string {
	if length <= 0 {
		return ""
	}

	start -= 1

	if len(s) <= int(start+length) {
		return s[start:]
	} else {
		return s[start : start+length]
	}
}

func Abs[T constraints.Integer](v T) T {
	if v < 0 {
		return -v
	} else {
		return v
	}
}

func Clamp[T constraints.Integer](x, minX, maxX T) T {
	if x < minX {
		return minX
	} else if x > maxX {
		return maxX
	} else {
		return x
	}
}
