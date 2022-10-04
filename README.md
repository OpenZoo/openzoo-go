# OpenZoo/Go

After C, C++ and Free Pascal, comes an attempt at a port of ZZT in Go.

## Compiling

Dependencies:

  * Latest version of Go
  * Latest version of SDL2 (on Windows, adding SDL2.dll to the build directory is sufficient)

Commands:

    $ go generate
    $ go build -x -tags sdl2
