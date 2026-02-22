// Package imageutil: normalize images before sending to iApp
package utils

import (
	"bytes"
	"errors"
	"github.com/rwcarlsen/goexif/exif"
	"golang.org/x/image/draw"
	"golang.org/x/image/webp"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"math"
)

// Features:
// - Decode JPEG/PNG/WebP
// - Apply EXIF Orientation (rotate/flip)
// - Resize (keep aspect) with maxWidth
// - Encode to JPEG (quality)
//
// Deps:
//   go get github.com/rwcarlsen/goexif/exif
//   go get golang.org/x/image/draw
//   go get golang.org/x/image/webp

// NormalizeToJPG decodes input (jpg/png/webp), applies EXIF orientation, resizes to maxWidth (if > 0),
// then encodes to JPEG with given quality (1..100). Returns JPEG bytes.
func NormalizeToJPG(input []byte, maxWidth int, quality int) ([]byte, error) {
	if len(input) == 0 {
		return nil, errors.New("empty image")
	}
	if quality <= 0 || quality > 100 {
		quality = 85
	}

	// 1) Decode
	img, format, err := decodeAny(bytes.NewReader(input))
	if err != nil {
		return nil, err
	}

	// 2) Read EXIF orientation (only meaningful for JPEG mostly, but harmless otherwise)
	ori := readEXIFOrientation(bytes.NewReader(input))
	img = applyOrientation(img, ori)

	// 3) Resize (optional)
	if maxWidth > 0 {
		img = resizeMaxWidth(img, maxWidth)
	}

	// 4) Encode as JPEG
	var out bytes.Buffer
	if err := jpeg.Encode(&out, img, &jpeg.Options{Quality: quality}); err != nil {
		return nil, err
	}

	_ = format // in case you want to log it
	return out.Bytes(), nil
}

func decodeAny(r io.Reader) (image.Image, string, error) {
	// We need a ReadSeeker? Not if we decode from bytes.NewReader above.
	// Try standard formats first:
	if img, fmt, err := image.Decode(r); err == nil {
		return img, fmt, nil
	}

	// If image.Decode failed because WebP decoder isn't registered in stdlib,
	// handle WebP explicitly.
	// NOTE: Since r is already consumed, caller should pass a fresh reader (we do).
	return nil, "", errors.New("unsupported image format (need jpeg/png/webp)")
}

// If you want robust decode (jpeg/png/webp) without relying on image.Decode registration:
func decodeAnyStrict(r *bytes.Reader) (image.Image, string, error) {
	// Try JPEG
	r.Seek(0, io.SeekStart)
	if img, err := jpeg.Decode(r); err == nil {
		return img, "jpeg", nil
	}
	// Try PNG
	r.Seek(0, io.SeekStart)
	if img, err := png.Decode(r); err == nil {
		return img, "png", nil
	}
	// Try WebP
	r.Seek(0, io.SeekStart)
	if img, err := webp.Decode(r); err == nil {
		return img, "webp", nil
	}
	return nil, "", errors.New("unsupported image format (jpeg/png/webp)")
}

func readEXIFOrientation(r io.Reader) int {
	// Default orientation = 1 (no transform)
	x, err := exif.Decode(r)
	if err != nil {
		return 1
	}
	tag, err := x.Get(exif.Orientation)
	if err != nil {
		return 1
	}
	ori, err := tag.Int(0)
	if err != nil {
		return 1
	}
	return ori
}

// EXIF orientation values:
// 1 = normal
// 2 = flip horizontal
// 3 = rotate 180
// 4 = flip vertical
// 5 = transpose (flip horizontal + rotate 90 CW)
// 6 = rotate 90 CW
// 7 = transverse (flip horizontal + rotate 90 CCW)
// 8 = rotate 270 CW (90 CCW)
func applyOrientation(src image.Image, ori int) image.Image {
	switch ori {
	case 2:
		return flipHorizontal(src)
	case 3:
		return rotate180(src)
	case 4:
		return flipVertical(src)
	case 5:
		return rotate90CW(flipHorizontal(src))
	case 6:
		return rotate90CW(src)
	case 7:
		return rotate90CCW(flipHorizontal(src))
	case 8:
		return rotate90CCW(src)
	default:
		return src
	}
}

func resizeMaxWidth(src image.Image, maxW int) image.Image {
	b := src.Bounds()
	w := b.Dx()
	h := b.Dy()
	if w <= 0 || h <= 0 {
		return src
	}
	if w <= maxW {
		return src
	}

	scale := float64(maxW) / float64(w)
	newW := maxW
	newH := int(math.Round(float64(h) * scale))
	if newH < 1 {
		newH = 1
	}

	dst := image.NewRGBA(image.Rect(0, 0, newW, newH))
	// High quality resampling
	draw.CatmullRom.Scale(dst, dst.Bounds(), src, b, draw.Over, nil)
	return dst
}

func rotate90CW(src image.Image) image.Image {
	b := src.Bounds()
	w, h := b.Dx(), b.Dy()
	dst := image.NewRGBA(image.Rect(0, 0, h, w))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			// (x,y) -> (h-1-y, x)
			dst.Set(h-1-y, x, src.At(b.Min.X+x, b.Min.Y+y))
		}
	}
	return dst
}

func rotate90CCW(src image.Image) image.Image {
	b := src.Bounds()
	w, h := b.Dx(), b.Dy()
	dst := image.NewRGBA(image.Rect(0, 0, h, w))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			// (x,y) -> (y, w-1-x)
			dst.Set(y, w-1-x, src.At(b.Min.X+x, b.Min.Y+y))
		}
	}
	return dst
}

func rotate180(src image.Image) image.Image {
	b := src.Bounds()
	w, h := b.Dx(), b.Dy()
	dst := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			dst.Set(w-1-x, h-1-y, src.At(b.Min.X+x, b.Min.Y+y))
		}
	}
	return dst
}

func flipHorizontal(src image.Image) image.Image {
	b := src.Bounds()
	w, h := b.Dx(), b.Dy()
	dst := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			dst.Set(w-1-x, y, src.At(b.Min.X+x, b.Min.Y+y))
		}
	}
	return dst
}

func flipVertical(src image.Image) image.Image {
	b := src.Bounds()
	w, h := b.Dx(), b.Dy()
	dst := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			dst.Set(x, h-1-y, src.At(b.Min.X+x, b.Min.Y+y))
		}
	}
	return dst
}
