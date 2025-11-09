// MIT License
//
// Copyright (c) 2025 TTBT Enterprises LLC
// Copyright (c) 2025 Robin Thellend <rthellend@rthellend.com>
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT of OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

//go:build x11 && wasm

package x11

import (
	"bytes"
	_ "embed"
	"encoding/base64"
	"encoding/binary"
	"strings"
	"syscall/js"
	"testing"
	"time"
)

//go:embed testdata/TestDrawRectangle.png
var testDrawRectangleSnapshot []byte

//go:embed testdata/TestDrawText.png
var testDrawTextSnapshot []byte

//go:embed testdata/TestOverlappingWindows.png
var testOverlappingWindowsSnapshot []byte

const (
	snapshotDir = "testdata"
)

func snapshot(t *testing.T, golden []byte) {
	t.Helper()
	// A small delay to ensure the canvas has been painted.
	time.Sleep(100 * time.Millisecond)

	var dataURL string
	var errStr string
	done := make(chan struct{})

	js.Global().Set("snapshotCallback", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		dataURL = args[0].String()
		errStr = args[1].String()
		close(done)
		return nil
	}))
	defer js.Global().Set("snapshotCallback", js.Undefined())

	js.Global().Call("eval", `
		(async () => {
			try {
				// There might be multiple canvases; find the one for our window.
				const canvas = document.querySelector('#x11-canvas-1-1');
				if (!canvas) {
					throw new Error('canvas not found');
				}
				snapshotCallback(canvas.toDataURL(), '');
			} catch (e) {
				snapshotCallback('', e.toString());
			}
		})();
	`)

	<-done
	if errStr != "" {
		t.Fatalf("Failed to get canvas data: %s", errStr)
	}

	prefix := "data:image/png;base64,"
	if !strings.HasPrefix(dataURL, prefix) {
		t.Fatalf("Unexpected data URL format")
	}
	b, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(dataURL, prefix))
	if err != nil {
		t.Fatalf("Failed to decode base64: %v", err)
	}

	snapshotName := strings.ReplaceAll(t.Name(), "/", "_") + ".png"

	if len(golden) == 0 {
		t.Fatalf("Snapshot is empty. To update, decode this base64 string and save it to %s:\n%s", snapshotName, base64.StdEncoding.EncodeToString(b))
	}

	if !bytes.Equal(golden, b) {
		t.Errorf("Snapshot mismatch. To update, decode this base64 string and save it to %s:\n%s", snapshotName, base64.StdEncoding.EncodeToString(b))
	}
}

func TestDrawRectangle(t *testing.T) {
	t.Log("Running TestDrawRectangle")
	setup := newDefaultSetup()
	s := &x11Server{
		logger:          &testLogger{t: t},
		windows:         make(map[xID]*window),
		gcs:             make(map[xID]GC),
		pixmaps:         make(map[xID]bool),
		cursors:         make(map[xID]bool),
		selections:      make(map[xID]uint32),
		colormaps:       make(map[xID]*colormap),
		clients:         make(map[uint32]*x11Client),
		byteOrder:       binary.LittleEndian,
		passiveGrabs:    make(map[xID][]*passiveGrab),
		rootVisual:      setup.screens[0].depths[0].visuals[0],
		blackPixel:      setup.screens[0].blackPixel,
		whitePixel:      setup.screens[0].whitePixel,
		defaultColormap: setup.screens[0].defaultColormap,
	}
	fe := newX11Frontend(&testLogger{t: t}, s)
	s.frontend = fe

	winID := xID{client: 1, local: 1}
	fe.CreateWindow(winID, s.rootWindowID(), 10, 10, 100, 80, 24, 0, WindowAttributes{})
	fe.MapWindow(winID)

	gcID := xID{client: 1, local: 2}
	fe.CreateGC(gcID, GCForeground, GC{Foreground: s.blackPixel})

	fe.PolyFillRectangle(winID, gcID, []uint32{20, 20, 50, 40})

	snapshot(t, testDrawRectangleSnapshot)
}

func TestDrawText(t *testing.T) {
	t.Log("Running TestDrawText")
	setup := newDefaultSetup()
	s := &x11Server{
		logger:          &testLogger{t: t},
		windows:         make(map[xID]*window),
		gcs:             make(map[xID]GC),
		pixmaps:         make(map[xID]bool),
		cursors:         make(map[xID]bool),
		selections:      make(map[xID]uint32),
		colormaps:       make(map[xID]*colormap),
		clients:         make(map[uint32]*x11Client),
		byteOrder:       binary.LittleEndian,
		passiveGrabs:    make(map[xID][]*passiveGrab),
		rootVisual:      setup.screens[0].depths[0].visuals[0],
		blackPixel:      setup.screens[0].blackPixel,
		whitePixel:      setup.screens[0].whitePixel,
		defaultColormap: setup.screens[0].defaultColormap,
	}
	fe := newX11Frontend(&testLogger{t: t}, s)
	s.frontend = fe

	winID := xID{client: 1, local: 1}
	fe.CreateWindow(winID, s.rootWindowID(), 10, 10, 100, 80, 24, 0, WindowAttributes{})
	fe.MapWindow(winID)

	gcID := xID{client: 1, local: 2}
	fe.CreateGC(gcID, GCForeground, GC{Foreground: s.blackPixel})

	fe.PolyText8(winID, gcID, 20, 40, []PolyTextItem{
		PolyText8String{Str: []byte("Hello, world!")},
	})

	snapshot(t, testDrawTextSnapshot)
}

func TestOverlappingWindows(t *testing.T) {
	t.Log("Running TestOverlappingWindows")
	setup := newDefaultSetup()
	s := &x11Server{
		logger:          &testLogger{t: t},
		windows:         make(map[xID]*window),
		gcs:             make(map[xID]GC),
		pixmaps:         make(map[xID]bool),
		cursors:         make(map[xID]bool),
		selections:      make(map[xID]uint32),
		colormaps:       make(map[xID]*colormap),
		clients:         make(map[uint32]*x11Client),
		byteOrder:       binary.LittleEndian,
		passiveGrabs:    make(map[xID][]*passiveGrab),
		rootVisual:      setup.screens[0].depths[0].visuals[0],
		blackPixel:      setup.screens[0].blackPixel,
		whitePixel:      setup.screens[0].whitePixel,
		defaultColormap: setup.screens[0].defaultColormap,
	}
	fe := newX11Frontend(&testLogger{t: t}, s)
	s.frontend = fe

	winID1 := xID{client: 1, local: 1}
	fe.CreateWindow(winID1, s.rootWindowID(), 10, 10, 100, 80, 24, 0, WindowAttributes{})
	fe.MapWindow(winID1)

	gcID1 := xID{client: 1, local: 2}
	fe.CreateGC(gcID1, GCForeground, GC{Foreground: s.blackPixel})
	fe.PolyFillRectangle(winID1, gcID1, []uint32{20, 20, 50, 40})

	winID2 := xID{client: 1, local: 3}
	fe.CreateWindow(winID2, s.rootWindowID(), 30, 30, 100, 80, 24, 0, WindowAttributes{})
	fe.MapWindow(winID2)

	gcID2 := xID{client: 1, local: 4}
	fe.CreateGC(gcID2, GCForeground, GC{Foreground: s.blackPixel})
	fe.PolyFillRectangle(winID2, gcID2, []uint32{20, 20, 50, 40})

	snapshot(t, testOverlappingWindowsSnapshot)
}
