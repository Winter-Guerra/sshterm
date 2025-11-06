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
	debugf("X11: Raw request header: %x", reqHeader)
	length := order.Uint16(reqHeader[2:4])
	if int(4*length) != len(raw) {
		return nil, fmt.Errorf("%w: mismatch length %d != %d", errParseError, 4*length, len(raw))
	}

	opcode := reqCode(reqHeader[0])
	data := reqHeader[1]
	body := raw[4:]

	switch opcode {
	case CreateWindow:
		req, err := parseCreateWindowRequest(order, data, body)
		if err != nil {
			return nil, err
		}
		return req, nil

	case ChangeWindowAttributes:
		req, err := parseChangeWindowAttributesRequest(order, body)
		if err != nil {
			return nil, err
		}
		return req, nil

	case GetWindowAttributes:
		req, err := parseGetWindowAttributesRequest(order, body)
		if err != nil {
			return nil, err
		}
		return req, nil

	case DestroyWindow:
		req, err := parseDestroyWindowRequest(order, body)
		if err != nil {
			return nil, err
		}
		return req, nil

	case DestroySubwindows:
		req, err := parseDestroySubwindowsRequest(order, body)
		if err != nil {
			return nil, err
		}
		return req, nil

	case ChangeSaveSet:
		req, err := parseChangeSaveSetRequest(order, body)
		if err != nil {
			return nil, err
		}
		return req, nil

	case ReparentWindow:
		req, err := parseReparentWindowRequest(order, body)
		if err != nil {
			return nil, err
		}
		return req, nil

	case MapWindow:
		req, err := parseMapWindowRequest(order, body)
		if err != nil {
			return nil, err
		}
		return req, nil

	case MapSubwindows:
		req, err := parseMapSubwindowsRequest(order, body)
		if err != nil {
			return nil, err
		}
		return req, nil

	case UnmapWindow:
		req, err := parseUnmapWindowRequest(order, body)
		if err != nil {
			return nil, err
		}
		return req, nil

	case UnmapSubwindows:
		req, err := parseUnmapSubwindowsRequest(order, body)
		if err != nil {
			return nil, err
		}
		return req, nil

	case ConfigureWindow:
		req, err := parseConfigureWindowRequest(order, body)
		if err != nil {
			return nil, err
		}
		return req, nil

	case CirculateWindow:
		req, err := parseCirculateWindowRequest(order, data, body)
		if err != nil {
			return nil, err
		}
		return req, nil

	case GetGeometry:
		req, err := parseGetGeometryRequest(order, body)
		if err != nil {
			return nil, err
		}
		return req, nil

	case QueryTree:
		req, err := parseQueryTreeRequest(order, body)
		if err != nil {
			return nil, err
		}
		return req, nil

	case InternAtom:
		req, err := parseInternAtomRequest(order, body)
		if err != nil {
			return nil, err
		}
		return req, nil

	case GetAtomName:
		req, err := parseGetAtomNameRequest(order, body)
		if err != nil {
			return nil, err
		}
		return req, nil

	case ChangeProperty:
		req, err := parseChangePropertyRequest(order, body)
		if err != nil {
			return nil, err
		}
		return req, nil

	case DeleteProperty:
		req, err := parseDeletePropertyRequest(order, body)
		if err != nil {
			return nil, err
		}
		return req, nil

	case GetProperty:
		req, err := parseGetPropertyRequest(order, body)
		if err != nil {
			return nil, err
		}
		return req, nil

	case ListProperties:
		req, err := parseListPropertiesRequest(order, body)
		if err != nil {
			return nil, err
		}
		return req, nil

	case SetSelectionOwner:
		req, err := parseSetSelectionOwnerRequest(order, body)
		if err != nil {
			return nil, err
		}
		return req, nil

	case GetSelectionOwner:
		req, err := parseGetSelectionOwnerRequest(order, body)
		if err != nil {
			return nil, err
		}
		return req, nil

	case ConvertSelection:
		req, err := parseConvertSelectionRequest(order, body)
		if err != nil {
			return nil, err
		}
		return req, nil

	case SendEvent:
		req, err := parseSendEventRequest(order, body)
		if err != nil {
			return nil, err
		}
		return req, nil

	case GrabPointer:
		req, err := parseGrabPointerRequest(order, body)
		if err != nil {
			return nil, err
		}
		return req, nil

	case UngrabPointer:
		req, err := parseUngrabPointerRequest(order, body)
		if err != nil {
			return nil, err
		}
		return req, nil

	case GrabButton:
		req, err := parseGrabButtonRequest(order, body)
		if err != nil {
			return nil, err
		}
		return req, nil

	case UngrabButton:
		req, err := parseUngrabButtonRequest(order, body)
		if err != nil {
			return nil, err
		}
		return req, nil

	case ChangeActivePointerGrab:
		req, err := parseChangeActivePointerGrabRequest(order, body)
		if err != nil {
			return nil, err
		}
		return req, nil

	case GrabKeyboard:
		req, err := parseGrabKeyboardRequest(order, body)
		if err != nil {
			return nil, err
		}
		return req, nil

	case UngrabKeyboard:
		req, err := parseUngrabKeyboardRequest(order, body)
		if err != nil {
			return nil, err
		}
		return req, nil

	case GrabKey:
		req, err := parseGrabKeyRequest(order, body)
		if err != nil {
			return nil, err
		}
		return req, nil

	case UngrabKey:
		req, err := parseUngrabKeyRequest(order, body)
		if err != nil {
			return nil, err
		}
		return req, nil

	case AllowEvents:
		req, err := parseAllowEventsRequest(order, data, body)
		if err != nil {
			return nil, err
		}
		return req, nil

	case GrabServer:
		req, err := parseGrabServerRequest(order, body)
		if err != nil {
			return nil, err
		}
		return req, nil

	case UngrabServer:
		req, err := parseUngrabServerRequest(order, body)
		if err != nil {
			return nil, err
		}
		return req, nil

	case QueryPointer:
		req, err := parseQueryPointerRequest(order, body)
		if err != nil {
			return nil, err
		}
		return req, nil

	case GetMotionEvents:
		req, err := parseGetMotionEventsRequest(order, body)
		if err != nil {
			return nil, err
		}
		return req, nil

	case TranslateCoords:
		req, err := parseTranslateCoordsRequest(order, body)
		if err != nil {
			return nil, err
		}
		return req, nil

	case WarpPointer:
		req, err := parseWarpPointerRequest(order, body)
		if err != nil {
			return nil, err
		}
		return req, nil

	case SetInputFocus:
		req, err := parseSetInputFocusRequest(order, body)
		if err != nil {
			return nil, err
		}
		return req, nil

	case GetInputFocus:
		req, err := parseGetInputFocusRequest(order, body)
		if err != nil {
			return nil, err
		}
		return req, nil

	case QueryKeymap:
		req, err := parseQueryKeymapRequest(order, body)
		if err != nil {
			return nil, err
		}
		return req, nil

	case OpenFont:
		req, err := parseOpenFontRequest(order, body)
		if err != nil {
			return nil, err
		}
		return req, nil

	case CloseFont:
		req, err := parseCloseFontRequest(order, body)
		if err != nil {
			return nil, err
		}
		return req, nil

	case QueryFont:
		req, err := parseQueryFontRequest(order, body)
		if err != nil {
			return nil, err
		}
		return req, nil

	case QueryTextExtents:
		req, err := parseQueryTextExtentsRequest(order, body)
		if err != nil {
			return nil, err
		}
		return req, nil

	case ListFonts:
		req, err := parseListFontsRequest(order, body)
		if err != nil {
			return nil, err
		}
		return req, nil

	case ListFontsWithInfo:
		req, err := parseListFontsWithInfoRequest(order, body)
		if err != nil {
			return nil, err
		}
		return req, nil

	case SetFontPath:
		req, err := parseSetFontPathRequest(order, body)
		if err != nil {
			return nil, err
		}
		return req, nil

	case GetFontPath:
		req, err := parseGetFontPathRequest(order, body)
		if err != nil {
			return nil, err
		}
		return req, nil

	case CreatePixmap:
		req, err := parseCreatePixmapRequest(order, data, body)
		if err != nil {
			return nil, err
		}
		return req, nil

	case FreePixmap:
		req, err := parseFreePixmapRequest(order, body)
		if err != nil {
			return nil, err
		}
		return req, nil

	case CreateGC:
		req, err := parseCreateGCRequest(order, body)
		if err != nil {
			return nil, err
		}
		return req, nil

	case ChangeGC:
		req, err := parseChangeGCRequest(order, body)
		if err != nil {
			return nil, err
		}
		return req, nil

	case CopyGC:
		req, err := parseCopyGCRequest(order, body)
		if err != nil {
			return nil, err
		}
		return req, nil

	case SetDashes:
		req, err := parseSetDashesRequest(order, body)
		if err != nil {
			return nil, err
		}
		return req, nil

	case SetClipRectangles:
		req, err := parseSetClipRectanglesRequest(order, data, body)
		if err != nil {
			return nil, err
		}
		return req, nil

	case FreeGC:
		req, err := parseFreeGCRequest(order, body)
		if err != nil {
			return nil, err
		}
		return req, nil

	case ClearArea:
		req, err := parseClearAreaRequest(order, body)
		if err != nil {
			return nil, err
		}
		return req, nil

	case CopyArea:
		req, err := parseCopyAreaRequest(order, body)
		if err != nil {
			return nil, err
		}
		return req, nil

	case PolyPoint:
		req, err := parsePolyPointRequest(order, body)
		if err != nil {
			return nil, err
		}
		return req, nil

	case PolyLine:
		req, err := parsePolyLineRequest(order, body)
		if err != nil {
			return nil, err
		}
		return req, nil

	case PolySegment:
		req, err := parsePolySegmentRequest(order, body)
		if err != nil {
			return nil, err
		}
		return req, nil

	case PolyRectangle:
		req, err := parsePolyRectangleRequest(order, body)
		if err != nil {
			return nil, err
		}
		return req, nil

	case PolyArc:
		req, err := parsePolyArcRequest(order, body)
		if err != nil {
			return nil, err
		}
		return req, nil

	case FillPoly:
		req, err := parseFillPolyRequest(order, body)
		if err != nil {
			return nil, err
		}
		return req, nil

	case PolyFillRectangle:
		req, err := parsePolyFillRectangleRequest(order, body)
		if err != nil {
			return nil, err
		}
		return req, nil

	case PolyFillArc:
		req, err := parsePolyFillArcRequest(order, body)
		if err != nil {
			return nil, err
		}
		return req, nil

	case PutImage:
		req, err := parsePutImageRequest(order, data, body)
		if err != nil {
			return nil, err
		}
		return req, nil

	case GetImage:
		req, err := parseGetImageRequest(order, data, body)
		if err != nil {
			return nil, err
		}
		return req, nil

	case PolyText8:
		req, err := parsePolyText8Request(order, body)
		if err != nil {
			return nil, err
		}
		return req, nil

	case PolyText16:
		req, err := parsePolyText16Request(order, body)
		if err != nil {
			return nil, err
		}
		return req, nil

	case ImageText8:
		req, err := parseImageText8Request(order, body)
		if err != nil {
			return nil, err
		}
		return req, nil

	case ImageText16:
		req, err := parseImageText16Request(order, body)
		if err != nil {
			return nil, err
		}
		return req, nil

	case CreateColormap:
		req, err := parseCreateColormapRequest(order, body)
		if err != nil {
			return nil, err
		}
		return req, nil

	case FreeColormap:
		req, err := parseFreeColormapRequest(order, body)
		if err != nil {
			return nil, err
		}
		return req, nil

	case InstallColormap:
		req, err := parseInstallColormapRequest(order, body)
		if err != nil {
			return nil, err
		}
		return req, nil

	case UninstallColormap:
		req, err := parseUninstallColormapRequest(order, body)
		if err != nil {
			return nil, err
		}
		return req, nil

	case ListInstalledColormaps:
		req, err := parseListInstalledColormapsRequest(order, body)
		if err != nil {
			return nil, err
		}
		return req, nil

	case AllocColor:
		req, err := parseAllocColorRequest(order, body)
		if err != nil {
			return nil, err
		}
		return req, nil

	case AllocNamedColor:
		req, err := parseAllocNamedColorRequest(order, body)
		if err != nil {
			return nil, err
		}
		return req, nil

	case FreeColors:
		req, err := parseFreeColorsRequest(order, body)
		if err != nil {
			return nil, err
		}
		return req, nil

	case StoreColors:
		req, err := parseStoreColorsRequest(order, body)
		if err != nil {
			return nil, err
		}
		return req, nil

	case StoreNamedColor:
		req, err := parseStoreNamedColorRequest(order, data, body)
		if err != nil {
			return nil, err
		}
		return req, nil

	case QueryColors:
		req, err := parseQueryColorsRequest(order, body)
		if err != nil {
			return nil, err
		}
		return req, nil

	case LookupColor:
		req, err := parseLookupColorRequest(order, body)
		if err != nil {
			return nil, err
		}
		return req, nil

	case CreateGlyphCursor:
		req, err := parseCreateGlyphCursorRequest(order, body)
		if err != nil {
			return nil, err
		}
		return req, nil

	case FreeCursor:
		req, err := parseFreeCursorRequest(order, body)
		if err != nil {
			return nil, err
		}
		return req, nil

	case RecolorCursor:
		req, err := parseRecolorCursorRequest(order, body)
		if err != nil {
			return nil, err
		}
		return req, nil

	case QueryBestSize:
		req, err := parseQueryBestSizeRequest(order, body)
		if err != nil {
			return nil, err
		}
		return req, nil

	case QueryExtension:
		req, err := parseQueryExtensionRequest(order, body)
		if err != nil {
			return nil, err
		}
		return req, nil

	case Bell:
		req, err := parseBellRequest(data)
		if err != nil {
			return nil, err
		}
		return req, nil

	case SetPointerMapping:
		req, err := parseSetPointerMappingRequest(order, body)
		if err != nil {
			return nil, err
		}
		return req, nil

	case GetPointerMapping:
		req, err := parseGetPointerMappingRequest(order, body)
		if err != nil {
			return nil, err
		}
		return req, nil

	case GetKeyboardMapping:
		req, err := parseGetKeyboardMappingRequest(order, body)
		if err != nil {
			return nil, err
		}
		return req, nil

	case ChangeKeyboardMapping:
		req, err := parseChangeKeyboardMappingRequest(order, data, body)
		if err != nil {
			return nil, err
		}
		return req, nil

	case ChangeKeyboardControl:
		req, err := parseChangeKeyboardControlRequest(order, body)
		if err != nil {
			return nil, err
		}
		return req, nil

	case GetKeyboardControl:
		req, err := parseGetKeyboardControlRequest(order, body)
		if err != nil {
			return nil, err
		}
		return req, nil

	case SetScreenSaver:
		req, err := parseSetScreenSaverRequest(order, body)
		if err != nil {
			return nil, err
		}
		return req, nil

	case GetScreenSaver:
		req, err := parseGetScreenSaverRequest(order, body)
		if err != nil {
			return nil, err
		}
		return req, nil

	case ChangeHosts:
		req, err := parseChangeHostsRequest(order, data, body)
		if err != nil {
			return nil, err
		}
		return req, nil

	case ListHosts:
		req, err := parseListHostsRequest(order, body)
		if err != nil {
			return nil, err
		}
		return req, nil

	case SetAccessControl:
		req, err := parseSetAccessControlRequest(order, data, body)
		if err != nil {
			return nil, err
		}
		return req, nil

	case SetCloseDownMode:
		req, err := parseSetCloseDownModeRequest(order, data, body)
		if err != nil {
			return nil, err
		}
		return req, nil

	case KillClient:
		req, err := parseKillClientRequest(order, body)
		if err != nil {
			return nil, err
		}
		return req, nil

	case RotateProperties:
		req, err := parseRotatePropertiesRequest(order, body)
		if err != nil {
			return nil, err
		}
		return req, nil

	case ForceScreenSaver:
		req, err := parseForceScreenSaverRequest(order, data, body)
		if err != nil {
			return nil, err
		}
		return req, nil

	case SetModifierMapping:
		req, err := parseSetModifierMappingRequest(order, body)
		if err != nil {
			return nil, err
		}
		return req, nil

	case GetModifierMapping:
		req, err := parseGetModifierMappingRequest(order, body)
		if err != nil {
			return nil, err
		}
		return req, nil

	case NoOperation:
		req, err := parseNoOperationRequest(order, body)
		if err != nil {
			return nil, err
		}
		return req, nil

	default:
		return nil, fmt.Errorf("x11: unhandled opcode %d", opcode)
	}
}

// auxiliary data structures

type WindowAttributes struct {
	BackgroundPixmap   Pixmap
	BackgroundPixel    uint32
	BackgroundPixelSet bool
	BorderPixmap       Pixmap
	BorderPixel        uint32
	BitGravity         uint32
	WinGravity         uint32
	BackingStore       uint32
	BackingPlanes      uint32
	BackingPixel       uint32
	OverrideRedirect   uint32
	SaveUnder          uint32
	EventMask          uint32
	DontPropagateMask  uint32
	Colormap           Colormap
	Cursor             Cursor
}

type PolyText8Item struct {
	Delta int8
	Str   []byte
}

type PolyText16Item struct {
	Delta int8
	Str   []uint16
}

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
	req.Values, _ = parseWindowAttributes(order, req.ValueMask, requestBody[28:])
	return req, nil
}

type ChangeWindowAttributesRequest struct {
	Window    Window
	ValueMask uint32
	Values    WindowAttributes
}

func (ChangeWindowAttributesRequest) OpCode() reqCode { return ChangeWindowAttributes }

func parseChangeWindowAttributesRequest(order binary.ByteOrder, requestBody []byte) (*ChangeWindowAttributesRequest, error) {
	req := &ChangeWindowAttributesRequest{}
	req.Window = Window(order.Uint32(requestBody[0:4]))
	req.ValueMask = order.Uint32(requestBody[4:8])
	req.Values, _ = parseWindowAttributes(order, req.ValueMask, requestBody[8:])
	return req, nil
}

type GetWindowAttributesRequest struct {
	Drawable Drawable
}

func (GetWindowAttributesRequest) OpCode() reqCode { return GetWindowAttributes }

func parseGetWindowAttributesRequest(order binary.ByteOrder, requestBody []byte) (*GetWindowAttributesRequest, error) {
	req := &GetWindowAttributesRequest{}
	req.Drawable = Drawable(order.Uint32(requestBody[0:4]))
	return req, nil
}

type DestroyWindowRequest struct {
	Window Window
}

func (DestroyWindowRequest) OpCode() reqCode { return DestroyWindow }

func parseDestroyWindowRequest(order binary.ByteOrder, requestBody []byte) (*DestroyWindowRequest, error) {
	req := &DestroyWindowRequest{}
	req.Window = Window(order.Uint32(requestBody[0:4]))
	return req, nil
}

type DestroySubwindowsRequest struct {
	Window Window
}

func (DestroySubwindowsRequest) OpCode() reqCode { return DestroySubwindows }

func parseDestroySubwindowsRequest(order binary.ByteOrder, requestBody []byte) (*DestroySubwindowsRequest, error) {
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
	req := &MapWindowRequest{}
	req.Window = Window(order.Uint32(requestBody[0:4]))
	return req, nil
}

type MapSubwindowsRequest struct {
	Window Window
}

func (MapSubwindowsRequest) OpCode() reqCode { return MapSubwindows }

func parseMapSubwindowsRequest(order binary.ByteOrder, requestBody []byte) (*MapSubwindowsRequest, error) {
	req := &MapSubwindowsRequest{}
	req.Window = Window(order.Uint32(requestBody[0:4]))
	return req, nil
}

type UnmapWindowRequest struct {
	Window Window
}

func (UnmapWindowRequest) OpCode() reqCode { return UnmapWindow }

func parseUnmapWindowRequest(order binary.ByteOrder, requestBody []byte) (*UnmapWindowRequest, error) {
	req := &UnmapWindowRequest{}
	req.Window = Window(order.Uint32(requestBody[0:4]))
	return req, nil
}

type UnmapSubwindowsRequest struct {
	Window Window
}

func (UnmapSubwindowsRequest) OpCode() reqCode { return UnmapSubwindows }

func parseUnmapSubwindowsRequest(order binary.ByteOrder, requestBody []byte) (*UnmapSubwindowsRequest, error) {
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
	req := &GetGeometryRequest{}
	req.Drawable = Drawable(order.Uint32(requestBody[0:4]))
	return req, nil
}

type QueryTreeRequest struct {
	Window Window
}

func (QueryTreeRequest) OpCode() reqCode { return QueryTree }

func parseQueryTreeRequest(order binary.ByteOrder, requestBody []byte) (*QueryTreeRequest, error) {
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
	req := &InternAtomRequest{}
	nameLen := order.Uint16(requestBody[0:2])
	req.Name = string(requestBody[4 : 4+nameLen])
	return req, nil
}

type GetAtomNameRequest struct {
	Atom Atom
}

func (GetAtomNameRequest) OpCode() reqCode { return GetAtomName }

func parseGetAtomNameRequest(order binary.ByteOrder, requestBody []byte) (*GetAtomNameRequest, error) {
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
	req := &ChangePropertyRequest{}
	req.Window = Window(order.Uint32(requestBody[0:4]))
	req.Property = Atom(order.Uint32(requestBody[4:8]))
	req.Type = Atom(order.Uint32(requestBody[8:12]))
	req.Format = requestBody[12]
	dataLen := order.Uint32(requestBody[16:20])
	req.Data = requestBody[20 : 20+dataLen]
	return req, nil
}

type DeletePropertyRequest struct {
	Window   Window
	Property Atom
}

func (DeletePropertyRequest) OpCode() reqCode { return DeleteProperty }

func parseDeletePropertyRequest(order binary.ByteOrder, requestBody []byte) (*DeletePropertyRequest, error) {
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
	req := &OpenFontRequest{}
	req.Fid = Font(order.Uint32(requestBody[0:4]))
	nameLen := order.Uint16(requestBody[4:6])
	req.Name = string(requestBody[8 : 8+nameLen])
	return req, nil
}

type CloseFontRequest struct {
	Fid Font
}

func (CloseFontRequest) OpCode() reqCode { return CloseFont }

func parseCloseFontRequest(order binary.ByteOrder, requestBody []byte) (*CloseFontRequest, error) {
	req := &CloseFontRequest{}
	req.Fid = Font(order.Uint32(requestBody[0:4]))
	return req, nil
}

type QueryFontRequest struct {
	Fid Font
}

func (QueryFontRequest) OpCode() reqCode { return QueryFont }

func parseQueryFontRequest(order binary.ByteOrder, requestBody []byte) (*QueryFontRequest, error) {
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
	req := &ListFontsRequest{}
	req.MaxNames = order.Uint16(requestBody[0:2])
	nameLen := order.Uint16(requestBody[2:4])
	req.Pattern = string(requestBody[4 : 4+nameLen])
	return req, nil
}

type ListFontsWithInfoRequest struct {
	MaxNames uint16
	Pattern  string
}

func (ListFontsWithInfoRequest) OpCode() reqCode { return ListFontsWithInfo }

func parseListFontsWithInfoRequest(order binary.ByteOrder, requestBody []byte) (*ListFontsWithInfoRequest, error) {
	req := &ListFontsWithInfoRequest{}
	req.MaxNames = order.Uint16(requestBody[0:2])
	nameLen := order.Uint16(requestBody[2:4])
	req.Pattern = string(requestBody[4 : 4+nameLen])
	return req, nil
}

type SetFontPathRequest struct {
	NumPaths uint16
	Paths    []string
}

func (SetFontPathRequest) OpCode() reqCode { return SetFontPath }

func parseSetFontPathRequest(order binary.ByteOrder, requestBody []byte) (*SetFontPathRequest, error) {
	req := &SetFontPathRequest{}
	req.NumPaths = order.Uint16(requestBody[0:2])
	pathsData := requestBody[4:]
	for i := 0; i < int(req.NumPaths); i++ {
		len := int(pathsData[0])
		pathsData = pathsData[1:]
		req.Paths = append(req.Paths, string(pathsData[:len]))
		pathsData = pathsData[len:]
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
	req := &CreateGCRequest{}
	req.Cid = GContext(order.Uint32(requestBody[0:4]))
	req.Drawable = Drawable(order.Uint32(requestBody[4:8]))
	req.ValueMask = order.Uint32(requestBody[8:12])
	req.Values, _ = parseGCValues(order, req.ValueMask, requestBody[12:])
	return req, nil
}

type ChangeGCRequest struct {
	Gc        GContext
	ValueMask uint32
	Values    GC
}

func (ChangeGCRequest) OpCode() reqCode { return ChangeGC }

func parseChangeGCRequest(order binary.ByteOrder, requestBody []byte) (*ChangeGCRequest, error) {
	req := &ChangeGCRequest{}
	req.Gc = GContext(order.Uint32(requestBody[0:4]))
	req.ValueMask = order.Uint32(requestBody[4:8])
	req.Values, _ = parseGCValues(order, req.ValueMask, requestBody[8:])
	return req, nil
}

type CopyGCRequest struct {
	SrcGC GContext
	DstGC GContext
}

func (CopyGCRequest) OpCode() reqCode { return CopyGC }

func parseCopyGCRequest(order binary.ByteOrder, requestBody []byte) (*CopyGCRequest, error) {
	req := &CopyGCRequest{}
	req.SrcGC = GContext(order.Uint32(requestBody[0:4]))
	req.DstGC = GContext(order.Uint32(requestBody[4:8]))
	return req, nil
}

type SetDashesRequest struct {
	GC GContext
	DashOffset uint16
	Dashes []byte
}

func (SetDashesRequest) OpCode() reqCode { return SetDashes }

func parseSetDashesRequest(order binary.ByteOrder, requestBody []byte) (*SetDashesRequest, error) {
	req := &SetDashesRequest{}
	req.GC = GContext(order.Uint32(requestBody[0:4]))
	req.DashOffset = order.Uint16(requestBody[4:6])
	nDashes := order.Uint16(requestBody[6:8])
	req.Dashes = requestBody[8 : 8+nDashes]
	return req, nil
}

type SetClipRectanglesRequest struct {
	GC          GContext
	ClippingX   int16
	ClippingY   int16
	Rectangles  []Rectangle
	Ordering    byte
}

func (SetClipRectanglesRequest) OpCode() reqCode { return SetClipRectangles }

func parseSetClipRectanglesRequest(order binary.ByteOrder, data byte, requestBody []byte) (*SetClipRectanglesRequest, error) {
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
	Gc       GContext
	X        int16
	Y        int16
	Items    []PolyText8Item
}

func (PolyText8Request) OpCode() reqCode { return PolyText8 }

func parsePolyText8Request(order binary.ByteOrder, requestBody []byte) (*PolyText8Request, error) {
	req := &PolyText8Request{}
	req.Drawable = Drawable(order.Uint32(requestBody[0:4]))
	req.Gc = GContext(order.Uint32(requestBody[4:8]))
	req.X = int16(order.Uint16(requestBody[8:10]))
	req.Y = int16(order.Uint16(requestBody[10:12]))

	currentPos := 12
	for currentPos < len(requestBody) {
		n := int(requestBody[currentPos])
		currentPos++

		if n == 255 {
			currentPos += 4
		} else if n > 0 {
			delta := int8(requestBody[currentPos])
			currentPos++
			str := requestBody[currentPos : currentPos+n]
			currentPos += n
			req.Items = append(req.Items, PolyText8Item{Delta: delta, Str: str})
		}
		padding := (4 - (n+2)%4) % 4
		currentPos += padding
	}
	return req, nil
}

type PolyText16Request struct {
	Drawable Drawable
	Gc       GContext
	X        int16
	Y        int16
	Items    []PolyText16Item
}

func (PolyText16Request) OpCode() reqCode { return PolyText16 }

func parsePolyText16Request(order binary.ByteOrder, requestBody []byte) (*PolyText16Request, error) {
	req := &PolyText16Request{}
	req.Drawable = Drawable(order.Uint32(requestBody[0:4]))
	req.Gc = GContext(order.Uint32(requestBody[4:8]))
	req.X = int16(order.Uint16(requestBody[8:10]))
	req.Y = int16(order.Uint16(requestBody[10:12]))

	currentPos := 12
	for currentPos < len(requestBody) {
		n := int(requestBody[currentPos])
		currentPos++

		if n == 255 {
			currentPos += 4
		} else if n > 0 {
			delta := int8(requestBody[currentPos])
			currentPos++
			var str []uint16
			for i := 0; i < n; i++ {
				str = append(str, order.Uint16(requestBody[currentPos:currentPos+2]))
				currentPos += 2
			}
			req.Items = append(req.Items, PolyText16Item{Delta: delta, Str: str})
		}
		padding := (4 - (n*2+2)%4) % 4
		currentPos += padding
	}
	return req, nil
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

func parseCreateColormapRequest(order binary.ByteOrder, payload []byte) (*CreateColormapRequest, error) {
	req := &CreateColormapRequest{}
	req.Alloc = payload[0]
	req.Mid = Colormap(order.Uint32(payload[4:8]))
	req.Window = Window(order.Uint32(payload[8:12]))
	req.Visual = VisualID(order.Uint32(payload[12:16]))
	return req, nil
}

type FreeColormapRequest struct {
	Cmap Colormap
}

func (FreeColormapRequest) OpCode() reqCode { return FreeColormap }

func parseFreeColormapRequest(order binary.ByteOrder, requestBody []byte) (*FreeColormapRequest, error) {
	req := &FreeColormapRequest{}
	req.Cmap = Colormap(order.Uint32(requestBody[0:4]))
	return req, nil
}

type InstallColormapRequest struct {
	Cmap Colormap
}

func (InstallColormapRequest) OpCode() reqCode { return InstallColormap }

func parseInstallColormapRequest(order binary.ByteOrder, requestBody []byte) (*InstallColormapRequest, error) {
	req := &InstallColormapRequest{}
	req.Cmap = Colormap(order.Uint32(requestBody[0:4]))
	return req, nil
}

type UninstallColormapRequest struct {
	Cmap Colormap
}

func (UninstallColormapRequest) OpCode() reqCode { return UninstallColormap }

func parseUninstallColormapRequest(order binary.ByteOrder, requestBody []byte) (*UninstallColormapRequest, error) {
	req := &UninstallColormapRequest{}
	req.Cmap = Colormap(order.Uint32(requestBody[0:4]))
	return req, nil
}

type ListInstalledColormapsRequest struct {
	Window Window
}

func (ListInstalledColormapsRequest) OpCode() reqCode { return ListInstalledColormaps }

func parseListInstalledColormapsRequest(order binary.ByteOrder, requestBody []byte) (*ListInstalledColormapsRequest, error) {
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
	req := &AllocColorRequest{}
	req.Cmap = Colormap(order.Uint32(payload[0:4]))
	req.Red = order.Uint16(payload[4:6])
	req.Green = order.Uint16(payload[6:8])
	req.Blue = order.Uint16(payload[8:10])
	return req, nil
}

type AllocNamedColorRequest struct {
	Cmap     xID
	Name     []byte
	Sequence uint16
	MinorOp  byte
	MajorOp  reqCode
}

func (AllocNamedColorRequest) OpCode() reqCode { return AllocNamedColor }

func parseAllocNamedColorRequest(order binary.ByteOrder, payload []byte) (*AllocNamedColorRequest, error) {
	req := &AllocNamedColorRequest{}
	req.Cmap = xID{local: order.Uint32(payload[0:4])}
	nameLen := order.Uint16(payload[4:6])
	req.Name = payload[8 : 8+nameLen]
	return req, nil
}

type FreeColorsRequest struct {
	Cmap      Colormap
	PlaneMask uint32
	Pixels    []uint32
}

func (FreeColorsRequest) OpCode() reqCode { return FreeColors }

func parseFreeColorsRequest(order binary.ByteOrder, requestBody []byte) (*FreeColorsRequest, error) {
	req := &FreeColorsRequest{}
	req.Cmap = Colormap(order.Uint32(requestBody[0:4]))
	req.PlaneMask = order.Uint32(requestBody[4:8])
	numPixels := (len(requestBody) - 8) / 4
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
	req := &StoreNamedColorRequest{}
	req.Cmap = Colormap(order.Uint32(requestBody[0:4]))
	req.Pixel = order.Uint32(requestBody[4:8])
	nameLen := order.Uint16(requestBody[8:10])
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
	req := &LookupColorRequest{}
	req.Cmap = Colormap(order.Uint32(payload[0:4]))
	nameLen := order.Uint16(payload[4:6])
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
	req := &FreeCursorRequest{}
	req.Cursor = Cursor(order.Uint32(requestBody[0:4]))
	return req, nil
}

type RecolorCursorRequest struct {
	Cursor      Cursor
	ForeColor   [3]uint16
	BackColor   [3]uint16
}

func (RecolorCursorRequest) OpCode() reqCode { return RecolorCursor }

func parseRecolorCursorRequest(order binary.ByteOrder, requestBody []byte) (*RecolorCursorRequest, error) {
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
	req := &QueryExtensionRequest{}
	nameLen := order.Uint16(requestBody[0:2])
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

type GetPointerMappingRequest struct {}

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
	req := &ChangeKeyboardMappingRequest{}
	req.KeyCodeCount = data
	req.FirstKeyCode = KeyCode(requestBody[0])
	req.KeySymsPerKeyCode = requestBody[1]
	numKeySyms := int(req.KeyCodeCount) * int(req.KeySymsPerKeyCode)
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
	req := &ChangeKeyboardControlRequest{}
	req.ValueMask = order.Uint32(requestBody[0:4])
	req.Values, _ = parseKeyboardControl(order, req.ValueMask, requestBody[4:])
	return req, nil
}

type GetKeyboardControlRequest struct{}

func (GetKeyboardControlRequest) OpCode() reqCode { return GetKeyboardControl }

func parseGetKeyboardControlRequest(order binary.ByteOrder, requestBody []byte) (*GetKeyboardControlRequest, error) {
	return &GetKeyboardControlRequest{}, nil
}

type SetScreenSaverRequest struct {
	Timeout       int16
	Interval      int16
	PreferBlank   byte
	AllowExpose   byte
}

func (SetScreenSaverRequest) OpCode() reqCode { return SetScreenSaver }

func parseSetScreenSaverRequest(order binary.ByteOrder, requestBody []byte) (*SetScreenSaverRequest, error) {
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
	req := &ChangeHostsRequest{}
	req.Mode = data
	family := requestBody[0]
	addressLen := order.Uint16(requestBody[2:4])
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
	req := &RotatePropertiesRequest{}
	req.Window = Window(order.Uint32(requestBody[0:4]))
	numAtoms := order.Uint16(requestBody[4:6])
	req.Delta = int16(order.Uint16(requestBody[6:8]))
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

func parseKeyboardControl(order binary.ByteOrder, valueMask uint32, valuesData []byte) (KeyboardControl, int) {
	kc := KeyboardControl{}
	offset := 0
	if valueMask&KBKeyClickPercent != 0 {
		kc.KeyClickPercent = int32(order.Uint32(valuesData[offset : offset+4]))
		offset += 4
	}
	if valueMask&KBBellPercent != 0 {
		kc.BellPercent = int32(order.Uint32(valuesData[offset : offset+4]))
		offset += 4
	}
	if valueMask&KBBellPitch != 0 {
		kc.BellPitch = int32(order.Uint32(valuesData[offset : offset+4]))
		offset += 4
	}
	if valueMask&KBBellDuration != 0 {
		kc.BellDuration = int32(order.Uint32(valuesData[offset : offset+4]))
		offset += 4
	}
	if valueMask&KBLed != 0 {
		kc.Led = order.Uint32(valuesData[offset : offset+4])
		offset += 4
	}
	if valueMask&KBLedMode != 0 {
		kc.LedMode = order.Uint32(valuesData[offset : offset+4])
		offset += 4
	}
	if valueMask&KBKey != 0 {
		kc.Key = KeyCode(valuesData[offset])
		offset += 4
	}
	if valueMask&KBAutoRepeatMode != 0 {
		kc.AutoRepeatMode = order.Uint32(valuesData[offset : offset+4])
		offset += 4
	}
	return kc, offset
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
func parseGCValues(order binary.ByteOrder, valueMask uint32, valuesData []byte) (GC, int) {
	gc := GC{}
	offset := 0
	if valueMask&GCFunction != 0 {
		gc.Function = uint32(valuesData[offset])
		offset += 4
	}
	if valueMask&GCPlaneMask != 0 {
		gc.PlaneMask = order.Uint32(valuesData[offset : offset+4])
		offset += 4
	}
	if valueMask&GCForeground != 0 {
		gc.Foreground = order.Uint32(valuesData[offset : offset+4])
		offset += 4
	}
	if valueMask&GCBackground != 0 {
		gc.Background = order.Uint32(valuesData[offset : offset+4])
		offset += 4
	}
	if valueMask&GCLineWidth != 0 {
		gc.LineWidth = order.Uint32(valuesData[offset : offset+4])
		offset += 4
	}
	if valueMask&GCLineStyle != 0 {
		gc.LineStyle = order.Uint32(valuesData[offset : offset+4])
		offset += 4
	}
	if valueMask&GCCapStyle != 0 {
		gc.CapStyle = order.Uint32(valuesData[offset : offset+4])
		offset += 4
	}
	if valueMask&GCJoinStyle != 0 {
		gc.JoinStyle = order.Uint32(valuesData[offset : offset+4])
		offset += 4
	}
	if valueMask&GCFillStyle != 0 {
		gc.FillStyle = order.Uint32(valuesData[offset : offset+4])
		offset += 4
	}
	if valueMask&GCFillRule != 0 {
		gc.FillRule = order.Uint32(valuesData[offset : offset+4])
		offset += 4
	}
	if valueMask&GCTile != 0 {
		gc.Tile = order.Uint32(valuesData[offset : offset+4])
		offset += 4
	}
	if valueMask&GCStipple != 0 {
		gc.Stipple = order.Uint32(valuesData[offset : offset+4])
		offset += 4
	}
	if valueMask&GCTileStipXOrigin != 0 {
		gc.TileStipXOrigin = order.Uint32(valuesData[offset : offset+4])
		offset += 4
	}
	if valueMask&GCTileStipYOrigin != 0 {
		gc.TileStipYOrigin = order.Uint32(valuesData[offset : offset+4])
		offset += 4
	}
	if valueMask&GCFont != 0 {
		gc.Font = order.Uint32(valuesData[offset : offset+4])
		offset += 4
	}
	if valueMask&GCSubwindowMode != 0 {
		gc.SubwindowMode = order.Uint32(valuesData[offset : offset+4])
		offset += 4
	}
	if valueMask&GCGraphicsExposures != 0 {
		gc.GraphicsExposures = order.Uint32(valuesData[offset : offset+4])
		offset += 4
	}
	if valueMask&GCClipXOrigin != 0 {
		gc.ClipXOrigin = int32(order.Uint32(valuesData[offset : offset+4]))
		offset += 4
	}
	if valueMask&GCClipYOrigin != 0 {
		gc.ClipYOrigin = int32(order.Uint32(valuesData[offset : offset+4]))
		offset += 4
	}
	if valueMask&GCClipMask != 0 {
		gc.ClipMask = order.Uint32(valuesData[offset : offset+4])
		offset += 4
	}
	if valueMask&GCDashOffset != 0 {
		gc.DashOffset = order.Uint32(valuesData[offset : offset+4])
		offset += 4
	}
	if valueMask&GCDashList != 0 {
		gc.Dashes = order.Uint32(valuesData[offset : offset+4])
		offset += 4
	}
	if valueMask&GCArcMode != 0 {
		gc.ArcMode = order.Uint32(valuesData[offset : offset+4])
		offset += 4
	}
	return gc, offset
}

func parseWindowAttributes(order binary.ByteOrder, valueMask uint32, valuesData []byte) (WindowAttributes, int) {
	wa := WindowAttributes{}
	offset := 0
	if valueMask&CWBackPixmap != 0 {
		wa.BackgroundPixmap = Pixmap(order.Uint32(valuesData[offset : offset+4]))
		offset += 4
	}
	if valueMask&CWBackPixel != 0 {
		wa.BackgroundPixel = order.Uint32(valuesData[offset : offset+4])
		wa.BackgroundPixelSet = true
		offset += 4
	}
	if valueMask&CWBorderPixmap != 0 {
		wa.BorderPixmap = Pixmap(order.Uint32(valuesData[offset : offset+4]))
		offset += 4
	}
	if valueMask&CWBorderPixel != 0 {
		wa.BorderPixel = order.Uint32(valuesData[offset : offset+4])
		offset += 4
	}
	if valueMask&CWBitGravity != 0 {
		wa.BitGravity = order.Uint32(valuesData[offset : offset+4])
		offset += 4
	}
	if valueMask&CWWinGravity != 0 {
		wa.WinGravity = order.Uint32(valuesData[offset : offset+4])
		offset += 4
	}
	if valueMask&CWBackingStore != 0 {
		wa.BackingStore = order.Uint32(valuesData[offset : offset+4])
		offset += 4
	}
	if valueMask&CWBackingPlanes != 0 {
		wa.BackingPlanes = order.Uint32(valuesData[offset : offset+4])
		offset += 4
	}
	if valueMask&CWBackingPixel != 0 {
		wa.BackingPixel = order.Uint32(valuesData[offset : offset+4])
		offset += 4
	}
	if valueMask&CWOverrideRedirect != 0 {
		wa.OverrideRedirect = order.Uint32(valuesData[offset : offset+4])
		offset += 4
	}
	if valueMask&CWSaveUnder != 0 {
		wa.SaveUnder = order.Uint32(valuesData[offset : offset+4])
		offset += 4
	}
	if valueMask&CWEventMask != 0 {
		wa.EventMask = order.Uint32(valuesData[offset : offset+4])
		offset += 4
	}
	if valueMask&CWDontPropagate != 0 {
		wa.DontPropagateMask = order.Uint32(valuesData[offset : offset+4])
		offset += 4
	}
	if valueMask&CWColormap != 0 {
		wa.Colormap = Colormap(order.Uint32(valuesData[offset : offset+4]))
		offset += 4
	}
	if valueMask&CWCursor != 0 {
		wa.Cursor = Cursor(order.Uint32(valuesData[offset : offset+4]))
		offset += 4
	}
	return wa, offset
}
