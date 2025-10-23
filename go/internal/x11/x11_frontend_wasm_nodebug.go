//go:build x11 && wasm && !debug

package x11

func (w *wasmX11Frontend) logf(format string, v ...interface{}) {}

func logf(format string, v ...interface{}) {}

func (w *wasmX11Frontend) recordOperation(op CanvasOperation) {}

func (w *wasmX11Frontend) GetCanvasOperations() []CanvasOperation {
	return nil
}

func (w *wasmX11Frontend) initCanvasOperations() {}
