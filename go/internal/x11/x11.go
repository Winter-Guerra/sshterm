//go:build x11

package x11

import (
	"bytes"
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
	CreateWindow(xid xID, parent, x, y, width, height, depth, valueMask uint32, values *WindowAttributes)
	ChangeWindowAttributes(xid xID, valueMask uint32, values *WindowAttributes)
	GetWindowAttributes(xid xID) *WindowAttributes
	ChangeProperty(xid xID, property, typeAtom, format uint32, data []byte)
	CreateGC(xid xID, gc *GC)
	ChangeGC(xid xID, valueMask uint32, gc *GC)
	DestroyWindow(xid xID)
	DestroyAllWindowsForClient(clientID uint32)
	MapWindow(xid xID)
	UnmapWindow(xid xID)
	ConfigureWindow(xid xID, valueMask uint16, values []uint32)
	PutImage(drawable xID, gc *GC, format uint8, width, height uint16, dstX, dstY int16, leftPad, depth uint8, data []byte)
	PolyLine(drawable xID, gc *GC, points []uint32)
	PolyFillRectangle(drawable xID, gc *GC, rects []uint32)
	FillPoly(drawable xID, gc *GC, points []uint32)
	PolySegment(drawable xID, gc *GC, segments []uint32)
	PolyPoint(drawable xID, gc *GC, points []uint32)
	PolyRectangle(drawable xID, gc *GC, rects []uint32)
	PolyArc(drawable xID, gc *GC, arcs []uint32)
	PolyFillArc(drawable xID, gc *GC, arcs []uint32)
	ClearArea(drawable xID, x, y, width, height int32)
	CopyArea(srcDrawable, dstDrawable xID, gc *GC, srcX, srcY, dstX, dstY, width, height int32)
	GetImage(drawable xID, x, y, width, height int32, format uint32) ([]byte, error)
	ReadClipboard() (string, error)
	WriteClipboard(string) error
	UpdatePointerPosition(x, y int16)
	Bell(percent int8)
	GetAtom(clientID uint32, name string) uint32
	GetAtomName(atom uint32) string
	ListProperties(window xID) []uint32
	GetProperty(window xID, property uint32) ([]byte, uint32, uint32)
	ImageText8(drawable xID, gc *GC, x, y int32, text []byte)
	ImageText16(drawable xID, gc *GC, x, y int32, text []uint16)
	PolyText8(drawable xID, gc *GC, x, y int32, items []PolyText8Item)
	PolyText16(drawable xID, gc *GC, x, y int32, items []PolyText16Item)
	CreatePixmap(xid, drawable xID, width, height, depth uint32)
	FreePixmap(xid xID)
	CopyPixmap(srcID, dstID, gcID xID, srcX, srcY, width, height, dstX, dstY uint32)
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
	GetRGBColor(colormap xID, pixel uint32) (r, g, b uint32)
	OpenFont(fid xID, name string)
	QueryFont(fid xID) (minBounds, maxBounds xCharInfo, minCharOrByte2, maxCharOrByte2, defaultChar uint16, drawDirection uint8, minByte1, maxByte1 uint8, allCharsExist bool, fontAscent, fontDescent int16, charInfos []xCharInfo)
	CloseFont(fid xID)
	ListFonts(maxNames uint16, pattern string) []string
	AllowEvents(clientID uint32, mode byte, time uint32)
}

type XError interface {
	Code() byte
	Sequence() uint16
	BadValue() uint32
	MinorOp() byte
	MajorOp() byte
}

// getWindowAttributesReply implements messageEncoder for GetWindowAttributes reply.
type getWindowAttributesReply struct {
	sequence           uint16
	backingStore       byte
	visualID           uint32
	class              uint16
	bitGravity         byte
	winGravity         byte
	backingPlanes      uint32
	backingPixel       uint32
	saveUnder          bool
	mapped             bool
	mapState           byte
	overrideRedirect   bool
	colormap           uint32
	allEventMasks      uint32
	yourEventMask      uint32
	doNotPropagateMask uint16
}

func (r *getWindowAttributesReply) encodeMessage(order binary.ByteOrder) []byte {
	reply := make([]byte, 44)
	reply[0] = 1 // Reply type
	reply[1] = r.backingStore
	order.PutUint16(reply[2:4], r.sequence)
	order.PutUint32(reply[4:8], 3) // Reply length (3 * 4 bytes = 12 bytes, plus 32 bytes header = 44 bytes total)
	order.PutUint32(reply[8:12], r.visualID)
	order.PutUint16(reply[12:14], r.class)
	reply[14] = r.bitGravity
	reply[15] = r.winGravity
	order.PutUint32(reply[16:20], r.backingPlanes)
	order.PutUint32(reply[20:24], r.backingPixel)
	reply[24] = boolToByte(r.saveUnder)
	reply[25] = boolToByte(r.mapped)
	reply[26] = r.mapState
	reply[27] = boolToByte(r.overrideRedirect)
	order.PutUint32(reply[28:32], r.colormap)
	order.PutUint32(reply[32:36], r.allEventMasks)
	order.PutUint32(reply[36:40], r.yourEventMask)
	order.PutUint16(reply[40:42], r.doNotPropagateMask)
	// reply[42:44] is padding
	return reply
}

// getGeometryReply implements messageEncoder for GetGeometry reply.
type getGeometryReply struct {
	sequence      uint16
	depth         byte
	root          uint32
	x, y          int16
	width, height uint16
	borderWidth   uint16
}

func (r *getGeometryReply) encodeMessage(order binary.ByteOrder) []byte {
	reply := make([]byte, 32)
	reply[0] = 1 // Reply type
	reply[1] = r.depth
	order.PutUint16(reply[2:4], r.sequence)
	order.PutUint32(reply[4:8], 0) // Reply length (0 * 4 bytes = 0 bytes, plus 32 bytes header = 32 bytes total)
	order.PutUint32(reply[8:12], r.root)
	order.PutUint16(reply[12:14], uint16(r.x))
	order.PutUint16(reply[14:16], uint16(r.y))
	order.PutUint16(reply[16:18], r.width)
	order.PutUint16(reply[18:20], r.height)
	order.PutUint16(reply[20:22], r.borderWidth)
	// reply[22:32] is padding
	return reply
}

// internAtomReply implements messageEncoder for InternAtom reply.
type internAtomReply struct {
	sequence uint16
	atom     uint32
}

func (r *internAtomReply) encodeMessage(order binary.ByteOrder) []byte {
	reply := make([]byte, 32)
	reply[0] = 1 // Reply type
	// byte 1 is unused
	order.PutUint16(reply[2:4], r.sequence)
	order.PutUint32(reply[4:8], 0) // Reply length (0 * 4 bytes = 0 bytes, plus 32 bytes header = 32 bytes total)
	order.PutUint32(reply[8:12], r.atom)
	// reply[12:32] is padding
	return reply
}

// getAtomNameReply implements messageEncoder for GetAtomName reply.
type getAtomNameReply struct {
	sequence   uint16
	nameLength uint16
	name       string
}

func (r *getAtomNameReply) encodeMessage(order binary.ByteOrder) []byte {
	nameLen := len(r.name)
	p := (4 - (nameLen % 4)) % 4
	reply := make([]byte, 32+nameLen+p)
	reply[0] = 1 // Reply type
	// byte 1 is unused
	order.PutUint16(reply[2:4], r.sequence)
	order.PutUint32(reply[4:8], uint32((nameLen+p)/4)) // Reply length
	order.PutUint16(reply[8:10], uint16(nameLen))
	// reply[10:32] is padding
	copy(reply[32:], r.name)
	return reply
}

// queryPointerReply implements messageEncoder for QueryPointer reply.
type queryPointerReply struct {
	sequence     uint16
	sameScreen   bool
	root         uint32
	child        uint32
	rootX, rootY int16
	winX, winY   int16
	mask         uint16
}

func (r *queryPointerReply) encodeMessage(order binary.ByteOrder) []byte {
	reply := make([]byte, 32)
	reply[0] = 1 // Reply type
	reply[1] = boolToByte(r.sameScreen)
	order.PutUint16(reply[2:4], r.sequence)
	order.PutUint32(reply[4:8], 0) // Reply length (0 * 4 bytes = 0 bytes, plus 32 bytes header = 32 bytes total)
	order.PutUint32(reply[8:12], r.root)
	order.PutUint32(reply[12:16], r.child)
	order.PutUint16(reply[16:18], uint16(r.rootX))
	order.PutUint16(reply[18:20], uint16(r.rootY))
	order.PutUint16(reply[20:22], uint16(r.winX))
	order.PutUint16(reply[22:24], uint16(r.winY))
	order.PutUint16(reply[24:26], r.mask)
	// reply[26:32] is padding
	return reply
}

// listPropertiesReply implements messageEncoder for ListProperties reply.
type listPropertiesReply struct {
	sequence      uint16
	numProperties uint16
	atoms         []uint32
}

func (r *listPropertiesReply) encodeMessage(order binary.ByteOrder) []byte {
	numAtoms := len(r.atoms)
	reply := make([]byte, 32+numAtoms*4)
	reply[0] = 1 // Reply type
	// byte 1 is unused
	order.PutUint16(reply[2:4], r.sequence)
	order.PutUint32(reply[4:8], uint32(numAtoms)) // Reply length
	order.PutUint16(reply[8:10], uint16(numAtoms))
	// reply[10:32] is padding
	for i, atom := range r.atoms {
		order.PutUint32(reply[32+i*4:], atom)
	}
	return reply
}

// getImageReply implements messageEncoder for GetImage reply.
type getImageReply struct {
	sequence  uint16
	depth     byte
	visualID  uint32
	imageData []byte
}

func (r *getImageReply) encodeMessage(order binary.ByteOrder) []byte {
	reply := make([]byte, 32+len(r.imageData))
	reply[0] = 1 // Reply type
	reply[1] = r.depth
	order.PutUint16(reply[2:4], r.sequence)
	order.PutUint32(reply[4:8], uint32(len(r.imageData)/4)) // Reply length
	order.PutUint32(reply[8:12], r.visualID)
	// reply[12:32] is padding
	copy(reply[32:], r.imageData)
	return reply
}

// getPropertyReply implements messageEncoder for GetProperty reply.
type getPropertyReply struct {
	sequence              uint16
	format                byte
	propertyType          uint32
	bytesAfter            uint32
	valueLenInFormatUnits uint32
	value                 []byte
}

func (r *getPropertyReply) encodeMessage(order binary.ByteOrder) []byte {
	n := len(r.value)
	p := (4 - (n % 4)) % 4
	replyLen := (n + p) / 4

	reply := make([]byte, 32+n+p)
	reply[0] = 1 // Reply type
	reply[1] = r.format
	order.PutUint16(reply[2:4], r.sequence)
	order.PutUint32(reply[4:8], uint32(replyLen)) // Reply length
	order.PutUint32(reply[8:12], r.propertyType)
	order.PutUint32(reply[12:16], r.bytesAfter)
	order.PutUint32(reply[16:20], r.valueLenInFormatUnits)
	// reply[20:32] is padding
	copy(reply[32:], r.value)
	return reply
}

// queryBestSizeReply implements messageEncoder for QueryBestSize reply.
type queryBestSizeReply struct {
	sequence uint16
	width    uint16
	height   uint16
}

func (r *queryBestSizeReply) encodeMessage(order binary.ByteOrder) []byte {
	reply := make([]byte, 32)
	reply[0] = 1 // Reply type
	// byte 1 is unused
	order.PutUint16(reply[2:4], r.sequence)
	order.PutUint32(reply[4:8], 0) // Reply length (0 * 4 bytes = 0 bytes, plus 32 bytes header = 32 bytes total)
	order.PutUint16(reply[8:10], r.width)
	order.PutUint16(reply[10:12], r.height)
	// reply[12:32] is padding
	return reply
}

// queryExtensionReply implements messageEncoder for QueryExtension reply.
type queryExtensionReply struct {
	sequence    uint16
	present     bool
	majorOpcode byte
	firstEvent  byte
	firstError  byte
}

func (r *queryExtensionReply) encodeMessage(order binary.ByteOrder) []byte {
	reply := make([]byte, 32)
	reply[0] = 1 // Reply type
	reply[1] = boolToByte(r.present)
	order.PutUint16(reply[2:4], r.sequence)
	order.PutUint32(reply[4:8], 0) // Reply length (0 * 4 bytes = 0 bytes, plus 32 bytes header = 32 bytes total)
	reply[8] = r.majorOpcode
	reply[9] = r.firstEvent
	reply[10] = r.firstError
	// reply[11:32] is padding
	return reply
}

// queryColorsReply implements messageEncoder for QueryColors reply.
type queryColorsReply struct {
	sequence uint16
	colors   []color
}

func (r *queryColorsReply) encodeMessage(order binary.ByteOrder) []byte {
	numColors := len(r.colors)
	replies := make([]byte, numColors*8)
	for i, color := range r.colors {
		order.PutUint16(replies[i*8:], color.Red)
		order.PutUint16(replies[i*8+2:], color.Green)
		order.PutUint16(replies[i*8+4:], color.Blue)
		// replies[i*8+6:i*8+8] unused
	}

	reply := make([]byte, 32+len(replies))
	reply[0] = 1 // Reply
	// byte 1 is unused
	order.PutUint16(reply[2:4], r.sequence)
	order.PutUint32(reply[4:8], uint32(len(replies)/4)) // Reply length
	order.PutUint16(reply[8:10], uint16(numColors))
	// reply[10:32] is padding
	copy(reply[32:], replies)
	return reply
}

// lookupColorReply implements messageEncoder for LookupColor reply.
type lookupColorReply struct {
	sequence   uint16
	red        uint16
	green      uint16
	blue       uint16
	exactRed   uint16
	exactGreen uint16
	exactBlue  uint16
}

func (r *lookupColorReply) encodeMessage(order binary.ByteOrder) []byte {
	reply := make([]byte, 32)
	reply[0] = 1 // Reply type
	// byte 1 is unused
	order.PutUint16(reply[2:4], r.sequence)
	order.PutUint32(reply[4:8], 0) // Reply length (0 * 4 bytes = 0 bytes, plus 32 bytes header = 32 bytes total)
	order.PutUint16(reply[8:10], r.red)
	order.PutUint16(reply[10:12], r.green)
	order.PutUint16(reply[12:14], r.blue)
	order.PutUint16(reply[14:16], r.exactRed)
	order.PutUint16(reply[16:18], r.exactGreen)
	order.PutUint16(reply[18:20], r.exactBlue)
	// reply[20:32] is padding
	return reply
}

// allocColorReply implements messageEncoder for AllocColor reply.
type allocColorReply struct {
	sequence uint16
	red      uint16
	green    uint16
	blue     uint16
	pixel    uint32
}

func (r *allocColorReply) encodeMessage(order binary.ByteOrder) []byte {
	reply := make([]byte, 32)
	reply[0] = 1 // Reply type
	// byte 1 is unused
	order.PutUint16(reply[2:4], r.sequence)
	order.PutUint32(reply[4:8], 0) // Reply length (0 * 4 bytes = 0 bytes, plus 32 bytes header = 32 bytes total)
	order.PutUint16(reply[8:10], r.red)
	order.PutUint16(reply[10:12], r.green)
	order.PutUint16(reply[12:14], r.blue)
	// reply[14:16] is padding
	order.PutUint32(reply[8:12], r.pixel)
	// reply[20:32] is padding
	return reply
}

// listFontsReply implements messageEncoder for ListFonts reply.
type listFontsReply struct {
	sequence  uint16
	numFonts  uint16
	fontNames []string
}

func (r *listFontsReply) encodeMessage(order binary.ByteOrder) []byte {
	var namesData []byte
	for _, name := range r.fontNames {
		namesData = append(namesData, byte(len(name)))
		namesData = append(namesData, []byte(name)...)
	}

	namesSize := len(namesData)
	padSize := (4 - (namesSize % 4)) % 4

	reply := make([]byte, 32+namesSize+padSize)
	reply[0] = 1 // Reply
	// byte 1 is unused
	order.PutUint16(reply[2:4], r.sequence)
	order.PutUint32(reply[4:8], uint32((namesSize+padSize)/4)) // Reply length
	order.PutUint16(reply[8:10], uint16(len(r.fontNames)))
	// reply[10:32] is padding
	copy(reply[32:], namesData)
	return reply
}

// queryFontReply implements messageEncoder for QueryFont reply.
type queryFontReply struct {
	sequence       uint16
	minBounds      xCharInfo
	maxBounds      xCharInfo
	minCharOrByte2 uint16
	maxCharOrByte2 uint16
	defaultChar    uint16
	numFontProps   uint16
	drawDirection  uint8
	minByte1       uint8
	maxByte1       uint8
	allCharsExist  bool
	fontAscent     int16
	fontDescent    int16
	numCharInfos   uint32
	charInfos      []xCharInfo
}

func (r *queryFontReply) encodeMessage(order binary.ByteOrder) []byte {
	numFontProps := 0 // Not implemented yet
	numCharInfos := len(r.charInfos)

	reply := make([]byte, 60+8*numFontProps+12*numCharInfos)
	reply[0] = 1 // Reply
	reply[1] = 1 // font-info-present (True)
	order.PutUint16(reply[2:4], r.sequence)
	order.PutUint32(reply[4:8], uint32(7+2*numFontProps+3*numCharInfos)) // Reply length

	// min-bounds
	order.PutUint16(reply[8:10], uint16(r.minBounds.LeftSideBearing))
	order.PutUint16(reply[10:12], uint16(r.minBounds.RightSideBearing))
	order.PutUint16(reply[12:14], uint16(r.minBounds.CharacterWidth))
	order.PutUint16(reply[14:16], uint16(r.minBounds.Ascent))
	order.PutUint16(reply[16:18], uint16(r.minBounds.Descent))
	order.PutUint16(reply[18:20], r.minBounds.Attributes)

	// max-bounds
	order.PutUint16(reply[24:26], uint16(r.maxBounds.LeftSideBearing))
	order.PutUint16(reply[26:28], uint16(r.maxBounds.RightSideBearing))
	order.PutUint16(reply[28:30], uint16(r.maxBounds.CharacterWidth))
	order.PutUint16(reply[30:32], uint16(r.maxBounds.Ascent))
	order.PutUint16(reply[32:34], uint16(r.maxBounds.Descent))
	order.PutUint16(reply[34:36], r.maxBounds.Attributes)

	order.PutUint16(reply[40:42], r.minCharOrByte2)
	order.PutUint16(reply[42:44], r.maxCharOrByte2)
	order.PutUint16(reply[44:46], r.defaultChar)
	order.PutUint16(reply[46:48], uint16(numFontProps))

	reply[48] = r.drawDirection & 0x1
	reply[49] = r.minByte1
	reply[50] = r.maxByte1
	reply[51] = boolToByte(r.allCharsExist)

	order.PutUint16(reply[52:54], uint16(r.fontAscent))
	order.PutUint16(reply[54:56], uint16(r.fontDescent))

	order.PutUint32(reply[56:60], uint32(numCharInfos))

	offset := 60 + 8*numFontProps
	for _, ci := range r.charInfos {
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

// listInstalledColormapsReply implements messageEncoder for ListInstalledColormaps reply.
type listInstalledColormapsReply struct {
	sequence     uint16
	numColormaps uint16
	colormaps    []uint32
}

func (r *listInstalledColormapsReply) encodeMessage(order binary.ByteOrder) []byte {
	nColormaps := len(r.colormaps)
	reply := make([]byte, 32+nColormaps*4)
	reply[0] = 1 // Reply
	// byte 1 is unused
	order.PutUint16(reply[2:4], r.sequence)
	order.PutUint32(reply[4:8], uint32(nColormaps)) // length
	order.PutUint16(reply[8:10], uint16(nColormaps))
	// reply[10:32] is padding
	for i, cmap := range r.colormaps {
		order.PutUint32(reply[32+i*4:], cmap)
	}
	return reply
}

// translateCoordsReply implements messageEncoder for TranslateCoords reply.
type translateCoordsReply struct {
	sequence   uint16
	sameScreen bool
	child      uint32
	dstX, dstY int16
}

func (r *translateCoordsReply) encodeMessage(order binary.ByteOrder) []byte {
	reply := make([]byte, 32)
	reply[0] = 1 // Reply type
	reply[1] = boolToByte(r.sameScreen)
	order.PutUint16(reply[2:4], r.sequence)
	order.PutUint32(reply[4:8], 0) // Reply length (0 * 4 bytes = 0 bytes, plus 32 bytes header = 32 bytes total)
	order.PutUint32(reply[8:12], r.child)
	order.PutUint16(reply[12:14], uint16(r.dstX))
	order.PutUint16(reply[14:16], uint16(r.dstY))
	// reply[16:32] is padding
	return reply
}

// getInputFocusReply implements messageEncoder for GetInputFocus reply.
type getInputFocusReply struct {
	sequence uint16
	revertTo byte
	focus    uint32
}

func (r *getInputFocusReply) encodeMessage(order binary.ByteOrder) []byte {
	reply := make([]byte, 32)
	reply[0] = 1 // Reply type
	reply[1] = r.revertTo
	order.PutUint16(reply[2:4], r.sequence)
	order.PutUint32(reply[4:8], 0) // Reply length (0 * 4 bytes = 0 bytes, plus 32 bytes header = 32 bytes total)
	order.PutUint32(reply[8:12], r.focus)
	// reply[12:32] is padding
	return reply
}

// getSelectionOwnerReply implements messageEncoder for GetSelectionOwner reply.
type getSelectionOwnerReply struct {
	sequence uint16
	owner    uint32
}

func (r *getSelectionOwnerReply) encodeMessage(order binary.ByteOrder) []byte {
	reply := make([]byte, 32)
	reply[0] = 1 // Reply type
	// byte 1 is unused
	order.PutUint16(reply[2:4], r.sequence)
	order.PutUint32(reply[4:8], 0) // Reply length (0 * 4 bytes = 0 bytes, plus 32 bytes header = 32 bytes total)
	order.PutUint32(reply[8:12], r.owner)
	// reply[12:32] is padding
	return reply
}

// grabPointerReply implements messageEncoder for GrabPointer reply.
type grabPointerReply struct {
	sequence uint16
	status   byte
}

func (r *grabPointerReply) encodeMessage(order binary.ByteOrder) []byte {
	reply := make([]byte, 32)
	reply[0] = 1 // Reply type
	reply[1] = r.status
	order.PutUint16(reply[2:4], r.sequence)
	order.PutUint32(reply[4:8], 0) // Reply length (0 * 4 bytes = 0 bytes, plus 32 bytes header = 32 bytes total)
	// reply[8:32] is padding
	return reply
}

// grabKeyboardReply implements messageEncoder for GrabKeyboard reply.
type grabKeyboardReply struct {
	sequence uint16
	status   byte
}

func (r *grabKeyboardReply) encodeMessage(order binary.ByteOrder) []byte {
	reply := make([]byte, 32)
	reply[0] = 1 // Reply type
	reply[1] = r.status
	order.PutUint16(reply[2:4], r.sequence)
	order.PutUint32(reply[4:8], 0) // Reply length (0 * 4 bytes = 0 bytes, plus 32 bytes header = 32 bytes total)
	// reply[8:32] is padding
	return reply
}

// setupResponse implements messageEncoder for the X11 setup response.
type setupResponse struct {
	success                  byte
	protocolVersion          uint16
	releaseNumber            uint32
	resourceIDBase           uint32
	resourceIDMask           uint32
	motionBufferSize         uint32
	vendorLength             uint16
	maxRequestLength         uint16
	numScreens               uint8
	numPixmapFormats         uint8
	imageByteOrder           uint8
	bitmapFormatBitOrder     byte
	bitmapFormatScanlineUnit byte
	bitmapFormatScanlinePad  byte
	minKeycode               uint8
	maxKeycode               uint8
	vendorString             string
	pixmapFormats            []format
	screens                  []screen
}

func (r *setupResponse) encodeMessage(order binary.ByteOrder) []byte {
	setup := newDefaultSetup() // This should probably be passed in or generated once
	setupData := setup.marshal(order)

	response := make([]byte, 8+len(setupData))
	response[0] = r.success
	// byte 1 is unused
	order.PutUint16(response[2:4], r.protocolVersion)
	order.PutUint16(response[4:6], 0) // length of additional data in 4-byte units
	order.PutUint16(response[6:8], uint16(len(setupData)/4))
	copy(response[8:], setupData)
	return response
}

// motionNotifyEvent implements messageEncoder for MotionNotify event.
type motionNotifyEvent struct {
	sequence       uint16
	detail         byte
	time           uint32
	root           uint32
	event          uint32
	child          uint32
	rootX, rootY   int16
	eventX, eventY int16
	state          uint16
	sameScreen     bool
}

func (e *motionNotifyEvent) encodeMessage(order binary.ByteOrder) []byte {
	event := make([]byte, 32)
	event[0] = 6 // MotionNotify event code
	event[1] = e.detail
	order.PutUint16(event[2:4], e.sequence)
	order.PutUint32(event[4:8], e.time)
	order.PutUint32(event[8:12], e.root)
	order.PutUint32(event[12:16], e.event)
	order.PutUint32(event[16:20], e.child)
	order.PutUint16(event[20:22], uint16(e.rootX))
	order.PutUint16(event[22:24], uint16(e.rootY))
	order.PutUint16(event[24:26], uint16(e.eventX))
	order.PutUint16(event[26:28], uint16(e.eventY))
	order.PutUint16(event[28:30], e.state)
	event[30] = boolToByte(e.sameScreen)
	// event[31] is unused
	return event
}

// configureNotifyEvent implements messageEncoder for ConfigureNotify event.
type configureNotifyEvent struct {
	sequence         uint16
	event            uint32
	window           uint32
	aboveSibling     uint32
	x, y             int16
	width, height    uint16
	borderWidth      uint16
	overrideRedirect bool
}

func (e *configureNotifyEvent) encodeMessage(order binary.ByteOrder) []byte {
	event := make([]byte, 32)
	event[0] = 22 // ConfigureNotify event code
	// byte 1 is unused
	order.PutUint16(event[2:4], e.sequence)
	order.PutUint32(event[4:8], e.event)
	order.PutUint32(event[8:12], e.window)
	order.PutUint32(event[12:16], e.aboveSibling)
	order.PutUint16(event[16:18], uint16(e.x))
	order.PutUint16(event[18:20], uint16(e.y))
	order.PutUint16(event[20:22], e.width)
	order.PutUint16(event[22:24], e.height)
	order.PutUint16(event[24:26], e.borderWidth)
	event[26] = boolToByte(e.overrideRedirect)
	// byte 27 is unused
	return event
}

// exposeEvent implements messageEncoder for Expose event.
type exposeEvent struct {
	sequence      uint16
	window        uint32
	x, y          uint16
	width, height uint16
	count         uint16
}

func (e *exposeEvent) encodeMessage(order binary.ByteOrder) []byte {
	event := make([]byte, 32)
	event[0] = 12 // Expose event code
	// byte 1 is unused
	order.PutUint16(event[2:4], e.sequence)
	order.PutUint32(event[4:8], e.window)
	order.PutUint16(event[8:10], e.x)
	order.PutUint16(event[10:12], e.y)
	order.PutUint16(event[12:14], e.width)
	order.PutUint16(event[14:16], e.height)
	order.PutUint16(event[16:18], e.count)
	// event[18:32] is unused
	return event
}

// clientMessageEvent implements messageEncoder for ClientMessage event.
type clientMessageEvent struct {
	sequence    uint16
	format      byte
	window      uint32
	messageType uint32
	data        [20]byte
}

func (e *clientMessageEvent) encodeMessage(order binary.ByteOrder) []byte {
	event := make([]byte, 32)
	event[0] = 33 // ClientMessage event code
	event[1] = e.format
	order.PutUint16(event[2:4], e.sequence)
	order.PutUint32(event[4:8], e.window)
	order.PutUint32(event[8:12], e.messageType)
	copy(event[12:32], e.data[:])
	return event
}

// selectionNotifyEvent implements messageEncoder for SelectionNotify event.
type selectionNotifyEvent struct {
	sequence  uint16
	requestor uint32
	selection uint32
	target    uint32
	property  uint32
	time      uint32
}

func (e *selectionNotifyEvent) encodeMessage(order binary.ByteOrder) []byte {
	event := make([]byte, 32)
	event[0] = 31 // SelectionNotify event code
	// byte 1 is unused
	order.PutUint16(event[2:4], e.sequence)
	order.PutUint32(event[4:8], e.requestor)
	order.PutUint32(event[8:12], e.selection)
	order.PutUint32(event[12:16], e.target)
	order.PutUint32(event[16:20], e.property)
	order.PutUint32(event[20:24], e.time)
	// event[24:32] is unused
	return event
}

// colormapNotifyEvent implements messageEncoder for ColormapNotify event.
type colormapNotifyEvent struct {
	sequence uint16
	window   uint32
	colormap uint32
	new      bool
	state    byte
}

func (e *colormapNotifyEvent) encodeMessage(order binary.ByteOrder) []byte {
	event := make([]byte, 32)
	event[0] = ColormapNotify // ColormapNotify event code
	// byte 1 is unused
	order.PutUint16(event[2:4], e.sequence)
	order.PutUint32(event[4:8], e.window)
	order.PutUint32(event[8:12], e.colormap)
	event[12] = boolToByte(e.new)
	event[13] = e.state
	// event[14:32] is unused
	return event
}

// keyEvent implements messageEncoder for KeyPress and KeyRelease events.
type keyEvent struct {
	sequence       uint16
	detail         byte // keycode
	time           uint32
	root           uint32
	event          uint32
	child          uint32
	rootX, rootY   int16
	eventX, eventY int16
	state          uint16 // keyboard state
	sameScreen     bool
}

func (e *keyEvent) encodeMessage(order binary.ByteOrder) []byte {
	event := make([]byte, 32)
	// event[0] will be set to KeyPress (2) or KeyRelease (3) by the caller
	event[1] = e.detail
	order.PutUint16(event[2:4], e.sequence)
	order.PutUint32(event[4:8], e.time)
	order.PutUint32(event[8:12], e.root)
	order.PutUint32(event[12:16], e.event)
	order.PutUint32(event[16:20], e.child)
	order.PutUint16(event[20:22], uint16(e.rootX))
	order.PutUint16(event[22:24], uint16(e.rootY))
	order.PutUint16(event[24:26], uint16(e.eventX))
	order.PutUint16(event[26:28], uint16(e.eventY))
	order.PutUint16(event[28:30], e.state)
	event[30] = boolToByte(e.sameScreen)
	// event[31] is unused
	return event
}

// x11RawEvent implements messageEncoder for raw X11 event data.
type x11RawEvent struct {
	data []byte
}

func (e *x11RawEvent) encodeMessage(order binary.ByteOrder) []byte {
	return e.data
}

// QueryColorsRequest represents the request for QueryColors.
type QueryColorsRequest struct {
	Cmap   xID
	Pixels []uint32
	// MinorOp and MajorOp are part of XError, but also useful for context
	MinorOp  byte
	MajorOp  reqCode
	Sequence uint16
}

func parseQueryColorsRequest(order binary.ByteOrder, body []byte) QueryColorsRequest {
	cmapID := order.Uint32(body[0:4])
	numPixels := (len(body) - 4) / 4
	pixels := make([]uint32, numPixels)
	for i := 0; i < numPixels; i++ {
		pixels[i] = order.Uint32(body[4+i*4 : 4+i*4+4])
	}
	return QueryColorsRequest{
		Cmap:   xID{local: cmapID},
		Pixels: pixels,
	}
}

// CanvasOperation represents a single canvas drawing operation captured from the frontend.
type CanvasOperation struct {
	Type        string `json:"type"`
	Args        []any  `json:"args"`
	FillStyle   string `json:"fillStyle"`
	StrokeStyle string `json:"strokeStyle"`
}

// request represents an X11 request.
type request struct {
	opcode   reqCode
	data     byte
	length   uint16
	sequence uint16
	body     []byte
}

type window struct {
	xid           xID
	parent        uint32
	x, y          int16
	width, height uint16
	mapped        bool
	depth         byte
	children      []uint32
	attributes    *WindowAttributes
	colormap      xID
}

func (w *window) mapState() byte {
	if !w.mapped {
		return 0 // Unmapped
	}
	return 2 // Viewable
}

type xCharInfo struct {
	LeftSideBearing  int16
	RightSideBearing int16
	CharacterWidth   uint16
	Ascent           int16
	Descent          int16
	Attributes       uint16
}

type color struct {
	Red   uint16
	Green uint16
	Blue  uint16
}

type colormap struct {
	pixels map[uint32]color
}

type BadColor struct {
	seq      uint16
	badValue uint32
	minorOp  byte
	majorOp  reqCode
}

func (e BadColor) Code() byte       { return ColormapError }
func (e BadColor) Sequence() uint16 { return e.seq }
func (e BadColor) BadValue() uint32 { return e.badValue }
func (e BadColor) MinorOp() byte    { return e.minorOp }
func (e BadColor) MajorOp() byte    { return byte(e.majorOp) }

type GenericError struct {
	seq      uint16
	badValue uint32
	minorOp  byte
	majorOp  reqCode
	code     byte
}

func (e GenericError) Code() byte       { return e.code }
func (e GenericError) Sequence() uint16 { return e.seq }
func (e GenericError) BadValue() uint32 { return e.badValue }
func (e GenericError) MinorOp() byte    { return e.minorOp }
func (e GenericError) MajorOp() byte    { return byte(e.majorOp) }

type x11Server struct {
	logger             Logger
	byteOrder          binary.ByteOrder
	frontend           X11FrontendAPI
	windows            map[xID]*window
	gcs                map[xID]*GC
	pixmaps            map[xID]bool
	cursors            map[xID]bool
	selections         map[xID]uint32
	colormaps          map[xID]*colormap
	defaultColormap    uint32
	installedColormap  xID
	visualID           uint32
	rootVisual         visualType
	blackPixel         uint32
	whitePixel         uint32
	pointerX, pointerY int16
	clients            map[uint32]*x11Client
	nextClientID       uint32
}

// Start of setup struct
type setup struct {
	releaseNumber            uint32
	resourceIDBase           uint32
	resourceIDMask           uint32
	motionBufferSize         uint32
	vendorLength             uint16
	maxRequestLength         uint16
	numScreens               uint8
	numPixmapFormats         uint8
	imageByteOrder           uint8
	bitmapFormatBitOrder     uint8
	bitmapFormatScanlineUnit uint8
	bitmapFormatScanlinePad  uint8
	minKeycode               uint8
	maxKeycode               uint8
	vendorString             string
	pixmapFormats            []format
	screens                  []screen
}

type format struct {
	depth        uint8
	bitsPerPixel uint8
	scanlinePad  uint8
}

type screen struct {
	root                uint32
	defaultColormap     uint32
	whitePixel          uint32
	blackPixel          uint32
	currentInputMasks   uint32
	widthInPixels       uint16
	heightInPixels      uint16
	widthInMillimeters  uint16
	heightInMillimeters uint16
	minInstalledMaps    uint16
	maxInstalledMaps    uint16
	rootVisual          uint32
	backingStores       uint8
	saveUnders          bool
	rootDepth           uint8
	numDepths           uint8
	depths              []depth
}

type depth struct {
	depth      uint8
	numVisuals uint16
	visuals    []visualType
}

type visualType struct {
	visualID        uint32 // visual-id
	class           uint8
	bitsPerRGBValue uint8
	colormapEntries uint16
	redMask         uint32
	greenMask       uint32
	blueMask        uint32
}

func (s *x11Server) UpdatePointerPosition(x, y int16) {
	s.pointerX = x
	s.pointerY = y
}

func (s *x11Server) SendMouseEvent(xid xID, eventType string, x, y, detail int32) {
	log.Printf("X11: SendMouseEvent xid=%s type=%s x=%d y=%d detail=%d", xid, eventType, x, y, detail)
	client, ok := s.clients[xid.client]
	if !ok {
		log.Print("X11: Failed to write mount event: client not found")
		return
	}

	event := &motionNotifyEvent{
		sequence:   client.sequence,
		detail:     0, // 0 for Normal
		time:       0, // 0 for now
		root:       s.rootWindowID(),
		event:      xid.local,
		child:      0, // 0 for now
		rootX:      int16(x),
		rootY:      int16(y),
		eventX:     int16(x),
		eventY:     int16(y),
		state:      uint16(detail),
		sameScreen: true,
	}

	if err := client.send(event); err != nil {
		log.Printf("X11: Failed to write mouse event: %v", err)
	}
}

func (s *x11Server) SendKeyboardEvent(xid xID, eventType string, keyCode int, altKey, ctrlKey, shiftKey, metaKey bool) {
	// Implement sending keyboard event to client
	// This will involve constructing an X11 event packet and writing it to client.conn
	log.Printf("X11: SendKeyboardEvent xid=%s type=%s keyCode=%d alt=%t ctrl=%t shift=%t meta=%t", xid, eventType, keyCode, altKey, ctrlKey, shiftKey, metaKey)
	client, ok := s.clients[xid.client]
	if !ok {
		log.Printf("X11: SendKeyboardEvent unknown client %d", xid.client)
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

	event := &keyEvent{
		sequence:   client.sequence,
		detail:     byte(keyCode),
		time:       0, // TODO: Get actual time
		root:       s.rootWindowID(),
		event:      xid.local,
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
		log.Printf("X11: Unknown keyboard event type: %s", eventType)
		return
	}

	if err := client.send(event); err != nil {
		log.Printf("X11: Failed to write keyboard event: %v", err)
	}
}

func (s *x11Server) sendConfigureNotifyEvent(windowID xID, x, y int16, width, height uint16) {
	log.Printf("X11: Sending ConfigureNotify event for window %d", windowID)
	client, ok := s.clients[windowID.client]
	if !ok {
		log.Print("X11: Failed to write ConfigureNotify event: client not found")
		return
	}

	event := &configureNotifyEvent{
		sequence:         client.sequence,
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
		log.Printf("X11: Failed to write ConfigureNotify event: %v", err)
	}
}

func (s *x11Server) sendExposeEvent(windowID xID, x, y, width, height uint16) {
	log.Printf("X11: Sending Expose event for window %s", windowID)
	client, ok := s.clients[windowID.client]
	if !ok {
		log.Printf("X11: sendExposeEvent unknown client %d", windowID.client)
		return
	}

	event := &exposeEvent{
		sequence: client.sequence,
		window:   windowID.local,
		x:        x,
		y:        y,
		width:    width,
		height:   height,
		count:    0, // count = 0, no more expose events to follow
	}

	if err := client.send(event); err != nil {
		log.Printf("X11: Failed to write Expose event: %v", err)
	}
}

func (s *x11Server) SendClientMessageEvent(windowID xID, messageTypeAtom uint32, data [20]byte) {
	log.Printf("X11: Sending ClientMessage event for window %s", windowID)
	client, ok := s.clients[windowID.client]
	if !ok {
		log.Printf("X11: SendClientMessageEvent unknown client %d", windowID.client)
		return
	}

	event := &clientMessageEvent{
		sequence:    client.sequence,
		format:      32, // Format is always 32 for ClientMessage
		window:      windowID.local,
		messageType: messageTypeAtom,
		data:        data,
	}

	if err := client.send(event); err != nil {
		log.Printf("X11: Failed to write ClientMessage event: %v", err)
	}
}

func (s *x11Server) SendSelectionNotify(requestor xID, selection, target, property uint32, data []byte) {
	client, ok := s.clients[requestor.client]
	if !ok {
		log.Printf("X11: SendSelectionNotify unknown client %d", requestor.client)
		return
	}

	event := &selectionNotifyEvent{
		sequence:  client.sequence,
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

func (s *x11Server) GetRGBColor(colormap xID, pixel uint32) (r, g, b uint32) {
	if colormap.local == s.defaultColormap {
		colormap.client = 0
	}
	if cm, ok := s.colormaps[colormap]; ok {
		if color, ok := cm.pixels[pixel]; ok {
			log.Printf("GetRGBColor: cmap:%s pixel:%x return %+v", colormap, pixel, color)
			return uint32(color.Red), uint32(color.Green), uint32(color.Blue)
		}
		r = (pixel & 0xff0000) >> 16
		g = (pixel & 0x00ff00) >> 8
		b = (pixel & 0x0000ff)
		log.Printf("GetRGBColor: cmap:%s pixel:%x return RGB for pixel", colormap, pixel)
		return r, g, b
	}
	// Explicitly handle black and white pixels based on server's setup
	if pixel == s.blackPixel {
		log.Printf("GetRGBColor: cmap:%s pixel:%x return blackPixel", colormap, pixel)
		return 0, 0, 0 // Black
	}
	if pixel == s.whitePixel {
		log.Printf("GetRGBColor: cmap:%s pixel:%x return whitePixel", colormap, pixel)
		return 0xFF, 0xFF, 0xFF // White
	}
	// For TrueColor visuals, the pixel value directly encodes RGB components.
	if s.rootVisual.class == 4 { // TrueColor
		r = (pixel & s.rootVisual.redMask) >> calculateShift(s.rootVisual.redMask)
		g = (pixel & s.rootVisual.greenMask) >> calculateShift(s.rootVisual.greenMask)
		b = (pixel & s.rootVisual.blueMask) >> calculateShift(s.rootVisual.blueMask)
		log.Printf("GetRGBColor: cmap:%s pixel:%x return RGB for pixel", colormap, pixel)
		return r, g, b
	}
	// Default to black if not found
	log.Printf("GetRGBColor: cmap:%s pixel:%x return black", colormap, pixel)
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

func (s *x11Server) readRequest(client *x11Client) (*request, error) {
	var reqHeader [4]byte
	if _, err := io.ReadFull(client.conn, reqHeader[:]); err != nil {
		return nil, err
	}
	log.Printf("X11: Raw request header: %x", reqHeader)
	client.sequence++
	req := &request{
		opcode:   reqCode(reqHeader[0]),
		data:     reqHeader[1],
		length:   client.byteOrder.Uint16(reqHeader[2:4]),
		sequence: client.sequence,
	}
	req.body = make([]byte, (req.length*4)-4)
	if _, err := io.ReadFull(client.conn, req.body); err != nil {
		return nil, err
	}
	return req, nil
}

func (s *x11Server) cleanupClient(client *x11Client) {
	s.frontend.DestroyAllWindowsForClient(client.id)
	delete(s.clients, client.id)
}

func (s *x11Server) serve(client *x11Client) {
	defer client.conn.Close()
	defer s.cleanupClient(client)
	for {
		req, err := s.readRequest(client)
		if err != nil {
			if err != io.EOF {
				s.logger.Errorf("Failed to read X11 request: %v", err)
			}
			break
		}
		reply := s.handleRequest(client, req)
		if reply != nil {
			if err := client.send(reply); err != nil {
				s.logger.Errorf("Failed to write reply: %v", err)
			}
		}
	}
}
func (s *x11Server) handleRequest(client *x11Client, req *request) (reply messageEncoder) {
	log.Printf("X11: Received opcode: %d", req.opcode)
	defer func() {
		if r := recover(); r != nil {
			s.logger.Errorf("X11 Request Handler Panic: %v\n%s", r, debug.Stack())
			// Construct a generic X11 error reply (Request error)
			reply = client.sendError(GenericError{
				seq:      req.sequence,
				badValue: uint32(req.opcode),
				minorOp:  0,
				majorOp:  req.opcode,
				code:     1, // Request error code
			})
		}
	}()

	switch req.opcode {
	case CreateWindow:
		drawable, parent, x, y, width, height, _, _, _, valueMask, values := parseCreateWindowRequest(s.byteOrder, req.body)
		xid := client.xID(drawable)
		parentXID := client.xID(parent)
		// Check if the window ID is already in use
		if _, exists := s.windows[xid]; exists {
			s.logger.Errorf("X11: CreateWindow: ID %d already in use", xid)
			return client.sendError(GenericError{seq: req.sequence, badValue: drawable, majorOp: CreateWindow, code: IDChoiceError})
		}

		newWindow := &window{
			xid:        xid,
			parent:     parent,
			x:          int16(x),
			y:          int16(y),
			width:      uint16(width),
			height:     uint16(height),
			depth:      byte(req.data),
			children:   []uint32{},
			attributes: values,
		}
		if values.Colormap > 0 {
			newWindow.colormap = client.xID(values.Colormap)
		} else {
			newWindow.colormap = xID{local: s.defaultColormap}
		}
		s.windows[xid] = newWindow

		// Add to parent's children list
		if parentWindow, ok := s.windows[parentXID]; ok {
			parentWindow.children = append(parentWindow.children, drawable)
		}
		s.frontend.CreateWindow(xid, parent, x, y, width, height, uint32(req.data), valueMask, values)

	case GetWindowAttributes:
		drawable := parseGetWindowAttributesRequest(s.byteOrder, req.body)
		xid := client.xID(drawable)
		w, ok := s.windows[xid]
		if !ok {
			return nil
		}
		return &getWindowAttributesReply{
			sequence:           req.sequence,
			backingStore:       byte(w.attributes.BackingStore),
			visualID:           s.visualID,
			class:              1, // Class: InputOutput
			bitGravity:         byte(w.attributes.BitGravity),
			winGravity:         byte(w.attributes.WinGravity),
			backingPlanes:      w.attributes.BackingPlanes,
			backingPixel:       w.attributes.BackingPixel,
			saveUnder:          w.attributes.SaveUnder != 0,
			mapped:             w.mapped,
			mapState:           w.mapState(),
			overrideRedirect:   w.attributes.OverrideRedirect != 0,
			colormap:           w.attributes.Colormap,
			allEventMasks:      w.attributes.EventMask,
			yourEventMask:      w.attributes.EventMask, // Assuming client's event mask is the same for now
			doNotPropagateMask: 0,                      // Not explicitly stored in window attributes
		}
	case DestroyWindow:
		drawable := s.byteOrder.Uint32(req.body[0:4])
		xid := client.xID(drawable)
		delete(s.windows, xid)
		s.frontend.DestroyWindow(xid)

	case UnmapWindow:
		drawable := s.byteOrder.Uint32(req.body[0:4])
		xid := client.xID(drawable)
		if w, ok := s.windows[xid]; ok {
			w.mapped = false
		}
		s.frontend.UnmapWindow(xid)

	case MapWindow:
		drawable := s.byteOrder.Uint32(req.body[0:4])
		xid := client.xID(drawable)
		if w, ok := s.windows[xid]; ok {
			w.mapped = true
			s.frontend.MapWindow(xid)
			s.sendExposeEvent(xid, 0, 0, w.width, w.height)
		}

	case MapSubwindows:
		drawable := s.byteOrder.Uint32(req.body[0:4])
		xid := client.xID(drawable)
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

	case ConfigureWindow:
		drawable, valueMask, values := parseConfigureWindowRequest(s.byteOrder, req.body)
		xid := client.xID(drawable)
		s.frontend.ConfigureWindow(xid, valueMask, values)

	case GetGeometry:
		drawable := parseGetGeometryRequest(s.byteOrder, req.body)
		xid := client.xID(drawable)
		w, ok := s.windows[xid]
		if !ok {
			return nil
		}
		return &getGeometryReply{
			sequence:    req.sequence,
			depth:       w.depth,
			root:        s.rootWindowID(),
			x:           w.x,
			y:           w.y,
			width:       w.width,
			height:      w.height,
			borderWidth: 0, // Border width is not stored in window struct, assuming 0 for now
		}
	case QueryTree:
		// Not implemented yet

	case InternAtom:
		nameLen := s.byteOrder.Uint16(req.body[0:2])
		name := string(req.body[4 : 4+nameLen])
		atomID := s.frontend.GetAtom(client.id, name)

		return &internAtomReply{
			sequence: req.sequence,
			atom:     atomID,
		}

	case GetAtomName:
		atom := parseGetAtomNameRequest(s.byteOrder, req.body)
		name := s.frontend.GetAtomName(atom)
		return &getAtomNameReply{
			sequence:   req.sequence,
			nameLength: uint16(len(name)),
			name:       name,
		}

	case ChangeProperty:
		drawable := s.byteOrder.Uint32(req.body[0:4])
		property := s.byteOrder.Uint32(req.body[4:8])
		typeAtom := s.byteOrder.Uint32(req.body[8:12])
		format := req.body[12]
		dataLen := s.byteOrder.Uint32(req.body[16:20])
		propData := req.body[20 : 20+dataLen]
		xid := client.xID(drawable)
		s.frontend.ChangeProperty(xid, property, typeAtom, uint32(format), propData)

	case SendEvent:
		// The X11 client sends an event to another client.
		// We need to forward this event to the appropriate frontend.
		// For now, we'll just log it and pass it to the frontend.
		s.frontend.SendEvent(&x11RawEvent{data: req.body})

	case QueryPointer:
		drawable := parseQueryPointerRequest(s.byteOrder, req.body)
		xid := client.xID(drawable)
		log.Printf("X11: QueryPointer drawable=%d", xid)
		return &queryPointerReply{
			sequence:   req.sequence,
			sameScreen: true,
			root:       s.rootWindowID(),
			child:      drawable,
			rootX:      s.pointerX,
			rootY:      s.pointerY,
			winX:       s.pointerX, // Assuming pointer is always in the window for now
			winY:       s.pointerY, // Assuming pointer is always in the window for now
			mask:       0,          // No buttons pressed
		}
	case ListProperties:
		window := parseListPropertiesRequest(s.byteOrder, req.body)
		xid := client.xID(window)
		atoms := s.frontend.ListProperties(xid)
		return &listPropertiesReply{
			sequence:      req.sequence,
			numProperties: uint16(len(atoms)),
			atoms:         atoms,
		}

	case CreateGC:
		xid := client.xID(s.byteOrder.Uint32(req.body[0:4]))
		valueMask := s.byteOrder.Uint32(req.body[8:12])
		gc, _ := parseGCValues(s.byteOrder, valueMask, req.body[12:])

		// Check if the GC ID is already in use
		if _, exists := s.gcs[xid]; exists {
			s.logger.Errorf("X11: CreateGC: ID %s already in use", xid)
			return client.sendError(GenericError{seq: req.sequence, badValue: xid.local, majorOp: CreateGC, code: IDChoiceError})
		}

		s.gcs[xid] = gc
		s.frontend.CreateGC(xid, gc)

	case ChangeGC:
		xid := client.xID(s.byteOrder.Uint32(req.body[0:4]))
		valueMask := s.byteOrder.Uint32(req.body[4:8])
		gc, _ := parseGCValues(s.byteOrder, valueMask, req.body[8:])
		if existingGC, ok := s.gcs[xid]; ok {
			if valueMask&GCFunction != 0 {
				existingGC.Function = gc.Function
			}
			if valueMask&GCPlaneMask != 0 {
				existingGC.PlaneMask = gc.PlaneMask
			}
			if valueMask&GCForeground != 0 {
				existingGC.Foreground = gc.Foreground
			}
			if valueMask&GCBackground != 0 {
				existingGC.Background = gc.Background
			}
			if valueMask&GCLineWidth != 0 {
				existingGC.LineWidth = gc.LineWidth
			}
			if valueMask&GCLineStyle != 0 {
				existingGC.LineStyle = gc.LineStyle
			}
			if valueMask&GCCapStyle != 0 {
				existingGC.CapStyle = gc.CapStyle
			}
			if valueMask&GCJoinStyle != 0 {
				existingGC.JoinStyle = gc.JoinStyle
			}
			if valueMask&GCFillStyle != 0 {
				existingGC.FillStyle = gc.FillStyle
			}
			if valueMask&GCFillRule != 0 {
				existingGC.FillRule = gc.FillRule
			}
			if valueMask&GCTile != 0 {
				existingGC.Tile = gc.Tile
			}
			if valueMask&GCStipple != 0 {
				existingGC.Stipple = gc.Stipple
			}
			if valueMask&GCTileStipXOrigin != 0 {
				existingGC.TileStipXOrigin = gc.TileStipXOrigin
			}
			if valueMask&GCTileStipYOrigin != 0 {
				existingGC.TileStipYOrigin = gc.TileStipYOrigin
			}
			if valueMask&GCFont != 0 {
				existingGC.Font = gc.Font
			}
			if valueMask&GCSubwindowMode != 0 {
				existingGC.SubwindowMode = gc.SubwindowMode
			}
			if valueMask&GCGraphicsExposures != 0 {
				existingGC.GraphicsExposures = gc.GraphicsExposures
			}
			if valueMask&GCClipXOrigin != 0 {
				existingGC.ClipXOrigin = gc.ClipXOrigin
			}
			if valueMask&GCClipYOrigin != 0 {
				existingGC.ClipYOrigin = gc.ClipYOrigin
			}
			if valueMask&GCClipMask != 0 {
				existingGC.ClipMask = gc.ClipMask
			}
			if valueMask&GCDashOffset != 0 {
				existingGC.DashOffset = gc.DashOffset
			}
			if valueMask&GCDashList != 0 {
				existingGC.Dashes = gc.Dashes
			}
			if valueMask&GCArcMode != 0 {
				existingGC.ArcMode = gc.ArcMode
			}
		}
		s.frontend.ChangeGC(xid, valueMask, gc)

	case ClearArea:
		drawable, x, y, width, height := parseClearAreaRequest(s.byteOrder, req.body)
		s.frontend.ClearArea(client.xID(drawable), int32(x), int32(y), int32(width), int32(height))

	case CopyArea:
		srcDrawable, dstDrawable, gcID, srcX, srcY, dstX, dstY, width, height := parseCopyAreaRequest(s.byteOrder, req.body)
		gc, ok := s.gcs[client.xID(gcID)]
		if !ok {
			return
		}
		s.frontend.CopyArea(client.xID(srcDrawable), client.xID(dstDrawable), gc, int32(srcX), int32(srcY), int32(dstX), int32(dstY), int32(width), int32(height))

	case PolyPoint:
		drawable, gcID, points := parsePolyPointRequest(s.byteOrder, req.body)
		gc, ok := s.gcs[client.xID(gcID)]
		if !ok {
			return
		}
		s.frontend.PolyPoint(client.xID(drawable), gc, points)

	case PolyLine:
		drawable, gcID, points := parsePolyLineRequest(s.byteOrder, req.body)
		gc, ok := s.gcs[client.xID(gcID)]
		if !ok {
			return
		}
		s.frontend.PolyLine(client.xID(drawable), gc, points)

	case PolySegment:
		drawable, gcID, segments := parsePolySegmentRequest(s.byteOrder, req.body)
		gc, ok := s.gcs[client.xID(gcID)]
		if !ok {
			return
		}
		s.frontend.PolySegment(client.xID(drawable), gc, segments)

	case PolyArc:
		drawable, gcID, arcs := parsePolyArcRequest(s.byteOrder, req.body)
		gc, ok := s.gcs[client.xID(gcID)]
		if !ok {
			return
		}
		s.frontend.PolyArc(client.xID(drawable), gc, arcs)

	case PolyRectangle:
		drawable, gcID, rects := parsePolyRectangleRequest(s.byteOrder, req.body)
		gc, ok := s.gcs[client.xID(gcID)]
		if !ok {
			return
		}
		s.frontend.PolyRectangle(client.xID(drawable), gc, rects)

	case FillPoly:
		drawable, gcID, points := parseFillPolyRequest(s.byteOrder, req.body)
		gc, ok := s.gcs[client.xID(gcID)]
		if !ok {
			return
		}
		s.frontend.FillPoly(client.xID(drawable), gc, points)

	case PolyFillRectangle:
		drawable, gcID, rects := parsePolyFillRectangleRequest(s.byteOrder, req.body)
		gc, ok := s.gcs[client.xID(gcID)]
		if !ok {
			return
		}
		s.frontend.PolyFillRectangle(client.xID(drawable), gc, rects)

	case PolyFillArc:
		drawable, gcID, arcs := parsePolyFillArcRequest(s.byteOrder, req.body)
		gc, ok := s.gcs[client.xID(gcID)]
		if !ok {
			return
		}
		s.frontend.PolyFillArc(client.xID(drawable), gc, arcs)

	case PutImage:
		log.Printf("X11: Server received PutImage request")
		drawable, gcID, width, height, dstX, dstY, leftPad, depth, imgData := parsePutImageRequest(s.byteOrder, req.body)
		gc, ok := s.gcs[client.xID(gcID)]
		if !ok {
			return
		}
		s.frontend.PutImage(client.xID(drawable), gc, req.data, width, height, dstX, dstY, leftPad, depth, imgData)

	case GetImage:
		drawable, x, y, width, height, _ := parseGetImageRequest(s.byteOrder, req.body)
		imgData, err := s.frontend.GetImage(client.xID(drawable), int32(x), int32(y), int32(width), int32(height), uint32(req.data))
		if err != nil {
			s.logger.Errorf("Failed to get image: %v", err)
			return nil
		}
		return &getImageReply{
			sequence:  req.sequence,
			depth:     24, // Assuming 24-bit depth for now
			visualID:  s.visualID,
			imageData: imgData,
		}
	case GetProperty:
		window := s.byteOrder.Uint32(req.body[0:4])
		property := s.byteOrder.Uint32(req.body[4:8])
		longOffset := s.byteOrder.Uint32(req.body[12:16])
		longLength := s.byteOrder.Uint32(req.body[16:20])

		data, typ, format := s.frontend.GetProperty(client.xID(window), property)

		// Handle offset and length
		var propData []byte
		bytesAfter := 0
		if longOffset*4 < uint32(len(data)) {
			start := longOffset * 4
			end := start + longLength*4
			if end > uint32(len(data)) {
				end = uint32(len(data))
			}
			propData = data[start:end]
			bytesAfter = len(data) - int(end)
		} else {
			bytesAfter = len(data)
		}

		n := len(propData)
		var valueLenInFormatUnits uint32
		if format == 8 {
			valueLenInFormatUnits = uint32(n)
		} else if format == 16 {
			valueLenInFormatUnits = uint32(n / 2)
		} else if format == 32 {
			valueLenInFormatUnits = uint32(n / 4)
		}

		return &getPropertyReply{
			sequence:              req.sequence,
			format:                byte(format),
			propertyType:          typ,
			bytesAfter:            uint32(bytesAfter),
			valueLenInFormatUnits: valueLenInFormatUnits,
			value:                 propData,
		}

	case ImageText8:
		drawable, gcID, x, y, text := parseImageText8Request(s.byteOrder, req.body)
		gc, ok := s.gcs[client.xID(gcID)]
		if !ok {
			return
		}
		s.frontend.ImageText8(client.xID(drawable), gc, x, y, text)

	case ImageText16:
		drawable, gcID, x, y, text := parseImageText16Request(s.byteOrder, req.body)
		gc, ok := s.gcs[client.xID(gcID)]
		if !ok {
			return
		}
		s.frontend.ImageText16(client.xID(drawable), gc, x, y, text)

	case PolyText8:
		drawable, gcID, x, y, items := parsePolyText8Request(s.byteOrder, req.body)
		gc, ok := s.gcs[client.xID(gcID)]
		if !ok {
			return
		}
		s.frontend.PolyText8(client.xID(drawable), gc, x, y, items)

	case PolyText16:
		drawable, gcID, x, y, items := parsePolyText16Request(s.byteOrder, req.body)
		gc, ok := s.gcs[client.xID(gcID)]
		if !ok {
			return
		}
		s.frontend.PolyText16(client.xID(drawable), gc, x, y, items)

	case Bell:
		s.frontend.Bell(int8(req.data))

	case CreatePixmap:
		pid, drawable, width, height := parseCreatePixmapRequest(s.byteOrder, req.body)
		xid := client.xID(pid)
		depth := uint32(req.data)

		// Check if the pixmap ID is already in use
		if _, exists := s.pixmaps[xid]; exists {
			s.logger.Errorf("X11: CreatePixmap: ID %s already in use", xid)
			return client.sendError(GenericError{seq: req.sequence, badValue: pid, majorOp: CreatePixmap, code: IDChoiceError})
		}

		s.pixmaps[xid] = true // Mark pixmap ID as used
		s.frontend.CreatePixmap(xid, client.xID(drawable), width, height, depth)

	case FreePixmap:
		pid := s.byteOrder.Uint32(req.body[0:4])
		xid := client.xID(pid)
		delete(s.pixmaps, xid)
		s.frontend.FreePixmap(xid)

	case CreateGlyphCursor:
		cid := s.byteOrder.Uint32(req.body[0:4])
		sourceChar := s.byteOrder.Uint16(req.body[8:10])

		// Check if the cursor ID is already in use
		if _, exists := s.cursors[client.xID(cid)]; exists {
			s.logger.Errorf("X11: CreateGlyphCursor: ID %d already in use", cid)
			return client.sendError(GenericError{seq: req.sequence, badValue: cid, majorOp: CreateGlyphCursor, code: IDChoiceError})
		}

		s.cursors[client.xID(cid)] = true
		s.frontend.CreateCursorFromGlyph(cid, sourceChar)

	case ChangeWindowAttributes:
		windowID := s.byteOrder.Uint32(req.body[0:4])
		xid := client.xID(windowID)
		valueMask := s.byteOrder.Uint32(req.body[4:8])
		values, _ := parseWindowAttributes(s.byteOrder, valueMask, req.body[8:])
		if w, ok := s.windows[xid]; ok {
			if valueMask&CWBackPixmap != 0 {
				w.attributes.BackgroundPixmap = values.BackgroundPixmap
			}
			if valueMask&CWBackPixel != 0 {
				w.attributes.BackgroundPixel = values.BackgroundPixel
			}
			if valueMask&CWBorderPixmap != 0 {
				w.attributes.BorderPixmap = values.BorderPixmap
			}
			if valueMask&CWBorderPixel != 0 {
				w.attributes.BorderPixel = values.BorderPixel
			}
			if valueMask&CWBitGravity != 0 {
				w.attributes.BitGravity = values.BitGravity
			}
			if valueMask&CWWinGravity != 0 {
				w.attributes.WinGravity = values.WinGravity
			}
			if valueMask&CWBackingStore != 0 {
				w.attributes.BackingStore = values.BackingStore
			}
			if valueMask&CWBackingPlanes != 0 {
				w.attributes.BackingPlanes = values.BackingPlanes
			}
			if valueMask&CWBackingPixel != 0 {
				w.attributes.BackingPixel = values.BackingPixel
			}
			if valueMask&CWOverrideRedirect != 0 {
				w.attributes.OverrideRedirect = values.OverrideRedirect
			}
			if valueMask&CWSaveUnder != 0 {
				w.attributes.SaveUnder = values.SaveUnder
			}
			if valueMask&CWEventMask != 0 {
				w.attributes.EventMask = values.EventMask
			}
			if valueMask&CWDontPropagate != 0 {
				w.attributes.DontPropagateMask = values.DontPropagateMask
			}
			if valueMask&CWColormap != 0 {
				w.attributes.Colormap = values.Colormap
			}
			if valueMask&CWCursor != 0 {
				w.attributes.Cursor = values.Cursor
				s.frontend.SetWindowCursor(xid, client.xID(values.Cursor))
			}
		}
		s.frontend.ChangeWindowAttributes(xid, valueMask, values)

	case CopyGC:
		srcGC := s.byteOrder.Uint32(req.body[0:4])
		dstGC := s.byteOrder.Uint32(req.body[4:8])
		s.frontend.CopyGC(client.xID(srcGC), client.xID(dstGC))

	case FreeGC:
		gcID := s.byteOrder.Uint32(req.body[0:4])
		s.frontend.FreeGC(client.xID(gcID))

	case FreeCursor:
		cursorID := s.byteOrder.Uint32(req.body[0:4])
		xid := client.xID(cursorID)
		delete(s.cursors, xid)
		s.frontend.FreeCursor(xid)

	case TranslateCoords:
		srcWindow := client.xID(s.byteOrder.Uint32(req.body[0:4]))
		dstWindow := client.xID(s.byteOrder.Uint32(req.body[4:8]))
		srcX := int16(s.byteOrder.Uint16(req.body[8:10]))
		srcY := int16(s.byteOrder.Uint16(req.body[10:12]))

		// Simplified implementation: assume windows are direct children of the root
		src, srcOk := s.windows[srcWindow]
		dst, dstOk := s.windows[dstWindow]
		if !srcOk || !dstOk {
			// One of the windows doesn't exist, can't translate
			return nil
		}

		dstX := src.x + srcX - dst.x
		dstY := src.y + srcY - dst.y

		return &translateCoordsReply{
			sequence:   req.sequence,
			sameScreen: true,
			child:      0, // No child for now
			dstX:       dstX,
			dstY:       dstY,
		}

	case GetInputFocus:
		return &getInputFocusReply{
			sequence: req.sequence,
			revertTo: 1, // RevertToParent
			focus:    s.frontend.GetFocusWindow(client.id).local,
		}

	case SetSelectionOwner:
		owner := s.byteOrder.Uint32(req.body[0:4])
		selection := s.byteOrder.Uint32(req.body[4:8])
		s.selections[client.xID(selection)] = owner

	case GetSelectionOwner:
		selection := s.byteOrder.Uint32(req.body[0:4])
		owner := s.selections[client.xID(selection)]
		return &getSelectionOwnerReply{
			sequence: req.sequence,
			owner:    owner,
		}

	case ConvertSelection:
		selection := s.byteOrder.Uint32(req.body[4:8])
		target := s.byteOrder.Uint32(req.body[8:12])
		property := s.byteOrder.Uint32(req.body[12:16])
		requestor := s.byteOrder.Uint32(req.body[0:4])
		s.frontend.ConvertSelection(selection, target, property, client.xID(requestor))

	case GrabPointer:
		grabWindow := client.xID(s.byteOrder.Uint32(req.body[0:4]))
		ownerEvents := req.data != 0 // ownerEvents is in req.data
		eventMask := s.byteOrder.Uint16(req.body[4:6])
		pointerMode := req.body[6]
		keyboardMode := req.body[7]
		confineTo := s.byteOrder.Uint32(req.body[8:12])
		cursor := s.byteOrder.Uint32(req.body[12:16])
		time := s.byteOrder.Uint32(req.body[16:20])
		status := s.frontend.GrabPointer(grabWindow, ownerEvents, eventMask, pointerMode, keyboardMode, confineTo, cursor, time)
		return &grabPointerReply{
			sequence: req.sequence,
			status:   status,
		}

	case UngrabPointer:
		time := s.byteOrder.Uint32(req.body[0:4])
		s.frontend.UngrabPointer(time)

	case GrabKeyboard:
		grabWindow := client.xID(s.byteOrder.Uint32(req.body[0:4]))
		ownerEvents := req.body[4] != 0
		time := s.byteOrder.Uint32(req.body[5:9])
		pointerMode := req.body[9]
		keyboardMode := req.body[10]
		status := s.frontend.GrabKeyboard(grabWindow, ownerEvents, time, pointerMode, keyboardMode)
		return &grabKeyboardReply{
			sequence: req.sequence,
			status:   status,
		}

	case UngrabKeyboard:
		time := s.byteOrder.Uint32(req.body[0:4])
		s.frontend.UngrabKeyboard(time)

	case AllowEvents:
		mode := req.data
		time := s.byteOrder.Uint32(req.body[0:4])
		s.frontend.AllowEvents(client.id, mode, time)

	case QueryBestSize:
		// TODO: Implement proper QueryBestSize logic
		// For now, just return the requested size
		class := req.data
		drawable := s.byteOrder.Uint32(req.body[0:4])
		width := s.byteOrder.Uint16(req.body[4:6])
		height := s.byteOrder.Uint16(req.body[6:8])

		log.Printf("X11: QueryBestSize class=%d drawable=%d width=%d height=%d", class, drawable, width, height)

		return &queryBestSizeReply{
			sequence: req.sequence,
			width:    width,
			height:   height,
		}

	case CreateColormap:
		alloc, mid, _, _ := parseCreateColormapRequest(s.byteOrder, req.body)
		xid := client.xID(mid)

		if _, exists := s.colormaps[xid]; exists {
			return client.sendError(GenericError{seq: req.sequence, badValue: mid, majorOp: CreateColormap, code: ColormapError})
		}

		newColormap := &colormap{
			pixels: make(map[uint32]color),
		}

		if alloc == 1 { // All
			// For TrueColor, pre-allocating doesn't make much sense as pixels are calculated.
			// For other visual types, this would be important.
			// For now, we'll just create an empty map.
		}

		s.colormaps[xid] = newColormap

	case FreeColormap:
		mid := s.byteOrder.Uint32(req.body[0:4])
		xid := client.xID(mid)
		if _, ok := s.colormaps[xid]; !ok {
			return client.sendError(GenericError{seq: req.sequence, badValue: mid, majorOp: FreeColormap, code: ColormapError})
		}
		delete(s.colormaps, xid)

	case QueryExtension:
		nameLen := s.byteOrder.Uint16(req.body[0:2])
		name := string(req.body[4 : 4+nameLen])
		log.Printf("X11: QueryExtension name=%s", name)

		return &queryExtensionReply{
			sequence:    req.sequence,
			present:     false,
			majorOpcode: 0,
			firstEvent:  0,
			firstError:  0,
		}

	case StoreNamedColor:
		log.Print("StoreNamedColor: not implemented")

	case StoreColors:
		cmapID := s.byteOrder.Uint32(req.body[0:4])
		xid := client.xID(cmapID)
		cm, ok := s.colormaps[xid]
		if !ok {
			return client.sendError(GenericError{seq: req.sequence, badValue: cmapID, majorOp: StoreColors, code: ColormapError})
		}

		numItems := (len(req.body) - 4) / 12
		for i := 0; i < numItems; i++ {
			offset := 4 + i*12
			pixel := s.byteOrder.Uint32(req.body[offset : offset+4])
			red := s.byteOrder.Uint16(req.body[offset+4 : offset+6])
			green := s.byteOrder.Uint16(req.body[offset+6 : offset+8])
			blue := s.byteOrder.Uint16(req.body[offset+8 : offset+10])
			flags := req.body[offset+10]

			c, exists := cm.pixels[pixel]
			if !exists {
				c = color{}
			}

			if flags&DoRed != 0 {
				c.Red = red
			}
			if flags&DoGreen != 0 {
				c.Green = green
			}
			if flags&DoBlue != 0 {
				c.Blue = blue
			}
			cm.pixels[pixel] = c
		}

	case AllocNamedColor:
		p := parseAllocNamedColorRequest(s.byteOrder, req.body, req.sequence, req.data, req.opcode)
		return s.handleAllocNamedColor(client, p)

	case QueryColors:
		p := parseQueryColorsRequest(s.byteOrder, req.body)
		cmapID := p.Cmap
		pixels := p.Pixels

		var colors []color
		for _, pixel := range pixels {
			color, ok := s.colormaps[cmapID].pixels[pixel]
			if !ok {
				return client.sendError(GenericError{seq: req.sequence, badValue: pixel, majorOp: QueryColors, code: ValueError})
			}
			colors = append(colors, color)
		}

		return &queryColorsReply{
			sequence: req.sequence,
			colors:   colors,
		}

	case LookupColor:
		cmapIDVal, nameStr := parseLookupColorRequest(s.byteOrder, req.body)
		cmapID := xID{local: cmapIDVal}
		name := nameStr

		color, ok := lookupColor(name)
		if !ok {
			// TODO: This should be BadName, not BadColor
			return client.sendError(GenericError{seq: req.sequence, badValue: cmapID.local, majorOp: LookupColor, code: ColormapError})
		}

		return &lookupColorReply{
			sequence:   req.sequence,
			red:        scale8to16(color.Red),
			green:      scale8to16(color.Green),
			blue:       scale8to16(color.Blue),
			exactRed:   scale8to16(color.Red),
			exactGreen: scale8to16(color.Green),
			exactBlue:  scale8to16(color.Blue),
		}

	case AllocColor:
		cmapID, red, green, blue := parseAllocColorRequest(s.byteOrder, req.body)

		xid := client.xID(cmapID)
		cm, ok := s.colormaps[xid]
		if !ok {
			return client.sendError(GenericError{seq: req.sequence, badValue: cmapID, majorOp: AllocColor, code: ColormapError})
		}

		// Simple allocation for TrueColor: construct pixel value from RGB
		r8 := byte(red >> 8)
		g8 := byte(green >> 8)
		b8 := byte(blue >> 8)
		pixel := (uint32(r8) << 16) | (uint32(g8) << 8) | uint32(b8)

		cm.pixels[pixel] = color{Red: red, Green: green, Blue: blue}

		return &allocColorReply{
			sequence: req.sequence,
			red:      red,
			green:    green,
			blue:     blue,
			pixel:    pixel,
		}

	case ListFonts:
		maxNames := s.byteOrder.Uint16(req.body[0:2])
		nameLen := s.byteOrder.Uint16(req.body[2:4])
		pattern := string(req.body[4 : 4+nameLen])

		fontNames := s.frontend.ListFonts(maxNames, pattern)

		return &listFontsReply{
			sequence:  req.sequence,
			numFonts:  uint16(len(fontNames)),
			fontNames: fontNames,
		}

	case OpenFont:
		fid, name := parseOpenFontRequest(s.byteOrder, req.body)
		s.frontend.OpenFont(client.xID(fid), name)

	case CloseFont:
		fid := s.byteOrder.Uint32(req.body[0:4])
		s.frontend.CloseFont(client.xID(fid))

	case QueryFont:
		fid := s.byteOrder.Uint32(req.body[0:4])
		minBounds, maxBounds, minCharOrByte2, maxCharOrByte2, defaultChar, drawDirection, minByte1, maxByte1, allCharsExist, fontAscent, fontDescent, charInfos := s.frontend.QueryFont(client.xID(fid))

		return &queryFontReply{
			sequence:       req.sequence,
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

	case FreeColors:
		cmapID := s.byteOrder.Uint32(req.body[0:4])
		// planeMask := s.byteOrder.Uint32(req.body[4:8]) // Not used for now
		xid := client.xID(cmapID)
		cm, ok := s.colormaps[xid]
		if !ok {
			return client.sendError(GenericError{seq: req.sequence, badValue: cmapID, majorOp: FreeColors, code: ColormapError})
		}

		numPixels := (len(req.body) - 8) / 4
		for i := 0; i < numPixels; i++ {
			offset := 8 + i*4
			pixel := s.byteOrder.Uint32(req.body[offset : offset+4])
			delete(cm.pixels, pixel)
		}

	case InstallColormap:
		cmapID := s.byteOrder.Uint32(req.body[0:4])
		xid := client.xID(cmapID)
		_, ok := s.colormaps[xid]
		if !ok {
			return client.sendError(GenericError{seq: req.sequence, badValue: cmapID, majorOp: InstallColormap, code: ColormapError})
		}

		s.installedColormap = xid

		for winID, win := range s.windows {
			if win.colormap == xid {
				client, ok := s.clients[winID.client]
				if !ok {
					log.Printf("X11: InstallColormap unknown client %d", winID.client)
					continue
				}
				event := &colormapNotifyEvent{
					sequence: client.sequence,
					window:   winID.local,
					colormap: cmapID,
					new:      true,
					state:    0, // Installed
				}
				s.sendEvent(client, event)
			}
		}
		return nil

	case UninstallColormap:
		cmapID := s.byteOrder.Uint32(req.body[0:4])
		xid := client.xID(cmapID)
		_, ok := s.colormaps[xid]
		if !ok {
			return client.sendError(GenericError{seq: req.sequence, badValue: cmapID, majorOp: UninstallColormap, code: ColormapError})
		}

		if s.installedColormap == xid {
			s.installedColormap = xID{local: s.defaultColormap}
		}

		for winID, win := range s.windows {
			if win.colormap == xid {
				client, ok := s.clients[winID.client]
				if !ok {
					log.Printf("X11: UninstallColormap unknown client %d", winID.client)
					continue
				}
				event := &colormapNotifyEvent{
					sequence: client.sequence,
					window:   winID.local,
					colormap: cmapID,
					new:      false,
					state:    1, // Uninstalled
				}
				s.sendEvent(client, event)
			}
		}
		return nil

	case ListInstalledColormaps:
		// windowID := s.byteOrder.Uint32(req.body[0:4]) // Not used for now

		var colormaps []uint32
		if s.installedColormap.local != 0 {
			colormaps = append(colormaps, s.installedColormap.local)
		}

		return &listInstalledColormapsReply{
			sequence:     req.sequence,
			numColormaps: uint16(len(colormaps)),
			colormaps:    colormaps,
		}

	default:
		log.Printf("Unknown X11 request opcode: %d", req.opcode)
	}
	return nil
}

func (s *x11Server) handleAllocNamedColor(client *x11Client, p AllocNamedColorRequest) messageEncoder {
	if _, ok := s.colormaps[p.Cmap]; !ok {
		return client.sendError(BadColor{
			seq:      p.Sequence,
			badValue: p.Cmap.local,
			minorOp:  p.MinorOp,
			majorOp:  p.MajorOp,
		})
	}

	name := string(p.Name)
	rgb, ok := lookupColor(name)
	if !ok {
		// TODO: This should be BadName, not BadColor
		return client.sendError(BadColor{
			seq:      p.Sequence,
			badValue: p.Cmap.local, // TODO: This should be the atom for the name, not the colormap
			minorOp:  p.MinorOp,
			majorOp:  p.MajorOp,
		})
	}

	exactRed := scale8to16(rgb.Red)
	exactGreen := scale8to16(rgb.Green)
	exactBlue := scale8to16(rgb.Blue)

	// For now, we only support TrueColor visuals, so we just allocate the color directly.
	// TODO: Implement proper colormap handling.
	pixel := (uint32(rgb.Red) << 16) | (uint32(rgb.Green) << 8) | uint32(rgb.Blue)

	return &allocColorReply{
		sequence: p.Sequence,
		red:      exactRed,
		green:    exactGreen,
		blue:     exactBlue,
		pixel:    pixel,
	}
}

func boolToByte(b bool) byte {
	if b {
		return 1
	}
	return 0
}

func (s *x11Server) handshake(client *x11Client) {
	var handshake [12]byte
	if _, err := io.ReadFull(client.conn, handshake[:]); err != nil {
		s.logger.Errorf("x11 handshake: %v", err)
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
	authLen := authProtoNameLen + authProtoDataLen
	if pad := authLen % 4; pad != 0 {
		authLen += 4 - pad
	}
	if _, err := io.CopyN(io.Discard, client.conn, int64(authLen)); err != nil {
		s.logger.Errorf("Failed to discard auth details: %v", err)
		return
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
func newDefaultSetup() *setup {
	vendorString := "sshterm"
	s := &setup{
		releaseNumber:            1,
		resourceIDBase:           0,
		resourceIDMask:           0x1FFFFF,
		motionBufferSize:         256,
		vendorLength:             uint16(len(vendorString)),
		maxRequestLength:         0xFFFF,
		numScreens:               1,
		numPixmapFormats:         1,
		imageByteOrder:           0, // LSBFirst
		bitmapFormatBitOrder:     0, // LeastSignificant
		bitmapFormatScanlineUnit: 8,
		bitmapFormatScanlinePad:  8,
		minKeycode:               8,
		maxKeycode:               255,
		vendorString:             vendorString,
		pixmapFormats: []format{
			{
				depth:        24,
				bitsPerPixel: 32,
				scanlinePad:  32,
			},
		},
		screens: []screen{
			{
				root:                0,
				defaultColormap:     1,
				whitePixel:          0xffffff,
				blackPixel:          0x000000,
				currentInputMasks:   0,
				widthInPixels:       1024,
				heightInPixels:      768,
				widthInMillimeters:  270,
				heightInMillimeters: 203,
				minInstalledMaps:    1,
				maxInstalledMaps:    1,
				rootVisual:          0x1,
				backingStores:       2, // Always
				saveUnders:          false,
				rootDepth:           24,
				numDepths:           1,
				depths: []depth{
					{
						depth:      24,
						numVisuals: 1,
						visuals: []visualType{
							{
								visualID:        0x1,
								class:           4, // TrueColor
								bitsPerRGBValue: 8,
								colormapEntries: 256,
								redMask:         0xff0000,
								greenMask:       0x00ff00,
								blueMask:        0x0000ff,
							},
						},
					},
				},
			},
		},
	}
	return s
}

func (s *setup) marshal(order binary.ByteOrder) []byte {
	buf := new(bytes.Buffer)
	binary.Write(buf, order, s.releaseNumber)
	binary.Write(buf, order, s.resourceIDBase)
	binary.Write(buf, order, s.resourceIDMask)
	binary.Write(buf, order, s.motionBufferSize)
	binary.Write(buf, order, s.vendorLength)
	binary.Write(buf, order, s.maxRequestLength)
	buf.WriteByte(s.numScreens)
	buf.WriteByte(s.numPixmapFormats)
	buf.WriteByte(s.imageByteOrder)
	buf.WriteByte(s.bitmapFormatBitOrder)
	buf.WriteByte(s.bitmapFormatScanlineUnit)
	buf.WriteByte(s.bitmapFormatScanlinePad)
	buf.WriteByte(s.minKeycode)
	buf.WriteByte(s.maxKeycode)
	buf.Write([]byte{0, 0, 0, 0}) // 4 bytes of padding
	buf.WriteString(s.vendorString)
	if pad := (4 - (len(s.vendorString) % 4)) % 4; pad > 0 {
		buf.Write(make([]byte, pad))
	}
	for _, f := range s.pixmapFormats {
		f.marshal(buf, order)
	}
	for _, scr := range s.screens {
		scr.marshal(buf, order)
	}
	return buf.Bytes()
}

func (f *format) marshal(buf *bytes.Buffer, order binary.ByteOrder) {
	buf.WriteByte(f.depth)
	buf.WriteByte(f.bitsPerPixel)
	buf.WriteByte(f.scanlinePad)
	buf.Write([]byte{0, 0, 0, 0, 0}) // 5 bytes of padding
}

func (s *screen) marshal(buf *bytes.Buffer, order binary.ByteOrder) {
	binary.Write(buf, order, s.root)
	binary.Write(buf, order, s.defaultColormap)
	binary.Write(buf, order, s.whitePixel)
	binary.Write(buf, order, s.blackPixel)
	binary.Write(buf, order, s.currentInputMasks)
	binary.Write(buf, order, s.widthInPixels)
	binary.Write(buf, order, s.heightInPixels)
	binary.Write(buf, order, s.widthInMillimeters)
	binary.Write(buf, order, s.heightInMillimeters)
	binary.Write(buf, order, s.minInstalledMaps)
	binary.Write(buf, order, s.maxInstalledMaps)
	binary.Write(buf, order, s.rootVisual)
	buf.WriteByte(s.backingStores)
	if s.saveUnders {
		buf.WriteByte(1)
	} else {
		buf.WriteByte(0)
	}
	buf.WriteByte(s.rootDepth)
	buf.WriteByte(s.numDepths)
	for _, d := range s.depths {
		d.marshal(buf, order)
	}
}

func (d *depth) marshal(buf *bytes.Buffer, order binary.ByteOrder) {
	buf.WriteByte(d.depth)
	buf.WriteByte(0) // padding
	binary.Write(buf, order, d.numVisuals)
	buf.Write([]byte{0, 0, 0, 0}) // 4 bytes of padding
	for _, v := range d.visuals {
		v.marshal(buf, order)
	}
}

func (v *visualType) marshal(buf *bytes.Buffer, order binary.ByteOrder) {
	binary.Write(buf, order, v.visualID)
	buf.WriteByte(v.class)
	buf.WriteByte(v.bitsPerRGBValue)
	binary.Write(buf, order, v.colormapEntries)
	binary.Write(buf, order, v.redMask)
	binary.Write(buf, order, v.greenMask)
	binary.Write(buf, order, v.blueMask)
	buf.Write([]byte{0, 0, 0, 0}) // 4 bytes of padding
}

func HandleX11Forwarding(logger Logger, client *ssh.Client) {
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
					gcs:        make(map[xID]*GC),
					pixmaps:    make(map[xID]bool),
					cursors:    make(map[xID]bool),
					selections: make(map[xID]uint32),
					colormaps: map[xID]*colormap{
						xID{local: 0x1}: {
							pixels: map[uint32]color{
								0x000000: color{0x00, 0x00, 0x00},
								1:        color{0xff, 0xff, 0xff},
								0xffffff: color{0xff, 0xff, 0xff},
							},
						},
					},
					defaultColormap: 0x1,
					clients:         make(map[uint32]*x11Client),
					nextClientID:    1,
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
