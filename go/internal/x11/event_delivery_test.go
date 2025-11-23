//go:build x11 && !wasm

package x11

import (
	"testing"

	"github.com/c2FmZQ/sshterm/internal/x11/wire"
	"github.com/stretchr/testify/assert"
)

func TestEventDelivery_SingleClient_CorrectWindow(t *testing.T) {
	server, client, _, buffer := setupTestServerWithClient(t)

	// 1. Create two windows
	window1ID := client.xID(100)
	window2ID := client.xID(200)

	server.windows[window1ID] = &window{
		xid:        window1ID,
		attributes: wire.WindowAttributes{EventMask: wire.ButtonPressMask},
	}
	server.windows[window2ID] = &window{
		xid:        window2ID,
		attributes: wire.WindowAttributes{EventMask: wire.KeyPressMask},
	}

	// 2. Send a mouse event to window 1
	server.SendMouseEvent(window1ID, "mousedown", 10, 10, 1)

	// 3. Verify window 1 gets the mouse event
	messages := drainMessages(t, buffer, client.byteOrder)
	assert.Len(t, messages, 1, "Expected one mouse event for window 1")
	if len(messages) == 1 {
		btnEvent, ok := messages[0].(*wire.ButtonPressEvent)
		assert.True(t, ok, "Expected a ButtonPressEvent")
		assert.Equal(t, window1ID.local, btnEvent.Event, "Event should be for window 1")
	}

	// 4. Send a keyboard event to window 2
	server.inputFocus = window2ID // Set focus to deliver keyboard events
	server.SendKeyboardEvent(window2ID, "keydown", "KeyB", false, false, false, false)

	// 5. Verify window 2 gets the keyboard event
	messages = drainMessages(t, buffer, client.byteOrder)
	assert.Len(t, messages, 1, "Expected one keyboard event for window 2")
	if len(messages) == 1 {
		keyEvent, ok := messages[0].(*wire.KeyEvent)
		assert.True(t, ok, "Expected a KeyEvent")
		assert.Equal(t, window2ID.local, keyEvent.Event, "Event should be for window 2")
	}

	// 6. Send a mouse event to window 2 (which is not listening for it)
	server.SendMouseEvent(window2ID, "mousedown", 10, 10, 1)
	messages = drainMessages(t, buffer, client.byteOrder)
	assert.Len(t, messages, 0, "Expected no mouse event for window 2")
}

func TestEventDelivery_TwoClients_SimpleGrab(t *testing.T) {
	server, clients, _, buffers := setupTestServerWithClients(t, 2)
	clientA := clients[0]
	clientB := clients[1]
	bufferA := buffers[0]
	bufferB := buffers[1]

	// 1. Client A creates a window and listens for button presses
	windowID := clientA.xID(100)
	server.windows[windowID] = &window{
		xid:        windowID,
		attributes: wire.WindowAttributes{EventMask: wire.ButtonPressMask},
	}

	// 2. Client B places a passive grab on Client A's window
	grabReq := &wire.GrabButtonRequest{
		GrabWindow:  wire.Window(windowID.local),
		OwnerEvents: false,
		EventMask:   wire.ButtonPressMask,
		Button:      1, // Grab button 1
		Modifiers:   wire.AnyModifier,
	}
	server.handleRequest(clientB, grabReq, 1)

	// 3. Simulate a mouse click on the window
	server.SendMouseEvent(windowID, "mousedown", 20, 20, 1) // Button 1

	// 4. Verify that only Client B (the grabber) gets the event
	messagesA := drainMessages(t, bufferA, clientA.byteOrder)
	assert.Len(t, messagesA, 0, "Client A should not receive the event due to the grab")

	messagesB := drainMessages(t, bufferB, clientB.byteOrder)
	assert.Len(t, messagesB, 1, "Client B should receive the grabbed event")
	if len(messagesB) == 1 {
		_, ok := messagesB[0].(*wire.ButtonPressEvent)
		assert.True(t, ok, "Expected a ButtonPressEvent for Client B")
	}
}

func TestEventDelivery_TwoClients_SimpleKeyGrab(t *testing.T) {
	server, clients, _, buffers := setupTestServerWithClients(t, 2)
	clientA := clients[0]
	clientB := clients[1]
	bufferA := buffers[0]
	bufferB := buffers[1]

	// 1. Client A creates a window and listens for key presses
	windowID := clientA.xID(100)
	server.windows[windowID] = &window{
		xid:        windowID,
		attributes: wire.WindowAttributes{EventMask: wire.KeyPressMask},
	}
	server.inputFocus = windowID

	// 2. Client B places a passive grab on Client A's window
	grabReq := &wire.GrabKeyRequest{
		GrabWindow:  wire.Window(windowID.local),
		OwnerEvents: false,
		Modifiers:   wire.AnyModifier,
		Key:         wire.KeyCode(jsCodeToX11Keycode["KeyC"]),
	}
	server.handleRequest(clientB, grabReq, 1)

	// 3. Simulate a key press on the window
	server.SendKeyboardEvent(windowID, "keydown", "KeyC", false, false, false, false)

	// 4. Verify that only Client B (the grabber) gets the event
	messagesA := drainMessages(t, bufferA, clientA.byteOrder)
	assert.Len(t, messagesA, 0, "Client A should not receive the event due to the grab")

	messagesB := drainMessages(t, bufferB, clientB.byteOrder)
	assert.Len(t, messagesB, 1, "Client B should receive the grabbed event")
	if len(messagesB) == 1 {
		_, ok := messagesB[0].(*wire.KeyEvent)
		assert.True(t, ok, "Expected a KeyEvent for Client B")
	}
}

func TestEventDelivery_ActivePointerGrab(t *testing.T) {
	server, clients, _, buffers := setupTestServerWithClients(t, 2)
	clientA := clients[0]
	clientB := clients[1]
	bufferA := buffers[0]
	bufferB := buffers[1]

	// 1. Client A creates a window and listens for button presses
	windowA_ID := clientA.xID(100)
	server.windows[windowA_ID] = &window{
		xid:        windowA_ID,
		attributes: wire.WindowAttributes{EventMask: wire.ButtonPressMask},
	}

	// 2. Client B creates a window to grab the pointer
	windowB_ID := clientB.xID(200)
	server.windows[windowB_ID] = &window{xid: windowB_ID}

	// 3. Client B actively grabs the pointer
	grabReq := &wire.GrabPointerRequest{
		GrabWindow:  wire.Window(windowB_ID.local),
		OwnerEvents: false,
		EventMask:   wire.ButtonPressMask,
	}
	reply := server.handleRequest(clientB, grabReq, 1)
	assert.Equal(t, byte(wire.GrabSuccess), reply.(*wire.GrabPointerReply).Status, "GrabPointer should succeed")

	// 4. Simulate a mouse click on Client A's window
	server.SendMouseEvent(windowA_ID, "mousedown", 5, 5, 1)

	// 5. Verify that only Client B (the grabber) gets the event
	messagesA := drainMessages(t, bufferA, clientA.byteOrder)
	assert.Len(t, messagesA, 0, "Client A should not receive the event")

	messagesB := drainMessages(t, bufferB, clientB.byteOrder)
	assert.Len(t, messagesB, 1, "Client B should receive the grabbed event")
}

func TestEventDelivery_ActiveKeyboardGrab(t *testing.T) {
	server, clients, _, buffers := setupTestServerWithClients(t, 2)
	clientA := clients[0]
	clientB := clients[1]
	bufferA := buffers[0]
	bufferB := buffers[1]

	// 1. Client A creates a window and listens for key presses
	windowA_ID := clientA.xID(100)
	server.windows[windowA_ID] = &window{
		xid:        windowA_ID,
		attributes: wire.WindowAttributes{EventMask: wire.KeyPressMask},
	}
	server.inputFocus = windowA_ID

	// 2. Client B creates a window to grab the keyboard
	windowB_ID := clientB.xID(200)
	server.windows[windowB_ID] = &window{xid: windowB_ID}

	// 3. Client B actively grabs the keyboard
	grabReq := &wire.GrabKeyboardRequest{
		GrabWindow:  wire.Window(windowB_ID.local),
		OwnerEvents: false,
	}
	reply := server.handleRequest(clientB, grabReq, 1)
	assert.Equal(t, byte(wire.GrabSuccess), reply.(*wire.GrabKeyboardReply).Status, "GrabKeyboard should succeed")

	// 4. Simulate a key press
	server.SendKeyboardEvent(windowA_ID, "keydown", "KeyD", false, false, false, false)

	// 5. Verify that only Client B (the grabber) gets the event
	messagesA := drainMessages(t, bufferA, clientA.byteOrder)
	assert.Len(t, messagesA, 0, "Client A should not receive the event")

	messagesB := drainMessages(t, bufferB, clientB.byteOrder)
	assert.Len(t, messagesB, 1, "Client B should receive the grabbed event")
}

func TestEventDelivery_OwnerEvents(t *testing.T) {
	server, clients, _, buffers := setupTestServerWithClients(t, 2)
	clientA := clients[0]
	clientB := clients[1]
	bufferA := buffers[0]
	bufferB := buffers[1]

	// 1. Client A creates a window and listens for button presses
	windowID := clientA.xID(100)
	server.windows[windowID] = &window{
		xid:        windowID,
		attributes: wire.WindowAttributes{EventMask: wire.ButtonPressMask},
	}

	// 2. Client B places a passive grab with OwnerEvents = true
	grabReq := &wire.GrabButtonRequest{
		GrabWindow:  wire.Window(windowID.local),
		OwnerEvents: true,
		EventMask:   wire.ButtonPressMask,
		Button:      1,
		Modifiers:   wire.AnyModifier,
	}
	server.handleRequest(clientB, grabReq, 1)

	// 3. Simulate a mouse click on the window
	server.SendMouseEvent(windowID, "mousedown", 20, 20, 1)

	// 4. Verify that both clients get the event
	messagesA := drainMessages(t, bufferA, clientA.byteOrder)
	assert.Len(t, messagesA, 1, "Client A should receive the event")

	messagesB := drainMessages(t, bufferB, clientB.byteOrder)
	assert.Len(t, messagesB, 1, "Client B should also receive the event")
}
