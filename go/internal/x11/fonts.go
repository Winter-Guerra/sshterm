//go:build x11

package x11

import (
	"fmt"
	"strconv"
	"strings"
)

// MapX11FontToCSS converts an X11 font name (XLFD) to CSS font properties.
// This is a simplified implementation.
func MapX11FontToCSS(x11FontName string) (size, family, weight, slant, cssFont string) {
	// Default values
	size = "12px"
	family = "monospace"
	weight = "normal"
	slant = "normal"

	// Handle common aliases first
	switch strings.ToLower(x11FontName) {
	case "fixed", "9x15", "10x20", "6x13", "7x14", "8x16":
		size = "12px"
		family = "monospace"
		weight = "normal"
		slant = "normal"
	case "variable":
		size = "12px"
		family = "sans-serif"
		weight = "normal"
		slant = "normal"
	case "lucidasans-10":
		size = "10px"
		family = "sans-serif"
		weight = "normal"
		slant = "normal"
	case "lucidasans-12":
		size = "12px"
		family = "sans-serif"
		weight = "normal"
		slant = "normal"
	case "lucidasans-14":
		size = "14px"
		family = "sans-serif"
		weight = "normal"
		slant = "normal"
	case "lucidasans-bold-10":
		size = "10px"
		family = "sans-serif"
		weight = "bold"
		slant = "normal"
	case "lucidasans-bold-12":
		size = "12px"
		family = "sans-serif"
		weight = "bold"
		slant = "normal"
	case "lucidasans-bold-14":
		size = "14px"
		family = "sans-serif"
		weight = "bold"
		slant = "normal"
	case "dejavu sans mono-10":
		size = "10px"
		family = "monospace"
		weight = "normal"
		slant = "normal"
	case "dejavu sans mono-12":
		size = "12px"
		family = "monospace"
		weight = "normal"
		slant = "normal"
	case "dejavu sans mono-14":
		size = "14px"
		family = "monospace"
		weight = "normal"
		slant = "normal"
	case "dejavu sans mono-bold-10":
		size = "10px"
		family = "monospace"
		weight = "bold"
		slant = "normal"
	case "dejavu sans mono-bold-12":
		size = "12px"
		family = "monospace"
		weight = "bold"
		slant = "normal"
	case "dejavu sans mono-bold-14":
		size = "14px"
		family = "monospace"
		weight = "bold"
		slant = "normal"
	}

	// Example XLFD: -*-helvetica-medium-r-normal-*-12-*-*-*-p-*-iso8859-1
	parts := strings.Split(x11FontName, "-")

	// Attempt to parse XLFD
	if len(parts) >= 13 {
		// Pixel size is usually the 7th field (0-indexed XLFD field 6)
		if len(parts[6]) > 0 && parts[6] != "*" {
			size = parts[6] + "px"
		} else if len(parts[7]) > 0 && parts[7] != "*" { // Point size is XLFD field 7
			// X11 point sizes are in tenths of a point, but sometimes used directly.
			if pointSize, err := strconv.ParseFloat(parts[7], 64); err == nil {
				// Heuristic: if pointSize is small (e.g., 6-48), treat as direct pixel size.
				// Otherwise, assume it's in tenths of a point.
				if pointSize >= 6 && pointSize <= 48 { // Common pixel sizes
					size = fmt.Sprintf("%.0fpx", pointSize)
				} else {
					size = fmt.Sprintf("%.0fpx", pointSize/10.0)
				}
			} else {
				size = "12px" // Default if conversion fails
			}
		} else if len(parts[11]) > 0 && parts[11] != "*" { // Fallback to average_width (XLFD field 11)
			if avgWidth, err := strconv.Atoi(parts[11]); err == nil {
				size = fmt.Sprintf("%dpx", avgWidth)
			} else {
				size = "12px" // Default if conversion fails
			}
		} else {
			size = "12px" // Default if neither is specified
		}

		// Family (usually the 3rd field, XLFD field 2)
		if len(parts[2]) > 0 && parts[2] != "*" {
			switch strings.ToLower(parts[2]) {
			case "helvetica", "arial", "lucida", "sans":
				family = "sans-serif"
			case "times", "serif", "charter", "new century schoolbook":
				family = "serif"
			case "courier", "typewriter", "lucidatypewriter", "mono":
				family = "monospace"
			default:
				family = strings.ToLower(parts[2])
			}
		} else {
			family = "sans-serif" // Default to sans-serif for wildcard family
		}

		// Weight (usually the 4th field)
		if len(parts[3]) > 0 && parts[3] != "*" {
			switch strings.ToLower(parts[3]) {
			case "medium":
				weight = "normal"
			case "bold", "demibold":
				weight = "bold"
			}
		}

		// Slant (usually the 5th field)
		if len(parts[4]) > 0 && parts[4] != "*" {
			switch strings.ToLower(parts[4]) {
			case "r": // Roman
				slant = "normal"
			case "i": // Italic
				slant = "italic"
			case "o": // Oblique
				slant = "oblique"
			}
		}
	}

	// Construct the CSS font string
	cssFont = fmt.Sprintf("%s %s %s %s", weight, slant, size, family)
	return
}

// GetAvailableFonts returns a hardcoded list of X11 font names.
func GetAvailableFonts() []string {
	return []string{
		"-*-helvetica-medium-r-normal-*-12-*-*-*-p-*-iso8859-1",
		"-*-helvetica-medium-r-normal-*-14-*-*-*-p-*-iso8859-1",
		"-*-helvetica-medium-r-normal-*-18-*-*-*-p-*-iso8859-1",
		"-*-helvetica-bold-r-normal-*-12-*-*-*-p-*-iso8859-1",
		"-*-courier-medium-r-normal-*-12-*-*-*-m-*-iso8859-1",
		"-*-courier-medium-r-normal-*-14-*-*-*-m-*-iso8859-1",
		"-*-times-medium-r-normal-*-12-*-*-*-p-*-iso8859-1",
		"-*-times-medium-r-normal-*-14-*-*-*-p-*-iso8859-1",
		"-*-lucida-medium-r-normal-*-12-*-*-*-p-*-iso8859-1",
		"-*-lucidatypewriter-medium-r-normal-*-12-*-*-*-m-*-iso8859-1",
		"-*-charter-medium-r-normal-*-12-*-*-*-p-*-iso8859-1",
		"-*-new century schoolbook-medium-r-normal-*-12-*-*-*-p-*-iso8859-1",
		"fixed",
		"variable",
		"9x15",  // Common fixed-width font
		"10x20", // Common fixed-width font
		"6x13",  // Common fixed-width font
		"7x14",  // Common fixed-width font
		"8x16",  // Common fixed-width font
		"lucidasans-10",
		"lucidasans-12",
		"lucidasans-14",
		"lucidasans-bold-10",
		"lucidasans-bold-12",
		"lucidasans-bold-14",
		"dejavu sans mono-10",
		"dejavu sans mono-12",
		"dejavu sans mono-14",
		"dejavu sans mono-bold-10",
		"dejavu sans mono-bold-12",
		"dejavu sans mono-bold-14",
		"monospace",  // Generic CSS fallback
		"sans-serif", // Generic CSS fallback
		"serif",      // Generic CSS fallback
	}
}
