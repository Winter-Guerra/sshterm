//go:build x11

package x11

import (
	"bytes"
	"encoding/binary"
	"testing"
)

func TestReplyMessages(t *testing.T) {
	t.Run("GetKeyboardControl", func(t *testing.T) {
		reply := &getKeyboardControlReply{
			sequence:         9,
			keyClickPercent:  1,
			bellPercent:      2,
			bellPitch:        3,
			bellDuration:     4,
			ledMask:          5,
			globalAutoRepeat: 1,
			autoRepeats:      [32]byte{1, 2, 3},
		}
		encoded := reply.encodeMessage(binary.LittleEndian)
		expected := make([]byte, 52)
		expected[0] = 1
		expected[1] = 1
		binary.LittleEndian.PutUint16(expected[2:4], 9)
		binary.LittleEndian.PutUint32(expected[4:8], 5)
		binary.LittleEndian.PutUint32(expected[8:12], 5)
		expected[12] = 1
		expected[13] = 2
		binary.LittleEndian.PutUint16(expected[14:16], 3)
		binary.LittleEndian.PutUint16(expected[16:18], 4)
		expected[20] = 1
		expected[21] = 2
		expected[22] = 3
		if !bytes.Equal(encoded, expected) {
			t.Errorf("GetKeyboardControlReply encoding failed. Got %v, want %v", encoded, expected)
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

	t.Run("GetWindowAttributes", func(t *testing.T) {
		reply := &getWindowAttributesReply{
			sequence:           10,
			backingStore:       1,
			visualID:           2,
			class:              3,
			bitGravity:         4,
			winGravity:         5,
			backingPlanes:      6,
			backingPixel:       7,
			saveUnder:          true,
			mapped:             true,
			mapState:           2,
			overrideRedirect:   true,
			colormap:           8,
			allEventMasks:      9,
			yourEventMask:      10,
			doNotPropagateMask: 11,
		}
		encoded := reply.encodeMessage(binary.LittleEndian)
		expected := make([]byte, 44)
		expected[0] = 1
		expected[1] = 1
		binary.LittleEndian.PutUint16(expected[2:4], 10)
		binary.LittleEndian.PutUint32(expected[4:8], 3)
		binary.LittleEndian.PutUint32(expected[8:12], 2)
		binary.LittleEndian.PutUint16(expected[12:14], 3)
		expected[14] = 4
		expected[15] = 5
		binary.LittleEndian.PutUint32(expected[16:20], 6)
		binary.LittleEndian.PutUint32(expected[20:24], 7)
		expected[24] = 1
		expected[25] = 1
		expected[26] = 2
		expected[27] = 1
		binary.LittleEndian.PutUint32(expected[28:32], 8)
		binary.LittleEndian.PutUint32(expected[32:36], 9)
		binary.LittleEndian.PutUint32(expected[36:40], 10)
		binary.LittleEndian.PutUint16(expected[40:42], 11)
		if !bytes.Equal(encoded, expected) {
			t.Errorf("GetWindowAttributesReply encoding failed. Got %v, want %v", encoded, expected)
		}
	})

	t.Run("GetGeometry", func(t *testing.T) {
		reply := &getGeometryReply{
			sequence:    11,
			depth:       1,
			root:        2,
			x:           3,
			y:           4,
			width:       5,
			height:      6,
			borderWidth: 7,
		}
		encoded := reply.encodeMessage(binary.LittleEndian)
		expected := make([]byte, 32)
		expected[0] = 1
		expected[1] = 1
		binary.LittleEndian.PutUint16(expected[2:4], 11)
		binary.LittleEndian.PutUint32(expected[8:12], 2)
		binary.LittleEndian.PutUint16(expected[12:14], 3)
		binary.LittleEndian.PutUint16(expected[14:16], 4)
		binary.LittleEndian.PutUint16(expected[16:18], 5)
		binary.LittleEndian.PutUint16(expected[18:20], 6)
		binary.LittleEndian.PutUint16(expected[20:22], 7)
		if !bytes.Equal(encoded, expected) {
			t.Errorf("GetGeometryReply encoding failed. Got %v, want %v", encoded, expected)
		}
	})

	t.Run("InternAtom", func(t *testing.T) {
		reply := &internAtomReply{
			sequence: 12,
			atom:     1,
		}
		encoded := reply.encodeMessage(binary.LittleEndian)
		expected := make([]byte, 32)
		expected[0] = 1
		binary.LittleEndian.PutUint16(expected[2:4], 12)
		binary.LittleEndian.PutUint32(expected[8:12], 1)
		if !bytes.Equal(encoded, expected) {
			t.Errorf("InternAtomReply encoding failed. Got %v, want %v", encoded, expected)
		}
	})

	t.Run("GetAtomName", func(t *testing.T) {
		reply := &getAtomNameReply{
			sequence:   13,
			nameLength: 4,
			name:       "test",
		}
		encoded := reply.encodeMessage(binary.LittleEndian)
		expected := make([]byte, 36)
		expected[0] = 1
		binary.LittleEndian.PutUint16(expected[2:4], 13)
		binary.LittleEndian.PutUint32(expected[4:8], 1)
		binary.LittleEndian.PutUint16(expected[8:10], 4)
		copy(expected[32:], "test")
		if !bytes.Equal(encoded, expected) {
			t.Errorf("GetAtomNameReply encoding failed. Got %v, want %v", encoded, expected)
		}
	})

	t.Run("GetProperty", func(t *testing.T) {
		reply := &getPropertyReply{
			sequence:              14,
			format:                8,
			propertyType:          1,
			bytesAfter:            2,
			valueLenInFormatUnits: 4,
			value:                 []byte{1, 2, 3, 4},
		}
		encoded := reply.encodeMessage(binary.LittleEndian)
		expected := make([]byte, 36)
		expected[0] = 1
		expected[1] = 8
		binary.LittleEndian.PutUint16(expected[2:4], 14)
		binary.LittleEndian.PutUint32(expected[4:8], 1)
		binary.LittleEndian.PutUint32(expected[8:12], 1)
		binary.LittleEndian.PutUint32(expected[12:16], 2)
		binary.LittleEndian.PutUint32(expected[16:20], 4)
		copy(expected[32:], []byte{1, 2, 3, 4})
		if !bytes.Equal(encoded, expected) {
			t.Errorf("GetPropertyReply encoding failed. Got %v, want %v", encoded, expected)
		}
	})

	t.Run("ListProperties", func(t *testing.T) {
		reply := &listPropertiesReply{
			sequence:      15,
			numProperties: 3,
			atoms:         []uint32{1, 2, 3},
		}
		encoded := reply.encodeMessage(binary.LittleEndian)
		expected := make([]byte, 44)
		expected[0] = 1
		binary.LittleEndian.PutUint16(expected[2:4], 15)
		binary.LittleEndian.PutUint32(expected[4:8], 3)
		binary.LittleEndian.PutUint16(expected[8:10], 3)
		binary.LittleEndian.PutUint32(expected[32:36], 1)
		binary.LittleEndian.PutUint32(expected[36:40], 2)
		binary.LittleEndian.PutUint32(expected[40:44], 3)
		if !bytes.Equal(encoded, expected) {
			t.Errorf("ListPropertiesReply encoding failed. Got %v, want %v", encoded, expected)
		}
	})

	t.Run("GetSelectionOwner", func(t *testing.T) {
		reply := &getSelectionOwnerReply{
			sequence: 16,
			owner:    1,
		}
		encoded := reply.encodeMessage(binary.LittleEndian)
		expected := make([]byte, 32)
		expected[0] = 1
		binary.LittleEndian.PutUint16(expected[2:4], 16)
		binary.LittleEndian.PutUint32(expected[8:12], 1)
		if !bytes.Equal(encoded, expected) {
			t.Errorf("GetSelectionOwnerReply encoding failed. Got %v, want %v", encoded, expected)
		}
	})

	t.Run("GrabPointer", func(t *testing.T) {
		reply := &grabPointerReply{
			sequence: 17,
			status:   1,
		}
		encoded := reply.encodeMessage(binary.LittleEndian)
		expected := make([]byte, 32)
		expected[0] = 1
		expected[1] = 1
		binary.LittleEndian.PutUint16(expected[2:4], 17)
		if !bytes.Equal(encoded, expected) {
			t.Errorf("GrabPointerReply encoding failed. Got %v, want %v", encoded, expected)
		}
	})

	t.Run("GrabKeyboard", func(t *testing.T) {
		reply := &grabKeyboardReply{
			sequence: 18,
			status:   1,
		}
		encoded := reply.encodeMessage(binary.LittleEndian)
		expected := make([]byte, 32)
		expected[0] = 1
		expected[1] = 1
		binary.LittleEndian.PutUint16(expected[2:4], 18)
		if !bytes.Equal(encoded, expected) {
			t.Errorf("GrabKeyboardReply encoding failed. Got %v, want %v", encoded, expected)
		}
	})

	t.Run("QueryPointer", func(t *testing.T) {
		reply := &queryPointerReply{
			sequence:   19,
			sameScreen: true,
			root:       1,
			child:      2,
			rootX:      3,
			rootY:      4,
			winX:       5,
			winY:       6,
			mask:       7,
		}
		encoded := reply.encodeMessage(binary.LittleEndian)
		expected := make([]byte, 32)
		expected[0] = 1
		expected[1] = 1
		binary.LittleEndian.PutUint16(expected[2:4], 19)
		binary.LittleEndian.PutUint32(expected[8:12], 1)
		binary.LittleEndian.PutUint32(expected[12:16], 2)
		binary.LittleEndian.PutUint16(expected[16:18], 3)
		binary.LittleEndian.PutUint16(expected[18:20], 4)
		binary.LittleEndian.PutUint16(expected[20:22], 5)
		binary.LittleEndian.PutUint16(expected[22:24], 6)
		binary.LittleEndian.PutUint16(expected[24:26], 7)
		if !bytes.Equal(encoded, expected) {
			t.Errorf("QueryPointerReply encoding failed. Got %v, want %v", encoded, expected)
		}
	})

	t.Run("TranslateCoords", func(t *testing.T) {
		reply := &translateCoordsReply{
			sequence:   20,
			sameScreen: true,
			child:      1,
			dstX:       2,
			dstY:       3,
		}
		encoded := reply.encodeMessage(binary.LittleEndian)
		expected := make([]byte, 32)
		expected[0] = 1
		expected[1] = 1
		binary.LittleEndian.PutUint16(expected[2:4], 20)
		binary.LittleEndian.PutUint32(expected[8:12], 1)
		binary.LittleEndian.PutUint16(expected[12:14], 2)
		binary.LittleEndian.PutUint16(expected[14:16], 3)
		if !bytes.Equal(encoded, expected) {
			t.Errorf("TranslateCoordsReply encoding failed. Got %v, want %v", encoded, expected)
		}
	})

	t.Run("GetInputFocus", func(t *testing.T) {
		reply := &getInputFocusReply{
			sequence: 21,
			revertTo: 1,
			focus:    2,
		}
		encoded := reply.encodeMessage(binary.LittleEndian)
		expected := make([]byte, 32)
		expected[0] = 1
		expected[1] = 1
		binary.LittleEndian.PutUint16(expected[2:4], 21)
		binary.LittleEndian.PutUint32(expected[8:12], 2)
		if !bytes.Equal(encoded, expected) {
			t.Errorf("GetInputFocusReply encoding failed. Got %v, want %v", encoded, expected)
		}
	})

	t.Run("QueryFont", func(t *testing.T) {
		reply := &queryFontReply{
			sequence:       22,
			minCharOrByte2: 1,
			maxCharOrByte2: 2,
			defaultChar:    3,
			drawDirection:  1,
			minByte1:       1,
			maxByte1:       2,
			allCharsExist:  true,
			fontAscent:     10,
			fontDescent:    2,
			charInfos: []xCharInfo{
				{1, 2, 3, 4, 5, 6},
			},
		}
		encoded := reply.encodeMessage(binary.LittleEndian)
		expected := make([]byte, 72)
		expected[0] = 1
		expected[1] = 1
		binary.LittleEndian.PutUint16(expected[2:4], 22)
		binary.LittleEndian.PutUint32(expected[4:8], 10)
		binary.LittleEndian.PutUint16(expected[40:42], 1)
		binary.LittleEndian.PutUint16(expected[42:44], 2)
		binary.LittleEndian.PutUint16(expected[44:46], 3)
		expected[48] = 1
		expected[49] = 1
		expected[50] = 2
		expected[51] = 1
		binary.LittleEndian.PutUint16(expected[52:54], uint16(10))
		binary.LittleEndian.PutUint16(expected[54:56], uint16(2))
		binary.LittleEndian.PutUint32(expected[56:60], 1)
		binary.LittleEndian.PutUint16(expected[60:62], 1)
		binary.LittleEndian.PutUint16(expected[62:64], 2)
		binary.LittleEndian.PutUint16(expected[64:66], 3)
		binary.LittleEndian.PutUint16(expected[66:68], 4)
		binary.LittleEndian.PutUint16(expected[68:70], 5)
		binary.LittleEndian.PutUint16(expected[70:72], 6)
		if !bytes.Equal(encoded, expected) {
			t.Errorf("QueryFontReply encoding failed. Got %v, want %v", encoded, expected)
		}
	})

	t.Run("ListFonts", func(t *testing.T) {
		reply := &listFontsReply{
			sequence:  23,
			fontNames: []string{"test", "test2"},
		}
		encoded := reply.encodeMessage(binary.LittleEndian)
		expected := make([]byte, 44)
		expected[0] = 1
		binary.LittleEndian.PutUint16(expected[2:4], 23)
		binary.LittleEndian.PutUint32(expected[4:8], 3)
		binary.LittleEndian.PutUint16(expected[8:10], 2)
		expected[32] = 4
		copy(expected[33:37], "test")
		expected[37] = 5
		copy(expected[38:43], "test2")
		if !bytes.Equal(encoded, expected) {
			t.Errorf("ListFontsReply encoding failed. Got %v, want %v", encoded, expected)
		}
	})

	t.Run("GetImage", func(t *testing.T) {
		reply := &getImageReply{
			sequence:  24,
			depth:     24,
			visualID:  1,
			imageData: []byte{1, 2, 3, 4},
		}
		encoded := reply.encodeMessage(binary.LittleEndian)
		expected := make([]byte, 36)
		expected[0] = 1
		expected[1] = 24
		binary.LittleEndian.PutUint16(expected[2:4], 24)
		binary.LittleEndian.PutUint32(expected[4:8], 1)
		binary.LittleEndian.PutUint32(expected[8:12], 1)
		copy(expected[32:], []byte{1, 2, 3, 4})
		if !bytes.Equal(encoded, expected) {
			t.Errorf("GetImageReply encoding failed. Got %v, want %v", encoded, expected)
		}
	})

	t.Run("AllocColor", func(t *testing.T) {
		reply := &allocColorReply{
			sequence: 25,
			red:      1,
			green:    2,
			blue:     3,
			pixel:    4,
		}
		encoded := reply.encodeMessage(binary.LittleEndian)
		expected := make([]byte, 32)
		expected[0] = 1
		binary.LittleEndian.PutUint16(expected[2:4], 25)
		binary.LittleEndian.PutUint16(expected[8:10], 1)
		binary.LittleEndian.PutUint16(expected[10:12], 2)
		binary.LittleEndian.PutUint16(expected[12:14], 3)
		binary.LittleEndian.PutUint32(expected[16:20], 4)
		if !bytes.Equal(encoded, expected) {
			t.Errorf("AllocColorReply encoding failed. Got %v, want %v", encoded, expected)
		}
	})

	t.Run("ListInstalledColormaps", func(t *testing.T) {
		reply := &listInstalledColormapsReply{
			sequence:     26,
			numColormaps: 3,
			colormaps:    []uint32{1, 2, 3},
		}
		encoded := reply.encodeMessage(binary.LittleEndian)
		expected := make([]byte, 44)
		expected[0] = 1
		binary.LittleEndian.PutUint16(expected[2:4], 26)
		binary.LittleEndian.PutUint32(expected[4:8], 3)
		binary.LittleEndian.PutUint16(expected[8:10], 3)
		binary.LittleEndian.PutUint32(expected[32:36], 1)
		binary.LittleEndian.PutUint32(expected[36:40], 2)
		binary.LittleEndian.PutUint32(expected[40:44], 3)
		if !bytes.Equal(encoded, expected) {
			t.Errorf("ListInstalledColormapsReply encoding failed. Got %v, want %v", encoded, expected)
		}
	})

	t.Run("QueryColors", func(t *testing.T) {
		reply := &queryColorsReply{
			sequence: 27,
			colors: []xColorItem{
				{Pixel: 0, Red: 1, Green: 2, Blue: 3, Flags: 0},
				{Pixel: 0, Red: 4, Green: 5, Blue: 6, Flags: 0},
			},
		}
		encoded := reply.encodeMessage(binary.LittleEndian)
		expected := make([]byte, 48)
		expected[0] = 1
		binary.LittleEndian.PutUint16(expected[2:4], 27)
		binary.LittleEndian.PutUint32(expected[4:8], 4)
		binary.LittleEndian.PutUint16(expected[8:10], 2)
		binary.LittleEndian.PutUint16(expected[32:34], 1)
		binary.LittleEndian.PutUint16(expected[34:36], 2)
		binary.LittleEndian.PutUint16(expected[36:38], 3)
		binary.LittleEndian.PutUint16(expected[40:42], 4)
		binary.LittleEndian.PutUint16(expected[42:44], 5)
		binary.LittleEndian.PutUint16(expected[44:46], 6)
		if !bytes.Equal(encoded, expected) {
			t.Errorf("QueryColorsReply encoding failed. Got %v, want %v", encoded, expected)
		}
	})

	t.Run("LookupColor", func(t *testing.T) {
		reply := &lookupColorReply{
			sequence:   28,
			red:        1,
			green:      2,
			blue:       3,
			exactRed:   4,
			exactGreen: 5,
			exactBlue:  6,
		}
		encoded := reply.encodeMessage(binary.LittleEndian)
		expected := make([]byte, 32)
		expected[0] = 1
		binary.LittleEndian.PutUint16(expected[2:4], 28)
		binary.LittleEndian.PutUint16(expected[8:10], 1)
		binary.LittleEndian.PutUint16(expected[10:12], 2)
		binary.LittleEndian.PutUint16(expected[12:14], 3)
		binary.LittleEndian.PutUint16(expected[14:16], 4)
		binary.LittleEndian.PutUint16(expected[16:18], 5)
		binary.LittleEndian.PutUint16(expected[18:20], 6)
		if !bytes.Equal(encoded, expected) {
			t.Errorf("LookupColorReply encoding failed. Got %v, want %v", encoded, expected)
		}
	})

	t.Run("QueryBestSize", func(t *testing.T) {
		reply := &queryBestSizeReply{
			sequence: 29,
			width:    1,
			height:   2,
		}
		encoded := reply.encodeMessage(binary.LittleEndian)
		expected := make([]byte, 32)
		expected[0] = 1
		binary.LittleEndian.PutUint16(expected[2:4], 29)
		binary.LittleEndian.PutUint16(expected[8:10], 1)
		binary.LittleEndian.PutUint16(expected[10:12], 2)
		if !bytes.Equal(encoded, expected) {
			t.Errorf("QueryBestSizeReply encoding failed. Got %v, want %v", encoded, expected)
		}
	})

	t.Run("QueryExtension", func(t *testing.T) {
		reply := &queryExtensionReply{
			sequence:    30,
			present:     true,
			majorOpcode: 1,
			firstEvent:  2,
			firstError:  3,
		}
		encoded := reply.encodeMessage(binary.LittleEndian)
		expected := make([]byte, 32)
		expected[0] = 1
		expected[1] = 1
		binary.LittleEndian.PutUint16(expected[2:4], 30)
		expected[8] = 1
		expected[9] = 2
		expected[10] = 3
		if !bytes.Equal(encoded, expected) {
			t.Errorf("QueryExtensionReply encoding failed. Got %v, want %v", encoded, expected)
		}
	})

	t.Run("SetPointerMapping", func(t *testing.T) {
		reply := &setPointerMappingReply{
			sequence: 31,
		}
		encoded := reply.encodeMessage(binary.LittleEndian)
		expected := make([]byte, 32)
		expected[0] = 1
		binary.LittleEndian.PutUint16(expected[2:4], 31)
		if !bytes.Equal(encoded, expected) {
			t.Errorf("SetPointerMappingReply encoding failed. Got %v, want %v", encoded, expected)
		}
	})

	t.Run("GetPointerMapping", func(t *testing.T) {
		reply := &getPointerMappingReply{
			sequence: 32,
			length:   4,
			pMap:     []byte{1, 2, 3, 4},
		}
		encoded := reply.encodeMessage(binary.LittleEndian)
		expected := make([]byte, 36)
		expected[0] = 1
		expected[1] = 4
		binary.LittleEndian.PutUint16(expected[2:4], 32)
		binary.LittleEndian.PutUint32(expected[4:8], 1)
		copy(expected[32:], []byte{1, 2, 3, 4})
		if !bytes.Equal(encoded, expected) {
			t.Errorf("GetPointerMappingReply encoding failed. Got %v, want %v", encoded, expected)
		}
	})

	t.Run("GetKeyboardMapping", func(t *testing.T) {
		reply := &getKeyboardMappingReply{
			sequence: 33,
			keySyms:  []uint32{1, 2, 3},
		}
		encoded := reply.encodeMessage(binary.LittleEndian)
		expected := make([]byte, 44)
		expected[0] = 1
		expected[1] = 3
		binary.LittleEndian.PutUint16(expected[2:4], 33)
		binary.LittleEndian.PutUint32(expected[4:8], 3)
		binary.LittleEndian.PutUint32(expected[32:36], 1)
		binary.LittleEndian.PutUint32(expected[36:40], 2)
		binary.LittleEndian.PutUint32(expected[40:44], 3)
		if !bytes.Equal(encoded, expected) {
			t.Errorf("GetKeyboardMappingReply encoding failed. Got %v, want %v", encoded, expected)
		}
	})

	t.Run("GetScreenSaver", func(t *testing.T) {
		reply := &getScreenSaverReply{
			sequence:    34,
			timeout:     1,
			interval:    2,
			preferBlank: 1,
			allowExpose: 1,
		}
		encoded := reply.encodeMessage(binary.LittleEndian)
		expected := make([]byte, 32)
		expected[0] = 1
		binary.LittleEndian.PutUint16(expected[2:4], 34)
		binary.LittleEndian.PutUint16(expected[8:10], 1)
		binary.LittleEndian.PutUint16(expected[10:12], 2)
		expected[12] = 1
		expected[13] = 1
		if !bytes.Equal(encoded, expected) {
			t.Errorf("GetScreenSaverReply encoding failed. Got %v, want %v", encoded, expected)
		}
	})

	t.Run("ListHosts", func(t *testing.T) {
		reply := &listHostsReply{
			sequence: 35,
			numHosts: 1,
			hosts: []Host{
				{
					Family: 1,
					Data:   []byte{1, 2, 3, 4},
				},
			},
		}
		encoded := reply.encodeMessage(binary.LittleEndian)
		expected := make([]byte, 40)
		expected[0] = 1
		binary.LittleEndian.PutUint16(expected[2:4], 35)
		binary.LittleEndian.PutUint32(expected[4:8], 2)
		binary.LittleEndian.PutUint16(expected[8:10], 1)
		expected[32] = 1
		binary.LittleEndian.PutUint16(expected[34:36], 4)
		copy(expected[36:], []byte{1, 2, 3, 4})
		if !bytes.Equal(encoded, expected) {
			t.Errorf("ListHostsReply encoding failed. Got %v, want %v", encoded, expected)
		}
	})

	t.Run("SetModifierMapping", func(t *testing.T) {
		reply := &setModifierMappingReply{
			sequence: 36,
		}
		encoded := reply.encodeMessage(binary.LittleEndian)
		expected := make([]byte, 32)
		expected[0] = 1
		binary.LittleEndian.PutUint16(expected[2:4], 36)
		if !bytes.Equal(encoded, expected) {
			t.Errorf("SetModifierMappingReply encoding failed. Got %v, want %v", encoded, expected)
		}
	})

	t.Run("GetModifierMapping", func(t *testing.T) {
		reply := &getModifierMappingReply{
			sequence: 37,
			keyCodes: []KeyCode{1, 2, 3, 4},
		}
		encoded := reply.encodeMessage(binary.LittleEndian)
		expected := make([]byte, 36)
		expected[0] = 1
		binary.LittleEndian.PutUint16(expected[2:4], 37)
		binary.LittleEndian.PutUint32(expected[4:8], 1)
		copy(expected[32:], []byte{1, 2, 3, 4})
		if !bytes.Equal(encoded, expected) {
			t.Errorf("GetModifierMappingReply encoding failed. Got %v, want %v", encoded, expected)
		}
	})

	t.Run("SetPointerMapping", func(t *testing.T) {
		reply := &setPointerMappingReply{
			sequence: 31,
			status:   1,
		}
		encoded := reply.encodeMessage(binary.LittleEndian)
		expected := make([]byte, 32)
		expected[0] = 1
		expected[1] = 1
		binary.LittleEndian.PutUint16(expected[2:4], 31)
		if !bytes.Equal(encoded, expected) {
			t.Errorf("SetPointerMappingReply encoding failed. Got %v, want %v", encoded, expected)
		}
	})

	t.Run("SetModifierMapping", func(t *testing.T) {
		reply := &setModifierMappingReply{
			sequence: 36,
			status:   1,
		}
		encoded := reply.encodeMessage(binary.LittleEndian)
		expected := make([]byte, 32)
		expected[0] = 1
		expected[1] = 1
		binary.LittleEndian.PutUint16(expected[2:4], 36)
		if !bytes.Equal(encoded, expected) {
			t.Errorf("SetModifierMappingReply encoding failed. Got %v, want %v", encoded, expected)
		}
	})

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
		if !bytes.Equal(encoded, expected) {
			t.Errorf("GetPointerControlReply encoding failed. Got %v, want %v", encoded, expected)
		}
	})

	t.Run("Setup", func(t *testing.T) {
		s := newDefaultSetup()
		b := s.marshal(binary.LittleEndian)
		if len(b) == 0 {
			t.Error("Setup marshaling failed")
		}
	})
}

func int32ToUint32(i int32) uint32 {
	return uint32(i)
}
