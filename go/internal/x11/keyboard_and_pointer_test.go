//go:build x11 && !wasm

package x11

import (
	"bytes"
	"encoding/binary"
	"testing"

	"github.com/stretchr/testify/assert"
)

// setupTestServerWithClient creates a new x11Server with a mock frontend and a single mock client.
// It returns the server instance, the client, the mock frontend, and a buffer that captures all data sent to the mock client.
func setupTestServerWithClient(t *testing.T) (*x11Server, *x11Client, *MockX11Frontend, *bytes.Buffer) {
	clientBuffer := &bytes.Buffer{}
	mockConn := &testConn{r: &bytes.Buffer{}, w: clientBuffer}

	client := &x11Client{
		id:        1,
		conn:      mockConn,
		byteOrder: binary.LittleEndian,
		sequence:  0, // Will be incremented to 1 by readRequest
	}

	mockFrontend := &MockX11Frontend{}
	server := &x11Server{
		logger:       &testLogger{t: t},
		windows:      make(map[xID]*window),
		clients:      map[uint32]*x11Client{1: client},
		frontend:     mockFrontend,
		byteOrder:    binary.LittleEndian,
		passiveGrabs: make(map[xID][]*passiveGrab),
		selections:   make(map[xID]uint32),
		keymap:       make(map[byte]uint32),
	}
	for k, v := range KeyCodeToKeysym {
		server.keymap[k] = v
	}

	return server, client, mockFrontend, clientBuffer
}

func TestGetKeyboardMappingRequest(t *testing.T) {
	server, client, _, clientBuffer := setupTestServerWithClient(t)
	// 1. Construct the raw request bytes
	firstKeyCode := KeyCode(10)
	count := byte(3)

	rawReq := new(bytes.Buffer)
	binary.Write(rawReq, client.byteOrder, uint8(GetKeyboardMapping)) // Opcode
	binary.Write(rawReq, client.byteOrder, byte(0))                   // Unused
	binary.Write(rawReq, client.byteOrder, uint16(2))                 // request length
	binary.Write(rawReq, client.byteOrder, firstKeyCode)
	binary.Write(rawReq, client.byteOrder, count)
	binary.Write(rawReq, client.byteOrder, uint16(0)) // Unused

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

	// 3. Verify the reply sent to the client
	encodedReply := reply.encodeMessage(client.byteOrder)
	if _, err := clientBuffer.Write(encodedReply); err != nil {
		t.Fatalf("Failed to write reply to buffer: %v", err)
	}

	// 4. Decode the reply and verify its contents
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

	keysymsPerKeycode := replyBytes[1]
	if keysymsPerKeycode != 1 {
		t.Errorf("Expected keysymsPerKeycode to be 1, got %d", keysymsPerKeycode)
	}

	length := client.byteOrder.Uint32(replyBytes[4:8])
	expectedLength := uint32(count) * uint32(keysymsPerKeycode)
	if length != expectedLength {
		t.Errorf("Expected reply length %d, got %d", expectedLength, length)
	}

	expectedKeysyms := []uint32{KeyCodeToKeysym[10], KeyCodeToKeysym[11], KeyCodeToKeysym[12]}
	keysyms := make([]uint32, count)
	for i := 0; i < int(count); i++ {
		keysyms[i] = client.byteOrder.Uint32(replyBytes[32+i*4 : 32+(i+1)*4])
	}
	assert.Equal(t, expectedKeysyms, keysyms, "keysyms mismatch")
}

func TestChangeKeyboardMappingRequest(t *testing.T) {
	// Make a copy of the original keymap to restore it later
	originalKeymap := make(map[byte]uint32)
	for k, v := range KeyCodeToKeysym {
		originalKeymap[k] = v
	}
	t.Cleanup(func() {
		KeyCodeToKeysym = originalKeymap
	})

	server, client, _, clientBuffer := setupTestServerWithClient(t)
	mockConn := client.conn.(*testConn)

	// 1. Send a ChangeKeyboardMapping request
	firstKeyCode := KeyCode(10)
	keySyms := []uint32{0x0031, 0x0032, 0x0033} // '1', '2', '3'
	keyCodeCount := byte(len(keySyms))
	keySymsPerKeyCode := byte(1)

	changeReqBuf := new(bytes.Buffer)
	binary.Write(changeReqBuf, client.byteOrder, uint8(ChangeKeyboardMapping))
	binary.Write(changeReqBuf, client.byteOrder, keyCodeCount)
	binary.Write(changeReqBuf, client.byteOrder, uint16(2+len(keySyms))) // request length
	binary.Write(changeReqBuf, client.byteOrder, firstKeyCode)
	binary.Write(changeReqBuf, client.byteOrder, keySymsPerKeyCode)
	binary.Write(changeReqBuf, client.byteOrder, uint16(0)) // unused
	for _, sym := range keySyms {
		binary.Write(changeReqBuf, client.byteOrder, sym)
	}
	mockConn.r = changeReqBuf

	changeReq, seq, err := server.readRequest(client)
	if err != nil {
		t.Fatalf("readRequest for ChangeKeyboardMapping failed: %v", err)
	}
	server.handleRequest(client, changeReq, seq)

	// 2. Send a GetKeyboardMapping request
	getReqBuf := new(bytes.Buffer)
	binary.Write(getReqBuf, client.byteOrder, uint8(GetKeyboardMapping))
	binary.Write(getReqBuf, client.byteOrder, byte(0))
	binary.Write(getReqBuf, client.byteOrder, uint16(2))
	binary.Write(getReqBuf, client.byteOrder, firstKeyCode)
	binary.Write(getReqBuf, client.byteOrder, keyCodeCount)
	binary.Write(getReqBuf, client.byteOrder, uint16(0))
	mockConn.r = getReqBuf

	getReq, seq, err := server.readRequest(client)
	if err != nil {
		t.Fatalf("readRequest for GetKeyboardMapping failed: %v", err)
	}
	reply := server.handleRequest(client, getReq, seq)

	// 3. Verify the reply
	encodedReply := reply.encodeMessage(client.byteOrder)
	clientBuffer.Write(encodedReply)

	replyBytes := clientBuffer.Bytes()
	if len(replyBytes) < 32+len(keySyms)*4 {
		t.Fatalf("Expected reply of at least %d bytes, got %d", 32+len(keySyms)*4, len(replyBytes))
	}
	returnedKeysyms := make([]uint32, len(keySyms))
	for i := range returnedKeysyms {
		returnedKeysyms[i] = client.byteOrder.Uint32(replyBytes[32+i*4 : 32+(i+1)*4])
	}
	assert.Equal(t, keySyms, returnedKeysyms, "returned keysyms should match the ones set")
}

func TestGetPointerMappingRequest(t *testing.T) {
	server, client, _, clientBuffer := setupTestServerWithClient(t)
	mockConn := client.conn.(*testConn)

	// 1. Send a GetPointerMapping request
	reqBuf := new(bytes.Buffer)
	binary.Write(reqBuf, client.byteOrder, uint8(GetPointerMapping))
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
	encodedReply := reply.encodeMessage(client.byteOrder)
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
	binary.Write(reqBuf, client.byteOrder, uint8(GetPointerControl))
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
	encodedReply := reply.encodeMessage(client.byteOrder)
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
	keyCodes := make([]KeyCode, 8*keyCodesPerModifier)
	for i := range keyCodes {
		keyCodes[i] = KeyCode(i + 1)
	}

	reqBuf := new(bytes.Buffer)
	binary.Write(reqBuf, client.byteOrder, uint8(SetModifierMapping))
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
	binary.Write(reqBuf, client.byteOrder, uint8(GetModifierMapping))
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
	encodedReply := reply.encodeMessage(client.byteOrder)
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
	paddedLen := n + padLen(n)
	reqLen := 1 + (paddedLen / 4) // Request length in 4-byte units

	rawReq := new(bytes.Buffer)
	binary.Write(rawReq, client.byteOrder, uint8(SetPointerMapping)) // Opcode
	binary.Write(rawReq, client.byteOrder, uint8(n))                 // length of map
	binary.Write(rawReq, client.byteOrder, uint16(reqLen))           // request length
	rawReq.Write(pointerMap)
	rawReq.Write(make([]byte, padLen(n))) // padding

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
	encodedReply := reply.encodeMessage(client.byteOrder)
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
