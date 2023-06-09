package labelimage

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"

	"github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
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

type FontDescription struct {
	Font    *truetype.Font
	DPI     float64
	Hinting font.Hinting
	Size    float64
	Ratio   float64
}

// PrepareFreetypeContext sets up all the bits and bobs related to drawing text on the image
func PrepareFreetypeContext(dst *image.RGBA, src image.Image, font FontDescription) (*freetype.Context, int) {
	c := freetype.NewContext()
	c.SetDPI(font.DPI)
	c.SetFont(font.Font)
	c.SetHinting(font.Hinting)
	c.SetFontSize(font.Size)
	c.SetSrc(src)
	c.SetDst(dst)
	c.SetClip(dst.Bounds())

	baseline := (int(c.PointToFixed(font.Size) >> 6))

	return c, baseline
}

// DrawText draws the given text in the given context, at the given location
func DrawText(c *freetype.Context, x int, y int, text string) error {
	pt := freetype.Pt(x, y)
	_, err := c.DrawString(text, pt)
	if err != nil {
		return fmt.Errorf("drawing text: %w", err)
	}
	return nil
}

func CreateWithFont(size image.Rectangle, fontFace *truetype.Font, textColor color.RGBA, text []string, border bool, shadow bool) *image.RGBA {

	if len(text) > 3 {
		text = text[:3]
	}

	img := image.NewRGBA(size)
	bounds := size.Bounds()
	draw.Draw(img, bounds, &image.Uniform{color.Transparent}, image.Point{}, draw.Src)

	if border {
		DrawBorder(img, 2, textColor)
	}

	fontDescription := FontDescription{
		Font:    fontFace,
		DPI:     72.0,
		Hinting: font.HintingNone,
		Size:    19.0,
		Ratio:   0.65,
	}
	opts := &truetype.Options{
		Size: 19.0,
	}
	face := truetype.NewFace(fontFace, opts)
	advance, _ := face.GlyphAdvance('M')

	center := float64(size.Max.X / 2)

	if shadow {
		shadowCtx, baseline := PrepareFreetypeContext(img, image.NewUniform(color.RGBA{0, 0, 0, 60}), fontDescription)
		for i, line := range text {
			lineSize := float64(len(line)) * (float64(advance) / 64.0)
			offset := int(center - (lineSize / 2))
			DrawText(shadowCtx, 4+offset, (i*baseline)+baseline+1, line) // FIXME: Don't ignore the error!
		}
	}

	ctx, baseline := PrepareFreetypeContext(img, &image.Uniform{textColor}, fontDescription)
	for i, line := range text {
		lineSize := float64(len(line)) * (float64(advance) / 64.0)
		offset := int(center - (lineSize / 2))
		DrawText(ctx, 3+offset, (i*baseline)+baseline, line)
	}

	return img

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
