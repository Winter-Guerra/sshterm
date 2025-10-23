package main

import (
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"time"

	"golang.org/x/crypto/ssh"
)

// X11Operation represents a single X11 drawing operation for testing purposes.
type X11Operation struct {
	Type  string
	Color uint32
	Args  []any
}

var x11Operations []X11Operation

func clearX11Operations() {
	x11Operations = nil
}

func (s *sshServer) simulateX11Application(serverConn *ssh.ServerConn) {
	s.t.Log("Simulating X11 application (client-side)")

	windowWidth, windowHeight := 600, 400

	// Open X11 channel back to the SSH client (WASM App)
	x11Channel, x11Requests, err := serverConn.OpenChannel("x11", nil)
	if err != nil {
		s.t.Logf("Failed to open X11 channel: %v", err)
		return
	}
	defer x11Channel.Close()
	go ssh.DiscardRequests(x11Requests)

	// 1. Send SetupRequest
	s.t.Log("Sending X11 SetupRequest")
	setupRequest := make([]byte, 12)
	setupRequest[0] = 'l'                                // LittleEndian
	binary.LittleEndian.PutUint16(setupRequest[2:4], 11) // Protocol Major Version
	binary.LittleEndian.PutUint16(setupRequest[4:6], 0)  // Protocol Minor Version

	_, err = x11Channel.Write(setupRequest)
	if err != nil {
		s.t.Logf("Failed to send X11 SetupRequest: %v", err)
		return
	}
	s.t.Log("X11 SetupRequest sent successfully")

	// 2. Receive SetupResponse
	s.t.Log("Waiting for X11 SetupResponse")
	setupResponseHeader := make([]byte, 8)
	_, err = io.ReadFull(x11Channel, setupResponseHeader)
	if err != nil {
		s.t.Logf("Failed to read X11 SetupResponse header: %v", err)
		return
	}
	s.t.Log("X11 SetupResponse header received")

	status := setupResponseHeader[0]
	protocolMajor := binary.LittleEndian.Uint16(setupResponseHeader[2:4])
	protocolMinor := binary.LittleEndian.Uint16(setupResponseHeader[4:6])
	additionalDataLength := binary.LittleEndian.Uint16(setupResponseHeader[6:8]) * 4 // in bytes

	s.t.Logf("X11 SetupResponse: status=%d, protocolMajor=%d, protocolMinor=%d, additionalDataLength=%d", status, protocolMajor, protocolMinor, additionalDataLength)

	if status != 1 { // 1 means success
		s.t.Logf("X11 SetupResponse indicates failure (status %d)", status)
		return
	}

	if additionalDataLength > 0 {
		remainingData := make([]byte, additionalDataLength)
		_, err := io.ReadFull(x11Channel, remainingData)
		if err != nil {
			s.t.Logf("Failed to read X11 SetupResponse remaining data: %v", err)
			return
		}
	}

	s.t.Logf("X11 client handshake successful")
	replyChan := s.readReplies(x11Channel)

	// Now send drawing commands
	s.t.Log("Sending drawing commands...")

	// Create Window
	s.t.Log("Sending CreateWindow")
	if err := s.createWindow(x11Channel, 1, 0, 10, 20, uint32(windowWidth), uint32(windowHeight)); err != nil {
		s.t.Logf("Failed to create window: %v", err)
		return
	}

	// MapWindow
	s.t.Log("Sending MapWindow")
	if err := s.mapWindow(x11Channel, 1); err != nil {
		s.t.Logf("Failed to map window: %v", err)
		return
	}

	// Fill window with white background
	s.t.Log("Filling window with white background")
	if err := s.createGCWithBackground(x11Channel, 99, 1, 0xFFFFFF, 0); err != nil {
		s.t.Logf("Failed to create white GC: %v", err)
		return
	}
	background := []int16{0, 0, int16(windowWidth), int16(windowHeight)}
	if err := s.polyFillRectangle(x11Channel, 1, 99, background); err != nil {
		s.t.Logf("Failed to fill window background: %v", err)
		return
	}

	// Create GCs
	colors := map[string]uint32{
		"red":    0xFF0000,
		"green":  0x008000, // Darker Green
		"blue":   0x0000FF,
		"yellow": 0xFFFF00,
		"brown":  0x8B4513,
		"cyan":   0x00FFFF,
	}
	gcs := make(map[string]uint32)
	i := uint32(100)
	for name, color := range colors {
		gcs[name] = i
		if err := s.createGCWithBackground(x11Channel, i, 1, color, 0); err != nil {
			s.t.Logf("Failed to create %s GC: %v", name, err)
			return
		}
		i++
	}

	// Draw ground
	s.t.Log("Drawing ground")
	ground := []int16{0, 300, 600, 100}
	if err := s.polyFillRectangle(x11Channel, 1, gcs["green"], ground); err != nil {
		s.t.Logf("Failed to draw ground: %v", err)
		return
	}

	// Draw house base
	s.t.Log("Drawing house base")
	houseBase := []int16{200, 200, 200, 150}
	if err := s.polyFillRectangle(x11Channel, 1, gcs["brown"], houseBase); err != nil {
		s.t.Logf("Failed to draw house base: %v", err)
		return
	}

	// Draw roof
	s.t.Log("Drawing roof")
	roof := []int16{180, 200, 300, 100, 420, 200}
	if err := s.fillPoly(x11Channel, 1, gcs["red"], 0, roof); err != nil {
		s.t.Logf("Failed to draw roof: %v", err)
		return
	}

	// Draw sun
	s.t.Log("Drawing sun")
	sun := []int16{500, 50, 40, 40, 0, 360 * 64}
	if err := s.polyFillArc(x11Channel, 1, gcs["yellow"], sun); err != nil {
		s.t.Logf("Failed to draw sun: %v", err)
		return
	}

	// Draw a star
	s.t.Log("Drawing cyan star")
	starPoints := []int16{
		100, 50, 110, 75, 135, 75, 115, 95, 125, 120,
		100, 105, 75, 120, 85, 95, 65, 75, 90, 75, 100, 50,
	}
	if err := s.polyLine(x11Channel, 1, gcs["cyan"], 0, starPoints); err != nil {
		s.t.Logf("Failed to draw cyan star: %v", err)
		return
	}

	// ChangeProperty (set window title)
	s.t.Log("Sending ChangeProperty")
	title := "SSHTERM X11 - House and Sun"
	if err := s.changeProperty(x11Channel, 1, 35, 36, 0, 8, []byte(title)); err != nil {
		s.t.Logf("Failed to change property: %v", err)
		return
	}

	if err := s.imageText8(x11Channel, 1, gcs["blue"], 50, 50, []byte("Hello X11!")); err != nil {
		s.t.Logf("Failed to draw text: %v", err)
		return
	}

	if err := s.imageText16(x11Channel, 1, gcs["red"], 50, 70, []uint16{0x0048, 0x0065, 0x006c, 0x006c, 0x006f, 0x0020, 0x0057, 0x006f, 0x0072, 0x006c, 0x0064, 0x0021}); err != nil {
		s.t.Logf("Failed to draw ImageText16: %v", err)
		return
	}

	polyText8Items := []PolyText8Item{
		{Delta: 0, Str: []byte("PolyText8 ")},
		{Delta: 10, Str: []byte("Example")},
	}
	if err := s.polyText8(x11Channel, 1, gcs["yellow"], 50, 90, polyText8Items); err != nil {
		s.t.Logf("Failed to draw PolyText8: %v", err)
		return
	}

	polyText16Items := []PolyText16Item{
		{Delta: 0, Str: []uint16{0x0050, 0x006f, 0x006c, 0x0079, 0x0054, 0x0065, 0x0078, 0x0074, 0x0031, 0x0036, 0x0020}},
		{Delta: 10, Str: []uint16{0x0045, 0x0078, 0x0061, 0x006d, 0x0070, 0x006c, 0x0065}},
	}
	if err := s.polyText16(x11Channel, 1, gcs["cyan"], 50, 110, polyText16Items); err != nil {
		s.t.Logf("Failed to draw PolyText16: %v", err)
		return
	}

	// Test Font API
	s.t.Log("Testing Font API")
	fontID := uint32(200)
	fontName := "-*-helvetica-medium-r-normal-*-12-*-*-*-p-*-iso8859-1"
	if err := s.openFont(x11Channel, fontID, fontName); err != nil {
		s.t.Logf("Failed to open font: %v", err)
		return
	}

	if err := s.queryFont(x11Channel, fontID, replyChan); err != nil {
		s.t.Logf("Failed to query font: %v", err)
		return
	}

	if err := s.listFonts(x11Channel, 10, "*", replyChan); err != nil {
		s.t.Logf("Failed to list fonts: %v", err)
		return
	}

	// Create a new GC with the font
	fontGC := uint32(106)
	if err := s.createGCWithFont(x11Channel, fontGC, 1, colors["blue"], 0, fontID); err != nil {
		s.t.Logf("Failed to create GC with font: %v", err)
		return
	}

	// Draw text with the new GC
	if err := s.imageText8(x11Channel, 1, fontGC, 50, 130, []byte("Text with font!")); err != nil {
		s.t.Logf("Failed to draw text with font: %v", err)
		return
	}

	if err := s.closeFont(x11Channel, fontID); err != nil {
		s.t.Logf("Failed to close font: %v", err)
		return
	}

	s.simulateXEyes(x11Channel)

	s.t.Log("All drawing commands sent successfully")
	time.Sleep(2 * time.Second)
	x11Channel.Close()
}

func (s *sshServer) readReplies(channel ssh.Channel) <-chan []byte {
	ch := make(chan []byte, 1)
	go func() {
		for {
			// Read Reply or Event
			replyHeader := make([]byte, 32)
			_, err := io.ReadFull(channel, replyHeader)
			if err != nil {
				s.t.Logf("failed to read X11 message header: %v", err)
				return
			}

			msgType := replyHeader[0]
			sequenceNumber := binary.LittleEndian.Uint16(replyHeader[2:4])

			s.t.Logf("Received X11 message: type=%d, sequence=%d", msgType, sequenceNumber)

			if msgType == 1 { // Reply
				replyLength := 4 * binary.LittleEndian.Uint32(replyHeader[4:8])
				if replyLength > 4096 {
					s.t.Errorf("Received X11 reply with length %d > 4096", replyLength)
					return
				}
				replyRemaining := make([]byte, replyLength)
				_, err = io.ReadFull(channel, replyRemaining)
				if err != nil {
					s.t.Errorf("failed to read reply remaining: %v", err)
					return
				}
				fullReply := append(replyHeader, replyRemaining...)
				ch <- fullReply

			} else if msgType == 0 { // Error
				// Error messages are 32 bytes long
				s.t.Logf("Received X11 error: code=%d, sequence=%d", replyHeader[1], sequenceNumber)
			} else if msgType >= 2 && msgType <= 127 { // Event
				// Events are 32 bytes long, so no additional data to read after the header
				s.t.Logf("Received X11 event: type=%d, sequence=%d. Discarding.", msgType, sequenceNumber)
			} else {
				s.t.Errorf("ERR received unknown X11 message type: %d", msgType)
			}
		}
	}()
	return ch
}

func (s *sshServer) simulateXEyes(x11Channel ssh.Channel) {
	s.t.Log("Simulating xeyes")
	// Create Window
	if err := s.createWindow(x11Channel, 10, 64, 0, 0, 150, 100); err != nil {
		s.t.Logf("Failed to create window: %v", err)
		return
	}
	if err := s.createWindow(x11Channel, 11, 10, 0, 0, 150, 100); err != nil {
		s.t.Logf("Failed to create window: %v", err)
		return
	}

	// Create GCs
	whiteGC := uint32(13)
	if err := s.createGCWithBackground(x11Channel, whiteGC, 11, 0xFFFFFF, 0); err != nil {
		s.t.Logf("Failed to create white GC: %v", err)
		return
	}
	blackGC := uint32(14)
	if err := s.createGCWithBackground(x11Channel, blackGC, 11, 0x000000, 0); err != nil {
		s.t.Logf("Failed to create black GC: %v", err)
		return
	}

	// MapWindow
	if err := s.mapWindow(x11Channel, 11); err != nil {
		s.t.Logf("Failed to map window: %v", err)
		return
	}
	if err := s.mapWindow(x11Channel, 10); err != nil {
		s.t.Logf("Failed to map window: %v", err)
		return
	}

	// Draw eyes
	leftEyeOutline := []int16{17, 25, 13, 18, 5760, 23040}
	if err := s.polyFillArc(x11Channel, 11, blackGC, leftEyeOutline); err != nil {
		s.t.Logf("Failed to draw left eye outline: %v", err)
		return
	}
	rightEyeOutline := []int16{92, 34, 13, 18, 5760, 23040}
	if err := s.polyFillArc(x11Channel, 11, blackGC, rightEyeOutline); err != nil {
		s.t.Logf("Failed to draw right eye outline: %v", err)
		return
	}

	leftEye := []int16{18, 26, 11, 16, 5760, 23040}
	if err := s.polyFillArc(x11Channel, 11, whiteGC, leftEye); err != nil {
		s.t.Logf("Failed to draw left eye: %v", err)
		return
	}
	rightEye := []int16{93, 35, 11, 16, 5760, 23040}
	if err := s.polyFillArc(x11Channel, 11, whiteGC, rightEye); err != nil {
		s.t.Logf("Failed to draw right eye: %v", err)
		return
	}

	// Draw pupils
	leftPupil := []int16{23, 31, 5, 8, 5760, 23040}
	if err := s.polyFillArc(x11Channel, 11, blackGC, leftPupil); err != nil {
		s.t.Logf("Failed to draw left pupil: %v", err)
		return
	}
	rightPupil := []int16{98, 40, 5, 8, 5760, 23040}
	if err := s.polyFillArc(x11Channel, 11, blackGC, rightPupil); err != nil {
		s.t.Logf("Failed to draw right pupil: %v", err)
		return
	}
}

func GetX11Operations() []X11Operation {
	log.Printf("X11: GetX11Operations returning %d operations", len(x11Operations))
	return x11Operations
}

func (s *sshServer) changeProperty(channel ssh.Channel, wid, property, typeAtom uint32, mode, format byte, data []byte) error {
	newOp := X11Operation{
		Type: "changeProperty",
		Args: []any{wid, property, typeAtom, uint32(format), string(data)},
	}
	x11Operations = append(x11Operations, newOp)
	// Opcode: 18
	// Mode: mode (0=Replace)
	// Request Length: calculated
	// Window ID: wid
	// Property Atom: property
	// Type Atom: typeAtom
	// Format: format (8, 16, 32)
	// Number of units in data: len(data) / (format / 8)
	// Data: data

	dataLen := uint32(len(data))
	if format == 16 {
		dataLen /= 2
	} else if format == 32 {
		dataLen /= 4
	}

	payload := make([]byte, 20) // Fixed part of payload
	binary.LittleEndian.PutUint32(payload[0:4], wid)
	binary.LittleEndian.PutUint32(payload[4:8], property)
	binary.LittleEndian.PutUint32(payload[8:12], typeAtom)
	payload[12] = format
	// payload[13:16] unused
	binary.LittleEndian.PutUint32(payload[16:20], dataLen)

	fullPayload := append(payload, data...)

	_, err := s.writeX11Request(channel, 18, mode, fullPayload)
	if err != nil {
		return err
	}
	return nil
}

func (s *sshServer) putImage(channel ssh.Channel, drawable, gc, x, y, width, height, leftPad, format uint32, imageData []byte) error {
	newOp := X11Operation{
		Type:  "putImage",
		Color: gcColors[gc],
		Args:  []any{drawable, gcToMap(gc), x, y, width, height, leftPad, format, len(imageData)},
	}
	x11Operations = append(x11Operations, newOp)

	// Opcode: 72
	// Format: format
	// Request Length: calculated
	// Drawable ID: drawable
	// GC ID: gc
	// X, Y, Width, Height, Left Pad: x, y, width, height, leftPad
	// Image Data: imageData

	payload := make([]byte, 20) // Fixed part of payload
	binary.LittleEndian.PutUint32(payload[0:4], drawable)
	binary.LittleEndian.PutUint32(payload[4:8], gc)
	binary.LittleEndian.PutUint16(payload[8:10], uint16(x))
	binary.LittleEndian.PutUint16(payload[10:12], uint16(y))
	binary.LittleEndian.PutUint16(payload[12:14], uint16(width))
	binary.LittleEndian.PutUint16(payload[14:16], uint16(height))
	payload[16] = byte(leftPad)
	// payload[17:20] unused

	fullPayload := append(payload, imageData...)

	_, err := s.writeX11Request(channel, 72, byte(format), fullPayload)
	if err != nil {
		return err
	}
	return nil
}

func (s *sshServer) polyFillArc(channel ssh.Channel, drawable, gc uint32, arcs []int16) error {
	newOp := X11Operation{
		Type:  "polyFillArc",
		Color: gcColors[gc],
		Args:  []any{drawable, gcToMap(gc), uint32Slice(arcs)},
	}
	x11Operations = append(x11Operations, newOp)

	// Opcode: 71
	// Request Length: calculated
	// Drawable ID: drawable
	// GC ID: gc
	// Arcs: list of X, Y, Width, Height, Angle1, Angle2 tuples

	payload := make([]byte, 8)
	binary.LittleEndian.PutUint32(payload[0:4], drawable)
	binary.LittleEndian.PutUint32(payload[4:8], gc)

	for i := 0; i < len(arcs); i += 6 {
		arcBytes := make([]byte, 12)
		binary.LittleEndian.PutUint16(arcBytes[0:2], uint16(arcs[i]))
		binary.LittleEndian.PutUint16(arcBytes[2:4], uint16(arcs[i+1]))
		binary.LittleEndian.PutUint16(arcBytes[4:6], uint16(arcs[i+2]))
		binary.LittleEndian.PutUint16(arcBytes[6:8], uint16(arcs[i+3]))
		binary.LittleEndian.PutUint16(arcBytes[8:10], uint16(arcs[i+4]))
		binary.LittleEndian.PutUint16(arcBytes[10:12], uint16(arcs[i+5]))
		payload = append(payload, arcBytes...)
	}

	_, err := s.writeX11Request(channel, 71, 0, payload)
	if err != nil {
		return err
	}
	return nil
}

func gcToMap(gcID uint32) map[string]interface{} {
	return map[string]interface{}{
		"Foreground": gcColors[gcID],
	}
}

func (s *sshServer) polyArc(channel ssh.Channel, drawable, gc uint32, arcs []int16) error {
	newOp := X11Operation{
		Type:  "polyArc",
		Color: gcColors[gc],
		Args:  []any{drawable, gcToMap(gc), uint32Slice(arcs)},
	}
	x11Operations = append(x11Operations, newOp)

	// Opcode: 68
	// Request Length: calculated
	// Drawable ID: drawable
	// GC ID: gc
	// Arcs: list of X, Y, Width, Height, Angle1, Angle2 tuples

	payload := make([]byte, 8)
	binary.LittleEndian.PutUint32(payload[0:4], drawable)
	binary.LittleEndian.PutUint32(payload[4:8], gc)

	for i := 0; i < len(arcs); i += 6 {
		arcBytes := make([]byte, 12)
		binary.LittleEndian.PutUint16(arcBytes[0:2], uint16(arcs[i]))
		binary.LittleEndian.PutUint16(arcBytes[2:4], uint16(arcs[i+1]))
		binary.LittleEndian.PutUint16(arcBytes[4:6], uint16(arcs[i+2]))
		binary.LittleEndian.PutUint16(arcBytes[6:8], uint16(arcs[i+3]))
		binary.LittleEndian.PutUint16(arcBytes[8:10], uint16(arcs[i+4]))
		binary.LittleEndian.PutUint16(arcBytes[10:12], uint16(arcs[i+5]))
		payload = append(payload, arcBytes...)
	}

	_, err := s.writeX11Request(channel, 68, 0, payload)
	if err != nil {
		return err
	}
	return nil
}

func (s *sshServer) polyRectangle(channel ssh.Channel, drawable, gc uint32, rects []int16) error {
	newOp := X11Operation{
		Type:  "polyRectangle",
		Color: gcColors[gc],
		Args:  []any{drawable, gcToMap(gc), uint32Slice(rects)},
	}
	x11Operations = append(x11Operations, newOp)

	// Opcode: 67
	// Request Length: calculated
	// Drawable ID: drawable
	// GC ID: gc
	// Rectangles: list of X, Y, Width, Height tuples

	payload := make([]byte, 8)
	binary.LittleEndian.PutUint32(payload[0:4], drawable)
	binary.LittleEndian.PutUint32(payload[4:8], gc)

	for i := 0; i < len(rects); i += 4 {
		rectBytes := make([]byte, 8)
		binary.LittleEndian.PutUint16(rectBytes[0:2], uint16(rects[i]))
		binary.LittleEndian.PutUint16(rectBytes[2:4], uint16(rects[i+1]))
		binary.LittleEndian.PutUint16(rectBytes[4:6], uint16(rects[i+2]))
		binary.LittleEndian.PutUint16(rectBytes[6:8], uint16(rects[i+3]))
		payload = append(payload, rectBytes...)
	}

	_, err := s.writeX11Request(channel, 67, 0, payload)
	if err != nil {
		return err
	}
	return nil
}

func (s *sshServer) polyPoint(channel ssh.Channel, drawable, gc uint32, coordinateMode byte, points []int16) error {
	newOp := X11Operation{
		Type:  "polyPoint",
		Color: gcColors[gc],
		Args:  []any{drawable, gcToMap(gc), uint32Slice(points)},
	}
	x11Operations = append(x11Operations, newOp)

	// Opcode: 64
	// Coordinate Mode: coordinateMode
	// Request Length: calculated
	// Drawable ID: drawable
	// GC ID: gc
	// Points: list of X, Y pairs

	payload := make([]byte, 8)
	binary.LittleEndian.PutUint32(payload[0:4], drawable)
	binary.LittleEndian.PutUint32(payload[4:8], gc)

	for i := 0; i < len(points); i += 2 {
		pointBytes := make([]byte, 4)
		binary.LittleEndian.PutUint16(pointBytes[0:2], uint16(points[i]))
		binary.LittleEndian.PutUint16(pointBytes[2:4], uint16(points[i+1]))
		payload = append(payload, pointBytes...)
	}

	_, err := s.writeX11Request(channel, 64, coordinateMode, payload)
	if err != nil {
		return err
	}
	return nil
}

func (s *sshServer) polySegment(channel ssh.Channel, drawable, gc uint32, segments []int16) error {
	newOp := X11Operation{
		Type:  "polySegment",
		Color: gcColors[gc],
		Args:  []any{drawable, gcToMap(gc), uint32Slice(segments)},
	}
	x11Operations = append(x11Operations, newOp)

	// Opcode: 66
	// Request Length: calculated
	// Drawable ID: drawable
	// GC ID: gc
	// Segments: list of X1, Y1, X2, Y2 tuples

	payload := make([]byte, 8)
	binary.LittleEndian.PutUint32(payload[0:4], drawable)
	binary.LittleEndian.PutUint32(payload[4:8], gc)

	for i := 0; i < len(segments); i += 4 {
		segmentBytes := make([]byte, 8)
		binary.LittleEndian.PutUint16(segmentBytes[0:2], uint16(segments[i]))
		binary.LittleEndian.PutUint16(segmentBytes[2:4], uint16(segments[i+1]))
		binary.LittleEndian.PutUint16(segmentBytes[4:6], uint16(segments[i+2]))
		binary.LittleEndian.PutUint16(segmentBytes[6:8], uint16(segments[i+3]))
		payload = append(payload, segmentBytes...)
	}

	_, err := s.writeX11Request(channel, 66, 0, payload)
	if err != nil {
		return err
	}
	return nil
}

func (s *sshServer) fillPoly(channel ssh.Channel, drawable, gc uint32, shape byte, points []int16) error {
	newOp := X11Operation{
		Type:  "fillPoly",
		Color: gcColors[gc],
		Args:  []any{drawable, gcToMap(gc), uint32Slice(points)},
	}
	x11Operations = append(x11Operations, newOp)

	// Opcode: 69
	// Shape: shape (passed as data byte)
	// Request Length: calculated
	// Drawable ID: drawable
	// GC ID: gc
	// Coordinate Mode: coordinateMode
	// Points: list of X, Y pairs

	payload := make([]byte, 12) // drawable (4) + gc (4) + coordinateMode (1) + padding (3)
	binary.LittleEndian.PutUint32(payload[0:4], drawable)
	binary.LittleEndian.PutUint32(payload[4:8], gc)
	payload[8] = 0 // coordinateMode (always 0 for now)
	// payload[9:12] is padding

	for i := 0; i < len(points); i += 2 {
		pointBytes := make([]byte, 4)
		binary.LittleEndian.PutUint16(pointBytes[0:2], uint16(points[i]))
		binary.LittleEndian.PutUint16(pointBytes[2:4], uint16(points[i+1]))
		payload = append(payload, pointBytes...)
	}

	_, err := s.writeX11Request(channel, 69, shape, payload)
	if err != nil {
		return err
	}
	return nil
}

func uint32Slice(in []int16) []any {
	out := make([]any, len(in))
	for i, v := range in {
		out[i] = uint32(v)
	}
	return out
}

func (s *sshServer) polyFillRectangle(channel ssh.Channel, drawable, gc uint32, rects []int16) error {
	newOp := X11Operation{
		Type:  "polyFillRectangle",
		Color: gcColors[gc],
		Args:  []any{drawable, gcToMap(gc), uint32Slice(rects)},
	}
	x11Operations = append(x11Operations, newOp)

	// Opcode: 63
	// Request Length: calculated
	// Drawable ID: drawable
	// GC ID: gc
	// Rectangles: list of X, Y, Width, Height tuples

	payload := make([]byte, 8)
	binary.LittleEndian.PutUint32(payload[0:4], drawable)
	binary.LittleEndian.PutUint32(payload[4:8], gc)

	for i := 0; i < len(rects); i += 4 {
		rectBytes := make([]byte, 8)
		binary.LittleEndian.PutUint16(rectBytes[0:2], uint16(rects[i]))
		binary.LittleEndian.PutUint16(rectBytes[2:4], uint16(rects[i+1]))
		binary.LittleEndian.PutUint16(rectBytes[4:6], uint16(rects[i+2]))
		binary.LittleEndian.PutUint16(rectBytes[6:8], uint16(rects[i+3]))
		payload = append(payload, rectBytes...)
	}

	_, err := s.writeX11Request(channel, 70, 0, payload)
	if err != nil {
		return err
	}
	return nil
}

var gcColors = make(map[uint32]uint32)

func (s *sshServer) createGC(channel ssh.Channel, gcID, drawable, foregroundColor uint32) error {
	return s.createGCWithBackground(channel, gcID, drawable, foregroundColor, 0)
}

func (s *sshServer) createGCWithBackground(channel ssh.Channel, gcID, drawable, foregroundColor, backgroundColor uint32) error {
	gcColors[gcID] = foregroundColor
	x11Operations = append(x11Operations, X11Operation{
		Type: "createGC",
		Args: []any{gcID},
	})

	// Opcode: 55
	// Request Length: calculated
	// GC ID: gcID
	// Drawable ID: drawable
	// Value Mask: GCForeground (1<<2) | GCBackground (1<<3)
	// Value List: foreground pixel, background pixel

	valueMask := uint32((1 << 2) | (1 << 3)) // GCForeground | GCBackground

	// Fixed part of payload (8 bytes)
	payload := make([]byte, 8)
	binary.LittleEndian.PutUint32(payload[0:4], gcID)
	binary.LittleEndian.PutUint32(payload[4:8], drawable)

	// Value Mask (4 bytes)
	valueMaskBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(valueMaskBytes[0:4], valueMask)

	// Value List (8 bytes: foregroundColor, backgroundColor)
	valueList := make([]byte, 8)
	binary.LittleEndian.PutUint32(valueList[0:4], foregroundColor)
	binary.LittleEndian.PutUint32(valueList[4:8], backgroundColor)

	fullPayload := append(payload, valueMaskBytes...)
	fullPayload = append(fullPayload, valueList...)

	_, err := s.writeX11Request(channel, 55, 0, fullPayload)
	return err
}

func (s *sshServer) polyLine(channel ssh.Channel, drawable, gc uint32, coordinateMode byte, points []int16) error {
	x11Operations = append(x11Operations, X11Operation{
		Type:  "polyLine",
		Color: gcColors[gc],
		Args:  []any{drawable, gcToMap(gc), uint32Slice(points)},
	})

	// Opcode: 65
	// Coordinate Mode: coordinateMode
	// Request Length: calculated
	// Drawable ID: drawable
	// GC ID: gc
	// Points: list of X, Y pairs

	payload := make([]byte, 8)
	binary.LittleEndian.PutUint32(payload[0:4], drawable)
	binary.LittleEndian.PutUint32(payload[4:8], gc)

	for i := 0; i < len(points); i += 2 {
		pointBytes := make([]byte, 4)
		binary.LittleEndian.PutUint16(pointBytes[0:2], uint16(points[i]))
		binary.LittleEndian.PutUint16(pointBytes[2:4], uint16(points[i+1]))
		payload = append(payload, pointBytes...)
	}

	_, err := s.writeX11Request(channel, 65, coordinateMode, payload)
	if err != nil {
		return err
	}
	return nil
}

func (s *sshServer) mapWindow(channel ssh.Channel, wid uint32) error {
	x11Operations = append(x11Operations, X11Operation{
		Type: "mapWindow",
		Args: []any{wid},
	})
	payload := make([]byte, 4)
	binary.LittleEndian.PutUint32(payload[0:4], wid)
	_, err := s.writeX11Request(channel, 8, 0, payload)
	if err != nil {
		return err
	}
	return nil
}

func (s *sshServer) imageText8(channel ssh.Channel, drawable, gc uint32, x, y int16, text []byte) error {
	x11Operations = append(x11Operations, X11Operation{
		Type:  "imageText8",
		Color: gcColors[gc],
		Args:  []any{uint32(drawable), gcToMap(gc), uint32(x), uint32(y), string(text)},
	})
	// Opcode: 76
	// Delta: len(text)
	// Request Length: calculated
	// Drawable ID: drawable
	// GC ID: gc
	// X, Y: x, y
	// Text: text

	payload := make([]byte, 12) // drawable (4) + gc (4) + x (2) + y (2)
	binary.LittleEndian.PutUint32(payload[0:4], drawable)
	binary.LittleEndian.PutUint32(payload[4:8], gc)
	binary.LittleEndian.PutUint16(payload[8:10], uint16(x))
	binary.LittleEndian.PutUint16(payload[10:12], uint16(y))

	fullPayload := append(payload, text...)

	_, err := s.writeX11Request(channel, 76, byte(len(text)), fullPayload)
	if err != nil {
		return err
	}
	return nil
}

func (s *sshServer) imageText16(channel ssh.Channel, drawable, gc uint32, x, y int16, text []uint16) error {
	x11Operations = append(x11Operations, X11Operation{
		Type:  "imageText16",
		Color: gcColors[gc],
		Args:  []any{uint32(drawable), gcToMap(gc), uint32(x), uint32(y), uint16SliceToString(text)},
	})

	payload := make([]byte, 12) // drawable (4) + gc (4) + x (2) + y (2)
	binary.LittleEndian.PutUint32(payload[0:4], drawable)
	binary.LittleEndian.PutUint32(payload[4:8], gc)
	binary.LittleEndian.PutUint16(payload[8:10], uint16(x))
	binary.LittleEndian.PutUint16(payload[10:12], uint16(y))

	textBytes := make([]byte, len(text)*2)
	for i, r := range text {
		binary.LittleEndian.PutUint16(textBytes[i*2:], r)
	}
	fullPayload := append(payload, textBytes...)

	_, err := s.writeX11Request(channel, 77, byte(len(text)), fullPayload)
	if err != nil {
		return err
	}
	return nil
}

type PolyText8Item struct {
	Delta int8
	Str   []byte
}

func (s *sshServer) polyText8(channel ssh.Channel, drawable, gc uint32, x, y int16, items []PolyText8Item) error {
	recordedItems := make([]any, len(items))
	for i, item := range items {
		recordedItems[i] = map[string]any{"delta": item.Delta, "text": string(item.Str)}
	}
	x11Operations = append(x11Operations, X11Operation{
		Type:  "polyText8",
		Color: gcColors[gc],
		Args:  []any{uint32(drawable), gcToMap(gc), uint32(x), uint32(y), recordedItems},
	})

	payload := make([]byte, 12) // drawable (4) + gc (4) + x (2) + y (2)
	binary.LittleEndian.PutUint32(payload[0:4], drawable)
	binary.LittleEndian.PutUint32(payload[4:8], gc)
	binary.LittleEndian.PutUint16(payload[8:10], uint16(x))
	binary.LittleEndian.PutUint16(payload[10:12], uint16(y))

	for _, item := range items {
		n := len(item.Str)
		payload = append(payload, byte(n))
		payload = append(payload, byte(item.Delta))
		payload = append(payload, item.Str...)
		padding := (4 - (n+2)%4) % 4
		payload = append(payload, make([]byte, padding)...)
	}

	_, err := s.writeX11Request(channel, 74, 0, payload)
	if err != nil {
		return err
	}
	return nil
}

type PolyText16Item struct {
	Delta int8
	Str   []uint16
}

func (s *sshServer) polyText16(channel ssh.Channel, drawable, gc uint32, x, y int16, items []PolyText16Item) error {
	recordedItems := make([]any, len(items))
	for i, item := range items {
		recordedItems[i] = map[string]any{"delta": item.Delta, "text": uint16SliceToString(item.Str)}
	}
	x11Operations = append(x11Operations, X11Operation{
		Type:  "polyText16",
		Color: gcColors[gc],
		Args:  []any{uint32(drawable), gcToMap(gc), uint32(x), uint32(y), recordedItems},
	})

	payload := make([]byte, 12) // drawable (4) + gc (4) + x (2) + y (2)
	binary.LittleEndian.PutUint32(payload[0:4], drawable)
	binary.LittleEndian.PutUint32(payload[4:8], gc)
	binary.LittleEndian.PutUint16(payload[8:10], uint16(x))
	binary.LittleEndian.PutUint16(payload[10:12], uint16(y))

	for _, item := range items {
		n := len(item.Str)
		payload = append(payload, byte(n))
		payload = append(payload, byte(item.Delta))
		textBytes := make([]byte, n*2)
		for i, r := range item.Str {
			binary.LittleEndian.PutUint16(textBytes[i*2:], r)
		}
		payload = append(payload, textBytes...)
		padding := (4 - (n*2+2)%4) % 4
		payload = append(payload, make([]byte, padding)...)
	}

	_, err := s.writeX11Request(channel, 75, 0, payload)
	if err != nil {
		return err
	}
	return nil
}

func (s *sshServer) openFont(channel ssh.Channel, fid uint32, name string) error {
	x11Operations = append(x11Operations, X11Operation{
		Type: "openFont",
		Args: []any{fid, name},
	})
	// Opcode: 45
	// Request Length: calculated
	// Font ID: fid
	// Name Length: len(name)
	// Name: name

	nameBytes := []byte(name)
	nameLen := uint16(len(nameBytes))

	payload := make([]byte, 4+2+2) // fid + nameLen + padding
	binary.LittleEndian.PutUint32(payload[0:4], fid)
	binary.LittleEndian.PutUint16(payload[4:6], nameLen)
	// payload[6:8] is padding

	payload = append(payload, nameBytes...)

	_, err := s.writeX11Request(channel, 45, 0, payload)
	if err != nil {
		return err
	}
	return nil
}

func (s *sshServer) closeFont(channel ssh.Channel, fid uint32) error {
	x11Operations = append(x11Operations, X11Operation{
		Type: "closeFont",
		Args: []any{fid},
	})
	// Opcode: 46
	// Request Length: 2
	// Font ID: fid
	payload := make([]byte, 4)
	binary.LittleEndian.PutUint32(payload[0:4], fid)
	_, err := s.writeX11Request(channel, 46, 0, payload)
	if err != nil {
		return err
	}
	return nil
}

func (s *sshServer) queryFont(channel ssh.Channel, fid uint32, replyChan <-chan []byte) error {
	x11Operations = append(x11Operations, X11Operation{
		Type: "queryFont",
		Args: []any{fid},
	})
	// Opcode: 47
	// Request Length: 2
	// Font ID: fid
	payload := make([]byte, 4)
	binary.LittleEndian.PutUint32(payload[0:4], fid)

	expectedSequence, err := s.writeX11Request(channel, 47, 0, payload)
	if err != nil {
		return fmt.Errorf("failed to write QueryFont request: %w", err)
	}

	reply := <-replyChan
	sequenceNumber := binary.LittleEndian.Uint16(reply[2:4])
	if sequenceNumber != expectedSequence {
		return fmt.Errorf("unexpected reply sequence number %d != %d", sequenceNumber, expectedSequence)
	}
	minBounds := xCharInfo{
		LeftSideBearing:  int16(binary.LittleEndian.Uint16(reply[8:10])),
		RightSideBearing: int16(binary.LittleEndian.Uint16(reply[10:12])),
		CharacterWidth:   binary.LittleEndian.Uint16(reply[12:14]),
		Ascent:           int16(binary.LittleEndian.Uint16(reply[14:16])),
		Descent:          int16(binary.LittleEndian.Uint16(reply[16:18])),
		Attributes:       binary.LittleEndian.Uint16(reply[18:20]),
	}

	maxBounds := xCharInfo{
		LeftSideBearing:  int16(binary.LittleEndian.Uint16(reply[24:26])),
		RightSideBearing: int16(binary.LittleEndian.Uint16(reply[26:28])),
		CharacterWidth:   binary.LittleEndian.Uint16(reply[28:30]),
		Ascent:           int16(binary.LittleEndian.Uint16(reply[30:32])),
		Descent:          int16(binary.LittleEndian.Uint16(reply[32:34])),
		Attributes:       binary.LittleEndian.Uint16(reply[34:36]),
	}

	minCharOrByte2 := binary.LittleEndian.Uint16(reply[40:42])
	maxCharOrByte2 := binary.LittleEndian.Uint16(reply[42:44])
	defaultChar := binary.LittleEndian.Uint16(reply[44:46])
	nFontProps := binary.LittleEndian.Uint16(reply[46:48])
	// drawDirection := reply[48]
	minByte1 := reply[49]
	maxByte1 := reply[50]
	allCharsExist := reply[51] != 0
	fontAscent := int16(binary.LittleEndian.Uint16(reply[52:54]))
	fontDescent := int16(binary.LittleEndian.Uint16(reply[54:56]))
	nCharInfos := binary.LittleEndian.Uint32(reply[56:60])

	s.t.Logf("QueryFont Reply: fid=%d, length=%d, minCharOrByte2=%d, maxCharOrByte2=%d, defaultChar=%d, nFontProps=%d, minByte1=%d, maxByte1=%d, allCharsExist=%t, fontAscent=%d, fontDescent=%d, nCharInfos=%d",
		fid, len(reply), minCharOrByte2, maxCharOrByte2, defaultChar, nFontProps, minByte1, maxByte1, allCharsExist, fontAscent, fontDescent, nCharInfos)
	s.t.Logf("  minBounds: %+v", minBounds)
	s.t.Logf("  maxBounds: %+v", maxBounds)

	// Validate values
	if fontAscent <= 0 || fontDescent <= 0 {
		s.t.Errorf("ERR fontAscent (%d) or fontDescent (%d) is not positive", fontAscent, fontDescent)
	}
	if minBounds.Ascent <= 0 || minBounds.Descent <= 0 {
		s.t.Errorf("ERR minBounds.Ascent (%d) or minBounds.Descent (%d) is not positive", minBounds.Ascent, minBounds.Descent)
	}
	if maxBounds.Ascent <= 0 || maxBounds.Descent <= 0 {
		s.t.Errorf("ERR maxBounds.Ascent (%d) or maxBounds.Descent (%d) is not positive", maxBounds.Ascent, maxBounds.Descent)
	}
	if nCharInfos != uint32(maxCharOrByte2-minCharOrByte2+1) {
		s.t.Errorf("ERR nCharInfos (%d) does not match expected count (%d)", nCharInfos, maxCharOrByte2-minCharOrByte2+1)
	}

	// Read font properties (if any)
	for i := 0; i < int(nFontProps); i++ {
		propBytes := make([]byte, 8)
		copy(propBytes, reply[60+8*i:])
		propName := binary.LittleEndian.Uint32(propBytes[0:4])
		propValue := binary.LittleEndian.Uint32(propBytes[4:8])
		s.t.Logf("  Font Property %d: Name=%d, Value=%d", i, propName, propValue)
	}

	// Read character info (if any)
	for i := 0; i < int(nCharInfos); i++ {
		charInfoBytes := make([]byte, 12)
		copy(charInfoBytes, reply[60+8*int(nFontProps)+12*i:])
		ci := xCharInfo{
			LeftSideBearing:  int16(binary.LittleEndian.Uint16(charInfoBytes[0:2])),
			RightSideBearing: int16(binary.LittleEndian.Uint16(charInfoBytes[2:4])),
			CharacterWidth:   binary.LittleEndian.Uint16(charInfoBytes[4:6]),
			Ascent:           int16(binary.LittleEndian.Uint16(charInfoBytes[6:8])),
			Descent:          int16(binary.LittleEndian.Uint16(charInfoBytes[8:10])),
			Attributes:       binary.LittleEndian.Uint16(charInfoBytes[10:12]),
		}
		if ci.Ascent <= 0 || ci.Descent <= 0 {
			s.t.Errorf("ERR char info %d: Ascent (%d) or Descent (%d) is not positive", i, ci.Ascent, ci.Descent)
		}
		// s.t.Logf("  Char Info %d: %+v", i, ci)
	}
	s.t.Log("QueryFont reply validated successfully")

	return nil
}

// xCharInfo matches the structure in go/internal/x11/x11.go
type xCharInfo struct {
	LeftSideBearing  int16
	RightSideBearing int16
	CharacterWidth   uint16
	Ascent           int16
	Descent          int16
	Attributes       uint16
}

// xFontProp matches the structure in X11 protocol (name: Atom, value: CARD32)
type xFontProp struct {
	Name  uint32
	Value uint32
}

func (s *sshServer) listFonts(channel ssh.Channel, maxNames uint16, pattern string, replyChan <-chan []byte) error {
	x11Operations = append(x11Operations, X11Operation{
		Type: "listFonts",
		Args: []any{maxNames, pattern},
	})
	// Opcode: 49
	// Request Length: calculated
	// Max Names: maxNames
	// Pattern Length: len(pattern)
	// Pattern: pattern

	patternBytes := []byte(pattern)
	patternLen := uint16(len(patternBytes))

	payload := make([]byte, 2+2) // maxNames + patternLen
	binary.LittleEndian.PutUint16(payload[0:2], maxNames)
	binary.LittleEndian.PutUint16(payload[2:4], patternLen)

	payload = append(payload, patternBytes...)

	_, err := s.writeX11Request(channel, 49, 0, payload)
	if err != nil {
		return err
	}
	<-replyChan
	return nil
}

func (s *sshServer) createGCWithFont(channel ssh.Channel, gcID, drawable, foregroundColor, backgroundColor, fontID uint32) error {
	gcColors[gcID] = foregroundColor
	x11Operations = append(x11Operations, X11Operation{
		Type: "createGC",
		Args: []any{gcID},
	})

	// Opcode: 55
	// Request Length: calculated
	// GC ID: gcID
	// Drawable ID: drawable
	// Value Mask: GCForeground (1<<2) | GCBackground (1<<3) | GCFont (1<<14)
	// Value List: foreground pixel, background pixel, font id

	valueMask := uint32((1 << 2) | (1 << 3) | (1 << 14)) // GCForeground | GCBackground | GCFont

	// Fixed part of payload (8 bytes)
	payload := make([]byte, 8)
	binary.LittleEndian.PutUint32(payload[0:4], gcID)
	binary.LittleEndian.PutUint32(payload[4:8], drawable)

	// Value Mask (4 bytes)
	valueMaskBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(valueMaskBytes[0:4], valueMask)

	// Value List (12 bytes: foregroundColor, backgroundColor, fontID)
	valueList := make([]byte, 12)
	binary.LittleEndian.PutUint32(valueList[0:4], foregroundColor)
	binary.LittleEndian.PutUint32(valueList[4:8], backgroundColor)
	binary.LittleEndian.PutUint32(valueList[8:12], fontID)

	_, err := s.writeX11Request(channel, 55, 0, append(append(payload, valueMaskBytes...), valueList...))
	if err != nil {
		return err
	}
	return nil
}

func uint16SliceToString(s []uint16) string {
	runes := make([]rune, len(s))
	for i, v := range s {
		runes[i] = rune(v)
	}
	return string(runes)
}

func (s *sshServer) writeX11Request(channel ssh.Channel, opcode byte, data byte, payload []byte) (uint16, error) {
	// X11 requests must be padded to a multiple of 4 bytes.
	padding := (4 - (len(payload) % 4)) % 4
	paddedPayload := make([]byte, len(payload)+padding)
	copy(paddedPayload, payload)

	requestLength := uint16((4 + len(paddedPayload)) / 4) // in 4-byte units

	header := make([]byte, 4)
	header[0] = opcode
	header[1] = data
	binary.LittleEndian.PutUint16(header[2:4], requestLength)

	_, err := channel.Write(header)
	if err != nil {
		return 0, fmt.Errorf("failed to write X11 request header: %w", err)
	}

	_, err = channel.Write(paddedPayload)
	if err != nil {
		return 0, fmt.Errorf("failed to write X11 request payload: %w", err)
	}

	s.clientSequence++
	s.t.Logf("Sent X11 request %d", s.clientSequence)
	return s.clientSequence, nil
}

func (s *sshServer) createWindow(channel ssh.Channel, wid, parent, x, y, width, height uint32) error {
	x11Operations = append(x11Operations, X11Operation{
		Type: "createWindow",
		Args: []any{wid, parent, x, y, width, height, uint32(24)},
	})
	// Opcode: 1
	// Depth: 24 (arbitrary)
	// Request Length: calculated
	// Window ID: wid
	// Parent ID: parent
	// X, Y, Width, Height, Border Width: x, y, width, height, 0
	// Class: InputOutput (1)
	// Visual: CopyFromParent (0)
	// Value Mask: CWBackPixel (1<<1) | CWEventMask (1<<11)
	// Value List: background pixel, event mask

	depth := byte(24)
	borderWidth := uint16(0)
	class := uint16(1)  // InputOutput
	visual := uint32(0) // CopyFromParent

	valueMask := uint32(1 << 1)         // CWBackPixel
	backgroundPixel := uint32(0xFFFFFF) // White

	// Fixed part of payload (24 bytes: Window ID, Parent ID, X, Y, Width, Height, Border Width, Class, Visual)
	payload := make([]byte, 24)
	binary.LittleEndian.PutUint32(payload[0:4], wid)
	binary.LittleEndian.PutUint32(payload[4:8], parent)
	binary.LittleEndian.PutUint16(payload[8:10], uint16(x))
	binary.LittleEndian.PutUint16(payload[10:12], uint16(y))
	binary.LittleEndian.PutUint16(payload[12:14], uint16(width))
	binary.LittleEndian.PutUint16(payload[14:16], uint16(height))
	binary.LittleEndian.PutUint16(payload[16:18], borderWidth)
	binary.LittleEndian.PutUint16(payload[18:20], class)
	binary.LittleEndian.PutUint32(payload[20:24], visual)

	// Value Mask (4 bytes)
	valueMaskBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(valueMaskBytes[0:4], valueMask)

	// Value List (4 bytes: CWBackPixel)
	valueList := make([]byte, 4)
	binary.LittleEndian.PutUint32(valueList[0:4], backgroundPixel)

	fullPayload := append(payload, valueMaskBytes...)
	fullPayload = append(fullPayload, valueList...)

	_, err := s.writeX11Request(channel, 1, depth, fullPayload)
	if err != nil {
		return err
	}
	return nil
}
