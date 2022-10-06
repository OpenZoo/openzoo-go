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

### WebAssembly (Web)

Dependencies:

  * Recent version of Node.js
  * Latest version of TinyGo (>= 0.26.0)

Commands:

    $ mkdir out
    $ cd web
    $ npm run build && cp zeta.js ../out/zeta.min.js && cp res/* ../out/
    $ cd ..
    $ tinygo build -target wasm -wasm-abi js -scheduler asyncify -opt z -llvm-features "+bulk-memory"
    $ cp openzoo-go.wasm out/

