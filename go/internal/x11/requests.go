//go:build x11

package x11

import (
	"encoding/binary"
	"log"
)

type WindowAttributes struct {
	BackgroundPixmap   uint32
	BackgroundPixel    uint32
	BackgroundPixelSet bool
	BorderPixmap       uint32
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
	Colormap           uint32
	Cursor             uint32
}

func parseWindowAttributes(order binary.ByteOrder, valueMask uint32, body []byte) (*WindowAttributes, int) {
	attributes := &WindowAttributes{
		BackgroundPixel: 1,
	}
	read := 0
	if valueMask&CWBackPixmap != 0 {
		attributes.BackgroundPixmap = order.Uint32(body[read : read+4])
		read += 4
	}
	if valueMask&CWBackPixel != 0 {
		attributes.BackgroundPixel = order.Uint32(body[read : read+4])
		attributes.BackgroundPixelSet = true
		read += 4
	}
	if valueMask&CWBorderPixmap != 0 {
		attributes.BorderPixmap = order.Uint32(body[read : read+4])
		read += 4
	}
	if valueMask&CWBorderPixel != 0 {
		attributes.BorderPixel = order.Uint32(body[read : read+4])
		read += 4
	}
	if valueMask&CWBitGravity != 0 {
		attributes.BitGravity = order.Uint32(body[read : read+4])
		read += 4
	}
	if valueMask&CWWinGravity != 0 {
		attributes.WinGravity = order.Uint32(body[read : read+4])
		read += 4
	}
	if valueMask&CWBackingStore != 0 {
		attributes.BackingStore = order.Uint32(body[read : read+4])
		read += 4
	}
	if valueMask&CWBackingPlanes != 0 {
		attributes.BackingPlanes = order.Uint32(body[read : read+4])
		read += 4
	}
	if valueMask&CWBackingPixel != 0 {
		attributes.BackingPixel = order.Uint32(body[read : read+4])
		read += 4
	}
	if valueMask&CWOverrideRedirect != 0 {
		attributes.OverrideRedirect = order.Uint32(body[read : read+4])
		read += 4
	}
	if valueMask&CWSaveUnder != 0 {
		attributes.SaveUnder = order.Uint32(body[read : read+4])
		read += 4
	}
	if valueMask&CWEventMask != 0 {
		attributes.EventMask = order.Uint32(body[read : read+4])
		read += 4
	}
	if valueMask&CWDontPropagate != 0 {
		attributes.DontPropagateMask = order.Uint32(body[read : read+4])
		read += 4
	}
	if valueMask&CWColormap != 0 {
		attributes.Colormap = order.Uint32(body[read : read+4])
		read += 4
	}
	if valueMask&CWCursor != 0 {
		attributes.Cursor = order.Uint32(body[read : read+4])
		read += 4
	}
	return attributes, read
}

func parseCreateWindowRequest(order binary.ByteOrder, requestBody []byte) (drawable, parent, x, y, width, height, borderWidth, class, visual, valueMask uint32, values *WindowAttributes) {
	drawable = order.Uint32(requestBody[0:4])
	parent = order.Uint32(requestBody[4:8])
	x = uint32(order.Uint16(requestBody[8:10]))
	y = uint32(order.Uint16(requestBody[10:12]))
	width = uint32(order.Uint16(requestBody[12:14]))
	height = uint32(order.Uint16(requestBody[14:16]))
	borderWidth = uint32(order.Uint16(requestBody[16:18]))
	class = uint32(order.Uint16(requestBody[18:20]))
	visual = order.Uint32(requestBody[20:24])
	valueMask = order.Uint32(requestBody[24:28])

	values, _ = parseWindowAttributes(order, valueMask, requestBody[28:])

	return
}

func parsePutImageRequest(order binary.ByteOrder, requestBody []byte) (drawable, gc uint32, width, height uint16, dstX, dstY int16, leftPad, depth uint8, imgData []byte) {
	log.Printf("parsePutImageRequest: requestBody length=%d, bytes[0:20]=%x", len(requestBody), requestBody[:min(len(requestBody), 20)])

	drawable = order.Uint32(requestBody[0:4])
	gc = order.Uint32(requestBody[4:8])
	width = order.Uint16(requestBody[8:10])
	height = order.Uint16(requestBody[10:12])
	dstX = int16(order.Uint16(requestBody[12:14]))
	dstY = int16(order.Uint16(requestBody[14:16]))
	leftPad = requestBody[16] // CARD8
	depth = requestBody[17]   // CARD8
	// requestBody[18:19] are unused (2 bytes padding)

	imgData = requestBody[20:]
	return
}

func parsePolyLineRequest(order binary.ByteOrder, requestBody []byte) (drawable, gc uint32, points []uint32) {
	drawable = order.Uint32(requestBody[0:4])
	gc = order.Uint32(requestBody[4:8])
	numPoints := (len(requestBody) - 8) / 4
	for i := 0; i < numPoints; i++ {
		offset := 8 + i*4
		x := int32(order.Uint16(requestBody[offset : offset+2]))
		y := int32(order.Uint16(requestBody[offset+2 : offset+4]))
		points = append(points, uint32(x), uint32(y))
	}
	return
}

func parseImageText8Request(order binary.ByteOrder, requestBody []byte) (drawable, gc uint32, x, y int32, text []byte) {
	drawable = order.Uint32(requestBody[0:4])
	gc = order.Uint32(requestBody[4:8])
	x = int32(order.Uint16(requestBody[8:10]))
	y = int32(order.Uint16(requestBody[10:12]))
	text = requestBody[12:]
	return
}

func parseImageText16Request(order binary.ByteOrder, requestBody []byte) (drawable, gc uint32, x, y int32, text []uint16) {
	drawable = order.Uint32(requestBody[0:4])
	gc = order.Uint32(requestBody[4:8])
	x = int32(order.Uint16(requestBody[8:10]))
	y = int32(order.Uint16(requestBody[10:12]))
	for i := 12; i < len(requestBody); i += 2 {
		text = append(text, order.Uint16(requestBody[i:i+2]))
	}
	return
}

type PolyText8Item struct {
	Delta int8
	Str   []byte
}

func parsePolyText8Request(order binary.ByteOrder, requestBody []byte) (drawable, gc uint32, x, y int32, items []PolyText8Item) {
	drawable = order.Uint32(requestBody[0:4])
	gc = order.Uint32(requestBody[4:8])
	x = int32(order.Uint16(requestBody[8:10]))
	y = int32(order.Uint16(requestBody[10:12]))

	currentPos := 12
	for currentPos < len(requestBody) {
		// Each item starts with a length byte (n)
		n := int(requestBody[currentPos])
		currentPos++

		if n == 255 { // Font change
			// This is a font change, not a text item. Skip for now.
			// In a full implementation, you'd parse the font ID here.
			currentPos += 4 // Skip font ID
		} else if n > 0 {
			// Text item
			delta := int8(requestBody[currentPos])
			currentPos++
			str := requestBody[currentPos : currentPos+n]
			currentPos += n
			items = append(items, PolyText8Item{Delta: delta, Str: str})
		}
		// Pad to multiple of 4 bytes
		padding := (4 - (n+2)%4) % 4 // n (string length) + 1 (n) + 1 (delta)
		currentPos += padding
	}
	return
}

type PolyText16Item struct {
	Delta int8
	Str   []uint16
}

func parsePolyText16Request(order binary.ByteOrder, requestBody []byte) (drawable, gc uint32, x, y int32, items []PolyText16Item) {
	drawable = order.Uint32(requestBody[0:4])
	gc = order.Uint32(requestBody[4:8])
	x = int32(order.Uint16(requestBody[8:10]))
	y = int32(order.Uint16(requestBody[10:12]))

	currentPos := 12
	for currentPos < len(requestBody) {
		// Each item starts with a length byte (n)
		n := int(requestBody[currentPos])
		currentPos++

		if n == 255 { // Font change
			// This is a font change, not a text item. Skip for now.
			// In a full implementation, you'd parse the font ID here.
			currentPos += 4 // Skip font ID
		} else if n > 0 {
			// Text item
			delta := int8(requestBody[currentPos])
			currentPos++
			var str []uint16
			for i := 0; i < n; i++ {
				str = append(str, order.Uint16(requestBody[currentPos:currentPos+2]))
				currentPos += 2
			}
			items = append(items, PolyText16Item{Delta: delta, Str: str})
		}
		// Pad to multiple of 4 bytes
		padding := (4 - (n*2+2)%4) % 4 // n*2 (string length) + 1 (n) + 1 (delta)
		currentPos += padding
	}
	return
}

func parsePolyFillRectangleRequest(order binary.ByteOrder, requestBody []byte) (drawable, gc uint32, rects []uint32) {
	drawable = order.Uint32(requestBody[0:4])
	gc = order.Uint32(requestBody[4:8])
	numRects := (len(requestBody) - 8) / 8
	for i := 0; i < numRects; i++ {
		offset := 8 + i*8
		x := uint32(order.Uint16(requestBody[offset : offset+2]))
		y := uint32(order.Uint16(requestBody[offset+2 : offset+4]))
		width := uint32(order.Uint16(requestBody[offset+4 : offset+6]))
		height := uint32(order.Uint16(requestBody[offset+6 : offset+8]))
		rects = append(rects, x, y, width, height)
	}
	return
}

func parseFillPolyRequest(order binary.ByteOrder, requestBody []byte) (drawable, gc uint32, points []uint32) {
	drawable = order.Uint32(requestBody[0:4])
	gc = order.Uint32(requestBody[4:8])
	// 4 bytes for shape, 4 for pad
	numPoints := (len(requestBody) - 12) / 4
	for i := 0; i < numPoints; i++ {
		offset := 12 + i*4
		x := int32(order.Uint16(requestBody[offset : offset+2]))
		y := int32(order.Uint16(requestBody[offset+2 : offset+4]))
		points = append(points, uint32(x), uint32(y))
	}
	return
}

func parsePolySegmentRequest(order binary.ByteOrder, requestBody []byte) (drawable, gc uint32, segments []uint32) {
	drawable = order.Uint32(requestBody[0:4])
	gc = order.Uint32(requestBody[4:8])
	numSegments := (len(requestBody) - 8) / 8
	for i := 0; i < numSegments; i++ {
		offset := 8 + i*8
		x1 := int32(order.Uint16(requestBody[offset : offset+2]))
		y1 := int32(order.Uint16(requestBody[offset+2 : offset+4]))
		x2 := int32(order.Uint16(requestBody[offset+4 : offset+6]))
		y2 := int32(order.Uint16(requestBody[offset+6 : offset+8]))
		segments = append(segments, uint32(x1), uint32(y1), uint32(x2), uint32(y2))
	}
	return
}

func parsePolyPointRequest(order binary.ByteOrder, requestBody []byte) (drawable, gc uint32, points []uint32) {
	drawable = order.Uint32(requestBody[0:4])
	gc = order.Uint32(requestBody[4:8])
	numPoints := (len(requestBody) - 8) / 4
	for i := 0; i < numPoints; i++ {
		offset := 8 + i*4
		x := int32(order.Uint16(requestBody[offset : offset+2]))
		y := int32(order.Uint16(requestBody[offset+2 : offset+4]))
		points = append(points, uint32(x), uint32(y))
	}
	return
}

func parsePolyRectangleRequest(order binary.ByteOrder, requestBody []byte) (drawable, gc uint32, rects []uint32) {
	drawable = order.Uint32(requestBody[0:4])
	gc = order.Uint32(requestBody[4:8])
	numRects := (len(requestBody) - 8) / 8
	for i := 0; i < numRects; i++ {
		offset := 8 + i*8
		x := uint32(order.Uint16(requestBody[offset : offset+2]))
		y := uint32(order.Uint16(requestBody[offset+2 : offset+4]))
		width := uint32(order.Uint16(requestBody[offset+4 : offset+6]))
		height := uint32(order.Uint16(requestBody[offset+6 : offset+8]))
		rects = append(rects, x, y, width, height)
	}
	return
}

func parsePolyArcRequest(order binary.ByteOrder, requestBody []byte) (drawable, gc uint32, arcs []uint32) {
	drawable = order.Uint32(requestBody[0:4])
	gc = order.Uint32(requestBody[4:8])
	numArcs := (len(requestBody) - 8) / 12
	for i := 0; i < numArcs; i++ {
		offset := 8 + i*12
		x := int32(order.Uint16(requestBody[offset : offset+2]))
		y := int32(order.Uint16(requestBody[offset+2 : offset+4]))
		width := uint32(order.Uint16(requestBody[offset+4 : offset+6]))
		height := uint32(order.Uint16(requestBody[offset+6 : offset+8]))
		angle1 := int32(order.Uint16(requestBody[offset+8 : offset+10]))
		angle2 := int32(order.Uint16(requestBody[offset+10 : offset+12]))
		arcs = append(arcs, uint32(x), uint32(y), width, height, uint32(angle1), uint32(angle2))
	}
	return
}

func parsePolyFillArcRequest(order binary.ByteOrder, requestBody []byte) (drawable, gc uint32, arcs []uint32) {
	drawable = order.Uint32(requestBody[0:4])
	gc = order.Uint32(requestBody[4:8])
	numArcs := (len(requestBody) - 8) / 12
	for i := 0; i < numArcs; i++ {
		offset := 8 + i*12
		x := int32(order.Uint16(requestBody[offset : offset+2]))
		y := int32(order.Uint16(requestBody[offset+2 : offset+4]))
		width := uint32(order.Uint16(requestBody[offset+4 : offset+6]))
		height := uint32(order.Uint16(requestBody[offset+6 : offset+8]))
		angle1 := int32(order.Uint16(requestBody[offset+8 : offset+10]))
		angle2 := int32(order.Uint16(requestBody[offset+10 : offset+12]))
		arcs = append(arcs, uint32(x), uint32(y), width, height, uint32(angle1), uint32(angle2))
	}
	return
}

func parseGetWindowAttributesRequest(order binary.ByteOrder, requestBody []byte) (drawable uint32) {
	drawable = order.Uint32(requestBody[0:4])
	return
}

func parseConfigureWindowRequest(order binary.ByteOrder, requestBody []byte) (drawable uint32, valueMask uint16, values []uint32) {
	drawable = order.Uint32(requestBody[0:4])
	valueMask = order.Uint16(requestBody[4:6])
	// The rest of the body is a list of values, corresponding to the bits in valueMask.
	for i := 8; i < len(requestBody); i += 4 {
		values = append(values, order.Uint32(requestBody[i:i+4]))
	}
	return
}

func parseGetGeometryRequest(order binary.ByteOrder, requestBody []byte) (drawable uint32) {
	drawable = order.Uint32(requestBody[0:4])
	return
}

func parseSendEventRequest(order binary.ByteOrder, requestBody []byte) (destination uint32, eventMask uint32, eventData []byte) {
	destination = order.Uint32(requestBody[4:8])
	eventMask = order.Uint32(requestBody[8:12])
	eventData = requestBody[12:44]
	return
}

func parseClearAreaRequest(order binary.ByteOrder, requestBody []byte) (drawable uint32, x int16, y int16, width uint16, height uint16) {
	drawable = order.Uint32(requestBody[0:4])
	x = int16(order.Uint16(requestBody[4:6]))
	y = int16(order.Uint16(requestBody[6:8]))
	width = order.Uint16(requestBody[8:10])
	height = order.Uint16(requestBody[10:12])
	return
}

func parseQueryPointerRequest(order binary.ByteOrder, requestBody []byte) (drawable uint32) {
	drawable = order.Uint32(requestBody[0:4])
	return
}

func parseCopyAreaRequest(order binary.ByteOrder, requestBody []byte) (srcDrawable, dstDrawable, gc uint32, srcX, srcY, dstX, dstY int16, width, height uint16) {
	srcDrawable = order.Uint32(requestBody[0:4])
	dstDrawable = order.Uint32(requestBody[4:8])
	gc = order.Uint32(requestBody[8:12])
	srcX = int16(order.Uint16(requestBody[12:14]))
	srcY = int16(order.Uint16(requestBody[14:16]))
	dstX = int16(order.Uint16(requestBody[16:18]))
	dstY = int16(order.Uint16(requestBody[18:20]))
	width = order.Uint16(requestBody[20:22])
	height = order.Uint16(requestBody[22:24])
	return
}

func parseGetImageRequest(order binary.ByteOrder, requestBody []byte) (drawable uint32, x int16, y int16, width, height uint16, planeMask uint32) {
	drawable = order.Uint32(requestBody[0:4])
	x = int16(order.Uint16(requestBody[4:6]))
	y = int16(order.Uint16(requestBody[6:8]))
	width = order.Uint16(requestBody[8:10])
	height = order.Uint16(requestBody[10:12])
	planeMask = order.Uint32(requestBody[12:16])
	return
}

func parseGetAtomNameRequest(order binary.ByteOrder, requestBody []byte) (atom uint32) {
	atom = order.Uint32(requestBody[0:4])
	return
}

func parseListPropertiesRequest(order binary.ByteOrder, requestBody []byte) (window uint32) {
	window = order.Uint32(requestBody[0:4])
	return
}

func parseGCValues(order binary.ByteOrder, valueMask uint32, body []byte) (*GC, int) {
	gc := &GC{
		Function:   GXcopy,
		Foreground: 0,
		Background: 1,
	}
	read := 0
	if valueMask&GCFunction != 0 {
		gc.Function = order.Uint32(body[read : read+4])
		read += 4
	}
	if valueMask&GCPlaneMask != 0 {
		gc.PlaneMask = order.Uint32(body[read : read+4])
		read += 4
	}
	if valueMask&GCForeground != 0 {
		gc.Foreground = order.Uint32(body[read : read+4])
		read += 4
	}
	if valueMask&GCBackground != 0 {
		gc.Background = order.Uint32(body[read : read+4])
		read += 4
	}
	if valueMask&GCLineWidth != 0 {
		gc.LineWidth = order.Uint32(body[read : read+4])
		read += 4
	}
	if valueMask&GCLineStyle != 0 {
		gc.LineStyle = order.Uint32(body[read : read+4])
		read += 4
	}
	if valueMask&GCCapStyle != 0 {
		gc.CapStyle = order.Uint32(body[read : read+4])
		read += 4
	}
	if valueMask&GCJoinStyle != 0 {
		gc.JoinStyle = order.Uint32(body[read : read+4])
		read += 4
	}
	if valueMask&GCFillStyle != 0 {
		gc.FillStyle = order.Uint32(body[read : read+4])
		read += 4
	}
	if valueMask&GCFillRule != 0 {
		gc.FillRule = order.Uint32(body[read : read+4])
		read += 4
	}
	if valueMask&GCTile != 0 {
		gc.Tile = order.Uint32(body[read : read+4])
		read += 4
	}
	if valueMask&GCStipple != 0 {
		gc.Stipple = order.Uint32(body[read : read+4])
		read += 4
	}
	if valueMask&GCTileStipXOrigin != 0 {
		gc.TileStipXOrigin = order.Uint32(body[read : read+4])
		read += 4
	}
	if valueMask&GCTileStipYOrigin != 0 {
		gc.TileStipYOrigin = order.Uint32(body[read : read+4])
		read += 4
	}
	if valueMask&GCFont != 0 {
		gc.Font = order.Uint32(body[read : read+4])
		read += 4
	}
	if valueMask&GCSubwindowMode != 0 {
		gc.SubwindowMode = order.Uint32(body[read : read+4])
		read += 4
	}
	if valueMask&GCGraphicsExposures != 0 {
		gc.GraphicsExposures = order.Uint32(body[read : read+4])
		read += 4
	}
	if valueMask&GCClipXOrigin != 0 {
		gc.ClipXOrigin = int32(order.Uint32(body[read : read+4]))
		read += 4
	}
	if valueMask&GCClipYOrigin != 0 {
		gc.ClipYOrigin = int32(order.Uint32(body[read : read+4]))
		read += 4
	}
	if valueMask&GCClipMask != 0 {
		gc.ClipMask = order.Uint32(body[read : read+4])
		read += 4
	}
	if valueMask&GCDashOffset != 0 {
		gc.DashOffset = order.Uint32(body[read : read+4])
		read += 4
	}
	if valueMask&GCDashList != 0 {
		gc.Dashes = order.Uint32(body[read : read+4])
		read += 4
	}
	if valueMask&GCArcMode != 0 {
		gc.ArcMode = order.Uint32(body[read : read+4])
		read += 4
	}
	return gc, read
}

func parseCreatePixmapRequest(order binary.ByteOrder, payload []byte) (uint32, uint32, uint32, uint32) {
	pid := order.Uint32(payload[0:4])
	drawable := order.Uint32(payload[4:8])
	width := uint32(order.Uint16(payload[8:10]))
	height := uint32(order.Uint16(payload[10:12]))
	return pid, drawable, width, height
}

func parseWarpPointerRequest(order binary.ByteOrder, payload []byte) (int16, int16) {
	dstX := int16(order.Uint16(payload[12:14]))
	dstY := int16(order.Uint16(payload[14:16]))
	return dstX, dstY
}

func parseAllowEventsRequest(order binary.ByteOrder, requestBody []byte) (mode byte, time uint32) {
	mode = requestBody[0]
	time = order.Uint32(requestBody[4:8])
	return
}

func parseOpenFontRequest(order binary.ByteOrder, requestBody []byte) (fid uint32, name string) {
	fid = order.Uint32(requestBody[0:4])
	nameLen := order.Uint16(requestBody[4:6])
	name = string(requestBody[8 : 8+nameLen])
	return
}

func parseCreateColormapRequest(order binary.ByteOrder, payload []byte) (byte, uint32, uint32, uint32) {
	alloc := payload[0]
	mid := order.Uint32(payload[4:8])
	window := order.Uint32(payload[8:12])
	visual := order.Uint32(payload[12:16])
	return alloc, mid, window, visual
}

func parseAllocColorRequest(order binary.ByteOrder, payload []byte) (uint32, uint16, uint16, uint16) {
	cmap := order.Uint32(payload[0:4])
	red := order.Uint16(payload[4:6])
	green := order.Uint16(payload[6:8])
	blue := order.Uint16(payload[8:10])
	return cmap, red, green, blue
}

type AllocNamedColorRequest struct {
	Cmap     xID
	Name     []byte
	Sequence uint16
	MinorOp  byte
	MajorOp  reqCode
}

func parseLookupColorRequest(order binary.ByteOrder, payload []byte) (cmapID uint32, name string) {
	cmapID = order.Uint32(payload[0:4])
	nameLen := order.Uint16(payload[4:6])
	name = string(payload[8 : 8+nameLen])
	return
}

func parseAllocNamedColorRequest(order binary.ByteOrder, payload []byte, sequence uint16, minorOp byte, majorOp reqCode) AllocNamedColorRequest {
	cmapID := order.Uint32(payload[0:4])
	nameLen := order.Uint16(payload[4:6])
	name := payload[8 : 8+nameLen]
	return AllocNamedColorRequest{
		Cmap:     xID{local: cmapID},
		Name:     name,
		Sequence: sequence,
		MinorOp:  minorOp,
		MajorOp:  majorOp,
	}
}
