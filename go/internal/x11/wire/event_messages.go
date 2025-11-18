//go:build x11

package wire

import (
	"encoding/binary"
)

type KeyEvent struct {
	Opcode         byte // KeyPress: 2, KeyRelease: 3
	Sequence       uint16
	Detail         byte // keycode
	Time           uint32
	Root           uint32
	Event          uint32
	Child          uint32
	RootX, RootY   int16
	EventX, EventY int16
	State          uint16 // keyboard state
	SameScreen     bool
}

func (e *KeyEvent) EncodeMessage(order binary.ByteOrder) []byte {
	event := make([]byte, 32)
	event[0] = e.Opcode
	event[1] = e.Detail
	order.PutUint16(event[2:4], e.Sequence)
	order.PutUint32(event[4:8], e.Time)
	order.PutUint32(event[8:12], e.Root)
	order.PutUint32(event[12:16], e.Event)
	order.PutUint32(event[16:20], e.Child)
	order.PutUint16(event[20:22], uint16(e.RootX))
	order.PutUint16(event[22:24], uint16(e.RootY))
	order.PutUint16(event[24:26], uint16(e.EventX))
	order.PutUint16(event[26:28], uint16(e.EventY))
	order.PutUint16(event[28:30], e.State)
	event[30] = BoolToByte(e.SameScreen)
	// event[31] is unused
	return event
}

// ButtonPress: 4
type ButtonPressEvent struct {
	Sequence       uint16
	Detail         byte // button
	Time           uint32
	Root           uint32
	Event          uint32
	Child          uint32
	RootX, RootY   int16
	EventX, EventY int16
	State          uint16
	SameScreen     bool
}

type DeviceButtonReleaseEvent struct {
	Sequence   uint16
	DeviceID   byte
	Time       uint32
	Button     byte
	Root       uint32
	Event      uint32
	Child      uint32
	RootX      int16
	RootY      int16
	EventX     int16
	EventY     int16
	State      uint16
	SameScreen bool
}

func (e *DeviceButtonReleaseEvent) EncodeMessage(order binary.ByteOrder) []byte {
	buf := make([]byte, 32)
	buf[0] = byte(XInputOpcode)
	buf[1] = DeviceButtonRelease
	order.PutUint16(buf[2:4], e.Sequence)
	order.PutUint32(buf[4:8], e.Time)
	order.PutUint32(buf[8:12], e.Root)
	order.PutUint32(buf[12:16], e.Event)
	order.PutUint32(buf[16:20], e.Child)
	order.PutUint16(buf[20:22], uint16(e.RootX))
	order.PutUint16(buf[22:24], uint16(e.RootY))
	order.PutUint16(buf[24:26], uint16(e.EventX))
	order.PutUint16(buf[26:28], uint16(e.EventY))
	order.PutUint16(buf[28:30], e.State)
	buf[30] = e.DeviceID
	buf[31] = e.Button
	return buf
}

func (e *ButtonPressEvent) EncodeMessage(order binary.ByteOrder) []byte {
	event := make([]byte, 32)
	event[0] = 4 // ButtonPress event code
	event[1] = e.Detail
	order.PutUint16(event[2:4], e.Sequence)
	order.PutUint32(event[4:8], e.Time)
	order.PutUint32(event[8:12], e.Root)
	order.PutUint32(event[12:16], e.Event)
	order.PutUint32(event[16:20], e.Child)
	order.PutUint16(event[20:22], uint16(e.RootX))
	order.PutUint16(event[22:24], uint16(e.RootY))
	order.PutUint16(event[24:26], uint16(e.EventX))
	order.PutUint16(event[26:28], uint16(e.EventY))
	order.PutUint16(event[28:30], e.State)
	event[30] = BoolToByte(e.SameScreen)
	// event[31] is unused
	return event
}

// ButtonRelease: 5
type ButtonReleaseEvent struct {
	Sequence       uint16
	Detail         byte // button
	Time           uint32
	Root           uint32
	Event          uint32
	Child          uint32
	RootX, RootY   int16
	EventX, EventY int16
	State          uint16
	SameScreen     bool
}

func (e *ButtonReleaseEvent) EncodeMessage(order binary.ByteOrder) []byte {
	event := make([]byte, 32)
	event[0] = 5 // ButtonRelease event code
	event[1] = e.Detail
	order.PutUint16(event[2:4], e.Sequence)
	order.PutUint32(event[4:8], e.Time)
	order.PutUint32(event[8:12], e.Root)
	order.PutUint32(event[12:16], e.Event)
	order.PutUint32(event[16:20], e.Child)
	order.PutUint16(event[20:22], uint16(e.RootX))
	order.PutUint16(event[22:24], uint16(e.RootY))
	order.PutUint16(event[24:26], uint16(e.EventX))
	order.PutUint16(event[26:28], uint16(e.EventY))
	order.PutUint16(event[28:30], e.State)
	event[30] = BoolToByte(e.SameScreen)
	// event[31] is unused
	return event
}

// MotionNotify: 6
type MotionNotifyEvent struct {
	Sequence       uint16
	Detail         byte
	Time           uint32
	Root           uint32
	Event          uint32
	Child          uint32
	RootX, RootY   int16
	EventX, EventY int16
	State          uint16
	SameScreen     bool
}

func (e *MotionNotifyEvent) EncodeMessage(order binary.ByteOrder) []byte {
	event := make([]byte, 32)
	event[0] = 6 // MotionNotify event code
	event[1] = e.Detail
	order.PutUint16(event[2:4], e.Sequence)
	order.PutUint32(event[4:8], e.Time)
	order.PutUint32(event[8:12], e.Root)
	order.PutUint32(event[12:16], e.Event)
	order.PutUint32(event[16:20], e.Child)
	order.PutUint16(event[20:22], uint16(e.RootX))
	order.PutUint16(event[22:24], uint16(e.RootY))
	order.PutUint16(event[24:26], uint16(e.EventX))
	order.PutUint16(event[26:28], uint16(e.EventY))
	order.PutUint16(event[28:30], e.State)
	event[30] = BoolToByte(e.SameScreen)
	// event[31] is unused
	return event
}

// EnterNotify: 7
type EnterNotifyEvent struct {
	Sequence       uint16
	Detail         byte
	Time           uint32
	Root           uint32
	Event          uint32
	Child          uint32
	RootX, RootY   int16
	EventX, EventY int16
	State          uint16
	Mode           byte
	SameScreen     bool
	Focus          bool
}

func (e *EnterNotifyEvent) EncodeMessage(order binary.ByteOrder) []byte {
	event := make([]byte, 32)
	event[0] = 7 // EnterNotify event code
	event[1] = e.Detail
	order.PutUint16(event[2:4], e.Sequence)
	order.PutUint32(event[4:8], e.Time)
	order.PutUint32(event[8:12], e.Root)
	order.PutUint32(event[12:16], e.Event)
	order.PutUint32(event[16:20], e.Child)
	order.PutUint16(event[20:22], uint16(e.RootX))
	order.PutUint16(event[22:24], uint16(e.RootY))
	order.PutUint16(event[24:26], uint16(e.EventX))
	order.PutUint16(event[26:28], uint16(e.EventY))
	order.PutUint16(event[28:30], e.State)
	event[30] = e.Mode
	var sameScreenFocusByte byte
	if e.SameScreen {
		sameScreenFocusByte |= 1
	}
	if e.Focus {
		sameScreenFocusByte |= 2
	}
	event[31] = sameScreenFocusByte
	return event
}

// LeaveNotify: 8
type LeaveNotifyEvent struct {
	Sequence       uint16
	Detail         byte
	Time           uint32
	Root           uint32
	Event          uint32
	Child          uint32
	RootX, RootY   int16
	EventX, EventY int16
	State          uint16
	Mode           byte
	SameScreen     bool
	Focus          bool
}

func (e *LeaveNotifyEvent) EncodeMessage(order binary.ByteOrder) []byte {
	event := make([]byte, 32)
	event[0] = 8 // LeaveNotify event code
	event[1] = e.Detail
	order.PutUint16(event[2:4], e.Sequence)
	order.PutUint32(event[4:8], e.Time)
	order.PutUint32(event[8:12], e.Root)
	order.PutUint32(event[12:16], e.Event)
	order.PutUint32(event[16:20], e.Child)
	order.PutUint16(event[20:22], uint16(e.RootX))
	order.PutUint16(event[22:24], uint16(e.RootY))
	order.PutUint16(event[24:26], uint16(e.EventX))
	order.PutUint16(event[26:28], uint16(e.EventY))
	order.PutUint16(event[28:30], e.State)
	event[30] = e.Mode
	var sameScreenFocusByte byte
	if e.SameScreen {
		sameScreenFocusByte |= 1
	}
	if e.Focus {
		sameScreenFocusByte |= 2
	}
	event[31] = sameScreenFocusByte
	return event
}

// Expose: 12
type ExposeEvent struct {
	Sequence      uint16
	Window        uint32
	X, Y          uint16
	Width, Height uint16
	Count         uint16
}

func (e *ExposeEvent) EncodeMessage(order binary.ByteOrder) []byte {
	event := make([]byte, 32)
	event[0] = 12 // Expose event code
	// byte 1 is unused
	order.PutUint16(event[2:4], e.Sequence)
	order.PutUint32(event[4:8], e.Window)
	order.PutUint16(event[8:10], e.X)
	order.PutUint16(event[10:12], e.Y)
	order.PutUint16(event[12:14], e.Width)
	order.PutUint16(event[14:16], e.Height)
	order.PutUint16(event[16:18], e.Count)
	// event[18:32] is unused
	return event
}

// ConfigureNotify: 22
type ConfigureNotifyEvent struct {
	Sequence         uint16
	Event            uint32
	Window           uint32
	AboveSibling     uint32
	X, Y             int16
	Width, Height    uint16
	BorderWidth      uint16
	OverrideRedirect bool
}

func (e *ConfigureNotifyEvent) EncodeMessage(order binary.ByteOrder) []byte {
	event := make([]byte, 32)
	event[0] = 22 // ConfigureNotify event code
	// byte 1 is unused
	order.PutUint16(event[2:4], e.Sequence)
	order.PutUint32(event[4:8], e.Event)
	order.PutUint32(event[8:12], e.Window)
	order.PutUint32(event[12:16], e.AboveSibling)
	order.PutUint16(event[16:18], uint16(e.X))
	order.PutUint16(event[18:20], uint16(e.Y))
	order.PutUint16(event[20:22], e.Width)
	order.PutUint16(event[22:24], e.Height)
	order.PutUint16(event[24:26], e.BorderWidth)
	event[26] = BoolToByte(e.OverrideRedirect)
	// byte 27 is unused
	return event
}

// SelectionNotify: 31
type SelectionNotifyEvent struct {
	Sequence  uint16
	Requestor uint32
	Selection uint32
	Target    uint32
	Property  uint32
	Time      uint32
}

func (e *SelectionNotifyEvent) EncodeMessage(order binary.ByteOrder) []byte {
	event := make([]byte, 32)
	event[0] = 31 // SelectionNotify event code
	// byte 1 is unused
	order.PutUint16(event[2:4], e.Sequence)
	order.PutUint32(event[4:8], e.Requestor)
	order.PutUint32(event[8:12], e.Selection)
	order.PutUint32(event[12:16], e.Target)
	order.PutUint32(event[16:20], e.Property)
	order.PutUint32(event[20:24], e.Time)
	// event[24:32] is unused
	return event
}

// ColormapNotify: 32
type ColormapNotifyEvent struct {
	Sequence uint16
	Window   uint32
	Colormap uint32
	New      bool
	State    byte
}

func (e *ColormapNotifyEvent) EncodeMessage(order binary.ByteOrder) []byte {
	event := make([]byte, 32)
	event[0] = ColormapNotifyCode // ColormapNotify event code
	// byte 1 is unused
	order.PutUint16(event[2:4], e.Sequence)
	order.PutUint32(event[4:8], e.Window)
	order.PutUint32(event[8:12], e.Colormap)
	event[12] = BoolToByte(e.New)
	event[13] = e.State
	// event[14:32] is unused
	return event
}

// ClientMessage: 33
type ClientMessageEvent struct {
	Sequence    uint16
	Format      byte
	Window      uint32
	MessageType uint32
	Data        [20]byte
}

func (e *ClientMessageEvent) EncodeMessage(order binary.ByteOrder) []byte {
	event := make([]byte, 32)
	event[0] = 33 // ClientMessage event code
	event[1] = e.Format
	order.PutUint16(event[2:4], e.Sequence)
	order.PutUint32(event[4:8], e.Window)
	order.PutUint32(event[8:12], e.MessageType)
	copy(event[12:32], e.Data[:])
	return event
}

// X11RawEvent implements messageEncoder for raw X11 event data.
type X11RawEvent struct {
	Data []byte
}

func (e *X11RawEvent) EncodeMessage(order binary.ByteOrder) []byte {
	return e.Data
}

// DeviceKeyPressEvent is an XInput key press event.
type DeviceKeyPressEvent struct {
	DeviceID   byte
	Sequence   uint16
	Time       uint32
	Root       uint32
	Event      uint32
	Child      uint32
	RootX      int16
	RootY      int16
	EventX     int16
	EventY     int16
	State      uint16
	SameScreen bool
	KeyCode    byte
}

func (e *DeviceKeyPressEvent) EncodeMessage(order binary.ByteOrder) []byte {
	buf := make([]byte, 32)
	buf[0] = byte(XInputOpcode)
	buf[1] = DeviceKeyPress
	order.PutUint16(buf[2:4], e.Sequence)
	order.PutUint32(buf[4:8], e.Time)
	order.PutUint32(buf[8:12], e.Root)
	order.PutUint32(buf[12:16], e.Event)
	order.PutUint32(buf[16:20], e.Child)
	order.PutUint16(buf[20:22], uint16(e.RootX))
	order.PutUint16(buf[22:24], uint16(e.RootY))
	order.PutUint16(buf[24:26], uint16(e.EventX))
	order.PutUint16(buf[26:28], uint16(e.EventY))
	order.PutUint16(buf[28:30], e.State)
	buf[30] = e.DeviceID
	buf[31] = e.KeyCode
	return buf
}

// DeviceKeyReleaseEvent is an XInput key release event.
type DeviceKeyReleaseEvent struct {
	DeviceID   byte
	Sequence   uint16
	Time       uint32
	Root       uint32
	Event      uint32
	Child      uint32
	RootX      int16
	RootY      int16
	EventX     int16
	EventY     int16
	State      uint16
	SameScreen bool
	KeyCode    byte
}

func (e *DeviceKeyReleaseEvent) EncodeMessage(order binary.ByteOrder) []byte {
	buf := make([]byte, 32)
	buf[0] = byte(XInputOpcode)
	buf[1] = DeviceKeyRelease
	order.PutUint16(buf[2:4], e.Sequence)
	order.PutUint32(buf[4:8], e.Time)
	order.PutUint32(buf[8:12], e.Root)
	order.PutUint32(buf[12:16], e.Event)
	order.PutUint32(buf[16:20], e.Child)
	order.PutUint16(buf[20:22], uint16(e.RootX))
	order.PutUint16(buf[22:24], uint16(e.RootY))
	order.PutUint16(buf[24:26], uint16(e.EventX))
	order.PutUint16(buf[26:28], uint16(e.EventY))
	order.PutUint16(buf[28:30], e.State)
	buf[30] = e.DeviceID
	buf[31] = e.KeyCode
	return buf
}

// DeviceButtonPressEvent is an XInput button press event.
type DeviceButtonPressEvent struct {
	DeviceID   byte
	Sequence   uint16
	Time       uint32
	Root       uint32
	Event      uint32
	Child      uint32
	RootX      int16
	RootY      int16
	EventX     int16
	EventY     int16
	State      uint16
	SameScreen bool
	Detail     byte // Button
}

func (e *DeviceButtonPressEvent) SetSequence(seq uint16) {
	e.Sequence = seq
}

func (e *DeviceButtonPressEvent) EncodeMessage(order binary.ByteOrder) []byte {
	buf := make([]byte, 32)
	buf[0] = byte(XInputOpcode)
	buf[1] = DeviceButtonPress
	order.PutUint16(buf[2:4], e.Sequence)
	order.PutUint32(buf[4:8], e.Time)
	order.PutUint32(buf[8:12], e.Root)
	order.PutUint32(buf[12:16], e.Event)
	order.PutUint32(buf[16:20], e.Child)
	order.PutUint16(buf[20:22], uint16(e.RootX))
	order.PutUint16(buf[22:24], uint16(e.RootY))
	order.PutUint16(buf[24:26], uint16(e.EventX))
	order.PutUint16(buf[26:28], uint16(e.EventY))
	order.PutUint16(buf[28:30], e.State)
	buf[30] = e.DeviceID
	buf[31] = e.Detail
	return buf
}
