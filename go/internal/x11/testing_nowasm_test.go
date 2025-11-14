//go:build x11 && !wasm

package x11

import (
	"bytes"
	"encoding/binary"
	"testing"
)

// setupTestServerWithClient creates a new x11Server with a mock frontend and a single mock client.
// It returns the server instance, the client, the mock frontend, and a buffer that captures all data sent to the mock client.
func setupTestServerWithClient(t *testing.T) (*x11Server, *x11Client, *MockX11Frontend, *bytes.Buffer) {
	clientBuffer := &bytes.Buffer{}
	mockConn := &testConn{r: &bytes.Buffer{}, w: clientBuffer}

	client := &x11Client{
		id:          1,
		conn:        mockConn,
		byteOrder:   binary.LittleEndian,
		sequence:    0, // Will be incremented to 1 by readRequest
		openDevices: make(map[byte]*deviceInfo),
	}

	mockFrontend := &MockX11Frontend{}
	server := &x11Server{
		logger:          &testLogger{t: t},
		windows:         make(map[xID]*window),
		clients:         map[uint32]*x11Client{1: client},
		frontend:        mockFrontend,
		byteOrder:       binary.LittleEndian,
		passiveGrabs:    make(map[xID][]*passiveGrab),
		selections:      make(map[xID]uint32),
		keymap:          make(map[byte]uint32),
		colormaps:       make(map[xID]*colormap),
		defaultColormap: 1,
	}
	server.colormaps[xID{local: server.defaultColormap}] = &colormap{pixels: make(map[uint32]xColorItem)}
	for k, v := range KeyCodeToKeysym {
		server.keymap[k] = v
	}

	t.Cleanup(func() {
		x11ServerInstance = nil
	})

	return server, client, mockFrontend, clientBuffer
}
