//go:build x11 && !wasm

package x11

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestXGetExtensionVersion(t *testing.T) {
	s, client, _, _ := setupTestServerWithClient(t)

	req := &XInputRequest{MinorOpcode: XGetExtensionVersion}
	reply := s.handleRequest(client, req, 1)
	require.NotNil(t, reply)

	versionReply, ok := reply.(*GetExtensionVersionReply)
	require.True(t, ok)
	assert.Equal(t, uint16(1), versionReply.MajorVersion)
	assert.Equal(t, uint16(5), versionReply.MinorVersion)
}

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

func TestXSetDeviceMode(t *testing.T) {
	s, client, _, _ := setupTestServerWithClient(t)

	// Open the virtual pointer device
	openReqBody := []byte{2, 0, 0, 0}
	openReq := &XInputRequest{MinorOpcode: XOpenDevice, Body: openReqBody}
	s.handleRequest(client, openReq, 1)

	device, ok := client.openDevices[2]
	require.True(t, ok)
	valuatorInfo, ok := device.classes[1].(*ValuatorClassInfo)
	require.True(t, ok)
	assert.Equal(t, byte(0), valuatorInfo.Mode) // Initially Relative

	// Set mode to Absolute
	setModeReqBody := []byte{2, 1, 0, 0} // device 2, mode Absolute
	setModeReq := &XInputRequest{MinorOpcode: XSetDeviceMode, Body: setModeReqBody}
	reply := s.handleRequest(client, setModeReq, 2)
	require.NotNil(t, reply)

	setModeReply, ok := reply.(*SetDeviceModeReply)
	require.True(t, ok)
	assert.Equal(t, byte(GrabSuccess), setModeReply.Status)
	assert.Equal(t, byte(1), valuatorInfo.Mode) // Should now be Absolute
}

func TestXSetDeviceValuators(t *testing.T) {
	s, client, _, _ := setupTestServerWithClient(t)

	// Open the virtual pointer device
	openReqBody := []byte{2, 0, 0, 0}
	openReq := &XInputRequest{MinorOpcode: XOpenDevice, Body: openReqBody}
	s.handleRequest(client, openReq, 1)

	// Set valuators
	setValuatorsReqBody := make([]byte, 12)
	setValuatorsReqBody[0] = 2 // device 2
	setValuatorsReqBody[1] = 0 // first_valuator
	setValuatorsReqBody[2] = 2 // num_valuators
	client.byteOrder.PutUint32(setValuatorsReqBody[4:8], 100)
	client.byteOrder.PutUint32(setValuatorsReqBody[8:12], 200)
	setValuatorsReq := &XInputRequest{MinorOpcode: XSetDeviceValuators, Body: setValuatorsReqBody}
	reply := s.handleRequest(client, setValuatorsReq, 2)
	require.NotNil(t, reply)

	setValuatorsReply, ok := reply.(*SetDeviceValuatorsReply)
	require.True(t, ok)
	assert.Equal(t, byte(GrabSuccess), setValuatorsReply.Status)

	device, ok := client.openDevices[2]
	require.True(t, ok)
	valuatorInfo, ok := device.classes[1].(*ValuatorClassInfo)
	require.True(t, ok)
	assert.Equal(t, int32(100), valuatorInfo.Axes[0].Value)
	assert.Equal(t, int32(200), valuatorInfo.Axes[1].Value)
}

func TestXDeviceControl(t *testing.T) {
	s, client, _, _ := setupTestServerWithClient(t)

	// Open the virtual pointer device
	openReqBody := []byte{2, 0, 0, 0}
	openReq := &XInputRequest{MinorOpcode: XOpenDevice, Body: openReqBody}
	s.handleRequest(client, openReq, 1)

	// Get device control (resolution)
	getControlReqBody := []byte{2, 1, 0, 0} // device 2, control DEVICE_RESOLUTION
	getControlReq := &XInputRequest{MinorOpcode: XGetDeviceControl, Body: getControlReqBody}
	reply := s.handleRequest(client, getControlReq, 2)
	require.NotNil(t, reply)
	getControlReply, ok := reply.(*GetDeviceControlReply)
	require.True(t, ok)
	resolutionState, ok := getControlReply.Control.(*DeviceResolutionState)
	require.True(t, ok)
	assert.Equal(t, uint32(1), resolutionState.Resolutions[0])

	// Change device control (resolution)
	changeControlReqBody := make([]byte, 12)
	changeControlReqBody[0] = 2                               // device 2
	client.byteOrder.PutUint16(changeControlReqBody[2:4], 1)  // control DEVICE_RESOLUTION
	client.byteOrder.PutUint16(changeControlReqBody[4:6], 12) // length
	changeControlReqBody[5] = 0                               // first_valuator
	changeControlReqBody[6] = 1                               // num_valuators
	client.byteOrder.PutUint32(changeControlReqBody[8:12], 500)
	changeControlReq := &XInputRequest{MinorOpcode: XChangeDeviceControl, Body: changeControlReqBody}
	reply = s.handleRequest(client, changeControlReq, 3)
	require.NotNil(t, reply)
	_, ok = reply.(*ChangeDeviceControlReply)
	require.True(t, ok)

	// Get device control again to verify
	reply = s.handleRequest(client, getControlReq, 4)
	require.NotNil(t, reply)
	getControlReply, ok = reply.(*GetDeviceControlReply)
	require.True(t, ok)
	resolutionState, ok = getControlReply.Control.(*DeviceResolutionState)
	require.True(t, ok)
	assert.Equal(t, uint32(500), resolutionState.Resolutions[0])
}

func TestXGetSelectedExtensionEvents(t *testing.T) {
	s, client, _, _ := setupTestServerWithClient(t)

	// Create a window
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

	// Select for button press events
	mask := (uint32(DeviceButtonPressMask) << 8) | uint32(2)
	selectEventReqBody := make([]byte, 8+4)
	client.byteOrder.PutUint32(selectEventReqBody[0:4], 1) // window
	client.byteOrder.PutUint16(selectEventReqBody[4:6], 1) // num_classes
	client.byteOrder.PutUint32(selectEventReqBody[8:12], mask)
	selectEventReq := &XInputRequest{MinorOpcode: XSelectExtensionEvent, Body: selectEventReqBody}
	s.handleRequest(client, selectEventReq, 3)

	// Get selected events
	getSelectedReqBody := make([]byte, 4)
	client.byteOrder.PutUint32(getSelectedReqBody[0:4], 1) // window
	getSelectedReq := &XInputRequest{MinorOpcode: XGetSelectedExtensionEvents, Body: getSelectedReqBody}
	reply := s.handleRequest(client, getSelectedReq, 4)
	require.NotNil(t, reply)

	getSelectedReply, ok := reply.(*GetSelectedExtensionEventsReply)
	require.True(t, ok)
	require.Len(t, getSelectedReply.ThisClientClasses, 1)
	assert.Equal(t, mask, getSelectedReply.ThisClientClasses[0])
}

func TestDeviceDontPropagateList(t *testing.T) {
	s, client, _, _ := setupTestServerWithClient(t)

	// Create a window
	createWindowReq := &CreateWindowRequest{
		Drawable: 1,
		Parent:   0,
		Width:    100,
		Height:   100,
	}
	s.handleRequest(client, createWindowReq, 1)

	// Change don't propagate list
	changeListReqBody := make([]byte, 12)
	client.byteOrder.PutUint32(changeListReqBody[0:4], 1) // window
	changeListReqBody[4] = 0                              // mode AddToList
	client.byteOrder.PutUint16(changeListReqBody[6:8], 1) // num_classes
	client.byteOrder.PutUint32(changeListReqBody[8:12], 1234)
	changeListReq := &XInputRequest{MinorOpcode: XChangeDeviceDontPropagateList, Body: changeListReqBody}
	s.handleRequest(client, changeListReq, 2)

	// Get don't propagate list
	getListReqBody := make([]byte, 4)
	client.byteOrder.PutUint32(getListReqBody[0:4], 1) // window
	getListReq := &XInputRequest{MinorOpcode: XGetDeviceDontPropagateList, Body: getListReqBody}
	reply := s.handleRequest(client, getListReq, 3)
	require.NotNil(t, reply)

	getListReply, ok := reply.(*GetDeviceDontPropagateListReply)
	require.True(t, ok)
	require.Len(t, getListReply.Classes, 1)
	assert.Equal(t, uint32(1234), getListReply.Classes[0])
}

func TestXAllowDeviceEvents(t *testing.T) {
	s, client, _, _ := setupTestServerWithClient(t)

	// Create a window
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
	s.handleRequest(client, grabReq, 3)

	// Allow events
	allowEventsReqBody := make([]byte, 8)
	client.byteOrder.PutUint32(allowEventsReqBody[0:4], 0) // time
	allowEventsReqBody[4] = 2                              // device_id
	allowEventsReqBody[5] = 0                              // event_mode AsyncThisDevice
	allowEventsReq := &XInputRequest{MinorOpcode: XAllowDeviceEvents, Body: allowEventsReqBody}
	s.handleRequest(client, allowEventsReq, 4)

	// The grab should still be active
	_, grabExists := s.deviceGrabs[2]
	assert.True(t, grabExists, "Device 2 should still be grabbed")
}

func TestXChangeCoreDevices(t *testing.T) {
	s, client, _, _ := setupTestServerWithClient(t)

	// Attempt to change the keyboard device, expecting an error
	changeKeyboardReqBody := []byte{3, 0, 0, 0} // device 3 (keyboard)
	changeKeyboardReq := &XInputRequest{MinorOpcode: XChangeKeyboardDevice, Body: changeKeyboardReqBody}
	reply := s.handleRequest(client, changeKeyboardReq, 1)
	require.NotNil(t, reply)
	_, ok := reply.(*DeviceError)
	require.True(t, ok)

	// Attempt to change the pointer device, expecting an error
	changePointerReqBody := []byte{2, 0, 1, 0} // device 2 (pointer), xaxis 0, yaxis 1
	changePointerReq := &XInputRequest{MinorOpcode: XChangePointerDevice, Body: changePointerReqBody}
	reply = s.handleRequest(client, changePointerReq, 2)
	require.NotNil(t, reply)
	_, ok = reply.(*DeviceError)
	require.True(t, ok)
}

func TestXSendExtensionEvent(t *testing.T) {
	s, client1, _, _ := setupTestServerWithClient(t)

	// Create a window
	createWindowReq := &CreateWindowRequest{
		Drawable: 1,
		Parent:   0,
		Width:    100,
		Height:   100,
	}
	s.handleRequest(client1, createWindowReq, 1)

	// Client 1 opens the pointer and selects for button press events
	openReqBody := []byte{2, 0, 0, 0} // device 2 (pointer)
	openReq := &XInputRequest{MinorOpcode: XOpenDevice, Body: openReqBody}
	s.handleRequest(client1, openReq, 2)

	class := (uint32(DeviceButtonPressMask) << 8) | uint32(2)
	selectEventReqBody := make([]byte, 8+4)
	client1.byteOrder.PutUint32(selectEventReqBody[0:4], 1) // window
	client1.byteOrder.PutUint16(selectEventReqBody[4:6], 1) // num_classes
	client1.byteOrder.PutUint32(selectEventReqBody[8:12], class)
	selectEventReq := &XInputRequest{MinorOpcode: XSelectExtensionEvent, Body: selectEventReqBody}
	s.handleRequest(client1, selectEventReq, 3)

	// Create a second client to send the event
	client2 := &x11Client{
		id:           s.nextClientID,
		sequence:     0,
		byteOrder:    s.byteOrder,
		saveSet:      make(map[uint32]bool),
		openDevices:  make(map[byte]*deviceInfo),
		sentMessages: make([]messageEncoder, 0),
	}
	s.clients[client2.id] = client2
	s.nextClientID++

	// Client 2 sends a button press event
	event := &DeviceButtonPressEvent{
		DeviceID: 2,
		Detail:   1, // Button 1
	}
	eventBytes := make([]byte, 32)
	eventBytes[0] = byte(xInputOpcode)
	eventBytes[1] = byte(DeviceButtonPress)
	eventBytes[31] = event.Detail

	sendEventReqBody := make([]byte, 12+32+4)
	client2.byteOrder.PutUint32(sendEventReqBody[0:4], 1)  // destination
	sendEventReqBody[4] = 2                                // device_id
	sendEventReqBody[5] = 1                                // propagate
	client2.byteOrder.PutUint16(sendEventReqBody[8:10], 1) // num_classes
	sendEventReqBody[10] = 1                               // num_events
	copy(sendEventReqBody[12:12+32], eventBytes)
	client2.byteOrder.PutUint32(sendEventReqBody[12+32:12+32+4], class)
	sendEventReq := &XInputRequest{MinorOpcode: XSendExtensionEvent, Body: sendEventReqBody}
	s.handleRequest(client2, sendEventReq, 1)

	// Check that client 1 received the event
	require.Len(t, client1.sentMessages, 1, "Client 1 should have received a message")
	rawEvent, ok := client1.sentMessages[0].(*x11RawEvent)
	require.True(t, ok, "Expected a raw event")
	assert.Equal(t, byte(xInputOpcode), rawEvent.data[0], "Opcode should be XInputExtension")
	assert.Equal(t, byte(DeviceButtonPress), rawEvent.data[1], "Sub-opcode should be DeviceButtonPress")
	assert.Equal(t, byte(1), rawEvent.data[31], "Button should be 1")
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
	assert.Equal(t, byte(1), buttonPressEvent.Detail, "Button should be 1")
	assert.Equal(t, byte(2), buttonPressEvent.DeviceID, "Device ID should be 2")
}
