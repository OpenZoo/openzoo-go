package main

import (
	"runtime"
	"sync"
)

const (
	freqSilence   int = iota
	freqTruncated     /* DOS ZZT behaviour */
	freqExact         /* ClassicZoo "HQ audio" behaviour */

	samplesNoteDelay = 16
	pitDivisor       = 1193182
)

var CurrentAudioSimulator *AudioSimulatorState

type audioSimulatorRenderer interface {
	setVolume(v byte)
	emitSilenceToEnd(samples []byte, streamPos int)
	emitNote(a *AudioSimulatorState, targetPos int, frequency uint32, emitType int, samples []byte, streamPos *int)
}

type AudioSimulatorState struct {
	SimulationAllowed bool

	mutex           sync.Mutex
	currentNote     int
	currentNotePos  int
	currentNoteMax  int
	buffer          string
	bufferPos       int
	bufferStopTicks int
	frequency       int
	samplesPerDrum  int
	samplesPerPit   int
	volume          byte
	renderer        audioSimulatorRenderer
}

func (a *AudioSimulatorState) calcJump(targetNotePos int, streamPos int, streamLen int) int {
	maxTargetChange := targetNotePos - a.currentNotePos
	if maxTargetChange < 0 {
		return 0
	}
	maxStreamChange := streamLen - streamPos
	if maxTargetChange < maxStreamChange {
		return maxTargetChange
	} else {
		return maxStreamChange
	}
}

func (a *AudioSimulatorState) jumpBy(amount int, streamPos *int) {
	a.currentNotePos += amount
	*streamPos += amount
	if a.currentNotePos >= a.currentNoteMax {
		a.currentNote = -1
	}
}

func (a *AudioSimulatorState) Queue(pattern string, clear bool) {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	if clear {
		a.buffer = pattern
	} else {
		a.buffer = a.buffer[a.bufferPos:] + pattern
	}

	a.bufferPos = 0
	a.bufferStopTicks = TimerTicks() + SoundCountTicks(pattern)
}

type AudioSimulatorRendererNearest struct {
	volumeMin byte
	volumeMax byte
}

func newAudioSimulatorState(frequency int) *AudioSimulatorState {
	return &AudioSimulatorState{
		frequency:      frequency,
		samplesPerDrum: (frequency + 500) / 1000,
		samplesPerPit:  (frequency*55 + 500) / 1000,
	}
}

func NewAudioSimulatorNearest(frequency int, volume byte) *AudioSimulatorState {
	a := newAudioSimulatorState(frequency)
	a.renderer = &AudioSimulatorRendererNearest{}
	a.SetVolume(volume)
	a.Clear()
	return a
}

func (a *AudioSimulatorState) Clear() {
	a.currentNote = -1
}

func (a *AudioSimulatorState) GetVolume() byte {
	return a.volume
}

func (a *AudioSimulatorRendererNearest) setVolume(v byte) {
	a.volumeMin = 128 - (v >> 2)
	a.volumeMax = 128 + (v >> 2)
}

func (a *AudioSimulatorState) SetVolume(v byte) {
	a.volume = v
	a.renderer.setVolume(v)
}

func (a *AudioSimulatorState) OnPitTick() {
	a.SimulationAllowed = SoundIsPlaying
}

func (a *AudioSimulatorRendererNearest) emitSilenceToEnd(samples []byte, streamPos int) {
	for i := streamPos; i < len(samples); i++ {
		samples[i] = 128
	}
}

func (a *AudioSimulatorRendererNearest) emitNote(as *AudioSimulatorState, targetPos int, frequency uint32, freqType int, samples []byte, streamPos *int) {
	iMax := as.calcJump(targetPos, *streamPos, len(samples))
	if freqType == freqSilence {
		for i := 0; i < iMax; i++ {
			samples[*streamPos+i] = 128
		}
	} else {
		var samplesPerChange uint32
		if freqType == freqTruncated {
			samplesPerChange = uint32((uint64(as.frequency*256) * uint64(pitDivisor/(frequency>>8))) / pitDivisor)
		} else {
			samplesPerChange = uint32(uint64(as.frequency*65536) / uint64(frequency))
		}
		for i := 0; i < iMax; i++ {
			var samplePos uint32 = uint32(as.currentNotePos+i) << 8
			if (samplePos % samplesPerChange) < (samplesPerChange >> 1) {
				samples[*streamPos+i] = a.volumeMin
			} else {
				samples[*streamPos+i] = a.volumeMax
			}
		}
	}
	as.jumpBy(iMax, streamPos)
}

func (a *AudioSimulatorState) Simulate(samples []byte) {
	freqType := freqExact // TODO: support "low quality"
	if runtime.GOOS == "js" {
		freqType = freqTruncated
	}

	slen := len(samples)
	a.mutex.Lock()
	defer a.mutex.Unlock()

	if !SoundEnabled || !SoundIsPlaying || !a.SimulationAllowed {
		a.currentNote = -1
		a.renderer.emitSilenceToEnd(samples, 0)
	} else {
		pos := 0
		for pos < slen {
			if a.currentNote < 0 {
				if a.bufferPos >= len(a.buffer) {
					SoundIsPlaying = false
					a.renderer.emitSilenceToEnd(samples, pos)
					break
				} else {
					// pop note
					a.currentNote = int(byte(a.buffer[a.bufferPos]))
					a.currentNotePos = 0
					a.currentNoteMax = int(SoundDurationMultiplier) * int(byte(a.buffer[a.bufferPos+1])) * a.samplesPerPit
					a.bufferPos += 2
				}
			}

			if a.currentNote >= 0 && a.currentNote < 240 {
				// note
				if a.currentNotePos < samplesNoteDelay {
					a.renderer.emitNote(a, samplesNoteDelay, 0, freqSilence, samples, &pos)
				} else {
					if SoundFreqTable[a.currentNote] >= 256 {
						a.renderer.emitNote(a, a.currentNoteMax, SoundFreqTable[a.currentNote], freqType, samples, &pos)
					} else {
						a.renderer.emitNote(a, a.currentNoteMax, 0, freqSilence, samples, &pos)
					}
				}
			} else if a.currentNote >= 240 && a.currentNote < 250 {
				// silence
				drum := SoundDrumTable[a.currentNote-240]
				drumPos := a.currentNotePos / a.samplesPerDrum
				if drumPos < len(drum) {
					a.renderer.emitNote(a, (drumPos+1)*a.samplesPerDrum, uint32(drum[drumPos])<<8, freqTruncated, samples, &pos)
				} else {
					a.renderer.emitNote(a, a.currentNoteMax, 0, freqSilence, samples, &pos)
				}
			} else {
				// unknown
				a.renderer.emitNote(a, a.currentNoteMax, 0, freqSilence, samples, &pos)
			}
		}
	}
}
