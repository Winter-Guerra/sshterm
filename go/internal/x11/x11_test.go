//go:build x11 && !wasm

package x11

import (
	"bytes"
	"encoding/binary"
	"testing"

	"github.com/c2FmZQ/sshterm/internal/x11/wire"
)

// setupTestServer creates a new x11Server with a mock frontend and a single mock client.
// It returns the server instance and a buffer that captures all data sent to the mock client.
func setupTestServer(t *testing.T) (*x11Server, *bytes.Buffer) {
	server, _, _, clientBuffer := setupTestServerWithClient(t)
	return server, clientBuffer
}

func TestSendMouseEvent_EventMask_Sent(t *testing.T) {
	server, clientBuffer := setupTestServer(t)
	client := server.clients[1]

	windowID := xID{client: 1, local: 10}
	server.windows[windowID] = &window{
		xid:        windowID,
		attributes: wire.WindowAttributes{EventMask: wire.ButtonPressMask},
	}

	// Send a button press event
	server.SendMouseEvent(windowID, "mousedown", 10, 20, (1<<16)|1)

	messages := drainMessages(t, clientBuffer, client.byteOrder)
	if len(messages) == 0 {
		t.Fatal("Expected event to be sent, but no message was recorded")
	}
}

func TestSendMouseEvent_EventMask_Blocked(t *testing.T) {
	server, clientBuffer := setupTestServer(t)
	client := server.clients[1]

	windowID := xID{client: 1, local: 10}
	server.windows[windowID] = &window{
		xid:        windowID,
		attributes: wire.WindowAttributes{EventMask: wire.PointerMotionMask}, // Does not include ButtonPressMask
	}

	// Send a button press event
	server.SendMouseEvent(windowID, "mousedown", 10, 20, (1<<16)|1)

	messages := drainMessages(t, clientBuffer, client.byteOrder)
	if len(messages) != 0 {
		t.Fatalf("Expected event to be blocked by event mask, but %d messages were sent", len(messages))
	}
}

func TestGetWindowAttributesRequest(t *testing.T) {
	server, clientBuffer := setupTestServer(t)
	client := server.clients[1]

	// 1. Create a window with known attributes
	windowID := xID{client: 1, local: 10}
	server.windows[windowID] = &window{
		xid: windowID,
		attributes: wire.WindowAttributes{
			Class:            wire.InputOutput,
			BitGravity:       wire.NorthWestGravity,
			WinGravity:       wire.NorthWestGravity,
			BackingStore:     wire.NotUseful,
			BackingPlanes:    0,
			BackingPixel:     0,
			SaveUnder:        false,
			MapIsInstalled:   false,
			MapState:         wire.IsUnmapped,
			OverrideRedirect: true,
			Colormap:         0,
			Cursor:           0,
			EventMask:        wire.ButtonPressMask | wire.KeyPressMask,
		},
	}

	// 2. Create and handle the GetWindowAttributes request
	req := &wire.GetWindowAttributesRequest{Window: wire.Window(windowID.local)}
	reply := server.handleRequest(client, req, 2)
	if reply == nil {
		t.Fatalf("handleRequest returned a nil reply")
	}

	// 3. Encode the reply and write to the client's buffer
	encodedReply := reply.EncodeMessage(client.byteOrder)
	if _, err := clientBuffer.Write(encodedReply); err != nil {
		t.Fatalf("Failed to write reply to buffer: %v", err)
	}

	// 4. Decode the reply from the buffer
	var parsedReply wire.GetWindowAttributesReply
	err := binary.Read(clientBuffer, binary.LittleEndian, &parsedReply)
	if err != nil {
		t.Fatalf("Failed to read reply from buffer: %v", err)
	}

	// 5. Assert the reply fields match the window attributes
	if parsedReply.ReplyType != 1 {
		t.Errorf("Expected ReplyType 1, got %d", parsedReply.ReplyType)
	}
	if parsedReply.BackingStore != wire.NotUseful {
		t.Errorf("Expected BackingStore %d, got %d", wire.NotUseful, parsedReply.BackingStore)
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
	if parsedReply.Class != wire.InputOutput {
		t.Errorf("Expected Class %d, got %d", wire.InputOutput, parsedReply.Class)
	}
	if parsedReply.BitGravity != wire.NorthWestGravity {
		t.Errorf("Expected BitGravity %d, got %d", wire.NorthWestGravity, parsedReply.BitGravity)
	}
	if parsedReply.WinGravity != wire.NorthWestGravity {
		t.Errorf("Expected WinGravity %d, got %d", wire.NorthWestGravity, parsedReply.WinGravity)
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
	if parsedReply.MapState != wire.IsUnmapped {
		t.Errorf("Expected MapState %d, got %d", wire.IsUnmapped, parsedReply.MapState)
	}
	if parsedReply.OverrideRedirect != 1 { // 1 for true
		t.Errorf("Expected OverrideRedirect true (1), got %d", parsedReply.OverrideRedirect)
	}
	if parsedReply.Colormap != 0 {
		t.Errorf("Expected Colormap 0, got %d", parsedReply.Colormap)
	}
	if parsedReply.AllEventMasks != (wire.ButtonPressMask | wire.KeyPressMask) {
		t.Errorf("Expected AllEventMasks %d, got %d", (wire.ButtonPressMask | wire.KeyPressMask), parsedReply.AllEventMasks)
	}
	if parsedReply.YourEventMask != (wire.ButtonPressMask | wire.KeyPressMask) {
		t.Errorf("Expected YourEventMask %d, got %d", (wire.ButtonPressMask | wire.KeyPressMask), parsedReply.YourEventMask)
	}
}

func TestSendKeyboardEvent_PassiveGrab_Activates(t *testing.T) {
	server, _ := setupTestServer(t)
	client := server.clients[1]

	windowID := xID{client: 1, local: 10}
	server.windows[windowID] = &window{
		xid:        windowID,
		attributes: wire.WindowAttributes{EventMask: wire.KeyPressMask},
	}
	req := &wire.GrabKeyRequest{
		GrabWindow: wire.Window(windowID.local),
		Modifiers:  wire.AnyModifier,
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
	grabReq := &wire.GrabKeyRequest{
		GrabWindow:  wire.Window(windowID.local),
		Modifiers:   wire.ShiftMask,
		Key:         38, // KeyA
		OwnerEvents: false,
	}
	server.handleRequest(client, grabReq, 2)

	// Verify grab exists
	if len(server.passiveGrabs[windowID]) != 1 {
		t.Fatalf("GrabKey did not create a passive grab. Expected 1, got %d", len(server.passiveGrabs[windowID]))
	}

	// 2. Ungrab the key
	ungrabReq := &wire.UngrabKeyRequest{
		GrabWindow: wire.Window(windowID.local),
		Modifiers:  wire.ShiftMask,
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
	grabReq1 := &wire.GrabKeyRequest{GrabWindow: wire.Window(windowID.local), Modifiers: wire.ShiftMask, Key: 38}
	grabReq2 := &wire.GrabKeyRequest{GrabWindow: wire.Window(windowID.local), Modifiers: wire.ControlMask, Key: 38}
	server.handleRequest(client, grabReq1, 2)
	server.handleRequest(client, grabReq2, 3)

	if len(server.passiveGrabs[windowID]) != 2 {
		t.Fatalf("GrabKey did not create passive grabs. Expected 2, got %d", len(server.passiveGrabs[windowID]))
	}

	// 2. Ungrab the key with AnyModifier
	ungrabReq := &wire.UngrabKeyRequest{
		GrabWindow: wire.Window(windowID.local),
		Modifiers:  wire.AnyModifier,
		Key:        38, // KeyA
	}
	server.handleRequest(client, ungrabReq, 4)

	// Verify all grabs for that key are removed
	if len(server.passiveGrabs[windowID]) != 0 {
		t.Errorf("UngrabKey with AnyModifier did not remove all passive grabs. Expected 0, got %d", len(server.passiveGrabs[windowID]))
	}
}

func TestSendMouseEvent_ActivePointerGrab_Redirected(t *testing.T) {
	server, clientBuffer := setupTestServer(t)
	client := server.clients[1]

	originalWindowID := xID{client: 1, local: 10}
	grabWindowID := xID{client: 1, local: 20}

	server.windows[originalWindowID] = &window{
		xid:        originalWindowID,
		attributes: wire.WindowAttributes{EventMask: wire.ButtonPressMask},
	}
	server.windows[grabWindowID] = &window{
		xid:        grabWindowID,
		attributes: wire.WindowAttributes{EventMask: 0}, // Grab window doesn't need the mask
	}

	// Grab the pointer on grabWindowID
	server.pointerGrabWindow = grabWindowID
	server.pointerGrabEventMask = wire.ButtonPressMask
	server.pointerGrabOwner = false // Event should be sent to grabWindowID

	// Send a button press event to the original window
	server.SendMouseEvent(originalWindowID, "mousedown", 10, 20, (1<<16)|1)

	messages := drainMessages(t, clientBuffer, client.byteOrder)
	if len(messages) == 0 {
		t.Fatal("Expected event to be sent, but no message was recorded")
	}

	// Verify the event was sent to the grab window
	event := messages[0].(*wire.ButtonPressEvent)
	if event.Event != grabWindowID.local {
		t.Errorf("Expected event to be redirected to window %d, but it was sent to %d", grabWindowID.local, event.Event)
	}
}

func TestSendMouseEvent_ActivePointerGrab_OwnerEventsTrue(t *testing.T) {
	server, clientBuffer := setupTestServer(t)
	client := server.clients[1]

	originalWindowID := xID{client: 1, local: 10}
	grabWindowID := xID{client: 1, local: 20}

	server.windows[originalWindowID] = &window{
		xid:        originalWindowID,
		attributes: wire.WindowAttributes{EventMask: wire.ButtonPressMask},
	}
	server.windows[grabWindowID] = &window{
		xid:        grabWindowID,
		attributes: wire.WindowAttributes{EventMask: 0},
	}

	// Grab the pointer with ownerEvents = true
	server.pointerGrabWindow = grabWindowID
	server.pointerGrabEventMask = wire.ButtonPressMask
	server.pointerGrabOwner = true // Event should be sent to originalWindowID

	// Send a button press event to the original window
	server.SendMouseEvent(originalWindowID, "mousedown", 10, 20, (1<<16)|1)

	messages := drainMessages(t, clientBuffer, client.byteOrder)
	if len(messages) == 0 {
		t.Fatal("Expected event to be sent, but no message was recorded")
	}

	event := messages[0].(*wire.ButtonPressEvent)
	if event.Event != originalWindowID.local {
		t.Errorf("Expected event to be sent to original window %d, but it was sent to %d", originalWindowID.local, event.Event)
	}
}

func TestSendMouseEvent_ActivePointerGrab_MaskBlocked(t *testing.T) {
	server, clientBuffer := setupTestServer(t)
	client := server.clients[1]

	originalWindowID := xID{client: 1, local: 10}
	grabWindowID := xID{client: 1, local: 20}

	server.windows[originalWindowID] = &window{xid: originalWindowID}
	server.windows[grabWindowID] = &window{xid: grabWindowID}

	// Grab the pointer, but with a mask that doesn't include ButtonPress
	server.pointerGrabWindow = grabWindowID
	server.pointerGrabEventMask = wire.PointerMotionMask

	// Send a button press event
	server.SendMouseEvent(originalWindowID, "mousedown", 10, 20, (1<<16)|1)

	messages := drainMessages(t, clientBuffer, client.byteOrder)
	if len(messages) != 0 {
		t.Fatalf("Expected event to be blocked by grab event mask, but %d messages were sent", len(messages))
	}
}

func TestSendMouseEvent_PassiveGrab_Activates(t *testing.T) {
	server, _ := setupTestServer(t)

	windowID := xID{client: 1, local: 10}
	server.windows[windowID] = &window{
		xid:        windowID,
		attributes: wire.WindowAttributes{EventMask: wire.ButtonPressMask},
	}

	// Setup a passive grab on the window for Button 1
	server.passiveGrabs[windowID] = []*passiveGrab{
		{
			button:    1,
			modifiers: 0,
			owner:     false,
			eventMask: wire.ButtonPressMask | wire.ButtonReleaseMask,
		},
	}

	// Send a button press event that should activate the grab
	// state = 0, button = 1
	server.SendMouseEvent(windowID, "mousedown", 10, 20, 1)

	// Check that the pointer grab is now active
	if server.pointerGrabWindow != windowID {
		t.Errorf("Expected pointer grab to be activated on window %s, but it was not", windowID)
	}
	if server.pointerGrabEventMask != (wire.ButtonPressMask | wire.ButtonReleaseMask) {
		t.Errorf("Expected grab event mask to be %d, but got %d", (wire.ButtonPressMask | wire.ButtonReleaseMask), server.pointerGrabEventMask)
	}
}

func TestSendKeyboardEvent_ActiveKeyboardGrab_Redirected(t *testing.T) {
	server, clientBuffer := setupTestServer(t)
	client := server.clients[1]

	originalWindowID := xID{client: 1, local: 10}
	grabWindowID := xID{client: 1, local: 20}

	server.windows[originalWindowID] = &window{
		xid:        originalWindowID,
		attributes: wire.WindowAttributes{EventMask: wire.KeyPressMask},
	}
	server.windows[grabWindowID] = &window{
		xid:        grabWindowID,
		attributes: wire.WindowAttributes{EventMask: 0},
	}

	// Grab the keyboard on grabWindowID
	server.keyboardGrabWindow = grabWindowID
	server.keyboardGrabOwner = false // Event should be sent to grabWindowID

	// Send a key press event to the original window
	server.SendKeyboardEvent(originalWindowID, "keydown", "KeyA", false, false, false, false)

	messages := drainMessages(t, clientBuffer, client.byteOrder)
	if len(messages) == 0 {
		t.Fatal("Expected event to be sent, but no message was recorded")
	}

	// Verify the event was sent to the grab window
	event := messages[0].(*wire.KeyEvent)
	if event.Event != grabWindowID.local {
		t.Errorf("Expected event to be redirected to window %d, but it was sent to %d", grabWindowID.local, event.Event)
	}
}

func TestSendKeyboardEvent_ActiveKeyboardGrab_OwnerEventsTrue(t *testing.T) {
	server, clientBuffer := setupTestServer(t)
	client := server.clients[1]

	originalWindowID := xID{client: 1, local: 10}
	grabWindowID := xID{client: 1, local: 20}

	server.windows[originalWindowID] = &window{
		xid:        originalWindowID,
		attributes: wire.WindowAttributes{EventMask: wire.KeyPressMask},
	}
	server.windows[grabWindowID] = &window{
		xid:        grabWindowID,
		attributes: wire.WindowAttributes{EventMask: 0},
	}

	// Grab the keyboard with ownerEvents = true
	server.keyboardGrabWindow = grabWindowID
	server.keyboardGrabOwner = true // Event should be sent to originalWindowID

	// Send a key press event to the original window
	server.SendKeyboardEvent(originalWindowID, "keydown", "KeyA", false, false, false, false)

	messages := drainMessages(t, clientBuffer, client.byteOrder)
	if len(messages) == 0 {
		t.Fatal("Expected event to be sent, but no message was recorded")
	}

	event := messages[0].(*wire.KeyEvent)
	if event.Event != originalWindowID.local {
		t.Errorf("Expected event to be sent to original window %d, but it was sent to %d", originalWindowID.local, event.Event)
	}
}

func TestSendKeyboardEvent_EventMask_Sent(t *testing.T) {
	server, clientBuffer := setupTestServer(t)
	client := server.clients[1]

	windowID := xID{client: 1, local: 10}
	server.windows[windowID] = &window{
		xid:        windowID,
		attributes: wire.WindowAttributes{EventMask: wire.KeyPressMask},
	}

	// Send a key press event
	server.handleRequest(client, &wire.SetInputFocusRequest{Focus: 10}, 0)
	server.SendKeyboardEvent(windowID, "keydown", "KeyA", false, false, false, false)

	messages := drainMessages(t, clientBuffer, client.byteOrder)
	if len(messages) == 0 {
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
	reparentReq := &wire.ReparentWindowRequest{Window: wire.Window(childID.local), Parent: wire.Window(newParentID.local), X: 10, Y: 20}
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
	circulateReq := &wire.CirculateWindowRequest{Window: wire.Window(childID.local), Direction: 0 /* RaiseLowest */}
	server.handleRequest(client, circulateReq, 3)
	if len(mockFrontend.CirculateWindowCalls) != 1 {
		t.Errorf("CirculateWindow: expected frontend to be called")
	}

	// 4. Test QueryTree
	queryTreeReq := &wire.QueryTreeRequest{Window: wire.Window(newParentID.local)}
	reply := server.handleRequest(client, queryTreeReq, 4)
	if reply == nil {
		t.Fatalf("QueryTree: handleRequest returned a nil reply")
	}
	encodedReply := reply.EncodeMessage(client.byteOrder)
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
	destroyReq := &wire.DestroySubwindowsRequest{Window: wire.Window(newParentID.local)}
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
	server, clientBuffer := setupTestServer(t)
	client := server.clients[1]

	windowID := xID{client: 1, local: 10}
	server.windows[windowID] = &window{
		xid:        windowID,
		attributes: wire.WindowAttributes{EventMask: wire.KeyReleaseMask}, // Does not include KeyPressMask
	}

	// Send a key press event
	server.SendKeyboardEvent(windowID, "keydown", "KeyA", false, false, false, false)

	messages := drainMessages(t, clientBuffer, client.byteOrder)
	if len(messages) != 0 {
		t.Fatalf("Expected event to be blocked by event mask, but %d messages were sent", len(messages))
	}
}

func TestColormapRequests(t *testing.T) {
	server, _ := setupTestServer(t)
	client := server.clients[1]

	// Test CreateColormapRequest
	t.Run("CreateColormapRequest", func(t *testing.T) {
		cmapID := xID{client: client.id, local: 100}
		req := &wire.CreateColormapRequest{Mid: wire.Colormap(cmapID.local), Alloc: 0, Visual: 0}
		reply := server.handleRequest(client, req, 2)
		if reply != nil {
			t.Fatalf("CreateColormapRequest: expected nil reply for success, got %v", reply)
		}
		if _, ok := server.colormaps[cmapID]; !ok {
			t.Errorf("CreateColormapRequest: colormap %s not created", cmapID)
		}
	})

	// Test AllocColorRequest
	t.Run("AllocColorRequest", func(t *testing.T) {
		cmapID := xID{client: client.id, local: 100}
		req := &wire.AllocColorRequest{Cmap: wire.Colormap(cmapID.local), Red: 0x1000, Green: 0x2000, Blue: 0x3000}
		reply := server.handleRequest(client, req, 3)
		if reply == nil {
			t.Fatal("AllocColorRequest: expected reply, got nil")
		}
		allocReply, ok := reply.(*wire.AllocColorReply)
		if !ok {
			t.Fatalf("AllocColorRequest: expected *wire.AllocColorReply, got %T", reply)
		}
		expectedPixel := (uint32(0x10) << 16) | (uint32(0x20) << 8) | uint32(0x30)
		if allocReply.Pixel != expectedPixel {
			t.Errorf("AllocColorRequest: expected pixel %x, got %x", expectedPixel, allocReply.Pixel)
		}
		if _, ok := server.colormaps[cmapID].pixels[expectedPixel]; !ok {
			t.Errorf("AllocColorRequest: pixel %x not allocated in colormap %s", expectedPixel, cmapID)
		}

		// Test with default colormap
		reqDefault := &wire.AllocColorRequest{Cmap: wire.Colormap(server.defaultColormap), Red: 0x4000, Green: 0x5000, Blue: 0x6000}
		replyDefault := server.handleRequest(client, reqDefault, 4)
		allocReplyDefault, ok := replyDefault.(*wire.AllocColorReply)
		if !ok {
			t.Fatalf("AllocColorRequest (default): expected *wire.AllocColorReply, got %T", replyDefault)
		}
		expectedPixelDefault := (uint32(0x40) << 16) | (uint32(0x50) << 8) | uint32(0x60)
		if allocReplyDefault.Pixel != expectedPixelDefault {
			t.Errorf("AllocColorRequest (default): expected pixel %x, got %x", expectedPixelDefault, allocReplyDefault.Pixel)
		}
		if _, ok := server.colormaps[xID{local: server.defaultColormap}].pixels[expectedPixelDefault]; !ok {
			t.Errorf("AllocColorRequest (default): pixel %x not allocated in default colormap", expectedPixelDefault)
		}
	})

	// Test AllocNamedColorRequest
	t.Run("AllocNamedColorRequest", func(t *testing.T) {
		cmapID := xID{client: client.id, local: 100}
		req := &wire.AllocNamedColorRequest{Cmap: wire.Colormap(cmapID.local), Name: []byte("red")}
		reply := server.handleRequest(client, req, 5)
		if reply == nil {
			t.Fatal("AllocNamedColorRequest: expected reply, got nil")
		}
		allocReply, ok := reply.(*wire.AllocNamedColorReply)
		if !ok {
			t.Fatalf("AllocNamedColorRequest: expected *wire.AllocNamedColorReply, got %T", reply)
		}
		// "red" is 0xFF0000, scaled to 16-bit is 0xFFFF00000000
		if allocReply.Red != 0xFFFF || allocReply.Green != 0 || allocReply.Blue != 0 {
			t.Errorf("AllocNamedColorRequest: expected red, got R:%x G:%x B:%x", allocReply.Red, allocReply.Green, allocReply.Blue)
		}
		expectedPixel := (uint32(0xFF) << 16) | (uint32(0x00) << 8) | uint32(0x00)
		if _, ok := server.colormaps[cmapID].pixels[expectedPixel]; !ok {
			t.Errorf("AllocNamedColorRequest: pixel %x not allocated in colormap %s", expectedPixel, cmapID)
		}

		// Test with default colormap
		reqDefault := &wire.AllocNamedColorRequest{Cmap: wire.Colormap(server.defaultColormap), Name: []byte("blue")}
		replyDefault := server.handleRequest(client, reqDefault, 6)
		allocReplyDefault, ok := replyDefault.(*wire.AllocNamedColorReply)
		if !ok {
			t.Fatalf("AllocNamedColorRequest (default): expected *wire.AllocNamedColorReply, got %T", replyDefault)
		}
		// "blue" is 0x0000FF, scaled to 16-bit is 0x00000000FFFF
		if allocReplyDefault.Red != 0 || allocReplyDefault.Green != 0 || allocReplyDefault.Blue != 0xFFFF {
			t.Errorf("AllocNamedColorRequest (default): expected blue, got R:%x G:%x B:%x", allocReplyDefault.Red, allocReplyDefault.Green, allocReplyDefault.Blue)
		}
		expectedPixelDefault := (uint32(0x00) << 16) | (uint32(0x00) << 8) | uint32(0xFF)
		if _, ok := server.colormaps[xID{local: server.defaultColormap}].pixels[expectedPixelDefault]; !ok {
			t.Errorf("AllocNamedColorRequest (default): pixel %x not allocated in default colormap", expectedPixelDefault)
		}
	})

	// Test FreeColorsRequest
	t.Run("FreeColorsRequest", func(t *testing.T) {
		cmapID := xID{client: client.id, local: 100}
		// Allocate a color first
		allocReq := &wire.AllocColorRequest{Cmap: wire.Colormap(cmapID.local), Red: 0x1000, Green: 0x2000, Blue: 0x3000}
		allocReply := server.handleRequest(client, allocReq, 7).(*wire.AllocColorReply)
		pixelToFree := allocReply.Pixel

		req := &wire.FreeColorsRequest{Cmap: wire.Colormap(cmapID.local), Pixels: []uint32{pixelToFree}}
		reply := server.handleRequest(client, req, 8)
		if reply != nil {
			t.Fatalf("FreeColorsRequest: expected nil reply for success, got %v", reply)
		}
		if _, ok := server.colormaps[cmapID].pixels[pixelToFree]; ok {
			t.Errorf("FreeColorsRequest: pixel %x not freed from colormap %s", pixelToFree, cmapID)
		}
	})

	// Test StoreColorsRequest
	t.Run("StoreColorsRequest", func(t *testing.T) {
		cmapID := xID{client: client.id, local: 100}
		// Allocate a color first
		allocReq := &wire.AllocColorRequest{Cmap: wire.Colormap(cmapID.local), Red: 0x1000, Green: 0x2000, Blue: 0x3000}
		allocReply := server.handleRequest(client, allocReq, 9).(*wire.AllocColorReply)
		pixelToStore := allocReply.Pixel

		// Store new values for the pixel
		req := &wire.StoreColorsRequest{
			Cmap: wire.Colormap(cmapID.local),
			Items: []wire.XColorItem{
				{Pixel: pixelToStore, Red: 0xAAAA, Green: 0xBBBB, Blue: 0xCCCC, Flags: wire.DoRed | wire.DoGreen | wire.DoBlue},
			},
		}
		reply := server.handleRequest(client, req, 10)
		if reply != nil {
			t.Fatalf("StoreColorsRequest: expected nil reply for success, got %v", reply)
		}
		storedColor := server.colormaps[cmapID].pixels[pixelToStore]
		if storedColor.Red != 0xAAAA || storedColor.Green != 0xBBBB || storedColor.Blue != 0xCCCC {
			t.Errorf("StoreColorsRequest: expected stored color R:AAAA G:BBBB B:CCCC, got R:%x G:%x B:%x", storedColor.Red, storedColor.Green, storedColor.Blue)
		}
	})

	// Test StoreNamedColorRequest
	t.Run("StoreNamedColorRequest", func(t *testing.T) {
		cmapID := xID{client: client.id, local: 100}
		// Allocate a color first
		allocReq := &wire.AllocColorRequest{Cmap: wire.Colormap(cmapID.local), Red: 0x1000, Green: 0x2000, Blue: 0x3000}
		allocReply := server.handleRequest(client, allocReq, 11).(*wire.AllocColorReply)
		pixelToStore := allocReply.Pixel

		// Store a named color for the pixel
		req := &wire.StoreNamedColorRequest{
			Cmap:  wire.Colormap(cmapID.local),
			Pixel: pixelToStore,
			Name:  "green",
			Flags: wire.DoRed | wire.DoGreen | wire.DoBlue,
		}
		reply := server.handleRequest(client, req, 12)
		if reply != nil {
			t.Fatalf("StoreNamedColorRequest: expected nil reply for success, got %v", reply)
		}
		storedColor := server.colormaps[cmapID].pixels[pixelToStore]
		// "green" is 0x008000, scaled to 16-bit is 0x000080800000
		if storedColor.Red != 0 || storedColor.Green != 0x8080 || storedColor.Blue != 0 {
			t.Errorf("StoreNamedColorRequest: expected stored color R:0 G:8080 B:0, got R:%x G:%x B:%x", storedColor.Red, storedColor.Green, storedColor.Blue)
		}
	})

	// Test LookupColorRequest
	t.Run("LookupColorRequest", func(t *testing.T) {
		cmapID := xID{client: client.id, local: 100}
		req := &wire.LookupColorRequest{Cmap: wire.Colormap(cmapID.local), Name: "white"}
		reply := server.handleRequest(client, req, 13)
		if reply == nil {
			t.Fatal("LookupColorRequest: expected reply, got nil")
		}
		lookupReply, ok := reply.(*wire.LookupColorReply)
		if !ok {
			t.Fatalf("LookupColorRequest: expected *wire.LookupColorReply, got %T", reply)
		}
		// "white" is 0xFFFFFF, scaled to 16-bit is 0xFFFFFFFF
		if lookupReply.Red != 0xFFFF || lookupReply.Green != 0xFFFF || lookupReply.Blue != 0xFFFF {
			t.Errorf("LookupColorRequest: expected white, got R:%x G:%x B:%x", lookupReply.Red, lookupReply.Green, lookupReply.Blue)
		}
	})

	// Test GetRGBColor
	t.Run("GetRGBColor", func(t *testing.T) {
		cmapID := xID{client: client.id, local: 100}
		// Allocate a color first
		allocReq := &wire.AllocColorRequest{Cmap: wire.Colormap(cmapID.local), Red: 0x1234, Green: 0x5678, Blue: 0x9ABC}
		allocReply := server.handleRequest(client, allocReq, 14).(*wire.AllocColorReply)
		pixel := allocReply.Pixel

		r, g, b := server.GetRGBColor(cmapID, pixel)
		if r != 0x12 || g != 0x56 || b != 0x9A {
			t.Errorf("GetRGBColor: expected R:0x12 G:0x56 B:0x9A, got R:0x%x G:0x%x B:0x%x", r, g, b)
		}

		// Test with default colormap
		allocReqDefault := &wire.AllocColorRequest{Cmap: wire.Colormap(server.defaultColormap), Red: 0xDEFF, Green: 0xADBE, Blue: 0xEF00}
		allocReplyDefault := server.handleRequest(client, allocReqDefault, 15).(*wire.AllocColorReply)
		pixelDefault := allocReplyDefault.Pixel

		r, g, b = server.GetRGBColor(xID{local: server.defaultColormap}, pixelDefault)
		if r != 0xDE || g != 0xAD || b != 0xEF {
			t.Errorf("GetRGBColor (default): expected R:0xDE G:0xAD B:0xEF, got R:0x%x G:0x%x B:0x%x", r, g, b)
		}

		// Test non-existent pixel, should return RGB from pixel value
		r, g, b = server.GetRGBColor(cmapID, 0x112233)
		if r != 0x11 || g != 0x22 || b != 0x33 {
			t.Errorf("GetRGBColor (non-existent): expected R:0x11 G:0x22 B:0x33, got R:0x%x G:0x%x B:0x%x", r, g, b)
		}
	})

	// Test CopyColormapAndFreeRequest
	t.Run("CopyColormapAndFreeRequest", func(t *testing.T) {
		srcCmapID := xID{client: client.id, local: 200}
		newCmapID := xID{client: client.id, local: 201}

		// 1. Create source colormap and allocate a color
		server.colormaps[srcCmapID] = &colormap{pixels: make(map[uint32]wire.XColorItem)}
		color := wire.XColorItem{Pixel: 0xABCDEF, Red: 0xAAAA, Green: 0xBBBB, Blue: 0xCCCC, ClientID: client.id}
		server.colormaps[srcCmapID].pixels[color.Pixel] = color

		// 2. Send CopyColormapAndFree request
		req := &wire.CopyColormapAndFreeRequest{Mid: wire.Colormap(newCmapID.local), SrcCmap: wire.Colormap(srcCmapID.local)}
		reply := server.handleRequest(client, req, 16)
		if reply != nil {
			t.Fatalf("CopyColormapAndFreeRequest: expected nil reply for success, got %v", reply)
		}

		// 3. Verify new colormap and copied color
		newCmap, ok := server.colormaps[newCmapID]
		if !ok {
			t.Fatalf("CopyColormapAndFreeRequest: new colormap %s not created", newCmapID)
		}
		if _, ok := newCmap.pixels[color.Pixel]; !ok {
			t.Errorf("CopyColormapAndFreeRequest: color not copied to new colormap")
		}

		// 4. Verify color is freed from source colormap
		if _, ok := server.colormaps[srcCmapID].pixels[color.Pixel]; ok {
			t.Errorf("CopyColormapAndFreeRequest: color not freed from source colormap")
		}
	})
}

func TestKeyboardMappingRequests(t *testing.T) {
	server, clientBuffer := setupTestServer(t)
	client := server.clients[1]

	// 1. Test GetKeyboardMappingRequest
	getReq := &wire.GetKeyboardMappingRequest{FirstKeyCode: 10, Count: 2}
	reply := server.handleRequest(client, getReq, 2)
	if reply == nil {
		t.Fatal("GetKeyboardMapping: handleRequest returned a nil reply")
	}
	encodedReply := reply.EncodeMessage(client.byteOrder)
	if _, err := clientBuffer.Write(encodedReply); err != nil {
		t.Fatalf("GetKeyboardMapping: failed to write reply to buffer: %v", err)
	}

	var getHeader struct {
		ReplyType         byte
		KeySymsPerKeycode byte
		Sequence          uint16
		Length            uint32
		Padding           [24]byte
	}
	if err := binary.Read(clientBuffer, binary.LittleEndian, &getHeader); err != nil {
		t.Fatalf("GetKeyboardMapping: failed to read reply header: %v", err)
	}
	keySyms := make([]uint32, getHeader.Length)
	if err := binary.Read(clientBuffer, binary.LittleEndian, &keySyms); err != nil {
		t.Fatalf("GetKeyboardMapping: failed to read reply keysyms: %v", err)
	}

	expectedKeySyms := []uint32{0x0031, 0x0032} // XK_1, XK_2
	if len(keySyms) != 2 || keySyms[0] != expectedKeySyms[0] || keySyms[1] != expectedKeySyms[1] {
		t.Errorf("GetKeyboardMapping: incorrect keysyms returned. Expected %v, got %v", expectedKeySyms, keySyms)
	}

	// 2. Test ChangeKeyboardMappingRequest
	changeReq := &wire.ChangeKeyboardMappingRequest{
		KeyCodeCount:      1,
		FirstKeyCode:      10,
		KeySymsPerKeyCode: 1,
		KeySyms:           []uint32{0x0041}, // XK_A
	}
	server.handleRequest(client, changeReq, 3)

	if server.keymap[10] != 0x0041 {
		t.Errorf("ChangeKeyboardMapping: keymap was not updated. Expected %x, got %x", 0x0041, server.keymap[10])
	}
}

func TestExtensionRequests(t *testing.T) {
	server, _ := setupTestServer(t)
	client := server.clients[1]

	// 1. Test QueryExtension for a known extension (BIG-REQUESTS)
	queryBigReq := &wire.QueryExtensionRequest{Name: wire.BigRequestsExtensionName}
	reply := server.handleRequest(client, queryBigReq, 2)
	if reply == nil {
		t.Fatal("QueryExtension (BIG-REQUESTS): handleRequest returned a nil reply")
	}
	queryBigReply, ok := reply.(*wire.QueryExtensionReply)
	if !ok {
		t.Fatalf("QueryExtension (BIG-REQUESTS): expected *wire.QueryExtensionReply, got %T", reply)
	}
	if !queryBigReply.Present || queryBigReply.MajorOpcode != byte(wire.BigRequestsOpcode) {
		t.Errorf("QueryExtension (BIG-REQUESTS): incorrect reply. Got present=%t, opcode=%d", queryBigReply.Present, queryBigReply.MajorOpcode)
	}

	// 2. Test QueryExtension for the new XInput extension
	queryXInputReq := &wire.QueryExtensionRequest{Name: wire.XInputExtensionName}
	reply = server.handleRequest(client, queryXInputReq, 3)
	if reply == nil {
		t.Fatal("QueryExtension (XInput): handleRequest returned a nil reply")
	}
	queryXInputReply, ok := reply.(*wire.QueryExtensionReply)
	if !ok {
		t.Fatalf("QueryExtension (XInput): expected *wire.QueryExtensionReply, got %T", reply)
	}
	if !queryXInputReply.Present || queryXInputReply.MajorOpcode != byte(wire.XInputOpcode) {
		t.Errorf("QueryExtension (XInput): incorrect reply. Got present=%t, opcode=%d", queryXInputReply.Present, queryXInputReply.MajorOpcode)
	}

	// 3. Test ListExtensions
	listReq := &wire.ListExtensionsRequest{}
	reply = server.handleRequest(client, listReq, 4)
	if reply == nil {
		t.Fatal("ListExtensions: handleRequest returned a nil reply")
	}
	listReply, ok := reply.(*wire.ListExtensionsReply)
	if !ok {
		t.Fatalf("ListExtensions: expected *wire.ListExtensionsReply, got %T", reply)
	}
	if len(listReply.Names) != 2 || listReply.Names[0] != wire.BigRequestsExtensionName || listReply.Names[1] != wire.XInputExtensionName {
		t.Errorf("ListExtensions: incorrect extension list. Got %v", listReply.Names)
	}

	// 4. Test XListInputDevices
	xinputReq := &wire.XInputRequest{MinorOpcode: wire.XListInputDevices}
	reply = server.handleRequest(client, xinputReq, 5)
	if reply == nil {
		t.Fatal("XListInputDevices: handleRequest returned a nil reply")
	}
	listInputDevicesReply, ok := reply.(*wire.ListInputDevicesReply)
	if !ok {
		t.Fatalf("XListInputDevices: expected *wire.ListInputDevicesReply, got %T", reply)
	}
	if len(listInputDevicesReply.Devices) != 2 {
		t.Errorf("XListInputDevices: incorrect number of devices. Expected 2, got %d", len(listInputDevicesReply.Devices))
	}
}
