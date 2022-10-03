package main // unit: Sounds

import (
	"math"
	"strings"

	"github.com/OpenZoo/openzoo-go/platform"
)

type TDrumData struct {
	Len  int16
	Data [255]uint16
}

var (
	SoundEnabled            bool
	SoundBlockQueueing      bool
	SoundCurrentPriority    int16
	SoundFreqTable          [255]uint16
	SoundDurationMultiplier byte
	SoundDurationCounter    byte
	SoundBuffer             string
	SoundBufferPos          int16
	SoundIsPlaying          bool
	SoundDrumTable          [10]TDrumData
)

// implementation uses: Crt, Dos

func SoundQueue(priority int16, pattern string) {
	if !SoundBlockQueueing && (!SoundIsPlaying || (priority >= SoundCurrentPriority && SoundCurrentPriority != -1 || priority == -1)) {
		if priority >= 0 || !SoundIsPlaying {
			SoundCurrentPriority = priority
			SoundBuffer = pattern
			SoundBufferPos = 1
			SoundDurationCounter = 1
		} else {
			SoundBuffer = Copy(SoundBuffer, SoundBufferPos, Length(SoundBuffer)-SoundBufferPos+1)
			SoundBufferPos = 1
			if Length(SoundBuffer)+Length(pattern) < 255 {
				SoundBuffer += pattern
			}
		}
		SoundIsPlaying = true
	}
}

func SoundClearQueue() {
	SoundBuffer = ""
	SoundIsPlaying = false
	NoSound()
}

func SoundInitFreqTable() {
	var (
		octave, note               int16
		freqC1, noteStep, noteBase float64
	)
	freqC1 = 32.0
	noteStep = math.Exp(math.Ln2 / 12.0)
	for octave = 1; octave <= 15; octave++ {
		noteBase = math.Exp(float64(octave)*math.Ln2) * freqC1
		for note = 0; note <= 11; note++ {
			SoundFreqTable[octave*16+note-1] = uint16(math.Floor(noteBase))
			noteBase = noteBase * noteStep
		}
	}
}

func SoundInitDrumTable() {
	var i int16
	SoundDrumTable[0].Len = 1
	SoundDrumTable[0].Data[0] = 3200
	for i = 1; i <= 9; i++ {
		SoundDrumTable[i].Len = 14
	}
	for i = 1; i <= 14; i++ {
		SoundDrumTable[1].Data[i-1] = uint16(i*100 + 1000)
	}
	for i = 1; i <= 16; i++ {
		SoundDrumTable[2].Data[i-1] = uint16(i%2*1600 + 1600 + i%4*1600)
	}
	for i = 1; i <= 14; i++ {
		SoundDrumTable[4].Data[i-1] = uint16(Random(5000) + 500)
	}
	for i = 1; i <= 8; i++ {
		SoundDrumTable[5].Data[i*2-1-1] = 1600
		SoundDrumTable[5].Data[i*2-1] = uint16(Random(1600) + 800)
	}
	for i = 1; i <= 14; i++ {
		SoundDrumTable[6].Data[i-1] = uint16(i%2*880 + 880 + i%3*440)
	}
	for i = 1; i <= 14; i++ {
		SoundDrumTable[7].Data[i-1] = uint16(700 - i*12)
	}
	for i = 1; i <= 14; i++ {
		SoundDrumTable[8].Data[i-1] = uint16(i*20 + 1200 - Random(i*40))
	}
	for i = 1; i <= 14; i++ {
		SoundDrumTable[9].Data[i-1] = uint16(Random(440) + 220)
	}
}

func SoundPlayDrum(drum *TDrumData) {
	var i int16
	for i = 1; i <= drum.Len; i++ {
		Sound(drum.Data[i-1])
		Delay(1)
	}
	NoSound()
}

func SoundHasTimeElapsed(counter *int16, duration int16) (SoundHasTimeElapsed bool) {
	var (
		hSecsDiff  uint16
		hSecsTotal int16
	)
	hSecsTotal = int16(platform.TimerTicks() * 11 / 2)
	hSecsDiff = uint16(hSecsTotal - *counter)
	if hSecsDiff >= uint16(duration) {
		SoundHasTimeElapsed = true
		*counter = hSecsTotal
	} else {
		SoundHasTimeElapsed = false
	}
	return
}

func SoundTimerHandler() {
	if !SoundEnabled {
		SoundIsPlaying = false
		NoSound()
	} else if SoundIsPlaying {
		SoundDurationCounter--
		if SoundDurationCounter <= 0 {
			NoSound()
			if SoundBufferPos >= Length(SoundBuffer) {
				NoSound()
				SoundIsPlaying = false
			} else {
				if SoundBuffer[SoundBufferPos-1] == '\x00' {
					NoSound()
				} else if SoundBuffer[SoundBufferPos-1] < '\xf0' {
					Sound(SoundFreqTable[SoundBuffer[SoundBufferPos-1]-1])
				} else {
					SoundPlayDrum(&SoundDrumTable[SoundBuffer[SoundBufferPos-1]-240])
				}

				SoundBufferPos++
				SoundDurationCounter = SoundDurationMultiplier * SoundBuffer[SoundBufferPos-1]
				SoundBufferPos++
			}
		}
	}

}

func SoundUninstall() {
	// stub SetIntVec(0x1C, SoundOldVector)
}

func SoundParse(input string) (SoundParse string) {
	var (
		noteOctave   int16
		noteDuration int16
		output       strings.Builder
		noteTone     int16
	)
	AdvanceInput := func() {
		input = Copy(input, 2, Length(input)-1)
	}

	noteOctave = 3
	noteDuration = 1
	for Length(input) != 0 {
		noteTone = -1
		switch UpCase(input[0]) {
		case 'T':
			noteDuration = 1
			AdvanceInput()
		case 'S':
			noteDuration = 2
			AdvanceInput()
		case 'I':
			noteDuration = 4
			AdvanceInput()
		case 'Q':
			noteDuration = 8
			AdvanceInput()
		case 'H':
			noteDuration = 16
			AdvanceInput()
		case 'W':
			noteDuration = 32
			AdvanceInput()
		case '.':
			noteDuration = noteDuration * 3 / 2
			AdvanceInput()
		case '3':
			noteDuration = noteDuration / 3
			AdvanceInput()
		case '+':
			if noteOctave < 6 {
				noteOctave++
			}
			AdvanceInput()
		case '-':
			if noteOctave > 1 {
				noteOctave--
			}
			AdvanceInput()
		case 'A', 'B', 'C', 'D', 'E', 'F', 'G':
			switch UpCase(input[0]) {
			case 'C':
				noteTone = 0
				AdvanceInput()
			case 'D':
				noteTone = 2
				AdvanceInput()
			case 'E':
				noteTone = 4
				AdvanceInput()
			case 'F':
				noteTone = 5
				AdvanceInput()
			case 'G':
				noteTone = 7
				AdvanceInput()
			case 'A':
				noteTone = 9
				AdvanceInput()
			case 'B':
				noteTone = 11
				AdvanceInput()
			}
			if len(input) > 0 {
				switch UpCase(input[0]) {
				case '!':
					noteTone--
					AdvanceInput()
				case '#':
					noteTone++
					AdvanceInput()
				}
			}
			output.WriteByte(byte(noteOctave*0x10 + noteTone))
			output.WriteByte(byte(noteDuration))
		case 'X':
			output.WriteByte(0x00)
			output.WriteByte(byte(noteDuration))
			AdvanceInput()
		case '0', '1', '2', '4', '5', '6', '7', '8', '9':
			output.WriteByte(input[0] + 0xF0 - '0')
			output.WriteByte(byte(noteDuration))
			AdvanceInput()
		default:
			AdvanceInput()
		}
	}
	SoundParse = output.String()
	return
}

func init() {
	SoundInitFreqTable()
	SoundInitDrumTable()
	SoundEnabled = true
	SoundBlockQueueing = false
	SoundClearQueue()
	SoundDurationMultiplier = 1
	SoundIsPlaying = false
}
