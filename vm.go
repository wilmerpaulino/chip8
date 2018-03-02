package chip8

import (
	"bytes"
	"errors"
	"fmt"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"
)

const (
	// memorySize represents the size in bytes of the virtual machine's
	// memory.
	memorySize = 4096

	// memoryOffset is the starting offset of where programs should be
	// loaded in.
	memoryOffset = 512

	// numRegisters is the number of registers of the virtual machine's CPU.
	numRegisters = 16

	// numFrames is the number of stack frames available on the virtual
	// machine's stack.
	numFrames = 16

	// numKeys is the number of supported keys of the virtual machine.
	numKeys = 16

	// defaultClockSpeed is the default clock speed of the virtual machine's
	// CPU in hertz.
	defaultClockSpeed = time.Duration(60)
)

var (
	// font is the font used in the CHIP-8 virtual machine.
	font = []byte{
		0xF0, 0x90, 0x90, 0x90, 0xF0,
		0x20, 0x60, 0x20, 0x20, 0x70,
		0xF0, 0x10, 0xF0, 0x80, 0xF0,
		0xF0, 0x10, 0xF0, 0x10, 0xF0,
		0x90, 0x90, 0xF0, 0x10, 0x10,
		0xF0, 0x80, 0xF0, 0x10, 0xF0,
		0xF0, 0x80, 0xF0, 0x90, 0xF0,
		0xF0, 0x10, 0x20, 0x40, 0x40,
		0xF0, 0x90, 0xF0, 0x90, 0xF0,
		0xF0, 0x90, 0xF0, 0x10, 0xF0,
		0xF0, 0x90, 0xF0, 0x90, 0x90,
		0xE0, 0x90, 0xE0, 0x90, 0xE0,
		0xF0, 0x80, 0x80, 0x80, 0xF0,
		0xE0, 0x90, 0x90, 0x90, 0xE0,
		0xF0, 0x80, 0xF0, 0x80, 0xF0,
		0xF0, 0x80, 0xF0, 0x80, 0x80,
	}
)

// VirtualMachine emulates the CHIP-8 virtual machine.
type VirtualMachine struct {
	started int32
	stopped int32

	// memory contains all of the different memory locations of the virtual
	// machine, which are a byte long. There are 4096 (0x1000) of them.
	//
	// NOTE: The CHIP-8 interpreter occupies the first 512 bytes of the
	// memory space. For this reason, most programs begin at memory location
	// 512 (0x0200) and do not access memory below that.
	memory [memorySize]byte

	// v represents the registers of the virtual machine's CPU. There are
	// 16 registers, named V0 to VF.
	//
	// NOTE: The last register can also function as a carry flag for some
	// instructions, so it should avoid being used.
	v [numRegisters]byte

	// i is the address register of the virtual machine. It is used with
	// several opcodes that involve memory operations.
	i uint16

	// pc is the program counter of the virtual machine.
	pc uint16

	// stack is the stack of the virtual machine.
	stack [numFrames]uint16

	// sp is the stack pointer of virtual machine.
	sp byte

	// display is the display of the virtual machine.
	display Display

	// renderer renders the display of the virtual machine.
	renderer Renderer

	// delayTimer is the virtual machine's delay timer used for timing
	// events.
	delayTimer byte

	// soundTimer is the virtual machine's sound timer used for sound
	// effects. When its value is non-zero, a beeping sound is made.
	soundTimer byte

	// keys holds the current state for all supported keys. If the key is
	// pressed, then the state is true. Otherwise, it is false.
	keys [numKeys]bool

	// clock is a ticker that represents the clock of the virtual machine.
	// The default clock speed is 60 Hz.
	clock <-chan time.Time

	quit chan struct{}
	wg   sync.WaitGroup
}

// New creates a new CHIP-8 virtual machine.
func New(r Renderer) *VirtualMachine {
	vm := &VirtualMachine{
		pc:    memoryOffset,
		clock: time.Tick(time.Second / defaultClockSpeed),
		quit:  make(chan struct{}),
	}

	for i := 0; i < len(font); i++ {
		vm.memory[i] = font[i]
	}

	vm.renderer = r

	return vm
}

// LoadROM loads the ROM's data, in bytes, into the virtual machine's memory.
func (vm *VirtualMachine) LoadROM(rom []byte) error {
	if len(rom) > memorySize-memoryOffset {
		return fmt.Errorf("chip8: size of rom data is too large, "+
			"must be at most %d", memorySize-memoryOffset)
	}

	r := bytes.NewReader(rom)
	_, err := r.Read(vm.memory[memoryOffset:])
	if err != nil {
		return fmt.Errorf("chip8: unable to read data from rom: %v", err)
	}

	return nil
}

// Start starts executing the virtual machine.
func (vm *VirtualMachine) Start() {
	if atomic.AddInt32(&vm.started, 1) != 1 {
		return
	}

	vm.wg.Add(1)
	go vm.run()
}

// Stop stops executing the virtual machine.
func (vm *VirtualMachine) Stop() {
	if atomic.AddInt32(&vm.stopped, 1) != 1 {
		return
	}

	close(vm.quit)

	vm.wg.Wait()
}

// Reset resets the virtual machine to its initial state.
func (vm *VirtualMachine) Reset() {
	vm.Stop()

	vm.started = 0
	vm.stopped = 0

	for i := 0; i < memorySize; i++ {
		vm.memory[i] = 0
	}

	for i := 0; i < len(font); i++ {
		vm.memory[i] = font[i]
	}

	for i := 0; i < numRegisters; i++ {
		vm.v[i] = 0
	}

	vm.i = 0
	vm.pc = memoryOffset
	vm.sp = 0

	vm.display.clear()
	vm.renderer.Render(vm.display)

	vm.delayTimer = 0
	vm.soundTimer = 0

	for i := 0; i < numKeys; i++ {
		vm.keys[i] = false
	}

	vm.quit = make(chan struct{})
}

// run executes the opcode at every step of the virtual machine's execution.
//
// NOTE: This MUST be run in a goroutine.
func (vm *VirtualMachine) run() {
	defer vm.wg.Done()

out:
	for {
		select {
		case <-vm.clock:
			if err := vm.step(); err != nil {
				err = fmt.Errorf("chip8: %v", err)
				panic(err)
			}
		case <-vm.quit:
			break out
		}
	}
}

// step steps through the next opcode.
func (vm *VirtualMachine) step() error {
	op, err := vm.decodeNextOpcode()
	if err != nil {
		return fmt.Errorf("failed retrieving next opcode: %v", err)
	}

	if err := vm.execute(op); err != nil {
		return fmt.Errorf("failed executing opcode %v: %v", op, err)
	}

	if vm.delayTimer > 0 {
		vm.delayTimer--
	}

	if vm.soundTimer > 0 {
		vm.renderer.Beep()
		vm.soundTimer--
	}

	return nil
}

// decodeNextOpcode decodes the next opcode available.
func (vm *VirtualMachine) decodeNextOpcode() (opcode, error) {
	if vm.pc+1 > memorySize {
		return 0, errors.New("program counter out of bounds")
	}

	op := uint16(vm.memory[vm.pc])<<8 | uint16(vm.memory[vm.pc+1])

	if vm.pc+2 > memorySize {
		return 0, errors.New("program counter reached end of memory")
	}
	vm.pc += 2

	return opcode(op), nil
}

// execute executes the opcode on the virtual machine.
func (vm *VirtualMachine) execute(op opcode) error {
	switch op & 0xf000 {
	case 0x0000:
		switch op {
		case 0x00e0:
			// Clear the screen.
			vm.display.clear()
			vm.renderer.Render(vm.display)
		case 0x00ee:
			// Return from a subroutine.
			if vm.sp == 0 {
				return errors.New("stack underflow")
			}

			vm.sp--
			vm.pc = vm.stack[vm.sp]
		default:
			return ErrUnknownOpcode
		}
	case 0x1000:
		// Jump to the address encoded in the opcode.
		vm.pc = op.Address()
	case 0x2000:
		// Call subroutine at the address encoded in the opcode.
		addr := op.Address()

		if vm.sp >= numFrames {
			return errors.New("stack overflow")
		}

		vm.stack[vm.sp] = vm.pc
		vm.sp++
		vm.pc = addr
	case 0x3000:
		// Skip the next instruction if VX is equal to the byte constant.
		x := op.RegisterIndex(true)
		val := op.ByteConstant()

		if vm.v[x] == val {
			vm.pc += 2
		}
	case 0x4000:
		// Skip the next instruction if VX is not equal to the byte
		// constant.
		x := op.RegisterIndex(true)
		val := op.ByteConstant()

		if vm.v[x] != val {
			vm.pc += 2
		}
	case 0x5000:
		switch op & 0xf00f {
		case 0x5000:
			// Skip the next instruction if VX is equal to VY.
			x := op.RegisterIndex(true)
			y := op.RegisterIndex(false)

			if vm.v[x] == vm.v[y] {
				vm.pc += 2
			}
		default:
			return ErrUnknownOpcode
		}
	case 0x6000:
		// Set VX to the byte constant.
		x := op.RegisterIndex(true)
		val := op.ByteConstant()

		vm.v[x] = val
	case 0x7000:
		// Add the byte constant to VX.
		x := op.RegisterIndex(true)
		val := op.ByteConstant()

		vm.v[x] += val
	case 0x8000:
		switch op & 0x000f {
		case 0x0000:
			// Set VX to VY.
			x := op.RegisterIndex(true)
			y := op.RegisterIndex(false)

			vm.v[x] = vm.v[y]
		case 0x0001:
			// Set VX to the bitwise or of VX and VY.
			x := op.RegisterIndex(true)
			y := op.RegisterIndex(false)

			vm.v[x] |= vm.v[y]
		case 0x0002:
			// Set VX to the bitwise and of VX and VY.
			x := op.RegisterIndex(true)
			y := op.RegisterIndex(false)

			vm.v[x] &= vm.v[y]
		case 0x0003:
			// Set VX to the xor of VX and VY.
			x := op.RegisterIndex(true)
			y := op.RegisterIndex(false)

			vm.v[x] ^= vm.v[y]
		case 0x0004:
			// Add VY to VX. Set VF to 1 if there is a carry,
			// otherwise set it to 0.
			x := op.RegisterIndex(true)
			y := op.RegisterIndex(false)
			res := uint16(vm.v[x]) + uint16(vm.v[y])

			vm.v[x] = uint8(res)
			if res >= 0x100 {
				vm.v[0xf] = 1
			} else {
				vm.v[0xf] = 0
			}
		case 0x0005:
			// Subtract VY from VX. Set VF to 0 if there is a
			// borrow, otherwise set it to 1.
			x := op.RegisterIndex(true)
			y := op.RegisterIndex(false)
			res := int16(vm.v[x]) - int16(vm.v[y])

			vm.v[x] = uint8(res)
			if res >= 0 {
				vm.v[0xf] = 0
			} else {
				vm.v[0xf] = 1
			}
		case 0x0006:
			// Shift VY right by one and copy it to VX. Set VF to
			// the least significant bit of VY before the shft.
			x := op.RegisterIndex(true)
			y := op.RegisterIndex(false)

			lsb := vm.v[y] & 1
			vm.v[0xf] = lsb
			vm.v[y] >>= 1
			vm.v[x] = vm.v[y]
		case 0x0007:
			// Set VX to VX subtracted from VY. Set VF to 0 if there
			// is a borrow, otherwise set it to 1.
			x := op.RegisterIndex(true)
			y := op.RegisterIndex(false)
			res := int16(vm.v[y]) - int16(vm.v[x])

			vm.v[x] = uint8(res)
			if res >= 0 {
				vm.v[0xf] = 0
			} else {
				vm.v[0xf] = 1
			}
		case 0x000e:
			// Shift VY left by one and copy it to VX. Set VF to the
			// most significant bit of VY before the shft.
			x := op.RegisterIndex(true)
			y := op.RegisterIndex(false)

			msb := vm.v[y] >> 7
			vm.v[0xf] = msb
			vm.v[y] <<= 1
			vm.v[x] = vm.v[y]
		default:
			return ErrUnknownOpcode
		}
	case 0x9000:
		switch op & 0xf00f {
		case 0x9000:
			// Skip the next instruction if VX is not equal to VY.
			x := op.RegisterIndex(true)
			y := op.RegisterIndex(false)

			if vm.v[x] != vm.v[y] {
				vm.pc += 2
			}
		default:
			return ErrUnknownOpcode
		}
	case 0xa000:
		// Set the address register to the address encoded in the opcode.
		vm.i = op.Address()
	case 0xb000:
		// Jump to the address encoded in the opcode plus the value
		// stored in V0.
		vm.pc = op.Address() + uint16(vm.v[0])
	case 0xc000:
		// Set VX to a bitwise and operation between a random number and
		// the byte constant encoded in the opcode.
		x := op.RegisterIndex(true)
		r := rand.New(rand.NewSource(time.Now().UnixNano())).Intn(255)
		val := op.ByteConstant()

		vm.v[x] = byte(r) & val
	case 0xd000:
		// Draw a sprite at coordinate (VX, VY) that has a width of 8
		// and a height of the nibble constant encoded in the opcode in
		// pixels. Each row of 8 pixels is read as bit-coded starting
		// from the memory location stored in the address register. The
		// address register should not change after. Set VF to 1 if any
		// screen pixels are flipped from set to unset when the sprite
		// is drawn, otherwise set it to 0.
		x := op.RegisterIndex(true)
		y := op.RegisterIndex(false)
		height := op.NibbleConstant()

		flipped := vm.display.drawSprite(
			vm.memory[vm.i:vm.i+uint16(height)], vm.v[x], vm.v[y],
		)

		if flipped {
			vm.v[0xf] = 1
		} else {
			vm.v[0xf] = 0
		}

		vm.renderer.Render(vm.display)
	case 0xe000:
		switch op & 0x00ff {
		case 0x009e:
			// Skip the next instruction if the key stored in VX is
			// pressed.
			x := op.RegisterIndex(true)

			if vm.keys[vm.v[x]] {
				vm.pc += 2
			}
		case 0x00a1:
			// Skip the next instruction if the key stored in VY is
			// not pressed.
			x := op.RegisterIndex(true)

			if !vm.keys[vm.v[x]] {
				vm.pc += 2
			}
		default:
			return ErrUnknownOpcode
		}
	case 0xf000:
		switch op & 0x00ff {
		case 0x0007:
			// Set VX to the delay timer.
			x := op.RegisterIndex(true)

			vm.v[x] = vm.delayTimer
		case 0x000a:
			// TODO: Wait for a key press, and then store it in VX.
			panic("unimplemented opcode")
		case 0x0015:
			// Set the delay timer to VX.
			x := op.RegisterIndex(true)

			vm.delayTimer = vm.v[x]
		case 0x0018:
			// Set the sound timer to VX.
			x := op.RegisterIndex(true)

			vm.soundTimer = vm.v[x]
		case 0x001e:
			// Add VX to the address register.
			x := op.RegisterIndex(true)

			vm.i += uint16(vm.v[x])
		case 0x0029:
			// Set the address register to the location of the
			// sprite for the character in VX. Characters 0-F (in
			// hexadecimal) are represented by a 4x5 font.
			x := op.RegisterIndex(true)

			vm.i = uint16(vm.v[x]) * 5
		case 0x0033:
			// Store the binary-coded decimal representation of VX,
			// with the most significant of three digits at the
			// address in I, the middle digit at I+1, and the least
			// significant digit as I+2.
			x := op.RegisterIndex(true)

			bcd := vm.v[x]
			vm.memory[vm.i] = bcd / 100
			vm.memory[vm.i+1] = (bcd / 10) % 10
			vm.memory[vm.i+2] = bcd % 10
		case 0x0055:
			// Store the values from registers V0-VX in memory
			// starting at address I. I is increased by 1 for each
			// value written.
			x := op.RegisterIndex(true)

			for i := uint16(0); i < x; i++ {
				vm.memory[vm.i] = vm.v[i]
				vm.i++
			}
		case 0x0065:
			// Fill the registers V0-VX with values from memory
			// starting at address I. I is increased by 1 for each
			// value written.
			x := op.RegisterIndex(true)

			for i := uint16(0); i < x; i++ {
				vm.v[i] = vm.memory[vm.i]
				vm.i++
			}
		default:
			return ErrUnknownOpcode
		}
	default:
		return ErrUnknownOpcode
	}

	return nil
}

// PressKey signals the virtual machine that the key was pressed.
func (vm *VirtualMachine) PressKey(idx int) {
	vm.updateKeyState(idx, true)
}

// ReleaseKey signals the virtual machine that the key was released.
func (vm *VirtualMachine) ReleaseKey(idx int) {
	vm.updateKeyState(idx, false)
}

// updateKeyState updates the state of a key.
func (vm *VirtualMachine) updateKeyState(idx int, pressed bool) {
	if idx < 0 || idx >= numKeys {
		return
	}

	vm.keys[idx] = pressed
}
