package chip8

const (
	// DisplayWidth is the width of the display in pixels.
	DisplayWidth = 64

	// DisplayHeight is the height of the display in pixels.
	DisplayHeight = 32
)

// Renderer represents the abstract renderer for the CHIP-8 virtual machine.
type Renderer interface {
	// Render renders the display.
	Render(display Display) error

	// Beep makes an audible beep.
	Beep() error
}

// Display represents the display of the CHIP-8 virtual machine.
type Display [DisplayHeight][DisplayWidth]byte

// drawSprite draws the sprite on the display. The bool returned signifies that
// a pixel was flipped from set to unset while the sprite was drawn.
func (d *Display) drawSprite(sprite []byte, x, y uint8) bool {
	flipped := false

	// First, we'll go through every byte of the sprite. Each byte
	// represents 8 pixels, one for each bit.
	n := uint8(len(sprite))
	for i := uint8(0); i < n; i++ {
		// Get the row of pixels.
		pixels := sprite[i]

		// Now, we'll go through every pixel in our row and draw it.
		for j := uint8(0); j < 8; j++ {
			// Get the coordinates of the pixel in our display.
			xPos := (x + j) % DisplayWidth
			yPos := (y + i) % DisplayHeight

			// Determine if this pixel in the sprite needs to be
			// drawn.
			pixel := (pixels >> (7 - j)) & 1
			set := pixel == 1
			drawn := d[yPos][xPos] == 1

			if !set && drawn {
				flipped = true
			}

			d[yPos][xPos] ^= pixel
		}
	}

	return flipped
}

// clear clears the display.
func (d *Display) clear() {
	for y := 0; y < DisplayHeight; y++ {
		for x := 0; x < DisplayWidth; x++ {
			d[y][x] = 0
		}
	}
}
