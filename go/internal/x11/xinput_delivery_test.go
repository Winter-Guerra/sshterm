//go:build x11 && !wasm

package x11

import (
	"testing"

	"github.com/c2FmZQ/sshterm/internal/x11/wire"
	"github.com/stretchr/testify/assert"
)

func TestXInputEventDelivery_Simple(t *testing.T) {
	server, client, _, buffer := setupTestServerWithClient(t)

	// 1. Create a window
	windowID := client.xID(100)
	server.windows[windowID] = &window{
		xid:        windowID,
		attributes: wire.WindowAttributes{},
	}

	// 2. Open the virtual pointer device (ID 2)
	openDevReq := &wire.OpenDeviceRequest{
		DeviceID: 2, // Virtual Pointer
	}
	reply := server.handleRequest(client, openDevReq, 1)
	assert.NotNil(t, reply)
	_, ok := reply.(*wire.OpenDeviceReply)
	assert.True(t, ok, "Expected OpenDeviceReply")

	// 3. Select XInput ButtonPress events on the window
	// Mask for DeviceButtonPress is DeviceButtonPressMask (1<<2 = 4)
	// Class = (Mask << 8) | DeviceID
	class := uint32(wire.DeviceButtonPressMask<<8) | 2
	selectReq := &wire.SelectExtensionEventRequest{
		Window:  wire.Window(windowID.local),
		Classes: []uint32{class},
	}
	server.handleRequest(client, selectReq, 1)

	// 4. Send a mouse down event
	// SendMouseEvent will trigger both core and XInput events if configured
	server.SendMouseEvent(windowID, "mousedown", 10, 10, 1)

	// 5. Verify delivery
	messages := drainMessages(t, buffer, client.byteOrder)

	// We might get core events if defaults allow, but we definitely want the XInput event
	var foundXInputEvent bool
	for _, msg := range messages {
		if inputEvent, ok := msg.(*wire.DeviceButtonPressEvent); ok {
			foundXInputEvent = true
			assert.Equal(t, byte(2), inputEvent.DeviceID)
			assert.Equal(t, windowID.local, inputEvent.Event)
			assert.Equal(t, byte(1), inputEvent.Detail)
		}
	}
	assert.True(t, foundXInputEvent, "Expected DeviceButtonPressEvent")
}

func TestXInputEventDelivery_Keyboard(t *testing.T) {
	server, client, _, buffer := setupTestServerWithClient(t)

	// 1. Create a window and focus it
	windowID := client.xID(100)
	server.windows[windowID] = &window{
		xid:        windowID,
		attributes: wire.WindowAttributes{},
	}
	server.inputFocus = windowID

	// 2. Open the virtual keyboard device (ID 3)
	openDevReq := &wire.OpenDeviceRequest{
		DeviceID: 3, // Virtual Keyboard
	}
	reply := server.handleRequest(client, openDevReq, 1)
	assert.NotNil(t, reply)

	// 3. Select XInput KeyPress events
	// Mask for DeviceKeyPress is DeviceKeyPressMask (1<<0 = 1)
	class := uint32(wire.DeviceKeyPressMask<<8) | 3
	selectReq := &wire.SelectExtensionEventRequest{
		Window:  wire.Window(windowID.local),
		Classes: []uint32{class},
	}
	server.handleRequest(client, selectReq, 1)

	// 4. Send a key down event
	server.SendKeyboardEvent(windowID, "keydown", "KeyA", false, false, false, false)

	// 5. Verify delivery
	messages := drainMessages(t, buffer, client.byteOrder)

	var foundXInputEvent bool
	for _, msg := range messages {
		if inputEvent, ok := msg.(*wire.DeviceKeyPressEvent); ok {
			foundXInputEvent = true
			assert.Equal(t, byte(3), inputEvent.DeviceID)
			assert.Equal(t, windowID.local, inputEvent.Event)
			// Detail/KeyCode checks depends on keymap, skipping for now
		}
	}
	assert.True(t, foundXInputEvent, "Expected DeviceKeyPressEvent")
}

func TestXInput2_QueryVersion(t *testing.T) {
	server, client, _, _ := setupTestServerWithClient(t)

	// XIQueryVersion
	req := &wire.XIQueryVersionRequest{
		MajorVersion: 2,
		MinorVersion: 2,
	}
	reply := server.handleRequest(client, req, 1)
	assert.NotNil(t, reply)
	xiReply, ok := reply.(*wire.XIQueryVersionReply)
	if assert.True(t, ok, "Expected XIQueryVersionReply, got %T", reply) {
		assert.Equal(t, uint16(2), xiReply.MajorVersion)
		assert.Equal(t, uint16(2), xiReply.MinorVersion)
	}
}

func TestXInput2_SelectEvents(t *testing.T) {
	server, client, _, _ := setupTestServerWithClient(t)
	windowID := client.xID(100)
	server.windows[windowID] = &window{xid: windowID}

	// XISelectEvents
	mask := []byte{0x00, 0x00, 0x00, 0x00} // Empty mask for now
	req := &wire.XISelectEventsRequest{
		Window:   wire.Window(windowID.local),
		NumMasks: 1,
		Masks: []wire.XIEventMask{
			{
				DeviceID: 2, // Virtual Pointer
				MaskLen:  uint16(len(mask)),
				Mask:     mask,
			},
		},
	}
	// This is expected to fail or do nothing currently as it is unhandled
	reply := server.handleRequest(client, req, 1)

	// Since it falls to default, it returns a RequestErrorCode (BadImplementation or similar default error)
	// The default case in handleXInputRequest returns:
	// wire.NewError(wire.RequestErrorCode, seq, 0, 0, wire.XInputOpcode)

	// If it was implemented, it would return nil (success).
	// Currently, it returns nil but functionality is not fully implemented (stub).
	assert.Nil(t, reply, "XISelectEvents should be handled (return nil)")
}

func TestXInputGrab_Active(t *testing.T) {
	server, client, _, buffer := setupTestServerWithClient(t)

	// 1. Create a window and listen for Core events
	windowID := client.xID(100)
	server.windows[windowID] = &window{
		xid:        windowID,
		attributes: wire.WindowAttributes{EventMask: wire.ButtonPressMask},
	}

	// 2. Open device
	server.handleRequest(client, &wire.OpenDeviceRequest{DeviceID: 2}, 1)

	// 3. Active Grab Device
	mask := uint32(wire.DeviceButtonPressMask<<8) | 2
	grabReq := &wire.GrabDeviceRequest{
		GrabWindow:  uint32(windowID.local),
		DeviceID:    2,
		OwnerEvents: false,
		NumClasses:  1,
		Classes:     []uint32{mask},
		Time:        0,
	}
	reply := server.handleRequest(client, grabReq, 2)
	assert.NotNil(t, reply)
	grabReply, ok := reply.(*wire.GrabDeviceReply)
	assert.True(t, ok, "Expected GrabDeviceReply")
	assert.Equal(t, wire.GrabSuccess, grabReply.Status)

	// 4. Send event
	server.SendMouseEvent(windowID, "mousedown", 10, 10, 1)

	// 5. Verify
	messages := drainMessages(t, buffer, client.byteOrder)
	var foundXInput bool
	var foundCore bool
	for _, msg := range messages {
		if _, ok := msg.(*wire.DeviceButtonPressEvent); ok {
			foundXInput = true
		}
		if _, ok := msg.(*wire.ButtonPressEvent); ok {
			foundCore = true
		}
	}
	assert.True(t, foundXInput, "Should receive XInput event")
	assert.False(t, foundCore, "Should NOT receive Core event due to device grab")
}

func TestXInputGrab_PassiveButton(t *testing.T) {
	server, client, _, buffer := setupTestServerWithClient(t)

	// 1. Create window with Core event mask
	windowID := client.xID(100)
	server.windows[windowID] = &window{
		xid:        windowID,
		attributes: wire.WindowAttributes{EventMask: wire.ButtonPressMask},
	}

	// 2. Open device
	server.handleRequest(client, &wire.OpenDeviceRequest{DeviceID: 2}, 1)

	// 3. Establish Passive Grab (GrabDeviceButton)
	mask := uint32(wire.DeviceButtonPressMask<<8) | 2
	passiveReq := &wire.GrabDeviceButtonRequest{
		GrabWindow:  wire.Window(windowID.local),
		DeviceID:    2,
		Button:      1, // Left button
		Modifiers:   wire.AnyModifier,
		OwnerEvents: false,
		NumClasses:  1,
		Classes:     []uint32{mask},
	}
	server.handleRequest(client, passiveReq, 2)

	// 4. Send mousedown (should activate grab)
	server.SendMouseEvent(windowID, "mousedown", 10, 10, 1)

	// 5. Verify
	messages := drainMessages(t, buffer, client.byteOrder)
	var foundXInput bool
	var foundCore bool
	for _, msg := range messages {
		if _, ok := msg.(*wire.DeviceButtonPressEvent); ok {
			foundXInput = true
		}
		if _, ok := msg.(*wire.ButtonPressEvent); ok {
			foundCore = true
		}
	}
	assert.True(t, foundXInput, "Should receive XInput event from passive grab")
	assert.False(t, foundCore, "Should NOT receive Core event due to activated device grab")
}

func TestXInputGrab_PassiveKey(t *testing.T) {
	server, client, _, buffer := setupTestServerWithClient(t)

	// 1. Create window
	windowID := client.xID(100)
	server.windows[windowID] = &window{
		xid:        windowID,
		attributes: wire.WindowAttributes{EventMask: wire.KeyPressMask},
	}
	server.inputFocus = windowID

	// 2. Open device
	server.handleRequest(client, &wire.OpenDeviceRequest{DeviceID: 3}, 1)

	// 3. Passive Grab Key
	mask := uint32(wire.DeviceKeyPressMask<<8) | 3
	passiveReq := &wire.GrabDeviceKeyRequest{
		GrabWindow:  wire.Window(windowID.local),
		DeviceID:    3,
		Key:         byte(jsCodeToX11Keycode["KeyA"]),
		Modifiers:   wire.AnyModifier,
		OwnerEvents: false,
		NumClasses:  1,
		Classes:     []uint32{mask},
	}
	server.handleRequest(client, passiveReq, 2)

	// 4. Send keydown
	server.SendKeyboardEvent(windowID, "keydown", "KeyA", false, false, false, false)

	// 5. Verify
	messages := drainMessages(t, buffer, client.byteOrder)
	var foundXInput bool
	var foundCore bool
	for _, msg := range messages {
		if _, ok := msg.(*wire.DeviceKeyPressEvent); ok {
			foundXInput = true
		}
		if _, ok := msg.(*wire.KeyEvent); ok {
			foundCore = true
		}
	}
	assert.True(t, foundXInput, "Should receive XInput event from passive grab")
	assert.False(t, foundCore, "Should NOT receive Core event due to activated device grab")
}
