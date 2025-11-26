//go:build x11 && !wasm

package x11

import (
	"testing"

	"github.com/c2FmZQ/sshterm/internal/x11/wire"
	"github.com/stretchr/testify/require"
)

func TestQueryBestSize(t *testing.T) {
	s, c, _, _ := setupTestServerWithClient(t)
	createWindow(t, s, c, 1, 0, 0, 0, 10, 10)

	// Test cursor
	reqCursor := &wire.QueryBestSizeRequest{
		Drawable: 1,
		Class:    0, // Cursor
		Width:    16,
		Height:   16,
	}
	replyCursor := s.handleQueryBestSize(c, reqCursor, 0).(*wire.QueryBestSizeReply)
	require.Equal(t, uint16(16), replyCursor.Width)
	require.Equal(t, uint16(16), replyCursor.Height)

	// Test tile
	reqTile := &wire.QueryBestSizeRequest{
		Drawable: 1,
		Class:    1, // Tile
		Width:    100,
		Height:   100,
	}
	replyTile := s.handleQueryBestSize(c, reqTile, 0).(*wire.QueryBestSizeReply)
	require.Equal(t, uint16(100), replyTile.Width)
	require.Equal(t, uint16(100), replyTile.Height)

	// Test stipple
	reqStipple := &wire.QueryBestSizeRequest{
		Drawable: 1,
		Class:    2, // Stipple
		Width:    200,
		Height:   200,
	}
	replyStipple := s.handleQueryBestSize(c, reqStipple, 0).(*wire.QueryBestSizeReply)
	require.Equal(t, uint16(200), replyStipple.Width)
	require.Equal(t, uint16(200), replyStipple.Height)
}

func TestDrawingBatching(t *testing.T) {
	s, c, fe, out := setupTestServerWithClient(t)

	createWindow(t, s, c, 1, 0, 10, 10, 100, 100)
	createGC(t, s, c, 2, 1)

	// Perform a drawing operation
	polyPointReq := &wire.PolyPointRequest{
		Drawable: 1,
		Gc:       2,
		Coordinates: []uint32{
			10, 20,
			30, 40,
		},
	}
	s.handlePolyPoint(c, polyPointReq, 0)
	s.handlePolyPoint(c, polyPointReq, 0)

	// Check that ComposeWindow was not called yet
	require.Equal(t, 0, fe.ComposeWindowCount)

	// Flush the dirty windows
	s.flushDirtyWindows()

	// Check that ComposeWindow was called once
	require.Equal(t, 1, fe.ComposeWindowCount)

	// Clear the buffer for the next check
	out.Reset()
	fe.ComposeWindowCount = 0

	// Perform another drawing operation
	s.handlePolyPoint(c, polyPointReq, 0)
	s.flushDirtyWindows()
	require.Equal(t, 1, fe.ComposeWindowCount)
}

func TestSendEventMask(t *testing.T) {
	s, c, _, out := setupTestServerWithClient(t)
	require.NotNil(t, s.byteOrder, "s.byteOrder should not be nil")

	// Create a window and select for key press events
	createWindow(t, s, c, 1, 0, 0, 0, 100, 100)
	changeWindowAttributes(t, s, c, 1, wire.CWEventMask, &wire.WindowAttributes{
		EventMask: wire.KeyPressMask,
	})

	// Send a KeyPress event
	keyPressEvent := &wire.KeyEvent{
		Opcode:     wire.KeyPress,
		Sequence:   1,
		Detail:     10,
		Time:       12345,
		Root:       s.rootWindowID(),
		Event:      1,
		Child:      0,
		RootX:      10,
		RootY:      10,
		EventX:     10,
		EventY:     10,
		State:      0,
		SameScreen: true,
	}
	eventData := keyPressEvent.EncodeMessage(s.byteOrder)
	sendEventReq := &wire.SendEventRequest{
		Propagate:   false,
		Destination: wire.Window(1),
		EventMask:   wire.KeyPressMask,
		EventData:   eventData,
	}
	s.handleSendEvent(c, sendEventReq, 0)

	// Check that the event was sent
	require.True(t, out.Len() > 0)
	out.Reset()

	// Send a KeyRelease event (not selected)
	keyReleaseEvent := &wire.KeyEvent{
		Opcode:     wire.KeyRelease,
		Sequence:   1,
		Detail:     10,
		Time:       12345,
		Root:       s.rootWindowID(),
		Event:      1,
		Child:      0,
		RootX:      10,
		RootY:      10,
		EventX:     10,
		EventY:     10,
		State:      0,
		SameScreen: true,
	}
	eventData = keyReleaseEvent.EncodeMessage(s.byteOrder)
	sendEventReq.EventData = eventData
	s.handleSendEvent(c, sendEventReq, 0)

	// Check that the event was not sent
	require.Equal(t, 0, out.Len())
}

func TestTranslateCoordsWithChild(t *testing.T) {
	s, c, _, _ := setupTestServerWithClient(t)

	// Create parent and child windows
	createWindow(t, s, c, 1, 0, 10, 10, 200, 200)
	s.windows[c.xID(1)].mapped = true
	createWindow(t, s, c, 2, 1, 50, 50, 100, 100)
	s.windows[c.xID(2)].mapped = true
	s.windows[c.xID(2)].parent = c.xID(1).local
	s.windows[c.xID(1)].children = []uint32{2}

	// Translate coordinates from root to parent, hitting the child
	req := &wire.TranslateCoordsRequest{
		SrcWindow: wire.Window(s.rootWindowID()),
		DstWindow: wire.Window(1),
		SrcX:      65,
		SrcY:      65,
	}
	reply := s.handleTranslateCoords(c, req, 0).(*wire.TranslateCoordsReply)

	require.Equal(t, int16(55), reply.DstX)
	require.Equal(t, int16(55), reply.DstY)
	require.Equal(t, uint32(2), reply.Child)
}

// Helper functions to create requests
func createWindow(t *testing.T, s *x11Server, c *x11Client, id, parent uint32, x, y, width, height int16) {
	req := &wire.CreateWindowRequest{
		Drawable: wire.Window(id),
		Parent:   wire.Window(parent),
		X:        x,
		Y:        y,
		Width:    uint16(width),
		Height:   uint16(height),
		Depth:    24,
	}
	s.handleCreateWindow(c, req, 0)
}

func createGC(t *testing.T, s *x11Server, c *x11Client, id, drawable uint32) {
	req := &wire.CreateGCRequest{
		Cid:      wire.GContext(id),
		Drawable: wire.Drawable(drawable),
	}
	s.handleCreateGC(c, req, 0)
}

func changeWindowAttributes(t *testing.T, s *x11Server, c *x11Client, id uint32, mask uint32, values *wire.WindowAttributes) {
	req := &wire.ChangeWindowAttributesRequest{
		Window:    wire.Window(id),
		ValueMask: uint32(mask),
		Values:    *values,
	}
	s.handleChangeWindowAttributes(c, req, 0)
}
