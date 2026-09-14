package processor

import (
	"bytes"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"

	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
)

type Processor struct{}

func NewProcessor() *Processor {
	return &Processor{}
}

func (p *Processor) Resize(img image.Image, width, height int) image.Image {
	bounds := img.Bounds()
	srcW, srcH := bounds.Dx(), bounds.Dy()

	if width == 0 && height == 0 {
		return img
	}

	if width == 0 {
		width = srcW * height / srcH
	}
	if height == 0 {
		height = srcH * width / srcW
	}

	dst := image.NewRGBA(image.Rect(0, 0, width, height))
	scaleX := float64(width) / float64(srcW)
	scaleY := float64(height) / float64(srcH)

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			srcX := int(float64(x) / scaleX)
			srcY := int(float64(y) / scaleY)
			if srcX < srcW && srcY < srcH {
				dst.Set(x, y, img.At(bounds.Min.X+srcX, bounds.Min.Y+srcY))
			}
		}
	}
	return dst
}

func (p *Processor) Thumbnail(img image.Image, size int) image.Image {
	bounds := img.Bounds()
	srcW, srcH := bounds.Dx(), bounds.Dy()

	var width, height int
	if srcW > srcH {
		width = size
		height = size * srcH / srcW
	} else {
		height = size
		width = size * srcW / srcH
	}

	return p.Resize(img, width, height)
}

func (p *Processor) AddWatermark(img image.Image, text string) image.Image {
	bounds := img.Bounds()
	dst := image.NewRGBA(bounds)
	draw.Draw(dst, bounds, img, bounds.Min, draw.Src)

	point := fixed.Point26_6{
		X: fixed.I(bounds.Dx() - 150),
		Y: fixed.I(bounds.Dy() - 10),
	}

	d := &font.Drawer{
		Dst:  dst,
		Src:  image.NewUniform(color.RGBA{255, 255, 255, 128}),
		Face: basicfont.Face7x13,
		Dot:  point,
	}
	d.DrawString(text)

	return dst
}

func (p *Processor) ProcessImage(inputPath, outputPath string, width, height, thumbSize int, watermark string) error {
	file, err := os.Open(inputPath)
	if err != nil {
		return err
	}
	defer file.Close()

	var img image.Image
	ext := filepath.Ext(inputPath)
	switch ext {
	case ".jpg", ".jpeg":
		img, err = jpeg.Decode(file)
	case ".png":
		img, err = png.Decode(file)
	default:
		img, _, err = image.Decode(file)
	}
	if err != nil {
		return err
	}

	processed := img
	if width > 0 || height > 0 {
		processed = p.Resize(processed, width, height)
	}
	if watermark != "" {
		processed = p.AddWatermark(processed, watermark)
	}

	return p.saveImage(processed, outputPath)
}

func (p *Processor) ProcessThumbnail(inputPath, outputPath string, size int) error {
	file, err := os.Open(inputPath)
	if err != nil {
		return err
	}
	defer file.Close()

	var img image.Image
	ext := filepath.Ext(inputPath)
	switch ext {
	case ".jpg", ".jpeg":
		img, err = jpeg.Decode(file)
	case ".png":
		img, err = png.Decode(file)
	default:
		img, _, err = image.Decode(file)
	}
	if err != nil {
		return err
	}

	thumb := p.Thumbnail(img, size)
	return p.saveImage(thumb, outputPath)
}

func (p *Processor) saveImage(img image.Image, path string) error {
	ext := filepath.Ext(path)
	buf := new(bytes.Buffer)

	switch ext {
	case ".jpg", ".jpeg":
		err := jpeg.Encode(buf, img, &jpeg.Options{Quality: 90})
		if err != nil {
			return err
		}
	case ".png":
		err := png.Encode(buf, img)
		if err != nil {
			return err
		}
	default:
		err := jpeg.Encode(buf, img, &jpeg.Options{Quality: 90})
		if err != nil {
			return err
		}
	}

	return os.WriteFile(path, buf.Bytes(), 0644)
}

