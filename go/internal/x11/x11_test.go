//go:build x11 && !wasm

package x11

import (
	"bytes"
	"encoding/binary"
	"testing"
)

// setupTestServer creates a new x11Server with a mock frontend and a single mock client.
// It returns the server instance and a buffer that captures all data sent to the mock client.
func setupTestServer(t *testing.T) (*x11Server, *bytes.Buffer) {
	clientBuffer := &bytes.Buffer{}
	mockConn := &testConn{r: &bytes.Buffer{}, w: clientBuffer}

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

	return server, clientBuffer
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

func TestGetWindowAttributesRequest(t *testing.T) {
	server, clientBuffer := setupTestServer(t)
	client := server.clients[1]

	// 1. Create a window with known attributes
	windowID := xID{client: 1, local: 10}
	server.windows[windowID] = &window{
		xid: windowID,
		attributes: WindowAttributes{
			Class:            InputOutput,
			BitGravity:       NorthWestGravity,
			WinGravity:       NorthWestGravity,
			BackingStore:     NotUseful,
			BackingPlanes:    0,
			BackingPixel:     0,
			SaveUnder:        false,
			MapIsInstalled:   false,
			MapState:         IsUnmapped,
			OverrideRedirect: true,
			Colormap:         0,
			Cursor:           0,
			EventMask:        ButtonPressMask | KeyPressMask,
		},
	}

	// 2. Create and handle the GetWindowAttributes request
	req := &GetWindowAttributesRequest{Window: Window(windowID.local)}
	reply := server.handleRequest(client, req, 2)
	if reply == nil {
		t.Fatalf("handleRequest returned a nil reply")
	}

	// 3. Encode the reply and write to the client's buffer
	encodedReply := reply.encodeMessage(client.byteOrder)
	if _, err := clientBuffer.Write(encodedReply); err != nil {
		t.Fatalf("Failed to write reply to buffer: %v", err)
	}

	// 4. Decode the reply from the buffer
	var parsedReply getWindowAttributesReply
	err := binary.Read(clientBuffer, binary.LittleEndian, &parsedReply)
	if err != nil {
		t.Fatalf("Failed to read reply from buffer: %v", err)
	}

	// 5. Assert the reply fields match the window attributes
	if parsedReply.ReplyType != 1 {
		t.Errorf("Expected ReplyType 1, got %d", parsedReply.ReplyType)
	}
	if parsedReply.BackingStore != NotUseful {
		t.Errorf("Expected BackingStore %d, got %d", NotUseful, parsedReply.BackingStore)
	}
	if parsedReply.Sequence != 2 {
		t.Errorf("Expected Sequence 2, got %d", parsedReply.Sequence)
	}
	if parsedReply.Length != 3 { // 12 bytes / 4 = 3
		t.Errorf("Expected Length 3, got %d", parsedReply.Length)
	}
	if parsedReply.VisualID != 0 {
		t.Errorf("Expected VisualID 0, got %d", parsedReply.VisualID)
	}
	if parsedReply.Class != InputOutput {
		t.Errorf("Expected Class %d, got %d", InputOutput, parsedReply.Class)
	}
	if parsedReply.BitGravity != NorthWestGravity {
		t.Errorf("Expected BitGravity %d, got %d", NorthWestGravity, parsedReply.BitGravity)
	}
	if parsedReply.WinGravity != NorthWestGravity {
		t.Errorf("Expected WinGravity %d, got %d", NorthWestGravity, parsedReply.WinGravity)
	}
	if parsedReply.BackingPlanes != 0 {
		t.Errorf("Expected BackingPlanes 0, got %d", parsedReply.BackingPlanes)
	}
	if parsedReply.BackingPixel != 0 {
		t.Errorf("Expected BackingPixel 0, got %d", parsedReply.BackingPixel)
	}
	if parsedReply.SaveUnder != 0 { // 0 for false
		t.Errorf("Expected SaveUnder false (0), got %d", parsedReply.SaveUnder)
	}
	if parsedReply.MapIsInstalled != 0 { // 0 for false
		t.Errorf("Expected MapIsInstalled false (0), got %d", parsedReply.MapIsInstalled)
	}
	if parsedReply.MapState != IsUnmapped {
		t.Errorf("Expected MapState %d, got %d", IsUnmapped, parsedReply.MapState)
	}
	if parsedReply.OverrideRedirect != 1 { // 1 for true
		t.Errorf("Expected OverrideRedirect true (1), got %d", parsedReply.OverrideRedirect)
	}
	if parsedReply.Colormap != 0 {
		t.Errorf("Expected Colormap 0, got %d", parsedReply.Colormap)
	}
	if parsedReply.AllEventMasks != (ButtonPressMask | KeyPressMask) {
		t.Errorf("Expected AllEventMasks %d, got %d", (ButtonPressMask | KeyPressMask), parsedReply.AllEventMasks)
	}
	if parsedReply.YourEventMask != (ButtonPressMask | KeyPressMask) {
		t.Errorf("Expected YourEventMask %d, got %d", (ButtonPressMask | KeyPressMask), parsedReply.YourEventMask)
	}
}

func TestSendKeyboardEvent_PassiveGrab_Activates(t *testing.T) {
	server, _ := setupTestServer(t)
	client := server.clients[1]

	windowID := xID{client: 1, local: 10}
	server.windows[windowID] = &window{
		xid:        windowID,
		attributes: WindowAttributes{EventMask: KeyPressMask},
	}
	req := &GrabKeyRequest{
		GrabWindow: Window(windowID.local),
		Modifiers:  AnyModifier,
		Key:        38, // KeyA
	}
	server.handleRequest(client, req, 2)

	// Send a key press event that should activate the grab
	server.SendKeyboardEvent(windowID, "keydown", "KeyA", false, false, false, false)

	// Check that the keyboard grab is now active
	if server.keyboardGrabWindow != windowID {
		t.Errorf("Expected keyboard grab to be activated on window %s, but it was not", windowID)
	}
}

func TestUngrabKeyRequest(t *testing.T) {
	server, _ := setupTestServer(t)
	client := server.clients[1]

	windowID := xID{client: 1, local: 10}
	server.windows[windowID] = &window{
		xid: windowID,
	}

	// 1. Grab a key
	grabReq := &GrabKeyRequest{
		GrabWindow:  Window(windowID.local),
		Modifiers:   ShiftMask,
		Key:         38, // KeyA
		OwnerEvents: false,
	}
	server.handleRequest(client, grabReq, 2)

	// Verify grab exists
	if len(server.passiveGrabs[windowID]) != 1 {
		t.Fatalf("GrabKey did not create a passive grab. Expected 1, got %d", len(server.passiveGrabs[windowID]))
	}

	// 2. Ungrab the key
	ungrabReq := &UngrabKeyRequest{
		GrabWindow: Window(windowID.local),
		Modifiers:  ShiftMask,
		Key:        38, // KeyA
	}
	server.handleRequest(client, ungrabReq, 3)

	// Verify grab is removed
	if len(server.passiveGrabs[windowID]) != 0 {
		t.Errorf("UngrabKey did not remove the passive grab. Expected 0, got %d", len(server.passiveGrabs[windowID]))
	}
}

func TestUngrabKeyRequest_AnyModifier(t *testing.T) {
	server, _ := setupTestServer(t)
	client := server.clients[1]

	windowID := xID{client: 1, local: 10}
	server.windows[windowID] = &window{
		xid: windowID,
	}

	// 1. Grab a key with different modifiers
	grabReq1 := &GrabKeyRequest{GrabWindow: Window(windowID.local), Modifiers: ShiftMask, Key: 38}
	grabReq2 := &GrabKeyRequest{GrabWindow: Window(windowID.local), Modifiers: ControlMask, Key: 38}
	server.handleRequest(client, grabReq1, 2)
	server.handleRequest(client, grabReq2, 3)

	if len(server.passiveGrabs[windowID]) != 2 {
		t.Fatalf("GrabKey did not create passive grabs. Expected 2, got %d", len(server.passiveGrabs[windowID]))
	}

	// 2. Ungrab the key with AnyModifier
	ungrabReq := &UngrabKeyRequest{
		GrabWindow: Window(windowID.local),
		Modifiers:  AnyModifier,
		Key:        38, // KeyA
	}
	server.handleRequest(client, ungrabReq, 4)

	// Verify all grabs for that key are removed
	if len(server.passiveGrabs[windowID]) != 0 {
		t.Errorf("UngrabKey with AnyModifier did not remove all passive grabs. Expected 0, got %d", len(server.passiveGrabs[windowID]))
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

func TestWindowHierarchyRequests(t *testing.T) {
	server, clientBuffer := setupTestServer(t)
	client := server.clients[1]
	mockFrontend := server.frontend.(*MockX11Frontend)

	// 1. Create a parent window and a child window
	parentID := xID{client: 1, local: 10}
	childID := xID{client: 1, local: 20}
	server.windows[parentID] = &window{xid: parentID, children: []uint32{childID.local}}
	server.windows[childID] = &window{xid: childID, parent: parentID.local}

	// 2. Test ReparentWindow
	newParentID := xID{client: 1, local: 30}
	server.windows[newParentID] = &window{xid: newParentID}
	reparentReq := &ReparentWindowRequest{Window: Window(childID.local), Parent: Window(newParentID.local), X: 10, Y: 20}
	server.handleRequest(client, reparentReq, 2)

	if server.windows[childID].parent != newParentID.local {
		t.Errorf("ReparentWindow: child's parent was not updated")
	}
	if len(server.windows[parentID].children) != 0 {
		t.Errorf("ReparentWindow: child was not removed from old parent's children")
	}
	if len(server.windows[newParentID].children) != 1 || server.windows[newParentID].children[0] != childID.local {
		t.Errorf("ReparentWindow: child was not added to new parent's children")
	}
	if len(mockFrontend.ReparentWindowCalls) != 1 {
		t.Errorf("ReparentWindow: expected frontend to be called")
	}

	// 3. Test CirculateWindow
	circulateReq := &CirculateWindowRequest{Window: Window(childID.local), Direction: 0 /* RaiseLowest */}
	server.handleRequest(client, circulateReq, 3)
	if len(mockFrontend.CirculateWindowCalls) != 1 {
		t.Errorf("CirculateWindow: expected frontend to be called")
	}

	// 4. Test QueryTree
	queryTreeReq := &QueryTreeRequest{Window: Window(newParentID.local)}
	reply := server.handleRequest(client, queryTreeReq, 4)
	if reply == nil {
		t.Fatalf("QueryTree: handleRequest returned a nil reply")
	}
	encodedReply := reply.encodeMessage(client.byteOrder)
	if _, err := clientBuffer.Write(encodedReply); err != nil {
		t.Fatalf("QueryTree: failed to write reply to buffer: %v", err)
	}

	var header struct {
		ReplyType   byte
		Unused      byte
		Sequence    uint16
		Length      uint32
		Root        uint32
		Parent      uint32
		NumChildren uint16
		Padding     [14]byte
	}
	if err := binary.Read(clientBuffer, binary.LittleEndian, &header); err != nil {
		t.Fatalf("QueryTree: failed to read reply header: %v", err)
	}
	children := make([]uint32, header.NumChildren)
	if err := binary.Read(clientBuffer, binary.LittleEndian, &children); err != nil {
		t.Fatalf("QueryTree: failed to read reply children: %v", err)
	}

	if header.NumChildren != 1 || children[0] != childID.local {
		t.Errorf("QueryTree: incorrect children returned. Got %d children: %v", header.NumChildren, children)
	}

	// 5. Test DestroySubwindows
	destroyReq := &DestroySubwindowsRequest{Window: Window(newParentID.local)}
	server.handleRequest(client, destroyReq, 5)
	if _, exists := server.windows[childID]; exists {
		t.Errorf("DestroySubwindows: child window was not destroyed")
	}
	if len(server.windows[newParentID].children) != 0 {
		t.Errorf("DestroySubwindows: children were not removed from parent")
	}
	if len(mockFrontend.DestroySubwindowsCalls) != 1 {
		t.Errorf("DestroySubwindows: expected frontend to be called")
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
