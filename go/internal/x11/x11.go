//go:build x11

package x11

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"log"
	"runtime/debug"
	"sync"
	"time"

	"golang.org/x/crypto/ssh"

	"github.com/c2FmZQ/sshterm/internal/x11/wire"
)

var errParseError = errors.New("x11: request parsing error")

type xID struct {
	client uint32
	local  uint32
}

func (x xID) String() string {
	return fmt.Sprintf("%d-%d", x.client, x.local)
}

var (
	x11ServerInstance *x11Server
	once              sync.Once
)

func Enabled() bool {
	return true
}

// Logger is the interface for logging.
type Logger interface {
	Errorf(format string, args ...interface{})
	Infof(format string, args ...interface{})
	Printf(format string, args ...interface{})
}

// X11FrontendAPI is the interface for the X11 frontend.
type X11FrontendAPI interface {
	CreateWindow(xid xID, parent, x, y, width, height, depth, valueMask uint32, values wire.WindowAttributes)
	ChangeWindowAttributes(xid xID, valueMask uint32, values wire.WindowAttributes)
	GetWindowAttributes(xid xID) wire.WindowAttributes
	CreateGC(xid xID, valueMask uint32, values wire.GC)
	ChangeGC(xid xID, valueMask uint32, gc wire.GC)
	DestroyWindow(xid xID)
	ReparentWindow(window xID, parent xID, x, y int16)
	DestroySubwindows(xid xID)
	DestroyAllWindowsForClient(clientID uint32)
	MapWindow(xid xID)
	UnmapWindow(xid xID)
	ConfigureWindow(xid xID, valueMask uint16, values []uint32)
	CirculateWindow(xid xID, direction byte)
	PutImage(drawable xID, gcID xID, format uint8, width, height uint16, dstX, dstY int16, leftPad, depth uint8, data []byte)
	PolyLine(drawable xID, gcID xID, points []uint32)
	PolyFillRectangle(drawable xID, gcID xID, rects []uint32)
	FillPoly(drawable xID, gcID xID, points []uint32)
	PolySegment(drawable xID, gcID xID, segments []uint32)
	PolyPoint(drawable xID, gcID xID, points []uint32)
	PolyRectangle(drawable xID, gcID xID, rects []uint32)
	PolyArc(drawable xID, gcID xID, arcs []uint32)
	PolyFillArc(drawable xID, gcID xID, arcs []uint32)
	ClearArea(drawable xID, x, y, width, height int32)
	CopyArea(srcDrawable, dstDrawable xID, gcID xID, srcX, srcY, dstX, dstY, width, height int32)
	CopyPlane(srcDrawable, dstDrawable xID, gcID xID, srcX, srcY, dstX, dstY, width, height, bitPlane int32)
	GetImage(drawable xID, x, y, width, height int32, format uint32) ([]byte, error)
	ReadClipboard() (string, error)
	WriteClipboard(string) error
	UpdatePointerPosition(x, y int16)
	Bell(percent int8)
	SetInputFocus(focus xID, revertTo byte)
	ImageText8(drawable xID, gcID xID, x, y int32, text []byte)
	ImageText16(drawable xID, gcID xID, x, y int32, text []uint16)
	PolyText8(drawable xID, gcID xID, x, y int32, items []wire.PolyTextItem)
	PolyText16(drawable xID, gcID xID, x, y int32, items []wire.PolyTextItem)
	CreatePixmap(xid, drawable xID, width, height, depth uint32)
	FreePixmap(xid xID)
	CopyPixmap(srcID, dstID, gcID xID, srcX, srcY, width, height, dstX, dstY uint32)
	CreateCursor(cursorID xID, source, mask xID, foreColor, backColor [3]uint16, x, y uint16)
	CreateCursorFromGlyph(cursorID uint32, glyphID uint16)
	SetWindowCursor(windowID xID, cursorID xID)
	CopyGC(srcGC, dstGC xID)
	FreeGC(gc xID)
	FreeCursor(cursorID xID)
	SendEvent(eventData messageEncoder)
	GetFocusWindow(clientID uint32) xID
	SetWindowTitle(xid xID, title string)
	GrabPointer(grabWindow xID, ownerEvents bool, eventMask uint16, pointerMode, keyboardMode byte, confineTo uint32, cursor uint32, time uint32) byte
	UngrabPointer(time uint32)
	GrabKeyboard(grabWindow xID, ownerEvents bool, time uint32, pointerMode, keyboardMode byte) byte
	UngrabKeyboard(time uint32)
	WarpPointer(x, y int16)
	GetCanvasOperations() []CanvasOperation
	GetRGBColor(colormap xID, pixel uint32) (r, g, b uint8)
	OpenFont(fid xID, name string)
	QueryFont(fid xID) (minBounds, maxBounds wire.XCharInfo, minCharOrByte2, maxCharOrByte2, defaultChar uint16, drawDirection uint8, minByte1, maxByte1 uint8, allCharsExist bool, fontAscent, fontDescent int16, charInfos []wire.XCharInfo, fontProps []wire.FontProp)
	QueryTextExtents(font xID, text []uint16) (drawDirection uint8, fontAscent, fontDescent, overallAscent, overallDescent, overallWidth, overallLeft, overallRight int16)
	CloseFont(fid xID)
	ListFonts(maxNames uint16, pattern string) []string
	AllowEvents(clientID uint32, mode byte, time uint32)
	SetDashes(gc xID, dashOffset uint16, dashes []byte)
	SetClipRectangles(gc xID, clippingX, clippingY int16, rectangles []wire.Rectangle, ordering byte)
	RecolorCursor(cursor xID, foreColor, backColor [3]uint16)
	SetPointerMapping(pMap []byte) (byte, error)
	GetPointerMapping() ([]byte, error)
	GetPointerControl() (accelNumerator, accelDenominator, threshold uint16, err error)
	ChangePointerControl(accelNum, accelDenom, threshold int16, doAccel, doThresh bool)
	ChangeKeyboardControl(valueMask uint32, values wire.KeyboardControl)
	GetKeyboardControl() (wire.KeyboardControl, error)
	SetScreenSaver(timeout, interval int16, preferBlank, allowExpose byte)
	GetScreenSaver() (timeout, interval int16, preferBlank, allowExpose byte, err error)
	ChangeHosts(mode byte, host wire.Host)
	ListHosts() ([]wire.Host, error)
	SetAccessControl(mode byte)
	SetCloseDownMode(mode byte)
	KillClient(resource uint32)
	ForceScreenSaver(mode byte)
	SetModifierMapping(keyCodesPerModifier byte, keyCodes []wire.KeyCode) (byte, error)
	GetModifierMapping() ([]wire.KeyCode, error)
	DeviceBell(deviceID byte, feedbackID byte, feedbackClass byte, percent int8)
	XIChangeHierarchy(changes []wire.XIChangeHierarchyChange)
	ChangeFeedbackControl(deviceID byte, feedbackID byte, mask uint32, control []byte)
	ChangeDeviceKeyMapping(deviceID byte, firstKey byte, keysymsPerKeycode byte, keycodeCount byte, keysyms []uint32)
	SetDeviceModifierMapping(deviceID byte, keycodes []byte) byte
	SetDeviceButtonMapping(deviceID byte, buttonMap []byte) byte
	GetFeedbackControl(deviceID byte) []wire.FeedbackState
	GetDeviceKeyMapping(deviceID byte, firstKey byte, count byte) (byte, []uint32)
	GetDeviceModifierMapping(deviceID byte) (byte, []byte)
	GetDeviceButtonMapping(deviceID byte) []byte
	QueryDeviceState(deviceID byte) []wire.InputClassInfo
	ComposeWindow(xid xID)
}

type XError interface {
	Code() byte
	Sequence() uint16
	BadValue() uint32
	MinorOp() byte
	MajorOp() byte
}

// CanvasOperation represents a single canvas drawing operation captured from the frontend.
type CanvasOperation struct {
	Type        string `json:"type"`
	Args        []any  `json:"args"`
	FillStyle   string `json:"fillStyle"`
	StrokeStyle string `json:"strokeStyle"`
}

type window struct {
	xid                       xID
	parent                    uint32
	x, y                      int16
	width, height             uint16
	mapped                    bool
	depth                     byte
	children                  []uint32
	attributes                wire.WindowAttributes
	colormap                  xID
	dontPropagateDeviceEvents map[uint32]bool
}

func (w *window) mapState() byte {
	if !w.mapped {
		return 0 // Unmapped
	}
	return 2 // Viewable
}

type colormap struct {
	pixels map[uint32]wire.XColorItem
}

type property struct {
	data     []byte
	typeAtom uint32
	format   byte
}

type selectionOwner struct {
	window xID
	time   uint32
}

type DeviceButtonPressEventData struct {
	Event  uint32
	RootX  uint16
	RootY  uint16
	EventX uint16
	EventY uint16
}

type x11Server struct {
	logger                Logger
	byteOrder             binary.ByteOrder
	frontend              X11FrontendAPI
	windows               map[xID]*window
	gcs                   map[xID]wire.GC
	pixmaps               map[xID]bool
	cursors               map[xID]bool
	selections            map[uint32]*selectionOwner
	atoms                 map[string]uint32
	atomNames             map[uint32]string
	nextAtomID            uint32
	properties            map[xID]map[uint32]*property
	colormaps             map[xID]*colormap
	defaultColormap       uint32
	installedColormap     xID
	visualID              uint32
	rootVisual            wire.VisualType
	rootWindowWidth       uint16
	rootWindowHeight      uint16
	blackPixel            uint32
	whitePixel            uint32
	pointerX, pointerY    int16
	clients               map[uint32]*x11Client
	nextClientID          uint32
	pointerGrabWindow     xID
	keyboardGrabWindow    xID
	pointerGrabTime       uint32
	keyboardGrabTime      uint32
	pointerGrabOwner      bool
	keyboardGrabOwner     bool
	pointerGrabEventMask  uint16
	keyboardGrabEventMask uint32
	inputFocus            xID
	passiveGrabs          map[xID][]*passiveGrab
	passiveDeviceGrabs    map[xID][]*passiveDeviceGrab
	deviceGrabs           map[byte]*deviceGrab // device id -> grab info
	authProtocol          string
	authCookie            []byte
	serverGrabbed         bool
	grabbingClientID      uint32
	fontPath              []string
	keymap                map[byte]uint32
	pointerState          uint16
	startTime             time.Time
	pointerGrabMode       byte
	keyboardGrabMode      byte
	pointerGrabConfineTo  xID
	pointerGrabCursor     xID
	fonts                 map[xID]bool
}

type passiveGrab struct {
	clientID     uint32
	button       byte
	key          wire.KeyCode
	modifiers    uint16
	owner        bool
	eventMask    uint16
	cursor       xID
	pointerMode  byte
	keyboardMode byte
	confineTo    xID
}

type passiveDeviceGrab struct {
	clientID     uint32
	deviceID     byte
	key          wire.KeyCode
	button       byte
	detail       uint32
	modifiers    uint16
	xi2Modifiers []uint32
	owner        bool
	eventMask    []uint32
	xi2EventMask []uint32
	xi2GrabType  int
}

type deviceGrab struct {
	window       xID
	ownerEvents  bool
	eventMask    []uint32
	xi2EventMask []uint32
	time         uint32
}

var virtualPointer = &wire.DeviceInfo{
	Header: wire.DeviceHeader{
		DeviceID:   2,
		DeviceType: 0,
		NumClasses: 2,
		Use:        0, // IsXPointer
		Name:       "Virtual Pointer",
	},
	Classes: []wire.InputClassInfo{
		&wire.ButtonClassInfo{NumButtons: 5},
		&wire.ValuatorClassInfo{
			NumAxes:    2,
			Mode:       0, // Relative
			MotionSize: 0,
			Axes: []wire.ValuatorAxisInfo{
				{Min: 0, Max: 65535, Resolution: 1},
				{Min: 0, Max: 65535, Resolution: 1},
			},
		},
	},
}

var virtualKeyboard = &wire.DeviceInfo{
	Header: wire.DeviceHeader{
		DeviceID:   3,
		DeviceType: 0,
		NumClasses: 1,
		Use:        1, // IsXKeyboard
		Name:       "Virtual Keyboard",
	},
	Classes: []wire.InputClassInfo{
		&wire.KeyClassInfo{
			NumKeys:    248,
			MinKeycode: 8,
			MaxKeycode: 255,
		},
	},
}

func (s *x11Server) SetRootWindowSize(width, height uint16) {
	s.rootWindowWidth = width
	s.rootWindowHeight = height
}

func (s *x11Server) serverTime() uint32 {
	return uint32(time.Since(s.startTime).Milliseconds())
}

func (s *x11Server) UpdatePointerPosition(x, y int16) {
	s.pointerX = x
	s.pointerY = y
}

func (s *x11Server) GetWindowAttributes(xid xID) (wire.WindowAttributes, bool) {
	w, ok := s.windows[xid]
	if !ok {
		return wire.WindowAttributes{}, false
	}
	return w.attributes, true
}

func (s *x11Server) checkWindow(xid xID, seq uint16, majorReq wire.ReqCode, minorReq byte) wire.Error {
	if xid.local == s.rootWindowID() {
		return nil
	}
	if _, ok := s.windows[xid]; !ok {
		return wire.NewGenericError(seq, xid.local, minorReq, majorReq, wire.WindowErrorCode)
	}
	return nil
}

func (s *x11Server) checkPixmap(xid xID, seq uint16, majorReq wire.ReqCode, minorReq byte) wire.Error {
	if _, ok := s.pixmaps[xid]; !ok {
		return wire.NewGenericError(seq, xid.local, minorReq, majorReq, wire.PixmapErrorCode)
	}
	return nil
}

func (s *x11Server) checkDrawable(xid xID, seq uint16, majorReq wire.ReqCode, minorReq byte) wire.Error {
	if _, ok := s.windows[xid]; ok {
		return nil
	}
	if _, ok := s.pixmaps[xid]; ok {
		return nil
	}
	if xid.local == s.rootWindowID() {
		return nil
	}
	return wire.NewGenericError(seq, xid.local, minorReq, majorReq, wire.DrawableErrorCode)
}

func (s *x11Server) checkGC(xid xID, seq uint16, majorReq wire.ReqCode, minorReq byte) wire.Error {
	if _, ok := s.gcs[xid]; !ok {
		return wire.NewGenericError(seq, xid.local, minorReq, majorReq, wire.GContextErrorCode)
	}
	return nil
}

func (s *x11Server) checkCursor(xid xID, seq uint16, majorReq wire.ReqCode, minorReq byte) wire.Error {
	if _, ok := s.cursors[xid]; !ok {
		return wire.NewGenericError(seq, xid.local, minorReq, majorReq, wire.CursorErrorCode)
	}
	return nil
}

func (s *x11Server) checkColormap(xid xID, seq uint16, majorReq wire.ReqCode, minorReq byte) wire.Error {
	if xid.local == s.defaultColormap {
		xid.client = 0
	}
	if _, ok := s.colormaps[xid]; !ok {
		return wire.NewGenericError(seq, xid.local, minorReq, majorReq, wire.ColormapErrorCode)
	}
	return nil
}

func (s *x11Server) checkFont(xid xID, seq uint16, majorReq wire.ReqCode, minorReq byte) wire.Error {
	if _, ok := s.fonts[xid]; !ok {
		return wire.NewGenericError(seq, xid.local, minorReq, majorReq, wire.FontErrorCode)
	}
	return nil
}

func NewWindowAttributes() wire.WindowAttributes {
	return wire.WindowAttributes{}
}

func (s *x11Server) SendMouseEvent(xid xID, eventType string, x, y, detail int32) {
	debugf("X11: SendMouseEvent xid=%s type=%s x=%d y=%d detail=%d", xid, eventType, x, y, detail)

	originalXID := xid
	if _, ok := s.windows[originalXID]; !ok {
		log.Printf("X11: Failed to write mouse event: window not found")
		return
	}

	// Calculate delta
	dx := float64(x - int32(s.pointerX))
	dy := float64(y - int32(s.pointerY))

	state := uint16(detail >> 16)
	s.pointerState = state
	button := byte(detail & 0xFFFF)
	deviceID := virtualPointer.Header.DeviceID

	// 0. Handle Raw Events (XI 2.x)
	// "Raw events are sent to all clients that have selected for the event type on the root window."
	// We iterate all clients to check if they selected Raw events on the Root Window.
	for _, client := range s.clients {
		if client.xi2EventMasks != nil {
			if devMasks, ok := client.xi2EventMasks[s.rootWindowID()]; ok {
				var mask []uint32
				if m, ok := devMasks[uint16(deviceID)]; ok {
					mask = m
				} else if m, ok := devMasks[wire.XIAllMasterDevices]; ok {
					mask = m
				}

				if mask != nil {
					var evType uint16
					switch eventType {
					case "mousedown":
						evType = wire.XI_RawButtonPress
					case "mouseup":
						evType = wire.XI_RawButtonRelease
					case "mousemove":
						evType = wire.XI_RawMotion
					}

					if evType > 0 {
						wordIdx := int(evType / 32)
						bitIdx := int(evType % 32)
						if wordIdx < len(mask) && (mask[wordIdx]&(1<<bitIdx)) != 0 {
							s.sendXInput2RawEvent(client, evType, deviceID, button, dx, dy)
						}
					}
				}
			}
		}
	}

	// 1. Handle Device Grabs (Active and Passive)
	activeDeviceGrab, deviceGrabbed := s.deviceGrabs[deviceID]

	// Check passive device grabs if no active grab
	if !deviceGrabbed && eventType == "mousedown" {
		if grabs, ok := s.passiveDeviceGrabs[originalXID]; ok {
			for _, grab := range grabs {
				xi1Match := grab.xi2GrabType == 0 && grab.deviceID == deviceID && grab.button == button && (grab.modifiers == wire.AnyModifier || grab.modifiers == state)

				xi2Match := false
				if grab.xi2GrabType == wire.XI_ButtonPress && grab.deviceID == deviceID && byte(grab.detail) == button {
					if len(grab.xi2Modifiers) == 0 {
						xi2Match = true
					} else {
						for _, mod := range grab.xi2Modifiers {
							if (mod & 0xFFFF) == uint32(state) {
								xi2Match = true
								break
							}
						}
					}
				}

				if xi1Match || xi2Match {
					activeDeviceGrab = &deviceGrab{
						window:       originalXID, // Grab is on this window
						ownerEvents:  grab.owner,
						eventMask:    grab.eventMask,
						xi2EventMask: grab.xi2EventMask,
						time:         s.serverTime(),
					}
					s.deviceGrabs[deviceID] = activeDeviceGrab
					deviceGrabbed = true
					activeDeviceGrab.window.client = grab.clientID // Ensure client is set correctly
					break
				}
			}
		}
	}

	var xiEventMask uint32
	switch eventType {
	case "mousedown":
		xiEventMask = wire.DeviceButtonPressMask
	case "mouseup":
		xiEventMask = wire.DeviceButtonReleaseMask
	}

	if deviceGrabbed {
		// Send XInput event to grabbing client if mask matches
		grabbingClient, clientExists := s.clients[activeDeviceGrab.window.client]
		if clientExists {
			// XI 1.x
			if xiEventMask > 0 {
				match := false
				for _, class := range activeDeviceGrab.eventMask {
					// class is (mask << 8) | deviceID
					if byte(class&0xFF) == deviceID {
						mask := class >> 8
						if mask&uint32(xiEventMask) != 0 {
							match = true
							break
						}
					}
				}
				if match {
					s.sendXInputMouseEvent(grabbingClient, eventType, deviceID, button, originalXID.local, x, y, state)
				}
			}

			// XI 2.x
			if activeDeviceGrab.xi2EventMask != nil {
				var evType uint16
				switch eventType {
				case "mousedown":
					evType = wire.XI_ButtonPress
				case "mouseup":
					evType = wire.XI_ButtonRelease
				case "mousemove":
					evType = wire.XI_Motion
				}

				if evType > 0 {
					wordIdx := int(evType / 32)
					bitIdx := int(evType % 32)
					if wordIdx < len(activeDeviceGrab.xi2EventMask) && (activeDeviceGrab.xi2EventMask[wordIdx]&(1<<bitIdx)) != 0 {
						s.sendXInput2MouseEvent(grabbingClient, evType, deviceID, button, originalXID.local, x, y, state)
					}
				}
			}
		}
		// Core events are suppressed when device is grabbed
		return
	}

	// 2. Handle Core Grabs and Events (only if device not grabbed)
	grabActive := s.pointerGrabWindow.local != 0
	if grabActive {
		xid = s.pointerGrabWindow
	}

	var eventMask uint32
	switch eventType {
	case "mousedown":
		eventMask = wire.ButtonPressMask
	case "mouseup":
		eventMask = wire.ButtonReleaseMask
	case "mousemove":
		eventMask = wire.PointerMotionMask
		if state&wire.Button1Mask != 0 {
			eventMask |= wire.Button1MotionMask
		}
		if state&wire.Button2Mask != 0 {
			eventMask |= wire.Button2MotionMask
		}
		if state&wire.Button3Mask != 0 {
			eventMask |= wire.Button3MotionMask
		}
		if state&wire.Button4Mask != 0 {
			eventMask |= wire.Button4MotionMask
		}
		if state&wire.Button5Mask != 0 {
			eventMask |= wire.Button5MotionMask
		}
		if state&uint16(wire.ButtonPressMask|wire.ButtonReleaseMask) != 0 {
			eventMask |= wire.ButtonMotionMask
		}
	}

	if !grabActive && eventType == "mousedown" {
		if grabs, ok := s.passiveGrabs[originalXID]; ok {
			for _, grab := range grabs {
				if grab.button == button && (grab.modifiers == wire.AnyModifier || grab.modifiers == state) {
					s.pointerGrabWindow = originalXID
					s.pointerGrabWindow.client = grab.clientID
					s.pointerGrabOwner = grab.owner
					s.pointerGrabEventMask = grab.eventMask
					s.pointerGrabMode = grab.pointerMode
					s.keyboardGrabMode = grab.keyboardMode
					s.pointerGrabConfineTo = grab.confineTo
					s.pointerGrabCursor = grab.cursor
					s.pointerGrabTime = s.serverTime()
					grabActive = true
					s.frontend.SetWindowCursor(originalXID, grab.cursor)
					break
				}
			}
		}
	}

	// Dispatch Core events.
	if grabActive {
		grabbingClient, grabberOk := s.clients[s.pointerGrabWindow.client]
		if grabberOk && (uint32(s.pointerGrabEventMask)&eventMask) != 0 {
			eventWindowID := originalXID.local
			if !s.pointerGrabOwner {
				eventWindowID = s.pointerGrabWindow.local
			}
			s.sendCoreMouseEvent(grabbingClient, eventType, button, eventWindowID, x, y, state)
		}

		if s.pointerGrabOwner {
			ownerClient, ownerOk := s.clients[originalXID.client]
			if ownerOk && (!grabberOk || ownerClient.id != grabbingClient.id) {
				if w, ok := s.windows[originalXID]; ok {
					if w.attributes.EventMask&eventMask != 0 {
						s.sendCoreMouseEvent(ownerClient, eventType, button, originalXID.local, x, y, state)
					}
				}
			}
		}
	} else {
		for _, client := range s.clients {
			w, ok := s.windows[client.xID(xid.local)]
			if ok && w.attributes.EventMask&eventMask != 0 {
				s.sendCoreMouseEvent(client, eventType, button, originalXID.local, x, y, state)
			}
		}
	}

	// 3. Send XInput events (non-grabbed)
	for _, client := range s.clients {
		// XI 1.x
		if xiEventMask > 0 {
			if deviceInfo, ok := client.openDevices[deviceID]; ok {
				if mask, ok := deviceInfo.EventMasks[originalXID.local]; ok {
					if mask&xiEventMask != 0 {
						s.sendXInputMouseEvent(client, eventType, deviceID, button, originalXID.local, x, y, state)
					}
				}
			}
		}

		// XI 2.x
		if client.xi2EventMasks != nil {
			if devMasks, ok := client.xi2EventMasks[originalXID.local]; ok {
				// Check specific device or AllDevices (0) or AllMasterDevices (1)
				// For now, check deviceID (2) and AllMasterDevices (1)
				var mask []uint32
				if m, ok := devMasks[uint16(deviceID)]; ok {
					mask = m
				} else if m, ok := devMasks[1]; ok { // XIAllMasterDevices
					mask = m
				}

				if mask != nil {
					var evType uint16
					switch eventType {
					case "mousedown":
						evType = 4 // XI_ButtonPress
					case "mouseup":
						evType = 5 // XI_ButtonRelease
					case "mousemove":
						evType = 6 // XI_Motion
					}

					if evType > 0 {
						// Check bit in mask
						// Mask is []uint32. Bit N is in word N/32 at bit N%32
						wordIdx := int(evType / 32)
						bitIdx := int(evType % 32)
						if wordIdx < len(mask) && (mask[wordIdx]&(1<<bitIdx)) != 0 {
							s.sendXInput2MouseEvent(client, evType, deviceID, button, originalXID.local, x, y, state)
						}
					}
				}
			}
		}
	}
}

func (s *x11Server) sendXInput2RawEvent(client *x11Client, evType uint16, deviceID byte, detail byte, dx, dy float64) {
	// Construct mask: set bits 0 (X) and 1 (Y) if they changed (non-zero delta) or just always set them?
	// Standard usually reports both axes for motion.
	// Mask is []uint32. We use 1 uint32 (enough for axes 0-31).

	var mask uint32
	var values []float64

	// Always report X and Y for RawMotion to be safe, or check non-zero?
	// xeyes expects relative motion.

	// Axis 0: X
	mask |= 1 << 0
	values = append(values, dx)

	// Axis 1: Y
	mask |= 1 << 1
	values = append(values, dy)

	valuatorsMask := []uint32{mask}

	event := &wire.XIRawEvent{
		Sequence:       client.sequence - 1,
		EventType:      evType,
		DeviceID:       uint16(deviceID),
		Time:           s.serverTime(),
		Detail:         uint32(detail),
		SourceID:       uint16(deviceID),
		ValuatorsMask:  valuatorsMask,
		ValuatorValues: values,
		RawValues:      values, // Assuming raw == relative for virtual device
	}

	if err := client.send(event); err != nil {
		debugf("X11: Failed to write XInput2 raw event: %v", err)
	}
}

func (s *x11Server) sendXInput2MouseEvent(client *x11Client, evType uint16, deviceID byte, button byte, eventWindowID uint32, x, y int32, state uint16) {
	// Modifiers
	mods := wire.ModifierInfo{
		Base:      uint32(state),
		Latched:   0,
		Locked:    0,
		Effective: uint32(state),
	}

	// Buttons mask
	// Convert X11 state mask to XI2 button mask
	buttonMask := uint32(0)
	if state&wire.Button1Mask != 0 {
		buttonMask |= (1 << 0)
	}
	if state&wire.Button2Mask != 0 {
		buttonMask |= (1 << 1)
	}
	if state&wire.Button3Mask != 0 {
		buttonMask |= (1 << 2)
	}
	if state&wire.Button4Mask != 0 {
		buttonMask |= (1 << 3)
	}
	if state&wire.Button5Mask != 0 {
		buttonMask |= (1 << 4)
	}

	// For ButtonPress/Release, we might need to ensure the button is set/unset correctly.
	// Assuming `state` reflects the state *after* the event (as per JS semantics).

	buttons := []uint32{buttonMask}

	// FP1616 coordinates
	rootX := x << 16
	rootY := y << 16
	eventX := x << 16
	eventY := y << 16

	event := &wire.XIDeviceEvent{
		Sequence:  client.sequence - 1,
		EventType: evType,
		DeviceID:  uint16(deviceID),
		Time:      s.serverTime(),
		Detail:    uint32(button),
		Root:      s.rootWindowID(),
		Event:     eventWindowID,
		Child:     0,
		RootX:     rootX,
		RootY:     rootY,
		EventX:    eventX,
		EventY:    eventY,
		Buttons:   buttons,
		Valuators: nil,
		SourceID:  uint16(deviceID), // For now, SourceID is same as DeviceID (Master)
		Mods:      mods,
		Group:     wire.GroupInfo{},
	}

	// Wrap in GenericEvent
	// XIDeviceEvent.EncodeMessage returns the full packet including the GenericEvent header.
	// client.send calls EncodeMessage, so we can pass the event directly.

	if err := client.send(event); err != nil {
		debugf("X11: Failed to write XInput2 mouse event: %v", err)
	}
}

func (s *x11Server) sendXInput2KeyboardEvent(client *x11Client, evType uint16, deviceID byte, keycode byte, eventWindowID uint32, state uint16) {
	mods := wire.ModifierInfo{
		Base:      uint32(state),
		Latched:   0,
		Locked:    0,
		Effective: uint32(state),
	}

	buttonMask := uint32(0)
	if state&wire.Button1Mask != 0 {
		buttonMask |= (1 << 0)
	}
	if state&wire.Button2Mask != 0 {
		buttonMask |= (1 << 1)
	}
	if state&wire.Button3Mask != 0 {
		buttonMask |= (1 << 2)
	}
	if state&wire.Button4Mask != 0 {
		buttonMask |= (1 << 3)
	}
	if state&wire.Button5Mask != 0 {
		buttonMask |= (1 << 4)
	}

	buttons := []uint32{buttonMask}

	event := &wire.XIDeviceEvent{
		Sequence:  client.sequence - 1,
		EventType: evType,
		DeviceID:  uint16(deviceID),
		Time:      s.serverTime(),
		Detail:    uint32(keycode),
		Root:      s.rootWindowID(),
		Event:     eventWindowID,
		Child:     0,
		RootX:     int32(s.pointerX) << 16,
		RootY:     int32(s.pointerY) << 16,
		EventX:    int32(s.pointerX) << 16,
		EventY:    int32(s.pointerY) << 16,
		Buttons:   buttons,
		Valuators: nil,
		SourceID:  uint16(deviceID),
		Mods:      mods,
		Group:     wire.GroupInfo{},
	}

	if err := client.send(event); err != nil {
		debugf("X11: Failed to write XInput2 keyboard event: %v", err)
	}
}

func (s *x11Server) sendCoreMouseEvent(client *x11Client, eventType string, button byte, eventWindowID uint32, x, y int32, state uint16) {
	var event messageEncoder
	switch eventType {
	case "mousedown":
		event = &wire.ButtonPressEvent{
			Sequence:   client.sequence - 1,
			Detail:     button,
			Time:       s.serverTime(),
			Root:       s.rootWindowID(),
			Event:      eventWindowID,
			Child:      0, // 0 for now
			RootX:      int16(x),
			RootY:      int16(y),
			EventX:     int16(x),
			EventY:     int16(y),
			State:      state,
			SameScreen: true,
		}
	case "mouseup":
		event = &wire.ButtonReleaseEvent{
			Sequence:   client.sequence - 1,
			Detail:     button,
			Time:       s.serverTime(),
			Root:       s.rootWindowID(),
			Event:      eventWindowID,
			Child:      0, // 0 for now
			RootX:      int16(x),
			RootY:      int16(y),
			EventX:     int16(x),
			EventY:     int16(y),
			State:      state,
			SameScreen: true,
		}
	case "mousemove":
		event = &wire.MotionNotifyEvent{
			Sequence:   client.sequence - 1,
			Detail:     0, // 0 for Normal
			Time:       s.serverTime(),
			Root:       s.rootWindowID(),
			Event:      eventWindowID,
			Child:      0, // 0 for now
			RootX:      int16(x),
			RootY:      int16(y),
			EventX:     int16(x),
			EventY:     int16(y),
			State:      state,
			SameScreen: true,
		}
	default:
		debugf("X11: Unknown mouse event type: %s", eventType)
		return
	}
	if err := client.send(event); err != nil {
		debugf("X11: Failed to write mouse event: %v", err)
	}
}

func (s *x11Server) sendXInputMouseEvent(client *x11Client, eventType string, deviceID, button byte, eventWindowID uint32, x, y int32, state uint16) {
	var xiEvent messageEncoder
	switch eventType {
	case "mousedown":
		xiEvent = &wire.DeviceButtonPressEvent{
			Sequence:   client.sequence - 1,
			DeviceID:   deviceID,
			Time:       s.serverTime(),
			Detail:     button,
			Root:       s.rootWindowID(),
			Event:      eventWindowID,
			Child:      0, // Or a child window ID if applicable
			RootX:      int16(x),
			RootY:      int16(y),
			EventX:     int16(x),
			EventY:     int16(y),
			State:      state,
			SameScreen: true,
		}
	case "mouseup":
		xiEvent = &wire.DeviceButtonReleaseEvent{
			Sequence:   client.sequence - 1,
			DeviceID:   deviceID,
			Time:       s.serverTime(),
			Button:     button,
			Root:       s.rootWindowID(),
			Event:      eventWindowID,
			Child:      0, // Or a child window ID if applicable
			RootX:      int16(x),
			RootY:      int16(y),
			EventX:     int16(x),
			EventY:     int16(y),
			State:      state,
			SameScreen: true,
		}
	}

	if xiEvent != nil {
		if err := client.send(xiEvent); err != nil {
			debugf("X11: Failed to write XInput mouse event: %v", err)
		}
	}
}

func (s *x11Server) SendKeyboardEvent(xid xID, eventType string, code string, altKey, ctrlKey, shiftKey, metaKey bool) {
	debugf("X11: SendKeyboardEvent xid=%s type=%s code=%s alt=%t ctrl=%t shift=%t meta=%t", xid, eventType, code, altKey, ctrlKey, shiftKey, metaKey)

	state := uint16(0)
	if shiftKey {
		state |= 1
	}
	if ctrlKey {
		state |= 4
	}
	if altKey {
		state |= 8
	}
	if metaKey {
		state |= 64
	}

	// Preserve button state
	state |= (s.pointerState & 0xFF00)
	s.pointerState = state

	keycode, ok := jsCodeToX11Keycode[code]
	if !ok {
		keycode = jsCodeToX11Keycode["Unidentified"]
	}

	deviceID := virtualKeyboard.Header.DeviceID

	// 1. Handle Device Grabs (Active and Passive)
	activeDeviceGrab, deviceGrabbed := s.deviceGrabs[deviceID]

	if !deviceGrabbed && eventType == "keydown" {
		if grabs, ok := s.passiveDeviceGrabs[xid]; ok {
			for _, grab := range grabs {
				xi1Match := grab.xi2GrabType == 0 && grab.deviceID == deviceID && grab.key == wire.KeyCode(keycode) && (grab.modifiers == wire.AnyModifier || grab.modifiers == state)

				xi2Match := false
				if grab.xi2GrabType == wire.XI_KeyPress && grab.deviceID == deviceID && byte(grab.detail) == byte(keycode) {
					if len(grab.xi2Modifiers) == 0 {
						xi2Match = true
					} else {
						for _, mod := range grab.xi2Modifiers {
							if (mod & 0xFFFF) == uint32(state) {
								xi2Match = true
								break
							}
						}
					}
				}

				if xi1Match || xi2Match {
					activeDeviceGrab = &deviceGrab{
						window:       xid,
						ownerEvents:  grab.owner,
						eventMask:    grab.eventMask,
						xi2EventMask: grab.xi2EventMask,
						time:         s.serverTime(),
					}
					s.deviceGrabs[deviceID] = activeDeviceGrab
					deviceGrabbed = true
					activeDeviceGrab.window.client = grab.clientID
					break
				}
			}
		}
	}

	var xiEventMask uint32
	switch eventType {
	case "keydown":
		xiEventMask = wire.DeviceKeyPressMask
	case "keyup":
		xiEventMask = wire.DeviceKeyReleaseMask
	}

	if deviceGrabbed {
		grabbingClient, clientExists := s.clients[activeDeviceGrab.window.client]
		if clientExists {
			if xiEventMask > 0 {
				match := false
				for _, class := range activeDeviceGrab.eventMask {
					if byte(class&0xFF) == deviceID {
						mask := class >> 8
						if mask&uint32(xiEventMask) != 0 {
							match = true
							break
						}
					}
				}
				if match {
					s.sendXInputKeyboardEvent(grabbingClient, eventType, keycode, s.inputFocus.local, state)
				}
			}

			if activeDeviceGrab.xi2EventMask != nil {
				var evType uint16
				switch eventType {
				case "keydown":
					evType = wire.XI_KeyPress
				case "keyup":
					evType = wire.XI_KeyRelease
				}

				if evType > 0 {
					wordIdx := int(evType / 32)
					bitIdx := int(evType % 32)
					if wordIdx < len(activeDeviceGrab.xi2EventMask) && (activeDeviceGrab.xi2EventMask[wordIdx]&(1<<bitIdx)) != 0 {
						s.sendXInput2KeyboardEvent(grabbingClient, evType, deviceID, keycode, s.inputFocus.local, state)
					}
				}
			}
		}
		return
	}

	// 2. Handle Core Grabs and Events
	grabActive := s.keyboardGrabWindow.local != 0

	var eventMask uint32
	switch eventType {
	case "keydown":
		eventMask = wire.KeyPressMask
	case "keyup":
		eventMask = wire.KeyReleaseMask
	}

	if !grabActive && eventType == "keydown" {
		if grabs, ok := s.passiveGrabs[xid]; ok {
			for _, grab := range grabs {
				if grab.key == wire.KeyCode(keycode) && (grab.modifiers == wire.AnyModifier || grab.modifiers == state) {
					s.keyboardGrabWindow = xid
					s.keyboardGrabWindow.client = grab.clientID
					s.keyboardGrabOwner = grab.owner
					s.pointerGrabMode = grab.pointerMode
					s.keyboardGrabMode = grab.keyboardMode
					s.keyboardGrabTime = s.serverTime()
					grabActive = true
					if client, ok := s.clients[s.keyboardGrabWindow.client]; ok {
						s.sendCoreKeyboardEvent(client, eventType, keycode, xid.local, state)
					}
					return
				}
			}
		}
	}

	if grabActive {
		grabbingClient, grabberOk := s.clients[s.keyboardGrabWindow.client]
		if grabberOk {
			eventWindow := s.keyboardGrabWindow.local
			if s.keyboardGrabOwner {
				eventWindow = xid.local
			}
			s.sendCoreKeyboardEvent(grabbingClient, eventType, keycode, eventWindow, state)
		}

		if s.keyboardGrabOwner {
			ownerClient, ownerOk := s.clients[xid.client]
			if ownerOk && (!grabberOk || ownerClient.id != grabbingClient.id) {
				if w, ok := s.windows[xid]; ok {
					if w.attributes.EventMask&eventMask != 0 {
						s.sendCoreKeyboardEvent(ownerClient, eventType, keycode, xid.local, state)
					}
				}
			}
		}
		return
	}

	// No active grab, send to interested clients
	for _, client := range s.clients {
		if w, ok := s.windows[client.xID(s.inputFocus.local)]; ok {
			if w.attributes.EventMask&eventMask != 0 {
				s.sendCoreKeyboardEvent(client, eventType, keycode, s.inputFocus.local, state)
			}
		}
	}

	// 3. Send XInput events (non-grabbed)
	if xiEventMask > 0 {
		for _, client := range s.clients {
			if deviceInfo, ok := client.openDevices[virtualKeyboard.Header.DeviceID]; ok {
				if mask, ok := deviceInfo.EventMasks[s.inputFocus.local]; ok {
					if mask&xiEventMask != 0 {
						s.sendXInputKeyboardEvent(client, eventType, keycode, s.inputFocus.local, state)
					}
				}
			}
		}
	}
}

func (s *x11Server) sendCoreKeyboardEvent(client *x11Client, eventType string, keycode byte, eventWindowID uint32, state uint16) {
	event := &wire.KeyEvent{
		Sequence:   client.sequence - 1,
		Detail:     keycode,
		Time:       s.serverTime(),
		Root:       s.rootWindowID(),
		Event:      eventWindowID,
		Child:      0, // No child for now
		RootX:      s.pointerX,
		RootY:      s.pointerY,
		EventX:     s.pointerX,
		EventY:     s.pointerY,
		State:      state,
		SameScreen: true,
	}

	if eventType == "keydown" {
		event.Opcode = 2 // KeyPress
	} else if eventType == "keyup" {
		event.Opcode = 3 // KeyRelease
	} else {
		debugf("X11: Unknown keyboard event type: %s", eventType)
		return
	}

	if err := client.send(event); err != nil {
		debugf("X11: Failed to write keyboard event: %v", err)
	}
}

func (s *x11Server) sendXInputKeyboardEvent(client *x11Client, eventType string, keycode byte, eventWindowID uint32, state uint16) {
	var xiEvent messageEncoder
	switch eventType {
	case "keydown":
		xiEvent = &wire.DeviceKeyPressEvent{
			Sequence:   client.sequence - 1,
			DeviceID:   virtualKeyboard.Header.DeviceID,
			Time:       s.serverTime(),
			KeyCode:    keycode,
			Root:       s.rootWindowID(),
			Event:      eventWindowID,
			Child:      0, // Or a child window ID if applicable
			RootX:      s.pointerX,
			RootY:      s.pointerY,
			EventX:     s.pointerX,
			EventY:     s.pointerY,
			State:      state,
			SameScreen: true,
		}
	case "keyup":
		xiEvent = &wire.DeviceKeyReleaseEvent{
			Sequence:   client.sequence - 1,
			DeviceID:   virtualKeyboard.Header.DeviceID,
			Time:       s.serverTime(),
			KeyCode:    keycode,
			Root:       s.rootWindowID(),
			Event:      eventWindowID,
			Child:      0, // Or a child window ID if applicable
			RootX:      s.pointerX,
			RootY:      s.pointerY,
			EventX:     s.pointerX,
			EventY:     s.pointerY,
			State:      state,
			SameScreen: true,
		}
	}

	if xiEvent != nil {
		if err := client.send(xiEvent); err != nil {
			debugf("X11: Failed to write XInput keyboard event: %v", err)
		}
	}
}

func (s *x11Server) SendPointerCrossingEvent(isEnter bool, xid xID, rootX, rootY, eventX, eventY int16, state uint16, mode, detail byte) {
	client, ok := s.clients[xid.client]
	if !ok {
		log.Printf("X11: Failed to write pointer crossing event: client %d not found", xid.client)
		return
	}

	var event messageEncoder
	if isEnter {
		event = &wire.EnterNotifyEvent{
			Sequence:   client.sequence - 1,
			Detail:     detail,
			Time:       s.serverTime(),
			Root:       s.rootWindowID(),
			Event:      xid.local,
			Child:      0, // Or a child window ID if applicable
			RootX:      rootX,
			RootY:      rootY,
			EventX:     eventX,
			EventY:     eventY,
			State:      state,
			Mode:       mode,
			SameScreen: true,
			Focus:      s.frontend.GetFocusWindow(client.id) == xid,
		}
	} else {
		event = &wire.LeaveNotifyEvent{
			Sequence:   client.sequence - 1,
			Detail:     detail,
			Time:       s.serverTime(),
			Root:       s.rootWindowID(),
			Event:      xid.local,
			Child:      0, // Or a child window ID if applicable
			RootX:      rootX,
			RootY:      rootY,
			EventX:     eventX,
			EventY:     eventY,
			State:      state,
			Mode:       mode,
			SameScreen: true,
			Focus:      s.frontend.GetFocusWindow(client.id) == xid,
		}
	}

	if err := client.send(event); err != nil {
		debugf("X11: Failed to write pointer crossing event: %v", err)
	}
}

func (s *x11Server) sendConfigureNotifyEvent(windowID xID, x, y int16, width, height uint16) {
	debugf("X11: Sending ConfigureNotify event for window %d", windowID)
	client, ok := s.clients[windowID.client]
	if !ok {
		log.Print("X11: Failed to write ConfigureNotify event: client not found")
		return
	}

	event := &wire.ConfigureNotifyEvent{
		Sequence:         client.sequence - 1,
		Event:            windowID.local,
		Window:           windowID.local,
		AboveSibling:     0, // None
		X:                x,
		Y:                y,
		Width:            width,
		Height:           height,
		BorderWidth:      0,
		OverrideRedirect: false,
	}

	if err := client.send(event); err != nil {
		debugf("X11: Failed to write ConfigureNotify event: %v", err)
	}
}

func (s *x11Server) sendExposeEvent(windowID xID, x, y, width, height uint16) {
	debugf("X11: Sending Expose event for window %s", windowID)
	client, ok := s.clients[windowID.client]
	if !ok {
		debugf("X11: sendExposeEvent unknown client %d", windowID.client)
		return
	}

	event := &wire.ExposeEvent{
		Sequence: client.sequence - 1,
		Window:   windowID.local,
		X:        x,
		Y:        y,
		Width:    width,
		Height:   height,
		Count:    0, // count = 0, no more expose events to follow
	}

	if err := client.send(event); err != nil {
		debugf("X11: Failed to write Expose event: %v", err)
	}
}

func (s *x11Server) SendClientMessageEvent(windowID xID, messageTypeAtom uint32, data [20]byte) {
	debugf("X11: Sending ClientMessage event for window %s", windowID)
	client, ok := s.clients[windowID.client]
	if !ok {
		debugf("X11: SendClientMessageEvent unknown client %d", windowID.client)
		return
	}

	event := &wire.ClientMessageEvent{
		Sequence:    client.sequence - 1,
		Format:      32, // Format is always 32 for ClientMessage
		Window:      windowID.local,
		MessageType: messageTypeAtom,
		Data:        data,
	}

	if err := client.send(event); err != nil {
		debugf("X11: Failed to write ClientMessage event: %v", err)
	}
}

func (s *x11Server) SendSelectionNotify(requestor xID, selection, target, property uint32, data []byte) {
	client, ok := s.clients[requestor.client]
	if !ok {
		debugf("X11: SendSelectionNotify unknown client %d", requestor.client)
		return
	}

	event := &wire.SelectionNotifyEvent{
		Sequence:  client.sequence - 1,
		Requestor: requestor.local,
		Selection: selection,
		Target:    target,
		Property:  property,
		Time:      s.serverTime(),
	}
	s.sendEvent(client, event)
}

func (s *x11Server) sendEvent(client *x11Client, event messageEncoder) {
	if err := client.send(event); err != nil {
		s.logger.Errorf("Failed to write event: %v", err)
	}
}

func (s *x11Server) GetRGBColor(colormap xID, pixel uint32) (r, g, b uint8) {
	if colormap.local == s.defaultColormap {
		colormap.client = 0
	}
	if cm, ok := s.colormaps[colormap]; ok {
		if color, ok := cm.pixels[pixel]; ok {
			debugf("GetRGBColor: cmap:%s pixel:%x return %+v", colormap, pixel, color)
			return uint8(color.Red >> 8), uint8(color.Green >> 8), uint8(color.Blue >> 8)
		}
		r = uint8((pixel & 0xff0000) >> 16)
		g = uint8((pixel & 0x00ff00) >> 8)
		b = uint8((pixel & 0x0000ff))
		debugf("GetRGBColor: cmap:%s pixel:%x return RGB for pixel", colormap, pixel)
		return r, g, b
	}
	// Explicitly handle black and white pixels based on server's setup
	if pixel == s.blackPixel {
		debugf("GetRGBColor: cmap:%s pixel:%x return blackPixel", colormap, pixel)
		return 0, 0, 0 // Black
	}
	if pixel == s.whitePixel {
		debugf("GetRGBColor: cmap:%s pixel:%x return whitePixel", colormap, pixel)
		return 0xFF, 0xFF, 0xFF // White
	}
	// For TrueColor visuals, the pixel value directly encodes RGB components.
	if s.rootVisual.Class == 4 { // TrueColor
		r = uint8((pixel & s.rootVisual.RedMask) >> calculateShift(s.rootVisual.RedMask))
		g = uint8((pixel & s.rootVisual.GreenMask) >> calculateShift(s.rootVisual.GreenMask))
		b = uint8((pixel & s.rootVisual.BlueMask) >> calculateShift(s.rootVisual.BlueMask))
		debugf("GetRGBColor: cmap:%s pixel:%x return RGB for pixel", colormap, pixel)
		return r, g, b
	}
	// Default to black if not found
	debugf("GetRGBColor: cmap:%s pixel:%x return black", colormap, pixel)
	return 0, 0, 0
}

// calculateShift determines the right shift needed to extract the color component.
func calculateShift(mask uint32) uint32 {
	if mask == 0 {
		return 0
	}
	shift := uint32(0)
	for (mask & 1) == 0 {
		mask >>= 1
		shift++
	}
	return shift
}

func (s *x11Server) initAtoms() {
	s.atoms = map[string]uint32{
		"PRIMARY":             1,
		"SECONDARY":           2,
		"ARC":                 3,
		"ATOM":                4,
		"BITMAP":              5,
		"CARDINAL":            6,
		"COLORMAP":            7,
		"CURSOR":              8,
		"CUT_BUFFER0":         9,
		"CUT_BUFFER1":         10,
		"CUT_BUFFER2":         11,
		"CUT_BUFFER3":         12,
		"CUT_BUFFER4":         13,
		"CUT_BUFFER5":         14,
		"CUT_BUFFER6":         15,
		"CUT_BUFFER7":         16,
		"DRAWABLE":            17,
		"FONT":                18,
		"INTEGER":             19,
		"PIXMAP":              20,
		"POINT":               21,
		"RECTANGLE":           22,
		"RESOURCE_MANAGER":    23,
		"RGB_COLOR_MAP":       24,
		"RGB_BEST_MAP":        25,
		"RGB_BLUE_MAP":        26,
		"RGB_DEFAULT_MAP":     27,
		"RGB_GRAY_MAP":        28,
		"RGB_GREEN_MAP":       29,
		"RGB_RED_MAP":         30,
		"STRING":              31,
		"VISUALID":            32,
		"WINDOW":              33,
		"WM_COMMAND":          34,
		"WM_HINTS":            35,
		"WM_CLIENT_MACHINE":   36,
		"WM_ICON_NAME":        37,
		"WM_ICON_SIZE":        38,
		"WM_NAME":             39,
		"WM_NORMAL_HINTS":     40,
		"WM_SIZE_HINTS":       41,
		"WM_ZOOM_HINTS":       42,
		"MIN_SPACE":           43,
		"NORM_SPACE":          44,
		"MAX_SPACE":           45,
		"END_SPACE":           46,
		"SUPERSCRIPT_X":       47,
		"SUPERSCRIPT_Y":       48,
		"SUBSCRIPT_X":         49,
		"SUBSCRIPT_Y":         50,
		"UNDERLINE_POSITION":  51,
		"UNDERLINE_THICKNESS": 52,
		"STRIKEOUT_ASCENT":    53,
		"STRIKEOUT_DESCENT":   54,
		"ITALIC_ANGLE":        55,
		"X_HEIGHT":            56,
		"QUAD_WIDTH":          57,
		"WEIGHT":              58,
		"POINT_SIZE":          59,
		"RESOLUTION":          60,
		"COPYRIGHT":           61,
		"NOTICE":              62,
		"FONT_NAME":           63,
		"FAMILY_NAME":         64,
		"FULL_NAME":           65,
		"CAP_HEIGHT":          66,
		"WM_CLASS":            67,
		"WM_TRANSIENT_FOR":    68,
	}
	s.atomNames = make(map[uint32]string)
	for name, id := range s.atoms {
		s.atomNames[id] = name
	}
	s.nextAtomID = 69
}

func (s *x11Server) GetAtom(name string) uint32 {
	if id, ok := s.atoms[name]; ok {
		return id
	}
	id := s.nextAtomID
	s.nextAtomID++
	s.atoms[name] = id
	s.atomNames[id] = name
	return id
}

func (s *x11Server) GetAtomName(atom uint32) string {
	return s.atomNames[atom]
}

func (s *x11Server) ChangeProperty(xid xID, propertyID, typeAtom uint32, format byte, data []byte) {
	props, ok := s.properties[xid]
	if !ok {
		props = make(map[uint32]*property)
		s.properties[xid] = props
	}
	props[propertyID] = &property{
		data:     data,
		typeAtom: typeAtom,
		format:   format,
	}

	s.sendPropertyNotify(xid, propertyID, 0) // PropertyNewValue

	// Check for WM_NAME etc.
	name := s.GetAtomName(propertyID)
	if name == "WM_NAME" || name == "_NET_WM_NAME" || name == "WM_ICON_NAME" {
		s.frontend.SetWindowTitle(xid, string(data))
	}
}

func (s *x11Server) DeleteProperty(xid xID, propertyID uint32) {
	if props, ok := s.properties[xid]; ok {
		delete(props, propertyID)
		s.sendPropertyNotify(xid, propertyID, 1) // PropertyDelete
	}
}

func (s *x11Server) sendPropertyNotify(windowID xID, atom uint32, state byte) {
	if client, ok := s.clients[windowID.client]; ok {
		if w, ok := s.windows[windowID]; ok {
			if w.attributes.EventMask&wire.PropertyChangeMask != 0 {
				event := &wire.PropertyNotifyEvent{
					Sequence: client.sequence - 1,
					Window:   windowID.local,
					Atom:     atom,
					Time:     s.serverTime(),
					State:    state,
				}
				s.sendEvent(client, event)
			}
		}
	}
}

func (s *x11Server) GetProperty(xid xID, propertyID uint32) *property {
	if props, ok := s.properties[xid]; ok {
		return props[propertyID]
	}
	return nil
}

func (s *x11Server) ListProperties(xid xID) []uint32 {
	var list []uint32
	if props, ok := s.properties[xid]; ok {
		for id := range props {
			list = append(list, id)
		}
	}
	return list
}

func (s *x11Server) RotateProperties(xid xID, delta int16, atoms []wire.Atom) error {
	props, ok := s.properties[xid]
	if !ok {
		return nil
	}

	// Check existence
	for _, atom := range atoms {
		if _, ok := props[uint32(atom)]; !ok {
			return errors.New("property not found")
		}
	}

	n := len(atoms)
	if n == 0 {
		return nil
	}

	newValues := make(map[uint32]*property)
	for i, atom := range atoms {
		prop := props[uint32(atom)]
		newIdx := (i + int(delta)) % n
		if newIdx < 0 {
			newIdx += n
		}
		newValues[uint32(atoms[newIdx])] = prop
	}

	for atom, prop := range newValues {
		props[atom] = prop
	}
	return nil
}

func (s *x11Server) rootWindowID() uint32 {
	return 0
}

func (s *x11Server) readRequest(client *x11Client) (wire.Request, uint16, error) {
	client.sequence++
	var header [4]byte
	if _, err := io.ReadFull(client.conn, header[:]); err != nil {
		return nil, 0, err
	}

	length := uint32(client.byteOrder.Uint16(header[2:4]))
	var extendedHeader []byte
	if client.bigRequestsEnabled && length == 0 {
		var extendedLengthBytes [4]byte
		if _, err := io.ReadFull(client.conn, extendedLengthBytes[:]); err != nil {
			return nil, 0, err
		}
		length = client.byteOrder.Uint32(extendedLengthBytes[:])
		extendedHeader = extendedLengthBytes[:]
	}

	if length == 0 {
		client.send(wire.NewError(wire.LengthErrorCode, client.sequence, 0, wire.Opcodes{Major: wire.ReqCode(header[0]), Minor: 0}))
		return nil, 0, errParseError
	}

	totalSize := 4 * length
	raw := make([]byte, totalSize)
	copy(raw[0:4], header[:])
	if extendedHeader != nil {
		copy(raw[4:8], extendedHeader)
	}

	readOffset := 4 + len(extendedHeader)
	if totalSize > uint32(readOffset) {
		if _, err := io.ReadFull(client.conn, raw[readOffset:]); err != nil {
			return nil, 0, err
		}
	}

	debugf("X11DEBUG: RAW Request: %x", raw)
	req, err := wire.ParseRequest(client.byteOrder, raw, client.sequence, client.bigRequestsEnabled)
	if err != nil {
		if x11Err, ok := err.(wire.Error); ok {
			client.send(x11Err)
		} else {
			client.send(wire.NewError(wire.LengthErrorCode, client.sequence, 0, wire.Opcodes{Major: wire.ReqCode(header[0]), Minor: 0}))
		}
		return nil, 0, err
	}
	return req, client.sequence, nil
}

func (s *x11Server) cleanupClient(client *x11Client) {
	s.frontend.DestroyAllWindowsForClient(client.id)
	delete(s.clients, client.id)
}

func (s *x11Server) serve(client *x11Client) {
	defer func() {
		if r := recover(); r != nil {
			debugf("X11 Request Handler Panic: %v\n%s", r, debug.Stack())
		}
	}()
	defer client.conn.Close()
	defer s.cleanupClient(client)
	for {
		req, seq, err := s.readRequest(client)
		if err != nil {
			if err != io.EOF {
				s.logger.Errorf("Failed to read X11 request: %v", err)
				debugf("readRequest failed: %v", err)
			}
			break
		}
		reply := s.handleRequest(client, req, seq)
		if reply != nil {
			if err := client.send(reply); err != nil {
				s.logger.Errorf("Failed to write reply: %v", err)
			}
		}
	}
}

func (s *x11Server) handleRequest(client *x11Client, req wire.Request, seq uint16) (reply messageEncoder) {
	if s.serverGrabbed && s.grabbingClientID != client.id {
		// Ignore requests from other clients while the server is grabbed
		return nil
	}
	debugf("X11DEBUG: handleRequest(%d) opcode: %d: %#v", seq, req.OpCode(), req)
	if req.OpCode() == wire.XInputOpcode {
		return s.handleXInputRequest(client, req, seq)
	}

	switch p := req.(type) {
	case *wire.CreateWindowRequest:
		xid := client.xID(uint32(p.Drawable))
		parentXID := client.xID(uint32(p.Parent))
		// Check if the window ID is already in use
		if _, exists := s.windows[xid]; exists {
			s.logger.Errorf("X11: CreateWindow: ID %d already in use", xid)
			return wire.NewGenericError(seq, uint32(p.Drawable), 0, wire.CreateWindow, wire.IDChoiceErrorCode)
		}

		newWindow := &window{
			xid:        xid,
			parent:     uint32(p.Parent),
			x:          p.X,
			y:          p.Y,
			width:      p.Width,
			height:     p.Height,
			depth:      p.Depth,
			children:   []uint32{},
			attributes: p.Values,
		}
		if p.Values.Colormap > 0 {
			newWindow.colormap = client.xID(uint32(p.Values.Colormap))
		} else {
			newWindow.colormap = xID{local: s.defaultColormap}
		}
		s.windows[xid] = newWindow

		// Add to parent's children list
		if parentWindow, ok := s.windows[parentXID]; ok {
			parentWindow.children = append(parentWindow.children, uint32(p.Drawable))
		}
		s.frontend.CreateWindow(xid, uint32(p.Parent), uint32(p.X), uint32(p.Y), uint32(p.Width), uint32(p.Height), uint32(p.Depth), p.ValueMask, p.Values)

	case *wire.ChangeWindowAttributesRequest:
		xid := client.xID(uint32(p.Window))
		if err := s.checkWindow(xid, seq, wire.ChangeWindowAttributes, 0); err != nil {
			return err
		}
		if w, ok := s.windows[xid]; ok {
			if p.ValueMask&wire.CWBackPixmap != 0 {
				w.attributes.BackgroundPixmap = p.Values.BackgroundPixmap
			}
			if p.ValueMask&wire.CWBackPixel != 0 {
				w.attributes.BackgroundPixel = p.Values.BackgroundPixel
			}
			if p.ValueMask&wire.CWBorderPixmap != 0 {
				w.attributes.BorderPixmap = p.Values.BorderPixmap
			}
			if p.ValueMask&wire.CWBorderPixel != 0 {
				w.attributes.BorderPixel = p.Values.BorderPixel
			}
			if p.ValueMask&wire.CWBitGravity != 0 {
				w.attributes.BitGravity = p.Values.BitGravity
			}
			if p.ValueMask&wire.CWWinGravity != 0 {
				w.attributes.WinGravity = p.Values.WinGravity
			}
			if p.ValueMask&wire.CWBackingStore != 0 {
				w.attributes.BackingStore = p.Values.BackingStore
			}
			if p.ValueMask&wire.CWBackingPlanes != 0 {
				w.attributes.BackingPlanes = p.Values.BackingPlanes
			}
			if p.ValueMask&wire.CWBackingPixel != 0 {
				w.attributes.BackingPixel = p.Values.BackingPixel
			}
			if p.ValueMask&wire.CWOverrideRedirect != 0 {
				w.attributes.OverrideRedirect = p.Values.OverrideRedirect
			}
			if p.ValueMask&wire.CWSaveUnder != 0 {
				w.attributes.SaveUnder = p.Values.SaveUnder
			}
			if p.ValueMask&wire.CWEventMask != 0 {
				w.attributes.EventMask = p.Values.EventMask
			}
			if p.ValueMask&wire.CWDontPropagate != 0 {
				w.attributes.DontPropagateMask = p.Values.DontPropagateMask
			}
			if p.ValueMask&wire.CWColormap != 0 {
				w.attributes.Colormap = p.Values.Colormap
			}
			if p.ValueMask&wire.CWCursor != 0 {
				w.attributes.Cursor = p.Values.Cursor
				s.frontend.SetWindowCursor(xid, client.xID(uint32(p.Values.Cursor)))
			}
		}
		s.frontend.ChangeWindowAttributes(xid, p.ValueMask, p.Values)

	case *wire.GetWindowAttributesRequest:
		xid := client.xID(uint32(p.Window))
		if err := s.checkWindow(xid, seq, wire.GetWindowAttributes, 0); err != nil {
			return err
		}
		attrs, _ := s.GetWindowAttributes(xid)
		return &wire.GetWindowAttributesReply{
			Sequence:           seq,
			BackingStore:       byte(attrs.BackingStore),
			VisualID:           s.visualID,
			Class:              uint16(attrs.Class),
			BitGravity:         byte(attrs.BitGravity),
			WinGravity:         byte(attrs.WinGravity),
			BackingPlanes:      attrs.BackingPlanes,
			BackingPixel:       attrs.BackingPixel,
			SaveUnder:          wire.BoolToByte(attrs.SaveUnder),
			MapIsInstalled:     wire.BoolToByte(attrs.MapIsInstalled),
			MapState:           byte(attrs.MapState),
			OverrideRedirect:   wire.BoolToByte(attrs.OverrideRedirect),
			Colormap:           uint32(attrs.Colormap),
			AllEventMasks:      attrs.EventMask,
			YourEventMask:      attrs.EventMask,
			DoNotPropagateMask: uint16(attrs.DontPropagateMask),
		}

	case *wire.DestroyWindowRequest:
		xid := client.xID(uint32(p.Window))
		if xid.local == s.rootWindowID() {
			return nil
		}
		if err := s.checkWindow(xid, seq, wire.DestroyWindow, 0); err != nil {
			return err
		}
		delete(s.windows, xid)
		s.frontend.DestroyWindow(xid)

	case *wire.DestroySubwindowsRequest:
		xid := client.xID(uint32(p.Window))
		if err := s.checkWindow(xid, seq, wire.DestroySubwindows, 0); err != nil {
			return err
		}
		if parent, ok := s.windows[xid]; ok {
			var destroy func(uint32)
			destroy = func(windowID uint32) {
				childXID := client.xID(windowID)
				if w, ok := s.windows[childXID]; ok {
					for _, child := range w.children {
						destroy(child)
					}
					delete(s.windows, childXID)
				}
			}
			for _, child := range parent.children {
				destroy(child)
			}
			parent.children = []uint32{}
		}
		s.frontend.DestroySubwindows(xid)

	case *wire.ChangeSaveSetRequest:
		if err := s.checkWindow(client.xID(uint32(p.Window)), seq, wire.ChangeSaveSet, 0); err != nil {
			return err
		}
		if p.Mode == 0 { // Insert
			client.saveSet[uint32(p.Window)] = true
		} else { // Delete
			delete(client.saveSet, uint32(p.Window))
		}

	case *wire.ReparentWindowRequest:
		windowXID := client.xID(uint32(p.Window))
		parentXID := client.xID(uint32(p.Parent))
		if err := s.checkWindow(windowXID, seq, wire.ReparentWindow, 0); err != nil {
			return err
		}
		if err := s.checkWindow(parentXID, seq, wire.ReparentWindow, 0); err != nil {
			return err
		}
		window := s.windows[windowXID]
		if window == nil {
			return wire.NewGenericError(seq, uint32(p.Window), 0, wire.ReparentWindow, wire.MatchErrorCode)
		}

		oldParent, ok := s.windows[client.xID(window.parent)]
		if !ok && window.parent != s.rootWindowID() {
			return wire.NewGenericError(seq, window.parent, 0, wire.ReparentWindow, wire.WindowErrorCode)
		}
		newParent := s.windows[parentXID]

		// Remove from old parent's children
		if ok {
			for i, childID := range oldParent.children {
				if childID == window.xid.local {
					oldParent.children = append(oldParent.children[:i], oldParent.children[i+1:]...)
					break
				}
			}
		}

		// Add to new parent's children
		if newParent != nil {
			newParent.children = append(newParent.children, window.xid.local)
		}

		// Update window's state
		window.parent = uint32(p.Parent)
		window.x = p.X
		window.y = p.Y

		s.frontend.ReparentWindow(windowXID, parentXID, p.X, p.Y)

	case *wire.MapWindowRequest:
		xid := client.xID(uint32(p.Window))
		if err := s.checkWindow(xid, seq, wire.MapWindow, 0); err != nil {
			return err
		}
		if w, ok := s.windows[xid]; ok {
			w.mapped = true
			s.frontend.MapWindow(xid)
			s.sendExposeEvent(xid, 0, 0, w.width, w.height)
		}

	case *wire.MapSubwindowsRequest:
		xid := client.xID(uint32(p.Window))
		if err := s.checkWindow(xid, seq, wire.MapSubwindows, 0); err != nil {
			return err
		}
		if parentWindow, ok := s.windows[xid]; ok {
			for _, childID := range parentWindow.children {
				childXID := xID{client: xid.client, local: childID}
				if childWindow, ok := s.windows[childXID]; ok {
					childWindow.mapped = true
					s.frontend.MapWindow(childXID)
					s.sendExposeEvent(childXID, 0, 0, childWindow.width, childWindow.height)
				}
			}
		}

	case *wire.UnmapWindowRequest:
		xid := client.xID(uint32(p.Window))
		if err := s.checkWindow(xid, seq, wire.UnmapWindow, 0); err != nil {
			return err
		}
		if w, ok := s.windows[xid]; ok {
			w.mapped = false
		}
		s.frontend.UnmapWindow(xid)

	case *wire.UnmapSubwindowsRequest:
		xid := client.xID(uint32(p.Window))
		if err := s.checkWindow(xid, seq, wire.UnmapSubwindows, 0); err != nil {
			return err
		}
		if parentWindow, ok := s.windows[xid]; ok {
			for _, childID := range parentWindow.children {
				childXID := xID{client: xid.client, local: childID}
				if childWindow, ok := s.windows[childXID]; ok {
					childWindow.mapped = false
					s.frontend.UnmapWindow(childXID)
				}
			}
		}

	case *wire.ConfigureWindowRequest:
		xid := client.xID(uint32(p.Window))
		if err := s.checkWindow(xid, seq, wire.ConfigureWindow, 0); err != nil {
			return err
		}
		if xid.local == s.rootWindowID() {
			return wire.NewGenericError(seq, uint32(p.Window), 0, wire.ConfigureWindow, wire.MatchErrorCode)
		}
		s.frontend.ConfigureWindow(xid, p.ValueMask, p.Values)

	case *wire.CirculateWindowRequest:
		xid := client.xID(uint32(p.Window))
		if err := s.checkWindow(xid, seq, wire.CirculateWindow, 0); err != nil {
			return err
		}
		window := s.windows[xid]
		if window == nil {
			return wire.NewGenericError(seq, uint32(p.Window), 0, wire.CirculateWindow, wire.MatchErrorCode)
		}
		parent, ok := s.windows[client.xID(window.parent)]
		if ok {
			// Find index of window in parent's children
			idx := -1
			for i, childID := range parent.children {
				if childID == xid.local {
					idx = i
					break
				}
			}

			if idx != -1 {
				// Remove window from children slice
				children := append(parent.children[:idx], parent.children[idx+1:]...)

				if p.Direction == 0 { // RaiseLowest
					// Add to end of slice
					parent.children = append(children, xid.local)
				} else { // LowerHighest
					// Add to beginning of slice
					parent.children = append([]uint32{xid.local}, children...)
				}
			}
		} else if window.parent != s.rootWindowID() {
			return wire.NewGenericError(seq, window.parent, 0, wire.CirculateWindow, wire.WindowErrorCode)
		}

		s.frontend.CirculateWindow(xid, p.Direction)

	case *wire.GetGeometryRequest:
		xid := client.xID(uint32(p.Drawable))
		if err := s.checkDrawable(xid, seq, wire.GetGeometry, 0); err != nil {
			return err
		}
		if xid.local == s.rootWindowID() {
			return &wire.GetGeometryReply{
				Sequence:    seq,
				Depth:       24, // TODO: Get this from rootVisual or screen info
				Root:        s.rootWindowID(),
				X:           0,
				Y:           0,
				Width:       s.rootWindowWidth,
				Height:      s.rootWindowHeight,
				BorderWidth: 0,
			}
		}
		if w, ok := s.windows[xid]; ok {
			return &wire.GetGeometryReply{
				Sequence:    seq,
				Depth:       w.depth,
				Root:        s.rootWindowID(),
				X:           w.x,
				Y:           w.y,
				Width:       w.width,
				Height:      w.height,
				BorderWidth: 0, // Border width is not stored in window struct, assuming 0 for now
			}
		}
		// Must be a pixmap
		return &wire.GetGeometryReply{
			Sequence:    seq,
			Depth:       24, // Assumption
			Root:        s.rootWindowID(),
			X:           0,
			Y:           0,
			Width:       1, // Dummy
			Height:      1, // Dummy
			BorderWidth: 0,
		}

	case *wire.QueryTreeRequest:
		xid := client.xID(uint32(p.Window))
		if err := s.checkWindow(xid, seq, wire.QueryTree, 0); err != nil {
			return err
		}
		window := s.windows[xid]
		if window == nil {
			var children []uint32
			for _, w := range s.windows {
				if w.parent == s.rootWindowID() {
					children = append(children, w.xid.local)
				}
			}
			return &wire.QueryTreeReply{
				Sequence:    seq,
				Root:        s.rootWindowID(),
				Parent:      s.rootWindowID(),
				NumChildren: uint16(len(children)),
				Children:    children,
			}
		}
		return &wire.QueryTreeReply{
			Sequence:    seq,
			Root:        s.rootWindowID(),
			Parent:      window.parent,
			NumChildren: uint16(len(window.children)),
			Children:    window.children,
		}

	case *wire.InternAtomRequest:
		atomID := s.GetAtom(p.Name)

		return &wire.InternAtomReply{
			Sequence: seq,
			Atom:     atomID,
		}

	case *wire.GetAtomNameRequest:
		name := s.GetAtomName(uint32(p.Atom))
		return &wire.GetAtomNameReply{
			Sequence:   seq,
			NameLength: uint16(len(name)),
			Name:       name,
		}

	case *wire.ChangePropertyRequest:
		xid := client.xID(uint32(p.Window))
		if err := s.checkWindow(xid, seq, wire.ChangeProperty, 0); err != nil {
			return err
		}
		s.ChangeProperty(xid, uint32(p.Property), uint32(p.Type), byte(p.Format), p.Data)

	case *wire.DeletePropertyRequest:
		xid := client.xID(uint32(p.Window))
		if err := s.checkWindow(xid, seq, wire.DeleteProperty, 0); err != nil {
			return err
		}
		s.DeleteProperty(xid, uint32(p.Property))

	case *wire.GetPropertyRequest:
		xid := client.xID(uint32(p.Window))
		if err := s.checkWindow(xid, seq, wire.GetProperty, 0); err != nil {
			return err
		}
		prop := s.GetProperty(xid, uint32(p.Property))

		if prop == nil {
			return &wire.GetPropertyReply{
				Sequence: seq,
				Format:   0,
			}
		}

		// Calculate slice
		byteOffset := p.Offset * 4
		byteLength := p.Length * 4
		totalLen := uint32(len(prop.data))

		if byteOffset >= totalLen {
			return &wire.GetPropertyReply{
				Sequence:              seq,
				Format:                prop.format,
				PropertyType:          prop.typeAtom,
				BytesAfter:            0,
				ValueLenInFormatUnits: 0,
				Value:                 nil,
			}
		}

		end := byteOffset + byteLength
		if end > totalLen {
			end = totalLen
		}
		dataToSend := prop.data[byteOffset:end]
		bytesAfter := totalLen - end

		var valueLenInFormatUnits uint32
		if prop.format == 8 {
			valueLenInFormatUnits = uint32(len(dataToSend))
		} else if prop.format == 16 {
			valueLenInFormatUnits = uint32(len(dataToSend) / 2)
		} else if prop.format == 32 {
			valueLenInFormatUnits = uint32(len(dataToSend) / 4)
		}

		if p.Delete && bytesAfter == 0 && (p.Type == 0 || prop.typeAtom == uint32(p.Type)) {
			s.DeleteProperty(xid, uint32(p.Property))
		}

		if p.Type != 0 && prop.typeAtom != uint32(p.Type) {
			return &wire.GetPropertyReply{
				Sequence:              seq,
				Format:                prop.format,
				PropertyType:          prop.typeAtom,
				BytesAfter:            totalLen, // Full length
				ValueLenInFormatUnits: 0,
				Value:                 nil,
			}
		}

		return &wire.GetPropertyReply{
			Sequence:              seq,
			Format:                prop.format,
			PropertyType:          prop.typeAtom,
			BytesAfter:            bytesAfter,
			ValueLenInFormatUnits: valueLenInFormatUnits,
			Value:                 dataToSend,
		}

	case *wire.ListPropertiesRequest:
		xid := client.xID(uint32(p.Window))
		if err := s.checkWindow(xid, seq, wire.ListProperties, 0); err != nil {
			return err
		}
		atoms := s.ListProperties(xid)
		return &wire.ListPropertiesReply{
			Sequence:      seq,
			NumProperties: uint16(len(atoms)),
			Atoms:         atoms,
		}

	case *wire.SetSelectionOwnerRequest:
		selectionAtom := uint32(p.Selection)
		ownerWindow := uint32(p.Owner)
		time := uint32(p.Time)

		if time == 0 { // CurrentTime
			time = s.serverTime()
		}

		// "If the timestamp is not CurrentTime and is less than the timestamp of the last successful SetSelectionOwner request for the selection, the request is ignored."
		currentOwner, ok := s.selections[selectionAtom]
		if ok && time < currentOwner.time && p.Time != 0 {
			return nil
		}

		if ownerWindow == 0 { // None
			delete(s.selections, selectionAtom)
		} else {
			if err := s.checkWindow(client.xID(ownerWindow), seq, wire.SetSelectionOwner, 0); err != nil {
				return err
			}
			s.selections[selectionAtom] = &selectionOwner{
				window: client.xID(ownerWindow),
				time:   time,
			}
		}

		if ok && currentOwner.window.local != 0 && (currentOwner.window != client.xID(ownerWindow)) {
			// Send SelectionClear to old owner
			if oldClient, ok := s.clients[currentOwner.window.client]; ok {
				event := &wire.SelectionClearEvent{
					Sequence:  oldClient.sequence - 1, // Approximate
					Time:      time,
					Owner:     currentOwner.window.local,
					Selection: selectionAtom,
				}
				s.sendEvent(oldClient, event)
			}
		}

	case *wire.GetSelectionOwnerRequest:
		selectionAtom := uint32(p.Selection)
		var owner uint32
		if o, ok := s.selections[selectionAtom]; ok {
			owner = o.window.local
		}
		return &wire.GetSelectionOwnerReply{
			Sequence: seq,
			Owner:    owner,
		}

	case *wire.ConvertSelectionRequest:
		selectionAtom := uint32(p.Selection)
		targetAtom := uint32(p.Target)
		propertyAtom := uint32(p.Property)
		requestor := client.xID(uint32(p.Requestor))
		time := uint32(p.Time)
		if time == 0 {
			time = s.serverTime()
		}

		if err := s.checkWindow(requestor, seq, wire.ConvertSelection, 0); err != nil {
			return err
		}

		owner, ok := s.selections[selectionAtom]

		// Special handling for CLIPBOARD if no owner
		clipboardAtom := s.GetAtom("CLIPBOARD")
		if !ok && selectionAtom == clipboardAtom {
			go func() {
				content, err := s.frontend.ReadClipboard()
				if err == nil {
					// Write to property
					// Assuming target is STRING or UTF8_STRING or TEXT
					// Ideally check targetAtom. For now, assume string.
					s.ChangeProperty(requestor, propertyAtom, s.GetAtom("STRING"), 8, []byte(content))

					s.SendSelectionNotify(requestor, selectionAtom, targetAtom, propertyAtom, nil)
				} else {
					// Failed
					s.SendSelectionNotify(requestor, selectionAtom, targetAtom, 0, nil)
				}
			}()
			return nil
		}

		if ok {
			// Send SelectionRequest to owner
			if ownerClient, ok := s.clients[owner.window.client]; ok {
				event := &wire.SelectionRequestEvent{
					Sequence:  ownerClient.sequence - 1,
					Time:      time,
					Owner:     owner.window.local,
					Requestor: requestor.local,
					Selection: selectionAtom,
					Target:    targetAtom,
					Property:  propertyAtom,
				}
				s.sendEvent(ownerClient, event)
			}
			return nil
		}

		// No owner, send SelectionNotify with property None
		s.SendSelectionNotify(requestor, selectionAtom, targetAtom, 0, nil)

	case *wire.SendEventRequest:
		// The X11 client sends an event to another client.
		// We need to forward this event to the appropriate frontend.
		// For now, we'll just log it and pass it to the frontend.
		s.frontend.SendEvent(&wire.X11RawEvent{Data: p.EventData})

	case *wire.GrabPointerRequest:
		grabWindow := client.xID(uint32(p.GrabWindow))
		if err := s.checkWindow(grabWindow, seq, wire.GrabPointer, 0); err != nil {
			return err
		}
		if p.ConfineTo != 0 {
			if err := s.checkWindow(client.xID(uint32(p.ConfineTo)), seq, wire.GrabPointer, 0); err != nil {
				return err
			}
		}
		if p.Cursor != 0 {
			if err := s.checkCursor(client.xID(uint32(p.Cursor)), seq, wire.GrabPointer, 0); err != nil {
				return err
			}
		}

		if p.Time != 0 {
			if uint32(p.Time) < s.pointerGrabTime || uint32(p.Time) > s.serverTime() {
				return &wire.GrabPointerReply{Sequence: seq, Status: wire.GrabInvalidTime}
			}
		}

		if s.pointerGrabWindow.local != 0 && s.pointerGrabWindow.client != client.id {
			return &wire.GrabPointerReply{
				Sequence: seq,
				Status:   wire.AlreadyGrabbed,
			}
		}

		s.pointerGrabWindow = grabWindow
		s.pointerGrabWindow.client = client.id
		s.pointerGrabOwner = p.OwnerEvents
		s.pointerGrabEventMask = p.EventMask
		s.pointerGrabTime = uint32(p.Time)
		if s.pointerGrabTime == 0 {
			s.pointerGrabTime = s.serverTime()
		}
		s.pointerGrabMode = p.PointerMode
		s.keyboardGrabMode = p.KeyboardMode
		s.pointerGrabConfineTo = client.xID(uint32(p.ConfineTo))
		s.pointerGrabCursor = client.xID(uint32(p.Cursor))

		if p.Cursor != 0 {
			s.frontend.SetWindowCursor(grabWindow, s.pointerGrabCursor)
		}

		return &wire.GrabPointerReply{
			Sequence: seq,
			Status:   wire.GrabSuccess,
		}

	case *wire.UngrabPointerRequest:
		if p.Time != 0 && uint32(p.Time) < s.pointerGrabTime {
			// Ignore
			return nil
		}
		s.pointerGrabWindow = xID{}
		s.pointerGrabOwner = false
		s.pointerGrabEventMask = 0
		s.pointerGrabTime = 0
		s.pointerGrabMode = 0
		s.keyboardGrabMode = 0
		s.pointerGrabConfineTo = xID{}
		s.pointerGrabCursor = xID{}

	case *wire.GrabButtonRequest:
		var grabWindow xID
		var found bool
		for wID := range s.windows {
			if wID.local == uint32(p.GrabWindow) {
				grabWindow = wID
				found = true
				break
			}
		}
		if !found {
			return wire.NewGenericError(seq, uint32(p.GrabWindow), 0, wire.GrabButton, wire.WindowErrorCode)
		}

		grab := &passiveGrab{
			clientID:     client.id,
			button:       p.Button,
			modifiers:    p.Modifiers,
			owner:        p.OwnerEvents,
			eventMask:    p.EventMask,
			cursor:       client.xID(uint32(p.Cursor)),
			pointerMode:  p.PointerMode,
			keyboardMode: p.KeyboardMode,
			confineTo:    client.xID(uint32(p.ConfineTo)),
		}
		s.passiveGrabs[grabWindow] = append(s.passiveGrabs[grabWindow], grab)

	case *wire.UngrabButtonRequest:
		grabWindow := client.xID(uint32(p.GrabWindow))
		if err := s.checkWindow(grabWindow, seq, wire.UngrabButton, 0); err != nil {
			return err
		}
		if grabs, ok := s.passiveGrabs[grabWindow]; ok {
			for i, grab := range grabs {
				if grab.button == p.Button && grab.modifiers == p.Modifiers {
					s.passiveGrabs[grabWindow] = append(grabs[:i], grabs[i+1:]...)
					break
				}
			}
		}

	case *wire.ChangeActivePointerGrabRequest:
		if s.pointerGrabWindow.client == client.id && s.pointerGrabWindow.local != 0 {
			if p.Cursor != 0 {
				cursorXID := client.xID(uint32(p.Cursor))
				if err := s.checkCursor(cursorXID, seq, wire.ChangeActivePointerGrab, 0); err != nil {
					return err
				}
				s.frontend.SetWindowCursor(s.pointerGrabWindow, cursorXID)
			}
			s.pointerGrabEventMask = p.EventMask
			if p.Time == 0 || uint32(p.Time) >= s.pointerGrabTime {
				s.pointerGrabTime = uint32(p.Time)
			}
		}

	case *wire.GrabKeyboardRequest:
		grabWindow := client.xID(uint32(p.GrabWindow))
		if err := s.checkWindow(grabWindow, seq, wire.GrabKeyboard, 0); err != nil {
			return err
		}
		if p.Time != 0 {
			if uint32(p.Time) < s.keyboardGrabTime || uint32(p.Time) > s.serverTime() {
				return &wire.GrabKeyboardReply{Sequence: seq, Status: wire.GrabInvalidTime}
			}
		}

		if s.keyboardGrabWindow.local != 0 && s.keyboardGrabWindow.client != client.id {
			return &wire.GrabKeyboardReply{
				Sequence: seq,
				Status:   wire.AlreadyGrabbed,
			}
		}

		s.keyboardGrabWindow = grabWindow
		s.keyboardGrabWindow.client = client.id
		s.keyboardGrabOwner = p.OwnerEvents
		s.keyboardGrabTime = uint32(p.Time)
		if s.keyboardGrabTime == 0 {
			s.keyboardGrabTime = s.serverTime()
		}
		s.keyboardGrabMode = p.KeyboardMode
		s.pointerGrabMode = p.PointerMode

		return &wire.GrabKeyboardReply{
			Sequence: seq,
			Status:   wire.GrabSuccess,
		}

	case *wire.UngrabKeyboardRequest:
		if p.Time != 0 && uint32(p.Time) < s.keyboardGrabTime {
			// Ignore
			return nil
		}
		s.keyboardGrabWindow = xID{}
		s.keyboardGrabOwner = false
		s.keyboardGrabTime = 0
		s.keyboardGrabMode = 0

	case *wire.GrabKeyRequest:
		var grabWindow xID
		var found bool
		for wID := range s.windows {
			if wID.local == uint32(p.GrabWindow) {
				grabWindow = wID
				found = true
				break
			}
		}
		if !found {
			return wire.NewGenericError(seq, uint32(p.GrabWindow), 0, wire.GrabKey, wire.WindowErrorCode)
		}
		grab := &passiveGrab{
			clientID:     client.id,
			key:          p.Key,
			modifiers:    p.Modifiers,
			owner:        p.OwnerEvents,
			pointerMode:  p.PointerMode,
			keyboardMode: p.KeyboardMode,
		}
		s.passiveGrabs[grabWindow] = append(s.passiveGrabs[grabWindow], grab)

	case *wire.UngrabKeyRequest:
		grabWindow := client.xID(uint32(p.GrabWindow))
		if err := s.checkWindow(grabWindow, seq, wire.UngrabKey, 0); err != nil {
			return err
		}
		if grabs, ok := s.passiveGrabs[grabWindow]; ok {
			newGrabs := make([]*passiveGrab, 0, len(grabs))
			for _, grab := range grabs {
				if !(grab.key == p.Key && (p.Modifiers == wire.AnyModifier || grab.modifiers == p.Modifiers)) {
					newGrabs = append(newGrabs, grab)
				}
			}
			s.passiveGrabs[grabWindow] = newGrabs
		}

	case *wire.AllowEventsRequest:
		s.frontend.AllowEvents(client.id, p.Mode, uint32(p.Time))

	case *wire.GrabServerRequest:
		if !s.serverGrabbed {
			s.serverGrabbed = true
			s.grabbingClientID = client.id
		}

	case *wire.UngrabServerRequest:
		s.serverGrabbed = false
		s.grabbingClientID = 0

	case *wire.QueryPointerRequest:
		xid := client.xID(uint32(p.Drawable))
		if err := s.checkDrawable(xid, seq, wire.QueryPointer, 0); err != nil {
			return err
		}
		debugf("X11: QueryPointer drawable=%d", xid)
		return &wire.QueryPointerReply{
			Sequence:   seq,
			SameScreen: true,
			Root:       s.rootWindowID(),
			Child:      uint32(p.Drawable),
			RootX:      s.pointerX,
			RootY:      s.pointerY,
			WinX:       s.pointerX, // Assuming pointer is always in the window for now
			WinY:       s.pointerY, // Assuming pointer is always in the window for now
			Mask:       0,          // No buttons pressed
		}

	case *wire.GetMotionEventsRequest:
		return &wire.GetMotionEventsReply{
			Sequence: seq,
			Events:   []wire.TimeCoord{},
		}

	case *wire.TranslateCoordsRequest:
		srcWindow := client.xID(uint32(p.SrcWindow))
		dstWindow := client.xID(uint32(p.DstWindow))

		if err := s.checkWindow(srcWindow, seq, wire.TranslateCoords, 0); err != nil {
			return err
		}
		if err := s.checkWindow(dstWindow, seq, wire.TranslateCoords, 0); err != nil {
			return err
		}

		// Simplified implementation: assume windows are direct children of the root
		src := s.windows[srcWindow]
		dst := s.windows[dstWindow]

		var srcAbsX, srcAbsY int16
		if src != nil {
			srcAbsX = src.x
			srcAbsY = src.y
		}
		var dstAbsX, dstAbsY int16
		if dst != nil {
			dstAbsX = dst.x
			dstAbsY = dst.y
		}

		dstX := srcAbsX + p.SrcX - dstAbsX
		dstY := srcAbsY + p.SrcY - dstAbsY

		return &wire.TranslateCoordsReply{
			Sequence:   seq,
			SameScreen: true,
			Child:      0, // No child for now
			DstX:       dstX,
			DstY:       dstY,
		}

	case *wire.WarpPointerRequest:
		if p.SrcWindow != 0 {
			if err := s.checkWindow(client.xID(p.SrcWindow), seq, wire.WarpPointer, 0); err != nil {
				return err
			}
		}
		if p.DstWindow != 0 {
			if err := s.checkWindow(client.xID(p.DstWindow), seq, wire.WarpPointer, 0); err != nil {
				return err
			}
		}
		s.frontend.WarpPointer(p.DstX, p.DstY)

	case *wire.SetInputFocusRequest:
		xid := client.xID(uint32(p.Focus))
		// Focus can be None(0) or PointerRoot(1).
		if uint32(p.Focus) > 1 {
			if err := s.checkWindow(xid, seq, wire.SetInputFocus, 0); err != nil {
				return err
			}
		}
		s.inputFocus = xid
		s.frontend.SetInputFocus(xid, p.RevertTo)

	case *wire.GetInputFocusRequest:
		return &wire.GetInputFocusReply{
			Sequence: seq,
			RevertTo: 1, // RevertToParent
			Focus:    s.frontend.GetFocusWindow(client.id).local,
		}

	case *wire.QueryKeymapRequest:
		return &wire.QueryKeymapReply{
			Sequence: seq,
			Keys:     [32]byte{},
		}

	case *wire.OpenFontRequest:
		fid := client.xID(uint32(p.Fid))
		if _, exists := s.fonts[fid]; exists {
			return wire.NewGenericError(seq, uint32(p.Fid), 0, wire.OpenFont, wire.IDChoiceErrorCode)
		}
		s.fonts[fid] = true
		s.frontend.OpenFont(fid, p.Name)

	case *wire.CloseFontRequest:
		fid := client.xID(uint32(p.Fid))
		if err := s.checkFont(fid, seq, wire.CloseFont, 0); err != nil {
			return err
		}
		delete(s.fonts, fid)
		s.frontend.CloseFont(fid)

	case *wire.QueryFontRequest:
		fid := client.xID(uint32(p.Fid))
		if err := s.checkFont(fid, seq, wire.QueryFont, 0); err != nil {
			return err
		}
		minBounds, maxBounds, minCharOrByte2, maxCharOrByte2, defaultChar, drawDirection, minByte1, maxByte1, allCharsExist, fontAscent, fontDescent, charInfos, fontProps := s.frontend.QueryFont(fid)

		return &wire.QueryFontReply{
			Sequence:       seq,
			MinBounds:      minBounds,
			MaxBounds:      maxBounds,
			MinCharOrByte2: minCharOrByte2,
			MaxCharOrByte2: maxCharOrByte2,
			DefaultChar:    defaultChar,
			NumFontProps:   uint16(len(fontProps)),
			DrawDirection:  drawDirection,
			MinByte1:       minByte1,
			MaxByte1:       maxByte1,
			AllCharsExist:  allCharsExist,
			FontAscent:     fontAscent,
			FontDescent:    fontDescent,
			NumCharInfos:   uint32(len(charInfos)),
			CharInfos:      charInfos,
			FontProps:      fontProps,
		}

	case *wire.QueryTextExtentsRequest:
		fid := client.xID(uint32(p.Fid))
		if err := s.checkFont(fid, seq, wire.QueryTextExtents, 0); err != nil {
			return err
		}
		drawDirection, fontAscent, fontDescent, overallAscent, overallDescent, overallWidth, overallLeft, overallRight := s.frontend.QueryTextExtents(fid, p.Text)
		return &wire.QueryTextExtentsReply{
			Sequence:       seq,
			DrawDirection:  drawDirection,
			FontAscent:     fontAscent,
			FontDescent:    fontDescent,
			OverallAscent:  overallAscent,
			OverallDescent: overallDescent,
			OverallWidth:   int32(overallWidth),
			OverallLeft:    int32(overallLeft),
			OverallRight:   int32(overallRight),
		}

	case *wire.ListFontsRequest:
		fontNames := s.frontend.ListFonts(p.MaxNames, p.Pattern)

		return &wire.ListFontsReply{
			Sequence:  seq,
			FontNames: fontNames,
		}

	case *wire.ListFontsWithInfoRequest:
		fontNames := s.frontend.ListFonts(p.MaxNames, p.Pattern)
		tempFID := xID{client: 0, local: 0xFFFFFFFF}

		for _, name := range fontNames {
			s.frontend.OpenFont(tempFID, name)
			minBounds, maxBounds, minCharOrByte2, maxCharOrByte2, defaultChar, drawDirection, minByte1, maxByte1, allCharsExist, fontAscent, fontDescent, _, fontProps := s.frontend.QueryFont(tempFID)
			s.frontend.CloseFont(tempFID)

			reply := &wire.ListFontsWithInfoReply{
				Sequence:      seq,
				NameLength:    byte(len(name)),
				MinBounds:     minBounds,
				MaxBounds:     maxBounds,
				MinChar:       minCharOrByte2,
				MaxChar:       maxCharOrByte2,
				DefaultChar:   defaultChar,
				NFontProps:    uint16(len(fontProps)),
				DrawDirection: drawDirection,
				MinByte1:      minByte1,
				MaxByte1:      maxByte1,
				AllCharsExist: allCharsExist,
				FontAscent:    fontAscent,
				FontDescent:   fontDescent,
				NReplies:      1 + uint32(len(fontNames)-1), // This field is not actually used by clients usually, but let's be approximate
				FontProps:     fontProps,
				FontName:      name,
			}
			client.send(reply)
		}

		// Final reply
		lastReply := &wire.ListFontsWithInfoReply{
			Sequence: seq,
			FontName: "",
		}
		client.send(lastReply)
		return nil

	case *wire.SetFontPathRequest:
		s.fontPath = p.Paths

	case *wire.GetFontPathRequest:
		return &wire.GetFontPathReply{
			Sequence: seq,
			Paths:    s.fontPath,
		}

	case *wire.CreatePixmapRequest:
		xid := client.xID(uint32(p.Pid))

		// Check if the pixmap ID is already in use
		if _, exists := s.pixmaps[xid]; exists {
			s.logger.Errorf("X11: CreatePixmap: ID %s already in use", xid)
			return wire.NewGenericError(seq, uint32(p.Pid), 0, wire.CreatePixmap, wire.IDChoiceErrorCode)
		}
		if err := s.checkDrawable(client.xID(uint32(p.Drawable)), seq, wire.CreatePixmap, 0); err != nil {
			return err
		}

		s.pixmaps[xid] = true // Mark pixmap ID as used
		s.frontend.CreatePixmap(xid, client.xID(uint32(p.Drawable)), uint32(p.Width), uint32(p.Height), uint32(p.Depth))

	case *wire.FreePixmapRequest:
		xid := client.xID(uint32(p.Pid))
		if err := s.checkPixmap(xid, seq, wire.FreePixmap, 0); err != nil {
			return err
		}
		delete(s.pixmaps, xid)
		s.frontend.FreePixmap(xid)

	case *wire.CreateGCRequest:
		xid := client.xID(uint32(p.Cid))

		// Check if the GC ID is already in use
		if _, exists := s.gcs[xid]; exists {
			s.logger.Errorf("X11: CreateGC: ID %s already in use", xid)
			return wire.NewGenericError(seq, uint32(xid.local), 0, wire.CreateGC, wire.IDChoiceErrorCode)
		}
		if err := s.checkDrawable(client.xID(uint32(p.Drawable)), seq, wire.CreateGC, 0); err != nil {
			return err
		}

		s.gcs[xid] = p.Values
		s.frontend.CreateGC(xid, p.ValueMask, p.Values)

	case *wire.ChangeGCRequest:
		xid := client.xID(uint32(p.Gc))
		if err := s.checkGC(xid, seq, wire.ChangeGC, 0); err != nil {
			return err
		}
		if existingGC, ok := s.gcs[xid]; ok {
			if p.ValueMask&wire.GCFunction != 0 {
				existingGC.Function = p.Values.Function
			}
			if p.ValueMask&wire.GCPlaneMask != 0 {
				existingGC.PlaneMask = p.Values.PlaneMask
			}
			if p.ValueMask&wire.GCForeground != 0 {
				existingGC.Foreground = p.Values.Foreground
			}
			if p.ValueMask&wire.GCBackground != 0 {
				existingGC.Background = p.Values.Background
			}
			if p.ValueMask&wire.GCLineWidth != 0 {
				existingGC.LineWidth = p.Values.LineWidth
			}
			if p.ValueMask&wire.GCLineStyle != 0 {
				existingGC.LineStyle = p.Values.LineStyle
			}
			if p.ValueMask&wire.GCCapStyle != 0 {
				existingGC.CapStyle = p.Values.CapStyle
			}
			if p.ValueMask&wire.GCJoinStyle != 0 {
				existingGC.JoinStyle = p.Values.JoinStyle
			}
			if p.ValueMask&wire.GCFillStyle != 0 {
				existingGC.FillStyle = p.Values.FillStyle
			}
			if p.ValueMask&wire.GCFillRule != 0 {
				existingGC.FillRule = p.Values.FillRule
			}
			if p.ValueMask&wire.GCTile != 0 {
				existingGC.Tile = p.Values.Tile
			}
			if p.ValueMask&wire.GCStipple != 0 {
				existingGC.Stipple = p.Values.Stipple
			}
			if p.ValueMask&wire.GCTileStipXOrigin != 0 {
				existingGC.TileStipXOrigin = p.Values.TileStipXOrigin
			}
			if p.ValueMask&wire.GCTileStipYOrigin != 0 {
				existingGC.TileStipYOrigin = p.Values.TileStipYOrigin
			}
			if p.ValueMask&wire.GCFont != 0 {
				existingGC.Font = p.Values.Font
			}
			if p.ValueMask&wire.GCSubwindowMode != 0 {
				existingGC.SubwindowMode = p.Values.SubwindowMode
			}
			if p.ValueMask&wire.GCGraphicsExposures != 0 {
				existingGC.GraphicsExposures = p.Values.GraphicsExposures
			}
			if p.ValueMask&wire.GCClipXOrigin != 0 {
				existingGC.ClipXOrigin = p.Values.ClipXOrigin
			}
			if p.ValueMask&wire.GCClipYOrigin != 0 {
				existingGC.ClipYOrigin = p.Values.ClipYOrigin
			}
			if p.ValueMask&wire.GCClipMask != 0 {
				existingGC.ClipMask = p.Values.ClipMask
			}
			if p.ValueMask&wire.GCDashOffset != 0 {
				existingGC.DashOffset = p.Values.DashOffset
			}
			if p.ValueMask&wire.GCDashes != 0 {
				existingGC.Dashes = p.Values.Dashes
			}
			if p.ValueMask&wire.GCArcMode != 0 {
				existingGC.ArcMode = p.Values.ArcMode
			}
		}
		s.frontend.ChangeGC(xid, p.ValueMask, p.Values)

	case *wire.CopyGCRequest:
		srcGC := client.xID(uint32(p.SrcGC))
		dstGC := client.xID(uint32(p.DstGC))
		if err := s.checkGC(srcGC, seq, wire.CopyGC, 0); err != nil {
			return err
		}
		if err := s.checkGC(dstGC, seq, wire.CopyGC, 0); err != nil {
			return err
		}
		s.frontend.CopyGC(srcGC, dstGC)

	case *wire.SetDashesRequest:
		gc := client.xID(uint32(p.GC))
		if err := s.checkGC(gc, seq, wire.SetDashes, 0); err != nil {
			return err
		}
		s.frontend.SetDashes(gc, p.DashOffset, p.Dashes)

	case *wire.SetClipRectanglesRequest:
		gc := client.xID(uint32(p.GC))
		if err := s.checkGC(gc, seq, wire.SetClipRectangles, 0); err != nil {
			return err
		}
		s.frontend.SetClipRectangles(gc, p.ClippingX, p.ClippingY, p.Rectangles, p.Ordering)

	case *wire.FreeGCRequest:
		gcID := client.xID(uint32(p.GC))
		if err := s.checkGC(gcID, seq, wire.FreeGC, 0); err != nil {
			return err
		}
		delete(s.gcs, gcID)
		s.frontend.FreeGC(gcID)

	case *wire.ClearAreaRequest:
		if err := s.checkWindow(client.xID(uint32(p.Window)), seq, wire.ClearArea, 0); err != nil {
			return err
		}
		s.frontend.ClearArea(client.xID(uint32(p.Window)), int32(p.X), int32(p.Y), int32(p.Width), int32(p.Height))
		s.frontend.ComposeWindow(client.xID(uint32(p.Window)))

	case *wire.CopyAreaRequest:
		gcID := client.xID(uint32(p.Gc))
		if err := s.checkGC(gcID, seq, wire.CopyArea, 0); err != nil {
			return err
		}
		if err := s.checkDrawable(client.xID(uint32(p.SrcDrawable)), seq, wire.CopyArea, 0); err != nil {
			return err
		}
		if err := s.checkDrawable(client.xID(uint32(p.DstDrawable)), seq, wire.CopyArea, 0); err != nil {
			return err
		}
		s.frontend.CopyArea(client.xID(uint32(p.SrcDrawable)), client.xID(uint32(p.DstDrawable)), gcID, int32(p.SrcX), int32(p.SrcY), int32(p.DstX), int32(p.DstY), int32(p.Width), int32(p.Height))
		s.frontend.ComposeWindow(client.xID(uint32(p.DstDrawable)))

	case *wire.CopyPlaneRequest:
		gcID := client.xID(uint32(p.Gc))
		if err := s.checkGC(gcID, seq, wire.CopyPlane, 0); err != nil {
			return err
		}
		if err := s.checkDrawable(client.xID(uint32(p.SrcDrawable)), seq, wire.CopyPlane, 0); err != nil {
			return err
		}
		if err := s.checkDrawable(client.xID(uint32(p.DstDrawable)), seq, wire.CopyPlane, 0); err != nil {
			return err
		}
		s.frontend.CopyPlane(client.xID(uint32(p.SrcDrawable)), client.xID(uint32(p.DstDrawable)), gcID, int32(p.SrcX), int32(p.SrcY), int32(p.DstX), int32(p.DstY), int32(p.Width), int32(p.Height), int32(p.PlaneMask))
		s.frontend.ComposeWindow(client.xID(uint32(p.DstDrawable)))

	case *wire.PolyPointRequest:
		gcID := client.xID(uint32(p.Gc))
		if err := s.checkGC(gcID, seq, wire.PolyPoint, 0); err != nil {
			return err
		}
		if err := s.checkDrawable(client.xID(uint32(p.Drawable)), seq, wire.PolyPoint, 0); err != nil {
			return err
		}
		s.frontend.PolyPoint(client.xID(uint32(p.Drawable)), gcID, p.Coordinates)
		s.frontend.ComposeWindow(client.xID(uint32(p.Drawable)))

	case *wire.PolyLineRequest:
		gcID := client.xID(uint32(p.Gc))
		if err := s.checkGC(gcID, seq, wire.PolyLine, 0); err != nil {
			return err
		}
		if err := s.checkDrawable(client.xID(uint32(p.Drawable)), seq, wire.PolyLine, 0); err != nil {
			return err
		}
		s.frontend.PolyLine(client.xID(uint32(p.Drawable)), gcID, p.Coordinates)
		s.frontend.ComposeWindow(client.xID(uint32(p.Drawable)))

	case *wire.PolySegmentRequest:
		gcID := client.xID(uint32(p.Gc))
		if err := s.checkGC(gcID, seq, wire.PolySegment, 0); err != nil {
			return err
		}
		if err := s.checkDrawable(client.xID(uint32(p.Drawable)), seq, wire.PolySegment, 0); err != nil {
			return err
		}
		s.frontend.PolySegment(client.xID(uint32(p.Drawable)), gcID, p.Segments)
		s.frontend.ComposeWindow(client.xID(uint32(p.Drawable)))

	case *wire.PolyRectangleRequest:
		gcID := client.xID(uint32(p.Gc))
		if err := s.checkGC(gcID, seq, wire.PolyRectangle, 0); err != nil {
			return err
		}
		if err := s.checkDrawable(client.xID(uint32(p.Drawable)), seq, wire.PolyRectangle, 0); err != nil {
			return err
		}
		s.frontend.PolyRectangle(client.xID(uint32(p.Drawable)), gcID, p.Rectangles)
		s.frontend.ComposeWindow(client.xID(uint32(p.Drawable)))

	case *wire.PolyArcRequest:
		gcID := client.xID(uint32(p.Gc))
		if err := s.checkGC(gcID, seq, wire.PolyArc, 0); err != nil {
			return err
		}
		if err := s.checkDrawable(client.xID(uint32(p.Drawable)), seq, wire.PolyArc, 0); err != nil {
			return err
		}
		s.frontend.PolyArc(client.xID(uint32(p.Drawable)), gcID, p.Arcs)
		s.frontend.ComposeWindow(client.xID(uint32(p.Drawable)))

	case *wire.FillPolyRequest:
		gcID := client.xID(uint32(p.Gc))
		if err := s.checkGC(gcID, seq, wire.FillPoly, 0); err != nil {
			return err
		}
		if err := s.checkDrawable(client.xID(uint32(p.Drawable)), seq, wire.FillPoly, 0); err != nil {
			return err
		}
		s.frontend.FillPoly(client.xID(uint32(p.Drawable)), gcID, p.Coordinates)
		s.frontend.ComposeWindow(client.xID(uint32(p.Drawable)))

	case *wire.PolyFillRectangleRequest:
		gcID := client.xID(uint32(p.Gc))
		if err := s.checkGC(gcID, seq, wire.PolyFillRectangle, 0); err != nil {
			return err
		}
		if err := s.checkDrawable(client.xID(uint32(p.Drawable)), seq, wire.PolyFillRectangle, 0); err != nil {
			return err
		}
		s.frontend.PolyFillRectangle(client.xID(uint32(p.Drawable)), gcID, p.Rectangles)
		s.frontend.ComposeWindow(client.xID(uint32(p.Drawable)))

	case *wire.PolyFillArcRequest:
		gcID := client.xID(uint32(p.Gc))
		if err := s.checkGC(gcID, seq, wire.PolyFillArc, 0); err != nil {
			return err
		}
		if err := s.checkDrawable(client.xID(uint32(p.Drawable)), seq, wire.PolyFillArc, 0); err != nil {
			return err
		}
		s.frontend.PolyFillArc(client.xID(uint32(p.Drawable)), gcID, p.Arcs)
		s.frontend.ComposeWindow(client.xID(uint32(p.Drawable)))

	case *wire.PutImageRequest:
		gcID := client.xID(uint32(p.Gc))
		if err := s.checkGC(gcID, seq, wire.PutImage, 0); err != nil {
			return err
		}
		if err := s.checkDrawable(client.xID(uint32(p.Drawable)), seq, wire.PutImage, 0); err != nil {
			return err
		}
		s.frontend.PutImage(client.xID(uint32(p.Drawable)), gcID, p.Format, p.Width, p.Height, p.DstX, p.DstY, p.LeftPad, p.Depth, p.Data)
		s.frontend.ComposeWindow(client.xID(uint32(p.Drawable)))

	case *wire.GetImageRequest:
		if err := s.checkDrawable(client.xID(uint32(p.Drawable)), seq, wire.GetImage, 0); err != nil {
			return err
		}
		imgData, err := s.frontend.GetImage(client.xID(uint32(p.Drawable)), int32(p.X), int32(p.Y), int32(p.Width), int32(p.Height), p.PlaneMask)
		if err != nil {
			return wire.NewGenericError(seq, 0, 0, wire.GetImage, wire.MatchErrorCode)
		}
		return &wire.GetImageReply{
			Sequence:  seq,
			Depth:     24, // Assuming 24-bit depth for now
			VisualID:  s.visualID,
			ImageData: imgData,
		}

	case *wire.PolyText8Request:
		gcID := client.xID(uint32(p.GC))
		if err := s.checkGC(gcID, seq, wire.PolyText8, 0); err != nil {
			return err
		}
		if err := s.checkDrawable(client.xID(uint32(p.Drawable)), seq, wire.PolyText8, 0); err != nil {
			return err
		}
		s.frontend.PolyText8(client.xID(uint32(p.Drawable)), gcID, int32(p.X), int32(p.Y), p.Items)
		s.frontend.ComposeWindow(client.xID(uint32(p.Drawable)))

	case *wire.PolyText16Request:
		gcID := client.xID(uint32(p.GC))
		if err := s.checkGC(gcID, seq, wire.PolyText16, 0); err != nil {
			return err
		}
		if err := s.checkDrawable(client.xID(uint32(p.Drawable)), seq, wire.PolyText16, 0); err != nil {
			return err
		}
		s.frontend.PolyText16(client.xID(uint32(p.Drawable)), gcID, int32(p.X), int32(p.Y), p.Items)
		s.frontend.ComposeWindow(client.xID(uint32(p.Drawable)))

	case *wire.ImageText8Request:
		gcID := client.xID(uint32(p.Gc))
		if err := s.checkGC(gcID, seq, wire.ImageText8, 0); err != nil {
			return err
		}
		if err := s.checkDrawable(client.xID(uint32(p.Drawable)), seq, wire.ImageText8, 0); err != nil {
			return err
		}
		s.frontend.ImageText8(client.xID(uint32(p.Drawable)), gcID, int32(p.X), int32(p.Y), p.Text)
		s.frontend.ComposeWindow(client.xID(uint32(p.Drawable)))

	case *wire.ImageText16Request:
		gcID := client.xID(uint32(p.Gc))
		if err := s.checkGC(gcID, seq, wire.ImageText16, 0); err != nil {
			return err
		}
		if err := s.checkDrawable(client.xID(uint32(p.Drawable)), seq, wire.ImageText16, 0); err != nil {
			return err
		}
		s.frontend.ImageText16(client.xID(uint32(p.Drawable)), gcID, int32(p.X), int32(p.Y), p.Text)
		s.frontend.ComposeWindow(client.xID(uint32(p.Drawable)))

	case *wire.CreateColormapRequest:
		xid := client.xID(uint32(p.Mid))

		if _, exists := s.colormaps[xid]; exists {
			return wire.NewGenericError(seq, uint32(p.Mid), 0, wire.CreateColormap, wire.IDChoiceErrorCode)
		}
		if err := s.checkWindow(client.xID(uint32(p.Window)), seq, wire.CreateColormap, 0); err != nil {
			return err
		}

		newColormap := &colormap{
			pixels: make(map[uint32]wire.XColorItem),
		}

		if p.Alloc == 1 { // All
			// For TrueColor, pre-allocating doesn't make much sense as pixels are calculated.
			// For other visual types, this would be important.
			// For now, we'll just create an empty map.
		}

		s.colormaps[xid] = newColormap

	case *wire.FreeColormapRequest:
		xid := client.xID(uint32(p.Cmap))
		if err := s.checkColormap(xid, seq, wire.FreeColormap, 0); err != nil {
			return err
		}
		delete(s.colormaps, xid)

	case *wire.CopyColormapAndFreeRequest:
		srcCmapID := client.xID(uint32(p.SrcCmap))
		if err := s.checkColormap(srcCmapID, seq, wire.CopyColormapAndFree, 0); err != nil {
			return err
		}
		srcCmap := s.colormaps[srcCmapID]

		newCmapID := client.xID(uint32(p.Mid))
		if _, exists := s.colormaps[newCmapID]; exists {
			return wire.NewGenericError(seq, uint32(p.Mid), 0, wire.CopyColormapAndFree, wire.IDChoiceErrorCode)
		}

		newCmap := &colormap{
			pixels: make(map[uint32]wire.XColorItem),
		}
		s.colormaps[newCmapID] = newCmap

		for pixel, color := range srcCmap.pixels {
			if color.ClientID == client.id {
				newCmap.pixels[pixel] = color
				delete(srcCmap.pixels, pixel)
			}
		}

	case *wire.InstallColormapRequest:
		xid := client.xID(uint32(p.Cmap))
		if err := s.checkColormap(xid, seq, wire.InstallColormap, 0); err != nil {
			return err
		}

		s.installedColormap = xid

		for winID, win := range s.windows {
			if win.colormap == xid {
				client, ok := s.clients[winID.client]
				if !ok {
					debugf("X11: InstallColormap unknown client %d", winID.client)
					continue
				}
				event := &wire.ColormapNotifyEvent{
					Sequence: client.sequence - 1,
					Window:   winID.local,
					Colormap: uint32(p.Cmap),
					New:      true,
					State:    0, // Installed
				}
				s.sendEvent(client, event)
			}
		}
		return nil

	case *wire.UninstallColormapRequest:
		xid := client.xID(uint32(p.Cmap))
		if err := s.checkColormap(xid, seq, wire.UninstallColormap, 0); err != nil {
			return err
		}

		if s.installedColormap == xid {
			s.installedColormap = xID{local: s.defaultColormap}
		}

		for winID, win := range s.windows {
			if win.colormap == xid {
				client, ok := s.clients[winID.client]
				if !ok {
					debugf("X11: UninstallColormap unknown client %d", winID.client)
					continue
				}
				event := &wire.ColormapNotifyEvent{
					Sequence: client.sequence - 1,
					Window:   winID.local,
					Colormap: uint32(p.Cmap),
					New:      false,
					State:    1, // Uninstalled
				}
				s.sendEvent(client, event)
			}
		}
		return nil

	case *wire.ListInstalledColormapsRequest:
		if err := s.checkWindow(client.xID(uint32(p.Window)), seq, wire.ListInstalledColormaps, 0); err != nil {
			return err
		}
		var colormaps []uint32
		if s.installedColormap.local != 0 {
			colormaps = append(colormaps, s.installedColormap.local)
		}

		return &wire.ListInstalledColormapsReply{
			Sequence:     seq,
			NumColormaps: uint16(len(colormaps)),
			Colormaps:    colormaps,
		}

	case *wire.AllocColorRequest:
		xid := client.xID(uint32(p.Cmap))
		if xid.local == s.defaultColormap {
			xid.client = 0
		}
		cm, ok := s.colormaps[xid]
		if !ok {
			return wire.NewGenericError(seq, uint32(p.Cmap), 0, wire.AllocColor, wire.ColormapErrorCode)
		}

		// Simple allocation for TrueColor: construct pixel value from RGB
		r8 := byte(p.Red >> 8)
		g8 := byte(p.Green >> 8)
		b8 := byte(p.Blue >> 8)
		pixel := (uint32(r8) << 16) | (uint32(g8) << 8) | uint32(b8)

		cm.pixels[pixel] = wire.XColorItem{Red: p.Red, Green: p.Green, Blue: p.Blue, ClientID: client.id}

		return &wire.AllocColorReply{
			Sequence: seq,
			Red:      p.Red,
			Green:    p.Green,
			Blue:     p.Blue,
			Pixel:    pixel,
		}

	case *wire.AllocNamedColorRequest:
		cmap := client.xID(uint32(p.Cmap))
		if cmap.local == s.defaultColormap {
			cmap.client = 0
		}
		cm, ok := s.colormaps[cmap]
		if !ok {
			return wire.NewError(wire.ColormapErrorCode, seq, uint32(p.Cmap), wire.Opcodes{Major: p.OpCode(), Minor: 0})
		}

		name := string(p.Name)
		rgb, ok := lookupColor(name)
		if !ok {
			return wire.NewError(wire.NameErrorCode, seq, 0, wire.Opcodes{Major: p.OpCode(), Minor: 0})
		}

		exactRed := scale8to16(rgb.Red)
		exactGreen := scale8to16(rgb.Green)
		exactBlue := scale8to16(rgb.Blue)

		pixel := (uint32(rgb.Red) << 16) | (uint32(rgb.Green) << 8) | uint32(rgb.Blue)
		cm.pixels[pixel] = wire.XColorItem{Red: exactRed, Green: exactGreen, Blue: exactBlue, ClientID: client.id}

		return &wire.AllocNamedColorReply{
			Sequence:   seq,
			Pixel:      pixel,
			ExactRed:   exactRed,
			ExactGreen: exactGreen,
			ExactBlue:  exactBlue,
			Red:        exactRed,
			Green:      exactGreen,
			Blue:       exactBlue,
		}

	case *wire.AllocColorCellsRequest:
		return wire.NewGenericError(seq, 0, 0, wire.AllocColorCells, wire.MatchErrorCode)

	case *wire.AllocColorPlanesRequest:
		return wire.NewGenericError(seq, 0, 0, wire.AllocColorPlanes, wire.MatchErrorCode)

	case *wire.FreeColorsRequest:
		xid := client.xID(uint32(p.Cmap))
		if xid.local == s.defaultColormap {
			xid.client = 0
		}
		cm, ok := s.colormaps[xid]
		if !ok {
			return wire.NewGenericError(seq, uint32(p.Cmap), 0, wire.FreeColors, wire.ColormapErrorCode)
		}

		for _, pixel := range p.Pixels {
			delete(cm.pixels, pixel)
		}

	case *wire.StoreColorsRequest:
		xid := client.xID(uint32(p.Cmap))
		if xid.local == s.defaultColormap {
			xid.client = 0
		}
		cm, ok := s.colormaps[xid]
		if !ok {
			return wire.NewGenericError(seq, uint32(p.Cmap), 0, wire.StoreColors, wire.ColormapErrorCode)
		}

		for _, item := range p.Items {
			c, exists := cm.pixels[item.Pixel]
			if !exists {
				c = wire.XColorItem{}
			}

			if item.Flags&wire.DoRed != 0 {
				c.Red = item.Red
			}
			if item.Flags&wire.DoGreen != 0 {
				c.Green = item.Green
			}
			if item.Flags&wire.DoBlue != 0 {
				c.Blue = item.Blue
			}
			cm.pixels[item.Pixel] = c
		}

	case *wire.StoreNamedColorRequest:
		xid := client.xID(uint32(p.Cmap))
		if xid.local == s.defaultColormap {
			xid.client = 0
		}
		cm, ok := s.colormaps[xid]
		if !ok {
			return wire.NewGenericError(seq, uint32(p.Cmap), 0, wire.StoreNamedColor, wire.ColormapErrorCode)
		}

		rgb, ok := lookupColor(p.Name)
		if !ok {
			return wire.NewGenericError(seq, 0, 0, wire.StoreNamedColor, wire.NameErrorCode)
		}

		c, exists := cm.pixels[p.Pixel]
		if !exists {
			c = wire.XColorItem{}
		}

		if p.Flags&wire.DoRed != 0 {
			c.Red = scale8to16(rgb.Red)
		}
		if p.Flags&wire.DoGreen != 0 {
			c.Green = scale8to16(rgb.Green)
		}
		if p.Flags&wire.DoBlue != 0 {
			c.Blue = scale8to16(rgb.Blue)
		}
		cm.pixels[p.Pixel] = c

	case *wire.QueryColorsRequest:
		cmap := client.xID(p.Cmap)
		if cmap.local == s.defaultColormap {
			cmap.client = 0
		}
		cm, ok := s.colormaps[cmap]
		if !ok {
			return wire.NewGenericError(seq, uint32(p.Cmap), 0, wire.QueryColors, wire.ColormapErrorCode)
		}

		colors := make([]wire.XColorItem, len(p.Pixels))
		for i, pixel := range p.Pixels {
			color, ok := cm.pixels[pixel]
			if !ok {
				// If the pixel is not in the colormap, return black
				color = wire.XColorItem{Red: 0, Green: 0, Blue: 0}
			}
			colors[i] = color
		}

		return &wire.QueryColorsReply{
			Sequence: seq,
			Colors:   colors,
		}

	case *wire.LookupColorRequest:
		// cmapID := client.xID(uint32(p.Cmap))

		color, ok := lookupColor(p.Name)
		if !ok {
			return wire.NewGenericError(seq, uint32(p.Cmap), 0, wire.LookupColor, wire.NameErrorCode)
		}

		return &wire.LookupColorReply{
			Sequence:   seq,
			Red:        scale8to16(color.Red),
			Green:      scale8to16(color.Green),
			Blue:       scale8to16(color.Blue),
			ExactRed:   scale8to16(color.Red),
			ExactGreen: scale8to16(color.Green),
			ExactBlue:  scale8to16(color.Blue),
		}

	case *wire.CreateCursorRequest:
		cursorXID := client.xID(uint32(p.Cid))
		if _, exists := s.cursors[cursorXID]; exists {
			s.logger.Errorf("X11: CreateCursor: ID %s already in use", cursorXID)
			return wire.NewGenericError(seq, uint32(p.Cid), 0, wire.CreateCursor, wire.IDChoiceErrorCode)
		}

		sourceXID := client.xID(uint32(p.Source))
		if err := s.checkPixmap(sourceXID, seq, wire.CreateCursor, 0); err != nil {
			return err
		}
		maskXID := client.xID(uint32(p.Mask))
		if p.Mask != 0 {
			if err := s.checkPixmap(maskXID, seq, wire.CreateCursor, 0); err != nil {
				return err
			}
		}

		s.cursors[cursorXID] = true
		foreColor := [3]uint16{p.ForeRed, p.ForeGreen, p.ForeBlue}
		backColor := [3]uint16{p.BackRed, p.BackGreen, p.BackBlue}
		s.frontend.CreateCursor(cursorXID, sourceXID, maskXID, foreColor, backColor, p.X, p.Y)

	case *wire.CreateGlyphCursorRequest:
		// Check if the cursor ID is already in use
		if _, exists := s.cursors[client.xID(uint32(p.Cid))]; exists {
			s.logger.Errorf("X11: CreateGlyphCursor: ID %d already in use", p.Cid)
			return wire.NewGenericError(seq, uint32(p.Cid), 0, wire.CreateGlyphCursor, wire.IDChoiceErrorCode)
		}
		if err := s.checkFont(client.xID(uint32(p.SourceFont)), seq, wire.CreateGlyphCursor, 0); err != nil {
			return err
		}
		if p.MaskFont != 0 {
			if err := s.checkFont(client.xID(uint32(p.MaskFont)), seq, wire.CreateGlyphCursor, 0); err != nil {
				return err
			}
		}

		s.cursors[client.xID(uint32(p.Cid))] = true
		s.frontend.CreateCursorFromGlyph(uint32(p.Cid), p.SourceChar)

	case *wire.FreeCursorRequest:
		xid := client.xID(uint32(p.Cursor))
		if err := s.checkCursor(xid, seq, wire.FreeCursor, 0); err != nil {
			return err
		}
		delete(s.cursors, xid)
		s.frontend.FreeCursor(xid)

	case *wire.RecolorCursorRequest:
		if err := s.checkCursor(client.xID(uint32(p.Cursor)), seq, wire.RecolorCursor, 0); err != nil {
			return err
		}
		s.frontend.RecolorCursor(client.xID(uint32(p.Cursor)), p.ForeColor, p.BackColor)

	case *wire.QueryBestSizeRequest:
		if err := s.checkDrawable(client.xID(uint32(p.Drawable)), seq, wire.QueryBestSize, 0); err != nil {
			return err
		}
		debugf("X11: QueryBestSize class=%d drawable=%d width=%d height=%d", p.Class, p.Drawable, p.Width, p.Height)

		return &wire.QueryBestSizeReply{
			Sequence: seq,
			Width:    p.Width,
			Height:   p.Height,
		}

	case *wire.QueryExtensionRequest:
		debugf("X11: QueryExtension name=%s", p.Name)

		switch p.Name {
		case wire.BigRequestsExtensionName:
			return &wire.QueryExtensionReply{
				Sequence:    seq,
				Present:     true,
				MajorOpcode: byte(wire.BigRequestsOpcode),
				FirstEvent:  0,
				FirstError:  0,
			}
		case wire.XInputExtensionName:
			return &wire.QueryExtensionReply{
				Sequence:    seq,
				Present:     true,
				MajorOpcode: byte(wire.XInputOpcode),
				FirstEvent:  0,
				FirstError:  0,
			}
		default:
			return &wire.QueryExtensionReply{
				Sequence:    seq,
				Present:     false,
				MajorOpcode: 0,
				FirstEvent:  0,
				FirstError:  0,
			}
		}

	case *wire.ListExtensionsRequest:
		extensions := []string{
			wire.BigRequestsExtensionName,
			wire.XInputExtensionName,
		}
		return &wire.ListExtensionsReply{
			Sequence: seq,
			NNames:   byte(len(extensions)),
			Names:    extensions,
		}

	case *wire.ChangeKeyboardMappingRequest:
		keySymIndex := 0
		for i := 0; i < int(p.KeyCodeCount); i++ {
			keyCode := p.FirstKeyCode + wire.KeyCode(i)
			for j := 0; j < int(p.KeySymsPerKeyCode); j++ {
				if j == 0 {
					s.keymap[byte(keyCode)] = p.KeySyms[keySymIndex]
				}
				keySymIndex++
			}
		}

	case *wire.GetKeyboardMappingRequest:
		keySyms := make([]uint32, 0, p.Count)
		for i := 0; i < int(p.Count); i++ {
			keyCode := p.FirstKeyCode + wire.KeyCode(i)
			keySym, ok := s.keymap[byte(keyCode)]
			if !ok {
				keySym = 0 // NoSymbol
			}
			keySyms = append(keySyms, keySym)
		}
		return &wire.GetKeyboardMappingReply{
			Sequence: seq,
			KeySyms:  keySyms,
		}

	case *wire.ChangeKeyboardControlRequest:
		s.frontend.ChangeKeyboardControl(p.ValueMask, p.Values)

	case *wire.GetKeyboardControlRequest:
		kc, _ := s.frontend.GetKeyboardControl()
		return &wire.GetKeyboardControlReply{
			Sequence:         seq,
			KeyClickPercent:  byte(kc.KeyClickPercent),
			BellPercent:      byte(kc.BellPercent),
			BellPitch:        uint16(kc.BellPitch),
			BellDuration:     uint16(kc.BellDuration),
			LedMask:          uint32(kc.Led),
			GlobalAutoRepeat: byte(kc.AutoRepeatMode),
			AutoRepeats:      [32]byte{},
		}

	case *wire.BellRequest:
		s.frontend.Bell(p.Percent)

	case *wire.ChangePointerControlRequest:
		s.frontend.ChangePointerControl(p.AccelerationNumerator, p.AccelerationDenominator, p.Threshold, p.DoAcceleration, p.DoThreshold)

	case *wire.GetPointerControlRequest:
		accelNumerator, accelDenominator, threshold, _ := s.frontend.GetPointerControl()
		return &wire.GetPointerControlReply{
			Sequence:         seq,
			AccelNumerator:   accelNumerator,
			AccelDenominator: accelDenominator,
			Threshold:        threshold,
		}

	case *wire.SetScreenSaverRequest:
		s.frontend.SetScreenSaver(p.Timeout, p.Interval, p.PreferBlank, p.AllowExpose)

	case *wire.GetScreenSaverRequest:
		timeout, interval, preferBlank, allowExpose, _ := s.frontend.GetScreenSaver()
		return &wire.GetScreenSaverReply{
			Sequence:    seq,
			Timeout:     uint16(timeout),
			Interval:    uint16(interval),
			PreferBlank: preferBlank,
			AllowExpose: allowExpose,
		}

	case *wire.ChangeHostsRequest:
		s.frontend.ChangeHosts(p.Mode, p.Host)

	case *wire.ListHostsRequest:
		hosts, _ := s.frontend.ListHosts()
		return &wire.ListHostsReply{
			Sequence: seq,
			NumHosts: uint16(len(hosts)),
			Hosts:    hosts,
		}

	case *wire.SetAccessControlRequest:
		s.frontend.SetAccessControl(p.Mode)

	case *wire.SetCloseDownModeRequest:
		s.frontend.SetCloseDownMode(p.Mode)

	case *wire.KillClientRequest:
		s.frontend.KillClient(p.Resource)

	case *wire.RotatePropertiesRequest:
		err := s.RotateProperties(client.xID(uint32(p.Window)), p.Delta, p.Atoms)
		if err != nil {
			return wire.NewGenericError(seq, uint32(p.Window), 0, wire.RotateProperties, wire.MatchErrorCode)
		}

	case *wire.ForceScreenSaverRequest:
		s.frontend.ForceScreenSaver(p.Mode)

	case *wire.SetPointerMappingRequest:
		status, _ := s.frontend.SetPointerMapping(p.Map)
		return &wire.SetPointerMappingReply{
			Sequence: seq,
			Status:   status,
		}

	case *wire.GetPointerMappingRequest:
		pMap, _ := s.frontend.GetPointerMapping()
		return &wire.GetPointerMappingReply{
			Sequence: seq,
			Length:   byte(len(pMap)),
			PMap:     pMap,
		}

	case *wire.SetModifierMappingRequest:
		status, _ := s.frontend.SetModifierMapping(p.KeyCodesPerModifier, p.KeyCodes)
		return &wire.SetModifierMappingReply{
			Sequence: seq,
			Status:   status,
		}

	case *wire.GetModifierMappingRequest:
		keyCodes, err := s.frontend.GetModifierMapping()
		if err != nil {
			return wire.NewGenericError(seq, 0, 0, wire.GetModifierMapping, wire.ImplementationErrorCode)
		}
		return &wire.GetModifierMappingReply{
			Sequence:            seq,
			KeyCodesPerModifier: byte(len(keyCodes) / 8),
			KeyCodes:            keyCodes,
		}

	case *wire.NoOperationRequest:

	case *wire.EnableBigRequestsRequest:
		client.bigRequestsEnabled = true
		return &wire.BigRequestsEnableReply{
			Sequence:         seq,
			MaxRequestLength: 0x100000,
		}

	default:
		debugf("Unknown X11 request opcode: %d", p.OpCode())
	}
	return nil
}

func (s *x11Server) handshake(client *x11Client) {
	var handshake [12]byte
	if _, err := io.ReadFull(client.conn, handshake[:]); err != nil {
		debugf("x11 handshake: %v", err)
		return
	}

	var order binary.ByteOrder
	if handshake[0] == 'B' {
		order = binary.BigEndian
	} else {
		order = binary.LittleEndian
	}
	s.byteOrder = order
	client.byteOrder = order
	authProtoNameLen := order.Uint16(handshake[6:8])
	authProtoDataLen := order.Uint16(handshake[8:10])

	authProtoName := make([]byte, authProtoNameLen)
	if _, err := io.ReadFull(client.conn, authProtoName); err != nil {
		debugf("Failed to read auth protocol name: %v", err)
		return
	}
	if pad := authProtoNameLen % 4; pad != 0 {
		if _, err := io.CopyN(io.Discard, client.conn, int64(4-pad)); err != nil {
			debugf("Failed to discard auth protocol name padding: %v", err)
			return
		}
	}
	authProtoData := make([]byte, authProtoDataLen)
	if _, err := io.ReadFull(client.conn, authProtoData); err != nil {
		debugf("Failed to read auth protocol data: %v", err)
		return
	}
	if pad := authProtoDataLen % 4; pad != 0 {
		if _, err := io.CopyN(io.Discard, client.conn, int64(4-pad)); err != nil {
			debugf("Failed to discard auth protocol data padding: %v", err)
			return
		}
	}

	if s.authProtocol != "" || s.authCookie != nil {
		if s.authProtocol != string(authProtoName) || string(s.authCookie) != string(authProtoData) {
			debugf("X11 auth failed: protocol=%q cookie=%q, expected protocol=%q cookie=%q",
				authProtoName, authProtoData, s.authProtocol, s.authCookie)
			client.send(&wire.SetupResponse{
				Success: 0, // Failed
				Reason:  "Invalid authorization",
			})
			return
		}
	}

	setup := wire.NewDefaultSetup()

	// Create the setup response message encoder
	responseMsg := &wire.SetupResponse{
		Success:                  1, // Success
		ProtocolVersion:          11,
		ReleaseNumber:            setup.ReleaseNumber,
		ResourceIDBase:           setup.ResourceIDBase,
		ResourceIDMask:           setup.ResourceIDMask,
		MotionBufferSize:         setup.MotionBufferSize,
		VendorLength:             setup.VendorLength,
		MaxRequestLength:         setup.MaxRequestLength,
		NumScreens:               setup.NumScreens,
		NumPixmapFormats:         setup.NumPixmapFormats,
		ImageByteOrder:           setup.ImageByteOrder,
		BitmapFormatBitOrder:     setup.BitmapFormatBitOrder,
		BitmapFormatScanlineUnit: setup.BitmapFormatScanlineUnit,
		BitmapFormatScanlinePad:  setup.BitmapFormatScanlinePad,
		MinKeycode:               setup.MinKeycode,
		MaxKeycode:               setup.MaxKeycode,
		VendorString:             setup.VendorString,
		PixmapFormats:            setup.PixmapFormats,
		Screens:                  setup.Screens,
	}

	if err := client.send(responseMsg); err != nil {
		s.logger.Errorf("x11 handshake write: %v", err)
		return
	}
	s.visualID = setup.Screens[0].RootVisual
	s.rootVisual = setup.Screens[0].Depths[0].Visuals[0]
	s.blackPixel = setup.Screens[0].BlackPixel
	s.whitePixel = setup.Screens[0].WhitePixel
}

func HandleX11Forwarding(logger Logger, client *ssh.Client, authProtocol string, authCookie []byte) {
	x11channels := client.HandleChannelOpen("x11")
	go func() {
		for ch := range x11channels {
			channel, requests, err := ch.Accept()
			if err != nil {
				logger.Errorf("x11 channel accept: %v", err)
				continue
			}
			go ssh.DiscardRequests(requests)

			once.Do(func() {
				x11ServerInstance = &x11Server{
					logger:     logger,
					windows:    make(map[xID]*window),
					gcs:        make(map[xID]wire.GC),
					pixmaps:    make(map[xID]bool),
					cursors:    make(map[xID]bool),
					selections: make(map[uint32]*selectionOwner),
					properties: make(map[xID]map[uint32]*property),
					colormaps: map[xID]*colormap{
						xID{local: 0x1}: {
							pixels: map[uint32]wire.XColorItem{
								0x000000: {Pixel: 0x000000, Red: 0x0000, Green: 0x0000, Blue: 0x0000, Flags: 0},
								1:        {Pixel: 1, Red: 0xffff, Green: 0xffff, Blue: 0xffff, Flags: 0},
								0xffffff: {Pixel: 0xffffff, Red: 0xffff, Green: 0xffff, Blue: 0xffff, Flags: 0},
							},
						},
					},
					defaultColormap:    0x1,
					clients:            make(map[uint32]*x11Client),
					nextClientID:       1,
					passiveGrabs:       make(map[xID][]*passiveGrab),
					passiveDeviceGrabs: make(map[xID][]*passiveDeviceGrab),
					deviceGrabs:        make(map[byte]*deviceGrab),
					authProtocol:       authProtocol,
					authCookie:         authCookie,
					keymap:             make(map[byte]uint32),
					fonts:              make(map[xID]bool),
					startTime:          time.Now(),
				}
				x11ServerInstance.initAtoms()
				for k, v := range KeyCodeToKeysym {
					x11ServerInstance.keymap[k] = v
				}
				x11ServerInstance.frontend = newX11Frontend(logger, x11ServerInstance)
			})

			client := &x11Client{
				id:            x11ServerInstance.nextClientID,
				conn:          channel,
				sequence:      0,
				byteOrder:     binary.LittleEndian, // Default, will be updated in handshake
				saveSet:       make(map[uint32]bool),
				openDevices:   make(map[byte]*wire.DeviceInfo),
				xi2EventMasks: make(map[uint32]map[uint16][]uint32),
			}
			x11ServerInstance.clients[client.id] = client
			x11ServerInstance.nextClientID++
			go func() {
				x11ServerInstance.handshake(client)
				x11ServerInstance.serve(client)
			}()
		}
	}()
}
