//go:build x11 && !wasm

package x11

import (
	"bytes"
	"encoding/binary"
	"io"
	"testing"

	"github.com/stretchr/testify/require"
)

// drainMessages reads all messages from the buffer and returns them as a slice of decoded messages.
func drainMessages(t *testing.T, buf *bytes.Buffer, order binary.ByteOrder) []interface{} {
	t.Helper()
	var messages []interface{}
	for buf.Len() > 0 {
		msg := decodeOneMessage(t, buf, order)
		if msg != nil {
			messages = append(messages, msg)
		}
	}
	return messages
}

// decodeOneMessage decodes a single X11 message from the buffer.
func decodeOneMessage(t *testing.T, buf *bytes.Buffer, order binary.ByteOrder) interface{} {
	t.Helper()
	if buf.Len() < 32 {
		return nil // Not enough data for a full message
	}

	var header [32]byte
	_, err := io.ReadFull(buf, header[:])
	if err == io.EOF {
		return nil
	}
	require.NoError(t, err)

	msgType := header[0]

	// Check for XInput extension events
	if msgType == xInputOpcode {
		subOpcode := header[1]
		switch subOpcode {
		case DeviceButtonPress:
			return decodeDeviceButtonPressEvent(header, order)
		case DeviceButtonRelease:
			return decodeDeviceButtonReleaseEvent(header, order)
		// Add other XInput event types here as needed
		default:
			return &x11RawEvent{data: header[:]}
		}
	}

	switch msgType {
	case 4: // ButtonPress
		return decodeButtonPressEvent(header, order)
	case 1, 2: // KeyPress, KeyRelease
		return decodeKeyEvent(header, order)
	// Add other core event types here as needed
	default:
		// For unknown message types, return a raw event
		return &x11RawEvent{data: header[:]}
	}
}

func decodeKeyEvent(data [32]byte, order binary.ByteOrder) *keyEvent {
	return &keyEvent{
		opcode:     data[0],
		detail:     data[1],
		sequence:   order.Uint16(data[2:4]),
		time:       order.Uint32(data[4:8]),
		root:       order.Uint32(data[8:12]),
		event:      order.Uint32(data[12:16]),
		child:      order.Uint32(data[16:20]),
		rootX:      int16(order.Uint16(data[20:22])),
		rootY:      int16(order.Uint16(data[22:24])),
		eventX:     int16(order.Uint16(data[24:26])),
		eventY:     int16(order.Uint16(data[26:28])),
		state:      order.Uint16(data[28:30]),
		sameScreen: data[30] != 0,
	}
}

func decodeButtonPressEvent(data [32]byte, order binary.ByteOrder) *ButtonPressEvent {
	return &ButtonPressEvent{
		detail:     data[1],
		sequence:   order.Uint16(data[2:4]),
		time:       order.Uint32(data[4:8]),
		root:       order.Uint32(data[8:12]),
		event:      order.Uint32(data[12:16]),
		child:      order.Uint32(data[16:20]),
		rootX:      int16(order.Uint16(data[20:22])),
		rootY:      int16(order.Uint16(data[22:24])),
		eventX:     int16(order.Uint16(data[24:26])),
		eventY:     int16(order.Uint16(data[26:28])),
		state:      order.Uint16(data[28:30]),
		sameScreen: data[30] != 0,
	}
}

func decodeDeviceButtonPressEvent(data [32]byte, order binary.ByteOrder) *DeviceButtonPressEvent {
	return &DeviceButtonPressEvent{
		Sequence: order.Uint16(data[2:4]),
		Time:     order.Uint32(data[4:8]),
		Root:     order.Uint32(data[8:12]),
		Event:    order.Uint32(data[12:16]),
		Child:    order.Uint32(data[16:20]),
		RootX:    int16(order.Uint16(data[20:22])),
		RootY:    int16(order.Uint16(data[22:24])),
		EventX:   int16(order.Uint16(data[24:26])),
		EventY:   int16(order.Uint16(data[26:28])),
		State:    order.Uint16(data[28:30]),
		DeviceID: data[30],
		Detail:   data[31],
	}
}

func decodeDeviceButtonReleaseEvent(data [32]byte, order binary.ByteOrder) *DeviceButtonReleaseEvent {
	return &DeviceButtonReleaseEvent{
		sequence: order.Uint16(data[2:4]),
		Time:     order.Uint32(data[4:8]),
		Root:     order.Uint32(data[8:12]),
		Event:    order.Uint32(data[12:16]),
		Child:    order.Uint32(data[16:20]),
		RootX:    int16(order.Uint16(data[20:22])),
		RootY:    int16(order.Uint16(data[22:24])),
		EventX:   int16(order.Uint16(data[24:26])),
		EventY:   int16(order.Uint16(data[26:28])),
		State:    order.Uint16(data[28:30]),
		DeviceID: data[30],
		Button:   data[31],
	}
}
