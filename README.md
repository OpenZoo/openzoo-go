# OpenZoo/Go

After C, C++ and Free Pascal, comes an attempt at a port of ZZT in Go.

## Compiling

### SDL2 (Desktop)

Dependencies:

  * Latest version of Go
  * Latest version of SDL2 (on Windows, adding SDL2.dll to the build directory is sufficient)

Commands:

    $ go generate
    $ go build -x -tags sdl2

### WebAssembly (Web, Go)

Dependencies:

  * Latest version of Go
  * Recent version of Node.js

Commands:

    $ mkdir out
    $ cd web
    $ npm run build && cp zeta.js ../out/zeta.min.js && cp res/* ../out/
    $ cd ..
    $ GOOS=js GOARCH=wasm go build -tags wasm -o out/openzoo-go.wasm

Tricks to make the binary a little smaller:

  * Add `-trimpath -ldflags="-s -w"` at the end of the `go build` comamnd to remove/reduce debugging information.
  * Replace `go build` with `garble -tiny build` to remove debugging information/stack traces even further. (Install instructions for garble [here](https://github.com/burrowers/garble).)
  * Use `wasm-opt -Oz -o out/openzoo-go-optimized.wasm out/openzoo-go.wasm` to run a separate optimization pass on the WASM binary.

### WebAssembly (Web, TinyGo)

Warning: This version generates much smaller binaries, but currently runs into high GC pauses.

Dependencies:

  * Latest version of TinyGo (>= 0.26.0)
  * Recent version of Node.js

Commands:

    $ mkdir out
    $ cd web
    $ npm run build && cp zeta.js ../out/zeta.min.js && cp res/* res_tinygo/* ../out/
    $ cd ..
    $ tinygo build -target wasm -wasm-abi js -scheduler asyncify -opt z -llvm-features "+bulk-memory"
    $ cp openzoo-go.wasm out/

