//go:build x11

package wire

import (
	"bytes"
	"encoding/binary"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseSelectExtensionEventRequest(t *testing.T) {
	t.Run("valid request", func(t *testing.T) {
		classes := []uint32{10, 20}
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.LittleEndian, uint32(123)) // window
		binary.Write(buf, binary.LittleEndian, uint16(len(classes)))
		buf.Write([]byte{0, 0}) // padding
		for _, class := range classes {
			binary.Write(buf, binary.LittleEndian, class)
		}

		req, err := ParseSelectExtensionEventRequest(binary.LittleEndian, buf.Bytes(), 1)
		assert.NoError(t, err)
		assert.NotNil(t, req)
		assert.Equal(t, Window(123), req.Window)
		assert.Equal(t, classes, req.Classes)
	})

	t.Run("body too short", func(t *testing.T) {
		_, err := ParseSelectExtensionEventRequest(binary.LittleEndian, []byte{1, 2}, 1)
		assert.Error(t, err)
		assert.IsType(t, &LengthError{}, err)
	})

	t.Run("body length mismatch", func(t *testing.T) {
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.LittleEndian, uint32(123)) // window
		binary.Write(buf, binary.LittleEndian, uint16(2))   // num_classes = 2
		buf.Write([]byte{0, 0})                             // padding
		binary.Write(buf, binary.LittleEndian, uint32(10))  // only one class

		_, err := ParseSelectExtensionEventRequest(binary.LittleEndian, buf.Bytes(), 1)
		assert.Error(t, err)
		assert.IsType(t, &LengthError{}, err)
	})
}

func TestParseGrabDeviceKeyRequest(t *testing.T) {
	t.Run("valid request", func(t *testing.T) {
		classes := []uint32{}
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.LittleEndian, uint32(123)) // grab_window
		binary.Write(buf, binary.LittleEndian, uint16(len(classes)))
		buf.WriteByte(1) // owner_events
		buf.WriteByte(1) // this_device_mode
		buf.WriteByte(0) // other_device_mode
		buf.WriteByte(5) // device_id
		binary.Write(buf, binary.LittleEndian, uint16(0)) // modifiers
		buf.WriteByte(10)   // key

		req, err := ParseGrabDeviceKeyRequest(binary.LittleEndian, buf.Bytes(), 1)
		assert.NoError(t, err)
		assert.NotNil(t, req)
		assert.Equal(t, Window(123), req.GrabWindow)
		assert.Equal(t, uint16(len(classes)), req.NumClasses)
		assert.True(t, req.OwnerEvents)
		assert.Equal(t, byte(1), req.ThisDeviceMode)
		assert.Equal(t, byte(0), req.OtherDeviceMode)
		assert.Equal(t, byte(5), req.DeviceID)
		assert.Equal(t, byte(10), req.Key)
		assert.Equal(t, uint16(0), req.Modifiers)
		assert.Equal(t, classes, req.Classes)
	})

	t.Run("invalid length", func(t *testing.T) {
		_, err := ParseGrabDeviceKeyRequest(binary.LittleEndian, []byte{1, 2, 3}, 1)
		assert.Error(t, err)
		assert.IsType(t, &LengthError{}, err)
	})

	t.Run("body length mismatch", func(t *testing.T) {
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.LittleEndian, uint32(123)) // grab_window
		binary.Write(buf, binary.LittleEndian, uint16(2))   // num_classes = 2
		buf.WriteByte(1) // owner_events
		buf.WriteByte(1) // this_device_mode
		buf.WriteByte(0) // other_device_mode
		buf.WriteByte(5) // device_id
		binary.Write(buf, binary.LittleEndian, uint16(0)) // modifiers
		buf.WriteByte(10)   // key
		binary.Write(buf, binary.LittleEndian, uint32(10))  // only one class

		_, err := ParseGrabDeviceKeyRequest(binary.LittleEndian, buf.Bytes(), 1)
		assert.Error(t, err)
		assert.IsType(t, &LengthError{}, err)
	})
}

func TestParseUngrabDeviceKeyRequest(t *testing.T) {
	t.Run("valid request", func(t *testing.T) {
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.LittleEndian, uint32(123)) // grab_window
		binary.Write(buf, binary.LittleEndian, uint16(1))   // modifiers
		buf.WriteByte(5)                                    // device_id
		buf.WriteByte(10)                                   // key
		buf.Write([]byte{0, 0, 0, 0})                       // padding

		req, err := ParseUngrabDeviceKeyRequest(binary.LittleEndian, buf.Bytes(), 1)
		assert.NoError(t, err)
		assert.NotNil(t, req)
		assert.Equal(t, Window(123), req.GrabWindow)
		assert.Equal(t, uint16(1), req.Modifiers)
		assert.Equal(t, byte(5), req.DeviceID)
		assert.Equal(t, byte(10), req.Key)
	})

	t.Run("invalid length", func(t *testing.T) {
		_, err := ParseUngrabDeviceKeyRequest(binary.LittleEndian, []byte{1, 2, 3}, 1)
		assert.Error(t, err)
		assert.IsType(t, &LengthError{}, err)
	})
}

func TestParseGrabDeviceButtonRequest(t *testing.T) {
	t.Run("valid request", func(t *testing.T) {
		classes := []uint32{}
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.LittleEndian, uint32(123)) // grab_window
		binary.Write(buf, binary.LittleEndian, uint16(len(classes)))
		buf.WriteByte(1) // owner_events
		buf.WriteByte(1) // this_device_mode
		buf.WriteByte(0) // other_device_mode
		buf.WriteByte(5) // device_id
		binary.Write(buf, binary.LittleEndian, uint16(1)) // modifiers
		buf.WriteByte(10)   // button

		req, err := ParseGrabDeviceButtonRequest(binary.LittleEndian, buf.Bytes(), 1)
		assert.NoError(t, err)
		assert.NotNil(t, req)
		assert.Equal(t, Window(123), req.GrabWindow)
		assert.Equal(t, uint16(len(classes)), req.NumClasses)
		assert.True(t, req.OwnerEvents)
		assert.Equal(t, byte(1), req.ThisDeviceMode)
		assert.Equal(t, byte(0), req.OtherDeviceMode)
		assert.Equal(t, byte(5), req.DeviceID)
		assert.Equal(t, byte(10), req.Button)
		assert.Equal(t, uint16(1), req.Modifiers)
		assert.Equal(t, classes, req.Classes)
	})
}

func TestParseUngrabDeviceButtonRequest(t *testing.T) {
	t.Run("valid request", func(t *testing.T) {
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.LittleEndian, uint32(123)) // grab_window
		binary.Write(buf, binary.LittleEndian, uint16(1))   // modifiers
		buf.WriteByte(5)                                    // device_id
		buf.WriteByte(10)                                   // button
		buf.Write([]byte{0, 0, 0, 0})                       // padding

		req, err := ParseUngrabDeviceButtonRequest(binary.LittleEndian, buf.Bytes(), 1)
		assert.NoError(t, err)
		assert.NotNil(t, req)
		assert.Equal(t, Window(123), req.GrabWindow)
		assert.Equal(t, uint16(1), req.Modifiers)
		assert.Equal(t, byte(5), req.DeviceID)
		assert.Equal(t, byte(10), req.Button)
	})
}

func TestParseGetDeviceMotionEventsRequest(t *testing.T) {
	t.Run("valid request", func(t *testing.T) {
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.LittleEndian, uint32(100)) // start
		binary.Write(buf, binary.LittleEndian, uint32(200)) // stop
		buf.WriteByte(5)                                    // device id
		buf.Write([]byte{0, 0, 0})                          // padding

		req, err := ParseGetDeviceMotionEventsRequest(binary.LittleEndian, buf.Bytes(), 1)
		assert.NoError(t, err)
		assert.NotNil(t, req)
		assert.Equal(t, uint32(100), req.Start)
		assert.Equal(t, uint32(200), req.Stop)
		assert.Equal(t, byte(5), req.DeviceID)
	})

	t.Run("invalid length", func(t *testing.T) {
		_, err := ParseGetDeviceMotionEventsRequest(binary.LittleEndian, []byte{1, 2, 3}, 1)
		assert.Error(t, err)
		assert.IsType(t, &LengthError{}, err)
	})
}

func TestParseChangeKeyboardDeviceRequest(t *testing.T) {
	t.Run("valid request", func(t *testing.T) {
		buf := new(bytes.Buffer)
		buf.WriteByte(5)           // device id
		buf.Write([]byte{0, 0, 0}) // padding

		req, err := ParseChangeKeyboardDeviceRequest(binary.LittleEndian, buf.Bytes(), 1)
		assert.NoError(t, err)
		assert.NotNil(t, req)
		assert.Equal(t, byte(5), req.DeviceID)
	})

	t.Run("invalid length", func(t *testing.T) {
		_, err := ParseChangeKeyboardDeviceRequest(binary.LittleEndian, []byte{1, 2}, 1)
		assert.Error(t, err)
		assert.IsType(t, &LengthError{}, err)
	})
}

func TestParseListInputDevicesRequest(t *testing.T) {
	t.Run("valid request", func(t *testing.T) {
		req, err := ParseListInputDevicesRequest(binary.LittleEndian, []byte{}, 1)
		assert.NoError(t, err)
		assert.NotNil(t, req)
	})

	t.Run("invalid length", func(t *testing.T) {
		_, err := ParseListInputDevicesRequest(binary.LittleEndian, []byte{1, 2}, 1)
		assert.Error(t, err)
		assert.IsType(t, &LengthError{}, err)
	})
}

func TestParseOpenDeviceRequest(t *testing.T) {
	t.Run("valid request", func(t *testing.T) {
		buf := new(bytes.Buffer)
		buf.WriteByte(5)
		buf.Write([]byte{0, 0, 0})
		req, err := ParseOpenDeviceRequest(binary.LittleEndian, buf.Bytes(), 1)
		assert.NoError(t, err)
		assert.NotNil(t, req)
		assert.Equal(t, byte(5), req.DeviceID)
	})

	t.Run("invalid length", func(t *testing.T) {
		_, err := ParseOpenDeviceRequest(binary.LittleEndian, []byte{1, 2}, 1)
		assert.Error(t, err)
		assert.IsType(t, &LengthError{}, err)
	})
}

func TestParseSetDeviceModeRequest(t *testing.T) {
	t.Run("valid request", func(t *testing.T) {
		buf := new(bytes.Buffer)
		buf.WriteByte(5)
		buf.WriteByte(1)
		buf.Write([]byte{0, 0})
		req, err := ParseSetDeviceModeRequest(binary.LittleEndian, buf.Bytes(), 1)
		assert.NoError(t, err)
		assert.NotNil(t, req)
		assert.Equal(t, byte(5), req.DeviceID)
		assert.Equal(t, byte(1), req.Mode)
	})

	t.Run("invalid length", func(t *testing.T) {
		_, err := ParseSetDeviceModeRequest(binary.LittleEndian, []byte{1, 2}, 1)
		assert.Error(t, err)
		assert.IsType(t, &LengthError{}, err)
	})
}

func TestParseSetDeviceValuatorsRequest(t *testing.T) {
	t.Run("valid request", func(t *testing.T) {
		valuators := []int32{100, 200}
		buf := new(bytes.Buffer)
		buf.WriteByte(5)
		buf.WriteByte(1)
		buf.WriteByte(byte(len(valuators)))
		buf.WriteByte(0)
		for _, v := range valuators {
			binary.Write(buf, binary.LittleEndian, v)
		}
		req, err := ParseSetDeviceValuatorsRequest(binary.LittleEndian, buf.Bytes(), 1)
		assert.NoError(t, err)
		assert.NotNil(t, req)
		assert.Equal(t, byte(5), req.DeviceID)
		assert.Equal(t, byte(1), req.FirstValuator)
		assert.Equal(t, byte(len(valuators)), req.NumValuators)
		assert.Equal(t, valuators, req.Valuators)
	})

	t.Run("invalid length", func(t *testing.T) {
		_, err := ParseSetDeviceValuatorsRequest(binary.LittleEndian, []byte{1, 2}, 1)
		assert.Error(t, err)
		assert.IsType(t, &LengthError{}, err)
	})
}

func TestParseGetDeviceControlRequest(t *testing.T) {
	t.Run("valid request", func(t *testing.T) {
		buf := new(bytes.Buffer)
		buf.WriteByte(5)
		buf.WriteByte(0)
		binary.Write(buf, binary.LittleEndian, uint16(1))
		req, err := ParseGetDeviceControlRequest(binary.LittleEndian, buf.Bytes(), 1)
		assert.NoError(t, err)
		assert.NotNil(t, req)
		assert.Equal(t, byte(5), req.DeviceID)
		assert.Equal(t, uint16(1), req.Control)
	})
}

func TestParseChangeDeviceControlRequest(t *testing.T) {
	t.Run("valid device resolution control", func(t *testing.T) {
		resolutions := []uint32{100, 200}
		control := &DeviceResolutionControl{
			FirstValuator: 1,
			NumValuators:  2,
			Resolutions:   resolutions,
		}
		buf := new(bytes.Buffer)
		buf.WriteByte(10) // device ID
		buf.WriteByte(0)  // padding
		binary.Write(buf, binary.LittleEndian, uint16(DeviceResolution))
		binary.Write(buf, binary.LittleEndian, uint16(8+len(resolutions)*4))
		buf.WriteByte(control.FirstValuator)
		buf.WriteByte(control.NumValuators)
		buf.Write([]byte{0, 0}) // padding
		for _, res := range resolutions {
			binary.Write(buf, binary.LittleEndian, res)
		}

		req, err := ParseChangeDeviceControlRequest(binary.LittleEndian, buf.Bytes(), 1)
		assert.NoError(t, err)
		assert.NotNil(t, req)
		assert.Equal(t, byte(10), req.DeviceID)
		assert.IsType(t, &DeviceResolutionControl{}, req.Control)
		drc := req.Control.(*DeviceResolutionControl)
		assert.Equal(t, control.FirstValuator, drc.FirstValuator)
		assert.Equal(t, control.NumValuators, drc.NumValuators)
		assert.Equal(t, control.Resolutions, drc.Resolutions)
	})

	t.Run("invalid control id", func(t *testing.T) {
		buf := new(bytes.Buffer)
		buf.WriteByte(10) // device ID
		buf.WriteByte(0)  // padding
		binary.Write(buf, binary.LittleEndian, uint16(99)) // invalid control ID
		_, err := ParseChangeDeviceControlRequest(binary.LittleEndian, buf.Bytes(), 1)
		assert.Error(t, err)
		assert.IsType(t, &ValueError{}, err)
		e := err.(Error)
		assert.Equal(t, byte(ValueErrorCode), e.Code())
	})

	t.Run("invalid length", func(t *testing.T) {
		buf := new(bytes.Buffer)
		buf.WriteByte(10) // device ID
		buf.WriteByte(0)  // padding
		binary.Write(buf, binary.LittleEndian, uint16(DeviceResolution))
		binary.Write(buf, binary.LittleEndian, uint16(10)) // invalid length
		_, err := ParseChangeDeviceControlRequest(binary.LittleEndian, buf.Bytes(), 1)
		assert.Error(t, err)
		assert.IsType(t, &LengthError{}, err)
		e := err.(Error)
		assert.Equal(t, byte(LengthErrorCode), e.Code())
	})
}

func TestParseChangePointerDeviceRequest(t *testing.T) {
	t.Run("valid request", func(t *testing.T) {
		buf := new(bytes.Buffer)
		buf.WriteByte(1)           // x axis
		buf.WriteByte(2)           // y axis
		buf.WriteByte(5)           // device id
		buf.WriteByte(0)           // padding

		req, err := ParseChangePointerDeviceRequest(binary.LittleEndian, buf.Bytes(), 1)
		assert.NoError(t, err)
		assert.NotNil(t, req)
		assert.Equal(t, byte(1), req.XAxis)
		assert.Equal(t, byte(2), req.YAxis)
		assert.Equal(t, byte(5), req.DeviceID)
	})

	t.Run("invalid length", func(t *testing.T) {
		_, err := ParseChangePointerDeviceRequest(binary.LittleEndian, []byte{1, 2}, 1)
		assert.Error(t, err)
		assert.IsType(t, &LengthError{}, err)
	})
}

func TestParseGetDeviceFocusRequest(t *testing.T) {
	t.Run("valid request", func(t *testing.T) {
		buf := new(bytes.Buffer)
		buf.WriteByte(5)
		buf.Write([]byte{0, 0, 0})
		req, err := ParseGetDeviceFocusRequest(binary.LittleEndian, buf.Bytes(), 1)
		assert.NoError(t, err)
		assert.NotNil(t, req)
		assert.Equal(t, byte(5), req.DeviceID)
	})
}

func TestParseSetDeviceFocusRequest(t *testing.T) {
	t.Run("valid request", func(t *testing.T) {
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.LittleEndian, uint32(123)) // focus
		binary.Write(buf, binary.LittleEndian, uint32(100)) // time
		buf.WriteByte(1)                                    // revert_to
		buf.WriteByte(5)                                    // device_id
		buf.Write([]byte{0, 0})                             // padding
		req, err := ParseSetDeviceFocusRequest(binary.LittleEndian, buf.Bytes(), 1)
		assert.NoError(t, err)
		assert.NotNil(t, req)
		assert.Equal(t, Window(123), req.Focus)
		assert.Equal(t, uint32(100), req.Time)
		assert.Equal(t, byte(1), req.RevertTo)
		assert.Equal(t, byte(5), req.DeviceID)
	})
}

func TestParseGetFeedbackControlRequest(t *testing.T) {
	t.Run("valid request", func(t *testing.T) {
		buf := new(bytes.Buffer)
		buf.WriteByte(5)
		buf.Write([]byte{0, 0, 0})
		req, err := ParseGetFeedbackControlRequest(binary.LittleEndian, buf.Bytes(), 1)
		assert.NoError(t, err)
		assert.NotNil(t, req)
		assert.Equal(t, byte(5), req.DeviceID)
	})
}

func TestParseChangeFeedbackControlRequest(t *testing.T) {
	t.Run("valid request", func(t *testing.T) {
		controlData := []byte{1, 2, 3, 4}
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.LittleEndian, uint32(1)) // mask
		buf.WriteByte(5)                                  // device_id
		buf.WriteByte(10)                                 // feedback_id
		buf.Write([]byte{0, 0})                           // padding
		buf.Write(controlData)
		req, err := ParseChangeFeedbackControlRequest(binary.LittleEndian, buf.Bytes(), 1)
		assert.NoError(t, err)
		assert.NotNil(t, req)
		assert.Equal(t, uint32(1), req.Mask)
		assert.Equal(t, byte(5), req.DeviceID)
		assert.Equal(t, byte(10), req.ControlID)
		assert.Equal(t, controlData, req.Control)
	})
}

func TestParseGetDeviceKeyMappingRequest(t *testing.T) {
	t.Run("valid request", func(t *testing.T) {
		buf := new(bytes.Buffer)
		buf.WriteByte(5)  // device_id
		buf.WriteByte(10) // first_keycode
		buf.WriteByte(2)  // count
		buf.WriteByte(0)  // padding
		req, err := ParseGetDeviceKeyMappingRequest(binary.LittleEndian, buf.Bytes(), 1)
		assert.NoError(t, err)
		assert.NotNil(t, req)
		assert.Equal(t, byte(5), req.DeviceID)
		assert.Equal(t, byte(10), req.FirstKey)
		assert.Equal(t, byte(2), req.Count)
	})
}

func TestParseChangeDeviceKeyMappingRequest(t *testing.T) {
	t.Run("valid request", func(t *testing.T) {
		keysyms := []uint32{1, 2, 3, 4}
		buf := new(bytes.Buffer)
		buf.WriteByte(5)  // device_id
		buf.WriteByte(10) // first_keycode
		buf.WriteByte(2)  // keysyms_per_keycode
		buf.WriteByte(2)  // keycode_count
		for _, ks := range keysyms {
			binary.Write(buf, binary.LittleEndian, ks)
		}
		req, err := ParseChangeDeviceKeyMappingRequest(binary.LittleEndian, buf.Bytes(), 1)
		assert.NoError(t, err)
		assert.NotNil(t, req)
		assert.Equal(t, byte(5), req.DeviceID)
		assert.Equal(t, byte(10), req.FirstKey)
		assert.Equal(t, byte(2), req.KeysymsPerKeycode)
		assert.Equal(t, byte(2), req.KeycodeCount)
		assert.Equal(t, keysyms, req.Keysyms)
	})
}

func TestParseGetDeviceModifierMappingRequest(t *testing.T) {
	t.Run("valid request", func(t *testing.T) {
		buf := new(bytes.Buffer)
		buf.WriteByte(5)
		buf.Write([]byte{0, 0, 0})
		req, err := ParseGetDeviceModifierMappingRequest(binary.LittleEndian, buf.Bytes(), 1)
		assert.NoError(t, err)
		assert.NotNil(t, req)
		assert.Equal(t, byte(5), req.DeviceID)
	})
}

func TestParseSetDeviceModifierMappingRequest(t *testing.T) {
	t.Run("valid request", func(t *testing.T) {
		keycodes := []byte{1, 2, 3, 4, 5, 6, 7, 8}
		buf := new(bytes.Buffer)
		buf.WriteByte(5) // device_id
		buf.WriteByte(1) // num_keycodes_per_modifier
		buf.Write([]byte{0, 0})
		buf.Write(keycodes)
		req, err := ParseSetDeviceModifierMappingRequest(binary.LittleEndian, buf.Bytes(), 1)
		assert.NoError(t, err)
		assert.NotNil(t, req)
		assert.Equal(t, byte(5), req.DeviceID)
		assert.Equal(t, keycodes, req.Keycodes)
	})
}

func TestParseGetDeviceButtonMappingRequest(t *testing.T) {
	t.Run("valid request", func(t *testing.T) {
		buf := new(bytes.Buffer)
		buf.WriteByte(5)
		buf.Write([]byte{0, 0, 0})
		req, err := ParseGetDeviceButtonMappingRequest(binary.LittleEndian, buf.Bytes(), 1)
		assert.NoError(t, err)
		assert.NotNil(t, req)
		assert.Equal(t, byte(5), req.DeviceID)
	})
}

func TestParseSetDeviceButtonMappingRequest(t *testing.T) {
	t.Run("valid request", func(t *testing.T) {
		buttonMap := []byte{1, 3, 2}
		buf := new(bytes.Buffer)
		buf.WriteByte(5) // device_id
		buf.WriteByte(byte(len(buttonMap)))
		buf.Write([]byte{0, 0})
		buf.Write(buttonMap)
		req, err := ParseSetDeviceButtonMappingRequest(binary.LittleEndian, buf.Bytes(), 1)
		assert.NoError(t, err)
		assert.NotNil(t, req)
		assert.Equal(t, byte(5), req.DeviceID)
		assert.Equal(t, buttonMap, req.Map)
	})
}

func TestParseQueryDeviceStateRequest(t *testing.T) {
	t.Run("valid request", func(t *testing.T) {
		buf := new(bytes.Buffer)
		buf.WriteByte(5)
		buf.Write([]byte{0, 0, 0})
		req, err := ParseQueryDeviceStateRequest(binary.LittleEndian, buf.Bytes(), 1)
		assert.NoError(t, err)
		assert.NotNil(t, req)
		assert.Equal(t, byte(5), req.DeviceID)
	})
}

func TestParseSendExtensionEventRequest(t *testing.T) {
	t.Run("valid request", func(t *testing.T) {
		events := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32}
		classes := []uint32{}
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.LittleEndian, uint32(123)) // destination
		binary.Write(buf, binary.LittleEndian, uint16(len(classes)))
		buf.WriteByte(byte(len(events) / 32)) // num_events
		buf.WriteByte(5)                      // device_id
		buf.WriteByte(1)                      // propagate
		buf.Write([]byte{0, 0, 0})            // padding
		buf.Write(events)

		req, err := ParseSendExtensionEventRequest(binary.LittleEndian, buf.Bytes(), 1)
		assert.NoError(t, err)
		assert.NotNil(t, req)
		assert.Equal(t, Window(123), req.Destination)
		assert.Equal(t, byte(5), req.DeviceID)
		assert.True(t, req.Propagate)
		assert.Equal(t, byte(len(events)/32), req.NumEvents)
		assert.Equal(t, uint16(len(classes)), req.NumClasses)
		assert.Equal(t, events, req.Events)
		assert.Equal(t, classes, req.Classes)
	})
}

func TestParseDeviceBellRequest(t *testing.T) {
	t.Run("valid request", func(t *testing.T) {
		buf := new(bytes.Buffer)
		buf.WriteByte(5)  // device_id
		buf.WriteByte(10) // feedback_id
		buf.WriteByte(1)  // feedback_class
		buf.WriteByte(50) // percent
		req, err := ParseDeviceBellRequest(binary.LittleEndian, buf.Bytes(), 1)
		assert.NoError(t, err)
		assert.NotNil(t, req)
		assert.Equal(t, byte(5), req.DeviceID)
		assert.Equal(t, byte(10), req.FeedbackID)
		assert.Equal(t, byte(1), req.FeedbackClass)
		assert.Equal(t, byte(50), req.Percent)
	})
}

func TestParseXIChangeHierarchyRequest(t *testing.T) {
	t.Run("valid request with detach slave", func(t *testing.T) {
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.LittleEndian, uint16(1)) // num_changes
		binary.Write(buf, binary.LittleEndian, uint16(0)) // padding
		binary.Write(buf, binary.LittleEndian, uint16(4)) // type = XIDetachSlave
		binary.Write(buf, binary.LittleEndian, uint16(8)) // length
		binary.Write(buf, binary.LittleEndian, uint16(5)) // deviceid
		binary.Write(buf, binary.LittleEndian, uint16(0)) // padding

		req, err := ParseXIChangeHierarchyRequest(binary.LittleEndian, buf.Bytes(), 1)
		assert.NoError(t, err)
		assert.NotNil(t, req)
		assert.Equal(t, uint16(1), req.NumChanges)
		assert.Len(t, req.Changes, 1)
		detach, ok := req.Changes[0].(*XIDetachSlave)
		assert.True(t, ok, "Expected XIDetachSlave change")
		assert.Equal(t, uint16(5), detach.DeviceID)
	})
}

func TestGetExtensionVersionReply(t *testing.T) {
	order := binary.LittleEndian
	reply := &GetExtensionVersionReply{
		Sequence:     1,
		MajorVersion: 2,
		MinorVersion: 3,
	}

	encoded := reply.EncodeMessage(order)
	decoded, err := ParseGetExtensionVersionReply(order, encoded)
	if err != nil {
		t.Fatalf("ParseGetExtensionVersionReply failed: %v", err)
	}

	if !reflect.DeepEqual(reply, decoded) {
		t.Errorf("expected %+v, got %+v", reply, decoded)
	}
}

func TestGetDeviceMotionEventsReply(t *testing.T) {
	order := binary.LittleEndian
	reply := &GetDeviceMotionEventsReply{
		Sequence: 1,
		NEvents:  2,
		Events: []TimeCoord{
			{Time: 1, X: 2, Y: 3},
			{Time: 4, X: 5, Y: 6},
		},
	}

	encoded := reply.EncodeMessage(order)
	decoded, err := ParseGetDeviceMotionEventsReply(order, encoded)
	if err != nil {
		t.Fatalf("ParseGetDeviceMotionEventsReply failed: %v", err)
	}

	if !reflect.DeepEqual(reply, decoded) {
		t.Errorf("expected %+v, got %+v", reply, decoded)
	}
}

func TestParseListInputDevicesReply(t *testing.T) {
	order := binary.LittleEndian
	reply := &ListInputDevicesReply{
		Sequence: 1,
		NDevices: 1,
		Devices: []*DeviceInfo{
			{
				Header: DeviceHeader{
					DeviceID:   2,
					DeviceType: 3,
					NumClasses: 1,
					Use:        4,
					Name:       "test",
				},
				Classes: []InputClassInfo{
					&KeyClassInfo{
						NumKeys:    10,
						MinKeycode: 8,
						MaxKeycode: 255,
					},
				},
			},
		},
	}

	encoded := reply.EncodeMessage(order)
	decoded, err := ParseListInputDevicesReply(order, encoded)
	if err != nil {
		t.Fatalf("ParseListInputDevicesReply failed: %v", err)
	}

	if !reflect.DeepEqual(reply, decoded) {
		t.Errorf("expected %+v, got %+v", reply, decoded)
	}
}

func TestQueryDeviceStateReply(t *testing.T) {
	order := binary.LittleEndian
	reply := &QueryDeviceStateReply{
		Sequence:  1,
		NumEvents: 1,
		Classes: []InputClassInfo{
			&KeyClassInfo{
				NumKeys:    10,
				MinKeycode: 8,
				MaxKeycode: 255,
			},
		},
	}

	encoded := reply.EncodeMessage(order)
	decoded, err := ParseQueryDeviceStateReply(order, encoded)
	if err != nil {
		t.Fatalf("ParseQueryDeviceStateReply failed: %v", err)
	}

	if !reflect.DeepEqual(reply, decoded) {
		t.Errorf("expected %+v, got %+v", reply, decoded)
	}
}

func TestGetDeviceButtonMappingReply(t *testing.T) {
	order := binary.LittleEndian
	reply := &GetDeviceButtonMappingReply{
		Sequence: 1,
		Map:      []byte{1, 2, 3, 4},
	}

	encoded := reply.EncodeMessage(order)
	decoded, err := ParseGetDeviceButtonMappingReply(order, encoded)
	if err != nil {
		t.Fatalf("ParseGetDeviceButtonMappingReply failed: %v", err)
	}

	if !reflect.DeepEqual(reply, decoded) {
		t.Errorf("expected %+v, got %+v", reply, decoded)
	}
}

func TestGetDeviceModifierMappingReply(t *testing.T) {
	order := binary.LittleEndian
	reply := &GetDeviceModifierMappingReply{
		Sequence:          1,
		NumKeycodesPerMod: 2,
		Keycodes:          []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16},
	}

	encoded := reply.EncodeMessage(order)
	decoded, err := ParseGetDeviceModifierMappingReply(order, encoded)
	if err != nil {
		t.Fatalf("ParseGetDeviceModifierMappingReply failed: %v", err)
	}

	if !reflect.DeepEqual(reply, decoded) {
		t.Errorf("expected %+v, got %+v", reply, decoded)
	}
}

func TestGetDeviceFocusReply(t *testing.T) {
	order := binary.LittleEndian
	reply := &GetDeviceFocusReply{
		Sequence: 1,
		Focus:    2,
		Time:     3,
		RevertTo: 4,
	}

	encoded := reply.EncodeMessage(order)
	decoded, err := ParseGetDeviceFocusReply(order, encoded)
	if err != nil {
		t.Fatalf("ParseGetDeviceFocusReply failed: %v", err)
	}

	if !reflect.DeepEqual(reply, decoded) {
		t.Errorf("expected %+v, got %+v", reply, decoded)
	}
}

func TestOpenDeviceReply(t *testing.T) {
	order := binary.LittleEndian
	reply := &OpenDeviceReply{
		Sequence: 1,
		Classes: []InputClassInfo{
			&KeyClassInfo{
				NumKeys:    10,
				MinKeycode: 8,
				MaxKeycode: 255,
			},
		},
	}

	encoded := reply.EncodeMessage(order)
	decoded, err := ParseOpenDeviceReply(order, encoded)
	if err != nil {
		t.Fatalf("ParseOpenDeviceReply failed: %v", err)
	}

	if !reflect.DeepEqual(reply, decoded) {
		t.Errorf("expected %+v, got %+v", reply, decoded)
	}
}

func TestSetDeviceModeReply(t *testing.T) {
	order := binary.LittleEndian
	reply := &SetDeviceModeReply{
		Sequence: 1,
		Status:   2,
	}

	encoded := reply.EncodeMessage(order)
	decoded, err := ParseSetDeviceModeReply(order, encoded)
	if err != nil {
		t.Fatalf("ParseSetDeviceModeReply failed: %v", err)
	}

	if !reflect.DeepEqual(reply, decoded) {
		t.Errorf("expected %+v, got %+v", reply, decoded)
	}
}

func TestSetDeviceValuatorsReply(t *testing.T) {
	order := binary.LittleEndian
	reply := &SetDeviceValuatorsReply{
		Sequence: 1,
		Status:   2,
	}

	encoded := reply.EncodeMessage(order)
	decoded, err := ParseSetDeviceValuatorsReply(order, encoded)
	if err != nil {
		t.Fatalf("ParseSetDeviceValuatorsReply failed: %v", err)
	}

	if !reflect.DeepEqual(reply, decoded) {
		t.Errorf("expected %+v, got %+v", reply, decoded)
	}
}

func TestGetDeviceControlReply(t *testing.T) {
	order := binary.LittleEndian
	reply := &GetDeviceControlReply{
		Sequence: 1,
		Control: &DeviceResolutionState{
			NumValuators:   2,
			Resolutions:    []uint32{1, 2},
			MinResolutions: []uint32{1, 2},
			MaxResolutions: []uint32{1, 2},
		},
	}

	encoded := reply.EncodeMessage(order)
	decoded, err := ParseGetDeviceControlReply(order, encoded)
	if err != nil {
		t.Fatalf("ParseGetDeviceControlReply failed: %v", err)
	}

	if !reflect.DeepEqual(reply, decoded) {
		t.Errorf("expected %+v, got %+v", reply, decoded)
	}
}

func TestChangeDeviceControlReply(t *testing.T) {
	order := binary.LittleEndian
	reply := &ChangeDeviceControlReply{
		Sequence: 1,
		Status:   2,
	}

	encoded := reply.EncodeMessage(order)
	decoded, err := ParseChangeDeviceControlReply(order, encoded)
	if err != nil {
		t.Fatalf("ParseChangeDeviceControlReply failed: %v", err)
	}

	if !reflect.DeepEqual(reply, decoded) {
		t.Errorf("expected %+v, got %+v", reply, decoded)
	}
}

func TestGetSelectedExtensionEventsReply(t *testing.T) {
	order := binary.LittleEndian
	reply := &GetSelectedExtensionEventsReply{
		Sequence:          1,
		ThisClientClasses: []uint32{1, 2},
		AllClientsClasses: []uint32{3, 4},
	}

	encoded := reply.EncodeMessage(order)
	decoded, err := ParseGetSelectedExtensionEventsReply(order, encoded)
	if err != nil {
		t.Fatalf("ParseGetSelectedExtensionEventsReply failed: %v", err)
	}

	if !reflect.DeepEqual(reply, decoded) {
		t.Errorf("expected %+v, got %+v", reply, decoded)
	}
}

func TestGetDeviceDontPropagateListReply(t *testing.T) {
	order := binary.LittleEndian
	reply := &GetDeviceDontPropagateListReply{
		Sequence: 1,
		Classes:  []uint32{1, 2},
	}

	encoded := reply.EncodeMessage(order)
	decoded, err := ParseGetDeviceDontPropagateListReply(order, encoded)
	if err != nil {
		t.Fatalf("ParseGetDeviceDontPropagateListReply failed: %v", err)
	}

	if !reflect.DeepEqual(reply, decoded) {
		t.Errorf("expected %+v, got %+v", reply, decoded)
	}
}

func TestCloseDeviceReply(t *testing.T) {
	order := binary.LittleEndian
	reply := &CloseDeviceReply{
		Sequence: 1,
	}

	encoded := reply.EncodeMessage(order)
	decoded, err := ParseCloseDeviceReply(order, encoded)
	if err != nil {
		t.Fatalf("ParseCloseDeviceReply failed: %v", err)
	}

	if !reflect.DeepEqual(reply, decoded) {
		t.Errorf("expected %+v, got %+v", reply, decoded)
	}
}

func TestGrabDeviceReply(t *testing.T) {
	order := binary.LittleEndian
	reply := &GrabDeviceReply{
		Sequence: 1,
		Status:   2,
	}

	encoded := reply.EncodeMessage(order)
	decoded, err := ParseGrabDeviceReply(order, encoded)
	if err != nil {
		t.Fatalf("ParseGrabDeviceReply failed: %v", err)
	}

	if !reflect.DeepEqual(reply, decoded) {
		t.Errorf("expected %+v, got %+v", reply, decoded)
	}
}

func TestGetDeviceKeyMappingReply(t *testing.T) {
	order := binary.LittleEndian
	reply := &GetDeviceKeyMappingReply{
		Sequence:          1,
		KeysymsPerKeycode: 2,
		Keysyms:           []uint32{1, 2},
	}

	encoded := reply.EncodeMessage(order)
	decoded, err := ParseGetDeviceKeyMappingReply(order, encoded)
	if err != nil {
		t.Fatalf("ParseGetDeviceKeyMappingReply failed: %v", err)
	}

	if !reflect.DeepEqual(reply, decoded) {
		t.Errorf("expected %+v, got %+v", reply, decoded)
	}
}

func TestSetDeviceModifierMappingReply(t *testing.T) {
	order := binary.LittleEndian
	reply := &SetDeviceModifierMappingReply{
		Sequence: 1,
		Status:   2,
	}

	encoded := reply.EncodeMessage(order)
	decoded, err := ParseSetDeviceModifierMappingReply(order, encoded)
	if err != nil {
		t.Fatalf("ParseSetDeviceModifierMappingReply failed: %v", err)
	}

	if !reflect.DeepEqual(reply, decoded) {
		t.Errorf("expected %+v, got %+v", reply, decoded)
	}
}

func TestSetDeviceButtonMappingReply(t *testing.T) {
	order := binary.LittleEndian
	reply := &SetDeviceButtonMappingReply{
		Sequence: 1,
		Status:   2,
	}

	encoded := reply.EncodeMessage(order)
	decoded, err := ParseSetDeviceButtonMappingReply(order, encoded)
	if err != nil {
		t.Fatalf("ParseSetDeviceButtonMappingReply failed: %v", err)
	}

	if !reflect.DeepEqual(reply, decoded) {
		t.Errorf("expected %+v, got %+v", reply, decoded)
	}
}
