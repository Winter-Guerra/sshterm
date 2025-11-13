//go:build x11

package x11

// A simplified map of X11 keycodes to keysyms.
// This is not a complete mapping and is intended for basic functionality.
var keycodeToKeysym = map[byte]uint32{
	10: 0x0031, // '1'
	11: 0x0032, // '2'
	12: 0x0033, // '3'
	24: 0x0071, // 'q'
	25: 0x0077, // 'w'
	26: 0x0065, // 'e'
	38: 0x0061, // 'a'
	39: 0x0073, // 's'
	40: 0x0064, // 'd'
}
