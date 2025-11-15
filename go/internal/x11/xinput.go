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
	XGetExtensionVersion           = 1
	XListInputDevices              = 2
	XOpenDevice                    = 3
	XCloseDevice                   = 4
	XSetDeviceMode                 = 5
	XSetDeviceValuators            = 6
	XGetDeviceControl              = 7
	XChangeDeviceControl           = 8
	XSelectExtensionEvent          = 16
	XGetSelectedExtensionEvents    = 17
	XChangeDeviceDontPropagateList = 18
	XGetDeviceDontPropagateList    = 19
	XSendExtensionEvent            = 20
	XAllowDeviceEvents             = 28
	XChangeKeyboardDevice          = 29
	XChangePointerDevice           = 30
	XGrabDevice                    = 26
	XUngrabDevice                  = 27
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

// GetExtensionVersion reply
type GetExtensionVersionReply struct {
	sequence     uint16
	MajorVersion uint16
	MinorVersion uint16
}

func (r *GetExtensionVersionReply) encodeMessage(order binary.ByteOrder) []byte {
	reply := make([]byte, 32)
	reply[0] = 1 // Reply
	reply[1] = 1 // Present
	order.PutUint16(reply[2:4], r.sequence)
	order.PutUint32(reply[4:8], 0) // length
	order.PutUint16(reply[8:10], r.MajorVersion)
	order.PutUint16(reply[10:12], r.MinorVersion)
	return reply
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

// SetDeviceMode request
type SetDeviceModeRequest struct {
	DeviceID byte
	Mode     byte
}

func parseSetDeviceModeRequest(order binary.ByteOrder, body []byte, seq uint16) (*SetDeviceModeRequest, error) {
	if len(body) != 4 {
		return nil, NewError(LengthErrorCode, seq, 0, xInputOpcode, XSetDeviceMode)
	}
	return &SetDeviceModeRequest{
		DeviceID: body[0],
		Mode:     body[1],
	}, nil
}

// SetDeviceMode reply
type SetDeviceModeReply struct {
	sequence uint16
	Status   byte
}

func (r *SetDeviceModeReply) encodeMessage(order binary.ByteOrder) []byte {
	reply := make([]byte, 32)
	reply[0] = 1 // Reply
	reply[1] = r.Status
	order.PutUint16(reply[2:4], r.sequence)
	order.PutUint32(reply[4:8], 0) // length
	return reply
}

// SetDeviceValuators request
type SetDeviceValuatorsRequest struct {
	DeviceID      byte
	FirstValuator byte
	NumValuators  byte
	Valuators     []int32
}

func parseSetDeviceValuatorsRequest(order binary.ByteOrder, body []byte, seq uint16) (*SetDeviceValuatorsRequest, error) {
	numValuators := body[2]
	if len(body) != 4+int(numValuators)*4 {
		return nil, NewError(LengthErrorCode, seq, 0, xInputOpcode, XSetDeviceValuators)
	}
	valuators := make([]int32, numValuators)
	for i := 0; i < int(numValuators); i++ {
		valuators[i] = int32(order.Uint32(body[4+i*4 : 8+i*4]))
	}
	return &SetDeviceValuatorsRequest{
		DeviceID:      body[0],
		FirstValuator: body[1],
		NumValuators:  numValuators,
		Valuators:     valuators,
	}, nil
}

// SetDeviceValuators reply
type SetDeviceValuatorsReply struct {
	sequence uint16
	Status   byte
}

func (r *SetDeviceValuatorsReply) encodeMessage(order binary.ByteOrder) []byte {
	reply := make([]byte, 32)
	reply[0] = 1 // Reply
	reply[1] = r.Status
	order.PutUint16(reply[2:4], r.sequence)
	order.PutUint32(reply[4:8], 0) // length
	return reply
}

// Device Control constants
const (
	DeviceResolution = 1
)

// DeviceControl interfaces
type DeviceControlState interface {
	encode(order binary.ByteOrder) []byte
}

type DeviceControl interface {
	encode(order binary.ByteOrder) []byte
}

type DeviceResolutionState struct {
	NumValuators   byte
	Resolutions    []uint32
	MinResolutions []uint32
	MaxResolutions []uint32
}

func (s *DeviceResolutionState) encode(order binary.ByteOrder) []byte {
	length := 8 + int(s.NumValuators)*12
	buf := new(bytes.Buffer)
	buf.Grow(length)
	binary.Write(buf, order, uint16(DeviceResolution))
	binary.Write(buf, order, uint16(length))
	buf.WriteByte(s.NumValuators)
	buf.Write([]byte{0, 0, 0}) // padding
	for _, res := range s.Resolutions {
		binary.Write(buf, order, res)
	}
	for _, res := range s.MinResolutions {
		binary.Write(buf, order, res)
	}
	for _, res := range s.MaxResolutions {
		binary.Write(buf, order, res)
	}
	return buf.Bytes()
}

type DeviceResolutionControl struct {
	FirstValuator byte
	NumValuators  byte
	Resolutions   []uint32
}

func (c *DeviceResolutionControl) encode(order binary.ByteOrder) []byte {
	length := 8 + int(c.NumValuators)*4
	buf := new(bytes.Buffer)
	buf.Grow(length)
	binary.Write(buf, order, uint16(DeviceResolution))
	binary.Write(buf, order, uint16(length))
	buf.WriteByte(c.FirstValuator)
	buf.WriteByte(c.NumValuators)
	buf.Write([]byte{0, 0}) // padding
	for _, res := range c.Resolutions {
		binary.Write(buf, order, res)
	}
	return buf.Bytes()
}

// GetDeviceControl request
type GetDeviceControlRequest struct {
	DeviceID byte
	Control  uint16
}

func parseGetDeviceControlRequest(order binary.ByteOrder, body []byte, seq uint16) (*GetDeviceControlRequest, error) {
	if len(body) != 4 {
		return nil, NewError(LengthErrorCode, seq, 0, xInputOpcode, XGetDeviceControl)
	}
	return &GetDeviceControlRequest{
		DeviceID: body[0],
		Control:  order.Uint16(body[2:4]),
	}, nil
}

// GetDeviceControl reply
type GetDeviceControlReply struct {
	sequence uint16
	Control  DeviceControlState
}

func (r *GetDeviceControlReply) encodeMessage(order binary.ByteOrder) []byte {
	controlBytes := r.Control.encode(order)
	reply := make([]byte, 32+len(controlBytes))
	reply[0] = 1 // Reply
	order.PutUint16(reply[2:4], r.sequence)
	order.PutUint32(reply[4:8], uint32((len(controlBytes)+3)/4)) // length
	copy(reply[32:], controlBytes)
	return reply
}

// ChangeDeviceControl request
type ChangeDeviceControlRequest struct {
	DeviceID byte
	Control  DeviceControl
}

func parseChangeDeviceControlRequest(order binary.ByteOrder, body []byte, seq uint16) (*ChangeDeviceControlRequest, error) {
	controlID := order.Uint16(body[2:4])
	if controlID != DeviceResolution {
		return nil, NewError(ValueErrorCode, seq, 0, xInputOpcode, XChangeDeviceControl)
	}
	length := order.Uint16(body[4:6])
	numValuators := body[6]
	if length != 8+uint16(numValuators)*4 {
		return nil, NewError(LengthErrorCode, seq, 0, xInputOpcode, XChangeDeviceControl)
	}
	resolutions := make([]uint32, numValuators)
	for i := 0; i < int(numValuators); i++ {
		resolutions[i] = order.Uint32(body[8+i*4 : 12+i*4])
	}
	return &ChangeDeviceControlRequest{
		DeviceID: body[0],
		Control: &DeviceResolutionControl{
			FirstValuator: body[5],
			NumValuators:  numValuators,
			Resolutions:   resolutions,
		},
	}, nil
}

// ChangeDeviceControl reply
type ChangeDeviceControlReply struct {
	sequence uint16
	Status   byte
}

func (r *ChangeDeviceControlReply) encodeMessage(order binary.ByteOrder) []byte {
	reply := make([]byte, 32)
	reply[0] = 1 // Reply
	reply[1] = r.Status
	order.PutUint16(reply[2:4], r.sequence)
	order.PutUint32(reply[4:8], 0)
	return reply
}

// GetSelectedExtensionEvents request
type GetSelectedExtensionEventsRequest struct {
	Window uint32
}

func parseGetSelectedExtensionEventsRequest(order binary.ByteOrder, body []byte, seq uint16) (*GetSelectedExtensionEventsRequest, error) {
	if len(body) != 4 {
		return nil, NewError(LengthErrorCode, seq, 0, xInputOpcode, XGetSelectedExtensionEvents)
	}
	return &GetSelectedExtensionEventsRequest{
		Window: order.Uint32(body[0:4]),
	}, nil
}

// GetSelectedExtensionEvents reply
type GetSelectedExtensionEventsReply struct {
	sequence          uint16
	ThisClientClasses []uint32
	AllClientsClasses []uint32
}

func (r *GetSelectedExtensionEventsReply) encodeMessage(order binary.ByteOrder) []byte {
	thisClientLen := len(r.ThisClientClasses)
	allClientsLen := len(r.AllClientsClasses)
	length := (thisClientLen + allClientsLen) * 4
	reply := make([]byte, 32+length)
	reply[0] = 1 // Reply
	order.PutUint16(reply[2:4], r.sequence)
	order.PutUint32(reply[4:8], uint32(length/4))
	order.PutUint16(reply[8:10], uint16(thisClientLen))
	order.PutUint16(reply[10:12], uint16(allClientsLen))
	offset := 32
	for _, class := range r.ThisClientClasses {
		order.PutUint32(reply[offset:offset+4], class)
		offset += 4
	}
	for _, class := range r.AllClientsClasses {
		order.PutUint32(reply[offset:offset+4], class)
		offset += 4
	}
	return reply
}

// ChangeDeviceDontPropagateList request
type ChangeDeviceDontPropagateListRequest struct {
	Window  uint32
	Mode    byte
	Classes []uint32
}

func parseChangeDeviceDontPropagateListRequest(order binary.ByteOrder, body []byte, seq uint16) (*ChangeDeviceDontPropagateListRequest, error) {
	numClasses := order.Uint16(body[6:8])
	if len(body) != 8+int(numClasses)*4 {
		return nil, NewError(LengthErrorCode, seq, 0, xInputOpcode, XChangeDeviceDontPropagateList)
	}
	classes := make([]uint32, numClasses)
	for i := 0; i < int(numClasses); i++ {
		classes[i] = order.Uint32(body[8+i*4 : 12+i*4])
	}
	return &ChangeDeviceDontPropagateListRequest{
		Window:  order.Uint32(body[0:4]),
		Mode:    body[4],
		Classes: classes,
	}, nil
}

// GetDeviceDontPropagateList request
type GetDeviceDontPropagateListRequest struct {
	Window uint32
}

func parseGetDeviceDontPropagateListRequest(order binary.ByteOrder, body []byte, seq uint16) (*GetDeviceDontPropagateListRequest, error) {
	if len(body) != 4 {
		return nil, NewError(LengthErrorCode, seq, 0, xInputOpcode, XGetDeviceDontPropagateList)
	}
	return &GetDeviceDontPropagateListRequest{
		Window: order.Uint32(body[0:4]),
	}, nil
}

// GetDeviceDontPropagateList reply
type GetDeviceDontPropagateListReply struct {
	sequence uint16
	Classes  []uint32
}

func (r *GetDeviceDontPropagateListReply) encodeMessage(order binary.ByteOrder) []byte {
	length := len(r.Classes) * 4
	reply := make([]byte, 32+length)
	reply[0] = 1 // Reply
	order.PutUint16(reply[2:4], r.sequence)
	order.PutUint32(reply[4:8], uint32(length/4))
	order.PutUint16(reply[8:10], uint16(len(r.Classes)))
	offset := 32
	for _, class := range r.Classes {
		order.PutUint32(reply[offset:offset+4], class)
		offset += 4
	}
	return reply
}

// AllowDeviceEvents request
type AllowDeviceEventsRequest struct {
	Time     uint32
	DeviceID byte
	Mode     byte
}

func parseAllowDeviceEventsRequest(order binary.ByteOrder, body []byte, seq uint16) (*AllowDeviceEventsRequest, error) {
	if len(body) != 8 {
		return nil, NewError(LengthErrorCode, seq, 0, xInputOpcode, XAllowDeviceEvents)
	}
	return &AllowDeviceEventsRequest{
		Time:     order.Uint32(body[0:4]),
		DeviceID: body[4],
		Mode:     body[5],
	}, nil
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

// GrabDevice request
type GrabDeviceRequest struct {
	DeviceID        byte
	GrabWindow      uint32
	Time            uint32
	OwnerEvents     bool
	ThisDeviceMode  byte
	OtherDeviceMode byte
	NumClasses      uint16
	Classes         []uint32
}

func parseGrabDeviceRequest(order binary.ByteOrder, body []byte, seq uint16) (*GrabDeviceRequest, error) {
	if len(body) < 16 {
		return nil, NewError(LengthErrorCode, seq, 0, xInputOpcode, XGrabDevice)
	}
	numClasses := order.Uint16(body[14:16])
	if len(body) != 16+int(numClasses)*4 {
		return nil, NewError(LengthErrorCode, seq, 0, xInputOpcode, XGrabDevice)
	}
	classes := make([]uint32, numClasses)
	for i := 0; i < int(numClasses); i++ {
		classes[i] = order.Uint32(body[16+i*4 : 20+i*4])
	}
	return &GrabDeviceRequest{
		GrabWindow:      order.Uint32(body[0:4]),
		Time:            order.Uint32(body[4:8]),
		OwnerEvents:     body[10] != 0,
		ThisDeviceMode:  body[12],
		OtherDeviceMode: body[13],
		DeviceID:        body[11],
		NumClasses:      numClasses,
		Classes:         classes,
	}, nil
}

// GrabDevice reply
type GrabDeviceReply struct {
	sequence uint16
	Status   byte
}

func (r *GrabDeviceReply) encodeMessage(order binary.ByteOrder) []byte {
	reply := make([]byte, 32)
	reply[0] = 1 // Reply
	reply[1] = r.Status
	order.PutUint16(reply[2:4], r.sequence)
	order.PutUint32(reply[4:8], 0) // length
	return reply
}

// UngrabDevice request
type UngrabDeviceRequest struct {
	DeviceID byte
	Time     uint32
}

func parseUngrabDeviceRequest(order binary.ByteOrder, body []byte, seq uint16) (*UngrabDeviceRequest, error) {
	if len(body) != 8 {
		return nil, NewError(LengthErrorCode, seq, 0, xInputOpcode, XUngrabDevice)
	}
	return &UngrabDeviceRequest{
		Time:     order.Uint32(body[0:4]),
		DeviceID: body[5],
	}, nil
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
	Value      int32
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

func (s *x11Server) handleXInputRequest(client *x11Client, minorOpcode byte, body []byte, seq uint16) (reply messageEncoder) {
	switch minorOpcode {
	case XGetExtensionVersion:
		return &GetExtensionVersionReply{
			sequence:     seq,
			MajorVersion: 1,
			MinorVersion: 5,
		}
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

	case XSetDeviceMode:
		req, err := parseSetDeviceModeRequest(client.byteOrder, body, seq)
		if err != nil {
			return err.(messageEncoder)
		}
		device, ok := client.openDevices[req.DeviceID]
		if !ok {
			return NewError(ValueErrorCode, seq, uint32(req.DeviceID), xInputOpcode, XSetDeviceMode)
		}
		var valuatorInfo *ValuatorClassInfo
		for _, class := range device.classes {
			if vc, ok := class.(*ValuatorClassInfo); ok {
				valuatorInfo = vc
				break
			}
		}
		if valuatorInfo == nil {
			return NewError(MatchErrorCode, seq, 0, xInputOpcode, XSetDeviceMode)
		}
		valuatorInfo.Mode = req.Mode
		return &SetDeviceModeReply{sequence: seq, Status: GrabSuccess}
	case XSetDeviceValuators:
		req, err := parseSetDeviceValuatorsRequest(client.byteOrder, body, seq)
		if err != nil {
			return err.(messageEncoder)
		}
		device, ok := client.openDevices[req.DeviceID]
		if !ok {
			return NewError(ValueErrorCode, seq, uint32(req.DeviceID), xInputOpcode, XSetDeviceValuators)
		}
		var valuatorInfo *ValuatorClassInfo
		for _, class := range device.classes {
			if vc, ok := class.(*ValuatorClassInfo); ok {
				valuatorInfo = vc
				break
			}
		}
		if valuatorInfo == nil {
			return NewError(MatchErrorCode, seq, 0, xInputOpcode, XSetDeviceValuators)
		}
		if int(req.FirstValuator)+int(req.NumValuators) > len(valuatorInfo.Axes) {
			return NewError(ValueErrorCode, seq, 0, xInputOpcode, XSetDeviceValuators)
		}
		for i := 0; i < int(req.NumValuators); i++ {
			valuatorInfo.Axes[int(req.FirstValuator)+i].Value = req.Valuators[i]
		}
		return &SetDeviceValuatorsReply{sequence: seq, Status: GrabSuccess}
	case XGetDeviceControl:
		req, err := parseGetDeviceControlRequest(client.byteOrder, body, seq)
		if err != nil {
			return err.(messageEncoder)
		}
		device, ok := client.openDevices[req.DeviceID]
		if !ok {
			return NewError(ValueErrorCode, seq, uint32(req.DeviceID), xInputOpcode, XGetDeviceControl)
		}
		var valuatorInfo *ValuatorClassInfo
		for _, class := range device.classes {
			if vc, ok := class.(*ValuatorClassInfo); ok {
				valuatorInfo = vc
				break
			}
		}
		if valuatorInfo == nil {
			return NewError(MatchErrorCode, seq, 0, xInputOpcode, XGetDeviceControl)
		}
		resolutions := make([]uint32, len(valuatorInfo.Axes))
		minResolutions := make([]uint32, len(valuatorInfo.Axes))
		maxResolutions := make([]uint32, len(valuatorInfo.Axes))
		for i, axis := range valuatorInfo.Axes {
			resolutions[i] = axis.Resolution
			minResolutions[i] = 0
			maxResolutions[i] = 1000
		}
		return &GetDeviceControlReply{
			sequence: seq,
			Control: &DeviceResolutionState{
				NumValuators:   byte(len(valuatorInfo.Axes)),
				Resolutions:    resolutions,
				MinResolutions: minResolutions,
				MaxResolutions: maxResolutions,
			},
		}
	case XChangeDeviceControl:
		req, err := parseChangeDeviceControlRequest(client.byteOrder, body, seq)
		if err != nil {
			return err.(messageEncoder)
		}
		device, ok := client.openDevices[req.DeviceID]
		if !ok {
			return NewError(ValueErrorCode, seq, uint32(req.DeviceID), xInputOpcode, XChangeDeviceControl)
		}
		var valuatorInfo *ValuatorClassInfo
		for _, class := range device.classes {
			if vc, ok := class.(*ValuatorClassInfo); ok {
				valuatorInfo = vc
				break
			}
		}
		if valuatorInfo == nil {
			return NewError(MatchErrorCode, seq, 0, xInputOpcode, XChangeDeviceControl)
		}
		resolutionControl, ok := req.Control.(*DeviceResolutionControl)
		if !ok {
			return NewError(ValueErrorCode, seq, 0, xInputOpcode, XChangeDeviceControl)
		}
		if int(resolutionControl.FirstValuator)+int(resolutionControl.NumValuators) > len(valuatorInfo.Axes) {
			return NewError(ValueErrorCode, seq, 0, xInputOpcode, XChangeDeviceControl)
		}
		for i := 0; i < int(resolutionControl.NumValuators); i++ {
			valuatorInfo.Axes[int(resolutionControl.FirstValuator)+i].Resolution = resolutionControl.Resolutions[i]
		}
		return &ChangeDeviceControlReply{sequence: seq, Status: GrabSuccess}
	case XGetSelectedExtensionEvents:
		req, err := parseGetSelectedExtensionEventsRequest(client.byteOrder, body, seq)
		if err != nil {
			return err.(messageEncoder)
		}
		var thisClientClasses, allClientsClasses []uint32
		for _, dev := range client.openDevices {
			if mask, ok := dev.eventMasks[req.Window]; ok {
				class := (mask << 8) | uint32(dev.header.DeviceID)
				thisClientClasses = append(thisClientClasses, class)
			}
		}
		for _, c := range s.clients {
			for _, dev := range c.openDevices {
				if mask, ok := dev.eventMasks[req.Window]; ok {
					class := (mask << 8) | uint32(dev.header.DeviceID)
					allClientsClasses = append(allClientsClasses, class)
				}
			}
		}
		return &GetSelectedExtensionEventsReply{
			sequence:          seq,
			ThisClientClasses: thisClientClasses,
			AllClientsClasses: allClientsClasses,
		}
	case XChangeDeviceDontPropagateList:
		req, err := parseChangeDeviceDontPropagateListRequest(client.byteOrder, body, seq)
		if err != nil {
			return err.(messageEncoder)
		}
		win, ok := s.windows[client.xID(req.Window)]
		if !ok {
			return NewError(WindowErrorCode, seq, req.Window, xInputOpcode, XChangeDeviceDontPropagateList)
		}
		if win.dontPropagateDeviceEvents == nil {
			win.dontPropagateDeviceEvents = make(map[uint32]bool)
		}
		for _, class := range req.Classes {
			if req.Mode == 0 { // AddToList
				win.dontPropagateDeviceEvents[class] = true
			} else { // DeleteFromList
				delete(win.dontPropagateDeviceEvents, class)
			}
		}
		return nil
	case XAllowDeviceEvents:
		req, err := parseAllowDeviceEventsRequest(client.byteOrder, body, seq)
		if err != nil {
			return err.(messageEncoder)
		}
		s.frontend.AllowEvents(client.id, req.Mode, req.Time)
		return nil
	case XChangeKeyboardDevice:
		return NewError(DeviceErrorCode, seq, 0, xInputOpcode, XChangeKeyboardDevice)
	case XChangePointerDevice:
		return NewError(DeviceErrorCode, seq, 0, xInputOpcode, XChangePointerDevice)
	case XGetDeviceDontPropagateList:
		req, err := parseGetDeviceDontPropagateListRequest(client.byteOrder, body, seq)
		if err != nil {
			return err.(messageEncoder)
		}
		win, ok := s.windows[client.xID(req.Window)]
		if !ok {
			return NewError(WindowErrorCode, seq, req.Window, xInputOpcode, XGetDeviceDontPropagateList)
		}
		classes := make([]uint32, 0, len(win.dontPropagateDeviceEvents))
		for class := range win.dontPropagateDeviceEvents {
			classes = append(classes, class)
		}
		return &GetDeviceDontPropagateListReply{
			sequence: seq,
			Classes:  classes,
		}
	case XSendExtensionEvent:
		dest := client.byteOrder.Uint32(body[0:4])
		numClasses := client.byteOrder.Uint16(body[8:10])
		numEvents := body[10]

		if len(body) < 12+int(numEvents)*32+int(numClasses)*4 {
			return NewError(LengthErrorCode, seq, 0, xInputOpcode, XSendExtensionEvent)
		}

		eventBytes := body[12 : 12+int(numEvents)*32]
		classesBytes := body[12+int(numEvents)*32:]

		// Assuming a 1-to-1 mapping between events and classes
		for i := 0; i < int(numEvents); i++ {
			eventData := eventBytes[i*32 : (i+1)*32]
			class := client.byteOrder.Uint32(classesBytes[i*4 : (i+1)*4])

			eventMask := class >> 8
			deviceID := byte(class & 0xFF)

			for _, c := range s.clients {
				if dev, ok := c.openDevices[deviceID]; ok {
					if mask, ok := dev.eventMasks[dest]; ok {
						if (mask & eventMask) != 0 {
							// The client has selected for this event.
							// Send the raw event, but update the sequence number.
							c.byteOrder.PutUint16(eventData[2:4], c.sequence-1)
							rawEvent := &x11RawEvent{data: eventData}
							c.send(rawEvent)
						}
					}
				}
			}
		}
		return nil
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
	case XGrabDevice:
		req, err := parseGrabDeviceRequest(client.byteOrder, body, seq)
		if err != nil {
			return err.(messageEncoder)
		}
		if _, ok := s.deviceGrabs[req.DeviceID]; ok {
			return &GrabDeviceReply{sequence: seq, Status: AlreadyGrabbed}
		}
		grab := &deviceGrab{
			window:      client.xID(req.GrabWindow),
			ownerEvents: req.OwnerEvents,
			eventMask:   req.Classes,
			time:        req.Time,
		}
		s.deviceGrabs[req.DeviceID] = grab
		return &GrabDeviceReply{sequence: seq, Status: GrabSuccess}
	case XUngrabDevice:
		req, err := parseUngrabDeviceRequest(client.byteOrder, body, seq)
		if err != nil {
			return err.(messageEncoder)
		}
		if grab, ok := s.deviceGrabs[req.DeviceID]; ok {
			// In a real server, we'd check the grabbing client ID.
			// Here we assume any client can ungrab.
			if grab.window.client == client.id {
				delete(s.deviceGrabs, req.DeviceID)
			}
		}
		return nil
	default:
		// TODO: Implement other XInput requests
		return nil
	}
}
