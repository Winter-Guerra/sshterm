//go:build x11

package x11

import (
	"bytes"
	"encoding/binary"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseImageText8Request(t *testing.T) {
	// ImageText8 request: drawable, gc, x, y, text
	drawable := uint32(1)
	gc := uint32(2)
	x := int16(10)
	y := int16(20)
	text := []byte("Hello")

	payload := make([]byte, 12) // drawable (4) + gc (4) + x (2) + y (2)
	binary.LittleEndian.PutUint32(payload[0:4], drawable)
	binary.LittleEndian.PutUint32(payload[4:8], gc)
	binary.LittleEndian.PutUint16(payload[8:10], uint16(x))
	binary.LittleEndian.PutUint16(payload[10:12], uint16(y))
	payload = append(payload, text...)

	d, g, parsedX, parsedY, parsedText := parseImageText8Request(binary.LittleEndian, payload)

	if d != drawable {
		t.Errorf("Expected drawable %d, got %d", drawable, d)
	}
	if g != gc {
		t.Errorf("Expected gc %d, got %d", gc, g)
	}
	if parsedX != int32(x) {
		t.Errorf("Expected x %d, got %d", x, parsedX)
	}
	if parsedY != int32(y) {
		t.Errorf("Expected y %d, got %d", y, parsedY)
	}
	if !bytes.Equal(parsedText, text) {
		t.Errorf("Expected text %s, got %s", text, parsedText)
	}
}

func TestParseImageText16Request(t *testing.T) {
	// ImageText16 request: drawable, gc, x, y, text
	drawable := uint32(1)
	gc := uint32(2)
	x := int16(10)
	y := int16(20)
	text := []uint16{0x0048, 0x0065, 0x006c, 0x006c, 0x006f} // "Hello"

	payload := make([]byte, 12) // drawable (4) + gc (4) + x (2) + y (2)
	binary.LittleEndian.PutUint32(payload[0:4], drawable)
	binary.LittleEndian.PutUint32(payload[4:8], gc)
	binary.LittleEndian.PutUint16(payload[8:10], uint16(x))
	binary.LittleEndian.PutUint16(payload[10:12], uint16(y))
	for _, r := range text {
		buf := make([]byte, 2)
		binary.LittleEndian.PutUint16(buf, r)
		payload = append(payload, buf...)
	}

	d, g, parsedX, parsedY, parsedText := parseImageText16Request(binary.LittleEndian, payload)

	assert.Equal(t, drawable, d, "drawable mismatch")
	assert.Equal(t, gc, g, "gc mismatch")
	assert.Equal(t, int32(x), parsedX, "x mismatch")
	assert.Equal(t, int32(y), parsedY, "y mismatch")
	assert.Equal(t, text, parsedText, "text mismatch")
}

func TestParsePolyText8Request(t *testing.T) {
	// PolyText8 request: drawable, gc, x, y, items
	drawable := uint32(1)
	gc := uint32(2)
	x := int16(10)
	y := int16(20)

	// Item 1: delta=5, text="Hi"
	// Item 2: delta=10, text="There"

	payload := make([]byte, 12) // drawable (4) + gc (4) + x (2) + y (2)
	binary.LittleEndian.PutUint32(payload[0:4], drawable)
	binary.LittleEndian.PutUint32(payload[4:8], gc)
	binary.LittleEndian.PutUint16(payload[8:10], uint16(x))
	binary.LittleEndian.PutUint16(payload[10:12], uint16(y))

	// Item 1
	payload = append(payload, 2)        // n = 2 (length of "Hi")
	payload = append(payload, 5)        // delta = 5
	payload = append(payload, 'H', 'i') // text = "Hi"
	// Padding for (1 byte n + 1 byte delta + 2 bytes string) = 4 bytes. No padding needed.

	// Item 2
	payload = append(payload, 5)                       // n = 5 (length of "There")
	payload = append(payload, 10)                      // delta = 10
	payload = append(payload, 'T', 'h', 'e', 'r', 'e') // text = "There"
	payload = append(payload, 0)                       // padding for (1 byte n + 1 byte delta + 5 bytes string) = 7 bytes. Need 1 byte padding.

	d, g, parsedX, parsedY, parsedItems := parsePolyText8Request(binary.LittleEndian, payload)

	assert.Equal(t, drawable, d, "drawable mismatch")
	assert.Equal(t, gc, g, "gc mismatch")
	assert.Equal(t, int32(x), parsedX, "x mismatch")
	assert.Equal(t, int32(y), parsedY, "y mismatch")

	expectedItems := []PolyText8Item{
		{Delta: 5, Str: []byte("Hi")},
		{Delta: 10, Str: []byte("There")},
	}
	assert.Equal(t, expectedItems, parsedItems, "items mismatch")
}

func TestParsePolyText16Request(t *testing.T) {
	// PolyText16 request: drawable, gc, x, y, items
	drawable := uint32(1)
	gc := uint32(2)
	x := int16(10)
	y := int16(20)

	// Item 1: delta=5, text="Hi"
	// Item 2: delta=10, text="There"

	payload := make([]byte, 12) // drawable (4) + gc (4) + x (2) + y (2)
	binary.LittleEndian.PutUint32(payload[0:4], drawable)
	binary.LittleEndian.PutUint32(payload[4:8], gc)
	binary.LittleEndian.PutUint16(payload[8:10], uint16(x))
	binary.LittleEndian.PutUint16(payload[10:12], uint16(y))

	// Item 1
	payload = append(payload, 2)                      // n = 2 (length of "Hi" in CHAR2B)
	payload = append(payload, 5)                      // delta = 5
	payload = append(payload, 0x48, 0x00, 0x69, 0x00) // text = "Hi"
	payload = append(payload, 0, 0)                   // padding for (1 byte n + 1 byte delta + 4 bytes string) = 6 bytes. Need 2 bytes padding.

	// Item 2
	payload = append(payload, 5)                                                          // n = 5 (length of "There" in CHAR2B)
	payload = append(payload, 10)                                                         // delta = 10
	payload = append(payload, 0x54, 0x00, 0x68, 0x00, 0x65, 0x00, 0x72, 0x00, 0x65, 0x00) // text = "There"
	// Padding for (1 byte n + 1 byte delta + 10 bytes string) = 12 bytes. No padding needed.

	d, g, parsedX, parsedY, parsedItems := parsePolyText16Request(binary.LittleEndian, payload)

	assert.Equal(t, drawable, d, "drawable mismatch")
	assert.Equal(t, gc, g, "gc mismatch")
	assert.Equal(t, int32(x), parsedX, "x mismatch")
	assert.Equal(t, int32(y), parsedY, "y mismatch")

	expectedItems := []PolyText16Item{
		{Delta: 5, Str: []uint16{0x0048, 0x0069}},
		{Delta: 10, Str: []uint16{0x0054, 0x0068, 0x0065, 0x0072, 0x0065}},
	}
	assert.Equal(t, expectedItems, parsedItems, "items mismatch")
}

func TestGetWindowAttributes(t *testing.T) {
	order := binary.LittleEndian
	s := &x11Server{
		byteOrder: order,
		windows:   make(map[xID]*window),
		visualID:  0x21,
		logger:    &testLogger{t: t},
	}

	wid := xID{0, 1}
	w := &window{
		xid:        wid,
		parent:     0,
		x:          10,
		y:          20,
		width:      100,
		height:     200,
		mapped:     true,
		attributes: &WindowAttributes{BackingStore: 1},
	}
	s.windows[wid] = w

	reqBody := make([]byte, 4)
	order.PutUint32(reqBody, wid.local)
	req := &request{
		opcode:   3, // GetWindowAttributes
		sequence: 10,
		body:     reqBody,
	}

	mockClient := &x11Client{conn: &testConn{}}
	reply := s.handleRequest(mockClient, req)

	replyMsg := reply.encodeMessage(order)

	assert.NotNil(t, replyMsg, "Reply should not be nil")
	assert.Equal(t, 44, len(replyMsg), "Reply length should be 44 bytes")

	assert.Equal(t, byte(1), replyMsg[0], "Reply type should be 1")
	assert.Equal(t, byte(1), replyMsg[1], "Backing store should be 1")
	assert.Equal(t, uint16(10), order.Uint16(replyMsg[2:4]), "Sequence number should match request")
	assert.Equal(t, uint32(3), order.Uint32(replyMsg[4:8]), "Reply length field should be 3")
	assert.Equal(t, s.visualID, order.Uint32(replyMsg[8:12]), "Visual ID should match server")
	assert.Equal(t, uint16(1), order.Uint16(replyMsg[12:14]), "Class should be InputOutput")
	assert.Equal(t, byte(2), replyMsg[26], "Map state should be Viewable")
}

func TestConfigureWindow(t *testing.T) {
	order := binary.LittleEndian
	mockFrontend := &MockX11Frontend{}
	s := &x11Server{
		byteOrder: order,
		windows:   make(map[xID]*window),
		logger:    &testLogger{t: t},
		frontend:  mockFrontend,
	}

	wid := xID{0, 1}
	w := &window{
		xid:    wid,
		x:      10,
		y:      20,
		width:  100,
		height: 200,
	}
	s.windows[wid] = w

	// Request to change x, y, width, height
	// valueMask: 1 (x) | 2 (y) | 4 (width) | 8 (height) = 15
	values := []uint32{50, 60, 300, 400}
	reqBody := make([]byte, 8+4*4)
	order.PutUint32(reqBody[0:4], wid.local)
	order.PutUint16(reqBody[4:6], 15)
	order.PutUint32(reqBody[8:12], values[0])  // new x
	order.PutUint32(reqBody[12:16], values[1]) // new y
	order.PutUint32(reqBody[16:20], values[2]) // new width
	order.PutUint32(reqBody[20:24], values[3]) // new height

	req := &request{
		opcode: 12, // ConfigureWindow
		body:   reqBody,
	}

	mockClient := &x11Client{conn: &testConn{}}
	reply := s.handleRequest(mockClient, req)

	assert.Nil(t, reply, "ConfigureWindow should not send a reply")
	assert.Len(t, mockFrontend.ConfigureWindowCalls, 1, "ConfigureWindow should be called once")
	call := mockFrontend.ConfigureWindowCalls[0]
	assert.Equal(t, wid, call.id, "Window ID should match")
	assert.Equal(t, uint16(15), call.valueMask, "Value mask should match")
	assert.Equal(t, values, call.values, "Values should match")
}

func TestGetGeometry(t *testing.T) {
	order := binary.LittleEndian
	s := &x11Server{
		byteOrder: order,
		windows:   make(map[xID]*window),
		logger:    &testLogger{t: t},
	}

	wid := xID{0, 1}
	w := &window{
		xid:    wid,
		x:      10,
		y:      20,
		width:  100,
		height: 200,
		depth:  24,
	}
	s.windows[wid] = w

	reqBody := make([]byte, 4)
	order.PutUint32(reqBody, wid.local)
	req := &request{
		opcode:   14, // GetGeometry
		sequence: 11,
		body:     reqBody,
	}

	mockClient := &x11Client{conn: &testConn{}}
	reply := s.handleRequest(mockClient, req)

	replyMsg := reply.encodeMessage(order)

	assert.NotNil(t, replyMsg, "Reply should not be nil")
	assert.Equal(t, 32, len(replyMsg), "Reply length should be 32 bytes")

	assert.Equal(t, byte(1), replyMsg[0], "Reply type should be 1")
	assert.Equal(t, byte(24), replyMsg[1], "Depth should match window")
	assert.Equal(t, uint16(11), order.Uint16(replyMsg[2:4]), "Sequence number should match request")
	assert.Equal(t, uint32(0), order.Uint32(replyMsg[4:8]), "Reply length field should be 0")
	assert.Equal(t, s.rootWindowID(), order.Uint32(replyMsg[8:12]), "Root window ID should match")
	assert.Equal(t, uint16(10), order.Uint16(replyMsg[12:14]), "X should match window")
	assert.Equal(t, uint16(20), order.Uint16(replyMsg[14:16]), "Y should match window")
	assert.Equal(t, uint16(100), order.Uint16(replyMsg[16:18]), "Width should match window")
	assert.Equal(t, uint16(200), order.Uint16(replyMsg[18:20]), "Height should match window")
	assert.Equal(t, uint16(0), order.Uint16(replyMsg[20:22]), "Border width should be 0")
}

func TestQueryPointer(t *testing.T) {
	order := binary.LittleEndian
	s := &x11Server{
		byteOrder: order,
		windows:   make(map[xID]*window),
		logger:    &testLogger{t: t},
	}

	s.pointerX = 123
	s.pointerY = 456

	reqBody := make([]byte, 4) // Drawable ID, not used by handler
	order.PutUint32(reqBody, 1)
	req := &request{
		opcode:   38, // QueryPointer
		sequence: 12,
		body:     reqBody,
	}

	mockClient := &x11Client{conn: &testConn{}}
	reply := s.handleRequest(mockClient, req)

	replyMsg := reply.encodeMessage(order)

	assert.NotNil(t, replyMsg, "Reply should not be nil")
	assert.Equal(t, 32, len(replyMsg), "Reply length should be 32 bytes")

	assert.Equal(t, byte(1), replyMsg[0], "Reply type should be 1")
	assert.Equal(t, byte(1), replyMsg[1], "Same screen should be true")
	assert.Equal(t, uint16(12), order.Uint16(replyMsg[2:4]), "Sequence number should match request")
	assert.Equal(t, s.rootWindowID(), order.Uint32(replyMsg[8:12]), "Root window ID should match")
	assert.Equal(t, uint16(123), order.Uint16(replyMsg[16:18]), "Root X should match pointer")
	assert.Equal(t, uint16(456), order.Uint16(replyMsg[18:20]), "Root Y should match pointer")
}

func TestSendEvent(t *testing.T) {
	order := binary.LittleEndian
	s := &x11Server{
		byteOrder: order,
		logger:    &testLogger{t: t},
		frontend:  &MockX11Frontend{},
	}

	// The body contains the destination window, event mask, and 32-byte event data.
	reqBody := make([]byte, 44)

	req := &request{
		opcode: 25, // SendEvent
		body:   reqBody,
	}

	mockClient := &x11Client{conn: &testConn{}}
	reply := s.handleRequest(mockClient, req)
	assert.Nil(t, reply, "SendEvent should not send a reply")
}

func TestClearArea(t *testing.T) {
	order := binary.LittleEndian
	mockFrontend := &MockX11Frontend{}
	s := &x11Server{
		byteOrder: order,
		logger:    &testLogger{t: t},
		frontend:  mockFrontend,
	}

	reqBody := make([]byte, 12)
	order.PutUint32(reqBody[0:4], 1)     // drawable
	order.PutUint16(reqBody[4:6], 10)    // x
	order.PutUint16(reqBody[6:8], 20)    // y
	order.PutUint16(reqBody[8:10], 100)  // width
	order.PutUint16(reqBody[10:12], 200) // height

	req := &request{
		opcode: 61, // ClearArea
		body:   reqBody,
	}

	mockClient := &x11Client{conn: &testConn{}}
	reply := s.handleRequest(mockClient, req)

	assert.Nil(t, reply, "ClearArea should not send a reply")
	assert.Len(t, mockFrontend.ClearAreaCalls, 1, "ClearArea should be called once")
	call := mockFrontend.ClearAreaCalls[0]
	assert.Equal(t, xID{0, 1}, call.drawable, "Drawable ID should match")
	assert.Equal(t, uint32(10), call.x, "X should match")
	assert.Equal(t, uint32(20), call.y, "Y should match")
	assert.Equal(t, uint32(100), call.width, "Width should match")
	assert.Equal(t, uint32(200), call.height, "Height should match")
}

func TestParseQueryPointerRequest(t *testing.T) {
	order := binary.LittleEndian
	reqBody := make([]byte, 4)
	order.PutUint32(reqBody, 123)
	drawable := parseQueryPointerRequest(order, reqBody)
	assert.Equal(t, uint32(123), drawable, "Drawable ID should be parsed correctly")

}

func TestParseCopyAreaRequest(t *testing.T) {
	order := binary.LittleEndian
	reqBody := make([]byte, 24)
	order.PutUint32(reqBody[0:4], 1)     // srcDrawable
	order.PutUint32(reqBody[4:8], 2)     // dstDrawable
	order.PutUint32(reqBody[8:12], 3)    // gc
	order.PutUint16(reqBody[12:14], 10)  // srcX
	order.PutUint16(reqBody[14:16], 20)  // srcY
	order.PutUint16(reqBody[16:18], 30)  // dstX
	order.PutUint16(reqBody[18:20], 40)  // dstY
	order.PutUint16(reqBody[20:22], 100) // width
	order.PutUint16(reqBody[22:24], 200) // height

	srcDrawable, dstDrawable, gc, srcX, srcY, dstX, dstY, width, height := parseCopyAreaRequest(order, reqBody)

	assert.Equal(t, uint32(1), srcDrawable, "srcDrawable should be parsed correctly")
	assert.Equal(t, uint32(2), dstDrawable, "dstDrawable should be parsed correctly")
	assert.Equal(t, uint32(3), gc, "gc should be parsed correctly")
	assert.Equal(t, int16(10), srcX, "srcX should be parsed correctly")
	assert.Equal(t, int16(20), srcY, "srcY should be parsed correctly")
	assert.Equal(t, int16(30), dstX, "dstX should be parsed correctly")
	assert.Equal(t, int16(40), dstY, "dstY should be parsed correctly")
	assert.Equal(t, uint16(100), width, "width should be parsed correctly")
	assert.Equal(t, uint16(200), height, "height should be parsed correctly")
}

func TestCopyArea(t *testing.T) {
	order := binary.LittleEndian
	mockFrontend := &MockX11Frontend{}
	s := &x11Server{
		byteOrder: order,
		logger:    &testLogger{t: t},
		frontend:  mockFrontend,
		gcs:       make(map[xID]*GC),
	}

	gcID := xID{0, 3}
	s.gcs[gcID] = &GC{Foreground: 3}

	reqBody := make([]byte, 24)
	order.PutUint32(reqBody[0:4], 1)     // srcDrawable
	order.PutUint32(reqBody[4:8], 2)     // dstDrawable
	order.PutUint32(reqBody[8:12], 3)    // gc
	order.PutUint16(reqBody[12:14], 10)  // srcX
	order.PutUint16(reqBody[14:16], 20)  // srcY
	order.PutUint16(reqBody[16:18], 30)  // dstX
	order.PutUint16(reqBody[18:20], 40)  // dstY
	order.PutUint16(reqBody[20:22], 100) // width
	order.PutUint16(reqBody[22:24], 200) // height

	req := &request{
		opcode: 62, // CopyArea
		body:   reqBody,
	}

	mockClient := &x11Client{conn: &testConn{}}
	reply := s.handleRequest(mockClient, req)

	assert.Nil(t, reply, "CopyArea should not send a reply")
	assert.Len(t, mockFrontend.CopyAreaCalls, 1, "CopyArea should be called once")
	call := mockFrontend.CopyAreaCalls[0]
	assert.Equal(t, xID{0, 1}, call.srcDrawable, "srcDrawable should match")
	assert.Equal(t, xID{0, 2}, call.dstDrawable, "dstDrawable should match")
	assert.Equal(t, uint32(3), call.gc, "gc should match")
	assert.Equal(t, uint32(10), call.srcX, "srcX should match")
	assert.Equal(t, uint32(20), call.srcY, "srcY should match")
	assert.Equal(t, uint32(30), call.dstX, "dstX should match")
	assert.Equal(t, uint32(40), call.dstY, "dstY should match")
	assert.Equal(t, uint32(100), call.width, "width should match")
	assert.Equal(t, uint32(200), call.height, "height should match")
}

func TestParseGetImageRequest(t *testing.T) {
	order := binary.LittleEndian
	reqBody := make([]byte, 16)
	order.PutUint32(reqBody[0:4], 1)            // drawable
	order.PutUint16(reqBody[4:6], 10)           // x
	order.PutUint16(reqBody[6:8], 20)           // y
	order.PutUint16(reqBody[8:10], 100)         // width
	order.PutUint16(reqBody[10:12], 200)        // height
	order.PutUint32(reqBody[12:16], 0xFFFFFFFF) // planeMask

	drawable, x, y, width, height, planeMask := parseGetImageRequest(order, reqBody)

	assert.Equal(t, uint32(1), drawable, "drawable should be parsed correctly")
	assert.Equal(t, int16(10), x, "x should be parsed correctly")
	assert.Equal(t, int16(20), y, "y should be parsed correctly")
	assert.Equal(t, uint16(100), width, "width should be parsed correctly")
	assert.Equal(t, uint16(200), height, "height should be parsed correctly")
	assert.Equal(t, uint32(0xFFFFFFFF), planeMask, "planeMask should be parsed correctly")
}

func TestGetImage(t *testing.T) {
	order := binary.LittleEndian
	mockFrontend := &MockX11Frontend{}
	s := &x11Server{
		byteOrder: order,
		logger:    &testLogger{t: t},
		frontend:  mockFrontend,
		visualID:  0x21,
	}

	reqBody := make([]byte, 16)
	order.PutUint32(reqBody[0:4], 1)            // drawable
	order.PutUint16(reqBody[4:6], 10)           // x
	order.PutUint16(reqBody[6:8], 20)           // y
	order.PutUint16(reqBody[8:10], 100)         // width
	order.PutUint16(reqBody[10:12], 200)        // height
	order.PutUint32(reqBody[12:16], 0xFFFFFFFF) // planeMask

	req := &request{
		opcode:   73, // GetImage
		data:     2,  // format ZPixmap
		sequence: 13,
		body:     reqBody,
	}

	imgData := []byte{1, 2, 3, 4}
	mockFrontend.GetImageReturn = imgData

	mockClient := &x11Client{conn: &testConn{}}
	reply := s.handleRequest(mockClient, req)

	replyMsg := reply.encodeMessage(order)

	assert.NotNil(t, replyMsg, "Reply should not be nil")
	assert.Len(t, mockFrontend.GetImageCalls, 1, "GetImage should be called once")
	call := mockFrontend.GetImageCalls[0]
	assert.Equal(t, xID{0, 1}, call.drawable, "drawable should match")
	assert.Equal(t, uint32(10), call.x, "x should match")
	assert.Equal(t, uint32(20), call.y, "y should match")
	assert.Equal(t, uint32(100), call.width, "width should match")
	assert.Equal(t, uint32(200), call.height, "height should match")
	assert.Equal(t, uint32(2), call.format, "format should match")

	assert.Equal(t, byte(1), replyMsg[0], "Reply type should be 1")
	assert.Equal(t, byte(24), replyMsg[1], "Depth should be 24")
	assert.Equal(t, uint16(13), order.Uint16(replyMsg[2:4]), "Sequence number should match request")
	assert.Equal(t, uint32(len(imgData)/4), order.Uint32(replyMsg[4:8]), "Reply length should be correct")
	assert.Equal(t, s.visualID, order.Uint32(replyMsg[8:12]), "Visual ID should match server")
	assert.Equal(t, imgData, replyMsg[32:], "Image data should match")
}

func TestParseGetAtomNameRequest(t *testing.T) {
	order := binary.LittleEndian
	reqBody := make([]byte, 4)
	order.PutUint32(reqBody[0:4], 123) // atom

	atom := parseGetAtomNameRequest(order, reqBody)

	assert.Equal(t, uint32(123), atom, "atom should be parsed correctly")
}

func TestGetAtomName(t *testing.T) {
	order := binary.LittleEndian
	mockFrontend := &MockX11Frontend{}
	s := &x11Server{
		byteOrder: order,
		logger:    &testLogger{t: t},
		frontend:  mockFrontend,
	}

	reqBody := make([]byte, 4)
	order.PutUint32(reqBody[0:4], 123) // atom

	req := &request{
		opcode:   17, // GetAtomName
		sequence: 14,
		body:     reqBody,
	}

	atomName := "TEST_ATOM"
	mockFrontend.GetAtomNameReturn = atomName

	mockClient := &x11Client{conn: &testConn{}}
	reply := s.handleRequest(mockClient, req)

	replyMsg := reply.encodeMessage(order)

	assert.NotNil(t, replyMsg, "Reply should not be nil")
	assert.Len(t, mockFrontend.GetAtomNameCalls, 1, "GetAtomName should be called once")
	assert.Equal(t, uint32(123), mockFrontend.GetAtomNameCalls[0], "atom should match")

	assert.Equal(t, byte(1), replyMsg[0], "Reply type should be 1")
	assert.Equal(t, uint16(14), order.Uint16(replyMsg[2:4]), "Sequence number should match request")
	assert.Equal(t, uint32((len(atomName)+3)/4), order.Uint32(replyMsg[4:8]), "Reply length should be correct")
	assert.Equal(t, uint16(len(atomName)), order.Uint16(replyMsg[8:10]), "Name length should be correct")
	assert.Equal(t, atomName, string(replyMsg[32:32+len(atomName)]), "Atom name should match")
}

func TestParseListPropertiesRequest(t *testing.T) {
	order := binary.LittleEndian
	reqBody := make([]byte, 4)
	order.PutUint32(reqBody[0:4], 123) // window

	window := parseListPropertiesRequest(order, reqBody)

	assert.Equal(t, uint32(123), window, "window should be parsed correctly")
}

func TestListProperties(t *testing.T) {
	order := binary.LittleEndian
	mockFrontend := &MockX11Frontend{}
	s := &x11Server{
		byteOrder: order,
		logger:    &testLogger{t: t},
		frontend:  mockFrontend,
	}

	reqBody := make([]byte, 4)
	order.PutUint32(reqBody[0:4], 123) // window

	req := &request{
		opcode:   21, // ListProperties
		sequence: 15,
		body:     reqBody,
	}

	atoms := []uint32{1, 2, 3}
	mockFrontend.ListPropertiesReturn = atoms

	mockClient := &x11Client{conn: &testConn{}}
	reply := s.handleRequest(mockClient, req)

	replyMsg := reply.encodeMessage(order)

	assert.NotNil(t, replyMsg, "Reply should not be nil")
	assert.Len(t, mockFrontend.ListPropertiesCalls, 1, "ListProperties should be called once")
	assert.Equal(t, xID{0, 123}, mockFrontend.ListPropertiesCalls[0].window, "window should match")

	assert.Equal(t, byte(1), replyMsg[0], "Reply type should be 1")
	assert.Equal(t, uint16(15), order.Uint16(replyMsg[2:4]), "Sequence number should match request")
	assert.Equal(t, uint32(len(atoms)), order.Uint32(replyMsg[4:8]), "Reply length should be correct")
	assert.Equal(t, uint16(len(atoms)), order.Uint16(replyMsg[8:10]), "Atoms count should be correct")

	for i, atom := range atoms {
		assert.Equal(t, atom, order.Uint32(replyMsg[32+i*4:]), "Atom should match")
	}
}

func TestColormap(t *testing.T) {
	order := binary.LittleEndian
	s := &x11Server{
		byteOrder: order,
		windows:   make(map[xID]*window),
		gcs:       make(map[xID]*GC),
		colormaps: map[xID]*colormap{
			{local: 1}: {pixels: make(map[uint32]color)},
		},
		defaultColormap: 1,
		logger:          &testLogger{t: t},
		frontend:        &MockX11Frontend{},
	}
	mockClient := &x11Client{conn: &testConn{}, byteOrder: order}

	// Test CreateColormap
	cmapID := uint32(2)
	reqBody := make([]byte, 16)
	reqBody[0] = 0 // alloc = None
	order.PutUint32(reqBody[4:8], cmapID)
	req := &request{
		opcode:   CreateColormap,
		sequence: 20,
		body:     reqBody,
	}
	reply := s.handleRequest(mockClient, req)
	assert.Nil(t, reply, "CreateColormap should not send a reply")
	_, ok := s.colormaps[mockClient.xID(cmapID)]
	assert.True(t, ok, "Colormap should be created")

	// Test AllocColor
	reqBody = make([]byte, 12)
	order.PutUint32(reqBody[0:4], cmapID)
	order.PutUint16(reqBody[4:6], 0x8000)  // red
	order.PutUint16(reqBody[6:8], 0x4000)  // green
	order.PutUint16(reqBody[8:10], 0x2000) // blue
	req = &request{
		opcode:   AllocColor,
		sequence: 21,
		body:     reqBody,
	}
	reply = s.handleRequest(mockClient, req)
	assert.NotNil(t, reply, "AllocColor should send a reply")
	replyMsg := reply.encodeMessage(order)
	assert.Equal(t, 32, len(replyMsg), "AllocColor reply length should be 32")
	pixel := order.Uint32(replyMsg[8:12])
	assert.Equal(t, uint32(0x804020), pixel, "Allocated pixel value should be correct")

	// Test QueryColors
	reqBody = make([]byte, 8)
	order.PutUint32(reqBody[0:4], cmapID)
	order.PutUint32(reqBody[4:8], pixel)
	req = &request{
		opcode:   QueryColors,
		sequence: 22,
		body:     reqBody,
	}
	reply = s.handleRequest(mockClient, req)
	assert.NotNil(t, reply, "QueryColors should send a reply")
	replyMsg = reply.encodeMessage(order)
	// 32 bytes header + 1 color * 8 bytes
	assert.Equal(t, 40, len(replyMsg), "QueryColors reply length should be 40")
	numColors := order.Uint16(replyMsg[8:10])
	assert.Equal(t, uint16(1), numColors, "Number of colors should be 1")
	red := order.Uint16(replyMsg[32:34])
	green := order.Uint16(replyMsg[34:36])
	blue := order.Uint16(replyMsg[36:38])
	assert.Equal(t, uint16(0x8000), red, "Red component should be correct")
	assert.Equal(t, uint16(0x4000), green, "Green component should be correct")
	assert.Equal(t, uint16(0x2000), blue, "Blue component should be correct")

	// Test StoreColors
	reqBody = make([]byte, 16)
	order.PutUint32(reqBody[0:4], cmapID)
	order.PutUint32(reqBody[4:8], pixel)
	order.PutUint16(reqBody[8:10], 0xA000)  // new red
	order.PutUint16(reqBody[10:12], 0xB000) // new green
	order.PutUint16(reqBody[12:14], 0xC000) // new blue
	reqBody[14] = DoRed | DoGreen | DoBlue
	req = &request{
		opcode:   StoreColors,
		sequence: 23,
		body:     reqBody,
	}
	reply = s.handleRequest(mockClient, req)
	assert.Nil(t, reply, "StoreColors should not send a reply")
	storedColor := s.colormaps[mockClient.xID(cmapID)].pixels[pixel]
	assert.Equal(t, uint16(0xA000), storedColor.Red, "Stored red component should be updated")

	// Test FreeColors
	reqBody = make([]byte, 12)
	order.PutUint32(reqBody[0:4], cmapID)
	order.PutUint32(reqBody[8:12], pixel)
	req = &request{
		opcode:   FreeColors,
		sequence: 24,
		body:     reqBody,
	}
	reply = s.handleRequest(mockClient, req)
	assert.Nil(t, reply, "FreeColors should not send a reply")
	_, ok = s.colormaps[mockClient.xID(cmapID)].pixels[pixel]
	assert.False(t, ok, "Pixel should be freed from colormap")

	// Test AllocNamedColor
	colorName := "saddle brown"
	reqBody = make([]byte, 8+len(colorName))
	order.PutUint32(reqBody[0:4], cmapID)
	order.PutUint16(reqBody[4:6], uint16(len(colorName)))
	copy(reqBody[8:], colorName)
	req = &request{
		opcode:   AllocNamedColor,
		sequence: 25,
		body:     reqBody,
	}
	reply = s.handleRequest(mockClient, req)
	assert.NotNil(t, reply, "AllocNamedColor should send a reply")
	replyMsg = reply.encodeMessage(order)
	assert.Equal(t, 32, len(replyMsg), "AllocNamedColor reply length should be 32")
	pixel = order.Uint32(replyMsg[8:12])
	assert.Equal(t, uint32(0x8b4513), pixel, "Allocated named color pixel value should be correct")

	// Test LookupColor
	reqBody = make([]byte, 8+len(colorName))
	order.PutUint32(reqBody[0:4], cmapID)
	order.PutUint16(reqBody[4:6], uint16(len(colorName)))
	copy(reqBody[8:], colorName)
	req = &request{
		opcode:   LookupColor,
		sequence: 26,
		body:     reqBody,
	}
	reply = s.handleRequest(mockClient, req)
	assert.NotNil(t, reply, "LookupColor should send a reply")
	replyMsg = reply.encodeMessage(order)
	assert.Equal(t, 32, len(replyMsg), "LookupColor reply length should be 32")
	exactRed := order.Uint16(replyMsg[8:10])
	assert.Equal(t, uint16(0x8B8B), exactRed, "Exact red component should be correct")

	// Test InstallColormap
	reqBody = make([]byte, 4)
	order.PutUint32(reqBody[0:4], cmapID)
	req = &request{
		opcode:   InstallColormap,
		sequence: 28,
		body:     reqBody,
	}
	reply = s.handleRequest(mockClient, req)
	assert.Nil(t, reply, "InstallColormap should not send a reply")
	assert.Equal(t, mockClient.xID(cmapID), s.installedColormap, "Installed colormap should be set")

	// Test ListInstalledColormaps
	reqBody = make([]byte, 4)
	order.PutUint32(reqBody[0:4], 0) // dummy window
	req = &request{
		opcode:   ListInstalledColormaps,
		sequence: 29,
		body:     reqBody,
	}
	reply = s.handleRequest(mockClient, req)
	assert.NotNil(t, reply, "ListInstalledColormaps should send a reply")
	replyMsg = reply.encodeMessage(order)
	assert.Equal(t, 36, len(replyMsg), "ListInstalledColormaps reply length should be 36")
	nColormaps := order.Uint16(replyMsg[8:10])
	assert.Equal(t, uint16(1), nColormaps, "Number of installed colormaps should be 1")
	installedCmapID := order.Uint32(replyMsg[32:36])
	assert.Equal(t, cmapID, installedCmapID, "Installed colormap ID should be correct")

	// Test UninstallColormap
	reqBody = make([]byte, 4)
	order.PutUint32(reqBody[0:4], cmapID)
	req = &request{
		opcode:   UninstallColormap,
		sequence: 30,
		body:     reqBody,
	}
	reply = s.handleRequest(mockClient, req)
	assert.Nil(t, reply, "UninstallColormap should not send a reply")
	assert.Equal(t, xID{local: s.defaultColormap}, s.installedColormap, "Installed colormap should be reset to default")

	// Test ListInstalledColormaps again
	reqBody = make([]byte, 4)
	order.PutUint32(reqBody[0:4], 0) // dummy window
	req = &request{
		opcode:   ListInstalledColormaps,
		sequence: 31,
		body:     reqBody,
	}
	reply = s.handleRequest(mockClient, req)
	assert.NotNil(t, reply, "ListInstalledColormaps should send a reply")
	replyMsg = reply.encodeMessage(order)
	assert.Equal(t, 36, len(replyMsg), "ListInstalledColormaps reply length should be 36")
	nColormaps = order.Uint16(replyMsg[8:10])
	assert.Equal(t, uint16(1), nColormaps, "Number of installed colormaps should be 1")
	installedCmapID = order.Uint32(replyMsg[32:36])
	assert.Equal(t, s.defaultColormap, installedCmapID, "Installed colormap ID should be the default one")

	// Test FreeColormap
	reqBody = make([]byte, 4)
	order.PutUint32(reqBody[0:4], cmapID)
	req = &request{
		opcode:   FreeColormap,
		sequence: 27,
		body:     reqBody,
	}
	reply = s.handleRequest(mockClient, req)
	assert.Nil(t, reply, "FreeColormap should not send a reply")
	_, ok = s.colormaps[mockClient.xID(cmapID)]
	assert.False(t, ok, "Colormap should be freed")
}
