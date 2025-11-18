//go:build x11 && !wasm

package wire

import (
	"bytes"
	"encoding/binary"
	"testing"
)

func TestReplyMessages(t *testing.T) {
	t.Run("GetKeyboardControl", func(t *testing.T) {
		reply := &GetKeyboardControlReply{
			Sequence:         9,
			KeyClickPercent:  1,
			BellPercent:      2,
			BellPitch:        3,
			BellDuration:     4,
			LedMask:          5,
			GlobalAutoRepeat: 1,
			AutoRepeats:      [32]byte{1, 2, 3},
		}
		encoded := reply.EncodeMessage(binary.LittleEndian)
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
		reply := &QueryTreeReply{
			Sequence:    5,
			Root:        1,
			Parent:      2,
			NumChildren: 1,
			Children:    []uint32{3},
		}
		encoded := reply.EncodeMessage(binary.LittleEndian)
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
		reply := &ListFontsWithInfoReply{
			Sequence:   6,
			NameLength: 4,
			FontName:   "test",
		}
		encoded := reply.EncodeMessage(binary.LittleEndian)
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
		reply := &GetFontPathReply{
			Sequence: 7,
			NPaths:   1,
			Paths:    []string{"/usr/share/fonts"},
		}
		encoded := reply.EncodeMessage(binary.LittleEndian)
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
		reply := &QueryKeymapReply{
			Sequence: 8,
			Keys:     [32]byte{1, 2, 3},
		}
		encoded := reply.EncodeMessage(binary.LittleEndian)
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
		reply := &GetMotionEventsReply{
			Sequence: 2,
			NEvents:  1,
			Events: []TimeCoord{
				{
					Time: 123,
					X:    10,
					Y:    20,
				},
			},
		}
		encoded := reply.EncodeMessage(binary.LittleEndian)
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
		reply := &QueryTextExtentsReply{
			Sequence:       3,
			DrawDirection:  0,
			FontAscent:     10,
			FontDescent:    2,
			OverallAscent:  11,
			OverallDescent: 3,
			OverallWidth:   100,
			OverallLeft:    -5,
			OverallRight:   95,
		}
		encoded := reply.EncodeMessage(binary.LittleEndian)
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
		reply := &ListExtensionsReply{
			Sequence: 4,
			NNames:   2,
			Names:    []string{"ext1", "ext2"},
		}
		encoded := reply.EncodeMessage(binary.LittleEndian)

		var data []byte
		data = append(data, 4, 'e', 'x', 't', '1')
		data = append(data, 4, 'e', 'x', 't', '2')

		expected := make([]byte, 32+len(data)+PadLen(len(data)))
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
		reply := &GetWindowAttributesReply{
			Sequence:           10,
			BackingStore:       1,
			VisualID:           2,
			Class:              3,
			BitGravity:         4,
			WinGravity:         5,
			BackingPlanes:      6,
			BackingPixel:       7,
			SaveUnder:          1,
			MapIsInstalled:     1,
			MapState:           2,
			OverrideRedirect:   1,
			Colormap:           8,
			AllEventMasks:      9,
			YourEventMask:      10,
			DoNotPropagateMask: 11,
		}
		encoded := reply.EncodeMessage(binary.LittleEndian)
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
		reply := &GetGeometryReply{
			Sequence:    11,
			Depth:       1,
			Root:        2,
			X:           3,
			Y:           4,
			Width:       5,
			Height:      6,
			BorderWidth: 7,
		}
		encoded := reply.EncodeMessage(binary.LittleEndian)
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
		reply := &InternAtomReply{
			Sequence: 12,
			Atom:     1,
		}
		encoded := reply.EncodeMessage(binary.LittleEndian)
		expected := make([]byte, 32)
		expected[0] = 1
		binary.LittleEndian.PutUint16(expected[2:4], 12)
		binary.LittleEndian.PutUint32(expected[8:12], 1)
		if !bytes.Equal(encoded, expected) {
			t.Errorf("InternAtomReply encoding failed. Got %v, want %v", encoded, expected)
		}
	})

	t.Run("GetAtomName", func(t *testing.T) {
		reply := &GetAtomNameReply{
			Sequence:   13,
			NameLength: 4,
			Name:       "test",
		}
		encoded := reply.EncodeMessage(binary.LittleEndian)
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
		reply := &GetPropertyReply{
			Sequence:              14,
			Format:                8,
			PropertyType:          1,
			BytesAfter:            2,
			ValueLenInFormatUnits: 4,
			Value:                 []byte{1, 2, 3, 4},
		}
		encoded := reply.EncodeMessage(binary.LittleEndian)
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
		reply := &ListPropertiesReply{
			Sequence:      15,
			NumProperties: 3,
			Atoms:         []uint32{1, 2, 3},
		}
		encoded := reply.EncodeMessage(binary.LittleEndian)
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
		reply := &GetSelectionOwnerReply{
			Sequence: 16,
			Owner:    1,
		}
		encoded := reply.EncodeMessage(binary.LittleEndian)
		expected := make([]byte, 32)
		expected[0] = 1
		binary.LittleEndian.PutUint16(expected[2:4], 16)
		binary.LittleEndian.PutUint32(expected[8:12], 1)
		if !bytes.Equal(encoded, expected) {
			t.Errorf("GetSelectionOwnerReply encoding failed. Got %v, want %v", encoded, expected)
		}
	})

	t.Run("GrabPointer", func(t *testing.T) {
		reply := &GrabPointerReply{
			Sequence: 17,
			Status:   1,
		}
		encoded := reply.EncodeMessage(binary.LittleEndian)
		expected := make([]byte, 32)
		expected[0] = 1
		expected[1] = 1
		binary.LittleEndian.PutUint16(expected[2:4], 17)
		if !bytes.Equal(encoded, expected) {
			t.Errorf("GrabPointerReply encoding failed. Got %v, want %v", encoded, expected)
		}
	})

	t.Run("GrabKeyboard", func(t *testing.T) {
		reply := &GrabKeyboardReply{
			Sequence: 18,
			Status:   1,
		}
		encoded := reply.EncodeMessage(binary.LittleEndian)
		expected := make([]byte, 32)
		expected[0] = 1
		expected[1] = 1
		binary.LittleEndian.PutUint16(expected[2:4], 18)
		if !bytes.Equal(encoded, expected) {
			t.Errorf("GrabKeyboardReply encoding failed. Got %v, want %v", encoded, expected)
		}
	})

	t.Run("QueryPointer", func(t *testing.T) {
		reply := &QueryPointerReply{
			Sequence:   19,
			SameScreen: true,
			Root:       1,
			Child:      2,
			RootX:      3,
			RootY:      4,
			WinX:       5,
			WinY:       6,
			Mask:       7,
		}
		encoded := reply.EncodeMessage(binary.LittleEndian)
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
		reply := &TranslateCoordsReply{
			Sequence:   20,
			SameScreen: true,
			Child:      1,
			DstX:       2,
			DstY:       3,
		}
		encoded := reply.EncodeMessage(binary.LittleEndian)
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
		reply := &GetInputFocusReply{
			Sequence: 21,
			RevertTo: 1,
			Focus:    2,
		}
		encoded := reply.EncodeMessage(binary.LittleEndian)
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
		reply := &QueryFontReply{
			Sequence:       22,
			MinCharOrByte2: 1,
			MaxCharOrByte2: 2,
			DefaultChar:    3,
			DrawDirection:  1,
			MinByte1:       1,
			MaxByte1:       2,
			AllCharsExist:  true,
			FontAscent:     10,
			FontDescent:    2,
			CharInfos: []XCharInfo{
				{1, 2, 3, 4, 5, 6},
			},
		}
		encoded := reply.EncodeMessage(binary.LittleEndian)
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
		reply := &ListFontsReply{
			Sequence:  23,
			FontNames: []string{"test", "test2"},
		}
		encoded := reply.EncodeMessage(binary.LittleEndian)
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
		reply := &GetImageReply{
			Sequence:  24,
			Depth:     24,
			VisualID:  1,
			ImageData: []byte{1, 2, 3, 4},
		}
		encoded := reply.EncodeMessage(binary.LittleEndian)
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
		reply := &AllocColorReply{
			Sequence: 25,
			Red:      1,
			Green:    2,
			Blue:     3,
			Pixel:    4,
		}
		encoded := reply.EncodeMessage(binary.LittleEndian)
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
		reply := &ListInstalledColormapsReply{
			Sequence:     26,
			NumColormaps: 3,
			Colormaps:    []uint32{1, 2, 3},
		}
		encoded := reply.EncodeMessage(binary.LittleEndian)
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
		reply := &QueryColorsReply{
			Sequence: 27,
			Colors: []XColorItem{
				{Pixel: 0, Red: 1, Green: 2, Blue: 3, Flags: 0},
				{Pixel: 0, Red: 4, Green: 5, Blue: 6, Flags: 0},
			},
		}
		encoded := reply.EncodeMessage(binary.LittleEndian)
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
		reply := &LookupColorReply{
			Sequence:   28,
			Red:        1,
			Green:      2,
			Blue:       3,
			ExactRed:   4,
			ExactGreen: 5,
			ExactBlue:  6,
		}
		encoded := reply.EncodeMessage(binary.LittleEndian)
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
		reply := &QueryBestSizeReply{
			Sequence: 29,
			Width:    1,
			Height:   2,
		}
		encoded := reply.EncodeMessage(binary.LittleEndian)
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
		reply := &QueryExtensionReply{
			Sequence:    30,
			Present:     true,
			MajorOpcode: 1,
			FirstEvent:  2,
			FirstError:  3,
		}
		encoded := reply.EncodeMessage(binary.LittleEndian)
		expected := make([]byte, 32)
		expected[0] = 1
		binary.LittleEndian.PutUint16(expected[2:4], 30)
		expected[8] = 1
		expected[9] = 1
		expected[10] = 2
		expected[11] = 3
		if !bytes.Equal(encoded, expected) {
			t.Errorf("QueryExtensionReply encoding failed. Got %v, want %v", encoded, expected)
		}
	})

	t.Run("SetPointerMapping", func(t *testing.T) {
		reply := &SetPointerMappingReply{
			Sequence: 31,
		}
		encoded := reply.EncodeMessage(binary.LittleEndian)
		expected := make([]byte, 32)
		expected[0] = 1
		binary.LittleEndian.PutUint16(expected[2:4], 31)
		if !bytes.Equal(encoded, expected) {
			t.Errorf("SetPointerMappingReply encoding failed. Got %v, want %v", encoded, expected)
		}
	})

	t.Run("GetPointerMapping", func(t *testing.T) {
		reply := &GetPointerMappingReply{
			Sequence: 32,
			Length:   4,
			PMap:     []byte{1, 2, 3, 4},
		}
		encoded := reply.EncodeMessage(binary.LittleEndian)
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
		reply := &GetKeyboardMappingReply{
			Sequence: 33,
			KeySyms:  []uint32{1, 2, 3},
		}
		encoded := reply.EncodeMessage(binary.LittleEndian)
		expected := make([]byte, 44)
		expected[0] = 1
		expected[1] = 1
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
		reply := &GetScreenSaverReply{
			Sequence:    34,
			Timeout:     1,
			Interval:    2,
			PreferBlank: 1,
			AllowExpose: 1,
		}
		encoded := reply.EncodeMessage(binary.LittleEndian)
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
		reply := &ListHostsReply{
			Sequence: 35,
			NumHosts: 1,
			Hosts: []Host{
				{
					Family: 1,
					Data:   []byte{1, 2, 3, 4},
				},
			},
		}
		encoded := reply.EncodeMessage(binary.LittleEndian)
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
		reply := &SetModifierMappingReply{
			Sequence: 36,
		}
		encoded := reply.EncodeMessage(binary.LittleEndian)
		expected := make([]byte, 32)
		expected[0] = 1
		binary.LittleEndian.PutUint16(expected[2:4], 36)
		if !bytes.Equal(encoded, expected) {
			t.Errorf("SetModifierMappingReply encoding failed. Got %v, want %v", encoded, expected)
		}
	})

	t.Run("GetModifierMapping", func(t *testing.T) {
		reply := &GetModifierMappingReply{
			Sequence: 37,
			KeyCodes: []KeyCode{1, 2, 3, 4},
		}
		encoded := reply.EncodeMessage(binary.LittleEndian)
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
		reply := &SetPointerMappingReply{
			Sequence: 31,
			Status:   1,
		}
		encoded := reply.EncodeMessage(binary.LittleEndian)
		expected := make([]byte, 32)
		expected[0] = 1
		expected[1] = 1
		binary.LittleEndian.PutUint16(expected[2:4], 31)
		if !bytes.Equal(encoded, expected) {
			t.Errorf("SetPointerMappingReply encoding failed. Got %v, want %v", encoded, expected)
		}
	})

	t.Run("SetModifierMapping", func(t *testing.T) {
		reply := &SetModifierMappingReply{
			Sequence: 36,
			Status:   1,
		}
		encoded := reply.EncodeMessage(binary.LittleEndian)
		expected := make([]byte, 32)
		expected[0] = 1
		expected[1] = 1
		binary.LittleEndian.PutUint16(expected[2:4], 36)
		if !bytes.Equal(encoded, expected) {
			t.Errorf("SetModifierMappingReply encoding failed. Got %v, want %v", encoded, expected)
		}
	})

	t.Run("GetPointerControl", func(t *testing.T) {
		reply := &GetPointerControlReply{
			Sequence:         1,
			AccelNumerator:   2,
			AccelDenominator: 3,
			Threshold:        4,
		}
		encoded := reply.EncodeMessage(binary.LittleEndian)
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
}

func int32ToUint32(i int32) uint32 {
	return uint32(i)
}
