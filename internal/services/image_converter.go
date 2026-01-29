package services

import (
	"errors"
	"image"
	"image/draw"
	_ "image/gif"
	_ "image/jpeg"
	"image/png"
	_ "image/png"
	"os"
	"path/filepath"
	"strconv"

	_ "golang.org/x/image/bmp"
	_ "golang.org/x/image/tiff"
	_ "golang.org/x/image/webp"
)

//const asciiRamp := "$@B%8&WM#*oahkbdpqwmZO0QLCJUYXzcvunxrjft/\|()1{}[]?-_+~<>i!lI;:,"^`. "
//const unicodeRamp := "█▉▊▋▌▍▎▏▓▒░■□@&%$#*+=-~:;!,\".^`' "

type ConvertDoneMsg struct {
	Err error
}

func ConvertImageToString(filePath string) error {
	isDebugEnv, _ := Shared().Get("debug")

	f, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()

	_ = Logger().Info("Successfully Loaded: " + filePath)

	inputImg, format, err := image.Decode(f)
	if err != nil {
		return err
	}

	imgBounds := inputImg.Bounds()
	imgWidth, imgHeight := imgBounds.Dx(), imgBounds.Dy()
	_ = Logger().Info("format: " + format)
	_ = Logger().Info("imgWidth: " + strconv.Itoa(imgWidth) + " imgHeight: " + strconv.Itoa(imgHeight))

	textSizeAny, ok := Shared().Get("textSize")
	if !ok || textSizeAny == nil {
		return errors.New("textSize is nil")
	}

	textSize, ok := textSizeAny.(int)
	if !ok || textSize <= 0 {
		textSize = 8
	}

	fontAspect := 2.0
	charWidth := textSize
	charHeight := int(float64(textSize) * fontAspect)

	//TODO DO I NEED TO DOWNSCALE for performance ?

	var rects []image.Rectangle
	for y := 0; y < imgHeight; y += charHeight {
		y1 := y + charHeight
		if y1 > imgHeight {
			y1 = imgHeight
		}
		for x := 0; x < imgWidth; x += charWidth {
			x1 := x + charWidth
			if x1 > imgWidth {
				x1 = imgWidth
			}

			rects = append(rects, image.Rect(
				imgBounds.Min.X+x, imgBounds.Min.Y+y,
				imgBounds.Min.X+x1, imgBounds.Min.Y+y1,
			))
		}
	}

	for i, r := range rects {
		var tile image.Image

		if si, ok := inputImg.(interface {
			SubImage(r image.Rectangle) image.Image
		}); ok {
			tile = si.SubImage(r)
		} else {
			// Fallback: copy crop if subimage fails
			tile = cropToRGBA(inputImg, r)
		}

		if isDebugEnv.(bool) {
			err := saveImageToDebugDir(tile, "image_"+strconv.Itoa(i)+".png")
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func cropToRGBA(src image.Image, r image.Rectangle) *image.RGBA {
	dst := image.NewRGBA(image.Rect(0, 0, r.Dx(), r.Dy()))
	draw.Draw(dst, dst.Bounds(), src, r.Min, draw.Src)
	return dst
}

func saveImageToDebugDir(img image.Image, outputName string) error {
	if filepath.Ext(outputName) == "" {
		outputName += ".png"
	}
	if err := os.MkdirAll("debugFolder/cropped_img", 0o755); err != nil {
		return err
	}

	outPath := filepath.Join("debugFolder/cropped_img", outputName)
	out, err := os.Create(outPath)
	if err != nil {
		return err
	}
	defer func() { _ = out.Close() }()

	return png.Encode(out, img)
}
