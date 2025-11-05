//go:build x11

package x11

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRequestParsing(t *testing.T) {
	b, err := os.ReadFile("testdata/requests.json")
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}

	var testdata []struct {
		Raw  string `json:"raw"`
		Want string `json:"want"`
	}
	if err := json.Unmarshal(b, &testdata); err != nil {
		t.Fatalf("json: %v", err)
	}
	_, update := os.LookupEnv("UPDATE_TESTDATA")
	for i, tc := range testdata {
		req, err := hex.DecodeString(tc.Raw)
		if err != nil {
			t.Errorf("#%d %q: %v", i, tc.Raw, err)
			continue
		}
		parsedReq, err := parseRequest(binary.LittleEndian, req)
		if err != nil {
			t.Errorf("#%d parseRequest(%q): %v", i, tc.Raw, err)
			continue
		}
		got := fmt.Sprintf("%#v", parsedReq)
		if update {
			testdata[i].Want = got
			continue
		}
		if got != tc.Want {
			t.Errorf("parseRequest(%q) = %s, want %s", tc.Raw, got, tc.Want)
		}
	}
	if update {
		b, err := json.MarshalIndent(testdata, "", "  ")
		if err != nil {
			t.Fatalf("json: %v", err)
		}
		if err := os.WriteFile("testdata/requests.json", b, 0o644); err != nil {
			t.Errorf("WriteFile: %v", err)
		}
	}
}

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

	p, err := parseImageText8Request(binary.LittleEndian, payload)
	assert.NoError(t, err, "parseImageText8Request should not return an error")

	if p.Drawable != drawable {
		t.Errorf("Expected drawable %d, got %d", drawable, p.Drawable)
	}
	if p.Gc != gc {
		t.Errorf("Expected gc %d, got %d", gc, p.Gc)
	}
	if p.X != x {
		t.Errorf("Expected x %d, got %d", x, p.X)
	}
	if p.Y != y {
		t.Errorf("Expected y %d, got %d", y, p.Y)
	}
	if !bytes.Equal(p.Text, text) {
		t.Errorf("Expected text %s, got %s", text, p.Text)
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

	p, err := parseImageText16Request(binary.LittleEndian, payload)
	assert.NoError(t, err, "parseImageText16Request should not return an error")

	assert.Equal(t, drawable, p.Drawable, "drawable mismatch")
	assert.Equal(t, gc, p.Gc, "gc mismatch")
	assert.Equal(t, x, p.X, "x mismatch")
	assert.Equal(t, y, p.Y, "y mismatch")
	assert.Equal(t, text, p.Text, "text mismatch")
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

	p, err := parsePolyText8Request(binary.LittleEndian, payload)
	assert.NoError(t, err, "parsePolyText8Request should not return an error")

	assert.Equal(t, drawable, p.Drawable, "drawable mismatch")
	assert.Equal(t, gc, p.Gc, "gc mismatch")
	assert.Equal(t, x, p.X, "x mismatch")
	assert.Equal(t, y, p.Y, "y mismatch")

	expectedItems := []PolyText8Item{
		{Delta: 5, Str: []byte("Hi")},
		{Delta: 10, Str: []byte("There")},
	}
	assert.Equal(t, expectedItems, p.Items, "items mismatch")
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

	p, err := parsePolyText16Request(binary.LittleEndian, payload)
	assert.NoError(t, err, "parsePolyText16Request should not return an error")

	assert.Equal(t, drawable, p.Drawable, "drawable mismatch")
	assert.Equal(t, gc, p.Gc, "gc mismatch")
	assert.Equal(t, x, p.X, "x mismatch")
	assert.Equal(t, y, p.Y, "y mismatch")

	expectedItems := []PolyText16Item{
		{Delta: 5, Str: []uint16{0x0048, 0x0069}},
		{Delta: 10, Str: []uint16{0x0054, 0x0068, 0x0065, 0x0072, 0x0065}},
	}
	assert.Equal(t, expectedItems, p.Items, "items mismatch")
}

func TestParseQueryPointerRequest(t *testing.T) {
	order := binary.LittleEndian
	reqBody := make([]byte, 4)
	order.PutUint32(reqBody, 123)
	p, err := parseQueryPointerRequest(order, reqBody)
	assert.NoError(t, err, "parseQueryPointerRequest should not return an error")
	assert.Equal(t, uint32(123), p.Drawable, "Drawable ID should be parsed correctly")

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

	p, err := parseCopyAreaRequest(order, reqBody)
	assert.NoError(t, err, "parseCopyAreaRequest should not return an error")

	assert.Equal(t, uint32(1), p.SrcDrawable, "srcDrawable should be parsed correctly")
	assert.Equal(t, uint32(2), p.DstDrawable, "dstDrawable should be parsed correctly")
	assert.Equal(t, uint32(3), p.Gc, "gc should be parsed correctly")
	assert.Equal(t, int16(10), p.SrcX, "srcX should be parsed correctly")
	assert.Equal(t, int16(20), p.SrcY, "srcY should be parsed correctly")
	assert.Equal(t, int16(30), p.DstX, "dstX should be parsed correctly")
	assert.Equal(t, int16(40), p.DstY, "dstY should be parsed correctly")
	assert.Equal(t, uint16(100), p.Width, "width should be parsed correctly")
	assert.Equal(t, uint16(200), p.Height, "height should be parsed correctly")
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

	p, err := parseGetImageRequest(order, 2, reqBody)
	assert.NoError(t, err, "parseGetImageRequest should not return an error")

	assert.Equal(t, uint32(1), p.Drawable, "drawable should be parsed correctly")
	assert.Equal(t, byte(2), p.Format, "format should be parsed correctly")
	assert.Equal(t, int16(10), p.X, "x should be parsed correctly")
	assert.Equal(t, int16(20), p.Y, "y should be parsed correctly")
	assert.Equal(t, uint16(100), p.Width, "width should be parsed correctly")
	assert.Equal(t, uint16(200), p.Height, "height should be parsed correctly")
	assert.Equal(t, uint32(0xFFFFFFFF), p.PlaneMask, "planeMask should be parsed correctly")
}

func TestParseGetAtomNameRequest(t *testing.T) {
	order := binary.LittleEndian
	reqBody := make([]byte, 4)
	order.PutUint32(reqBody[0:4], 123) // atom

	p, err := parseGetAtomNameRequest(order, reqBody)
	assert.NoError(t, err, "parseGetAtomNameRequest should not return an error")

	assert.Equal(t, uint32(123), p.Atom, "atom should be parsed correctly")
}

func TestParseListPropertiesRequest(t *testing.T) {
	order := binary.LittleEndian
	reqBody := make([]byte, 4)
	order.PutUint32(reqBody[0:4], 123) // window

	p, err := parseListPropertiesRequest(order, reqBody)
	assert.NoError(t, err, "parseListPropertiesRequest should not return an error")

	assert.Equal(t, uint32(123), p.Window, "window should be parsed correctly")
}

func TestParseChangeWindowAttributesRequest(t *testing.T) {
	order := binary.LittleEndian
	reqBody := make([]byte, 8)
	order.PutUint32(reqBody[0:4], 123)                               // window
	order.PutUint32(reqBody[4:8], uint32(CWBackPixel|CWCursor))       // valueMask
	reqBody = append(reqBody, make([]byte, 8)...)
	order.PutUint32(reqBody[8:12], 0xFF00FF)                         // background pixel
	order.PutUint32(reqBody[12:16], 456)                             // cursor

	p, err := parseChangeWindowAttributesRequest(order, reqBody)
	assert.NoError(t, err, "parseChangeWindowAttributesRequest should not return an error")

	assert.Equal(t, uint32(123), p.Window, "window should be parsed correctly")
	assert.Equal(t, uint32(CWBackPixel|CWCursor), p.ValueMask, "valueMask should be parsed correctly")
	assert.True(t, p.Values.BackgroundPixelSet, "BackgroundPixelSet should be true")
	assert.Equal(t, uint32(0xFF00FF), p.Values.BackgroundPixel, "background pixel should be parsed correctly")
	assert.Equal(t, uint32(456), p.Values.Cursor, "cursor should be parsed correctly")
}

func TestParseGetWindowAttributesRequest(t *testing.T) {
	order := binary.LittleEndian
	reqBody := make([]byte, 4)
	order.PutUint32(reqBody[0:4], 123)

	p, err := parseGetWindowAttributesRequest(order, reqBody)
	assert.NoError(t, err, "parseGetWindowAttributesRequest should not return an error")
	assert.Equal(t, uint32(123), p.Drawable, "Drawable should be parsed correctly")
}

func TestParseUnmapWindowRequest(t *testing.T) {
	order := binary.LittleEndian
	reqBody := make([]byte, 4)
	order.PutUint32(reqBody[0:4], 123)

	p, err := parseUnmapWindowRequest(order, reqBody)
	assert.NoError(t, err, "parseUnmapWindowRequest should not return an error")
	assert.Equal(t, uint32(123), p.Window, "Window should be parsed correctly")
}

func TestParseUnmapSubwindowsRequest(t *testing.T) {
	order := binary.LittleEndian
	reqBody := make([]byte, 4)
	order.PutUint32(reqBody[0:4], 123)

	p, err := parseUnmapSubwindowsRequest(order, reqBody)
	assert.NoError(t, err, "parseUnmapSubwindowsRequest should not return an error")
	assert.Equal(t, uint32(123), p.Window, "Window should be parsed correctly")
}

func TestParseGetGeometryRequest(t *testing.T) {
	order := binary.LittleEndian
	reqBody := make([]byte, 4)
	order.PutUint32(reqBody[0:4], 123)

	p, err := parseGetGeometryRequest(order, reqBody)
	assert.NoError(t, err, "parseGetGeometryRequest should not return an error")
	assert.Equal(t, uint32(123), p.Drawable, "Drawable should be parsed correctly")
}

func TestParseDeletePropertyRequest(t *testing.T) {
	order := binary.LittleEndian
	reqBody := make([]byte, 8)
	order.PutUint32(reqBody[0:4], 123)
	order.PutUint32(reqBody[4:8], 456)

	p, err := parseDeletePropertyRequest(order, reqBody)
	assert.NoError(t, err, "parseDeletePropertyRequest should not return an error")
	assert.Equal(t, uint32(123), p.Window, "Window should be parsed correctly")
	assert.Equal(t, uint32(456), p.Property, "Property should be parsed correctly")
}

func TestParseSetSelectionOwnerRequest(t *testing.T) {
	order := binary.LittleEndian
	reqBody := make([]byte, 12)
	order.PutUint32(reqBody[0:4], 123)
	order.PutUint32(reqBody[4:8], 456)
	order.PutUint32(reqBody[8:12], 789)

	p, err := parseSetSelectionOwnerRequest(order, reqBody)
	assert.NoError(t, err, "parseSetSelectionOwnerRequest should not return an error")
	assert.Equal(t, uint32(123), p.Owner, "Owner should be parsed correctly")
	assert.Equal(t, uint32(456), p.Selection, "Selection should be parsed correctly")
	assert.Equal(t, uint32(789), p.Time, "Time should be parsed correctly")
}

func TestParseGetSelectionOwnerRequest(t *testing.T) {
	order := binary.LittleEndian
	reqBody := make([]byte, 4)
	order.PutUint32(reqBody[0:4], 123)

	p, err := parseGetSelectionOwnerRequest(order, reqBody)
	assert.NoError(t, err, "parseGetSelectionOwnerRequest should not return an error")
	assert.Equal(t, uint32(123), p.Selection, "Selection should be parsed correctly")
}

func TestParseConvertSelectionRequest(t *testing.T) {
	order := binary.LittleEndian
	reqBody := make([]byte, 20)
	order.PutUint32(reqBody[0:4], 1)
	order.PutUint32(reqBody[4:8], 2)
	order.PutUint32(reqBody[8:12], 3)
	order.PutUint32(reqBody[12:16], 4)
	order.PutUint32(reqBody[16:20], 5)

	p, err := parseConvertSelectionRequest(order, reqBody)
	assert.NoError(t, err, "parseConvertSelectionRequest should not return an error")
	assert.Equal(t, uint32(1), p.Requestor, "Requestor should be parsed correctly")
	assert.Equal(t, uint32(2), p.Selection, "Selection should be parsed correctly")
	assert.Equal(t, uint32(3), p.Target, "Target should be parsed correctly")
	assert.Equal(t, uint32(4), p.Property, "Property should be parsed correctly")
	assert.Equal(t, uint32(5), p.Time, "Time should be parsed correctly")
}

func TestParseSendEventRequest(t *testing.T) {
	order := binary.LittleEndian
	reqBody := make([]byte, 44)
	order.PutUint32(reqBody[4:8], 123)
	order.PutUint32(reqBody[8:12], 456)
	for i := 12; i < 44; i++ {
		reqBody[i] = byte(i)
	}

	p, err := parseSendEventRequest(order, reqBody)
	assert.NoError(t, err, "parseSendEventRequest should not return an error")
	assert.Equal(t, uint32(123), p.Destination, "Destination should be parsed correctly")
	assert.Equal(t, uint32(456), p.EventMask, "EventMask should be parsed correctly")
	assert.Equal(t, reqBody[12:44], p.EventData, "EventData should be parsed correctly")
}

func TestParseGrabPointerRequest(t *testing.T) {
	order := binary.LittleEndian
	reqBody := make([]byte, 20)
	order.PutUint32(reqBody[0:4], 123)
	order.PutUint16(reqBody[4:6], 456)
	reqBody[6] = 1
	reqBody[7] = 2
	order.PutUint32(reqBody[8:12], 789)
	order.PutUint32(reqBody[12:16], 101)
	order.PutUint32(reqBody[16:20], 112)

	p, err := parseGrabPointerRequest(order, reqBody)
	assert.NoError(t, err, "parseGrabPointerRequest should not return an error")
	assert.Equal(t, uint32(123), p.GrabWindow, "GrabWindow should be parsed correctly")
	assert.Equal(t, uint16(456), p.EventMask, "EventMask should be parsed correctly")
	assert.Equal(t, byte(1), p.PointerMode, "PointerMode should be parsed correctly")
	assert.Equal(t, byte(2), p.KeyboardMode, "KeyboardMode should be parsed correctly")
	assert.Equal(t, uint32(789), p.ConfineTo, "ConfineTo should be parsed correctly")
	assert.Equal(t, uint32(101), p.Cursor, "Cursor should be parsed correctly")
	assert.Equal(t, uint32(112), p.Time, "Time should be parsed correctly")
}

func TestParseUngrabPointerRequest(t *testing.T) {
	order := binary.LittleEndian
	reqBody := make([]byte, 4)
	order.PutUint32(reqBody[0:4], 123)

	p, err := parseUngrabPointerRequest(order, reqBody)
	assert.NoError(t, err, "parseUngrabPointerRequest should not return an error")
	assert.Equal(t, uint32(123), p.Time, "Time should be parsed correctly")
}

func TestParseGrabKeyboardRequest(t *testing.T) {
	order := binary.LittleEndian
	reqBody := make([]byte, 12)
	order.PutUint32(reqBody[0:4], 123)
	order.PutUint32(reqBody[4:8], 456)
	reqBody[8] = 1
	reqBody[9] = 2

	p, err := parseGrabKeyboardRequest(order, reqBody)
	assert.NoError(t, err, "parseGrabKeyboardRequest should not return an error")
	assert.Equal(t, uint32(123), p.GrabWindow, "GrabWindow should be parsed correctly")
	assert.Equal(t, uint32(456), p.Time, "Time should be parsed correctly")
	assert.Equal(t, byte(1), p.PointerMode, "PointerMode should be parsed correctly")
	assert.Equal(t, byte(2), p.KeyboardMode, "KeyboardMode should be parsed correctly")
}

func TestParseUngrabKeyboardRequest(t *testing.T) {
	order := binary.LittleEndian
	reqBody := make([]byte, 4)
	order.PutUint32(reqBody[0:4], 123)

	p, err := parseUngrabKeyboardRequest(order, reqBody)
	assert.NoError(t, err, "parseUngrabKeyboardRequest should not return an error")
	assert.Equal(t, uint32(123), p.Time, "Time should be parsed correctly")
}

func TestParseAllowEventsRequest(t *testing.T) {
	order := binary.LittleEndian
	reqBody := make([]byte, 4)
	order.PutUint32(reqBody[0:4], 123)

	p, err := parseAllowEventsRequest(order, 5, reqBody)
	assert.NoError(t, err, "parseAllowEventsRequest should not return an error")
	assert.Equal(t, byte(5), p.Mode, "Mode should be parsed correctly")
	assert.Equal(t, uint32(123), p.Time, "Time should be parsed correctly")
}

func TestParseTranslateCoordsRequest(t *testing.T) {
	order := binary.LittleEndian
	reqBody := make([]byte, 12)
	order.PutUint32(reqBody[0:4], 1)
	order.PutUint32(reqBody[4:8], 2)
	order.PutUint16(reqBody[8:10], 10)
	order.PutUint16(reqBody[10:12], 20)

	p, err := parseTranslateCoordsRequest(order, reqBody)
	assert.NoError(t, err, "parseTranslateCoordsRequest should not return an error")
	assert.Equal(t, uint32(1), p.SrcWindow, "SrcWindow should be parsed correctly")
	assert.Equal(t, uint32(2), p.DstWindow, "DstWindow should be parsed correctly")
	assert.Equal(t, int16(10), p.SrcX, "SrcX should be parsed correctly")
	assert.Equal(t, int16(20), p.SrcY, "SrcY should be parsed correctly")
}

func TestParseWarpPointerRequest(t *testing.T) {
	order := binary.LittleEndian
	reqBody := make([]byte, 16)
	order.PutUint16(reqBody[12:14], 10)
	order.PutUint16(reqBody[14:16], 20)

	p, err := parseWarpPointerRequest(order, reqBody)
	assert.NoError(t, err, "parseWarpPointerRequest should not return an error")
	assert.Equal(t, int16(10), p.DstX, "DstX should be parsed correctly")
	assert.Equal(t, int16(20), p.DstY, "DstY should be parsed correctly")
}

func TestParseSetInputFocusRequest(t *testing.T) {
	order := binary.LittleEndian
	reqBody := make([]byte, 12)
	order.PutUint32(reqBody[0:4], 123)
	reqBody[4] = 2
	order.PutUint32(reqBody[8:12], 456)

	p, err := parseSetInputFocusRequest(order, reqBody)
	assert.NoError(t, err, "parseSetInputFocusRequest should not return an error")
	assert.Equal(t, uint32(123), p.Focus, "Focus should be parsed correctly")
	assert.Equal(t, byte(2), p.RevertTo, "RevertTo should be parsed correctly")
	assert.Equal(t, uint32(456), p.Time, "Time should be parsed correctly")
}

func TestParseCloseFontRequest(t *testing.T) {
	order := binary.LittleEndian
	reqBody := make([]byte, 4)
	order.PutUint32(reqBody[0:4], 123)

	p, err := parseCloseFontRequest(order, reqBody)
	assert.NoError(t, err, "parseCloseFontRequest should not return an error")
	assert.Equal(t, uint32(123), p.Fid, "Fid should be parsed correctly")
}

func TestParseFreePixmapRequest(t *testing.T) {
	order := binary.LittleEndian
	reqBody := make([]byte, 4)
	order.PutUint32(reqBody[0:4], 123)

	p, err := parseFreePixmapRequest(order, reqBody)
	assert.NoError(t, err, "parseFreePixmapRequest should not return an error")
	assert.Equal(t, uint32(123), p.Pid, "Pid should be parsed correctly")
}

func TestParseChangeGCRequest(t *testing.T) {
	order := binary.LittleEndian
	reqBody := make([]byte, 8)
	order.PutUint32(reqBody[0:4], 123)
	order.PutUint32(reqBody[4:8], uint32(GCForeground|GCBackground))
	reqBody = append(reqBody, make([]byte, 8)...)
	order.PutUint32(reqBody[8:12], 0xFF00FF)
	order.PutUint32(reqBody[12:16], 0x00FF00)

	p, err := parseChangeGCRequest(order, reqBody)
	assert.NoError(t, err, "parseChangeGCRequest should not return an error")
	assert.Equal(t, uint32(123), p.Gc, "Gc should be parsed correctly")
	assert.Equal(t, uint32(GCForeground|GCBackground), p.ValueMask, "ValueMask should be parsed correctly")
	assert.Equal(t, uint32(0xFF00FF), p.Values.Foreground, "Foreground should be parsed correctly")
	assert.Equal(t, uint32(0x00FF00), p.Values.Background, "Background should be parsed correctly")
}

func TestParseCopyGCRequest(t *testing.T) {
	order := binary.LittleEndian
	reqBody := make([]byte, 8)
	order.PutUint32(reqBody[0:4], 123)
	order.PutUint32(reqBody[4:8], 456)

	p, err := parseCopyGCRequest(order, reqBody)
	assert.NoError(t, err, "parseCopyGCRequest should not return an error")
	assert.Equal(t, uint32(123), p.SrcGC, "SrcGC should be parsed correctly")
	assert.Equal(t, uint32(456), p.DstGC, "DstGC should be parsed correctly")
}

func TestParseClearAreaRequest(t *testing.T) {
	order := binary.LittleEndian
	reqBody := make([]byte, 12)
	order.PutUint32(reqBody[0:4], 123)
	order.PutUint16(reqBody[4:6], 10)
	order.PutUint16(reqBody[6:8], 20)
	order.PutUint16(reqBody[8:10], 100)
	order.PutUint16(reqBody[10:12], 200)

	p, err := parseClearAreaRequest(order, reqBody)
	assert.NoError(t, err, "parseClearAreaRequest should not return an error")
	assert.Equal(t, uint32(123), p.Window, "Window should be parsed correctly")
	assert.Equal(t, int16(10), p.X, "X should be parsed correctly")
	assert.Equal(t, int16(20), p.Y, "Y should be parsed correctly")
	assert.Equal(t, uint16(100), p.Width, "Width should be parsed correctly")
	assert.Equal(t, uint16(200), p.Height, "Height should be parsed correctly")
}

func TestParsePolyPointRequest(t *testing.T) {
	order := binary.LittleEndian
	reqBody := make([]byte, 16)
	order.PutUint32(reqBody[0:4], 123)
	order.PutUint32(reqBody[4:8], 456)
	order.PutUint16(reqBody[8:10], 10)
	order.PutUint16(reqBody[10:12], 20)
	order.PutUint16(reqBody[12:14], 30)
	order.PutUint16(reqBody[14:16], 40)

	p, err := parsePolyPointRequest(order, reqBody)
	assert.NoError(t, err, "parsePolyPointRequest should not return an error")
	assert.Equal(t, uint32(123), p.Drawable, "Drawable should be parsed correctly")
	assert.Equal(t, uint32(456), p.Gc, "Gc should be parsed correctly")
	assert.Equal(t, []uint32{10, 20, 30, 40}, p.Coordinates, "Coordinates should be parsed correctly")
}

func TestParsePolyRectangleRequest(t *testing.T) {
	order := binary.LittleEndian
	reqBody := make([]byte, 24)
	order.PutUint32(reqBody[0:4], 123)
	order.PutUint32(reqBody[4:8], 456)
	order.PutUint16(reqBody[8:10], 10)
	order.PutUint16(reqBody[10:12], 20)
	order.PutUint16(reqBody[12:14], 100)
	order.PutUint16(reqBody[14:16], 200)
	order.PutUint16(reqBody[16:18], 30)
	order.PutUint16(reqBody[18:20], 40)
	order.PutUint16(reqBody[20:22], 50)
	order.PutUint16(reqBody[22:24], 60)

	p, err := parsePolyRectangleRequest(order, reqBody)
	assert.NoError(t, err, "parsePolyRectangleRequest should not return an error")
	assert.Equal(t, uint32(123), p.Drawable, "Drawable should be parsed correctly")
	assert.Equal(t, uint32(456), p.Gc, "Gc should be parsed correctly")
	assert.Equal(t, []uint32{10, 20, 100, 200, 30, 40, 50, 60}, p.Rectangles, "Rectangles should be parsed correctly")
}

func TestParsePolyArcRequest(t *testing.T) {
	order := binary.LittleEndian
	reqBody := make([]byte, 32)
	order.PutUint32(reqBody[0:4], 123)
	order.PutUint32(reqBody[4:8], 456)
	order.PutUint16(reqBody[8:10], 10)
	order.PutUint16(reqBody[10:12], 20)
	order.PutUint16(reqBody[12:14], 100)
	order.PutUint16(reqBody[14:16], 200)
	order.PutUint16(reqBody[16:18], 90)
	order.PutUint16(reqBody[18:20], 180)
	order.PutUint16(reqBody[20:22], 30)
	order.PutUint16(reqBody[22:24], 40)
	order.PutUint16(reqBody[24:26], 50)
	order.PutUint16(reqBody[26:28], 60)
	order.PutUint16(reqBody[28:30], 270)
	order.PutUint16(reqBody[30:32], 360)

	p, err := parsePolyArcRequest(order, reqBody)
	assert.NoError(t, err, "parsePolyArcRequest should not return an error")
	assert.Equal(t, uint32(123), p.Drawable, "Drawable should be parsed correctly")
	assert.Equal(t, uint32(456), p.Gc, "Gc should be parsed correctly")
	assert.Equal(t, []uint32{10, 20, 100, 200, 90, 180, 30, 40, 50, 60, 270, 360}, p.Arcs, "Arcs should be parsed correctly")
}

func TestParseCreateColormapRequest(t *testing.T) {
	order := binary.LittleEndian
	reqBody := make([]byte, 16)
	reqBody[0] = 1
	order.PutUint32(reqBody[4:8], 123)
	order.PutUint32(reqBody[8:12], 456)
	order.PutUint32(reqBody[12:16], 789)

	p, err := parseCreateColormapRequest(order, reqBody)
	assert.NoError(t, err, "parseCreateColormapRequest should not return an error")
	assert.Equal(t, byte(1), p.Alloc, "Alloc should be parsed correctly")
	assert.Equal(t, uint32(123), p.Mid, "Mid should be parsed correctly")
	assert.Equal(t, uint32(456), p.Window, "Window should be parsed correctly")
	assert.Equal(t, uint32(789), p.Visual, "Visual should be parsed correctly")
}

func TestParseFreeColormapRequest(t *testing.T) {
	order := binary.LittleEndian
	reqBody := make([]byte, 4)
	order.PutUint32(reqBody[0:4], 123)

	p, err := parseFreeColormapRequest(order, reqBody)
	assert.NoError(t, err, "parseFreeColormapRequest should not return an error")
	assert.Equal(t, uint32(123), p.Cmap, "Cmap should be parsed correctly")
}

func TestParseInstallColormapRequest(t *testing.T) {
	order := binary.LittleEndian
	reqBody := make([]byte, 4)
	order.PutUint32(reqBody[0:4], 123)

	p, err := parseInstallColormapRequest(order, reqBody)
	assert.NoError(t, err, "parseInstallColormapRequest should not return an error")
	assert.Equal(t, uint32(123), p.Cmap, "Cmap should be parsed correctly")
}

func TestParseUninstallColormapRequest(t *testing.T) {
	order := binary.LittleEndian
	reqBody := make([]byte, 4)
	order.PutUint32(reqBody[0:4], 123)

	p, err := parseUninstallColormapRequest(order, reqBody)
	assert.NoError(t, err, "parseUninstallColormapRequest should not return an error")
	assert.Equal(t, uint32(123), p.Cmap, "Cmap should be parsed correctly")
}

func TestParseListInstalledColormapsRequest(t *testing.T) {
	order := binary.LittleEndian
	reqBody := make([]byte, 4)
	order.PutUint32(reqBody[0:4], 123)

	p, err := parseListInstalledColormapsRequest(order, reqBody)
	assert.NoError(t, err, "parseListInstalledColormapsRequest should not return an error")
	assert.Equal(t, uint32(123), p.Window, "Window should be parsed correctly")
}

func TestParseAllocColorRequest(t *testing.T) {
	order := binary.LittleEndian
	reqBody := make([]byte, 12)
	order.PutUint32(reqBody[0:4], 123)
	order.PutUint16(reqBody[4:6], 100)
	order.PutUint16(reqBody[6:8], 200)
	order.PutUint16(reqBody[8:10], 255)

	p, err := parseAllocColorRequest(order, reqBody)
	assert.NoError(t, err, "parseAllocColorRequest should not return an error")
	assert.Equal(t, uint32(123), p.Cmap, "Cmap should be parsed correctly")
	assert.Equal(t, uint16(100), p.Red, "Red should be parsed correctly")
	assert.Equal(t, uint16(200), p.Green, "Green should be parsed correctly")
	assert.Equal(t, uint16(255), p.Blue, "Blue should be parsed correctly")
}

func TestParseAllocNamedColorRequest(t *testing.T) {
	order := binary.LittleEndian
	reqBody := make([]byte, 12)
	order.PutUint32(reqBody[0:4], 123)
	order.PutUint16(reqBody[4:6], 4)
	copy(reqBody[8:12], []byte("blue"))

	p, err := parseAllocNamedColorRequest(order, reqBody)
	assert.NoError(t, err, "parseAllocNamedColorRequest should not return an error")
	assert.Equal(t, xID{local: 123}, p.Cmap, "Cmap should be parsed correctly")
	assert.Equal(t, []byte("blue"), p.Name, "Name should be parsed correctly")
}

func TestParseFreeColorsRequest(t *testing.T) {
	order := binary.LittleEndian
	reqBody := make([]byte, 16)
	order.PutUint32(reqBody[0:4], 123)
	order.PutUint32(reqBody[4:8], 0xFF)
	order.PutUint32(reqBody[8:12], 1)
	order.PutUint32(reqBody[12:16], 2)

	p, err := parseFreeColorsRequest(order, reqBody)
	assert.NoError(t, err, "parseFreeColorsRequest should not return an error")
	assert.Equal(t, uint32(123), p.Cmap, "Cmap should be parsed correctly")
	assert.Equal(t, uint32(0xFF), p.PlaneMask, "PlaneMask should be parsed correctly")
	assert.Equal(t, []uint32{1, 2}, p.Pixels, "Pixels should be parsed correctly")
}

func TestParseStoreColorsRequest(t *testing.T) {
	order := binary.LittleEndian
	reqBody := make([]byte, 28)
	order.PutUint32(reqBody[0:4], 123)
	// Item 1
	order.PutUint32(reqBody[4:8], 1)
	order.PutUint16(reqBody[8:10], 10)
	order.PutUint16(reqBody[10:12], 20)
	order.PutUint16(reqBody[12:14], 30)
	reqBody[14] = 7
	// Item 2
	order.PutUint32(reqBody[16:20], 2)
	order.PutUint16(reqBody[20:22], 40)
	order.PutUint16(reqBody[22:24], 50)
	order.PutUint16(reqBody[24:26], 60)
	reqBody[26] = 3

	p, err := parseStoreColorsRequest(order, reqBody)
	assert.NoError(t, err, "parseStoreColorsRequest should not return an error")
	assert.Equal(t, uint32(123), p.Cmap, "Cmap should be parsed correctly")
	assert.Equal(t, uint32(1), p.Items[0].Pixel, "Item 1 Pixel should be parsed correctly")
	assert.Equal(t, uint16(10), p.Items[0].Red, "Item 1 Red should be parsed correctly")
	assert.Equal(t, uint16(20), p.Items[0].Green, "Item 1 Green should be parsed correctly")
	assert.Equal(t, uint16(30), p.Items[0].Blue, "Item 1 Blue should be parsed correctly")
	assert.Equal(t, byte(7), p.Items[0].Flags, "Item 1 Flags should be parsed correctly")
	assert.Equal(t, uint32(2), p.Items[1].Pixel, "Item 2 Pixel should be parsed correctly")
	assert.Equal(t, uint16(40), p.Items[1].Red, "Item 2 Red should be parsed correctly")
	assert.Equal(t, uint16(50), p.Items[1].Green, "Item 2 Green should be parsed correctly")
	assert.Equal(t, uint16(60), p.Items[1].Blue, "Item 2 Blue should be parsed correctly")
	assert.Equal(t, byte(3), p.Items[1].Flags, "Item 2 Flags should be parsed correctly")
}

func TestParseStoreNamedColorRequest(t *testing.T) {
	order := binary.LittleEndian
	reqBody := make([]byte, 16)
	order.PutUint32(reqBody[0:4], 123)
	order.PutUint32(reqBody[4:8], 456)
	order.PutUint16(reqBody[8:10], 4)
	copy(reqBody[12:16], []byte("blue"))

	p, err := parseStoreNamedColorRequest(order, 7, reqBody)
	assert.NoError(t, err, "parseStoreNamedColorRequest should not return an error")
	assert.Equal(t, uint32(123), p.Cmap, "Cmap should be parsed correctly")
	assert.Equal(t, uint32(456), p.Pixel, "Pixel should be parsed correctly")
	assert.Equal(t, "blue", p.Name, "Name should be parsed correctly")
	assert.Equal(t, byte(7), p.Flags, "Flags should be parsed correctly")
}

func TestParseLookupColorRequest(t *testing.T) {
	order := binary.LittleEndian
	reqBody := make([]byte, 12)
	order.PutUint32(reqBody[0:4], 123)
	order.PutUint16(reqBody[4:6], 4)
	copy(reqBody[8:12], []byte("blue"))

	p, err := parseLookupColorRequest(order, reqBody)
	assert.NoError(t, err, "parseLookupColorRequest should not return an error")
	assert.Equal(t, uint32(123), p.Cmap, "Cmap should be parsed correctly")
	assert.Equal(t, "blue", p.Name, "Name should be parsed correctly")
}

func TestParseFreeCursorRequest(t *testing.T) {
	order := binary.LittleEndian
	reqBody := make([]byte, 4)
	order.PutUint32(reqBody[0:4], 123)

	p, err := parseFreeCursorRequest(order, reqBody)
	assert.NoError(t, err, "parseFreeCursorRequest should not return an error")
	assert.Equal(t, uint32(123), p.Cursor, "Cursor should be parsed correctly")
}

func TestParseQueryBestSizeRequest(t *testing.T) {
	order := binary.LittleEndian
	reqBody := make([]byte, 8)
	order.PutUint32(reqBody[0:4], 123)
	order.PutUint16(reqBody[4:6], 100)
	order.PutUint16(reqBody[6:8], 200)

	p, err := parseQueryBestSizeRequest(order, reqBody)
	assert.NoError(t, err, "parseQueryBestSizeRequest should not return an error")
	assert.Equal(t, uint32(123), p.Drawable, "Drawable should be parsed correctly")
	assert.Equal(t, uint16(100), p.Width, "Width should be parsed correctly")
	assert.Equal(t, uint16(200), p.Height, "Height should be parsed correctly")
}

func TestParseBellRequest(t *testing.T) {
	p, err := parseBellRequest(50)
	assert.NoError(t, err, "parseBellRequest should not return an error")
	assert.Equal(t, int8(50), p.Percent, "Percent should be parsed correctly")
}
