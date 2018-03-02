package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/veandco/go-sdl2/sdl"
	"github.com/wilmerpaulino/chip8"
)

const (
	windowTitle = "chip8"
	pixelSize   = 16
)

var (
	keyMap = map[sdl.Keycode]int{
		sdl.K_1: 0x1,
		sdl.K_2: 0x2,
		sdl.K_3: 0x3,
		sdl.K_4: 0xc,
		sdl.K_q: 0x4,
		sdl.K_w: 0x5,
		sdl.K_e: 0x6,
		sdl.K_r: 0xd,
		sdl.K_a: 0x7,
		sdl.K_s: 0x8,
		sdl.K_d: 0x9,
		sdl.K_f: 0xe,
		sdl.K_z: 0xa,
		sdl.K_x: 0x0,
		sdl.K_c: 0xb,
		sdl.K_v: 0xf,
	}

	_ chip8.Renderer = (*sdlRenderer)(nil)
)

type sdlRenderer struct {
	*sdl.Renderer
}

func createSdlRenderer(window *sdl.Window, flags uint32) (*sdlRenderer, error) {
	r, err := sdl.CreateRenderer(window, -1, flags)
	if err != nil {
		return nil, err
	}

	return &sdlRenderer{r}, nil
}

func (r *sdlRenderer) Render(display chip8.Display) error {
	for x := 0; x < chip8.DisplayWidth; x++ {
		for y := 0; y < chip8.DisplayHeight; y++ {
			rect := &sdl.Rect{
				X: int32(x * pixelSize),
				Y: int32(y * pixelSize),
				W: pixelSize,
				H: pixelSize,
			}

			color := display[y][x] * 0xff
			r.SetDrawColor(color, color, color, color)

			if err := r.FillRect(rect); err != nil {
				return fmt.Errorf("unable to draw rect: %v", err)
			}
		}
	}

	r.Present()

	return nil
}

func (r *sdlRenderer) Beep() error {
	// TODO: Implement beep.
	return nil
}

func vmMain() error {
	romPath := flag.String("rom", "", "path to ROM file")

	flag.Parse()

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
		os.Exit(1)
	}

	if err := sdl.Init(sdl.INIT_VIDEO); err != nil {
		return fmt.Errorf("unable to initialize sdl: %v", err)
	}
	defer sdl.Quit()

	window, err := sdl.CreateWindow(
		windowTitle,
		sdl.WINDOWPOS_UNDEFINED,
		sdl.WINDOWPOS_UNDEFINED,
		chip8.DisplayWidth*pixelSize,
		chip8.DisplayHeight*pixelSize,
		sdl.WINDOW_SHOWN,
	)
	if err != nil {
		return fmt.Errorf("unable to create window: %v", err)
	}
	defer window.Destroy()

	renderer, err := createSdlRenderer(window, sdl.RENDERER_ACCELERATED)
	if err != nil {
		return fmt.Errorf("unable to create window renderer: %v", err)
	}
	defer renderer.Destroy()

	if *romPath == "" {
		flag.Usage()
	}

	rom, err := ioutil.ReadFile(*romPath)
	if err != nil {
		return fmt.Errorf("failed to read from file: %v", err)
	}

	vm := chip8.New(renderer)
	if err := vm.LoadROM(rom); err != nil {
		return fmt.Errorf("failed to load rom: %v", err)
	}

	vm.Start()
	defer vm.Stop()

out:
	for {
		for {
			e := sdl.PollEvent()
			if e == nil {
				break
			}

			switch e := e.(type) {
			case *sdl.QuitEvent:
				break out
			case *sdl.KeyboardEvent:
				keycode := e.Keysym.Sym
				switch e.Type {
				case sdl.KEYDOWN:
					vm.PressKey(keyMap[keycode])
				case sdl.KEYUP:
					vm.ReleaseKey(keyMap[keycode])
				}
			}
		}
	}

	return nil
}

func main() {
	if err := vmMain(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
