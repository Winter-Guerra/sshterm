//go:build x11

package x11

import (
	"github.com/c2FmZQ/sshterm/internal/x11/wire"
)

func (s *x11Server) handleXInputRequest(client *x11Client, minorOpcode byte, body []byte, seq uint16) (reply messageEncoder) {
	switch minorOpcode {
	case wire.XGetExtensionVersion:
		return &wire.GetExtensionVersionReply{
			Sequence:     seq,
			MajorVersion: 1,
			MinorVersion: 5,
		}
	case wire.XListInputDevices:
		return &wire.ListInputDevicesReply{
			Sequence: seq,
			Devices:  []*wire.DeviceInfo{virtualPointer, virtualKeyboard},
		}
	case wire.XOpenDevice:
		req, err := wire.ParseOpenDeviceRequest(client.byteOrder, body, seq)
		if err != nil {
			return err.(messageEncoder)
		}

		var selectedDevice *wire.DeviceInfo
		if req.DeviceID == virtualPointer.Header.DeviceID {
			selectedDevice = virtualPointer
		} else if req.DeviceID == virtualKeyboard.Header.DeviceID {
			selectedDevice = virtualKeyboard
		} else {
			return wire.NewError(wire.ValueErrorCode, seq, uint32(req.DeviceID), byte(wire.XInputOpcode), wire.XOpenDevice)
		}

		// Create a new deviceInfo instance for the client, so event masks are not shared.
		newClasses := make([]wire.InputClassInfo, len(selectedDevice.Classes))
		copy(newClasses, selectedDevice.Classes)
		newDeviceInfo := &wire.DeviceInfo{
			Header:     selectedDevice.Header,
			Classes:    newClasses,
			EventMasks: make(map[uint32]uint32),
		}
		client.openDevices[req.DeviceID] = newDeviceInfo
		return &wire.OpenDeviceReply{Sequence: seq, Classes: newDeviceInfo.Classes}

	case wire.XSetDeviceMode:
		req, err := wire.ParseSetDeviceModeRequest(client.byteOrder, body, seq)
		if err != nil {
			return err.(messageEncoder)
		}
		device, ok := client.openDevices[req.DeviceID]
		if !ok {
			return wire.NewError(wire.ValueErrorCode, seq, uint32(req.DeviceID), byte(wire.XInputOpcode), wire.XSetDeviceMode)
		}
		var valuatorInfo *wire.ValuatorClassInfo
		for _, class := range device.Classes {
			if vc, ok := class.(*wire.ValuatorClassInfo); ok {
				valuatorInfo = vc
				break
			}
		}
		if valuatorInfo == nil {
			return wire.NewError(wire.MatchErrorCode, seq, 0, byte(wire.XInputOpcode), wire.XSetDeviceMode)
		}
		valuatorInfo.Mode = req.Mode
		return &wire.SetDeviceModeReply{Sequence: seq, Status: wire.GrabSuccess}

	case wire.XSetDeviceValuators:
		req, err := wire.ParseSetDeviceValuatorsRequest(client.byteOrder, body, seq)
		if err != nil {
			return err.(messageEncoder)
		}
		device, ok := client.openDevices[req.DeviceID]
		if !ok {
			return wire.NewError(wire.ValueErrorCode, seq, uint32(req.DeviceID), byte(wire.XInputOpcode), wire.XSetDeviceValuators)
		}
		var valuatorInfo *wire.ValuatorClassInfo
		for _, class := range device.Classes {
			if vc, ok := class.(*wire.ValuatorClassInfo); ok {
				valuatorInfo = vc
				break
			}
		}
		if valuatorInfo == nil {
			return wire.NewError(wire.MatchErrorCode, seq, 0, byte(wire.XInputOpcode), wire.XSetDeviceValuators)
		}
		if int(req.FirstValuator)+int(req.NumValuators) > len(valuatorInfo.Axes) {
			return wire.NewError(wire.ValueErrorCode, seq, 0, byte(wire.XInputOpcode), wire.XSetDeviceValuators)
		}
		for i := 0; i < int(req.NumValuators); i++ {
			valuatorInfo.Axes[int(req.FirstValuator)+i].Value = req.Valuators[i]
		}
		return &wire.SetDeviceValuatorsReply{Sequence: seq, Status: wire.GrabSuccess}

	case wire.XGetDeviceControl:
		req, err := wire.ParseGetDeviceControlRequest(client.byteOrder, body, seq)
		if err != nil {
			return err.(messageEncoder)
		}
		device, ok := client.openDevices[req.DeviceID]
		if !ok {
			return wire.NewError(wire.ValueErrorCode, seq, uint32(req.DeviceID), byte(wire.XInputOpcode), wire.XGetDeviceControl)
		}
		var valuatorInfo *wire.ValuatorClassInfo
		for _, class := range device.Classes {
			if vc, ok := class.(*wire.ValuatorClassInfo); ok {
				valuatorInfo = vc
				break
			}
		}
		if valuatorInfo == nil {
			return wire.NewError(wire.MatchErrorCode, seq, 0, byte(wire.XInputOpcode), wire.XGetDeviceControl)
		}
		resolutions := make([]uint32, len(valuatorInfo.Axes))
		minResolutions := make([]uint32, len(valuatorInfo.Axes))
		maxResolutions := make([]uint32, len(valuatorInfo.Axes))
		for i, axis := range valuatorInfo.Axes {
			resolutions[i] = axis.Resolution
			minResolutions[i] = 0
			maxResolutions[i] = 1000
		}
		return &wire.GetDeviceControlReply{
			Sequence: seq,
			Control: &wire.DeviceResolutionState{
				NumValuators:   byte(len(valuatorInfo.Axes)),
				Resolutions:    resolutions,
				MinResolutions: minResolutions,
				MaxResolutions: maxResolutions,
			},
		}

	case wire.XChangeDeviceControl:
		req, err := wire.ParseChangeDeviceControlRequest(client.byteOrder, body, seq)
		if err != nil {
			return err.(messageEncoder)
		}
		device, ok := client.openDevices[req.DeviceID]
		if !ok {
			return wire.NewError(wire.ValueErrorCode, seq, uint32(req.DeviceID), byte(wire.XInputOpcode), wire.XChangeDeviceControl)
		}
		var valuatorInfo *wire.ValuatorClassInfo
		for _, class := range device.Classes {
			if vc, ok := class.(*wire.ValuatorClassInfo); ok {
				valuatorInfo = vc
				break
			}
		}
		if valuatorInfo == nil {
			return wire.NewError(wire.MatchErrorCode, seq, 0, byte(wire.XInputOpcode), wire.XChangeDeviceControl)
		}
		resolutionControl, ok := req.Control.(*wire.DeviceResolutionControl)
		if !ok {
			return wire.NewError(wire.ValueErrorCode, seq, 0, byte(wire.XInputOpcode), wire.XChangeDeviceControl)
		}
		if int(resolutionControl.FirstValuator)+int(resolutionControl.NumValuators) > len(valuatorInfo.Axes) {
			return wire.NewError(wire.ValueErrorCode, seq, 0, byte(wire.XInputOpcode), wire.XChangeDeviceControl)
		}
		for i := 0; i < int(resolutionControl.NumValuators); i++ {
			valuatorInfo.Axes[int(resolutionControl.FirstValuator)+i].Resolution = resolutionControl.Resolutions[i]
		}
		return &wire.ChangeDeviceControlReply{Sequence: seq, Status: wire.GrabSuccess}

	case wire.XGetSelectedExtensionEvents:
		req, err := wire.ParseGetSelectedExtensionEventsRequest(client.byteOrder, body, seq)
		if err != nil {
			return err.(messageEncoder)
		}
		var thisClientClasses, allClientsClasses []uint32
		for _, dev := range client.openDevices {
			if mask, ok := dev.EventMasks[req.Window]; ok {
				class := (mask << 8) | uint32(dev.Header.DeviceID)
				thisClientClasses = append(thisClientClasses, class)
			}
		}
		for _, c := range s.clients {
			for _, dev := range c.openDevices {
				if mask, ok := dev.EventMasks[req.Window]; ok {
					class := (mask << 8) | uint32(dev.Header.DeviceID)
					allClientsClasses = append(allClientsClasses, class)
				}
			}
		}
		return &wire.GetSelectedExtensionEventsReply{
			Sequence:          seq,
			ThisClientClasses: thisClientClasses,
			AllClientsClasses: allClientsClasses,
		}

	case wire.XChangeDeviceDontPropagateList:
		req, err := wire.ParseChangeDeviceDontPropagateListRequest(client.byteOrder, body, seq)
		if err != nil {
			return err.(messageEncoder)
		}
		win, ok := s.windows[client.xID(req.Window)]
		if !ok {
			return wire.NewError(wire.WindowErrorCode, seq, req.Window, byte(wire.XInputOpcode), wire.XChangeDeviceDontPropagateList)
		}
		if win.dontPropagateDeviceEvents == nil {
			win.dontPropagateDeviceEvents = make(map[uint32]bool)
		}
		for _, class := range req.Classes {
			if req.Mode == 0 { // AddToList
				win.dontPropagateDeviceEvents[class] = true
			} else { // DeleteFromList
				delete(win.dontPropagateDeviceEvents, class)
			}
		}
		return nil

	case wire.XAllowDeviceEvents:
		req, err := wire.ParseAllowDeviceEventsRequest(client.byteOrder, body, seq)
		if err != nil {
			return err.(messageEncoder)
		}
		s.frontend.AllowEvents(client.id, req.Mode, req.Time)
		return nil

	case wire.XChangeKeyboardDevice:
		return wire.NewError(wire.DeviceErrorCode, seq, 0, byte(wire.XInputOpcode), wire.XChangeKeyboardDevice)

	case wire.XChangePointerDevice:
		return wire.NewError(wire.DeviceErrorCode, seq, 0, byte(wire.XInputOpcode), wire.XChangePointerDevice)

	case wire.XGetDeviceDontPropagateList:
		req, err := wire.ParseGetDeviceDontPropagateListRequest(client.byteOrder, body, seq)
		if err != nil {
			return err.(messageEncoder)
		}
		win, ok := s.windows[client.xID(req.Window)]
		if !ok {
			return wire.NewError(wire.WindowErrorCode, seq, req.Window, byte(wire.XInputOpcode), wire.XGetDeviceDontPropagateList)
		}
		classes := make([]uint32, 0, len(win.dontPropagateDeviceEvents))
		for class := range win.dontPropagateDeviceEvents {
			classes = append(classes, class)
		}
		return &wire.GetDeviceDontPropagateListReply{
			Sequence: seq,
			Classes:  classes,
		}

	case wire.XSendExtensionEvent:
		dest := client.byteOrder.Uint32(body[0:4])
		numClasses := client.byteOrder.Uint16(body[8:10])
		numEvents := body[10]

		if len(body) < 12+int(numEvents)*32+int(numClasses)*4 {
			return wire.NewError(wire.LengthErrorCode, seq, 0, byte(wire.XInputOpcode), wire.XSendExtensionEvent)
		}

		eventBytes := body[12 : 12+int(numEvents)*32]
		classesBytes := body[12+int(numEvents)*32:]

		// Assuming a 1-to-1 mapping between events and classes
		for i := 0; i < int(numEvents); i++ {
			eventData := eventBytes[i*32 : (i+1)*32]
			class := client.byteOrder.Uint32(classesBytes[i*4 : (i+1)*4])

			eventMask := class >> 8
			deviceID := byte(class & 0xFF)

			for _, c := range s.clients {
				if dev, ok := c.openDevices[deviceID]; ok {
					if mask, ok := dev.EventMasks[dest]; ok {
						if (mask & eventMask) != 0 {
							// The client has selected for this event.
							// Send the raw event, but update the sequence number.
							c.byteOrder.PutUint16(eventData[2:4], c.sequence-1)
							rawEvent := &wire.X11RawEvent{Data: eventData}
							c.send(rawEvent)
						}
					}
				}
			}
		}
		return nil

	case wire.XCloseDevice:
		req, err := wire.ParseCloseDeviceRequest(client.byteOrder, body, seq)
		if err != nil {
			return err.(messageEncoder)
		}
		delete(client.openDevices, req.DeviceID)
		return &wire.CloseDeviceReply{Sequence: seq}

	case wire.XSelectExtensionEvent:
		windowID := client.byteOrder.Uint32(body[0:4])
		numClasses := client.byteOrder.Uint16(body[4:6])
		for i := 0; i < int(numClasses); i++ {
			class := client.byteOrder.Uint32(body[8+i*4 : 12+i*4])
			deviceID := byte(class & 0xFF)
			mask := class >> 8
			if dev, ok := client.openDevices[deviceID]; ok {
				if dev.EventMasks == nil {
					dev.EventMasks = make(map[uint32]uint32)
				}
				dev.EventMasks[windowID] = mask
			}
		}
		return nil

	case wire.XGrabDevice:
		req, err := wire.ParseGrabDeviceRequest(client.byteOrder, body, seq)
		if err != nil {
			return err.(messageEncoder)
		}
		if _, ok := s.deviceGrabs[req.DeviceID]; ok {
			return &wire.GrabDeviceReply{Sequence: seq, Status: wire.AlreadyGrabbed}
		}
		grab := &deviceGrab{
			window:      client.xID(req.GrabWindow),
			ownerEvents: req.OwnerEvents,
			eventMask:   req.Classes,
			time:        req.Time,
		}
		s.deviceGrabs[req.DeviceID] = grab
		return &wire.GrabDeviceReply{Sequence: seq, Status: wire.GrabSuccess}
	case wire.XUngrabDevice:
		req, err := wire.ParseUngrabDeviceRequest(client.byteOrder, body, seq)
		if err != nil {
			return err.(messageEncoder)
		}
		if grab, ok := s.deviceGrabs[req.DeviceID]; ok {
			if grab.window.client == client.id {
				delete(s.deviceGrabs, req.DeviceID)
			}
		}
		return nil

	default:
		// TODO: Implement other XInput requests
		return nil
	}
}
