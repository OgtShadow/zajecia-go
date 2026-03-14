package main

import (
	"image"
	"image/color"

	"golang.org/x/tour/pic"
)

// Image implements image.Image.
type Image struct{ W, H int }

func (m Image) ColorModel() color.Model { return color.RGBAModel }
func (m Image) Bounds() image.Rectangle { return image.Rect(0, 0, m.W, m.H) }
func (m Image) At(x, y int) color.Color {
	v := uint8(x ^ y)
	return color.RGBA{v, v, 255, 255}
}

func main() {
	m := Image{W: 128, H: 128}
	pic.ShowImage(m)
}
