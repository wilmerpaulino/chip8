package chip8

import (
	"errors"
	"fmt"
)

var (
	// errUnknownOpcode is returned when the virtual machine has tried to
	// execute an unknown opcode.
	errUnknownOpcode = errors.New("unknown opcode")
)

// opcode represents an opcode for the CHIP-8 virtual machine. It is composed
// of two bytes in big endian notation.
type opcode uint16

// Address returns the 12-bit address encoded in the opcode.
func (op opcode) Address() uint16 {
	return uint16(op) & 0x0fff
}

// ByteConstant returns the byte constant encoded in the opcode.
func (op opcode) ByteConstant() byte {
	return byte(op) & 0x00ff
}

// NibbleConstant returns the nibble constant encoded in the opcode.
func (op opcode) NibbleConstant() byte {
	return byte(op) & 0x000f
}

// RegisterIndex returns the register index encoded in the opcode.
func (op opcode) RegisterIndex(first bool) uint16 {
	if first {
		return (uint16(op) & 0x0f00) >> 8
	}

	return (uint16(op) & 0x00f0) >> 4
}

// String returns the string representation of an opcode.
func (op opcode) String() string {
	return fmt.Sprintf("0x%04X", uint16(op))
}
