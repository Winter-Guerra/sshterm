//go:build x11

package wire

// Window is a 32-bit value representing a window.
type Window uint32

// Drawable is a 32-bit value representing a drawable (window or pixmap).
type Drawable uint32

// Font is a 32-bit value representing a font.
type Font uint32

// Pixmap is a 32-bit value representing a pixmap.
type Pixmap uint32

// Cursor is a 32-bit value representing a cursor.
type Cursor uint32

// Colormap is a 32-bit value representing a colormap.
type Colormap uint32

// GContext is a 32-bit value representing a graphics context.
type GContext uint32

// Atom is a 32-bit value representing an atom.
type Atom uint32

// VisualID is a 32-bit value representing a visual ID.
type VisualID uint32

// Timestamp is a 32-bit value representing a timestamp.
type Timestamp uint32

// KeyCode is an 8-bit value representing a key code.
type KeyCode uint8

// Rectangle defines a rectangle.
type Rectangle struct {
	X      int16
	Y      int16
	Width  uint16
	Height uint16
}

// KeyboardControl defines the keyboard control attributes.
type KeyboardControl struct {
	KeyClickPercent int32
	BellPercent     int32
	BellPitch       int32
	BellDuration    int32
	Led             uint32
	LedMode         uint32
	Key             KeyCode
	AutoRepeatMode  uint32
}

// Host defines a host address.
type Host struct {
	Family byte
	Data   []byte
}

// XColorItem defines a color item.
type XColorItem struct {
	Pixel    uint32
	Red      uint16
	Green    uint16
	Blue     uint16
	Flags    byte
	ClientID uint32
}

// XID is a generic X resource identifier.
type XID uint32
