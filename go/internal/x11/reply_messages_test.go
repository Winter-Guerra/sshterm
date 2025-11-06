//go:build x11

package x11

import (
	"bytes"
	"encoding/binary"
	"testing"
)

func TestReplyMessages(t *testing.T) {
	t.Run("GetPointerControl", func(t *testing.T) {
		reply := &getPointerControlReply{
			sequence:         1,
			accelNumerator:   2,
			accelDenominator: 3,
			threshold:        4,
			doAccel:          true,
			doThreshold:      true,
		}
		encoded := reply.encodeMessage(binary.LittleEndian)
		expected := make([]byte, 32)
		expected[0] = 1
		binary.LittleEndian.PutUint16(expected[2:4], 1)
		binary.LittleEndian.PutUint32(expected[4:8], 0)
		binary.LittleEndian.PutUint16(expected[8:10], 2)
		binary.LittleEndian.PutUint16(expected[10:12], 3)
		binary.LittleEndian.PutUint16(expected[12:14], 4)
		expected[14] = 1
		expected[15] = 1
		if !bytes.Equal(encoded, expected) {
			t.Errorf("GetPointerControlReply encoding failed. Got %v, want %v", encoded, expected)
		}
	})

	t.Run("GetMotionEvents", func(t *testing.T) {
		reply := &getMotionEventsReply{
			sequence: 2,
			nEvents:  1,
			events: []TimeCoord{
				{
					Time: 123,
					X:    10,
					Y:    20,
				},
			},
		}
		encoded := reply.encodeMessage(binary.LittleEndian)
		expected := make([]byte, 40)
		expected[0] = 1
		binary.LittleEndian.PutUint16(expected[2:4], 2)
		binary.LittleEndian.PutUint32(expected[4:8], 2)
		binary.LittleEndian.PutUint32(expected[8:12], 1)
		binary.LittleEndian.PutUint32(expected[32:36], 123)
		binary.LittleEndian.PutUint16(expected[36:38], 10)
		binary.LittleEndian.PutUint16(expected[38:40], 20)
		if !bytes.Equal(encoded, expected) {
			t.Errorf("GetMotionEventsReply encoding failed. Got %v, want %v", encoded, expected)
		}
	})

	t.Run("QueryTextExtents", func(t *testing.T) {
		reply := &queryTextExtentsReply{
			sequence:       3,
			drawDirection:  0,
			fontAscent:     10,
			fontDescent:    2,
			overallAscent:  11,
			overallDescent: 3,
			overallWidth:   100,
			overallLeft:    -5,
			overallRight:   95,
		}
		encoded := reply.encodeMessage(binary.LittleEndian)
		expected := make([]byte, 32)
		expected[0] = 1
		expected[1] = 0
		binary.LittleEndian.PutUint16(expected[2:4], 3)
		binary.LittleEndian.PutUint32(expected[4:8], 0)
		binary.LittleEndian.PutUint16(expected[8:10], uint16(10))
		binary.LittleEndian.PutUint16(expected[10:12], uint16(2))
		binary.LittleEndian.PutUint16(expected[12:14], uint16(11))
		binary.LittleEndian.PutUint16(expected[14:16], uint16(3))
		binary.LittleEndian.PutUint32(expected[16:20], uint32(100))
		binary.LittleEndian.PutUint32(expected[20:24], int32ToUint32(-5))
		binary.LittleEndian.PutUint32(expected[24:28], int32ToUint32(95))
		if !bytes.Equal(encoded, expected) {
			t.Errorf("QueryTextExtentsReply encoding failed. Got %v, want %v", encoded, expected)
		}
	})

	t.Run("ListExtensions", func(t *testing.T) {
		reply := &listExtensionsReply{
			sequence: 4,
			nNames:   2,
			names:    []string{"ext1", "ext2"},
		}
		encoded := reply.encodeMessage(binary.LittleEndian)

		var data []byte
		data = append(data, 4, 'e', 'x', 't', '1')
		data = append(data, 4, 'e', 'x', 't', '2')

		expected := make([]byte, 32+len(data))
		expected[0] = 1
		expected[1] = 2
		binary.LittleEndian.PutUint16(expected[2:4], 4)
		binary.LittleEndian.PutUint32(expected[4:8], uint32((len(data)+3)/4))
		copy(expected[32:], data)

		if !bytes.Equal(encoded, expected) {
			t.Errorf("ListExtensionsReply encoding failed. Got %v, want %v", encoded, expected)
		}
	})

	t.Run("QueryTree", func(t *testing.T) {
		reply := &queryTreeReply{
			sequence:  5,
			root:      1,
			parent:    2,
			nChildren: 1,
			children:  []uint32{3},
		}
		encoded := reply.encodeMessage(binary.LittleEndian)
		expected := make([]byte, 36)
		expected[0] = 1
		binary.LittleEndian.PutUint16(expected[2:4], 5)
		binary.LittleEndian.PutUint32(expected[4:8], 1)
		binary.LittleEndian.PutUint32(expected[8:12], 1)
		binary.LittleEndian.PutUint32(expected[12:16], 2)
		binary.LittleEndian.PutUint16(expected[16:18], 1)
		binary.LittleEndian.PutUint32(expected[32:36], 3)
		if !bytes.Equal(encoded, expected) {
			t.Errorf("QueryTreeReply encoding failed. Got %v, want %v", encoded, expected)
		}
	})

	t.Run("ListFontsWithInfo", func(t *testing.T) {
		reply := &listFontsWithInfoReply{
			sequence:   6,
			nameLength: 4,
			fontName:   "test",
		}
		encoded := reply.encodeMessage(binary.LittleEndian)
		expected := make([]byte, 64)
		expected[0] = 1
		expected[1] = 4
		binary.LittleEndian.PutUint16(expected[2:4], 6)
		binary.LittleEndian.PutUint32(expected[4:8], 8)
		copy(expected[60:], "test")
		if !bytes.Equal(encoded, expected) {
			t.Errorf("ListFontsWithInfoReply encoding failed. Got %v, want %v", encoded, expected)
		}
	})

	t.Run("GetFontPath", func(t *testing.T) {
		reply := &getFontPathReply{
			sequence: 7,
			nPaths:   1,
			paths:    []string{"/usr/share/fonts"},
		}
		encoded := reply.encodeMessage(binary.LittleEndian)
		expected := make([]byte, 52)
		expected[0] = 1
		binary.LittleEndian.PutUint16(expected[2:4], 7)
		binary.LittleEndian.PutUint32(expected[4:8], 5)
		binary.LittleEndian.PutUint16(expected[8:10], 1)
		expected[32] = 16
		copy(expected[33:], "/usr/share/fonts")
		if !bytes.Equal(encoded, expected) {
			t.Errorf("GetFontPathReply encoding failed. Got %v, want %v", encoded, expected)
		}
	})

	t.Run("QueryKeymap", func(t *testing.T) {
		reply := &queryKeymapReply{
			sequence: 8,
			keys:     [32]byte{1, 2, 3},
		}
		encoded := reply.encodeMessage(binary.LittleEndian)
		expected := make([]byte, 40)
		expected[0] = 1
		binary.LittleEndian.PutUint16(expected[2:4], 8)
		binary.LittleEndian.PutUint32(expected[4:8], 2)
		expected[8] = 1
		expected[9] = 2
		expected[10] = 3
		if !bytes.Equal(encoded, expected) {
			t.Errorf("QueryKeymapReply encoding failed. Got %v, want %v", encoded, expected)
		}
	})

}

func int32ToUint32(i int32) uint32 {
	return uint32(i)
}
