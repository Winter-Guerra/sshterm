//go:build x11

package x11

import (
	"bytes"
	"encoding/binary"
)

type xCharInfo struct {
	LeftSideBearing  int16
	RightSideBearing int16
	CharacterWidth   uint16
	Ascent           int16
	Descent          int16
	Attributes       uint16
}

func boolToByte(b bool) byte {
	if b {
		return 1
	}
	return 0
}

// GetWindowAttributes: 3
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

// GetGeometry: 14
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

// InternAtom: 16
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

// GetAtomName: 17
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

// GetProperty: 20
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

// ListProperties: 21
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

// QueryTextExtents: 48
type queryTextExtentsReply struct {
	sequence       uint16
	drawDirection  byte
	fontAscent     int16
	fontDescent    int16
	overallAscent  int16
	overallDescent int16
	overallWidth   int32
	overallLeft    int32
	overallRight   int32
}

func (r *queryTextExtentsReply) encodeMessage(order binary.ByteOrder) []byte {
	reply := make([]byte, 32)
	reply[0] = 1 // Reply type
	reply[1] = r.drawDirection
	order.PutUint16(reply[2:4], r.sequence)
	order.PutUint32(reply[4:8], 0)
	order.PutUint16(reply[8:10], uint16(r.fontAscent))
	order.PutUint16(reply[10:12], uint16(r.fontDescent))
	order.PutUint16(reply[12:14], uint16(r.overallAscent))
	order.PutUint16(reply[14:16], uint16(r.overallDescent))
	order.PutUint32(reply[16:20], uint32(r.overallWidth))
	order.PutUint32(reply[20:24], uint32(r.overallLeft))
	order.PutUint32(reply[24:28], uint32(r.overallRight))
	return reply
}

// GetMotionEvents: 39
type getMotionEventsReply struct {
	sequence uint16
	nEvents  uint32
	events   []TimeCoord
}

type TimeCoord struct {
	Time uint32
	X, Y int16
}

func (r *getMotionEventsReply) encodeMessage(order binary.ByteOrder) []byte {
	reply := make([]byte, 32+len(r.events)*8)
	reply[0] = 1 // Reply type
	order.PutUint16(reply[2:4], r.sequence)
	order.PutUint32(reply[4:8], uint32(len(r.events)*2))
	order.PutUint32(reply[8:12], r.nEvents)
	for i, event := range r.events {
		order.PutUint32(reply[32+i*8:], event.Time)
		order.PutUint16(reply[32+i*8+4:], uint16(event.X))
		order.PutUint16(reply[32+i*8+6:], uint16(event.Y))
	}
	return reply
}

// GetSelectionOwner: 23
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

// GrabPointer: 26
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

// GrabKeyboard: 31
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

// QueryPointer: 38
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

// TranslateCoords: 40
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

// GetInputFocus: 43
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

// QueryFont: 47
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

// ListFonts: 50
type listFontsReply struct {
	sequence  uint16
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

// GetImage: 73
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

// AllocColor: 84
type allocColorReply struct {
	sequence uint16
	red      uint16
	green    uint16
	blue     uint16
	pixel    uint32
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
	order.PutUint32(reply[16:20], r.pixel)
	// reply[20:32] is padding
	return reply
}

// AllocNamedColor: 85
type allocNamedColorReply struct {
	sequence uint16
	red      uint16
	green    uint16
	blue     uint16
	pixel    uint32
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
func (r *allocNamedColorReply) encodeMessage(order binary.ByteOrder) []byte {
	reply := make([]byte, 32)
	reply[0] = 1 // Reply type
	// byte 1 is unused
	order.PutUint16(reply[2:4], r.sequence)
	order.PutUint32(reply[4:8], 0) // Reply length (0 * 4 bytes = 0 bytes, plus 32 bytes header = 32 bytes total)
	order.PutUint32(reply[8:12], r.pixel)
	order.PutUint16(reply[12:14], r.red)
	order.PutUint16(reply[14:16], r.green)
	order.PutUint16(reply[16:18], r.blue)
	order.PutUint16(reply[18:20], r.red)
	order.PutUint16(reply[20:22], r.green)
	order.PutUint16(reply[22:24], r.blue)
	return reply
}

// ListInstalledColormaps: 85
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

// QueryColors: 91
type queryColorsReply struct {
	sequence uint16
	colors   []xColorItem
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

// LookupColor: 92
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

// QueryBestSize: 97
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

// QueryExtension: 98
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

// SetPointerMapping: 116
type setPointerMappingReply struct {
	sequence uint16
	status   byte
}

func (r *setPointerMappingReply) encodeMessage(order binary.ByteOrder) []byte {
	reply := make([]byte, 32)
	reply[0] = 1 // Reply type
	reply[1] = r.status
	order.PutUint16(reply[2:4], r.sequence)
	order.PutUint32(reply[4:8], 0)
	return reply
}

// GetPointerMapping: 117
type getPointerMappingReply struct {
	sequence uint16
	length   byte
	pMap     []byte
}

func (r *getPointerMappingReply) encodeMessage(order binary.ByteOrder) []byte {
	reply := make([]byte, 32+len(r.pMap))
	reply[0] = 1 // Reply type
	reply[1] = r.length
	order.PutUint16(reply[2:4], r.sequence)
	order.PutUint32(reply[4:8], uint32((len(r.pMap)+3)/4))
	copy(reply[32:], r.pMap)
	return reply
}

// GetKeyboardMapping: 101
type getKeyboardMappingReply struct {
	sequence uint16
	keySyms  []uint32
}

func (r *getKeyboardMappingReply) encodeMessage(order binary.ByteOrder) []byte {
	reply := make([]byte, 32+len(r.keySyms)*4)
	reply[0] = 1 // Reply type
	reply[1] = byte(len(r.keySyms))
	order.PutUint16(reply[2:4], r.sequence)
	order.PutUint32(reply[4:8], uint32(len(r.keySyms)))
	for i, keySym := range r.keySyms {
		order.PutUint32(reply[32+i*4:], keySym)
	}
	return reply
}

// GetKeyboardControl: 103
type getKeyboardControlReply struct {
	sequence         uint16
	keyClickPercent  byte
	bellPercent      byte
	bellPitch        uint16
	bellDuration     uint16
	ledMask          uint32
	globalAutoRepeat byte
	autoRepeats      [32]byte
}

func (r *getKeyboardControlReply) encodeMessage(order binary.ByteOrder) []byte {
	reply := make([]byte, 52)
	reply[0] = 1 // Reply type
	reply[1] = r.globalAutoRepeat
	order.PutUint16(reply[2:4], r.sequence)
	order.PutUint32(reply[4:8], 5) // Reply length
	order.PutUint32(reply[8:12], r.ledMask)
	reply[12] = r.keyClickPercent
	reply[13] = r.bellPercent
	order.PutUint16(reply[14:16], r.bellPitch)
	order.PutUint16(reply[16:18], r.bellDuration)
	// reply[18:20] is padding
	copy(reply[20:52], r.autoRepeats[:])
	return reply
}

// GetScreenSaver: 108
type getScreenSaverReply struct {
	sequence    uint16
	timeout     uint16
	interval    uint16
	preferBlank byte
	allowExpose byte
}

func (r *getScreenSaverReply) encodeMessage(order binary.ByteOrder) []byte {
	reply := make([]byte, 32)
	reply[0] = 1 // Reply type
	order.PutUint16(reply[2:4], r.sequence)
	order.PutUint32(reply[4:8], 0)
	order.PutUint16(reply[8:10], r.timeout)
	order.PutUint16(reply[10:12], r.interval)
	reply[12] = r.preferBlank
	reply[13] = r.allowExpose
	return reply
}

// ListHosts: 110
type listHostsReply struct {
	sequence uint16
	numHosts uint16
	hosts    []Host
}

func (r *listHostsReply) encodeMessage(order binary.ByteOrder) []byte {
	var data []byte
	for _, host := range r.hosts {
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
	order.PutUint16(reply[2:4], r.sequence)
	order.PutUint32(reply[4:8], uint32(len(data)/4))
	order.PutUint16(reply[8:10], r.numHosts)
	copy(reply[32:], data)
	return reply
}

// SetModifierMapping: 118
type setModifierMappingReply struct {
	sequence uint16
	status   byte
}

func (r *setModifierMappingReply) encodeMessage(order binary.ByteOrder) []byte {
	reply := make([]byte, 32)
	reply[0] = 1 // Reply type
	reply[1] = r.status
	order.PutUint16(reply[2:4], r.sequence)
	order.PutUint32(reply[4:8], 0)
	return reply
}

// GetModifierMapping: 119
type getModifierMappingReply struct {
	sequence            uint16
	keyCodesPerModifier byte
	keyCodes            []KeyCode
}

func (r *getModifierMappingReply) encodeMessage(order binary.ByteOrder) []byte {
	keyCodes := make([]byte, len(r.keyCodes))
	for i, kc := range r.keyCodes {
		keyCodes[i] = byte(kc)
	}
	reply := make([]byte, 32+len(keyCodes))
	reply[0] = 1 // Reply type
	reply[1] = r.keyCodesPerModifier
	order.PutUint16(reply[2:4], r.sequence)
	order.PutUint32(reply[4:8], uint32((len(keyCodes)+3)/4))
	copy(reply[32:], keyCodes)
	return reply
}

// QueryKeymap: 44
type queryKeymapReply struct {
	sequence uint16
	keys     [32]byte
}

func (r *queryKeymapReply) encodeMessage(order binary.ByteOrder) []byte {
	reply := make([]byte, 40)
	reply[0] = 1 // Reply type
	order.PutUint16(reply[2:4], r.sequence)
	order.PutUint32(reply[4:8], 2)
	copy(reply[8:], r.keys[:])
	return reply
}

// GetFontPath: 52
type getFontPathReply struct {
	sequence uint16
	nPaths   uint16
	paths    []string
}

func (r *getFontPathReply) encodeMessage(order binary.ByteOrder) []byte {
	var data []byte
	for _, path := range r.paths {
		data = append(data, byte(len(path)))
		data = append(data, path...)
	}
	p := (4 - (len(data) % 4)) % 4
	totalLen := 32 + len(data) + p

	reply := make([]byte, totalLen)
	reply[0] = 1 // Reply type
	order.PutUint16(reply[2:4], r.sequence)
	order.PutUint32(reply[4:8], uint32((len(data)+p)/4))
	order.PutUint16(reply[8:10], r.nPaths)
	copy(reply[32:], data)
	return reply
}

// ListFontsWithInfo: 50
type listFontsWithInfoReply struct {
	sequence      uint16
	nameLength    byte
	minBounds     xCharInfo
	maxBounds     xCharInfo
	minChar       uint16
	maxChar       uint16
	defaultChar   uint16
	nFontProps    uint16
	drawDirection byte
	minByte1      byte
	maxByte1      byte
	allCharsExist bool
	fontAscent    int16
	fontDescent   int16
	nReplies      uint32
	fontProps     []FontProp
	fontName      string
}

type FontProp struct {
	Name  uint32
	Value uint32
}

func (r *listFontsWithInfoReply) encodeMessage(order binary.ByteOrder) []byte {
	fontNameBytes := []byte(r.fontName)
	fontNameLen := len(fontNameBytes)
	p := (4 - (fontNameLen % 4)) % 4
	totalLen := 60 + len(r.fontProps)*8 + fontNameLen + p

	reply := make([]byte, totalLen)
	reply[0] = 1 // Reply type
	reply[1] = r.nameLength
	order.PutUint16(reply[2:4], r.sequence)
	order.PutUint32(reply[4:8], uint32((totalLen-32)/4))

	// min-bounds
	order.PutUint16(reply[8:10], uint16(r.minBounds.LeftSideBearing))
	order.PutUint16(reply[10:12], uint16(r.minBounds.RightSideBearing))
	order.PutUint16(reply[12:14], r.minBounds.CharacterWidth)
	order.PutUint16(reply[14:16], uint16(r.minBounds.Ascent))
	order.PutUint16(reply[16:18], uint16(r.minBounds.Descent))
	order.PutUint16(reply[18:20], r.minBounds.Attributes)

	// max-bounds
	order.PutUint16(reply[24:26], uint16(r.maxBounds.LeftSideBearing))
	order.PutUint16(reply[26:28], uint16(r.maxBounds.RightSideBearing))
	order.PutUint16(reply[28:30], r.maxBounds.CharacterWidth)
	order.PutUint16(reply[30:32], uint16(r.maxBounds.Ascent))
	order.PutUint16(reply[32:34], uint16(r.maxBounds.Descent))
	order.PutUint16(reply[34:36], r.maxBounds.Attributes)

	order.PutUint16(reply[40:42], r.minChar)
	order.PutUint16(reply[42:44], r.maxChar)
	order.PutUint16(reply[44:46], r.defaultChar)
	order.PutUint16(reply[46:48], r.nFontProps)
	reply[48] = r.drawDirection
	reply[49] = r.minByte1
	reply[50] = r.maxByte1
	reply[51] = boolToByte(r.allCharsExist)
	order.PutUint16(reply[52:54], uint16(r.fontAscent))
	order.PutUint16(reply[54:56], uint16(r.fontDescent))
	order.PutUint32(reply[56:60], r.nReplies)

	offset := 60
	for _, prop := range r.fontProps {
		order.PutUint32(reply[offset:offset+4], prop.Name)
		order.PutUint32(reply[offset+4:offset+8], prop.Value)
		offset += 8
	}

	copy(reply[offset:], fontNameBytes)
	return reply
}

// QueryTree: 15
type queryTreeReply struct {
	sequence  uint16
	root      uint32
	parent    uint32
	nChildren uint16
	children  []uint32
}

func (r *queryTreeReply) encodeMessage(order binary.ByteOrder) []byte {
	reply := make([]byte, 32+len(r.children)*4)
	reply[0] = 1 // Reply type
	order.PutUint16(reply[2:4], r.sequence)
	order.PutUint32(reply[4:8], uint32(len(r.children)))
	order.PutUint32(reply[8:12], r.root)
	order.PutUint32(reply[12:16], r.parent)
	order.PutUint16(reply[16:18], r.nChildren)
	for i, child := range r.children {
		order.PutUint32(reply[32+i*4:], child)
	}
	return reply
}

// AllocColorCells: 86
type allocColorCellsReply struct {
	sequence uint16
	nPixels  uint16
	nMasks   uint16
	pixels   []uint32
	masks    []uint32
}

func (r *allocColorCellsReply) encodeMessage(order binary.ByteOrder) []byte {
	numPixels := len(r.pixels)
	numMasks := len(r.masks)
	reply := make([]byte, 32+(numPixels+numMasks)*4)
	reply[0] = 1 // Reply type
	// byte 1 is unused
	order.PutUint16(reply[2:4], r.sequence)
	order.PutUint32(reply[4:8], uint32(numPixels+numMasks)) // Reply length
	order.PutUint16(reply[8:10], uint16(numPixels))
	order.PutUint16(reply[10:12], uint16(numMasks))
	// reply[12:32] is padding
	for i, pixel := range r.pixels {
		order.PutUint32(reply[32+i*4:], pixel)
	}
	for i, mask := range r.masks {
		order.PutUint32(reply[32+numPixels*4+i*4:], mask)
	}
	return reply
}

// AllocColorPlanes: 87
type allocColorPlanesReply struct {
	sequence  uint16
	nPixels   uint16
	redMask   uint32
	greenMask uint32
	blueMask  uint32
	pixels    []uint32
}

func (r *allocColorPlanesReply) encodeMessage(order binary.ByteOrder) []byte {
	numPixels := len(r.pixels)
	reply := make([]byte, 32+numPixels*4)
	reply[0] = 1 // Reply type
	// byte 1 is unused
	order.PutUint16(reply[2:4], r.sequence)
	order.PutUint32(reply[4:8], uint32(numPixels)) // Reply length
	order.PutUint16(reply[8:10], uint16(numPixels))
	// reply[10:12] is padding
	order.PutUint32(reply[12:16], r.redMask)
	order.PutUint32(reply[16:20], r.greenMask)
	order.PutUint32(reply[20:24], r.blueMask)
	// reply[24:32] is padding
	for i, pixel := range r.pixels {
		order.PutUint32(reply[32+i*4:], pixel)
	}
	return reply
}

// ListExtensions: 99
type listExtensionsReply struct {
	sequence uint16
	nNames   byte
	names    []string
}

func (r *listExtensionsReply) encodeMessage(order binary.ByteOrder) []byte {
	var data []byte
	for _, name := range r.names {
		data = append(data, byte(len(name)))
		data = append(data, name...)
	}
	reply := make([]byte, 32+len(data))
	reply[0] = 1 // Reply type
	reply[1] = r.nNames
	order.PutUint16(reply[2:4], r.sequence)
	order.PutUint32(reply[4:8], uint32((len(data)+3)/4))
	copy(reply[32:], data)
	return reply
}

// GetPointerControl: 106
type getPointerControlReply struct {
	sequence         uint16
	accelNumerator   uint16
	accelDenominator uint16
	threshold        uint16
	doAccel          bool
	doThreshold      bool
}

func (r *getPointerControlReply) encodeMessage(order binary.ByteOrder) []byte {
	reply := make([]byte, 32)
	reply[0] = 1 // Reply type
	// byte 1 is unused
	order.PutUint16(reply[2:4], r.sequence)
	order.PutUint32(reply[4:8], 0) // Reply length
	order.PutUint16(reply[8:10], r.accelNumerator)
	order.PutUint16(reply[10:12], r.accelDenominator)
	order.PutUint16(reply[12:14], r.threshold)
	// reply[14:32] is padding
	return reply
}
