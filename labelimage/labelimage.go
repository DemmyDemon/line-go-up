package labelimage

import (
	"image"
	"image/color"
	"image/draw"

	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
)

func DrawBorder(img *image.RGBA, borderWidth int, col color.RGBA) {
	max := img.Bounds().Max
	imgWidth := max.X
	imgHeight := max.Y
	for x := 0; x < imgWidth; x++ {
		for y := 0; y < imgHeight; y++ {
			switch {
			case x < borderWidth:
				img.Set(x, y, col)
			case x >= imgWidth-borderWidth:
				img.Set(x, y, col)
			case y < borderWidth:
				img.Set(x, y, col)
			case y >= imgHeight-borderWidth:
				img.Set(x, y, col)
			}
		}
	}
}

func Crosshair(img *image.RGBA) {
	max := img.Bounds().Max
	halfHeight := int(max.X / 2)
	halfWidth := int(max.Y / 2)
	col := color.RGBA{255, 0, 0, 255}
	for x := 0; x < max.X; x++ {
		for y := 0; y < max.Y; y++ {
			switch {
			case y == halfWidth:
				img.Set(x, y, col)
			case x == halfHeight:
				img.Set(x, y, col)
			}
		}
	}
}

func Create(size image.Rectangle, textColor color.RGBA, text string, border bool, shadow bool) *image.RGBA {
	img := image.NewRGBA(size)
	bounds := size.Bounds()
	draw.Draw(img, bounds, &image.Uniform{color.Transparent}, image.Point{}, draw.Src)
	offset := int(len(text)/2) * 7
	point := fixed.Point26_6{X: fixed.I((bounds.Max.X / 2) - offset), Y: fixed.I((bounds.Max.Y / 2) + 6)}

	if border {
		DrawBorder(img, 3, textColor)
	}

	if shadow {
		shadowPoint := fixed.Point26_6{X: fixed.I((bounds.Max.X / 2) - (offset - 1)), Y: fixed.I((bounds.Max.Y / 2) + 7)}
		shadowDrawer := &font.Drawer{
			Dst:  img,
			Src:  image.NewUniform(color.RGBA{0, 0, 0, 100}),
			Face: basicfont.Face7x13,
			Dot:  shadowPoint,
		}
		shadowDrawer.DrawString(text)
	}
	drawer := &font.Drawer{
		Dst:  img,
		Src:  image.NewUniform(textColor),
		Face: basicfont.Face7x13,
		Dot:  point,
	}
	drawer.DrawString(text)
	// Crosshair(img)
	return img
}
