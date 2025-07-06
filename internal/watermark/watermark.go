package watermark

import (
	_ "embed"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"math"
	"os"

	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/font/sfnt"
	"golang.org/x/image/math/fixed"

	_ "image/jpeg" // Register JPEG format
	_ "image/png"  // Register PNG format
)

type (
	Position string

	WatermarkHandle struct {
		Text  string   `json:"text"`
		Size  float64  `json:"size"`
		Dpi   float64  `json:"dpi"`
		Color string   `json:"color"` // hex color; eg: #FF0000
		X     int      `json:"x"`
		Y     int      `json:"y"`
		Posi  Position `json:"position"`
		Angle float64  `json:"angle"` // 旋转角度，单位度
	}
)

func (p *Position) String() string {
	return string(*p)
}

var (
	//go:embed SourceHanSansCN-Normal.otf
	sourceHanSansCNNormal []byte

	LeftTop     Position = "LeftTop"
	LeftBottom  Position = "LeftBottom"
	RightTop    Position = "RightTop"
	RightBottom Position = "RightBottom"
	Center      Position = "Center"
	Full        Position = "Full"
)

func (wh *WatermarkHandle) Check() bool {
	return wh.Text != ""
}

func (wh *WatermarkHandle) Do(filePath string) (image.Image, error) {
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

	// 使用filePath创建dstImg
	bgImg, err := OpenImg(filePath)
	if err != nil {
		return nil, err
	}
	dstImg := image.NewRGBA(bgImg.Bounds())
	draw.Draw(dstImg, bgImg.Bounds(), bgImg, image.Point{}, draw.Src)

	if wh.Posi != "" {
		WriteWordMaskEasy(wh.Text, face, dstImg, wh.Posi, ParseHexColorFast(wh.Color), wh.Angle)
		return dstImg, nil
	}

	if wh.X < 0 {
		wh.X = 100
	}

	if wh.Y < 0 {
		wh.Y = 100
	}
	WriteWordMask(wh.Text, face, dstImg, wh.X, wh.Y, ParseHexColorFast(wh.Color), wh.Angle)
	return dstImg, nil
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

func WriteWordMaskEasy(word string, face font.Face, dst draw.Image, position Position, color color.Color, angle float64) {
	bgImgBounds := dst.Bounds()

	dx, dy := int(float64(bgImgBounds.Max.X)*0.1), int(float64(bgImgBounds.Max.Y)*0.1)
	x, y := dx, dy
	ang := angle

	if position == Full {
		// 平铺水印
		bounds, _ := font.BoundString(face, word)
		textWidth := (bounds.Max.X - bounds.Min.X).Ceil()
		textHeight := (bounds.Max.Y - bounds.Min.Y).Ceil()
		intervalX := textWidth + 40  // 水印横向间隔
		intervalY := textHeight + 40 // 水印纵向间隔

		for y := 0; y < bgImgBounds.Dy(); y += intervalY {
			for x := 0; x < bgImgBounds.Dx(); x += intervalX {
				WriteWordMask(word, face, dst, x, y, color, ang)
			}
		}
		return
	}

	switch position {
	case LeftTop:
		x, y = dx, dy
	case LeftBottom:
		x, y = dx, bgImgBounds.Max.Y-dy
	case RightTop:
		x, y = bgImgBounds.Max.X-dx, dy
	case RightBottom:
		x, y = bgImgBounds.Max.X-dx, bgImgBounds.Max.Y-dy
	case Center:
		x, y = (bgImgBounds.Max.X-dx)/2, (bgImgBounds.Max.Y-dy)/2
	}

	WriteWordMask(word, face, dst, x, y, color, ang)
}

// 修改为直接在传入的dst上绘制水印，不再返回新图片
func WriteWordMask(word string, face font.Face, dst draw.Image, x, y int, color color.Color, angle float64) {
	// Calculate text bounds
	bounds, _ := font.BoundString(face, word)
	textWidth := (bounds.Max.X - bounds.Min.X).Ceil()
	textHeight := (bounds.Max.Y - bounds.Min.Y).Ceil()

	if angle == 0 {
		// Create a drawer for the text
		drawer := &font.Drawer{
			Dst:  dst,
			Src:  image.NewUniform(color),
			Face: face,
		}

		// Adjust position to center the text if x or y is negative
		if x < 0 {
			x = (dst.Bounds().Dx() - textWidth) / 2
		}
		if y < 0 {
			y = (dst.Bounds().Dy()-textHeight)/2 + (textHeight * 4 / 5) // Adjust for baseline
		} else {
			y += textHeight // Adjust for baseline
		}

		// Draw the text
		drawer.Dot = fixed.P(x, y)
		drawer.DrawString(word)
		return
	}

	// 旋转模式：先绘制到足够大的小图再旋转粘贴，避免裁剪
	diag := int(math.Ceil(math.Hypot(float64(textWidth), float64(textHeight))))
	canvasW, canvasH := diag, diag
	txtImg := image.NewRGBA(image.Rect(0, 0, canvasW, canvasH))
	// 将文本绘制在中心
	drawer := &font.Drawer{
		Dst:  txtImg,
		Src:  image.NewUniform(color),
		Face: face,
		Dot:  fixed.P((canvasW-textWidth)/2, (canvasH+textHeight*4/5)/2), // baseline居中
	}
	drawer.DrawString(word)
	rotImg := RotateImage(txtImg, angle)
	// 粘贴到目标图
	offset := image.Pt(x, y)
	for py := 0; py < rotImg.Bounds().Dy(); py++ {
		for px := 0; px < rotImg.Bounds().Dx(); px++ {
			c := rotImg.At(px, py)
			_, _, _, a := c.RGBA()
			if a > 0 {
				dst.Set(px+offset.X, py+offset.Y, c)
			}
		}
	}
}

// 旋转文本绘制辅助函数
func DrawRotatedText(dst *image.RGBA, word string, face font.Face, x, y int, color color.Color, angle float64) {
	// 1. 先将文本绘制到一个小图
	bounds, _ := font.BoundString(face, word)
	w := (bounds.Max.X - bounds.Min.X).Ceil()
	h := (bounds.Max.Y - bounds.Min.Y).Ceil()
	if w <= 0 || h <= 0 {
		return
	}
	txtImg := image.NewRGBA(image.Rect(0, 0, w, h))
	drawer := &font.Drawer{
		Dst:  txtImg,
		Src:  image.NewUniform(color),
		Face: face,
		Dot:  fixed.P(0, h*4/5), // baseline
	}
	drawer.DrawString(word)

	// 2. 旋转小图
	rotImg := RotateImage(txtImg, angle)

	// 3. 粘贴到目标图
	offset := image.Pt(x, y)
	for py := 0; py < rotImg.Bounds().Dy(); py++ {
		for px := 0; px < rotImg.Bounds().Dx(); px++ {
			c := rotImg.At(px, py)
			_, _, _, a := c.RGBA()
			if a > 0 {
				dst.Set(px+offset.X, py+offset.Y, c)
			}
		}
	}
}

// 简单最近邻旋转实现（角度为度）
func RotateImage(src *image.RGBA, angle float64) *image.RGBA {
	rad := angle * 3.14159265 / 180.0
	cosA := math.Cos(rad)
	sinA := math.Sin(rad)
	w, h := src.Bounds().Dx(), src.Bounds().Dy()
	cx, cy := float64(w)/2, float64(h)/2
	rotImg := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			dx := float64(x) - cx
			dy := float64(y) - cy
			srcX := int(cx + dx*cosA - dy*sinA)
			srcY := int(cy + dx*sinA + dy*cosA)
			if srcX >= 0 && srcX < w && srcY >= 0 && srcY < h {
				rotImg.Set(x, y, src.At(srcX, srcY))
			}
		}
	}
	return rotImg
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
