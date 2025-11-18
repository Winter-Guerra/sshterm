//go:build x11

package wire

import (
	"bytes"
	"encoding/binary"
)

// XInput minor opcodes from XIproto.h
const (
	XGetExtensionVersion           = 1
	XListInputDevices              = 2
	XOpenDevice                    = 3
	XCloseDevice                   = 4
	XSetDeviceMode                 = 5
	XSelectExtensionEvent          = 6
	XGetSelectedExtensionEvents    = 7
	XChangeDeviceDontPropagateList = 8
	XGetDeviceDontPropagateList    = 9
	XGetDeviceMotionEvents         = 10
	XChangeKeyboardDevice          = 11
	XChangePointerDevice           = 12
	XGrabDevice                    = 13
	XUngrabDevice                  = 14
	XGrabDeviceKey                 = 15
	XUngrabDeviceKey               = 16
	XGrabDeviceButton              = 17
	XUngrabDeviceButton            = 18
	XAllowDeviceEvents             = 19
	XGetDeviceFocus                = 20
	XSetDeviceFocus                = 21
	XGetFeedbackControl            = 22
	XChangeFeedbackControl         = 23
	XGetDeviceKeyMapping           = 24
	XChangeDeviceKeyMapping        = 25
	XGetDeviceModifierMapping      = 26
	XSetDeviceModifierMapping      = 27
	XGetDeviceButtonMapping        = 28
	XSetDeviceButtonMapping        = 29
	XQueryDeviceState              = 30
	XSendExtensionEvent            = 31
	XDeviceBell                    = 32
	XSetDeviceValuators            = 33
	XGetDeviceControl              = 34
	XChangeDeviceControl           = 35
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
	if len(body) < 4 {
		return nil, NewError(LengthErrorCode, seq, 0, byte(XInputOpcode), XSetDeviceValuators)
	}
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
	if len(body) < 4 {
		return nil, NewError(LengthErrorCode, seq, 0, byte(XInputOpcode), XChangeDeviceControl)
	}
	controlID := order.Uint16(body[2:4])
	if controlID != DeviceResolution {
		return nil, NewError(ValueErrorCode, seq, 0, byte(XInputOpcode), XChangeDeviceControl)
	}
	if len(body) < 10 {
		return nil, NewError(LengthErrorCode, seq, 0, byte(XInputOpcode), XChangeDeviceControl)
	}
	length := order.Uint16(body[4:6])
	firstValuator := body[6]
	numValuators := body[7]
	expectedControlLength := uint16(8) + uint16(numValuators)*4
	if length != expectedControlLength {
		return nil, NewError(LengthErrorCode, seq, 0, byte(XInputOpcode), XChangeDeviceControl)
	}
	expectedBodyLength := 2 + int(length)
	if len(body) != expectedBodyLength {
		return nil, NewError(LengthErrorCode, seq, 0, byte(XInputOpcode), XChangeDeviceControl)
	}
	resolutions := make([]uint32, numValuators)
	for i := 0; i < int(numValuators); i++ {
		start := 10 + i*4
		resolutions[i] = order.Uint32(body[start : start+4])
	}
	return &ChangeDeviceControlRequest{
		DeviceID: body[0],
		Control: &DeviceResolutionControl{
			FirstValuator: firstValuator,
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
	if len(body) < 8 {
		return nil, NewError(LengthErrorCode, seq, 0, byte(XInputOpcode), XChangeDeviceDontPropagateList)
	}
	numClasses := order.Uint16(body[4:6])
	if len(body) != 8+int(numClasses)*4 {
		return nil, NewError(LengthErrorCode, seq, 0, byte(XInputOpcode), XChangeDeviceDontPropagateList)
	}
	classes := make([]uint32, numClasses)
	for i := 0; i < int(numClasses); i++ {
		classes[i] = order.Uint32(body[8+i*4 : 12+i*4])
	}
	return &ChangeDeviceDontPropagateListRequest{
		Window:  order.Uint32(body[0:4]),
		Mode:    body[6],
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
	numClasses := order.Uint16(body[12:14])
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
		DeviceID:        body[8],
		OwnerEvents:     body[9] != 0,
		ThisDeviceMode:  body[10],
		OtherDeviceMode: body[11],
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
		Time:     order.Uint32(body[4:8]),
		DeviceID: body[0],
	}, nil
}

// ListInputDevices reply
type DeviceInfo struct {
	Header     DeviceHeader
	Classes    []InputClassInfo
	EventMasks map[uint32]uint32 // window ID -> event mask
}

func (d DeviceInfo) EncodeMessage(order binary.ByteOrder) []byte {
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

func (c *KeyClassInfo) ClassID() byte { return 0 }
func (c *KeyClassInfo) Length() int   { return 8 }
func (c *KeyClassInfo) EncodeMessage(order binary.ByteOrder) []byte {
	buf := new(bytes.Buffer)
	buf.WriteByte(c.ClassID())
	buf.WriteByte(byte(c.Length()))
	binary.Write(buf, order, c.NumKeys)
	buf.WriteByte(c.MinKeycode)
	buf.WriteByte(c.MaxKeycode)
	buf.Write([]byte{0, 0}) // padding
	return buf.Bytes()
}

type ButtonClassInfo struct {
	NumButtons uint16
}

func (c *ButtonClassInfo) ClassID() byte { return 1 }
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

func (c *ValuatorClassInfo) ClassID() byte { return 2 }
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

//
// NEWLY IMPLEMENTED REQUESTS START HERE
//

// SelectExtensionEvent request
type SelectExtensionEventRequest struct {
	Window  Window
	Classes []uint32
}

func ParseSelectExtensionEventRequest(order binary.ByteOrder, body []byte, seq uint16) (*SelectExtensionEventRequest, error) {
	if len(body) < 8 {
		return nil, NewError(LengthErrorCode, seq, 0, byte(XInputOpcode), XSelectExtensionEvent)
	}
	numClasses := order.Uint16(body[4:6])
	if len(body) != 8+int(numClasses)*4 {
		return nil, NewError(LengthErrorCode, seq, 0, byte(XInputOpcode), XSelectExtensionEvent)
	}
	classes := make([]uint32, numClasses)
	for i := 0; i < int(numClasses); i++ {
		classes[i] = order.Uint32(body[8+i*4 : 12+i*4])
	}
	return &SelectExtensionEventRequest{
		Window:  Window(order.Uint32(body[0:4])),
		Classes: classes,
	}, nil
}

// GetDeviceMotionEvents request
type GetDeviceMotionEventsRequest struct {
	Start    uint32
	Stop     uint32
	DeviceID byte
}

func ParseGetDeviceMotionEventsRequest(order binary.ByteOrder, body []byte, seq uint16) (*GetDeviceMotionEventsRequest, error) {
	if len(body) != 12 {
		return nil, NewError(LengthErrorCode, seq, 0, byte(XInputOpcode), XGetDeviceMotionEvents)
	}
	return &GetDeviceMotionEventsRequest{
		Start:    order.Uint32(body[0:4]),
		Stop:     order.Uint32(body[4:8]),
		DeviceID: body[8],
	}, nil
}

// ChangeKeyboardDevice request
type ChangeKeyboardDeviceRequest struct {
	DeviceID byte
}

func ParseChangeKeyboardDeviceRequest(order binary.ByteOrder, body []byte, seq uint16) (*ChangeKeyboardDeviceRequest, error) {
	if len(body) != 4 {
		return nil, NewError(LengthErrorCode, seq, 0, byte(XInputOpcode), XChangeKeyboardDevice)
	}
	return &ChangeKeyboardDeviceRequest{
		DeviceID: body[0],
	}, nil
}

// ChangePointerDevice request
type ChangePointerDeviceRequest struct {
	XAxis    byte
	YAxis    byte
	DeviceID byte
}

func ParseChangePointerDeviceRequest(order binary.ByteOrder, body []byte, seq uint16) (*ChangePointerDeviceRequest, error) {
	if len(body) != 4 {
		return nil, NewError(LengthErrorCode, seq, 0, byte(XInputOpcode), XChangePointerDevice)
	}
	return &ChangePointerDeviceRequest{
		XAxis:    body[0],
		YAxis:    body[1],
		DeviceID: body[2],
	}, nil
}

// GrabDeviceKey request
type GrabDeviceKeyRequest struct {
	GrabWindow      Window
	Modifiers       uint16
	Key             byte
	DeviceID        byte
	OwnerEvents     bool
	ThisDeviceMode  byte
	OtherDeviceMode byte
	NumClasses      uint16
	Classes         []uint32
}

func ParseGrabDeviceKeyRequest(order binary.ByteOrder, body []byte, seq uint16) (*GrabDeviceKeyRequest, error) {
	if len(body) < 12 {
		return nil, NewError(LengthErrorCode, seq, 0, byte(XInputOpcode), XGrabDeviceKey)
	}
	numClasses := order.Uint16(body[4:6])
	if len(body) != 12+int(numClasses)*4 {
		return nil, NewError(LengthErrorCode, seq, 0, byte(XInputOpcode), XGrabDeviceKey)
	}
	classes := make([]uint32, numClasses)
	for i := 0; i < int(numClasses); i++ {
		classes[i] = order.Uint32(body[12+i*4 : 16+i*4])
	}
	return &GrabDeviceKeyRequest{
		GrabWindow:      Window(order.Uint32(body[0:4])),
		NumClasses:      numClasses,
		OwnerEvents:     body[6] != 0,
		ThisDeviceMode:  body[7],
		OtherDeviceMode: body[8],
		DeviceID:        body[9],
		Key:             body[11],
		Modifiers:       order.Uint16(body[8:10]),
		Classes:         classes,
	}, nil
}

// UngrabDeviceKey request
type UngrabDeviceKeyRequest struct {
	GrabWindow Window
	Modifiers  uint16
	Key        byte
	DeviceID   byte
}

func ParseUngrabDeviceKeyRequest(order binary.ByteOrder, body []byte, seq uint16) (*UngrabDeviceKeyRequest, error) {
	if len(body) != 12 {
		return nil, NewError(LengthErrorCode, seq, 0, byte(XInputOpcode), XUngrabDeviceKey)
	}
	return &UngrabDeviceKeyRequest{
		GrabWindow: Window(order.Uint32(body[0:4])),
		Modifiers:  order.Uint16(body[4:6]),
		Key:        body[7],
		DeviceID:   body[6],
	}, nil
}

// GrabDeviceButton request
type GrabDeviceButtonRequest struct {
	GrabWindow      Window
	Modifiers       uint16
	Button          byte
	DeviceID        byte
	OwnerEvents     bool
	ThisDeviceMode  byte
	OtherDeviceMode byte
	NumClasses      uint16
	Classes         []uint32
}

func ParseGrabDeviceButtonRequest(order binary.ByteOrder, body []byte, seq uint16) (*GrabDeviceButtonRequest, error) {
	if len(body) < 12 {
		return nil, NewError(LengthErrorCode, seq, 0, byte(XInputOpcode), XGrabDeviceButton)
	}
	numClasses := order.Uint16(body[4:6])
	expectedLen := 12 + int(numClasses)*4
	if len(body) != expectedLen {
		return nil, NewError(LengthErrorCode, seq, 0, byte(XInputOpcode), XGrabDeviceButton)
	}
	classes := make([]uint32, numClasses)
	for i := 0; i < int(numClasses); i++ {
		classes[i] = order.Uint32(body[12+i*4 : 16+i*4])
	}
	return &GrabDeviceButtonRequest{
		GrabWindow:      Window(order.Uint32(body[0:4])),
		NumClasses:      numClasses,
		OwnerEvents:     body[6] != 0,
		ThisDeviceMode:  body[7],
		OtherDeviceMode: body[8],
		DeviceID:        body[9],
		Button:          body[11],
		Modifiers:       order.Uint16(body[8:10]),
		Classes:         classes,
	}, nil
}

// UngrabDeviceButton request
type UngrabDeviceButtonRequest struct {
	GrabWindow Window
	Modifiers  uint16
	Button     byte
	DeviceID   byte
}

func ParseUngrabDeviceButtonRequest(order binary.ByteOrder, body []byte, seq uint16) (*UngrabDeviceButtonRequest, error) {
	if len(body) != 12 {
		return nil, NewError(LengthErrorCode, seq, 0, byte(XInputOpcode), XUngrabDeviceButton)
	}
	return &UngrabDeviceButtonRequest{
		GrabWindow: Window(order.Uint32(body[0:4])),
		Modifiers:  order.Uint16(body[4:6]),
		Button:     body[7],
		DeviceID:   body[6],
	}, nil
}

// GetDeviceFocus request
type GetDeviceFocusRequest struct {
	DeviceID byte
}

func ParseGetDeviceFocusRequest(order binary.ByteOrder, body []byte, seq uint16) (*GetDeviceFocusRequest, error) {
	if len(body) != 4 {
		return nil, NewError(LengthErrorCode, seq, 0, byte(XInputOpcode), XGetDeviceFocus)
	}
	return &GetDeviceFocusRequest{
		DeviceID: body[0],
	}, nil
}

// SetDeviceFocus request
type SetDeviceFocusRequest struct {
	Focus    Window
	Time     uint32
	RevertTo byte
	DeviceID byte
}

func ParseSetDeviceFocusRequest(order binary.ByteOrder, body []byte, seq uint16) (*SetDeviceFocusRequest, error) {
	if len(body) != 12 {
		return nil, NewError(LengthErrorCode, seq, 0, byte(XInputOpcode), XSetDeviceFocus)
	}
	return &SetDeviceFocusRequest{
		Focus:    Window(order.Uint32(body[0:4])),
		Time:     order.Uint32(body[4:8]),
		RevertTo: body[8],
		DeviceID: body[9],
	}, nil
}

// GetFeedbackControl request
type GetFeedbackControlRequest struct {
	DeviceID byte
}

func ParseGetFeedbackControlRequest(order binary.ByteOrder, body []byte, seq uint16) (*GetFeedbackControlRequest, error) {
	if len(body) != 4 {
		return nil, NewError(LengthErrorCode, seq, 0, byte(XInputOpcode), XGetFeedbackControl)
	}
	return &GetFeedbackControlRequest{
		DeviceID: body[0],
	}, nil
}

// ChangeFeedbackControl request
type ChangeFeedbackControlRequest struct {
	Mask      uint32
	DeviceID  byte
	ControlID byte
	Control   []byte
}

func ParseChangeFeedbackControlRequest(order binary.ByteOrder, body []byte, seq uint16) (*ChangeFeedbackControlRequest, error) {
	if len(body) < 8 {
		return nil, NewError(LengthErrorCode, seq, 0, byte(XInputOpcode), XChangeFeedbackControl)
	}
	return &ChangeFeedbackControlRequest{
		Mask:      order.Uint32(body[0:4]),
		DeviceID:  body[4],
		ControlID: body[5],
		Control:   body[8:],
	}, nil
}

// GetDeviceKeyMapping request
type GetDeviceKeyMappingRequest struct {
	DeviceID   byte
	FirstKey   byte
	Count      byte
}

func ParseGetDeviceKeyMappingRequest(order binary.ByteOrder, body []byte, seq uint16) (*GetDeviceKeyMappingRequest, error) {
	if len(body) != 4 {
		return nil, NewError(LengthErrorCode, seq, 0, byte(XInputOpcode), XGetDeviceKeyMapping)
	}
	return &GetDeviceKeyMappingRequest{
		DeviceID:   body[0],
		FirstKey: body[1],
		Count:    body[2],
	}, nil
}

// ChangeDeviceKeyMapping request
type ChangeDeviceKeyMappingRequest struct {
	DeviceID         byte
	FirstKey         byte
	KeysymsPerKeycode byte
	KeycodeCount     byte
	Keysyms          []uint32
}

func ParseChangeDeviceKeyMappingRequest(order binary.ByteOrder, body []byte, seq uint16) (*ChangeDeviceKeyMappingRequest, error) {
	if len(body) < 4 {
		return nil, NewError(LengthErrorCode, seq, 0, byte(XInputOpcode), XChangeDeviceKeyMapping)
	}
	keycodeCount := body[3]
	keysymsPerKeycode := body[2]
	expectedLen := 4 + int(keycodeCount)*int(keysymsPerKeycode)*4
	if len(body) != expectedLen {
		return nil, NewError(LengthErrorCode, seq, 0, byte(XInputOpcode), XChangeDeviceKeyMapping)
	}
	keysyms := make([]uint32, int(keycodeCount)*int(keysymsPerKeycode))
	for i := range keysyms {
		keysyms[i] = order.Uint32(body[4+i*4 : 8+i*4])
	}
	return &ChangeDeviceKeyMappingRequest{
		DeviceID:         body[0],
		FirstKey:         body[1],
		KeysymsPerKeycode: keysymsPerKeycode,
		KeycodeCount:     keycodeCount,
		Keysyms:          keysyms,
	}, nil
}

// GetDeviceModifierMapping request
type GetDeviceModifierMappingRequest struct {
	DeviceID byte
}

func ParseGetDeviceModifierMappingRequest(order binary.ByteOrder, body []byte, seq uint16) (*GetDeviceModifierMappingRequest, error) {
	if len(body) != 4 {
		return nil, NewError(LengthErrorCode, seq, 0, byte(XInputOpcode), XGetDeviceModifierMapping)
	}
	return &GetDeviceModifierMappingRequest{
		DeviceID: body[0],
	}, nil
}

// SetDeviceModifierMapping request
type SetDeviceModifierMappingRequest struct {
	DeviceID byte
	Keycodes []byte
}

func ParseSetDeviceModifierMappingRequest(order binary.ByteOrder, body []byte, seq uint16) (*SetDeviceModifierMappingRequest, error) {
	if len(body) < 4 {
		return nil, NewError(LengthErrorCode, seq, 0, byte(XInputOpcode), XSetDeviceModifierMapping)
	}
	numKeycodesPerModifier := body[1]
	// There are always 8 modifiers.
	expectedLen := 4 + int(numKeycodesPerModifier)*8
	if len(body) != expectedLen {
		return nil, NewError(LengthErrorCode, seq, 0, byte(XInputOpcode), XSetDeviceModifierMapping)
	}
	return &SetDeviceModifierMappingRequest{
		DeviceID: body[0],
		Keycodes: body[4:],
	}, nil
}

// GetDeviceButtonMapping request
type GetDeviceButtonMappingRequest struct {
	DeviceID byte
}

func ParseGetDeviceButtonMappingRequest(order binary.ByteOrder, body []byte, seq uint16) (*GetDeviceButtonMappingRequest, error) {
	if len(body) != 4 {
		return nil, NewError(LengthErrorCode, seq, 0, byte(XInputOpcode), XGetDeviceButtonMapping)
	}
	return &GetDeviceButtonMappingRequest{
		DeviceID: body[0],
	}, nil
}

// SetDeviceButtonMapping request
type SetDeviceButtonMappingRequest struct {
	DeviceID byte
	Map      []byte
}

func ParseSetDeviceButtonMappingRequest(order binary.ByteOrder, body []byte, seq uint16) (*SetDeviceButtonMappingRequest, error) {
	if len(body) < 4 {
		return nil, NewError(LengthErrorCode, seq, 0, byte(XInputOpcode), XSetDeviceButtonMapping)
	}
	map_size := body[1]
	expectedLen := 4 + int(map_size)
	if len(body) != expectedLen {
		return nil, NewError(LengthErrorCode, seq, 0, byte(XInputOpcode), XSetDeviceButtonMapping)
	}
	return &SetDeviceButtonMappingRequest{
		DeviceID: body[0],
		Map:      body[4:],
	}, nil
}

// QueryDeviceState request
type QueryDeviceStateRequest struct {
	DeviceID byte
}

func ParseQueryDeviceStateRequest(order binary.ByteOrder, body []byte, seq uint16) (*QueryDeviceStateRequest, error) {
	if len(body) != 4 {
		return nil, NewError(LengthErrorCode, seq, 0, byte(XInputOpcode), XQueryDeviceState)
	}
	return &QueryDeviceStateRequest{
		DeviceID: body[0],
	}, nil
}

// SendExtensionEvent request
type SendExtensionEventRequest struct {
	Destination Window
	DeviceID    byte
	Propagate   bool
	NumClasses  uint16
	NumEvents   byte
	Events      []byte
	Classes     []uint32
}

func ParseSendExtensionEventRequest(order binary.ByteOrder, body []byte, seq uint16) (*SendExtensionEventRequest, error) {
	if len(body) < 8 {
		return nil, NewError(LengthErrorCode, seq, 0, byte(XInputOpcode), XSendExtensionEvent)
	}
	numEvents := body[6]
	numClasses := order.Uint16(body[4:6])
	expectedLen := 8 + int(numEvents)*32 + int(numClasses)*4
	if len(body) != expectedLen {
		return nil, NewError(LengthErrorCode, seq, 0, byte(XInputOpcode), XSendExtensionEvent)
	}
	events := body[8 : 8+int(numEvents)*32]
	classesStart := 8 + int(numEvents)*32
	classes := make([]uint32, numClasses)
	for i := 0; i < int(numClasses); i++ {
		classes[i] = order.Uint32(body[classesStart+i*4 : classesStart+(i+1)*4])
	}
	return &SendExtensionEventRequest{
		Destination: Window(order.Uint32(body[0:4])),
		DeviceID:    body[7],
		Propagate:   body[8] != 0,
		NumClasses:  numClasses,
		NumEvents:   numEvents,
		Events:      events,
		Classes:     classes,
	}, nil
}

// DeviceBell request
type DeviceBellRequest struct {
	DeviceID       byte
	FeedbackID     byte
	FeedbackClass  byte
	Percent        byte
}

func ParseDeviceBellRequest(order binary.ByteOrder, body []byte, seq uint16) (*DeviceBellRequest, error) {
	if len(body) != 4 {
		return nil, NewError(LengthErrorCode, seq, 0, byte(XInputOpcode), XDeviceBell)
	}
	return &DeviceBellRequest{
		DeviceID:       body[0],
		FeedbackID:     body[1],
		FeedbackClass:  body[2],
		Percent:        body[3],
	}, nil
}
