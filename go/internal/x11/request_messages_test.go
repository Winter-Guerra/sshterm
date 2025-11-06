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

func TestRequestParsingErrors(t *testing.T) {
	testCases := []struct {
		reqType reqCode
		raw     []byte
	}{
		{CreateWindow, make([]byte, 27)},
		{ChangeWindowAttributes, make([]byte, 7)},
		{GetWindowAttributes, make([]byte, 3)},
		{DestroyWindow, make([]byte, 3)},
		{DestroySubwindows, make([]byte, 3)},
		{ChangeSaveSet, make([]byte, 4)},
		{ReparentWindow, make([]byte, 11)},
		{MapWindow, make([]byte, 3)},
		{MapSubwindows, make([]byte, 3)},
		{UnmapWindow, make([]byte, 3)},
		{UnmapSubwindows, make([]byte, 3)},
		{ConfigureWindow, make([]byte, 7)},
		{CirculateWindow, make([]byte, 3)},
		{GetGeometry, make([]byte, 3)},
		{QueryTree, make([]byte, 3)},
		{InternAtom, make([]byte, 3)},
		{GetAtomName, make([]byte, 3)},
		{ChangeProperty, make([]byte, 19)},
		{DeleteProperty, make([]byte, 7)},
		{GetProperty, make([]byte, 19)},
		{ListProperties, make([]byte, 3)},
		{SetSelectionOwner, make([]byte, 11)},
		{GetSelectionOwner, make([]byte, 3)},
		{ConvertSelection, make([]byte, 19)},
		{SendEvent, make([]byte, 43)},
		{GrabPointer, make([]byte, 19)},
		{UngrabPointer, make([]byte, 3)},
		{GrabButton, make([]byte, 23)},
		{UngrabButton, make([]byte, 7)},
		{ChangeActivePointerGrab, make([]byte, 11)},
		{GrabKeyboard, make([]byte, 11)},
		{UngrabKeyboard, make([]byte, 3)},
		{GrabKey, make([]byte, 12)},
		{UngrabKey, make([]byte, 7)},
		{AllowEvents, make([]byte, 3)},
		{QueryPointer, make([]byte, 3)},
		{GetMotionEvents, make([]byte, 11)},
		{TranslateCoords, make([]byte, 11)},
		{WarpPointer, make([]byte, 15)},
		{SetInputFocus, make([]byte, 11)},
		{OpenFont, make([]byte, 7)},
		{CloseFont, make([]byte, 3)},
		{QueryFont, make([]byte, 3)},
		{QueryTextExtents, make([]byte, 3)},
		{ListFonts, make([]byte, 3)},
		{ListFontsWithInfo, make([]byte, 3)},
		{SetFontPath, make([]byte, 3)},
		{CreatePixmap, make([]byte, 11)},
		{FreePixmap, make([]byte, 3)},
		{CreateGC, make([]byte, 11)},
		{ChangeGC, make([]byte, 7)},
		{CopyGC, make([]byte, 7)},
		{SetDashes, make([]byte, 7)},
		{SetClipRectangles, make([]byte, 7)},
		{FreeGC, make([]byte, 3)},
		{ClearArea, make([]byte, 11)},
		{CopyArea, make([]byte, 27)},
		{PolyPoint, make([]byte, 7)},
		{PolyLine, make([]byte, 7)},
		{PolySegment, make([]byte, 7)},
		{PolyRectangle, make([]byte, 7)},
		{PolyArc, make([]byte, 7)},
		{FillPoly, make([]byte, 11)},
		{PolyFillRectangle, make([]byte, 7)},
		{PolyFillArc, make([]byte, 7)},
		{PutImage, make([]byte, 19)},
		{GetImage, make([]byte, 15)},
		{PolyText8, make([]byte, 11)},
		{PolyText16, make([]byte, 11)},
		{ImageText8, make([]byte, 11)},
		{ImageText16, make([]byte, 11)},
		{CreateColormap, make([]byte, 15)},
		{FreeColormap, make([]byte, 3)},
		{InstallColormap, make([]byte, 3)},
		{UninstallColormap, make([]byte, 3)},
		{ListInstalledColormaps, make([]byte, 3)},
		{AllocColor, make([]byte, 9)},
		{AllocNamedColor, make([]byte, 7)},
		{FreeColors, make([]byte, 7)},
		{StoreColors, make([]byte, 3)},
		{StoreNamedColor, make([]byte, 11)},
		{QueryColors, make([]byte, 3)},
		{LookupColor, make([]byte, 7)},
		{CreateGlyphCursor, make([]byte, 27)},
		{FreeCursor, make([]byte, 3)},
		{RecolorCursor, make([]byte, 15)},
		{QueryBestSize, make([]byte, 7)},
		{QueryExtension, make([]byte, 3)},
		{GetKeyboardMapping, make([]byte, 1)},
		{ChangeKeyboardMapping, make([]byte, 3)},
		{ChangeKeyboardControl, make([]byte, 3)},
		{SetScreenSaver, make([]byte, 5)},
		{ChangeHosts, make([]byte, 3)},
		{KillClient, make([]byte, 3)},
		{RotateProperties, make([]byte, 7)},
		{SetModifierMapping, make([]byte, 0)},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("%T", tc.reqType), func(t *testing.T) {
			hdr := make([]byte, 4)
			hdr[0] = byte(tc.reqType)
			binary.LittleEndian.PutUint16(hdr[2:4], uint16(len(tc.raw)/4))
			_, err := parseRequest(binary.LittleEndian, append(hdr, tc.raw...))
			assert.Error(t, err, "parseRequest should return an error for undersized requests")
		})
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

	if p.Drawable != Drawable(drawable) {
		t.Errorf("Expected drawable %d, got %d", drawable, p.Drawable)
	}
	if p.Gc != GContext(gc) {
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

	assert.Equal(t, Drawable(drawable), p.Drawable, "drawable mismatch")
	assert.Equal(t, GContext(gc), p.Gc, "gc mismatch")
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

	assert.Equal(t, Drawable(drawable), p.Drawable, "drawable mismatch")
	assert.Equal(t, GContext(gc), p.Gc, "gc mismatch")
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

	assert.Equal(t, Drawable(drawable), p.Drawable, "drawable mismatch")
	assert.Equal(t, GContext(gc), p.Gc, "gc mismatch")
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
	assert.Equal(t, Drawable(123), p.Drawable, "Drawable ID should be parsed correctly")

}

func TestParseGetMotionEventsRequest(t *testing.T) {
	order := binary.LittleEndian
	reqBody := make([]byte, 12)
	order.PutUint32(reqBody[0:4], 123)
	order.PutUint32(reqBody[4:8], 456)
	order.PutUint32(reqBody[8:12], 789)

	p, err := parseGetMotionEventsRequest(order, reqBody)
	assert.NoError(t, err, "parseGetMotionEventsRequest should not return an error")
	assert.Equal(t, Window(123), p.Window, "Window should be parsed correctly")
	assert.Equal(t, Timestamp(456), p.Start, "Start should be parsed correctly")
	assert.Equal(t, Timestamp(789), p.Stop, "Stop should be parsed correctly")
}

func TestParseCopyAreaRequest(t *testing.T) {
	order := binary.LittleEndian
	reqBody := make([]byte, 28)
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

	assert.Equal(t, Drawable(1), p.SrcDrawable, "srcDrawable should be parsed correctly")
	assert.Equal(t, Drawable(2), p.DstDrawable, "dstDrawable should be parsed correctly")
	assert.Equal(t, GContext(3), p.Gc, "gc should be parsed correctly")
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

	assert.Equal(t, Drawable(1), p.Drawable, "drawable should be parsed correctly")
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

	assert.Equal(t, Atom(123), p.Atom, "atom should be parsed correctly")
}

func TestParseListPropertiesRequest(t *testing.T) {
	order := binary.LittleEndian
	reqBody := make([]byte, 4)
	order.PutUint32(reqBody[0:4], 123) // window

	p, err := parseListPropertiesRequest(order, reqBody)
	assert.NoError(t, err, "parseListPropertiesRequest should not return an error")

	assert.Equal(t, Window(123), p.Window, "window should be parsed correctly")
}

func TestParseChangeWindowAttributesRequest(t *testing.T) {
	order := binary.LittleEndian
	reqBody := make([]byte, 8)
	order.PutUint32(reqBody[0:4], 123)                          // window
	order.PutUint32(reqBody[4:8], uint32(CWBackPixel|CWCursor)) // valueMask
	reqBody = append(reqBody, make([]byte, 8)...)
	order.PutUint32(reqBody[8:12], 0xFF00FF) // background pixel
	order.PutUint32(reqBody[12:16], 456)     // cursor

	p, err := parseChangeWindowAttributesRequest(order, reqBody)
	assert.NoError(t, err, "parseChangeWindowAttributesRequest should not return an error")

	assert.Equal(t, Window(123), p.Window, "window should be parsed correctly")
	assert.Equal(t, uint32(CWBackPixel|CWCursor), p.ValueMask, "valueMask should be parsed correctly")
	assert.True(t, p.Values.BackgroundPixelSet, "BackgroundPixelSet should be true")
	assert.Equal(t, uint32(0xFF00FF), p.Values.BackgroundPixel, "background pixel should be parsed correctly")
	assert.Equal(t, Cursor(456), p.Values.Cursor, "cursor should be parsed correctly")
}

func TestParseGetWindowAttributesRequest(t *testing.T) {
	order := binary.LittleEndian
	reqBody := make([]byte, 4)
	order.PutUint32(reqBody[0:4], 123)

	p, err := parseGetWindowAttributesRequest(order, reqBody)
	assert.NoError(t, err, "parseGetWindowAttributesRequest should not return an error")
	assert.Equal(t, Drawable(123), p.Drawable, "Drawable should be parsed correctly")
}

func TestParseDestroyWindowRequest(t *testing.T) {
	order := binary.LittleEndian
	reqBody := make([]byte, 4)
	order.PutUint32(reqBody[0:4], 123)

	p, err := parseDestroyWindowRequest(order, reqBody)
	assert.NoError(t, err, "parseDestroyWindowRequest should not return an error")
	assert.Equal(t, Window(123), p.Window, "Window should be parsed correctly")
}

func TestParseDestroySubwindowsRequest(t *testing.T) {
	order := binary.LittleEndian
	reqBody := make([]byte, 4)
	order.PutUint32(reqBody[0:4], 123)

	p, err := parseDestroySubwindowsRequest(order, reqBody)
	assert.NoError(t, err, "parseDestroySubwindowsRequest should not return an error")
	assert.Equal(t, Window(123), p.Window, "Window should be parsed correctly")
}

func TestParseChangeSaveSetRequest(t *testing.T) {
	order := binary.LittleEndian
	reqBody := make([]byte, 5)
	order.PutUint32(reqBody[0:4], 123)
	reqBody[4] = 1

	p, err := parseChangeSaveSetRequest(order, reqBody)
	assert.NoError(t, err, "parseChangeSaveSetRequest should not return an error")
	assert.Equal(t, Window(123), p.Window, "Window should be parsed correctly")
	assert.Equal(t, byte(1), p.Mode, "Mode should be parsed correctly")
}

func TestParseReparentWindowRequest(t *testing.T) {
	order := binary.LittleEndian
	reqBody := make([]byte, 12)
	order.PutUint32(reqBody[0:4], 123)
	order.PutUint32(reqBody[4:8], 456)
	order.PutUint16(reqBody[8:10], 10)
	order.PutUint16(reqBody[10:12], 20)

	p, err := parseReparentWindowRequest(order, reqBody)
	assert.NoError(t, err, "parseReparentWindowRequest should not return an error")
	assert.Equal(t, Window(123), p.Window, "Window should be parsed correctly")
	assert.Equal(t, Window(456), p.Parent, "Parent should be parsed correctly")
	assert.Equal(t, int16(10), p.X, "X should be parsed correctly")
	assert.Equal(t, int16(20), p.Y, "Y should be parsed correctly")
}

func TestParseCirculateWindowRequest(t *testing.T) {
	order := binary.LittleEndian
	reqBody := make([]byte, 4)
	order.PutUint32(reqBody[0:4], 123)

	p, err := parseCirculateWindowRequest(order, 1, reqBody)
	assert.NoError(t, err, "parseCirculateWindowRequest should not return an error")
	assert.Equal(t, Window(123), p.Window, "Window should be parsed correctly")
	assert.Equal(t, byte(1), p.Direction, "Direction should be parsed correctly")
}

func TestParseQueryTreeRequest(t *testing.T) {
	order := binary.LittleEndian
	reqBody := make([]byte, 4)
	order.PutUint32(reqBody[0:4], 123)

	p, err := parseQueryTreeRequest(order, reqBody)
	assert.NoError(t, err, "parseQueryTreeRequest should not return an error")
	assert.Equal(t, Window(123), p.Window, "Window should be parsed correctly")
}

func TestParseUnmapWindowRequest(t *testing.T) {
	order := binary.LittleEndian
	reqBody := make([]byte, 4)
	order.PutUint32(reqBody[0:4], 123)

	p, err := parseUnmapWindowRequest(order, reqBody)
	assert.NoError(t, err, "parseUnmapWindowRequest should not return an error")
	assert.Equal(t, Window(123), p.Window, "Window should be parsed correctly")
}

func TestParseUnmapSubwindowsRequest(t *testing.T) {
	order := binary.LittleEndian
	reqBody := make([]byte, 4)
	order.PutUint32(reqBody[0:4], 123)

	p, err := parseUnmapSubwindowsRequest(order, reqBody)
	assert.NoError(t, err, "parseUnmapSubwindowsRequest should not return an error")
	assert.Equal(t, Window(123), p.Window, "Window should be parsed correctly")
}

func TestParseGetGeometryRequest(t *testing.T) {
	order := binary.LittleEndian
	reqBody := make([]byte, 4)
	order.PutUint32(reqBody[0:4], 123)

	p, err := parseGetGeometryRequest(order, reqBody)
	assert.NoError(t, err, "parseGetGeometryRequest should not return an error")
	assert.Equal(t, Drawable(123), p.Drawable, "Drawable should be parsed correctly")
}

func TestParseDeletePropertyRequest(t *testing.T) {
	order := binary.LittleEndian
	reqBody := make([]byte, 8)
	order.PutUint32(reqBody[0:4], 123)
	order.PutUint32(reqBody[4:8], 456)

	p, err := parseDeletePropertyRequest(order, reqBody)
	assert.NoError(t, err, "parseDeletePropertyRequest should not return an error")
	assert.Equal(t, Window(123), p.Window, "Window should be parsed correctly")
	assert.Equal(t, Atom(456), p.Property, "Property should be parsed correctly")
}

func TestParseSetSelectionOwnerRequest(t *testing.T) {
	order := binary.LittleEndian
	reqBody := make([]byte, 12)
	order.PutUint32(reqBody[0:4], 123)
	order.PutUint32(reqBody[4:8], 456)
	order.PutUint32(reqBody[8:12], 789)

	p, err := parseSetSelectionOwnerRequest(order, reqBody)
	assert.NoError(t, err, "parseSetSelectionOwnerRequest should not return an error")
	assert.Equal(t, Window(123), p.Owner, "Owner should be parsed correctly")
	assert.Equal(t, Atom(456), p.Selection, "Selection should be parsed correctly")
	assert.Equal(t, Timestamp(789), p.Time, "Time should be parsed correctly")
}

func TestParseGetSelectionOwnerRequest(t *testing.T) {
	order := binary.LittleEndian
	reqBody := make([]byte, 4)
	order.PutUint32(reqBody[0:4], 123)

	p, err := parseGetSelectionOwnerRequest(order, reqBody)
	assert.NoError(t, err, "parseGetSelectionOwnerRequest should not return an error")
	assert.Equal(t, Atom(123), p.Selection, "Selection should be parsed correctly")
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
	assert.Equal(t, Window(1), p.Requestor, "Requestor should be parsed correctly")
	assert.Equal(t, Atom(2), p.Selection, "Selection should be parsed correctly")
	assert.Equal(t, Atom(3), p.Target, "Target should be parsed correctly")
	assert.Equal(t, Atom(4), p.Property, "Property should be parsed correctly")
	assert.Equal(t, Timestamp(5), p.Time, "Time should be parsed correctly")
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
	assert.Equal(t, Window(123), p.Destination, "Destination should be parsed correctly")
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
	assert.Equal(t, Window(123), p.GrabWindow, "GrabWindow should be parsed correctly")
	assert.Equal(t, uint16(456), p.EventMask, "EventMask should be parsed correctly")
	assert.Equal(t, byte(1), p.PointerMode, "PointerMode should be parsed correctly")
	assert.Equal(t, byte(2), p.KeyboardMode, "KeyboardMode should be parsed correctly")
	assert.Equal(t, Window(789), p.ConfineTo, "ConfineTo should be parsed correctly")
	assert.Equal(t, Cursor(101), p.Cursor, "Cursor should be parsed correctly")
	assert.Equal(t, Timestamp(112), p.Time, "Time should be parsed correctly")
}

func TestParseUngrabPointerRequest(t *testing.T) {
	order := binary.LittleEndian
	reqBody := make([]byte, 4)
	order.PutUint32(reqBody[0:4], 123)

	p, err := parseUngrabPointerRequest(order, reqBody)
	assert.NoError(t, err, "parseUngrabPointerRequest should not return an error")
	assert.Equal(t, Timestamp(123), p.Time, "Time should be parsed correctly")
}

func TestParseGrabButtonRequest(t *testing.T) {
	order := binary.LittleEndian
	reqBody := make([]byte, 24)
	reqBody[0] = 1
	order.PutUint32(reqBody[4:8], 123)
	order.PutUint16(reqBody[8:10], 456)
	reqBody[10] = 1
	reqBody[11] = 2
	order.PutUint32(reqBody[12:16], 789)
	order.PutUint32(reqBody[16:20], 101)
	reqBody[20] = 3
	order.PutUint16(reqBody[22:24], 112)

	p, err := parseGrabButtonRequest(order, reqBody)
	assert.NoError(t, err, "parseGrabButtonRequest should not return an error")
	assert.True(t, p.OwnerEvents, "OwnerEvents should be true")
	assert.Equal(t, Window(123), p.GrabWindow, "GrabWindow should be parsed correctly")
	assert.Equal(t, uint16(456), p.EventMask, "EventMask should be parsed correctly")
	assert.Equal(t, byte(1), p.PointerMode, "PointerMode should be parsed correctly")
	assert.Equal(t, byte(2), p.KeyboardMode, "KeyboardMode should be parsed correctly")
	assert.Equal(t, Window(789), p.ConfineTo, "ConfineTo should be parsed correctly")
	assert.Equal(t, Cursor(101), p.Cursor, "Cursor should be parsed correctly")
	assert.Equal(t, byte(3), p.Button, "Button should be parsed correctly")
	assert.Equal(t, uint16(112), p.Modifiers, "Modifiers should be parsed correctly")
}

func TestParseUngrabButtonRequest(t *testing.T) {
	order := binary.LittleEndian
	reqBody := make([]byte, 8)
	order.PutUint32(reqBody[0:4], 123)
	reqBody[4] = 3
	order.PutUint16(reqBody[6:8], 112)

	p, err := parseUngrabButtonRequest(order, reqBody)
	assert.NoError(t, err, "parseUngrabButtonRequest should not return an error")
	assert.Equal(t, Window(123), p.GrabWindow, "GrabWindow should be parsed correctly")
	assert.Equal(t, byte(3), p.Button, "Button should be parsed correctly")
	assert.Equal(t, uint16(112), p.Modifiers, "Modifiers should be parsed correctly")
}

func TestParseChangeActivePointerGrabRequest(t *testing.T) {
	order := binary.LittleEndian
	reqBody := make([]byte, 12)
	order.PutUint32(reqBody[0:4], 123)
	order.PutUint32(reqBody[4:8], 456)
	order.PutUint16(reqBody[8:10], 789)

	p, err := parseChangeActivePointerGrabRequest(order, reqBody)
	assert.NoError(t, err, "parseChangeActivePointerGrabRequest should not return an error")
	assert.Equal(t, Cursor(123), p.Cursor, "Cursor should be parsed correctly")
	assert.Equal(t, Timestamp(456), p.Time, "Time should be parsed correctly")
	assert.Equal(t, uint16(789), p.EventMask, "EventMask should be parsed correctly")
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
	assert.Equal(t, Window(123), p.GrabWindow, "GrabWindow should be parsed correctly")
	assert.Equal(t, Timestamp(456), p.Time, "Time should be parsed correctly")
	assert.Equal(t, byte(1), p.PointerMode, "PointerMode should be parsed correctly")
	assert.Equal(t, byte(2), p.KeyboardMode, "KeyboardMode should be parsed correctly")
}

func TestParseUngrabKeyboardRequest(t *testing.T) {
	order := binary.LittleEndian
	reqBody := make([]byte, 4)
	order.PutUint32(reqBody[0:4], 123)

	p, err := parseUngrabKeyboardRequest(order, reqBody)
	assert.NoError(t, err, "parseUngrabKeyboardRequest should not return an error")
	assert.Equal(t, Timestamp(123), p.Time, "Time should be parsed correctly")
}

func TestParseGrabKeyRequest(t *testing.T) {
	order := binary.LittleEndian
	reqBody := make([]byte, 13)
	reqBody[0] = 1
	order.PutUint32(reqBody[4:8], 123)
	order.PutUint16(reqBody[8:10], 456)
	reqBody[10] = 7
	reqBody[11] = 1
	reqBody[12] = 2

	p, err := parseGrabKeyRequest(order, reqBody)
	assert.NoError(t, err, "parseGrabKeyRequest should not return an error")
	assert.True(t, p.OwnerEvents, "OwnerEvents should be true")
	assert.Equal(t, Window(123), p.GrabWindow, "GrabWindow should be parsed correctly")
	assert.Equal(t, uint16(456), p.Modifiers, "Modifiers should be parsed correctly")
	assert.Equal(t, KeyCode(7), p.Key, "Key should be parsed correctly")
	assert.Equal(t, byte(1), p.PointerMode, "PointerMode should be parsed correctly")
	assert.Equal(t, byte(2), p.KeyboardMode, "KeyboardMode should be parsed correctly")
}

func TestParseUngrabKeyRequest(t *testing.T) {
	order := binary.LittleEndian
	reqBody := make([]byte, 8)
	order.PutUint32(reqBody[0:4], 123)
	order.PutUint16(reqBody[4:6], 456)
	reqBody[6] = 7

	p, err := parseUngrabKeyRequest(order, reqBody)
	assert.NoError(t, err, "parseUngrabKeyRequest should not return an error")
	assert.Equal(t, Window(123), p.GrabWindow, "GrabWindow should be parsed correctly")
	assert.Equal(t, uint16(456), p.Modifiers, "Modifiers should be parsed correctly")
	assert.Equal(t, KeyCode(7), p.Key, "Key should be parsed correctly")
}

func TestParseAllowEventsRequest(t *testing.T) {
	order := binary.LittleEndian
	reqBody := make([]byte, 4)
	order.PutUint32(reqBody[0:4], 123)

	p, err := parseAllowEventsRequest(order, 5, reqBody)
	assert.NoError(t, err, "parseAllowEventsRequest should not return an error")
	assert.Equal(t, byte(5), p.Mode, "Mode should be parsed correctly")
	assert.Equal(t, Timestamp(123), p.Time, "Time should be parsed correctly")
}

func TestParseGrabServerRequest(t *testing.T) {
	order := binary.LittleEndian
	reqBody := make([]byte, 0)

	_, err := parseGrabServerRequest(order, reqBody)
	assert.NoError(t, err, "parseGrabServerRequest should not return an error")
}

func TestParseUngrabServerRequest(t *testing.T) {
	order := binary.LittleEndian
	reqBody := make([]byte, 0)

	_, err := parseUngrabServerRequest(order, reqBody)
	assert.NoError(t, err, "parseUngrabServerRequest should not return an error")
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
	assert.Equal(t, Window(1), p.SrcWindow, "SrcWindow should be parsed correctly")
	assert.Equal(t, Window(2), p.DstWindow, "DstWindow should be parsed correctly")
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
	assert.Equal(t, Window(123), p.Focus, "Focus should be parsed correctly")
	assert.Equal(t, byte(2), p.RevertTo, "RevertTo should be parsed correctly")
	assert.Equal(t, Timestamp(456), p.Time, "Time should be parsed correctly")
}

func TestParseQueryKeymapRequest(t *testing.T) {
	order := binary.LittleEndian
	reqBody := make([]byte, 0)

	_, err := parseQueryKeymapRequest(order, reqBody)
	assert.NoError(t, err, "parseQueryKeymapRequest should not return an error")
}

func TestParseCloseFontRequest(t *testing.T) {
	order := binary.LittleEndian
	reqBody := make([]byte, 4)
	order.PutUint32(reqBody[0:4], 123)

	p, err := parseCloseFontRequest(order, reqBody)
	assert.NoError(t, err, "parseCloseFontRequest should not return an error")
	assert.Equal(t, Font(123), p.Fid, "Fid should be parsed correctly")
}

func TestParseQueryTextExtentsRequest(t *testing.T) {
	order := binary.LittleEndian
	reqBody := make([]byte, 8)
	order.PutUint32(reqBody[0:4], 123)
	order.PutUint16(reqBody[4:6], 0x0048)
	order.PutUint16(reqBody[6:8], 0x0065)

	p, err := parseQueryTextExtentsRequest(order, reqBody)
	assert.NoError(t, err, "parseQueryTextExtentsRequest should not return an error")
	assert.Equal(t, Font(123), p.Fid, "Fid should be parsed correctly")
	assert.Equal(t, []uint16{0x0048, 0x0065}, p.Text, "Text should be parsed correctly")
}

func TestParseListFontsWithInfoRequest(t *testing.T) {
	order := binary.LittleEndian
	reqBody := make([]byte, 8)
	order.PutUint16(reqBody[0:2], 10)
	order.PutUint16(reqBody[2:4], 4)
	copy(reqBody[4:8], []byte("test"))

	p, err := parseListFontsWithInfoRequest(order, reqBody)
	assert.NoError(t, err, "parseListFontsWithInfoRequest should not return an error")
	assert.Equal(t, uint16(10), p.MaxNames, "MaxNames should be parsed correctly")
	assert.Equal(t, "test", p.Pattern, "Pattern should be parsed correctly")
}

func TestParseSetFontPathRequest(t *testing.T) {
	order := binary.LittleEndian
	reqBody := make([]byte, 4) // 2 for num paths, 2 unused
	order.PutUint16(reqBody[0:2], 2)

	// Add two paths
	path1 := "path1"
	path2 := "path2"
	reqBody = append(reqBody, byte(len(path1)))
	reqBody = append(reqBody, []byte(path1)...)
	reqBody = append(reqBody, byte(len(path2)))
	reqBody = append(reqBody, []byte(path2)...)

	p, err := parseSetFontPathRequest(order, reqBody)
	assert.NoError(t, err, "parseSetFontPathRequest should not return an error")
	assert.Equal(t, uint16(2), p.NumPaths, "NumPaths should be parsed correctly")
	assert.Equal(t, []string{path1, path2}, p.Paths, "Paths should be parsed correctly")
}

func TestParseGetFontPathRequest(t *testing.T) {
	order := binary.LittleEndian
	reqBody := make([]byte, 0)

	_, err := parseGetFontPathRequest(order, reqBody)
	assert.NoError(t, err, "parseGetFontPathRequest should not return an error")
}

func TestParseFreePixmapRequest(t *testing.T) {
	order := binary.LittleEndian
	reqBody := make([]byte, 4)
	order.PutUint32(reqBody[0:4], 123)

	p, err := parseFreePixmapRequest(order, reqBody)
	assert.NoError(t, err, "parseFreePixmapRequest should not return an error")
	assert.Equal(t, Pixmap(123), p.Pid, "Pid should be parsed correctly")
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
	assert.Equal(t, GContext(123), p.Gc, "Gc should be parsed correctly")
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
	assert.Equal(t, GContext(123), p.SrcGC, "SrcGC should be parsed correctly")
	assert.Equal(t, GContext(456), p.DstGC, "DstGC should be parsed correctly")
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
	assert.Equal(t, Window(123), p.Window, "Window should be parsed correctly")
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
	assert.Equal(t, Drawable(123), p.Drawable, "Drawable should be parsed correctly")
	assert.Equal(t, GContext(456), p.Gc, "Gc should be parsed correctly")
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
	assert.Equal(t, Drawable(123), p.Drawable, "Drawable should be parsed correctly")
	assert.Equal(t, GContext(456), p.Gc, "Gc should be parsed correctly")
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
	assert.Equal(t, Drawable(123), p.Drawable, "Drawable should be parsed correctly")
	assert.Equal(t, GContext(456), p.Gc, "Gc should be parsed correctly")
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
	assert.Equal(t, Colormap(123), p.Mid, "Mid should be parsed correctly")
	assert.Equal(t, Window(456), p.Window, "Window should be parsed correctly")
	assert.Equal(t, VisualID(789), p.Visual, "Visual should be parsed correctly")
}

func TestParseFreeColormapRequest(t *testing.T) {
	order := binary.LittleEndian
	reqBody := make([]byte, 4)
	order.PutUint32(reqBody[0:4], 123)

	p, err := parseFreeColormapRequest(order, reqBody)
	assert.NoError(t, err, "parseFreeColormapRequest should not return an error")
	assert.Equal(t, Colormap(123), p.Cmap, "Cmap should be parsed correctly")
}

func TestParseInstallColormapRequest(t *testing.T) {
	order := binary.LittleEndian
	reqBody := make([]byte, 4)
	order.PutUint32(reqBody[0:4], 123)

	p, err := parseInstallColormapRequest(order, reqBody)
	assert.NoError(t, err, "parseInstallColormapRequest should not return an error")
	assert.Equal(t, Colormap(123), p.Cmap, "Cmap should be parsed correctly")
}

func TestParseUninstallColormapRequest(t *testing.T) {
	order := binary.LittleEndian
	reqBody := make([]byte, 4)
	order.PutUint32(reqBody[0:4], 123)

	p, err := parseUninstallColormapRequest(order, reqBody)
	assert.NoError(t, err, "parseUninstallColormapRequest should not return an error")
	assert.Equal(t, Colormap(123), p.Cmap, "Cmap should be parsed correctly")
}

func TestParseListInstalledColormapsRequest(t *testing.T) {
	order := binary.LittleEndian
	reqBody := make([]byte, 4)
	order.PutUint32(reqBody[0:4], 123)

	p, err := parseListInstalledColormapsRequest(order, reqBody)
	assert.NoError(t, err, "parseListInstalledColormapsRequest should not return an error")
	assert.Equal(t, Window(123), p.Window, "Window should be parsed correctly")
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
	assert.Equal(t, Colormap(123), p.Cmap, "Cmap should be parsed correctly")
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
	assert.Equal(t, Colormap(123), p.Cmap, "Cmap should be parsed correctly")
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
	assert.Equal(t, Colormap(123), p.Cmap, "Cmap should be parsed correctly")
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
	assert.Equal(t, Colormap(123), p.Cmap, "Cmap should be parsed correctly")
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
	assert.Equal(t, Colormap(123), p.Cmap, "Cmap should be parsed correctly")
	assert.Equal(t, "blue", p.Name, "Name should be parsed correctly")
}

func TestParseFreeCursorRequest(t *testing.T) {
	order := binary.LittleEndian
	reqBody := make([]byte, 4)
	order.PutUint32(reqBody[0:4], 123)

	p, err := parseFreeCursorRequest(order, reqBody)
	assert.NoError(t, err, "parseFreeCursorRequest should not return an error")
	assert.Equal(t, Cursor(123), p.Cursor, "Cursor should be parsed correctly")
}

func TestParseQueryBestSizeRequest(t *testing.T) {
	order := binary.LittleEndian
	reqBody := make([]byte, 8)
	order.PutUint32(reqBody[0:4], 123)
	order.PutUint16(reqBody[4:6], 100)
	order.PutUint16(reqBody[6:8], 200)

	p, err := parseQueryBestSizeRequest(order, reqBody)
	assert.NoError(t, err, "parseQueryBestSizeRequest should not return an error")
	assert.Equal(t, Drawable(123), p.Drawable, "Drawable should be parsed correctly")
	assert.Equal(t, uint16(100), p.Width, "Width should be parsed correctly")
	assert.Equal(t, uint16(200), p.Height, "Height should be parsed correctly")
}

func TestParseBellRequest(t *testing.T) {
	p, err := parseBellRequest(50)
	assert.NoError(t, err, "parseBellRequest should not return an error")
	assert.Equal(t, int8(50), p.Percent, "Percent should be parsed correctly")
}

func TestParseSetDashesRequest(t *testing.T) {
	order := binary.LittleEndian
	reqBody := make([]byte, 10)
	order.PutUint32(reqBody[0:4], 123)
	order.PutUint16(reqBody[4:6], 456)
	order.PutUint16(reqBody[6:8], 2)
	reqBody[8] = 10
	reqBody[9] = 20

	p, err := parseSetDashesRequest(order, reqBody)
	assert.NoError(t, err, "parseSetDashesRequest should not return an error")
	assert.Equal(t, GContext(123), p.GC, "GC should be parsed correctly")
	assert.Equal(t, uint16(456), p.DashOffset, "DashOffset should be parsed correctly")
	assert.Equal(t, []byte{10, 20}, p.Dashes, "Dashes should be parsed correctly")
}

func TestParseSetClipRectanglesRequest(t *testing.T) {
	order := binary.LittleEndian
	reqBody := make([]byte, 24)
	order.PutUint32(reqBody[0:4], 123)
	order.PutUint16(reqBody[4:6], 10)
	order.PutUint16(reqBody[6:8], 20)
	// Rectangle 1
	order.PutUint16(reqBody[8:10], 1)
	order.PutUint16(reqBody[10:12], 2)
	order.PutUint16(reqBody[12:14], 3)
	order.PutUint16(reqBody[14:16], 4)
	// Rectangle 2
	order.PutUint16(reqBody[16:18], 5)
	order.PutUint16(reqBody[18:20], 6)
	order.PutUint16(reqBody[20:22], 7)
	order.PutUint16(reqBody[22:24], 8)

	p, err := parseSetClipRectanglesRequest(order, 1, reqBody)
	assert.NoError(t, err, "parseSetClipRectanglesRequest should not return an error")
	assert.Equal(t, GContext(123), p.GC, "GC should be parsed correctly")
	assert.Equal(t, int16(10), p.ClippingX, "ClippingX should be parsed correctly")
	assert.Equal(t, int16(20), p.ClippingY, "ClippingY should be parsed correctly")
	assert.Equal(t, byte(1), p.Ordering, "Ordering should be parsed correctly")
	assert.Equal(t, []Rectangle{{1, 2, 3, 4}, {5, 6, 7, 8}}, p.Rectangles, "Rectangles should be parsed correctly")
}

func TestParseRecolorCursorRequest(t *testing.T) {
	order := binary.LittleEndian
	reqBody := make([]byte, 16)
	order.PutUint32(reqBody[0:4], 123)
	order.PutUint16(reqBody[4:6], 10)
	order.PutUint16(reqBody[6:8], 20)
	order.PutUint16(reqBody[8:10], 30)
	order.PutUint16(reqBody[10:12], 40)
	order.PutUint16(reqBody[12:14], 50)
	order.PutUint16(reqBody[14:16], 60)

	p, err := parseRecolorCursorRequest(order, reqBody)
	assert.NoError(t, err, "parseRecolorCursorRequest should not return an error")
	assert.Equal(t, Cursor(123), p.Cursor, "Cursor should be parsed correctly")
	assert.Equal(t, [3]uint16{10, 20, 30}, p.ForeColor, "ForeColor should be parsed correctly")
	assert.Equal(t, [3]uint16{40, 50, 60}, p.BackColor, "BackColor should be parsed correctly")
}

func TestParseSetPointerMappingRequest(t *testing.T) {
	order := binary.LittleEndian
	reqBody := []byte{1, 2, 3, 4}

	p, err := parseSetPointerMappingRequest(order, reqBody)
	assert.NoError(t, err, "parseSetPointerMappingRequest should not return an error")
	assert.Equal(t, []byte{1, 2, 3, 4}, p.Map, "Map should be parsed correctly")
}

func TestParseGetKeyboardMappingRequest(t *testing.T) {
	order := binary.LittleEndian
	reqBody := []byte{10, 5}

	p, err := parseGetKeyboardMappingRequest(order, reqBody)
	assert.NoError(t, err, "parseGetKeyboardMappingRequest should not return an error")
	assert.Equal(t, KeyCode(10), p.FirstKeyCode, "FirstKeyCode should be parsed correctly")
	assert.Equal(t, byte(5), p.Count, "Count should be parsed correctly")
}

func TestParseChangeKeyboardMappingRequest(t *testing.T) {
	order := binary.LittleEndian
	reqBody := make([]byte, 20)
	reqBody[0] = 10
	reqBody[1] = 2
	order.PutUint32(reqBody[4:8], 123)
	order.PutUint32(reqBody[8:12], 456)
	order.PutUint32(reqBody[12:16], 789)
	order.PutUint32(reqBody[16:20], 101)

	p, err := parseChangeKeyboardMappingRequest(order, 2, reqBody)
	assert.NoError(t, err, "parseChangeKeyboardMappingRequest should not return an error")
	assert.Equal(t, byte(2), p.KeyCodeCount, "KeyCodeCount should be parsed correctly")
	assert.Equal(t, KeyCode(10), p.FirstKeyCode, "FirstKeyCode should be parsed correctly")
	assert.Equal(t, byte(2), p.KeySymsPerKeyCode, "KeySymsPerKeyCode should be parsed correctly")
	assert.Equal(t, []uint32{123, 456, 789, 101}, p.KeySyms, "KeySyms should be parsed correctly")
}

func TestParseChangeKeyboardControlRequest(t *testing.T) {
	order := binary.LittleEndian
	reqBody := make([]byte, 12)
	order.PutUint32(reqBody[0:4], uint32(KBKeyClickPercent|KBBellPercent))
	order.PutUint32(reqBody[4:8], 50)
	order.PutUint32(reqBody[8:12], 60)

	p, err := parseChangeKeyboardControlRequest(order, reqBody)
	assert.NoError(t, err, "parseChangeKeyboardControlRequest should not return an error")
	assert.Equal(t, uint32(KBKeyClickPercent|KBBellPercent), p.ValueMask, "ValueMask should be parsed correctly")
	assert.Equal(t, int32(50), p.Values.KeyClickPercent, "KeyClickPercent should be parsed correctly")
	assert.Equal(t, int32(60), p.Values.BellPercent, "BellPercent should be parsed correctly")
}

func TestParseSetScreenSaverRequest(t *testing.T) {
	order := binary.LittleEndian
	reqBody := make([]byte, 6)
	order.PutUint16(reqBody[0:2], 10)
	order.PutUint16(reqBody[2:4], 20)
	reqBody[4] = 1
	reqBody[5] = 2

	p, err := parseSetScreenSaverRequest(order, reqBody)
	assert.NoError(t, err, "parseSetScreenSaverRequest should not return an error")
	assert.Equal(t, int16(10), p.Timeout, "Timeout should be parsed correctly")
	assert.Equal(t, int16(20), p.Interval, "Interval should be parsed correctly")
	assert.Equal(t, byte(1), p.PreferBlank, "PreferBlank should be parsed correctly")
	assert.Equal(t, byte(2), p.AllowExpose, "AllowExpose should be parsed correctly")
}

func TestParseChangeHostsRequest(t *testing.T) {
	order := binary.LittleEndian
	reqBody := make([]byte, 8)
	reqBody[0] = 1
	order.PutUint16(reqBody[2:4], 4)
	copy(reqBody[4:8], []byte{1, 2, 3, 4})

	p, err := parseChangeHostsRequest(order, 2, reqBody)
	assert.NoError(t, err, "parseChangeHostsRequest should not return an error")
	assert.Equal(t, byte(2), p.Mode, "Mode should be parsed correctly")
	assert.Equal(t, byte(1), p.Host.Family, "Family should be parsed correctly")
	assert.Equal(t, []byte{1, 2, 3, 4}, p.Host.Data, "Data should be parsed correctly")
}

func TestParseKillClientRequest(t *testing.T) {
	order := binary.LittleEndian
	reqBody := make([]byte, 4)
	order.PutUint32(reqBody[0:4], 123)

	p, err := parseKillClientRequest(order, reqBody)
	assert.NoError(t, err, "parseKillClientRequest should not return an error")
	assert.Equal(t, uint32(123), p.Resource, "Resource should be parsed correctly")
}

func TestParseRotatePropertiesRequest(t *testing.T) {
	order := binary.LittleEndian
	reqBody := make([]byte, 16)
	order.PutUint32(reqBody[0:4], 123)
	order.PutUint16(reqBody[4:6], 2)
	order.PutUint16(reqBody[6:8], 10)
	order.PutUint32(reqBody[8:12], 456)
	order.PutUint32(reqBody[12:16], 789)

	p, err := parseRotatePropertiesRequest(order, reqBody)
	assert.NoError(t, err, "parseRotatePropertiesRequest should not return an error")
	assert.Equal(t, Window(123), p.Window, "Window should be parsed correctly")
	assert.Equal(t, int16(10), p.Delta, "Delta should be parsed correctly")
	assert.Equal(t, []Atom{456, 789}, p.Atoms, "Atoms should be parsed correctly")
}

func TestParseForceScreenSaverRequest(t *testing.T) {
	p, err := parseForceScreenSaverRequest(nil, 1, nil)
	assert.NoError(t, err, "parseForceScreenSaverRequest should not return an error")
	assert.Equal(t, byte(1), p.Mode, "Mode should be parsed correctly")
}

func TestParseSetModifierMappingRequest(t *testing.T) {
	order := binary.LittleEndian
	reqBody := []byte{2, 1, 2, 3, 4}

	p, err := parseSetModifierMappingRequest(order, reqBody)
	assert.NoError(t, err, "parseSetModifierMappingRequest should not return an error")
	assert.Equal(t, byte(2), p.KeyCodesPerModifier, "KeyCodesPerModifier should be parsed correctly")
	assert.Equal(t, []KeyCode{1, 2, 3, 4}, p.KeyCodes, "KeyCodes should be parsed correctly")
}
