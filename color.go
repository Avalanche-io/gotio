// SPDX-License-Identifier: Apache-2.0
// Copyright Contributors to the OpenTimelineIO project

package gotio

// Color represents an RGBA color.
type Color struct {
	R float64 `json:"r"`
	G float64 `json:"g"`
	B float64 `json:"b"`
	A float64 `json:"a"`
}

// NewColor creates a new Color.
func NewColor(r, g, b, a float64) *Color {
	return &Color{R: r, G: g, B: b, A: a}
}

// NewColorRGB creates a new Color with alpha 1.0.
func NewColorRGB(r, g, b float64) *Color {
	return &Color{R: r, G: g, B: b, A: 1.0}
}

// Predefined colors matching OpenTimelineIO's marker colors.
var (
	ColorPink    = NewColorRGB(1.0, 0.42, 0.78)
	ColorRed     = NewColorRGB(1.0, 0.13, 0.13)
	ColorOrange  = NewColorRGB(1.0, 0.55, 0.13)
	ColorYellow  = NewColorRGB(1.0, 0.87, 0.13)
	ColorGreen   = NewColorRGB(0.13, 0.87, 0.13)
	ColorCyan    = NewColorRGB(0.13, 0.87, 0.87)
	ColorBlue    = NewColorRGB(0.13, 0.55, 1.0)
	ColorPurple  = NewColorRGB(0.55, 0.13, 1.0)
	ColorMagenta = NewColorRGB(0.87, 0.13, 0.87)
	ColorBlack   = NewColorRGB(0.0, 0.0, 0.0)
	ColorWhite   = NewColorRGB(1.0, 1.0, 1.0)
)

// MarkerColor represents standard marker colors.
type MarkerColor string

const (
	MarkerColorPink    MarkerColor = "PINK"
	MarkerColorRed     MarkerColor = "RED"
	MarkerColorOrange  MarkerColor = "ORANGE"
	MarkerColorYellow  MarkerColor = "YELLOW"
	MarkerColorGreen   MarkerColor = "GREEN"
	MarkerColorCyan    MarkerColor = "CYAN"
	MarkerColorBlue    MarkerColor = "BLUE"
	MarkerColorPurple  MarkerColor = "PURPLE"
	MarkerColorMagenta MarkerColor = "MAGENTA"
	MarkerColorBlack   MarkerColor = "BLACK"
	MarkerColorWhite   MarkerColor = "WHITE"
)

// ToColor converts a MarkerColor to a Color.
func (mc MarkerColor) ToColor() *Color {
	switch mc {
	case MarkerColorPink:
		return ColorPink
	case MarkerColorRed:
		return ColorRed
	case MarkerColorOrange:
		return ColorOrange
	case MarkerColorYellow:
		return ColorYellow
	case MarkerColorGreen:
		return ColorGreen
	case MarkerColorCyan:
		return ColorCyan
	case MarkerColorBlue:
		return ColorBlue
	case MarkerColorPurple:
		return ColorPurple
	case MarkerColorMagenta:
		return ColorMagenta
	case MarkerColorBlack:
		return ColorBlack
	case MarkerColorWhite:
		return ColorWhite
	default:
		return ColorGreen
	}
}

// cloneColor creates a copy of a Color pointer.
func cloneColor(c *Color) *Color {
	if c == nil {
		return nil
	}
	return &Color{R: c.R, G: c.G, B: c.B, A: c.A}
}
