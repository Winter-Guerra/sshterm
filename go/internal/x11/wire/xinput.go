//go:build x11

package wire

import (
	"bytes"
	"encoding/binary"
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

func (r *XInputRequest) OpCode() ReqCode {
	return XInputOpcode
}

func ParseXInputRequest(order binary.ByteOrder, data byte, body []byte, seq uint16) (*XInputRequest, error) {
	return &XInputRequest{
		MinorOpcode: data,
		Body:        body,
	}, nil
}

// GetExtensionVersion reply
type GetExtensionVersionReply struct {
	Sequence     uint16
	MajorVersion uint16
	MinorVersion uint16
}

func (r *GetExtensionVersionReply) EncodeMessage(order binary.ByteOrder) []byte {
	reply := make([]byte, 32)
	reply[0] = 1 // Reply
	reply[1] = 1 // Present
	order.PutUint16(reply[2:4], r.Sequence)
	order.PutUint32(reply[4:8], 0) // length
	order.PutUint16(reply[8:10], r.MajorVersion)
	order.PutUint16(reply[10:12], r.MinorVersion)
	return reply
}

// ListInputDevices request
type ListInputDevicesRequest struct{}

func ParseListInputDevicesRequest(order binary.ByteOrder, body []byte, seq uint16) (*ListInputDevicesRequest, error) {
	if len(body) != 0 {
		// As per spec, the request has no body.
		return nil, NewError(LengthErrorCode, seq, 0, byte(XInputOpcode), XListInputDevices)
	}
	return &ListInputDevicesRequest{}, nil
}

// OpenDevice request
type OpenDeviceRequest struct {
	DeviceID byte
}

func ParseOpenDeviceRequest(order binary.ByteOrder, body []byte, seq uint16) (*OpenDeviceRequest, error) {
	if len(body) != 4 {
		return nil, NewError(LengthErrorCode, seq, 0, byte(XInputOpcode), XOpenDevice)
	}
	return &OpenDeviceRequest{
		DeviceID: body[0],
	}, nil
}

// OpenDevice reply
type OpenDeviceReply struct {
	Sequence uint16
	Classes  []InputClassInfo
}

func (r *OpenDeviceReply) EncodeMessage(order binary.ByteOrder) []byte {
	classesBuf := new(bytes.Buffer)
	for _, class := range r.Classes {
		classesBuf.Write(class.EncodeMessage(order))
	}
	classesBytes := classesBuf.Bytes()

	reply := make([]byte, 32+len(classesBytes))
	reply[0] = 1 // Reply
	reply[1] = byte(len(r.Classes))
	order.PutUint16(reply[2:4], r.Sequence)
	order.PutUint32(reply[4:8], uint32((len(classesBytes)+3)/4)) // length
	copy(reply[32:], classesBytes)
	return reply
}

// SetDeviceMode request
type SetDeviceModeRequest struct {
	DeviceID byte
	Mode     byte
}

func ParseSetDeviceModeRequest(order binary.ByteOrder, body []byte, seq uint16) (*SetDeviceModeRequest, error) {
	if len(body) != 4 {
		return nil, NewError(LengthErrorCode, seq, 0, byte(XInputOpcode), XSetDeviceMode)
	}
	return &SetDeviceModeRequest{
		DeviceID: body[0],
		Mode:     body[1],
	}, nil
}

// SetDeviceMode reply
type SetDeviceModeReply struct {
	Sequence uint16
	Status   byte
}

func (r *SetDeviceModeReply) EncodeMessage(order binary.ByteOrder) []byte {
	reply := make([]byte, 32)
	reply[0] = 1 // Reply
	reply[1] = r.Status
	order.PutUint16(reply[2:4], r.Sequence)
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

func ParseSetDeviceValuatorsRequest(order binary.ByteOrder, body []byte, seq uint16) (*SetDeviceValuatorsRequest, error) {
	numValuators := body[2]
	if len(body) != 4+int(numValuators)*4 {
		return nil, NewError(LengthErrorCode, seq, 0, byte(XInputOpcode), XSetDeviceValuators)
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
	Sequence uint16
	Status   byte
}

func (r *SetDeviceValuatorsReply) EncodeMessage(order binary.ByteOrder) []byte {
	reply := make([]byte, 32)
	reply[0] = 1 // Reply
	reply[1] = r.Status
	order.PutUint16(reply[2:4], r.Sequence)
	order.PutUint32(reply[4:8], 0) // length
	return reply
}

// Device Control constants
const (
	DeviceResolution = 1
)

// DeviceControl interfaces
type DeviceControlState interface {
	EncodeMessage(order binary.ByteOrder) []byte
}

type DeviceControl interface {
	EncodeMessage(order binary.ByteOrder) []byte
}

type DeviceResolutionState struct {
	NumValuators   byte
	Resolutions    []uint32
	MinResolutions []uint32
	MaxResolutions []uint32
}

func (s *DeviceResolutionState) EncodeMessage(order binary.ByteOrder) []byte {
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

func (c *DeviceResolutionControl) EncodeMessage(order binary.ByteOrder) []byte {
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

func ParseGetDeviceControlRequest(order binary.ByteOrder, body []byte, seq uint16) (*GetDeviceControlRequest, error) {
	if len(body) != 4 {
		return nil, NewError(LengthErrorCode, seq, 0, byte(XInputOpcode), XGetDeviceControl)
	}
	return &GetDeviceControlRequest{
		DeviceID: body[0],
		Control:  order.Uint16(body[2:4]),
	}, nil
}

// GetDeviceControl reply
type GetDeviceControlReply struct {
	Sequence uint16
	Control  DeviceControlState
}

func (r *GetDeviceControlReply) EncodeMessage(order binary.ByteOrder) []byte {
	controlBytes := r.Control.EncodeMessage(order)
	reply := make([]byte, 32+len(controlBytes))
	reply[0] = 1 // Reply
	order.PutUint16(reply[2:4], r.Sequence)
	order.PutUint32(reply[4:8], uint32((len(controlBytes)+3)/4)) // length
	copy(reply[32:], controlBytes)
	return reply
}

// ChangeDeviceControl request
type ChangeDeviceControlRequest struct {
	DeviceID byte
	Control  DeviceControl
}

func ParseChangeDeviceControlRequest(order binary.ByteOrder, body []byte, seq uint16) (*ChangeDeviceControlRequest, error) {
	controlID := order.Uint16(body[2:4])
	if controlID != DeviceResolution {
		return nil, NewError(ValueErrorCode, seq, 0, byte(XInputOpcode), XChangeDeviceControl)
	}
	length := order.Uint16(body[4:6])
	numValuators := body[6]
	if length != 8+uint16(numValuators)*4 {
		return nil, NewError(LengthErrorCode, seq, 0, byte(XInputOpcode), XChangeDeviceControl)
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
	Sequence uint16
	Status   byte
}

func (r *ChangeDeviceControlReply) EncodeMessage(order binary.ByteOrder) []byte {
	reply := make([]byte, 32)
	reply[0] = 1 // Reply
	reply[1] = r.Status
	order.PutUint16(reply[2:4], r.Sequence)
	order.PutUint32(reply[4:8], 0)
	return reply
}

// GetSelectedExtensionEvents request
type GetSelectedExtensionEventsRequest struct {
	Window uint32
}

func ParseGetSelectedExtensionEventsRequest(order binary.ByteOrder, body []byte, seq uint16) (*GetSelectedExtensionEventsRequest, error) {
	if len(body) != 4 {
		return nil, NewError(LengthErrorCode, seq, 0, byte(XInputOpcode), XGetSelectedExtensionEvents)
	}
	return &GetSelectedExtensionEventsRequest{
		Window: order.Uint32(body[0:4]),
	}, nil
}

// GetSelectedExtensionEvents reply
type GetSelectedExtensionEventsReply struct {
	Sequence          uint16
	ThisClientClasses []uint32
	AllClientsClasses []uint32
}

func (r *GetSelectedExtensionEventsReply) EncodeMessage(order binary.ByteOrder) []byte {
	thisClientLen := len(r.ThisClientClasses)
	allClientsLen := len(r.AllClientsClasses)
	length := (thisClientLen + allClientsLen) * 4
	reply := make([]byte, 32+length)
	reply[0] = 1 // Reply
	order.PutUint16(reply[2:4], r.Sequence)
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

func ParseChangeDeviceDontPropagateListRequest(order binary.ByteOrder, body []byte, seq uint16) (*ChangeDeviceDontPropagateListRequest, error) {
	numClasses := order.Uint16(body[6:8])
	if len(body) != 8+int(numClasses)*4 {
		return nil, NewError(LengthErrorCode, seq, 0, byte(XInputOpcode), XChangeDeviceDontPropagateList)
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

func ParseGetDeviceDontPropagateListRequest(order binary.ByteOrder, body []byte, seq uint16) (*GetDeviceDontPropagateListRequest, error) {
	if len(body) != 4 {
		return nil, NewError(LengthErrorCode, seq, 0, byte(XInputOpcode), XGetDeviceDontPropagateList)
	}
	return &GetDeviceDontPropagateListRequest{
		Window: order.Uint32(body[0:4]),
	}, nil
}

// GetDeviceDontPropagateList reply
type GetDeviceDontPropagateListReply struct {
	Sequence uint16
	Classes  []uint32
}

func (r *GetDeviceDontPropagateListReply) EncodeMessage(order binary.ByteOrder) []byte {
	length := len(r.Classes) * 4
	reply := make([]byte, 32+length)
	reply[0] = 1 // Reply
	order.PutUint16(reply[2:4], r.Sequence)
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

func ParseAllowDeviceEventsRequest(order binary.ByteOrder, body []byte, seq uint16) (*AllowDeviceEventsRequest, error) {
	if len(body) != 8 {
		return nil, NewError(LengthErrorCode, seq, 0, byte(XInputOpcode), XAllowDeviceEvents)
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

func ParseCloseDeviceRequest(order binary.ByteOrder, body []byte, seq uint16) (*CloseDeviceRequest, error) {
	if len(body) != 4 {
		return nil, NewError(LengthErrorCode, seq, 0, byte(XInputOpcode), XCloseDevice)
	}
	return &CloseDeviceRequest{
		DeviceID: body[0],
	}, nil
}

// CloseDevice reply
type CloseDeviceReply struct {
	Sequence uint16
}

func (r *CloseDeviceReply) EncodeMessage(order binary.ByteOrder) []byte {
	reply := make([]byte, 32)
	reply[0] = 1 // Reply
	reply[1] = 0
	order.PutUint16(reply[2:4], r.Sequence)
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

func ParseGrabDeviceRequest(order binary.ByteOrder, body []byte, seq uint16) (*GrabDeviceRequest, error) {
	if len(body) < 16 {
		return nil, NewError(LengthErrorCode, seq, 0, byte(XInputOpcode), XGrabDevice)
	}
	numClasses := order.Uint16(body[14:16])
	if len(body) != 16+int(numClasses)*4 {
		return nil, NewError(LengthErrorCode, seq, 0, byte(XInputOpcode), XGrabDevice)
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
	Sequence uint16
	Status   byte
}

func (r *GrabDeviceReply) EncodeMessage(order binary.ByteOrder) []byte {
	reply := make([]byte, 32)
	reply[0] = 1 // Reply
	reply[1] = r.Status
	order.PutUint16(reply[2:4], r.Sequence)
	order.PutUint32(reply[4:8], 0) // length
	return reply
}

// UngrabDevice request
type UngrabDeviceRequest struct {
	DeviceID byte
	Time     uint32
}

func ParseUngrabDeviceRequest(order binary.ByteOrder, body []byte, seq uint16) (*UngrabDeviceRequest, error) {
	if len(body) != 8 {
		return nil, NewError(LengthErrorCode, seq, 0, byte(XInputOpcode), XUngrabDevice)
	}
	return &UngrabDeviceRequest{
		Time:     order.Uint32(body[0:4]),
		DeviceID: body[5],
	}, nil
}

// ListInputDevices reply
// This is a complex, variable-length reply.

type DeviceInfo struct {
	Header     DeviceHeader
	Classes    []InputClassInfo
	EventMasks map[uint32]uint32 // window ID -> event mask
}

func (d DeviceInfo) EncodeMessage(order binary.ByteOrder) []byte {
	// TODO: Implement this
	return nil
}

type ListInputDevicesReply struct {
	Sequence uint16
	Devices  []*DeviceInfo
	NDevices byte
}

type DeviceHeader struct {
	DeviceID   byte
	DeviceType Atom
	NumClasses byte
	Use        byte // 0: IsXPointer, 1: IsXKeyboard, 2: IsXExtensionDevice
	Name       string
}

type InputClassInfo interface {
	EncodeMessage(order binary.ByteOrder) []byte
	ClassID() byte
	Length() int
}

type KeyClassInfo struct {
	NumKeys    uint16
	MinKeycode byte
	MaxKeycode byte
}

func (c *KeyClassInfo) ClassID() byte { return 0 } // KeyClass
func (c *KeyClassInfo) Length() int   { return 8 }
func (c *KeyClassInfo) EncodeMessage(order binary.ByteOrder) []byte {
	buf := new(bytes.Buffer)
	buf.WriteByte(c.ClassID())
	buf.WriteByte(byte(c.Length()))
	buf.WriteByte(c.MinKeycode)
	buf.WriteByte(c.MaxKeycode)
	binary.Write(buf, order, c.NumKeys)
	buf.Write([]byte{0, 0}) // padding
	return buf.Bytes()
}

type ButtonClassInfo struct {
	NumButtons uint16
}

func (c *ButtonClassInfo) ClassID() byte { return 1 } // ButtonClass
func (c *ButtonClassInfo) Length() int   { return 8 }
func (c *ButtonClassInfo) EncodeMessage(order binary.ByteOrder) []byte {
	buf := new(bytes.Buffer)
	buf.WriteByte(c.ClassID())
	buf.WriteByte(byte(c.Length()))
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

func (c *ValuatorClassInfo) ClassID() byte { return 2 } // ValuatorClass
func (c *ValuatorClassInfo) Length() int {
	return 8 + len(c.Axes)*12
}
func (c *ValuatorClassInfo) EncodeMessage(order binary.ByteOrder) []byte {
	buf := new(bytes.Buffer)
	buf.WriteByte(c.ClassID())
	buf.WriteByte(byte(c.Length()))
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

func (r *ListInputDevicesReply) EncodeMessage(order binary.ByteOrder) []byte {
	var devicesData []byte
	for _, dev := range r.Devices {
		devicesData = append(devicesData, dev.EncodeMessage(order)...)
	}

	p := (4 - (len(devicesData) % 4)) % 4
	if p == 4 {
		p = 0
	}
	finalDeviceData := make([]byte, len(devicesData)+p)
	copy(finalDeviceData, devicesData)

	reply := make([]byte, 32+len(finalDeviceData))
	reply[0] = 1 // Reply
	reply[1] = byte(len(r.Devices))
	order.PutUint16(reply[2:4], r.Sequence)
	order.PutUint32(reply[4:8], uint32((len(finalDeviceData)+3)/4)) // length
	copy(reply[32:], finalDeviceData)
	return reply
}
