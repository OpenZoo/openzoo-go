/**
 * Copyright (c) 2018, 2019, 2020, 2021 Adrian Siekierka
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 * SOFTWARE.
 */

import { time_ms } from "./util.js";

let audioCtx = undefined;

document.addEventListener('mousedown', function(event) {
	if (audioCtx == undefined) {
		audioCtx = new (window.AudioContext || window.webkitAudioContext) ();
	}
});

document.addEventListener('keydown', function(event) {
	if (audioCtx == undefined) {
		audioCtx = new (window.AudioContext || window.webkitAudioContext) ();
	}
});

export class BufferBasedAudio {
	constructor(emu, options) {
		this.emu = emu;
		// this.sampleRate = (options && options.sampleRate) || 48000;
		// this.bufferSize = (options && options.bufferSize) || 2048;
		this.sampleRate = 48000;
		this.bufferSize = 2048;
		this.nativeBuffer = new Uint8Array(this.bufferSize);
		this.volume = Math.min(1.0, Math.max(0.0, (options && options.volume) || 0.2));
		this.timeUnit = (this.bufferSize / this.sampleRate);
		this.initialized = false;
	}

	_queueBufferSource(populateFunc) {
		this.time += this.timeUnit;
		if (this.time < audioCtx.currentTime) this.time = audioCtx.currentTime;

		const source = audioCtx.createBufferSource();
		const buffer = audioCtx.createBuffer(1, this.bufferSize, this.sampleRate);
		if (populateFunc) populateFunc(buffer, source);
		source.buffer = buffer; // Firefox makes buffer immutable here! :^)
		source.onended = () => this._queueNextSpeakerBuffer();
		source.connect(audioCtx.destination);
		source.start(this.time);
	}

	_queueNextSpeakerBuffer() {
		const self = this;

		this._queueBufferSource((buffer, source) => {
			const bufferSize = self.bufferSize;
			const nativeBuffer = self.nativeBuffer;

			const out0 = buffer.getChannelData(0);
			ozg_audioCallback(nativeBuffer);
			// let nativeBuffer = ozg_audioCallback();
			for (let i = 0; i < bufferSize; i++) {
				out0[i] = (nativeBuffer[i] - 128) / 128.0;
			}
			for (var channel = 1; channel < buffer.numberOfChannels; channel++) {
				buffer.getChannelData(channel).set(out0);
			}
		})
	}

	_initSpeaker() {
		if (this.initialized) return true;
		if (audioCtx == undefined) return false;
		this.initialized = true;

		this.time = audioCtx.currentTime;

		this._queueBufferSource(() => {});
		this._queueBufferSource(() => {});
		this._queueNextSpeakerBuffer();
		return true;
	}

	setVolume(volume) {
		this.volume = Math.min(1.0, Math.max(0.0, volume));
		if (this.initialized) {
			// this.emu._audio_stream_set_volume(Math.floor(this.volume * this.emu._audio_stream_get_max_volume()));
		}
	}
}
