//go:build x11 && wasm && debug

package x11

import (
	"log"
	"syscall/js"
)

func logf(format string, v ...interface{}) {
	log.Printf(format, v...)
}

func (w *wasmX11Frontend) logf(format string, v ...interface{}) {
	log.Printf(format, v...)
}

func (w *wasmX11Frontend) recordOperation(op CanvasOperation) {
	for i, arg := range op.Args {
		if gc, ok := arg.(*GC); ok {
			op.Args[i] = map[string]interface{}{
				"Function":          gc.Function,
				"PlaneMask":         gc.PlaneMask,
				"Foreground":        gc.Foreground,
				"Background":        gc.Background,
				"LineWidth":         gc.LineWidth,
				"LineStyle":         gc.LineStyle,
				"CapStyle":          gc.CapStyle,
				"JoinStyle":         gc.JoinStyle,
				"FillStyle":         gc.FillStyle,
				"FillRule":          gc.FillRule,
				"Tile":              gc.Tile,
				"Stipple":           gc.Stipple,
				"TileStipXOrigin":   gc.TileStipXOrigin,
				"TileStipYOrigin":   gc.TileStipYOrigin,
				"Font":              gc.Font,
				"SubwindowMode":     gc.SubwindowMode,
				"GraphicsExposures": gc.GraphicsExposures,
				"ClipXOrigin":       gc.ClipXOrigin,
				"ClipYOrigin":       gc.ClipYOrigin,
				"ClipMask":          gc.ClipMask,
				"DashOffset":        gc.DashOffset,
				"Dashes":            gc.Dashes,
				"ArcMode":           gc.ArcMode,
			}
		} else if slice, ok := arg.([]uint32); ok {
			anySlice := make([]any, len(slice))
			for j, v := range slice {
				anySlice[j] = v
			}
			op.Args[i] = anySlice
		} else if items, ok := arg.([]PolyText8Item); ok {
			anySlice := make([]any, len(items))
			for j, v := range items {
				anySlice[j] = map[string]any{"delta": v.Delta, "text": string(v.Str)}
			}
			op.Args[i] = anySlice
		} else if items, ok := arg.([]PolyText16Item); ok {
			anySlice := make([]any, len(items))
			for j, v := range items {
				anySlice[j] = map[string]any{"delta": v.Delta, "text": uint16SliceToString(v.Str)}
			}
			op.Args[i] = anySlice
		}
	}
	w.canvasOperations = append(w.canvasOperations, op)
}

func uint16SliceToString(s []uint16) string {
	runes := make([]rune, len(s))
	for i, v := range s {
		runes[i] = rune(v)
	}
	return string(runes)
}

func (w *wasmX11Frontend) GetCanvasOperations() []CanvasOperation {
	return w.canvasOperations
}

func (w *wasmX11Frontend) initCanvasOperations() {
	w.canvasOperations = []CanvasOperation{}
	js.Global().Set("getCanvasOperations", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		ops := w.GetCanvasOperations()
		jsOps := make([]interface{}, len(ops))
		for i, op := range ops {
			jsOps[i] = map[string]interface{}{
				"Type":        op.Type,
				"Args":        op.Args,
				"FillStyle":   op.FillStyle,
				"StrokeStyle": op.StrokeStyle,
			}
		}
		return js.ValueOf(jsOps)
	}))
}
