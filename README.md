# chip8

[![MIT license](https://img.shields.io/badge/license-MIT-blue.svg)](https://github.com/lightningnetwork/lnd/blob/master/LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/wilmerpaulino/chip8)](https://goreportcard.com/report/github.com/wilmerpaulino/chip8)
[![GoDoc](https://godoc.org/github.com/wilmerpaulino/chip8?status.svg)](https://godoc.org/github.com/wilmerpaulino/chip8?status.svg)

`chip8` is a graphics library-agnostic implementation of the CHIP-8 virtual machine. This allows one to render the CHIP-8's display as they wish.

An example is included in `cmd/chip8` using the [SDL2](https://www.libsdl.org/index.php) library.

## Install Library

```bash
$ go get -u github.com/wilmerpaulino/chip8
```

## Import Library

```go
import "github.com/wilmerpaulino/chip8"
```

## Install SDL2 CHIP-8 Implementation

```bash
$ go get -u github.com/wilmerpaulino/chip8/cmd/chip8
```

## Run SDL2 CHIP-8 Implementation

Once installed, it can be run directly from `$GOPATH`:
```bash
$ $GOPATH/bin/chip8 -rom $GOPATH/src/wilmerpaulino/chip8/roms/PONG2
```

## License

This project is distributed under the [MIT license](https://github.com/wilmerpaulino/chip8/blob/master/LICENSE).
