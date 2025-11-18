//go:build x11 && !wasm

package x11

import (
	"bytes"
	"encoding/binary"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/c2FmZQ/sshterm/internal/x11/wire"
)

func TestGetPointerMappingRequest(t *testing.T) {
	server, client, _, clientBuffer := setupTestServerWithClient(t)
	mockConn := client.conn.(*testConn)

	// 1. Send a GetPointerMapping request
	reqBuf := new(bytes.Buffer)
	binary.Write(reqBuf, client.byteOrder, uint8(wire.GetPointerMapping))
	binary.Write(reqBuf, client.byteOrder, byte(0))
	binary.Write(reqBuf, client.byteOrder, uint16(1))
	binary.Write(reqBuf, client.byteOrder, uint32(0))
	mockConn.r = reqBuf

	req, seq, err := server.readRequest(client)
	if err != nil {
		t.Fatalf("readRequest for GetPointerMapping failed: %v", err)
	}
	reply := server.handleRequest(client, req, seq)

	// 2. Verify the reply
	encodedReply := reply.EncodeMessage(client.byteOrder)
	clientBuffer.Write(encodedReply)

	replyBytes := clientBuffer.Bytes()
	if len(replyBytes) < 32 {
		t.Fatalf("Expected reply of at least 32 bytes, got %d", len(replyBytes))
	}

	replyType := replyBytes[0]
	if replyType != 1 {
		t.Errorf("Expected ReplyType 1, got %d", replyType)
	}

	replySeq := client.byteOrder.Uint16(replyBytes[2:4])
	if replySeq != seq {
		t.Errorf("Expected Sequence %d, got %d", seq, replySeq)
	}

	length := replyBytes[1]
	if length != 3 {
		t.Errorf("Expected length to be 3, got %d", length)
	}

	expectedMap := []byte{1, 2, 3}
	pointerMap := replyBytes[32 : 32+length]
	assert.Equal(t, expectedMap, pointerMap, "pointer map mismatch")
}

func TestGetPointerControlRequest(t *testing.T) {
	server, client, _, clientBuffer := setupTestServerWithClient(t)
	mockConn := client.conn.(*testConn)

	// 1. Send a GetPointerControl request
	reqBuf := new(bytes.Buffer)
	binary.Write(reqBuf, client.byteOrder, uint8(wire.GetPointerControl))
	binary.Write(reqBuf, client.byteOrder, byte(0))
	binary.Write(reqBuf, client.byteOrder, uint16(1))
	// Pad to 4 bytes
	binary.Write(reqBuf, client.byteOrder, uint16(0))
	mockConn.r = reqBuf

	req, seq, err := server.readRequest(client)
	if err != nil {
		t.Fatalf("readRequest for GetPointerControl failed: %v", err)
	}
	reply := server.handleRequest(client, req, seq)

	// 2. Verify the reply
	encodedReply := reply.EncodeMessage(client.byteOrder)
	clientBuffer.Write(encodedReply)

	replyBytes := clientBuffer.Bytes()
	if len(replyBytes) < 32 {
		t.Fatalf("Expected reply of at least 32 bytes, got %d", len(replyBytes))
	}

	// Manual parsing to avoid reflection on unexported fields
	replyType := replyBytes[0]
	replySeq := client.byteOrder.Uint16(replyBytes[2:4])
	accelNumerator := client.byteOrder.Uint16(replyBytes[8:10])
	accelDenominator := client.byteOrder.Uint16(replyBytes[10:12])
	threshold := client.byteOrder.Uint16(replyBytes[12:14])

	if replyType != 1 {
		t.Errorf("Expected ReplyType 1, got %d", replyType)
	}
	if replySeq != seq {
		t.Errorf("Expected Sequence %d, got %d", seq, replySeq)
	}
	if accelNumerator != 1 {
		t.Errorf("Expected accelNumerator to be 1, got %d", accelNumerator)
	}
	if accelDenominator != 1 {
		t.Errorf("Expected accelDenominator to be 1, got %d", accelDenominator)
	}
	if threshold != 1 {
		t.Errorf("Expected threshold to be 1, got %d", threshold)
	}
}

func TestSetModifierMappingRequest(t *testing.T) {
	server, client, mockFrontend, _ := setupTestServerWithClient(t)
	mockConn := client.conn.(*testConn)

	// 1. Send a SetModifierMapping request
	const keyCodesPerModifier = 2
	keyCodes := make([]wire.KeyCode, 8*keyCodesPerModifier)
	for i := range keyCodes {
		keyCodes[i] = wire.KeyCode(i + 1)
	}

	reqBuf := new(bytes.Buffer)
	binary.Write(reqBuf, client.byteOrder, uint8(wire.SetModifierMapping))
	binary.Write(reqBuf, client.byteOrder, byte(keyCodesPerModifier))
	binary.Write(reqBuf, client.byteOrder, uint16(1+2*keyCodesPerModifier)) // request length
	for _, kc := range keyCodes {
		binary.Write(reqBuf, client.byteOrder, kc)
	}
	mockConn.r = reqBuf

	req, seq, err := server.readRequest(client)
	if err != nil {
		t.Fatalf("readRequest for SetModifierMapping failed: %v", err)
	}
	server.handleRequest(client, req, seq)

	// 2. Verify the mock frontend was updated
	assert.Equal(t, keyCodes, mockFrontend.modifierMap, "modifierMap mismatch")
}

func TestGetModifierMappingRequest(t *testing.T) {
	server, client, _, clientBuffer := setupTestServerWithClient(t)
	mockConn := client.conn.(*testConn)

	// 1. Send a GetModifierMapping request
	reqBuf := new(bytes.Buffer)
	binary.Write(reqBuf, client.byteOrder, uint8(wire.GetModifierMapping))
	binary.Write(reqBuf, client.byteOrder, byte(0))
	binary.Write(reqBuf, client.byteOrder, uint16(1))
	binary.Write(reqBuf, client.byteOrder, uint32(0))
	mockConn.r = reqBuf

	req, seq, err := server.readRequest(client)
	if err != nil {
		t.Fatalf("readRequest for GetModifierMapping failed: %v", err)
	}
	reply := server.handleRequest(client, req, seq)

	// 2. Verify the reply
	encodedReply := reply.EncodeMessage(client.byteOrder)
	clientBuffer.Write(encodedReply)

	replyBytes := clientBuffer.Bytes()
	if len(replyBytes) < 32 {
		t.Fatalf("Expected reply of at least 32 bytes, got %d", len(replyBytes))
	}

	replyType := replyBytes[0]
	if replyType != 1 {
		t.Errorf("Expected ReplyType 1, got %d", replyType)
	}

	replySeq := client.byteOrder.Uint16(replyBytes[2:4])
	if replySeq != seq {
		t.Errorf("Expected Sequence %d, got %d", seq, replySeq)
	}

	keyCodesPerModifier := replyBytes[1]
	if keyCodesPerModifier != 1 {
		t.Errorf("Expected keyCodesPerModifier to be 1, got %d", keyCodesPerModifier)
	}

	expectedMap := make([]byte, 8)
	keyCodes := replyBytes[32 : 32+len(expectedMap)]
	assert.Equal(t, expectedMap, keyCodes, "keyCodes mismatch")
}

func TestSetPointerMappingRequest(t *testing.T) {
	server, client, mockFrontend, clientBuffer := setupTestServerWithClient(t)

	// 1. Define the pointer map and construct the raw request bytes
	pointerMap := []byte{3, 1, 2}
	n := len(pointerMap)
	paddedLen := n + wire.PadLen(n)
	reqLen := 1 + (paddedLen / 4) // Request length in 4-byte units

	rawReq := new(bytes.Buffer)
	binary.Write(rawReq, client.byteOrder, uint8(wire.SetPointerMapping)) // Opcode
	binary.Write(rawReq, client.byteOrder, uint8(n))                      // length of map
	binary.Write(rawReq, client.byteOrder, uint16(reqLen))                // request length
	rawReq.Write(pointerMap)
	rawReq.Write(make([]byte, wire.PadLen(n))) // padding

	// The mock connection needs the raw request to be read
	mockConn := client.conn.(*testConn)
	mockConn.r = rawReq

	// 2. Read and handle the request
	req, seq, err := server.readRequest(client)
	if err != nil {
		t.Fatalf("readRequest failed: %v", err)
	}
	reply := server.handleRequest(client, req, seq)
	if reply == nil {
		t.Fatalf("handleRequest returned a nil reply")
	}

	// 3. Verify the frontend was called correctly
	if len(mockFrontend.SetPointerMappingCalls) != 1 {
		t.Fatalf("Expected SetPointerMapping to be called once, but it was called %d times", len(mockFrontend.SetPointerMappingCalls))
	}
	if !bytes.Equal(mockFrontend.SetPointerMappingCalls[0], pointerMap) {
		t.Errorf("Expected SetPointerMapping to be called with %v, but got %v", pointerMap, mockFrontend.SetPointerMappingCalls[0])
	}

	// 4. Verify the reply sent to the client
	encodedReply := reply.EncodeMessage(client.byteOrder)
	if _, err := clientBuffer.Write(encodedReply); err != nil {
		t.Fatalf("Failed to write reply to buffer: %v", err)
	}

	// We cannot use binary.Read directly on the struct because it is not fixed size.
	// We read the fields manually.
	replyBytes := clientBuffer.Bytes()
	if len(replyBytes) < 32 {
		t.Fatalf("Expected reply of at least 32 bytes, got %d", len(replyBytes))
	}

	replyType := replyBytes[0]
	status := replyBytes[1]
	replySeq := client.byteOrder.Uint16(replyBytes[2:4])

	if replyType != 1 {
		t.Errorf("Expected ReplyType 1, got %d", replyType)
	}
	if replySeq != seq {
		t.Errorf("Expected Sequence %d, got %d", seq, replySeq)
	}
	if status != 0 { // Success
		t.Errorf("Expected Status 0 (Success), got %d", status)
	}
}
