# chip8

[![MIT license](https://img.shields.io/badge/license-MIT-blue.svg)](https://github.com/lightningnetwork/lnd/blob/master/LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/wilmerpaulino/chip8)](https://goreportcard.com/report/github.com/wilmerpaulino/chip8)
[![GoDoc](https://godoc.org/github.com/wilmerpaulino/chip8?status.svg)](https://godoc.org/github.com/wilmerpaulino/chip8?status.svg)

chip8 is a graphics library independent implementation of the CHIP-8 virtual machine.

An example is included in [`cmd/chip8`](https://github.com/wilmerpaulino/chip8/tree/master/cmd/chip8) using the [SDL2](https://www.libsdl.org/index.php) library.

## Install Library

```bash
$ go get -u github.com/wilmerpaulino/chip8
```

## Import Library

```go
import "github.com/wilmerpaulino/chip8"
```

## Install SDL2 CHIP-8 Implementation

First, you'll need to install the SDL2 library. You can do so [here](https://www.libsdl.org/download-2.0.php).

It may also be available in your operating system's package manager. For example, on macOS:

```bash
$ brew install sdl2
```

Once SDL2 is installed, you can install the CHIP-8 emulator to your `$GOPATH`. Make sure it is set beforehand.

```bash
$ go get -u github.com/wilmerpaulino/chip8/cmd/chip8
```

## Run SDL2 CHIP-8 Implementation

Once installed, it can be run directly from your `$GOPATH`:
```bash
$ $GOPATH/bin/chip8 -rom $GOPATH/src/wilmerpaulino/chip8/roms/PONG2
```

## License

This project is distributed under the [MIT license](https://github.com/wilmerpaulino/chip8/blob/master/LICENSE).
