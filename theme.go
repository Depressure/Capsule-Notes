package main

import "image/color"

// Dark sci-fi palette inspired by the EVE Online client UI.
var palette = struct {
	bg        color.NRGBA
	panel     color.NRGBA
	tab       color.NRGBA
	tabHover  color.NRGBA
	tabActive color.NRGBA
	line      color.NRGBA
	fg        color.NRGBA
	dim       color.NRGBA
	hint      color.NRGBA
	accent    color.NRGBA
	gold      color.NRGBA
	selection color.NRGBA
}{
	bg:        color.NRGBA{R: 0x0e, G: 0x12, B: 0x17, A: 0xff},
	panel:     color.NRGBA{R: 0x14, G: 0x19, B: 0x21, A: 0xff},
	tab:       color.NRGBA{R: 0x1a, G: 0x21, B: 0x2b, A: 0xff},
	tabHover:  color.NRGBA{R: 0x20, G: 0x28, B: 0x34, A: 0xff},
	tabActive: color.NRGBA{R: 0x24, G: 0x2e, B: 0x3c, A: 0xff},
	line:      color.NRGBA{R: 0x2a, G: 0x33, B: 0x41, A: 0xff},
	fg:        color.NRGBA{R: 0xd6, G: 0xdc, B: 0xe4, A: 0xff},
	dim:       color.NRGBA{R: 0x8a, G: 0x94, B: 0xa3, A: 0xff},
	hint:      color.NRGBA{R: 0x5c, G: 0x66, B: 0x74, A: 0xff},
	accent:    color.NRGBA{R: 0x4f, G: 0x8f, B: 0xc4, A: 0xff},
	gold:      color.NRGBA{R: 0xd8, G: 0xb2, B: 0x5c, A: 0xff},
	selection: color.NRGBA{R: 0x2c, G: 0x4a, B: 0x6b, A: 0xff},
}
