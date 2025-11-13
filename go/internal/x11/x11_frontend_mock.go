//go:build x11 && !wasm

package x11

type propertyChange struct {
	id                 xID
	property, typeAtom uint32
	format             byte
	data               []byte
}

type putImageCall struct {
	drawable      xID
	gcID          xID
	depth         uint8
	width, height uint16
	dstX, dstY    int16
	leftPad       uint8
	format        uint8
	data          []byte
}

type polyLineCall struct {
	drawable xID
	gcID     xID
	points   []uint32
}

type polyFillRectCall struct {
	drawable xID
	gcID     xID
	rects    []uint32
}

type fillPolyCall struct {
	drawable xID
	gcID     xID
	points   []uint32
}

type polySegmentCall struct {
	drawable xID
	gcID     xID
	segments []uint32
}

type polyPointCall struct {
	drawable xID
	gcID     xID
	points   []uint32
}

type polyRectCall struct {
	drawable xID
	gcID     xID
	rects    []uint32
}

type polyArcCall struct {
	drawable xID
	gcID     xID
	arcs     []uint32
}

type polyFillArcCall struct {
	drawable xID
	gcID     xID
	arcs     []uint32
}

type clearAreaCall struct {
	drawable            xID
	x, y, width, height uint32
}

type copyAreaCall struct {
	srcDrawable, dstDrawable              xID
	gcID                                  xID
	srcX, srcY, dstX, dstY, width, height uint32
}

type copyPlaneCall struct {
	srcDrawable, dstDrawable                        xID
	gcID                                            xID
	srcX, srcY, dstX, dstY, width, height, bitPlane uint32
}

type getImageCall struct {
	drawable                    xID
	x, y, width, height, format uint32
}

type listPropertiesCall struct {
	window xID
}

type configureWindowCall struct {
	id        xID
	valueMask uint16
	values    []uint32
}

type circulateWindowCall struct {
	id        xID
	direction byte
}

type getPropertyCall struct {
	window     xID
	property   uint32
	longOffset uint32
	longLength uint32
}

type reparentWindowCall struct {
	window xID
	parent xID
	x, y   int16
}

// MockX11Frontend is a mock implementation of the X11FrontendAPI for testing.
type MockX11Frontend struct {
	CreateWindowCalls               []*window
	ReparentWindowCalls             []*reparentWindowCall
	DestroyWindowCalls              []xID
	DestroySubwindowsCalls          []xID
	DestroyAllWindowsForClientCalls []uint32
	MapWindowCalls                  []xID
	UnmapWindowCalls                []xID
	ConfigureWindowCalls            []*configureWindowCall
	CirculateWindowCalls            []*circulateWindowCall
	CreatedGCs                      map[xID]GC
	ChangedGCs                      map[xID]GC
	ChangedProperties               []*propertyChange
	PutImageCalls                   []*putImageCall
	PolyLineCalls                   []*polyLineCall
	PolyFillRectangleCalls          []*polyFillRectCall
	FillPolyCalls                   []*fillPolyCall
	PolySegmentCalls                []*polySegmentCall
	PolyPointCalls                  []*polyPointCall
	PolyRectangleCalls              []*polyRectCall
	PolyArcCalls                    []*polyArcCall
	PolyFillArcCalls                []*polyFillArcCall
	ClearAreaCalls                  []*clearAreaCall
	CopyAreaCalls                   []*copyAreaCall
	CopyPlaneCalls                  []*copyPlaneCall
	GetImageCalls                   []*getImageCall
	GetImageReturn                  []byte
	GetImageError                   error
	GetAtomNameCalls                []uint32
	GetAtomNameReturn               string
	ListPropertiesCalls             []*listPropertiesCall
	ListPropertiesReturn            []uint32
	ClipboardContent                string
	WrittenClipboard                string
	ReadClipboardCalls              []struct{}
	ReadClipboardReturn             string
	ReadClipboardError              error
	ImageText8Calls                 []*imageText8Call
	ImageText16Calls                []*imageText16Call
	PolyText8Calls                  []*polyText8Call
	PolyText16Calls                 []*polyText16Call
	BellCalls                       []int8
	ChangePropertyCalls             []*propertyChange
	GetPropertyCalls                []*getPropertyCall
	GetPropertyReturn               []byte
	GetPropertyTypeReturn           uint32
	GetPropertyFormatReturn         uint32
	GetPropertyBytesAfterReturn     uint32
	CanvasOperations                []CanvasOperation
	SetInputFocusCalls              []setInputFocusCall
}

type setInputFocusCall struct {
	focus    xID
	revertTo byte
}

type imageText8Call struct {
	drawable xID
	gcID     xID
	x, y     int32
	text     []byte
}

type imageText16Call struct {
	drawable xID
	gcID     xID
	x, y     int32
	text     []uint16
}

type polyText8Call struct {
	drawable xID
	gcID     xID
	x, y     int32
	items    []PolyTextItem
}

type polyText16Call struct {
	drawable xID
	gcID     xID
	x, y     int32
	items    []PolyTextItem
}

func (m *MockX11Frontend) CreateWindow(xid xID, parent, x, y, width, height, depth, valueMask uint32, values WindowAttributes) {
	// For mock, we can just log the call or do nothing.
	// No internal state to clean up for windows in the mock.
}

func (m *MockX11Frontend) ChangeWindowAttributes(xid xID, valueMask uint32, values WindowAttributes) {
	// Not implemented for mock
}

func (m *MockX11Frontend) GetWindowAttributes(xid xID) WindowAttributes {
	// Not implemented for mock
	return WindowAttributes{}
}

func (m *MockX11Frontend) DestroyWindow(xid xID) {
	m.DestroyWindowCalls = append(m.DestroyWindowCalls, xid)
}

func (m *MockX11Frontend) ReparentWindow(window xID, parent xID, x, y int16) {
	m.ReparentWindowCalls = append(m.ReparentWindowCalls, &reparentWindowCall{window, parent, x, y})
}

func (m *MockX11Frontend) DestroySubwindows(xid xID) {
	m.DestroySubwindowsCalls = append(m.DestroySubwindowsCalls, xid)
}

func (m *MockX11Frontend) DestroyAllWindowsForClient(clientID uint32) {
	m.DestroyAllWindowsForClientCalls = append(m.DestroyAllWindowsForClientCalls, clientID)
}

func (m *MockX11Frontend) MapWindow(xid xID) {
	m.MapWindowCalls = append(m.MapWindowCalls, xid)
}

func (m *MockX11Frontend) UnmapWindow(xid xID) {
	m.UnmapWindowCalls = append(m.UnmapWindowCalls, xid)
}

func (m *MockX11Frontend) ConfigureWindow(xid xID, valueMask uint16, values []uint32) {
	m.ConfigureWindowCalls = append(m.ConfigureWindowCalls, &configureWindowCall{xid, valueMask, values})
}

func (m *MockX11Frontend) CirculateWindow(xid xID, direction byte) {
	m.CirculateWindowCalls = append(m.CirculateWindowCalls, &circulateWindowCall{xid, direction})
}

func (w *MockX11Frontend) PutImage(drawable xID, gcID xID, format uint8, width, height uint16, dstX, dstY int16, leftPad, depth uint8, data []byte) {
	w.PutImageCalls = append(w.PutImageCalls, &putImageCall{drawable, gcID, depth, width, height, dstX, dstY, leftPad, format, data})
}

func (m *MockX11Frontend) PolyLine(drawable xID, gcID xID, points []uint32) {
	m.PolyLineCalls = append(m.PolyLineCalls, &polyLineCall{drawable, gcID, points})
}

func (m *MockX11Frontend) PolyFillRectangle(drawable xID, gcID xID, rects []uint32) {
	m.PolyFillRectangleCalls = append(m.PolyFillRectangleCalls, &polyFillRectCall{drawable, gcID, rects})
}

func (m *MockX11Frontend) FillPoly(drawable xID, gcID xID, points []uint32) {
	m.FillPolyCalls = append(m.FillPolyCalls, &fillPolyCall{drawable, gcID, points})
}

func (m *MockX11Frontend) PolySegment(drawable xID, gcID xID, segments []uint32) {
	m.PolySegmentCalls = append(m.PolySegmentCalls, &polySegmentCall{drawable, gcID, segments})
}

func (m *MockX11Frontend) PolyPoint(drawable xID, gcID xID, points []uint32) {
	m.PolyPointCalls = append(m.PolyPointCalls, &polyPointCall{drawable, gcID, points})
}

func (m *MockX11Frontend) PolyRectangle(drawable xID, gcID xID, rects []uint32) {
	m.PolyRectangleCalls = append(m.PolyRectangleCalls, &polyRectCall{drawable, gcID, rects})
}

func (m *MockX11Frontend) PolyArc(drawable xID, gcID xID, arcs []uint32) {
	m.PolyArcCalls = append(m.PolyArcCalls, &polyArcCall{drawable, gcID, arcs})
}

func (m *MockX11Frontend) PolyFillArc(drawable xID, gcID xID, arcs []uint32) {
	m.PolyFillArcCalls = append(m.PolyFillArcCalls, &polyFillArcCall{drawable, gcID, arcs})
}

func (m *MockX11Frontend) ClearArea(drawable xID, x, y, width, height int32) {
	m.ClearAreaCalls = append(m.ClearAreaCalls, &clearAreaCall{drawable, uint32(x), uint32(y), uint32(width), uint32(height)})
}

func (m *MockX11Frontend) CopyArea(srcDrawable, dstDrawable xID, gcID xID, srcX, srcY, dstX, dstY, width, height int32) {
	m.CopyAreaCalls = append(m.CopyAreaCalls, &copyAreaCall{srcDrawable, dstDrawable, gcID, uint32(srcX), uint32(srcY), uint32(dstX), uint32(dstY), uint32(width), uint32(height)})
}

func (m *MockX11Frontend) CopyPlane(srcDrawable, dstDrawable xID, gcID xID, srcX, srcY, dstX, dstY, width, height, bitPlane int32) {
	m.CopyPlaneCalls = append(m.CopyPlaneCalls, &copyPlaneCall{srcDrawable, dstDrawable, gcID, uint32(srcX), uint32(srcY), uint32(dstX), uint32(dstY), uint32(width), uint32(height), uint32(bitPlane)})
}

func (m *MockX11Frontend) GetImage(drawable xID, x, y, width, height int32, format uint32) ([]byte, error) {
	m.GetImageCalls = append(m.GetImageCalls, &getImageCall{drawable, uint32(x), uint32(y), uint32(width), uint32(height), format})
	return m.GetImageReturn, nil
}

func (m *MockX11Frontend) ImageText8(drawable xID, gcID xID, x, y int32, text []byte) {
	m.ImageText8Calls = append(m.ImageText8Calls, &imageText8Call{drawable, gcID, x, y, text})
}

func (m *MockX11Frontend) ImageText16(drawable xID, gcID xID, x, y int32, text []uint16) {
	m.ImageText16Calls = append(m.ImageText16Calls, &imageText16Call{drawable, gcID, x, y, text})
}

func (m *MockX11Frontend) PolyText8(drawable xID, gcID xID, x, y int32, items []PolyTextItem) {
	m.PolyText8Calls = append(m.PolyText8Calls, &polyText8Call{drawable, gcID, x, y, items})
}

func (m *MockX11Frontend) PolyText16(drawable xID, gcID xID, x, y int32, items []PolyTextItem) {
	m.PolyText16Calls = append(m.PolyText16Calls, &polyText16Call{drawable, gcID, x, y, items})
}

func (m *MockX11Frontend) CreatePixmap(id, drawable xID, width, height, depth uint32) {}

func (m *MockX11Frontend) FreePixmap(xid xID) {}

func (m *MockX11Frontend) CopyPixmap(srcID, dstID, gcID xID, srcX, srcY, width, height, dstX, dstY uint32) {
}

func (m *MockX11Frontend) CreateCursor(cursorID xID, source, mask xID, foreColor, backColor [3]uint16, x, y uint16) {
}

func (m *MockX11Frontend) CreateCursorFromGlyph(cursorID uint32, glyphID uint16) {}

func (m *MockX11Frontend) SetWindowCursor(windowID xID, cursorID xID) {}

func (m *MockX11Frontend) CopyGC(srcGC, dstGC xID) {}

func (m *MockX11Frontend) FreeGC(gc xID) {}

func (m *MockX11Frontend) FreeCursor(cursorID xID) {
	// For mock, we can just log the call or do nothing.
	// No internal state to clean up for cursors in the mock.
}

func (m *MockX11Frontend) SendEvent(eventData messageEncoder) {}

func (m *MockX11Frontend) GetFocusWindow(uint32) xID { return xID{} }

func (m *MockX11Frontend) ConvertSelection(selection, target, property uint32, requestor xID) {}

func (m *MockX11Frontend) GrabPointer(grabWindow xID, ownerEvents bool, eventMask uint16, pointerMode, keyboardMode byte, confineTo uint32, cursor uint32, time uint32) byte {
	return 0
}

func (m *MockX11Frontend) UngrabPointer(time uint32) {}

func (m *MockX11Frontend) GrabKeyboard(grabWindow xID, ownerEvents bool, time uint32, pointerMode, keyboardMode byte) byte {
	return 0
}

func (m *MockX11Frontend) UngrabKeyboard(time uint32) {}

func (m *MockX11Frontend) SetCursor(window xID, cursor uint32) {}

func (m *MockX11Frontend) WarpPointer(x, y int16) {}

func (m *MockX11Frontend) Bell(percent int8) {
	m.BellCalls = append(m.BellCalls, percent)
}

func (m *MockX11Frontend) ListProperties(window xID) []uint32 {
	m.ListPropertiesCalls = append(m.ListPropertiesCalls, &listPropertiesCall{window})
	return m.ListPropertiesReturn
}

func (m *MockX11Frontend) GetProperty(window xID, property uint32, longOffset, longLength uint32) ([]byte, uint32, uint32, uint32) {
	m.GetPropertyCalls = append(m.GetPropertyCalls, &getPropertyCall{window, property, longOffset, longLength})
	return m.GetPropertyReturn, m.GetPropertyTypeReturn, m.GetPropertyFormatReturn, m.GetPropertyBytesAfterReturn
}

func (m *MockX11Frontend) QueryFont(fid xID) (minBounds, maxBounds xCharInfo, minCharOrByte2, maxCharOrByte2, defaultChar uint16, drawDirection uint8, minByte1, maxByte1 uint8, allCharsExist bool, fontAscent, fontDescent int16, charInfos []xCharInfo) {
	// Dummy implementation for mock
	return
}

func (m *MockX11Frontend) QueryTextExtents(font xID, text []uint16) (drawDirection uint8, fontAscent, fontDescent, overallAscent, overallDescent, overallWidth, overallLeft, overallRight int16) {
	// Dummy implementation for mock
	return
}

func (m *MockX11Frontend) CloseFont(fid xID) {
	// Dummy implementation for mock
}

func (m *MockX11Frontend) ListFonts(maxNames uint16, pattern string) []string {
	// Dummy implementation for mock
	return []string{}
}

func (m *MockX11Frontend) ReadClipboard() (string, error) {
	m.ReadClipboardCalls = append(m.ReadClipboardCalls, struct{}{})
	return m.ReadClipboardReturn, m.ReadClipboardError
}

func (m *MockX11Frontend) WriteClipboard(s string) error {
	m.WrittenClipboard = s
	return nil
}

func (m *MockX11Frontend) UpdatePointerPosition(x, y int16) {} // No-op for mock

func (m *MockX11Frontend) GetAtom(clientID uint32, name string) uint32 {
	return 0 // No-op for mock
}

func (m *MockX11Frontend) GetAtomName(atom uint32) string {
	m.GetAtomNameCalls = append(m.GetAtomNameCalls, atom)
	return m.GetAtomNameReturn
}

func (m *MockX11Frontend) CreateGC(id xID, valueMask uint32, values GC) {
	if m.CreatedGCs == nil {
		m.CreatedGCs = make(map[xID]GC)
	}
	m.CreatedGCs[id] = values
}

func (m *MockX11Frontend) ChangeGC(id xID, valueMask uint32, gc GC) {
	if m.ChangedGCs == nil {
		m.ChangedGCs = make(map[xID]GC)
	}
	m.ChangedGCs[id] = gc
}

func (m *MockX11Frontend) ChangeProperty(id xID, property, typ, format uint32, data []byte) {
	m.ChangedProperties = append(m.ChangedProperties, &propertyChange{id, property, typ, byte(format), data})
}

func (m *MockX11Frontend) DeleteProperty(xid xID, property uint32) {
}

func newX11Frontend(logger Logger, s *x11Server) X11FrontendAPI {
	return &MockX11Frontend{}
}

func (m *MockX11Frontend) GetCanvasOperations() []CanvasOperation {
	return m.CanvasOperations
}

func (m *MockX11Frontend) GetRGBColor(colormap xID, pixel uint32) (r, g, b uint8) {
	if pixel == 0 {
		return 0xFF, 0xFF, 0xFF // White
	}
	return 0, 0, 0 // Black
}

func (m *MockX11Frontend) OpenFont(fid xID, name string) {
	// Dummy implementation for mock
}

func (m *MockX11Frontend) AllowEvents(clientID uint32, mode byte, time uint32) {
	// Dummy implementation for mock
}

func (m *MockX11Frontend) SendConfigureAndExposeEvent(windowID xID, x, y int16, width, height uint16) {
	// Dummy implementation for mock
}

func (m *MockX11Frontend) SetDashes(gc xID, dashOffset uint16, dashes []byte) {
}

func (m *MockX11Frontend) SetClipRectangles(gc xID, clippingX, clippingY int16, rectangles []Rectangle, ordering byte) {
}

func (m *MockX11Frontend) RecolorCursor(cursor xID, foreColor, backColor [3]uint16) {
}

func (m *MockX11Frontend) SetPointerMapping(pMap []byte) (byte, error) {
	return 0, nil
}

func (m *MockX11Frontend) GetPointerMapping() ([]byte, error) {
	return nil, nil
}

func (m *MockX11Frontend) GetKeyboardMapping(firstKeyCode KeyCode, count byte) ([]uint32, error) {
	return nil, nil
}

func (m *MockX11Frontend) ChangeKeyboardMapping(keyCodeCount byte, firstKeyCode KeyCode, keySymsPerKeyCode byte, keySyms []uint32) {
}

func (m *MockX11Frontend) ChangeKeyboardControl(valueMask uint32, values KeyboardControl) {
}

func (m *MockX11Frontend) GetKeyboardControl() (KeyboardControl, error) {
	return KeyboardControl{}, nil
}

func (m *MockX11Frontend) SetScreenSaver(timeout, interval int16, preferBlank, allowExpose byte) {
}

func (m *MockX11Frontend) GetScreenSaver() (timeout, interval int16, preferBlank, allowExpose byte, err error) {
	return 0, 0, 0, 0, nil
}

func (m *MockX11Frontend) ChangeHosts(mode byte, host Host) {
}

func (m *MockX11Frontend) ListHosts() ([]Host, error) {
	return nil, nil
}

func (m *MockX11Frontend) SetAccessControl(mode byte) {
}

func (m *MockX11Frontend) SetCloseDownMode(mode byte) {
}

func (m *MockX11Frontend) KillClient(resource uint32) {
}

func (m *MockX11Frontend) RotateProperties(window xID, delta int16, atoms []Atom) {
}

func (m *MockX11Frontend) ForceScreenSaver(mode byte) {
}

func (m *MockX11Frontend) SetModifierMapping(keyCodesPerModifier byte, keyCodes []KeyCode) (byte, error) {
	return 0, nil
}

func (m *MockX11Frontend) GetModifierMapping() (keyCodesPerModifier byte, keyCodes []KeyCode, err error) {
	return 0, nil, nil
}

func (m *MockX11Frontend) SetInputFocus(focus xID, revertTo byte) {
	m.SetInputFocusCalls = append(m.SetInputFocusCalls, setInputFocusCall{focus, revertTo})
}
