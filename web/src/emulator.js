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

import { time_ms, drawErrorMessage } from "./util.js";
import { keymap, keychrmap } from "./keymap.js";

class Emulator {
    constructor(element, emu, render, audio, vfs, options) {
        this.element = element;
        this.emu = emu;
        this.render = render;
        this.audio = audio;
        this.vfs = vfs;

        this.frameQueued = false;

        const self = this;

        var curr_keymod = 0x00
        var curr_key = -1

        var check_modifiers = function (event) {
            if (event.shiftKey === true) { curr_keymod |= 0x01; } else if (event.shiftKey === false) { curr_keymod &= ~0x01; }
            if (event.ctrltKey === true) { curr_keymod |= 0x04; } else if (event.ctrlKey === false) { curr_keymod &= ~0x04; }
            if (event.altKey === true) { curr_keymod |= 0x08; } else if (event.altKey === false) { curr_keymod &= ~0x08; }
        }

        document.addEventListener('keydown', function (event) {
            if (event.target != element) return false;
            let ret = false;

            check_modifiers(event);
            if (event.key == "Shift") { curr_keymod |= 0x01; }
            else if (event.key == "Control") { curr_keymod |= 0x04; }
            else if (event.key == "Alt" || event.key == "AltGraph") { curr_keymod |= 0x08; }
            else ret = false;

            let chr = (event.key.length == 1) ? event.key.charCodeAt(0) : (keychrmap[event.keyCode] || 0);
            let key = keymap[event.key] || 0;
            if (key >= 0x46 && key <= 0x53) chr = 0;
            if (chr > 0 || key > 0) {
                curr_key = (chr > 0 && chr < 127) ? chr : (key + 128);
                ret = true;
            }

            if (ret) {
                event.preventDefault();
            }
            return false;
        }, false);

        document.addEventListener('keyup', function (event) {
            if (event.target != element) return false;
            let ret = true;

            check_modifiers(event);
            if (event.key == "Shift") { curr_keymod &= ~0x01; }
            else if (event.key == "Control") { curr_keymod &= ~0x04; }
            else if (event.key == "Alt" || event.key == "AltGraph") { curr_keymod &= ~0x08; }
            else ret = false;

            var key = keymap[event.key] || 0;
            if (key > 0) {
                // emu._zzt_keyup(key);
                ret = true;
            }

            if (ret) {
                event.preventDefault();
            }
            return false;
        }, false);

        window["ozg_key"] = function (doRemove) {
            var key = curr_key;
            if (doRemove) curr_key = -1;
            return key;
        }

        window["ozg_setCharset"] = function(width, height, charset) {
            render.setCharset(width, height, charset);
        }

        window["ozg_setPalette"] = function(palette) {
            render.setPalette(palette);
        }

        var textBuffer = new Uint8Array(4000);

        console.log(emu)
        window["ozg_init"] = function() {
            var draw;

            draw = function() {
                let doRender = ozg_videoCopyTextBuffer(textBuffer, false);
                if (doRender) {
                    render.render(textBuffer, 3, 0);
                }
                window.requestAnimationFrame(draw);
                audio._initSpeaker();
            };
            window.requestAnimationFrame(draw);
        }

        this.element.addEventListener("mousemove", function (e) {
        });

        this.element.addEventListener("mousedown", function (e) {
            element.requestPointerLock();
        });

        this.element.addEventListener("mouseup", function (e) {
        });
    }

    setVolume(volume) {
        this.audio.setVolume(volume);
        return true;
    }

    getFile(filename) {
        return this.vfs.get(filename);
    }

    listFiles() {
        return this.vfs.list();
    }
}

export function runEmulator(render, audio, vfs, options) {
    const go = new Go();

    WebAssembly.instantiateStreaming(fetch("openzoo-go.wasm"), go.importObject).then((result) => {
        let emu = result.instance;
        const emuObj = new Emulator(options.render.canvas, emu, render(emu), audio ? audio(emu) : undefined, vfs, options);

        window["ozg_vfsRead"] = function(fn) {
            fn = fn.toUpperCase();
    
            let data = vfs.get(fn);
            return data ? Uint8Array.from(data) : null;
        }
    
        window["ozg_vfsList"] = function(path) {
            // TODO: not ignore path
    
            return vfs.list(key => true);
        }
    
        // TODO/GO: move to Go
        emuObj.render.setPalette(Array(
            0x000000,
            0x0000AA,
            0x00AA00,
            0x00AAAA,
            0xAA0000,
            0xAA00AA,
            0xAA5500,
            0xAAAAAA,
            0x555555,
            0x5555FF,
            0x55FF55,
            0x55FFFF,
            0xFF5555,
            0xFF55FF,
            0xFFFF55,
            0xFFFFFF
        ));
        
        go.run(emu);
    }).catch((err) => {
        console.error(err);
    });
}
