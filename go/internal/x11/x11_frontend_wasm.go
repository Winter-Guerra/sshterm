//go:build x11 && wasm

package x11

import (
	"encoding/binary"
	"fmt"
	"math"
	"strconv"
	"strings"
	"syscall/js"

	"github.com/c2FmZQ/sshterm/internal/jsutil"
)

type property struct {
	data     []byte
	typeAtom uint32
	format   uint32
}

type windowInfo struct {
	div             js.Value
	canvas          js.Value
	ctx             js.Value // 2D rendering context
	mouseEvents     map[string]js.Func
	focusEvent      js.Func
	blurEvent       js.Func
	keyDownEvent    js.Func
	keyUpEvent      js.Func
	zIndex          int
	properties      map[uint32]*property
	backgroundPixel uint32
	colormap        xID
	isTopLevel      bool

	titleBar      js.Value
	windowTitle   js.Value
	dragMouseDown js.Func
	dragMouseMove js.Func
	dragMouseUp   js.Func

	resizeHandles   map[string]js.Value
	resizeMouseDown js.Func
	resizeMouseMove js.Func
	resizeMouseUp   js.Func
}

type pixmapInfo struct {
	canvas  js.Value
	context js.Value
}

type fontInfo struct {
	x11Name string
	cssFont string // CSS font string, e.g., "12px monospace"
}

type cursorInfo struct {
	style  string
	source xID
	mask   xID
	x, y   uint16
}

type wasmX11Frontend struct {
	document         js.Value
	body             js.Value
	windows          map[xID]*windowInfo    // Map to store window elements (div)
	pixmaps          map[xID]*pixmapInfo    // Map to store pixmap elements (canvas)
	gcs              map[xID]GC             // Map to store graphics contexts (Go representation)
	fonts            map[xID]*fontInfo      // Map to store opened fonts
	cursors          map[xID]*cursorInfo    // Map to store cursor info
	focusedWindowID  xID                    // Track the currently focused window
	server           *x11Server             // To call back into the server for pointer updates
	canvasOperations []CanvasOperation      // Store canvas operations for testing
	atoms            map[string]uint32      // Map atom names to IDs
	nextAtomID       uint32                 // Next available atom ID
	cursorStyles     map[uint32]*cursorInfo // Map X11 cursor IDs to CSS cursor styles
}

func (w *wasmX11Frontend) initPredefinedAtoms() {
	w.atoms = map[string]uint32{
		"PRIMARY":             1,
		"SECONDARY":           2,
		"ARC":                 3,
		"ATOM":                4,
		"BITMAP":              5,
		"CARDINAL":            6,
		"COLORMAP":            7,
		"CURSOR":              8,
		"CUT_BUFFER0":         9,
		"CUT_BUFFER1":         10,
		"CUT_BUFFER2":         11,
		"CUT_BUFFER3":         12,
		"CUT_BUFFER4":         13,
		"CUT_BUFFER5":         14,
		"CUT_BUFFER6":         15,
		"CUT_BUFFER7":         16,
		"DRAWABLE":            17,
		"FONT":                18,
		"INTEGER":             19,
		"PIXMAP":              20,
		"POINT":               21,
		"RECTANGLE":           22,
		"RESOURCE_MANAGER":    23,
		"RGB_COLOR_MAP":       24,
		"RGB_BEST_MAP":        25,
		"RGB_BLUE_MAP":        26,
		"RGB_DEFAULT_MAP":     27,
		"RGB_GRAY_MAP":        28,
		"RGB_GREEN_MAP":       29,
		"RGB_RED_MAP":         30,
		"STRING":              31,
		"VISUALID":            32,
		"WINDOW":              33,
		"WM_COMMAND":          34,
		"WM_HINTS":            35,
		"WM_CLIENT_MACHINE":   36,
		"WM_ICON_NAME":        37,
		"WM_ICON_SIZE":        38,
		"WM_NAME":             39,
		"WM_NORMAL_HINTS":     40,
		"WM_SIZE_HINTS":       41,
		"WM_ZOOM_HINTS":       42,
		"MIN_SPACE":           43,
		"NORM_SPACE":          44,
		"MAX_SPACE":           45,
		"END_SPACE":           46,
		"SUPERSCRIPT_X":       47,
		"SUPERSCRIPT_Y":       48,
		"SUBSCRIPT_X":         49,
		"SUBSCRIPT_Y":         50,
		"UNDERLINE_POSITION":  51,
		"UNDERLINE_THICKNESS": 52,
		"STRIKEOUT_ASCENT":    53,
		"STRIKEOUT_DESCENT":   54,
		"ITALIC_ANGLE":        55,
		"X_HEIGHT":            56,
		"QUAD_WIDTH":          57,
		"WEIGHT":              58,
		"POINT_SIZE":          59,
		"RESOLUTION":          60,
		"COPYRIGHT":           61,
		"NOTICE":              62,
		"FONT_NAME":           63,
		"FAMILY_NAME":         64,
		"FULL_NAME":           65,
		"CAP_HEIGHT":          66,
		"WM_CLASS":            67,
		"WM_TRANSIENT_FOR":    68,
	}
	w.nextAtomID = 69
}

func newX11Frontend(logger Logger, s *x11Server) X11FrontendAPI {
	document := js.Global().Get("document")
	body := document.Get("body")
	frontend := &wasmX11Frontend{
		document:     document,
		body:         body,
		windows:      make(map[xID]*windowInfo),
		pixmaps:      make(map[xID]*pixmapInfo),
		gcs:          make(map[xID]GC),
		fonts:        make(map[xID]*fontInfo),
		cursors:      make(map[xID]*cursorInfo),
		server:       s,
		atoms:        make(map[string]uint32),
		nextAtomID:   1,
		cursorStyles: make(map[uint32]*cursorInfo),
	}
	frontend.initDefaultCursors()
	frontend.initCanvasOperations()
	frontend.initPredefinedAtoms()

	// Set initial root window size and add resize listener
	win := js.Global().Get("window")
	width := win.Get("innerWidth").Int()
	height := win.Get("innerHeight").Int()
	frontend.server.SetRootWindowSize(uint16(width), uint16(height))

	resizeHandler := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		newWidth := win.Get("innerWidth").Int()
		newHeight := win.Get("innerHeight").Int()
		frontend.server.SetRootWindowSize(uint16(newWidth), uint16(newHeight))
		return nil
	})
	win.Call("addEventListener", "resize", resizeHandler)

	return frontend
}

func (w *wasmX11Frontend) getForegroundColor(cmap xID, gc GC) (out string) {
	defer func() {
		debugf("getForegroundColor: cmap:%s gc=%+v %s", cmap, gc, out)
	}()
	r, g, b := w.GetRGBColor(cmap, gc.Foreground)
	return fmt.Sprintf("#%02x%02x%02x", r, g, b)
}

func (w *wasmX11Frontend) CreateWindow(xid xID, parent, x, y, width, height, depth, valueMask uint32, values WindowAttributes) {
	debugf("X11: createWindow xid=%s parent=%d x=%d y=%d width=%d height=%d depth=%d values=%+v", xid, parent, x, y, width, height, depth, values)

	windowDiv := w.document.Call("createElement", "div")
	windowDiv.Set("id", js.ValueOf(fmt.Sprintf("x11-window-%s", xid)))
	style := windowDiv.Get("style")
	style.Set("position", "absolute")
	style.Set("width", js.ValueOf(fmt.Sprintf("%dpx", width)))
	style.Set("border", "1px solid black")
	style.Set("zIndex", "100")      // Ensure it's on top
	style.Set("overflow", "hidden") // Hide overflow during resize

	// Create canvas first so it can be referenced in handlers, but don't append yet.
	canvas := w.document.Call("createElement", "canvas")
	canvas.Set("id", js.ValueOf(fmt.Sprintf("x11-canvas-%s", xid)))
	canvas.Set("width", width)
	canvas.Set("height", height)
	canvas.Get("style").Set("display", "block")

	isTopLevel := parent == w.server.rootWindowID()
	var titleBarHeight int
	var titleBar, windowTitleSpan js.Value
	var dragMouseDown, dragMouseMove, dragMouseUp js.Func
	var resizeHandlesMap map[string]js.Value
	var resizeMouseDown, resizeMouseMove, resizeMouseUp js.Func
	var titleBarStyle js.Value

	// These need to be accessible by blurEvent
	var isDragging bool
	var isResizing bool

	if isTopLevel {
		style.Set("backgroundColor", "white")

		titleBarHeight = 20

		// Title bar
		titleBar = w.document.Call("createElement", "div")
		titleBar.Set("id", js.ValueOf(fmt.Sprintf("x11-titlebar-%s", xid)))
		titleBarStyle = titleBar.Get("style")
		titleBarStyle.Set("height", "20px")
		titleBarStyle.Set("backgroundColor", "#333")
		titleBarStyle.Set("color", "white")
		titleBarStyle.Set("fontFamily", "monospace")
		titleBarStyle.Set("fontSize", "14px")
		titleBarStyle.Set("lineHeight", "20px")
		titleBarStyle.Set("paddingLeft", "5px")
		titleBarStyle.Set("cursor", "move")
		titleBarStyle.Set("userSelect", "none")
		windowDiv.Call("appendChild", titleBar)

		// Window title text
		windowTitleSpan = w.document.Call("createElement", "span")
		windowTitleSpan.Set("id", js.ValueOf(fmt.Sprintf("x11-window-title-%s", xid)))
		windowTitleSpan.Set("textContent", fmt.Sprintf("Window %s", xid)) // Default title
		titleBar.Call("appendChild", windowTitleSpan)

		// Close button
		closeButton := w.document.Call("createElement", "button")
		closeButton.Set("textContent", "X")
		closeButton.Set("ariaLabel", "Close Window")
		closeButtonStyle := closeButton.Get("style")
		closeButtonStyle.Set("float", "right")
		closeButtonStyle.Set("backgroundColor", "#f00")
		closeButtonStyle.Set("color", "white")
		closeButtonStyle.Set("border", "none")
		closeButtonStyle.Set("height", "100%")
		closeButtonStyle.Set("cursor", "pointer")
		closeButton.Call("addEventListener", "click", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			w.CloseWindow(xid)
			return nil
		}))
		titleBar.Call("appendChild", closeButton)

		// Dragging functionality
		var dragOffsetX, dragOffsetY int

		dragMouseMove = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			event := args[0]
			if isDragging {
				newX := event.Get("clientX").Int() - dragOffsetX
				newY := event.Get("clientY").Int() - dragOffsetY
				style.Set("left", js.ValueOf(fmt.Sprintf("%dpx", newX)))
				style.Set("top", js.ValueOf(fmt.Sprintf("%dpx", newY)))
			}
			return nil
		})

		dragMouseUp = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			isDragging = false
			titleBarStyle.Set("cursor", "move")
			w.document.Call("removeEventListener", "mousemove", dragMouseMove)
			w.document.Call("removeEventListener", "mouseup", dragMouseUp)
			w.SendConfigureAndExposeEvent(xid, int16(windowDiv.Get("offsetLeft").Int()), int16(windowDiv.Get("offsetTop").Int()), uint16(canvas.Get("width").Int()), uint16(canvas.Get("height").Int()))
			return nil
		})

		dragMouseDown = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			event := args[0]
			isDragging = true
			dragOffsetX = event.Get("clientX").Int() - windowDiv.Get("offsetLeft").Int()
			dragOffsetY = event.Get("clientY").Int() - windowDiv.Get("offsetTop").Int()
			titleBarStyle.Set("cursor", "grabbing")
			w.document.Call("addEventListener", "mousemove", dragMouseMove)
			w.document.Call("addEventListener", "mouseup", dragMouseUp)
			return nil
		})

		titleBar.Call("addEventListener", "mousedown", dragMouseDown)

		// Resizing functionality
		var resizeStartX, resizeStartY, resizeStartWidth, resizeStartHeight, resizeStartLeft, resizeStartTop int
		var resizeHandle string

		resizeHandlesMap = make(map[string]js.Value)
		handleNames := []string{"n", "s", "e", "w", "nw", "ne", "sw", "se"}
		for _, name := range handleNames {
			handle := w.document.Call("createElement", "div")
			handle.Set("className", "resize-handle "+name)
			handleStyle := handle.Get("style")
			handleStyle.Set("position", "absolute")
			handleStyle.Set("backgroundColor", "rgba(0, 0, 0, 0)") // Transparent
			handleStyle.Set("zIndex", "101")
			const handleSize = 8 // pixels

			switch name {
			case "n":
				handleStyle.Set("height", fmt.Sprintf("%dpx", handleSize))
				handleStyle.Set("left", fmt.Sprintf("%dpx", handleSize))
				handleStyle.Set("right", fmt.Sprintf("%dpx", handleSize))
				handleStyle.Set("top", fmt.Sprintf("-%dpx", handleSize/2))
				handleStyle.Set("cursor", "ns-resize")
			case "s":
				handleStyle.Set("height", fmt.Sprintf("%dpx", handleSize))
				handleStyle.Set("left", fmt.Sprintf("%dpx", handleSize))
				handleStyle.Set("right", fmt.Sprintf("%dpx", handleSize))
				handleStyle.Set("bottom", fmt.Sprintf("-%dpx", handleSize/2))
				handleStyle.Set("cursor", "ns-resize")
			case "e":
				handleStyle.Set("width", fmt.Sprintf("%dpx", handleSize))
				handleStyle.Set("top", fmt.Sprintf("%dpx", handleSize))
				handleStyle.Set("bottom", fmt.Sprintf("%dpx", handleSize))
				handleStyle.Set("right", fmt.Sprintf("-%dpx", handleSize/2))
				handleStyle.Set("cursor", "ew-resize")
			case "w":
				handleStyle.Set("width", fmt.Sprintf("%dpx", handleSize))
				handleStyle.Set("top", fmt.Sprintf("%dpx", handleSize))
				handleStyle.Set("bottom", fmt.Sprintf("%dpx", handleSize))
				handleStyle.Set("left", fmt.Sprintf("-%dpx", handleSize/2))
				handleStyle.Set("cursor", "ew-resize")
			case "nw":
				handleStyle.Set("width", fmt.Sprintf("%dpx", handleSize))
				handleStyle.Set("height", fmt.Sprintf("%dpx", handleSize))
				handleStyle.Set("top", fmt.Sprintf("-%dpx", handleSize/2))
				handleStyle.Set("left", fmt.Sprintf("-%dpx", handleSize/2))
				handleStyle.Set("cursor", "nwse-resize")
			case "ne":
				handleStyle.Set("width", fmt.Sprintf("%dpx", handleSize))
				handleStyle.Set("height", fmt.Sprintf("%dpx", handleSize))
				handleStyle.Set("top", fmt.Sprintf("-%dpx", handleSize/2))
				handleStyle.Set("right", fmt.Sprintf("-%dpx", handleSize/2))
				handleStyle.Set("cursor", "nesw-resize")
			case "sw":
				handleStyle.Set("width", fmt.Sprintf("%dpx", handleSize))
				handleStyle.Set("height", fmt.Sprintf("%dpx", handleSize))
				handleStyle.Set("bottom", fmt.Sprintf("-%dpx", handleSize/2))
				handleStyle.Set("left", fmt.Sprintf("-%dpx", handleSize/2))
				handleStyle.Set("cursor", "nesw-resize")
			case "se":
				handleStyle.Set("width", fmt.Sprintf("%dpx", handleSize))
				handleStyle.Set("height", fmt.Sprintf("%dpx", handleSize))
				handleStyle.Set("bottom", fmt.Sprintf("-%dpx", handleSize/2))
				handleStyle.Set("right", fmt.Sprintf("-%dpx", handleSize/2))
				handleStyle.Set("cursor", "nwse-resize")
			}
			windowDiv.Call("appendChild", handle)
			resizeHandlesMap[name] = handle
		}

		resizeMouseMove = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			event := args[0]
			if !isResizing {
				return nil
			}

			currentX := event.Get("clientX").Int()
			currentY := event.Get("clientY").Int()

			deltaX := currentX - resizeStartX
			deltaY := currentY - resizeStartY

			newWidth := resizeStartWidth
			newHeight := resizeStartHeight
			newX := resizeStartLeft
			newY := resizeStartTop

			name := strings.TrimPrefix(resizeHandle, "resize-handle ")
			switch {
			case strings.Contains(name, "n"):
				newHeight = resizeStartHeight - deltaY
				newY = resizeStartTop + deltaY
			case strings.Contains(name, "s"):
				newHeight = resizeStartHeight + deltaY
			}
			switch {
			case strings.Contains(name, "w"):
				newWidth = resizeStartWidth - deltaX
				newX = resizeStartLeft + deltaX
			case strings.Contains(name, "e"):
				newWidth = resizeStartWidth + deltaX
			}

			// Minimum size
			if newWidth < 50 {
				newWidth = 50
			}
			if newHeight < 50 {
				newHeight = 50
			}

			style.Set("width", fmt.Sprintf("%dpx", newWidth))
			style.Set("height", fmt.Sprintf("%dpx", newHeight))
			style.Set("left", js.ValueOf(fmt.Sprintf("%dpx", newX)))
			style.Set("top", js.ValueOf(fmt.Sprintf("%dpx", newY)))

			canvas.Set("width", newWidth)
			canvas.Set("height", newHeight-20) // Adjust for title bar height

			return nil
		})

		resizeMouseUp = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			isResizing = false
			w.document.Call("removeEventListener", "mousemove", resizeMouseMove)
			w.document.Call("removeEventListener", "mouseup", resizeMouseUp)
			winInfo, ok := w.windows[xid]
			if !ok {
				return nil
			}
			w.SendConfigureAndExposeEvent(xid, int16(winInfo.div.Get("offsetLeft").Int()), int16(winInfo.div.Get("offsetTop").Int()), uint16(winInfo.canvas.Get("width").Int()), uint16(winInfo.canvas.Get("height").Int()))
			return nil
		})

		resizeMouseDown = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			event := args[0]
			isResizing = true
			resizeStartX = event.Get("clientX").Int()
			resizeStartY = event.Get("clientY").Int()
			resizeStartWidth = windowDiv.Get("offsetWidth").Int()
			resizeStartHeight = windowDiv.Get("offsetHeight").Int()
			resizeStartLeft = windowDiv.Get("offsetLeft").Int()
			resizeStartTop = windowDiv.Get("offsetTop").Int()
			resizeHandle = this.Get("className").String() // e.g., "resize-handle n"
			w.document.Call("addEventListener", "mousemove", resizeMouseMove)
			w.document.Call("addEventListener", "mouseup", resizeMouseUp)
			return nil
		})

		for _, handle := range resizeHandlesMap {
			handle.Call("addEventListener", "mousedown", resizeMouseDown)
		}
	}

	windowDiv.Call("appendChild", canvas)

	ctx := canvas.Call("getContext", "2d")

	var finalX, finalY uint32 = x, y
	var parentDiv js.Value = w.body

	if !isTopLevel {
		if parentInfo, ok := w.windows[xID{xid.client, parent}]; ok {
			parentDiv = parentInfo.div
			if parentInfo.isTopLevel {
				finalY = y + 20
			}
		}
	}
	style.Set("left", js.ValueOf(fmt.Sprintf("%dpx", finalX)))
	style.Set("top", js.ValueOf(fmt.Sprintf("%dpx", finalY)))
	style.Set("height", js.ValueOf(fmt.Sprintf("%dpx", height+uint32(titleBarHeight))))

	// Create and store event listeners
	mouseEvents := make(map[string]js.Func)
	mouseEvents["mousedown"] = w.mouseEventHandler(xid, "mousedown")
	mouseEvents["mouseup"] = w.mouseEventHandler(xid, "mouseup")
	mouseEvents["mousemove"] = w.mouseEventHandler(xid, "mousemove")
	mouseEvents["wheel"] = w.mouseEventHandler(xid, "wheel")

	keyDownEvent := w.keyboardEventHandler(xid, "keydown")
	keyUpEvent := w.keyboardEventHandler(xid, "keyup")

	focusEvent := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		debugf("X11: Window %s focused", xid)
		w.focusedWindowID = xid
		w.document.Call("addEventListener", "keydown", keyDownEvent)
		w.document.Call("addEventListener", "keyup", keyUpEvent)
		return nil
	})
	blurEvent := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		debugf("X11: Window %s blurred", xid)
		w.focusedWindowID = xID{}
		w.document.Call("removeEventListener", "keydown", keyDownEvent)
		w.document.Call("removeEventListener", "keyup", keyUpEvent)
		if isTopLevel {
			if isDragging {
				isDragging = false
				titleBarStyle.Set("cursor", "move")
				w.document.Call("removeEventListener", "mousemove", dragMouseMove)
				w.document.Call("removeEventListener", "mouseup", dragMouseUp)
			}
			if isResizing {
				isResizing = false
				w.document.Call("removeEventListener", "mousemove", resizeMouseMove)
				w.document.Call("removeEventListener", "mouseup", resizeMouseUp)
			}
		}
		return nil
	})

	// Attach mouse event listeners
	canvas.Call("addEventListener", "mousedown", mouseEvents["mousedown"])
	canvas.Call("addEventListener", "mouseup", mouseEvents["mouseup"])
	canvas.Call("addEventListener", "mousemove", mouseEvents["mousemove"])
	canvas.Call("addEventListener", "wheel", mouseEvents["wheel"])

	// Attach focus/blur event listeners
	windowDiv.Set("tabIndex", 0) // Make the div focusable
	windowDiv.Call("addEventListener", "focus", focusEvent)
	windowDiv.Call("addEventListener", "blur", blurEvent)

	// Store window info in the map
	w.windows[xid] = &windowInfo{
		div:             windowDiv,
		canvas:          canvas,
		ctx:             ctx,
		mouseEvents:     mouseEvents,
		focusEvent:      focusEvent,
		blurEvent:       blurEvent,
		keyDownEvent:    keyDownEvent, // Store for removal
		keyUpEvent:      keyUpEvent,   // Store for removal
		zIndex:          100,
		properties:      make(map[uint32]*property),
		isTopLevel:      isTopLevel,
		titleBar:        titleBar,
		windowTitle:     windowTitleSpan,
		dragMouseDown:   dragMouseDown,
		dragMouseMove:   dragMouseMove,
		dragMouseUp:     dragMouseUp,
		resizeHandles:   resizeHandlesMap,
		resizeMouseDown: resizeMouseDown,
		resizeMouseMove: resizeMouseMove,
		resizeMouseUp:   resizeMouseUp,
	}
	if values.Colormap != 0 {
		w.windows[xid].colormap = xID{client: xid.client, local: uint32(values.Colormap)}
	}

	parentDiv.Call("appendChild", windowDiv)

	w.recordOperation(CanvasOperation{
		Type: "createWindow",
		Args: []any{xid.local, parent, x, y, width, height, depth},
	})
}

func (w *wasmX11Frontend) DestroyWindow(wid xID) {
	w.destroyWindow(wid, true)
}

func (w *wasmX11Frontend) destroyWindow(wid xID, logit bool) {
	if winInfo, ok := w.windows[wid]; ok {
		// Remove event listeners from the document and window elements
		if winInfo.isTopLevel {
			winInfo.titleBar.Call("removeEventListener", "mousedown", winInfo.dragMouseDown)
			w.document.Call("removeEventListener", "mousemove", winInfo.dragMouseMove)
			w.document.Call("removeEventListener", "mouseup", winInfo.dragMouseUp)

			for _, handle := range winInfo.resizeHandles {
				handle.Call("removeEventListener", "mousedown", winInfo.resizeMouseDown)
			}
			w.document.Call("removeEventListener", "mousemove", winInfo.resizeMouseMove)
			w.document.Call("removeEventListener", "mouseup", winInfo.resizeMouseUp)
		}

		winInfo.canvas.Call("removeEventListener", "mousedown", winInfo.mouseEvents["mousedown"])
		winInfo.canvas.Call("removeEventListener", "mouseup", winInfo.mouseEvents["mouseup"])
		winInfo.canvas.Call("removeEventListener", "mousemove", winInfo.mouseEvents["mousemove"])
		winInfo.canvas.Call("removeEventListener", "wheel", winInfo.mouseEvents["wheel"])

		winInfo.div.Call("removeEventListener", "focus", winInfo.focusEvent)
		winInfo.div.Call("removeEventListener", "blur", winInfo.blurEvent)

		// If the window is focused, remove the keyboard listeners from the document
		if w.focusedWindowID == wid {
			w.document.Call("removeEventListener", "keydown", winInfo.keyDownEvent)
			w.document.Call("removeEventListener", "keyup", winInfo.keyUpEvent)
		}

		winInfo.div.Call("remove")
		// Release all js.Func objects to prevent memory leaks
		for _, fn := range winInfo.mouseEvents {
			fn.Release()
		}
		winInfo.focusEvent.Release()
		winInfo.blurEvent.Release()
		winInfo.keyDownEvent.Release() // Release keyboard event listeners
		winInfo.keyUpEvent.Release()   // Release keyboard event listeners

		if winInfo.isTopLevel {
			winInfo.dragMouseDown.Release()
			winInfo.dragMouseMove.Release()
			winInfo.dragMouseUp.Release()
			winInfo.resizeMouseDown.Release()
			winInfo.resizeMouseMove.Release()
			winInfo.resizeMouseUp.Release()
		}

		delete(w.windows, wid)
	}
	if logit {
		w.recordOperation(CanvasOperation{
			Type: "destroyWindow",
			Args: []any{wid.local},
		})
	}
}

func (w *wasmX11Frontend) CloseWindow(xid xID) {
	winInfo, ok := w.windows[xid]
	if !ok {
		return
	}

	wmProtocolsAtom := w.GetAtom(xid.client, "WM_PROTOCOLS")
	wmDeleteWindowAtom := w.GetAtom(xid.client, "WM_DELETE_WINDOW")

	supportsDelete := false
	if protocolsProp, ok := winInfo.properties[wmProtocolsAtom]; ok {
		// The property contains a list of atoms (CARD32).
		if protocolsProp.format == 32 {
			for i := 0; i < len(protocolsProp.data); i += 4 {
				atom := w.server.byteOrder.Uint32(protocolsProp.data[i : i+4])
				if atom == wmDeleteWindowAtom {
					supportsDelete = true
					break
				}
			}
		}
	}

	if supportsDelete {
		debugf("X11: Sending WM_DELETE_WINDOW ClientMessage to window %s", xid)
		var data [20]byte
		w.server.byteOrder.PutUint32(data[0:4], wmDeleteWindowAtom)
		// The second element is a timestamp, which we can leave as 0 for now.
		w.server.byteOrder.PutUint32(data[4:8], 0) // Timestamp
		w.server.SendClientMessageEvent(xid, wmProtocolsAtom, data)
	} else {
		debugf("X11: WM_DELETE_WINDOW not supported for window %s, destroying directly", xid)
		w.destroyWindow(xid, false)
	}
}

func (w *wasmX11Frontend) MapWindow(wid xID) {
	if winInfo, ok := w.windows[wid]; ok {
		winInfo.div.Get("style").Set("display", "block")
	}
	w.recordOperation(CanvasOperation{
		Type: "mapWindow",
		Args: []any{wid.local},
	})
}
func (w *wasmX11Frontend) UnmapWindow(wid xID) {
	if winInfo, ok := w.windows[wid]; ok {
		winInfo.div.Get("style").Set("display", "none")
	}
	w.recordOperation(CanvasOperation{
		Type: "unmapWindow",
		Args: []any{wid.local},
	})
}

func (w *wasmX11Frontend) ConfigureWindow(xid xID, valueMask uint16, values []uint32) {
	const (
		CWX           = 1 << 0
		CWY           = 1 << 1
		CWWidth       = 1 << 2
		CWHeight      = 1 << 3
		CWBorderWidth = 1 << 4
		CWSibling     = 1 << 5
		CWStackMode   = 1 << 6
	)
	debugf("X11: configureWindow id=%s valueMask=%d values=%v", xid, valueMask, values)
	if winInfo, ok := w.windows[xid]; ok {
		style := winInfo.div.Get("style")
		var valueIndex int
		if valueMask&CWX != 0 {
			style.Set("left", fmt.Sprintf("%dpx", values[valueIndex]))
			valueIndex++
		}
		if valueMask&CWY != 0 {
			style.Set("top", fmt.Sprintf("%dpx", values[valueIndex]))
			valueIndex++
		}
		if valueMask&CWWidth != 0 {
			style.Set("width", fmt.Sprintf("%dpx", values[valueIndex]))
			winInfo.canvas.Set("width", values[valueIndex])
			valueIndex++
		}
		if valueMask&CWHeight != 0 {
			style.Set("height", fmt.Sprintf("%dpx", values[valueIndex]))
			winInfo.canvas.Set("height", values[valueIndex])
			valueIndex++
		}
		if valueMask&CWSibling != 0 {
			// Sibling is not implemented yet
			valueIndex++
		}
		if valueMask&CWStackMode != 0 {
			stackMode := values[valueIndex]
			switch stackMode {
			case 0: // Above
				winInfo.zIndex = w.getHighestZIndex() + 1
			case 1: // Below
				winInfo.zIndex = w.getLowestZIndex() - 1
			}
			style.Set("zIndex", fmt.Sprintf("%d", winInfo.zIndex))
			valueIndex++
		}
	}
	w.recordOperation(CanvasOperation{
		Type: "configureWindow",
		Args: []any{xid.local, valueMask, values},
	})
}

func (w *wasmX11Frontend) getHighestZIndex() int {
	highest := 0
	for _, winInfo := range w.windows {
		if winInfo.zIndex > highest {
			highest = winInfo.zIndex
		}
	}
	return highest
}

func (w *wasmX11Frontend) getLowestZIndex() int {
	lowest := 0
	for _, winInfo := range w.windows {
		if winInfo.zIndex < lowest {
			lowest = winInfo.zIndex
		}
	}
	return lowest
}
func (w *wasmX11Frontend) CreateGC(xid xID, valueMask uint32, values GC) {
	debugf("X11: createGC id=%s gc=%+v", xid, values)
	w.gcs[xid] = values
	w.recordOperation(CanvasOperation{
		Type: "createGC",
		Args: []any{xid.local, valueMask, values},
	})
}

func (w *wasmX11Frontend) ChangeGC(xid xID, valueMask uint32, gc GC) {
	debugf("X11: changeGC id=%s valueMask=%d gc=%+v", xid, valueMask, gc)
	existingGC, ok := w.gcs[xid]
	if !ok {
		// This shouldn't happen, but if it does, treat it as a CreateGC
		w.gcs[xid] = gc
		w.recordOperation(CanvasOperation{
			Type: "createGC",
			Args: []any{xid.local},
		})
		return
	}

	if valueMask&GCFunction != 0 {
		existingGC.Function = gc.Function
	}
	if valueMask&GCPlaneMask != 0 {
		existingGC.PlaneMask = gc.PlaneMask
	}
	if valueMask&GCForeground != 0 {
		existingGC.Foreground = gc.Foreground
	}
	if valueMask&GCBackground != 0 {
		existingGC.Background = gc.Background
	}
	if valueMask&GCLineWidth != 0 {
		existingGC.LineWidth = gc.LineWidth
	}
	if valueMask&GCLineStyle != 0 {
		existingGC.LineStyle = gc.LineStyle
	}
	if valueMask&GCCapStyle != 0 {
		existingGC.CapStyle = gc.CapStyle
	}
	if valueMask&GCJoinStyle != 0 {
		existingGC.JoinStyle = gc.JoinStyle
	}
	if valueMask&GCFillStyle != 0 {
		existingGC.FillStyle = gc.FillStyle
	}
	if valueMask&GCFillRule != 0 {
		existingGC.FillRule = gc.FillRule
	}
	if valueMask&GCTile != 0 {
		existingGC.Tile = gc.Tile
	}
	if valueMask&GCStipple != 0 {
		existingGC.Stipple = gc.Stipple
	}
	if valueMask&GCTileStipXOrigin != 0 {
		existingGC.TileStipXOrigin = gc.TileStipXOrigin
	}
	if valueMask&GCTileStipYOrigin != 0 {
		existingGC.TileStipYOrigin = gc.TileStipYOrigin
	}
	if valueMask&GCFont != 0 {
		existingGC.Font = gc.Font
	}
	if valueMask&GCSubwindowMode != 0 {
		existingGC.SubwindowMode = gc.SubwindowMode
	}
	if valueMask&GCGraphicsExposures != 0 {
		existingGC.GraphicsExposures = gc.GraphicsExposures
	}
	if valueMask&GCClipXOrigin != 0 {
		existingGC.ClipXOrigin = gc.ClipXOrigin
	}
	if valueMask&GCClipYOrigin != 0 {
		existingGC.ClipYOrigin = gc.ClipYOrigin
	}
	if valueMask&GCClipMask != 0 {
		existingGC.ClipMask = gc.ClipMask
	}
	if valueMask&GCDashOffset != 0 {
		existingGC.DashOffset = gc.DashOffset
	}
	if valueMask&GCDashes != 0 {
		existingGC.Dashes = gc.Dashes
	}
	if valueMask&GCArcMode != 0 {
		existingGC.ArcMode = gc.ArcMode
	}

	w.gcs[xid] = existingGC
	w.recordOperation(CanvasOperation{
		Type: "changeGC",
		Args: []any{xid.local, valueMask},
	})
}

func (w *wasmX11Frontend) ChangeProperty(xid xID, p, typ, format uint32, data []byte) {
	debugf("X11: changeProperty id=%s property=%d type=%d format=%d data=%s", xid, p, typ, format, string(data))
	if winInfo, ok := w.windows[xid]; ok {
		winInfo.properties[p] = &property{data: data, typeAtom: typ, format: format}

		wmNameAtom := w.GetAtom(xid.client, "WM_NAME")
		netWmNameAtom := w.GetAtom(xid.client, "_NET_WM_NAME")

		if p == wmNameAtom || p == netWmNameAtom {
			title := string(data)
			// Set HTML title attribute for tooltip
			winInfo.div.Set("title", title)
			// Set the text in the title bar, if it exists
			if !winInfo.windowTitle.IsUndefined() {
				winInfo.windowTitle.Set("textContent", title)
			}
			debugf("X11: Window %s title set to: %s", xid, title)
		}
	}
	w.recordOperation(CanvasOperation{
		Type: "changeProperty",
		Args: []any{xid.local, p, typ, format, string(data)},
	})
}

func (w *wasmX11Frontend) PutImage(drawable xID, gcID xID, format uint8, width, height uint16, dstX, dstY int16, leftPad, depth uint8, imgData []byte) {
	gc, ok := w.gcs[gcID]
	if !ok {
		return
	}
	debugf("X11: putImage drawable=%s gc=%v format=%d width=%d height=%d dstX=%d dstY=%d leftPad=%d depth=%d data length=%d first 16 bytes of data: %x", drawable, gc, format, width, height, dstX, dstY, leftPad, depth, len(imgData), imgData[:min(len(imgData), 16)])

	var currentColormap xID
	var ctx js.Value
	winInfo, ok := w.windows[drawable]
	if ok {
		ctx = winInfo.ctx
		currentColormap = winInfo.colormap
	} else if pixmapInfo, ok := w.pixmaps[drawable]; ok {
		ctx = pixmapInfo.context
		// For pixmaps, use the default colormap of the screen
		currentColormap = xID{0, w.server.defaultColormap}
	} else {
		debugf("X11: PutImage on unknown drawable %s", drawable)
		return
	}

	if ctx.IsNull() || width == 0 || height == 0 {
		return
	}
	switch format {
	case 0: // Bitmap
		r, g, b := w.GetRGBColor(currentColormap, gc.Foreground)
		fgR, fgG, fgB := r, g, b

		r, g, b = w.GetRGBColor(currentColormap, gc.Background)
		bgR, bgG, bgB := r, g, b

		jsImgData := js.Global().Get("Uint8ClampedArray").New(int(width * height * 4))
		dataIndex := 0
		scanlineStride := (int(width) + int(leftPad) + 7) / 8

		for row := 0; row < int(height); row++ {
			scanlineOffset := row * scanlineStride
			for col := 0; col < int(width); col++ {
				bitPos := int(leftPad) + col
				byteIndex := scanlineOffset + (bitPos / 8)
				bitIndex := bitPos % 8

				if (imgData[byteIndex]>>(bitIndex))&1 == 1 {
					jsImgData.SetIndex(dataIndex, int(fgR))
					jsImgData.SetIndex(dataIndex+1, int(fgG))
					jsImgData.SetIndex(dataIndex+2, int(fgB))
				} else {
					jsImgData.SetIndex(dataIndex, int(bgR))
					jsImgData.SetIndex(dataIndex+1, int(bgG))
					jsImgData.SetIndex(dataIndex+2, int(bgB))
				}
				jsImgData.SetIndex(dataIndex+3, 255) // Alpha
				dataIndex += 4
			}
		}
		imageData := js.Global().Get("ImageData").New(jsImgData, width, height)
		ctx.Call("putImageData", imageData, dstX, dstY)

	case 1: // XYPixmap
		jsImgData := js.Global().Get("Uint8ClampedArray").New(int(width * height * 4))
		pixelValues := make([]uint32, width*height)
		scanlineStride := (int(width) + 7) / 8

		for d := 0; d < int(depth); d++ {
			plane := imgData[d*int(height)*scanlineStride:]
			for y := 0; y < int(height); y++ {
				for x := 0; x < int(width); x++ {
					byteIndex := (y*scanlineStride + x/8)
					bitIndex := x % 8
					if (plane[byteIndex]>>bitIndex)&1 != 0 {
						pixelValues[y*int(width)+x] |= (1 << d)
					}
				}
			}
		}

		for i, pixel := range pixelValues {
			r, g, b := w.GetRGBColor(currentColormap, pixel)
			jsImgData.SetIndex(i*4+0, int(r))
			jsImgData.SetIndex(i*4+1, int(g))
			jsImgData.SetIndex(i*4+2, int(b))
			jsImgData.SetIndex(i*4+3, 255)
		}

		imageData := js.Global().Get("ImageData").New(jsImgData, width, height)
		ctx.Call("putImageData", imageData, dstX, dstY)

	case 2: // ZPixmap
		jsImgData := js.Global().Get("Uint8ClampedArray").New(int(width * height * 4))
		// Assuming depth 24, 32bpp, BGRX byte order
		for i := 0; i < int(width*height); i++ {
			if (i*4 + 2) < len(imgData) {
				jsImgData.SetIndex(i*4+0, int(imgData[i*4+2])) // R
				jsImgData.SetIndex(i*4+1, int(imgData[i*4+1])) // G
				jsImgData.SetIndex(i*4+2, int(imgData[i*4+0])) // B
				jsImgData.SetIndex(i*4+3, 255)                 // A
			}
		}
		imageData := js.Global().Get("ImageData").New(jsImgData, width, height)
		ctx.Call("putImageData", imageData, dstX, dstY)
	}

	w.recordOperation(CanvasOperation{
		Type: "putImage",
		Args: []any{drawable.local, gc, dstX, dstY, width, height, leftPad, format, len(imgData)},
	})
}

func (w *wasmX11Frontend) applyGC(ctx js.Value, colormap xID, gcID xID) {
	gc, ok := w.gcs[gcID]
	if !ok {
		return
	}
	// function (logic op)
	// https://developer.mozilla.org/en-US/docs/Web/API/CanvasRenderingContext2D/globalCompositeOperation
	switch gc.Function {
	case FunctionClear: // 0
		ctx.Set("globalCompositeOperation", "destination-out") // Essentially, clear the destination where the source is drawn
	case FunctionAnd: // src AND dst
		ctx.Set("globalCompositeOperation", "source-in")
	case FunctionAndReverse: // src AND NOT dst
		ctx.Set("globalCompositeOperation", "destination-in")
	case FunctionCopy: // src
		ctx.Set("globalCompositeOperation", "source-over")
	case FunctionAndInverted: // NOT src AND dst
		ctx.Set("globalCompositeOperation", "source-out")
	case FunctionNoOp: // dst
		// No operation, so we do nothing to the context
	case FunctionXor: // src XOR dst
		ctx.Set("globalCompositeOperation", "xor")
	case FunctionOr: // src OR dst
		ctx.Set("globalCompositeOperation", "lighter")
	case FunctionNor: // NOT src AND NOT dst
		// No direct equivalent, difference might be visually interesting but incorrect
	case FunctionEquiv: // NOT src XOR dst
		// No direct equivalent
	case FunctionInvert: // NOT dst
		ctx.Set("globalCompositeOperation", "difference")
	case FunctionOrReverse: // src OR NOT dst
		ctx.Set("globalCompositeOperation", "destination-over")
	case FunctionCopyInverted: // NOT src
		// No direct equivalent
	case FunctionOrInverted: // NOT src OR dst
		ctx.Set("globalCompositeOperation", "destination-out")
	case FunctionNand: // NOT src OR NOT dst
		// No direct equivalent
	case FunctionSet: // 1
		// No direct equivalent to setting all bits to 1, but 'source-over' with a white fill would be close for a specific color
	}
	// foreground
	color := w.getForegroundColor(colormap, gc)
	ctx.Set("strokeStyle", color)
	ctx.Set("fillStyle", color)

	// line-width
	ctx.Set("lineWidth", gc.LineWidth)

	// line-style, cap-style, join-style
	switch gc.LineStyle {
	case LineStyleOnOffDash, LineStyleDoubleDash:
		// Dashes are handled in applyDashes
	case LineStyleSolid:
		ctx.Call("setLineDash", js.Global().Get("Array").New())
	}

	switch gc.CapStyle {
	case CapStyleNotLast, CapStyleButt:
		ctx.Set("lineCap", "butt")
	case CapStyleRound:
		ctx.Set("lineCap", "round")
	case CapStyleProjecting:
		ctx.Set("lineCap", "square")
	}

	switch gc.JoinStyle {
	case JoinStyleMiter:
		ctx.Set("lineJoin", "miter")
	case JoinStyleRound:
		ctx.Set("lineJoin", "round")
	case JoinStyleBevel:
		ctx.Set("lineJoin", "bevel")
	}

	// fill-style
	switch gc.FillStyle {
	case FillStyleSolid:
		ctx.Set("fillStyle", color)
	case FillStyleTiled:
		if tilePixmap, ok := w.pixmaps[xID{local: gc.Tile}]; ok {
			pattern := ctx.Call("createPattern", tilePixmap.canvas, "repeat")
			ctx.Set("fillStyle", pattern)
		}
	case FillStyleStippled:
		if stipplePixmap, ok := w.pixmaps[xID{local: gc.Stipple}]; ok {
			stippleCanvas := w.document.Call("createElement", "canvas")
			stippleCanvas.Set("width", stipplePixmap.canvas.Get("width"))
			stippleCanvas.Set("height", stipplePixmap.canvas.Get("height"))
			stippleCtx := stippleCanvas.Call("getContext", "2d")

			stippleCtx.Set("fillStyle", color)
			stippleCtx.Call("fillRect", 0, 0, stippleCanvas.Get("width"), stippleCanvas.Get("height"))
			stippleCtx.Set("globalCompositeOperation", "destination-in")
			stippleCtx.Call("drawImage", stipplePixmap.canvas, 0, 0)

			pattern := ctx.Call("createPattern", stippleCanvas, "repeat")
			ctx.Set("fillStyle", pattern)
		}
	case FillStyleOpaqueStippled:
		if stipplePixmap, ok := w.pixmaps[xID{local: gc.Stipple}]; ok {
			stippleCanvas := w.document.Call("createElement", "canvas")
			stippleCanvas.Set("width", stipplePixmap.canvas.Get("width"))
			stippleCanvas.Set("height", stipplePixmap.canvas.Get("height"))
			stippleCtx := stippleCanvas.Call("getContext", "2d")

			r, g, b := w.GetRGBColor(colormap, gc.Background)
			bgColor := fmt.Sprintf("#%02x%02x%02x", r, g, b)
			stippleCtx.Set("fillStyle", bgColor)
			stippleCtx.Call("fillRect", 0, 0, stippleCanvas.Get("width"), stippleCanvas.Get("height"))

			stippleCtx.Set("fillStyle", color)
			stippleCtx.Call("fillRect", 0, 0, stippleCanvas.Get("width"), stippleCanvas.Get("height"))
			stippleCtx.Set("globalCompositeOperation", "destination-in")
			stippleCtx.Call("drawImage", stipplePixmap.canvas, 0, 0)

			pattern := ctx.Call("createPattern", stippleCanvas, "repeat")
			ctx.Set("fillStyle", pattern)
		}
	}

	// font
	if gc.Font != 0 {
		if font, ok := w.fonts[xID{local: gc.Font}]; ok {
			ctx.Set("font", font.cssFont)
		}
	}
	// clipping
	if gc.ClippingRectangles != nil && len(gc.ClippingRectangles) > 0 {
		ctx.Call("beginPath")
		for _, rect := range gc.ClippingRectangles {
			ctx.Call("rect", rect.X, rect.Y, rect.Width, rect.Height)
		}
		ctx.Call("clip")
	}

	// dashes
	if gc.DashPattern != nil && len(gc.DashPattern) > 0 {
		jsDashes := js.Global().Get("Array").New(len(gc.DashPattern))
		for i, v := range gc.DashPattern {
			jsDashes.SetIndex(i, js.ValueOf(v))
		}
		ctx.Call("setLineDash", jsDashes)
		ctx.Set("lineDashOffset", gc.DashOffset)
	}
}

func (w *wasmX11Frontend) PolyLine(drawable xID, gcID xID, points []uint32) {
	gc, ok := w.gcs[gcID]
	if !ok {
		return
	}
	debugf("X11: polyLine drawable=%s gc=%v points=%v", drawable, gc, points)
	color := "??????"
	var ctx js.Value
	var colormap xID
	if winInfo, ok := w.windows[drawable]; ok {
		ctx = winInfo.ctx
		colormap = winInfo.colormap
	} else if pixmapInfo, ok := w.pixmaps[drawable]; ok {
		ctx = pixmapInfo.context
		colormap = xID{0, w.server.defaultColormap}
	}

	if !ctx.IsUndefined() {
		ctx.Call("save")
		defer ctx.Call("restore")
		w.applyGC(ctx, colormap, gcID)
		color = w.getForegroundColor(colormap, gc)
		ctx.Call("beginPath")
		if len(points) >= 2 {
			ctx.Call("moveTo", points[0], points[1])
			for i := 2; i < len(points); i += 2 {
				ctx.Call("lineTo", points[i], points[i+1])
			}
		}
		ctx.Call("stroke")
	}
	w.recordOperation(CanvasOperation{
		Type:        "polyLine",
		Args:        []any{drawable.local, gc, points},
		StrokeStyle: color,
	})
}

func (w *wasmX11Frontend) PolyFillRectangle(drawable xID, gcID xID, rects []uint32) {
	gc, ok := w.gcs[gcID]
	if !ok {
		return
	}
	debugf("X11: polyFillRectangle drawable=%s gc=%v rects=%v GCFunction=%d", drawable, gc, rects, gc.Function)
	color := "??????"
	var ctx js.Value
	var colormap xID
	if winInfo, ok := w.windows[drawable]; ok {
		ctx = winInfo.ctx
		colormap = winInfo.colormap
	} else if pixmapInfo, ok := w.pixmaps[drawable]; ok {
		ctx = pixmapInfo.context
		colormap = xID{0, w.server.defaultColormap}
	}

	if !ctx.IsUndefined() {
		ctx.Call("save")
		defer ctx.Call("restore")
		w.applyGC(ctx, colormap, gcID)
		color = w.getForegroundColor(colormap, gc)
		for i := 0; i < len(rects); i += 4 {
			ctx.Call("fillRect", rects[i], rects[i+1], rects[i+2], rects[i+3])
		}
	}
	w.recordOperation(CanvasOperation{
		Type:      "polyFillRectangle",
		Args:      []any{drawable.local, gc, rects},
		FillStyle: color,
	})
}

func (w *wasmX11Frontend) FillPoly(drawable xID, gcID xID, points []uint32) {
	gc, ok := w.gcs[gcID]
	if !ok {
		return
	}
	debugf("X11: fillPoly drawable=%s gc=%v points=%v", drawable, gc, points)
	color := "??????"
	var ctx js.Value
	var colormap xID
	if winInfo, ok := w.windows[drawable]; ok {
		ctx = winInfo.ctx
		colormap = winInfo.colormap
	} else if pixmapInfo, ok := w.pixmaps[drawable]; ok {
		ctx = pixmapInfo.context
		colormap = xID{0, w.server.defaultColormap}
	}

	if !ctx.IsUndefined() {
		ctx.Call("save")
		defer ctx.Call("restore")
		w.applyGC(ctx, colormap, gcID)
		color = w.getForegroundColor(colormap, gc)
		ctx.Call("beginPath")
		if len(points) >= 2 {
			ctx.Call("moveTo", points[0], points[1])
			for i := 2; i < len(points); i += 2 {
				ctx.Call("lineTo", points[i], points[i+1])
			}
		}
		ctx.Call("closePath")
		var fillRule string
		switch gc.FillRule {
		case 0: // EvenOddRule
			fillRule = "evenodd"
		case 1: // WindingRule
			fillRule = "nonzero"
		default:
			fillRule = "nonzero"
		}
		ctx.Call("fill", fillRule)
	}
	w.recordOperation(CanvasOperation{
		Type:      "fillPoly",
		Args:      []any{drawable.local, gc, points},
		FillStyle: color,
	})
}

func (w *wasmX11Frontend) PolySegment(drawable xID, gcID xID, segments []uint32) {
	gc, ok := w.gcs[gcID]
	if !ok {
		return
	}
	debugf("X11: polySegment drawable=%s gc=%v segments=%v", drawable, gc, segments)
	color := "??????"
	var ctx js.Value
	var colormap xID
	if winInfo, ok := w.windows[drawable]; ok {
		ctx = winInfo.ctx
		colormap = winInfo.colormap
	} else if pixmapInfo, ok := w.pixmaps[drawable]; ok {
		ctx = pixmapInfo.context
		colormap = xID{0, w.server.defaultColormap}
	}

	if !ctx.IsUndefined() {
		ctx.Call("save")
		defer ctx.Call("restore")
		w.applyGC(ctx, colormap, gcID)
		color = w.getForegroundColor(colormap, gc)
		for i := 0; i < len(segments); i += 4 {
			ctx.Call("beginPath")
			ctx.Call("moveTo", segments[i], segments[i+1])
			ctx.Call("lineTo", segments[i+2], segments[i+3])
			ctx.Call("stroke")
		}
	}
	w.recordOperation(CanvasOperation{
		Type:        "polySegment",
		Args:        []any{drawable.local, gc, segments},
		StrokeStyle: color,
	})
}

func (w *wasmX11Frontend) PolyPoint(drawable xID, gcID xID, points []uint32) {
	gc, ok := w.gcs[gcID]
	if !ok {
		return
	}
	debugf("X11: polyPoint drawable=%s gc=%v points=%v", drawable, gc, points)
	color := "??????"
	var ctx js.Value
	var colormap xID
	if winInfo, ok := w.windows[drawable]; ok {
		ctx = winInfo.ctx
		colormap = winInfo.colormap
	} else if pixmapInfo, ok := w.pixmaps[drawable]; ok {
		ctx = pixmapInfo.context
		colormap = xID{0, w.server.defaultColormap}
	}

	if !ctx.IsUndefined() {
		ctx.Call("save")
		defer ctx.Call("restore")
		w.applyGC(ctx, colormap, gcID)
		color = w.getForegroundColor(colormap, gc)
		for i := 0; i < len(points); i += 2 {
			ctx.Call("fillRect", points[i], points[i+1], 1, 1)
		}
	}
	w.recordOperation(CanvasOperation{
		Type:      "polyPoint",
		Args:      []any{drawable.local, gc, points},
		FillStyle: color,
	})
}

func (w *wasmX11Frontend) PolyRectangle(drawable xID, gcID xID, rects []uint32) {
	gc, ok := w.gcs[gcID]
	if !ok {
		return
	}
	debugf("X11: polyRectangle drawable=%s gc=%v rects=%v", drawable, gc, rects)
	color := "??????"
	var ctx js.Value
	var colormap xID
	if winInfo, ok := w.windows[drawable]; ok {
		ctx = winInfo.ctx
		colormap = winInfo.colormap
	} else if pixmapInfo, ok := w.pixmaps[drawable]; ok {
		ctx = pixmapInfo.context
		colormap = xID{0, w.server.defaultColormap}
	}

	if !ctx.IsUndefined() {
		ctx.Call("save")
		defer ctx.Call("restore")
		w.applyGC(ctx, colormap, gcID)
		color = w.getForegroundColor(colormap, gc)
		for i := 0; i < len(rects); i += 4 {
			ctx.Call("strokeRect", rects[i], rects[i+1], rects[i+2], rects[i+3])
		}
	}
	w.recordOperation(CanvasOperation{
		Type:        "polyRectangle",
		Args:        []any{drawable.local, gc, rects},
		StrokeStyle: color,
	})
}

func (w *wasmX11Frontend) PolyArc(drawable xID, gcID xID, arcs []uint32) {
	gc, ok := w.gcs[gcID]
	if !ok {
		return
	}
	debugf("X11: polyArc drawable=%s gc=%v arcs=%v", drawable, gc, arcs)
	color := "??????"
	var ctx js.Value
	var colormap xID
	if winInfo, ok := w.windows[drawable]; ok {
		ctx = winInfo.ctx
		colormap = winInfo.colormap
	} else if pixmapInfo, ok := w.pixmaps[drawable]; ok {
		ctx = pixmapInfo.context
		colormap = xID{0, w.server.defaultColormap}
	}

	if !ctx.IsUndefined() {
		ctx.Call("save")
		defer ctx.Call("restore")
		w.applyGC(ctx, colormap, gcID)
		color = w.getForegroundColor(colormap, gc)
		for i := 0; i < len(arcs); i += 6 {
			ctx.Call("beginPath")
			// X11 angles are in 1/64th degrees, clockwise. Canvas angles are in radians, clockwise.
			// Start angle: arcs[i+4] / 64 * (Math.PI / 180)
			// End angle: (arcs[i+4] + arcs[i+5]) / 64 * (Math.PI / 180)
			startAngle := float64(arcs[i+4]) / 64 * (math.Pi / 180)
			endAngle := float64(arcs[i+4]+arcs[i+5]) / 64 * (math.Pi / 180)
			rx := uint32(arcs[i+2] / 2)
			ry := uint32(arcs[i+3] / 2)
			x := uint32(arcs[i] + rx)
			y := uint32(arcs[i+1] + ry)
			ctx.Call("ellipse", x, y, rx, ry, 0, startAngle, endAngle)
			ctx.Call("stroke")
		}
	}
	w.recordOperation(CanvasOperation{
		Type:        "polyArc",
		Args:        []any{drawable.local, gc, arcs},
		StrokeStyle: color,
	})
}

func (w *wasmX11Frontend) PolyFillArc(drawable xID, gcID xID, arcs []uint32) {
	gc, ok := w.gcs[gcID]
	if !ok {
		return
	}
	debugf("X11: polyFillArc drawable=%s gc=%v arcs=%v", drawable, gc, arcs)
	color := "??????"
	var ctx js.Value
	var colormap xID
	if winInfo, ok := w.windows[drawable]; ok {
		ctx = winInfo.ctx
		colormap = winInfo.colormap
	} else if pixmapInfo, ok := w.pixmaps[drawable]; ok {
		ctx = pixmapInfo.context
		colormap = xID{0, w.server.defaultColormap}
	}

	if !ctx.IsUndefined() {
		ctx.Call("save")
		defer ctx.Call("restore")
		w.applyGC(ctx, colormap, gcID)
		color = w.getForegroundColor(colormap, gc)
		for i := 0; i < len(arcs); i += 6 {
			ctx.Call("beginPath")
			startAngle := float64(arcs[i+4]) / 64 * (math.Pi / 180)
			endAngle := float64(arcs[i+4]+arcs[i+5]) / 64 * (math.Pi / 180)
			rx := uint32(arcs[i+2] / 2)
			ry := uint32(arcs[i+3] / 2)
			x := uint32(arcs[i] + rx)
			y := uint32(arcs[i+1] + ry)
			ctx.Call("ellipse", x, y, rx, ry, 0, startAngle, endAngle)
			if gc.ArcMode == 1 { // Pie
				ctx.Call("lineTo", x, y)
				ctx.Call("closePath")
			}

			var fillRule string
			switch gc.FillRule {
			case 0: // EvenOddRule
				fillRule = "evenodd"
			case 1: // WindingRule
				fillRule = "nonzero"
			}
			ctx.Call("fill", fillRule)
		}
	}
	w.recordOperation(CanvasOperation{
		Type:      "polyFillArc",
		Args:      []any{drawable.local, gc, arcs},
		FillStyle: color,
	})
}

func (w *wasmX11Frontend) ClearArea(drawable xID, x, y, width, height int32) {
	if width == 0 {
		width = int32(w.server.windows[drawable].width) - x
	}
	if height == 0 {
		height = int32(w.server.windows[drawable].height) - y
	}
	debugf("X11: clearArea drawable=%s x=%d y=%d width=%d height=%d", drawable, x, y, width, height)
	if winInfo, ok := w.windows[drawable]; ok {
		if !winInfo.canvas.IsNull() {
			// Clear the area with the window's background color
			var r, g, b uint8 = 0xff, 0xff, 0xff
			if w.server.windows[drawable].attributes.BackgroundPixelSet {
				// Get RGB color from server's colormap or visual
				r, g, b = w.GetRGBColor(winInfo.colormap, w.server.windows[drawable].attributes.BackgroundPixel)
			}
			bgColor := fmt.Sprintf("rgb(%d, %d, %d)", r, g, b)
			debugf("X11: ClearArea filling with fillStyle: %s", bgColor)
			winInfo.ctx.Set("fillStyle", bgColor)
			winInfo.ctx.Call("fillRect", x, y, width, height)
		}
	}
	w.recordOperation(CanvasOperation{
		Type: "clearArea",
		Args: []any{drawable.local, x, y, width, height},
	})
}

func (w *wasmX11Frontend) CopyArea(srcDrawable, dstDrawable xID, gcID xID, srcX, srcY, dstX, dstY, width, height int32) {
	gc, ok := w.gcs[gcID]
	if !ok {
		return
	}
	debugf("X11: copyArea src=%s dst=%s gc=%v srcX=%d srcY=%d dstX=%d dstY=%d width=%d height=%d", srcDrawable, dstDrawable, gc, srcX, srcY, dstX, dstY, width, height)
	var srcCanvas js.Value
	srcWinInfo, srcIsWindow := w.windows[srcDrawable]
	srcPixmapInfo, srcIsPixmap := w.pixmaps[srcDrawable]

	if srcIsWindow {
		srcCanvas = srcWinInfo.canvas
	} else if srcIsPixmap {
		srcCanvas = srcPixmapInfo.canvas
	} else {
		debugf("X11: CopyArea source drawable %d not found", srcDrawable)
		return
	}

	dstWinInfo, dstIsWindow := w.windows[dstDrawable]
	if !dstIsWindow {
		debugf("X11: CopyArea destination drawable %d not found or not a window", dstDrawable)
		return
	}

	if !srcCanvas.IsNull() && !dstWinInfo.canvas.IsNull() {
		dstWinInfo.ctx.Call("drawImage", srcCanvas, srcX, srcY, width, height, dstX, dstY, width, height)
	}
	w.recordOperation(CanvasOperation{
		Type: "copyArea",
		Args: []any{srcDrawable.local, dstDrawable.local, gc, srcX, srcY, dstX, dstY, width, height},
	})
}

func (w *wasmX11Frontend) CopyPlane(srcDrawable, dstDrawable xID, gcID xID, srcX, srcY, dstX, dstY, width, height, bitPlane int32) {
	gc, ok := w.gcs[gcID]
	if !ok {
		return
	}
	debugf("X11: copyPlane src=%s dst=%s gc=%v srcX=%d srcY=%d dstX=%d dstY=%d width=%d height=%d bitPlane=%d", srcDrawable, dstDrawable, gc, srcX, srcY, dstX, dstY, width, height, bitPlane)
	var srcCanvas js.Value
	srcWinInfo, srcIsWindow := w.windows[srcDrawable]
	srcPixmapInfo, srcIsPixmap := w.pixmaps[srcDrawable]

	if srcIsWindow {
		srcCanvas = srcWinInfo.canvas
	} else if srcIsPixmap {
		srcCanvas = srcPixmapInfo.canvas
	} else {
		debugf("X11: CopyPlane source drawable %d not found", srcDrawable)
		return
	}

	var dstCtx js.Value
	var currentColormap xID
	dstWinInfo, dstIsWindow := w.windows[dstDrawable]
	dstPixmapInfo, dstIsPixmap := w.pixmaps[dstDrawable]

	if dstIsWindow {
		dstCtx = dstWinInfo.ctx
		currentColormap = dstWinInfo.colormap
	} else if dstIsPixmap {
		dstCtx = dstPixmapInfo.context
		currentColormap = xID{0, w.server.defaultColormap}
	} else {
		debugf("X11: CopyPlane destination drawable %d not found", dstDrawable)
		return
	}

	if !srcCanvas.IsNull() && !dstCtx.IsUndefined() {
		// 1. Create a temporary canvas to prepare the source image.
		tempCanvas := w.document.Call("createElement", "canvas")
		tempCanvas.Set("width", width)
		tempCanvas.Set("height", height)
		tempCtx := tempCanvas.Call("getContext", "2d")

		// 2. Get the image data from the source drawable.
		srcImageData := srcCanvas.Call("getContext", "2d").Call("getImageData", srcX, srcY, width, height)
		srcData := srcImageData.Get("data")
		jsImgData := js.Global().Get("Uint8ClampedArray").New(int(width * height * 4))

		r, g, b := w.GetRGBColor(currentColormap, gc.Foreground)
		fgR, fgG, fgB := r, g, b

		r, g, b = w.GetRGBColor(currentColormap, gc.Background)
		bgR, bgG, bgB := r, g, b

		// 3. Iterate through the source image data and check the bitPlane.
		for i := 0; i < srcData.Length(); i += 4 {
			// The source is treated as a bitmap. We get the pixel value from the source,
			// and if the bit corresponding to bitPlane is set, we use the foreground color.
			// Otherwise, we use the background color.
			pixelValue := uint32(srcData.Index(i).Int()) | (uint32(srcData.Index(i+1).Int()) << 8) | (uint32(srcData.Index(i+2).Int()) << 16)

			// 4. Populate the temporary canvas with foreground or background color.
			if (pixelValue & uint32(bitPlane)) != 0 {
				jsImgData.SetIndex(i+0, int(fgR))
				jsImgData.SetIndex(i+1, int(fgG))
				jsImgData.SetIndex(i+2, int(fgB))
				jsImgData.SetIndex(i+3, 255) // Alpha for foreground
			} else {
				jsImgData.SetIndex(i+0, int(bgR))
				jsImgData.SetIndex(i+1, int(bgG))
				jsImgData.SetIndex(i+2, int(bgB))
				jsImgData.SetIndex(i+3, 255) // Alpha for background
			}
		}

		newImageData := js.Global().Get("ImageData").New(jsImgData, width, height)
		tempCtx.Call("putImageData", newImageData, 0, 0)

		// 5. Apply the GC to the destination context.
		dstCtx.Call("save")
		defer dstCtx.Call("restore")
		w.applyGC(dstCtx, currentColormap, gcID)

		// 6. Draw the temporary canvas image to the destination drawable.
		dstCtx.Call("drawImage", tempCanvas, dstX, dstY)
	}
	w.recordOperation(CanvasOperation{
		Type: "copyPlane",
		Args: []any{srcDrawable.local, dstDrawable.local, gc, srcX, srcY, dstX, dstY, width, height, bitPlane},
	})
}
func (w *wasmX11Frontend) GetImage(drawable xID, x, y, width, height int32, format uint32) ([]byte, error) {
	if winInfo, ok := w.windows[drawable]; ok {
		if !winInfo.canvas.IsNull() {
			imageData := winInfo.ctx.Call("getImageData", x, y, width, height)
			data := imageData.Get("data") // Uint8ClampedArray
			byteSlice := make([]byte, data.Length())
			js.CopyBytesToGo(byteSlice, data)
			return byteSlice, nil
		}
	}
	return nil, fmt.Errorf("window or canvas not found for drawable %d", drawable)
}

func (w *wasmX11Frontend) ImageText8(drawable xID, gcID xID, x, y int32, text []byte) {
	gc, ok := w.gcs[gcID]
	if !ok {
		return
	}
	decodedTextForLog := js.Global().Get("TextDecoder").New().Call("decode", jsutil.Uint8ArrayFromBytes(text)).String()
	decodedTextForLog = strings.ReplaceAll(decodedTextForLog, "\x00", "") // Trim null terminators
	debugf("X11: imageText8 drawable=%s gc=%v x=%d y=%d text=%s", drawable, gc, x, y, decodedTextForLog)
	color := "??????"
	if winInfo, ok := w.windows[drawable]; ok {
		if !winInfo.canvas.IsNull() {
			winInfo.ctx.Call("save")
			defer winInfo.ctx.Call("restore")
			w.applyGC(winInfo.ctx, winInfo.colormap, gcID)
			color = w.getForegroundColor(winInfo.colormap, gc)
			winInfo.ctx.Set("fillStyle", color)

			// Get font from GC
			fontCSS := "12px monospace" // Default fallback
			if gc.Font != 0 {
				if font, ok := w.fonts[xID{drawable.client, gc.Font}]; ok {
					fontCSS = font.cssFont
				}
			}
			winInfo.ctx.Set("font", fontCSS)

			decodedText := js.Global().Get("TextDecoder").New().Call("decode", jsutil.Uint8ArrayFromBytes(text)).String()
			decodedText = strings.ReplaceAll(decodedText, "\x00", "") // Trim null terminators
			winInfo.ctx.Call("fillText", decodedText, x, y)
		}
	}
	w.recordOperation(CanvasOperation{
		Type:      "imageText8",
		Args:      []any{drawable.local, gc, x, y, decodedTextForLog},
		FillStyle: color,
	})
}

func (w *wasmX11Frontend) ImageText16(drawable xID, gcID xID, x, y int32, text []uint16) {
	gc, ok := w.gcs[gcID]
	if !ok {
		return
	}
	// Convert []uint16 to []byte for TextDecoder
	textBytes := make([]byte, len(text)*2)
	for i, r := range text {
		binary.LittleEndian.PutUint16(textBytes[i*2:], r)
	}
	decodedTextForLog := js.Global().Get("TextDecoder").New().Call("decode", jsutil.Uint8ArrayFromBytes(textBytes)).String()
	decodedTextForLog = strings.ReplaceAll(decodedTextForLog, "\x00", "") // Trim null terminators
	debugf("X11: imageText16 drawable=%s gc=%v x=%d y=%d text=%s", drawable, gc, x, y, decodedTextForLog)
	color := "??????"
	if winInfo, ok := w.windows[drawable]; ok {
		if !winInfo.canvas.IsNull() {
			winInfo.ctx.Call("save")
			defer winInfo.ctx.Call("restore")
			w.applyGC(winInfo.ctx, winInfo.colormap, gcID)
			color = w.getForegroundColor(winInfo.colormap, gc)
			winInfo.ctx.Set("fillStyle", color)

			// Get font from GC
			fontCSS := "12px monospace" // Default fallback
			if gc.Font != 0 {
				if font, ok := w.fonts[xID{drawable.client, gc.Font}]; ok {
					fontCSS = font.cssFont
				}
			}
			winInfo.ctx.Set("font", fontCSS)

			decodedText := js.Global().Get("TextDecoder").New().Call("decode", jsutil.Uint8ArrayFromBytes(textBytes)).String()
			decodedText = strings.ReplaceAll(decodedText, "\x00", "") // Trim null terminators
			winInfo.ctx.Call("fillText", decodedText, x, y)
		}
	}
	w.recordOperation(CanvasOperation{
		Type:      "imageText16",
		Args:      []any{drawable.local, gc, x, y, decodedTextForLog},
		FillStyle: color,
	})
}

func (w *wasmX11Frontend) PolyText8(drawable xID, gcID xID, x, y int32, items []PolyTextItem) {
	gc, ok := w.gcs[gcID]
	if !ok {
		return
	}
	debugf("X11: polyText8 drawable=%s gc=%v x=%d y=%d items=%v", drawable, gc, x, y, items)
	color := "?????"
	winInfo, ok := w.windows[drawable]
	if !ok {
		return
	}
	if winInfo.canvas.IsNull() {
		return
	}
	winInfo.ctx.Call("save")
	defer winInfo.ctx.Call("restore")
	w.applyGC(winInfo.ctx, winInfo.colormap, gcID)
	color = w.getForegroundColor(winInfo.colormap, gc)
	winInfo.ctx.Set("fillStyle", color)
	// Get font from GC
	fontCSS := "12px monospace" // Default fallback
	if gc.Font != 0 {
		if font, ok := w.fonts[xID{drawable.client, gc.Font}]; ok {
			fontCSS = font.cssFont
		}
	}
	winInfo.ctx.Set("font", fontCSS)
	currentX := x
	var recordedItems []any
	for _, item := range items {
		switch it := item.(type) {
		case PolyText8String:
			currentX += int32(it.Delta)
			decodedText := js.Global().Get("TextDecoder").New().Call("decode", jsutil.Uint8ArrayFromBytes(it.Str)).String()
			decodedText = strings.ReplaceAll(decodedText, "\x00", "") // Trim null terminators
			winInfo.ctx.Call("fillText", decodedText, currentX, y)
			recordedItems = append(recordedItems, map[string]any{"delta": it.Delta, "text": decodedText})
			// The delta for the next item is relative to the start of the current item.
			// The browser's measureText is not reliable enough to calculate the next position.
		case PolyTextFont:
			if font, ok := w.fonts[xID{drawable.client, uint32(it.Font)}]; ok {
				winInfo.ctx.Set("font", font.cssFont)
				recordedItems = append(recordedItems, map[string]any{"font": it.Font})
			}
		}
	}
	w.recordOperation(CanvasOperation{
		Type:      "polyText8",
		Args:      []any{drawable.local, gc, x, y, recordedItems},
		FillStyle: color,
	})
}

func (w *wasmX11Frontend) SetDashes(gcID xID, dashOffset uint16, dashes []byte) {
	debugf("X11: setDashes gc=%s dashOffset=%d dashes=%v", gcID, dashOffset, dashes)
	if gc, ok := w.gcs[gcID]; ok {
		gc.DashOffset = uint32(dashOffset)
		gc.DashPattern = dashes
		w.gcs[gcID] = gc
	}
	w.recordOperation(CanvasOperation{
		Type: "setDashes",
		Args: []any{gcID.local, dashOffset, dashes},
	})
}

func (w *wasmX11Frontend) SetClipRectangles(gcID xID, clippingX, clippingY int16, rectangles []Rectangle, ordering byte) {
	debugf("X11: setClipRectangles gc=%s clippingX=%d clippingY=%d rectangles=%v ordering=%d", gcID, clippingX, clippingY, rectangles, ordering)
	if gc, ok := w.gcs[gcID]; ok {
		gc.ClipXOrigin = int32(clippingX)
		gc.ClipYOrigin = int32(clippingY)
		gc.ClippingRectangles = rectangles
		w.gcs[gcID] = gc
	}
	w.recordOperation(CanvasOperation{
		Type: "setClipRectangles",
		Args: []any{gcID.local, clippingX, clippingY, rectangles, ordering},
	})
}

func (w *wasmX11Frontend) RecolorCursor(cursorID xID, foreColor, backColor [3]uint16) {
	debugf("X11: RecolorCursor id=%s", cursorID)
	cursor, ok := w.cursorStyles[cursorID.local]
	if !ok {
		debugf("X11: RecolorCursor cursor %d not found", cursorID)
		return
	}

	w.CreateCursor(cursorID, cursor.source, cursor.mask, foreColor, backColor, cursor.x, cursor.y)
	w.recordOperation(CanvasOperation{
		Type: "recolorCursor",
		Args: []any{cursorID.local},
	})
}

func (w *wasmX11Frontend) SetPointerMapping(pMap []byte) (byte, error) {
	debugf("X11: SetPointerMapping (not implemented)")
	w.recordOperation(CanvasOperation{
		Type: "setPointerMapping",
		Args: []any{},
	})
	return 0, nil
}

func (w *wasmX11Frontend) GetPointerMapping() ([]byte, error) {
	debugf("X11: GetPointerMapping (not implemented)")
	w.recordOperation(CanvasOperation{
		Type: "getPointerMapping",
		Args: []any{},
	})
	return nil, nil
}

func (w *wasmX11Frontend) GetKeyboardMapping(firstKeyCode KeyCode, count byte) ([]uint32, error) {
	debugf("X11: GetKeyboardMapping (not implemented)")
	w.recordOperation(CanvasOperation{
		Type: "getKeyboardMapping",
		Args: []any{},
	})
	return nil, nil
}

func (w *wasmX11Frontend) ChangeKeyboardMapping(keyCodeCount byte, firstKeyCode KeyCode, keySymsPerKeyCode byte, keySyms []uint32) {
	debugf("X11: ChangeKeyboardMapping (not implemented)")
	w.recordOperation(CanvasOperation{
		Type: "changeKeyboardMapping",
		Args: []any{},
	})
}

func (w *wasmX11Frontend) ChangeKeyboardControl(valueMask uint32, values KeyboardControl) {
	debugf("X11: ChangeKeyboardControl (not implemented)")
	w.recordOperation(CanvasOperation{
		Type: "changeKeyboardControl",
		Args: []any{},
	})
}

func (w *wasmX11Frontend) GetKeyboardControl() (KeyboardControl, error) {
	debugf("X11: GetKeyboardControl (not implemented)")
	w.recordOperation(CanvasOperation{
		Type: "getKeyboardControl",
		Args: []any{},
	})
	return KeyboardControl{}, nil
}

func (w *wasmX11Frontend) SetScreenSaver(timeout, interval int16, preferBlank, allowExpose byte) {
	debugf("X11: SetScreenSaver (not implemented)")
	w.recordOperation(CanvasOperation{
		Type: "setScreenSaver",
		Args: []any{},
	})
}

func (w *wasmX11Frontend) GetScreenSaver() (timeout, interval int16, preferBlank, allowExpose byte, err error) {
	debugf("X11: GetScreenSaver (not implemented)")
	w.recordOperation(CanvasOperation{
		Type: "getScreenSaver",
		Args: []any{},
	})
	return 0, 0, 0, 0, nil
}

func (w *wasmX11Frontend) ChangeHosts(mode byte, host Host) {
	debugf("X11: ChangeHosts (not implemented)")
	w.recordOperation(CanvasOperation{
		Type: "changeHosts",
		Args: []any{},
	})
}

func (w *wasmX11Frontend) ListHosts() ([]Host, error) {
	debugf("X11: ListHosts (not implemented)")
	w.recordOperation(CanvasOperation{
		Type: "listHosts",
		Args: []any{},
	})
	return nil, nil
}

func (w *wasmX11Frontend) SetAccessControl(mode byte) {
	debugf("X11: SetAccessControl (not implemented)")
	w.recordOperation(CanvasOperation{
		Type: "setAccessControl",
		Args: []any{},
	})
}

func (w *wasmX11Frontend) SetCloseDownMode(mode byte) {
	debugf("X11: SetCloseDownMode (not implemented)")
	w.recordOperation(CanvasOperation{
		Type: "setCloseDownMode",
		Args: []any{},
	})
}

func (w *wasmX11Frontend) KillClient(resource uint32) {
	debugf("X11: KillClient (not implemented)")
	w.recordOperation(CanvasOperation{
		Type: "killClient",
		Args: []any{},
	})
}

func (w *wasmX11Frontend) RotateProperties(window xID, delta int16, atoms []Atom) {
	debugf("X11: RotateProperties (not implemented)")
	w.recordOperation(CanvasOperation{
		Type: "rotateProperties",
		Args: []any{},
	})
}

func (w *wasmX11Frontend) ForceScreenSaver(mode byte) {
	debugf("X11: ForceScreenSaver (not implemented)")
	w.recordOperation(CanvasOperation{
		Type: "forceScreenSaver",
		Args: []any{},
	})
}

func (w *wasmX11Frontend) SetModifierMapping(keyCodesPerModifier byte, keyCodes []KeyCode) (byte, error) {
	debugf("X11: SetModifierMapping (not implemented)")
	w.recordOperation(CanvasOperation{
		Type: "setModifierMapping",
		Args: []any{},
	})
	return 0, nil
}

func (w *wasmX11Frontend) GetModifierMapping() (keyCodesPerModifier byte, keyCodes []KeyCode, err error) {
	debugf("X11: GetModifierMapping (not implemented)")
	w.recordOperation(CanvasOperation{
		Type: "getModifierMapping",
		Args: []any{},
	})
	return 0, nil, nil
}

func (w *wasmX11Frontend) PolyText16(drawable xID, gcID xID, x, y int32, items []PolyTextItem) {
	gc, ok := w.gcs[gcID]
	if !ok {
		return
	}
	debugf("X11: polyText16 drawable=%s gc=%v x=%d y=%d items=%v", drawable, gc, x, y, items)
	color := "?????"
	winInfo, ok := w.windows[drawable]
	if !ok {
		return
	}
	if winInfo.canvas.IsNull() {
		return
	}
	winInfo.ctx.Call("save")
	defer winInfo.ctx.Call("restore")
	w.applyGC(winInfo.ctx, winInfo.colormap, gcID)
	color = w.getForegroundColor(winInfo.colormap, gc)
	winInfo.ctx.Set("fillStyle", color)
	// Get font from GC
	fontCSS := "12px monospace" // Default fallback
	if gc.Font != 0 {
		if font, ok := w.fonts[xID{drawable.client, gc.Font}]; ok {
			fontCSS = font.cssFont
		}
	}
	winInfo.ctx.Set("font", fontCSS)
	currentX := x
	var recordedItems []any
	for _, item := range items {
		switch it := item.(type) {
		case PolyText16String:
			currentX += int32(it.Delta)
			textBytes := make([]byte, len(it.Str)*2)
			for i, r := range it.Str {
				binary.LittleEndian.PutUint16(textBytes[i*2:], r)
			}
			decodedText := js.Global().Get("TextDecoder").New().Call("decode", jsutil.Uint8ArrayFromBytes(textBytes)).String()
			decodedText = strings.ReplaceAll(decodedText, "\x00", "") // Trim null terminators
			winInfo.ctx.Call("fillText", decodedText, currentX, y)
			recordedItems = append(recordedItems, map[string]any{"delta": it.Delta, "text": decodedText})
			// The delta for the next item is relative to the start of the current item.
			// The browser's measureText is not reliable enough to calculate the next position.
		case PolyTextFont:
			if font, ok := w.fonts[xID{drawable.client, uint32(it.Font)}]; ok {
				winInfo.ctx.Set("font", font.cssFont)
				recordedItems = append(recordedItems, map[string]any{"font": it.Font})
			}
		}
	}
	w.recordOperation(CanvasOperation{
		Type:      "polyText16",
		Args:      []any{drawable.local, gc, x, y, recordedItems},
		FillStyle: color,
	})
}

func (w *wasmX11Frontend) CreatePixmap(xid, drawable xID, width, height, depth uint32) {
	debugf("X11: createPixmap id=%s drawable=%s width=%d height=%d depth=%d", xid, drawable, width, height, depth)
	canvas := w.document.Call("createElement", "canvas")
	canvas.Set("width", width)
	canvas.Set("height", height)
	ctx := canvas.Call("getContext", "2d")
	w.pixmaps[xid] = &pixmapInfo{
		canvas:  canvas,
		context: ctx,
	}
	w.recordOperation(CanvasOperation{
		Type: "createPixmap",
		Args: []any{xid.local, drawable.local, width, height, depth},
	})
}

func (w *wasmX11Frontend) FreePixmap(xid xID) {
	debugf("X11: freePixmap id=%s", xid)
	delete(w.pixmaps, xid)
	w.recordOperation(CanvasOperation{
		Type: "freePixmap",
		Args: []any{xid.local},
	})
}

func (w *wasmX11Frontend) CopyPixmap(srcID, dstID, gcID xID, srcX, srcY, width, height, dstX, dstY uint32) {
	debugf("X11: copyPixmap src=%s dst=%s gc=%s srcX=%d srcY=%d width=%d height=%d dstX=%d dstY=%d", srcID, dstID, gcID, srcX, srcY, width, height, dstX, dstY)
	srcPixmap, srcOk := w.pixmaps[srcID]
	dstWin, dstOk := w.windows[dstID]
	if !srcOk || !dstOk {
		return
	}
	if !srcPixmap.canvas.IsNull() && !dstWin.canvas.IsNull() {
		dstWin.ctx.Call("drawImage", srcPixmap.canvas, srcX, srcY, width, height, dstX, dstY, width, height)
	}
	w.recordOperation(CanvasOperation{
		Type: "copyPixmap",
		Args: []any{srcID.local, dstID.local, gcID.local, srcX, srcY, width, height, dstX, dstY},
	})
}

func (w *wasmX11Frontend) CreateCursor(cursorID xID, source, mask xID, foreColor, backColor [3]uint16, x, y uint16) {
	debugf("X11: CreateCursor id=%s source=%s mask=%s", cursorID, source, mask)

	sourcePixmap, sourceOk := w.pixmaps[source]
	if !sourceOk {
		debugf("X11: CreateCursor source pixmap %s not found", source)
		return
	}

	maskPixmap, maskOk := w.pixmaps[mask]
	if !maskOk && mask.local != 0 {
		debugf("X11: CreateCursor mask pixmap %s not found", mask)
		return
	}

	width := sourcePixmap.canvas.Get("width").Int()
	height := sourcePixmap.canvas.Get("height").Int()

	if width == 0 || height == 0 {
		return
	}

	// Create a temporary canvas to generate the cursor image
	tempCanvas := w.document.Call("createElement", "canvas")
	tempCanvas.Set("width", width)
	tempCanvas.Set("height", height)
	tempCtx := tempCanvas.Call("getContext", "2d")

	// Get image data from source and mask pixmaps
	sourceJSData := sourcePixmap.context.Call("getImageData", 0, 0, width, height).Get("data")
	sourceBytes := make([]byte, sourceJSData.Length())
	js.CopyBytesToGo(sourceBytes, sourceJSData)

	var maskBytes []byte
	if mask.local != 0 {
		maskJSData := maskPixmap.context.Call("getImageData", 0, 0, width, height).Get("data")
		maskBytes = make([]byte, maskJSData.Length())
		js.CopyBytesToGo(maskBytes, maskJSData)
	}

	cursorBytes := make([]byte, width*height*4)

	fgR := uint8(foreColor[0] >> 8)
	fgG := uint8(foreColor[1] >> 8)
	fgB := uint8(foreColor[2] >> 8)
	bgR := uint8(backColor[0] >> 8)
	bgG := uint8(backColor[1] >> 8)
	bgB := uint8(backColor[2] >> 8)

	for i := 0; i < width*height; i++ {
		idx := i * 4

		maskBitOn := true
		if mask.local != 0 {
			maskBitOn = maskBytes[idx+3] > 0
		}

		if maskBitOn {
			sourceBitOn := sourceBytes[idx+3] > 0
			if sourceBitOn {
				// Foreground
				cursorBytes[idx+0] = fgR
				cursorBytes[idx+1] = fgG
				cursorBytes[idx+2] = fgB
				cursorBytes[idx+3] = 255
			} else {
				// Background
				cursorBytes[idx+0] = bgR
				cursorBytes[idx+1] = bgG
				cursorBytes[idx+2] = bgB
				cursorBytes[idx+3] = 255
			}
		} else {
			// Transparent
			cursorBytes[idx+0] = 0
			cursorBytes[idx+1] = 0
			cursorBytes[idx+2] = 0
			cursorBytes[idx+3] = 0
		}
	}

	cursorDataArray := jsutil.Uint8ClampedArrayFromBytes(cursorBytes)

	cursorImageData := js.Global().Get("ImageData").New(cursorDataArray, width, height)
	tempCtx.Call("putImageData", cursorImageData, 0, 0)

	dataURL := tempCanvas.Call("toDataURL").String()
	cursorStyle := fmt.Sprintf("url(%s) %d %d, auto", dataURL, x, y)

	w.cursorStyles[cursorID.local] = &cursorInfo{
		style:  cursorStyle,
		source: source,
		mask:   mask,
		x:      x,
		y:      y,
	}

	w.recordOperation(CanvasOperation{
		Type: "createCursor",
		Args: []any{cursorID.local, source.local, mask.local, x, y},
	})
}

func (w *wasmX11Frontend) WarpPointer(x, y int16) {
	debugf("X11: warpPointer x=%d y=%d", x, y)
	w.server.UpdatePointerPosition(x, y)
	w.recordOperation(CanvasOperation{
		Type: "warpPointer",
		Args: []any{x, y},
	})
}

func (w *wasmX11Frontend) CreateCursorFromGlyph(cursorID uint32, glyphID uint16) {
	debugf("X11: createCursorFromGlyph cursorID=%d glyphID=%d", cursorID, glyphID)
	// This is a simplified mapping from X11 cursor font glyphs to CSS cursor styles.
	var style string
	switch glyphID {
	case 68: // XC_xterm
		style = "text"
	case 34: // XC_crosshair
		style = "crosshair"
	case 58: // XC_hand1
		style = "pointer"
	case 52: // XC_fleur
		style = "move"
	case 138: // XC_right_ptr
		style = "pointer"
	case 108: // XC_watch
		style = "wait"
	case 118: // XC_sb_h_double_arrow
		style = "ew-resize"
	case 120: // XC_sb_v_double_arrow
		style = "ns-resize"
	default:
		style = "default"
	}
	w.cursorStyles[cursorID] = &cursorInfo{style: style}
	w.recordOperation(CanvasOperation{
		Type: "createCursorFromGlyph",
		Args: []any{cursorID, glyphID},
	})
}

func (w *wasmX11Frontend) SetWindowCursor(windowID xID, cursorID xID) {
	debugf("X11: setWindowCursor window=%s cursor=%s", windowID, cursorID)
	if winInfo, ok := w.windows[windowID]; ok {
		if cursor, ok := w.cursorStyles[cursorID.local]; ok {
			winInfo.canvas.Get("style").Set("cursor", cursor.style)
		} else {
			winInfo.canvas.Get("style").Set("cursor", "default")
		}
	}
	w.recordOperation(CanvasOperation{
		Type: "setWindowCursor",
		Args: []any{windowID.local, cursorID.local},
	})
}

func (w *wasmX11Frontend) CopyGC(srcGCID, dstGCID xID) {
	debugf("X11: copyGC src=%s dst=%s", srcGCID, dstGCID)
	if srcGC, ok := w.gcs[srcGCID]; ok {
		newGC := srcGC
		w.gcs[dstGCID] = newGC
	}
	w.recordOperation(CanvasOperation{
		Type: "copyGC",
		Args: []any{srcGCID.local, dstGCID.local},
	})
}

func (w *wasmX11Frontend) FreeGC(gcID xID) {
	debugf("X11: freeGC id=%s", gcID)
	delete(w.gcs, gcID)
	w.recordOperation(CanvasOperation{
		Type: "freeGC",
		Args: []any{gcID.local},
	})
}

func (w *wasmX11Frontend) FreeCursor(cursorID xID) {
	debugf("X11: freeCursor id=%s", cursorID)
	// In the wasm frontend, we only store the CSS style mapping.
	// We don't need to "free" a DOM element for a cursor.
	// We just remove it from our internal map.
	delete(w.cursorStyles, cursorID.local) // Note: cursorStyles map uses uint32 as key
	w.recordOperation(CanvasOperation{
		Type: "freeCursor",
		Args: []any{cursorID.local},
	})
}

func (w *wasmX11Frontend) SendEvent(eventData messageEncoder) {
	encodedData := eventData.encodeMessage(w.server.byteOrder)
	debugf("X11: SendEvent data=%v", encodedData)
	// In a real implementation, this would send the event data back to the Go server
	// which would then forward it to the X11 client.
	w.recordOperation(CanvasOperation{
		Type: "sendEvent",
		Args: []any{encodedData},
	})
}

func (w *wasmX11Frontend) GetFocusWindow(clientID uint32) xID {
	if w.focusedWindowID.client == clientID {
		return w.focusedWindowID
	}
	return xID{}
}

func (w *wasmX11Frontend) GetProperty(window xID, property uint32) ([]byte, uint32, uint32) {
	if winInfo, ok := w.windows[window]; ok {
		if prop, ok := winInfo.properties[property]; ok {
			return prop.data, prop.typeAtom, prop.format
		}
	}
	return nil, 0, 0
}

func (w *wasmX11Frontend) ConvertSelection(selection, target, property uint32, requestor xID) {
	debugf("X11: convertSelection selection=%d target=%d property=%d requestor=%s", selection, target, property, requestor)
	// This is a simplified implementation. A real implementation would send a SelectionRequest
	// event to the owner of the selection and wait for a SelectionNotify event.
	if selection == w.GetAtom(requestor.client, "CLIPBOARD") {
		// For now, we only support clipboard operations.
		// We will read the clipboard and send a SelectionNotify event.
		go func() {
			clipboardContent, err := w.ReadClipboard()
			if err != nil {
				return
			}
			// Find the client associated with the requestor window
			if _, ok := w.windows[requestor]; !ok {
				debugf("X11: ConvertSelection: Requestor window %s not found", requestor)
				return
			}
			w.server.SendSelectionNotify(requestor, selection, target, property, []byte(clipboardContent))
		}()
	}
}

func (w *wasmX11Frontend) GrabPointer(grabWindow xID, ownerEvents bool, eventMask uint16, pointerMode, keyboardMode byte, confineTo uint32, cursor uint32, time uint32) byte {
	debugf("X11: GrabPointer (not implemented)")
	return 0 // Success
}

func (w *wasmX11Frontend) UngrabPointer(time uint32) {
	debugf("X11: UngrabPointer (not implemented)")
}

func (w *wasmX11Frontend) GrabKeyboard(grabWindow xID, ownerEvents bool, time uint32, pointerMode, keyboardMode byte) byte {
	debugf("X11: GrabKeyboard (not implemented)")
	return 0 // Success
}

func (w *wasmX11Frontend) UngrabKeyboard(time uint32) {
	debugf("X11: UngrabKeyboard (not implemented)")
}

func (w *wasmX11Frontend) initDefaultCursors() {
	// This is a minimal mapping from X11 cursor names to CSS cursor values.
	// The cursor IDs are taken from the standard X11 cursor font.
	w.cursorStyles[68] = &cursorInfo{style: "pointer"}
	w.cursorStyles[34] = &cursorInfo{style: "crosshair"}
	w.cursorStyles[58] = &cursorInfo{style: "help"}
	w.cursorStyles[52] = &cursorInfo{style: "move"}
	w.cursorStyles[138] = &cursorInfo{style: "text"}
	w.cursorStyles[108] = &cursorInfo{style: "wait"}
	w.cursorStyles[116] = &cursorInfo{style: "wait"}
	w.cursorStyles[118] = &cursorInfo{style: "w-resize"}
	w.cursorStyles[120] = &cursorInfo{style: "e-resize"}
	w.cursorStyles[76] = &cursorInfo{style: "n-resize"}
	w.cursorStyles[14] = &cursorInfo{style: "s-resize"}
	w.cursorStyles[10] = &cursorInfo{style: "nw-resize"}
	w.cursorStyles[12] = &cursorInfo{style: "ne-resize"}
	w.cursorStyles[134] = &cursorInfo{style: "sw-resize"}
	w.cursorStyles[136] = &cursorInfo{style: "se-resize"}
}

func (w *wasmX11Frontend) SetCursor(windowID xID, cursorID uint32) {
	debugf("X11: setCursor window=%s cursor=%d", windowID, cursorID)
	if winInfo, ok := w.windows[windowID]; ok {
		if style, ok := w.cursorStyles[cursorID]; ok {
			winInfo.canvas.Get("style").Set("cursor", style)
		} else {
			winInfo.canvas.Get("style").Set("cursor", "default")
		}
	}
	w.recordOperation(CanvasOperation{
		Type: "setCursor",
		Args: []any{windowID.local, cursorID},
	})
}

func (w *wasmX11Frontend) ListProperties(window xID) []uint32 {
	// For now, return an empty slice.
	return []uint32{}
}

func (w *wasmX11Frontend) GetAtom(clientID uint32, name string) uint32 {
	if id, ok := w.atoms[name]; ok {
		return id
	}
	id := w.nextAtomID
	w.nextAtomID++
	w.atoms[name] = id
	return id
}

func (w *wasmX11Frontend) GetAtomName(atom uint32) string {
	for name, id := range w.atoms {
		if id == atom {
			return name
		}
	}
	return ""
}

func (w *wasmX11Frontend) ReadClipboard() (string, error) {
	return jsutil.ReadClipboard()
}

func (w *wasmX11Frontend) WriteClipboard(s string) error {
	return jsutil.WriteClipboard(s)
}

func (w *wasmX11Frontend) UpdatePointerPosition(x, y int16) {
	w.server.UpdatePointerPosition(x, y)
}

func (w *wasmX11Frontend) Bell(percent int8) {
	debugf("X11: bell percent=%d", percent)
	w.recordOperation(CanvasOperation{
		Type: "bell",
		Args: []any{percent},
	})
}

func (w *wasmX11Frontend) GetRGBColor(colormap xID, pixel uint32) (r, g, b uint8) {
	return w.server.GetRGBColor(colormap, pixel)
}

func (w *wasmX11Frontend) OpenFont(fid xID, name string) {
	debugf("X11: OpenFont fid=%s name=%s", fid, name)
	debugf("X11: OpenFont received font name: %s", name)

	_, _, _, _, cssFont := MapX11FontToCSS(name)

	w.fonts[fid] = &fontInfo{
		x11Name: name,
		cssFont: cssFont,
	}

	w.recordOperation(CanvasOperation{
		Type: "openFont",
		Args: []any{fid.local, name},
	})
}

func (w *wasmX11Frontend) CloseFont(fid xID) {
	debugf("X11: CloseFont fid=%s", fid)
	delete(w.fonts, fid)
	w.recordOperation(CanvasOperation{
		Type: "closeFont",
		Args: []any{fid.local},
	})
}

func (w *wasmX11Frontend) AllowEvents(clientID uint32, mode byte, time uint32) {
	debugf("X11: AllowEvents mode=%d time=%d (not implemented)", mode, time)
	w.recordOperation(CanvasOperation{
		Type: "allowEvents",
		Args: []any{mode, time},
	})
}

func (w *wasmX11Frontend) SendConfigureAndExposeEvent(windowID xID, x, y int16, width, height uint16) {
	w.server.sendConfigureNotifyEvent(windowID, x, y, width, height)
	w.server.sendExposeEvent(windowID, 0, 0, width, height) // Send expose for the entire window
	if win, ok := w.server.windows[windowID]; ok {
		for _, childID := range win.children {
			childXID := xID{client: windowID.client, local: childID}
			if childWin, ok := w.server.windows[childXID]; ok {
				w.server.sendExposeEvent(childXID, 0, 0, childWin.width, childWin.height)
			}
		}
	}
}

// mouseEventHandler creates a js.Func for mouse events.
func (w *wasmX11Frontend) mouseEventHandler(xid xID, eventType string) js.Func {
	return js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		if _, ok := w.windows[xid]; !ok {
			return nil
		}
		event := args[0]
		valOffsetX := event.Get("offsetX")
		offsetX := int32(0)
		if valOffsetX.Type() == js.TypeNumber {
			offsetX = int32(valOffsetX.Int())
		}

		valOffsetY := event.Get("offsetY")
		offsetY := int32(0)
		if valOffsetY.Type() == js.TypeNumber {
			offsetY = int32(valOffsetY.Int())
		}

		valButtons := event.Get("buttons")
		buttons := uint32(0)
		if valButtons.Type() == js.TypeNumber {
			buttons = uint32(valButtons.Int())
		}

		valDeltaY := event.Get("deltaY")
		deltaY := int32(0)
		if valDeltaY.Type() == js.TypeNumber {
			deltaY = int32(valDeltaY.Int())
		}

		if eventType == "wheel" {
			event.Call("preventDefault") // Prevent page scrolling
			w.server.SendMouseEvent(xid, eventType, offsetX, offsetY, deltaY)
			debugf("Mouse wheel event: window=%s, x=%d, y=%d, deltaY=%d", xid, offsetX, offsetY, deltaY)
		} else {
			w.server.SendMouseEvent(xid, eventType, offsetX, offsetY, int32(buttons))
			debugf("Mouse event: window=%s, type=%s, x=%d, y=%d, buttons=%d", xid, eventType, offsetX, offsetY, buttons)
		}
		if eventType == "mousemove" {
			debugf("Mouse move event: window=%s, x=%d, y=%d", xid, offsetX, offsetY)
			w.server.UpdatePointerPosition(int16(offsetX), int16(offsetY))
		}
		return nil
	})
}

// keyboardEventHandler creates a js.Func for keyboard events.
func (w *wasmX11Frontend) keyboardEventHandler(xid xID, eventType string) js.Func {
	return js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		if _, ok := w.windows[xid]; !ok {
			return nil
		}
		event := args[0]
		keyCode := event.Get("keyCode").Int()
		altKey := event.Get("altKey").Bool()
		ctrlKey := event.Get("ctrlKey").Bool()
		shiftKey := event.Get("shiftKey").Bool()
		metaKey := event.Get("metaKey").Bool()

		w.server.SendKeyboardEvent(w.focusedWindowID, eventType, keyCode, altKey, ctrlKey, shiftKey, metaKey)
		debugf("Keyboard event: window=%s, type=%s, keyCode=%d, alt=%t, ctrl=%t, shift=%t, meta=%t", w.focusedWindowID, eventType, keyCode, altKey, ctrlKey, shiftKey, metaKey)
		return nil
	})
}

func (w *wasmX11Frontend) QueryFont(fid xID) (minBounds, maxBounds xCharInfo, minCharOrByte2, maxCharOrByte2, defaultChar uint16, drawDirection uint8, minByte1, maxByte1 uint8, allCharsExist bool, fontAscent, fontDescent int16, charInfos []xCharInfo) {
	w.recordOperation(CanvasOperation{
		Type: "queryFont",
		Args: []any{fid.local},
	})
	debugf("X11: QueryFont fid=%s", fid)

	fontDescent = 5

	// Try to get font info from the opened fonts map
	var cssFont string = "12px monospace" // Default fallback
	if font, ok := w.fonts[fid]; ok {
		cssFont = font.cssFont
		// Parse font size from cssFont string (e.g., "12px monospace")
		sizeStr := strings.Split(font.cssFont, " ")[0] // Get "12px"
		sizeStr = strings.TrimSuffix(sizeStr, "px")
		if size, err := strconv.ParseFloat(sizeStr, 64); err == nil {
			// Derive ascent, descent from the font size
			fontAscent = int16(math.Round(size * 0.8))
			fontDescent = int16(math.Round(size * 0.2))
		}
	}

	// Create a temporary off-screen canvas for font measurement
	canvas := w.document.Call("createElement", "canvas")
	ctx := canvas.Call("getContext", "2d")
	ctx.Set("font", cssFont)

	// Measure overall font metrics using a dummy character (e.g., 'M')
	overallMetrics := ctx.Call("measureText", "M")
	if !overallMetrics.Get("fontBoundingBoxAscent").IsUndefined() {
		fontAscent = int16(math.Round(overallMetrics.Get("fontBoundingBoxAscent").Float()))
	}
	if !overallMetrics.Get("fontBoundingBoxDescent").IsUndefined() {
		fontDescent = int16(math.Round(overallMetrics.Get("fontBoundingBoxDescent").Float()))
	}
	if fontAscent <= 0 {
		fontAscent = 1
	}
	if fontDescent <= 0 {
		fontDescent = 1
	}

	// Measure metrics for a space character to initialize min/max bounds
	spaceMetrics := ctx.Call("measureText", " ")
	initialCharWidth := uint16(math.Round(spaceMetrics.Get("width").Float()))
	initialAscent := int16(math.Round(spaceMetrics.Get("actualBoundingBoxAscent").Float()))
	initialDescent := int16(math.Round(spaceMetrics.Get("actualBoundingBoxDescent").Float()))
	initialLSB := int16(math.Round(spaceMetrics.Get("actualBoundingBoxLeft").Float()))
	initialRSB := int16(math.Round(spaceMetrics.Get("actualBoundingBoxRight").Float()))

	minBounds = xCharInfo{
		LeftSideBearing:  initialLSB,
		RightSideBearing: initialRSB,
		CharacterWidth:   initialCharWidth,
		Ascent:           initialAscent,
		Descent:          initialDescent,
	}
	maxBounds = xCharInfo{
		LeftSideBearing:  initialLSB,
		RightSideBearing: initialRSB,
		CharacterWidth:   initialCharWidth,
		Ascent:           initialAscent,
		Descent:          initialDescent,
	}

	minCharOrByte2 = 0
	maxCharOrByte2 = 255 // ASCII range
	defaultChar = 0      // Will be set to ' ' (32) if not all chars exist
	drawDirection = 0    // LeftToRight
	minByte1 = 0
	maxByte1 = 0
	allCharsExist = true // Assume true, set to false if any char has 0 width

	charInfos = make([]xCharInfo, maxCharOrByte2-minCharOrByte2+1)

	for i := minCharOrByte2; i <= maxCharOrByte2; i++ {
		char := string(rune(i))
		metrics := ctx.Call("measureText", char)

		var charLSB, charRSB int16
		var charWidth uint16
		var charAscent, charDescent int16

		// Use actualBoundingBox properties for more accurate metrics
		if !metrics.Get("actualBoundingBoxLeft").IsUndefined() {
			charLSB = int16(math.Round(metrics.Get("actualBoundingBoxLeft").Float()))
		}
		if !metrics.Get("actualBoundingBoxRight").IsUndefined() {
			charRSB = int16(math.Round(metrics.Get("actualBoundingBoxRight").Float()))
		}
		if !metrics.Get("width").IsUndefined() {
			charWidth = uint16(math.Round(metrics.Get("width").Float()))
			if charWidth == 0 { // Ensure minimum width
				charWidth = 1
				if i != 0 { // If it's not the null character, and width is 0, then it doesn't exist
					allCharsExist = false
				}
			}
		} else {
			charWidth = 1 // Default to 1 if width is undefined
			if i != 0 {
				allCharsExist = false
			}
		}

		if !metrics.Get("actualBoundingBoxAscent").IsUndefined() {
			charAscent = int16(math.Round(math.Abs(metrics.Get("actualBoundingBoxAscent").Float())))
		} else {
			charAscent = fontAscent // Fallback to overall font ascent
		}
		if !metrics.Get("actualBoundingBoxDescent").IsUndefined() {
			charDescent = int16(math.Round(math.Abs(metrics.Get("actualBoundingBoxDescent").Float())))
		} else {
			charDescent = fontAscent // Fallback to overall font ascent
		}

		// Ensure ascent and descent are at least 1
		if charAscent <= 0 {
			charAscent = 1
		}
		if charDescent <= 0 {
			charDescent = 1
		}

		ci := xCharInfo{
			LeftSideBearing:  charLSB,
			RightSideBearing: charRSB,
			CharacterWidth:   charWidth,
			Ascent:           charAscent,
			Descent:          charDescent,
			Attributes:       0,
		}
		charInfos[i] = ci

		// Update minBounds
		if ci.LeftSideBearing < minBounds.LeftSideBearing {
			minBounds.LeftSideBearing = ci.LeftSideBearing
		}
		if ci.RightSideBearing < minBounds.RightSideBearing {
			minBounds.RightSideBearing = ci.RightSideBearing
		}
		if ci.CharacterWidth < minBounds.CharacterWidth {
			minBounds.CharacterWidth = ci.CharacterWidth
		}
		if ci.Ascent < minBounds.Ascent {
			minBounds.Ascent = ci.Ascent
		}
		if ci.Descent < minBounds.Descent {
			minBounds.Descent = ci.Descent
		}

		// Update maxBounds
		if ci.LeftSideBearing > maxBounds.LeftSideBearing {
			maxBounds.LeftSideBearing = ci.LeftSideBearing
		}
		if ci.RightSideBearing > maxBounds.RightSideBearing {
			maxBounds.RightSideBearing = ci.RightSideBearing
		}
		if ci.CharacterWidth > maxBounds.CharacterWidth {
			maxBounds.CharacterWidth = ci.CharacterWidth
		}
		if ci.Ascent > maxBounds.Ascent {
			maxBounds.Ascent = ci.Ascent
		}
		if ci.Descent > maxBounds.Descent {
			maxBounds.Descent = ci.Descent
		}
	}

	// Ensure minBounds ascent and descent are at least 1
	if minBounds.Ascent <= 0 {
		minBounds.Ascent = 1
	}
	if minBounds.Descent <= 0 {
		minBounds.Descent = 1
	}

	if !allCharsExist {
		defaultChar = 32 // Set defaultChar to space if not all characters exist
	}

	// Release the temporary canvas element
	canvas.Call("remove")

	debugf("X11: QueryFont fid=%s reply: minBounds=%+v, maxBounds=%+v, minCharOrByte2=%d, maxCharOrByte2=%d, defaultChar=%d, drawDirection=%d, minByte1=%d, maxByte1=%d, allCharsExist=%t, fontAscent=%d, fontDescent=%d, len(charInfos)=%d", fid, minBounds, maxBounds, minCharOrByte2, maxCharOrByte2, defaultChar, drawDirection, minByte1, maxByte1, allCharsExist, fontAscent, fontDescent, len(charInfos))

	return
}

func (w *wasmX11Frontend) ListFonts(maxNames uint16, pattern string) []string {
	debugf("X11: ListFonts maxNames=%d pattern=%s", maxNames, pattern)

	// Simplified implementation: return a hardcoded list of fonts.
	// In a real implementation, this would query available fonts.
	// The pattern matching is also simplified.

	var matchingFonts []string

	availableFonts := GetAvailableFonts()

	for _, font := range availableFonts {
		if strings.Contains(font, pattern) || pattern == "*" || pattern == "" {
			matchingFonts = append(matchingFonts, font)
			if len(matchingFonts) >= int(maxNames) && maxNames != 0 {
				break
			}
		}
	}

	w.recordOperation(CanvasOperation{
		Type: "listFonts",
		Args: []any{maxNames, pattern},
	})

	return matchingFonts
}

func (w *wasmX11Frontend) GetWindowAttributes(xid xID) WindowAttributes {
	// Not implemented for wasm
	w.recordOperation(CanvasOperation{
		Type: "getWindowAttributes",
		Args: []any{xid.local},
	})
	return WindowAttributes{}
}

func (w *wasmX11Frontend) ChangeWindowAttributes(xid xID, valueMask uint32, values WindowAttributes) {
	debugf("X11: changeWindowAttributes id=%s valueMask=%d values=%+v", xid, valueMask, values)
	if winInfo, ok := w.windows[xid]; ok {
		style := winInfo.div.Get("style")
		if valueMask&CWColormap != 0 {
			winInfo.colormap = xID{client: xid.client, local: uint32(values.Colormap)}
		}
		if valueMask&CWBackPixel != 0 {
			r, g, b := w.GetRGBColor(winInfo.colormap, values.BackgroundPixel)
			bgColor := fmt.Sprintf("rgb(%d, %d, %d)", r, g, b)
			style.Set("backgroundColor", bgColor)
		}
		if valueMask&CWBorderPixel != 0 {
			r, g, b := w.GetRGBColor(winInfo.colormap, values.BorderPixel)
			borderColor := fmt.Sprintf("rgb(%d, %d, %d)", r, g, b)
			style.Set("borderColor", borderColor)
		}
		if valueMask&CWCursor != 0 {
			w.SetWindowCursor(xid, xID{client: xid.client, local: uint32(values.Cursor)})
		}
	}
	w.recordOperation(CanvasOperation{
		Type: "changeWindowAttributes",
		Args: []any{xid.local, valueMask},
	})
}

func uint32SliceToAnySlice(s []uint32) []any {
	anySlice := make([]any, len(s))
	for i, v := range s {
		anySlice[i] = v
	}
	return anySlice
}

func (w *wasmX11Frontend) DestroyAllWindowsForClient(client uint32) {
	for xid := range w.windows {
		if xid.client == client {
			w.destroyWindow(xid, false)
		}
	}
}
