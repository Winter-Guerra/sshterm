//go:build x11

package x11

import (
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"runtime/debug"
	"sync"

	"golang.org/x/crypto/ssh"
)

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
	CreateWindow(xid xID, parent, x, y, width, height, depth, valueMask uint32, values WindowAttributes)
	ChangeWindowAttributes(xid xID, valueMask uint32, values WindowAttributes)
	GetWindowAttributes(xid xID) WindowAttributes
	ChangeProperty(xid xID, property, typeAtom, format uint32, data []byte)
	CreateGC(xid xID, valueMask uint32, values GC)
	ChangeGC(xid xID, valueMask uint32, gc GC)
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
	ImageText8(drawable xID, gcID xID, x, y int32, text []byte)
	ImageText16(drawable xID, gcID xID, x, y int32, text []uint16)
	PolyText8(drawable xID, gcID xID, x, y int32, items []PolyTextItem)
	PolyText16(drawable xID, gcID xID, x, y int32, items []PolyTextItem)
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
	GetCanvasOperations() []CanvasOperation
	GetRGBColor(colormap xID, pixel uint32) (r, g, b uint8)
	OpenFont(fid xID, name string)
	QueryFont(fid xID) (minBounds, maxBounds xCharInfo, minCharOrByte2, maxCharOrByte2, defaultChar uint16, drawDirection uint8, minByte1, maxByte1 uint8, allCharsExist bool, fontAscent, fontDescent int16, charInfos []xCharInfo)
	QueryTextExtents(font xID, text []uint16) (drawDirection uint8, fontAscent, fontDescent, overallAscent, overallDescent, overallWidth, overallLeft, overallRight int16)
	CloseFont(fid xID)
	ListFonts(maxNames uint16, pattern string) []string
	AllowEvents(clientID uint32, mode byte, time uint32)
	SetDashes(gc xID, dashOffset uint16, dashes []byte)
	SetClipRectangles(gc xID, clippingX, clippingY int16, rectangles []Rectangle, ordering byte)
	RecolorCursor(cursor xID, foreColor, backColor [3]uint16)
	SetPointerMapping(pMap []byte) (byte, error)
	GetPointerMapping() ([]byte, error)
	GetKeyboardMapping(firstKeyCode KeyCode, count byte) ([]uint32, error)
	ChangeKeyboardMapping(keyCodeCount byte, firstKeyCode KeyCode, keySymsPerKeyCode byte, keySyms []uint32)
	ChangeKeyboardControl(valueMask uint32, values KeyboardControl)
	GetKeyboardControl() (KeyboardControl, error)
	SetScreenSaver(timeout, interval int16, preferBlank, allowExpose byte)
	GetScreenSaver() (timeout, interval int16, preferBlank, allowExpose byte, err error)
	ChangeHosts(mode byte, host Host)
	ListHosts() ([]Host, error)
	SetAccessControl(mode byte)
	SetCloseDownMode(mode byte)
	KillClient(resource uint32)
	RotateProperties(window xID, delta int16, atoms []Atom)
	ForceScreenSaver(mode byte)
	SetModifierMapping(keyCodesPerModifier byte, keyCodes []KeyCode) (byte, error)
	GetModifierMapping() (keyCodesPerModifier byte, keyCodes []KeyCode, err error)
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
	xid           xID
	parent        uint32
	x, y          int16
	width, height uint16
	mapped        bool
	depth         byte
	children      []uint32
	attributes    WindowAttributes
	colormap      xID
}

func (w *window) mapState() byte {
	if !w.mapped {
		return 0 // Unmapped
	}
	return 2 // Viewable
}

type colormap struct {
	pixels map[uint32]xColorItem
}

type x11Server struct {
	logger               Logger
	byteOrder            binary.ByteOrder
	frontend             X11FrontendAPI
	windows              map[xID]*window
	gcs                  map[xID]GC
	pixmaps              map[xID]bool
	cursors              map[xID]bool
	selections           map[xID]uint32
	colormaps            map[xID]*colormap
	defaultColormap      uint32
	installedColormap    xID
	visualID             uint32
	rootVisual           visualType
	rootWindowWidth      uint16
	rootWindowHeight     uint16
	blackPixel           uint32
	whitePixel           uint32
	pointerX, pointerY   int16
	clients              map[uint32]*x11Client
	nextClientID         uint32
	pointerGrabWindow    xID
	keyboardGrabWindow   xID
	pointerGrabTime      uint32
	keyboardGrabTime     uint32
	pointerGrabOwner     bool
	keyboardGrabOwner    bool
	pointerGrabEventMask uint16
	passiveGrabs         map[xID][]*passiveGrab
	authProtocol         string
	authCookie           []byte
}

type passiveGrab struct {
	button    byte
	key       KeyCode
	modifiers uint16
	owner     bool
	eventMask uint16
	cursor    xID
}

func (s *x11Server) SetRootWindowSize(width, height uint16) {
	s.rootWindowWidth = width
	s.rootWindowHeight = height
}

func (s *x11Server) UpdatePointerPosition(x, y int16) {
	s.pointerX = x
	s.pointerY = y
}

func (s *x11Server) GetWindowAttributes(xid xID) (WindowAttributes, bool) {
	w, ok := s.windows[xid]
	if !ok {
		return WindowAttributes{}, false
	}
	return w.attributes, true
}

func (s *x11Server) SendMouseEvent(xid xID, eventType string, x, y, detail int32) {
	debugf("X11: SendMouseEvent xid=%s type=%s x=%d y=%d detail=%d", xid, eventType, x, y, detail)

	originalXID := xid
	grabActive := s.pointerGrabWindow.local != 0

	if grabActive {
		xid = s.pointerGrabWindow
	}
	var eventMask uint32
	state := uint16(detail >> 16)

	switch eventType {
	case "mousedown":
		eventMask = ButtonPressMask
	case "mouseup":
		eventMask = ButtonReleaseMask
	case "mousemove":
		eventMask = PointerMotionMask
		if state&Button1Mask != 0 {
			eventMask |= Button1MotionMask
		}
		if state&Button2Mask != 0 {
			eventMask |= Button2MotionMask
		}
		if state&Button3Mask != 0 {
			eventMask |= Button3MotionMask
		}
		if state&Button4Mask != 0 {
			eventMask |= Button4MotionMask
		}
		if state&Button5Mask != 0 {
			eventMask |= Button5MotionMask
		}
		if state&uint16(ButtonPressMask|ButtonReleaseMask) != 0 {
			eventMask |= ButtonMotionMask
		}
	}

	if grabActive {
		if (uint32(s.pointerGrabEventMask) & eventMask) == 0 {
			return
		}
	} else {
		w, ok := s.windows[xid]
		if !ok {
			log.Printf("X11: Failed to write mount event: window not found")
			return
		}
		if w.attributes.EventMask&eventMask == 0 {
			return
		}
	}

	client, ok := s.clients[xid.client]
	if !ok {
		log.Print("X11: Failed to write mount event: client not found")
		return
	}

	eventWindowID := originalXID.local
	if grabActive && !s.pointerGrabOwner {
		eventWindowID = s.pointerGrabWindow.local
	}

	// Unpack button and state from detail
	button := byte(detail & 0xFFFF)

	if !grabActive && eventType == "mousedown" {
		if grabs, ok := s.passiveGrabs[originalXID]; ok {
			for _, grab := range grabs {
				if grab.button == button && (grab.modifiers == AnyModifier || grab.modifiers == state) {
					s.pointerGrabWindow = originalXID
					s.pointerGrabOwner = grab.owner
					s.pointerGrabEventMask = grab.eventMask
					grabActive = true
					s.frontend.SetWindowCursor(originalXID, grab.cursor)
					break
				}
			}
		}
	}

	var event messageEncoder
	switch eventType {
	case "mousedown":
		event = &ButtonPressEvent{
			sequence:   client.sequence - 1,
			detail:     button,
			time:       0, // 0 for now
			root:       s.rootWindowID(),
			event:      eventWindowID,
			child:      0, // 0 for now
			rootX:      int16(x),
			rootY:      int16(y),
			eventX:     int16(x),
			eventY:     int16(y),
			state:      state,
			sameScreen: true,
		}
	case "mouseup":
		event = &ButtonReleaseEvent{
			sequence:   client.sequence - 1,
			detail:     button,
			time:       0, // 0 for now
			root:       s.rootWindowID(),
			event:      eventWindowID,
			child:      0, // 0 for now
			rootX:      int16(x),
			rootY:      int16(y),
			eventX:     int16(x),
			eventY:     int16(y),
			state:      state,
			sameScreen: true,
		}
	case "mousemove":
		event = &motionNotifyEvent{
			sequence:   client.sequence - 1,
			detail:     0, // 0 for Normal
			time:       0, // 0 for now
			root:       s.rootWindowID(),
			event:      eventWindowID,
			child:      0, // 0 for now
			rootX:      int16(x),
			rootY:      int16(y),
			eventX:     int16(x),
			eventY:     int16(y),
			state:      state,
			sameScreen: true,
		}
	default:
		debugf("X11: Unknown mouse event type: %s", eventType)
		return
	}

	if err := client.send(event); err != nil {
		debugf("X11: Failed to write mouse event: %v", err)
	}
}

func (s *x11Server) SendKeyboardEvent(xid xID, eventType string, code string, altKey, ctrlKey, shiftKey, metaKey bool) {
	// Implement sending keyboard event to client
	// This will involve constructing an X11 event packet and writing it to client.conn
	debugf("X11: SendKeyboardEvent xid=%s type=%s code=%s alt=%t ctrl=%t shift=%t meta=%t", xid, eventType, code, altKey, ctrlKey, shiftKey, metaKey)

	originalXID := xid
	grabActive := s.keyboardGrabWindow.local != 0

	if grabActive {
		xid = s.keyboardGrabWindow
	}

	var eventMask uint32
	if eventType == "keydown" {
		eventMask = KeyPressMask
	} else if eventType == "keyup" {
		eventMask = KeyReleaseMask
	}

	w, ok := s.windows[xid]
	if !ok {
		debugf("X11: SendKeyboardEvent unknown window %d", xid)
		return
	}
	if !grabActive && w.attributes.EventMask&eventMask == 0 {
		return
	}

	client, ok := s.clients[xid.client]
	if !ok {
		debugf("X11: SendKeyboardEvent unknown client %d", xid.client)
		return
	}

	state := uint16(0)
	if shiftKey {
		state |= 1 // ShiftMask
	}
	if ctrlKey {
		state |= 4 // ControlMask
	}
	if altKey {
		state |= 8 // Mod1Mask (Alt key)
	}
	if metaKey {
		state |= 64 // Mod4Mask (Meta key)
	}

	keycode, ok := jsCodeToX11Keycode[code]
	if !ok {
		keycode = jsCodeToX11Keycode["Unidentified"]
	}

	eventWindowID := originalXID.local
	if grabActive && !s.keyboardGrabOwner {
		eventWindowID = s.keyboardGrabWindow.local
	}

	if !grabActive && eventType == "keydown" {
		if grabs, ok := s.passiveGrabs[originalXID]; ok {
			for _, grab := range grabs {
				if grab.key == KeyCode(keycode) && (grab.modifiers == AnyModifier || grab.modifiers == state) {
					s.keyboardGrabWindow = originalXID
					s.keyboardGrabOwner = grab.owner
					grabActive = true
					break
				}
			}
		}
	}

	event := &keyEvent{
		sequence:   client.sequence - 1,
		detail:     keycode,
		time:       0, // TODO: Get actual time
		root:       s.rootWindowID(),
		event:      eventWindowID,
		child:      0, // No child for now
		rootX:      s.pointerX,
		rootY:      s.pointerY,
		eventX:     s.pointerX, // Assuming pointer is always in the window for now
		eventY:     s.pointerY, // Assuming pointer is always in the window for now
		state:      state,
		sameScreen: true,
	}

	if eventType == "keydown" {
		event.encodeMessage(client.byteOrder)[0] = 2 // KeyPress
	} else if eventType == "keyup" {
		event.encodeMessage(client.byteOrder)[0] = 3 // KeyRelease
	} else {
		debugf("X11: Unknown keyboard event type: %s", eventType)
		return
	}

	if err := client.send(event); err != nil {
		debugf("X11: Failed to write keyboard event: %v", err)
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
		event = &EnterNotifyEvent{
			sequence:   client.sequence - 1,
			detail:     detail,
			time:       0, // Timestamp (not implemented)
			root:       s.rootWindowID(),
			event:      xid.local,
			child:      0, // Or a child window ID if applicable
			rootX:      rootX,
			rootY:      rootY,
			eventX:     eventX,
			eventY:     eventY,
			state:      state,
			mode:       mode,
			sameScreen: true,
			focus:      s.frontend.GetFocusWindow(client.id) == xid,
		}
	} else {
		event = &LeaveNotifyEvent{
			sequence:   client.sequence - 1,
			detail:     detail,
			time:       0, // Timestamp (not implemented)
			root:       s.rootWindowID(),
			event:      xid.local,
			child:      0, // Or a child window ID if applicable
			rootX:      rootX,
			rootY:      rootY,
			eventX:     eventX,
			eventY:     eventY,
			state:      state,
			mode:       mode,
			sameScreen: true,
			focus:      s.frontend.GetFocusWindow(client.id) == xid,
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

	event := &configureNotifyEvent{
		sequence:         client.sequence - 1,
		event:            windowID.local,
		window:           windowID.local,
		aboveSibling:     0, // None
		x:                x,
		y:                y,
		width:            width,
		height:           height,
		borderWidth:      0,
		overrideRedirect: false,
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

	event := &exposeEvent{
		sequence: client.sequence - 1,
		window:   windowID.local,
		x:        x,
		y:        y,
		width:    width,
		height:   height,
		count:    0, // count = 0, no more expose events to follow
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

	event := &clientMessageEvent{
		sequence:    client.sequence - 1,
		format:      32, // Format is always 32 for ClientMessage
		window:      windowID.local,
		messageType: messageTypeAtom,
		data:        data,
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

	event := &selectionNotifyEvent{
		sequence:  client.sequence - 1,
		requestor: requestor.local,
		selection: selection,
		target:    target,
		property:  property,
		time:      0, // TODO: Get actual time
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
	if s.rootVisual.class == 4 { // TrueColor
		r = uint8((pixel & s.rootVisual.redMask) >> calculateShift(s.rootVisual.redMask))
		g = uint8((pixel & s.rootVisual.greenMask) >> calculateShift(s.rootVisual.greenMask))
		b = uint8((pixel & s.rootVisual.blueMask) >> calculateShift(s.rootVisual.blueMask))
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

func (s *x11Server) readRequest(client *x11Client) (request, uint16, error) {
	client.sequence++
	var header [4]byte
	if _, err := io.ReadFull(client.conn, header[:]); err != nil {
		return nil, 0, err
	}
	length := client.byteOrder.Uint16(header[2:4])
	if length == 0 {
		client.send(NewError(LengthErrorCode, client.sequence, 0, 0, reqCode(header[0])))
		return nil, 0, errParseError
	}
	raw := make([]byte, 4*length)
	copy(raw, header[:])
	if length > 1 {
		if _, err := io.ReadFull(client.conn, raw[4:]); err != nil {
			return nil, 0, err
		}
	}
	debugf("X11DEBUG: RAW Request: %x", raw)
	req, err := parseRequest(client.byteOrder, raw, client.sequence)
	if err != nil {
		if x11Err, ok := err.(Error); ok {
			client.send(x11Err)
		} else {
			client.send(NewError(LengthErrorCode, client.sequence, 0, 0, reqCode(header[0])))
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

func (s *x11Server) handleRequest(client *x11Client, req request, seq uint16) (reply messageEncoder) {
	debugf("X11DEBUG: handleRequest(%d) opcode: %d: %#v", seq, req.OpCode(), req)
	switch p := req.(type) {
	case *CreateWindowRequest:
		xid := client.xID(uint32(p.Drawable))
		parentXID := client.xID(uint32(p.Parent))
		// Check if the window ID is already in use
		if _, exists := s.windows[xid]; exists {
			s.logger.Errorf("X11: CreateWindow: ID %d already in use", xid)
			return client.sendError(&GenericError{seq: seq, badValue: uint32(p.Drawable), majorOp: CreateWindow, code: IDChoiceErrorCode})
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

	case *ChangeWindowAttributesRequest:
		xid := client.xID(uint32(p.Window))
		if w, ok := s.windows[xid]; ok {
			if p.ValueMask&CWBackPixmap != 0 {
				w.attributes.BackgroundPixmap = p.Values.BackgroundPixmap
			}
			if p.ValueMask&CWBackPixel != 0 {
				w.attributes.BackgroundPixel = p.Values.BackgroundPixel
			}
			if p.ValueMask&CWBorderPixmap != 0 {
				w.attributes.BorderPixmap = p.Values.BorderPixmap
			}
			if p.ValueMask&CWBorderPixel != 0 {
				w.attributes.BorderPixel = p.Values.BorderPixel
			}
			if p.ValueMask&CWBitGravity != 0 {
				w.attributes.BitGravity = p.Values.BitGravity
			}
			if p.ValueMask&CWWinGravity != 0 {
				w.attributes.WinGravity = p.Values.WinGravity
			}
			if p.ValueMask&CWBackingStore != 0 {
				w.attributes.BackingStore = p.Values.BackingStore
			}
			if p.ValueMask&CWBackingPlanes != 0 {
				w.attributes.BackingPlanes = p.Values.BackingPlanes
			}
			if p.ValueMask&CWBackingPixel != 0 {
				w.attributes.BackingPixel = p.Values.BackingPixel
			}
			if p.ValueMask&CWOverrideRedirect != 0 {
				w.attributes.OverrideRedirect = p.Values.OverrideRedirect
			}
			if p.ValueMask&CWSaveUnder != 0 {
				w.attributes.SaveUnder = p.Values.SaveUnder
			}
			if p.ValueMask&CWEventMask != 0 {
				w.attributes.EventMask = p.Values.EventMask
			}
			if p.ValueMask&CWDontPropagate != 0 {
				w.attributes.DontPropagateMask = p.Values.DontPropagateMask
			}
			if p.ValueMask&CWColormap != 0 {
				w.attributes.Colormap = p.Values.Colormap
			}
			if p.ValueMask&CWCursor != 0 {
				w.attributes.Cursor = p.Values.Cursor
				s.frontend.SetWindowCursor(xid, client.xID(uint32(p.Values.Cursor)))
			}
		}
		s.frontend.ChangeWindowAttributes(xid, p.ValueMask, p.Values)

	case *GetWindowAttributesRequest:
		xid := client.xID(uint32(p.Window))
		attrs, ok := s.GetWindowAttributes(xid)
		if !ok {
			return client.sendError(&GenericError{seq: seq, badValue: uint32(p.Window), majorOp: GetWindowAttributes, code: WindowErrorCode})
		}
		return &getWindowAttributesReply{
			Sequence:           seq,
			BackingStore:       byte(attrs.BackingStore),
			VisualID:           s.visualID,
			Class:              uint16(attrs.Class),
			BitGravity:         byte(attrs.BitGravity),
			WinGravity:         byte(attrs.WinGravity),
			BackingPlanes:      attrs.BackingPlanes,
			BackingPixel:       attrs.BackingPixel,
			SaveUnder:          boolToByte(attrs.SaveUnder),
			MapIsInstalled:     boolToByte(attrs.MapIsInstalled),
			MapState:           byte(attrs.MapState),
			OverrideRedirect:   boolToByte(attrs.OverrideRedirect),
			Colormap:           uint32(attrs.Colormap),
			AllEventMasks:      attrs.EventMask,
			YourEventMask:      attrs.EventMask,
			DoNotPropagateMask: uint16(attrs.DontPropagateMask),
		}

	case *DestroyWindowRequest:
		xid := client.xID(uint32(p.Window))
		delete(s.windows, xid)
		s.frontend.DestroyWindow(xid)

	case *DestroySubwindowsRequest:
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

	case *ChangeSaveSetRequest:
		// TODO: implement

	case *ReparentWindowRequest:
		windowXID := client.xID(uint32(p.Window))
		parentXID := client.xID(uint32(p.Parent))
		window, ok := s.windows[windowXID]
		if !ok {
			return client.sendError(&GenericError{seq: seq, badValue: uint32(p.Window), majorOp: ReparentWindow, code: WindowErrorCode})
		}
		oldParent, ok := s.windows[client.xID(window.parent)]
		if !ok {
			return client.sendError(&GenericError{seq: seq, badValue: window.parent, majorOp: ReparentWindow, code: WindowErrorCode})
		}
		newParent, ok := s.windows[parentXID]
		if !ok {
			return client.sendError(&GenericError{seq: seq, badValue: uint32(p.Parent), majorOp: ReparentWindow, code: WindowErrorCode})
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

	case *MapWindowRequest:
		xid := client.xID(uint32(p.Window))
		if w, ok := s.windows[xid]; ok {
			w.mapped = true
			s.frontend.MapWindow(xid)
			s.sendExposeEvent(xid, 0, 0, w.width, w.height)
		}

	case *MapSubwindowsRequest:
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

	case *UnmapWindowRequest:
		xid := client.xID(uint32(p.Window))
		if w, ok := s.windows[xid]; ok {
			w.mapped = false
		}
		s.frontend.UnmapWindow(xid)

	case *UnmapSubwindowsRequest:
		// TODO: implement

	case *ConfigureWindowRequest:
		xid := client.xID(uint32(p.Window))
		s.frontend.ConfigureWindow(xid, p.ValueMask, p.Values)

	case *CirculateWindowRequest:
		xid := client.xID(uint32(p.Window))
		window, ok := s.windows[xid]
		if !ok {
			return client.sendError(&GenericError{seq: seq, badValue: uint32(p.Window), majorOp: CirculateWindow, code: WindowErrorCode})
		}
		parent, ok := s.windows[client.xID(window.parent)]
		if !ok {
			return client.sendError(&GenericError{seq: seq, badValue: window.parent, majorOp: CirculateWindow, code: WindowErrorCode})
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

	case *GetGeometryRequest:
		xid := client.xID(uint32(p.Drawable))
		if xid.local == s.rootWindowID() {
			return &getGeometryReply{
				sequence:    seq,
				depth:       24, // TODO: Get this from rootVisual or screen info
				root:        s.rootWindowID(),
				x:           0,
				y:           0,
				width:       s.rootWindowWidth,
				height:      s.rootWindowHeight,
				borderWidth: 0,
			}
		}
		w, ok := s.windows[xid]
		if !ok {
			return nil
		}
		return &getGeometryReply{
			sequence:    seq,
			depth:       w.depth,
			root:        s.rootWindowID(),
			x:           w.x,
			y:           w.y,
			width:       w.width,
			height:      w.height,
			borderWidth: 0, // Border width is not stored in window struct, assuming 0 for now
		}

	case *QueryTreeRequest:
		xid := client.xID(uint32(p.Window))
		window, ok := s.windows[xid]
		if !ok {
			return client.sendError(&GenericError{seq: seq, badValue: uint32(p.Window), majorOp: QueryTree, code: WindowErrorCode})
		}
		return &queryTreeReply{
			sequence:    seq,
			root:        s.rootWindowID(),
			parent:      window.parent,
			numChildren: uint16(len(window.children)),
			children:    window.children,
		}

	case *InternAtomRequest:
		atomID := s.frontend.GetAtom(client.id, p.Name)

		return &internAtomReply{
			sequence: seq,
			atom:     atomID,
		}

	case *GetAtomNameRequest:
		name := s.frontend.GetAtomName(uint32(p.Atom))
		return &getAtomNameReply{
			sequence:   seq,
			nameLength: uint16(len(name)),
			name:       name,
		}

	case *ChangePropertyRequest:
		xid := client.xID(uint32(p.Window))
		s.frontend.ChangeProperty(xid, uint32(p.Property), uint32(p.Type), uint32(p.Format), p.Data)

	case *DeletePropertyRequest:
		// TODO: implement

	case *GetPropertyRequest:
		data, typ, format, bytesAfter := s.frontend.GetProperty(client.xID(uint32(p.Window)), uint32(p.Property), p.Offset, p.Length)

		if data == nil {
			return &getPropertyReply{
				sequence: seq,
				format:   0,
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

		if p.Delete && len(data) > 0 {
			// TODO: delete property
		}

		if p.Type != 0 && typ != uint32(p.Type) {
			// TODO: return empty property with the correct type
			// and format 0
		}

		return &getPropertyReply{
			sequence:              seq,
			format:                byte(format),
			propertyType:          typ,
			bytesAfter:            bytesAfter,
			valueLenInFormatUnits: valueLenInFormatUnits,
			value:                 data,
		}

	case *ListPropertiesRequest:
		xid := client.xID(uint32(p.Window))
		atoms := s.frontend.ListProperties(xid)
		return &listPropertiesReply{
			sequence:      seq,
			numProperties: uint16(len(atoms)),
			atoms:         atoms,
		}

	case *SetSelectionOwnerRequest:
		s.selections[client.xID(uint32(p.Selection))] = uint32(p.Owner)

	case *GetSelectionOwnerRequest:
		owner := s.selections[client.xID(uint32(p.Selection))]
		return &getSelectionOwnerReply{
			sequence: seq,
			owner:    owner,
		}

	case *ConvertSelectionRequest:
		s.frontend.ConvertSelection(uint32(p.Selection), uint32(p.Target), uint32(p.Property), client.xID(uint32(p.Requestor)))

	case *SendEventRequest:
		// The X11 client sends an event to another client.
		// We need to forward this event to the appropriate frontend.
		// For now, we'll just log it and pass it to the frontend.
		s.frontend.SendEvent(&x11RawEvent{data: p.EventData})

	case *GrabPointerRequest:
		grabWindow := client.xID(uint32(p.GrabWindow))
		if s.pointerGrabWindow.local != 0 {
			return &grabPointerReply{
				sequence: seq,
				status:   AlreadyGrabbed,
			}
		}

		s.pointerGrabWindow = grabWindow
		s.pointerGrabOwner = p.OwnerEvents
		s.pointerGrabEventMask = p.EventMask
		s.pointerGrabTime = uint32(p.Time)

		return &grabPointerReply{
			sequence: seq,
			status:   GrabSuccess,
		}

	case *UngrabPointerRequest:
		s.pointerGrabWindow = xID{}
		s.pointerGrabOwner = false
		s.pointerGrabEventMask = 0
		s.pointerGrabTime = 0

	case *GrabButtonRequest:
		grabWindow := client.xID(uint32(p.GrabWindow))
		grab := &passiveGrab{
			button:    p.Button,
			modifiers: p.Modifiers,
			owner:     p.OwnerEvents,
			eventMask: p.EventMask,
			cursor:    client.xID(uint32(p.Cursor)),
		}
		s.passiveGrabs[grabWindow] = append(s.passiveGrabs[grabWindow], grab)

	case *UngrabButtonRequest:
		grabWindow := client.xID(uint32(p.GrabWindow))
		if grabs, ok := s.passiveGrabs[grabWindow]; ok {
			for i, grab := range grabs {
				if grab.button == p.Button && grab.modifiers == p.Modifiers {
					s.passiveGrabs[grabWindow] = append(grabs[:i], grabs[i+1:]...)
					break
				}
			}
		}

	case *ChangeActivePointerGrabRequest:
		// TODO: implement

	case *GrabKeyboardRequest:
		grabWindow := client.xID(uint32(p.GrabWindow))
		if s.keyboardGrabWindow.local != 0 {
			return &grabKeyboardReply{
				sequence: seq,
				status:   AlreadyGrabbed,
			}
		}

		s.keyboardGrabWindow = grabWindow
		s.keyboardGrabOwner = p.OwnerEvents
		s.keyboardGrabTime = uint32(p.Time)

		return &grabKeyboardReply{
			sequence: seq,
			status:   GrabSuccess,
		}

	case *UngrabKeyboardRequest:
		s.keyboardGrabWindow = xID{}
		s.keyboardGrabOwner = false
		s.keyboardGrabTime = 0

	case *GrabKeyRequest:
		grabWindow := client.xID(uint32(p.GrabWindow))
		grab := &passiveGrab{
			key:       p.Key,
			modifiers: p.Modifiers,
			owner:     p.OwnerEvents,
		}
		s.passiveGrabs[grabWindow] = append(s.passiveGrabs[grabWindow], grab)

	case *UngrabKeyRequest:
		grabWindow := client.xID(uint32(p.GrabWindow))
		if grabs, ok := s.passiveGrabs[grabWindow]; ok {
			newGrabs := make([]*passiveGrab, 0, len(grabs))
			for _, grab := range grabs {
				if !(grab.key == p.Key && (p.Modifiers == AnyModifier || grab.modifiers == p.Modifiers)) {
					newGrabs = append(newGrabs, grab)
				}
			}
			s.passiveGrabs[grabWindow] = newGrabs
		}

	case *AllowEventsRequest:
		s.frontend.AllowEvents(client.id, p.Mode, uint32(p.Time))

	case *GrabServerRequest:
		// TODO: implement

	case *UngrabServerRequest:
		// TODO: implement

	case *QueryPointerRequest:
		xid := client.xID(uint32(p.Drawable))
		debugf("X11: QueryPointer drawable=%d", xid)
		return &queryPointerReply{
			sequence:   seq,
			sameScreen: true,
			root:       s.rootWindowID(),
			child:      uint32(p.Drawable),
			rootX:      s.pointerX,
			rootY:      s.pointerY,
			winX:       s.pointerX, // Assuming pointer is always in the window for now
			winY:       s.pointerY, // Assuming pointer is always in the window for now
			mask:       0,          // No buttons pressed
		}

	case *GetMotionEventsRequest:
		// TODO: implement
		return &getMotionEventsReply{
			sequence: seq,
		}

	case *TranslateCoordsRequest:
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

		return &translateCoordsReply{
			sequence:   seq,
			sameScreen: true,
			child:      0, // No child for now
			dstX:       dstX,
			dstY:       dstY,
		}

	case *WarpPointerRequest:
		// TODO: implement

	case *SetInputFocusRequest:
		// TODO: implement

	case *GetInputFocusRequest:
		return &getInputFocusReply{
			sequence: seq,
			revertTo: 1, // RevertToParent
			focus:    s.frontend.GetFocusWindow(client.id).local,
		}

	case *QueryKeymapRequest:
		// TODO: implement
		return &queryKeymapReply{
			sequence: seq,
		}

	case *OpenFontRequest:
		s.frontend.OpenFont(client.xID(uint32(p.Fid)), p.Name)

	case *CloseFontRequest:
		s.frontend.CloseFont(client.xID(uint32(p.Fid)))

	case *QueryFontRequest:
		minBounds, maxBounds, minCharOrByte2, maxCharOrByte2, defaultChar, drawDirection, minByte1, maxByte1, allCharsExist, fontAscent, fontDescent, charInfos := s.frontend.QueryFont(client.xID(uint32(p.Fid)))

		return &queryFontReply{
			sequence:       seq,
			minBounds:      minBounds,
			maxBounds:      maxBounds,
			minCharOrByte2: minCharOrByte2,
			maxCharOrByte2: maxCharOrByte2,
			defaultChar:    defaultChar,
			numFontProps:   0, // Not implemented yet
			drawDirection:  drawDirection,
			minByte1:       minByte1,
			maxByte1:       maxByte1,
			allCharsExist:  allCharsExist,
			fontAscent:     fontAscent,
			fontDescent:    fontDescent,
			numCharInfos:   uint32(len(charInfos)),
			charInfos:      charInfos,
		}

	case *QueryTextExtentsRequest:
		drawDirection, fontAscent, fontDescent, overallAscent, overallDescent, overallWidth, overallLeft, overallRight := s.frontend.QueryTextExtents(client.xID(uint32(p.Fid)), p.Text)
		return &queryTextExtentsReply{
			sequence:       seq,
			drawDirection:  drawDirection,
			fontAscent:     fontAscent,
			fontDescent:    fontDescent,
			overallAscent:  overallAscent,
			overallDescent: overallDescent,
			overallWidth:   int32(overallWidth),
			overallLeft:    int32(overallLeft),
			overallRight:   int32(overallRight),
		}

	case *ListFontsRequest:
		fontNames := s.frontend.ListFonts(p.MaxNames, p.Pattern)

		return &listFontsReply{
			sequence:  seq,
			fontNames: fontNames,
		}

	case *ListFontsWithInfoRequest:
		// TODO: implement
		return &listFontsWithInfoReply{
			sequence: seq,
		}

	case *SetFontPathRequest:
		// TODO: implement

	case *GetFontPathRequest:
		// TODO: implement
		return &getFontPathReply{
			sequence: seq,
		}

	case *CreatePixmapRequest:
		xid := client.xID(uint32(p.Pid))

		// Check if the pixmap ID is already in use
		if _, exists := s.pixmaps[xid]; exists {
			s.logger.Errorf("X11: CreatePixmap: ID %s already in use", xid)
			return client.sendError(&GenericError{seq: seq, badValue: uint32(p.Pid), majorOp: CreatePixmap, code: IDChoiceErrorCode})
		}

		s.pixmaps[xid] = true // Mark pixmap ID as used
		s.frontend.CreatePixmap(xid, client.xID(uint32(p.Drawable)), uint32(p.Width), uint32(p.Height), uint32(p.Depth))

	case *FreePixmapRequest:
		xid := client.xID(uint32(p.Pid))
		delete(s.pixmaps, xid)
		s.frontend.FreePixmap(xid)

	case *CreateGCRequest:
		xid := client.xID(uint32(p.Cid))

		// Check if the GC ID is already in use
		if _, exists := s.gcs[xid]; exists {
			s.logger.Errorf("X11: CreateGC: ID %s already in use", xid)
			return client.sendError(&GenericError{seq: seq, badValue: uint32(xid.local), majorOp: CreateGC, code: IDChoiceErrorCode})
		}

		s.gcs[xid] = p.Values
		s.frontend.CreateGC(xid, p.ValueMask, p.Values)

	case *ChangeGCRequest:
		xid := client.xID(uint32(p.Gc))
		if existingGC, ok := s.gcs[xid]; ok {
			if p.ValueMask&GCFunction != 0 {
				existingGC.Function = p.Values.Function
			}
			if p.ValueMask&GCPlaneMask != 0 {
				existingGC.PlaneMask = p.Values.PlaneMask
			}
			if p.ValueMask&GCForeground != 0 {
				existingGC.Foreground = p.Values.Foreground
			}
			if p.ValueMask&GCBackground != 0 {
				existingGC.Background = p.Values.Background
			}
			if p.ValueMask&GCLineWidth != 0 {
				existingGC.LineWidth = p.Values.LineWidth
			}
			if p.ValueMask&GCLineStyle != 0 {
				existingGC.LineStyle = p.Values.LineStyle
			}
			if p.ValueMask&GCCapStyle != 0 {
				existingGC.CapStyle = p.Values.CapStyle
			}
			if p.ValueMask&GCJoinStyle != 0 {
				existingGC.JoinStyle = p.Values.JoinStyle
			}
			if p.ValueMask&GCFillStyle != 0 {
				existingGC.FillStyle = p.Values.FillStyle
			}
			if p.ValueMask&GCFillRule != 0 {
				existingGC.FillRule = p.Values.FillRule
			}
			if p.ValueMask&GCTile != 0 {
				existingGC.Tile = p.Values.Tile
			}
			if p.ValueMask&GCStipple != 0 {
				existingGC.Stipple = p.Values.Stipple
			}
			if p.ValueMask&GCTileStipXOrigin != 0 {
				existingGC.TileStipXOrigin = p.Values.TileStipXOrigin
			}
			if p.ValueMask&GCTileStipYOrigin != 0 {
				existingGC.TileStipYOrigin = p.Values.TileStipYOrigin
			}
			if p.ValueMask&GCFont != 0 {
				existingGC.Font = p.Values.Font
			}
			if p.ValueMask&GCSubwindowMode != 0 {
				existingGC.SubwindowMode = p.Values.SubwindowMode
			}
			if p.ValueMask&GCGraphicsExposures != 0 {
				existingGC.GraphicsExposures = p.Values.GraphicsExposures
			}
			if p.ValueMask&GCClipXOrigin != 0 {
				existingGC.ClipXOrigin = p.Values.ClipXOrigin
			}
			if p.ValueMask&GCClipYOrigin != 0 {
				existingGC.ClipYOrigin = p.Values.ClipYOrigin
			}
			if p.ValueMask&GCClipMask != 0 {
				existingGC.ClipMask = p.Values.ClipMask
			}
			if p.ValueMask&GCDashOffset != 0 {
				existingGC.DashOffset = p.Values.DashOffset
			}
			if p.ValueMask&GCDashes != 0 {
				existingGC.Dashes = p.Values.Dashes
			}
			if p.ValueMask&GCArcMode != 0 {
				existingGC.ArcMode = p.Values.ArcMode
			}
		}
		s.frontend.ChangeGC(xid, p.ValueMask, p.Values)

	case *CopyGCRequest:
		srcGC := client.xID(uint32(p.SrcGC))
		dstGC := client.xID(uint32(p.DstGC))
		s.frontend.CopyGC(srcGC, dstGC)

	case *SetDashesRequest:
		s.frontend.SetDashes(client.xID(uint32(p.GC)), p.DashOffset, p.Dashes)

	case *SetClipRectanglesRequest:
		s.frontend.SetClipRectangles(client.xID(uint32(p.GC)), p.ClippingX, p.ClippingY, p.Rectangles, p.Ordering)

	case *FreeGCRequest:
		gcID := client.xID(uint32(p.GC))
		s.frontend.FreeGC(gcID)

	case *ClearAreaRequest:
		s.frontend.ClearArea(client.xID(uint32(p.Window)), int32(p.X), int32(p.Y), int32(p.Width), int32(p.Height))

	case *CopyAreaRequest:
		gcID := client.xID(uint32(p.Gc))
		s.frontend.CopyArea(client.xID(uint32(p.SrcDrawable)), client.xID(uint32(p.DstDrawable)), gcID, int32(p.SrcX), int32(p.SrcY), int32(p.DstX), int32(p.DstY), int32(p.Width), int32(p.Height))

	case *CopyPlaneRequest:
		gcID := client.xID(uint32(p.Gc))
		s.frontend.CopyPlane(client.xID(uint32(p.SrcDrawable)), client.xID(uint32(p.DstDrawable)), gcID, int32(p.SrcX), int32(p.SrcY), int32(p.DstX), int32(p.DstY), int32(p.Width), int32(p.Height), int32(p.PlaneMask))

	case *PolyPointRequest:
		gcID := client.xID(uint32(p.Gc))
		s.frontend.PolyPoint(client.xID(uint32(p.Drawable)), gcID, p.Coordinates)

	case *PolyLineRequest:
		gcID := client.xID(uint32(p.Gc))
		s.frontend.PolyLine(client.xID(uint32(p.Drawable)), gcID, p.Coordinates)

	case *PolySegmentRequest:
		gcID := client.xID(uint32(p.Gc))
		s.frontend.PolySegment(client.xID(uint32(p.Drawable)), gcID, p.Segments)

	case *PolyRectangleRequest:
		gcID := client.xID(uint32(p.Gc))
		s.frontend.PolyRectangle(client.xID(uint32(p.Drawable)), gcID, p.Rectangles)

	case *PolyArcRequest:
		gcID := client.xID(uint32(p.Gc))
		s.frontend.PolyArc(client.xID(uint32(p.Drawable)), gcID, p.Arcs)

	case *FillPolyRequest:
		gcID := client.xID(uint32(p.Gc))
		s.frontend.FillPoly(client.xID(uint32(p.Drawable)), gcID, p.Coordinates)

	case *PolyFillRectangleRequest:
		gcID := client.xID(uint32(p.Gc))
		s.frontend.PolyFillRectangle(client.xID(uint32(p.Drawable)), gcID, p.Rectangles)

	case *PolyFillArcRequest:
		gcID := client.xID(uint32(p.Gc))
		s.frontend.PolyFillArc(client.xID(uint32(p.Drawable)), gcID, p.Arcs)

	case *PutImageRequest:
		gcID := client.xID(uint32(p.Gc))
		s.frontend.PutImage(client.xID(uint32(p.Drawable)), gcID, p.Format, p.Width, p.Height, p.DstX, p.DstY, p.LeftPad, p.Depth, p.Data)

	case *GetImageRequest:
		imgData, err := s.frontend.GetImage(client.xID(uint32(p.Drawable)), int32(p.X), int32(p.Y), int32(p.Width), int32(p.Height), p.PlaneMask)
		if err != nil {
			s.logger.Errorf("Failed to get image: %v", err)
			return nil
		}
		return &getImageReply{
			sequence:  seq,
			depth:     24, // Assuming 24-bit depth for now
			visualID:  s.visualID,
			imageData: imgData,
		}

	case *PolyText8Request:
		gcID := client.xID(uint32(p.GC))
		s.frontend.PolyText8(client.xID(uint32(p.Drawable)), gcID, int32(p.X), int32(p.Y), p.Items)

	case *PolyText16Request:
		gcID := client.xID(uint32(p.GC))
		s.frontend.PolyText16(client.xID(uint32(p.Drawable)), gcID, int32(p.X), int32(p.Y), p.Items)

	case *ImageText8Request:
		gcID := client.xID(uint32(p.Gc))
		s.frontend.ImageText8(client.xID(uint32(p.Drawable)), gcID, int32(p.X), int32(p.Y), p.Text)

	case *ImageText16Request:
		gcID := client.xID(uint32(p.Gc))
		s.frontend.ImageText16(client.xID(uint32(p.Drawable)), gcID, int32(p.X), int32(p.Y), p.Text)

	case *CreateColormapRequest:
		xid := client.xID(uint32(p.Mid))

		if _, exists := s.colormaps[xid]; exists {
			return client.sendError(&GenericError{seq: seq, badValue: uint32(p.Mid), majorOp: CreateColormap, code: ColormapErrorCode})
		}

		newColormap := &colormap{
			pixels: make(map[uint32]xColorItem),
		}

		if p.Alloc == 1 { // All
			// For TrueColor, pre-allocating doesn't make much sense as pixels are calculated.
			// For other visual types, this would be important.
			// For now, we'll just create an empty map.
		}

		s.colormaps[xid] = newColormap

	case *FreeColormapRequest:
		xid := client.xID(uint32(p.Cmap))
		if _, ok := s.colormaps[xid]; !ok {
			return client.sendError(&GenericError{seq: seq, badValue: uint32(p.Cmap), majorOp: FreeColormap, code: ColormapErrorCode})
		}
		delete(s.colormaps, xid)

	case *InstallColormapRequest:
		xid := client.xID(uint32(p.Cmap))
		_, ok := s.colormaps[xid]
		if !ok {
			return client.sendError(&GenericError{seq: seq, badValue: uint32(p.Cmap), majorOp: InstallColormap, code: ColormapErrorCode})
		}

		s.installedColormap = xid

		for winID, win := range s.windows {
			if win.colormap == xid {
				client, ok := s.clients[winID.client]
				if !ok {
					debugf("X11: InstallColormap unknown client %d", winID.client)
					continue
				}
				event := &colormapNotifyEvent{
					sequence: client.sequence - 1,
					window:   winID.local,
					colormap: uint32(p.Cmap),
					new:      true,
					state:    0, // Installed
				}
				s.sendEvent(client, event)
			}
		}
		return nil

	case *UninstallColormapRequest:
		xid := client.xID(uint32(p.Cmap))
		_, ok := s.colormaps[xid]
		if !ok {
			return client.sendError(&GenericError{seq: seq, badValue: uint32(p.Cmap), majorOp: UninstallColormap, code: ColormapErrorCode})
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
				event := &colormapNotifyEvent{
					sequence: client.sequence - 1,
					window:   winID.local,
					colormap: uint32(p.Cmap),
					new:      false,
					state:    1, // Uninstalled
				}
				s.sendEvent(client, event)
			}
		}
		return nil

	case *ListInstalledColormapsRequest:
		var colormaps []uint32
		if s.installedColormap.local != 0 {
			colormaps = append(colormaps, s.installedColormap.local)
		}

		return &listInstalledColormapsReply{
			sequence:     seq,
			numColormaps: uint16(len(colormaps)),
			colormaps:    colormaps,
		}

	case *AllocColorRequest:
		xid := client.xID(uint32(p.Cmap))
		if xid.local == s.defaultColormap {
			xid.client = 0
		}
		cm, ok := s.colormaps[xid]
		if !ok {
			return client.sendError(&GenericError{seq: seq, badValue: uint32(p.Cmap), majorOp: AllocColor, code: ColormapErrorCode})
		}

		// Simple allocation for TrueColor: construct pixel value from RGB
		r8 := byte(p.Red >> 8)
		g8 := byte(p.Green >> 8)
		b8 := byte(p.Blue >> 8)
		pixel := (uint32(r8) << 16) | (uint32(g8) << 8) | uint32(b8)

		cm.pixels[pixel] = xColorItem{Red: p.Red, Green: p.Green, Blue: p.Blue}

		return &allocColorReply{
			sequence: seq,
			red:      p.Red,
			green:    p.Green,
			blue:     p.Blue,
			pixel:    pixel,
		}

	case *AllocNamedColorRequest:
		cmap := client.xID(uint32(p.Cmap))
		if cmap.local == s.defaultColormap {
			cmap.client = 0
		}
		cm, ok := s.colormaps[cmap]
		if !ok {
			return NewError(ColormapErrorCode, seq, uint32(p.Cmap), 0, p.OpCode())
		}

		name := string(p.Name)
		rgb, ok := lookupColor(name)
		if !ok {
			return NewError(NameErrorCode, seq, 0, 0, p.OpCode())
		}

		exactRed := scale8to16(rgb.Red)
		exactGreen := scale8to16(rgb.Green)
		exactBlue := scale8to16(rgb.Blue)

		pixel := (uint32(rgb.Red) << 16) | (uint32(rgb.Green) << 8) | uint32(rgb.Blue)
		cm.pixels[pixel] = xColorItem{Red: exactRed, Green: exactGreen, Blue: exactBlue}

		return &allocNamedColorReply{
			sequence: seq,
			red:      exactRed,
			green:    exactGreen,
			blue:     exactBlue,
			pixel:    pixel,
		}

	case *AllocColorCellsRequest:
		return &allocColorCellsReply{
			sequence: seq,
		}

	case *AllocColorPlanesRequest:
		return &allocColorPlanesReply{
			sequence: seq,
		}

	case *FreeColorsRequest:
		xid := client.xID(uint32(p.Cmap))
		if xid.local == s.defaultColormap {
			xid.client = 0
		}
		cm, ok := s.colormaps[xid]
		if !ok {
			return client.sendError(&GenericError{seq: seq, badValue: uint32(p.Cmap), majorOp: FreeColors, code: ColormapErrorCode})
		}

		for _, pixel := range p.Pixels {
			delete(cm.pixels, pixel)
		}

	case *StoreColorsRequest:
		xid := client.xID(uint32(p.Cmap))
		if xid.local == s.defaultColormap {
			xid.client = 0
		}
		cm, ok := s.colormaps[xid]
		if !ok {
			return client.sendError(&GenericError{seq: seq, badValue: uint32(p.Cmap), majorOp: StoreColors, code: ColormapErrorCode})
		}

		for _, item := range p.Items {
			c, exists := cm.pixels[item.Pixel]
			if !exists {
				c = xColorItem{}
			}

			if item.Flags&DoRed != 0 {
				c.Red = item.Red
			}
			if item.Flags&DoGreen != 0 {
				c.Green = item.Green
			}
			if item.Flags&DoBlue != 0 {
				c.Blue = item.Blue
			}
			cm.pixels[item.Pixel] = c
		}

	case *StoreNamedColorRequest:
		xid := client.xID(uint32(p.Cmap))
		if xid.local == s.defaultColormap {
			xid.client = 0
		}
		cm, ok := s.colormaps[xid]
		if !ok {
			return client.sendError(&GenericError{seq: seq, badValue: uint32(p.Cmap), majorOp: StoreNamedColor, code: ColormapErrorCode})
		}

		rgb, ok := lookupColor(p.Name)
		if !ok {
			return client.sendError(&GenericError{seq: seq, badValue: 0, majorOp: StoreNamedColor, code: NameErrorCode})
		}

		c, exists := cm.pixels[p.Pixel]
		if !exists {
			c = xColorItem{}
		}

		if p.Flags&DoRed != 0 {
			c.Red = scale8to16(rgb.Red)
		}
		if p.Flags&DoGreen != 0 {
			c.Green = scale8to16(rgb.Green)
		}
		if p.Flags&DoBlue != 0 {
			c.Blue = scale8to16(rgb.Blue)
		}
		cm.pixels[p.Pixel] = c

	case *QueryColorsRequest:
		cmap := client.xID(p.Cmap.local)
		if cmap.local == s.defaultColormap {
			cmap.client = 0
		}
		cm, ok := s.colormaps[cmap]
		if !ok {
			return client.sendError(&GenericError{seq: seq, badValue: p.Cmap.local, majorOp: QueryColors, code: ColormapErrorCode})
		}

		colors := make([]xColorItem, len(p.Pixels))
		for i, pixel := range p.Pixels {
			color, ok := cm.pixels[pixel]
			if !ok {
				// If the pixel is not in the colormap, return black
				color = xColorItem{Red: 0, Green: 0, Blue: 0}
			}
			colors[i] = color
		}

		return &queryColorsReply{
			sequence: seq,
			colors:   colors,
		}

	case *LookupColorRequest:
		// cmapID := client.xID(uint32(p.Cmap))

		color, ok := lookupColor(p.Name)
		if !ok {
			// TODO: This should be BadName, not BadColor
			return client.sendError(&GenericError{seq: seq, badValue: uint32(p.Cmap), majorOp: LookupColor, code: NameErrorCode})
		}

		return &lookupColorReply{
			sequence:   seq,
			red:        scale8to16(color.Red),
			green:      scale8to16(color.Green),
			blue:       scale8to16(color.Blue),
			exactRed:   scale8to16(color.Red),
			exactGreen: scale8to16(color.Green),
			exactBlue:  scale8to16(color.Blue),
		}

	case *CreateCursorRequest:
		cursorXID := client.xID(uint32(p.Cid))
		if _, exists := s.cursors[cursorXID]; exists {
			s.logger.Errorf("X11: CreateCursor: ID %s already in use", cursorXID)
			return client.sendError(&GenericError{seq: seq, badValue: uint32(p.Cid), majorOp: CreateCursor, code: IDChoiceErrorCode})
		}

		s.cursors[cursorXID] = true
		sourceXID := client.xID(uint32(p.Source))
		maskXID := client.xID(uint32(p.Mask))
		foreColor := [3]uint16{p.ForeRed, p.ForeGreen, p.ForeBlue}
		backColor := [3]uint16{p.BackRed, p.BackGreen, p.BackBlue}
		s.frontend.CreateCursor(cursorXID, sourceXID, maskXID, foreColor, backColor, p.X, p.Y)

	case *CreateGlyphCursorRequest:
		// Check if the cursor ID is already in use
		if _, exists := s.cursors[client.xID(uint32(p.Cid))]; exists {
			s.logger.Errorf("X11: CreateGlyphCursor: ID %d already in use", p.Cid)
			return client.sendError(&GenericError{seq: seq, badValue: uint32(p.Cid), majorOp: CreateGlyphCursor, code: IDChoiceErrorCode})
		}

		s.cursors[client.xID(uint32(p.Cid))] = true
		s.frontend.CreateCursorFromGlyph(uint32(p.Cid), p.SourceChar)

	case *FreeCursorRequest:
		xid := client.xID(uint32(p.Cursor))
		delete(s.cursors, xid)
		s.frontend.FreeCursor(xid)

	case *RecolorCursorRequest:
		s.frontend.RecolorCursor(client.xID(uint32(p.Cursor)), p.ForeColor, p.BackColor)

	case *QueryBestSizeRequest:
		debugf("X11: QueryBestSize class=%d drawable=%d width=%d height=%d", p.Class, p.Drawable, p.Width, p.Height)

		return &queryBestSizeReply{
			sequence: seq,
			width:    p.Width,
			height:   p.Height,
		}

	case *QueryExtensionRequest:
		debugf("X11: QueryExtension name=%s", p.Name)

		return &queryExtensionReply{
			sequence:    seq,
			present:     false,
			majorOpcode: 0,
			firstEvent:  0,
			firstError:  0,
		}

	case *ListExtensionsRequest:
		debugf("X11: ListExtensionsRequest not implemented")
		// TODO: Implement
		return &listExtensionsReply{
			sequence: seq,
		}

	case *ChangeKeyboardMappingRequest:
		s.frontend.ChangeKeyboardMapping(p.KeyCodeCount, p.FirstKeyCode, p.KeySymsPerKeyCode, p.KeySyms)

	case *GetKeyboardMappingRequest:
		keySyms, _ := s.frontend.GetKeyboardMapping(p.FirstKeyCode, p.Count)
		return &getKeyboardMappingReply{
			sequence: seq,
			keySyms:  keySyms,
		}

	case *ChangeKeyboardControlRequest:
		s.frontend.ChangeKeyboardControl(p.ValueMask, p.Values)

	case *GetKeyboardControlRequest:
		kc, _ := s.frontend.GetKeyboardControl()
		return &getKeyboardControlReply{
			sequence:         seq,
			keyClickPercent:  byte(kc.KeyClickPercent),
			bellPercent:      byte(kc.BellPercent),
			bellPitch:        uint16(kc.BellPitch),
			bellDuration:     uint16(kc.BellDuration),
			ledMask:          uint32(kc.Led),
			globalAutoRepeat: byte(kc.AutoRepeatMode),
			autoRepeats:      [32]byte{},
		}

	case *BellRequest:
		s.frontend.Bell(p.Percent)

	case *ChangePointerControlRequest:
		debugf("X11: ChangePointerControlRequest not implemented")
		// TODO: Implement

	case *GetPointerControlRequest:
		debugf("X11: GetPointerControlRequest not implemented")
		// TODO: Implement
		return &getPointerControlReply{
			sequence:         seq,
			accelNumerator:   1,
			accelDenominator: 1,
			threshold:        1,
		}

	case *SetScreenSaverRequest:
		s.frontend.SetScreenSaver(p.Timeout, p.Interval, p.PreferBlank, p.AllowExpose)

	case *GetScreenSaverRequest:
		timeout, interval, preferBlank, allowExpose, _ := s.frontend.GetScreenSaver()
		return &getScreenSaverReply{
			sequence:    seq,
			timeout:     uint16(timeout),
			interval:    uint16(interval),
			preferBlank: preferBlank,
			allowExpose: allowExpose,
		}

	case *ChangeHostsRequest:
		s.frontend.ChangeHosts(p.Mode, p.Host)

	case *ListHostsRequest:
		hosts, _ := s.frontend.ListHosts()
		return &listHostsReply{
			sequence: seq,
			numHosts: uint16(len(hosts)),
			hosts:    hosts,
		}

	case *SetAccessControlRequest:
		s.frontend.SetAccessControl(p.Mode)

	case *SetCloseDownModeRequest:
		s.frontend.SetCloseDownMode(p.Mode)

	case *KillClientRequest:
		s.frontend.KillClient(p.Resource)

	case *RotatePropertiesRequest:
		s.frontend.RotateProperties(client.xID(uint32(p.Window)), p.Delta, p.Atoms)

	case *ForceScreenSaverRequest:
		s.frontend.ForceScreenSaver(p.Mode)

	case *SetPointerMappingRequest:
		status, _ := s.frontend.SetPointerMapping(p.Map)
		return &setPointerMappingReply{
			sequence: seq,
			status:   status,
		}

	case *GetPointerMappingRequest:
		pMap, _ := s.frontend.GetPointerMapping()
		return &getPointerMappingReply{
			sequence: seq,
			length:   byte(len(pMap)),
			pMap:     pMap,
		}

	case *SetModifierMappingRequest:
		status, _ := s.frontend.SetModifierMapping(p.KeyCodesPerModifier, p.KeyCodes)
		return &setModifierMappingReply{
			sequence: seq,
			status:   status,
		}

	case *GetModifierMappingRequest:
		keyCodesPerModifier, keyCodes, _ := s.frontend.GetModifierMapping()
		return &getModifierMappingReply{
			sequence:            seq,
			keyCodesPerModifier: keyCodesPerModifier,
			keyCodes:            keyCodes,
		}

	case *NoOperationRequest:
		// TODO: implement

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
			client.send(&setupResponse{
				success: 0, // Failed
				reason:  "Invalid authorization",
			})
			return
		}
	}

	setup := newDefaultSetup()

	// Create the setup response message encoder
	responseMsg := &setupResponse{
		success:                  1, // Success
		protocolVersion:          11,
		releaseNumber:            setup.releaseNumber,
		resourceIDBase:           setup.resourceIDBase,
		resourceIDMask:           setup.resourceIDMask,
		motionBufferSize:         setup.motionBufferSize,
		vendorLength:             setup.vendorLength,
		maxRequestLength:         setup.maxRequestLength,
		numScreens:               setup.numScreens,
		numPixmapFormats:         setup.numPixmapFormats,
		imageByteOrder:           setup.imageByteOrder,
		bitmapFormatBitOrder:     setup.bitmapFormatBitOrder,
		bitmapFormatScanlineUnit: setup.bitmapFormatScanlineUnit,
		bitmapFormatScanlinePad:  setup.bitmapFormatScanlinePad,
		minKeycode:               setup.minKeycode,
		maxKeycode:               setup.maxKeycode,
		vendorString:             setup.vendorString,
		pixmapFormats:            setup.pixmapFormats,
		screens:                  setup.screens,
	}

	if err := client.send(responseMsg); err != nil {
		s.logger.Errorf("x11 handshake write: %v", err)
		return
	}
	s.visualID = setup.screens[0].rootVisual
	s.rootVisual = setup.screens[0].depths[0].visuals[0]
	s.blackPixel = setup.screens[0].blackPixel
	s.whitePixel = setup.screens[0].whitePixel
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
					gcs:        make(map[xID]GC),
					pixmaps:    make(map[xID]bool),
					cursors:    make(map[xID]bool),
					selections: make(map[xID]uint32),
					colormaps: map[xID]*colormap{
						xID{local: 0x1}: {
							pixels: map[uint32]xColorItem{
								0x000000: {Pixel: 0x000000, Red: 0x0000, Green: 0x0000, Blue: 0x0000, Flags: 0},
								1:        {Pixel: 1, Red: 0xffff, Green: 0xffff, Blue: 0xffff, Flags: 0},
								0xffffff: {Pixel: 0xffffff, Red: 0xffff, Green: 0xffff, Blue: 0xffff, Flags: 0},
							},
						},
					},
					defaultColormap: 0x1,
					clients:         make(map[uint32]*x11Client),
					nextClientID:    1,
					passiveGrabs:    make(map[xID][]*passiveGrab),
					authProtocol:    authProtocol,
					authCookie:      authCookie,
				}
				x11ServerInstance.frontend = newX11Frontend(logger, x11ServerInstance)
			})

			client := &x11Client{
				id:        x11ServerInstance.nextClientID,
				conn:      channel,
				sequence:  0,
				byteOrder: binary.LittleEndian, // Default, will be updated in handshake
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
