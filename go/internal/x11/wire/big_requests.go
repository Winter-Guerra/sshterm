//go:build x11

package wire

import "encoding/binary"

const (
	BigRequestsExtensionName = "BIG-REQUESTS"
)

type EnableBigRequestsRequest struct{}

func (r *EnableBigRequestsRequest) OpCode() ReqCode {
	return BigRequestsOpcode
}

func ParseEnableBigRequestsRequest(order binary.ByteOrder, raw []byte, seq uint16) (*EnableBigRequestsRequest, error) {
	return &EnableBigRequestsRequest{}, nil
}

type BigRequestsEnableReply struct {
	Sequence         uint16
	MaxRequestLength uint32
}

func (r *BigRequestsEnableReply) EncodeMessage(order binary.ByteOrder) []byte {
	reply := make([]byte, 32)
	reply[0] = 1 // Reply
	order.PutUint16(reply[2:4], r.Sequence)
	order.PutUint32(reply[8:12], r.MaxRequestLength)
	return reply
}
