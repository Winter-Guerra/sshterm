//go:build x11

package x11

import (
	"encoding/binary"
)

type keyEvent struct {
	opcode         byte // KeyPress: 2, KeyRelease: 3
	sequence       uint16
	detail         byte // keycode
	time           uint32
	root           uint32
	event          uint32
	child          uint32
	rootX, rootY   int16
	eventX, eventY int16
	state          uint16 // keyboard state
	sameScreen     bool
}

func (e *keyEvent) encodeMessage(order binary.ByteOrder) []byte {
	event := make([]byte, 32)
	event[0] = e.opcode
	event[1] = e.detail
	order.PutUint16(event[2:4], e.sequence)
	order.PutUint32(event[4:8], e.time)
	order.PutUint32(event[8:12], e.root)
	order.PutUint32(event[12:16], e.event)
	order.PutUint32(event[16:20], e.child)
	order.PutUint16(event[20:22], uint16(e.rootX))
	order.PutUint16(event[22:24], uint16(e.rootY))
	order.PutUint16(event[24:26], uint16(e.eventX))
	order.PutUint16(event[26:28], uint16(e.eventY))
	order.PutUint16(event[28:30], e.state)
	event[30] = boolToByte(e.sameScreen)
	// event[31] is unused
	return event
}

// ButtonPress: 4
type ButtonPressEvent struct {
	sequence       uint16
	detail         byte // button
	time           uint32
	root           uint32
	event          uint32
	child          uint32
	rootX, rootY   int16
	eventX, eventY int16
	state          uint16
	sameScreen     bool
}

type DeviceButtonReleaseEvent struct {
	sequence   uint16
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

func (e *DeviceButtonReleaseEvent) encodeMessage(order binary.ByteOrder) []byte {
	buf := make([]byte, 32)
	buf[0] = xInputOpcode
	buf[1] = DeviceButtonRelease
	order.PutUint16(buf[2:4], e.sequence)
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

func (e *ButtonPressEvent) encodeMessage(order binary.ByteOrder) []byte {
	event := make([]byte, 32)
	event[0] = 4 // ButtonPress event code
	event[1] = e.detail
	order.PutUint16(event[2:4], e.sequence)
	order.PutUint32(event[4:8], e.time)
	order.PutUint32(event[8:12], e.root)
	order.PutUint32(event[12:16], e.event)
	order.PutUint32(event[16:20], e.child)
	order.PutUint16(event[20:22], uint16(e.rootX))
	order.PutUint16(event[22:24], uint16(e.rootY))
	order.PutUint16(event[24:26], uint16(e.eventX))
	order.PutUint16(event[26:28], uint16(e.eventY))
	order.PutUint16(event[28:30], e.state)
	event[30] = boolToByte(e.sameScreen)
	// event[31] is unused
	return event
}

// ButtonRelease: 5
type ButtonReleaseEvent struct {
	sequence       uint16
	detail         byte // button
	time           uint32
	root           uint32
	event          uint32
	child          uint32
	rootX, rootY   int16
	eventX, eventY int16
	state          uint16
	sameScreen     bool
}

func (e *ButtonReleaseEvent) encodeMessage(order binary.ByteOrder) []byte {
	event := make([]byte, 32)
	event[0] = 5 // ButtonRelease event code
	event[1] = e.detail
	order.PutUint16(event[2:4], e.sequence)
	order.PutUint32(event[4:8], e.time)
	order.PutUint32(event[8:12], e.root)
	order.PutUint32(event[12:16], e.event)
	order.PutUint32(event[16:20], e.child)
	order.PutUint16(event[20:22], uint16(e.rootX))
	order.PutUint16(event[22:24], uint16(e.rootY))
	order.PutUint16(event[24:26], uint16(e.eventX))
	order.PutUint16(event[26:28], uint16(e.eventY))
	order.PutUint16(event[28:30], e.state)
	event[30] = boolToByte(e.sameScreen)
	// event[31] is unused
	return event
}

// MotionNotify: 6
type motionNotifyEvent struct {
	sequence       uint16
	detail         byte
	time           uint32
	root           uint32
	event          uint32
	child          uint32
	rootX, rootY   int16
	eventX, eventY int16
	state          uint16
	sameScreen     bool
}

func (e *motionNotifyEvent) encodeMessage(order binary.ByteOrder) []byte {
	event := make([]byte, 32)
	event[0] = 6 // MotionNotify event code
	event[1] = e.detail
	order.PutUint16(event[2:4], e.sequence)
	order.PutUint32(event[4:8], e.time)
	order.PutUint32(event[8:12], e.root)
	order.PutUint32(event[12:16], e.event)
	order.PutUint32(event[16:20], e.child)
	order.PutUint16(event[20:22], uint16(e.rootX))
	order.PutUint16(event[22:24], uint16(e.rootY))
	order.PutUint16(event[24:26], uint16(e.eventX))
	order.PutUint16(event[26:28], uint16(e.eventY))
	order.PutUint16(event[28:30], e.state)
	event[30] = boolToByte(e.sameScreen)
	// event[31] is unused
	return event
}

// EnterNotify: 7
type EnterNotifyEvent struct {
	sequence       uint16
	detail         byte
	time           uint32
	root           uint32
	event          uint32
	child          uint32
	rootX, rootY   int16
	eventX, eventY int16
	state          uint16
	mode           byte
	sameScreen     bool
	focus          bool
}

func (e *EnterNotifyEvent) encodeMessage(order binary.ByteOrder) []byte {
	event := make([]byte, 32)
	event[0] = 7 // EnterNotify event code
	event[1] = e.detail
	order.PutUint16(event[2:4], e.sequence)
	order.PutUint32(event[4:8], e.time)
	order.PutUint32(event[8:12], e.root)
	order.PutUint32(event[12:16], e.event)
	order.PutUint32(event[16:20], e.child)
	order.PutUint16(event[20:22], uint16(e.rootX))
	order.PutUint16(event[22:24], uint16(e.rootY))
	order.PutUint16(event[24:26], uint16(e.eventX))
	order.PutUint16(event[26:28], uint16(e.eventY))
	order.PutUint16(event[28:30], e.state)
	event[30] = e.mode
	var sameScreenFocusByte byte
	if e.sameScreen {
		sameScreenFocusByte |= 1
	}
	if e.focus {
		sameScreenFocusByte |= 2
	}
	event[31] = sameScreenFocusByte
	return event
}

// LeaveNotify: 8
type LeaveNotifyEvent struct {
	sequence       uint16
	detail         byte
	time           uint32
	root           uint32
	event          uint32
	child          uint32
	rootX, rootY   int16
	eventX, eventY int16
	state          uint16
	mode           byte
	sameScreen     bool
	focus          bool
}

func (e *LeaveNotifyEvent) encodeMessage(order binary.ByteOrder) []byte {
	event := make([]byte, 32)
	event[0] = 8 // LeaveNotify event code
	event[1] = e.detail
	order.PutUint16(event[2:4], e.sequence)
	order.PutUint32(event[4:8], e.time)
	order.PutUint32(event[8:12], e.root)
	order.PutUint32(event[12:16], e.event)
	order.PutUint32(event[16:20], e.child)
	order.PutUint16(event[20:22], uint16(e.rootX))
	order.PutUint16(event[22:24], uint16(e.rootY))
	order.PutUint16(event[24:26], uint16(e.eventX))
	order.PutUint16(event[26:28], uint16(e.eventY))
	order.PutUint16(event[28:30], e.state)
	event[30] = e.mode
	var sameScreenFocusByte byte
	if e.sameScreen {
		sameScreenFocusByte |= 1
	}
	if e.focus {
		sameScreenFocusByte |= 2
	}
	event[31] = sameScreenFocusByte
	return event
}

// Expose: 12
type exposeEvent struct {
	sequence      uint16
	window        uint32
	x, y          uint16
	width, height uint16
	count         uint16
}

func (e *exposeEvent) encodeMessage(order binary.ByteOrder) []byte {
	event := make([]byte, 32)
	event[0] = 12 // Expose event code
	// byte 1 is unused
	order.PutUint16(event[2:4], e.sequence)
	order.PutUint32(event[4:8], e.window)
	order.PutUint16(event[8:10], e.x)
	order.PutUint16(event[10:12], e.y)
	order.PutUint16(event[12:14], e.width)
	order.PutUint16(event[14:16], e.height)
	order.PutUint16(event[16:18], e.count)
	// event[18:32] is unused
	return event
}

// ConfigureNotify: 22
type configureNotifyEvent struct {
	sequence         uint16
	event            uint32
	window           uint32
	aboveSibling     uint32
	x, y             int16
	width, height    uint16
	borderWidth      uint16
	overrideRedirect bool
}

func (e *configureNotifyEvent) encodeMessage(order binary.ByteOrder) []byte {
	event := make([]byte, 32)
	event[0] = 22 // ConfigureNotify event code
	// byte 1 is unused
	order.PutUint16(event[2:4], e.sequence)
	order.PutUint32(event[4:8], e.event)
	order.PutUint32(event[8:12], e.window)
	order.PutUint32(event[12:16], e.aboveSibling)
	order.PutUint16(event[16:18], uint16(e.x))
	order.PutUint16(event[18:20], uint16(e.y))
	order.PutUint16(event[20:22], e.width)
	order.PutUint16(event[22:24], e.height)
	order.PutUint16(event[24:26], e.borderWidth)
	event[26] = boolToByte(e.overrideRedirect)
	// byte 27 is unused
	return event
}

// SelectionNotify: 31
type selectionNotifyEvent struct {
	sequence  uint16
	requestor uint32
	selection uint32
	target    uint32
	property  uint32
	time      uint32
}

func (e *selectionNotifyEvent) encodeMessage(order binary.ByteOrder) []byte {
	event := make([]byte, 32)
	event[0] = 31 // SelectionNotify event code
	// byte 1 is unused
	order.PutUint16(event[2:4], e.sequence)
	order.PutUint32(event[4:8], e.requestor)
	order.PutUint32(event[8:12], e.selection)
	order.PutUint32(event[12:16], e.target)
	order.PutUint32(event[16:20], e.property)
	order.PutUint32(event[20:24], e.time)
	// event[24:32] is unused
	return event
}

// ColormapNotify: 32
type colormapNotifyEvent struct {
	sequence uint16
	window   uint32
	colormap uint32
	new      bool
	state    byte
}

func (e *colormapNotifyEvent) encodeMessage(order binary.ByteOrder) []byte {
	event := make([]byte, 32)
	event[0] = ColormapNotifyCode // ColormapNotify event code
	// byte 1 is unused
	order.PutUint16(event[2:4], e.sequence)
	order.PutUint32(event[4:8], e.window)
	order.PutUint32(event[8:12], e.colormap)
	event[12] = boolToByte(e.new)
	event[13] = e.state
	// event[14:32] is unused
	return event
}

// ClientMessage: 33
type clientMessageEvent struct {
	sequence    uint16
	format      byte
	window      uint32
	messageType uint32
	data        [20]byte
}

func (e *clientMessageEvent) encodeMessage(order binary.ByteOrder) []byte {
	event := make([]byte, 32)
	event[0] = 33 // ClientMessage event code
	event[1] = e.format
	order.PutUint16(event[2:4], e.sequence)
	order.PutUint32(event[4:8], e.window)
	order.PutUint32(event[8:12], e.messageType)
	copy(event[12:32], e.data[:])
	return event
}

// x11RawEvent implements messageEncoder for raw X11 event data.
type x11RawEvent struct {
	data []byte
}

func (e *x11RawEvent) encodeMessage(order binary.ByteOrder) []byte {
	return e.data
}

// DeviceKeyPressEvent is an XInput key press event.
type DeviceKeyPressEvent struct {
	DeviceID   byte
	sequence   uint16
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

func (e *DeviceKeyPressEvent) encodeMessage(order binary.ByteOrder) []byte {
	buf := make([]byte, 32)
	buf[0] = xInputOpcode
	buf[1] = DeviceKeyPress
	order.PutUint16(buf[2:4], e.sequence)
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
	sequence   uint16
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

func (e *DeviceKeyReleaseEvent) encodeMessage(order binary.ByteOrder) []byte {
	buf := make([]byte, 32)
	buf[0] = xInputOpcode
	buf[1] = DeviceKeyRelease
	order.PutUint16(buf[2:4], e.sequence)
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
	sequence   uint16
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
	Button     byte
}

func (e *DeviceButtonPressEvent) encodeMessage(order binary.ByteOrder) []byte {
	buf := make([]byte, 32)
	buf[0] = xInputOpcode
	buf[1] = DeviceButtonPress
	order.PutUint16(buf[2:4], e.sequence)
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
