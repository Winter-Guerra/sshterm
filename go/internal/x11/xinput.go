//go:build x11

package x11

import (
	"bytes"
	"encoding/binary"
)

const (
	XInputExtensionName = "XInputExtension"
	xInputOpcode        = 134
)

// XInput minor opcodes
const (
	XListInputDevices = 2
)

// XInput request types
type XInputRequest struct {
	MinorOpcode byte
	Body        []byte
}

func (r *XInputRequest) OpCode() reqCode {
	return xInputOpcode
}

func parseXInputRequest(order binary.ByteOrder, data byte, body []byte, seq uint16) (*XInputRequest, error) {
	return &XInputRequest{
		MinorOpcode: data,
		Body:        body,
	}, nil
}

// ListInputDevices request
type ListInputDevicesRequest struct{}

func parseListInputDevicesRequest(order binary.ByteOrder, body []byte, seq uint16) (*ListInputDevicesRequest, error) {
	if len(body) != 0 {
		// As per spec, the request has no body.
		return nil, NewError(LengthErrorCode, seq, 0, xInputOpcode, XListInputDevices)
	}
	return &ListInputDevicesRequest{}, nil
}

// ListInputDevices reply
// This is a complex, variable-length reply.

type ListInputDevicesReply struct {
	sequence uint16
	devices  []DeviceInfo
}

type DeviceInfo interface {
	encode(order binary.ByteOrder) []byte
}

type deviceHeader struct {
	DeviceID   byte
	DeviceType Atom
	NumClasses byte
	Use        byte // 0: IsXPointer, 1: IsXKeyboard, 2: IsXExtensionDevice
	Name       string
}

type InputClassInfo interface {
	encode(order binary.ByteOrder) []byte
	classID() byte
	length() int
}

type KeyClassInfo struct {
	NumKeys    uint16
	MinKeycode byte
	MaxKeycode byte
}

func (c *KeyClassInfo) classID() byte { return 0 } // KeyClass
func (c *KeyClassInfo) length() int   { return 8 }
func (c *KeyClassInfo) encode(order binary.ByteOrder) []byte {
	buf := new(bytes.Buffer)
	buf.WriteByte(c.classID())
	buf.WriteByte(byte(c.length()))
	buf.WriteByte(c.MinKeycode)
	buf.WriteByte(c.MaxKeycode)
	binary.Write(buf, order, c.NumKeys)
	buf.Write([]byte{0, 0}) // padding
	return buf.Bytes()
}

type ButtonClassInfo struct {
	NumButtons uint16
}

func (c *ButtonClassInfo) classID() byte { return 1 } // ButtonClass
func (c *ButtonClassInfo) length() int   { return 8 }
func (c *ButtonClassInfo) encode(order binary.ByteOrder) []byte {
	buf := new(bytes.Buffer)
	buf.WriteByte(c.classID())
	buf.WriteByte(byte(c.length()))
	binary.Write(buf, order, c.NumButtons)
	buf.Write([]byte{0, 0, 0, 0}) // padding
	return buf.Bytes()
}

type ValuatorClassInfo struct {
	NumAxes    byte
	Mode       byte
	MotionSize uint32
	Axes       []ValuatorAxisInfo
}

type ValuatorAxisInfo struct {
	Min        int32
	Max        int32
	Resolution uint32
}

func (c *ValuatorClassInfo) classID() byte { return 2 } // ValuatorClass
func (c *ValuatorClassInfo) length() int {
	return 8 + len(c.Axes)*12
}
func (c *ValuatorClassInfo) encode(order binary.ByteOrder) []byte {
	buf := new(bytes.Buffer)
	buf.WriteByte(c.classID())
	buf.WriteByte(byte(c.length()))
	buf.WriteByte(c.NumAxes)
	buf.WriteByte(c.Mode)
	binary.Write(buf, order, c.MotionSize)
	for _, axis := range c.Axes {
		binary.Write(buf, order, axis.Resolution)
		binary.Write(buf, order, axis.Min)
		binary.Write(buf, order, axis.Max)
	}
	return buf.Bytes()
}

type deviceImpl struct {
	header  deviceHeader
	classes []InputClassInfo
}

func (d *deviceImpl) encode(order binary.ByteOrder) []byte {
	buf := new(bytes.Buffer)
	nameBytes := []byte(d.header.Name)
	nameLen := len(nameBytes)

	// Encode classes to a temporary buffer to get their total length
	classesBuf := new(bytes.Buffer)
	for _, class := range d.classes {
		classesBuf.Write(class.encode(order))
	}
	classesBytes := classesBuf.Bytes()
	classesLen := len(classesBytes)

	buf.WriteByte(d.header.DeviceID)
	buf.WriteByte(byte(len(d.classes)))
	buf.WriteByte(d.header.Use)
	buf.WriteByte(0) // unused
	binary.Write(buf, order, d.header.DeviceType)
	binary.Write(buf, order, uint16(nameLen))

	// Write class data length
	binary.Write(buf, order, uint16(classesLen))

	buf.Write(nameBytes)

	for _, class := range d.classes {
		buf.Write(class.encode(order))
	}

	return buf.Bytes()
}

func (r *ListInputDevicesReply) encodeMessage(order binary.ByteOrder) []byte {
	var devicesData []byte
	for _, dev := range r.devices {
		devicesData = append(devicesData, dev.encode(order)...)
	}

	p := (4 - (len(devicesData) % 4)) % 4
	if p == 4 {
		p = 0
	}
	finalDeviceData := make([]byte, len(devicesData)+p)
	copy(finalDeviceData, devicesData)

	reply := make([]byte, 32+len(finalDeviceData))
	reply[0] = 1 // Reply
	reply[1] = byte(len(r.devices))
	order.PutUint16(reply[2:4], r.sequence)
	order.PutUint32(reply[4:8], uint32((len(finalDeviceData)+3)/4)) // length
	copy(reply[32:], finalDeviceData)
	return reply
}

func handleXInputRequest(client *x11Client, minorOpcode byte, body []byte, seq uint16) (reply messageEncoder) {
	switch minorOpcode {
	case XListInputDevices:
		return &ListInputDevicesReply{
			sequence: seq,
			devices:  []DeviceInfo{},
		}
	default:
		// TODO: Implement other XInput requests
		return nil
	}
}
