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
	ChangeProperty(xid xID, property, typeAtom, format uint32, data []byte)
	DeleteProperty(xid xID, property uint32)
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
	GetAtom(clientID uint32, name string) uint32
	GetAtomName(atom uint32) string
	ListProperties(window xID) []uint32
	GetProperty(window xID, property uint32, longOffset, longLength uint32) (data []byte, typ, format, bytesAfter uint32)
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
	ConvertSelection(selection, target, property uint32, requestor xID)
	GrabPointer(grabWindow xID, ownerEvents bool, eventMask uint16, pointerMode, keyboardMode byte, confineTo uint32, cursor uint32, time uint32) byte
	UngrabPointer(time uint32)
	GrabKeyboard(grabWindow xID, ownerEvents bool, time uint32, pointerMode, keyboardMode byte) byte
	UngrabKeyboard(time uint32)
	WarpPointer(x, y int16)
	GetCanvasOperations() []CanvasOperation
	GetRGBColor(colormap xID, pixel uint32) (r, g, b uint8)
	OpenFont(fid xID, name string)
	QueryFont(fid xID) (minBounds, maxBounds wire.XCharInfo, minCharOrByte2, maxCharOrByte2, defaultChar uint16, drawDirection uint8, minByte1, maxByte1 uint8, allCharsExist bool, fontAscent, fontDescent int16, charInfos []wire.XCharInfo)
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
	ChangeKeyboardControl(valueMask uint32, values wire.KeyboardControl)
	GetKeyboardControl() (wire.KeyboardControl, error)
	SetScreenSaver(timeout, interval int16, preferBlank, allowExpose byte)
	GetScreenSaver() (timeout, interval int16, preferBlank, allowExpose byte, err error)
	ChangeHosts(mode byte, host wire.Host)
	ListHosts() ([]wire.Host, error)
	SetAccessControl(mode byte)
	SetCloseDownMode(mode byte)
	KillClient(resource uint32)
	RotateProperties(window xID, delta int16, atoms []wire.Atom)
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
	selections            map[xID]uint32
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
	pointerGrabClientID   uint32
	keyboardGrabClientID  uint32
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
}

type passiveGrab struct {
	clientID  uint32
	button    byte
	key       wire.KeyCode
	modifiers uint16
	owner     bool
	eventMask uint16
	cursor    xID
}

type passiveDeviceGrab struct {
	deviceID  byte
	key       wire.KeyCode
	button    byte
	modifiers uint16
	owner     bool
	eventMask []uint32
}

type deviceGrab struct {
	window      xID
	ownerEvents bool
	eventMask   []uint32
	time        uint32
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

func (s *x11Server) UpdatePointerPosition(x, y int16) {
	s.pointerX = x
	s.pointerY = y
}

func (s *x11Server) resolveWindowID(localID uint32) (xID, bool) {
	for id := range s.windows {
		if id.local == localID {
			return id, true
		}
	}
	return xID{}, false
}

func (s *x11Server) GetWindowAttributes(xid xID) (wire.WindowAttributes, bool) {
	w, ok := s.windows[xid]
	if !ok {
		return wire.WindowAttributes{}, false
	}
	return w.attributes, true
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

	grabActive := s.pointerGrabWindow.local != 0

	if grabActive {
		xid = s.pointerGrabWindow
	}
	var eventMask uint32
	state := uint16(detail >> 16)

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

	// Unpack button and state from detail
	button := byte(detail & 0xFFFF)

	if !grabActive && eventType == "mousedown" {
		if grabs, ok := s.passiveGrabs[originalXID]; ok {
			for _, grab := range grabs {
				if grab.button == button && (grab.modifiers == wire.AnyModifier || grab.modifiers == state) {
					s.pointerGrabWindow = originalXID
					s.pointerGrabClientID = grab.clientID
					s.pointerGrabOwner = grab.owner
					s.pointerGrabEventMask = grab.eventMask
					grabActive = true
					s.frontend.SetWindowCursor(originalXID, grab.cursor)
					break
				}
			}
		}
	}

	eventWindowID := originalXID.local
	if grabActive && !s.pointerGrabOwner {
		eventWindowID = s.pointerGrabWindow.local
	}

	// Dispatch events.
	// The core event is subject to grabs, but XInput events are not.
	if grabActive {
		client, ok := s.clients[s.pointerGrabClientID]
		if ok && (uint32(s.pointerGrabEventMask)&eventMask) != 0 {
			s.sendCoreMouseEvent(client, eventType, button, eventWindowID, x, y, state)
		}
	} else {
		// If no grab is active, send core events to clients that have selected for them.
		for _, client := range s.clients {
			w, ok := s.windows[client.xID(xid.local)]
			if ok && w.attributes.EventMask&eventMask != 0 {
				s.sendCoreMouseEvent(client, eventType, button, eventWindowID, x, y, state)
			}
		}
	}

	// Send XInput events, respecting device grabs.
	var xiEventMask uint32
	switch eventType {
	case "mousedown":
		xiEventMask = wire.DeviceButtonPressMask
	case "mouseup":
		xiEventMask = wire.DeviceButtonReleaseMask
	}

	if xiEventMask > 0 {
		if grab, ok := s.deviceGrabs[virtualPointer.Header.DeviceID]; ok {
			// A device grab is active. Send event only to the grabbing client.
			grabbingClient, clientExists := s.clients[grab.window.client]
			if clientExists {
				if deviceInfo, ok := grabbingClient.openDevices[virtualPointer.Header.DeviceID]; ok {
					if mask, ok := deviceInfo.EventMasks[originalXID.local]; ok {
						if mask&xiEventMask != 0 {
							s.sendXInputMouseEvent(grabbingClient, eventType, virtualPointer.Header.DeviceID, button, originalXID.local, x, y, state)
						}
					}
				}
			}
		} else {
			// No device grab, send to all interested clients.
			for _, client := range s.clients {
				if deviceInfo, ok := client.openDevices[virtualPointer.Header.DeviceID]; ok {
					if mask, ok := deviceInfo.EventMasks[originalXID.local]; ok {
						if mask&xiEventMask != 0 {
							s.sendXInputMouseEvent(client, eventType, virtualPointer.Header.DeviceID, button, originalXID.local, x, y, state)
						}
					}
				}
			}
		}
	}
}

func (s *x11Server) sendCoreMouseEvent(client *x11Client, eventType string, button byte, eventWindowID uint32, x, y int32, state uint16) {
	var event messageEncoder
	switch eventType {
	case "mousedown":
		event = &wire.ButtonPressEvent{
			Sequence:   client.sequence - 1,
			Detail:     button,
			Time:       0, // 0 for now
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
			Time:       0, // 0 for now
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
			Time:       0, // 0 for now
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
			Time:       0, // Timestamp
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
			Time:       0, // Timestamp
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

	keycode, ok := jsCodeToX11Keycode[code]
	if !ok {
		keycode = jsCodeToX11Keycode["Unidentified"]
	}

	// Handle device grabs first
	if grab, ok := s.deviceGrabs[virtualKeyboard.Header.DeviceID]; ok {
		grabbingClient, clientExists := s.clients[grab.window.client]
		if clientExists {
			var xiEventMask uint32
			switch eventType {
			case "keydown":
				xiEventMask = wire.DeviceKeyPressMask
			case "keyup":
				xiEventMask = wire.DeviceKeyReleaseMask
			}
			var grabMask uint32
			for _, m := range grab.eventMask {
				grabMask |= m
			}
			if grabMask&xiEventMask != 0 {
				s.sendXInputKeyboardEvent(grabbingClient, eventType, keycode, s.inputFocus.local, state)
			}
		}
		// If a device grab is active, core events might be suppressed depending on grab parameters.
		// For simplicity, we suppress them.
		return
	}

	grabActive := s.keyboardGrabWindow.local != 0

	var eventMask uint32
	switch eventType {
	case "keydown":
		eventMask = wire.KeyPressMask
	case "keyup":
		eventMask = wire.KeyReleaseMask
	}

	// Handle passive grabs first
	if !grabActive && eventType == "keydown" {
		if grabs, ok := s.passiveGrabs[xid]; ok {
			for _, grab := range grabs {
				if grab.key == wire.KeyCode(keycode) && (grab.modifiers == wire.AnyModifier || grab.modifiers == state) {
					s.keyboardGrabWindow = xid
					s.keyboardGrabClientID = grab.clientID
					s.keyboardGrabOwner = grab.owner
					grabActive = true
					if client, ok := s.clients[grab.clientID]; ok {
						s.sendCoreKeyboardEvent(client, eventType, keycode, xid.local, state)
					}
					return
				}
			}
		}
	}

	if grabActive {
		client, ok := s.clients[s.keyboardGrabClientID]
		if !ok {
			return
		}

		eventWindow := s.keyboardGrabWindow.local
		if s.keyboardGrabOwner {
			eventWindow = xid.local
		}

		s.sendCoreKeyboardEvent(client, eventType, keycode, eventWindow, state)
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

	// Send XInput events.
	var xiEventMask uint32
	switch eventType {
	case "keydown":
		xiEventMask = wire.DeviceKeyPressMask
	case "keyup":
		xiEventMask = wire.DeviceKeyReleaseMask
	}

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
		Time:       0, // TODO: Get actual time
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
			Time:       0, // Timestamp
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
			Time:       0, // Timestamp
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
			Time:       0, // Timestamp (not implemented)
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
			Time:       0, // Timestamp (not implemented)
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
		Time:      0, // TODO: Get actual time
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
		client.send(wire.NewError(wire.LengthErrorCode, client.sequence, 0, 0, wire.ReqCode(header[0])))
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
			client.send(wire.NewError(wire.LengthErrorCode, client.sequence, 0, 0, wire.ReqCode(header[0])))
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
		attrs, ok := s.GetWindowAttributes(xid)
		if !ok {
			return wire.NewGenericError(seq, uint32(p.Window), 0, wire.GetWindowAttributes, wire.WindowErrorCode)
		}
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
		delete(s.windows, xid)
		s.frontend.DestroyWindow(xid)

	case *wire.DestroySubwindowsRequest:
		xid := client.xID(uint32(p.Window))
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
		if p.Mode == 0 { // Insert
			client.saveSet[uint32(p.Window)] = true
		} else { // Delete
			delete(client.saveSet, uint32(p.Window))
		}

	case *wire.ReparentWindowRequest:
		windowXID := client.xID(uint32(p.Window))
		parentXID := client.xID(uint32(p.Parent))
		window, ok := s.windows[windowXID]
		if !ok {
			return wire.NewGenericError(seq, uint32(p.Window), 0, wire.ReparentWindow, wire.WindowErrorCode)
		}
		oldParent, ok := s.windows[client.xID(window.parent)]
		if !ok {
			return wire.NewGenericError(seq, window.parent, 0, wire.ReparentWindow, wire.WindowErrorCode)
		}
		newParent, ok := s.windows[parentXID]
		if !ok {
			return wire.NewGenericError(seq, uint32(p.Parent), 0, wire.ReparentWindow, wire.WindowErrorCode)
		}

		// Remove from old parent's children
		for i, childID := range oldParent.children {
			if childID == window.xid.local {
				oldParent.children = append(oldParent.children[:i], oldParent.children[i+1:]...)
				break
			}
		}

		// Add to new parent's children
		newParent.children = append(newParent.children, window.xid.local)

		// Update window's state
		window.parent = uint32(p.Parent)
		window.x = p.X
		window.y = p.Y

		s.frontend.ReparentWindow(windowXID, parentXID, p.X, p.Y)

	case *wire.MapWindowRequest:
		xid := client.xID(uint32(p.Window))
		if w, ok := s.windows[xid]; ok {
			w.mapped = true
			s.frontend.MapWindow(xid)
			s.sendExposeEvent(xid, 0, 0, w.width, w.height)
		}

	case *wire.MapSubwindowsRequest:
		xid := client.xID(uint32(p.Window))
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
		if w, ok := s.windows[xid]; ok {
			w.mapped = false
		}
		s.frontend.UnmapWindow(xid)

	case *wire.UnmapSubwindowsRequest:
		xid := client.xID(uint32(p.Window))
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
		s.frontend.ConfigureWindow(xid, p.ValueMask, p.Values)

	case *wire.CirculateWindowRequest:
		xid := client.xID(uint32(p.Window))
		window, ok := s.windows[xid]
		if !ok {
			return wire.NewGenericError(seq, uint32(p.Window), 0, wire.CirculateWindow, wire.WindowErrorCode)
		}
		parent, ok := s.windows[client.xID(window.parent)]
		if !ok {
			return wire.NewGenericError(seq, window.parent, 0, wire.CirculateWindow, wire.WindowErrorCode)
		}

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

		s.frontend.CirculateWindow(xid, p.Direction)

	case *wire.GetGeometryRequest:
		xid := client.xID(uint32(p.Drawable))
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
		w, ok := s.windows[xid]
		if !ok {
			return nil
		}
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

	case *wire.QueryTreeRequest:
		xid := client.xID(uint32(p.Window))
		window, ok := s.windows[xid]
		if !ok {
			return wire.NewGenericError(seq, uint32(p.Window), 0, wire.QueryTree, wire.WindowErrorCode)
		}
		return &wire.QueryTreeReply{
			Sequence:    seq,
			Root:        s.rootWindowID(),
			Parent:      window.parent,
			NumChildren: uint16(len(window.children)),
			Children:    window.children,
		}

	case *wire.InternAtomRequest:
		atomID := s.frontend.GetAtom(client.id, p.Name)

		return &wire.InternAtomReply{
			Sequence: seq,
			Atom:     atomID,
		}

	case *wire.GetAtomNameRequest:
		name := s.frontend.GetAtomName(uint32(p.Atom))
		return &wire.GetAtomNameReply{
			Sequence:   seq,
			NameLength: uint16(len(name)),
			Name:       name,
		}

	case *wire.ChangePropertyRequest:
		xid := client.xID(uint32(p.Window))
		s.frontend.ChangeProperty(xid, uint32(p.Property), uint32(p.Type), uint32(p.Format), p.Data)

	case *wire.DeletePropertyRequest:
		xid := client.xID(uint32(p.Window))
		s.frontend.DeleteProperty(xid, uint32(p.Property))

	case *wire.GetPropertyRequest:
		xid := client.xID(uint32(p.Window))
		data, typ, format, bytesAfter := s.frontend.GetProperty(xid, uint32(p.Property), p.Offset, p.Length)

		if data == nil {
			return &wire.GetPropertyReply{
				Sequence: seq,
				Format:   0,
			}
		}

		var valueLenInFormatUnits uint32
		if format == 8 {
			valueLenInFormatUnits = uint32(len(data))
		} else if format == 16 {
			valueLenInFormatUnits = uint32(len(data) / 2)
		} else if format == 32 {
			valueLenInFormatUnits = uint32(len(data) / 4)
		}

		if p.Delete {
			s.frontend.DeleteProperty(xid, uint32(p.Property))
		}

		if p.Type != 0 && typ != uint32(p.Type) {
			// TODO: return empty property with the correct type
			// and format 0
		}

		return &wire.GetPropertyReply{
			Sequence:              seq,
			Format:                byte(format),
			PropertyType:          typ,
			BytesAfter:            bytesAfter,
			ValueLenInFormatUnits: valueLenInFormatUnits,
			Value:                 data,
		}

	case *wire.ListPropertiesRequest:
		xid := client.xID(uint32(p.Window))
		atoms := s.frontend.ListProperties(xid)
		return &wire.ListPropertiesReply{
			Sequence:      seq,
			NumProperties: uint16(len(atoms)),
			Atoms:         atoms,
		}

	case *wire.SetSelectionOwnerRequest:
		s.selections[client.xID(uint32(p.Selection))] = uint32(p.Owner)

	case *wire.GetSelectionOwnerRequest:
		owner := s.selections[client.xID(uint32(p.Selection))]
		return &wire.GetSelectionOwnerReply{
			Sequence: seq,
			Owner:    owner,
		}

	case *wire.ConvertSelectionRequest:
		s.frontend.ConvertSelection(uint32(p.Selection), uint32(p.Target), uint32(p.Property), client.xID(uint32(p.Requestor)))

	case *wire.SendEventRequest:
		// The X11 client sends an event to another client.
		// We need to forward this event to the appropriate frontend.
		// For now, we'll just log it and pass it to the frontend.
		s.frontend.SendEvent(&wire.X11RawEvent{Data: p.EventData})

	case *wire.GrabPointerRequest:
		grabWindow := client.xID(uint32(p.GrabWindow))
		if s.pointerGrabWindow.local != 0 {
			return &wire.GrabPointerReply{
				Sequence: seq,
				Status:   wire.AlreadyGrabbed,
			}
		}

		s.pointerGrabWindow = grabWindow
		s.pointerGrabOwner = p.OwnerEvents
		s.pointerGrabEventMask = p.EventMask
		s.pointerGrabTime = uint32(p.Time)

		return &wire.GrabPointerReply{
			Sequence: seq,
			Status:   wire.GrabSuccess,
		}

	case *wire.UngrabPointerRequest:
		s.pointerGrabWindow = xID{}
		s.pointerGrabOwner = false
		s.pointerGrabEventMask = 0
		s.pointerGrabTime = 0

	case *wire.GrabButtonRequest:
		grabWindow, ok := s.resolveWindowID(uint32(p.GrabWindow))
		if !ok {
			return wire.NewGenericError(seq, uint32(p.GrabWindow), 0, wire.GrabButton, wire.WindowErrorCode)
		}
		grab := &passiveGrab{
			clientID:  client.id,
			button:    p.Button,
			modifiers: p.Modifiers,
			owner:     p.OwnerEvents,
			eventMask: p.EventMask,
			cursor:    client.xID(uint32(p.Cursor)),
		}
		s.passiveGrabs[grabWindow] = append(s.passiveGrabs[grabWindow], grab)

	case *wire.UngrabButtonRequest:
		grabWindow := client.xID(uint32(p.GrabWindow))
		if grabs, ok := s.passiveGrabs[grabWindow]; ok {
			for i, grab := range grabs {
				if grab.button == p.Button && grab.modifiers == p.Modifiers {
					s.passiveGrabs[grabWindow] = append(grabs[:i], grabs[i+1:]...)
					break
				}
			}
		}

	case *wire.ChangeActivePointerGrabRequest:
		// TODO: implement

	case *wire.GrabKeyboardRequest:
		grabWindow := client.xID(uint32(p.GrabWindow))
		if s.keyboardGrabWindow.local != 0 {
			return &wire.GrabKeyboardReply{
				Sequence: seq,
				Status:   wire.AlreadyGrabbed,
			}
		}

		s.keyboardGrabWindow = grabWindow
		s.keyboardGrabOwner = p.OwnerEvents
		s.keyboardGrabTime = uint32(p.Time)

		return &wire.GrabKeyboardReply{
			Sequence: seq,
			Status:   wire.GrabSuccess,
		}

	case *wire.UngrabKeyboardRequest:
		s.keyboardGrabWindow = xID{}
		s.keyboardGrabOwner = false
		s.keyboardGrabTime = 0

	case *wire.GrabKeyRequest:
		grabWindow, ok := s.resolveWindowID(uint32(p.GrabWindow))
		if !ok {
			return wire.NewGenericError(seq, uint32(p.GrabWindow), 0, wire.GrabKey, wire.WindowErrorCode)
		}
		grab := &passiveGrab{
			clientID:  client.id,
			key:       p.Key,
			modifiers: p.Modifiers,
			owner:     p.OwnerEvents,
		}
		s.passiveGrabs[grabWindow] = append(s.passiveGrabs[grabWindow], grab)

	case *wire.UngrabKeyRequest:
		grabWindow := client.xID(uint32(p.GrabWindow))
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

		// Simplified implementation: assume windows are direct children of the root
		src, srcOk := s.windows[srcWindow]
		dst, dstOk := s.windows[dstWindow]
		if !srcOk || !dstOk {
			// One of the windows doesn't exist, can't translate
			return nil
		}

		dstX := src.x + p.SrcX - dst.x
		dstY := src.y + p.SrcY - dst.y

		return &wire.TranslateCoordsReply{
			Sequence:   seq,
			SameScreen: true,
			Child:      0, // No child for now
			DstX:       dstX,
			DstY:       dstY,
		}

	case *wire.WarpPointerRequest:
		s.frontend.WarpPointer(p.DstX, p.DstY)

	case *wire.SetInputFocusRequest:
		xid := client.xID(uint32(p.Focus))
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
		s.frontend.OpenFont(client.xID(uint32(p.Fid)), p.Name)

	case *wire.CloseFontRequest:
		s.frontend.CloseFont(client.xID(uint32(p.Fid)))

	case *wire.QueryFontRequest:
		minBounds, maxBounds, minCharOrByte2, maxCharOrByte2, defaultChar, drawDirection, minByte1, maxByte1, allCharsExist, fontAscent, fontDescent, charInfos := s.frontend.QueryFont(client.xID(uint32(p.Fid)))

		return &wire.QueryFontReply{
			Sequence:       seq,
			MinBounds:      minBounds,
			MaxBounds:      maxBounds,
			MinCharOrByte2: minCharOrByte2,
			MaxCharOrByte2: maxCharOrByte2,
			DefaultChar:    defaultChar,
			NumFontProps:   0, // Not implemented yet
			DrawDirection:  drawDirection,
			MinByte1:       minByte1,
			MaxByte1:       maxByte1,
			AllCharsExist:  allCharsExist,
			FontAscent:     fontAscent,
			FontDescent:    fontDescent,
			NumCharInfos:   uint32(len(charInfos)),
			CharInfos:      charInfos,
		}

	case *wire.QueryTextExtentsRequest:
		drawDirection, fontAscent, fontDescent, overallAscent, overallDescent, overallWidth, overallLeft, overallRight := s.frontend.QueryTextExtents(client.xID(uint32(p.Fid)), p.Text)
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
		return &wire.ListFontsWithInfoReply{
			Sequence: seq,
			FontName: "",
		}

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

		s.pixmaps[xid] = true // Mark pixmap ID as used
		s.frontend.CreatePixmap(xid, client.xID(uint32(p.Drawable)), uint32(p.Width), uint32(p.Height), uint32(p.Depth))

	case *wire.FreePixmapRequest:
		xid := client.xID(uint32(p.Pid))
		delete(s.pixmaps, xid)
		s.frontend.FreePixmap(xid)

	case *wire.CreateGCRequest:
		xid := client.xID(uint32(p.Cid))

		// Check if the GC ID is already in use
		if _, exists := s.gcs[xid]; exists {
			s.logger.Errorf("X11: CreateGC: ID %s already in use", xid)
			return wire.NewGenericError(seq, uint32(xid.local), 0, wire.CreateGC, wire.IDChoiceErrorCode)
		}

		s.gcs[xid] = p.Values
		s.frontend.CreateGC(xid, p.ValueMask, p.Values)

	case *wire.ChangeGCRequest:
		xid := client.xID(uint32(p.Gc))
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
		s.frontend.CopyGC(srcGC, dstGC)

	case *wire.SetDashesRequest:
		s.frontend.SetDashes(client.xID(uint32(p.GC)), p.DashOffset, p.Dashes)

	case *wire.SetClipRectanglesRequest:
		s.frontend.SetClipRectangles(client.xID(uint32(p.GC)), p.ClippingX, p.ClippingY, p.Rectangles, p.Ordering)

	case *wire.FreeGCRequest:
		gcID := client.xID(uint32(p.GC))
		s.frontend.FreeGC(gcID)

	case *wire.ClearAreaRequest:
		s.frontend.ClearArea(client.xID(uint32(p.Window)), int32(p.X), int32(p.Y), int32(p.Width), int32(p.Height))

	case *wire.CopyAreaRequest:
		gcID := client.xID(uint32(p.Gc))
		s.frontend.CopyArea(client.xID(uint32(p.SrcDrawable)), client.xID(uint32(p.DstDrawable)), gcID, int32(p.SrcX), int32(p.SrcY), int32(p.DstX), int32(p.DstY), int32(p.Width), int32(p.Height))

	case *wire.CopyPlaneRequest:
		gcID := client.xID(uint32(p.Gc))
		s.frontend.CopyPlane(client.xID(uint32(p.SrcDrawable)), client.xID(uint32(p.DstDrawable)), gcID, int32(p.SrcX), int32(p.SrcY), int32(p.DstX), int32(p.DstY), int32(p.Width), int32(p.Height), int32(p.PlaneMask))

	case *wire.PolyPointRequest:
		gcID := client.xID(uint32(p.Gc))
		s.frontend.PolyPoint(client.xID(uint32(p.Drawable)), gcID, p.Coordinates)

	case *wire.PolyLineRequest:
		gcID := client.xID(uint32(p.Gc))
		s.frontend.PolyLine(client.xID(uint32(p.Drawable)), gcID, p.Coordinates)

	case *wire.PolySegmentRequest:
		gcID := client.xID(uint32(p.Gc))
		s.frontend.PolySegment(client.xID(uint32(p.Drawable)), gcID, p.Segments)

	case *wire.PolyRectangleRequest:
		gcID := client.xID(uint32(p.Gc))
		s.frontend.PolyRectangle(client.xID(uint32(p.Drawable)), gcID, p.Rectangles)

	case *wire.PolyArcRequest:
		gcID := client.xID(uint32(p.Gc))
		s.frontend.PolyArc(client.xID(uint32(p.Drawable)), gcID, p.Arcs)

	case *wire.FillPolyRequest:
		gcID := client.xID(uint32(p.Gc))
		s.frontend.FillPoly(client.xID(uint32(p.Drawable)), gcID, p.Coordinates)

	case *wire.PolyFillRectangleRequest:
		gcID := client.xID(uint32(p.Gc))
		s.frontend.PolyFillRectangle(client.xID(uint32(p.Drawable)), gcID, p.Rectangles)

	case *wire.PolyFillArcRequest:
		gcID := client.xID(uint32(p.Gc))
		s.frontend.PolyFillArc(client.xID(uint32(p.Drawable)), gcID, p.Arcs)

	case *wire.PutImageRequest:
		gcID := client.xID(uint32(p.Gc))
		s.frontend.PutImage(client.xID(uint32(p.Drawable)), gcID, p.Format, p.Width, p.Height, p.DstX, p.DstY, p.LeftPad, p.Depth, p.Data)

	case *wire.GetImageRequest:
		imgData, err := s.frontend.GetImage(client.xID(uint32(p.Drawable)), int32(p.X), int32(p.Y), int32(p.Width), int32(p.Height), p.PlaneMask)
		if err != nil {
			s.logger.Errorf("Failed to get image: %v", err)
			return nil
		}
		return &wire.GetImageReply{
			Sequence:  seq,
			Depth:     24, // Assuming 24-bit depth for now
			VisualID:  s.visualID,
			ImageData: imgData,
		}

	case *wire.PolyText8Request:
		gcID := client.xID(uint32(p.GC))
		s.frontend.PolyText8(client.xID(uint32(p.Drawable)), gcID, int32(p.X), int32(p.Y), p.Items)

	case *wire.PolyText16Request:
		gcID := client.xID(uint32(p.GC))
		s.frontend.PolyText16(client.xID(uint32(p.Drawable)), gcID, int32(p.X), int32(p.Y), p.Items)

	case *wire.ImageText8Request:
		gcID := client.xID(uint32(p.Gc))
		s.frontend.ImageText8(client.xID(uint32(p.Drawable)), gcID, int32(p.X), int32(p.Y), p.Text)

	case *wire.ImageText16Request:
		gcID := client.xID(uint32(p.Gc))
		s.frontend.ImageText16(client.xID(uint32(p.Drawable)), gcID, int32(p.X), int32(p.Y), p.Text)

	case *wire.CreateColormapRequest:
		xid := client.xID(uint32(p.Mid))

		if _, exists := s.colormaps[xid]; exists {
			return wire.NewGenericError(seq, uint32(p.Mid), 0, wire.CreateColormap, wire.ColormapErrorCode)
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
		if _, ok := s.colormaps[xid]; !ok {
			return wire.NewGenericError(seq, uint32(p.Cmap), 0, wire.FreeColormap, wire.ColormapErrorCode)
		}
		delete(s.colormaps, xid)

	case *wire.CopyColormapAndFreeRequest:
		srcCmapID := client.xID(uint32(p.SrcCmap))
		srcCmap, ok := s.colormaps[srcCmapID]
		if !ok {
			return wire.NewGenericError(seq, uint32(p.SrcCmap), 0, wire.CopyColormapAndFree, wire.ColormapErrorCode)
		}

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
		_, ok := s.colormaps[xid]
		if !ok {
			return wire.NewGenericError(seq, uint32(p.Cmap), 0, wire.InstallColormap, wire.ColormapErrorCode)
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
		_, ok := s.colormaps[xid]
		if !ok {
			return wire.NewGenericError(seq, uint32(p.Cmap), 0, wire.UninstallColormap, wire.ColormapErrorCode)
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
			return wire.NewError(wire.ColormapErrorCode, seq, uint32(p.Cmap), 0, p.OpCode())
		}

		name := string(p.Name)
		rgb, ok := lookupColor(name)
		if !ok {
			return wire.NewError(wire.NameErrorCode, seq, 0, 0, p.OpCode())
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
		return &wire.AllocColorCellsReply{
			Sequence: seq,
		}

	case *wire.AllocColorPlanesRequest:
		return &wire.AllocColorPlanesReply{
			Sequence: seq,
		}

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

		s.cursors[cursorXID] = true
		sourceXID := client.xID(uint32(p.Source))
		maskXID := client.xID(uint32(p.Mask))
		foreColor := [3]uint16{p.ForeRed, p.ForeGreen, p.ForeBlue}
		backColor := [3]uint16{p.BackRed, p.BackGreen, p.BackBlue}
		s.frontend.CreateCursor(cursorXID, sourceXID, maskXID, foreColor, backColor, p.X, p.Y)

	case *wire.CreateGlyphCursorRequest:
		// Check if the cursor ID is already in use
		if _, exists := s.cursors[client.xID(uint32(p.Cid))]; exists {
			s.logger.Errorf("X11: CreateGlyphCursor: ID %d already in use", p.Cid)
			return wire.NewGenericError(seq, uint32(p.Cid), 0, wire.CreateGlyphCursor, wire.IDChoiceErrorCode)
		}

		s.cursors[client.xID(uint32(p.Cid))] = true
		s.frontend.CreateCursorFromGlyph(uint32(p.Cid), p.SourceChar)

	case *wire.FreeCursorRequest:
		xid := client.xID(uint32(p.Cursor))
		delete(s.cursors, xid)
		s.frontend.FreeCursor(xid)

	case *wire.RecolorCursorRequest:
		s.frontend.RecolorCursor(client.xID(uint32(p.Cursor)), p.ForeColor, p.BackColor)

	case *wire.QueryBestSizeRequest:
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
		debugf("X11: ChangePointerControlRequest not implemented")
		// TODO: Implement

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
		s.frontend.RotateProperties(client.xID(uint32(p.Window)), p.Delta, p.Atoms)

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
			// TODO: proper error handling
			return nil
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
					selections: make(map[xID]uint32),
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
				}
				for k, v := range KeyCodeToKeysym {
					x11ServerInstance.keymap[k] = v
				}
				x11ServerInstance.frontend = newX11Frontend(logger, x11ServerInstance)
			})

			client := &x11Client{
				id:          x11ServerInstance.nextClientID,
				conn:        channel,
				sequence:    0,
				byteOrder:   binary.LittleEndian, // Default, will be updated in handshake
				saveSet:     make(map[uint32]bool),
				openDevices: make(map[byte]*wire.DeviceInfo),
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
