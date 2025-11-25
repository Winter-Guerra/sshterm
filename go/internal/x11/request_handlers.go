//go:build x11

package x11

import (
	"github.com/c2FmZQ/sshterm/internal/x11/wire"
)

//
// Handlers for X11 requests
//

func (s *x11Server) handleCreateWindow(client *x11Client, req wire.Request, seq uint16) messageEncoder {
	p := req.(*wire.CreateWindowRequest)
	xid := client.xID(uint32(p.Drawable))
	parentXID := client.xID(uint32(p.Parent))
	// Check if the window ID is already in use
	if _, exists := s.windows[xid]; exists {
		s.logger.Errorf("X11: CreateWindow: ID %d already in use", xid)
		return wire.NewGenericError(seq, uint32(p.Drawable), 0, wire.CreateWindow, wire.IDChoiceErrorCode)
	}

	newWindow := &window{
		xid:        xid,
		parent:     uint32(p.Parent),
		x:          p.X,
		y:          p.Y,
		width:      p.Width,
		height:     p.Height,
		depth:      p.Depth,
		children:   []uint32{},
		attributes: p.Values,
	}
	if p.Values.Colormap > 0 {
		newWindow.colormap = client.xID(uint32(p.Values.Colormap))
	} else {
		newWindow.colormap = xID{local: s.defaultColormap}
	}
	s.windows[xid] = newWindow

	// Add to parent's children list
	if parentWindow, ok := s.windows[parentXID]; ok {
		parentWindow.children = append(parentWindow.children, uint32(p.Drawable))
	}
	s.frontend.CreateWindow(xid, uint32(p.Parent), uint32(p.X), uint32(p.Y), uint32(p.Width), uint32(p.Height), uint32(p.Depth), p.ValueMask, p.Values)
	return nil
}

func (s *x11Server) handleChangeWindowAttributes(client *x11Client, req wire.Request, seq uint16) messageEncoder {
	p := req.(*wire.ChangeWindowAttributesRequest)
	xid := client.xID(uint32(p.Window))
	if err := s.checkWindow(xid, seq, wire.ChangeWindowAttributes, 0); err != nil {
		return err
	}
	if w, ok := s.windows[xid]; ok {
		if p.ValueMask&wire.CWBackPixmap != 0 {
			w.attributes.BackgroundPixmap = p.Values.BackgroundPixmap
		}
		if p.ValueMask&wire.CWBackPixel != 0 {
			w.attributes.BackgroundPixel = p.Values.BackgroundPixel
		}
		if p.ValueMask&wire.CWBorderPixmap != 0 {
			w.attributes.BorderPixmap = p.Values.BorderPixmap
		}
		if p.ValueMask&wire.CWBorderPixel != 0 {
			w.attributes.BorderPixel = p.Values.BorderPixel
		}
		if p.ValueMask&wire.CWBitGravity != 0 {
			w.attributes.BitGravity = p.Values.BitGravity
		}
		if p.ValueMask&wire.CWWinGravity != 0 {
			w.attributes.WinGravity = p.Values.WinGravity
		}
		if p.ValueMask&wire.CWBackingStore != 0 {
			w.attributes.BackingStore = p.Values.BackingStore
		}
		if p.ValueMask&wire.CWBackingPlanes != 0 {
			w.attributes.BackingPlanes = p.Values.BackingPlanes
		}
		if p.ValueMask&wire.CWBackingPixel != 0 {
			w.attributes.BackingPixel = p.Values.BackingPixel
		}
		if p.ValueMask&wire.CWOverrideRedirect != 0 {
			w.attributes.OverrideRedirect = p.Values.OverrideRedirect
		}
		if p.ValueMask&wire.CWSaveUnder != 0 {
			w.attributes.SaveUnder = p.Values.SaveUnder
		}
		if p.ValueMask&wire.CWEventMask != 0 {
			w.attributes.EventMask = p.Values.EventMask
		}
		if p.ValueMask&wire.CWDontPropagate != 0 {
			w.attributes.DontPropagateMask = p.Values.DontPropagateMask
		}
		if p.ValueMask&wire.CWColormap != 0 {
			w.attributes.Colormap = p.Values.Colormap
		}
		if p.ValueMask&wire.CWCursor != 0 {
			w.attributes.Cursor = p.Values.Cursor
			s.frontend.SetWindowCursor(xid, client.xID(uint32(p.Values.Cursor)))
		}
	}
	s.frontend.ChangeWindowAttributes(xid, p.ValueMask, p.Values)
	return nil
}

func (s *x11Server) handleGetWindowAttributes(client *x11Client, req wire.Request, seq uint16) messageEncoder {
	p := req.(*wire.GetWindowAttributesRequest)
	xid := client.xID(uint32(p.Window))
	if err := s.checkWindow(xid, seq, wire.GetWindowAttributes, 0); err != nil {
		return err
	}
	attrs, _ := s.GetWindowAttributes(xid)
	return &wire.GetWindowAttributesReply{
		Sequence:           seq,
		BackingStore:       byte(attrs.BackingStore),
		VisualID:           s.visualID,
		Class:              uint16(attrs.Class),
		BitGravity:         byte(attrs.BitGravity),
		WinGravity:         byte(attrs.WinGravity),
		BackingPlanes:      attrs.BackingPlanes,
		BackingPixel:       attrs.BackingPixel,
		SaveUnder:          wire.BoolToByte(attrs.SaveUnder),
		MapIsInstalled:     wire.BoolToByte(attrs.MapIsInstalled),
		MapState:           byte(attrs.MapState),
		OverrideRedirect:   wire.BoolToByte(attrs.OverrideRedirect),
		Colormap:           uint32(attrs.Colormap),
		AllEventMasks:      attrs.EventMask,
		YourEventMask:      attrs.EventMask,
		DoNotPropagateMask: uint16(attrs.DontPropagateMask),
	}
}

func (s *x11Server) handleDestroyWindow(client *x11Client, req wire.Request, seq uint16) messageEncoder {
	p := req.(*wire.DestroyWindowRequest)
	xid := client.xID(uint32(p.Window))
	if xid.local == s.rootWindowID() {
		return nil
	}
	if err := s.checkWindow(xid, seq, wire.DestroyWindow, 0); err != nil {
		return err
	}
	delete(s.windows, xid)
	s.frontend.DestroyWindow(xid)
	return nil
}

func (s *x11Server) handleDestroySubwindows(client *x11Client, req wire.Request, seq uint16) messageEncoder {
	p := req.(*wire.DestroySubwindowsRequest)
	xid := client.xID(uint32(p.Window))
	if err := s.checkWindow(xid, seq, wire.DestroySubwindows, 0); err != nil {
		return err
	}
	if parent, ok := s.windows[xid]; ok {
		var destroy func(uint32)
		destroy = func(windowID uint32) {
			childXID := client.xID(windowID)
			if w, ok := s.windows[childXID]; ok {
				for _, child := range w.children {
					destroy(child)
				}
				delete(s.windows, childXID)
			}
		}
		for _, child := range parent.children {
			destroy(child)
		}
		parent.children = []uint32{}
	}
	s.frontend.DestroySubwindows(xid)
	return nil
}

func (s *x11Server) handleChangeSaveSet(client *x11Client, req wire.Request, seq uint16) messageEncoder {
	p := req.(*wire.ChangeSaveSetRequest)
	if err := s.checkWindow(client.xID(uint32(p.Window)), seq, wire.ChangeSaveSet, 0); err != nil {
		return err
	}
	if p.Mode == 0 { // Insert
		client.saveSet[uint32(p.Window)] = true
	} else { // Delete
		delete(client.saveSet, uint32(p.Window))
	}
	return nil
}

func (s *x11Server) handleReparentWindow(client *x11Client, req wire.Request, seq uint16) messageEncoder {
	p := req.(*wire.ReparentWindowRequest)
	windowXID := client.xID(uint32(p.Window))
	parentXID := client.xID(uint32(p.Parent))
	if err := s.checkWindow(windowXID, seq, wire.ReparentWindow, 0); err != nil {
		return err
	}
	if err := s.checkWindow(parentXID, seq, wire.ReparentWindow, 0); err != nil {
		return err
	}
	window := s.windows[windowXID]
	if window == nil {
		return wire.NewGenericError(seq, uint32(p.Window), 0, wire.ReparentWindow, wire.MatchErrorCode)
	}

	oldParent, ok := s.windows[client.xID(window.parent)]
	if !ok && window.parent != s.rootWindowID() {
		return wire.NewGenericError(seq, window.parent, 0, wire.ReparentWindow, wire.WindowErrorCode)
	}
	newParent := s.windows[parentXID]

	// Remove from old parent's children
	if ok {
		for i, childID := range oldParent.children {
			if childID == window.xid.local {
				oldParent.children = append(oldParent.children[:i], oldParent.children[i+1:]...)
				break
			}
		}
	}

	// Add to new parent's children
	if newParent != nil {
		newParent.children = append(newParent.children, window.xid.local)
	}

	// Update window's state
	window.parent = uint32(p.Parent)
	window.x = p.X
	window.y = p.Y

	s.frontend.ReparentWindow(windowXID, parentXID, p.X, p.Y)
	return nil
}

func (s *x11Server) handleMapWindow(client *x11Client, req wire.Request, seq uint16) messageEncoder {
	p := req.(*wire.MapWindowRequest)
	xid := client.xID(uint32(p.Window))
	if err := s.checkWindow(xid, seq, wire.MapWindow, 0); err != nil {
		return err
	}
	if w, ok := s.windows[xid]; ok {
		w.mapped = true
		s.frontend.MapWindow(xid)
		s.sendExposeEvent(xid, 0, 0, w.width, w.height)
	}
	return nil
}

func (s *x11Server) handleMapSubwindows(client *x11Client, req wire.Request, seq uint16) messageEncoder {
	p := req.(*wire.MapSubwindowsRequest)
	xid := client.xID(uint32(p.Window))
	if err := s.checkWindow(xid, seq, wire.MapSubwindows, 0); err != nil {
		return err
	}
	if parentWindow, ok := s.windows[xid]; ok {
		for _, childID := range parentWindow.children {
			childXID := xID{client: xid.client, local: childID}
			if childWindow, ok := s.windows[childXID]; ok {
				childWindow.mapped = true
				s.frontend.MapWindow(childXID)
				s.sendExposeEvent(childXID, 0, 0, childWindow.width, childWindow.height)
			}
		}
	}
	return nil
}

func (s *x11Server) handleUnmapWindow(client *x11Client, req wire.Request, seq uint16) messageEncoder {
	p := req.(*wire.UnmapWindowRequest)
	xid := client.xID(uint32(p.Window))
	if err := s.checkWindow(xid, seq, wire.UnmapWindow, 0); err != nil {
		return err
	}
	if w, ok := s.windows[xid]; ok {
		w.mapped = false
	}
	s.frontend.UnmapWindow(xid)
	return nil
}

func (s *x11Server) handleUnmapSubwindows(client *x11Client, req wire.Request, seq uint16) messageEncoder {
	p := req.(*wire.UnmapSubwindowsRequest)
	xid := client.xID(uint32(p.Window))
	if err := s.checkWindow(xid, seq, wire.UnmapSubwindows, 0); err != nil {
		return err
	}
	if parentWindow, ok := s.windows[xid]; ok {
		for _, childID := range parentWindow.children {
			childXID := xID{client: xid.client, local: childID}
			if childWindow, ok := s.windows[childXID]; ok {
				childWindow.mapped = false
				s.frontend.UnmapWindow(childXID)
			}
		}
	}
	return nil
}

func (s *x11Server) handleConfigureWindow(client *x11Client, req wire.Request, seq uint16) messageEncoder {
	p := req.(*wire.ConfigureWindowRequest)
	xid := client.xID(uint32(p.Window))
	if err := s.checkWindow(xid, seq, wire.ConfigureWindow, 0); err != nil {
		return err
	}
	if xid.local == s.rootWindowID() {
		return wire.NewGenericError(seq, uint32(p.Window), 0, wire.ConfigureWindow, wire.MatchErrorCode)
	}
	s.frontend.ConfigureWindow(xid, p.ValueMask, p.Values)
	return nil
}

func (s *x11Server) handleCirculateWindow(client *x11Client, req wire.Request, seq uint16) messageEncoder {
	p := req.(*wire.CirculateWindowRequest)
	xid := client.xID(uint32(p.Window))
	if err := s.checkWindow(xid, seq, wire.CirculateWindow, 0); err != nil {
		return err
	}
	window := s.windows[xid]
	if window == nil {
		return wire.NewGenericError(seq, uint32(p.Window), 0, wire.CirculateWindow, wire.MatchErrorCode)
	}
	parent, ok := s.windows[client.xID(window.parent)]
	if ok {
		// Find index of window in parent's children
		idx := -1
		for i, childID := range parent.children {
			if childID == xid.local {
				idx = i
				break
			}
		}

		if idx != -1 {
			// Remove window from children slice
			children := append(parent.children[:idx], parent.children[idx+1:]...)

			if p.Direction == 0 { // RaiseLowest
				// Add to end of slice
				parent.children = append(children, xid.local)
			} else { // LowerHighest
				// Add to beginning of slice
				parent.children = append([]uint32{xid.local}, children...)
			}
		}
	} else if window.parent != s.rootWindowID() {
		return wire.NewGenericError(seq, window.parent, 0, wire.CirculateWindow, wire.WindowErrorCode)
	}

	s.frontend.CirculateWindow(xid, p.Direction)
	return nil
}

func (s *x11Server) handleGetGeometry(client *x11Client, req wire.Request, seq uint16) messageEncoder {
	p := req.(*wire.GetGeometryRequest)
	xid := client.xID(uint32(p.Drawable))
	if err := s.checkDrawable(xid, seq, wire.GetGeometry, 0); err != nil {
		return err
	}
	if xid.local == s.rootWindowID() {
		return &wire.GetGeometryReply{
			Sequence:    seq,
			Depth:       24, // TODO: Get this from rootVisual or screen info
			Root:        s.rootWindowID(),
			X:           0,
			Y:           0,
			Width:       s.rootWindowWidth,
			Height:      s.rootWindowHeight,
			BorderWidth: 0,
		}
	}
	if w, ok := s.windows[xid]; ok {
		return &wire.GetGeometryReply{
			Sequence:    seq,
			Depth:       w.depth,
			Root:        s.rootWindowID(),
			X:           w.x,
			Y:           w.y,
			Width:       w.width,
			Height:      w.height,
			BorderWidth: 0, // Border width is not stored in window struct, assuming 0 for now
		}
	}
	// Must be a pixmap
	return &wire.GetGeometryReply{
		Sequence:    seq,
		Depth:       24, // Assumption
		Root:        s.rootWindowID(),
		X:           0,
		Y:           0,
		Width:       1, // Dummy
		Height:      1, // Dummy
		BorderWidth: 0,
	}
}

func (s *x11Server) handleQueryTree(client *x11Client, req wire.Request, seq uint16) messageEncoder {
	p := req.(*wire.QueryTreeRequest)
	xid := client.xID(uint32(p.Window))
	if err := s.checkWindow(xid, seq, wire.QueryTree, 0); err != nil {
		return err
	}
	window := s.windows[xid]
	if window == nil {
		var children []uint32
		for _, w := range s.windows {
			if w.parent == s.rootWindowID() {
				children = append(children, w.xid.local)
			}
		}
		return &wire.QueryTreeReply{
			Sequence:    seq,
			Root:        s.rootWindowID(),
			Parent:      s.rootWindowID(),
			NumChildren: uint16(len(children)),
			Children:    children,
		}
	}
	return &wire.QueryTreeReply{
		Sequence:    seq,
		Root:        s.rootWindowID(),
		Parent:      window.parent,
		NumChildren: uint16(len(window.children)),
		Children:    window.children,
	}
}

func (s *x11Server) handleInternAtom(client *x11Client, req wire.Request, seq uint16) messageEncoder {
	p := req.(*wire.InternAtomRequest)
	atomID := s.GetAtom(p.Name)

	return &wire.InternAtomReply{
		Sequence: seq,
		Atom:     atomID,
	}
}

func (s *x11Server) handleGetAtomName(client *x11Client, req wire.Request, seq uint16) messageEncoder {
	p := req.(*wire.GetAtomNameRequest)
	name := s.GetAtomName(uint32(p.Atom))
	return &wire.GetAtomNameReply{
		Sequence:   seq,
		NameLength: uint16(len(name)),
		Name:       name,
	}
}

func (s *x11Server) handleChangeProperty(client *x11Client, req wire.Request, seq uint16) messageEncoder {
	p := req.(*wire.ChangePropertyRequest)
	xid := client.xID(uint32(p.Window))
	if err := s.checkWindow(xid, seq, wire.ChangeProperty, 0); err != nil {
		return err
	}
	s.ChangeProperty(xid, uint32(p.Property), uint32(p.Type), byte(p.Format), p.Data)
	return nil
}

func (s *x11Server) handleDeleteProperty(client *x11Client, req wire.Request, seq uint16) messageEncoder {
	p := req.(*wire.DeletePropertyRequest)
	xid := client.xID(uint32(p.Window))
	if err := s.checkWindow(xid, seq, wire.DeleteProperty, 0); err != nil {
		return err
	}
	s.DeleteProperty(xid, uint32(p.Property))
	return nil
}

func (s *x11Server) handleGetProperty(client *x11Client, req wire.Request, seq uint16) messageEncoder {
	p := req.(*wire.GetPropertyRequest)
	xid := client.xID(uint32(p.Window))
	if err := s.checkWindow(xid, seq, wire.GetProperty, 0); err != nil {
		return err
	}
	prop := s.GetProperty(xid, uint32(p.Property))

	if prop == nil {
		return &wire.GetPropertyReply{
			Sequence: seq,
			Format:   0,
		}
	}

	// Calculate slice
	byteOffset := p.Offset * 4
	byteLength := p.Length * 4
	totalLen := uint32(len(prop.data))

	if byteOffset >= totalLen {
		return &wire.GetPropertyReply{
			Sequence:              seq,
			Format:                prop.format,
			PropertyType:          prop.typeAtom,
			BytesAfter:            0,
			ValueLenInFormatUnits: 0,
			Value:                 nil,
		}
	}

	end := byteOffset + byteLength
	if end > totalLen {
		end = totalLen
	}
	dataToSend := prop.data[byteOffset:end]
	bytesAfter := totalLen - end

	var valueLenInFormatUnits uint32
	if prop.format == 8 {
		valueLenInFormatUnits = uint32(len(dataToSend))
	} else if prop.format == 16 {
		valueLenInFormatUnits = uint32(len(dataToSend) / 2)
	} else if prop.format == 32 {
		valueLenInFormatUnits = uint32(len(dataToSend) / 4)
	}

	if p.Delete && bytesAfter == 0 && (p.Type == 0 || prop.typeAtom == uint32(p.Type)) {
		s.DeleteProperty(xid, uint32(p.Property))
	}

	if p.Type != 0 && prop.typeAtom != uint32(p.Type) {
		return &wire.GetPropertyReply{
			Sequence:              seq,
			Format:                prop.format,
			PropertyType:          prop.typeAtom,
			BytesAfter:            totalLen, // Full length
			ValueLenInFormatUnits: 0,
			Value:                 nil,
		}
	}

	return &wire.GetPropertyReply{
		Sequence:              seq,
		Format:                prop.format,
		PropertyType:          prop.typeAtom,
		BytesAfter:            bytesAfter,
		ValueLenInFormatUnits: valueLenInFormatUnits,
		Value:                 dataToSend,
	}
}

func (s *x11Server) handleListProperties(client *x11Client, req wire.Request, seq uint16) messageEncoder {
	p := req.(*wire.ListPropertiesRequest)
	xid := client.xID(uint32(p.Window))
	if err := s.checkWindow(xid, seq, wire.ListProperties, 0); err != nil {
		return err
	}
	atoms := s.ListProperties(xid)
	return &wire.ListPropertiesReply{
		Sequence:      seq,
		NumProperties: uint16(len(atoms)),
		Atoms:         atoms,
	}
}

func (s *x11Server) handleSetSelectionOwner(client *x11Client, req wire.Request, seq uint16) messageEncoder {
	p := req.(*wire.SetSelectionOwnerRequest)
	selectionAtom := uint32(p.Selection)
	ownerWindow := uint32(p.Owner)
	time := uint32(p.Time)

	if time == 0 { // CurrentTime
		time = s.serverTime()
	}

	// "If the timestamp is not CurrentTime and is less than the timestamp of the last successful SetSelectionOwner request for the selection, the request is ignored."
	currentOwner, ok := s.selections[selectionAtom]
	if ok && time < currentOwner.time && p.Time != 0 {
		return nil
	}

	if ownerWindow == 0 { // None
		delete(s.selections, selectionAtom)
	} else {
		if err := s.checkWindow(client.xID(ownerWindow), seq, wire.SetSelectionOwner, 0); err != nil {
			return err
		}
		s.selections[selectionAtom] = &selectionOwner{
			window: client.xID(ownerWindow),
			time:   time,
		}
	}

	if ok && currentOwner.window.local != 0 && (currentOwner.window != client.xID(ownerWindow)) {
		// Send SelectionClear to old owner
		if oldClient, ok := s.clients[currentOwner.window.client]; ok {
			event := &wire.SelectionClearEvent{
				Sequence:  oldClient.sequence - 1, // Approximate
				Time:      time,
				Owner:     currentOwner.window.local,
				Selection: selectionAtom,
			}
			s.sendEvent(oldClient, event)
		}
	}
	return nil
}

func (s *x11Server) handleGetSelectionOwner(client *x11Client, req wire.Request, seq uint16) messageEncoder {
	p := req.(*wire.GetSelectionOwnerRequest)
	selectionAtom := uint32(p.Selection)
	var owner uint32
	if o, ok := s.selections[selectionAtom]; ok {
		owner = o.window.local
	}
	return &wire.GetSelectionOwnerReply{
		Sequence: seq,
		Owner:    owner,
	}
}

func (s *x11Server) handleConvertSelection(client *x11Client, req wire.Request, seq uint16) messageEncoder {
	p := req.(*wire.ConvertSelectionRequest)
	selectionAtom := uint32(p.Selection)
	targetAtom := uint32(p.Target)
	propertyAtom := uint32(p.Property)
	requestor := client.xID(uint32(p.Requestor))
	time := uint32(p.Time)
	if time == 0 {
		time = s.serverTime()
	}

	if err := s.checkWindow(requestor, seq, wire.ConvertSelection, 0); err != nil {
		return err
	}

	owner, ok := s.selections[selectionAtom]

	// Special handling for CLIPBOARD if no owner
	clipboardAtom := s.GetAtom("CLIPBOARD")
	if !ok && selectionAtom == clipboardAtom {
		go func() {
			content, err := s.frontend.ReadClipboard()
			if err == nil {
				// Write to property
				// Assuming target is STRING or UTF8_STRING or TEXT
				// Ideally check targetAtom. For now, assume string.
				s.ChangeProperty(requestor, propertyAtom, s.GetAtom("STRING"), 8, []byte(content))

				s.SendSelectionNotify(requestor, selectionAtom, targetAtom, propertyAtom, nil)
			} else {
				// Failed
				s.SendSelectionNotify(requestor, selectionAtom, targetAtom, 0, nil)
			}
		}()
		return nil
	}

	if ok {
		// Send SelectionRequest to owner
		if ownerClient, ok := s.clients[owner.window.client]; ok {
			event := &wire.SelectionRequestEvent{
				Sequence:  ownerClient.sequence - 1,
				Time:      time,
				Owner:     owner.window.local,
				Requestor: requestor.local,
				Selection: selectionAtom,
				Target:    targetAtom,
				Property:  propertyAtom,
			}
			s.sendEvent(ownerClient, event)
		}
		return nil
	}

	// No owner, send SelectionNotify with property None
	s.SendSelectionNotify(requestor, selectionAtom, targetAtom, 0, nil)
	return nil
}

func (s *x11Server) handleSendEvent(client *x11Client, req wire.Request, seq uint16) messageEncoder {
	p := req.(*wire.SendEventRequest)
	// The X11 client sends an event to another client.
	// We need to forward this event to the appropriate frontend.
	// For now, we'll just log it and pass it to the frontend.
	s.frontend.SendEvent(&wire.X11RawEvent{Data: p.EventData})
	return nil
}

func (s *x11Server) handleGrabPointer(client *x11Client, req wire.Request, seq uint16) messageEncoder {
	p := req.(*wire.GrabPointerRequest)
	grabWindow := client.xID(uint32(p.GrabWindow))
	if err := s.checkWindow(grabWindow, seq, wire.GrabPointer, 0); err != nil {
		return err
	}
	if p.ConfineTo != 0 {
		if err := s.checkWindow(client.xID(uint32(p.ConfineTo)), seq, wire.GrabPointer, 0); err != nil {
			return err
		}
	}
	if p.Cursor != 0 {
		if err := s.checkCursor(client.xID(uint32(p.Cursor)), seq, wire.GrabPointer, 0); err != nil {
			return err
		}
	}

	if p.Time != 0 {
		if uint32(p.Time) < s.pointerGrabTime || uint32(p.Time) > s.serverTime() {
			return &wire.GrabPointerReply{Sequence: seq, Status: wire.GrabInvalidTime}
		}
	}

	if s.pointerGrabWindow.local != 0 && s.pointerGrabWindow.client != client.id {
		return &wire.GrabPointerReply{
			Sequence: seq,
			Status:   wire.AlreadyGrabbed,
		}
	}

	s.pointerGrabWindow = grabWindow
	s.pointerGrabWindow.client = client.id
	s.pointerGrabOwner = p.OwnerEvents
	s.pointerGrabEventMask = p.EventMask
	s.pointerGrabTime = uint32(p.Time)
	if s.pointerGrabTime == 0 {
		s.pointerGrabTime = s.serverTime()
	}
	s.pointerGrabMode = p.PointerMode
	s.keyboardGrabMode = p.KeyboardMode
	s.pointerGrabConfineTo = client.xID(uint32(p.ConfineTo))
	s.pointerGrabCursor = client.xID(uint32(p.Cursor))

	if p.Cursor != 0 {
		s.frontend.SetWindowCursor(grabWindow, s.pointerGrabCursor)
	}

	return &wire.GrabPointerReply{
		Sequence: seq,
		Status:   wire.GrabSuccess,
	}
}

func (s *x11Server) handleUngrabPointer(client *x11Client, req wire.Request, seq uint16) messageEncoder {
	p := req.(*wire.UngrabPointerRequest)
	if p.Time != 0 && uint32(p.Time) < s.pointerGrabTime {
		// Ignore
		return nil
	}
	s.pointerGrabWindow = xID{}
	s.pointerGrabOwner = false
	s.pointerGrabEventMask = 0
	s.pointerGrabTime = 0
	s.pointerGrabMode = 0
	s.keyboardGrabMode = 0
	s.pointerGrabConfineTo = xID{}
	s.pointerGrabCursor = xID{}
	return nil
}

func (s *x11Server) handleGrabButton(client *x11Client, req wire.Request, seq uint16) messageEncoder {
	p := req.(*wire.GrabButtonRequest)
	var grabWindow xID
	var found bool
	for wID := range s.windows {
		if wID.local == uint32(p.GrabWindow) {
			grabWindow = wID
			found = true
			break
		}
	}
	if !found {
		return wire.NewGenericError(seq, uint32(p.GrabWindow), 0, wire.GrabButton, wire.WindowErrorCode)
	}

	grab := &passiveGrab{
		clientID:     client.id,
		button:       p.Button,
		modifiers:    p.Modifiers,
		owner:        p.OwnerEvents,
		eventMask:    p.EventMask,
		cursor:       client.xID(uint32(p.Cursor)),
		pointerMode:  p.PointerMode,
		keyboardMode: p.KeyboardMode,
		confineTo:    client.xID(uint32(p.ConfineTo)),
	}
	s.passiveGrabs[grabWindow] = append(s.passiveGrabs[grabWindow], grab)
	return nil
}

func (s *x11Server) handleUngrabButton(client *x11Client, req wire.Request, seq uint16) messageEncoder {
	p := req.(*wire.UngrabButtonRequest)
	grabWindow := client.xID(uint32(p.GrabWindow))
	if err := s.checkWindow(grabWindow, seq, wire.UngrabButton, 0); err != nil {
		return err
	}
	if grabs, ok := s.passiveGrabs[grabWindow]; ok {
		for i, grab := range grabs {
			if grab.button == p.Button && grab.modifiers == p.Modifiers {
				s.passiveGrabs[grabWindow] = append(grabs[:i], grabs[i+1:]...)
				break
			}
		}
	}
	return nil
}

func (s *x11Server) handleChangeActivePointerGrab(client *x11Client, req wire.Request, seq uint16) messageEncoder {
	p := req.(*wire.ChangeActivePointerGrabRequest)
	if s.pointerGrabWindow.client == client.id && s.pointerGrabWindow.local != 0 {
		if p.Cursor != 0 {
			cursorXID := client.xID(uint32(p.Cursor))
			if err := s.checkCursor(cursorXID, seq, wire.ChangeActivePointerGrab, 0); err != nil {
				return err
			}
			s.frontend.SetWindowCursor(s.pointerGrabWindow, cursorXID)
		}
		s.pointerGrabEventMask = p.EventMask
		if p.Time == 0 || uint32(p.Time) >= s.pointerGrabTime {
			s.pointerGrabTime = uint32(p.Time)
		}
	}
	return nil
}

func (s *x11Server) handleGrabKeyboard(client *x11Client, req wire.Request, seq uint16) messageEncoder {
	p := req.(*wire.GrabKeyboardRequest)
	grabWindow := client.xID(uint32(p.GrabWindow))
	if err := s.checkWindow(grabWindow, seq, wire.GrabKeyboard, 0); err != nil {
		return err
	}
	if p.Time != 0 {
		if uint32(p.Time) < s.keyboardGrabTime || uint32(p.Time) > s.serverTime() {
			return &wire.GrabKeyboardReply{Sequence: seq, Status: wire.GrabInvalidTime}
		}
	}

	if s.keyboardGrabWindow.local != 0 && s.keyboardGrabWindow.client != client.id {
		return &wire.GrabKeyboardReply{
			Sequence: seq,
			Status:   wire.AlreadyGrabbed,
		}
	}

	s.keyboardGrabWindow = grabWindow
	s.keyboardGrabWindow.client = client.id
	s.keyboardGrabOwner = p.OwnerEvents
	s.keyboardGrabTime = uint32(p.Time)
	if s.keyboardGrabTime == 0 {
		s.keyboardGrabTime = s.serverTime()
	}
	s.keyboardGrabMode = p.KeyboardMode
	s.pointerGrabMode = p.PointerMode

	return &wire.GrabKeyboardReply{
		Sequence: seq,
		Status:   wire.GrabSuccess,
	}
}

func (s *x11Server) handleUngrabKeyboard(client *x11Client, req wire.Request, seq uint16) messageEncoder {
	p := req.(*wire.UngrabKeyboardRequest)
	if p.Time != 0 && uint32(p.Time) < s.keyboardGrabTime {
		// Ignore
		return nil
	}
	s.keyboardGrabWindow = xID{}
	s.keyboardGrabOwner = false
	s.keyboardGrabTime = 0
	s.keyboardGrabMode = 0
	return nil
}

func (s *x11Server) handleGrabKey(client *x11Client, req wire.Request, seq uint16) messageEncoder {
	p := req.(*wire.GrabKeyRequest)
	var grabWindow xID
	var found bool
	for wID := range s.windows {
		if wID.local == uint32(p.GrabWindow) {
			grabWindow = wID
			found = true
			break
		}
	}
	if !found {
		return wire.NewGenericError(seq, uint32(p.GrabWindow), 0, wire.GrabKey, wire.WindowErrorCode)
	}
	grab := &passiveGrab{
		clientID:     client.id,
		key:          p.Key,
		modifiers:    p.Modifiers,
		owner:        p.OwnerEvents,
		pointerMode:  p.PointerMode,
		keyboardMode: p.KeyboardMode,
	}
	s.passiveGrabs[grabWindow] = append(s.passiveGrabs[grabWindow], grab)
	return nil
}

func (s *x11Server) handleUngrabKey(client *x11Client, req wire.Request, seq uint16) messageEncoder {
	p := req.(*wire.UngrabKeyRequest)
	grabWindow := client.xID(uint32(p.GrabWindow))
	if err := s.checkWindow(grabWindow, seq, wire.UngrabKey, 0); err != nil {
		return err
	}
	if grabs, ok := s.passiveGrabs[grabWindow]; ok {
		newGrabs := make([]*passiveGrab, 0, len(grabs))
		for _, grab := range grabs {
			if !(grab.key == p.Key && (p.Modifiers == wire.AnyModifier || grab.modifiers == p.Modifiers)) {
				newGrabs = append(newGrabs, grab)
			}
		}
		s.passiveGrabs[grabWindow] = newGrabs
	}
	return nil
}

func (s *x11Server) handleAllowEvents(client *x11Client, req wire.Request, seq uint16) messageEncoder {
	p := req.(*wire.AllowEventsRequest)
	s.frontend.AllowEvents(client.id, p.Mode, uint32(p.Time))
	return nil
}

func (s *x11Server) handleGrabServer(client *x11Client, req wire.Request, seq uint16) messageEncoder {
	if !s.serverGrabbed {
		s.serverGrabbed = true
		s.grabbingClientID = client.id
	}
	return nil
}

func (s *x11Server) handleUngrabServer(client *x11Client, req wire.Request, seq uint16) messageEncoder {
	s.serverGrabbed = false
	s.grabbingClientID = 0
	return nil
}

func (s *x11Server) handleQueryPointer(client *x11Client, req wire.Request, seq uint16) messageEncoder {
	p := req.(*wire.QueryPointerRequest)
	xid := client.xID(uint32(p.Drawable))
	if err := s.checkDrawable(xid, seq, wire.QueryPointer, 0); err != nil {
		return err
	}
	debugf("X11: QueryPointer drawable=%d", xid)
	return &wire.QueryPointerReply{
		Sequence:   seq,
		SameScreen: true,
		Root:       s.rootWindowID(),
		Child:      uint32(p.Drawable),
		RootX:      s.pointerX,
		RootY:      s.pointerY,
		WinX:       s.pointerX, // Assuming pointer is always in the window for now
		WinY:       s.pointerY, // Assuming pointer is always in the window for now
		Mask:       0,          // No buttons pressed
	}
}

func (s *x11Server) handleGetMotionEvents(client *x11Client, req wire.Request, seq uint16) messageEncoder {
	return &wire.GetMotionEventsReply{
		Sequence: seq,
		Events:   []wire.TimeCoord{},
	}
}

func (s *x11Server) handleTranslateCoords(client *x11Client, req wire.Request, seq uint16) messageEncoder {
	p := req.(*wire.TranslateCoordsRequest)
	srcWindow := client.xID(uint32(p.SrcWindow))
	dstWindow := client.xID(uint32(p.DstWindow))

	if err := s.checkWindow(srcWindow, seq, wire.TranslateCoords, 0); err != nil {
		return err
	}
	if err := s.checkWindow(dstWindow, seq, wire.TranslateCoords, 0); err != nil {
		return err
	}

	// Simplified implementation: assume windows are direct children of the root
	src := s.windows[srcWindow]
	dst := s.windows[dstWindow]

	var srcAbsX, srcAbsY int16
	if src != nil {
		srcAbsX = src.x
		srcAbsY = src.y
	}
	var dstAbsX, dstAbsY int16
	if dst != nil {
		dstAbsX = dst.x
		dstAbsY = dst.y
	}

	dstX := srcAbsX + p.SrcX - dstAbsX
	dstY := srcAbsY + p.SrcY - dstAbsY

	return &wire.TranslateCoordsReply{
		Sequence:   seq,
		SameScreen: true,
		Child:      0, // No child for now
		DstX:       dstX,
		DstY:       dstY,
	}
}

func (s *x11Server) handleWarpPointer(client *x11Client, req wire.Request, seq uint16) messageEncoder {
	p := req.(*wire.WarpPointerRequest)
	if p.SrcWindow != 0 {
		if err := s.checkWindow(client.xID(p.SrcWindow), seq, wire.WarpPointer, 0); err != nil {
			return err
		}
	}
	if p.DstWindow != 0 {
		if err := s.checkWindow(client.xID(p.DstWindow), seq, wire.WarpPointer, 0); err != nil {
			return err
		}
	}
	s.frontend.WarpPointer(p.DstX, p.DstY)
	return nil
}

func (s *x11Server) handleSetInputFocus(client *x11Client, req wire.Request, seq uint16) messageEncoder {
	p := req.(*wire.SetInputFocusRequest)
	xid := client.xID(uint32(p.Focus))
	// Focus can be None(0) or PointerRoot(1).
	if uint32(p.Focus) > 1 {
		if err := s.checkWindow(xid, seq, wire.SetInputFocus, 0); err != nil {
			return err
		}
	}
	s.inputFocus = xid
	s.frontend.SetInputFocus(xid, p.RevertTo)
	return nil
}

func (s *x11Server) handleGetInputFocus(client *x11Client, req wire.Request, seq uint16) messageEncoder {
	return &wire.GetInputFocusReply{
		Sequence: seq,
		RevertTo: 1, // RevertToParent
		Focus:    s.frontend.GetFocusWindow(client.id).local,
	}
}

func (s *x11Server) handleQueryKeymap(client *x11Client, req wire.Request, seq uint16) messageEncoder {
	return &wire.QueryKeymapReply{
		Sequence: seq,
		Keys:     [32]byte{},
	}
}

func (s *x11Server) handleOpenFont(client *x11Client, req wire.Request, seq uint16) messageEncoder {
	p := req.(*wire.OpenFontRequest)
	fid := client.xID(uint32(p.Fid))
	if _, exists := s.fonts[fid]; exists {
		return wire.NewGenericError(seq, uint32(p.Fid), 0, wire.OpenFont, wire.IDChoiceErrorCode)
	}
	s.fonts[fid] = true
	s.frontend.OpenFont(fid, p.Name)
	return nil
}

func (s *x11Server) handleCloseFont(client *x11Client, req wire.Request, seq uint16) messageEncoder {
	p := req.(*wire.CloseFontRequest)
	fid := client.xID(uint32(p.Fid))
	if err := s.checkFont(fid, seq, wire.CloseFont, 0); err != nil {
		return err
	}
	delete(s.fonts, fid)
	s.frontend.CloseFont(fid)
	return nil
}

func (s *x11Server) handleQueryFont(client *x11Client, req wire.Request, seq uint16) messageEncoder {
	p := req.(*wire.QueryFontRequest)
	fid := client.xID(uint32(p.Fid))
	if err := s.checkFont(fid, seq, wire.QueryFont, 0); err != nil {
		return err
	}
	minBounds, maxBounds, minCharOrByte2, maxCharOrByte2, defaultChar, drawDirection, minByte1, maxByte1, allCharsExist, fontAscent, fontDescent, charInfos, fontProps := s.frontend.QueryFont(fid)

	return &wire.QueryFontReply{
		Sequence:       seq,
		MinBounds:      minBounds,
		MaxBounds:      maxBounds,
		MinCharOrByte2: minCharOrByte2,
		MaxCharOrByte2: maxCharOrByte2,
		DefaultChar:    defaultChar,
		NumFontProps:   uint16(len(fontProps)),
		DrawDirection:  drawDirection,
		MinByte1:       minByte1,
		MaxByte1:       maxByte1,
		AllCharsExist:  allCharsExist,
		FontAscent:     fontAscent,
		FontDescent:    fontDescent,
		NumCharInfos:   uint32(len(charInfos)),
		CharInfos:      charInfos,
		FontProps:      fontProps,
	}
}

func (s *x11Server) handleQueryTextExtents(client *x11Client, req wire.Request, seq uint16) messageEncoder {
	p := req.(*wire.QueryTextExtentsRequest)
	fid := client.xID(uint32(p.Fid))
	if err := s.checkFont(fid, seq, wire.QueryTextExtents, 0); err != nil {
		return err
	}
	drawDirection, fontAscent, fontDescent, overallAscent, overallDescent, overallWidth, overallLeft, overallRight := s.frontend.QueryTextExtents(fid, p.Text)
	return &wire.QueryTextExtentsReply{
		Sequence:       seq,
		DrawDirection:  drawDirection,
		FontAscent:     fontAscent,
		FontDescent:    fontDescent,
		OverallAscent:  overallAscent,
		OverallDescent: overallDescent,
		OverallWidth:   int32(overallWidth),
		OverallLeft:    int32(overallLeft),
		OverallRight:   int32(overallRight),
	}
}

func (s *x11Server) handleListFonts(client *x11Client, req wire.Request, seq uint16) messageEncoder {
	p := req.(*wire.ListFontsRequest)
	fontNames := s.frontend.ListFonts(p.MaxNames, p.Pattern)

	return &wire.ListFontsReply{
		Sequence:  seq,
		FontNames: fontNames,
	}
}

func (s *x11Server) handleListFontsWithInfo(client *x11Client, req wire.Request, seq uint16) messageEncoder {
	p := req.(*wire.ListFontsWithInfoRequest)
	fontNames := s.frontend.ListFonts(p.MaxNames, p.Pattern)
	tempFID := xID{client: 0, local: 0xFFFFFFFF}

	for _, name := range fontNames {
		s.frontend.OpenFont(tempFID, name)
		minBounds, maxBounds, minCharOrByte2, maxCharOrByte2, defaultChar, drawDirection, minByte1, maxByte1, allCharsExist, fontAscent, fontDescent, _, fontProps := s.frontend.QueryFont(tempFID)
		s.frontend.CloseFont(tempFID)

		reply := &wire.ListFontsWithInfoReply{
			Sequence:      seq,
			NameLength:    byte(len(name)),
			MinBounds:     minBounds,
			MaxBounds:     maxBounds,
			MinChar:       minCharOrByte2,
			MaxChar:       maxCharOrByte2,
			DefaultChar:   defaultChar,
			NFontProps:    uint16(len(fontProps)),
			DrawDirection: drawDirection,
			MinByte1:      minByte1,
			MaxByte1:      maxByte1,
			AllCharsExist: allCharsExist,
			FontAscent:    fontAscent,
			FontDescent:   fontDescent,
			NReplies:      1 + uint32(len(fontNames)-1), // This field is not actually used by clients usually, but let's be approximate
			FontProps:     fontProps,
			FontName:      name,
		}
		client.send(reply)
	}

	// Final reply
	lastReply := &wire.ListFontsWithInfoReply{
		Sequence: seq,
		FontName: "",
	}
	client.send(lastReply)
	return nil
}

func (s *x11Server) handleSetFontPath(client *x11Client, req wire.Request, seq uint16) messageEncoder {
	p := req.(*wire.SetFontPathRequest)
	s.fontPath = p.Paths
	return nil
}

func (s *x11Server) handleGetFontPath(client *x11Client, req wire.Request, seq uint16) messageEncoder {
	return &wire.GetFontPathReply{
		Sequence: seq,
		Paths:    s.fontPath,
	}
}

func (s *x11Server) handleCreatePixmap(client *x11Client, req wire.Request, seq uint16) messageEncoder {
	p := req.(*wire.CreatePixmapRequest)
	xid := client.xID(uint32(p.Pid))

	// Check if the pixmap ID is already in use
	if _, exists := s.pixmaps[xid]; exists {
		s.logger.Errorf("X11: CreatePixmap: ID %s already in use", xid)
		return wire.NewGenericError(seq, uint32(p.Pid), 0, wire.CreatePixmap, wire.IDChoiceErrorCode)
	}
	if err := s.checkDrawable(client.xID(uint32(p.Drawable)), seq, wire.CreatePixmap, 0); err != nil {
		return err
	}

	s.pixmaps[xid] = true // Mark pixmap ID as used
	s.frontend.CreatePixmap(xid, client.xID(uint32(p.Drawable)), uint32(p.Width), uint32(p.Height), uint32(p.Depth))
	return nil
}

func (s *x11Server) handleFreePixmap(client *x11Client, req wire.Request, seq uint16) messageEncoder {
	p := req.(*wire.FreePixmapRequest)
	xid := client.xID(uint32(p.Pid))
	if err := s.checkPixmap(xid, seq, wire.FreePixmap, 0); err != nil {
		return err
	}
	delete(s.pixmaps, xid)
	s.frontend.FreePixmap(xid)
	return nil
}

func (s *x11Server) handleCreateGC(client *x11Client, req wire.Request, seq uint16) messageEncoder {
	p := req.(*wire.CreateGCRequest)
	xid := client.xID(uint32(p.Cid))

	// Check if the GC ID is already in use
	if _, exists := s.gcs[xid]; exists {
		s.logger.Errorf("X11: CreateGC: ID %s already in use", xid)
		return wire.NewGenericError(seq, uint32(xid.local), 0, wire.CreateGC, wire.IDChoiceErrorCode)
	}
	if err := s.checkDrawable(client.xID(uint32(p.Drawable)), seq, wire.CreateGC, 0); err != nil {
		return err
	}

	s.gcs[xid] = p.Values
	s.frontend.CreateGC(xid, p.ValueMask, p.Values)
	return nil
}

func (s *x11Server) handleChangeGC(client *x11Client, req wire.Request, seq uint16) messageEncoder {
	p := req.(*wire.ChangeGCRequest)
	xid := client.xID(uint32(p.Gc))
	if err := s.checkGC(xid, seq, wire.ChangeGC, 0); err != nil {
		return err
	}
	if existingGC, ok := s.gcs[xid]; ok {
		if p.ValueMask&wire.GCFunction != 0 {
			existingGC.Function = p.Values.Function
		}
		if p.ValueMask&wire.GCPlaneMask != 0 {
			existingGC.PlaneMask = p.Values.PlaneMask
		}
		if p.ValueMask&wire.GCForeground != 0 {
			existingGC.Foreground = p.Values.Foreground
		}
		if p.ValueMask&wire.GCBackground != 0 {
			existingGC.Background = p.Values.Background
		}
		if p.ValueMask&wire.GCLineWidth != 0 {
			existingGC.LineWidth = p.Values.LineWidth
		}
		if p.ValueMask&wire.GCLineStyle != 0 {
			existingGC.LineStyle = p.Values.LineStyle
		}
		if p.ValueMask&wire.GCCapStyle != 0 {
			existingGC.CapStyle = p.Values.CapStyle
		}
		if p.ValueMask&wire.GCJoinStyle != 0 {
			existingGC.JoinStyle = p.Values.JoinStyle
		}
		if p.ValueMask&wire.GCFillStyle != 0 {
			existingGC.FillStyle = p.Values.FillStyle
		}
		if p.ValueMask&wire.GCFillRule != 0 {
			existingGC.FillRule = p.Values.FillRule
		}
		if p.ValueMask&wire.GCTile != 0 {
			existingGC.Tile = p.Values.Tile
		}
		if p.ValueMask&wire.GCStipple != 0 {
			existingGC.Stipple = p.Values.Stipple
		}
		if p.ValueMask&wire.GCTileStipXOrigin != 0 {
			existingGC.TileStipXOrigin = p.Values.TileStipXOrigin
		}
		if p.ValueMask&wire.GCTileStipYOrigin != 0 {
			existingGC.TileStipYOrigin = p.Values.TileStipYOrigin
		}
		if p.ValueMask&wire.GCFont != 0 {
			existingGC.Font = p.Values.Font
		}
		if p.ValueMask&wire.GCSubwindowMode != 0 {
			existingGC.SubwindowMode = p.Values.SubwindowMode
		}
		if p.ValueMask&wire.GCGraphicsExposures != 0 {
			existingGC.GraphicsExposures = p.Values.GraphicsExposures
		}
		if p.ValueMask&wire.GCClipXOrigin != 0 {
			existingGC.ClipXOrigin = p.Values.ClipXOrigin
		}
		if p.ValueMask&wire.GCClipYOrigin != 0 {
			existingGC.ClipYOrigin = p.Values.ClipYOrigin
		}
		if p.ValueMask&wire.GCClipMask != 0 {
			existingGC.ClipMask = p.Values.ClipMask
		}
		if p.ValueMask&wire.GCDashOffset != 0 {
			existingGC.DashOffset = p.Values.DashOffset
		}
		if p.ValueMask&wire.GCDashes != 0 {
			existingGC.Dashes = p.Values.Dashes
		}
		if p.ValueMask&wire.GCArcMode != 0 {
			existingGC.ArcMode = p.Values.ArcMode
		}
	}
	s.frontend.ChangeGC(xid, p.ValueMask, p.Values)
	return nil
}

func (s *x11Server) handleCopyGC(client *x11Client, req wire.Request, seq uint16) messageEncoder {
	p := req.(*wire.CopyGCRequest)
	srcGC := client.xID(uint32(p.SrcGC))
	dstGC := client.xID(uint32(p.DstGC))
	if err := s.checkGC(srcGC, seq, wire.CopyGC, 0); err != nil {
		return err
	}
	if err := s.checkGC(dstGC, seq, wire.CopyGC, 0); err != nil {
		return err
	}
	s.frontend.CopyGC(srcGC, dstGC)
	return nil
}

func (s *x11Server) handleSetDashes(client *x11Client, req wire.Request, seq uint16) messageEncoder {
	p := req.(*wire.SetDashesRequest)
	gc := client.xID(uint32(p.GC))
	if err := s.checkGC(gc, seq, wire.SetDashes, 0); err != nil {
		return err
	}
	s.frontend.SetDashes(gc, p.DashOffset, p.Dashes)
	return nil
}

func (s *x11Server) handleSetClipRectangles(client *x11Client, req wire.Request, seq uint16) messageEncoder {
	p := req.(*wire.SetClipRectanglesRequest)
	gc := client.xID(uint32(p.GC))
	if err := s.checkGC(gc, seq, wire.SetClipRectangles, 0); err != nil {
		return err
	}
	s.frontend.SetClipRectangles(gc, p.ClippingX, p.ClippingY, p.Rectangles, p.Ordering)
	return nil
}

func (s *x11Server) handleFreeGC(client *x11Client, req wire.Request, seq uint16) messageEncoder {
	p := req.(*wire.FreeGCRequest)
	gcID := client.xID(uint32(p.GC))
	if err := s.checkGC(gcID, seq, wire.FreeGC, 0); err != nil {
		return err
	}
	delete(s.gcs, gcID)
	s.frontend.FreeGC(gcID)
	return nil
}

func (s *x11Server) handleClearArea(client *x11Client, req wire.Request, seq uint16) messageEncoder {
	p := req.(*wire.ClearAreaRequest)
	if err := s.checkWindow(client.xID(uint32(p.Window)), seq, wire.ClearArea, 0); err != nil {
		return err
	}
	s.frontend.ClearArea(client.xID(uint32(p.Window)), int32(p.X), int32(p.Y), int32(p.Width), int32(p.Height))
	s.frontend.ComposeWindow(client.xID(uint32(p.Window)))
	return nil
}

func (s *x11Server) handleCopyArea(client *x11Client, req wire.Request, seq uint16) messageEncoder {
	p := req.(*wire.CopyAreaRequest)
	gcID := client.xID(uint32(p.Gc))
	if err := s.checkGC(gcID, seq, wire.CopyArea, 0); err != nil {
		return err
	}
	if err := s.checkDrawable(client.xID(uint32(p.SrcDrawable)), seq, wire.CopyArea, 0); err != nil {
		return err
	}
	if err := s.checkDrawable(client.xID(uint32(p.DstDrawable)), seq, wire.CopyArea, 0); err != nil {
		return err
	}
	s.frontend.CopyArea(client.xID(uint32(p.SrcDrawable)), client.xID(uint32(p.DstDrawable)), gcID, int32(p.SrcX), int32(p.SrcY), int32(p.DstX), int32(p.DstY), int32(p.Width), int32(p.Height))
	s.frontend.ComposeWindow(client.xID(uint32(p.DstDrawable)))
	return nil
}

func (s *x11Server) handleCopyPlane(client *x11Client, req wire.Request, seq uint16) messageEncoder {
	p := req.(*wire.CopyPlaneRequest)
	gcID := client.xID(uint32(p.Gc))
	if err := s.checkGC(gcID, seq, wire.CopyPlane, 0); err != nil {
		return err
	}
	if err := s.checkDrawable(client.xID(uint32(p.SrcDrawable)), seq, wire.CopyPlane, 0); err != nil {
		return err
	}
	if err := s.checkDrawable(client.xID(uint32(p.DstDrawable)), seq, wire.CopyPlane, 0); err != nil {
		return err
	}
	s.frontend.CopyPlane(client.xID(uint32(p.SrcDrawable)), client.xID(uint32(p.DstDrawable)), gcID, int32(p.SrcX), int32(p.SrcY), int32(p.DstX), int32(p.DstY), int32(p.Width), int32(p.Height), int32(p.PlaneMask))
	s.frontend.ComposeWindow(client.xID(uint32(p.DstDrawable)))
	return nil
}

func (s *x11Server) handlePolyPoint(client *x11Client, req wire.Request, seq uint16) messageEncoder {
	p := req.(*wire.PolyPointRequest)
	gcID := client.xID(uint32(p.Gc))
	if err := s.checkGC(gcID, seq, wire.PolyPoint, 0); err != nil {
		return err
	}
	if err := s.checkDrawable(client.xID(uint32(p.Drawable)), seq, wire.PolyPoint, 0); err != nil {
		return err
	}
	s.frontend.PolyPoint(client.xID(uint32(p.Drawable)), gcID, p.Coordinates)
	s.frontend.ComposeWindow(client.xID(uint32(p.Drawable)))
	return nil
}

func (s *x11Server) handlePolyLine(client *x11Client, req wire.Request, seq uint16) messageEncoder {
	p := req.(*wire.PolyLineRequest)
	gcID := client.xID(uint32(p.Gc))
	if err := s.checkGC(gcID, seq, wire.PolyLine, 0); err != nil {
		return err
	}
	if err := s.checkDrawable(client.xID(uint32(p.Drawable)), seq, wire.PolyLine, 0); err != nil {
		return err
	}
	s.frontend.PolyLine(client.xID(uint32(p.Drawable)), gcID, p.Coordinates)
	s.frontend.ComposeWindow(client.xID(uint32(p.Drawable)))
	return nil
}

func (s *x11Server) handlePolySegment(client *x11Client, req wire.Request, seq uint16) messageEncoder {
	p := req.(*wire.PolySegmentRequest)
	gcID := client.xID(uint32(p.Gc))
	if err := s.checkGC(gcID, seq, wire.PolySegment, 0); err != nil {
		return err
	}
	if err := s.checkDrawable(client.xID(uint32(p.Drawable)), seq, wire.PolySegment, 0); err != nil {
		return err
	}
	s.frontend.PolySegment(client.xID(uint32(p.Drawable)), gcID, p.Segments)
	s.frontend.ComposeWindow(client.xID(uint32(p.Drawable)))
	return nil
}

func (s *x11Server) handlePolyRectangle(client *x11Client, req wire.Request, seq uint16) messageEncoder {
	p := req.(*wire.PolyRectangleRequest)
	gcID := client.xID(uint32(p.Gc))
	if err := s.checkGC(gcID, seq, wire.PolyRectangle, 0); err != nil {
		return err
	}
	if err := s.checkDrawable(client.xID(uint32(p.Drawable)), seq, wire.PolyRectangle, 0); err != nil {
		return err
	}
	s.frontend.PolyRectangle(client.xID(uint32(p.Drawable)), gcID, p.Rectangles)
	s.frontend.ComposeWindow(client.xID(uint32(p.Drawable)))
	return nil
}

func (s *x11Server) handlePolyArc(client *x11Client, req wire.Request, seq uint16) messageEncoder {
	p := req.(*wire.PolyArcRequest)
	gcID := client.xID(uint32(p.Gc))
	if err := s.checkGC(gcID, seq, wire.PolyArc, 0); err != nil {
		return err
	}
	if err := s.checkDrawable(client.xID(uint32(p.Drawable)), seq, wire.PolyArc, 0); err != nil {
		return err
	}
	s.frontend.PolyArc(client.xID(uint32(p.Drawable)), gcID, p.Arcs)
	s.frontend.ComposeWindow(client.xID(uint32(p.Drawable)))
	return nil
}

func (s *x11Server) handleFillPoly(client *x11Client, req wire.Request, seq uint16) messageEncoder {
	p := req.(*wire.FillPolyRequest)
	gcID := client.xID(uint32(p.Gc))
	if err := s.checkGC(gcID, seq, wire.FillPoly, 0); err != nil {
		return err
	}
	if err := s.checkDrawable(client.xID(uint32(p.Drawable)), seq, wire.FillPoly, 0); err != nil {
		return err
	}
	s.frontend.FillPoly(client.xID(uint32(p.Drawable)), gcID, p.Coordinates)
	s.frontend.ComposeWindow(client.xID(uint32(p.Drawable)))
	return nil
}

func (s *x11Server) handlePolyFillRectangle(client *x11Client, req wire.Request, seq uint16) messageEncoder {
	p := req.(*wire.PolyFillRectangleRequest)
	gcID := client.xID(uint32(p.Gc))
	if err := s.checkGC(gcID, seq, wire.PolyFillRectangle, 0); err != nil {
		return err
	}
	if err := s.checkDrawable(client.xID(uint32(p.Drawable)), seq, wire.PolyFillRectangle, 0); err != nil {
		return err
	}
	s.frontend.PolyFillRectangle(client.xID(uint32(p.Drawable)), gcID, p.Rectangles)
	s.frontend.ComposeWindow(client.xID(uint32(p.Drawable)))
	return nil
}

func (s *x11Server) handlePolyFillArc(client *x11Client, req wire.Request, seq uint16) messageEncoder {
	p := req.(*wire.PolyFillArcRequest)
	gcID := client.xID(uint32(p.Gc))
	if err := s.checkGC(gcID, seq, wire.PolyFillArc, 0); err != nil {
		return err
	}
	if err := s.checkDrawable(client.xID(uint32(p.Drawable)), seq, wire.PolyFillArc, 0); err != nil {
		return err
	}
	s.frontend.PolyFillArc(client.xID(uint32(p.Drawable)), gcID, p.Arcs)
	s.frontend.ComposeWindow(client.xID(uint32(p.Drawable)))
	return nil
}

func (s *x11Server) handlePutImage(client *x11Client, req wire.Request, seq uint16) messageEncoder {
	p := req.(*wire.PutImageRequest)
	gcID := client.xID(uint32(p.Gc))
	if err := s.checkGC(gcID, seq, wire.PutImage, 0); err != nil {
		return err
	}
	if err := s.checkDrawable(client.xID(uint32(p.Drawable)), seq, wire.PutImage, 0); err != nil {
		return err
	}
	s.frontend.PutImage(client.xID(uint32(p.Drawable)), gcID, p.Format, p.Width, p.Height, p.DstX, p.DstY, p.LeftPad, p.Depth, p.Data)
	s.frontend.ComposeWindow(client.xID(uint32(p.Drawable)))
	return nil
}

func (s *x11Server) handleGetImage(client *x11Client, req wire.Request, seq uint16) messageEncoder {
	p := req.(*wire.GetImageRequest)
	if err := s.checkDrawable(client.xID(uint32(p.Drawable)), seq, wire.GetImage, 0); err != nil {
		return err
	}
	imgData, err := s.frontend.GetImage(client.xID(uint32(p.Drawable)), int32(p.X), int32(p.Y), int32(p.Width), int32(p.Height), p.PlaneMask)
	if err != nil {
		return wire.NewGenericError(seq, 0, 0, wire.GetImage, wire.MatchErrorCode)
	}
	return &wire.GetImageReply{
		Sequence:  seq,
		Depth:     24, // Assuming 24-bit depth for now
		VisualID:  s.visualID,
		ImageData: imgData,
	}
}

func (s *x11Server) handlePolyText8(client *x11Client, req wire.Request, seq uint16) messageEncoder {
	p := req.(*wire.PolyText8Request)
	gcID := client.xID(uint32(p.GC))
	if err := s.checkGC(gcID, seq, wire.PolyText8, 0); err != nil {
		return err
	}
	if err := s.checkDrawable(client.xID(uint32(p.Drawable)), seq, wire.PolyText8, 0); err != nil {
		return err
	}
	s.frontend.PolyText8(client.xID(uint32(p.Drawable)), gcID, int32(p.X), int32(p.Y), p.Items)
	s.frontend.ComposeWindow(client.xID(uint32(p.Drawable)))
	return nil
}

func (s *x11Server) handlePolyText16(client *x11Client, req wire.Request, seq uint16) messageEncoder {
	p := req.(*wire.PolyText16Request)
	gcID := client.xID(uint32(p.GC))
	if err := s.checkGC(gcID, seq, wire.PolyText16, 0); err != nil {
		return err
	}
	if err := s.checkDrawable(client.xID(uint32(p.Drawable)), seq, wire.PolyText16, 0); err != nil {
		return err
	}
	s.frontend.PolyText16(client.xID(uint32(p.Drawable)), gcID, int32(p.X), int32(p.Y), p.Items)
	s.frontend.ComposeWindow(client.xID(uint32(p.Drawable)))
	return nil
}

func (s *x11Server) handleImageText8(client *x11Client, req wire.Request, seq uint16) messageEncoder {
	p := req.(*wire.ImageText8Request)
	gcID := client.xID(uint32(p.Gc))
	if err := s.checkGC(gcID, seq, wire.ImageText8, 0); err != nil {
		return err
	}
	if err := s.checkDrawable(client.xID(uint32(p.Drawable)), seq, wire.ImageText8, 0); err != nil {
		return err
	}
	s.frontend.ImageText8(client.xID(uint32(p.Drawable)), gcID, int32(p.X), int32(p.Y), p.Text)
	s.frontend.ComposeWindow(client.xID(uint32(p.Drawable)))
	return nil
}

func (s *x11Server) handleImageText16(client *x11Client, req wire.Request, seq uint16) messageEncoder {
	p := req.(*wire.ImageText16Request)
	gcID := client.xID(uint32(p.Gc))
	if err := s.checkGC(gcID, seq, wire.ImageText16, 0); err != nil {
		return err
	}
	if err := s.checkDrawable(client.xID(uint32(p.Drawable)), seq, wire.ImageText16, 0); err != nil {
		return err
	}
	s.frontend.ImageText16(client.xID(uint32(p.Drawable)), gcID, int32(p.X), int32(p.Y), p.Text)
	s.frontend.ComposeWindow(client.xID(uint32(p.Drawable)))
	return nil
}

func (s *x11Server) handleCreateColormap(client *x11Client, req wire.Request, seq uint16) messageEncoder {
	p := req.(*wire.CreateColormapRequest)
	xid := client.xID(uint32(p.Mid))

	if _, exists := s.colormaps[xid]; exists {
		return wire.NewGenericError(seq, uint32(p.Mid), 0, wire.CreateColormap, wire.IDChoiceErrorCode)
	}
	if err := s.checkWindow(client.xID(uint32(p.Window)), seq, wire.CreateColormap, 0); err != nil {
		return err
	}

	newColormap := &colormap{
		pixels: make(map[uint32]wire.XColorItem),
	}

	if p.Alloc == 1 { // All
		// For TrueColor, pre-allocating doesn't make much sense as pixels are calculated.
		// For other visual types, this would be important.
		// For now, we'll just create an empty map.
	}

	s.colormaps[xid] = newColormap
	return nil
}

func (s *x11Server) handleFreeColormap(client *x11Client, req wire.Request, seq uint16) messageEncoder {
	p := req.(*wire.FreeColormapRequest)
	xid := client.xID(uint32(p.Cmap))
	if err := s.checkColormap(xid, seq, wire.FreeColormap, 0); err != nil {
		return err
	}
	delete(s.colormaps, xid)
	return nil
}

func (s *x11Server) handleCopyColormapAndFree(client *x11Client, req wire.Request, seq uint16) messageEncoder {
	p := req.(*wire.CopyColormapAndFreeRequest)
	srcCmapID := client.xID(uint32(p.SrcCmap))
	if err := s.checkColormap(srcCmapID, seq, wire.CopyColormapAndFree, 0); err != nil {
		return err
	}
	srcCmap := s.colormaps[srcCmapID]

	newCmapID := client.xID(uint32(p.Mid))
	if _, exists := s.colormaps[newCmapID]; exists {
		return wire.NewGenericError(seq, uint32(p.Mid), 0, wire.CopyColormapAndFree, wire.IDChoiceErrorCode)
	}

	newCmap := &colormap{
		pixels: make(map[uint32]wire.XColorItem),
	}
	s.colormaps[newCmapID] = newCmap

	for pixel, color := range srcCmap.pixels {
		if color.ClientID == client.id {
			newCmap.pixels[pixel] = color
			delete(srcCmap.pixels, pixel)
		}
	}
	return nil
}

func (s *x11Server) handleInstallColormap(client *x11Client, req wire.Request, seq uint16) messageEncoder {
	p := req.(*wire.InstallColormapRequest)
	xid := client.xID(uint32(p.Cmap))
	if err := s.checkColormap(xid, seq, wire.InstallColormap, 0); err != nil {
		return err
	}

	s.installedColormap = xid

	for winID, win := range s.windows {
		if win.colormap == xid {
			client, ok := s.clients[winID.client]
			if !ok {
				debugf("X11: InstallColormap unknown client %d", winID.client)
				continue
			}
			event := &wire.ColormapNotifyEvent{
				Sequence: client.sequence - 1,
				Window:   winID.local,
				Colormap: uint32(p.Cmap),
				New:      true,
				State:    0, // Installed
			}
			s.sendEvent(client, event)
		}
	}
	return nil
}

func (s *x11Server) handleUninstallColormap(client *x11Client, req wire.Request, seq uint16) messageEncoder {
	p := req.(*wire.UninstallColormapRequest)
	xid := client.xID(uint32(p.Cmap))
	if err := s.checkColormap(xid, seq, wire.UninstallColormap, 0); err != nil {
		return err
	}

	if s.installedColormap == xid {
		s.installedColormap = xID{local: s.defaultColormap}
	}

	for winID, win := range s.windows {
		if win.colormap == xid {
			client, ok := s.clients[winID.client]
			if !ok {
				debugf("X11: UninstallColormap unknown client %d", winID.client)
				continue
			}
			event := &wire.ColormapNotifyEvent{
				Sequence: client.sequence - 1,
				Window:   winID.local,
				Colormap: uint32(p.Cmap),
				New:      false,
				State:    1, // Uninstalled
			}
			s.sendEvent(client, event)
		}
	}
	return nil
}

func (s *x11Server) handleListInstalledColormaps(client *x11Client, req wire.Request, seq uint16) messageEncoder {
	p := req.(*wire.ListInstalledColormapsRequest)
	if err := s.checkWindow(client.xID(uint32(p.Window)), seq, wire.ListInstalledColormaps, 0); err != nil {
		return err
	}
	var colormaps []uint32
	if s.installedColormap.local != 0 {
		colormaps = append(colormaps, s.installedColormap.local)
	}

	return &wire.ListInstalledColormapsReply{
		Sequence:     seq,
		NumColormaps: uint16(len(colormaps)),
		Colormaps:    colormaps,
	}
}

func (s *x11Server) handleAllocColor(client *x11Client, req wire.Request, seq uint16) messageEncoder {
	p := req.(*wire.AllocColorRequest)
	xid := client.xID(uint32(p.Cmap))
	if xid.local == s.defaultColormap {
		xid.client = 0
	}
	cm, ok := s.colormaps[xid]
	if !ok {
		return wire.NewGenericError(seq, uint32(p.Cmap), 0, wire.AllocColor, wire.ColormapErrorCode)
	}

	// Simple allocation for TrueColor: construct pixel value from RGB
	r8 := byte(p.Red >> 8)
	g8 := byte(p.Green >> 8)
	b8 := byte(p.Blue >> 8)
	pixel := (uint32(r8) << 16) | (uint32(g8) << 8) | uint32(b8)

	cm.pixels[pixel] = wire.XColorItem{Red: p.Red, Green: p.Green, Blue: p.Blue, ClientID: client.id}

	return &wire.AllocColorReply{
		Sequence: seq,
		Red:      p.Red,
		Green:    p.Green,
		Blue:     p.Blue,
		Pixel:    pixel,
	}
}

func (s *x11Server) handleAllocNamedColor(client *x11Client, req wire.Request, seq uint16) messageEncoder {
	p := req.(*wire.AllocNamedColorRequest)
	cmap := client.xID(uint32(p.Cmap))
	if cmap.local == s.defaultColormap {
		cmap.client = 0
	}
	cm, ok := s.colormaps[cmap]
	if !ok {
		return wire.NewError(wire.ColormapErrorCode, seq, uint32(p.Cmap), wire.Opcodes{Major: p.OpCode(), Minor: 0})
	}

	name := string(p.Name)
	rgb, ok := lookupColor(name)
	if !ok {
		return wire.NewError(wire.NameErrorCode, seq, 0, wire.Opcodes{Major: p.OpCode(), Minor: 0})
	}

	exactRed := scale8to16(rgb.Red)
	exactGreen := scale8to16(rgb.Green)
	exactBlue := scale8to16(rgb.Blue)

	pixel := (uint32(rgb.Red) << 16) | (uint32(rgb.Green) << 8) | uint32(rgb.Blue)
	cm.pixels[pixel] = wire.XColorItem{Red: exactRed, Green: exactGreen, Blue: exactBlue, ClientID: client.id}

	return &wire.AllocNamedColorReply{
		Sequence:   seq,
		Pixel:      pixel,
		ExactRed:   exactRed,
		ExactGreen: exactGreen,
		ExactBlue:  exactBlue,
		Red:        exactRed,
		Green:      exactGreen,
		Blue:       exactBlue,
	}
}

func (s *x11Server) handleAllocColorCells(client *x11Client, req wire.Request, seq uint16) messageEncoder {
	return wire.NewGenericError(seq, 0, 0, wire.AllocColorCells, wire.MatchErrorCode)
}

func (s *x11Server) handleAllocColorPlanes(client *x11Client, req wire.Request, seq uint16) messageEncoder {
	return wire.NewGenericError(seq, 0, 0, wire.AllocColorPlanes, wire.MatchErrorCode)
}

func (s *x11Server) handleFreeColors(client *x11Client, req wire.Request, seq uint16) messageEncoder {
	p := req.(*wire.FreeColorsRequest)
	xid := client.xID(uint32(p.Cmap))
	if xid.local == s.defaultColormap {
		xid.client = 0
	}
	cm, ok := s.colormaps[xid]
	if !ok {
		return wire.NewGenericError(seq, uint32(p.Cmap), 0, wire.FreeColors, wire.ColormapErrorCode)
	}

	for _, pixel := range p.Pixels {
		delete(cm.pixels, pixel)
	}
	return nil
}

func (s *x11Server) handleStoreColors(client *x11Client, req wire.Request, seq uint16) messageEncoder {
	p := req.(*wire.StoreColorsRequest)
	xid := client.xID(uint32(p.Cmap))
	if xid.local == s.defaultColormap {
		xid.client = 0
	}
	cm, ok := s.colormaps[xid]
	if !ok {
		return wire.NewGenericError(seq, uint32(p.Cmap), 0, wire.StoreColors, wire.ColormapErrorCode)
	}

	for _, item := range p.Items {
		c, exists := cm.pixels[item.Pixel]
		if !exists {
			c = wire.XColorItem{}
		}

		if item.Flags&wire.DoRed != 0 {
			c.Red = item.Red
		}
		if item.Flags&wire.DoGreen != 0 {
			c.Green = item.Green
		}
		if item.Flags&wire.DoBlue != 0 {
			c.Blue = item.Blue
		}
		cm.pixels[item.Pixel] = c
	}
	return nil
}

func (s *x11Server) handleStoreNamedColor(client *x11Client, req wire.Request, seq uint16) messageEncoder {
	p := req.(*wire.StoreNamedColorRequest)
	xid := client.xID(uint32(p.Cmap))
	if xid.local == s.defaultColormap {
		xid.client = 0
	}
	cm, ok := s.colormaps[xid]
	if !ok {
		return wire.NewGenericError(seq, uint32(p.Cmap), 0, wire.StoreNamedColor, wire.ColormapErrorCode)
	}

	rgb, ok := lookupColor(p.Name)
	if !ok {
		return wire.NewGenericError(seq, 0, 0, wire.StoreNamedColor, wire.NameErrorCode)
	}

	c, exists := cm.pixels[p.Pixel]
	if !exists {
		c = wire.XColorItem{}
	}

	if p.Flags&wire.DoRed != 0 {
		c.Red = scale8to16(rgb.Red)
	}
	if p.Flags&wire.DoGreen != 0 {
		c.Green = scale8to16(rgb.Green)
	}
	if p.Flags&wire.DoBlue != 0 {
		c.Blue = scale8to16(rgb.Blue)
	}
	cm.pixels[p.Pixel] = c
	return nil
}

func (s *x11Server) handleQueryColors(client *x11Client, req wire.Request, seq uint16) messageEncoder {
	p := req.(*wire.QueryColorsRequest)
	cmap := client.xID(p.Cmap)
	if cmap.local == s.defaultColormap {
		cmap.client = 0
	}
	cm, ok := s.colormaps[cmap]
	if !ok {
		return wire.NewGenericError(seq, uint32(p.Cmap), 0, wire.QueryColors, wire.ColormapErrorCode)
	}

	colors := make([]wire.XColorItem, len(p.Pixels))
	for i, pixel := range p.Pixels {
		color, ok := cm.pixels[pixel]
		if !ok {
			// If the pixel is not in the colormap, return black
			color = wire.XColorItem{Red: 0, Green: 0, Blue: 0}
		}
		colors[i] = color
	}

	return &wire.QueryColorsReply{
		Sequence: seq,
		Colors:   colors,
	}
}

func (s *x11Server) handleLookupColor(client *x11Client, req wire.Request, seq uint16) messageEncoder {
	p := req.(*wire.LookupColorRequest)
	// cmapID := client.xID(uint32(p.Cmap))

	color, ok := lookupColor(p.Name)
	if !ok {
		return wire.NewGenericError(seq, uint32(p.Cmap), 0, wire.LookupColor, wire.NameErrorCode)
	}

	return &wire.LookupColorReply{
		Sequence:   seq,
		Red:        scale8to16(color.Red),
		Green:      scale8to16(color.Green),
		Blue:       scale8to16(color.Blue),
		ExactRed:   scale8to16(color.Red),
		ExactGreen: scale8to16(color.Green),
		ExactBlue:  scale8to16(color.Blue),
	}
}

func (s *x11Server) handleCreateCursor(client *x11Client, req wire.Request, seq uint16) messageEncoder {
	p := req.(*wire.CreateCursorRequest)
	cursorXID := client.xID(uint32(p.Cid))
	if _, exists := s.cursors[cursorXID]; exists {
		s.logger.Errorf("X11: CreateCursor: ID %s already in use", cursorXID)
		return wire.NewGenericError(seq, uint32(p.Cid), 0, wire.CreateCursor, wire.IDChoiceErrorCode)
	}

	sourceXID := client.xID(uint32(p.Source))
	if err := s.checkPixmap(sourceXID, seq, wire.CreateCursor, 0); err != nil {
		return err
	}
	maskXID := client.xID(uint32(p.Mask))
	if p.Mask != 0 {
		if err := s.checkPixmap(maskXID, seq, wire.CreateCursor, 0); err != nil {
			return err
		}
	}

	s.cursors[cursorXID] = true
	foreColor := [3]uint16{p.ForeRed, p.ForeGreen, p.ForeBlue}
	backColor := [3]uint16{p.BackRed, p.BackGreen, p.BackBlue}
	s.frontend.CreateCursor(cursorXID, sourceXID, maskXID, foreColor, backColor, p.X, p.Y)
	return nil
}

func (s *x11Server) handleCreateGlyphCursor(client *x11Client, req wire.Request, seq uint16) messageEncoder {
	p := req.(*wire.CreateGlyphCursorRequest)
	// Check if the cursor ID is already in use
	if _, exists := s.cursors[client.xID(uint32(p.Cid))]; exists {
		s.logger.Errorf("X11: CreateGlyphCursor: ID %d already in use", p.Cid)
		return wire.NewGenericError(seq, uint32(p.Cid), 0, wire.CreateGlyphCursor, wire.IDChoiceErrorCode)
	}
	if err := s.checkFont(client.xID(uint32(p.SourceFont)), seq, wire.CreateGlyphCursor, 0); err != nil {
		return err
	}
	if p.MaskFont != 0 {
		if err := s.checkFont(client.xID(uint32(p.MaskFont)), seq, wire.CreateGlyphCursor, 0); err != nil {
			return err
		}
	}

	s.cursors[client.xID(uint32(p.Cid))] = true
	s.frontend.CreateCursorFromGlyph(uint32(p.Cid), p.SourceChar)
	return nil
}

func (s *x11Server) handleFreeCursor(client *x11Client, req wire.Request, seq uint16) messageEncoder {
	p := req.(*wire.FreeCursorRequest)
	xid := client.xID(uint32(p.Cursor))
	if err := s.checkCursor(xid, seq, wire.FreeCursor, 0); err != nil {
		return err
	}
	delete(s.cursors, xid)
	s.frontend.FreeCursor(xid)
	return nil
}

func (s *x11Server) handleRecolorCursor(client *x11Client, req wire.Request, seq uint16) messageEncoder {
	p := req.(*wire.RecolorCursorRequest)
	if err := s.checkCursor(client.xID(uint32(p.Cursor)), seq, wire.RecolorCursor, 0); err != nil {
		return err
	}
	s.frontend.RecolorCursor(client.xID(uint32(p.Cursor)), p.ForeColor, p.BackColor)
	return nil
}

func (s *x11Server) handleQueryBestSize(client *x11Client, req wire.Request, seq uint16) messageEncoder {
	p := req.(*wire.QueryBestSizeRequest)
	if err := s.checkDrawable(client.xID(uint32(p.Drawable)), seq, wire.QueryBestSize, 0); err != nil {
		return err
	}
	debugf("X11: QueryBestSize class=%d drawable=%d width=%d height=%d", p.Class, p.Drawable, p.Width, p.Height)

	return &wire.QueryBestSizeReply{
		Sequence: seq,
		Width:    p.Width,
		Height:   p.Height,
	}
}

func (s *x11Server) handleQueryExtension(client *x11Client, req wire.Request, seq uint16) messageEncoder {
	p := req.(*wire.QueryExtensionRequest)
	debugf("X11: QueryExtension name=%s", p.Name)

	switch p.Name {
	case wire.BigRequestsExtensionName:
		return &wire.QueryExtensionReply{
			Sequence:    seq,
			Present:     true,
			MajorOpcode: byte(wire.BigRequestsOpcode),
			FirstEvent:  0,
			FirstError:  0,
		}
	case wire.XInputExtensionName:
		return &wire.QueryExtensionReply{
			Sequence:    seq,
			Present:     true,
			MajorOpcode: byte(wire.XInputOpcode),
			FirstEvent:  0,
			FirstError:  0,
		}
	default:
		return &wire.QueryExtensionReply{
			Sequence:    seq,
			Present:     false,
			MajorOpcode: 0,
			FirstEvent:  0,
			FirstError:  0,
		}
	}
}

func (s *x11Server) handleListExtensions(client *x11Client, req wire.Request, seq uint16) messageEncoder {
	extensions := []string{
		wire.BigRequestsExtensionName,
		wire.XInputExtensionName,
	}
	return &wire.ListExtensionsReply{
		Sequence: seq,
		NNames:   byte(len(extensions)),
		Names:    extensions,
	}
}

func (s *x11Server) handleChangeKeyboardMapping(client *x11Client, req wire.Request, seq uint16) messageEncoder {
	p := req.(*wire.ChangeKeyboardMappingRequest)
	keySymIndex := 0
	for i := 0; i < int(p.KeyCodeCount); i++ {
		keyCode := p.FirstKeyCode + wire.KeyCode(i)
		for j := 0; j < int(p.KeySymsPerKeyCode); j++ {
			if j == 0 {
				s.keymap[byte(keyCode)] = p.KeySyms[keySymIndex]
			}
			keySymIndex++
		}
	}
	return nil
}

func (s *x11Server) handleGetKeyboardMapping(client *x11Client, req wire.Request, seq uint16) messageEncoder {
	p := req.(*wire.GetKeyboardMappingRequest)
	keySyms := make([]uint32, 0, p.Count)
	for i := 0; i < int(p.Count); i++ {
		keyCode := p.FirstKeyCode + wire.KeyCode(i)
		keySym, ok := s.keymap[byte(keyCode)]
		if !ok {
			keySym = 0 // NoSymbol
		}
		keySyms = append(keySyms, keySym)
	}
	return &wire.GetKeyboardMappingReply{
		Sequence: seq,
		KeySyms:  keySyms,
	}
}

func (s *x11Server) handleChangeKeyboardControl(client *x11Client, req wire.Request, seq uint16) messageEncoder {
	p := req.(*wire.ChangeKeyboardControlRequest)
	s.frontend.ChangeKeyboardControl(p.ValueMask, p.Values)
	return nil
}

func (s *x11Server) handleGetKeyboardControl(client *x11Client, req wire.Request, seq uint16) messageEncoder {
	kc, _ := s.frontend.GetKeyboardControl()
	return &wire.GetKeyboardControlReply{
		Sequence:         seq,
		KeyClickPercent:  byte(kc.KeyClickPercent),
		BellPercent:      byte(kc.BellPercent),
		BellPitch:        uint16(kc.BellPitch),
		BellDuration:     uint16(kc.BellDuration),
		LedMask:          uint32(kc.Led),
		GlobalAutoRepeat: byte(kc.AutoRepeatMode),
		AutoRepeats:      [32]byte{},
	}
}

func (s *x11Server) handleBell(client *x11Client, req wire.Request, seq uint16) messageEncoder {
	p := req.(*wire.BellRequest)
	s.frontend.Bell(p.Percent)
	return nil
}

func (s *x11Server) handleChangePointerControl(client *x11Client, req wire.Request, seq uint16) messageEncoder {
	p := req.(*wire.ChangePointerControlRequest)
	s.frontend.ChangePointerControl(p.AccelerationNumerator, p.AccelerationDenominator, p.Threshold, p.DoAcceleration, p.DoThreshold)
	return nil
}

func (s *x11Server) handleGetPointerControl(client *x11Client, req wire.Request, seq uint16) messageEncoder {
	accelNumerator, accelDenominator, threshold, _ := s.frontend.GetPointerControl()
	return &wire.GetPointerControlReply{
		Sequence:         seq,
		AccelNumerator:   accelNumerator,
		AccelDenominator: accelDenominator,
		Threshold:        threshold,
	}
}

func (s *x11Server) handleSetScreenSaver(client *x11Client, req wire.Request, seq uint16) messageEncoder {
	p := req.(*wire.SetScreenSaverRequest)
	s.frontend.SetScreenSaver(p.Timeout, p.Interval, p.PreferBlank, p.AllowExpose)
	return nil
}

func (s *x11Server) handleGetScreenSaver(client *x11Client, req wire.Request, seq uint16) messageEncoder {
	timeout, interval, preferBlank, allowExpose, _ := s.frontend.GetScreenSaver()
	return &wire.GetScreenSaverReply{
		Sequence:    seq,
		Timeout:     uint16(timeout),
		Interval:    uint16(interval),
		PreferBlank: preferBlank,
		AllowExpose: allowExpose,
	}
}

func (s *x11Server) handleChangeHosts(client *x11Client, req wire.Request, seq uint16) messageEncoder {
	p := req.(*wire.ChangeHostsRequest)
	s.frontend.ChangeHosts(p.Mode, p.Host)
	return nil
}

func (s *x11Server) handleListHosts(client *x11Client, req wire.Request, seq uint16) messageEncoder {
	hosts, _ := s.frontend.ListHosts()
	return &wire.ListHostsReply{
		Sequence: seq,
		NumHosts: uint16(len(hosts)),
		Hosts:    hosts,
	}
}

func (s *x11Server) handleSetAccessControl(client *x11Client, req wire.Request, seq uint16) messageEncoder {
	p := req.(*wire.SetAccessControlRequest)
	s.frontend.SetAccessControl(p.Mode)
	return nil
}

func (s *x11Server) handleSetCloseDownMode(client *x11Client, req wire.Request, seq uint16) messageEncoder {
	p := req.(*wire.SetCloseDownModeRequest)
	s.frontend.SetCloseDownMode(p.Mode)
	return nil
}

func (s *x11Server) handleKillClient(client *x11Client, req wire.Request, seq uint16) messageEncoder {
	p := req.(*wire.KillClientRequest)
	s.frontend.KillClient(p.Resource)
	return nil
}

func (s *x11Server) handleRotateProperties(client *x11Client, req wire.Request, seq uint16) messageEncoder {
	p := req.(*wire.RotatePropertiesRequest)
	err := s.RotateProperties(client.xID(uint32(p.Window)), p.Delta, p.Atoms)
	if err != nil {
		return wire.NewGenericError(seq, uint32(p.Window), 0, wire.RotateProperties, wire.MatchErrorCode)
	}
	return nil
}

func (s *x11Server) handleForceScreenSaver(client *x11Client, req wire.Request, seq uint16) messageEncoder {
	p := req.(*wire.ForceScreenSaverRequest)
	s.frontend.ForceScreenSaver(p.Mode)
	return nil
}

func (s *x11Server) handleSetPointerMapping(client *x11Client, req wire.Request, seq uint16) messageEncoder {
	p := req.(*wire.SetPointerMappingRequest)
	status, _ := s.frontend.SetPointerMapping(p.Map)
	return &wire.SetPointerMappingReply{
		Sequence: seq,
		Status:   status,
	}
}

func (s *x11Server) handleGetPointerMapping(client *x11Client, req wire.Request, seq uint16) messageEncoder {
	pMap, _ := s.frontend.GetPointerMapping()
	return &wire.GetPointerMappingReply{
		Sequence: seq,
		Length:   byte(len(pMap)),
		PMap:     pMap,
	}
}

func (s *x11Server) handleSetModifierMapping(client *x11Client, req wire.Request, seq uint16) messageEncoder {
	p := req.(*wire.SetModifierMappingRequest)
	status, _ := s.frontend.SetModifierMapping(p.KeyCodesPerModifier, p.KeyCodes)
	return &wire.SetModifierMappingReply{
		Sequence: seq,
		Status:   status,
	}
}

func (s *x11Server) handleGetModifierMapping(client *x11Client, req wire.Request, seq uint16) messageEncoder {
	keyCodes, err := s.frontend.GetModifierMapping()
	if err != nil {
		return wire.NewGenericError(seq, 0, 0, wire.GetModifierMapping, wire.ImplementationErrorCode)
	}
	return &wire.GetModifierMappingReply{
		Sequence:            seq,
		KeyCodesPerModifier: byte(len(keyCodes) / 8),
		KeyCodes:            keyCodes,
	}
}

func (s *x11Server) handleNoOperation(client *x11Client, req wire.Request, seq uint16) messageEncoder {
	return nil
}

func (s *x11Server) handleEnableBigRequests(client *x11Client, req wire.Request, seq uint16) messageEncoder {
	client.bigRequestsEnabled = true
	return &wire.BigRequestsEnableReply{
		Sequence:         seq,
		MaxRequestLength: 0x100000,
	}
}
