package levels

import "image"

// AverageColour gets the average colour from an image.
// Return values are R, G, B, A respectively.
func AverageColour(img image.Image) (red uint8, green uint8, blue uint8, alpha uint8) {
	bounds := img.Bounds()
	minX, minY := bounds.Min.X, bounds.Min.Y
	maxX, maxY := bounds.Max.X, bounds.Max.Y

	var pixels int
	var r, g, b, a int

	for x := minX; x < maxX; x++ {
		for y := minY; y < maxY; y++ {
			if rd, gr, bl, al := img.At(x, y).RGBA(); al != 0 {
				pixels++

				r += int(rd >> 8)
				g += int(gr >> 8)
				b += int(bl >> 8)
				a += int(al >> 8)
			}
		}
	}

	if pixels == 0 {
		return 0, 0, 0, 0
	}

	return uint8(r / pixels), uint8(g / pixels), uint8(b / pixels), uint8(a / pixels)
}
