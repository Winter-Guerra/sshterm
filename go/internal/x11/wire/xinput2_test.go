//go:build x11

package wire

import (
	"encoding/binary"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var order = binary.LittleEndian

func TestParseXIQueryVersion(t *testing.T) {
	// https://www.x.org/releases/X11R7.7/doc/inputproto/XI2proto.txt
	// 2.3. XIQueryVersion
	t.Run("valid request", func(t *testing.T) {
		raw := []byte{byte(XIQueryVersion), 0, 2, 0, 2, 0}
		req, err := ParseXInputRequest(binary.LittleEndian, raw[0], raw[2:], 1)
		require.NoError(t, err)
		xiReq, ok := req.(*XIQueryVersionRequest)
		require.True(t, ok)
		assert.Equal(t, uint16(2), xiReq.MajorVersion)
		assert.Equal(t, uint16(2), xiReq.MinorVersion)
	})

	t.Run("invalid request", func(t *testing.T) {
		raw := []byte{byte(XIQueryVersion), 0, 2, 0}
		_, err := ParseXInputRequest(binary.LittleEndian, raw[0], raw[2:], 1)
		var target Error
		require.ErrorAs(t, err, &target)
		assert.Equal(t, byte(LengthErrorCode), target.Code())
	})
}

func TestParseXIChangeProperty(t *testing.T) {
	// https://www.x.org/releases/X11R7.7/doc/inputproto/XI2proto.txt
	// 5.18. XIChangeProperty
	t.Run("valid request", func(t *testing.T) {
		body := []byte{
			2, 0, // deviceid
			0,          // mode
			8,          // format
			3, 0, 0, 0, // property
			4, 0, 0, 0, // type
			1, 0, 0, 0, // num_items
			5, 0, 0, 0, // item + padding
		}

		req, err := ParseXInputRequest(binary.LittleEndian, XIChangeProperty, body, 1)
		require.NoError(t, err)
		xiReq, ok := req.(*XIChangePropertyRequest)
		require.True(t, ok)
		assert.Equal(t, uint16(2), xiReq.DeviceID)
		assert.Equal(t, byte(0), xiReq.Mode)
		assert.Equal(t, byte(8), xiReq.Format)
		assert.Equal(t, uint32(3), xiReq.Property)
		assert.Equal(t, uint32(4), xiReq.Type)
		assert.Equal(t, uint32(1), xiReq.NumItems)
		assert.Equal(t, []byte{5}, xiReq.Data)
	})
}

func TestParseXIListProperties(t *testing.T) {
	// https://www.x.org/releases/X11R7.7/doc/inputproto/XI2proto.txt
	// 5.17. XIListProperties
	t.Run("valid request", func(t *testing.T) {
		body := []byte{
			2, 0, // deviceid
			0, 0, // pad
		}
		require.Len(t, body, 4)

		req, err := ParseXInputRequest(binary.LittleEndian, XIListProperties, body, 1)
		require.NoError(t, err)
		xiReq, ok := req.(*XIListPropertiesRequest)
		require.True(t, ok)
		assert.Equal(t, uint16(2), xiReq.DeviceID)
	})

	t.Run("invalid request", func(t *testing.T) {
		body := []byte{0}
		_, err := ParseXInputRequest(binary.LittleEndian, XIListProperties, body, 1)
		var target Error
		require.ErrorAs(t, err, &target)
		assert.Equal(t, byte(LengthErrorCode), target.Code())
	})
}

func TestParseXIPassiveUngrabDevice(t *testing.T) {
	// https://www.x.org/releases/X11R7.7/doc/inputproto/XI2proto.txt
	// 5.16. XIPassiveUngrabDevice
	t.Run("valid request", func(t *testing.T) {
		body := []byte{
			2, 0, // deviceid
			0, 0, // pad
			1, 0, 0, 0, // grab_window
			4, 0, 0, 0, // detail
			1, 0, // num_modifiers
			0, 0, // pad
			1,    // grab_type
			0,    // pad
			0, 0, // pad
			5, 0, 0, 0, // modifiers
		}
		require.Len(t, body, 24)

		req, err := ParseXInputRequest(binary.LittleEndian, XIPassiveUngrabDevice, body, 1)
		require.NoError(t, err)
		xiReq, ok := req.(*XIPassiveUngrabDeviceRequest)
		require.True(t, ok)
		assert.Equal(t, uint16(2), xiReq.DeviceID)
		assert.Equal(t, Window(1), xiReq.GrabWindow)
		assert.Equal(t, uint32(4), xiReq.Detail)
		assert.Equal(t, uint16(1), xiReq.NumModifiers)
		assert.Equal(t, byte(1), xiReq.GrabType)
		assert.Equal(t, []byte{5, 0, 0, 0}, xiReq.Modifiers)
	})

	t.Run("invalid request", func(t *testing.T) {
		body := []byte{0}
		_, err := ParseXInputRequest(binary.LittleEndian, XIPassiveUngrabDevice, body, 1)
		var target Error
		require.ErrorAs(t, err, &target)
		assert.Equal(t, byte(LengthErrorCode), target.Code())
	})
}

func TestParseXIPassiveGrabDevice(t *testing.T) {
	// https://www.x.org/releases/X11R7.7/doc/inputproto/XI2proto.txt
	// 5.15. XIPassiveGrabDevice
	t.Run("valid request", func(t *testing.T) {
		body := []byte{
			2, 0, // deviceid
			0, 0, // pad
			1, 0, 0, 0, // grab_window
			0, 0, 0, 0, // time
			3, 0, 0, 0, // cursor
			4, 0, 0, 0, // detail
			1, 0, // num_modifiers
			0, 0, // pad
			1,          // grab_type
			1,          // grab_mode
			1,          // paired_device_mode
			1,          // owner_events
			5, 0, 0, 0, // modifiers
		}
		require.Len(t, body, 32)

		req, err := ParseXInputRequest(binary.LittleEndian, XIPassiveGrabDevice, body, 1)
		require.NoError(t, err)
		xiReq, ok := req.(*XIPassiveGrabDeviceRequest)
		require.True(t, ok)
		assert.Equal(t, uint16(2), xiReq.DeviceID)
		assert.Equal(t, Window(1), xiReq.GrabWindow)
		assert.Equal(t, uint32(0), xiReq.Time)
		assert.Equal(t, uint32(3), xiReq.Cursor)
		assert.Equal(t, uint32(4), xiReq.Detail)
		assert.Equal(t, uint16(1), xiReq.NumModifiers)
		assert.Equal(t, byte(1), xiReq.GrabType)
		assert.Equal(t, byte(1), xiReq.GrabMode)
		assert.Equal(t, byte(1), xiReq.PairedDeviceMode)
		assert.True(t, xiReq.OwnerEvents)
		assert.Equal(t, []byte{5, 0, 0, 0}, xiReq.Modifiers)
	})

	t.Run("invalid request", func(t *testing.T) {
		body := []byte{0}
		_, err := ParseXInputRequest(binary.LittleEndian, XIPassiveGrabDevice, body, 1)
		var target Error
		require.ErrorAs(t, err, &target)
		assert.Equal(t, byte(LengthErrorCode), target.Code())
	})
}

func TestParseXIAllowEvents(t *testing.T) {
	// https://www.x.org/releases/X11R7.7/doc/inputproto/XI2proto.txt
	// 5.14. XIAllowEvents
	t.Run("valid request", func(t *testing.T) {
		body := []byte{
			2, 0, // deviceid
			0,          // event_mode
			0,          // pad
			0, 0, 0, 0, // time
			1, 0, 0, 0, // touchid
			1, 0, 0, 0, // grab_window
		}
		require.Len(t, body, 16)

		req, err := ParseXInputRequest(binary.LittleEndian, XIAllowEvents, body, 1)
		require.NoError(t, err)
		xiReq, ok := req.(*XIAllowEventsRequest)
		require.True(t, ok)
		assert.Equal(t, uint16(2), xiReq.DeviceID)
		assert.Equal(t, byte(0), xiReq.EventMode)
		assert.Equal(t, uint32(0), xiReq.Time)
		assert.Equal(t, uint32(1), xiReq.TouchID)
		assert.Equal(t, Window(1), xiReq.GrabWindow)
	})

	t.Run("invalid request", func(t *testing.T) {
		body := []byte{0}
		_, err := ParseXInputRequest(binary.LittleEndian, XIAllowEvents, body, 1)
		var target Error
		require.ErrorAs(t, err, &target)
		assert.Equal(t, byte(LengthErrorCode), target.Code())
	})
}

func TestParseXIUngrabDevice(t *testing.T) {
	// https://www.x.org/releases/X11R7.7/doc/inputproto/XI2proto.txt
	// 5.13. XIUngrabDevice
	t.Run("valid request", func(t *testing.T) {
		body := []byte{
			2, 0, // deviceid
			0, 0, // pad
			0, 0, 0, 0, // time
		}
		require.Len(t, body, 8)

		req, err := ParseXInputRequest(binary.LittleEndian, XIUngrabDevice, body, 1)
		require.NoError(t, err)
		xiReq, ok := req.(*XIUngrabDeviceRequest)
		require.True(t, ok)
		assert.Equal(t, uint16(2), xiReq.DeviceID)
		assert.Equal(t, uint32(0), xiReq.Time)
	})

	t.Run("invalid request", func(t *testing.T) {
		body := []byte{0}
		_, err := ParseXInputRequest(binary.LittleEndian, XIUngrabDevice, body, 1)
		var target Error
		require.ErrorAs(t, err, &target)
		assert.Equal(t, byte(LengthErrorCode), target.Code())
	})
}

func TestParseXIGrabDevice(t *testing.T) {
	// https://www.x.org/releases/X11R7.7/doc/inputproto/XI2proto.txt
	// 5.12. XIGrabDevice
	t.Run("valid request", func(t *testing.T) {
		body := []byte{
			2, 0, // deviceid
			1, 0, 0, 0, // grab_window
			0, 0, 0, 0, // time
			3, 0, 0, 0, // cursor
			1,    // grab_mode
			1,    // paired_device_mode
			1,    // owner_events
			0,    // pad
			1, 0, // mask_len = 1 (4 bytes)
			5, 0, 0, 0, // mask
		}
		require.Len(t, body, 24)

		req, err := ParseXInputRequest(binary.LittleEndian, XIGrabDevice, body, 1)
		require.NoError(t, err)
		xiReq, ok := req.(*XIGrabDeviceRequest)
		require.True(t, ok)
		assert.Equal(t, uint16(2), xiReq.DeviceID)
		assert.Equal(t, Window(1), xiReq.GrabWindow)
		assert.Equal(t, uint32(0), xiReq.Time)
		assert.Equal(t, uint32(3), xiReq.Cursor)
		assert.Equal(t, byte(1), xiReq.GrabMode)
		assert.Equal(t, byte(1), xiReq.PairedDeviceMode)
		assert.True(t, xiReq.OwnerEvents)
		assert.Equal(t, uint16(1), xiReq.MaskLen)
		assert.Equal(t, []byte{5, 0, 0, 0}, xiReq.Mask)
	})

	t.Run("invalid request", func(t *testing.T) {
		body := []byte{0}
		_, err := ParseXInputRequest(binary.LittleEndian, XIGrabDevice, body, 1)
		var target Error
		require.ErrorAs(t, err, &target)
		assert.Equal(t, byte(LengthErrorCode), target.Code())
	})
}

func TestParseXIGetFocus(t *testing.T) {
	// https://www.x.org/releases/X11R7.7/doc/inputproto/XI2proto.txt
	// 5.11. XIGetFocus
	t.Run("valid request", func(t *testing.T) {
		body := []byte{
			2, 0, // deviceid
			0, 0, // pad
		}
		require.Len(t, body, 4)

		req, err := ParseXInputRequest(binary.LittleEndian, XIGetFocus, body, 1)
		require.NoError(t, err)
		xiReq, ok := req.(*XIGetFocusRequest)
		require.True(t, ok)
		assert.Equal(t, uint16(2), xiReq.DeviceID)
	})

	t.Run("invalid request", func(t *testing.T) {
		body := []byte{0}
		_, err := ParseXInputRequest(binary.LittleEndian, XIGetFocus, body, 1)
		var target Error
		require.ErrorAs(t, err, &target)
		assert.Equal(t, byte(LengthErrorCode), target.Code())
	})
}

func TestParseXISetFocus(t *testing.T) {
	// https://www.x.org/releases/X11R7.7/doc/inputproto/XI2proto.txt
	// 5.10. XISetFocus
	t.Run("valid request", func(t *testing.T) {
		body := []byte{
			2, 0, // deviceid
			0, 0, // pad
			1, 0, 0, 0, // focus
			0, 0, 0, 0, // time
		}
		require.Len(t, body, 12)

		req, err := ParseXInputRequest(binary.LittleEndian, XISetFocus, body, 1)
		require.NoError(t, err)
		xiReq, ok := req.(*XISetFocusRequest)
		require.True(t, ok)
		assert.Equal(t, uint16(2), xiReq.DeviceID)
		assert.Equal(t, Window(1), xiReq.Focus)
		assert.Equal(t, uint32(0), xiReq.Time)
	})

	t.Run("invalid request", func(t *testing.T) {
		body := []byte{0}
		_, err := ParseXInputRequest(binary.LittleEndian, XISetFocus, body, 1)
		var target Error
		require.ErrorAs(t, err, &target)
		assert.Equal(t, byte(LengthErrorCode), target.Code())
	})
}

func TestParseXIQueryDevice(t *testing.T) {
	// https://www.x.org/releases/X11R7.7/doc/inputproto/XI2proto.txt
	// 5.9. XIQueryDevice
	t.Run("valid request", func(t *testing.T) {
		body := []byte{
			2, 0, // deviceid
			0, 0, // pad
		}
		require.Len(t, body, 4)

		req, err := ParseXInputRequest(binary.LittleEndian, XIQueryDevice, body, 1)
		require.NoError(t, err)
		xiReq, ok := req.(*XIQueryDeviceRequest)
		require.True(t, ok)
		assert.Equal(t, uint16(2), xiReq.DeviceID)
	})

	t.Run("invalid request", func(t *testing.T) {
		body := []byte{0}
		_, err := ParseXInputRequest(binary.LittleEndian, XIQueryDevice, body, 1)
		var target Error
		require.ErrorAs(t, err, &target)
		assert.Equal(t, byte(LengthErrorCode), target.Code())
	})
}

func TestParseXISelectEvents(t *testing.T) {
	// https://www.x.org/releases/X11R7.7/doc/inputproto/XI2proto.txt
	// 5.8. XISelectEvents
	t.Run("valid request", func(t *testing.T) {
		body := []byte{
			1, 0, 0, 0, // window
			1, 0, // num_masks
			0, 0, // pad
			2, 0, // deviceid
			1, 0, // mask_len = 1 (4 bytes)
			4, 0, 0, 0, // mask
		}
		require.Len(t, body, 16)

		req, err := ParseXInputRequest(binary.LittleEndian, XISelectEvents, body, 1)
		require.NoError(t, err)
		xiReq, ok := req.(*XISelectEventsRequest)
		require.True(t, ok)
		assert.Equal(t, Window(1), xiReq.Window)
		require.Len(t, xiReq.Masks, 1)
		assert.Equal(t, uint16(2), xiReq.Masks[0].DeviceID)
		assert.Equal(t, uint16(1), xiReq.Masks[0].MaskLen)
		assert.Equal(t, []byte{4, 0, 0, 0}, xiReq.Masks[0].Mask)
	})

	t.Run("invalid request", func(t *testing.T) {
		body := []byte{0}
		_, err := ParseXInputRequest(binary.LittleEndian, XISelectEvents, body, 1)
		var target Error
		require.ErrorAs(t, err, &target)
		assert.Equal(t, byte(LengthErrorCode), target.Code())
	})
}

func TestParseXIGetClientPointer(t *testing.T) {
	// https://www.x.org/releases/X11R7.7/doc/inputproto/XI2proto.txt
	// 5.7. XIGetClientPointer
	t.Run("valid request", func(t *testing.T) {
		body := []byte{
			1, 0, 0, 0, // window
		}
		require.Len(t, body, 4)

		req, err := ParseXInputRequest(binary.LittleEndian, XIGetClientPointer, body, 1)
		require.NoError(t, err)
		xiReq, ok := req.(*XIGetClientPointerRequest)
		require.True(t, ok)
		assert.Equal(t, Window(1), xiReq.Window)
	})

	t.Run("invalid request", func(t *testing.T) {
		body := []byte{0}
		_, err := ParseXInputRequest(binary.LittleEndian, XIGetClientPointer, body, 1)
		var target Error
		require.ErrorAs(t, err, &target)
		assert.Equal(t, byte(LengthErrorCode), target.Code())
	})
}

func TestParseXISetClientPointer(t *testing.T) {
	// https://www.x.org/releases/X11R7.7/doc/inputproto/XI2proto.txt
	// 5.6. XISetClientPointer
	t.Run("valid request", func(t *testing.T) {
		body := []byte{
			2, 0, // deviceid
			0, 0, // pad
			1, 0, 0, 0, // window
		}
		require.Len(t, body, 8)

		req, err := ParseXInputRequest(binary.LittleEndian, XISetClientPointer, body, 1)
		require.NoError(t, err)
		xiReq, ok := req.(*XISetClientPointerRequest)
		require.True(t, ok)
		assert.Equal(t, uint16(2), xiReq.DeviceID)
		assert.Equal(t, Window(1), xiReq.Window)
	})

	t.Run("invalid request", func(t *testing.T) {
		body := []byte{0}
		_, err := ParseXInputRequest(binary.LittleEndian, XISetClientPointer, body, 1)
		var target Error
		require.ErrorAs(t, err, &target)
		assert.Equal(t, byte(LengthErrorCode), target.Code())
	})
}

func TestParseXIChangeHierarchy(t *testing.T) {
	// https://www.x.org/releases/X11R7.7/doc/inputproto/XI2proto.txt
	// 5.5. XIChangeHierarchy
	t.Run("valid request", func(t *testing.T) {
		body := []byte{
			1, 0, // num_changes
			0, 0, // pad
			3, 0, // type XIAttachSlave
			8, 0, // length
			2, 0, // deviceid
			4, 0, // masterid
		}
		require.Len(t, body, 12)

		req, err := ParseXInputRequest(binary.LittleEndian, XIChangeHierarchy, body, 1)
		require.NoError(t, err)
		xiReq, ok := req.(*XIChangeHierarchyRequest)
		require.True(t, ok)
		assert.Equal(t, uint16(1), xiReq.NumChanges)
		require.Len(t, xiReq.Changes, 1)
		change, ok := xiReq.Changes[0].(*XIAttachSlave)
		require.True(t, ok)
		assert.Equal(t, uint16(3), change.Type)
		assert.Equal(t, uint16(8), change.Length)
		assert.Equal(t, uint16(2), change.DeviceID)
		assert.Equal(t, uint16(4), change.MasterID)
	})

	t.Run("invalid request", func(t *testing.T) {
		body := []byte{0}
		_, err := ParseXInputRequest(binary.LittleEndian, XIChangeHierarchy, body, 1)
		var target Error
		require.ErrorAs(t, err, &target)
		assert.Equal(t, byte(LengthErrorCode), target.Code())
	})
}

func TestParseXIChangeCursor(t *testing.T) {
	// https://www.x.org/releases/X11R7.7/doc/inputproto/XI2proto.txt
	// 5.4. XIChangeCursor
	t.Run("valid request", func(t *testing.T) {
		body := []byte{
			2, 0, // deviceid
			0, 0, // pad
			1, 0, 0, 0, // window
			3, 0, 0, 0, // cursor
		}
		require.Len(t, body, 12)

		req, err := ParseXInputRequest(binary.LittleEndian, XIChangeCursor, body, 1)
		require.NoError(t, err)
		xiReq, ok := req.(*XIChangeCursorRequest)
		require.True(t, ok)
		assert.Equal(t, uint16(2), xiReq.DeviceID)
		assert.Equal(t, Window(1), xiReq.Window)
		assert.Equal(t, uint32(3), xiReq.Cursor)
	})

	t.Run("invalid request", func(t *testing.T) {
		body := []byte{0}
		_, err := ParseXInputRequest(binary.LittleEndian, XIChangeCursor, body, 1)
		var target Error
		require.ErrorAs(t, err, &target)
		assert.Equal(t, byte(LengthErrorCode), target.Code())
	})
}

func TestParseXIWarpPointer(t *testing.T) {
	// https://www.x.org/releases/X11R7.7/doc/inputproto/XI2proto.txt
	// 5.2. XIWarpPointer
	t.Run("valid request", func(t *testing.T) {
		body := []byte{
			2, 0, // deviceid
			0, 0, // pad
			1, 0, 0, 0, // src-window
			2, 0, 0, 0, // dst-window
			0, 0, 0, 0, // src-x
			0, 0, 0, 0, // src-y
			10, 0, // src-w
			20, 0, // src-h
			0, 0, 0, 0, // dst-x
			0, 0, 0, 0, // dst-y
		}
		require.Len(t, body, 32)

		req, err := ParseXInputRequest(binary.LittleEndian, XIWarpPointer, body, 1)
		require.NoError(t, err)
		xiReq, ok := req.(*XIWarpPointerRequest)
		require.True(t, ok)
		assert.Equal(t, uint16(2), xiReq.DeviceID)
		assert.Equal(t, Window(1), xiReq.SrcWindow)
		assert.Equal(t, Window(2), xiReq.DstWindow)
		assert.Equal(t, int32(0), xiReq.SrcX)
		assert.Equal(t, int32(0), xiReq.SrcY)
		assert.Equal(t, uint16(10), xiReq.SrcW)
		assert.Equal(t, uint16(20), xiReq.SrcH)
		assert.Equal(t, int32(0), xiReq.DstX)
		assert.Equal(t, int32(0), xiReq.DstY)
	})

	t.Run("invalid request", func(t *testing.T) {
		body := []byte{0}
		_, err := ParseXInputRequest(binary.LittleEndian, XIWarpPointer, body, 1)
		var target Error
		require.ErrorAs(t, err, &target)
		assert.Equal(t, byte(LengthErrorCode), target.Code())
	})
}
