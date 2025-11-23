//go:build x11

package x11

import (
	"github.com/c2FmZQ/sshterm/internal/x11/wire"
)

func (s *x11Server) handleXInputRequest(client *x11Client, req wire.Request, seq uint16) (reply messageEncoder) {
	switch p := req.(type) {
	case *wire.GetExtensionVersionRequest:
		return &wire.GetExtensionVersionReply{
			Sequence:     seq,
			MajorVersion: 2,
			MinorVersion: 2,
		}

	case *wire.ListInputDevicesRequest:
		return &wire.ListInputDevicesReply{
			Sequence: seq,
			Devices:  []*wire.DeviceInfo{virtualPointer, virtualKeyboard},
		}

	case *wire.OpenDeviceRequest:
		var selectedDevice *wire.DeviceInfo
		if p.DeviceID == virtualPointer.Header.DeviceID {
			selectedDevice = virtualPointer
		} else if p.DeviceID == virtualKeyboard.Header.DeviceID {
			selectedDevice = virtualKeyboard
		} else {
			return wire.NewError(wire.ValueErrorCode, seq, uint32(p.DeviceID), wire.Opcodes{Major: wire.XInputOpcode, Minor: wire.XOpenDevice})
		}

		// Create a new deviceInfo instance for the client, so event masks are not shared.
		newClasses := make([]wire.InputClassInfo, len(selectedDevice.Classes))
		copy(newClasses, selectedDevice.Classes)
		newDeviceInfo := &wire.DeviceInfo{
			Header:     selectedDevice.Header,
			Classes:    newClasses,
			EventMasks: make(map[uint32]uint32),
		}
		client.openDevices[p.DeviceID] = newDeviceInfo
		return &wire.OpenDeviceReply{Sequence: seq, Classes: newDeviceInfo.Classes}

	case *wire.SetDeviceModeRequest:
		device, ok := client.openDevices[p.DeviceID]
		if !ok {
			return wire.NewError(wire.ValueErrorCode, seq, uint32(p.DeviceID), wire.Opcodes{Major: wire.XInputOpcode, Minor: wire.XSetDeviceMode})
		}
		var valuatorInfo *wire.ValuatorClassInfo
		for _, class := range device.Classes {
			if vc, ok := class.(*wire.ValuatorClassInfo); ok {
				valuatorInfo = vc
				break
			}
		}
		if valuatorInfo == nil {
			return wire.NewError(wire.MatchErrorCode, seq, 0, wire.Opcodes{Major: wire.XInputOpcode, Minor: wire.XSetDeviceMode})
		}
		valuatorInfo.Mode = p.Mode
		return &wire.SetDeviceModeReply{Sequence: seq, Status: wire.GrabSuccess}

	case *wire.SetDeviceValuatorsRequest:
		device, ok := client.openDevices[p.DeviceID]
		if !ok {
			return wire.NewError(wire.ValueErrorCode, seq, uint32(p.DeviceID), wire.Opcodes{Major: wire.XInputOpcode, Minor: wire.XSetDeviceValuators})
		}
		var valuatorInfo *wire.ValuatorClassInfo
		for _, class := range device.Classes {
			if vc, ok := class.(*wire.ValuatorClassInfo); ok {
				valuatorInfo = vc
				break
			}
		}
		if valuatorInfo == nil {
			return wire.NewError(wire.MatchErrorCode, seq, 0, wire.Opcodes{Major: wire.XInputOpcode, Minor: wire.XSetDeviceValuators})
		}
		if int(p.FirstValuator)+int(p.NumValuators) > len(valuatorInfo.Axes) {
			return wire.NewError(wire.ValueErrorCode, seq, 0, wire.Opcodes{Major: wire.XInputOpcode, Minor: wire.XSetDeviceValuators})
		}
		for i := 0; i < int(p.NumValuators); i++ {
			valuatorInfo.Axes[int(p.FirstValuator)+i].Value = p.Valuators[i]
		}
		return &wire.SetDeviceValuatorsReply{Sequence: seq, Status: wire.GrabSuccess}

	case *wire.GetDeviceControlRequest:
		device, ok := client.openDevices[p.DeviceID]
		if !ok {
			return wire.NewError(wire.ValueErrorCode, seq, uint32(p.DeviceID), wire.Opcodes{Major: wire.XInputOpcode, Minor: wire.XGetDeviceControl})
		}
		var valuatorInfo *wire.ValuatorClassInfo
		for _, class := range device.Classes {
			if vc, ok := class.(*wire.ValuatorClassInfo); ok {
				valuatorInfo = vc
				break
			}
		}
		if valuatorInfo == nil {
			return wire.NewError(wire.MatchErrorCode, seq, 0, wire.Opcodes{Major: wire.XInputOpcode, Minor: wire.XGetDeviceControl})
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

	case *wire.ChangeDeviceControlRequest:
		device, ok := client.openDevices[p.DeviceID]
		if !ok {
			return wire.NewError(wire.ValueErrorCode, seq, uint32(p.DeviceID), wire.Opcodes{Major: wire.XInputOpcode, Minor: wire.XChangeDeviceControl})
		}
		var valuatorInfo *wire.ValuatorClassInfo
		for _, class := range device.Classes {
			if vc, ok := class.(*wire.ValuatorClassInfo); ok {
				valuatorInfo = vc
				break
			}
		}
		if valuatorInfo == nil {
			return wire.NewError(wire.MatchErrorCode, seq, 0, wire.Opcodes{Major: wire.XInputOpcode, Minor: wire.XChangeDeviceControl})
		}
		resolutionControl, ok := p.Control.(*wire.DeviceResolutionControl)
		if !ok {
			return wire.NewError(wire.ValueErrorCode, seq, 0, wire.Opcodes{Major: wire.XInputOpcode, Minor: wire.XChangeDeviceControl})
		}
		if int(resolutionControl.FirstValuator)+int(resolutionControl.NumValuators) > len(valuatorInfo.Axes) {
			return wire.NewError(wire.ValueErrorCode, seq, 0, wire.Opcodes{Major: wire.XInputOpcode, Minor: wire.XChangeDeviceControl})
		}
		for i := 0; i < int(resolutionControl.NumValuators); i++ {
			valuatorInfo.Axes[int(resolutionControl.FirstValuator)+i].Resolution = resolutionControl.Resolutions[i]
		}
		return &wire.ChangeDeviceControlReply{Sequence: seq, Status: wire.GrabSuccess}

	case *wire.GetSelectedExtensionEventsRequest:
		var thisClientClasses, allClientsClasses []uint32
		for _, dev := range client.openDevices {
			if mask, ok := dev.EventMasks[p.Window]; ok {
				class := (mask << 8) | uint32(dev.Header.DeviceID)
				thisClientClasses = append(thisClientClasses, class)
			}
		}
		for _, c := range s.clients {
			for _, dev := range c.openDevices {
				if mask, ok := dev.EventMasks[p.Window]; ok {
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

	case *wire.ChangeDeviceDontPropagateListRequest:
		win, ok := s.windows[client.xID(p.Window)]
		if !ok {
			return wire.NewError(wire.WindowErrorCode, seq, p.Window, wire.Opcodes{Major: wire.XInputOpcode, Minor: wire.XChangeDeviceDontPropagateList})
		}
		if win.dontPropagateDeviceEvents == nil {
			win.dontPropagateDeviceEvents = make(map[uint32]bool)
		}
		for _, class := range p.Classes {
			if p.Mode == 0 { // AddToList
				win.dontPropagateDeviceEvents[class] = true
			} else { // DeleteFromList
				delete(win.dontPropagateDeviceEvents, class)
			}
		}
		return nil

	case *wire.AllowDeviceEventsRequest:
		s.frontend.AllowEvents(client.id, p.Mode, p.Time)
		return nil

	case *wire.ChangeKeyboardDeviceRequest:
		return wire.NewError(wire.DeviceErrorCode, seq, 0, wire.Opcodes{Major: wire.XInputOpcode, Minor: wire.XChangeKeyboardDevice})

	case *wire.ChangePointerDeviceRequest:
		return wire.NewError(wire.DeviceErrorCode, seq, 0, wire.Opcodes{Major: wire.XInputOpcode, Minor: wire.XChangePointerDevice})

	case *wire.GetDeviceDontPropagateListRequest:
		win, ok := s.windows[client.xID(p.Window)]
		if !ok {
			return wire.NewError(wire.WindowErrorCode, seq, p.Window, wire.Opcodes{Major: wire.XInputOpcode, Minor: wire.XGetDeviceDontPropagateList})
		}
		classes := make([]uint32, 0, len(win.dontPropagateDeviceEvents))
		for class := range win.dontPropagateDeviceEvents {
			classes = append(classes, class)
		}
		return &wire.GetDeviceDontPropagateListReply{
			Sequence: seq,
			Classes:  classes,
		}

	case *wire.SendExtensionEventRequest:
		dest := p.Destination
		numEvents := p.NumEvents
		events := p.Events
		classes := p.Classes

		// Assuming a 1-to-1 mapping between events and classes
		for i := 0; i < int(numEvents); i++ {
			eventData := events[i*32 : (i+1)*32]
			class := classes[i] // classes array already holds uint32

			eventMask := class >> 8
			deviceID := byte(class & 0xFF)

			for _, c := range s.clients {
				if dev, ok := c.openDevices[deviceID]; ok {
					if mask, ok := dev.EventMasks[uint32(dest)]; ok { // Cast dest to uint32
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

	case *wire.CloseDeviceRequest:
		delete(client.openDevices, p.DeviceID)
		return &wire.CloseDeviceReply{Sequence: seq}

	case *wire.SelectExtensionEventRequest:
		windowID := uint32(p.Window) // p.Window is wire.Window, an alias for uint32
		// p.Classes is []uint32, so its length gives numClasses
		for _, class := range p.Classes {
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

	case *wire.GrabDeviceRequest:
		if _, ok := s.deviceGrabs[p.DeviceID]; ok {
			return &wire.GrabDeviceReply{Sequence: seq, Status: wire.AlreadyGrabbed}
		}
		grab := &deviceGrab{
			window:      client.xID(p.GrabWindow),
			ownerEvents: p.OwnerEvents,
			eventMask:   p.Classes,
			time:        p.Time,
		}
		s.deviceGrabs[p.DeviceID] = grab
		return &wire.GrabDeviceReply{Sequence: seq, Status: wire.GrabSuccess}

	case *wire.UngrabDeviceRequest:
		if grab, ok := s.deviceGrabs[p.DeviceID]; ok {
			if grab.window.client == client.id {
				delete(s.deviceGrabs, p.DeviceID)
			}
		}
		return nil

	case *wire.GrabDeviceKeyRequest:
		grabWindow := client.xID(uint32(p.GrabWindow))
		grab := &passiveDeviceGrab{
			clientID:  client.id,
			deviceID:  p.DeviceID,
			key:       wire.KeyCode(p.Key),
			modifiers: p.Modifiers,
			owner:     p.OwnerEvents,
			eventMask: p.Classes,
		}
		s.passiveDeviceGrabs[grabWindow] = append(s.passiveDeviceGrabs[grabWindow], grab)
		return nil

	case *wire.UngrabDeviceKeyRequest:
		grabWindow := client.xID(uint32(p.GrabWindow))
		if grabs, ok := s.passiveDeviceGrabs[grabWindow]; ok {
			newGrabs := make([]*passiveDeviceGrab, 0, len(grabs))
			for _, grab := range grabs {
				if !(grab.key == wire.KeyCode(p.Key) && (p.Modifiers == wire.AnyModifier || grab.modifiers == p.Modifiers)) {
					newGrabs = append(newGrabs, grab)
				}
			}
			s.passiveDeviceGrabs[grabWindow] = newGrabs
		}
		return nil

	case *wire.GrabDeviceButtonRequest:
		grabWindow := client.xID(uint32(p.GrabWindow))
		grab := &passiveDeviceGrab{
			clientID:  client.id,
			deviceID:  p.DeviceID,
			button:    p.Button,
			modifiers: p.Modifiers,
			owner:     p.OwnerEvents,
			eventMask: p.Classes,
		}
		s.passiveDeviceGrabs[grabWindow] = append(s.passiveDeviceGrabs[grabWindow], grab)
		return nil

	case *wire.UngrabDeviceButtonRequest:
		grabWindow := client.xID(uint32(p.GrabWindow))
		if grabs, ok := s.passiveDeviceGrabs[grabWindow]; ok {
			newGrabs := make([]*passiveDeviceGrab, 0, len(grabs))
			for _, grab := range grabs {
				if !(grab.button == p.Button && (p.Modifiers == wire.AnyModifier || grab.modifiers == p.Modifiers)) {
					newGrabs = append(newGrabs, grab)
				}
			}
			s.passiveDeviceGrabs[grabWindow] = newGrabs
		}
		return nil

	case *wire.GetDeviceFocusRequest:
		return &wire.GetDeviceFocusReply{
			Sequence: seq,
			Focus:    s.inputFocus.local,
			RevertTo: 1, // RevertToParent
		}

	case *wire.SetDeviceFocusRequest:
		s.inputFocus = client.xID(uint32(p.Focus))
		return nil

	case *wire.GetFeedbackControlRequest:
		feedbacks := s.frontend.GetFeedbackControl(p.DeviceID)
		return &wire.GetFeedbackControlReply{
			Sequence:  seq,
			Feedbacks: feedbacks,
			NumEvents: uint16(len(feedbacks)),
		}

	case *wire.ChangeFeedbackControlRequest:
		s.frontend.ChangeFeedbackControl(p.DeviceID, p.ControlID, p.Mask, p.Control)
		return nil

	case *wire.GetDeviceKeyMappingRequest:
		keysymsPerKeycode, keysyms := s.frontend.GetDeviceKeyMapping(p.DeviceID, p.FirstKey, p.Count)
		return &wire.GetDeviceKeyMappingReply{
			Sequence:          seq,
			KeysymsPerKeycode: keysymsPerKeycode,
			Keysyms:           keysyms,
		}

	case *wire.ChangeDeviceKeyMappingRequest:
		s.frontend.ChangeDeviceKeyMapping(p.DeviceID, p.FirstKey, p.KeysymsPerKeycode, p.KeycodeCount, p.Keysyms)
		return nil

	case *wire.GetDeviceModifierMappingRequest:
		numKeycodesPerMod, keycodes := s.frontend.GetDeviceModifierMapping(p.DeviceID)
		return &wire.GetDeviceModifierMappingReply{
			Sequence:          seq,
			NumKeycodesPerMod: numKeycodesPerMod,
			Keycodes:          keycodes,
		}

	case *wire.SetDeviceModifierMappingRequest:
		status := s.frontend.SetDeviceModifierMapping(p.DeviceID, p.Keycodes)
		return &wire.SetDeviceModifierMappingReply{
			Sequence: seq,
			Status:   status,
		}

	case *wire.GetDeviceButtonMappingRequest:
		buttonMap := s.frontend.GetDeviceButtonMapping(p.DeviceID)
		return &wire.GetDeviceButtonMappingReply{
			Sequence: seq,
			Map:      buttonMap,
		}

	case *wire.SetDeviceButtonMappingRequest:
		status := s.frontend.SetDeviceButtonMapping(p.DeviceID, p.Map)
		return &wire.SetDeviceButtonMappingReply{
			Sequence: seq,
			Status:   status,
		}

	case *wire.QueryDeviceStateRequest:
		classes := s.frontend.QueryDeviceState(p.DeviceID)
		return &wire.QueryDeviceStateReply{
			Sequence:  seq,
			Classes:   classes,
			NumEvents: uint16(len(classes)),
		}

	case *wire.DeviceBellRequest:
		s.frontend.DeviceBell(p.DeviceID, p.FeedbackID, p.FeedbackClass, int8(p.Percent))
		return nil

	case *wire.XIChangeHierarchyRequest:
		s.frontend.XIChangeHierarchy(p.Changes)
		return nil

	case *wire.XIQueryVersionRequest:
		return &wire.XIQueryVersionReply{
			Sequence:     seq,
			MajorVersion: 2,
			MinorVersion: 2,
		}

	case *wire.XISelectEventsRequest:
		// TODO: Implement XI 2.x event masks storage
		return nil

	default:
		return wire.NewError(wire.RequestErrorCode, seq, 0, wire.Opcodes{Major: wire.XInputOpcode, Minor: 0})
	}
}
