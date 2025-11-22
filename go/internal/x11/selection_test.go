//go:build x11 && !wasm

package x11

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/c2FmZQ/sshterm/internal/x11/wire"
)

func TestSetSelectionOwnerRequest(t *testing.T) {
	server, client, _, _ := setupTestServerWithClient(t)
	mockConn := client.conn.(*testConn)

	// 1. Send a SetSelectionOwner request
	selection := wire.Atom(1)
	owner := wire.Window(2)
	r := &wire.SetSelectionOwnerRequest{
		Owner:     owner,
		Selection: selection,
		Time:      0,
	}
	mockConn.r = bytes.NewBuffer(r.EncodeMessage(client.byteOrder))

	req, seq, err := server.readRequest(client)
	if err != nil {
		t.Fatalf("readRequest for SetSelectionOwner failed: %v", err)
	}
	server.handleRequest(client, req, seq)

	// 2. Verify the selection owner was set
	assert.Equal(t, uint32(owner), server.selections[client.xID(uint32(selection))], "selection owner mismatch")
}

func TestGetSelectionOwnerRequest(t *testing.T) {
	server, client, _, clientBuffer := setupTestServerWithClient(t)
	mockConn := client.conn.(*testConn)

	// 1. Set a selection owner
	selection := wire.Atom(1)
	owner := wire.Window(2)
	server.selections[client.xID(uint32(selection))] = uint32(owner)

	// 2. Send a GetSelectionOwner request
	r := &wire.GetSelectionOwnerRequest{
		Selection: selection,
	}
	mockConn.r = bytes.NewBuffer(r.EncodeMessage(client.byteOrder))

	req, seq, err := server.readRequest(client)
	if err != nil {
		t.Fatalf("readRequest for GetSelectionOwner failed: %v", err)
	}
	reply := server.handleRequest(client, req, seq)

	// 3. Verify the reply
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

	returnedOwner := client.byteOrder.Uint32(replyBytes[8:12])
	assert.Equal(t, uint32(owner), returnedOwner, "selection owner mismatch")
}

func TestConvertSelectionRequest(t *testing.T) {
	server, client, mockFrontend, _ := setupTestServerWithClient(t)
	mockConn := client.conn.(*testConn)

	// 1. Send a ConvertSelection request
	selection := wire.Atom(1)
	target := wire.Atom(2)
	property := wire.Atom(3)
	requestor := wire.Window(4)
	r := &wire.ConvertSelectionRequest{
		Requestor: requestor,
		Selection: selection,
		Target:    target,
		Property:  property,
		Time:      0,
	}
	mockConn.r = bytes.NewBuffer(r.EncodeMessage(client.byteOrder))

	req, seq, err := server.readRequest(client)
	if err != nil {
		t.Fatalf("readRequest for ConvertSelection failed: %v", err)
	}
	server.handleRequest(client, req, seq)

	// 2. Verify the mock frontend was called correctly
	assert.Equal(t, uint32(selection), mockFrontend.ConvertSelectionCalls[0].selection, "selection mismatch")
	assert.Equal(t, uint32(target), mockFrontend.ConvertSelectionCalls[0].target, "target mismatch")
	assert.Equal(t, uint32(property), mockFrontend.ConvertSelectionCalls[0].property, "property mismatch")
	assert.Equal(t, client.xID(uint32(requestor)), mockFrontend.ConvertSelectionCalls[0].requestor, "requestor mismatch")
}
