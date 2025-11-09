//go:build x11 && !wasm

package x11

import (
	"testing"
)

func TestMapX11FontToCSS(t *testing.T) {
	tests := []struct {
		name        string
		x11FontName string
		expectedCSS string
	}{
		{
			name:        "XLFD with point size",
			x11FontName: "-*-*-*-R-*-*-*-120-*-*-*-*-ISO8859-*",
			expectedCSS: "normal normal 12px sans-serif",
		},
		{
			name:        "XLFD with pixel size",
			x11FontName: "-*-helvetica-medium-r-normal-*-12-*-*-*-p-*-iso8859-1",
			expectedCSS: "normal normal 12px sans-serif",
		},
		{
			name:        "XLFD with bold weight",
			x11FontName: "-*-helvetica-bold-r-normal-*-14-*-*-*-p-*-iso8859-1",
			expectedCSS: "bold normal 14px sans-serif",
		},
		{
			name:        "XLFD with italic slant",
			x11FontName: "-*-times-medium-i-normal-*-12-*-*-*-p-*-iso8859-1",
			expectedCSS: "normal italic 12px serif",
		},
		{
			name:        "Fixed alias",
			x11FontName: "fixed",
			expectedCSS: "normal normal 12px monospace",
		},
		{
			name:        "Variable alias",
			x11FontName: "variable",
			expectedCSS: "normal normal 12px sans-serif",
		},
		{
			name:        "Unknown XLFD, fallback",
			x11FontName: "some-random-font",
			expectedCSS: "normal normal 12px monospace", // Default fallback
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, _, _, cssFont := MapX11FontToCSS(tt.x11FontName)
			if cssFont != tt.expectedCSS {
				t.Errorf("MapX11FontToCSS(%q) got %q, want %q", tt.x11FontName, cssFont, tt.expectedCSS)
			}
		})
	}
}
