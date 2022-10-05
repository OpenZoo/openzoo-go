package main // unit: Sounds

import (
	"math"
	"strings"
)

var (
	SoundEnabled            bool
	SoundBlockQueueing      bool
	SoundCurrentPriority    int16
	SoundFreqTable          [255]uint32
	SoundDurationMultiplier byte
	SoundDurationCounter    byte
	SoundBuffer             string
	SoundBufferPos          int16
	SoundIsPlaying          bool
	SoundDrumTable          [10][]uint16
)

// implementation uses: Crt, Dos

func SoundQueue(priority int16, pattern string) {
	if !SoundBlockQueueing && (!SoundIsPlaying || (priority >= SoundCurrentPriority && SoundCurrentPriority != -1 || priority == -1)) {
		if priority >= 0 || !SoundIsPlaying {
			SoundCurrentPriority = priority
			SoundBuffer = pattern
			SoundBufferPos = 1
			SoundDurationCounter = 1
			if CurrentAudioSimulator != nil {
				CurrentAudioSimulator.Queue(pattern, true)
			}
		} else {
			SoundBuffer = Copy(SoundBuffer, SoundBufferPos, Length(SoundBuffer)-SoundBufferPos+1)
			SoundBufferPos = 1
			if Length(SoundBuffer)+Length(pattern) < 255 {
				SoundBuffer += pattern
				if CurrentAudioSimulator != nil {
					CurrentAudioSimulator.Queue(pattern, false)
				}
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

func SoundInitFreqTable(highQuality bool) {
	var (
		octave, note               int16
		freqC1, noteStep, noteBase float64
	)
	freqC1 = 32.0
	noteStep = math.Exp(math.Ln2 / 12.0)
	for octave = 1; octave <= 15; octave++ {
		noteBase = math.Exp(float64(octave)*math.Ln2) * freqC1
		for note = 0; note <= 11; note++ {
			if highQuality {
				SoundFreqTable[octave*16+note] = uint32(math.Round(noteBase * 256))
			} else {
				SoundFreqTable[octave*16+note] = uint32(math.Floor(noteBase) * 256)
			}
			noteBase = noteBase * noteStep
		}
	}
}

func SoundInitDrumTable() {
	var i int16
	SoundDrumTable[0] = []uint16{3200}
	for i = 1; i <= 9; i++ {
		SoundDrumTable[i] = make([]uint16, 14)
	}
	for i = 1; i <= 14; i++ {
		SoundDrumTable[1][i-1] = uint16(i*100 + 1000)
	}
	for i = 1; i <= 14; i++ {
		SoundDrumTable[2][i-1] = uint16(i%2*1600 + 1600 + i%4*1600)
	}
	for i = 1; i <= 14; i++ {
		SoundDrumTable[4][i-1] = uint16(Random(5000) + 500)
	}
	for i = 1; i <= 7; i++ {
		SoundDrumTable[5][i*2-1-1] = 1600
		SoundDrumTable[5][i*2-1] = uint16(Random(1600) + 800)
	}
	Random(1600) // ensure RNG matches
	for i = 1; i <= 14; i++ {
		SoundDrumTable[6][i-1] = uint16(i%2*880 + 880 + i%3*440)
	}
	for i = 1; i <= 14; i++ {
		SoundDrumTable[7][i-1] = uint16(700 - i*12)
	}
	for i = 1; i <= 14; i++ {
		SoundDrumTable[8][i-1] = uint16(i*20 + 1200 - Random(i*40))
	}
	for i = 1; i <= 14; i++ {
		SoundDrumTable[9][i-1] = uint16(Random(440) + 220)
	}
}

func SoundPlayDrum(drum []uint16) {
	for i := 0; i < len(drum); i++ {
		Sound(drum[i])
		Delay(1)
	}
	NoSound()
}

func SoundHasTimeElapsed(counter *int16, duration int16) (SoundHasTimeElapsed bool) {
	var (
		hSecsDiff  uint16
		hSecsTotal int16
	)
	hSecsTotal = int16(TimerTicks() * 11 / 2)
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
					Sound(uint16(SoundFreqTable[SoundBuffer[SoundBufferPos-1]-1] >> 8))
				} else {
					SoundPlayDrum(SoundDrumTable[SoundBuffer[SoundBufferPos-1]-240])
				}

				SoundBufferPos++
				SoundDurationCounter = SoundDurationMultiplier * SoundBuffer[SoundBufferPos-1]
				SoundBufferPos++
			}
		}
	}

	if CurrentAudioSimulator != nil {
		CurrentAudioSimulator.OnPitTick()
	}
}

func SoundCountTicks(pattern string) int {
	ticks := 0
	for i := 1; i < len(pattern); i += 2 {
		ticks += int(byte(pattern[i]))
	}
	return ticks
}

func SoundParse(input string) (SoundParse string) {
	var (
		noteOctave   int16
		noteDuration int16
		output       strings.Builder
		noteTone     int16
	)
	noteOctave = 3
	noteDuration = 1
	for Length(input) != 0 {
		noteTone = -1
		switch UpCase(input[0]) {
		case 'T':
			noteDuration = 1
			input = input[1:]
		case 'S':
			noteDuration = 2
			input = input[1:]
		case 'I':
			noteDuration = 4
			input = input[1:]
		case 'Q':
			noteDuration = 8
			input = input[1:]
		case 'H':
			noteDuration = 16
			input = input[1:]
		case 'W':
			noteDuration = 32
			input = input[1:]
		case '.':
			noteDuration = noteDuration * 3 / 2
			input = input[1:]
		case '3':
			noteDuration = noteDuration / 3
			input = input[1:]
		case '+':
			if noteOctave < 6 {
				noteOctave++
			}
			input = input[1:]
		case '-':
			if noteOctave > 1 {
				noteOctave--
			}
			input = input[1:]
		case 'A', 'B', 'C', 'D', 'E', 'F', 'G':
			switch UpCase(input[0]) {
			case 'C':
				noteTone = 0
				input = input[1:]
			case 'D':
				noteTone = 2
				input = input[1:]
			case 'E':
				noteTone = 4
				input = input[1:]
			case 'F':
				noteTone = 5
				input = input[1:]
			case 'G':
				noteTone = 7
				input = input[1:]
			case 'A':
				noteTone = 9
				input = input[1:]
			case 'B':
				noteTone = 11
				input = input[1:]
			}
			if len(input) > 0 {
				switch UpCase(input[0]) {
				case '!':
					noteTone--
					input = input[1:]
				case '#':
					noteTone++
					input = input[1:]
				}
			}
			output.WriteByte(byte(noteOctave*0x10 + noteTone))
			output.WriteByte(byte(noteDuration))
		case 'X':
			output.WriteByte(0x00)
			output.WriteByte(byte(noteDuration))
			input = input[1:]
		case '0', '1', '2', '4', '5', '6', '7', '8', '9':
			output.WriteByte(input[0] + 0xF0 - '0')
			output.WriteByte(byte(noteDuration))
			input = input[1:]
		default:
			input = input[1:]
		}
	}
	SoundParse = output.String()
	return
}

func init() {
	SoundInitFreqTable(true)
	SoundInitDrumTable()
	SoundEnabled = true
	SoundBlockQueueing = false
	SoundClearQueue()
	SoundDurationMultiplier = 1
	SoundIsPlaying = false
}
