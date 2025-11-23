//go:build x11

package wire

import (
	"encoding/binary"
	"fmt"
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

// GraphicsExposure: 13
type GraphicsExposureEvent struct {
	Sequence      uint16
	Drawable      uint32
	X, Y          uint16
	Width, Height uint16
	MinorOpcode   uint16
	Count         uint16
	MajorOpcode   byte
}

func (e *GraphicsExposureEvent) EncodeMessage(order binary.ByteOrder) []byte {
	event := make([]byte, 32)
	event[0] = 13 // GraphicsExposure event code
	// byte 1 is unused
	order.PutUint16(event[2:4], e.Sequence)
	order.PutUint32(event[4:8], e.Drawable)
	order.PutUint16(event[8:10], e.X)
	order.PutUint16(event[10:12], e.Y)
	order.PutUint16(event[12:14], e.Width)
	order.PutUint16(event[14:16], e.Height)
	order.PutUint16(event[16:18], e.MinorOpcode)
	order.PutUint16(event[18:20], e.Count)
	event[20] = e.MajorOpcode
	// event[21:32] is unused
	return event
}

// NoExposure: 14
type NoExposureEvent struct {
	Sequence    uint16
	Drawable    uint32
	MinorOpcode uint16
	MajorOpcode byte
}

func (e *NoExposureEvent) EncodeMessage(order binary.ByteOrder) []byte {
	event := make([]byte, 32)
	event[0] = 14 // NoExposure event code
	// byte 1 is unused
	order.PutUint16(event[2:4], e.Sequence)
	order.PutUint32(event[4:8], e.Drawable)
	order.PutUint16(event[8:10], e.MinorOpcode)
	event[10] = e.MajorOpcode
	// event[11:32] is unused
	return event
}

// VisibilityNotify: 15
type VisibilityNotifyEvent struct {
	Sequence uint16
	Window   uint32
	State    byte
}

func (e *VisibilityNotifyEvent) EncodeMessage(order binary.ByteOrder) []byte {
	event := make([]byte, 32)
	event[0] = 15 // VisibilityNotify event code
	// byte 1 is unused
	order.PutUint16(event[2:4], e.Sequence)
	order.PutUint32(event[4:8], e.Window)
	event[8] = e.State
	// event[9:32] is unused
	return event
}

// CreateNotify: 16
type CreateNotifyEvent struct {
	Sequence         uint16
	Parent           uint32
	Window           uint32
	X, Y             int16
	Width, Height    uint16
	BorderWidth      uint16
	OverrideRedirect bool
}

func (e *CreateNotifyEvent) EncodeMessage(order binary.ByteOrder) []byte {
	event := make([]byte, 32)
	event[0] = 16 // CreateNotify event code
	// byte 1 is unused
	order.PutUint16(event[2:4], e.Sequence)
	order.PutUint32(event[4:8], e.Parent)
	order.PutUint32(event[8:12], e.Window)
	order.PutUint16(event[12:14], uint16(e.X))
	order.PutUint16(event[14:16], uint16(e.Y))
	order.PutUint16(event[16:18], e.Width)
	order.PutUint16(event[18:20], e.Height)
	order.PutUint16(event[20:22], e.BorderWidth)
	event[22] = BoolToByte(e.OverrideRedirect)
	// byte 23 is unused
	return event
}

// DestroyNotify: 17
type DestroyNotifyEvent struct {
	Sequence uint16
	Event    uint32
	Window   uint32
}

func (e *DestroyNotifyEvent) EncodeMessage(order binary.ByteOrder) []byte {
	event := make([]byte, 32)
	event[0] = 17 // DestroyNotify event code
	// byte 1 is unused
	order.PutUint16(event[2:4], e.Sequence)
	order.PutUint32(event[4:8], e.Event)
	order.PutUint32(event[8:12], e.Window)
	// event[12:32] is unused
	return event
}

// UnmapNotify: 18
type UnmapNotifyEvent struct {
	Sequence      uint16
	Event         uint32
	Window        uint32
	FromConfigure bool
}

func (e *UnmapNotifyEvent) EncodeMessage(order binary.ByteOrder) []byte {
	event := make([]byte, 32)
	event[0] = 18 // UnmapNotify event code
	// byte 1 is unused
	order.PutUint16(event[2:4], e.Sequence)
	order.PutUint32(event[4:8], e.Event)
	order.PutUint32(event[8:12], e.Window)
	event[12] = BoolToByte(e.FromConfigure)
	// event[13:32] is unused
	return event
}

// MapNotify: 19
type MapNotifyEvent struct {
	Sequence         uint16
	Event            uint32
	Window           uint32
	OverrideRedirect bool
}

func (e *MapNotifyEvent) EncodeMessage(order binary.ByteOrder) []byte {
	event := make([]byte, 32)
	event[0] = 19 // MapNotify event code
	// byte 1 is unused
	order.PutUint16(event[2:4], e.Sequence)
	order.PutUint32(event[4:8], e.Event)
	order.PutUint32(event[8:12], e.Window)
	event[12] = BoolToByte(e.OverrideRedirect)
	// event[13:32] is unused
	return event
}

// MapRequest: 20
type MapRequestEvent struct {
	Sequence uint16
	Parent   uint32
	Window   uint32
}

func (e *MapRequestEvent) EncodeMessage(order binary.ByteOrder) []byte {
	event := make([]byte, 32)
	event[0] = 20 // MapRequest event code
	// byte 1 is unused
	order.PutUint16(event[2:4], e.Sequence)
	order.PutUint32(event[4:8], e.Parent)
	order.PutUint32(event[8:12], e.Window)
	// event[12:32] is unused
	return event
}

// ReparentNotify: 21
type ReparentNotifyEvent struct {
	Sequence         uint16
	Event            uint32
	Window           uint32
	Parent           uint32
	X, Y             int16
	OverrideRedirect bool
}

func (e *ReparentNotifyEvent) EncodeMessage(order binary.ByteOrder) []byte {
	event := make([]byte, 32)
	event[0] = 21 // ReparentNotify event code
	// byte 1 is unused
	order.PutUint16(event[2:4], e.Sequence)
	order.PutUint32(event[4:8], e.Event)
	order.PutUint32(event[8:12], e.Window)
	order.PutUint32(event[12:16], e.Parent)
	order.PutUint16(event[16:18], uint16(e.X))
	order.PutUint16(event[18:20], uint16(e.Y))
	event[20] = BoolToByte(e.OverrideRedirect)
	// event[21:32] is unused
	return event
}

// ConfigureRequest: 23
type ConfigureRequestEvent struct {
	Sequence      uint16
	StackMode     byte
	Parent        uint32
	Window        uint32
	Sibling       uint32
	X, Y          int16
	Width, Height uint16
	BorderWidth   uint16
	ValueMask     uint16
}

func (e *ConfigureRequestEvent) EncodeMessage(order binary.ByteOrder) []byte {
	event := make([]byte, 32)
	event[0] = 23 // ConfigureRequest event code
	event[1] = e.StackMode
	order.PutUint16(event[2:4], e.Sequence)
	order.PutUint32(event[4:8], e.Parent)
	order.PutUint32(event[8:12], e.Window)
	order.PutUint32(event[12:16], e.Sibling)
	order.PutUint16(event[16:18], uint16(e.X))
	order.PutUint16(event[18:20], uint16(e.Y))
	order.PutUint16(event[20:22], e.Width)
	order.PutUint16(event[22:24], e.Height)
	order.PutUint16(event[24:26], e.BorderWidth)
	order.PutUint16(event[26:28], e.ValueMask)
	// event[28:32] is unused
	return event
}

// GravityNotify: 24
type GravityNotifyEvent struct {
	Sequence uint16
	Event    uint32
	Window   uint32
	X, Y     int16
}

func (e *GravityNotifyEvent) EncodeMessage(order binary.ByteOrder) []byte {
	event := make([]byte, 32)
	event[0] = 24 // GravityNotify event code
	// byte 1 is unused
	order.PutUint16(event[2:4], e.Sequence)
	order.PutUint32(event[4:8], e.Event)
	order.PutUint32(event[8:12], e.Window)
	order.PutUint16(event[12:14], uint16(e.X))
	order.PutUint16(event[14:16], uint16(e.Y))
	// event[16:32] is unused
	return event
}

// ResizeRequest: 25
type ResizeRequestEvent struct {
	Sequence      uint16
	Window        uint32
	Width, Height uint16
}

func (e *ResizeRequestEvent) EncodeMessage(order binary.ByteOrder) []byte {
	event := make([]byte, 32)
	event[0] = 25 // ResizeRequest event code
	// byte 1 is unused
	order.PutUint16(event[2:4], e.Sequence)
	order.PutUint32(event[4:8], e.Window)
	order.PutUint16(event[8:10], e.Width)
	order.PutUint16(event[10:12], e.Height)
	// event[12:32] is unused
	return event
}

// CirculateNotify: 26
type CirculateNotifyEvent struct {
	Sequence uint16
	Event    uint32
	Window   uint32
	Place    byte
}

func (e *CirculateNotifyEvent) EncodeMessage(order binary.ByteOrder) []byte {
	event := make([]byte, 32)
	event[0] = 26 // CirculateNotify event code
	// byte 1 is unused
	order.PutUint16(event[2:4], e.Sequence)
	order.PutUint32(event[4:8], e.Event)
	order.PutUint32(event[8:12], e.Window)
	event[16] = e.Place
	// event[17:32] is unused
	return event
}

// CirculateRequest: 27
type CirculateRequestEvent struct {
	Sequence uint16
	Parent   uint32
	Window   uint32
	Place    byte
}

func (e *CirculateRequestEvent) EncodeMessage(order binary.ByteOrder) []byte {
	event := make([]byte, 32)
	event[0] = 27 // CirculateRequest event code
	// byte 1 is unused
	order.PutUint16(event[2:4], e.Sequence)
	order.PutUint32(event[4:8], e.Parent)
	order.PutUint32(event[8:12], e.Window)
	event[16] = e.Place
	// event[17:32] is unused
	return event
}

// PropertyNotify: 28
type PropertyNotifyEvent struct {
	Sequence uint16
	Window   uint32
	Atom     uint32
	Time     uint32
	State    byte
}

func (e *PropertyNotifyEvent) EncodeMessage(order binary.ByteOrder) []byte {
	event := make([]byte, 32)
	event[0] = 28 // PropertyNotify event code
	// byte 1 is unused
	order.PutUint16(event[2:4], e.Sequence)
	order.PutUint32(event[4:8], e.Window)
	order.PutUint32(event[8:12], e.Atom)
	order.PutUint32(event[12:16], e.Time)
	event[16] = e.State
	// event[17:32] is unused
	return event
}

// SelectionClear: 29
type SelectionClearEvent struct {
	Sequence  uint16
	Owner     uint32
	Selection uint32
	Time      uint32
}

func (e *SelectionClearEvent) EncodeMessage(order binary.ByteOrder) []byte {
	event := make([]byte, 32)
	event[0] = 29 // SelectionClear event code
	// byte 1 is unused
	order.PutUint16(event[2:4], e.Sequence)
	order.PutUint32(event[4:8], e.Time)
	order.PutUint32(event[8:12], e.Owner)
	order.PutUint32(event[12:16], e.Selection)
	// event[16:32] is unused
	return event
}

// SelectionRequest: 30
type SelectionRequestEvent struct {
	Sequence  uint16
	Owner     uint32
	Requestor uint32
	Selection uint32
	Target    uint32
	Property  uint32
	Time      uint32
}

func (e *SelectionRequestEvent) EncodeMessage(order binary.ByteOrder) []byte {
	event := make([]byte, 32)
	event[0] = 30 // SelectionRequest event code
	// byte 1 is unused
	order.PutUint16(event[2:4], e.Sequence)
	order.PutUint32(event[4:8], e.Time)
	order.PutUint32(event[8:12], e.Owner)
	order.PutUint32(event[12:16], e.Requestor)
	order.PutUint32(event[16:20], e.Selection)
	order.PutUint32(event[20:24], e.Target)
	order.PutUint32(event[24:28], e.Property)
	return event
}

// MappingNotify: 34
type MappingNotifyEvent struct {
	Sequence     uint16
	Request      byte
	FirstKeycode byte
	Count        byte
}

func (e *MappingNotifyEvent) EncodeMessage(order binary.ByteOrder) []byte {
	event := make([]byte, 32)
	event[0] = 34 // MappingNotify event code
	// byte 1 is unused
	order.PutUint16(event[2:4], e.Sequence)
	event[4] = e.Request
	event[5] = e.FirstKeycode
	event[6] = e.Count
	// event[7:32] is unused
	return event
}

// GenericEvent: 35
type GenericEventData struct {
	Sequence  uint16
	Extension byte
	EventType uint16
	Length    uint32
	EventData []byte
}

func (e *GenericEventData) EncodeMessage(order binary.ByteOrder) []byte {
	event := make([]byte, 32)
	event[0] = 35 // GenericEvent event code
	event[1] = e.Extension
	order.PutUint16(event[2:4], e.Sequence)
	order.PutUint32(event[4:8], e.Length)
	order.PutUint16(event[8:10], e.EventType)
	copy(event[12:32], e.EventData)
	return event
}

func (e *DeviceMotionNotifyEvent) EncodeMessage(order binary.ByteOrder) []byte {
	buf := make([]byte, 32)
	buf[0] = byte(XInputOpcode)
	buf[1] = 6 // DeviceMotionNotify
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

func (e *ProximityInEvent) EncodeMessage(order binary.ByteOrder) []byte {
	buf := make([]byte, 32)
	buf[0] = byte(XInputOpcode)
	buf[1] = 8 // ProximityIn
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

func (e *ProximityOutEvent) EncodeMessage(order binary.ByteOrder) []byte {
	buf := make([]byte, 32)
	buf[0] = byte(XInputOpcode)
	buf[1] = 9 // ProximityOut
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

func ParseKeyEvent(buf []byte, order binary.ByteOrder) (*KeyEvent, error) {
	e := &KeyEvent{}
	e.Opcode = buf[0]
	e.Detail = buf[1]
	e.Sequence = order.Uint16(buf[2:4])
	e.Time = order.Uint32(buf[4:8])
	e.Root = order.Uint32(buf[8:12])
	e.Event = order.Uint32(buf[12:16])
	e.Child = order.Uint32(buf[16:20])
	e.RootX = int16(order.Uint16(buf[20:22]))
	e.RootY = int16(order.Uint16(buf[22:24]))
	e.EventX = int16(order.Uint16(buf[24:26]))
	e.EventY = int16(order.Uint16(buf[26:28]))
	e.State = order.Uint16(buf[28:30])
	e.SameScreen = ByteToBool(buf[30])
	return e, nil
}

func ParseButtonPressEvent(buf []byte, order binary.ByteOrder) (*ButtonPressEvent, error) {
	e := &ButtonPressEvent{}
	e.Detail = buf[1]
	e.Sequence = order.Uint16(buf[2:4])
	e.Time = order.Uint32(buf[4:8])
	e.Root = order.Uint32(buf[8:12])
	e.Event = order.Uint32(buf[12:16])
	e.Child = order.Uint32(buf[16:20])
	e.RootX = int16(order.Uint16(buf[20:22]))
	e.RootY = int16(order.Uint16(buf[22:24]))
	e.EventX = int16(order.Uint16(buf[24:26]))
	e.EventY = int16(order.Uint16(buf[26:28]))
	e.State = order.Uint16(buf[28:30])
	e.SameScreen = ByteToBool(buf[30])
	return e, nil
}

func ParseDeviceButtonReleaseEvent(buf []byte, order binary.ByteOrder) (*DeviceButtonReleaseEvent, error) {
	e := &DeviceButtonReleaseEvent{}
	e.Sequence = order.Uint16(buf[2:4])
	e.Time = order.Uint32(buf[4:8])
	e.Root = order.Uint32(buf[8:12])
	e.Event = order.Uint32(buf[12:16])
	e.Child = order.Uint32(buf[16:20])
	e.RootX = int16(order.Uint16(buf[20:22]))
	e.RootY = int16(order.Uint16(buf[22:24]))
	e.EventX = int16(order.Uint16(buf[24:26]))
	e.EventY = int16(order.Uint16(buf[26:28]))
	e.State = order.Uint16(buf[28:30])
	e.DeviceID = buf[30]
	e.Button = buf[31]
	return e, nil
}

func ParseButtonReleaseEvent(buf []byte, order binary.ByteOrder) (*ButtonReleaseEvent, error) {
	e := &ButtonReleaseEvent{}
	e.Detail = buf[1]
	e.Sequence = order.Uint16(buf[2:4])
	e.Time = order.Uint32(buf[4:8])
	e.Root = order.Uint32(buf[8:12])
	e.Event = order.Uint32(buf[12:16])
	e.Child = order.Uint32(buf[16:20])
	e.RootX = int16(order.Uint16(buf[20:22]))
	e.RootY = int16(order.Uint16(buf[22:24]))
	e.EventX = int16(order.Uint16(buf[24:26]))
	e.EventY = int16(order.Uint16(buf[26:28]))
	e.State = order.Uint16(buf[28:30])
	e.SameScreen = ByteToBool(buf[30])
	return e, nil
}

func ParseMotionNotifyEvent(buf []byte, order binary.ByteOrder) (*MotionNotifyEvent, error) {
	e := &MotionNotifyEvent{}
	e.Detail = buf[1]
	e.Sequence = order.Uint16(buf[2:4])
	e.Time = order.Uint32(buf[4:8])
	e.Root = order.Uint32(buf[8:12])
	e.Event = order.Uint32(buf[12:16])
	e.Child = order.Uint32(buf[16:20])
	e.RootX = int16(order.Uint16(buf[20:22]))
	e.RootY = int16(order.Uint16(buf[22:24]))
	e.EventX = int16(order.Uint16(buf[24:26]))
	e.EventY = int16(order.Uint16(buf[26:28]))
	e.State = order.Uint16(buf[28:30])
	e.SameScreen = ByteToBool(buf[30])
	return e, nil
}

func ParseEnterNotifyEvent(buf []byte, order binary.ByteOrder) (*EnterNotifyEvent, error) {
	e := &EnterNotifyEvent{}
	e.Detail = buf[1]
	e.Sequence = order.Uint16(buf[2:4])
	e.Time = order.Uint32(buf[4:8])
	e.Root = order.Uint32(buf[8:12])
	e.Event = order.Uint32(buf[12:16])
	e.Child = order.Uint32(buf[16:20])
	e.RootX = int16(order.Uint16(buf[20:22]))
	e.RootY = int16(order.Uint16(buf[22:24]))
	e.EventX = int16(order.Uint16(buf[24:26]))
	e.EventY = int16(order.Uint16(buf[26:28]))
	e.State = order.Uint16(buf[28:30])
	e.Mode = buf[30]
	e.SameScreen = (buf[31] & 1) != 0
	e.Focus = (buf[31] & 2) != 0
	return e, nil
}

func ParseLeaveNotifyEvent(buf []byte, order binary.ByteOrder) (*LeaveNotifyEvent, error) {
	e := &LeaveNotifyEvent{}
	e.Detail = buf[1]
	e.Sequence = order.Uint16(buf[2:4])
	e.Time = order.Uint32(buf[4:8])
	e.Root = order.Uint32(buf[8:12])
	e.Event = order.Uint32(buf[12:16])
	e.Child = order.Uint32(buf[16:20])
	e.RootX = int16(order.Uint16(buf[20:22]))
	e.RootY = int16(order.Uint16(buf[22:24]))
	e.EventX = int16(order.Uint16(buf[24:26]))
	e.EventY = int16(order.Uint16(buf[26:28]))
	e.State = order.Uint16(buf[28:30])
	e.Mode = buf[30]
	e.SameScreen = (buf[31] & 1) != 0
	e.Focus = (buf[31] & 2) != 0
	return e, nil
}

func ParseExposeEvent(buf []byte, order binary.ByteOrder) (*ExposeEvent, error) {
	e := &ExposeEvent{}
	e.Sequence = order.Uint16(buf[2:4])
	e.Window = order.Uint32(buf[4:8])
	e.X = order.Uint16(buf[8:10])
	e.Y = order.Uint16(buf[10:12])
	e.Width = order.Uint16(buf[12:14])
	e.Height = order.Uint16(buf[14:16])
	e.Count = order.Uint16(buf[16:18])
	return e, nil
}

func ParseConfigureNotifyEvent(buf []byte, order binary.ByteOrder) (*ConfigureNotifyEvent, error) {
	e := &ConfigureNotifyEvent{}
	e.Sequence = order.Uint16(buf[2:4])
	e.Event = order.Uint32(buf[4:8])
	e.Window = order.Uint32(buf[8:12])
	e.AboveSibling = order.Uint32(buf[12:16])
	e.X = int16(order.Uint16(buf[16:18]))
	e.Y = int16(order.Uint16(buf[18:20]))
	e.Width = order.Uint16(buf[20:22])
	e.Height = order.Uint16(buf[22:24])
	e.BorderWidth = order.Uint16(buf[24:26])
	e.OverrideRedirect = ByteToBool(buf[26])
	return e, nil
}

func ParseSelectionNotifyEvent(buf []byte, order binary.ByteOrder) (*SelectionNotifyEvent, error) {
	e := &SelectionNotifyEvent{}
	e.Sequence = order.Uint16(buf[2:4])
	e.Requestor = order.Uint32(buf[4:8])
	e.Selection = order.Uint32(buf[8:12])
	e.Target = order.Uint32(buf[12:16])
	e.Property = order.Uint32(buf[16:20])
	e.Time = order.Uint32(buf[20:24])
	return e, nil
}

func ParseColormapNotifyEvent(buf []byte, order binary.ByteOrder) (*ColormapNotifyEvent, error) {
	e := &ColormapNotifyEvent{}
	e.Sequence = order.Uint16(buf[2:4])
	e.Window = order.Uint32(buf[4:8])
	e.Colormap = order.Uint32(buf[8:12])
	e.New = ByteToBool(buf[12])
	e.State = buf[13]
	return e, nil
}

func ParseClientMessageEvent(buf []byte, order binary.ByteOrder) (*ClientMessageEvent, error) {
	e := &ClientMessageEvent{}
	e.Format = buf[1]
	e.Sequence = order.Uint16(buf[2:4])
	e.Window = order.Uint32(buf[4:8])
	e.MessageType = order.Uint32(buf[8:12])
	copy(e.Data[:], buf[12:32])
	return e, nil
}

func ParseDeviceKeyPressEvent(buf []byte, order binary.ByteOrder) (*DeviceKeyPressEvent, error) {
	e := &DeviceKeyPressEvent{}
	e.Sequence = order.Uint16(buf[2:4])
	e.Time = order.Uint32(buf[4:8])
	e.Root = order.Uint32(buf[8:12])
	e.Event = order.Uint32(buf[12:16])
	e.Child = order.Uint32(buf[16:20])
	e.RootX = int16(order.Uint16(buf[20:22]))
	e.RootY = int16(order.Uint16(buf[22:24]))
	e.EventX = int16(order.Uint16(buf[24:26]))
	e.EventY = int16(order.Uint16(buf[26:28]))
	e.State = order.Uint16(buf[28:30])
	e.DeviceID = buf[30]
	e.KeyCode = buf[31]
	return e, nil
}

func ParseDeviceKeyReleaseEvent(buf []byte, order binary.ByteOrder) (*DeviceKeyReleaseEvent, error) {
	e := &DeviceKeyReleaseEvent{}
	e.Sequence = order.Uint16(buf[2:4])
	e.Time = order.Uint32(buf[4:8])
	e.Root = order.Uint32(buf[8:12])
	e.Event = order.Uint32(buf[12:16])
	e.Child = order.Uint32(buf[16:20])
	e.RootX = int16(order.Uint16(buf[20:22]))
	e.RootY = int16(order.Uint16(buf[22:24]))
	e.EventX = int16(order.Uint16(buf[24:26]))
	e.EventY = int16(order.Uint16(buf[26:28]))
	e.State = order.Uint16(buf[28:30])
	e.DeviceID = buf[30]
	e.KeyCode = buf[31]
	return e, nil
}

func ParseDeviceButtonPressEvent(buf []byte, order binary.ByteOrder) (*DeviceButtonPressEvent, error) {
	e := &DeviceButtonPressEvent{}
	e.Sequence = order.Uint16(buf[2:4])
	e.Time = order.Uint32(buf[4:8])
	e.Root = order.Uint32(buf[8:12])
	e.Event = order.Uint32(buf[12:16])
	e.Child = order.Uint32(buf[16:20])
	e.RootX = int16(order.Uint16(buf[20:22]))
	e.RootY = int16(order.Uint16(buf[22:24]))
	e.EventX = int16(order.Uint16(buf[24:26]))
	e.EventY = int16(order.Uint16(buf[26:28]))
	e.State = order.Uint16(buf[28:30])
	e.DeviceID = buf[30]
	e.Detail = buf[31]
	return e, nil
}

func ParseGraphicsExposureEvent(buf []byte, order binary.ByteOrder) (*GraphicsExposureEvent, error) {
	e := &GraphicsExposureEvent{}
	e.Sequence = order.Uint16(buf[2:4])
	e.Drawable = order.Uint32(buf[4:8])
	e.X = order.Uint16(buf[8:10])
	e.Y = order.Uint16(buf[10:12])
	e.Width = order.Uint16(buf[12:14])
	e.Height = order.Uint16(buf[14:16])
	e.MinorOpcode = order.Uint16(buf[16:18])
	e.Count = order.Uint16(buf[18:20])
	e.MajorOpcode = buf[20]
	return e, nil
}

func ParseNoExposureEvent(buf []byte, order binary.ByteOrder) (*NoExposureEvent, error) {
	e := &NoExposureEvent{}
	e.Sequence = order.Uint16(buf[2:4])
	e.Drawable = order.Uint32(buf[4:8])
	e.MinorOpcode = order.Uint16(buf[8:10])
	e.MajorOpcode = buf[10]
	return e, nil
}

func ParseVisibilityNotifyEvent(buf []byte, order binary.ByteOrder) (*VisibilityNotifyEvent, error) {
	e := &VisibilityNotifyEvent{}
	e.Sequence = order.Uint16(buf[2:4])
	e.Window = order.Uint32(buf[4:8])
	e.State = buf[8]
	return e, nil
}

func ParseCreateNotifyEvent(buf []byte, order binary.ByteOrder) (*CreateNotifyEvent, error) {
	e := &CreateNotifyEvent{}
	e.Sequence = order.Uint16(buf[2:4])
	e.Parent = order.Uint32(buf[4:8])
	e.Window = order.Uint32(buf[8:12])
	e.X = int16(order.Uint16(buf[12:14]))
	e.Y = int16(order.Uint16(buf[14:16]))
	e.Width = order.Uint16(buf[16:18])
	e.Height = order.Uint16(buf[18:20])
	e.BorderWidth = order.Uint16(buf[20:22])
	e.OverrideRedirect = ByteToBool(buf[22])
	return e, nil
}

func ParseDestroyNotifyEvent(buf []byte, order binary.ByteOrder) (*DestroyNotifyEvent, error) {
	e := &DestroyNotifyEvent{}
	e.Sequence = order.Uint16(buf[2:4])
	e.Event = order.Uint32(buf[4:8])
	e.Window = order.Uint32(buf[8:12])
	return e, nil
}

func ParseUnmapNotifyEvent(buf []byte, order binary.ByteOrder) (*UnmapNotifyEvent, error) {
	e := &UnmapNotifyEvent{}
	e.Sequence = order.Uint16(buf[2:4])
	e.Event = order.Uint32(buf[4:8])
	e.Window = order.Uint32(buf[8:12])
	e.FromConfigure = ByteToBool(buf[12])
	return e, nil
}

func ParseMapNotifyEvent(buf []byte, order binary.ByteOrder) (*MapNotifyEvent, error) {
	e := &MapNotifyEvent{}
	e.Sequence = order.Uint16(buf[2:4])
	e.Event = order.Uint32(buf[4:8])
	e.Window = order.Uint32(buf[8:12])
	e.OverrideRedirect = ByteToBool(buf[12])
	return e, nil
}

func ParseMapRequestEvent(buf []byte, order binary.ByteOrder) (*MapRequestEvent, error) {
	e := &MapRequestEvent{}
	e.Sequence = order.Uint16(buf[2:4])
	e.Parent = order.Uint32(buf[4:8])
	e.Window = order.Uint32(buf[8:12])
	return e, nil
}

func ParseReparentNotifyEvent(buf []byte, order binary.ByteOrder) (*ReparentNotifyEvent, error) {
	e := &ReparentNotifyEvent{}
	e.Sequence = order.Uint16(buf[2:4])
	e.Event = order.Uint32(buf[4:8])
	e.Window = order.Uint32(buf[8:12])
	e.Parent = order.Uint32(buf[12:16])
	e.X = int16(order.Uint16(buf[16:18]))
	e.Y = int16(order.Uint16(buf[18:20]))
	e.OverrideRedirect = ByteToBool(buf[20])
	return e, nil
}

func ParseConfigureRequestEvent(buf []byte, order binary.ByteOrder) (*ConfigureRequestEvent, error) {
	e := &ConfigureRequestEvent{}
	e.StackMode = buf[1]
	e.Sequence = order.Uint16(buf[2:4])
	e.Parent = order.Uint32(buf[4:8])
	e.Window = order.Uint32(buf[8:12])
	e.Sibling = order.Uint32(buf[12:16])
	e.X = int16(order.Uint16(buf[16:18]))
	e.Y = int16(order.Uint16(buf[18:20]))
	e.Width = order.Uint16(buf[20:22])
	e.Height = order.Uint16(buf[22:24])
	e.BorderWidth = order.Uint16(buf[24:26])
	e.ValueMask = order.Uint16(buf[26:28])
	return e, nil
}

func ParseGravityNotifyEvent(buf []byte, order binary.ByteOrder) (*GravityNotifyEvent, error) {
	e := &GravityNotifyEvent{}
	e.Sequence = order.Uint16(buf[2:4])
	e.Event = order.Uint32(buf[4:8])
	e.Window = order.Uint32(buf[8:12])
	e.X = int16(order.Uint16(buf[12:14]))
	e.Y = int16(order.Uint16(buf[14:16]))
	return e, nil
}

func ParseResizeRequestEvent(buf []byte, order binary.ByteOrder) (*ResizeRequestEvent, error) {
	e := &ResizeRequestEvent{}
	e.Sequence = order.Uint16(buf[2:4])
	e.Window = order.Uint32(buf[4:8])
	e.Width = order.Uint16(buf[8:10])
	e.Height = order.Uint16(buf[10:12])
	return e, nil
}

func ParseCirculateNotifyEvent(buf []byte, order binary.ByteOrder) (*CirculateNotifyEvent, error) {
	e := &CirculateNotifyEvent{}
	e.Sequence = order.Uint16(buf[2:4])
	e.Event = order.Uint32(buf[4:8])
	e.Window = order.Uint32(buf[8:12])
	e.Place = buf[16]
	return e, nil
}

func ParseCirculateRequestEvent(buf []byte, order binary.ByteOrder) (*CirculateRequestEvent, error) {
	e := &CirculateRequestEvent{}
	e.Sequence = order.Uint16(buf[2:4])
	e.Parent = order.Uint32(buf[4:8])
	e.Window = order.Uint32(buf[8:12])
	e.Place = buf[16]
	return e, nil
}

func ParsePropertyNotifyEvent(buf []byte, order binary.ByteOrder) (*PropertyNotifyEvent, error) {
	e := &PropertyNotifyEvent{}
	e.Sequence = order.Uint16(buf[2:4])
	e.Window = order.Uint32(buf[4:8])
	e.Atom = order.Uint32(buf[8:12])
	e.Time = order.Uint32(buf[12:16])
	e.State = buf[16]
	return e, nil
}

func ParseSelectionClearEvent(buf []byte, order binary.ByteOrder) (*SelectionClearEvent, error) {
	e := &SelectionClearEvent{}
	e.Sequence = order.Uint16(buf[2:4])
	e.Time = order.Uint32(buf[4:8])
	e.Owner = order.Uint32(buf[8:12])
	e.Selection = order.Uint32(buf[12:16])
	return e, nil
}

func ParseSelectionRequestEvent(buf []byte, order binary.ByteOrder) (*SelectionRequestEvent, error) {
	e := &SelectionRequestEvent{}
	e.Sequence = order.Uint16(buf[2:4])
	e.Time = order.Uint32(buf[4:8])
	e.Owner = order.Uint32(buf[8:12])
	e.Requestor = order.Uint32(buf[12:16])
	e.Selection = order.Uint32(buf[16:20])
	e.Target = order.Uint32(buf[20:24])
	e.Property = order.Uint32(buf[24:28])
	return e, nil
}

func ParseMappingNotifyEvent(buf []byte, order binary.ByteOrder) (*MappingNotifyEvent, error) {
	e := &MappingNotifyEvent{}
	e.Sequence = order.Uint16(buf[2:4])
	e.Request = buf[4]
	e.FirstKeycode = buf[5]
	e.Count = buf[6]
	return e, nil
}

func ParseGenericEvent(buf []byte, order binary.ByteOrder) (*GenericEventData, error) {
	e := &GenericEventData{}
	e.Extension = buf[1]
	e.Sequence = order.Uint16(buf[2:4])
	e.Length = order.Uint32(buf[4:8])
	e.EventType = order.Uint16(buf[8:10])
	e.EventData = buf[12:32]
	return e, nil
}

func ParseDeviceMotionNotifyEvent(buf []byte, order binary.ByteOrder) (*DeviceMotionNotifyEvent, error) {
	e := &DeviceMotionNotifyEvent{}
	e.Sequence = order.Uint16(buf[2:4])
	e.Time = order.Uint32(buf[4:8])
	e.Root = order.Uint32(buf[8:12])
	e.Event = order.Uint32(buf[12:16])
	e.Child = order.Uint32(buf[16:20])
	e.RootX = int16(order.Uint16(buf[20:22]))
	e.RootY = int16(order.Uint16(buf[22:24]))
	e.EventX = int16(order.Uint16(buf[24:26]))
	e.EventY = int16(order.Uint16(buf[26:28]))
	e.State = order.Uint16(buf[28:30])
	e.DeviceID = buf[30]
	e.Detail = buf[31]
	return e, nil
}

func ParseProximityInEvent(buf []byte, order binary.ByteOrder) (*ProximityInEvent, error) {
	e := &ProximityInEvent{}
	e.Sequence = order.Uint16(buf[2:4])
	e.Time = order.Uint32(buf[4:8])
	e.Root = order.Uint32(buf[8:12])
	e.Event = order.Uint32(buf[12:16])
	e.Child = order.Uint32(buf[16:20])
	e.RootX = int16(order.Uint16(buf[20:22]))
	e.RootY = int16(order.Uint16(buf[22:24]))
	e.EventX = int16(order.Uint16(buf[24:26]))
	e.EventY = int16(order.Uint16(buf[26:28]))
	e.State = order.Uint16(buf[28:30])
	e.DeviceID = buf[30]
	e.Detail = buf[31]
	return e, nil
}

func ParseProximityOutEvent(buf []byte, order binary.ByteOrder) (*ProximityOutEvent, error) {
	e := &ProximityOutEvent{}
	e.Sequence = order.Uint16(buf[2:4])
	e.Time = order.Uint32(buf[4:8])
	e.Root = order.Uint32(buf[8:12])
	e.Event = order.Uint32(buf[12:16])
	e.Child = order.Uint32(buf[16:20])
	e.RootX = int16(order.Uint16(buf[20:22]))
	e.RootY = int16(order.Uint16(buf[22:24]))
	e.EventX = int16(order.Uint16(buf[24:26]))
	e.EventY = int16(order.Uint16(buf[26:28]))
	e.State = order.Uint16(buf[28:30])
	e.DeviceID = buf[30]
	e.Detail = buf[31]
	return e, nil
}

type Event interface {
	EncodeMessage(order binary.ByteOrder) []byte
}

func ParseEvent(buf []byte, order binary.ByteOrder) (Event, error) {
	if len(buf) < 32 {
		return nil, fmt.Errorf("event message too short: %d", len(buf))
	}
	switch buf[0] {
	case KeyPress, KeyRelease:
		return ParseKeyEvent(buf, order)
	case ButtonPress:
		return ParseButtonPressEvent(buf, order)
	case ButtonRelease:
		return ParseButtonReleaseEvent(buf, order)
	case MotionNotify:
		return ParseMotionNotifyEvent(buf, order)
	case EnterNotify:
		return ParseEnterNotifyEvent(buf, order)
	case LeaveNotify:
		return ParseLeaveNotifyEvent(buf, order)
	case Expose:
		return ParseExposeEvent(buf, order)
	case ConfigureNotify:
		return ParseConfigureNotifyEvent(buf, order)
	case GraphicsExposure:
		return ParseGraphicsExposureEvent(buf, order)
	case NoExposure:
		return ParseNoExposureEvent(buf, order)
	case VisibilityNotify:
		return ParseVisibilityNotifyEvent(buf, order)
	case CreateNotify:
		return ParseCreateNotifyEvent(buf, order)
	case DestroyNotify:
		return ParseDestroyNotifyEvent(buf, order)
	case UnmapNotify:
		return ParseUnmapNotifyEvent(buf, order)
	case MapNotify:
		return ParseMapNotifyEvent(buf, order)
	case MapRequest:
		return ParseMapRequestEvent(buf, order)
	case ReparentNotify:
		return ParseReparentNotifyEvent(buf, order)
	case ConfigureRequest:
		return ParseConfigureRequestEvent(buf, order)
	case GravityNotify:
		return ParseGravityNotifyEvent(buf, order)
	case ResizeRequest:
		return ParseResizeRequestEvent(buf, order)
	case CirculateNotify:
		return ParseCirculateNotifyEvent(buf, order)
	case CirculateRequest:
		return ParseCirculateRequestEvent(buf, order)
	case PropertyNotify:
		return ParsePropertyNotifyEvent(buf, order)
	case SelectionClear:
		return ParseSelectionClearEvent(buf, order)
	case SelectionRequest:
		return ParseSelectionRequestEvent(buf, order)
	case SelectionNotify:
		return ParseSelectionNotifyEvent(buf, order)
	case ColormapNotifyCode:
		return ParseColormapNotifyEvent(buf, order)
	case ClientMessage:
		return ParseClientMessageEvent(buf, order)
	case MappingNotify:
		return ParseMappingNotifyEvent(buf, order)
	case GenericEvent:
		return ParseGenericEvent(buf, order)
	case byte(XInputOpcode):
		switch buf[1] {
		case DeviceKeyPress:
			return ParseDeviceKeyPressEvent(buf, order)
		case DeviceKeyRelease:
			return ParseDeviceKeyReleaseEvent(buf, order)
		case DeviceButtonPress:
			return ParseDeviceButtonPressEvent(buf, order)
		case DeviceButtonRelease:
			return ParseDeviceButtonReleaseEvent(buf, order)
		case DeviceMotionNotify:
			return ParseDeviceMotionNotifyEvent(buf, order)
		case ProximityIn:
			return ParseProximityInEvent(buf, order)
		case ProximityOut:
			return ParseProximityOutEvent(buf, order)
		}
	}
	return nil, fmt.Errorf("unknown event opcode: %d", buf[0])
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

type DeviceMotionNotifyEvent struct {
	Sequence   uint16
	DeviceID   byte
	Time       uint32
	Detail     byte
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

type ProximityInEvent struct {
	Sequence   uint16
	DeviceID   byte
	Time       uint32
	Detail     byte
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

type ProximityOutEvent struct {
	Sequence   uint16
	DeviceID   byte
	Time       uint32
	Detail     byte
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
