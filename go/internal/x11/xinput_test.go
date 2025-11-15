//go:build x11 && !wasm

package x11

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestXListInputDevices(t *testing.T) {
	s, client, _, _ := setupTestServerWithClient(t)

	req := &XInputRequest{MinorOpcode: XListInputDevices}
	reply := s.handleRequest(client, req, 1)
	require.NotNil(t, reply)

	listReply, ok := reply.(*ListInputDevicesReply)
	require.True(t, ok)
	assert.Equal(t, 2, len(listReply.devices))
}

func TestXOpenDevice(t *testing.T) {
	s, client, _, _ := setupTestServerWithClient(t)

	openReqBody := []byte{2, 0, 0, 0}
	req := &XInputRequest{MinorOpcode: XOpenDevice, Body: openReqBody}
	reply := s.handleRequest(client, req, 1)
	require.NotNil(t, reply)

	_, ok := client.openDevices[2]
	assert.True(t, ok, "Device 2 should be open")
}

func TestXCloseDevice(t *testing.T) {
	s, client, _, _ := setupTestServerWithClient(t)

	// First, open a device
	openReqBody := []byte{3, 0, 0, 0}
	openReq := &XInputRequest{MinorOpcode: XOpenDevice, Body: openReqBody}
	s.handleRequest(client, openReq, 1)

	_, ok := client.openDevices[3]
	require.True(t, ok, "Device 3 should be open before closing")

	// Now, close the device
	closeReqBody := []byte{3, 0, 0, 0}
	closeReq := &XInputRequest{MinorOpcode: XCloseDevice, Body: closeReqBody}
	reply := s.handleRequest(client, closeReq, 2)
	require.NotNil(t, reply)

	_, ok = client.openDevices[3]
	assert.False(t, ok, "Device 3 should be closed")
}

func TestXGrabDevice(t *testing.T) {
	s, client, _, _ := setupTestServerWithClient(t)
	// Create a window to grab
	createWindowReq := &CreateWindowRequest{
		Drawable: 1,
		Parent:   0,
		Width:    100,
		Height:   100,
	}
	s.handleRequest(client, createWindowReq, 1)

	// Open the virtual pointer device
	openReqBody := []byte{2, 0, 0, 0}
	openReq := &XInputRequest{MinorOpcode: XOpenDevice, Body: openReqBody}
	s.handleRequest(client, openReq, 2)

	// Grab the device
	grabReqBody := make([]byte, 20)
	client.byteOrder.PutUint32(grabReqBody[0:4], 1)   // grab_window
	grabReqBody[11] = 2                               // device_id
	client.byteOrder.PutUint16(grabReqBody[14:16], 1) // num_classes
	client.byteOrder.PutUint32(grabReqBody[16:20], 0) // classes

	grabReq := &XInputRequest{MinorOpcode: XGrabDevice, Body: grabReqBody}
	reply := s.handleRequest(client, grabReq, 3)
	require.NotNil(t, reply)

	grabReply, ok := reply.(*GrabDeviceReply)
	require.True(t, ok)
	assert.Equal(t, byte(GrabSuccess), grabReply.Status)

	_, grabExists := s.deviceGrabs[2]
	assert.True(t, grabExists, "Device 2 should be grabbed")

	// Attempt to grab again
	reply = s.handleRequest(client, grabReq, 4)
	require.NotNil(t, reply)

	grabReply, ok = reply.(*GrabDeviceReply)
	require.True(t, ok)
	assert.Equal(t, byte(AlreadyGrabbed), grabReply.Status)

	// Ungrab the device
	ungrabReqBody := make([]byte, 8)
	client.byteOrder.PutUint32(ungrabReqBody[0:4], 0) // time
	ungrabReqBody[5] = 2                              // device_id
	ungrabReq := &XInputRequest{MinorOpcode: XUngrabDevice, Body: ungrabReqBody}
	s.handleRequest(client, ungrabReq, 5)

	_, grabExists = s.deviceGrabs[2]
	assert.False(t, grabExists, "Device 2 should be ungrabbed")
}

func TestXGrabDeviceEventRedirection(t *testing.T) {
	s, client1, _, client1Buffer := setupTestServerWithClient(t)

	// Create a window
	createWindowReq := &CreateWindowRequest{
		Drawable: 1,
		Parent:   0,
		Width:    100,
		Height:   100,
	}
	s.handleRequest(client1, createWindowReq, 1)

	// Create a second client
	client2 := &x11Client{
		id:          s.nextClientID,
		sequence:    0,
		byteOrder:   s.byteOrder,
		saveSet:     make(map[uint32]bool),
		openDevices: make(map[byte]*deviceInfo),
	}
	s.clients[client2.id] = client2
	s.nextClientID++
	client2.openDevices = make(map[byte]*deviceInfo)

	// Client 1 opens the pointer
	openReqBody := []byte{2, 0, 0, 0} // device 2 (pointer)
	openReq := &XInputRequest{MinorOpcode: XOpenDevice, Body: openReqBody}
	s.handleRequest(client1, openReq, 2)

	// Client 2 opens the pointer and selects for button press events
	s.handleRequest(client2, openReq, 1)

	selectEventReqBody := make([]byte, 12)
	client2.byteOrder.PutUint32(selectEventReqBody[0:4], 1) // window
	client2.byteOrder.PutUint16(selectEventReqBody[4:6], 1) // num_events
	mask := uint32(DeviceButtonPressMask) << 8
	mask |= uint32(2)
	client2.byteOrder.PutUint32(selectEventReqBody[8:12], mask) // event_mask | device_id
	selectEventReq := &XInputRequest{MinorOpcode: XSelectExtensionEvent, Body: selectEventReqBody}
	s.handleRequest(client2, selectEventReq, 2)

	// Client 2 grabs the device
	grabReqBody := make([]byte, 20)
	client2.byteOrder.PutUint32(grabReqBody[0:4], 1)   // grab_window
	grabReqBody[11] = 2                                // device_id
	client2.byteOrder.PutUint16(grabReqBody[14:16], 1) // num_classes
	client2.byteOrder.PutUint32(grabReqBody[16:20], DeviceButtonPressMask)

	grabReq := &XInputRequest{MinorOpcode: XGrabDevice, Body: grabReqBody}
	s.handleRequest(client2, grabReq, 3)

	// Send a mouse event
	s.SendMouseEvent(client1.xID(1), "mousedown", 10, 10, (0<<16)|1) // button 1

	// Check that only client 2 received the event
	assert.Zero(t, client1Buffer.Len(), "Client 1 should not have received any messages")

	require.Len(t, client2.sentMessages, 1, "Client 2 should have received a message")
	buttonPressEvent, ok := client2.sentMessages[0].(*DeviceButtonPressEvent)
	require.True(t, ok, "Client 2 should have received a *DeviceButtonPressEvent")

	// Basic validation of the event
	assert.Equal(t, byte(1), buttonPressEvent.Button, "Button should be 1")
	assert.Equal(t, byte(2), buttonPressEvent.DeviceID, "Device ID should be 2")
}
