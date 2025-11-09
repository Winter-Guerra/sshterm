//go:build x11

package x11

// GC attribute masks
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
	GCDashes            = 1 << 21
	GCArcMode           = 1 << 22
)

const (
	// Graphics functions
	FunctionClear        = 0
	FunctionAnd          = 1
	FunctionAndReverse   = 2
	FunctionCopy         = 3
	FunctionAndInverted  = 4
	FunctionNoOp         = 5
	FunctionXor          = 6
	FunctionOr           = 7
	FunctionNor          = 8
	FunctionEquiv        = 9
	FunctionInvert       = 10
	FunctionOrReverse    = 11
	FunctionCopyInverted = 12
	FunctionOrInverted   = 13
	FunctionNand         = 14
	FunctionSet          = 15
)

const (
	LineStyleSolid      = 0
	LineStyleOnOffDash  = 1
	LineStyleDoubleDash = 2
)

const (
	CapStyleNotLast    = 0
	CapStyleButt       = 1
	CapStyleRound      = 2
	CapStyleProjecting = 3
)

const (
	JoinStyleMiter = 0
	JoinStyleRound = 1
	JoinStyleBevel = 2
)

const (
	FillStyleSolid          = 0
	FillStyleTiled          = 1
	FillStyleStippled       = 2
	FillStyleOpaqueStippled = 3
)

const (
	FillRuleEvenOdd = 0
	FillRuleWinding = 1
)

const (
	SubwindowModeClipByChildren   = 0
	SubwindowModeIncludeInferiors = 1
)

const (
	ArcModeChord    = 0
	ArcModePieSlice = 1
)

// GC represents a Graphics Context.
// See: https://www.x.org/releases/X11R7.6/doc/xproto/x11protocol.html#requests:CreateGC
type GC struct {
	Function           uint32
	PlaneMask          uint32
	Foreground         uint32
	Background         uint32
	LineWidth          uint32
	LineStyle          uint32
	CapStyle           uint32
	JoinStyle          uint32
	FillStyle          uint32
	FillRule           uint32
	Tile               uint32
	Stipple            uint32
	TileStipXOrigin    uint32
	TileStipYOrigin    uint32
	Font               uint32
	SubwindowMode      uint32
	GraphicsExposures  uint32
	ClipXOrigin        int32
	ClipYOrigin        int32
	ClipMask           uint32
	DashOffset         uint32
	Dashes             uint32
	ArcMode            uint32
	ClippingRectangles []Rectangle
	DashPattern        []byte
}
