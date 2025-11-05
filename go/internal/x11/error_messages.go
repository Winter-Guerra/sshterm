//go:build x11

package x11

import (
	"encoding/binary"
)

// ColormapError: 12
type BadColor struct {
	seq      uint16
	badValue uint32
	minorOp  byte
	majorOp  reqCode
}

func (e BadColor) Code() byte       { return ColormapError }
func (e BadColor) Sequence() uint16 { return e.seq }
func (e BadColor) BadValue() uint32 { return e.badValue }
func (e BadColor) MinorOp() byte    { return e.minorOp }
func (e BadColor) MajorOp() byte    { return byte(e.majorOp) }

func (e *BadColor) encodeMessage(order binary.ByteOrder) []byte {
	reply := make([]byte, 32)
	reply[0] = 0 // Error type
	reply[1] = e.Code()
	order.PutUint16(reply[2:4], e.Sequence())
	order.PutUint32(reply[4:8], e.BadValue())
	order.PutUint16(reply[8:10], uint16(e.MinorOp()))
	reply[10] = e.MajorOp()
	return reply
}

type GenericError struct {
	seq      uint16
	badValue uint32
	minorOp  byte
	majorOp  reqCode
	code     byte
}

func (e GenericError) Code() byte       { return e.code }
func (e GenericError) Sequence() uint16 { return e.seq }
func (e GenericError) BadValue() uint32 { return e.badValue }
func (e GenericError) MinorOp() byte    { return e.minorOp }
func (e GenericError) MajorOp() byte    { return byte(e.majorOp) }

func (e *GenericError) encodeMessage(order binary.ByteOrder) []byte {
	reply := make([]byte, 32)
	reply[0] = 0 // Error type
	reply[1] = e.Code()
	order.PutUint16(reply[2:4], e.Sequence())
	order.PutUint32(reply[4:8], e.BadValue())
	order.PutUint16(reply[8:10], uint16(e.MinorOp()))
	reply[10] = e.MajorOp()
	return reply
}
