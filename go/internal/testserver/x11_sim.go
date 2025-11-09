package main

import (
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"io"
	"sort"
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
	s.simulateColorOperations(x11Channel, replyChan)
	s.simulateGCOperations(x11Channel)

	s.t.Log("All drawing commands sent successfully")
	close(s.x11SimDone)
	time.Sleep(2 * time.Second)
	x11Channel.Close()
}

const (
	GCFunction          = 1 << 0
	GCPlaneMask         = 1 << 1
	GCForeground        = 1 << 2
	GCBackground        = 1 << 3
	GCLineWidth         = 1 << 4
	GCLineStyle         = 1 << 5
	GCCapStyle          = 1 << 6
	GCJoinStyle         = 1 << 7
	GCFillStyle         = 1 << 8
	GCFillRule          = 1 << 9
	GCTile              = 1 << 10
	GCStipple           = 1 << 11
	GCTileStipXOrigin   = 1 << 12
	GCTileStipYOrigin   = 1 << 13
	GCFont              = 1 << 14
	GCSubwindowMode     = 1 << 15
	GCGraphicsExposures = 1 << 16
	GCClipXOrigin       = 1 << 17
	GCClipYOrigin       = 1 << 18
	GCClipMask          = 1 << 19
	GCDashOffset        = 1 << 20
	GCDashList          = 1 << 21
	GCArcMode           = 1 << 22
)

func gcValuesToMap(values map[uint32]uint32) map[string]interface{} {
	m := make(map[string]interface{})
	if v, ok := values[GCFunction]; ok {
		m["Function"] = v
	}
	if v, ok := values[GCPlaneMask]; ok {
		m["PlaneMask"] = v
	}
	if v, ok := values[GCForeground]; ok {
		m["Foreground"] = v
	}
	if v, ok := values[GCBackground]; ok {
		m["Background"] = v
	}
	if v, ok := values[GCLineWidth]; ok {
		m["LineWidth"] = v
	}
	if v, ok := values[GCLineStyle]; ok {
		m["LineStyle"] = v
	}
	if v, ok := values[GCCapStyle]; ok {
		m["CapStyle"] = v
	}
	if v, ok := values[GCJoinStyle]; ok {
		m["JoinStyle"] = v
	}
	if v, ok := values[GCFillStyle]; ok {
		m["FillStyle"] = v
	}
	if v, ok := values[GCFillRule]; ok {
		m["FillRule"] = v
	}
	if v, ok := values[GCTile]; ok {
		m["Tile"] = v
	}
	if v, ok := values[GCStipple]; ok {
		m["Stipple"] = v
	}
	if v, ok := values[GCTileStipXOrigin]; ok {
		m["TileStipXOrigin"] = v
	}
	if v, ok := values[GCTileStipYOrigin]; ok {
		m["TileStipYOrigin"] = v
	}
	if v, ok := values[GCFont]; ok {
		m["Font"] = v
	}
	if v, ok := values[GCSubwindowMode]; ok {
		m["SubwindowMode"] = v
	}
	if v, ok := values[GCGraphicsExposures]; ok {
		m["GraphicsExposures"] = v
	}
	if v, ok := values[GCClipXOrigin]; ok {
		m["ClipXOrigin"] = v
	}
	if v, ok := values[GCClipYOrigin]; ok {
		m["ClipYOrigin"] = v
	}
	if v, ok := values[GCClipMask]; ok {
		m["ClipMask"] = v
	}
	if v, ok := values[GCDashOffset]; ok {
		m["DashOffset"] = v
	}
	if v, ok := values[GCDashList]; ok {
		m["Dashes"] = v
	}
	if v, ok := values[GCArcMode]; ok {
		m["ArcMode"] = v
	}
	return m
}

func (s *sshServer) simulateGCOperations(channel ssh.Channel) {
	s.t.Log("Simulating GC operations")

	// Create a new window for GC tests
	gcWindowID := uint32(30)
	if err := s.createWindow(channel, gcWindowID, 1, 220, 220, 300, 300); err != nil {
		s.t.Errorf("Failed to create GC test window: %v", err)
		return
	}
	if err := s.mapWindow(channel, gcWindowID); err != nil {
		s.t.Errorf("Failed to map GC test window: %v", err)
		return
	}

	// Create GCs with different attributes
	// GC for thick red line
	gcThickRed := uint32(300)
	if err := s.createGCWithAttributes(channel, gcThickRed, gcWindowID, map[uint32]uint32{
		GCForeground: 0xFF0000,
		GCLineWidth:  5,
	}); err != nil {
		s.t.Errorf("Failed to create thick red GC: %v", err)
		return
	}
	s.polyLine(channel, gcWindowID, gcThickRed, 0, []int16{10, 10, 100, 10})

	// GC for dashed blue line with round caps and joins
	gcDashedBlue := uint32(301)
	if err := s.createGCWithAttributes(channel, gcDashedBlue, gcWindowID, map[uint32]uint32{
		GCForeground: 0x0000FF,
		GCCapStyle:   2, // Round
		GCJoinStyle:  1, // Round
		GCDashList:   4,
	}); err != nil {
		s.t.Errorf("Failed to create dashed blue GC: %v", err)
		return
	}
	if err := s.setDashes(channel, gcDashedBlue, 0, []byte{4, 4}); err != nil {
		s.t.Errorf("Failed to set dashes: %v", err)
		return
	}
	s.polyLine(channel, gcWindowID, gcDashedBlue, 0, []int16{10, 30, 100, 30, 100, 50})

	// GC for winding fill rule
	gcWindingFill := uint32(302)
	if err := s.createGCWithAttributes(channel, gcWindingFill, gcWindowID, map[uint32]uint32{
		GCForeground: 0x00FF00,
		GCFillRule:   1, // Winding
	}); err != nil {
		s.t.Errorf("Failed to create winding fill GC: %v", err)
		return
	}
	points := []int16{150, 10, 180, 60, 120, 60, 150, 10}
	s.fillPoly(channel, gcWindowID, gcWindingFill, 0, points)

	// Tiled rectangle
	tilePixmapID := uint32(400)
	s.createPixmap(channel, tilePixmapID, gcWindowID, 8, 8, 24)
	gcTile := uint32(303)
	if err := s.createGCWithAttributes(channel, gcTile, tilePixmapID, map[uint32]uint32{
		GCForeground: 0xFF00FF,
	}); err != nil {
		s.t.Errorf("Failed to create tile GC: %v", err)
		return
	}
	s.polyFillRectangle(channel, tilePixmapID, gcTile, []int16{0, 0, 4, 4})
	s.polyFillRectangle(channel, tilePixmapID, gcTile, []int16{4, 4, 4, 4})
	gcTiledFill := uint32(304)
	if err := s.createGCWithAttributes(channel, gcTiledFill, gcWindowID, map[uint32]uint32{
		GCFillStyle: 1, // Tiled
		GCTile:      tilePixmapID,
	}); err != nil {
		s.t.Errorf("Failed to create tiled fill GC: %v", err)
		return
	}
	s.polyFillRectangle(channel, gcWindowID, gcTiledFill, []int16{10, 70, 100, 50})

	// Stippled rectangle
	stipplePixmapID := uint32(401)
	s.createPixmap(channel, stipplePixmapID, gcWindowID, 8, 8, 1)
	gcStipple := uint32(305)
	if err := s.createGCWithAttributes(channel, gcStipple, stipplePixmapID, map[uint32]uint32{
		GCForeground: 0x000000,
	}); err != nil {
		s.t.Errorf("Failed to create stipple GC: %v", err)
		return
	}
	s.polyPoint(channel, stipplePixmapID, gcStipple, 0, []int16{0, 0, 1, 1, 2, 2, 3, 3, 4, 4, 5, 5, 6, 6, 7, 7})
	gcStippledFill := uint32(306)
	if err := s.createGCWithAttributes(channel, gcStippledFill, gcWindowID, map[uint32]uint32{
		GCForeground: 0x800080, // Purple
		GCFillStyle:  2,        // Stippled
		GCStipple:    stipplePixmapID,
	}); err != nil {
		s.t.Errorf("Failed to create stippled fill GC: %v", err)
		return
	}
	s.polyFillRectangle(channel, gcWindowID, gcStippledFill, []int16{120, 70, 100, 50})

	// GC for XOR function
	gcXOR := uint32(307)
	if err := s.createGCWithAttributes(channel, gcXOR, gcWindowID, map[uint32]uint32{
		GCFunction:   6, // GXxor
		GCForeground: 0xFF00FF,
	}); err != nil {
		s.t.Errorf("Failed to create XOR GC: %v", err)
		return
	}
	s.polyFillRectangle(channel, gcWindowID, gcXOR, []int16{10, 130, 50, 50})
	s.polyFillRectangle(channel, gcWindowID, gcXOR, []int16{40, 160, 50, 50})

	// GC for ArcMode and font
	fontID := uint32(402)
	s.openFont(channel, fontID, "-*-helvetica-bold-r-normal--25-*-*-*-*-*-iso8859-1")
	gcArcFont := uint32(308)
	if err := s.createGCWithAttributes(channel, gcArcFont, gcWindowID, map[uint32]uint32{
		GCForeground: 0x000000,
		GCArcMode:    1, // PieSlice
		GCFont:       fontID,
	}); err != nil {
		s.t.Errorf("Failed to create arc/font GC: %v", err)
		return
	}
	s.polyFillArc(channel, gcWindowID, gcArcFont, []int16{120, 130, 100, 100, 0, 90 * 64})
	s.imageText8(channel, gcWindowID, gcArcFont, 120, 250, []byte("Arc"))
	s.closeFont(channel, fontID)
}

func (s *sshServer) readReplies(channel ssh.Channel) <-chan []byte {
	ch := make(chan []byte, 1)
	go func() {
		for {
			// Read Reply or Event
			replyHeader := make([]byte, 32)
			_, err := io.ReadFull(channel, replyHeader)
			if err != nil {
				if err != io.EOF {
					s.t.Logf("failed to read X11 message header: %v", err)
				}
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
				s.t.Errorf("Received X11 error: code=%d, sequence=%d", replyHeader[1], sequenceNumber)
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

func (s *sshServer) simulateColorOperations(channel ssh.Channel, replyChan <-chan []byte) {
	s.t.Log("Simulating color operations")

	// 1. Create a new colormap
	colormapID := uint32(2)
	if err := s.createColormap(channel, colormapID, 1, 0); err != nil {
		s.t.Errorf("Failed to create colormap: %v", err)
		return
	}

	// 2. Allocate some colors
	pixel, r, g, b, err := s.allocNamedColor(channel, colormapID, "blue", replyChan)
	if err != nil {
		s.t.Errorf("Failed to allocate named color: %v", err)
		return
	}
	if pixel != 0x0000ff || r != 0 || g != 0 || b != 0xffff {
		s.t.Errorf("ERR allocNamedColor(_, %d, blue, _) = %06x, (%04x, %04x, %04x)", colormapID, pixel, r, g, b)
	}
	pixel, r, g, b, err = s.allocColor(channel, colormapID, 0, 0, 65535, replyChan)
	if err != nil {
		s.t.Errorf("Failed to allocate color: %v", err)
		return
	}
	if pixel != 0x0000ff || r != 0 || g != 0 || b != 0xffff {
		s.t.Errorf("ERR allocColor(_, %d, 0, 0, 65535, _) = %06x, (%04x, %04x, %04x)", colormapID, pixel, r, g, b)
	}

	// 3. Create a new window with the new colormap
	if err := s.createWindowWithColormap(channel, 20, 1, 10, 20, 200, 200, colormapID); err != nil {
		s.t.Errorf("Failed to create window with colormap: %v", err)
		return
	}
	if err := s.mapWindow(channel, 20); err != nil {
		s.t.Errorf("Failed to map window: %v", err)
		return
	}

	// 4. Draw something in the new window
	blueGC := uint32(200)
	if err := s.createGCWithBackground(channel, blueGC, 20, 0x0000FF, 0); err != nil {
		s.t.Errorf("Failed to create blue GC: %v", err)
		return
	}
	rect := []int16{10, 10, 180, 180}
	if err := s.polyFillRectangle(channel, 20, blueGC, rect); err != nil {
		s.t.Errorf("Failed to draw rectangle: %v", err)
		return
	}

	// 5. Query colors
	if _, err := s.queryColors(channel, colormapID, []uint32{0x0000FF}, replyChan); err != nil {
		s.t.Errorf("Failed to query colors: %v", err)
		return
	}

	// 6. Install colormap
	if err := s.installColormap(channel, colormapID); err != nil {
		s.t.Errorf("Failed to install colormap: %v", err)
		return
	}

	// 7. List installed colormaps
	if _, err := s.listInstalledColormaps(channel, replyChan); err != nil {
		s.t.Errorf("Failed to list installed colormaps: %v", err)
		return
	}

	// 8. Free colors
	if err := s.freeColors(channel, colormapID, 0, []uint32{0x0000FF}); err != nil {
		s.t.Errorf("Failed to free colors: %v", err)
		return
	}

	// 9. Free colormap
	if err := s.freeColormap(channel, colormapID); err != nil {
		s.t.Errorf("Failed to free colormap: %v", err)
		return
	}
}

func GetX11Operations() []X11Operation {
	for i := range x11Operations {
		for j := range x11Operations[i].Args {
			x11Operations[i].Args[j] = fmt.Sprint(x11Operations[i].Args[j])
		}
	}
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
	return s.createGCWithAttributes(channel, gcID, drawable, map[uint32]uint32{
		1 << 2: foregroundColor,
		1 << 3: backgroundColor,
	})
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
	return s.createGCWithAttributes(channel, gcID, drawable, map[uint32]uint32{
		1 << 2:  foregroundColor,
		1 << 3:  backgroundColor,
		1 << 14: fontID,
	})
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
	return s.createWindowWithColormap(channel, wid, parent, x, y, width, height, 0)
}

func (s *sshServer) createWindowWithColormap(channel ssh.Channel, wid, parent, x, y, width, height, colormap uint32) error {
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
	// Value Mask: CWBackPixel (1<<1) | CWEventMask (1<<11) | CWColormap (1<<13)
	// Value List: background pixel, event mask, colormap

	depth := byte(24)
	borderWidth := uint16(0)
	class := uint16(1)  // InputOutput
	visual := uint32(0) // CopyFromParent

	valueMask := uint32((1 << 1) | (1 << 11) | (1 << 13)) // CWBackPixel | CWEventMask | CWColormap
	backgroundPixel := uint32(0xFFFFFF)                   // White
	eventMask := uint32(0)                                // No events

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

	// Value List (12 bytes: CWBackPixel, CWEventMask, CWColormap)
	valueList := make([]byte, 12)
	binary.LittleEndian.PutUint32(valueList[0:4], backgroundPixel)
	binary.LittleEndian.PutUint32(valueList[4:8], eventMask)
	binary.LittleEndian.PutUint32(valueList[8:12], colormap)

	fullPayload := append(payload, valueMaskBytes...)
	fullPayload = append(fullPayload, valueList...)

	_, err := s.writeX11Request(channel, 1, depth, fullPayload)
	if err != nil {
		return err
	}
	return nil
}

func (s *sshServer) createColormap(channel ssh.Channel, mid, window, visual uint32) error {
	payload := make([]byte, 12)
	binary.LittleEndian.PutUint32(payload[0:4], mid)
	binary.LittleEndian.PutUint32(payload[4:8], window)
	binary.LittleEndian.PutUint32(payload[8:12], visual)

	_, err := s.writeX11Request(channel, 78, 0, payload)
	return err
}

func (s *sshServer) freeColormap(channel ssh.Channel, cmap uint32) error {
	payload := make([]byte, 4)
	binary.LittleEndian.PutUint32(payload[0:4], cmap)

	_, err := s.writeX11Request(channel, 79, 0, payload)
	return err
}

func (s *sshServer) allocNamedColor(channel ssh.Channel, cmap uint32, name string, replyChan <-chan []byte) (uint32, uint16, uint16, uint16, error) {
	nameBytes := []byte(name)
	nameLen := uint16(len(nameBytes))

	payload := make([]byte, 4+2+2)
	binary.LittleEndian.PutUint32(payload[0:4], cmap)
	binary.LittleEndian.PutUint16(payload[4:6], nameLen)

	payload = append(payload, nameBytes...)

	expectedSequence, err := s.writeX11Request(channel, 85, 0, payload)
	if err != nil {
		return 0, 0, 0, 0, err
	}

	reply := <-replyChan
	sequenceNumber := binary.LittleEndian.Uint16(reply[2:4])
	if sequenceNumber != expectedSequence {
		return 0, 0, 0, 0, fmt.Errorf("unexpected reply sequence number %d != %d", sequenceNumber, expectedSequence)
	}

	pixel := binary.LittleEndian.Uint32(reply[8:12])
	red := binary.LittleEndian.Uint16(reply[12:14])
	green := binary.LittleEndian.Uint16(reply[14:16])
	blue := binary.LittleEndian.Uint16(reply[16:18])

	return pixel, red, green, blue, nil
}

func (s *sshServer) allocColor(channel ssh.Channel, cmap uint32, red, green, blue uint16, replyChan <-chan []byte) (uint32, uint16, uint16, uint16, error) {
	payload := make([]byte, 12)
	binary.LittleEndian.PutUint32(payload[0:4], cmap)
	binary.LittleEndian.PutUint16(payload[4:6], red)
	binary.LittleEndian.PutUint16(payload[6:8], green)
	binary.LittleEndian.PutUint16(payload[8:10], blue)

	expectedSequence, err := s.writeX11Request(channel, 84, 0, payload)
	if err != nil {
		return 0, 0, 0, 0, err
	}

	reply := <-replyChan
	sequenceNumber := binary.LittleEndian.Uint16(reply[2:4])
	if sequenceNumber != expectedSequence {
		return 0, 0, 0, 0, fmt.Errorf("unexpected reply sequence number %d != %d", sequenceNumber, expectedSequence)
	}

	red = binary.LittleEndian.Uint16(reply[8:10])
	green = binary.LittleEndian.Uint16(reply[10:12])
	blue = binary.LittleEndian.Uint16(reply[12:14])
	pixel := binary.LittleEndian.Uint32(reply[16:20])

	return pixel, red, green, blue, nil
}

func (s *sshServer) queryColors(channel ssh.Channel, cmap uint32, pixels []uint32, replyChan <-chan []byte) ([]uint16, error) {
	payload := make([]byte, 4)
	binary.LittleEndian.PutUint32(payload[0:4], cmap)

	for _, pixel := range pixels {
		pixelBytes := make([]byte, 4)
		binary.LittleEndian.PutUint32(pixelBytes, pixel)
		payload = append(payload, pixelBytes...)
	}

	expectedSequence, err := s.writeX11Request(channel, 91, 0, payload)
	if err != nil {
		return nil, err
	}

	reply := <-replyChan
	sequenceNumber := binary.LittleEndian.Uint16(reply[2:4])
	if sequenceNumber != expectedSequence {
		return nil, fmt.Errorf("unexpected reply sequence number %d != %d", sequenceNumber, expectedSequence)
	}

	numColors := binary.LittleEndian.Uint16(reply[8:10])
	colors := make([]uint16, numColors*3)
	for i := 0; i < int(numColors); i++ {
		offset := 12 + i*8
		colors[i*3] = binary.LittleEndian.Uint16(reply[offset : offset+2])
		colors[i*3+1] = binary.LittleEndian.Uint16(reply[offset+2 : offset+4])
		colors[i*3+2] = binary.LittleEndian.Uint16(reply[offset+4 : offset+6])
	}

	return colors, nil
}

func (s *sshServer) installColormap(channel ssh.Channel, cmap uint32) error {
	payload := make([]byte, 4)
	binary.LittleEndian.PutUint32(payload[0:4], cmap)

	_, err := s.writeX11Request(channel, 81, 0, payload)
	return err
}

func (s *sshServer) listInstalledColormaps(channel ssh.Channel, replyChan <-chan []byte) ([]uint32, error) {
	payload := make([]byte, 4)
	binary.LittleEndian.PutUint32(payload[0:4], 1) // dummy window

	expectedSequence, err := s.writeX11Request(channel, 83, 0, payload)
	if err != nil {
		return nil, err
	}

	reply := <-replyChan
	sequenceNumber := binary.LittleEndian.Uint16(reply[2:4])
	if sequenceNumber != expectedSequence {
		return nil, fmt.Errorf("unexpected reply sequence number %d != %d", sequenceNumber, expectedSequence)
	}

	numColormaps := binary.LittleEndian.Uint16(reply[8:10])
	colormaps := make([]uint32, numColormaps)
	for i := 0; i < int(numColormaps); i++ {
		colormaps[i] = binary.LittleEndian.Uint32(reply[12+i*4 : 16+i*4])
	}

	return colormaps, nil
}

func (s *sshServer) freeColors(channel ssh.Channel, cmap, planeMask uint32, pixels []uint32) error {
	payload := make([]byte, 8+len(pixels)*4)
	binary.LittleEndian.PutUint32(payload[0:4], cmap)
	binary.LittleEndian.PutUint32(payload[4:8], planeMask)

	for _, pixel := range pixels {
		pixelBytes := make([]byte, 4)
		binary.LittleEndian.PutUint32(pixelBytes, pixel)
		payload = append(payload, pixelBytes...)
	}

	_, err := s.writeX11Request(channel, 88, 0, payload)
	return err
}

func (s *sshServer) createGCWithAttributes(channel ssh.Channel, gcID, drawable uint32, values map[uint32]uint32) error {
	if foregroundColor, ok := values[GCForeground]; ok {
		gcColors[gcID] = foregroundColor
	}

	var valueMask uint32
	var sortedMasks []uint32
	for mask := range values {
		sortedMasks = append(sortedMasks, mask)
	}
	sort.Slice(sortedMasks, func(i, j int) bool {
		return sortedMasks[i] < sortedMasks[j]
	})

	var valueList []byte
	for _, mask := range sortedMasks {
		valueMask |= mask
		valueBytes := make([]byte, 4)
		binary.LittleEndian.PutUint32(valueBytes, values[mask])
		valueList = append(valueList, valueBytes...)
	}

	payload := make([]byte, 12) // gcID, drawable, valueMask
	binary.LittleEndian.PutUint32(payload[0:4], gcID)
	binary.LittleEndian.PutUint32(payload[4:8], drawable)
	binary.LittleEndian.PutUint32(payload[8:12], valueMask)
	payload = append(payload, valueList...)

	x11Operations = append(x11Operations, X11Operation{
		Type: "createGC",
		Args: []any{gcID, valueMask, gcValuesToMap(values)},
	})

	_, err := s.writeX11Request(channel, 55, 0, payload)
	return err
}

func (s *sshServer) setDashes(channel ssh.Channel, gcID uint32, dashOffset uint16, dashes []byte) error {
	x11Operations = append(x11Operations, X11Operation{
		Type: "setDashes",
		Args: []any{gcID, dashOffset, base64.StdEncoding.EncodeToString(dashes)},
	})
	payload := make([]byte, 8+len(dashes))
	binary.LittleEndian.PutUint32(payload[0:4], gcID)
	binary.LittleEndian.PutUint16(payload[4:6], dashOffset)
	binary.LittleEndian.PutUint16(payload[6:8], uint16(len(dashes)))
	copy(payload[8:], dashes)

	_, err := s.writeX11Request(channel, 58, 0, payload)
	return err
}

func (s *sshServer) createPixmap(channel ssh.Channel, pid, drawable, width, height, depth uint32) error {
	x11Operations = append(x11Operations, X11Operation{
		Type: "createPixmap",
		Args: []any{pid, drawable, width, height, depth},
	})
	payload := make([]byte, 16)
	binary.LittleEndian.PutUint32(payload[0:4], pid)
	binary.LittleEndian.PutUint32(payload[4:8], drawable)
	binary.LittleEndian.PutUint16(payload[8:10], uint16(width))
	binary.LittleEndian.PutUint16(payload[10:12], uint16(height))
	payload[12] = 0 // unused
	_, err := s.writeX11Request(channel, 53, byte(depth), payload)
	return err
}
