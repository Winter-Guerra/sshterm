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
	XListInputDevices     = 2
	XOpenDevice           = 3
	XCloseDevice          = 4
	XSelectExtensionEvent = 16
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

// OpenDevice request
type OpenDeviceRequest struct {
	DeviceID byte
}

func parseOpenDeviceRequest(order binary.ByteOrder, body []byte, seq uint16) (*OpenDeviceRequest, error) {
	if len(body) != 4 {
		return nil, NewError(LengthErrorCode, seq, 0, xInputOpcode, XOpenDevice)
	}
	return &OpenDeviceRequest{
		DeviceID: body[0],
	}, nil
}

// OpenDevice reply
type OpenDeviceReply struct {
	sequence uint16
	classes  []InputClassInfo
}

func (r *OpenDeviceReply) encodeMessage(order binary.ByteOrder) []byte {
	classesBuf := new(bytes.Buffer)
	for _, class := range r.classes {
		classesBuf.Write(class.encode(order))
	}
	classesBytes := classesBuf.Bytes()

	reply := make([]byte, 32+len(classesBytes))
	reply[0] = 1 // Reply
	reply[1] = byte(len(r.classes))
	order.PutUint16(reply[2:4], r.sequence)
	order.PutUint32(reply[4:8], uint32((len(classesBytes)+3)/4)) // length
	copy(reply[32:], classesBytes)
	return reply
}

// CloseDevice request
type CloseDeviceRequest struct {
	DeviceID byte
}

func parseCloseDeviceRequest(order binary.ByteOrder, body []byte, seq uint16) (*CloseDeviceRequest, error) {
	if len(body) != 4 {
		return nil, NewError(LengthErrorCode, seq, 0, xInputOpcode, XCloseDevice)
	}
	return &CloseDeviceRequest{
		DeviceID: body[0],
	}, nil
}

// CloseDevice reply
type CloseDeviceReply struct {
	sequence uint16
}

func (r *CloseDeviceReply) encodeMessage(order binary.ByteOrder) []byte {
	reply := make([]byte, 32)
	reply[0] = 1 // Reply
	reply[1] = 0
	order.PutUint16(reply[2:4], r.sequence)
	order.PutUint32(reply[4:8], 0) // length
	return reply
}

// ListInputDevices reply
// This is a complex, variable-length reply.

type ListInputDevicesReply struct {
	sequence uint16
	devices  []DeviceInfo
	NDevices byte
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

type deviceInfo struct {
	header     deviceHeader
	classes    []InputClassInfo
	eventMasks map[uint32]uint32 // window ID -> event mask
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

func (d *deviceInfo) encode(order binary.ByteOrder) []byte {
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

var virtualPointer = &deviceInfo{
	header: deviceHeader{
		DeviceID:   2,
		DeviceType: 0,
		NumClasses: 2,
		Use:        0, // IsXPointer
		Name:       "Virtual Pointer",
	},
	classes: []InputClassInfo{
		&ButtonClassInfo{NumButtons: 5},
		&ValuatorClassInfo{
			NumAxes:    2,
			Mode:       0, // Relative
			MotionSize: 0,
			Axes: []ValuatorAxisInfo{
				{Min: 0, Max: 65535, Resolution: 1},
				{Min: 0, Max: 65535, Resolution: 1},
			},
		},
	},
}

var virtualKeyboard = &deviceInfo{
	header: deviceHeader{
		DeviceID:   3,
		DeviceType: 0,
		NumClasses: 1,
		Use:        1, // IsXKeyboard
		Name:       "Virtual Keyboard",
	},
	classes: []InputClassInfo{
		&KeyClassInfo{
			NumKeys:    248,
			MinKeycode: 8,
			MaxKeycode: 255,
		},
	},
}

func handleXInputRequest(client *x11Client, minorOpcode byte, body []byte, seq uint16) (reply messageEncoder) {
	switch minorOpcode {
	case XListInputDevices:
		return &ListInputDevicesReply{
			sequence: seq,
			devices:  []DeviceInfo{virtualPointer, virtualKeyboard},
		}
	case XOpenDevice:
		req, err := parseOpenDeviceRequest(client.byteOrder, body, seq)
		if err != nil {
			return err.(messageEncoder)
		}

		var selectedDevice *deviceInfo
		if req.DeviceID == virtualPointer.header.DeviceID {
			selectedDevice = virtualPointer
		} else if req.DeviceID == virtualKeyboard.header.DeviceID {
			selectedDevice = virtualKeyboard
		} else {
			return NewError(ValueErrorCode, seq, uint32(req.DeviceID), xInputOpcode, XOpenDevice)
		}

		// Create a new deviceInfo instance for the client, so event masks are not shared.
		newClasses := make([]InputClassInfo, len(selectedDevice.classes))
		copy(newClasses, selectedDevice.classes)
		newDeviceInfo := &deviceInfo{
			header:     selectedDevice.header,
			classes:    newClasses,
			eventMasks: make(map[uint32]uint32),
		}
		client.openDevices[req.DeviceID] = newDeviceInfo
		return &OpenDeviceReply{sequence: seq, classes: newDeviceInfo.classes}

	case XCloseDevice:
		req, err := parseCloseDeviceRequest(client.byteOrder, body, seq)
		if err != nil {
			return err.(messageEncoder)
		}
		delete(client.openDevices, req.DeviceID)
		return &CloseDeviceReply{sequence: seq}
	case XSelectExtensionEvent:
		windowID := client.byteOrder.Uint32(body[0:4])
		numClasses := client.byteOrder.Uint16(body[4:6])
		for i := 0; i < int(numClasses); i++ {
			class := client.byteOrder.Uint32(body[8+i*4 : 12+i*4])
			deviceID := byte(class & 0xFF)
			mask := class >> 8
			if dev, ok := client.openDevices[deviceID]; ok {
				if dev.eventMasks == nil {
					dev.eventMasks = make(map[uint32]uint32)
				}
				dev.eventMasks[windowID] = mask
			}
		}
		return nil
	default:
		// TODO: Implement other XInput requests
		return nil
	}
}
