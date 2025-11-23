//go:build x11 && !wasm

package x11

import (
	"testing"

	"github.com/c2FmZQ/sshterm/internal/x11/wire"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCore_PassiveKeyGrab_DeliveredToGrabberNotOwner(t *testing.T) {
	s, client1, _, client1Buffer := setupTestServerWithClient(t)
	client2, client2Buffer := newClient(t, s)

	// Window 1 is owned by Client 1
	window1 := client1.xID(1)
	s.handleRequest(client1, &wire.CreateWindowRequest{
		Drawable:  wire.Window(window1.local),
		ValueMask: wire.CWEventMask,
		Values:    wire.WindowAttributes{EventMask: wire.KeyPressMask},
	}, 1)
	s.inputFocus = window1

	// Client 2 grabs 'A' on Window 1
	s.handleRequest(client2, &wire.GrabKeyRequest{
		GrabWindow: wire.Window(window1.local),
		Key:        wire.KeyCode(jsCodeToX11Keycode["KeyA"]),
		Modifiers:  wire.AnyModifier,
	}, 1)

	// Send key press 'A' to Window 1, which should activate the grab
	s.SendKeyboardEvent(window1, "keydown", "KeyA", false, false, false, false)

	// Assert Client 2 (the grabber) receives the event
	require.NotEmpty(t, client2Buffer.Bytes(), "client 2 (grabber) should receive a key press event")
	msgBytes := client2Buffer.Bytes()
	assert.Equal(t, byte(wire.KeyPress), msgBytes[0], "event should be a key press")

	// Assert Client 1 (the owner) does NOT receive the event
	assert.Empty(t, client1Buffer.Bytes(), "client 1 (owner) should not receive the event")
}
