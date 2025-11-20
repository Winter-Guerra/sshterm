//go:build x11

package wire

import (
	"bytes"
	"encoding/binary"
)

// XInput minor opcodes from XIproto.h
const (
	CorePointerDeviceID = 2
	CoreKeyboardDeviceID = 3

	KeyClass    = 0
	ButtonClass = 1
	ValuatorClass = 2

	KbdFeedbackClass  = 0
	PtrFeedbackClass  = 1
	IntFeedbackClass  = 2
	StringFeedbackClass = 3
	BellFeedbackClass = 4
	LedFeedbackClass  = 5

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
	XIQueryPointer                 = 40
	XIWarpPointer                  = 41
	XIChangeCursor                 = 42
	XIChangeHierarchy              = 43
	XISetClientPointer             = 44
	XIGetClientPointer             = 45
	XISelectEvents                 = 46
	XIQueryVersion                 = 47
	XIQueryDevice                  = 48
	XISetFocus                     = 49
	XIGetFocus                     = 50
	XIGrabDevice                   = 51
	XIUngrabDevice                 = 52
	XIAllowEvents                  = 53
	XIPassiveGrabDevice            = 54
	XIPassiveUngrabDevice          = 55
	XIListProperties               = 56
	XIChangeProperty               = 57
	XIDeleteProperty               = 58
	XIGetProperty                  = 59
	XIGetSelectedEvents            = 60
	XIBarrierReleasePointer        = 61
)

// XInput request types
type XInputRequest struct {
	MinorOpcode byte
	Body        []byte
}

func (r *XInputRequest) OpCode() ReqCode {
	return XInputOpcode
}

func ParseXInputRequest(order binary.ByteOrder, data byte, body []byte, seq uint16) (Request, error) {
	switch data {
	case XGetExtensionVersion:
		return ParseGetExtensionVersionRequest(order, body, seq)
	case XListInputDevices:
		return ParseListInputDevicesRequest(order, body, seq)
	case XOpenDevice:
		return ParseOpenDeviceRequest(order, body, seq)
	case XCloseDevice:
		return ParseCloseDeviceRequest(order, body, seq)
	case XSetDeviceMode:
		return ParseSetDeviceModeRequest(order, body, seq)
	case XSelectExtensionEvent:
		return ParseSelectExtensionEventRequest(order, body, seq)
	case XGetSelectedExtensionEvents:
		return ParseGetSelectedExtensionEventsRequest(order, body, seq)
	case XChangeDeviceDontPropagateList:
		return ParseChangeDeviceDontPropagateListRequest(order, body, seq)
	case XGetDeviceDontPropagateList:
		return ParseGetDeviceDontPropagateListRequest(order, body, seq)
	case XGetDeviceMotionEvents:
		return ParseGetDeviceMotionEventsRequest(order, body, seq)
	case XChangeKeyboardDevice:
		return ParseChangeKeyboardDeviceRequest(order, body, seq)
	case XChangePointerDevice:
		return ParseChangePointerDeviceRequest(order, body, seq)
	case XGrabDevice:
		return ParseGrabDeviceRequest(order, body, seq)
	case XUngrabDevice:
		return ParseUngrabDeviceRequest(order, body, seq)
	case XGrabDeviceKey:
		return ParseGrabDeviceKeyRequest(order, body, seq)
	case XUngrabDeviceKey:
		return ParseUngrabDeviceKeyRequest(order, body, seq)
	case XGrabDeviceButton:
		return ParseGrabDeviceButtonRequest(order, body, seq)
	case XUngrabDeviceButton:
		return ParseUngrabDeviceButtonRequest(order, body, seq)
	case XAllowDeviceEvents:
		return ParseAllowDeviceEventsRequest(order, body, seq)
	case XGetDeviceFocus:
		return ParseGetDeviceFocusRequest(order, body, seq)
	case XSetDeviceFocus:
		return ParseSetDeviceFocusRequest(order, body, seq)
	case XGetFeedbackControl:
		return ParseGetFeedbackControlRequest(order, body, seq)
	case XChangeFeedbackControl:
		return ParseChangeFeedbackControlRequest(order, body, seq)
	case XGetDeviceKeyMapping:
		return ParseGetDeviceKeyMappingRequest(order, body, seq)
	case XChangeDeviceKeyMapping:
		return ParseChangeDeviceKeyMappingRequest(order, body, seq)
	case XGetDeviceModifierMapping:
		return ParseGetDeviceModifierMappingRequest(order, body, seq)
	case XSetDeviceModifierMapping:
		return ParseSetDeviceModifierMappingRequest(order, body, seq)
	case XGetDeviceButtonMapping:
		return ParseGetDeviceButtonMappingRequest(order, body, seq)
	case XSetDeviceButtonMapping:
		return ParseSetDeviceButtonMappingRequest(order, body, seq)
	case XQueryDeviceState:
		return ParseQueryDeviceStateRequest(order, body, seq)
	case XSendExtensionEvent:
		return ParseSendExtensionEventRequest(order, body, seq)
	case XDeviceBell:
		return ParseDeviceBellRequest(order, body, seq)
	case XSetDeviceValuators:
		return ParseSetDeviceValuatorsRequest(order, body, seq)
	case XGetDeviceControl:
		return ParseGetDeviceControlRequest(order, body, seq)
	case XChangeDeviceControl:
		return ParseChangeDeviceControlRequest(order, body, seq)
	case XIQueryVersion:
		return ParseXIQueryVersionRequest(order, body, seq)
	case XIWarpPointer:
		return ParseXIWarpPointerRequest(order, body, seq)
	case XIChangeCursor:
		return ParseXIChangeCursorRequest(order, body, seq)
	case XIChangeHierarchy:
		return ParseXIChangeHierarchyRequest(order, body, seq)
	case XISetClientPointer:
		return ParseXISetClientPointerRequest(order, body, seq)
	case XIGetClientPointer:
		return ParseXIGetClientPointerRequest(order, body, seq)
	case XISelectEvents:
		return ParseXISelectEventsRequest(order, body, seq)
	case XIQueryDevice:
		return ParseXIQueryDeviceRequest(order, body, seq)
	case XISetFocus:
		return ParseXISetFocusRequest(order, body, seq)
	case XIGetFocus:
		return ParseXIGetFocusRequest(order, body, seq)
	case XIGrabDevice:
		return ParseXIGrabDeviceRequest(order, body, seq)
	case XIUngrabDevice:
		return ParseXIUngrabDeviceRequest(order, body, seq)
	case XIAllowEvents:
		return ParseXIAllowEventsRequest(order, body, seq)
	case XIPassiveGrabDevice:
		return ParseXIPassiveGrabDeviceRequest(order, body, seq)
	case XIPassiveUngrabDevice:
		return ParseXIPassiveUngrabDeviceRequest(order, body, seq)
	case XIListProperties:
		return ParseXIListPropertiesRequest(order, body, seq)
	case XIChangeProperty:
		return ParseXIChangePropertyRequest(order, body, seq)
	case XIDeleteProperty:
		return ParseXIDeletePropertyRequest(order, body, seq)
	case XIGetProperty:
		return ParseXIGetPropertyRequest(order, body, seq)
	case XIGetSelectedEvents:
		return ParseXIGetSelectedEventsRequest(order, body, seq)
	case XIBarrierReleasePointer:
		return ParseXIBarrierReleasePointerRequest(order, body, seq)
	default:
		return nil, NewError(RequestErrorCode, seq, 0, byte(XInputOpcode), ReqCode(data))
	}
}

// GetExtensionVersion request
type GetExtensionVersionRequest struct {
	MajorVersion uint16
	MinorVersion uint16
}

func (r *GetExtensionVersionRequest) OpCode() ReqCode {
	return XInputOpcode
}

func ParseGetExtensionVersionRequest(order binary.ByteOrder, body []byte, seq uint16) (*GetExtensionVersionRequest, error) {
	if len(body) != 4 {
		return nil, NewError(LengthErrorCode, seq, 0, byte(XInputOpcode), XGetExtensionVersion)
	}
	return &GetExtensionVersionRequest{
		MajorVersion: order.Uint16(body[0:2]),
		MinorVersion: order.Uint16(body[2:4]),
	}, nil
}

// XIWarpPointer request
type XIWarpPointerRequest struct {
	DeviceID  uint16
	SrcWindow Window
	DstWindow Window
	SrcX      int32
	SrcY      int32
	SrcW      uint16
	SrcH      uint16
	DstX      int32
	DstY      int32
}

func (r *XIWarpPointerRequest) OpCode() ReqCode {
	return XInputOpcode
}

func ParseXIWarpPointerRequest(order binary.ByteOrder, body []byte, seq uint16) (*XIWarpPointerRequest, error) {
	if len(body) != 32 {
		return nil, NewError(LengthErrorCode, seq, 0, byte(XInputOpcode), XIWarpPointer)
	}
	// The protocol defines 'fp1616' fixed point numbers. We'll treat them as int32 for now,
	// effectively ignoring the fractional part by taking the integer part.
	return &XIWarpPointerRequest{
		DeviceID:  order.Uint16(body[0:2]),
		SrcWindow: Window(order.Uint32(body[4:8])),
		DstWindow: Window(order.Uint32(body[8:12])),
		SrcX:      int32(order.Uint32(body[12:16])),
		SrcY:      int32(order.Uint32(body[16:20])),
		SrcW:      order.Uint16(body[20:22]),
		SrcH:      order.Uint16(body[22:24]),
		DstX:      int32(order.Uint32(body[24:28])),
		DstY:      int32(order.Uint32(body[28:32])),
	}, nil
}

// XIChangeCursor request
type XIChangeCursorRequest struct {
	DeviceID uint16
	Window   Window
	Cursor   uint32
}

func (r *XIChangeCursorRequest) OpCode() ReqCode {
	return XInputOpcode
}

func ParseXIChangeCursorRequest(order binary.ByteOrder, body []byte, seq uint16) (*XIChangeCursorRequest, error) {
	if len(body) != 12 {
		return nil, NewError(LengthErrorCode, seq, 0, byte(XInputOpcode), XIChangeCursor)
	}
	return &XIChangeCursorRequest{
		DeviceID: order.Uint16(body[0:2]),
		Window:   Window(order.Uint32(body[4:8])),
		Cursor:   order.Uint32(body[8:12]),
	}, nil
}

// XIChangeHierarchy request
type XIChangeHierarchyRequest struct {
	NumChanges uint16
	Changes    []XIChangeHierarchyChange
}

type XIChangeHierarchyChange interface {
	Op() uint16
}

type XIAnyHierarchyChange struct {
	Type   uint16
	Length uint16
}

type XIAddMaster struct {
	Type     uint16
	Length   uint16
	Name     string
	SendCore bool
	Enable   bool
}

func (c *XIAddMaster) Op() uint16 { return 1 }

type XIRemoveMaster struct {
	Type           uint16
	Length         uint16
	DeviceID       uint16
	ReturnMode     byte
	ReturnPointer  uint16
	ReturnKeyboard uint16
}

func (c *XIRemoveMaster) Op() uint16 { return 2 }

type XIAttachSlave struct {
	Type     uint16
	Length   uint16
	DeviceID uint16
	MasterID uint16
}

func (c *XIAttachSlave) Op() uint16 { return 3 }

type XIDetachSlave struct {
	Type     uint16
	Length   uint16
	DeviceID uint16
}

func (c *XIDetachSlave) Op() uint16 { return 4 }

func (r *XIChangeHierarchyRequest) OpCode() ReqCode {
	return XInputOpcode
}

func ParseXIChangeHierarchyRequest(order binary.ByteOrder, body []byte, seq uint16) (*XIChangeHierarchyRequest, error) {
	if len(body) < 4 {
		return nil, NewError(LengthErrorCode, seq, 0, byte(XInputOpcode), XIChangeHierarchy)
	}
	numChanges := order.Uint16(body[0:2])
	changes := make([]XIChangeHierarchyChange, 0, numChanges)
	offset := 4
	for i := 0; i < int(numChanges); i++ {
		if len(body) < offset+4 {
			return nil, NewError(LengthErrorCode, seq, 0, byte(XInputOpcode), XIChangeHierarchy)
		}
		changeType := order.Uint16(body[offset : offset+2])
		length := order.Uint16(body[offset+2 : offset+4])
		if len(body) < offset+int(length) {
			return nil, NewError(LengthErrorCode, seq, 0, byte(XInputOpcode), XIChangeHierarchy)
		}
		changeBody := body[offset : offset+int(length)]
		var change XIChangeHierarchyChange
		switch changeType {
		case 1: // XIAddMaster
			return nil, NewError(ImplementationErrorCode, seq, 0, byte(XInputOpcode), XIChangeHierarchy)
		case 2: // XIRemoveMaster
			return nil, NewError(ImplementationErrorCode, seq, 0, byte(XInputOpcode), XIChangeHierarchy)
		case 3: // XIAttachSlave
			if len(changeBody) != 8 {
				return nil, NewError(LengthErrorCode, seq, 0, byte(XInputOpcode), XIChangeHierarchy)
			}
			change = &XIAttachSlave{
				Type:     changeType,
				Length:   length,
				DeviceID: order.Uint16(changeBody[4:6]),
				MasterID: order.Uint16(changeBody[6:8]),
			}
		case 4: // XIDetachSlave
			if len(changeBody) != 8 {
				return nil, NewError(LengthErrorCode, seq, 0, byte(XInputOpcode), XIChangeHierarchy)
			}
			change = &XIDetachSlave{
				Type:     changeType,
				Length:   length,
				DeviceID: order.Uint16(changeBody[4:6]),
			}
		default:
			return nil, NewError(ValueErrorCode, seq, 0, byte(XInputOpcode), XIChangeHierarchy)
		}
		changes = append(changes, change)
		offset += int(length)
	}
	return &XIChangeHierarchyRequest{
		NumChanges: numChanges,
		Changes:    changes,
	}, nil
}

// XISetClientPointer request
type XISetClientPointerRequest struct {
	DeviceID uint16
	Window   Window
}

func (r *XISetClientPointerRequest) OpCode() ReqCode {
	return XInputOpcode
}

func ParseXISetClientPointerRequest(order binary.ByteOrder, body []byte, seq uint16) (*XISetClientPointerRequest, error) {
	if len(body) != 8 {
		return nil, NewError(LengthErrorCode, seq, 0, byte(XInputOpcode), XISetClientPointer)
	}
	return &XISetClientPointerRequest{
		DeviceID: order.Uint16(body[0:2]),
		Window:   Window(order.Uint32(body[4:8])),
	}, nil
}

// XIGetClientPointer request
type XIGetClientPointerRequest struct {
	Window Window
}

func (r *XIGetClientPointerRequest) OpCode() ReqCode {
	return XInputOpcode
}

func ParseXIGetClientPointerRequest(order binary.ByteOrder, body []byte, seq uint16) (*XIGetClientPointerRequest, error) {
	if len(body) != 4 {
		return nil, NewError(LengthErrorCode, seq, 0, byte(XInputOpcode), XIGetClientPointer)
	}
	return &XIGetClientPointerRequest{
		Window: Window(order.Uint32(body[0:4])),
	}, nil
}

// XISelectEvents request
type XISelectEventsRequest struct {
	Window   Window
	NumMasks uint16
	Masks    []XIEventMask
}

type XIEventMask struct {
	DeviceID uint16
	MaskLen  uint16
	Mask     []byte
}

func (r *XISelectEventsRequest) OpCode() ReqCode {
	return XInputOpcode
}

func ParseXISelectEventsRequest(order binary.ByteOrder, body []byte, seq uint16) (*XISelectEventsRequest, error) {
	if len(body) < 8 {
		return nil, NewError(LengthErrorCode, seq, 0, byte(XInputOpcode), XISelectEvents)
	}
	window := Window(order.Uint32(body[0:4]))
	numMasks := order.Uint16(body[4:6])
	masks := make([]XIEventMask, 0, numMasks)
	offset := 8
	for i := 0; i < int(numMasks); i++ {
		if len(body) < offset+4 {
			return nil, NewError(LengthErrorCode, seq, 0, byte(XInputOpcode), XISelectEvents)
		}
		deviceID := order.Uint16(body[offset : offset+2])
		maskLen := order.Uint16(body[offset+2 : offset+4])
		maskBytes := int(maskLen) * 4
		if len(body) < offset+4+maskBytes {
			return nil, NewError(LengthErrorCode, seq, 0, byte(XInputOpcode), XISelectEvents)
		}
		mask := body[offset+4 : offset+4+maskBytes]
		masks = append(masks, XIEventMask{
			DeviceID: deviceID,
			MaskLen:  maskLen,
			Mask:     mask,
		})
		offset += 4 + maskBytes
	}
	return &XISelectEventsRequest{
		Window:   window,
		NumMasks: numMasks,
		Masks:    masks,
	}, nil
}

// XIQueryDevice request
type XIQueryDeviceRequest struct {
	DeviceID uint16
}

func (r *XIQueryDeviceRequest) OpCode() ReqCode {
	return XInputOpcode
}

func ParseXIQueryDeviceRequest(order binary.ByteOrder, body []byte, seq uint16) (*XIQueryDeviceRequest, error) {
	if len(body) != 4 {
		return nil, NewError(LengthErrorCode, seq, 0, byte(XInputOpcode), XIQueryDevice)
	}
	return &XIQueryDeviceRequest{
		DeviceID: order.Uint16(body[0:2]),
	}, nil
}

// XISetFocus request
type XISetFocusRequest struct {
	DeviceID uint16
	Focus    Window
	Time     uint32
}

func (r *XISetFocusRequest) OpCode() ReqCode {
	return XInputOpcode
}

func ParseXISetFocusRequest(order binary.ByteOrder, body []byte, seq uint16) (*XISetFocusRequest, error) {
	if len(body) != 12 {
		return nil, NewError(LengthErrorCode, seq, 0, byte(XInputOpcode), XISetFocus)
	}
	return &XISetFocusRequest{
		DeviceID: order.Uint16(body[0:2]),
		Focus:    Window(order.Uint32(body[4:8])),
		Time:     order.Uint32(body[8:12]),
	}, nil
}

// XIGetFocus request
type XIGetFocusRequest struct {
	DeviceID uint16
}

func (r *XIGetFocusRequest) OpCode() ReqCode {
	return XInputOpcode
}

func ParseXIGetFocusRequest(order binary.ByteOrder, body []byte, seq uint16) (*XIGetFocusRequest, error) {
	if len(body) != 4 {
		return nil, NewError(LengthErrorCode, seq, 0, byte(XInputOpcode), XIGetFocus)
	}
	return &XIGetFocusRequest{
		DeviceID: order.Uint16(body[0:2]),
	}, nil
}

// XIGrabDevice request
type XIGrabDeviceRequest struct {
	DeviceID         uint16
	GrabWindow       Window
	Time             uint32
	Cursor           uint32
	GrabMode         byte
	PairedDeviceMode byte
	OwnerEvents      bool
	MaskLen          uint16
	Mask             []byte
}

func (r *XIGrabDeviceRequest) OpCode() ReqCode {
	return XInputOpcode
}

func ParseXIGrabDeviceRequest(order binary.ByteOrder, body []byte, seq uint16) (*XIGrabDeviceRequest, error) {
	if len(body) < 20 {
		return nil, NewError(LengthErrorCode, seq, 0, byte(XInputOpcode), XIGrabDevice)
	}
	maskLen := order.Uint16(body[18:20])
	if len(body) != 20+int(maskLen)*4 {
		return nil, NewError(LengthErrorCode, seq, 0, byte(XInputOpcode), XIGrabDevice)
	}
	return &XIGrabDeviceRequest{
		DeviceID:         order.Uint16(body[0:2]),
		GrabWindow:       Window(order.Uint32(body[2:6])),
		Time:             order.Uint32(body[6:10]),
		Cursor:           order.Uint32(body[10:14]),
		GrabMode:         body[14],
		PairedDeviceMode: body[15],
		OwnerEvents:      body[16] != 0,
		MaskLen:          maskLen,
		Mask:             body[20:],
	}, nil
}

// XIUngrabDevice request
type XIUngrabDeviceRequest struct {
	DeviceID uint16
	Time     uint32
}

func (r *XIUngrabDeviceRequest) OpCode() ReqCode {
	return XInputOpcode
}

func ParseXIUngrabDeviceRequest(order binary.ByteOrder, body []byte, seq uint16) (*XIUngrabDeviceRequest, error) {
	if len(body) != 8 {
		return nil, NewError(LengthErrorCode, seq, 0, byte(XInputOpcode), XIUngrabDevice)
	}
	return &XIUngrabDeviceRequest{
		DeviceID: order.Uint16(body[0:2]),
		Time:     order.Uint32(body[4:8]),
	}, nil
}

// XIAllowEvents request
type XIAllowEventsRequest struct {
	DeviceID   uint16
	EventMode  byte
	Time       uint32
	TouchID    uint32
	GrabWindow Window
}

func (r *XIAllowEventsRequest) OpCode() ReqCode {
	return XInputOpcode
}

func ParseXIAllowEventsRequest(order binary.ByteOrder, body []byte, seq uint16) (*XIAllowEventsRequest, error) {
	if len(body) != 16 {
		return nil, NewError(LengthErrorCode, seq, 0, byte(XInputOpcode), XIAllowEvents)
	}
	return &XIAllowEventsRequest{
		DeviceID:   order.Uint16(body[0:2]),
		EventMode:  body[2],
		Time:       order.Uint32(body[4:8]),
		TouchID:    order.Uint32(body[8:12]),
		GrabWindow: Window(order.Uint32(body[12:16])),
	}, nil
}

// XIPassiveGrabDevice request
type XIPassiveGrabDeviceRequest struct {
	DeviceID         uint16
	GrabWindow       Window
	Time             uint32
	Cursor           uint32
	Detail           uint32
	NumModifiers     uint16
	GrabType         byte
	GrabMode         byte
	PairedDeviceMode byte
	OwnerEvents      bool
	Modifiers        []byte
}

func (r *XIPassiveGrabDeviceRequest) OpCode() ReqCode {
	return XInputOpcode
}

func ParseXIPassiveGrabDeviceRequest(order binary.ByteOrder, body []byte, seq uint16) (*XIPassiveGrabDeviceRequest, error) {
	if len(body) < 28 {
		return nil, NewError(LengthErrorCode, seq, 0, byte(XInputOpcode), XIPassiveGrabDevice)
	}
	numModifiers := order.Uint16(body[20:22])
	if len(body) != 28+int(numModifiers)*4 {
		return nil, NewError(LengthErrorCode, seq, 0, byte(XInputOpcode), XIPassiveGrabDevice)
	}
	return &XIPassiveGrabDeviceRequest{
		DeviceID:         order.Uint16(body[0:2]),
		GrabWindow:       Window(order.Uint32(body[4:8])),
		Time:             order.Uint32(body[8:12]),
		Cursor:           order.Uint32(body[12:16]),
		Detail:           order.Uint32(body[16:20]),
		NumModifiers:     numModifiers,
		GrabType:         body[24],
		GrabMode:         body[25],
		PairedDeviceMode: body[26],
		OwnerEvents:      body[27] != 0,
		Modifiers:        body[28:],
	}, nil
}

// XIPassiveUngrabDevice request
type XIPassiveUngrabDeviceRequest struct {
	DeviceID     uint16
	GrabWindow   Window
	Detail       uint32
	NumModifiers uint16
	GrabType     byte
	Modifiers    []byte
}

func (r *XIPassiveUngrabDeviceRequest) OpCode() ReqCode {
	return XInputOpcode
}

func ParseXIPassiveUngrabDeviceRequest(order binary.ByteOrder, body []byte, seq uint16) (*XIPassiveUngrabDeviceRequest, error) {
	if len(body) < 16 {
		return nil, NewError(LengthErrorCode, seq, 0, byte(XInputOpcode), XIPassiveUngrabDevice)
	}
	numModifiers := order.Uint16(body[12:14])
	if len(body) != 20+int(numModifiers)*4 {
		return nil, NewError(LengthErrorCode, seq, 0, byte(XInputOpcode), XIPassiveUngrabDevice)
	}
	return &XIPassiveUngrabDeviceRequest{
		DeviceID:     order.Uint16(body[0:2]),
		GrabWindow:   Window(order.Uint32(body[4:8])),
		Detail:       order.Uint32(body[8:12]),
		NumModifiers: numModifiers,
		GrabType:     body[16],
		Modifiers:    body[20:],
	}, nil
}

// XIListProperties request
type XIListPropertiesRequest struct {
	DeviceID uint16
}

func (r *XIListPropertiesRequest) OpCode() ReqCode {
	return XInputOpcode
}

func ParseXIListPropertiesRequest(order binary.ByteOrder, body []byte, seq uint16) (*XIListPropertiesRequest, error) {
	if len(body) != 4 {
		return nil, NewError(LengthErrorCode, seq, 0, byte(XInputOpcode), XIListProperties)
	}
	return &XIListPropertiesRequest{
		DeviceID: order.Uint16(body[0:2]),
	}, nil
}

// XIChangeProperty request
type XIChangePropertyRequest struct {
	DeviceID uint16
	Mode     byte
	Format   byte
	Property uint32
	Type     uint32
	NumItems uint32
	Data     []byte
}

func (r *XIChangePropertyRequest) OpCode() ReqCode {
	return XInputOpcode
}

func pad(i int) int {
	return (i + 3) & ^3
}

func ParseXIChangePropertyRequest(order binary.ByteOrder, body []byte, seq uint16) (*XIChangePropertyRequest, error) {
	if len(body) < 16 {
		return nil, NewError(LengthErrorCode, seq, 0, byte(XInputOpcode), XIChangeProperty)
	}
	format := body[3]
	numItems := order.Uint32(body[12:16])
	var dataLen int
	switch format {
	case 8:
		dataLen = int(numItems)
	case 16:
		dataLen = int(numItems) * 2
	case 32:
		dataLen = int(numItems) * 4
	default:
		return nil, NewError(ValueErrorCode, seq, 0, byte(XInputOpcode), XIChangeProperty)
	}
	if len(body) != 16+pad(dataLen) {
		return nil, NewError(LengthErrorCode, seq, 0, byte(XInputOpcode), XIChangeProperty)
	}
	return &XIChangePropertyRequest{
		DeviceID: order.Uint16(body[0:2]),
		Mode:     body[2],
		Format:   format,
		Property: order.Uint32(body[4:8]),
		Type:     order.Uint32(body[8:12]),
		NumItems: numItems,
		Data:     body[16 : 16+dataLen],
	}, nil
}

// XIDeleteProperty request
type XIDeletePropertyRequest struct {
	DeviceID uint16
	Property uint32
}

func (r *XIDeletePropertyRequest) OpCode() ReqCode {
	return XInputOpcode
}

func ParseXIDeletePropertyRequest(order binary.ByteOrder, body []byte, seq uint16) (*XIDeletePropertyRequest, error) {
	if len(body) != 8 {
		return nil, NewError(LengthErrorCode, seq, 0, byte(XInputOpcode), XIDeleteProperty)
	}
	return &XIDeletePropertyRequest{
		DeviceID: order.Uint16(body[0:2]),
		Property: order.Uint32(body[4:8]),
	}, nil
}

// XIGetProperty request
type XIGetPropertyRequest struct {
	DeviceID uint16
	Delete   bool
	Property uint32
	Type     uint32
	Offset   uint32
	Len      uint32
}

func (r *XIGetPropertyRequest) OpCode() ReqCode {
	return XInputOpcode
}

func ParseXIGetPropertyRequest(order binary.ByteOrder, body []byte, seq uint16) (*XIGetPropertyRequest, error) {
	if len(body) != 20 {
		return nil, NewError(LengthErrorCode, seq, 0, byte(XInputOpcode), XIGetProperty)
	}
	return &XIGetPropertyRequest{
		DeviceID: order.Uint16(body[0:2]),
		Delete:   body[2] != 0,
		Property: order.Uint32(body[4:8]),
		Type:     order.Uint32(body[8:12]),
		Offset:   order.Uint32(body[12:16]),
		Len:      order.Uint32(body[16:20]),
	}, nil
}

// XIGetSelectedEvents request
type XIGetSelectedEventsRequest struct {
	Window Window
}

func (r *XIGetSelectedEventsRequest) OpCode() ReqCode {
	return XInputOpcode
}

func ParseXIGetSelectedEventsRequest(order binary.ByteOrder, body []byte, seq uint16) (*XIGetSelectedEventsRequest, error) {
	if len(body) != 4 {
		return nil, NewError(LengthErrorCode, seq, 0, byte(XInputOpcode), XIGetSelectedEvents)
	}
	return &XIGetSelectedEventsRequest{
		Window: Window(order.Uint32(body[0:4])),
	}, nil
}

// XIBarrierReleasePointer request
type XIBarrierReleasePointerRequest struct {
	NumBarriers uint32
	Barriers    []XIBarrier
}

type XIBarrier struct {
	Barrier uint32
	EventID uint32
}

func (r *XIBarrierReleasePointerRequest) OpCode() ReqCode {
	return XInputOpcode
}

func ParseXIBarrierReleasePointerRequest(order binary.ByteOrder, body []byte, seq uint16) (*XIBarrierReleasePointerRequest, error) {
	if len(body) < 4 {
		return nil, NewError(LengthErrorCode, seq, 0, byte(XInputOpcode), XIBarrierReleasePointer)
	}
	numBarriers := order.Uint32(body[0:4])
	if len(body) != 4+int(numBarriers)*8 {
		return nil, NewError(LengthErrorCode, seq, 0, byte(XInputOpcode), XIBarrierReleasePointer)
	}
	barriers := make([]XIBarrier, numBarriers)
	for i := 0; i < int(numBarriers); i++ {
		offset := 4 + i*8
		barriers[i] = XIBarrier{
			Barrier: order.Uint32(body[offset : offset+4]),
			EventID: order.Uint32(body[offset+4 : offset+8]),
		}
	}
	return &XIBarrierReleasePointerRequest{
		NumBarriers: numBarriers,
		Barriers:    barriers,
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

func (r *ListInputDevicesRequest) OpCode() ReqCode {
	return XInputOpcode
}

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

func (r *OpenDeviceRequest) OpCode() ReqCode {
	return XInputOpcode
}

func ParseOpenDeviceRequest(order binary.ByteOrder, body []byte, seq uint16) (*OpenDeviceRequest, error) {
	if len(body) != 4 {
		return nil, NewError(LengthErrorCode, seq, 0, byte(XInputOpcode), XOpenDevice)
	}
	return &OpenDeviceRequest{
		DeviceID: body[0],
	}, nil
}

// QueryDeviceState reply
type QueryDeviceStateReply struct {
	Sequence  uint16
	NumEvents uint16
	Classes   []InputClassInfo
}

func (r *QueryDeviceStateReply) EncodeMessage(order binary.ByteOrder) []byte {
	var classBytes []byte
	for _, c := range r.Classes {
		classBytes = append(classBytes, c.EncodeMessage(order)...)
	}
	length := (len(classBytes) + 3) / 4

	reply := make([]byte, 32+len(classBytes))
	reply[0] = 1 // Reply
	order.PutUint16(reply[2:4], r.Sequence)
	order.PutUint32(reply[4:8], uint32(length))
	reply[8] = byte(len(r.Classes))
	copy(reply[32:], classBytes)
	return reply
}

// GetDeviceButtonMapping reply
type GetDeviceButtonMappingReply struct {
	Sequence uint16
	Map      []byte
}

func (r *GetDeviceButtonMappingReply) EncodeMessage(order binary.ByteOrder) []byte {
	length := len(r.Map)
	reply := make([]byte, 32+length)
	reply[0] = 1 // Reply
	reply[1] = byte(length)
	order.PutUint16(reply[2:4], r.Sequence)
	order.PutUint32(reply[4:8], uint32((length+3)/4))
	copy(reply[32:], r.Map)
	return reply
}

// GetDeviceModifierMapping reply
type GetDeviceModifierMappingReply struct {
	Sequence            uint16
	NumKeycodesPerMod byte
	Keycodes            []byte
}

func (r *GetDeviceModifierMappingReply) EncodeMessage(order binary.ByteOrder) []byte {
	length := len(r.Keycodes)
	reply := make([]byte, 32+length)
	reply[0] = 1 // Reply
	reply[1] = r.NumKeycodesPerMod
	order.PutUint16(reply[2:4], r.Sequence)
	order.PutUint32(reply[4:8], uint32((length+3)/4))
	copy(reply[32:], r.Keycodes)
	return reply
}

// GetFeedbackControl reply
type GetFeedbackControlReply struct {
	ReplyType byte
	Unused    byte
	Sequence  uint16
	Length    uint32
	NumEvents uint16
	Padding   [22]byte
	Feedbacks []FeedbackState
}

func (r *GetFeedbackControlReply) EncodeMessage(order binary.ByteOrder) []byte {
	var feedbackBytes []byte
	for _, f := range r.Feedbacks {
		feedbackBytes = append(feedbackBytes, f.EncodeMessage(order)...)
	}
	length := (len(feedbackBytes) + 3) / 4

	reply := make([]byte, 32+len(feedbackBytes))
	reply[0] = 1 // Reply
	order.PutUint16(reply[2:4], r.Sequence)
	order.PutUint32(reply[4:8], uint32(length))
	order.PutUint16(reply[8:10], r.NumEvents)
	copy(reply[32:], feedbackBytes)
	return reply
}

type FeedbackState interface {
	EncodeMessage(order binary.ByteOrder) []byte
}

type KbdFeedbackState struct {
	ClassID          byte
	ID               byte
	Len              uint16
	Pitch            uint16
	Duration         uint16
	LedMask          uint32
	LedValues        uint32
	GlobalAutoRepeat bool
	Click            byte
	Percent          byte
	AutoRepeats      [32]byte
}

func (f *KbdFeedbackState) EncodeMessage(order binary.ByteOrder) []byte {
	buf := new(bytes.Buffer)
	buf.WriteByte(f.ClassID)
	buf.WriteByte(f.ID)
	order.PutUint16(buf.Bytes()[2:4], f.Len)
	order.PutUint16(buf.Bytes()[4:6], f.Pitch)
	order.PutUint16(buf.Bytes()[6:8], f.Duration)
	order.PutUint32(buf.Bytes()[8:12], f.LedMask)
	order.PutUint32(buf.Bytes()[12:16], f.LedValues)
	if f.GlobalAutoRepeat {
		buf.WriteByte(1)
	} else {
		buf.WriteByte(0)
	}
	buf.WriteByte(f.Click)
	buf.WriteByte(f.Percent)
	buf.Write(f.AutoRepeats[:])
	return buf.Bytes()
}

type PtrFeedbackState struct {
	ClassID    byte
	ID         byte
	Len        uint16
	AccelNum   uint16
	AccelDenom uint16
	Threshold  uint16
}

func (f *PtrFeedbackState) EncodeMessage(order binary.ByteOrder) []byte {
	buf := new(bytes.Buffer)
	buf.WriteByte(f.ClassID)
	buf.WriteByte(f.ID)
	order.PutUint16(buf.Bytes()[2:4], f.Len)
	order.PutUint16(buf.Bytes()[4:6], f.AccelNum)
	order.PutUint16(buf.Bytes()[6:8], f.AccelDenom)
	order.PutUint16(buf.Bytes()[8:10], f.Threshold)
	return buf.Bytes()
}

// GetDeviceFocus reply
type GetDeviceFocusReply struct {
	ReplyType byte
	Unused    byte
	Sequence  uint16
	Length    uint32
	Focus     uint32
	Time      uint32
	RevertTo  byte
	Padding   [15]byte
}

func (r *GetDeviceFocusReply) EncodeMessage(order binary.ByteOrder) []byte {
	reply := make([]byte, 32)
	reply[0] = 1 // Reply
	order.PutUint16(reply[2:4], r.Sequence)
	order.PutUint32(reply[4:8], 0) // length
	order.PutUint32(reply[8:12], r.Focus)
	order.PutUint32(reply[12:16], r.Time)
	reply[16] = r.RevertTo
	return reply
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

func (r *SetDeviceModeRequest) OpCode() ReqCode {
	return XInputOpcode
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

func (r *SetDeviceValuatorsRequest) OpCode() ReqCode {
	return XInputOpcode
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

func (r *GetDeviceControlRequest) OpCode() ReqCode {
	return XInputOpcode
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

func (r *ChangeDeviceControlRequest) OpCode() ReqCode {
	return XInputOpcode
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

func (r *GetSelectedExtensionEventsRequest) OpCode() ReqCode {
	return XInputOpcode
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

func (r *ChangeDeviceDontPropagateListRequest) OpCode() ReqCode {
	return XInputOpcode
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

func (r *GetDeviceDontPropagateListRequest) OpCode() ReqCode {
	return XInputOpcode
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

func (r *AllowDeviceEventsRequest) OpCode() ReqCode {
	return XInputOpcode
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

func (r *CloseDeviceRequest) OpCode() ReqCode {
	return XInputOpcode
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

func (r *GrabDeviceRequest) OpCode() ReqCode {
	return XInputOpcode
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

func (r *UngrabDeviceRequest) OpCode() ReqCode {
	return XInputOpcode
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

func (r *SelectExtensionEventRequest) OpCode() ReqCode {
	return XInputOpcode
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

func (r *GetDeviceMotionEventsRequest) OpCode() ReqCode {
	return XInputOpcode
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

func (r *ChangeKeyboardDeviceRequest) OpCode() ReqCode {
	return XInputOpcode
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

func (r *ChangePointerDeviceRequest) OpCode() ReqCode {
	return XInputOpcode
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

func (r *GrabDeviceKeyRequest) OpCode() ReqCode {
	return XInputOpcode
}

func ParseGrabDeviceKeyRequest(order binary.ByteOrder, body []byte, seq uint16) (*GrabDeviceKeyRequest, error) {
	if len(body) < 13 {
		return nil, NewError(LengthErrorCode, seq, 0, byte(XInputOpcode), XGrabDeviceKey)
	}
	numClasses := order.Uint16(body[4:6])
	if len(body) != 13+int(numClasses)*4 {
		return nil, NewError(LengthErrorCode, seq, 0, byte(XInputOpcode), XGrabDeviceKey)
	}
	classes := make([]uint32, numClasses)
	for i := 0; i < int(numClasses); i++ {
		classes[i] = order.Uint32(body[13+i*4 : 17+i*4])
	}
	return &GrabDeviceKeyRequest{
		GrabWindow:      Window(order.Uint32(body[0:4])),
		NumClasses:      numClasses,
		OwnerEvents:     body[6] != 0,
		ThisDeviceMode:  body[7],
		OtherDeviceMode: body[8],
		DeviceID:        body[9],
		Modifiers:       order.Uint16(body[10:12]),
		Key:             body[12],
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

func (r *UngrabDeviceKeyRequest) OpCode() ReqCode {
	return XInputOpcode
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

func (r *GrabDeviceButtonRequest) OpCode() ReqCode {
	return XInputOpcode
}

func ParseGrabDeviceButtonRequest(order binary.ByteOrder, body []byte, seq uint16) (*GrabDeviceButtonRequest, error) {
	if len(body) < 13 {
		return nil, NewError(LengthErrorCode, seq, 0, byte(XInputOpcode), XGrabDeviceButton)
	}
	numClasses := order.Uint16(body[4:6])
	expectedLen := 13 + int(numClasses)*4
	if len(body) != expectedLen {
		return nil, NewError(LengthErrorCode, seq, 0, byte(XInputOpcode), XGrabDeviceButton)
	}
	classes := make([]uint32, numClasses)
	for i := 0; i < int(numClasses); i++ {
		classes[i] = order.Uint32(body[13+i*4 : 17+i*4])
	}
	return &GrabDeviceButtonRequest{
		GrabWindow:      Window(order.Uint32(body[0:4])),
		NumClasses:      numClasses,
		OwnerEvents:     body[6] != 0,
		ThisDeviceMode:  body[7],
		OtherDeviceMode: body[8],
		DeviceID:        body[9],
		Modifiers:       order.Uint16(body[10:12]),
		Button:          body[12],
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

func (r *UngrabDeviceButtonRequest) OpCode() ReqCode {
	return XInputOpcode
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

func (r *GetDeviceFocusRequest) OpCode() ReqCode {
	return XInputOpcode
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

func (r *SetDeviceFocusRequest) OpCode() ReqCode {
	return XInputOpcode
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

func (r *GetFeedbackControlRequest) OpCode() ReqCode {
	return XInputOpcode
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

func (r *ChangeFeedbackControlRequest) OpCode() ReqCode {
	return XInputOpcode
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
	DeviceID byte
	FirstKey byte
	Count    byte
}

func (r *GetDeviceKeyMappingRequest) OpCode() ReqCode {
	return XInputOpcode
}

func ParseGetDeviceKeyMappingRequest(order binary.ByteOrder, body []byte, seq uint16) (*GetDeviceKeyMappingRequest, error) {
	if len(body) != 4 {
		return nil, NewError(LengthErrorCode, seq, 0, byte(XInputOpcode), XGetDeviceKeyMapping)
	}
	return &GetDeviceKeyMappingRequest{
		DeviceID: body[0],
		FirstKey: body[1],
		Count:    body[2],
	}, nil
}

// GetDeviceKeyMapping reply
type GetDeviceKeyMappingReply struct {
	Sequence          uint16
	KeysymsPerKeycode byte
	Keysyms           []uint32
}

func (r *GetDeviceKeyMappingReply) EncodeMessage(order binary.ByteOrder) []byte {
	length := len(r.Keysyms) * 4
	reply := make([]byte, 32+length)
	reply[0] = 1 // Reply
	reply[1] = r.KeysymsPerKeycode
	order.PutUint16(reply[2:4], r.Sequence)
	order.PutUint32(reply[4:8], uint32(length/4))
	offset := 32
	for _, keysym := range r.Keysyms {
		order.PutUint32(reply[offset:offset+4], keysym)
		offset += 4
	}
	return reply
}

// ChangeDeviceKeyMapping request
type ChangeDeviceKeyMappingRequest struct {
	DeviceID          byte
	FirstKey          byte
	KeysymsPerKeycode byte
	KeycodeCount      byte
	Keysyms           []uint32
}

func (r *ChangeDeviceKeyMappingRequest) OpCode() ReqCode {
	return XInputOpcode
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
		DeviceID:          body[0],
		FirstKey:          body[1],
		KeysymsPerKeycode: keysymsPerKeycode,
		KeycodeCount:      keycodeCount,
		Keysyms:           keysyms,
	}, nil
}

// GetDeviceModifierMapping request
type GetDeviceModifierMappingRequest struct {
	DeviceID byte
}

func (r *GetDeviceModifierMappingRequest) OpCode() ReqCode {
	return XInputOpcode
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

func (r *SetDeviceModifierMappingRequest) OpCode() ReqCode {
	return XInputOpcode
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

// SetDeviceModifierMapping reply
type SetDeviceModifierMappingReply struct {
	ReplyType byte
	Unused    byte
	Sequence  uint16
	Length    uint32
	Status    byte
	Padding   [23]byte
}

func (r *SetDeviceModifierMappingReply) EncodeMessage(order binary.ByteOrder) []byte {
	reply := make([]byte, 32)
	reply[0] = 1 // Reply
	reply[1] = r.Status
	order.PutUint16(reply[2:4], r.Sequence)
	order.PutUint32(reply[4:8], 0)
	return reply
}

// GetDeviceButtonMapping request
type GetDeviceButtonMappingRequest struct {
	DeviceID byte
}

func (r *GetDeviceButtonMappingRequest) OpCode() ReqCode {
	return XInputOpcode
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

func (r *SetDeviceButtonMappingRequest) OpCode() ReqCode {
	return XInputOpcode
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

// SetDeviceButtonMapping reply
type SetDeviceButtonMappingReply struct {
	ReplyType byte
	Unused    byte
	Sequence  uint16
	Length    uint32
	Status    byte
	Padding   [23]byte
}

func (r *SetDeviceButtonMappingReply) EncodeMessage(order binary.ByteOrder) []byte {
	reply := make([]byte, 32)
	reply[0] = 1 // Reply
	reply[1] = r.Status
	order.PutUint16(reply[2:4], r.Sequence)
	order.PutUint32(reply[4:8], 0)
	return reply
}

// QueryDeviceState request
type QueryDeviceStateRequest struct {
	DeviceID byte
}

func (r *QueryDeviceStateRequest) OpCode() ReqCode {
	return XInputOpcode
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

func (r *SendExtensionEventRequest) OpCode() ReqCode {
	return XInputOpcode
}

func ParseSendExtensionEventRequest(order binary.ByteOrder, body []byte, seq uint16) (*SendExtensionEventRequest, error) {
	// Base length before events and classes
	fixedBaseLen := 12 // 4 (Destination) + 2 (NumClasses) + 1 (NumEvents) + 1 (DeviceID) + 1 (Propagate) + 3 (Padding)

	if len(body) < fixedBaseLen {
		return nil, NewError(LengthErrorCode, seq, 0, byte(XInputOpcode), XSendExtensionEvent)
	}

	destination := Window(order.Uint32(body[0:4]))
	numClasses := order.Uint16(body[4:6])
	numEvents := body[6]
	deviceID := body[7]
	propagate := body[8] != 0

	// Calculate expected total length
	expectedLen := fixedBaseLen + int(numEvents)*32 + int(numClasses)*4
	if len(body) != expectedLen {
		return nil, NewError(LengthErrorCode, seq, 0, byte(XInputOpcode), XSendExtensionEvent)
	}

	eventsStart := fixedBaseLen // Events start after the fixed part (including padding)
	eventsEnd := eventsStart + int(numEvents)*32
	events := body[eventsStart:eventsEnd]

	classesStart := eventsEnd
	classes := make([]uint32, numClasses)
	for i := 0; i < int(numClasses); i++ {
		classes[i] = order.Uint32(body[classesStart+i*4 : classesStart+(i+1)*4])
	}

	return &SendExtensionEventRequest{
		Destination: destination,
		DeviceID:    deviceID,
		Propagate:   propagate,
		NumClasses:  numClasses,
		NumEvents:   numEvents,
		Events:      events,
		Classes:     classes,
	}, nil
}

// DeviceBell request
type DeviceBellRequest struct {
	DeviceID      byte
	FeedbackID    byte
	FeedbackClass byte
	Percent       byte
}

func (r *DeviceBellRequest) OpCode() ReqCode {
	return XInputOpcode
}

func ParseDeviceBellRequest(order binary.ByteOrder, body []byte, seq uint16) (*DeviceBellRequest, error) {
	if len(body) != 4 {
		return nil, NewError(LengthErrorCode, seq, 0, byte(XInputOpcode), XDeviceBell)
	}
	return &DeviceBellRequest{
		DeviceID:      body[0],
		FeedbackID:    body[1],
		FeedbackClass: body[2],
		Percent:       body[3],
	}, nil
}

//
// XInput 2.0 requests
//

// XIQueryVersion request
type XIQueryVersionRequest struct {
	MajorVersion uint16
	MinorVersion uint16
}

func (r *XIQueryVersionRequest) OpCode() ReqCode {
	return XInputOpcode
}

func ParseXIQueryVersionRequest(order binary.ByteOrder, body []byte, seq uint16) (*XIQueryVersionRequest, error) {
	if len(body) != 4 {
		return nil, NewError(LengthErrorCode, seq, 0, byte(XInputOpcode), XIQueryVersion)
	}
	return &XIQueryVersionRequest{
		MajorVersion: order.Uint16(body[0:2]),
		MinorVersion: order.Uint16(body[2:4]),
	}, nil
}
