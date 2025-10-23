//go:build x11

package x11

import "strings"

type rgb8bit struct {
	Red, Green, Blue uint8
}

var colorNames = map[string]rgb8bit{
	"black":           {0, 0, 0},
	"white":           {255, 255, 255},
	"red":             {255, 0, 0},
	"green":           {0, 255, 0},
	"blue":            {0, 0, 255},
	"yellow":          {255, 255, 0},
	"cyan":            {0, 255, 255},
	"magenta":         {255, 0, 255},
	"saddle brown":    {139, 69, 19},
	"sienna":          {160, 82, 45},
	"chocolate":       {210, 105, 30},
	"peru":            {205, 133, 63},
	"sandy brown":     {244, 164, 96},
	"burlywood":       {222, 184, 135},
	"tan":             {210, 180, 140},
	"rosy brown":      {188, 143, 143},
	"moccasin":        {255, 228, 181},
	"navajo white":    {255, 222, 173},
	"peach puff":      {255, 218, 185},
	"misty rose":      {255, 228, 225},
	"lavender blush":  {255, 240, 245},
	"linen":           {250, 240, 230},
	"old lace":        {253, 245, 230},
	"papaya whip":     {255, 239, 213},
	"blanched almond": {255, 235, 205},
	"bisque":          {255, 228, 196},
	"wheat":           {245, 222, 179},
	"cornsilk":        {255, 248, 220},
	"lemon chiffon":   {255, 250, 205},
	"light goldenrod": {250, 250, 210},
	"light yellow":    {255, 255, 224},
}

func lookupColor(name string) (rgb8bit, bool) {
	c, ok := colorNames[strings.ToLower(name)]
	return c, ok
}

// scale8to16 scales an 8-bit color component to a 16-bit color component
func scale8to16(c uint8) uint16 {
	return uint16(c) | (uint16(c) << 8)
}
