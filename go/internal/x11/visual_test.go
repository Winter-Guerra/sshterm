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
	_ "embed"
	"encoding/binary"
	"fmt"
	"image"
	"image/color"
	"syscall/js"
	"testing"
	"time"
)

func getCanvasData(t *testing.T, winID xID, x, y, w, h int) *image.RGBA {
	t.Helper()
	// A small delay to ensure the canvas has been painted.
	time.Sleep(100 * time.Millisecond)

	doc := js.Global().Get("document")
	if !doc.Truthy() {
		t.Fatal("document not found")
	}
	canvas := doc.Call("querySelector", fmt.Sprintf("#x11-canvas-%d-%d", winID.client, winID.local))
	if !canvas.Truthy() {
		t.Fatalf("canvas #x11-canvas-%d-%d not found", winID.client, winID.local)
	}
	ctx := canvas.Call("getContext", "2d")
	if !ctx.Truthy() {
		t.Fatal("canvas context not found")
	}
	imgData := ctx.Call("getImageData", x, y, w, h).Get("data")
	if !imgData.Truthy() {
		t.Fatal("failed to get image data")
	}
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	js.CopyBytesToGo(img.Pix, imgData)
	return img
}

func getWindowBounds(t *testing.T, winID xID) image.Rectangle {
	t.Helper()
	doc := js.Global().Get("document")
	if !doc.Truthy() {
		t.Fatal("document not found")
	}
	div := doc.Call("querySelector", fmt.Sprintf("#x11-window-%d-%d", winID.client, winID.local))
	if !div.Truthy() {
		t.Fatalf("div #x11-window-%d-%d not found", winID.client, winID.local)
	}
	style := div.Get("style")
	if !style.Truthy() {
		t.Fatal("div style not found")
	}
	var x, y, w, h int
	fmt.Sscanf(style.Get("left").String(), "%dpx", &x)
	fmt.Sscanf(style.Get("top").String(), "%dpx", &y)
	fmt.Sscanf(style.Get("width").String(), "%dpx", &w)
	fmt.Sscanf(style.Get("height").String(), "%dpx", &h)
	return image.Rect(x, y, x+w, y+h)
}

func canvasExists(t *testing.T, winID xID) bool {
	t.Helper()
	doc := js.Global().Get("document")
	if !doc.Truthy() {
		t.Fatal("document not found")
	}
	return doc.Call("querySelector", fmt.Sprintf("#x11-canvas-%d-%d", winID.client, winID.local)).Truthy()
}

func assertRectangle(t *testing.T, img *image.RGBA, rect image.Rectangle, c color.Color) {
	t.Helper()
	for y := rect.Min.Y; y < rect.Max.Y; y++ {
		for x := rect.Min.X; x < rect.Max.X; x++ {
			if got := img.At(x, y); got != c {
				t.Errorf("at (%d, %d): got %v, want %v", x, y, got, c)
			}
		}
	}
}

func assertWindow(t *testing.T, got, want image.Rectangle) {
	t.Helper()
	if got != want {
		t.Errorf("got %v, want %v", got, want)
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

	img := getCanvasData(t, winID, 20, 20, 50, 40)
	assertRectangle(t, img, image.Rect(0, 0, 50, 40), color.Black)
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

	img := getCanvasData(t, winID, 20, 30, 80, 20)
	var foundPixel bool
	bounds := img.Bounds()
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			if _, _, _, a := img.At(x, y).RGBA(); a != 0 {
				foundPixel = true
				break
			}
		}
		if foundPixel {
			break
		}
	}
	if !foundPixel {
		t.Error("did not find any drawn text")
	}
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

	assertWindow(t, getWindowBounds(t, winID1), image.Rect(10, 10, 110, 90))
	img1 := getCanvasData(t, winID1, 20, 20, 50, 40)
	assertRectangle(t, img1, image.Rect(0, 0, 50, 40), color.Black)

	assertWindow(t, getWindowBounds(t, winID2), image.Rect(30, 30, 130, 110))
	img2 := getCanvasData(t, winID2, 20, 20, 50, 40)
	assertRectangle(t, img2, image.Rect(0, 0, 50, 40), color.Black)
}
