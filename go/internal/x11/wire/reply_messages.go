//go:build x11

package wire

import (
	"bytes"
	"encoding/binary"
	"io"
	"sync"
)

type ServerMessage interface {
	EncodeMessage(order binary.ByteOrder) []byte
}

var (
	sequenceToOpcode = make(map[uint16]Opcodes)
	seqMutex         sync.Mutex
)

func ExpectReply(sequence uint16, opcodes Opcodes) {
	seqMutex.Lock()
	defer seqMutex.Unlock()
	sequenceToOpcode[sequence] = opcodes
}

func ReadServerMessages(conn io.Reader, order binary.ByteOrder) <-chan ServerMessage {
	ch := make(chan ServerMessage, 1)
	go func() {
		defer close(ch)
		for {
			header := make([]byte, 32)
			if _, err := io.ReadFull(conn, header); err != nil {
				if err != io.EOF {
					debugf("X11: failed to read server message header: %v", err)
				}
				return
			}

			msgType := header[0]
			sequenceNumber := order.Uint16(header[2:4])

			debugf("X11 received server message: type=%d, sequence=%d", msgType, sequenceNumber)

			switch msgType {
			case 0:
				p, err := ParseError(header, order)
				if err != nil {
					debugf("X11 ReadServerMessages: ParseError(%x): %v", header, err)
					continue
				}
				ch <- p
			case 1:
				replyLength := 4 * order.Uint32(header[4:8])
				msg := append(header, make([]byte, replyLength)...)
				if _, err := io.ReadFull(conn, msg[32:]); err != nil {
					debugf("X11: failed to read remaining server message: %v", err)
					return
				}
				seqMutex.Lock()
				opcodes, ok := sequenceToOpcode[sequenceNumber]
				if ok {
					delete(sequenceToOpcode, sequenceNumber)
				}
				seqMutex.Unlock()
				if !ok {
					debugf("X11: unknown sequence number %d", sequenceNumber)
					continue
				}

				p, err := ParseReply(opcodes, msg, order)
				if err != nil {
					debugf("X11 ReadServerMessages: ParseReply(%x): %v", msg, err)
					continue
				}
				ch <- p
			default:
				p, err := ParseEvent(header, order)
				if err != nil {
					debugf("X11 ReadServerMessages: ParseEvent(%x): %v", header, err)
					continue
				}
				ch <- p
			}
		}
	}()

	return ch
}

func ParseReply(opcodes Opcodes, msg []byte, order binary.ByteOrder) (ServerMessage, error) {
	switch opcodes.Major {
	case GetWindowAttributes:
		return ParseGetWindowAttributesReply(order, msg)
	case GetGeometry:
		return ParseGetGeometryReply(order, msg)
	case InternAtom:
		return ParseInternAtomReply(order, msg)
	case GetAtomName:
		return ParseGetAtomNameReply(order, msg)
	case GetProperty:
		return ParseGetPropertyReply(order, msg)
	case ListProperties:
		return ParseListPropertiesReply(order, msg)
	case QueryTextExtents:
		return ParseQueryTextExtentsReply(order, msg)
	case GetMotionEvents:
		return ParseGetMotionEventsReply(order, msg)
	case GetSelectionOwner:
		return ParseGetSelectionOwnerReply(order, msg)
	case GrabPointer:
		return ParseGrabPointerReply(order, msg)
	case GrabKeyboard:
		return ParseGrabKeyboardReply(order, msg)
	case QueryPointer:
		return ParseQueryPointerReply(order, msg)
	case TranslateCoords:
		return ParseTranslateCoordsReply(order, msg)
	case GetInputFocus:
		return ParseGetInputFocusReply(order, msg)
	case QueryFont:
		return ParseQueryFontReply(order, msg)
	case ListFonts:
		return ParseListFontsReply(order, msg)
	case GetImage:
		return ParseGetImageReply(order, msg)
	case AllocColor:
		return ParseAllocColorReply(order, msg)
	case AllocNamedColor:
		return ParseAllocNamedColorReply(order, msg)
	case ListInstalledColormaps:
		return ParseListInstalledColormapsReply(order, msg)
	case QueryColors:
		return ParseQueryColorsReply(order, msg)
	case LookupColor:
		return ParseLookupColorReply(order, msg)
	case QueryBestSize:
		return ParseQueryBestSizeReply(order, msg)
	case QueryExtension:
		return ParseQueryExtensionReply(order, msg)
	case GetKeyboardMapping:
		return ParseGetKeyboardMappingReply(order, msg)
	case GetKeyboardControl:
		return ParseGetKeyboardControlReply(order, msg)
	case GetPointerMapping:
		return ParseGetPointerMappingReply(order, msg)
	case SetPointerMapping:
		return ParseSetPointerMappingReply(order, msg)
	case GetModifierMapping:
		return ParseGetModifierMappingReply(order, msg)
	case SetModifierMapping:
		return ParseSetModifierMappingReply(order, msg)
	case GetScreenSaver:
		return ParseGetScreenSaverReply(order, msg)
	case ListHosts:
		return ParseListHostsReply(order, msg)
	case QueryKeymap:
		return ParseQueryKeymapReply(order, msg)
	case GetFontPath:
		return ParseGetFontPathReply(order, msg)
	case ListFontsWithInfo:
		return ParseListFontsWithInfoReply(order, msg)
	case QueryTree:
		return ParseQueryTreeReply(order, msg)
	case AllocColorCells:
		return ParseAllocColorCellsReply(order, msg)
	case AllocColorPlanes:
		return ParseAllocColorPlanesReply(order, msg)
	case ListExtensions:
		return ParseListExtensionsReply(order, msg)
	case GetPointerControl:
		return ParseGetPointerControlReply(order, msg)
	case XInputOpcode:
		return parseXInputReply(opcodes.Minor, order, msg)
	case BigRequestsOpcode:
		return &BigRequestsEnableReply{
			Sequence:         order.Uint16(msg[2:4]),
			MaxRequestLength: order.Uint32(msg[8:12]),
		}, nil
	default:
		return nil, NewError(RequestErrorCode, 0, 0, Opcodes{Major: opcodes.Major, Minor: 0})
	}
}

func parseXInputReply(minorOpcode uint8, order binary.ByteOrder, b []byte) (ServerMessage, error) {
	switch minorOpcode {
	case XGetExtensionVersion:
		return ParseGetExtensionVersionReply(order, b)
	case XListInputDevices:
		return ParseListInputDevicesReply(order, b)
	case XOpenDevice:
		return ParseOpenDeviceReply(order, b)
	case XCloseDevice:
		return ParseCloseDeviceReply(order, b)
	case XSetDeviceMode:
		return ParseSetDeviceModeReply(order, b)
	case XGetSelectedExtensionEvents:
		return ParseGetSelectedExtensionEventsReply(order, b)
	case XGetDeviceDontPropagateList:
		return ParseGetDeviceDontPropagateListReply(order, b)
	case XGetDeviceMotionEvents:
		return ParseGetDeviceMotionEventsReply(order, b)
	case XChangeKeyboardDevice:
		return ParseChangeKeyboardDeviceReply(order, b)
	case XChangePointerDevice:
		return ParseChangePointerDeviceReply(order, b)
	case XGrabDevice:
		return ParseGrabDeviceReply(order, b)
	case XGetDeviceFocus:
		return ParseGetDeviceFocusReply(order, b)
	case XGetFeedbackControl:
		return ParseGetFeedbackControlReply(order, b)
	case XGetDeviceKeyMapping:
		return ParseGetDeviceKeyMappingReply(order, b)
	case XGetDeviceModifierMapping:
		return ParseGetDeviceModifierMappingReply(order, b)
	case XSetDeviceModifierMapping:
		return ParseSetDeviceModifierMappingReply(order, b)
	case XGetDeviceButtonMapping:
		return ParseGetDeviceButtonMappingReply(order, b)
	case XSetDeviceButtonMapping:
		return ParseSetDeviceButtonMappingReply(order, b)
	case XQueryDeviceState:
		return ParseQueryDeviceStateReply(order, b)
	case XSetDeviceValuators:
		return ParseSetDeviceValuatorsReply(order, b)
	case XGetDeviceControl:
		return ParseGetDeviceControlReply(order, b)
	case XChangeDeviceControl:
		return ParseChangeDeviceControlReply(order, b)
	}
	return nil, NewError(RequestErrorCode, 0, 0, Opcodes{Major: XInputOpcode, Minor: minorOpcode})
}

type XCharInfo struct {
	LeftSideBearing  int16
	RightSideBearing int16
	CharacterWidth   uint16
	Ascent           int16
	Descent          int16
	Attributes       uint16
}

func BoolToByte(b bool) byte {
	if b {
		return 1
	}
	return 0
}

func ByteToBool(b byte) bool {
	return b != 0
}

// GetWindowAttributes: 3
type GetWindowAttributesReply struct {
	ReplyType          byte
	BackingStore       byte
	Sequence           uint16
	Length             uint32
	VisualID           uint32
	Class              uint16
	BitGravity         byte
	WinGravity         byte
	BackingPlanes      uint32
	BackingPixel       uint32
	SaveUnder          byte
	MapIsInstalled     byte
	MapState           byte
	OverrideRedirect   byte
	Colormap           uint32
	AllEventMasks      uint32
	YourEventMask      uint32
	DoNotPropagateMask uint16
}

func (r *GetWindowAttributesReply) EncodeMessage(order binary.ByteOrder) []byte {
	reply := make([]byte, 44)
	reply[0] = 1 // Reply type
	reply[1] = r.BackingStore
	order.PutUint16(reply[2:4], r.Sequence)
	order.PutUint32(reply[4:8], 3) // Reply length (3 * 4 bytes = 12 bytes, plus 32 bytes header = 44 bytes total)
	order.PutUint32(reply[8:12], r.VisualID)
	order.PutUint16(reply[12:14], r.Class)
	reply[14] = r.BitGravity
	reply[15] = r.WinGravity
	order.PutUint32(reply[16:20], r.BackingPlanes)
	order.PutUint32(reply[20:24], r.BackingPixel)
	reply[24] = r.SaveUnder
	reply[25] = r.MapIsInstalled
	reply[26] = r.MapState
	reply[27] = r.OverrideRedirect
	order.PutUint32(reply[28:32], r.Colormap)
	order.PutUint32(reply[32:36], r.AllEventMasks)
	order.PutUint32(reply[36:40], r.YourEventMask)
	order.PutUint16(reply[40:42], r.DoNotPropagateMask)
	// reply[42:44] is padding
	return reply
}

func ParseGetWindowAttributesReply(order binary.ByteOrder, b []byte) (*GetWindowAttributesReply, error) {
	if len(b) < 44 {
		return nil, NewError(LengthErrorCode, 0, 0, Opcodes{Major: 0, Minor: 0})
	}
	r := &GetWindowAttributesReply{
		ReplyType:          b[0],
		BackingStore:       b[1],
		Sequence:           order.Uint16(b[2:4]),
		Length:             order.Uint32(b[4:8]),
		VisualID:           order.Uint32(b[8:12]),
		Class:              order.Uint16(b[12:14]),
		BitGravity:         b[14],
		WinGravity:         b[15],
		BackingPlanes:      order.Uint32(b[16:20]),
		BackingPixel:       order.Uint32(b[20:24]),
		SaveUnder:          b[24],
		MapIsInstalled:     b[25],
		MapState:           b[26],
		OverrideRedirect:   b[27],
		Colormap:           order.Uint32(b[28:32]),
		AllEventMasks:      order.Uint32(b[32:36]),
		YourEventMask:      order.Uint32(b[36:40]),
		DoNotPropagateMask: order.Uint16(b[40:42]),
	}
	return r, nil
}

// GetGeometry: 14
type GetGeometryReply struct {
	Sequence      uint16
	Depth         byte
	Root          uint32
	X, Y          int16
	Width, Height uint16
	BorderWidth   uint16
}

func (r *GetGeometryReply) EncodeMessage(order binary.ByteOrder) []byte {
	reply := make([]byte, 32)
	reply[0] = 1 // Reply type
	reply[1] = r.Depth
	order.PutUint16(reply[2:4], r.Sequence)
	order.PutUint32(reply[4:8], 0) // Reply length (0 * 4 bytes = 0 bytes, plus 32 bytes header = 32 bytes total)
	order.PutUint32(reply[8:12], r.Root)
	order.PutUint16(reply[12:14], uint16(r.X))
	order.PutUint16(reply[14:16], uint16(r.Y))
	order.PutUint16(reply[16:18], r.Width)
	order.PutUint16(reply[18:20], r.Height)
	order.PutUint16(reply[20:22], r.BorderWidth)
	return reply
}

func ParseGetGeometryReply(order binary.ByteOrder, b []byte) (*GetGeometryReply, error) {
	if len(b) < 32 {
		return nil, NewError(LengthErrorCode, 0, 0, Opcodes{Major: 0, Minor: 0})
	}
	r := &GetGeometryReply{
		Depth:       b[1],
		Sequence:    order.Uint16(b[2:4]),
		Root:        order.Uint32(b[8:12]),
		X:           int16(order.Uint16(b[12:14])),
		Y:           int16(order.Uint16(b[14:16])),
		Width:       order.Uint16(b[16:18]),
		Height:      order.Uint16(b[18:20]),
		BorderWidth: order.Uint16(b[20:22]),
	}
	return r, nil
}

// InternAtom: 16
type InternAtomReply struct {
	Sequence uint16
	Atom     uint32
}

func (r *InternAtomReply) EncodeMessage(order binary.ByteOrder) []byte {
	reply := make([]byte, 32)
	reply[0] = 1 // Reply type
	// byte 1 is unused
	order.PutUint16(reply[2:4], r.Sequence)
	order.PutUint32(reply[4:8], 0) // Reply length (0 * 4 bytes = 0 bytes, plus 32 bytes header = 32 bytes total)
	order.PutUint32(reply[8:12], r.Atom)
	// reply[12:32] is padding
	return reply
}

func ParseInternAtomReply(order binary.ByteOrder, b []byte) (*InternAtomReply, error) {
	if len(b) < 32 {
		return nil, NewError(LengthErrorCode, 0, 0, Opcodes{Major: 0, Minor: 0})
	}
	r := &InternAtomReply{
		Sequence: order.Uint16(b[2:4]),
		Atom:     order.Uint32(b[8:12]),
	}
	return r, nil
}

// GetAtomName: 17
type GetAtomNameReply struct {
	Sequence   uint16
	NameLength uint16
	Name       string
}

func (r *GetAtomNameReply) EncodeMessage(order binary.ByteOrder) []byte {
	nameLen := len(r.Name)
	p := (4 - (nameLen % 4)) % 4
	reply := make([]byte, 32+nameLen+p)
	reply[0] = 1 // Reply type
	// byte 1 is unused
	order.PutUint16(reply[2:4], r.Sequence)
	order.PutUint32(reply[4:8], uint32((nameLen+p)/4)) // Reply length
	order.PutUint16(reply[8:10], uint16(nameLen))
	// reply[10:32] is padding
	copy(reply[32:], r.Name)
	return reply
}

func ParseGetAtomNameReply(order binary.ByteOrder, b []byte) (*GetAtomNameReply, error) {
	if len(b) < 32 {
		return nil, NewError(LengthErrorCode, 0, 0, Opcodes{Major: 0, Minor: 0})
	}
	nameLen := order.Uint16(b[8:10])
	if len(b) < 32+int(nameLen) {
		return nil, NewError(LengthErrorCode, 0, 0, Opcodes{Major: 0, Minor: 0})
	}
	r := &GetAtomNameReply{
		Sequence:   order.Uint16(b[2:4]),
		NameLength: nameLen,
		Name:       string(b[32 : 32+nameLen]),
	}
	return r, nil
}

// GetProperty: 20
type GetPropertyReply struct {
	Sequence              uint16
	Format                byte
	PropertyType          uint32
	BytesAfter            uint32
	ValueLenInFormatUnits uint32
	Value                 []byte
}

func (r *GetPropertyReply) EncodeMessage(order binary.ByteOrder) []byte {
	n := len(r.Value)
	p := (4 - (n % 4)) % 4
	replyLen := (n + p) / 4

	reply := make([]byte, 32+n+p)
	reply[0] = 1 // Reply type
	reply[1] = r.Format
	order.PutUint16(reply[2:4], r.Sequence)
	order.PutUint32(reply[4:8], uint32(replyLen)) // Reply length
	order.PutUint32(reply[8:12], r.PropertyType)
	order.PutUint32(reply[12:16], r.BytesAfter)
	order.PutUint32(reply[16:20], r.ValueLenInFormatUnits)
	// reply[20:32] is padding
	copy(reply[32:], r.Value)
	return reply
}

func ParseGetPropertyReply(order binary.ByteOrder, b []byte) (*GetPropertyReply, error) {
	if len(b) < 32 {
		return nil, NewError(LengthErrorCode, 0, 0, Opcodes{Major: 0, Minor: 0})
	}
	valLen := order.Uint32(b[4:8]) * 4
	if len(b) < 32+int(valLen) {
		return nil, NewError(LengthErrorCode, 0, 0, Opcodes{Major: 0, Minor: 0})
	}
	r := &GetPropertyReply{
		Sequence:              order.Uint16(b[2:4]),
		Format:                b[1],
		PropertyType:          order.Uint32(b[8:12]),
		BytesAfter:            order.Uint32(b[12:16]),
		ValueLenInFormatUnits: order.Uint32(b[16:20]),
		Value:                 b[32 : 32+valLen],
	}
	return r, nil
}

// ListProperties: 21
type ListPropertiesReply struct {
	Sequence      uint16
	NumProperties uint16
	Atoms         []uint32
}

func (r *ListPropertiesReply) EncodeMessage(order binary.ByteOrder) []byte {
	numAtoms := len(r.Atoms)
	reply := make([]byte, 32+numAtoms*4)
	reply[0] = 1 // Reply type
	// byte 1 is unused
	order.PutUint16(reply[2:4], r.Sequence)
	order.PutUint32(reply[4:8], uint32(numAtoms)) // Reply length
	order.PutUint16(reply[8:10], uint16(numAtoms))
	// reply[10:32] is padding
	for i, atom := range r.Atoms {
		order.PutUint32(reply[32+i*4:], atom)
	}
	return reply
}

func ParseListPropertiesReply(order binary.ByteOrder, b []byte) (*ListPropertiesReply, error) {
	if len(b) < 32 {
		return nil, NewError(LengthErrorCode, 0, 0, Opcodes{Major: 0, Minor: 0})
	}
	numAtoms := order.Uint16(b[8:10])
	if len(b) < 32+int(numAtoms)*4 {
		return nil, NewError(LengthErrorCode, 0, 0, Opcodes{Major: 0, Minor: 0})
	}
	atoms := make([]uint32, numAtoms)
	for i := 0; i < int(numAtoms); i++ {
		atoms[i] = order.Uint32(b[32+i*4:])
	}
	r := &ListPropertiesReply{
		Sequence:      order.Uint16(b[2:4]),
		NumProperties: numAtoms,
		Atoms:         atoms,
	}
	return r, nil
}

// QueryTextExtents: 48
type QueryTextExtentsReply struct {
	Sequence       uint16
	DrawDirection  byte
	FontAscent     int16
	FontDescent    int16
	OverallAscent  int16
	OverallDescent int16
	OverallWidth   int32
	OverallLeft    int32
	OverallRight   int32
}

func (r *QueryTextExtentsReply) EncodeMessage(order binary.ByteOrder) []byte {
	reply := make([]byte, 32)
	reply[0] = 1 // Reply type
	reply[1] = r.DrawDirection
	order.PutUint16(reply[2:4], r.Sequence)
	order.PutUint32(reply[4:8], 0)
	order.PutUint16(reply[8:10], uint16(r.FontAscent))
	order.PutUint16(reply[10:12], uint16(r.FontDescent))
	order.PutUint16(reply[12:14], uint16(r.OverallAscent))
	order.PutUint16(reply[14:16], uint16(r.OverallDescent))
	order.PutUint32(reply[16:20], uint32(r.OverallWidth))
	order.PutUint32(reply[20:24], uint32(r.OverallLeft))
	order.PutUint32(reply[24:28], uint32(r.OverallRight))
	return reply
}

func ParseQueryTextExtentsReply(order binary.ByteOrder, b []byte) (*QueryTextExtentsReply, error) {
	if len(b) < 32 {
		return nil, NewError(LengthErrorCode, 0, 0, Opcodes{Major: 0, Minor: 0})
	}
	r := &QueryTextExtentsReply{
		Sequence:       order.Uint16(b[2:4]),
		DrawDirection:  b[1],
		FontAscent:     int16(order.Uint16(b[8:10])),
		FontDescent:    int16(order.Uint16(b[10:12])),
		OverallAscent:  int16(order.Uint16(b[12:14])),
		OverallDescent: int16(order.Uint16(b[14:16])),
		OverallWidth:   int32(order.Uint32(b[16:20])),
		OverallLeft:    int32(order.Uint32(b[20:24])),
		OverallRight:   int32(order.Uint32(b[24:28])),
	}
	return r, nil
}

// GetMotionEvents: 39
type GetMotionEventsReply struct {
	Sequence uint16
	NEvents  uint32
	Events   []TimeCoord
}

type TimeCoord struct {
	Time uint32
	X, Y int16
}

func (r *GetMotionEventsReply) EncodeMessage(order binary.ByteOrder) []byte {
	reply := make([]byte, 32+len(r.Events)*8)
	reply[0] = 1 // Reply type
	order.PutUint16(reply[2:4], r.Sequence)
	order.PutUint32(reply[4:8], uint32(len(r.Events)*2))
	order.PutUint32(reply[8:12], r.NEvents)
	for i, event := range r.Events {
		order.PutUint32(reply[32+i*8:], event.Time)
		order.PutUint16(reply[32+i*8+4:], uint16(event.X))
		order.PutUint16(reply[32+i*8+6:], uint16(event.Y))
	}
	return reply
}

func ParseGetMotionEventsReply(order binary.ByteOrder, b []byte) (*GetMotionEventsReply, error) {
	if len(b) < 32 {
		return nil, NewError(LengthErrorCode, 0, 0, Opcodes{Major: 0, Minor: 0})
	}
	nEvents := order.Uint32(b[8:12])
	if len(b) < 32+int(nEvents)*8 {
		return nil, NewError(LengthErrorCode, 0, 0, Opcodes{Major: 0, Minor: 0})
	}
	events := make([]TimeCoord, nEvents)
	for i := 0; i < int(nEvents); i++ {
		events[i] = TimeCoord{
			Time: order.Uint32(b[32+i*8:]),
			X:    int16(order.Uint16(b[32+i*8+4:])),
			Y:    int16(order.Uint16(b[32+i*8+6:])),
		}
	}
	r := &GetMotionEventsReply{
		Sequence: order.Uint16(b[2:4]),
		NEvents:  nEvents,
		Events:   events,
	}
	return r, nil
}

// GetSelectionOwner: 23
type GetSelectionOwnerReply struct {
	Sequence uint16
	Owner    uint32
}

func (r *GetSelectionOwnerReply) EncodeMessage(order binary.ByteOrder) []byte {
	reply := make([]byte, 32)
	reply[0] = 1 // Reply type
	// byte 1 is unused
	order.PutUint16(reply[2:4], r.Sequence)
	order.PutUint32(reply[4:8], 0) // Reply length (0 * 4 bytes = 0 bytes, plus 32 bytes header = 32 bytes total)
	order.PutUint32(reply[8:12], r.Owner)
	// reply[12:32] is padding
	return reply
}

func ParseGetSelectionOwnerReply(order binary.ByteOrder, b []byte) (*GetSelectionOwnerReply, error) {
	if len(b) < 32 {
		return nil, NewError(LengthErrorCode, 0, 0, Opcodes{Major: 0, Minor: 0})
	}
	r := &GetSelectionOwnerReply{
		Sequence: order.Uint16(b[2:4]),
		Owner:    order.Uint32(b[8:12]),
	}
	return r, nil
}

// GrabPointer: 26
type GrabPointerReply struct {
	Sequence uint16
	Status   byte
}

func (r *GrabPointerReply) EncodeMessage(order binary.ByteOrder) []byte {
	reply := make([]byte, 32)
	reply[0] = 1 // Reply type
	reply[1] = r.Status
	order.PutUint16(reply[2:4], r.Sequence)
	order.PutUint32(reply[4:8], 0) // Reply length (0 * 4 bytes = 0 bytes, plus 32 bytes header = 32 bytes total)
	// reply[8:32] is padding
	return reply
}

func ParseGrabPointerReply(order binary.ByteOrder, b []byte) (*GrabPointerReply, error) {
	if len(b) < 32 {
		return nil, NewError(LengthErrorCode, 0, 0, Opcodes{Major: 0, Minor: 0})
	}
	r := &GrabPointerReply{
		Sequence: order.Uint16(b[2:4]),
		Status:   b[1],
	}
	return r, nil
}

// GrabKeyboard: 31
type GrabKeyboardReply struct {
	Sequence uint16
	Status   byte
}

func (r *GrabKeyboardReply) EncodeMessage(order binary.ByteOrder) []byte {
	reply := make([]byte, 32)
	reply[0] = 1 // Reply type
	reply[1] = r.Status
	order.PutUint16(reply[2:4], r.Sequence)
	order.PutUint32(reply[4:8], 0) // Reply length (0 * 4 bytes = 0 bytes, plus 32 bytes header = 32 bytes total)
	// reply[8:32] is padding
	return reply
}

func ParseGrabKeyboardReply(order binary.ByteOrder, b []byte) (*GrabKeyboardReply, error) {
	if len(b) < 32 {
		return nil, NewError(LengthErrorCode, 0, 0, Opcodes{Major: 0, Minor: 0})
	}
	r := &GrabKeyboardReply{
		Sequence: order.Uint16(b[2:4]),
		Status:   b[1],
	}
	return r, nil
}

// QueryPointer: 38
type QueryPointerReply struct {
	Sequence     uint16
	SameScreen   bool
	Root         uint32
	Child        uint32
	RootX, RootY int16
	WinX, WinY   int16
	Mask         uint16
}

func (r *QueryPointerReply) EncodeMessage(order binary.ByteOrder) []byte {
	reply := make([]byte, 32)
	reply[0] = 1 // Reply type
	reply[1] = BoolToByte(r.SameScreen)
	order.PutUint16(reply[2:4], r.Sequence)
	order.PutUint32(reply[4:8], 0) // Reply length (0 * 4 bytes = 0 bytes, plus 32 bytes header = 32 bytes total)
	order.PutUint32(reply[8:12], r.Root)
	order.PutUint32(reply[12:16], r.Child)
	order.PutUint16(reply[16:18], uint16(r.RootX))
	order.PutUint16(reply[18:20], uint16(r.RootY))
	order.PutUint16(reply[20:22], uint16(r.WinX))
	order.PutUint16(reply[22:24], uint16(r.WinY))
	order.PutUint16(reply[24:26], r.Mask)
	// reply[26:32] is padding
	return reply
}

func ParseQueryPointerReply(order binary.ByteOrder, b []byte) (*QueryPointerReply, error) {
	if len(b) < 32 {
		return nil, NewError(LengthErrorCode, 0, 0, Opcodes{Major: 0, Minor: 0})
	}
	r := &QueryPointerReply{
		Sequence:   order.Uint16(b[2:4]),
		SameScreen: b[1] != 0,
		Root:       order.Uint32(b[8:12]),
		Child:      order.Uint32(b[12:16]),
		RootX:      int16(order.Uint16(b[16:18])),
		RootY:      int16(order.Uint16(b[18:20])),
		WinX:       int16(order.Uint16(b[20:22])),
		WinY:       int16(order.Uint16(b[22:24])),
		Mask:       order.Uint16(b[24:26]),
	}
	return r, nil
}

// TranslateCoords: 40
type TranslateCoordsReply struct {
	Sequence   uint16
	SameScreen bool
	Child      uint32
	DstX, DstY int16
}

func (r *TranslateCoordsReply) EncodeMessage(order binary.ByteOrder) []byte {
	reply := make([]byte, 32)
	reply[0] = 1 // Reply type
	reply[1] = BoolToByte(r.SameScreen)
	order.PutUint16(reply[2:4], r.Sequence)
	order.PutUint32(reply[4:8], 0) // Reply length (0 * 4 bytes = 0 bytes, plus 32 bytes header = 32 bytes total)
	order.PutUint32(reply[8:12], r.Child)
	order.PutUint16(reply[12:14], uint16(r.DstX))
	order.PutUint16(reply[14:16], uint16(r.DstY))
	// reply[16:32] is padding
	return reply
}

func ParseTranslateCoordsReply(order binary.ByteOrder, b []byte) (*TranslateCoordsReply, error) {
	if len(b) < 32 {
		return nil, NewError(LengthErrorCode, 0, 0, Opcodes{Major: 0, Minor: 0})
	}
	r := &TranslateCoordsReply{
		Sequence:   order.Uint16(b[2:4]),
		SameScreen: b[1] != 0,
		Child:      order.Uint32(b[8:12]),
		DstX:       int16(order.Uint16(b[12:14])),
		DstY:       int16(order.Uint16(b[14:16])),
	}
	return r, nil
}

// GetInputFocus: 43
type GetInputFocusReply struct {
	Sequence uint16
	RevertTo byte
	Focus    uint32
}

func (r *GetInputFocusReply) EncodeMessage(order binary.ByteOrder) []byte {
	reply := make([]byte, 32)
	reply[0] = 1 // Reply type
	reply[1] = r.RevertTo
	order.PutUint16(reply[2:4], r.Sequence)
	order.PutUint32(reply[4:8], 0) // Reply length (0 * 4 bytes = 0 bytes, plus 32 bytes header = 32 bytes total)
	order.PutUint32(reply[8:12], r.Focus)
	// reply[12:32] is padding
	return reply
}

func ParseGetInputFocusReply(order binary.ByteOrder, b []byte) (*GetInputFocusReply, error) {
	if len(b) < 32 {
		return nil, NewError(LengthErrorCode, 0, 0, Opcodes{Major: 0, Minor: 0})
	}
	r := &GetInputFocusReply{
		Sequence: order.Uint16(b[2:4]),
		RevertTo: b[1],
		Focus:    order.Uint32(b[8:12]),
	}
	return r, nil
}

// QueryFont: 47
type QueryFontReply struct {
	Sequence       uint16
	MinBounds      XCharInfo
	MaxBounds      XCharInfo
	MinCharOrByte2 uint16
	MaxCharOrByte2 uint16
	DefaultChar    uint16
	NumFontProps   uint16
	DrawDirection  uint8
	MinByte1       uint8
	MaxByte1       uint8
	AllCharsExist  bool
	FontAscent     int16
	FontDescent    int16
	NumCharInfos   uint32
	CharInfos      []XCharInfo
	FontProps      []FontProp
}

func (r *QueryFontReply) EncodeMessage(order binary.ByteOrder) []byte {
	numFontProps := len(r.FontProps)
	numCharInfos := len(r.CharInfos)

	reply := make([]byte, 60+8*numFontProps+12*numCharInfos)
	reply[0] = 1 // Reply
	reply[1] = 1 // font-info-present (True)
	order.PutUint16(reply[2:4], r.Sequence)
	order.PutUint32(reply[4:8], uint32(7+2*numFontProps+3*numCharInfos)) // Reply length

	// min-bounds
	order.PutUint16(reply[8:10], uint16(r.MinBounds.LeftSideBearing))
	order.PutUint16(reply[10:12], uint16(r.MinBounds.RightSideBearing))
	order.PutUint16(reply[12:14], uint16(r.MinBounds.CharacterWidth))
	order.PutUint16(reply[14:16], uint16(r.MinBounds.Ascent))
	order.PutUint16(reply[16:18], uint16(r.MinBounds.Descent))
	order.PutUint16(reply[18:20], r.MinBounds.Attributes)

	// max-bounds
	order.PutUint16(reply[24:26], uint16(r.MaxBounds.LeftSideBearing))
	order.PutUint16(reply[26:28], uint16(r.MaxBounds.RightSideBearing))
	order.PutUint16(reply[28:30], uint16(r.MaxBounds.CharacterWidth))
	order.PutUint16(reply[30:32], uint16(r.MaxBounds.Ascent))
	order.PutUint16(reply[32:34], uint16(r.MaxBounds.Descent))
	order.PutUint16(reply[34:36], r.MaxBounds.Attributes)

	order.PutUint16(reply[40:42], r.MinCharOrByte2)
	order.PutUint16(reply[42:44], r.MaxCharOrByte2)
	order.PutUint16(reply[44:46], r.DefaultChar)
	order.PutUint16(reply[46:48], uint16(numFontProps))

	reply[48] = r.DrawDirection & 0x1
	reply[49] = r.MinByte1
	reply[50] = r.MaxByte1
	reply[51] = BoolToByte(r.AllCharsExist)

	order.PutUint16(reply[52:54], uint16(r.FontAscent))
	order.PutUint16(reply[54:56], uint16(r.FontDescent))

	order.PutUint32(reply[56:60], uint32(len(r.CharInfos)))

	offset := 60 + 8*numFontProps
	for _, ci := range r.CharInfos {
		order.PutUint16(reply[offset:offset+2], uint16(ci.LeftSideBearing))
		order.PutUint16(reply[offset+2:offset+4], uint16(ci.RightSideBearing))
		order.PutUint16(reply[offset+4:offset+6], uint16(ci.CharacterWidth))
		order.PutUint16(reply[offset+6:offset+8], uint16(ci.Ascent))
		order.PutUint16(reply[offset+8:offset+10], uint16(ci.Descent))
		order.PutUint16(reply[offset+10:offset+12], ci.Attributes)
		offset += 12
	}
	return reply
}

func ParseQueryFontReply(order binary.ByteOrder, b []byte) (*QueryFontReply, error) {
	if len(b) < 60 {
		return nil, NewError(LengthErrorCode, 0, 0, Opcodes{Major: 0, Minor: 0})
	}
	numFontProps := order.Uint16(b[46:48])
	numCharInfos := order.Uint32(b[56:60])
	if len(b) < 60+8*int(numFontProps)+12*int(numCharInfos) {
		return nil, NewError(LengthErrorCode, 0, 0, Opcodes{Major: 0, Minor: 0})
	}
	var charInfos []XCharInfo
	if numCharInfos > 0 {
		charInfos = make([]XCharInfo, numCharInfos)
		offset := 60 + 8*int(numFontProps)
		for i := 0; i < int(numCharInfos); i++ {
			charInfos[i] = XCharInfo{
				LeftSideBearing:  int16(order.Uint16(b[offset:])),
				RightSideBearing: int16(order.Uint16(b[offset+2:])),
				CharacterWidth:   order.Uint16(b[offset+4:]),
				Ascent:           int16(order.Uint16(b[offset+6:])),
				Descent:          int16(order.Uint16(b[offset+8:])),
				Attributes:       order.Uint16(b[offset+10:]),
			}
			offset += 12
		}
	}

	var fontProps []FontProp
	if numFontProps > 0 {
		fontProps = make([]FontProp, numFontProps)
		offset := 60
		for i := 0; i < int(numFontProps); i++ {
			fontProps[i] = FontProp{
				Name:  order.Uint32(b[offset:]),
				Value: order.Uint32(b[offset+4:]),
			}
			offset += 8
		}
	}

	r := &QueryFontReply{
		Sequence: order.Uint16(b[2:4]),
		MinBounds: XCharInfo{
			LeftSideBearing:  int16(order.Uint16(b[8:10])),
			RightSideBearing: int16(order.Uint16(b[10:12])),
			CharacterWidth:   order.Uint16(b[12:14]),
			Ascent:           int16(order.Uint16(b[14:16])),
			Descent:          int16(order.Uint16(b[16:18])),
			Attributes:       order.Uint16(b[18:20]),
		},
		MaxBounds: XCharInfo{
			LeftSideBearing:  int16(order.Uint16(b[24:26])),
			RightSideBearing: int16(order.Uint16(b[26:28])),
			CharacterWidth:   order.Uint16(b[28:30]),
			Ascent:           int16(order.Uint16(b[30:32])),
			Descent:          int16(order.Uint16(b[32:34])),
			Attributes:       order.Uint16(b[34:36]),
		},
		MinCharOrByte2: order.Uint16(b[40:42]),
		MaxCharOrByte2: order.Uint16(b[42:44]),
		DefaultChar:    order.Uint16(b[44:46]),
		NumFontProps:   order.Uint16(b[46:48]),
		DrawDirection:  b[48],
		MinByte1:       b[49],
		MaxByte1:       b[50],
		AllCharsExist:  b[51] != 0,
		FontAscent:     int16(order.Uint16(b[52:54])),
		FontDescent:    int16(order.Uint16(b[54:56])),
		NumCharInfos:   numCharInfos,
		CharInfos:      charInfos,
		FontProps:      fontProps,
	}
	return r, nil
}

// ListFonts: 50
type ListFontsReply struct {
	Sequence  uint16
	FontNames []string
}

func (r *ListFontsReply) EncodeMessage(order binary.ByteOrder) []byte {
	var namesData []byte
	for _, name := range r.FontNames {
		namesData = append(namesData, byte(len(name)))
		namesData = append(namesData, []byte(name)...)
	}

	namesSize := len(namesData)
	padSize := (4 - (namesSize % 4)) % 4

	reply := make([]byte, 32+namesSize+padSize)
	reply[0] = 1 // Reply
	// byte 1 is unused
	order.PutUint16(reply[2:4], r.Sequence)
	order.PutUint32(reply[4:8], uint32((namesSize+padSize)/4)) // Reply length
	order.PutUint16(reply[8:10], uint16(len(r.FontNames)))
	// reply[10:32] is padding
	copy(reply[32:], namesData)
	return reply
}

func ParseListFontsReply(order binary.ByteOrder, b []byte) (*ListFontsReply, error) {
	if len(b) < 32 {
		return nil, NewError(LengthErrorCode, 0, 0, Opcodes{Major: 0, Minor: 0})
	}
	numFonts := order.Uint16(b[8:10])
	fontNames := make([]string, numFonts)
	offset := 32
	for i := 0; i < int(numFonts); i++ {
		if len(b) < offset+1 {
			return nil, NewError(LengthErrorCode, 0, 0, Opcodes{Major: 0, Minor: 0})
		}
		length := int(b[offset])
		if len(b) < offset+1+length {
			return nil, NewError(LengthErrorCode, 0, 0, Opcodes{Major: 0, Minor: 0})
		}
		fontNames[i] = string(b[offset+1 : offset+1+length])
		offset += 1 + length
	}
	r := &ListFontsReply{
		Sequence:  order.Uint16(b[2:4]),
		FontNames: fontNames,
	}
	return r, nil
}

// GetImage: 73
type GetImageReply struct {
	Sequence  uint16
	Depth     byte
	VisualID  uint32
	ImageData []byte
}

func (r *GetImageReply) EncodeMessage(order binary.ByteOrder) []byte {
	reply := make([]byte, 32+len(r.ImageData))
	reply[0] = 1 // Reply type
	reply[1] = r.Depth
	order.PutUint16(reply[2:4], r.Sequence)
	order.PutUint32(reply[4:8], uint32(len(r.ImageData)/4)) // Reply length
	order.PutUint32(reply[8:12], r.VisualID)
	// reply[12:32] is padding
	copy(reply[32:], r.ImageData)
	return reply
}

func ParseGetImageReply(order binary.ByteOrder, b []byte) (*GetImageReply, error) {
	if len(b) < 32 {
		return nil, NewError(LengthErrorCode, 0, 0, Opcodes{Major: 0, Minor: 0})
	}
	length := order.Uint32(b[4:8]) * 4
	if len(b) < 32+int(length) {
		return nil, NewError(LengthErrorCode, 0, 0, Opcodes{Major: 0, Minor: 0})
	}
	r := &GetImageReply{
		Sequence:  order.Uint16(b[2:4]),
		Depth:     b[1],
		VisualID:  order.Uint32(b[8:12]),
		ImageData: b[32 : 32+length],
	}
	return r, nil
}

// AllocColor: 84
type AllocColorReply struct {
	Sequence uint16
	Red      uint16
	Green    uint16
	Blue     uint16
	Pixel    uint32
}

/*
1     1                               Reply
1                                     unused
2     CARD16                          sequence number
4     0                               reply length
2     CARD16                          red
2     CARD16                          green
2     CARD16                          blue
2                                     unused
4     CARD32                          pixel
12                                    unused
*/
func (r *AllocColorReply) EncodeMessage(order binary.ByteOrder) []byte {
	reply := make([]byte, 32)
	reply[0] = 1 // Reply type
	// byte 1 is unused
	order.PutUint16(reply[2:4], r.Sequence)
	order.PutUint32(reply[4:8], 0) // Reply length (0 * 4 bytes = 0 bytes, plus 32 bytes header = 32 bytes total)
	order.PutUint16(reply[8:10], r.Red)
	order.PutUint16(reply[10:12], r.Green)
	order.PutUint16(reply[12:14], r.Blue)
	// reply[14:16] is padding
	order.PutUint32(reply[16:20], r.Pixel)
	// reply[20:32] is padding
	return reply
}

func ParseAllocColorReply(order binary.ByteOrder, b []byte) (*AllocColorReply, error) {
	if len(b) < 32 {
		return nil, NewError(LengthErrorCode, 0, 0, Opcodes{Major: 0, Minor: 0})
	}
	r := &AllocColorReply{
		Sequence: order.Uint16(b[2:4]),
		Red:      order.Uint16(b[8:10]),
		Green:    order.Uint16(b[10:12]),
		Blue:     order.Uint16(b[12:14]),
		Pixel:    order.Uint32(b[16:20]),
	}
	return r, nil
}

// AllocNamedColor: 85
type AllocNamedColorReply struct {
	Sequence   uint16
	Red        uint16
	Green      uint16
	Blue       uint16
	ExactRed   uint16
	ExactGreen uint16
	ExactBlue  uint16
	Pixel      uint32
}

/*
1     1                               Reply
1                                     unused
2     CARD16                          sequence number
4     0                               reply length
4     CARD32                          pixel
2     CARD16                          exact-red
2     CARD16                          exact-green
2     CARD16                          exact-blue
2     CARD16                          visual-red
2     CARD16                          visual-green
2     CARD16                          visual-blue
8                                     unused
*/
func (r *AllocNamedColorReply) EncodeMessage(order binary.ByteOrder) []byte {
	reply := make([]byte, 32)
	reply[0] = 1 // Reply type
	// byte 1 is unused
	order.PutUint16(reply[2:4], r.Sequence)
	order.PutUint32(reply[4:8], 0) // Reply length (0 * 4 bytes = 0 bytes, plus 32 bytes header = 32 bytes total)
	order.PutUint32(reply[8:12], r.Pixel)
	order.PutUint16(reply[12:14], r.ExactRed)
	order.PutUint16(reply[14:16], r.ExactGreen)
	order.PutUint16(reply[16:18], r.ExactBlue)
	order.PutUint16(reply[18:20], r.Red)
	order.PutUint16(reply[20:22], r.Green)
	order.PutUint16(reply[22:24], r.Blue)
	return reply
}

func ParseAllocNamedColorReply(order binary.ByteOrder, b []byte) (*AllocNamedColorReply, error) {
	if len(b) < 32 {
		return nil, NewError(LengthErrorCode, 0, 0, Opcodes{Major: 0, Minor: 0})
	}
	r := &AllocNamedColorReply{
		Sequence:   order.Uint16(b[2:4]),
		Pixel:      order.Uint32(b[8:12]),
		ExactRed:   order.Uint16(b[12:14]),
		ExactGreen: order.Uint16(b[14:16]),
		ExactBlue:  order.Uint16(b[16:18]),
		Red:        order.Uint16(b[18:20]),
		Green:      order.Uint16(b[20:22]),
		Blue:       order.Uint16(b[22:24]),
	}
	return r, nil
}

// ListInstalledColormaps: 85
type ListInstalledColormapsReply struct {
	Sequence     uint16
	NumColormaps uint16
	Colormaps    []uint32
}

func (r *ListInstalledColormapsReply) EncodeMessage(order binary.ByteOrder) []byte {
	nColormaps := len(r.Colormaps)
	reply := make([]byte, 32+nColormaps*4)
	reply[0] = 1 // Reply
	// byte 1 is unused
	order.PutUint16(reply[2:4], r.Sequence)
	order.PutUint32(reply[4:8], uint32(nColormaps)) // length
	order.PutUint16(reply[8:10], uint16(nColormaps))
	// reply[10:32] is padding
	for i, cmap := range r.Colormaps {
		order.PutUint32(reply[32+i*4:], cmap)
	}
	return reply
}

func ParseListInstalledColormapsReply(order binary.ByteOrder, b []byte) (*ListInstalledColormapsReply, error) {
	if len(b) < 32 {
		return nil, NewError(LengthErrorCode, 0, 0, Opcodes{Major: 0, Minor: 0})
	}
	numColormaps := order.Uint16(b[8:10])
	if len(b) < 32+int(numColormaps)*4 {
		return nil, NewError(LengthErrorCode, 0, 0, Opcodes{Major: 0, Minor: 0})
	}
	colormaps := make([]uint32, numColormaps)
	for i := 0; i < int(numColormaps); i++ {
		colormaps[i] = order.Uint32(b[32+i*4:])
	}
	r := &ListInstalledColormapsReply{
		Sequence:     order.Uint16(b[2:4]),
		NumColormaps: numColormaps,
		Colormaps:    colormaps,
	}
	return r, nil
}

// QueryColors: 91
type QueryColorsReply struct {
	Sequence uint16
	Colors   []XColorItem
}

func (r *QueryColorsReply) EncodeMessage(order binary.ByteOrder) []byte {
	numColors := len(r.Colors)
	replies := make([]byte, numColors*8)
	for i, color := range r.Colors {
		order.PutUint16(replies[i*8:], color.Red)
		order.PutUint16(replies[i*8+2:], color.Green)
		order.PutUint16(replies[i*8+4:], color.Blue)
		// replies[i*8+6:i*8+8] unused
	}

	reply := make([]byte, 32+len(replies))
	reply[0] = 1 // Reply
	// byte 1 is unused
	order.PutUint16(reply[2:4], r.Sequence)
	order.PutUint32(reply[4:8], uint32(len(replies)/4)) // Reply length
	order.PutUint16(reply[8:10], uint16(numColors))
	// reply[10:32] is padding
	copy(reply[32:], replies)
	return reply
}

func ParseQueryColorsReply(order binary.ByteOrder, b []byte) (*QueryColorsReply, error) {
	if len(b) < 32 {
		return nil, NewError(LengthErrorCode, 0, 0, Opcodes{Major: 0, Minor: 0})
	}
	numColors := order.Uint16(b[8:10])
	if len(b) < 32+int(numColors)*8 {
		return nil, NewError(LengthErrorCode, 0, 0, Opcodes{Major: 0, Minor: 0})
	}
	colors := make([]XColorItem, numColors)
	for i := 0; i < int(numColors); i++ {
		colors[i] = XColorItem{
			Red:   order.Uint16(b[32+i*8:]),
			Green: order.Uint16(b[32+i*8+2:]),
			Blue:  order.Uint16(b[32+i*8+4:]),
		}
	}
	r := &QueryColorsReply{
		Sequence: order.Uint16(b[2:4]),
		Colors:   colors,
	}
	return r, nil
}

// LookupColor: 92
type LookupColorReply struct {
	Sequence   uint16
	Red        uint16
	Green      uint16
	Blue       uint16
	ExactRed   uint16
	ExactGreen uint16
	ExactBlue  uint16
}

func (r *LookupColorReply) EncodeMessage(order binary.ByteOrder) []byte {
	reply := make([]byte, 32)
	reply[0] = 1 // Reply type
	// byte 1 is unused
	order.PutUint16(reply[2:4], r.Sequence)
	order.PutUint32(reply[4:8], 0) // Reply length (0 * 4 bytes = 0 bytes, plus 32 bytes header = 32 bytes total)
	order.PutUint16(reply[8:10], r.Red)
	order.PutUint16(reply[10:12], r.Green)
	order.PutUint16(reply[12:14], r.Blue)
	order.PutUint16(reply[14:16], r.ExactRed)
	order.PutUint16(reply[16:18], r.ExactGreen)
	order.PutUint16(reply[18:20], r.ExactBlue)
	// reply[20:32] is padding
	return reply
}

func ParseLookupColorReply(order binary.ByteOrder, b []byte) (*LookupColorReply, error) {
	if len(b) < 32 {
		return nil, NewError(LengthErrorCode, 0, 0, Opcodes{Major: 0, Minor: 0})
	}
	r := &LookupColorReply{
		Sequence:   order.Uint16(b[2:4]),
		Red:        order.Uint16(b[8:10]),
		Green:      order.Uint16(b[10:12]),
		Blue:       order.Uint16(b[12:14]),
		ExactRed:   order.Uint16(b[14:16]),
		ExactGreen: order.Uint16(b[16:18]),
		ExactBlue:  order.Uint16(b[18:20]),
	}
	return r, nil
}

// QueryBestSize: 97
type QueryBestSizeReply struct {
	Sequence uint16
	Width    uint16
	Height   uint16
}

func (r *QueryBestSizeReply) EncodeMessage(order binary.ByteOrder) []byte {
	reply := make([]byte, 32)
	reply[0] = 1 // Reply type
	// byte 1 is unused
	order.PutUint16(reply[2:4], r.Sequence)
	order.PutUint32(reply[4:8], 0) // Reply length (0 * 4 bytes = 0 bytes, plus 32 bytes header = 32 bytes total)
	order.PutUint16(reply[8:10], r.Width)
	order.PutUint16(reply[10:12], r.Height)
	// reply[12:32] is padding
	return reply
}

func ParseQueryBestSizeReply(order binary.ByteOrder, b []byte) (*QueryBestSizeReply, error) {
	if len(b) < 32 {
		return nil, NewError(LengthErrorCode, 0, 0, Opcodes{Major: 0, Minor: 0})
	}
	r := &QueryBestSizeReply{
		Sequence: order.Uint16(b[2:4]),
		Width:    order.Uint16(b[8:10]),
		Height:   order.Uint16(b[10:12]),
	}
	return r, nil
}

// QueryExtension: 98
type QueryExtensionReply struct {
	Sequence    uint16
	Present     bool
	MajorOpcode byte
	FirstEvent  byte
	FirstError  byte
}

// 1     1                               Reply
// 1                                     unused
// 2     CARD16                          sequence number
// 4     0                               reply length
// 1     BOOL                            present
// 1     CARD8                           major-opcode
// 1     CARD8                           first-event
// 1     CARD8                           first-error
// 20                                    unused
func (r *QueryExtensionReply) EncodeMessage(order binary.ByteOrder) []byte {
	reply := make([]byte, 32)
	reply[0] = 1 // Reply type
	order.PutUint16(reply[2:4], r.Sequence)
	order.PutUint32(reply[4:8], 0) // Reply length (0 * 4 bytes = 0 bytes, plus 32 bytes header = 32 bytes total)
	reply[8] = BoolToByte(r.Present)
	reply[9] = r.MajorOpcode
	reply[10] = r.FirstEvent
	reply[11] = r.FirstError
	// reply[12:32] is padding
	return reply
}

func ParseQueryExtensionReply(order binary.ByteOrder, b []byte) (*QueryExtensionReply, error) {
	if len(b) < 32 {
		return nil, NewError(LengthErrorCode, 0, 0, Opcodes{Major: 0, Minor: 0})
	}
	r := &QueryExtensionReply{
		Sequence:    order.Uint16(b[2:4]),
		Present:     b[8] != 0,
		MajorOpcode: b[9],
		FirstEvent:  b[10],
		FirstError:  b[11],
	}
	return r, nil
}

// SetupResponse implements messageEncoder for the X11 setup response.
type SetupResponse struct {
	Success                  byte
	Reason                   string
	ProtocolVersion          uint16
	ReleaseNumber            uint32
	ResourceIDBase           uint32
	ResourceIDMask           uint32
	MotionBufferSize         uint32
	VendorLength             uint16
	MaxRequestLength         uint16
	NumScreens               uint8
	NumPixmapFormats         uint8
	ImageByteOrder           uint8
	BitmapFormatBitOrder     byte
	BitmapFormatScanlineUnit byte
	BitmapFormatScanlinePad  byte
	MinKeycode               uint8
	MaxKeycode               uint8
	VendorString             string
	PixmapFormats            []Format
	Screens                  []Screen
}

func (r *SetupResponse) EncodeMessage(order binary.ByteOrder) []byte {
	if r.Success == 0 {
		response := make([]byte, 8+len(r.Reason))
		order.PutUint16(response[2:4], r.ProtocolVersion)
		order.PutUint16(response[4:6], 0) // length of additional data in 4-byte units
		order.PutUint16(response[6:8], uint16(len(r.Reason)/4))
		copy(response[8:], []byte(r.Reason))
		return response
	}
	setup := NewDefaultSetup() // This should probably be passed in or generated once
	setupData := setup.marshal(order)

	response := make([]byte, 8+len(setupData))
	response[0] = r.Success
	// byte 1 is unused
	order.PutUint16(response[2:4], r.ProtocolVersion)
	order.PutUint16(response[4:6], 0) // length of additional data in 4-byte units
	order.PutUint16(response[6:8], uint16(len(setupData)/4))
	copy(response[8:], setupData)
	return response
}

// Start of setup struct
type Setup struct {
	ReleaseNumber            uint32
	ResourceIDBase           uint32
	ResourceIDMask           uint32
	MotionBufferSize         uint32
	VendorLength             uint16
	MaxRequestLength         uint16
	NumScreens               uint8
	NumPixmapFormats         uint8
	ImageByteOrder           uint8
	BitmapFormatBitOrder     uint8
	BitmapFormatScanlineUnit uint8
	BitmapFormatScanlinePad  uint8
	MinKeycode               uint8
	MaxKeycode               uint8
	VendorString             string
	PixmapFormats            []Format
	Screens                  []Screen
}

type Format struct {
	Depth        uint8
	BitsPerPixel uint8
	ScanlinePad  uint8
}

type Screen struct {
	Root                uint32
	DefaultColormap     uint32
	WhitePixel          uint32
	BlackPixel          uint32
	CurrentInputMasks   uint32
	WidthInPixels       uint16
	HeightInPixels      uint16
	WidthInMillimeters  uint16
	HeightInMillimeters uint16
	MinInstalledMaps    uint16
	MaxInstalledMaps    uint16
	RootVisual          uint32
	BackingStores       uint8
	SaveUnders          bool
	RootDepth           uint8
	NumDepths           uint8
	Depths              []Depth
}

type Depth struct {
	Depth      uint8
	NumVisuals uint16
	Visuals    []VisualType
}

type VisualType struct {
	VisualID        uint32 // visual-id
	Class           uint8
	BitsPerRGBValue uint8
	ColormapEntries uint16
	RedMask         uint32
	GreenMask       uint32
	BlueMask        uint32
}

func NewDefaultSetup() *Setup {
	vendorString := "sshterm"
	s := &Setup{
		ReleaseNumber:            1,
		ResourceIDBase:           0,
		ResourceIDMask:           0x1FFFFF,
		MotionBufferSize:         256,
		VendorLength:             uint16(len(vendorString)),
		MaxRequestLength:         0xFFFF,
		NumScreens:               1,
		NumPixmapFormats:         1,
		ImageByteOrder:           0, // LSBFirst
		BitmapFormatBitOrder:     0, // LeastSignificant
		BitmapFormatScanlineUnit: 8,
		BitmapFormatScanlinePad:  8,
		MinKeycode:               8,
		MaxKeycode:               255,
		VendorString:             vendorString,
		PixmapFormats: []Format{
			{
				Depth:        24,
				BitsPerPixel: 32,
				ScanlinePad:  32,
			},
		},
		Screens: []Screen{
			{
				Root:                0,
				DefaultColormap:     1,
				WhitePixel:          0xffffff,
				BlackPixel:          0x000000,
				CurrentInputMasks:   0,
				WidthInPixels:       1024,
				HeightInPixels:      768,
				WidthInMillimeters:  270,
				HeightInMillimeters: 203,
				MinInstalledMaps:    1,
				MaxInstalledMaps:    1,
				RootVisual:          0x1,
				BackingStores:       2, // Always
				SaveUnders:          false,
				RootDepth:           24,
				NumDepths:           1,
				Depths: []Depth{
					{
						Depth:      24,
						NumVisuals: 1,
						Visuals: []VisualType{
							{
								VisualID:        0x1,
								Class:           4, // TrueColor
								BitsPerRGBValue: 8,
								ColormapEntries: 256,
								RedMask:         0xff0000,
								GreenMask:       0x00ff00,
								BlueMask:        0x0000ff,
							},
						},
					},
				},
			},
		},
	}
	return s
}

func (s *Setup) marshal(order binary.ByteOrder) []byte {
	buf := new(bytes.Buffer)
	binary.Write(buf, order, s.ReleaseNumber)
	binary.Write(buf, order, s.ResourceIDBase)
	binary.Write(buf, order, s.ResourceIDMask)
	binary.Write(buf, order, s.MotionBufferSize)
	binary.Write(buf, order, s.VendorLength)
	binary.Write(buf, order, s.MaxRequestLength)
	buf.WriteByte(s.NumScreens)
	buf.WriteByte(s.NumPixmapFormats)
	buf.WriteByte(s.ImageByteOrder)
	buf.WriteByte(s.BitmapFormatBitOrder)
	buf.WriteByte(s.BitmapFormatScanlineUnit)
	buf.WriteByte(s.BitmapFormatScanlinePad)
	buf.WriteByte(s.MinKeycode)
	buf.WriteByte(s.MaxKeycode)
	buf.Write([]byte{0, 0, 0, 0}) // 4 bytes of padding
	buf.WriteString(s.VendorString)
	if pad := (4 - (len(s.VendorString) % 4)) % 4; pad > 0 {
		buf.Write(make([]byte, pad))
	}
	for _, f := range s.PixmapFormats {
		f.marshal(buf, order)
	}
	for _, scr := range s.Screens {
		scr.marshal(buf, order)
	}
	return buf.Bytes()
}

func (f *Format) marshal(buf *bytes.Buffer, order binary.ByteOrder) {
	buf.WriteByte(f.Depth)
	buf.WriteByte(f.BitsPerPixel)
	buf.WriteByte(f.ScanlinePad)
	buf.Write([]byte{0, 0, 0, 0, 0}) // 5 bytes of padding
}

func (s *Screen) marshal(buf *bytes.Buffer, order binary.ByteOrder) {
	binary.Write(buf, order, s.Root)
	binary.Write(buf, order, s.DefaultColormap)
	binary.Write(buf, order, s.WhitePixel)
	binary.Write(buf, order, s.BlackPixel)
	binary.Write(buf, order, s.CurrentInputMasks)
	binary.Write(buf, order, s.WidthInPixels)
	binary.Write(buf, order, s.HeightInPixels)
	binary.Write(buf, order, s.WidthInMillimeters)
	binary.Write(buf, order, s.HeightInMillimeters)
	binary.Write(buf, order, s.MinInstalledMaps)
	binary.Write(buf, order, s.MaxInstalledMaps)
	binary.Write(buf, order, s.RootVisual)
	buf.WriteByte(s.BackingStores)
	if s.SaveUnders {
		buf.WriteByte(1)
	} else {
		buf.WriteByte(0)
	}
	buf.WriteByte(s.RootDepth)
	buf.WriteByte(s.NumDepths)
	for _, d := range s.Depths {
		d.marshal(buf, order)
	}
}

func (d *Depth) marshal(buf *bytes.Buffer, order binary.ByteOrder) {
	buf.WriteByte(d.Depth)
	buf.WriteByte(0) // padding
	binary.Write(buf, order, d.NumVisuals)
	buf.Write([]byte{0, 0, 0, 0}) // 4 bytes of padding
	for _, v := range d.Visuals {
		v.marshal(buf, order)
	}
}

func (v *VisualType) marshal(buf *bytes.Buffer, order binary.ByteOrder) {
	binary.Write(buf, order, v.VisualID)
	buf.WriteByte(v.Class)
	buf.WriteByte(v.BitsPerRGBValue)
	binary.Write(buf, order, v.ColormapEntries)
	binary.Write(buf, order, v.RedMask)
	binary.Write(buf, order, v.GreenMask)
	binary.Write(buf, order, v.BlueMask)
	buf.Write([]byte{0, 0, 0, 0}) // 4 bytes of padding
}

// SetPointerMapping: 116
type SetPointerMappingReply struct {
	Sequence uint16
	Status   byte
}

func (r *SetPointerMappingReply) EncodeMessage(order binary.ByteOrder) []byte {
	reply := make([]byte, 32)
	reply[0] = 1 // Reply type
	reply[1] = r.Status
	order.PutUint16(reply[2:4], r.Sequence)
	order.PutUint32(reply[4:8], 0)
	return reply
}

func ParseSetPointerMappingReply(order binary.ByteOrder, b []byte) (*SetPointerMappingReply, error) {
	if len(b) < 32 {
		return nil, NewError(LengthErrorCode, 0, 0, Opcodes{Major: 0, Minor: 0})
	}
	r := &SetPointerMappingReply{
		Sequence: order.Uint16(b[2:4]),
		Status:   b[1],
	}
	return r, nil
}

// GetPointerMapping: 117
type GetPointerMappingReply struct {
	Sequence uint16
	Length   byte
	PMap     []byte
}

func (r *GetPointerMappingReply) EncodeMessage(order binary.ByteOrder) []byte {
	reply := make([]byte, 32+len(r.PMap))
	reply[0] = 1 // Reply type
	reply[1] = r.Length
	order.PutUint16(reply[2:4], r.Sequence)
	order.PutUint32(reply[4:8], uint32((len(r.PMap)+3)/4))
	copy(reply[32:], r.PMap)
	return reply
}

func ParseGetPointerMappingReply(order binary.ByteOrder, b []byte) (*GetPointerMappingReply, error) {
	if len(b) < 32 {
		return nil, NewError(LengthErrorCode, 0, 0, Opcodes{Major: 0, Minor: 0})
	}
	length := b[1]
	if len(b) < 32+int(length) {
		return nil, NewError(LengthErrorCode, 0, 0, Opcodes{Major: 0, Minor: 0})
	}
	r := &GetPointerMappingReply{
		Sequence: order.Uint16(b[2:4]),
		Length:   length,
		PMap:     b[32 : 32+length],
	}
	return r, nil
}

// GetKeyboardMapping: 101
type GetKeyboardMappingReply struct {
	Sequence          uint16
	KeySymsPerKeycode byte
	KeySyms           []uint32
}

func (r *GetKeyboardMappingReply) OpCode() ReqCode { return GetKeyboardMapping }

func (r *GetKeyboardMappingReply) EncodeMessage(order binary.ByteOrder) []byte {
	numKeysyms := len(r.KeySyms)
	length := uint32(numKeysyms)

	reply := make([]byte, 32+numKeysyms*4)
	reply[0] = 1 // Reply type
	reply[1] = 1
	order.PutUint16(reply[2:4], r.Sequence)
	order.PutUint32(reply[4:8], length)
	// bytes 8-31 are unused
	for i, keySym := range r.KeySyms {
		order.PutUint32(reply[32+i*4:], keySym)
	}
	return reply
}

func ParseGetKeyboardMappingReply(order binary.ByteOrder, b []byte) (*GetKeyboardMappingReply, error) {
	if len(b) < 32 {
		return nil, NewError(LengthErrorCode, 0, 0, Opcodes{Major: 0, Minor: 0})
	}
	length := order.Uint32(b[4:8])
	if len(b) < 32+int(length)*4 {
		return nil, NewError(LengthErrorCode, 0, 0, Opcodes{Major: 0, Minor: 0})
	}
	keySyms := make([]uint32, length)
	for i := 0; i < int(length); i++ {
		keySyms[i] = order.Uint32(b[32+i*4:])
	}
	r := &GetKeyboardMappingReply{
		Sequence:          order.Uint16(b[2:4]),
		KeySymsPerKeycode: 1,
		KeySyms:           keySyms,
	}
	return r, nil
}

// GetKeyboardControl: 103
type GetKeyboardControlReply struct {
	Sequence         uint16
	KeyClickPercent  byte
	BellPercent      byte
	BellPitch        uint16
	BellDuration     uint16
	LedMask          uint32
	GlobalAutoRepeat byte
	AutoRepeats      [32]byte
}

func (r *GetKeyboardControlReply) EncodeMessage(order binary.ByteOrder) []byte {
	reply := make([]byte, 52)
	reply[0] = 1 // Reply type
	reply[1] = r.GlobalAutoRepeat
	order.PutUint16(reply[2:4], r.Sequence)
	order.PutUint32(reply[4:8], 5) // Reply length
	order.PutUint32(reply[8:12], r.LedMask)
	reply[12] = r.KeyClickPercent
	reply[13] = r.BellPercent
	order.PutUint16(reply[14:16], r.BellPitch)
	order.PutUint16(reply[16:18], r.BellDuration)
	// reply[18:20] is padding
	copy(reply[20:52], r.AutoRepeats[:])
	return reply
}

func ParseGetKeyboardControlReply(order binary.ByteOrder, b []byte) (*GetKeyboardControlReply, error) {
	if len(b) < 52 {
		return nil, NewError(LengthErrorCode, 0, 0, Opcodes{Major: 0, Minor: 0})
	}
	r := &GetKeyboardControlReply{
		Sequence:         order.Uint16(b[2:4]),
		GlobalAutoRepeat: b[1],
		LedMask:          order.Uint32(b[8:12]),
		KeyClickPercent:  b[12],
		BellPercent:      b[13],
		BellPitch:        order.Uint16(b[14:16]),
		BellDuration:     order.Uint16(b[16:18]),
	}
	copy(r.AutoRepeats[:], b[20:52])
	return r, nil
}

// GetScreenSaver: 108
type GetScreenSaverReply struct {
	Sequence    uint16
	Timeout     uint16
	Interval    uint16
	PreferBlank byte
	AllowExpose byte
}

func (r *GetScreenSaverReply) EncodeMessage(order binary.ByteOrder) []byte {
	reply := make([]byte, 32)
	reply[0] = 1 // Reply type
	order.PutUint16(reply[2:4], r.Sequence)
	order.PutUint32(reply[4:8], 0)
	order.PutUint16(reply[8:10], r.Timeout)
	order.PutUint16(reply[10:12], r.Interval)
	reply[12] = r.PreferBlank
	reply[13] = r.AllowExpose
	return reply
}

func ParseGetScreenSaverReply(order binary.ByteOrder, b []byte) (*GetScreenSaverReply, error) {
	if len(b) < 32 {
		return nil, NewError(LengthErrorCode, 0, 0, Opcodes{Major: 0, Minor: 0})
	}
	r := &GetScreenSaverReply{
		Sequence:    order.Uint16(b[2:4]),
		Timeout:     order.Uint16(b[8:10]),
		Interval:    order.Uint16(b[10:12]),
		PreferBlank: b[12],
		AllowExpose: b[13],
	}
	return r, nil
}

// ListHosts: 110
type ListHostsReply struct {
	Sequence uint16
	NumHosts uint16
	Hosts    []Host
}

func (r *ListHostsReply) EncodeMessage(order binary.ByteOrder) []byte {
	var data []byte
	for _, host := range r.Hosts {
		var hostData []byte
		hostData = append(hostData, host.Family)
		hostData = append(hostData, 0) // padding
		hostData = append(hostData, make([]byte, 2)...)
		order.PutUint16(hostData[2:4], uint16(len(host.Data)))
		hostData = append(hostData, host.Data...)
		pad := (4 - (len(host.Data) % 4)) % 4
		hostData = append(hostData, make([]byte, pad)...)
		data = append(data, hostData...)
	}
	reply := make([]byte, 32+len(data))
	reply[0] = 1 // Reply type
	order.PutUint16(reply[2:4], r.Sequence)
	order.PutUint32(reply[4:8], uint32(len(data)/4))
	order.PutUint16(reply[8:10], r.NumHosts)
	copy(reply[32:], data)
	return reply
}

func ParseListHostsReply(order binary.ByteOrder, b []byte) (*ListHostsReply, error) {
	if len(b) < 32 {
		return nil, NewError(LengthErrorCode, 0, 0, Opcodes{Major: 0, Minor: 0})
	}
	numHosts := order.Uint16(b[8:10])
	hosts := make([]Host, numHosts)
	offset := 32
	for i := 0; i < int(numHosts); i++ {
		if len(b) < offset+4 {
			return nil, NewError(LengthErrorCode, 0, 0, Opcodes{Major: 0, Minor: 0})
		}
		family := b[offset]
		length := int(order.Uint16(b[offset+2 : offset+4]))
		if len(b) < offset+4+length {
			return nil, NewError(LengthErrorCode, 0, 0, Opcodes{Major: 0, Minor: 0})
		}
		data := b[offset+4 : offset+4+length]
		hosts[i] = Host{
			Family: family,
			Data:   data,
		}
		offset += 4 + length + PadLen(length)
	}
	r := &ListHostsReply{
		Sequence: order.Uint16(b[2:4]),
		NumHosts: numHosts,
		Hosts:    hosts,
	}
	return r, nil
}

// SetModifierMapping: 118
type SetModifierMappingReply struct {
	Sequence uint16
	Status   byte
}

func (r *SetModifierMappingReply) EncodeMessage(order binary.ByteOrder) []byte {
	reply := make([]byte, 32)
	reply[0] = 1 // Reply type
	reply[1] = r.Status
	order.PutUint16(reply[2:4], r.Sequence)
	order.PutUint32(reply[4:8], 0)
	return reply
}

func ParseSetModifierMappingReply(order binary.ByteOrder, b []byte) (*SetModifierMappingReply, error) {
	if len(b) < 32 {
		return nil, NewError(LengthErrorCode, 0, 0, Opcodes{Major: 0, Minor: 0})
	}
	r := &SetModifierMappingReply{
		Sequence: order.Uint16(b[2:4]),
		Status:   b[1],
	}
	return r, nil
}

// GetModifierMapping: 119
type GetModifierMappingReply struct {
	Sequence            uint16
	KeyCodesPerModifier byte
	KeyCodes            []KeyCode
}

func (r *GetModifierMappingReply) EncodeMessage(order binary.ByteOrder) []byte {
	keyCodes := make([]byte, len(r.KeyCodes))
	for i, kc := range r.KeyCodes {
		keyCodes[i] = byte(kc)
	}
	reply := make([]byte, 32+len(keyCodes))
	reply[0] = 1 // Reply type
	reply[1] = r.KeyCodesPerModifier
	order.PutUint16(reply[2:4], r.Sequence)
	order.PutUint32(reply[4:8], uint32((len(keyCodes)+3)/4))
	copy(reply[32:], keyCodes)
	return reply
}

func ParseGetModifierMappingReply(order binary.ByteOrder, b []byte) (*GetModifierMappingReply, error) {
	if len(b) < 32 {
		return nil, NewError(LengthErrorCode, 0, 0, Opcodes{Major: 0, Minor: 0})
	}
	keyCodesPerModifier := b[1]
	if len(b) < 32+int(keyCodesPerModifier)*8 {
		return nil, NewError(LengthErrorCode, 0, 0, Opcodes{Major: 0, Minor: 0})
	}
	keyCodes := make([]KeyCode, int(keyCodesPerModifier)*8)
	for i := 0; i < len(keyCodes); i++ {
		keyCodes[i] = KeyCode(b[32+i])
	}
	r := &GetModifierMappingReply{
		Sequence:            order.Uint16(b[2:4]),
		KeyCodesPerModifier: keyCodesPerModifier,
		KeyCodes:            keyCodes,
	}
	return r, nil
}

// QueryKeymap: 44
type QueryKeymapReply struct {
	Sequence uint16
	Keys     [32]byte
}

func (r *QueryKeymapReply) EncodeMessage(order binary.ByteOrder) []byte {
	reply := make([]byte, 40)
	reply[0] = 1 // Reply type
	order.PutUint16(reply[2:4], r.Sequence)
	order.PutUint32(reply[4:8], 2)
	copy(reply[8:], r.Keys[:])
	return reply
}

func ParseQueryKeymapReply(order binary.ByteOrder, b []byte) (*QueryKeymapReply, error) {
	if len(b) < 40 {
		return nil, NewError(LengthErrorCode, 0, 0, Opcodes{Major: 0, Minor: 0})
	}
	r := &QueryKeymapReply{
		Sequence: order.Uint16(b[2:4]),
	}
	copy(r.Keys[:], b[8:40])
	return r, nil
}

// GetFontPath: 52
type GetFontPathReply struct {
	Sequence uint16
	NPaths   uint16
	Paths    []string
}

func (r *GetFontPathReply) EncodeMessage(order binary.ByteOrder) []byte {
	var data []byte
	for _, path := range r.Paths {
		data = append(data, byte(len(path)))
		data = append(data, path...)
	}
	p := (4 - (len(data) % 4)) % 4
	totalLen := 32 + len(data) + p

	reply := make([]byte, totalLen)
	reply[0] = 1 // Reply type
	order.PutUint16(reply[2:4], r.Sequence)
	order.PutUint32(reply[4:8], uint32((len(data)+p)/4))
	order.PutUint16(reply[8:10], r.NPaths)
	copy(reply[32:], data)
	return reply
}

func ParseGetFontPathReply(order binary.ByteOrder, b []byte) (*GetFontPathReply, error) {
	if len(b) < 32 {
		return nil, NewError(LengthErrorCode, 0, 0, Opcodes{Major: 0, Minor: 0})
	}
	nPaths := order.Uint16(b[8:10])
	paths := make([]string, nPaths)
	offset := 32
	for i := 0; i < int(nPaths); i++ {
		if len(b) < offset+1 {
			return nil, NewError(LengthErrorCode, 0, 0, Opcodes{Major: 0, Minor: 0})
		}
		length := int(b[offset])
		if len(b) < offset+1+length {
			return nil, NewError(LengthErrorCode, 0, 0, Opcodes{Major: 0, Minor: 0})
		}
		paths[i] = string(b[offset+1 : offset+1+length])
		offset += 1 + length
	}
	r := &GetFontPathReply{
		Sequence: order.Uint16(b[2:4]),
		NPaths:   nPaths,
		Paths:    paths,
	}
	return r, nil
}

// ListFontsWithInfo: 50
type ListFontsWithInfoReply struct {
	Sequence      uint16
	NameLength    byte
	MinBounds     XCharInfo
	MaxBounds     XCharInfo
	MinChar       uint16
	MaxChar       uint16
	DefaultChar   uint16
	NFontProps    uint16
	DrawDirection byte
	MinByte1      byte
	MaxByte1      byte
	AllCharsExist bool
	FontAscent    int16
	FontDescent   int16
	NReplies      uint32
	FontProps     []FontProp
	FontName      string
}

type FontProp struct {
	Name  uint32
	Value uint32
}

func (r *ListFontsWithInfoReply) EncodeMessage(order binary.ByteOrder) []byte {
	fontNameBytes := []byte(r.FontName)
	fontNameLen := len(fontNameBytes)
	p := (4 - (fontNameLen % 4)) % 4
	totalLen := 60 + len(r.FontProps)*8 + fontNameLen + p

	reply := make([]byte, totalLen)
	reply[0] = 1 // Reply type
	reply[1] = r.NameLength
	order.PutUint16(reply[2:4], r.Sequence)
	order.PutUint32(reply[4:8], uint32((totalLen-32)/4))

	// min-bounds
	order.PutUint16(reply[8:10], uint16(r.MinBounds.LeftSideBearing))
	order.PutUint16(reply[10:12], uint16(r.MinBounds.RightSideBearing))
	order.PutUint16(reply[12:14], r.MinBounds.CharacterWidth)
	order.PutUint16(reply[14:16], uint16(r.MinBounds.Ascent))
	order.PutUint16(reply[16:18], uint16(r.MinBounds.Descent))
	order.PutUint16(reply[18:20], r.MinBounds.Attributes)

	// max-bounds
	order.PutUint16(reply[24:26], uint16(r.MaxBounds.LeftSideBearing))
	order.PutUint16(reply[26:28], uint16(r.MaxBounds.RightSideBearing))
	order.PutUint16(reply[28:30], r.MaxBounds.CharacterWidth)
	order.PutUint16(reply[30:32], uint16(r.MaxBounds.Ascent))
	order.PutUint16(reply[32:34], uint16(r.MaxBounds.Descent))
	order.PutUint16(reply[34:36], r.MaxBounds.Attributes)

	order.PutUint16(reply[40:42], r.MinChar)
	order.PutUint16(reply[42:44], r.MaxChar)
	order.PutUint16(reply[44:46], r.DefaultChar)
	order.PutUint16(reply[46:48], r.NFontProps)
	reply[48] = r.DrawDirection
	reply[49] = r.MinByte1
	reply[50] = r.MaxByte1
	reply[51] = BoolToByte(r.AllCharsExist)
	order.PutUint16(reply[52:54], uint16(r.FontAscent))
	order.PutUint16(reply[54:56], uint16(r.FontDescent))
	order.PutUint32(reply[56:60], r.NReplies)

	offset := 60
	for _, prop := range r.FontProps {
		order.PutUint32(reply[offset:offset+4], prop.Name)
		order.PutUint32(reply[offset+4:offset+8], prop.Value)
		offset += 8
	}

	copy(reply[offset:], fontNameBytes)
	return reply
}

func ParseListFontsWithInfoReply(order binary.ByteOrder, b []byte) (*ListFontsWithInfoReply, error) {
	if len(b) < 60 {
		return nil, NewError(LengthErrorCode, 0, 0, Opcodes{Major: 0, Minor: 0})
	}
	nameLength := b[1]
	nFontProps := order.Uint16(b[46:48])
	if len(b) < 60+int(nFontProps)*8+int(nameLength) {
		return nil, NewError(LengthErrorCode, 0, 0, Opcodes{Major: 0, Minor: 0})
	}
	fontProps := make([]FontProp, nFontProps)
	offset := 60
	for i := 0; i < int(nFontProps); i++ {
		fontProps[i] = FontProp{
			Name:  order.Uint32(b[offset:]),
			Value: order.Uint32(b[offset+4:]),
		}
		offset += 8
	}
	fontName := string(b[offset : offset+int(nameLength)])

	r := &ListFontsWithInfoReply{
		Sequence:   order.Uint16(b[2:4]),
		NameLength: nameLength,
		MinBounds: XCharInfo{
			LeftSideBearing:  int16(order.Uint16(b[8:10])),
			RightSideBearing: int16(order.Uint16(b[10:12])),
			CharacterWidth:   order.Uint16(b[12:14]),
			Ascent:           int16(order.Uint16(b[14:16])),
			Descent:          int16(order.Uint16(b[16:18])),
			Attributes:       order.Uint16(b[18:20]),
		},
		MaxBounds: XCharInfo{
			LeftSideBearing:  int16(order.Uint16(b[24:26])),
			RightSideBearing: int16(order.Uint16(b[26:28])),
			CharacterWidth:   order.Uint16(b[28:30]),
			Ascent:           int16(order.Uint16(b[30:32])),
			Descent:          int16(order.Uint16(b[32:34])),
			Attributes:       order.Uint16(b[34:36]),
		},
		MinChar:       order.Uint16(b[40:42]),
		MaxChar:       order.Uint16(b[42:44]),
		DefaultChar:   order.Uint16(b[44:46]),
		NFontProps:    nFontProps,
		DrawDirection: b[48],
		MinByte1:      b[49],
		MaxByte1:      b[50],
		AllCharsExist: b[51] != 0,
		FontAscent:    int16(order.Uint16(b[52:54])),
		FontDescent:   int16(order.Uint16(b[54:56])),
		NReplies:      order.Uint32(b[56:60]),
		FontProps:     fontProps,
		FontName:      fontName,
	}
	return r, nil
}

// QueryTree: 15
type QueryTreeReply struct {
	Sequence    uint16
	Root        uint32
	Parent      uint32
	NumChildren uint16
	Children    []uint32
}

func (r *QueryTreeReply) EncodeMessage(order binary.ByteOrder) []byte {
	reply := make([]byte, 32+len(r.Children)*4)
	reply[0] = 1 // Reply type
	order.PutUint16(reply[2:4], r.Sequence)
	order.PutUint32(reply[4:8], uint32(len(r.Children)))
	order.PutUint32(reply[8:12], r.Root)
	order.PutUint32(reply[12:16], r.Parent)
	order.PutUint16(reply[16:18], r.NumChildren)
	for i, child := range r.Children {
		order.PutUint32(reply[32+i*4:], child)
	}
	return reply
}

func ParseQueryTreeReply(order binary.ByteOrder, b []byte) (*QueryTreeReply, error) {
	if len(b) < 32 {
		return nil, NewError(LengthErrorCode, 0, 0, Opcodes{Major: 0, Minor: 0})
	}
	numChildren := order.Uint16(b[16:18])
	if len(b) < 32+int(numChildren)*4 {
		return nil, NewError(LengthErrorCode, 0, 0, Opcodes{Major: 0, Minor: 0})
	}
	children := make([]uint32, numChildren)
	for i := 0; i < int(numChildren); i++ {
		children[i] = order.Uint32(b[32+i*4:])
	}
	r := &QueryTreeReply{
		Sequence:    order.Uint16(b[2:4]),
		Root:        order.Uint32(b[8:12]),
		Parent:      order.Uint32(b[12:16]),
		NumChildren: numChildren,
		Children:    children,
	}
	return r, nil
}

// AllocColorCells: 86
type AllocColorCellsReply struct {
	Sequence uint16
	NPixels  uint16
	NMasks   uint16
	Pixels   []uint32
	Masks    []uint32
}

func (r *AllocColorCellsReply) EncodeMessage(order binary.ByteOrder) []byte {
	numPixels := len(r.Pixels)
	numMasks := len(r.Masks)
	reply := make([]byte, 32+(numPixels+numMasks)*4)
	reply[0] = 1 // Reply type
	// byte 1 is unused
	order.PutUint16(reply[2:4], r.Sequence)
	order.PutUint32(reply[4:8], uint32(numPixels+numMasks)) // Reply length
	order.PutUint16(reply[8:10], uint16(numPixels))
	order.PutUint16(reply[10:12], uint16(numMasks))
	// reply[12:32] is padding
	for i, pixel := range r.Pixels {
		order.PutUint32(reply[32+i*4:], pixel)
	}
	for i, mask := range r.Masks {
		order.PutUint32(reply[32+numPixels*4+i*4:], mask)
	}
	return reply
}

func ParseAllocColorCellsReply(order binary.ByteOrder, b []byte) (*AllocColorCellsReply, error) {
	if len(b) < 32 {
		return nil, NewError(LengthErrorCode, 0, 0, Opcodes{Major: 0, Minor: 0})
	}
	nPixels := order.Uint16(b[8:10])
	nMasks := order.Uint16(b[10:12])
	if len(b) < 32+int(nPixels)*4+int(nMasks)*4 {
		return nil, NewError(LengthErrorCode, 0, 0, Opcodes{Major: 0, Minor: 0})
	}
	pixels := make([]uint32, nPixels)
	masks := make([]uint32, nMasks)
	for i := 0; i < int(nPixels); i++ {
		pixels[i] = order.Uint32(b[32+i*4:])
	}
	for i := 0; i < int(nMasks); i++ {
		masks[i] = order.Uint32(b[32+int(nPixels)*4+i*4:])
	}
	r := &AllocColorCellsReply{
		Sequence: order.Uint16(b[2:4]),
		NPixels:  nPixels,
		NMasks:   nMasks,
		Pixels:   pixels,
		Masks:    masks,
	}
	return r, nil
}

// AllocColorPlanes: 87
type AllocColorPlanesReply struct {
	Sequence  uint16
	NPixels   uint16
	RedMask   uint32
	GreenMask uint32
	BlueMask  uint32
	Pixels    []uint32
}

func (r *AllocColorPlanesReply) EncodeMessage(order binary.ByteOrder) []byte {
	numPixels := len(r.Pixels)
	reply := make([]byte, 32+numPixels*4)
	reply[0] = 1 // Reply type
	// byte 1 is unused
	order.PutUint16(reply[2:4], r.Sequence)
	order.PutUint32(reply[4:8], uint32(numPixels)) // Reply length
	order.PutUint16(reply[8:10], uint16(numPixels))
	// reply[10:12] is padding
	order.PutUint32(reply[12:16], r.RedMask)
	order.PutUint32(reply[16:20], r.GreenMask)
	order.PutUint32(reply[20:24], r.BlueMask)
	// reply[24:32] is padding
	for i, pixel := range r.Pixels {
		order.PutUint32(reply[32+i*4:], pixel)
	}
	return reply
}

func ParseAllocColorPlanesReply(order binary.ByteOrder, b []byte) (*AllocColorPlanesReply, error) {
	if len(b) < 32 {
		return nil, NewError(LengthErrorCode, 0, 0, Opcodes{Major: 0, Minor: 0})
	}
	nPixels := order.Uint16(b[8:10])
	if len(b) < 32+int(nPixels)*4 {
		return nil, NewError(LengthErrorCode, 0, 0, Opcodes{Major: 0, Minor: 0})
	}
	pixels := make([]uint32, nPixels)
	for i := 0; i < int(nPixels); i++ {
		pixels[i] = order.Uint32(b[32+i*4:])
	}
	r := &AllocColorPlanesReply{
		Sequence:  order.Uint16(b[2:4]),
		NPixels:   nPixels,
		RedMask:   order.Uint32(b[12:16]),
		GreenMask: order.Uint32(b[16:20]),
		BlueMask:  order.Uint32(b[20:24]),
		Pixels:    pixels,
	}
	return r, nil
}

// ListExtensions: 99
type ListExtensionsReply struct {
	Sequence uint16
	NNames   byte
	Names    []string
}

func ParseListExtensionsReply(order binary.ByteOrder, b []byte) (*ListExtensionsReply, error) {
	if len(b) < 32 {
		return nil, NewError(LengthErrorCode, 0, 0, Opcodes{Major: 0, Minor: 0})
	}
	nNames := b[1]
	names := make([]string, nNames)
	offset := 32
	for i := 0; i < int(nNames); i++ {
		if len(b) < offset+1 {
			return nil, NewError(LengthErrorCode, 0, 0, Opcodes{Major: 0, Minor: 0})
		}
		length := int(b[offset])
		if len(b) < offset+1+length {
			return nil, NewError(LengthErrorCode, 0, 0, Opcodes{Major: 0, Minor: 0})
		}
		names[i] = string(b[offset+1 : offset+1+length])
		offset += 1 + length
	}
	r := &ListExtensionsReply{
		Sequence: order.Uint16(b[2:4]),
		NNames:   nNames,
		Names:    names,
	}
	return r, nil
}

// GetPointerControl: 106
type GetPointerControlReply struct {
	Sequence         uint16
	AccelNumerator   uint16
	AccelDenominator uint16
	Threshold        uint16
}

func (r *GetPointerControlReply) EncodeMessage(order binary.ByteOrder) []byte {
	reply := make([]byte, 32)
	reply[0] = 1 // Reply type
	// byte 1 is unused
	order.PutUint16(reply[2:4], r.Sequence)
	order.PutUint32(reply[4:8], 0) // Reply length
	order.PutUint16(reply[8:10], r.AccelNumerator)
	order.PutUint16(reply[10:12], r.AccelDenominator)
	order.PutUint16(reply[12:14], r.Threshold)
	// reply[14:32] is padding
	return reply
}

func ParseGetPointerControlReply(order binary.ByteOrder, b []byte) (*GetPointerControlReply, error) {
	if len(b) < 32 {
		return nil, NewError(LengthErrorCode, 0, 0, Opcodes{Major: 0, Minor: 0})
	}
	r := &GetPointerControlReply{
		Sequence:         order.Uint16(b[2:4]),
		AccelNumerator:   order.Uint16(b[8:10]),
		AccelDenominator: order.Uint16(b[10:12]),
		Threshold:        order.Uint16(b[12:14]),
	}
	return r, nil
}

func (r *ListExtensionsReply) EncodeMessage(order binary.ByteOrder) []byte {
	var data []byte
	for _, name := range r.Names {
		data = append(data, byte(len(name)))
		data = append(data, name...)
	}
	p := (4 - (len(data) % 4)) % 4
	reply := make([]byte, 32+len(data)+p)
	reply[0] = 1 // Reply type
	reply[1] = r.NNames
	order.PutUint16(reply[2:4], r.Sequence)
	order.PutUint32(reply[4:8], uint32((len(data)+p)/4))
	copy(reply[32:], data)
	return reply
}
