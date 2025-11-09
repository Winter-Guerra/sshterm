//go:build x11

package x11

import (
	"encoding/binary"
	"errors"
	"fmt"
)

var (
	errParseError = errors.New("x11: request parsing error")
)

type request interface {
	OpCode() reqCode
}

func parseRequest(order binary.ByteOrder, raw []byte) (request, error) {
	var reqHeader [4]byte
	if n := copy(reqHeader[:], raw); n != 4 {
		return nil, fmt.Errorf("%w: header too short", errParseError)
	}
	length := order.Uint16(reqHeader[2:4])
	if int(4*length) != len(raw) {
		debugf("X11: Raw request header: %x", reqHeader)
		return nil, fmt.Errorf("%w: mismatch length %d != %d", errParseError, 4*length, len(raw))
	}

	opcode := reqCode(reqHeader[0])
	data := reqHeader[1]
	body := raw[4:]

	switch opcode {
	case CreateWindow:
		return parseCreateWindowRequest(order, data, body)

	case ChangeWindowAttributes:
		return parseChangeWindowAttributesRequest(order, body)

	case GetWindowAttributes:
		return parseGetWindowAttributesRequest(order, body)

	case DestroyWindow:
		return parseDestroyWindowRequest(order, body)

	case DestroySubwindows:
		return parseDestroySubwindowsRequest(order, body)

	case ChangeSaveSet:
		return parseChangeSaveSetRequest(order, body)

	case ReparentWindow:
		return parseReparentWindowRequest(order, body)

	case MapWindow:
		return parseMapWindowRequest(order, body)

	case MapSubwindows:
		return parseMapSubwindowsRequest(order, body)

	case UnmapWindow:
		return parseUnmapWindowRequest(order, body)

	case UnmapSubwindows:
		return parseUnmapSubwindowsRequest(order, body)

	case ConfigureWindow:
		return parseConfigureWindowRequest(order, body)

	case CirculateWindow:
		return parseCirculateWindowRequest(order, data, body)

	case GetGeometry:
		return parseGetGeometryRequest(order, body)

	case QueryTree:
		return parseQueryTreeRequest(order, body)

	case InternAtom:
		return parseInternAtomRequest(order, body)

	case GetAtomName:
		return parseGetAtomNameRequest(order, body)

	case ChangeProperty:
		return parseChangePropertyRequest(order, body)

	case DeleteProperty:
		return parseDeletePropertyRequest(order, body)

	case GetProperty:
		return parseGetPropertyRequest(order, body)

	case ListProperties:
		return parseListPropertiesRequest(order, body)

	case SetSelectionOwner:
		return parseSetSelectionOwnerRequest(order, body)

	case GetSelectionOwner:
		return parseGetSelectionOwnerRequest(order, body)

	case ConvertSelection:
		return parseConvertSelectionRequest(order, body)

	case SendEvent:
		return parseSendEventRequest(order, body)

	case GrabPointer:
		return parseGrabPointerRequest(order, body)

	case UngrabPointer:
		return parseUngrabPointerRequest(order, body)

	case GrabButton:
		return parseGrabButtonRequest(order, body)

	case UngrabButton:
		return parseUngrabButtonRequest(order, body)

	case ChangeActivePointerGrab:
		return parseChangeActivePointerGrabRequest(order, body)

	case GrabKeyboard:
		return parseGrabKeyboardRequest(order, body)

	case UngrabKeyboard:
		return parseUngrabKeyboardRequest(order, body)

	case GrabKey:
		return parseGrabKeyRequest(order, body)

	case UngrabKey:
		return parseUngrabKeyRequest(order, body)

	case AllowEvents:
		return parseAllowEventsRequest(order, data, body)

	case GrabServer:
		return parseGrabServerRequest(order, body)

	case UngrabServer:
		return parseUngrabServerRequest(order, body)

	case QueryPointer:
		return parseQueryPointerRequest(order, body)

	case GetMotionEvents:
		return parseGetMotionEventsRequest(order, body)

	case TranslateCoords:
		return parseTranslateCoordsRequest(order, body)

	case WarpPointer:
		return parseWarpPointerRequest(order, body)

	case SetInputFocus:
		return parseSetInputFocusRequest(order, body)

	case GetInputFocus:
		return parseGetInputFocusRequest(order, body)

	case QueryKeymap:
		return parseQueryKeymapRequest(order, body)

	case OpenFont:
		return parseOpenFontRequest(order, body)

	case CloseFont:
		return parseCloseFontRequest(order, body)

	case QueryFont:
		return parseQueryFontRequest(order, body)

	case QueryTextExtents:
		return parseQueryTextExtentsRequest(order, body)

	case ListFonts:
		return parseListFontsRequest(order, body)

	case ListFontsWithInfo:
		return parseListFontsWithInfoRequest(order, body)

	case SetFontPath:
		return parseSetFontPathRequest(order, body)

	case GetFontPath:
		return parseGetFontPathRequest(order, body)

	case CreatePixmap:
		return parseCreatePixmapRequest(order, data, body)

	case FreePixmap:
		return parseFreePixmapRequest(order, body)

	case CreateGC:
		return parseCreateGCRequest(order, body)

	case ChangeGC:
		return parseChangeGCRequest(order, body)

	case CopyGC:
		return parseCopyGCRequest(order, body)

	case SetDashes:
		return parseSetDashesRequest(order, body)

	case SetClipRectangles:
		return parseSetClipRectanglesRequest(order, data, body)

	case FreeGC:
		return parseFreeGCRequest(order, body)

	case ClearArea:
		return parseClearAreaRequest(order, body)

	case CopyArea:
		return parseCopyAreaRequest(order, body)

	case PolyPoint:
		return parsePolyPointRequest(order, body)

	case PolyLine:
		return parsePolyLineRequest(order, body)

	case PolySegment:
		return parsePolySegmentRequest(order, body)

	case PolyRectangle:
		return parsePolyRectangleRequest(order, body)

	case PolyArc:
		return parsePolyArcRequest(order, body)

	case FillPoly:
		return parseFillPolyRequest(order, body)

	case PolyFillRectangle:
		return parsePolyFillRectangleRequest(order, body)

	case PolyFillArc:
		return parsePolyFillArcRequest(order, body)

	case PutImage:
		return parsePutImageRequest(order, data, body)

	case GetImage:
		return parseGetImageRequest(order, data, body)

	case PolyText8:
		return parsePolyText8Request(order, body)

	case PolyText16:
		return parsePolyText16Request(order, body)

	case ImageText8:
		return parseImageText8Request(order, body)

	case ImageText16:
		return parseImageText16Request(order, body)

	case CreateColormap:
		return parseCreateColormapRequest(order, data, body)

	case FreeColormap:
		return parseFreeColormapRequest(order, body)

	case InstallColormap:
		return parseInstallColormapRequest(order, body)

	case UninstallColormap:
		return parseUninstallColormapRequest(order, body)

	case ListInstalledColormaps:
		return parseListInstalledColormapsRequest(order, body)

	case AllocColor:
		return parseAllocColorRequest(order, body)

	case AllocNamedColor:
		return parseAllocNamedColorRequest(order, body)

	case FreeColors:
		return parseFreeColorsRequest(order, body)

	case StoreColors:
		return parseStoreColorsRequest(order, body)

	case StoreNamedColor:
		return parseStoreNamedColorRequest(order, data, body)

	case QueryColors:
		return parseQueryColorsRequest(order, body)

	case LookupColor:
		return parseLookupColorRequest(order, body)

	case CreateGlyphCursor:
		return parseCreateGlyphCursorRequest(order, body)

	case FreeCursor:
		return parseFreeCursorRequest(order, body)

	case RecolorCursor:
		return parseRecolorCursorRequest(order, body)

	case QueryBestSize:
		return parseQueryBestSizeRequest(order, body)

	case QueryExtension:
		return parseQueryExtensionRequest(order, body)

	case Bell:
		return parseBellRequest(data)

	case SetPointerMapping:
		return parseSetPointerMappingRequest(order, body)

	case GetPointerMapping:
		return parseGetPointerMappingRequest(order, body)

	case GetKeyboardMapping:
		return parseGetKeyboardMappingRequest(order, body)

	case ChangeKeyboardMapping:
		return parseChangeKeyboardMappingRequest(order, data, body)

	case ChangeKeyboardControl:
		return parseChangeKeyboardControlRequest(order, body)

	case GetKeyboardControl:
		return parseGetKeyboardControlRequest(order, body)

	case SetScreenSaver:
		return parseSetScreenSaverRequest(order, body)

	case GetScreenSaver:
		return parseGetScreenSaverRequest(order, body)

	case ChangeHosts:
		return parseChangeHostsRequest(order, data, body)

	case ListHosts:
		return parseListHostsRequest(order, body)

	case SetAccessControl:
		return parseSetAccessControlRequest(order, data, body)

	case SetCloseDownMode:
		return parseSetCloseDownModeRequest(order, data, body)

	case KillClient:
		return parseKillClientRequest(order, body)

	case RotateProperties:
		return parseRotatePropertiesRequest(order, body)

	case ForceScreenSaver:
		return parseForceScreenSaverRequest(order, data, body)

	case SetModifierMapping:
		return parseSetModifierMappingRequest(order, body)

	case GetModifierMapping:
		return parseGetModifierMappingRequest(order, body)

	case NoOperation:
		return parseNoOperationRequest(order, body)

	case AllocColorCells:
		return parseAllocColorCellsRequest(order, data, body)

	case AllocColorPlanes:
		return parseAllocColorPlanesRequest(order, data, body)

	case CreateCursor:
		return parseCreateCursorRequest(order, body)

	case CopyPlane:
		return parseCopyPlaneRequest(order, body)

	case ListExtensions:
		return parseListExtensionsRequest(order, raw)

	case ChangePointerControl:
		return parseChangePointerControlRequest(order, body)

	case GetPointerControl:
		return parseGetPointerControlRequest(order, raw)

	default:
		return nil, fmt.Errorf("x11: unhandled opcode %d", opcode)
	}
}

// auxiliary data structures

type WindowAttributes struct {
	BackgroundPixmap  Pixmap
	BackgroundPixel   uint32
	BorderPixmap      Pixmap
	BorderPixel       uint32
	BitGravity        uint32
	WinGravity        uint32
	BackingStore      uint32
	BackingPlanes     uint32
	BackingPixel      uint32
	OverrideRedirect  bool
	SaveUnder         bool
	EventMask         uint32
	DontPropagateMask uint32
	Colormap          Colormap
	Cursor            Cursor

	// Not part of value-mask, but part of window state
	Class          uint32
	MapIsInstalled bool
	MapState       uint32
}

// PolyTextItem is an interface for items in a PolyText request.
type PolyTextItem interface {
	isPolyTextItem()
}

// PolyText8String represents a string in a PolyText8 request.
type PolyText8String struct {
	Delta int8
	Str   []byte
}

func (PolyText8String) isPolyTextItem() {}

// PolyText16String represents a string in a PolyText16 request.
type PolyText16String struct {
	Delta int8
	Str   []uint16
}

func (PolyText16String) isPolyTextItem() {}

// PolyTextFont represents a font change in a PolyText request.
type PolyTextFont struct {
	Font Font
}

func (PolyTextFont) isPolyTextItem() {}

// request messages

type CreateWindowRequest struct {
	Depth       uint8
	Drawable    Window
	Parent      Window
	X           int16
	Y           int16
	Width       uint16
	Height      uint16
	BorderWidth uint16
	Class       uint16
	Visual      VisualID
	ValueMask   uint32
	Values      WindowAttributes
}

func (CreateWindowRequest) OpCode() reqCode { return CreateWindow }

func parseCreateWindowRequest(order binary.ByteOrder, data byte, requestBody []byte) (*CreateWindowRequest, error) {
	if len(requestBody) < 28 {
		return nil, fmt.Errorf("%w: create window request too short", errParseError)
	}
	req := &CreateWindowRequest{}
	req.Depth = data
	req.Drawable = Window(order.Uint32(requestBody[0:4]))
	req.Parent = Window(order.Uint32(requestBody[4:8]))
	req.X = int16(order.Uint16(requestBody[8:10]))
	req.Y = int16(order.Uint16(requestBody[10:12]))
	req.Width = order.Uint16(requestBody[12:14])
	req.Height = order.Uint16(requestBody[14:16])
	req.BorderWidth = order.Uint16(requestBody[16:18])
	req.Class = order.Uint16(requestBody[18:20])
	req.Visual = VisualID(order.Uint32(requestBody[20:24]))
	req.ValueMask = order.Uint32(requestBody[24:28])
	values, _, err := parseWindowAttributes(order, req.ValueMask, requestBody[28:])
	if err != nil {
		return nil, err
	}
	req.Values = values
	return req, nil
}

type ChangeWindowAttributesRequest struct {
	Window    Window
	ValueMask uint32
	Values    WindowAttributes
}

func (ChangeWindowAttributesRequest) OpCode() reqCode { return ChangeWindowAttributes }

func parseChangeWindowAttributesRequest(order binary.ByteOrder, requestBody []byte) (*ChangeWindowAttributesRequest, error) {
	if len(requestBody) < 8 {
		return nil, fmt.Errorf("%w: change window attributes request too short", errParseError)
	}
	req := &ChangeWindowAttributesRequest{}
	req.Window = Window(order.Uint32(requestBody[0:4]))
	req.ValueMask = order.Uint32(requestBody[4:8])
	values, _, err := parseWindowAttributes(order, req.ValueMask, requestBody[8:])
	if err != nil {
		return nil, err
	}
	req.Values = values
	return req, nil
}

type GetWindowAttributesRequest struct {
	Window Window
}

func (GetWindowAttributesRequest) OpCode() reqCode { return GetWindowAttributes }

func parseGetWindowAttributesRequest(order binary.ByteOrder, requestBody []byte) (*GetWindowAttributesRequest, error) {
	if len(requestBody) < 4 {
		return nil, fmt.Errorf("%w: get window attributes request too short", errParseError)
	}
	req := &GetWindowAttributesRequest{}
	req.Window = Window(order.Uint32(requestBody[0:4]))
	return req, nil
}

type DestroyWindowRequest struct {
	Window Window
}

func (DestroyWindowRequest) OpCode() reqCode { return DestroyWindow }

func parseDestroyWindowRequest(order binary.ByteOrder, requestBody []byte) (*DestroyWindowRequest, error) {
	if len(requestBody) < 4 {
		return nil, fmt.Errorf("%w: destroy window request too short", errParseError)
	}
	req := &DestroyWindowRequest{}
	req.Window = Window(order.Uint32(requestBody[0:4]))
	return req, nil
}

type DestroySubwindowsRequest struct {
	Window Window
}

func (DestroySubwindowsRequest) OpCode() reqCode { return DestroySubwindows }

func parseDestroySubwindowsRequest(order binary.ByteOrder, requestBody []byte) (*DestroySubwindowsRequest, error) {
	if len(requestBody) < 4 {
		return nil, fmt.Errorf("%w: destroy subwindows request too short", errParseError)
	}
	req := &DestroySubwindowsRequest{}
	req.Window = Window(order.Uint32(requestBody[0:4]))
	return req, nil
}

type ChangeSaveSetRequest struct {
	Window Window
	Mode   byte
}

func (ChangeSaveSetRequest) OpCode() reqCode { return ChangeSaveSet }

func parseChangeSaveSetRequest(order binary.ByteOrder, requestBody []byte) (*ChangeSaveSetRequest, error) {
	if len(requestBody) < 5 {
		return nil, fmt.Errorf("%w: change save set request too short", errParseError)
	}
	req := &ChangeSaveSetRequest{}
	req.Window = Window(order.Uint32(requestBody[0:4]))
	req.Mode = requestBody[4]
	return req, nil
}

type ReparentWindowRequest struct {
	Window Window
	Parent Window
	X      int16
	Y      int16
}

func (ReparentWindowRequest) OpCode() reqCode { return ReparentWindow }

func parseReparentWindowRequest(order binary.ByteOrder, requestBody []byte) (*ReparentWindowRequest, error) {
	if len(requestBody) < 12 {
		return nil, fmt.Errorf("%w: reparent window request too short", errParseError)
	}
	req := &ReparentWindowRequest{}
	req.Window = Window(order.Uint32(requestBody[0:4]))
	req.Parent = Window(order.Uint32(requestBody[4:8]))
	req.X = int16(order.Uint16(requestBody[8:10]))
	req.Y = int16(order.Uint16(requestBody[10:12]))
	return req, nil
}

type MapWindowRequest struct {
	Window Window
}

func (MapWindowRequest) OpCode() reqCode { return MapWindow }

func parseMapWindowRequest(order binary.ByteOrder, requestBody []byte) (*MapWindowRequest, error) {
	if len(requestBody) < 4 {
		return nil, fmt.Errorf("%w: map window request too short", errParseError)
	}
	req := &MapWindowRequest{}
	req.Window = Window(order.Uint32(requestBody[0:4]))
	return req, nil
}

type MapSubwindowsRequest struct {
	Window Window
}

func (MapSubwindowsRequest) OpCode() reqCode { return MapSubwindows }

func parseMapSubwindowsRequest(order binary.ByteOrder, requestBody []byte) (*MapSubwindowsRequest, error) {
	if len(requestBody) < 4 {
		return nil, fmt.Errorf("%w: map subwindows request too short", errParseError)
	}
	req := &MapSubwindowsRequest{}
	req.Window = Window(order.Uint32(requestBody[0:4]))
	return req, nil
}

type UnmapWindowRequest struct {
	Window Window
}

func (UnmapWindowRequest) OpCode() reqCode { return UnmapWindow }

func parseUnmapWindowRequest(order binary.ByteOrder, requestBody []byte) (*UnmapWindowRequest, error) {
	if len(requestBody) < 4 {
		return nil, fmt.Errorf("%w: unmap window request too short", errParseError)
	}
	req := &UnmapWindowRequest{}
	req.Window = Window(order.Uint32(requestBody[0:4]))
	return req, nil
}

type UnmapSubwindowsRequest struct {
	Window Window
}

func (UnmapSubwindowsRequest) OpCode() reqCode { return UnmapSubwindows }

func parseUnmapSubwindowsRequest(order binary.ByteOrder, requestBody []byte) (*UnmapSubwindowsRequest, error) {
	if len(requestBody) < 4 {
		return nil, fmt.Errorf("%w: unmap subwindows request too short", errParseError)
	}
	req := &UnmapSubwindowsRequest{}
	req.Window = Window(order.Uint32(requestBody[0:4]))
	return req, nil
}

type ConfigureWindowRequest struct {
	Window    Window
	ValueMask uint16
	Values    []uint32
}

func (ConfigureWindowRequest) OpCode() reqCode { return ConfigureWindow }

func parseConfigureWindowRequest(order binary.ByteOrder, requestBody []byte) (*ConfigureWindowRequest, error) {
	if len(requestBody) < 8 {
		return nil, fmt.Errorf("%w: configure window request too short", errParseError)
	}
	req := &ConfigureWindowRequest{}
	req.Window = Window(order.Uint32(requestBody[0:4]))
	req.ValueMask = order.Uint16(requestBody[4:6])
	for i := 8; i < len(requestBody); i += 4 {
		req.Values = append(req.Values, order.Uint32(requestBody[i:i+4]))
	}
	return req, nil
}

type CirculateWindowRequest struct {
	Window    Window
	Direction byte
}

func (CirculateWindowRequest) OpCode() reqCode { return CirculateWindow }

func parseCirculateWindowRequest(order binary.ByteOrder, data byte, requestBody []byte) (*CirculateWindowRequest, error) {
	if len(requestBody) < 4 {
		return nil, fmt.Errorf("%w: circulate window request too short", errParseError)
	}
	req := &CirculateWindowRequest{}
	req.Window = Window(order.Uint32(requestBody[0:4]))
	req.Direction = data
	return req, nil
}

type GetGeometryRequest struct {
	Drawable Drawable
}

func (GetGeometryRequest) OpCode() reqCode { return GetGeometry }

func parseGetGeometryRequest(order binary.ByteOrder, requestBody []byte) (*GetGeometryRequest, error) {
	if len(requestBody) < 4 {
		return nil, fmt.Errorf("%w: get geometry request too short", errParseError)
	}
	req := &GetGeometryRequest{}
	req.Drawable = Drawable(order.Uint32(requestBody[0:4]))
	return req, nil
}

type QueryTreeRequest struct {
	Window Window
}

func (QueryTreeRequest) OpCode() reqCode { return QueryTree }

func parseQueryTreeRequest(order binary.ByteOrder, requestBody []byte) (*QueryTreeRequest, error) {
	if len(requestBody) < 4 {
		return nil, fmt.Errorf("%w: query tree request too short", errParseError)
	}
	req := &QueryTreeRequest{}
	req.Window = Window(order.Uint32(requestBody[0:4]))
	return req, nil
}

type InternAtomRequest struct {
	Name         string
	OnlyIfExists bool
}

func (InternAtomRequest) OpCode() reqCode { return InternAtom }

func parseInternAtomRequest(order binary.ByteOrder, requestBody []byte) (*InternAtomRequest, error) {
	if len(requestBody) < 4 {
		return nil, fmt.Errorf("%w: intern atom request too short", errParseError)
	}
	req := &InternAtomRequest{}
	nameLen := order.Uint16(requestBody[0:2])
	if len(requestBody) < 4+int(nameLen) {
		return nil, fmt.Errorf("%w: intern atom request too short for name", errParseError)
	}
	req.Name = string(requestBody[4 : 4+nameLen])
	return req, nil
}

type GetAtomNameRequest struct {
	Atom Atom
}

func (GetAtomNameRequest) OpCode() reqCode { return GetAtomName }

func parseGetAtomNameRequest(order binary.ByteOrder, requestBody []byte) (*GetAtomNameRequest, error) {
	if len(requestBody) < 4 {
		return nil, fmt.Errorf("%w: get atom name request too short", errParseError)
	}
	req := &GetAtomNameRequest{}
	req.Atom = Atom(order.Uint32(requestBody[0:4]))
	return req, nil
}

type ChangePropertyRequest struct {
	Window   Window
	Property Atom
	Type     Atom
	Format   byte
	Data     []byte
}

func (ChangePropertyRequest) OpCode() reqCode { return ChangeProperty }

func parseChangePropertyRequest(order binary.ByteOrder, requestBody []byte) (*ChangePropertyRequest, error) {
	if len(requestBody) < 20 {
		return nil, fmt.Errorf("%w: change property request too short", errParseError)
	}
	req := &ChangePropertyRequest{}
	req.Window = Window(order.Uint32(requestBody[0:4]))
	req.Property = Atom(order.Uint32(requestBody[4:8]))
	req.Type = Atom(order.Uint32(requestBody[8:12]))
	req.Format = requestBody[12]
	dataLen := order.Uint32(requestBody[16:20])
	if len(requestBody) < 20+int(dataLen) {
		return nil, fmt.Errorf("%w: change property request too short for data", errParseError)
	}
	req.Data = requestBody[20 : 20+dataLen]
	return req, nil
}

type DeletePropertyRequest struct {
	Window   Window
	Property Atom
}

func (DeletePropertyRequest) OpCode() reqCode { return DeleteProperty }

func parseDeletePropertyRequest(order binary.ByteOrder, requestBody []byte) (*DeletePropertyRequest, error) {
	if len(requestBody) < 8 {
		return nil, fmt.Errorf("%w: delete property request too short", errParseError)
	}
	req := &DeletePropertyRequest{}
	req.Window = Window(order.Uint32(requestBody[0:4]))
	req.Property = Atom(order.Uint32(requestBody[4:8]))
	return req, nil
}

type GetPropertyRequest struct {
	Window   Window
	Property Atom
	Type     Atom
	Delete   bool
	Offset   uint32
	Length   uint32
}

func (GetPropertyRequest) OpCode() reqCode { return GetProperty }

func parseGetPropertyRequest(order binary.ByteOrder, requestBody []byte) (*GetPropertyRequest, error) {
	if len(requestBody) < 20 {
		return nil, fmt.Errorf("%w: get property request too short", errParseError)
	}
	req := &GetPropertyRequest{}
	req.Window = Window(order.Uint32(requestBody[0:4]))
	req.Property = Atom(order.Uint32(requestBody[4:8]))
	req.Delete = requestBody[8] != 0
	req.Offset = order.Uint32(requestBody[12:16])
	req.Length = order.Uint32(requestBody[16:20])
	return req, nil
}

type ListPropertiesRequest struct {
	Window Window
}

func (ListPropertiesRequest) OpCode() reqCode { return ListProperties }

func parseListPropertiesRequest(order binary.ByteOrder, requestBody []byte) (*ListPropertiesRequest, error) {
	if len(requestBody) < 4 {
		return nil, fmt.Errorf("%w: list properties request too short", errParseError)
	}
	req := &ListPropertiesRequest{}
	req.Window = Window(order.Uint32(requestBody[0:4]))
	return req, nil
}

type SetSelectionOwnerRequest struct {
	Owner     Window
	Selection Atom
	Time      Timestamp
}

func (SetSelectionOwnerRequest) OpCode() reqCode { return SetSelectionOwner }

func parseSetSelectionOwnerRequest(order binary.ByteOrder, requestBody []byte) (*SetSelectionOwnerRequest, error) {
	if len(requestBody) < 12 {
		return nil, fmt.Errorf("%w: set selection owner request too short", errParseError)
	}
	req := &SetSelectionOwnerRequest{}
	req.Owner = Window(order.Uint32(requestBody[0:4]))
	req.Selection = Atom(order.Uint32(requestBody[4:8]))
	req.Time = Timestamp(order.Uint32(requestBody[8:12]))
	return req, nil
}

type GetSelectionOwnerRequest struct {
	Selection Atom
}

func (GetSelectionOwnerRequest) OpCode() reqCode { return GetSelectionOwner }

func parseGetSelectionOwnerRequest(order binary.ByteOrder, requestBody []byte) (*GetSelectionOwnerRequest, error) {
	if len(requestBody) < 4 {
		return nil, fmt.Errorf("%w: get selection owner request too short", errParseError)
	}
	req := &GetSelectionOwnerRequest{}
	req.Selection = Atom(order.Uint32(requestBody[0:4]))
	return req, nil
}

type ConvertSelectionRequest struct {
	Requestor Window
	Selection Atom
	Target    Atom
	Property  Atom
	Time      Timestamp
}

func (ConvertSelectionRequest) OpCode() reqCode { return ConvertSelection }

func parseConvertSelectionRequest(order binary.ByteOrder, requestBody []byte) (*ConvertSelectionRequest, error) {
	if len(requestBody) < 20 {
		return nil, fmt.Errorf("%w: convert selection request too short", errParseError)
	}
	req := &ConvertSelectionRequest{}
	req.Requestor = Window(order.Uint32(requestBody[0:4]))
	req.Selection = Atom(order.Uint32(requestBody[4:8]))
	req.Target = Atom(order.Uint32(requestBody[8:12]))
	req.Property = Atom(order.Uint32(requestBody[12:16]))
	req.Time = Timestamp(order.Uint32(requestBody[16:20]))
	return req, nil
}

type SendEventRequest struct {
	Propagate   bool
	Destination Window
	EventMask   uint32
	EventData   []byte
}

func (SendEventRequest) OpCode() reqCode { return SendEvent }

func parseSendEventRequest(order binary.ByteOrder, requestBody []byte) (*SendEventRequest, error) {
	if len(requestBody) < 44 {
		return nil, fmt.Errorf("%w: send event request too short", errParseError)
	}
	req := &SendEventRequest{}
	req.Destination = Window(order.Uint32(requestBody[4:8]))
	req.EventMask = order.Uint32(requestBody[8:12])
	req.EventData = requestBody[12:44]
	return req, nil
}

type GrabPointerRequest struct {
	OwnerEvents  bool
	GrabWindow   Window
	EventMask    uint16
	PointerMode  byte
	KeyboardMode byte
	ConfineTo    Window
	Cursor       Cursor
	Time         Timestamp
}

func (GrabPointerRequest) OpCode() reqCode { return GrabPointer }

func parseGrabPointerRequest(order binary.ByteOrder, requestBody []byte) (*GrabPointerRequest, error) {
	if len(requestBody) < 20 {
		return nil, fmt.Errorf("%w: grab pointer request too short", errParseError)
	}
	req := &GrabPointerRequest{}
	req.GrabWindow = Window(order.Uint32(requestBody[0:4]))
	req.EventMask = order.Uint16(requestBody[4:6])
	req.PointerMode = requestBody[6]
	req.KeyboardMode = requestBody[7]
	req.ConfineTo = Window(order.Uint32(requestBody[8:12]))
	req.Cursor = Cursor(order.Uint32(requestBody[12:16]))
	req.Time = Timestamp(order.Uint32(requestBody[16:20]))
	return req, nil
}

type UngrabPointerRequest struct {
	Time Timestamp
}

func (UngrabPointerRequest) OpCode() reqCode { return UngrabPointer }

func parseUngrabPointerRequest(order binary.ByteOrder, requestBody []byte) (*UngrabPointerRequest, error) {
	if len(requestBody) < 4 {
		return nil, fmt.Errorf("%w: ungrab pointer request too short", errParseError)
	}
	req := &UngrabPointerRequest{}
	req.Time = Timestamp(order.Uint32(requestBody[0:4]))
	return req, nil
}

type GrabButtonRequest struct {
	OwnerEvents  bool
	GrabWindow   Window
	EventMask    uint16
	PointerMode  byte
	KeyboardMode byte
	ConfineTo    Window
	Cursor       Cursor
	Button       byte
	Modifiers    uint16
}

func (GrabButtonRequest) OpCode() reqCode { return GrabButton }

func parseGrabButtonRequest(order binary.ByteOrder, requestBody []byte) (*GrabButtonRequest, error) {
	if len(requestBody) < 24 {
		return nil, fmt.Errorf("%w: grab button request too short", errParseError)
	}
	req := &GrabButtonRequest{}
	req.OwnerEvents = requestBody[0] != 0
	req.GrabWindow = Window(order.Uint32(requestBody[4:8]))
	req.EventMask = order.Uint16(requestBody[8:10])
	req.PointerMode = requestBody[10]
	req.KeyboardMode = requestBody[11]
	req.ConfineTo = Window(order.Uint32(requestBody[12:16]))
	req.Cursor = Cursor(order.Uint32(requestBody[16:20]))
	req.Button = requestBody[20]
	req.Modifiers = order.Uint16(requestBody[22:24])
	return req, nil
}

type UngrabButtonRequest struct {
	GrabWindow Window
	Button     byte
	Modifiers  uint16
}

func (UngrabButtonRequest) OpCode() reqCode { return UngrabButton }

func parseUngrabButtonRequest(order binary.ByteOrder, requestBody []byte) (*UngrabButtonRequest, error) {
	if len(requestBody) < 8 {
		return nil, fmt.Errorf("%w: ungrab button request too short", errParseError)
	}
	req := &UngrabButtonRequest{}
	req.GrabWindow = Window(order.Uint32(requestBody[0:4]))
	req.Button = requestBody[4]
	req.Modifiers = order.Uint16(requestBody[6:8])
	return req, nil
}

type ChangeActivePointerGrabRequest struct {
	Cursor    Cursor
	Time      Timestamp
	EventMask uint16
}

func (ChangeActivePointerGrabRequest) OpCode() reqCode { return ChangeActivePointerGrab }

func parseChangeActivePointerGrabRequest(order binary.ByteOrder, requestBody []byte) (*ChangeActivePointerGrabRequest, error) {
	if len(requestBody) < 12 {
		return nil, fmt.Errorf("%w: change active pointer grab request too short", errParseError)
	}
	req := &ChangeActivePointerGrabRequest{}
	req.Cursor = Cursor(order.Uint32(requestBody[0:4]))
	req.Time = Timestamp(order.Uint32(requestBody[4:8]))
	req.EventMask = order.Uint16(requestBody[8:10])
	return req, nil
}

type GrabKeyboardRequest struct {
	OwnerEvents  bool
	GrabWindow   Window
	Time         Timestamp
	PointerMode  byte
	KeyboardMode byte
}

func (GrabKeyboardRequest) OpCode() reqCode { return GrabKeyboard }

func parseGrabKeyboardRequest(order binary.ByteOrder, requestBody []byte) (*GrabKeyboardRequest, error) {
	if len(requestBody) < 12 {
		return nil, fmt.Errorf("%w: grab keyboard request too short", errParseError)
	}
	req := &GrabKeyboardRequest{}
	req.GrabWindow = Window(order.Uint32(requestBody[0:4]))
	req.Time = Timestamp(order.Uint32(requestBody[4:8]))
	req.PointerMode = requestBody[8]
	req.KeyboardMode = requestBody[9]
	return req, nil
}

type UngrabKeyboardRequest struct {
	Time Timestamp
}

func (UngrabKeyboardRequest) OpCode() reqCode { return UngrabKeyboard }

func parseUngrabKeyboardRequest(order binary.ByteOrder, requestBody []byte) (*UngrabKeyboardRequest, error) {
	if len(requestBody) < 4 {
		return nil, fmt.Errorf("%w: ungrab keyboard request too short", errParseError)
	}
	req := &UngrabKeyboardRequest{}
	req.Time = Timestamp(order.Uint32(requestBody[0:4]))
	return req, nil
}

type GrabKeyRequest struct {
	OwnerEvents  bool
	GrabWindow   Window
	Modifiers    uint16
	Key          KeyCode
	PointerMode  byte
	KeyboardMode byte
}

func (GrabKeyRequest) OpCode() reqCode { return GrabKey }

func parseGrabKeyRequest(order binary.ByteOrder, requestBody []byte) (*GrabKeyRequest, error) {
	if len(requestBody) < 13 {
		return nil, fmt.Errorf("%w: grab key request too short", errParseError)
	}
	req := &GrabKeyRequest{}
	req.OwnerEvents = requestBody[0] != 0
	req.GrabWindow = Window(order.Uint32(requestBody[4:8]))
	req.Modifiers = order.Uint16(requestBody[8:10])
	req.Key = KeyCode(requestBody[10])
	req.PointerMode = requestBody[11]
	req.KeyboardMode = requestBody[12]
	return req, nil
}

type UngrabKeyRequest struct {
	GrabWindow Window
	Modifiers  uint16
	Key        KeyCode
}

func (UngrabKeyRequest) OpCode() reqCode { return UngrabKey }

func parseUngrabKeyRequest(order binary.ByteOrder, requestBody []byte) (*UngrabKeyRequest, error) {
	if len(requestBody) < 8 {
		return nil, fmt.Errorf("%w: ungrab key request too short", errParseError)
	}
	req := &UngrabKeyRequest{}
	req.GrabWindow = Window(order.Uint32(requestBody[0:4]))
	req.Modifiers = order.Uint16(requestBody[4:6])
	req.Key = KeyCode(requestBody[6])
	return req, nil
}

type AllowEventsRequest struct {
	Mode byte
	Time Timestamp
}

func (AllowEventsRequest) OpCode() reqCode { return AllowEvents }

func parseAllowEventsRequest(order binary.ByteOrder, data byte, requestBody []byte) (*AllowEventsRequest, error) {
	if len(requestBody) < 4 {
		return nil, fmt.Errorf("%w: allow events request too short", errParseError)
	}
	req := &AllowEventsRequest{}
	req.Mode = data
	req.Time = Timestamp(order.Uint32(requestBody[0:4]))
	return req, nil
}

type GrabServerRequest struct{}

func (GrabServerRequest) OpCode() reqCode { return GrabServer }

func parseGrabServerRequest(order binary.ByteOrder, requestBody []byte) (*GrabServerRequest, error) {
	return &GrabServerRequest{}, nil
}

type UngrabServerRequest struct{}

func (UngrabServerRequest) OpCode() reqCode { return UngrabServer }

func parseUngrabServerRequest(order binary.ByteOrder, requestBody []byte) (*UngrabServerRequest, error) {
	return &UngrabServerRequest{}, nil
}

type QueryPointerRequest struct {
	Drawable Drawable
}

func (QueryPointerRequest) OpCode() reqCode { return QueryPointer }

func parseQueryPointerRequest(order binary.ByteOrder, requestBody []byte) (*QueryPointerRequest, error) {
	if len(requestBody) < 4 {
		return nil, fmt.Errorf("%w: query pointer request too short", errParseError)
	}
	req := &QueryPointerRequest{}
	req.Drawable = Drawable(order.Uint32(requestBody[0:4]))
	return req, nil
}

type GetMotionEventsRequest struct {
	Window Window
	Start  Timestamp
	Stop   Timestamp
}

func (GetMotionEventsRequest) OpCode() reqCode { return GetMotionEvents }

func parseGetMotionEventsRequest(order binary.ByteOrder, requestBody []byte) (*GetMotionEventsRequest, error) {
	if len(requestBody) < 12 {
		return nil, fmt.Errorf("%w: get motion events request too short", errParseError)
	}
	req := &GetMotionEventsRequest{}
	req.Window = Window(order.Uint32(requestBody[0:4]))
	req.Start = Timestamp(order.Uint32(requestBody[4:8]))
	req.Stop = Timestamp(order.Uint32(requestBody[8:12]))
	return req, nil
}

type TranslateCoordsRequest struct {
	SrcWindow Window
	DstWindow Window
	SrcX      int16
	SrcY      int16
}

func (TranslateCoordsRequest) OpCode() reqCode { return TranslateCoords }

func parseTranslateCoordsRequest(order binary.ByteOrder, requestBody []byte) (*TranslateCoordsRequest, error) {
	if len(requestBody) < 12 {
		return nil, fmt.Errorf("%w: translate coords request too short", errParseError)
	}
	req := &TranslateCoordsRequest{}
	req.SrcWindow = Window(order.Uint32(requestBody[0:4]))
	req.DstWindow = Window(order.Uint32(requestBody[4:8]))
	req.SrcX = int16(order.Uint16(requestBody[8:10]))
	req.SrcY = int16(order.Uint16(requestBody[10:12]))
	return req, nil
}

type WarpPointerRequest struct {
	SrcWindow uint32
	DstWindow uint32
	SrcX      int16
	SrcY      int16
	SrcWidth  uint16
	SrcHeight uint16
	DstX      int16
	DstY      int16
}

func (WarpPointerRequest) OpCode() reqCode { return WarpPointer }

func parseWarpPointerRequest(order binary.ByteOrder, payload []byte) (*WarpPointerRequest, error) {
	if len(payload) < 16 {
		return nil, fmt.Errorf("%w: warp pointer request too short", errParseError)
	}
	req := &WarpPointerRequest{}
	req.DstX = int16(order.Uint16(payload[12:14]))
	req.DstY = int16(order.Uint16(payload[14:16]))
	return req, nil
}

type SetInputFocusRequest struct {
	Focus    Window
	RevertTo byte
	Time     Timestamp
}

func (SetInputFocusRequest) OpCode() reqCode { return SetInputFocus }

func parseSetInputFocusRequest(order binary.ByteOrder, requestBody []byte) (*SetInputFocusRequest, error) {
	if len(requestBody) < 12 {
		return nil, fmt.Errorf("%w: set input focus request too short", errParseError)
	}
	req := &SetInputFocusRequest{}
	req.Focus = Window(order.Uint32(requestBody[0:4]))
	req.RevertTo = requestBody[4]
	req.Time = Timestamp(order.Uint32(requestBody[8:12]))
	return req, nil
}

type GetInputFocusRequest struct{}

func (GetInputFocusRequest) OpCode() reqCode { return GetInputFocus }

func parseGetInputFocusRequest(order binary.ByteOrder, requestBody []byte) (*GetInputFocusRequest, error) {
	return &GetInputFocusRequest{}, nil
}

type QueryKeymapRequest struct{}

func (QueryKeymapRequest) OpCode() reqCode { return QueryKeymap }

func parseQueryKeymapRequest(order binary.ByteOrder, requestBody []byte) (*QueryKeymapRequest, error) {
	return &QueryKeymapRequest{}, nil
}

type OpenFontRequest struct {
	Fid  Font
	Name string
}

func (OpenFontRequest) OpCode() reqCode { return OpenFont }

func parseOpenFontRequest(order binary.ByteOrder, requestBody []byte) (*OpenFontRequest, error) {
	if len(requestBody) < 8 {
		return nil, fmt.Errorf("%w: open font request too short", errParseError)
	}
	req := &OpenFontRequest{}
	req.Fid = Font(order.Uint32(requestBody[0:4]))
	nameLen := order.Uint16(requestBody[4:6])
	if len(requestBody) < 8+int(nameLen) {
		return nil, fmt.Errorf("%w: open font request too short for name", errParseError)
	}
	req.Name = string(requestBody[8 : 8+nameLen])
	return req, nil
}

type CloseFontRequest struct {
	Fid Font
}

func (CloseFontRequest) OpCode() reqCode { return CloseFont }

func parseCloseFontRequest(order binary.ByteOrder, requestBody []byte) (*CloseFontRequest, error) {
	if len(requestBody) < 4 {
		return nil, fmt.Errorf("%w: close font request too short", errParseError)
	}
	req := &CloseFontRequest{}
	req.Fid = Font(order.Uint32(requestBody[0:4]))
	return req, nil
}

type QueryFontRequest struct {
	Fid Font
}

func (QueryFontRequest) OpCode() reqCode { return QueryFont }

func parseQueryFontRequest(order binary.ByteOrder, requestBody []byte) (*QueryFontRequest, error) {
	if len(requestBody) < 4 {
		return nil, fmt.Errorf("%w: query font request too short", errParseError)
	}
	req := &QueryFontRequest{}
	req.Fid = Font(order.Uint32(requestBody[0:4]))
	return req, nil
}

type QueryTextExtentsRequest struct {
	Fid  Font
	Text []uint16
}

func (QueryTextExtentsRequest) OpCode() reqCode { return QueryTextExtents }

func parseQueryTextExtentsRequest(order binary.ByteOrder, requestBody []byte) (*QueryTextExtentsRequest, error) {
	if len(requestBody) < 4 {
		return nil, fmt.Errorf("%w: query text extents request too short", errParseError)
	}
	req := &QueryTextExtentsRequest{}
	req.Fid = Font(order.Uint32(requestBody[0:4]))
	for i := 4; i < len(requestBody); i += 2 {
		req.Text = append(req.Text, order.Uint16(requestBody[i:i+2]))
	}
	return req, nil
}

type ListFontsRequest struct {
	MaxNames uint16
	Pattern  string
}

func (ListFontsRequest) OpCode() reqCode { return ListFonts }

func parseListFontsRequest(order binary.ByteOrder, requestBody []byte) (*ListFontsRequest, error) {
	if len(requestBody) < 4 {
		return nil, fmt.Errorf("%w: list fonts request too short", errParseError)
	}
	req := &ListFontsRequest{}
	req.MaxNames = order.Uint16(requestBody[0:2])
	nameLen := order.Uint16(requestBody[2:4])
	if len(requestBody) < 4+int(nameLen) {
		return nil, fmt.Errorf("%w: list fonts request too short for pattern", errParseError)
	}
	req.Pattern = string(requestBody[4 : 4+nameLen])
	return req, nil
}

type ListFontsWithInfoRequest struct {
	MaxNames uint16
	Pattern  string
}

func (ListFontsWithInfoRequest) OpCode() reqCode { return ListFontsWithInfo }

func parseListFontsWithInfoRequest(order binary.ByteOrder, requestBody []byte) (*ListFontsWithInfoRequest, error) {
	if len(requestBody) < 4 {
		return nil, fmt.Errorf("%w: list fonts with info request too short", errParseError)
	}
	req := &ListFontsWithInfoRequest{}
	req.MaxNames = order.Uint16(requestBody[0:2])
	nameLen := order.Uint16(requestBody[2:4])
	if len(requestBody) < 4+int(nameLen) {
		return nil, fmt.Errorf("%w: list fonts with info request too short for pattern", errParseError)
	}
	req.Pattern = string(requestBody[4 : 4+nameLen])
	return req, nil
}

type SetFontPathRequest struct {
	NumPaths uint16
	Paths    []string
}

func (SetFontPathRequest) OpCode() reqCode { return SetFontPath }

func parseSetFontPathRequest(order binary.ByteOrder, requestBody []byte) (*SetFontPathRequest, error) {
	if len(requestBody) < 4 {
		return nil, fmt.Errorf("%w: set font path request too short", errParseError)
	}
	req := &SetFontPathRequest{}
	req.NumPaths = order.Uint16(requestBody[0:2])
	pathsData := requestBody[4:]
	for i := 0; i < int(req.NumPaths); i++ {
		if len(pathsData) == 0 {
			return nil, fmt.Errorf("%w: set font path request too short for path length", errParseError)
		}
		pathLen := int(pathsData[0])
		pathsData = pathsData[1:]
		if len(pathsData) < pathLen {
			return nil, fmt.Errorf("%w: set font path request too short for path", errParseError)
		}
		req.Paths = append(req.Paths, string(pathsData[:pathLen]))
		pathsData = pathsData[pathLen:]
	}
	return req, nil
}

type GetFontPathRequest struct{}

func (GetFontPathRequest) OpCode() reqCode { return GetFontPath }

func parseGetFontPathRequest(order binary.ByteOrder, requestBody []byte) (*GetFontPathRequest, error) {
	return &GetFontPathRequest{}, nil
}

type CreatePixmapRequest struct {
	Pid      Pixmap
	Drawable Drawable
	Width    uint16
	Height   uint16
	Depth    byte
}

func (CreatePixmapRequest) OpCode() reqCode { return CreatePixmap }

func parseCreatePixmapRequest(order binary.ByteOrder, data byte, payload []byte) (*CreatePixmapRequest, error) {
	if len(payload) < 12 {
		return nil, fmt.Errorf("%w: create pixmap request too short", errParseError)
	}
	req := &CreatePixmapRequest{}
	req.Depth = data
	req.Pid = Pixmap(order.Uint32(payload[0:4]))
	req.Drawable = Drawable(order.Uint32(payload[4:8]))
	req.Width = order.Uint16(payload[8:10])
	req.Height = order.Uint16(payload[10:12])
	return req, nil
}

type FreePixmapRequest struct {
	Pid Pixmap
}

func (FreePixmapRequest) OpCode() reqCode { return FreePixmap }

func parseFreePixmapRequest(order binary.ByteOrder, requestBody []byte) (*FreePixmapRequest, error) {
	if len(requestBody) < 4 {
		return nil, fmt.Errorf("%w: free pixmap request too short", errParseError)
	}
	req := &FreePixmapRequest{}
	req.Pid = Pixmap(order.Uint32(requestBody[0:4]))
	return req, nil
}

type CreateGCRequest struct {
	Cid       GContext
	Drawable  Drawable
	ValueMask uint32
	Values    GC
}

func (CreateGCRequest) OpCode() reqCode { return CreateGC }

func parseCreateGCRequest(order binary.ByteOrder, requestBody []byte) (*CreateGCRequest, error) {
	if len(requestBody) < 12 {
		return nil, fmt.Errorf("%w: create gc request too short", errParseError)
	}
	req := &CreateGCRequest{}
	req.Cid = GContext(order.Uint32(requestBody[0:4]))
	req.Drawable = Drawable(order.Uint32(requestBody[4:8]))
	req.ValueMask = order.Uint32(requestBody[8:12])
	values, _, err := parseGCValues(order, req.ValueMask, requestBody[12:])
	if err != nil {
		return nil, err
	}
	req.Values = values
	return req, nil
}

type ChangeGCRequest struct {
	Gc        GContext
	ValueMask uint32
	Values    GC
}

func (ChangeGCRequest) OpCode() reqCode { return ChangeGC }

func parseChangeGCRequest(order binary.ByteOrder, requestBody []byte) (*ChangeGCRequest, error) {
	if len(requestBody) < 8 {
		return nil, fmt.Errorf("%w: change gc request too short", errParseError)
	}
	req := &ChangeGCRequest{}
	req.Gc = GContext(order.Uint32(requestBody[0:4]))
	req.ValueMask = order.Uint32(requestBody[4:8])
	values, _, err := parseGCValues(order, req.ValueMask, requestBody[8:])
	if err != nil {
		return nil, err
	}
	req.Values = values
	return req, nil
}

type CopyGCRequest struct {
	SrcGC GContext
	DstGC GContext
}

func (CopyGCRequest) OpCode() reqCode { return CopyGC }

func parseCopyGCRequest(order binary.ByteOrder, requestBody []byte) (*CopyGCRequest, error) {
	if len(requestBody) < 8 {
		return nil, fmt.Errorf("%w: copy gc request too short", errParseError)
	}
	req := &CopyGCRequest{}
	req.SrcGC = GContext(order.Uint32(requestBody[0:4]))
	req.DstGC = GContext(order.Uint32(requestBody[4:8]))
	return req, nil
}

type SetDashesRequest struct {
	GC         GContext
	DashOffset uint16
	Dashes     []byte
}

func (SetDashesRequest) OpCode() reqCode { return SetDashes }

func parseSetDashesRequest(order binary.ByteOrder, requestBody []byte) (*SetDashesRequest, error) {
	if len(requestBody) < 8 {
		return nil, fmt.Errorf("%w: set dashes request too short", errParseError)
	}
	req := &SetDashesRequest{}
	req.GC = GContext(order.Uint32(requestBody[0:4]))
	req.DashOffset = order.Uint16(requestBody[4:6])
	nDashes := order.Uint16(requestBody[6:8])
	if len(requestBody) < 8+int(nDashes) {
		return nil, fmt.Errorf("%w: set dashes request too short for dashes", errParseError)
	}
	req.Dashes = requestBody[8 : 8+nDashes]
	return req, nil
}

type SetClipRectanglesRequest struct {
	GC         GContext
	ClippingX  int16
	ClippingY  int16
	Rectangles []Rectangle
	Ordering   byte
}

func (SetClipRectanglesRequest) OpCode() reqCode { return SetClipRectangles }

func parseSetClipRectanglesRequest(order binary.ByteOrder, data byte, requestBody []byte) (*SetClipRectanglesRequest, error) {
	if len(requestBody) < 8 {
		return nil, fmt.Errorf("%w: set clip rectangles request too short", errParseError)
	}
	req := &SetClipRectanglesRequest{}
	req.Ordering = data
	req.GC = GContext(order.Uint32(requestBody[0:4]))
	req.ClippingX = int16(order.Uint16(requestBody[4:6]))
	req.ClippingY = int16(order.Uint16(requestBody[6:8]))
	numRects := (len(requestBody) - 8) / 8
	for i := 0; i < numRects; i++ {
		offset := 8 + i*8
		rect := Rectangle{
			X:      int16(order.Uint16(requestBody[offset : offset+2])),
			Y:      int16(order.Uint16(requestBody[offset+2 : offset+4])),
			Width:  order.Uint16(requestBody[offset+4 : offset+6]),
			Height: order.Uint16(requestBody[offset+6 : offset+8]),
		}
		req.Rectangles = append(req.Rectangles, rect)
	}
	return req, nil
}

type FreeGCRequest struct {
	GC GContext
}

func (FreeGCRequest) OpCode() reqCode { return FreeGC }

func parseFreeGCRequest(order binary.ByteOrder, requestBody []byte) (*FreeGCRequest, error) {
	if len(requestBody) < 4 {
		return nil, fmt.Errorf("%w: free gc request too short", errParseError)
	}
	req := &FreeGCRequest{}
	req.GC = GContext(order.Uint32(requestBody[0:4]))
	return req, nil
}

type ClearAreaRequest struct {
	Exposures bool
	Window    Window
	X         int16
	Y         int16
	Width     uint16
	Height    uint16
}

func (ClearAreaRequest) OpCode() reqCode { return ClearArea }

func parseClearAreaRequest(order binary.ByteOrder, requestBody []byte) (*ClearAreaRequest, error) {
	if len(requestBody) < 12 {
		return nil, fmt.Errorf("%w: clear area request too short", errParseError)
	}
	req := &ClearAreaRequest{}
	req.Window = Window(order.Uint32(requestBody[0:4]))
	req.X = int16(order.Uint16(requestBody[4:6]))
	req.Y = int16(order.Uint16(requestBody[6:8]))
	req.Width = order.Uint16(requestBody[8:10])
	req.Height = order.Uint16(requestBody[10:12])
	return req, nil
}

type CopyAreaRequest struct {
	SrcDrawable Drawable
	DstDrawable Drawable
	Gc          GContext
	SrcX        int16
	SrcY        int16
	DstX        int16
	DstY        int16
	Width       uint16
	Height      uint16
}

func (CopyAreaRequest) OpCode() reqCode { return CopyArea }

func parseCopyAreaRequest(order binary.ByteOrder, requestBody []byte) (*CopyAreaRequest, error) {
	if len(requestBody) < 28 {
		return nil, fmt.Errorf("%w: copy area request too short", errParseError)
	}
	req := &CopyAreaRequest{}
	req.SrcDrawable = Drawable(order.Uint32(requestBody[0:4]))
	req.DstDrawable = Drawable(order.Uint32(requestBody[4:8]))
	req.Gc = GContext(order.Uint32(requestBody[8:12]))
	req.SrcX = int16(order.Uint16(requestBody[12:14]))
	req.SrcY = int16(order.Uint16(requestBody[14:16]))
	req.DstX = int16(order.Uint16(requestBody[16:18]))
	req.DstY = int16(order.Uint16(requestBody[18:20]))
	req.Width = order.Uint16(requestBody[20:22])
	req.Height = order.Uint16(requestBody[22:24])
	return req, nil
}

type PolyPointRequest struct {
	Drawable    Drawable
	Gc          GContext
	Coordinates []uint32
}

func (PolyPointRequest) OpCode() reqCode { return PolyPoint }

func parsePolyPointRequest(order binary.ByteOrder, requestBody []byte) (*PolyPointRequest, error) {
	if len(requestBody) < 8 {
		return nil, fmt.Errorf("%w: poly point request too short", errParseError)
	}
	req := &PolyPointRequest{}
	req.Drawable = Drawable(order.Uint32(requestBody[0:4]))
	req.Gc = GContext(order.Uint32(requestBody[4:8]))
	numPoints := (len(requestBody) - 8) / 4
	for i := 0; i < numPoints; i++ {
		offset := 8 + i*4
		x := int32(order.Uint16(requestBody[offset : offset+2]))
		y := int32(order.Uint16(requestBody[offset+2 : offset+4]))
		req.Coordinates = append(req.Coordinates, uint32(x), uint32(y))
	}
	return req, nil
}

type PolyLineRequest struct {
	Drawable    Drawable
	Gc          GContext
	Coordinates []uint32
}

func (PolyLineRequest) OpCode() reqCode { return PolyLine }

func parsePolyLineRequest(order binary.ByteOrder, requestBody []byte) (*PolyLineRequest, error) {
	if len(requestBody) < 8 {
		return nil, fmt.Errorf("%w: poly line request too short", errParseError)
	}
	req := &PolyLineRequest{}
	req.Drawable = Drawable(order.Uint32(requestBody[0:4]))
	req.Gc = GContext(order.Uint32(requestBody[4:8]))
	numPoints := (len(requestBody) - 8) / 4
	for i := 0; i < numPoints; i++ {
		offset := 8 + i*4
		x := int32(order.Uint16(requestBody[offset : offset+2]))
		y := int32(order.Uint16(requestBody[offset+2 : offset+4]))
		req.Coordinates = append(req.Coordinates, uint32(x), uint32(y))
	}
	return req, nil
}

type PolySegmentRequest struct {
	Drawable Drawable
	Gc       GContext
	Segments []uint32
}

func (PolySegmentRequest) OpCode() reqCode { return PolySegment }

func parsePolySegmentRequest(order binary.ByteOrder, requestBody []byte) (*PolySegmentRequest, error) {
	if len(requestBody) < 8 {
		return nil, fmt.Errorf("%w: poly segment request too short", errParseError)
	}
	req := &PolySegmentRequest{}
	req.Drawable = Drawable(order.Uint32(requestBody[0:4]))
	req.Gc = GContext(order.Uint32(requestBody[4:8]))
	numSegments := (len(requestBody) - 8) / 8
	for i := 0; i < numSegments; i++ {
		offset := 8 + i*8
		x1 := int32(order.Uint16(requestBody[offset : offset+2]))
		y1 := int32(order.Uint16(requestBody[offset+2 : offset+4]))
		x2 := int32(order.Uint16(requestBody[offset+4 : offset+6]))
		y2 := int32(order.Uint16(requestBody[offset+6 : offset+8]))
		req.Segments = append(req.Segments, uint32(x1), uint32(y1), uint32(x2), uint32(y2))
	}
	return req, nil
}

type PolyRectangleRequest struct {
	Drawable   Drawable
	Gc         GContext
	Rectangles []uint32
}

func (PolyRectangleRequest) OpCode() reqCode { return PolyRectangle }

func parsePolyRectangleRequest(order binary.ByteOrder, requestBody []byte) (*PolyRectangleRequest, error) {
	if len(requestBody) < 8 {
		return nil, fmt.Errorf("%w: poly rectangle request too short", errParseError)
	}
	req := &PolyRectangleRequest{}
	req.Drawable = Drawable(order.Uint32(requestBody[0:4]))
	req.Gc = GContext(order.Uint32(requestBody[4:8]))
	numRects := (len(requestBody) - 8) / 8
	for i := 0; i < numRects; i++ {
		offset := 8 + i*8
		x := uint32(order.Uint16(requestBody[offset : offset+2]))
		y := uint32(order.Uint16(requestBody[offset+2 : offset+4]))
		width := uint32(order.Uint16(requestBody[offset+4 : offset+6]))
		height := uint32(order.Uint16(requestBody[offset+6 : offset+8]))
		req.Rectangles = append(req.Rectangles, x, y, width, height)
	}
	return req, nil
}

type PolyArcRequest struct {
	Drawable Drawable
	Gc       GContext
	Arcs     []uint32
}

func (PolyArcRequest) OpCode() reqCode { return PolyArc }

func parsePolyArcRequest(order binary.ByteOrder, requestBody []byte) (*PolyArcRequest, error) {
	if len(requestBody) < 8 {
		return nil, fmt.Errorf("%w: poly arc request too short", errParseError)
	}
	req := &PolyArcRequest{}
	req.Drawable = Drawable(order.Uint32(requestBody[0:4]))
	req.Gc = GContext(order.Uint32(requestBody[4:8]))
	numArcs := (len(requestBody) - 8) / 12
	for i := 0; i < numArcs; i++ {
		offset := 8 + i*12
		x := int32(order.Uint16(requestBody[offset : offset+2]))
		y := int32(order.Uint16(requestBody[offset+2 : offset+4]))
		width := uint32(order.Uint16(requestBody[offset+4 : offset+6]))
		height := uint32(order.Uint16(requestBody[offset+6 : offset+8]))
		angle1 := int32(order.Uint16(requestBody[offset+8 : offset+10]))
		angle2 := int32(order.Uint16(requestBody[offset+10 : offset+12]))
		req.Arcs = append(req.Arcs, uint32(x), uint32(y), width, height, uint32(angle1), uint32(angle2))
	}
	return req, nil
}

type FillPolyRequest struct {
	Drawable    Drawable
	Gc          GContext
	Shape       byte
	Coordinates []uint32
}

func (FillPolyRequest) OpCode() reqCode { return FillPoly }

func parseFillPolyRequest(order binary.ByteOrder, requestBody []byte) (*FillPolyRequest, error) {
	if len(requestBody) < 12 {
		return nil, fmt.Errorf("%w: fill poly request too short", errParseError)
	}
	req := &FillPolyRequest{}
	req.Drawable = Drawable(order.Uint32(requestBody[0:4]))
	req.Gc = GContext(order.Uint32(requestBody[4:8]))
	numPoints := (len(requestBody) - 12) / 4
	for i := 0; i < numPoints; i++ {
		offset := 12 + i*4
		x := int32(order.Uint16(requestBody[offset : offset+2]))
		y := int32(order.Uint16(requestBody[offset+2 : offset+4]))
		req.Coordinates = append(req.Coordinates, uint32(x), uint32(y))
	}
	return req, nil
}

type PolyFillRectangleRequest struct {
	Drawable   Drawable
	Gc         GContext
	Rectangles []uint32
}

func (PolyFillRectangleRequest) OpCode() reqCode { return PolyFillRectangle }

func parsePolyFillRectangleRequest(order binary.ByteOrder, requestBody []byte) (*PolyFillRectangleRequest, error) {
	if len(requestBody) < 8 {
		return nil, fmt.Errorf("%w: poly fill rectangle request too short", errParseError)
	}
	req := &PolyFillRectangleRequest{}
	req.Drawable = Drawable(order.Uint32(requestBody[0:4]))
	req.Gc = GContext(order.Uint32(requestBody[4:8]))
	numRects := (len(requestBody) - 8) / 8
	for i := 0; i < numRects; i++ {
		offset := 8 + i*8
		x := uint32(order.Uint16(requestBody[offset : offset+2]))
		y := uint32(order.Uint16(requestBody[offset+2 : offset+4]))
		width := uint32(order.Uint16(requestBody[offset+4 : offset+6]))
		height := uint32(order.Uint16(requestBody[offset+6 : offset+8]))
		req.Rectangles = append(req.Rectangles, x, y, width, height)
	}
	return req, nil
}

type PolyFillArcRequest struct {
	Drawable Drawable
	Gc       GContext
	Arcs     []uint32
}

func (PolyFillArcRequest) OpCode() reqCode { return PolyFillArc }

func parsePolyFillArcRequest(order binary.ByteOrder, requestBody []byte) (*PolyFillArcRequest, error) {
	if len(requestBody) < 8 {
		return nil, fmt.Errorf("%w: poly fill arc request too short", errParseError)
	}
	req := &PolyFillArcRequest{}
	req.Drawable = Drawable(order.Uint32(requestBody[0:4]))
	req.Gc = GContext(order.Uint32(requestBody[4:8]))
	numArcs := (len(requestBody) - 8) / 12
	for i := 0; i < numArcs; i++ {
		offset := 8 + i*12
		x := int32(order.Uint16(requestBody[offset : offset+2]))
		y := int32(order.Uint16(requestBody[offset+2 : offset+4]))
		width := uint32(order.Uint16(requestBody[offset+4 : offset+6]))
		height := uint32(order.Uint16(requestBody[offset+6 : offset+8]))
		angle1 := int32(order.Uint16(requestBody[offset+8 : offset+10]))
		angle2 := int32(order.Uint16(requestBody[offset+10 : offset+12]))
		req.Arcs = append(req.Arcs, uint32(x), uint32(y), width, height, uint32(angle1), uint32(angle2))
	}
	return req, nil
}

type PutImageRequest struct {
	Drawable Drawable
	Gc       GContext
	Width    uint16
	Height   uint16
	DstX     int16
	DstY     int16
	LeftPad  byte
	Depth    byte
	Format   byte
	Data     []byte
}

func (PutImageRequest) OpCode() reqCode { return PutImage }

func parsePutImageRequest(order binary.ByteOrder, data byte, requestBody []byte) (*PutImageRequest, error) {
	if len(requestBody) < 20 {
		return nil, fmt.Errorf("%w: put image request too short", errParseError)
	}
	req := &PutImageRequest{}
	req.Format = data
	req.Drawable = Drawable(order.Uint32(requestBody[0:4]))
	req.Gc = GContext(order.Uint32(requestBody[4:8]))
	req.Width = order.Uint16(requestBody[8:10])
	req.Height = order.Uint16(requestBody[10:12])
	req.DstX = int16(order.Uint16(requestBody[12:14]))
	req.DstY = int16(order.Uint16(requestBody[14:16]))
	req.LeftPad = requestBody[16]
	req.Depth = requestBody[17]
	req.Data = requestBody[20:]
	return req, nil
}

type GetImageRequest struct {
	Drawable  Drawable
	X         int16
	Y         int16
	Width     uint16
	Height    uint16
	PlaneMask uint32
	Format    byte
}

func (GetImageRequest) OpCode() reqCode { return GetImage }

func parseGetImageRequest(order binary.ByteOrder, data byte, requestBody []byte) (*GetImageRequest, error) {
	if len(requestBody) < 16 {
		return nil, fmt.Errorf("%w: get image request too short", errParseError)
	}
	req := &GetImageRequest{}
	req.Format = data
	req.Drawable = Drawable(order.Uint32(requestBody[0:4]))
	req.X = int16(order.Uint16(requestBody[4:6]))
	req.Y = int16(order.Uint16(requestBody[6:8]))
	req.Width = order.Uint16(requestBody[8:10])
	req.Height = order.Uint16(requestBody[10:12])
	req.PlaneMask = order.Uint32(requestBody[12:16])
	return req, nil
}

type PolyText8Request struct {
	Drawable Drawable
	GC       GContext
	X, Y     int16
	Items    []PolyTextItem
}

func (PolyText8Request) OpCode() reqCode { return PolyText8 }

func parsePolyText8Request(order binary.ByteOrder, data []byte) (*PolyText8Request, error) {
	var req PolyText8Request
	if len(data) < 12 {
		return nil, fmt.Errorf("poly text 8 request too short for header")
	}
	req.Drawable = Drawable(order.Uint32(data[0:4]))
	req.GC = GContext(order.Uint32(data[4:8]))
	req.X = int16(order.Uint16(data[8:10]))
	req.Y = int16(order.Uint16(data[10:12]))

	i := 12
	for i < len(data) {
		if i+1 > len(data) {
			break
		}
		length := int(data[i])
		if length == 0 { // Invalid length, must be padding
			break
		}
		if length == 255 {
			itemSize := 5
			if i+itemSize > len(data) {
				break
			}
			font := Font(order.Uint32(data[i+1 : i+5]))
			req.Items = append(req.Items, PolyTextFont{Font: font})
			i += itemSize
		} else {
			itemSize := 2 + length
			if i+itemSize > len(data) {
				break
			}
			delta := int8(data[i+1])
			str := data[i+2 : i+2+length]
			req.Items = append(req.Items, PolyText8String{Delta: delta, Str: str})
			i += itemSize
		}
	}
	return &req, nil
}

type PolyText16Request struct {
	Drawable Drawable
	GC       GContext
	X, Y     int16
	Items    []PolyTextItem
}

func (PolyText16Request) OpCode() reqCode { return PolyText16 }

func parsePolyText16Request(order binary.ByteOrder, data []byte) (*PolyText16Request, error) {
	var req PolyText16Request
	if len(data) < 12 {
		return nil, fmt.Errorf("poly text 16 request too short for header")
	}
	req.Drawable = Drawable(order.Uint32(data[0:4]))
	req.GC = GContext(order.Uint32(data[4:8]))
	req.X = int16(order.Uint16(data[8:10]))
	req.Y = int16(order.Uint16(data[10:12]))

	i := 12
	for i < len(data) {
		if i+1 > len(data) {
			break
		}
		length := int(data[i])
		if length == 0 { // Invalid length, must be padding
			break
		}
		if length == 255 {
			itemSize := 5
			if i+itemSize > len(data) {
				break
			}
			font := Font(order.Uint32(data[i+1 : i+5]))
			req.Items = append(req.Items, PolyTextFont{Font: font})
			i += itemSize
		} else {
			itemSize := 2 + length*2
			if i+itemSize > len(data) {
				break
			}

			delta := int8(data[i+1])
			var str []uint16
			for j := 0; j < length; j++ {
				str = append(str, order.Uint16(data[i+2+j*2:i+2+(j+1)*2]))
			}
			req.Items = append(req.Items, PolyText16String{Delta: delta, Str: str})
			i += itemSize
		}
	}
	return &req, nil
}

type ImageText8Request struct {
	Drawable Drawable
	Gc       GContext
	X        int16
	Y        int16
	Text     []byte
}

func (ImageText8Request) OpCode() reqCode { return ImageText8 }

func parseImageText8Request(order binary.ByteOrder, requestBody []byte) (*ImageText8Request, error) {
	if len(requestBody) < 12 {
		return nil, fmt.Errorf("%w: image text 8 request too short", errParseError)
	}
	req := &ImageText8Request{}
	req.Drawable = Drawable(order.Uint32(requestBody[0:4]))
	req.Gc = GContext(order.Uint32(requestBody[4:8]))
	req.X = int16(order.Uint16(requestBody[8:10]))
	req.Y = int16(order.Uint16(requestBody[10:12]))
	req.Text = requestBody[12:]
	return req, nil
}

type ImageText16Request struct {
	Drawable Drawable
	Gc       GContext
	X        int16
	Y        int16
	Text     []uint16
}

func (ImageText16Request) OpCode() reqCode { return ImageText16 }

func parseImageText16Request(order binary.ByteOrder, requestBody []byte) (*ImageText16Request, error) {
	if len(requestBody) < 12 {
		return nil, fmt.Errorf("%w: image text 16 request too short", errParseError)
	}
	req := &ImageText16Request{}
	req.Drawable = Drawable(order.Uint32(requestBody[0:4]))
	req.Gc = GContext(order.Uint32(requestBody[4:8]))
	req.X = int16(order.Uint16(requestBody[8:10]))
	req.Y = int16(order.Uint16(requestBody[10:12]))
	for i := 12; i < len(requestBody); i += 2 {
		req.Text = append(req.Text, order.Uint16(requestBody[i:i+2]))
	}
	return req, nil
}

type CreateColormapRequest struct {
	Alloc  byte
	Mid    Colormap
	Window Window
	Visual VisualID
}

func (CreateColormapRequest) OpCode() reqCode { return CreateColormap }

func parseCreateColormapRequest(order binary.ByteOrder, data byte, payload []byte) (*CreateColormapRequest, error) {
	if len(payload) < 12 {
		return nil, fmt.Errorf("%w: create colormap request too short", errParseError)
	}
	req := &CreateColormapRequest{}
	req.Alloc = data
	req.Mid = Colormap(order.Uint32(payload[0:4]))
	req.Window = Window(order.Uint32(payload[4:8]))
	req.Visual = VisualID(order.Uint32(payload[8:12]))
	return req, nil
}

type FreeColormapRequest struct {
	Cmap Colormap
}

func (FreeColormapRequest) OpCode() reqCode { return FreeColormap }

func parseFreeColormapRequest(order binary.ByteOrder, requestBody []byte) (*FreeColormapRequest, error) {
	if len(requestBody) < 4 {
		return nil, fmt.Errorf("%w: free colormap request too short", errParseError)
	}
	req := &FreeColormapRequest{}
	req.Cmap = Colormap(order.Uint32(requestBody[0:4]))
	return req, nil
}

type InstallColormapRequest struct {
	Cmap Colormap
}

func (InstallColormapRequest) OpCode() reqCode { return InstallColormap }

func parseInstallColormapRequest(order binary.ByteOrder, requestBody []byte) (*InstallColormapRequest, error) {
	if len(requestBody) < 4 {
		return nil, fmt.Errorf("%w: install colormap request too short", errParseError)
	}
	req := &InstallColormapRequest{}
	req.Cmap = Colormap(order.Uint32(requestBody[0:4]))
	return req, nil
}

type UninstallColormapRequest struct {
	Cmap Colormap
}

func (UninstallColormapRequest) OpCode() reqCode { return UninstallColormap }

func parseUninstallColormapRequest(order binary.ByteOrder, requestBody []byte) (*UninstallColormapRequest, error) {
	if len(requestBody) < 4 {
		return nil, fmt.Errorf("%w: uninstall colormap request too short", errParseError)
	}
	req := &UninstallColormapRequest{}
	req.Cmap = Colormap(order.Uint32(requestBody[0:4]))
	return req, nil
}

type ListInstalledColormapsRequest struct {
	Window Window
}

func (ListInstalledColormapsRequest) OpCode() reqCode { return ListInstalledColormaps }

func parseListInstalledColormapsRequest(order binary.ByteOrder, requestBody []byte) (*ListInstalledColormapsRequest, error) {
	if len(requestBody) < 4 {
		return nil, fmt.Errorf("%w: list installed colormaps request too short", errParseError)
	}
	req := &ListInstalledColormapsRequest{}
	req.Window = Window(order.Uint32(requestBody[0:4]))
	return req, nil
}

type AllocColorRequest struct {
	Cmap  Colormap
	Red   uint16
	Green uint16
	Blue  uint16
}

func (AllocColorRequest) OpCode() reqCode { return AllocColor }

func parseAllocColorRequest(order binary.ByteOrder, payload []byte) (*AllocColorRequest, error) {
	if len(payload) < 10 {
		return nil, fmt.Errorf("%w: alloc color request too short", errParseError)
	}
	req := &AllocColorRequest{}
	req.Cmap = Colormap(order.Uint32(payload[0:4]))
	req.Red = order.Uint16(payload[4:6])
	req.Green = order.Uint16(payload[6:8])
	req.Blue = order.Uint16(payload[8:10])
	return req, nil
}

type AllocNamedColorRequest struct {
	Cmap xID
	Name []byte
}

func (AllocNamedColorRequest) OpCode() reqCode { return AllocNamedColor }

func parseAllocNamedColorRequest(order binary.ByteOrder, payload []byte) (*AllocNamedColorRequest, error) {
	if len(payload) < 8 {
		return nil, fmt.Errorf("%w: alloc named color request too short", errParseError)
	}
	req := &AllocNamedColorRequest{}
	req.Cmap = xID{local: order.Uint32(payload[0:4])}
	nameLen := order.Uint16(payload[4:6])
	if len(payload) < 8+int(nameLen) {
		return nil, fmt.Errorf("%w: alloc named color request too short for name", errParseError)
	}
	req.Name = payload[8 : 8+nameLen]
	return req, nil
}

type FreeColorsRequest struct {
	Cmap      Colormap
	PlaneMask uint32
	Pixels    []uint32
}

func (FreeColorsRequest) OpCode() reqCode { return FreeColors }

/*
FreeColors

	1     88                              opcode
	1                                     unused
	2     3+n                             request length
	4     COLORMAP                        cmap
	4     CARD32                          plane-mask
	4n     LISTofCARD32                   pixels
*/
func parseFreeColorsRequest(order binary.ByteOrder, requestBody []byte) (*FreeColorsRequest, error) {
	if len(requestBody) < 8 {
		return nil, fmt.Errorf("%w: free colors request too short", errParseError)
	}
	req := &FreeColorsRequest{}
	req.Cmap = Colormap(order.Uint32(requestBody[0:4]))
	req.PlaneMask = order.Uint32(requestBody[4:8])
	numPixels := (len(requestBody) - 8) / 4
	if len(requestBody) < 8+numPixels*4 {
		return nil, fmt.Errorf("%w: free colors request too short for %d colors", errParseError, numPixels)
	}
	for i := 0; i < numPixels; i++ {
		offset := 8 + i*4
		req.Pixels = append(req.Pixels, order.Uint32(requestBody[offset:offset+4]))
	}
	return req, nil
}

type StoreColorsRequest struct {
	Cmap  Colormap
	Items []struct {
		Pixel uint32
		Red   uint16
		Green uint16
		Blue  uint16
		Flags byte
	}
}

func (StoreColorsRequest) OpCode() reqCode { return StoreColors }

func parseStoreColorsRequest(order binary.ByteOrder, requestBody []byte) (*StoreColorsRequest, error) {
	if len(requestBody) < 4 {
		return nil, fmt.Errorf("%w: store colors request too short", errParseError)
	}
	req := &StoreColorsRequest{}
	req.Cmap = Colormap(order.Uint32(requestBody[0:4]))
	numItems := (len(requestBody) - 4) / 12
	for i := 0; i < numItems; i++ {
		offset := 4 + i*12
		item := struct {
			Pixel uint32
			Red   uint16
			Green uint16
			Blue  uint16
			Flags byte
		}{
			Pixel: order.Uint32(requestBody[offset : offset+4]),
			Red:   order.Uint16(requestBody[offset+4 : offset+6]),
			Green: order.Uint16(requestBody[offset+6 : offset+8]),
			Blue:  order.Uint16(requestBody[offset+8 : offset+10]),
			Flags: requestBody[offset+10],
		}
		req.Items = append(req.Items, item)
	}
	return req, nil
}

type StoreNamedColorRequest struct {
	Cmap  Colormap
	Pixel uint32
	Name  string
	Flags byte
}

func (StoreNamedColorRequest) OpCode() reqCode { return StoreNamedColor }

func parseStoreNamedColorRequest(order binary.ByteOrder, data byte, requestBody []byte) (*StoreNamedColorRequest, error) {
	if len(requestBody) < 12 {
		return nil, fmt.Errorf("%w: store named color request too short", errParseError)
	}
	req := &StoreNamedColorRequest{}
	req.Cmap = Colormap(order.Uint32(requestBody[0:4]))
	req.Pixel = order.Uint32(requestBody[4:8])
	nameLen := order.Uint16(requestBody[8:10])
	if len(requestBody) < 12+int(nameLen) {
		return nil, fmt.Errorf("%w: store named color request too short for name", errParseError)
	}
	req.Name = string(requestBody[12 : 12+nameLen])
	req.Flags = data
	return req, nil
}

type QueryColorsRequest struct {
	Cmap   xID
	Pixels []uint32
}

func (QueryColorsRequest) OpCode() reqCode { return QueryColors }

func parseQueryColorsRequest(order binary.ByteOrder, requestBody []byte) (*QueryColorsRequest, error) {
	if len(requestBody) < 4 {
		return nil, fmt.Errorf("%w: query colors request too short", errParseError)
	}
	req := &QueryColorsRequest{}
	req.Cmap = xID{local: order.Uint32(requestBody[0:4])}
	numPixels := (len(requestBody) - 4) / 4
	for i := 0; i < numPixels; i++ {
		offset := 4 + i*4
		req.Pixels = append(req.Pixels, order.Uint32(requestBody[offset:offset+4]))
	}
	return req, nil
}

type LookupColorRequest struct {
	Cmap Colormap
	Name string
}

func (LookupColorRequest) OpCode() reqCode { return LookupColor }

func parseLookupColorRequest(order binary.ByteOrder, payload []byte) (*LookupColorRequest, error) {
	if len(payload) < 8 {
		return nil, fmt.Errorf("%w: lookup color request too short", errParseError)
	}
	req := &LookupColorRequest{}
	req.Cmap = Colormap(order.Uint32(payload[0:4]))
	nameLen := order.Uint16(payload[4:6])
	if len(payload) < 8+int(nameLen) {
		return nil, fmt.Errorf("%w: lookup color request too short for name", errParseError)
	}
	req.Name = string(payload[8 : 8+nameLen])
	return req, nil
}

type CreateGlyphCursorRequest struct {
	Cid        Cursor
	SourceFont Font
	MaskFont   Font
	SourceChar uint16
	MaskChar   uint16
	ForeColor  [3]uint16
	BackColor  [3]uint16
}

func (CreateGlyphCursorRequest) OpCode() reqCode { return CreateGlyphCursor }

func parseCreateGlyphCursorRequest(order binary.ByteOrder, requestBody []byte) (*CreateGlyphCursorRequest, error) {
	if len(requestBody) < 28 {
		return nil, fmt.Errorf("%w: create glyph cursor request too short", errParseError)
	}
	req := &CreateGlyphCursorRequest{}
	req.Cid = Cursor(order.Uint32(requestBody[0:4]))
	req.SourceFont = Font(order.Uint32(requestBody[4:8]))
	req.MaskFont = Font(order.Uint32(requestBody[8:12]))
	req.SourceChar = order.Uint16(requestBody[12:14])
	req.MaskChar = order.Uint16(requestBody[14:16])
	req.ForeColor[0] = order.Uint16(requestBody[16:18])
	req.ForeColor[1] = order.Uint16(requestBody[18:20])
	req.ForeColor[2] = order.Uint16(requestBody[20:22])
	req.BackColor[0] = order.Uint16(requestBody[22:24])
	req.BackColor[1] = order.Uint16(requestBody[24:26])
	req.BackColor[2] = order.Uint16(requestBody[26:28])
	return req, nil
}

type FreeCursorRequest struct {
	Cursor Cursor
}

func (FreeCursorRequest) OpCode() reqCode { return FreeCursor }

func parseFreeCursorRequest(order binary.ByteOrder, requestBody []byte) (*FreeCursorRequest, error) {
	if len(requestBody) < 4 {
		return nil, fmt.Errorf("%w: free cursor request too short", errParseError)
	}
	req := &FreeCursorRequest{}
	req.Cursor = Cursor(order.Uint32(requestBody[0:4]))
	return req, nil
}

type RecolorCursorRequest struct {
	Cursor    Cursor
	ForeColor [3]uint16
	BackColor [3]uint16
}

func (RecolorCursorRequest) OpCode() reqCode { return RecolorCursor }

func parseRecolorCursorRequest(order binary.ByteOrder, requestBody []byte) (*RecolorCursorRequest, error) {
	if len(requestBody) < 16 {
		return nil, fmt.Errorf("%w: recolor cursor request too short", errParseError)
	}
	req := &RecolorCursorRequest{}
	req.Cursor = Cursor(order.Uint32(requestBody[0:4]))
	req.ForeColor[0] = order.Uint16(requestBody[4:6])
	req.ForeColor[1] = order.Uint16(requestBody[6:8])
	req.ForeColor[2] = order.Uint16(requestBody[8:10])
	req.BackColor[0] = order.Uint16(requestBody[10:12])
	req.BackColor[1] = order.Uint16(requestBody[12:14])
	req.BackColor[2] = order.Uint16(requestBody[14:16])
	return req, nil
}

type QueryBestSizeRequest struct {
	Class    byte
	Drawable Drawable
	Width    uint16
	Height   uint16
}

func (QueryBestSizeRequest) OpCode() reqCode { return QueryBestSize }

func parseQueryBestSizeRequest(order binary.ByteOrder, requestBody []byte) (*QueryBestSizeRequest, error) {
	if len(requestBody) < 8 {
		return nil, fmt.Errorf("%w: query best size request too short", errParseError)
	}
	req := &QueryBestSizeRequest{}
	req.Drawable = Drawable(order.Uint32(requestBody[0:4]))
	req.Width = order.Uint16(requestBody[4:6])
	req.Height = order.Uint16(requestBody[6:8])
	return req, nil
}

type QueryExtensionRequest struct {
	Name string
}

func (QueryExtensionRequest) OpCode() reqCode { return QueryExtension }

func parseQueryExtensionRequest(order binary.ByteOrder, requestBody []byte) (*QueryExtensionRequest, error) {
	if len(requestBody) < 4 {
		return nil, fmt.Errorf("%w: query extension request too short", errParseError)
	}
	req := &QueryExtensionRequest{}
	nameLen := order.Uint16(requestBody[0:2])
	if len(requestBody) < 4+int(nameLen) {
		return nil, fmt.Errorf("%w: query extension request too short for name", errParseError)
	}
	req.Name = string(requestBody[4 : 4+nameLen])
	return req, nil
}

type BellRequest struct {
	Percent int8
}

func (BellRequest) OpCode() reqCode { return Bell }

func parseBellRequest(requestBody byte) (*BellRequest, error) {
	req := &BellRequest{}
	req.Percent = int8(requestBody)
	return req, nil
}

type SetPointerMappingRequest struct {
	Map []byte
}

func (SetPointerMappingRequest) OpCode() reqCode { return SetPointerMapping }

func parseSetPointerMappingRequest(order binary.ByteOrder, requestBody []byte) (*SetPointerMappingRequest, error) {
	req := &SetPointerMappingRequest{}
	req.Map = requestBody
	return req, nil
}

type GetPointerMappingRequest struct{}

func (GetPointerMappingRequest) OpCode() reqCode { return GetPointerMapping }

func parseGetPointerMappingRequest(order binary.ByteOrder, requestBody []byte) (*GetPointerMappingRequest, error) {
	return &GetPointerMappingRequest{}, nil
}

type GetKeyboardMappingRequest struct {
	FirstKeyCode KeyCode
	Count        byte
}

func (GetKeyboardMappingRequest) OpCode() reqCode { return GetKeyboardMapping }

func parseGetKeyboardMappingRequest(order binary.ByteOrder, requestBody []byte) (*GetKeyboardMappingRequest, error) {
	if len(requestBody) < 2 {
		return nil, fmt.Errorf("%w: get keyboard mapping request too short", errParseError)
	}
	req := &GetKeyboardMappingRequest{}
	req.FirstKeyCode = KeyCode(requestBody[0])
	req.Count = requestBody[1]
	return req, nil
}

type ChangeKeyboardMappingRequest struct {
	KeyCodeCount      byte
	FirstKeyCode      KeyCode
	KeySymsPerKeyCode byte
	KeySyms           []uint32
}

func (ChangeKeyboardMappingRequest) OpCode() reqCode { return ChangeKeyboardMapping }

func parseChangeKeyboardMappingRequest(order binary.ByteOrder, data byte, requestBody []byte) (*ChangeKeyboardMappingRequest, error) {
	if len(requestBody) < 4 {
		return nil, fmt.Errorf("%w: change keyboard mapping request too short", errParseError)
	}
	req := &ChangeKeyboardMappingRequest{}
	req.KeyCodeCount = data
	req.FirstKeyCode = KeyCode(requestBody[0])
	req.KeySymsPerKeyCode = requestBody[1]
	numKeySyms := int(req.KeyCodeCount) * int(req.KeySymsPerKeyCode)
	if len(requestBody) < 4+numKeySyms*4 {
		return nil, fmt.Errorf("%w: change keyboard mapping request too short for key syms", errParseError)
	}
	for i := 0; i < numKeySyms; i++ {
		offset := 4 + i*4
		req.KeySyms = append(req.KeySyms, order.Uint32(requestBody[offset:offset+4]))
	}
	return req, nil
}

type ChangeKeyboardControlRequest struct {
	ValueMask uint32
	Values    KeyboardControl
}

func (ChangeKeyboardControlRequest) OpCode() reqCode { return ChangeKeyboardControl }

func parseChangeKeyboardControlRequest(order binary.ByteOrder, requestBody []byte) (*ChangeKeyboardControlRequest, error) {
	if len(requestBody) < 4 {
		return nil, fmt.Errorf("%w: change keyboard control request too short", errParseError)
	}
	req := &ChangeKeyboardControlRequest{}
	req.ValueMask = order.Uint32(requestBody[0:4])
	values, _, err := parseKeyboardControl(order, req.ValueMask, requestBody[4:])
	if err != nil {
		return nil, err
	}
	req.Values = values
	return req, nil
}

type GetKeyboardControlRequest struct{}

func (GetKeyboardControlRequest) OpCode() reqCode { return GetKeyboardControl }

func parseGetKeyboardControlRequest(order binary.ByteOrder, requestBody []byte) (*GetKeyboardControlRequest, error) {
	return &GetKeyboardControlRequest{}, nil
}

type SetScreenSaverRequest struct {
	Timeout     int16
	Interval    int16
	PreferBlank byte
	AllowExpose byte
}

func (SetScreenSaverRequest) OpCode() reqCode { return SetScreenSaver }

func parseSetScreenSaverRequest(order binary.ByteOrder, requestBody []byte) (*SetScreenSaverRequest, error) {
	if len(requestBody) < 6 {
		return nil, fmt.Errorf("%w: set screen saver request too short", errParseError)
	}
	req := &SetScreenSaverRequest{}
	req.Timeout = int16(order.Uint16(requestBody[0:2]))
	req.Interval = int16(order.Uint16(requestBody[2:4]))
	req.PreferBlank = requestBody[4]
	req.AllowExpose = requestBody[5]
	return req, nil
}

type GetScreenSaverRequest struct{}

func (GetScreenSaverRequest) OpCode() reqCode { return GetScreenSaver }

func parseGetScreenSaverRequest(order binary.ByteOrder, requestBody []byte) (*GetScreenSaverRequest, error) {
	return &GetScreenSaverRequest{}, nil
}

type ChangeHostsRequest struct {
	Mode byte
	Host Host
}

func (ChangeHostsRequest) OpCode() reqCode { return ChangeHosts }

func parseChangeHostsRequest(order binary.ByteOrder, data byte, requestBody []byte) (*ChangeHostsRequest, error) {
	if len(requestBody) < 4 {
		return nil, fmt.Errorf("%w: change hosts request too short", errParseError)
	}
	req := &ChangeHostsRequest{}
	req.Mode = data
	family := requestBody[0]
	addressLen := order.Uint16(requestBody[2:4])
	if len(requestBody) < 4+int(addressLen) {
		return nil, fmt.Errorf("%w: change hosts request too short for host data", errParseError)
	}
	req.Host = Host{
		Family: family,
		Data:   requestBody[4 : 4+addressLen],
	}
	return req, nil
}

type ListHostsRequest struct{}

func (ListHostsRequest) OpCode() reqCode { return ListHosts }

func parseListHostsRequest(order binary.ByteOrder, requestBody []byte) (*ListHostsRequest, error) {
	return &ListHostsRequest{}, nil
}

type SetAccessControlRequest struct {
	Mode byte
}

func (SetAccessControlRequest) OpCode() reqCode { return SetAccessControl }

func parseSetAccessControlRequest(order binary.ByteOrder, data byte, requestBody []byte) (*SetAccessControlRequest, error) {
	req := &SetAccessControlRequest{}
	req.Mode = data
	return req, nil
}

type SetCloseDownModeRequest struct {
	Mode byte
}

func (SetCloseDownModeRequest) OpCode() reqCode { return SetCloseDownMode }

func parseSetCloseDownModeRequest(order binary.ByteOrder, data byte, requestBody []byte) (*SetCloseDownModeRequest, error) {
	req := &SetCloseDownModeRequest{}
	req.Mode = data
	return req, nil
}

type KillClientRequest struct {
	Resource uint32
}

func (KillClientRequest) OpCode() reqCode { return KillClient }

func parseKillClientRequest(order binary.ByteOrder, requestBody []byte) (*KillClientRequest, error) {
	if len(requestBody) < 4 {
		return nil, fmt.Errorf("%w: kill client request too short", errParseError)
	}
	req := &KillClientRequest{}
	req.Resource = order.Uint32(requestBody[0:4])
	return req, nil
}

type RotatePropertiesRequest struct {
	Window Window
	Delta  int16
	Atoms  []Atom
}

func (RotatePropertiesRequest) OpCode() reqCode { return RotateProperties }

func parseRotatePropertiesRequest(order binary.ByteOrder, requestBody []byte) (*RotatePropertiesRequest, error) {
	if len(requestBody) < 8 {
		return nil, fmt.Errorf("%w: rotate properties request too short", errParseError)
	}
	req := &RotatePropertiesRequest{}
	req.Window = Window(order.Uint32(requestBody[0:4]))
	numAtoms := order.Uint16(requestBody[4:6])
	req.Delta = int16(order.Uint16(requestBody[6:8]))
	if len(requestBody) < 8+int(numAtoms)*4 {
		return nil, fmt.Errorf("%w: rotate properties request too short for atoms", errParseError)
	}
	for i := 0; i < int(numAtoms); i++ {
		offset := 8 + i*4
		req.Atoms = append(req.Atoms, Atom(order.Uint32(requestBody[offset:offset+4])))
	}
	return req, nil
}

type ForceScreenSaverRequest struct {
	Mode byte
}

func (ForceScreenSaverRequest) OpCode() reqCode { return ForceScreenSaver }

func parseForceScreenSaverRequest(order binary.ByteOrder, data byte, requestBody []byte) (*ForceScreenSaverRequest, error) {
	req := &ForceScreenSaverRequest{}
	req.Mode = data
	return req, nil
}

type SetModifierMappingRequest struct {
	KeyCodesPerModifier byte
	KeyCodes            []KeyCode
}

func (SetModifierMappingRequest) OpCode() reqCode { return SetModifierMapping }

func parseSetModifierMappingRequest(order binary.ByteOrder, requestBody []byte) (*SetModifierMappingRequest, error) {
	if len(requestBody) < 1 {
		return nil, fmt.Errorf("%w: set modifier mapping request too short", errParseError)
	}
	req := &SetModifierMappingRequest{}
	req.KeyCodesPerModifier = requestBody[0]
	for i := 1; i < len(requestBody); i++ {
		req.KeyCodes = append(req.KeyCodes, KeyCode(requestBody[i]))
	}
	return req, nil
}

type GetModifierMappingRequest struct{}

func (GetModifierMappingRequest) OpCode() reqCode { return GetModifierMapping }

func parseGetModifierMappingRequest(order binary.ByteOrder, requestBody []byte) (*GetModifierMappingRequest, error) {
	return &GetModifierMappingRequest{}, nil
}

func parseKeyboardControl(order binary.ByteOrder, valueMask uint32, valuesData []byte) (KeyboardControl, int, error) {
	kc := KeyboardControl{}
	offset := 0
	if valueMask&KBKeyClickPercent != 0 {
		if len(valuesData) < offset+4 {
			return kc, 0, fmt.Errorf("%w: keyboard control values too short for key click percent", errParseError)
		}
		kc.KeyClickPercent = int32(order.Uint32(valuesData[offset : offset+4]))
		offset += 4
	}
	if valueMask&KBBellPercent != 0 {
		if len(valuesData) < offset+4 {
			return kc, 0, fmt.Errorf("%w: keyboard control values too short for bell percent", errParseError)
		}
		kc.BellPercent = int32(order.Uint32(valuesData[offset : offset+4]))
		offset += 4
	}
	if valueMask&KBBellPitch != 0 {
		if len(valuesData) < offset+4 {
			return kc, 0, fmt.Errorf("%w: keyboard control values too short for bell pitch", errParseError)
		}
		kc.BellPitch = int32(order.Uint32(valuesData[offset : offset+4]))
		offset += 4
	}
	if valueMask&KBBellDuration != 0 {
		if len(valuesData) < offset+4 {
			return kc, 0, fmt.Errorf("%w: keyboard control values too short for bell duration", errParseError)
		}
		kc.BellDuration = int32(order.Uint32(valuesData[offset : offset+4]))
		offset += 4
	}
	if valueMask&KBLed != 0 {
		if len(valuesData) < offset+4 {
			return kc, 0, fmt.Errorf("%w: keyboard control values too short for led", errParseError)
		}
		kc.Led = order.Uint32(valuesData[offset : offset+4])
		offset += 4
	}
	if valueMask&KBLedMode != 0 {
		if len(valuesData) < offset+4 {
			return kc, 0, fmt.Errorf("%w: keyboard control values too short for led mode", errParseError)
		}
		kc.LedMode = order.Uint32(valuesData[offset : offset+4])
		offset += 4
	}
	if valueMask&KBKey != 0 {
		if len(valuesData) < offset+4 {
			return kc, 0, fmt.Errorf("%w: keyboard control values too short for key", errParseError)
		}
		kc.Key = KeyCode(valuesData[offset])
		offset += 4
	}
	if valueMask&KBAutoRepeatMode != 0 {
		if len(valuesData) < offset+4 {
			return kc, 0, fmt.Errorf("%w: keyboard control values too short for auto repeat mode", errParseError)
		}
		kc.AutoRepeatMode = order.Uint32(valuesData[offset : offset+4])
		offset += 4
	}
	return kc, offset, nil
}

type NoOperationRequest struct{}

func (NoOperationRequest) OpCode() reqCode { return NoOperation }

func parseNoOperationRequest(order binary.ByteOrder, requestBody []byte) (*NoOperationRequest, error) {
	return &NoOperationRequest{}, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
func parseGCValues(order binary.ByteOrder, valueMask uint32, valuesData []byte) (GC, int, error) {
	// http://www.x.org/releases/X11R7.6/doc/xproto/x11protocol.html#requests:CreateGC
	gc := GC{
		Function:          FunctionCopy,
		PlaneMask:         ^uint32(0),
		Foreground:        0,
		Background:        1,
		LineWidth:         0,
		LineStyle:         LineStyleSolid,
		CapStyle:          CapStyleButt,
		JoinStyle:         JoinStyleMiter,
		FillStyle:         FillStyleSolid,
		FillRule:          FillRuleEvenOdd,
		Tile:              0, // pixmap of unspecified size filled with foreground pixel
		Stipple:           0, // pixmap of unspecified size filled with ones
		TileStipXOrigin:   0,
		TileStipYOrigin:   0,
		Font:              0, // server-dependent
		SubwindowMode:     SubwindowModeClipByChildren,
		GraphicsExposures: 1, // true
		ClipXOrigin:       0,
		ClipYOrigin:       0,
		ClipMask:          0, // no clip mask
		DashOffset:        0,
		Dashes:            4,
		ArcMode:           ArcModePieSlice,
	}
	offset := 0
	if valueMask&GCFunction != 0 {
		if len(valuesData) < offset+4 {
			return gc, 0, fmt.Errorf("%w: gc values too short for function", errParseError)
		}
		gc.Function = order.Uint32(valuesData[offset : offset+4])
		offset += 4
	}
	if valueMask&GCPlaneMask != 0 {
		if len(valuesData) < offset+4 {
			return gc, 0, fmt.Errorf("%w: gc values too short for plane mask", errParseError)
		}
		gc.PlaneMask = order.Uint32(valuesData[offset : offset+4])
		offset += 4
	}
	if valueMask&GCForeground != 0 {
		if len(valuesData) < offset+4 {
			return gc, 0, fmt.Errorf("%w: gc values too short for foreground", errParseError)
		}
		gc.Foreground = order.Uint32(valuesData[offset : offset+4])
		offset += 4
	}
	if valueMask&GCBackground != 0 {
		if len(valuesData) < offset+4 {
			return gc, 0, fmt.Errorf("%w: gc values too short for background", errParseError)
		}
		gc.Background = order.Uint32(valuesData[offset : offset+4])
		offset += 4
	}
	if valueMask&GCLineWidth != 0 {
		if len(valuesData) < offset+4 {
			return gc, 0, fmt.Errorf("%w: gc values too short for line width", errParseError)
		}
		gc.LineWidth = order.Uint32(valuesData[offset : offset+4])
		offset += 4
	}
	if valueMask&GCLineStyle != 0 {
		if len(valuesData) < offset+4 {
			return gc, 0, fmt.Errorf("%w: gc values too short for line style", errParseError)
		}
		gc.LineStyle = order.Uint32(valuesData[offset : offset+4])
		offset += 4
	}
	if valueMask&GCCapStyle != 0 {
		if len(valuesData) < offset+4 {
			return gc, 0, fmt.Errorf("%w: gc values too short for cap style", errParseError)
		}
		gc.CapStyle = order.Uint32(valuesData[offset : offset+4])
		offset += 4
	}
	if valueMask&GCJoinStyle != 0 {
		if len(valuesData) < offset+4 {
			return gc, 0, fmt.Errorf("%w: gc values too short for join style", errParseError)
		}
		gc.JoinStyle = order.Uint32(valuesData[offset : offset+4])
		offset += 4
	}
	if valueMask&GCFillStyle != 0 {
		if len(valuesData) < offset+4 {
			return gc, 0, fmt.Errorf("%w: gc values too short for fill style", errParseError)
		}
		gc.FillStyle = order.Uint32(valuesData[offset : offset+4])
		offset += 4
	}
	if valueMask&GCFillRule != 0 {
		if len(valuesData) < offset+4 {
			return gc, 0, fmt.Errorf("%w: gc values too short for fill rule", errParseError)
		}
		gc.FillRule = order.Uint32(valuesData[offset : offset+4])
		offset += 4
	}
	if valueMask&GCTile != 0 {
		if len(valuesData) < offset+4 {
			return gc, 0, fmt.Errorf("%w: gc values too short for tile", errParseError)
		}
		gc.Tile = order.Uint32(valuesData[offset : offset+4])
		offset += 4
	}
	if valueMask&GCStipple != 0 {
		if len(valuesData) < offset+4 {
			return gc, 0, fmt.Errorf("%w: gc values too short for stipple", errParseError)
		}
		gc.Stipple = order.Uint32(valuesData[offset : offset+4])
		offset += 4
	}
	if valueMask&GCTileStipXOrigin != 0 {
		if len(valuesData) < offset+4 {
			return gc, 0, fmt.Errorf("%w: gc values too short for tile stip x origin", errParseError)
		}
		gc.TileStipXOrigin = order.Uint32(valuesData[offset : offset+4])
		offset += 4
	}
	if valueMask&GCTileStipYOrigin != 0 {
		if len(valuesData) < offset+4 {
			return gc, 0, fmt.Errorf("%w: gc values too short for tile stip y origin", errParseError)
		}
		gc.TileStipYOrigin = order.Uint32(valuesData[offset : offset+4])
		offset += 4
	}
	if valueMask&GCFont != 0 {
		if len(valuesData) < offset+4 {
			return gc, 0, fmt.Errorf("%w: gc values too short for font", errParseError)
		}
		gc.Font = order.Uint32(valuesData[offset : offset+4])
		offset += 4
	}
	if valueMask&GCSubwindowMode != 0 {
		if len(valuesData) < offset+4 {
			return gc, 0, fmt.Errorf("%w: gc values too short for subwindow mode", errParseError)
		}
		gc.SubwindowMode = order.Uint32(valuesData[offset : offset+4])
		offset += 4
	}
	if valueMask&GCGraphicsExposures != 0 {
		if len(valuesData) < offset+4 {
			return gc, 0, fmt.Errorf("%w: gc values too short for graphics exposures", errParseError)
		}
		gc.GraphicsExposures = order.Uint32(valuesData[offset : offset+4])
		offset += 4
	}
	if valueMask&GCClipXOrigin != 0 {
		if len(valuesData) < offset+4 {
			return gc, 0, fmt.Errorf("%w: gc values too short for clip x origin", errParseError)
		}
		gc.ClipXOrigin = int32(order.Uint32(valuesData[offset : offset+4]))
		offset += 4
	}
	if valueMask&GCClipYOrigin != 0 {
		if len(valuesData) < offset+4 {
			return gc, 0, fmt.Errorf("%w: gc values too short for clip y origin", errParseError)
		}
		gc.ClipYOrigin = int32(order.Uint32(valuesData[offset : offset+4]))
		offset += 4
	}
	if valueMask&GCClipMask != 0 {
		if len(valuesData) < offset+4 {
			return gc, 0, fmt.Errorf("%w: gc values too short for clip mask", errParseError)
		}
		gc.ClipMask = order.Uint32(valuesData[offset : offset+4])
		offset += 4
	}
	if valueMask&GCDashOffset != 0 {
		if len(valuesData) < offset+4 {
			return gc, 0, fmt.Errorf("%w: gc values too short for dash offset", errParseError)
		}
		gc.DashOffset = order.Uint32(valuesData[offset : offset+4])
		offset += 4
	}
	if valueMask&GCDashes != 0 {
		if len(valuesData) < offset+4 {
			return gc, 0, fmt.Errorf("%w: gc values too short for dash list", errParseError)
		}
		gc.Dashes = order.Uint32(valuesData[offset : offset+4])
		offset += 4
	}
	if valueMask&GCArcMode != 0 {
		if len(valuesData) < offset+4 {
			return gc, 0, fmt.Errorf("%w: gc values too short for arc mode", errParseError)
		}
		gc.ArcMode = order.Uint32(valuesData[offset : offset+4])
		offset += 4
	}
	return gc, offset, nil
}

func parseWindowAttributes(order binary.ByteOrder, valueMask uint32, valuesData []byte) (WindowAttributes, int, error) {
	wa := WindowAttributes{}
	offset := 0
	if valueMask&CWBackPixmap != 0 {
		if len(valuesData) < offset+4 {
			return wa, 0, fmt.Errorf("%w: window attributes too short for background pixmap", errParseError)
		}
		wa.BackgroundPixmap = Pixmap(order.Uint32(valuesData[offset : offset+4]))
		offset += 4
	}
	if valueMask&CWBackPixel != 0 {
		if len(valuesData) < offset+4 {
			return wa, 0, fmt.Errorf("%w: window attributes too short for background pixel", errParseError)
		}
		wa.BackgroundPixel = order.Uint32(valuesData[offset : offset+4])
		offset += 4
	}
	if valueMask&CWBorderPixmap != 0 {
		if len(valuesData) < offset+4 {
			return wa, 0, fmt.Errorf("%w: window attributes too short for border pixmap", errParseError)
		}
		wa.BorderPixmap = Pixmap(order.Uint32(valuesData[offset : offset+4]))
		offset += 4
	}
	if valueMask&CWBorderPixel != 0 {
		if len(valuesData) < offset+4 {
			return wa, 0, fmt.Errorf("%w: window attributes too short for border pixel", errParseError)
		}
		wa.BorderPixel = order.Uint32(valuesData[offset : offset+4])
		offset += 4
	}
	if valueMask&CWBitGravity != 0 {
		if len(valuesData) < offset+4 {
			return wa, 0, fmt.Errorf("%w: window attributes too short for bit gravity", errParseError)
		}
		wa.BitGravity = order.Uint32(valuesData[offset : offset+4])
		offset += 4
	}
	if valueMask&CWWinGravity != 0 {
		if len(valuesData) < offset+4 {
			return wa, 0, fmt.Errorf("%w: window attributes too short for win gravity", errParseError)
		}
		wa.WinGravity = order.Uint32(valuesData[offset : offset+4])
		offset += 4
	}
	if valueMask&CWBackingStore != 0 {
		if len(valuesData) < offset+4 {
			return wa, 0, fmt.Errorf("%w: window attributes too short for backing store", errParseError)
		}
		wa.BackingStore = order.Uint32(valuesData[offset : offset+4])
		offset += 4
	}
	if valueMask&CWBackingPlanes != 0 {
		if len(valuesData) < offset+4 {
			return wa, 0, fmt.Errorf("%w: window attributes too short for backing planes", errParseError)
		}
		wa.BackingPlanes = order.Uint32(valuesData[offset : offset+4])
		offset += 4
	}
	if valueMask&CWBackingPixel != 0 {
		if len(valuesData) < offset+4 {
			return wa, 0, fmt.Errorf("%w: window attributes too short for backing pixel", errParseError)
		}
		wa.BackingPixel = order.Uint32(valuesData[offset : offset+4])
		offset += 4
	}
	if valueMask&CWOverrideRedirect != 0 {
		if len(valuesData) < offset+4 {
			return wa, 0, fmt.Errorf("%w: window attributes too short for override redirect", errParseError)
		}
		wa.OverrideRedirect = order.Uint32(valuesData[offset:offset+4]) != 0
		offset += 4
	}
	if valueMask&CWSaveUnder != 0 {
		if len(valuesData) < offset+4 {
			return wa, 0, fmt.Errorf("%w: window attributes too short for save under", errParseError)
		}
		wa.SaveUnder = order.Uint32(valuesData[offset:offset+4]) != 0
		offset += 4
	}
	if valueMask&CWEventMask != 0 {
		if len(valuesData) < offset+4 {
			return wa, 0, fmt.Errorf("%w: window attributes too short for event mask", errParseError)
		}
		wa.EventMask = order.Uint32(valuesData[offset : offset+4])
		offset += 4
	}
	if valueMask&CWDontPropagate != 0 {
		if len(valuesData) < offset+4 {
			return wa, 0, fmt.Errorf("%w: window attributes too short for dont propagate mask", errParseError)
		}
		wa.DontPropagateMask = order.Uint32(valuesData[offset : offset+4])
		offset += 4
	}
	if valueMask&CWColormap != 0 {
		if len(valuesData) < offset+4 {
			return wa, 0, fmt.Errorf("%w: window attributes too short for colormap", errParseError)
		}
		wa.Colormap = Colormap(order.Uint32(valuesData[offset : offset+4]))
		offset += 4
	}
	if valueMask&CWCursor != 0 {
		if len(valuesData) < offset+4 {
			return wa, 0, fmt.Errorf("%w: window attributes too short for cursor", errParseError)
		}
		wa.Cursor = Cursor(order.Uint32(valuesData[offset : offset+4]))
		offset += 4
	}
	return wa, offset, nil
}

// AllocColorCells: 86
type AllocColorCellsRequest struct {
	Contiguous bool
	Cmap       Colormap
	Colors     uint16
	Planes     uint16
}

func (r *AllocColorCellsRequest) OpCode() reqCode { return AllocColorCells }

func parseAllocColorCellsRequest(order binary.ByteOrder, data byte, body []byte) (*AllocColorCellsRequest, error) {
	if len(body) < 8 {
		return nil, fmt.Errorf("%w: alloc color cells request too short", errParseError)
	}
	req := &AllocColorCellsRequest{}
	req.Contiguous = data != 0
	req.Cmap = Colormap(order.Uint32(body[0:4]))
	req.Colors = order.Uint16(body[4:6])
	req.Planes = order.Uint16(body[6:8])
	return req, nil
}

// AllocColorPlanes: 87
type AllocColorPlanesRequest struct {
	Contiguous bool
	Cmap       Colormap
	Colors     uint16
	Reds       uint16
	Greens     uint16
	Blues      uint16
}

func (r *AllocColorPlanesRequest) OpCode() reqCode { return AllocColorPlanes }

func parseAllocColorPlanesRequest(order binary.ByteOrder, data byte, body []byte) (*AllocColorPlanesRequest, error) {
	if len(body) < 12 {
		return nil, fmt.Errorf("%w: alloc color planes request too short", errParseError)
	}
	req := &AllocColorPlanesRequest{}
	req.Contiguous = data != 0
	req.Cmap = Colormap(order.Uint32(body[0:4]))
	req.Colors = order.Uint16(body[4:6])
	req.Reds = order.Uint16(body[6:8])
	req.Greens = order.Uint16(body[8:10])
	req.Blues = order.Uint16(body[10:12])
	return req, nil
}

// reqCodeCreateCursor:
type CreateCursorRequest struct {
	Cid       Cursor
	Source    Pixmap
	Mask      Pixmap
	ForeRed   uint16
	ForeGreen uint16
	ForeBlue  uint16
	BackRed   uint16
	BackGreen uint16
	BackBlue  uint16
	X         uint16
	Y         uint16
}

func (r *CreateCursorRequest) OpCode() reqCode { return CreateCursor }

func parseCreateCursorRequest(order binary.ByteOrder, body []byte) (*CreateCursorRequest, error) {
	if len(body) < 28 {
		return nil, fmt.Errorf("%w: create cursor request too short", errParseError)
	}
	req := &CreateCursorRequest{}
	req.Cid = Cursor(order.Uint32(body[0:4]))
	req.Source = Pixmap(order.Uint32(body[4:8]))
	req.Mask = Pixmap(order.Uint32(body[8:12]))
	req.ForeRed = order.Uint16(body[12:14])
	req.ForeGreen = order.Uint16(body[14:16])
	req.ForeBlue = order.Uint16(body[16:18])
	req.BackRed = order.Uint16(body[18:20])
	req.BackGreen = order.Uint16(body[20:22])
	req.BackBlue = order.Uint16(body[22:24])
	req.X = order.Uint16(body[24:26])
	req.Y = order.Uint16(body[26:28])
	return req, nil
}

// reqCodeCopyPlane:
type CopyPlaneRequest struct {
	SrcDrawable Drawable
	DstDrawable Drawable
	Gc          GContext
	SrcX        int16
	SrcY        int16
	DstX        int16
	DstY        int16
	Width       uint16
	Height      uint16
	PlaneMask   uint32
}

func (r *CopyPlaneRequest) OpCode() reqCode { return CopyPlane }

func parseCopyPlaneRequest(order binary.ByteOrder, body []byte) (*CopyPlaneRequest, error) {
	if len(body) < 28 {
		return nil, fmt.Errorf("%w: copy plane request too short", errParseError)
	}
	req := &CopyPlaneRequest{}
	req.SrcDrawable = Drawable(order.Uint32(body[0:4]))
	req.DstDrawable = Drawable(order.Uint32(body[4:8]))
	req.Gc = GContext(order.Uint32(body[8:12]))
	req.SrcX = int16(order.Uint16(body[12:14]))
	req.SrcY = int16(order.Uint16(body[14:16]))
	req.DstX = int16(order.Uint16(body[16:18]))
	req.DstY = int16(order.Uint16(body[18:20]))
	req.Width = order.Uint16(body[20:22])
	req.Height = order.Uint16(body[22:24])
	req.PlaneMask = order.Uint32(body[24:28])
	return req, nil
}

// reqCodeListExtensions:
type ListExtensionsRequest struct{}

func (r *ListExtensionsRequest) OpCode() reqCode { return ListExtensions }

func parseListExtensionsRequest(order binary.ByteOrder, raw []byte) (*ListExtensionsRequest, error) {
	return &ListExtensionsRequest{}, nil
}

// reqCodeChangePointerControl:
type ChangePointerControlRequest struct {
	AccelerationNumerator   int16
	AccelerationDenominator int16
	Threshold               int16
	DoAcceleration          bool
	DoThreshold             bool
}

func (r *ChangePointerControlRequest) OpCode() reqCode { return ChangePointerControl }

func parseChangePointerControlRequest(order binary.ByteOrder, body []byte) (*ChangePointerControlRequest, error) {
	if len(body) < 8 {
		return nil, fmt.Errorf("%w: change pointer control request too short", errParseError)
	}
	req := &ChangePointerControlRequest{}
	req.AccelerationNumerator = int16(order.Uint16(body[0:2]))
	req.AccelerationDenominator = int16(order.Uint16(body[2:4]))
	req.Threshold = int16(order.Uint16(body[4:6]))
	req.DoAcceleration = body[6] != 0
	req.DoThreshold = body[7] != 0
	return req, nil
}

// reqCodeGetPointerControl:
type GetPointerControlRequest struct{}

func (r *GetPointerControlRequest) OpCode() reqCode { return GetPointerControl }

func parseGetPointerControlRequest(order binary.ByteOrder, raw []byte) (*GetPointerControlRequest, error) {
	return &GetPointerControlRequest{}, nil
}
