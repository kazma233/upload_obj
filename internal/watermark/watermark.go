package watermark

import (
	_ "embed"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"

	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/font/sfnt"
	"golang.org/x/image/math/fixed"

	_ "image/jpeg" // Register JPEG format
	_ "image/png"  // Register PNG format
)

type (
	WatermarkHandle struct {
		Text  string  `json:"text"`
		Size  float64 `json:"size"`
		Dpi   float64 `json:"dpi"`
		Color string  `json:"color"` // hex color; eg: #FF0000
		X     int     `json:"x"`
		Y     int     `json:"y"`
	}
)

var (
	//go:embed SourceHanSansCN-Normal.otf
	sourceHanSansCNNormal []byte
)

func (wh *WatermarkHandle) Check() bool {
	return wh.Text != ""
}

func (wh *WatermarkHandle) Do(bgImg image.Image) (image.Image, error) {
	face, err := NewFace(wh.Size, wh.Dpi)
	if err != nil {
		return nil, err
	}

	if wh.Text == "" {
		return nil, errors.New("text is empty")
	}

	if wh.Size <= 0 {
		wh.Size = 40
	}

	if wh.Dpi <= 0 {
		wh.Dpi = 100
	}

	if wh.Color == "" {
		wh.Color = "#000000"
	}

	if wh.X < 0 {
		wh.X = 100
	}

	if wh.Y < 0 {
		wh.Y = 100
	}

	return WriteWordMask(wh.Text, face, bgImg, wh.X, wh.Y, ParseHexColorFast(wh.Color)), nil
}

func NewFace(size float64, dpi float64) (font.Face, error) {
	parse, err := sfnt.Parse(sourceHanSansCNNormal)
	if err != nil {
		fmt.Printf("Failed to parse font: %v\n", err)
		// return basicfont.Face7x13, nil
		return nil, err
	}
	return opentype.NewFace(parse, &opentype.FaceOptions{
		Size:    size,
		DPI:     dpi,
		Hinting: font.HintingNone,
	})
}

func WriteWordMask(word string, face font.Face, bgImg image.Image, x, y int, color color.Color) image.Image {
	// Calculate text bounds
	bounds, _ := font.BoundString(face, word)
	textWidth := (bounds.Max.X - bounds.Min.X).Ceil()
	textHeight := (bounds.Max.Y - bounds.Min.Y).Ceil()

	// Create a new RGBA image with the same size as the original
	dstImg := image.NewRGBA(bgImg.Bounds())

	// Copy the original image to the new image
	boundsImg := bgImg.Bounds()
	for py := boundsImg.Min.Y; py < boundsImg.Max.Y; py++ {
		for px := boundsImg.Min.X; px < boundsImg.Max.X; px++ {
			dstImg.Set(px, py, bgImg.At(px, py))
		}
	}

	// Create a drawer for the text
	drawer := &font.Drawer{
		Dst:  dstImg,
		Src:  image.NewUniform(color),
		Face: face,
	}

	// Adjust position to center the text if x or y is negative
	if x < 0 {
		x = (bgImg.Bounds().Dx() - textWidth) / 2
	}
	if y < 0 {
		y = (bgImg.Bounds().Dy()-textHeight)/2 + (textHeight * 4 / 5) // Adjust for baseline
	} else {
		y += textHeight // Adjust for baseline
	}

	// Draw the text
	drawer.Dot = fixed.P(x, y)
	drawer.DrawString(word)

	return dstImg
}

func ParseHexColorFast(s string) color.RGBA {
	var c color.RGBA

	c.A = 0xff

	if s[0] != '#' {
		return c
	}

	switch len(s) {
	case 7:
		c.R = hexToByte(s[1])<<4 + hexToByte(s[2])
		c.G = hexToByte(s[3])<<4 + hexToByte(s[4])
		c.B = hexToByte(s[5])<<4 + hexToByte(s[6])
	case 4:
		c.R = hexToByte(s[1]) * 17
		c.G = hexToByte(s[2]) * 17
		c.B = hexToByte(s[3]) * 17
	}

	return c
}

func hexToByte(b byte) byte {
	switch {
	case b >= '0' && b <= '9':
		return b - '0'
	case b >= 'a' && b <= 'f':
		return b - 'a' + 10
	case b >= 'A' && b <= 'F':
		return b - 'A' + 10
	}

	return '0'
}

func OpenImg(filePath string) (image.Image, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	bgImg, _, err := image.Decode(file)
	if err != nil {
		return nil, err
	}
	return bgImg, nil
}

func Save2PngImg(img image.Image, filePath string) error {
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()
	return png.Encode(file, img)
}
