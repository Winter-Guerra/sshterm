//go:build x11

package x11

import (
	"encoding/binary"
	"fmt"
)

// The X11 protocol defines a set of errors that can be returned by the server.
// Each error has a unique code, and some errors have additional data.
// The following structs define the errors that can be returned by the server.

// Error is an interface that all X11 errors implement.
type Error interface {
	// Code returns the error code.
	Code() byte
	// Sequence returns the sequence number of the request that caused the error.
	Sequence() uint16
	// BadValue returns the bad value that caused the error, if any.
	BadValue() uint32
	// MinorOp returns the minor opcode of the request that caused the error.
	MinorOp() byte
	// MajorOp returns the major opcode of the request that caused the error.
	MajorOp() byte
	// encodeMessage encodes the error message into a byte slice.
	encodeMessage(order binary.ByteOrder) []byte

	error
}

// baseError is a helper struct that implements the Error interface.
type baseError struct {
	seq      uint16
	badValue uint32
	minorOp  byte
	majorOp  reqCode
	code     byte
}

func (e baseError) Code() byte       { return e.code }
func (e baseError) Sequence() uint16 { return e.seq }
func (e baseError) BadValue() uint32 { return e.badValue }
func (e baseError) MinorOp() byte    { return e.minorOp }
func (e baseError) MajorOp() byte    { return byte(e.majorOp) }
func (e baseError) Error() string {
	return fmt.Sprintf("X11 error: %d", e.code)
}

func (e *baseError) encodeMessage(order binary.ByteOrder) []byte {
	reply := make([]byte, 32)
	reply[0] = 0 // Error type
	reply[1] = e.Code()
	order.PutUint16(reply[2:4], e.Sequence())
	order.PutUint32(reply[4:8], e.BadValue())
	order.PutUint16(reply[8:10], uint16(e.MinorOp()))
	reply[10] = e.MajorOp()
	return reply
}

// RequestError: 1
type RequestError struct {
	baseError
}

// ValueError: 2
type ValueError struct {
	baseError
}

// WindowError: 3
type WindowError struct {
	baseError
}

// PixmapError: 4
type PixmapError struct {
	baseError
}

// AtomError: 5
type AtomError struct {
	baseError
}

// CursorError: 6
type CursorError struct {
	baseError
}

// FontError: 7
type FontError struct {
	baseError
}

// MatchError: 8
type MatchError struct {
	baseError
}

// DrawableError: 9
type DrawableError struct {
	baseError
}

// AccessError: 10
type AccessError struct {
	baseError
}

// AllocError: 11
type AllocError struct {
	baseError
}

// ColormapError: 12
type ColormapError struct {
	baseError
}

// GContextError: 13
type GContextError struct {
	baseError
}

// IDChoiceError: 14
type IDChoiceError struct {
	baseError
}

// NameError: 15
type NameError struct {
	baseError
}

// LengthError: 16
type LengthError struct {
	baseError
}

// ImplementationError: 17
type ImplementationError struct {
	baseError
}

func NewError(code byte, seq uint16, badValue uint32, minorOp byte, majorOp reqCode) Error {
	base := baseError{
		code:     code,
		seq:      seq,
		badValue: badValue,
		minorOp:  minorOp,
		majorOp:  majorOp,
	}
	switch code {
	case 1:
		return &RequestError{base}
	case ValueErrorCode:
		return &ValueError{base}
	case WindowErrorCode:
		return &WindowError{base}
	case PixmapErrorCode:
		return &PixmapError{base}
	case 5:
		return &AtomError{base}
	case CursorErrorCode:
		return &CursorError{base}
	case 7:
		return &FontError{base}
	case 8:
		return &MatchError{base}
	case 9:
		return &DrawableError{base}
	case 10:
		return &AccessError{base}
	case 11:
		return &AllocError{base}
	case ColormapErrorCode:
		return &ColormapError{base}
	case GContextErrorCode:
		return &GContextError{base}
	case IDChoiceErrorCode:
		return &IDChoiceError{base}
	case 15:
		return &NameError{base}
	case 16:
		return &LengthError{base}
	case 17:
		return &ImplementationError{base}
	default:
		return &GenericError{
			seq:      seq,
			badValue: badValue,
			minorOp:  minorOp,
			majorOp:  majorOp,
			code:     code,
		}
	}
}

// GenericError is used for unknown errors.
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
func (e GenericError) Error() string {
	return fmt.Sprintf("unknown X11 error: %d", e.code)
}

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
