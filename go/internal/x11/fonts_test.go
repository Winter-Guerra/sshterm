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
		// Extended tests
		{"Helvetica 12", "-*-helvetica-medium-r-normal-*-12-*-*-*-p-*-iso8859-1", "normal normal 12px sans-serif"},
		{"Courier 12", "-*-courier-medium-r-normal-*-12-*-*-*-m-*-iso8859-1", "normal normal 12px monospace"},
		{"Courier Bold 12", "-*-courier-bold-r-normal-*-12-*-*-*-m-*-iso8859-1", "bold normal 12px monospace"},
		{"Courier Bold 14", "-*-courier-bold-r-normal-*-14-*-*-*-m-*-iso8859-1", "bold normal 14px monospace"},
		{"Courier Bold 18", "-*-courier-bold-r-normal-*-18-*-*-*-m-*-iso8859-1", "bold normal 18px monospace"},
		{"Courier Oblique 12", "-*-courier-medium-o-normal-*-12-*-*-*-m-*-iso8859-1", "normal oblique 12px monospace"},
		{"Times Bold 12", "-*-times-bold-r-normal-*-12-*-*-*-p-*-iso8859-1", "bold normal 12px serif"},
		{"Times Bold 18", "-*-times-bold-r-normal-*-18-*-*-*-p-*-iso8859-1", "bold normal 18px serif"},
		{"Times Italic 12", "-*-times-medium-i-normal-*-12-*-*-*-p-*-iso8859-1", "normal italic 12px serif"},
		{"Helvetica Oblique 12", "-*-helvetica-medium-o-normal-*-12-*-*-*-p-*-iso8859-1", "normal oblique 12px sans-serif"},
		{"Helvetica Bold 14", "-*-helvetica-bold-r-normal-*-14-*-*-*-p-*-iso8859-1", "bold normal 14px sans-serif"},
		{"Fixed 13", "-misc-fixed-medium-r-normal--13-120-75-75-c-70-iso8859-1", "normal normal 13px monospace"},
		{"Fixed 6x13", "-misc-fixed-medium-r-semicondensed--13-120-75-75-c-60-iso8859-1", "normal normal 13px monospace"},
		{"cursor", "cursor", "normal normal 12px monospace"},
		{"9x15 alias", "9x15", "normal normal 15px monospace"},
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

func TestGetAvailableFonts(t *testing.T) {
	fonts := GetAvailableFonts()
	if len(fonts) < 50 {
		t.Errorf("Expected at least 50 fonts, got %d", len(fonts))
	}
}
