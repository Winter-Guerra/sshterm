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
