package ansi

// An index into a standard 256-color ANSI palette
type ColorIndex uint8

const Black = ColorIndex(0)
const White = ColorIndex(15)

// RGB maps rgb color to an index in standard ANSI 256-color palette
//
// input color component values [0..ff] are snapped to
// - 0x00
// - 0x33
// - 0x66
// - 0x99
// - 0xcc
// - 0xff
//
func RGB(r, g, b byte) ColorIndex {
	rr := (uint(r) * 3) >> 7
	gg := (uint(g) * 3) >> 7
	bb := (uint(b) * 3) >> 7
	return ColorIndex(rr*36 + gg*6 + bb + 16)
}

// Gray maps luminance to an index in standard ANSI 256-color palette
//
// input values [0..ff] are snapped to
// - 0x00
// - 0x0A
// - 0x14
// - 0x1E
// - 0x28
// - 0x33
// - 0x3D
// - 0x47
// - 0x51
// - 0x5B
// - 0x66
// - 0x70
// - 0x7A
// - 0x84
// - 0x8E
// - 0x99
// - 0xA3
// - 0xAD
// - 0xB7
// - 0xC1
// - 0xCC
// - 0xD6
// - 0xE0
// - 0xEA
// - 0xF4
// - 0xFF
//
func Gray(l byte) ColorIndex {
	v := (uint32(l)*25 + 128) >> 8
	if v == 0 {
		return Black
	} else if v == 25 {
		return White
	}
	v += 231
	return ColorIndex(v)
}
