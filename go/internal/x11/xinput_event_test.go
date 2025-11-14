//go:build x11 && !wasm

package x11

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSendMouseEvent_XInput_EventSent(t *testing.T) {
	server, client, _, _ := setupTestServerWithClient(t)
	windowID := client.xID(10)

	// 1. Open the virtual pointer device
	openReqBody := []byte{2, 0, 0, 0}
	openReq := &XInputRequest{MinorOpcode: XOpenDevice, Body: openReqBody}
	reply := server.handleRequest(client, openReq, 1)
	require.NotNil(t, reply)

	// 2. Create a window
	createReq := &CreateWindowRequest{
		Drawable: Window(windowID.local),
		Parent:   Window(server.rootWindowID()),
		Width:    100,
		Height:   100,
	}
	server.handleRequest(client, createReq, 2)

	// 3. Send a mouse button press event
	server.SendMouseEvent(windowID, "mousedown", 10, 20, (0<<16)|1)

	// 5. Verify that a DeviceButtonPressEvent was sent
	require.NotEmpty(t, client.sentMessages, "Expected a button press event to be sent")
	pressEvent, ok := client.sentMessages[0].(*DeviceButtonPressEvent)
	require.True(t, ok, "Expected a DeviceButtonPressEvent")
	assert.Equal(t, byte(2), pressEvent.DeviceID)
	assert.Equal(t, uint32(windowID.local), pressEvent.Event)

	// 6. Send a mouse button release event
	server.SendMouseEvent(windowID, "mouseup", 10, 20, (0<<16)|1)

	// 7. Verify that a DeviceButtonReleaseEvent was sent
	require.Len(t, client.sentMessages, 2, "Expected a button release event to be sent")
	releaseEvent, ok := client.sentMessages[1].(*DeviceButtonReleaseEvent)
	require.True(t, ok, "Expected a DeviceButtonReleaseEvent")
	assert.Equal(t, byte(2), releaseEvent.DeviceID)
	assert.Equal(t, uint32(windowID.local), releaseEvent.Event)
}
