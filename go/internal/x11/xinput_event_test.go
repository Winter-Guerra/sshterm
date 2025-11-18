//go:build x11 && !wasm

package x11

import (
	"encoding/binary"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/c2FmZQ/sshterm/internal/x11/wire"
)

func TestSendMouseEvent_XInput_EventSent(t *testing.T) {
	server, client, _, clientBuffer := setupTestServerWithClient(t)
	windowID := client.xID(10)

	// 1. Open the virtual pointer device
	openReqBody := []byte{2, 0, 0, 0}
	openReq := &wire.XInputRequest{MinorOpcode: wire.XOpenDevice, Body: openReqBody}
	reply := server.handleRequest(client, openReq, 1)
	require.NotNil(t, reply)

	// 2. Create a window
	createReq := &wire.CreateWindowRequest{
		Drawable: wire.Window(windowID.local),
		Parent:   wire.Window(server.rootWindowID()),
		Width:    100,
		Height:   100,
	}
	server.handleRequest(client, createReq, 2)

	// 3. Select for XInput events on the window
	selectReqBody := make([]byte, 12)
	// window ID
	binary.LittleEndian.PutUint32(selectReqBody[0:4], windowID.local)
	// num_classes
	binary.LittleEndian.PutUint16(selectReqBody[4:6], 1)
	// class
	mask := uint32(wire.DeviceButtonPressMask | wire.DeviceButtonReleaseMask)
	deviceID := byte(2) // Virtual pointer
	class := (mask << 8) | uint32(deviceID)
	binary.LittleEndian.PutUint32(selectReqBody[8:12], class)

	selectReq := &wire.XInputRequest{MinorOpcode: wire.XSelectExtensionEvent, Body: selectReqBody}
	server.handleRequest(client, selectReq, 3)

	// 4. Send a mouse button press event
	server.SendMouseEvent(windowID, "mousedown", 10, 20, (0<<16)|1)

	// 5. Verify that a DeviceButtonPressEvent was sent
	messages := drainMessages(t, clientBuffer, client.byteOrder)
	require.NotEmpty(t, messages, "Expected a button press event to be sent")
	pressEvent, ok := messages[0].(*wire.DeviceButtonPressEvent)
	require.True(t, ok, "Expected a DeviceButtonPressEvent")
	assert.Equal(t, deviceID, pressEvent.DeviceID)
	assert.Equal(t, uint32(windowID.local), pressEvent.Event)

	// 6. Send a mouse button release event and verify
	server.SendMouseEvent(windowID, "mouseup", 10, 20, (0<<16)|1)
	messages = drainMessages(t, clientBuffer, client.byteOrder)
	require.Len(t, messages, 1, "Expected a button release event to be sent")
	releaseEvent, ok := messages[0].(*wire.DeviceButtonReleaseEvent)
	require.True(t, ok, "Expected a DeviceButtonReleaseEvent")
	assert.Equal(t, deviceID, releaseEvent.DeviceID)
	assert.Equal(t, uint32(windowID.local), releaseEvent.Event)
}
