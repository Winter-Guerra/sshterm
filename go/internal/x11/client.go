//go:build x11

package x11

import (
	"encoding/binary"
	"fmt"
	"io"
	"testing"
)

// messageEncoder is an interface for types that can encode themselves into a byte slice.
type messageEncoder interface {
	encodeMessage(order binary.ByteOrder) []byte
}

// x11Error implements messageEncoder for X11 errors.
type x11Error struct {
	data []byte
}

func (e *x11Error) encodeMessage(order binary.ByteOrder) []byte {
	// The error data is already encoded with the correct byte order.
	return e.data
}

type x11Client struct {
	id           uint32
	conn         io.ReadWriteCloser
	sequence     uint16
	byteOrder    binary.ByteOrder
	sentMessages []messageEncoder
}

func (c *x11Client) xID(local uint32) xID {
	return xID{c.id, local}
}

func (c *x11Client) sendError(err XError) messageEncoder {
	reply := make([]byte, 32)
	reply[0] = 0 // Error code
	reply[1] = err.Code()
	c.byteOrder.PutUint16(reply[2:4], err.Sequence())
	c.byteOrder.PutUint32(reply[4:8], err.BadValue())
	c.byteOrder.PutUint16(reply[8:10], uint16(err.MinorOp()))
	reply[10] = byte(err.MajorOp())
	return &x11Error{data: reply}
}

// send sends a message to the client.
func (c *x11Client) send(m messageEncoder) error {
	// For testing, we can append the message to a slice.
	if testing.Testing() {
		c.sentMessages = append(c.sentMessages, m)
		return nil
	}
	encodedMsg := m.encodeMessage(c.byteOrder)
	debugf("X11DEBUG: client.send(%#v) encoded: %x", m, encodedMsg)
	if _, err := c.conn.Write(encodedMsg); err != nil {
		return fmt.Errorf("failed to write message to client: %w", err)
	}
	return nil
}
