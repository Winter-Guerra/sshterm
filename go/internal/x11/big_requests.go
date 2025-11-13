//go:build x11

package x11

import "encoding/binary"

const (
	BigRequestsExtensionName = "BIG-REQUESTS"
	bigRequestsOpcode        = 133
)

type EnableBigRequestsRequest struct{}

func (r *EnableBigRequestsRequest) OpCode() reqCode {
	return bigRequestsOpcode
}

func parseEnableBigRequestsRequest(order binary.ByteOrder, raw []byte, seq uint16) (*EnableBigRequestsRequest, error) {
	return &EnableBigRequestsRequest{}, nil
}

type BigRequestsEnableReply struct {
	sequence         uint16
	maxRequestLength uint32
}

func (r *BigRequestsEnableReply) encodeMessage(order binary.ByteOrder) []byte {
	reply := make([]byte, 32)
	reply[0] = 1 // Reply
	order.PutUint16(reply[2:4], r.sequence)
	order.PutUint32(reply[8:12], r.maxRequestLength)
	return reply
}
