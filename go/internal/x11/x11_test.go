//go:build x11

package x11

import (
	"bytes"
	"encoding/binary"
	"testing"
)

// setupTestServer creates a new x11Server with a mock frontend and a single mock client.
// It returns the server instance and a buffer that captures all data sent to the mock client.
func setupTestServer(t *testing.T) (*x11Server, *bytes.Buffer) {
	var clientBuffer bytes.Buffer
	mockConn := &testConn{r: &bytes.Buffer{}, w: &clientBuffer}

	client := &x11Client{
		id:        1,
		conn:      mockConn,
		byteOrder: binary.LittleEndian,
		sequence:  1, // Start sequence at 1
	}

	server := &x11Server{
		logger:       &testLogger{t: t},
		windows:      make(map[xID]*window),
		clients:      map[uint32]*x11Client{1: client},
		frontend:     &MockX11Frontend{},
		byteOrder:    binary.LittleEndian,
		passiveGrabs: make(map[xID][]*passiveGrab),
	}

	return server, &clientBuffer
}

func TestSendMouseEvent_EventMask_Sent(t *testing.T) {
	server, _ := setupTestServer(t)
	client := server.clients[1]

	windowID := xID{client: 1, local: 10}
	server.windows[windowID] = &window{
		xid:        windowID,
		attributes: WindowAttributes{EventMask: ButtonPressMask},
	}

	// Send a button press event
	server.SendMouseEvent(windowID, "mousedown", 10, 20, (1<<16)|1)

	if len(client.sentMessages) == 0 {
		t.Fatal("Expected event to be sent, but no message was recorded")
	}
}

func TestSendMouseEvent_EventMask_Blocked(t *testing.T) {
	server, _ := setupTestServer(t)
	client := server.clients[1]

	windowID := xID{client: 1, local: 10}
	server.windows[windowID] = &window{
		xid:        windowID,
		attributes: WindowAttributes{EventMask: PointerMotionMask}, // Does not include ButtonPressMask
	}

	// Send a button press event
	server.SendMouseEvent(windowID, "mousedown", 10, 20, (1<<16)|1)

	if len(client.sentMessages) != 0 {
		t.Fatalf("Expected event to be blocked by event mask, but %d messages were sent", len(client.sentMessages))
	}
}

func TestSendMouseEvent_ActivePointerGrab_Redirected(t *testing.T) {
	server, _ := setupTestServer(t)
	client := server.clients[1]

	originalWindowID := xID{client: 1, local: 10}
	grabWindowID := xID{client: 1, local: 20}

	server.windows[originalWindowID] = &window{
		xid:        originalWindowID,
		attributes: WindowAttributes{EventMask: ButtonPressMask},
	}
	server.windows[grabWindowID] = &window{
		xid:        grabWindowID,
		attributes: WindowAttributes{EventMask: 0}, // Grab window doesn't need the mask
	}

	// Grab the pointer on grabWindowID
	server.pointerGrabWindow = grabWindowID
	server.pointerGrabEventMask = ButtonPressMask
	server.pointerGrabOwner = false // Event should be sent to grabWindowID

	// Send a button press event to the original window
	server.SendMouseEvent(originalWindowID, "mousedown", 10, 20, (1<<16)|1)

	if len(client.sentMessages) == 0 {
		t.Fatal("Expected event to be sent, but no message was recorded")
	}

	// Verify the event was sent to the grab window
	event := client.sentMessages[0].(*ButtonPressEvent)
	if event.event != grabWindowID.local {
		t.Errorf("Expected event to be redirected to window %d, but it was sent to %d", grabWindowID.local, event.event)
	}
}

func TestSendMouseEvent_ActivePointerGrab_OwnerEventsTrue(t *testing.T) {
	server, _ := setupTestServer(t)
	client := server.clients[1]

	originalWindowID := xID{client: 1, local: 10}
	grabWindowID := xID{client: 1, local: 20}

	server.windows[originalWindowID] = &window{
		xid:        originalWindowID,
		attributes: WindowAttributes{EventMask: ButtonPressMask},
	}
	server.windows[grabWindowID] = &window{
		xid:        grabWindowID,
		attributes: WindowAttributes{EventMask: 0},
	}

	// Grab the pointer with ownerEvents = true
	server.pointerGrabWindow = grabWindowID
	server.pointerGrabEventMask = ButtonPressMask
	server.pointerGrabOwner = true // Event should be sent to originalWindowID

	// Send a button press event to the original window
	server.SendMouseEvent(originalWindowID, "mousedown", 10, 20, (1<<16)|1)

	if len(client.sentMessages) == 0 {
		t.Fatal("Expected event to be sent, but no message was recorded")
	}

	event := client.sentMessages[0].(*ButtonPressEvent)
	if event.event != originalWindowID.local {
		t.Errorf("Expected event to be sent to original window %d, but it was sent to %d", originalWindowID.local, event.event)
	}
}

func TestSendMouseEvent_ActivePointerGrab_MaskBlocked(t *testing.T) {
	server, _ := setupTestServer(t)
	client := server.clients[1]

	originalWindowID := xID{client: 1, local: 10}
	grabWindowID := xID{client: 1, local: 20}

	server.windows[originalWindowID] = &window{xid: originalWindowID}
	server.windows[grabWindowID] = &window{xid: grabWindowID}

	// Grab the pointer, but with a mask that doesn't include ButtonPress
	server.pointerGrabWindow = grabWindowID
	server.pointerGrabEventMask = PointerMotionMask

	// Send a button press event
	server.SendMouseEvent(originalWindowID, "mousedown", 10, 20, (1<<16)|1)

	if len(client.sentMessages) != 0 {
		t.Fatalf("Expected event to be blocked by grab event mask, but %d messages were sent", len(client.sentMessages))
	}
}

func TestSendMouseEvent_PassiveGrab_Activates(t *testing.T) {
	server, _ := setupTestServer(t)

	windowID := xID{client: 1, local: 10}
	server.windows[windowID] = &window{
		xid:        windowID,
		attributes: WindowAttributes{EventMask: ButtonPressMask},
	}

	// Setup a passive grab on the window for Button 1
	server.passiveGrabs[windowID] = []*passiveGrab{
		{
			button:    1,
			modifiers: 0,
			owner:     false,
			eventMask: ButtonPressMask | ButtonReleaseMask,
		},
	}

	// Send a button press event that should activate the grab
	// state = 0, button = 1
	server.SendMouseEvent(windowID, "mousedown", 10, 20, 1)

	// Check that the pointer grab is now active
	if server.pointerGrabWindow != windowID {
		t.Errorf("Expected pointer grab to be activated on window %s, but it was not", windowID)
	}
	if server.pointerGrabEventMask != (ButtonPressMask | ButtonReleaseMask) {
		t.Errorf("Expected grab event mask to be %d, but got %d", (ButtonPressMask | ButtonReleaseMask), server.pointerGrabEventMask)
	}
}

func TestSendKeyboardEvent_ActiveKeyboardGrab_Redirected(t *testing.T) {
	server, _ := setupTestServer(t)
	client := server.clients[1]

	originalWindowID := xID{client: 1, local: 10}
	grabWindowID := xID{client: 1, local: 20}

	server.windows[originalWindowID] = &window{
		xid:        originalWindowID,
		attributes: WindowAttributes{EventMask: KeyPressMask},
	}
	server.windows[grabWindowID] = &window{
		xid:        grabWindowID,
		attributes: WindowAttributes{EventMask: 0},
	}

	// Grab the keyboard on grabWindowID
	server.keyboardGrabWindow = grabWindowID
	server.keyboardGrabOwner = false // Event should be sent to grabWindowID

	// Send a key press event to the original window
	server.SendKeyboardEvent(originalWindowID, "keydown", "KeyA", false, false, false, false)

	if len(client.sentMessages) == 0 {
		t.Fatal("Expected event to be sent, but no message was recorded")
	}

	// Verify the event was sent to the grab window
	event := client.sentMessages[0].(*keyEvent)
	if event.event != grabWindowID.local {
		t.Errorf("Expected event to be redirected to window %d, but it was sent to %d", grabWindowID.local, event.event)
	}
}

func TestSendKeyboardEvent_ActiveKeyboardGrab_OwnerEventsTrue(t *testing.T) {
	server, _ := setupTestServer(t)
	client := server.clients[1]

	originalWindowID := xID{client: 1, local: 10}
	grabWindowID := xID{client: 1, local: 20}

	server.windows[originalWindowID] = &window{
		xid:        originalWindowID,
		attributes: WindowAttributes{EventMask: KeyPressMask},
	}
	server.windows[grabWindowID] = &window{
		xid:        grabWindowID,
		attributes: WindowAttributes{EventMask: 0},
	}

	// Grab the keyboard with ownerEvents = true
	server.keyboardGrabWindow = grabWindowID
	server.keyboardGrabOwner = true // Event should be sent to originalWindowID

	// Send a key press event to the original window
	server.SendKeyboardEvent(originalWindowID, "keydown", "KeyA", false, false, false, false)

	if len(client.sentMessages) == 0 {
		t.Fatal("Expected event to be sent, but no message was recorded")
	}

	event := client.sentMessages[0].(*keyEvent)
	if event.event != originalWindowID.local {
		t.Errorf("Expected event to be sent to original window %d, but it was sent to %d", originalWindowID.local, event.event)
	}
}

func TestSendKeyboardEvent_EventMask_Sent(t *testing.T) {
	server, _ := setupTestServer(t)
	client := server.clients[1]

	windowID := xID{client: 1, local: 10}
	server.windows[windowID] = &window{
		xid:        windowID,
		attributes: WindowAttributes{EventMask: KeyPressMask},
	}

	// Send a key press event
	server.SendKeyboardEvent(windowID, "keydown", "KeyA", false, false, false, false)

	if len(client.sentMessages) == 0 {
		t.Fatal("Expected event to be sent, but no message was recorded")
	}
}

func TestSendKeyboardEvent_EventMask_Blocked(t *testing.T) {
	server, _ := setupTestServer(t)
	client := server.clients[1]

	windowID := xID{client: 1, local: 10}
	server.windows[windowID] = &window{
		xid:        windowID,
		attributes: WindowAttributes{EventMask: KeyReleaseMask}, // Does not include KeyPressMask
	}

	// Send a key press event
	server.SendKeyboardEvent(windowID, "keydown", "KeyA", false, false, false, false)

	if len(client.sentMessages) != 0 {
		t.Fatalf("Expected event to be blocked by event mask, but %d messages were sent", len(client.sentMessages))
	}
}
