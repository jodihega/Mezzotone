package services

import (
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	_ "image/gif"
	_ "image/jpeg"
	"image/png"
	_ "image/png"
	"os"
	"path/filepath"
	"sort"
	"strconv"

	_ "golang.org/x/image/bmp"
	_ "golang.org/x/image/tiff"
	_ "golang.org/x/image/webp"
)

const asciiRampDarkToBrightStr = "$@B%8&WM#*oahkbdpqwmZO0QLCJUYXzcvunxrjft/\\|()1{}[]?-_+~<>i!lI;:,^`. "
const unicodeRampDarkToBrightStr = "█▉▊▋▌▍▎▏▓▒░■□@&%$#*+=-~:;!,\".^`' "

const asciiRampBrightToDarkStr = " .`^,:;Il!i><~+_-?][}{1)(|\\/tfjrxnuvczXYUJCLQ0OZmwqpdbkhao*#MW&8%B@$"
const unicodeRampBrightToDarkStr = " '`^.\",!;:~-=+*#$%&@□■░▒▓▏▎▍▌▋▊▉█"

type ConvertDoneMsg struct {
	Err error
}

type ConvertedImageArray struct {
	Cols       int
	Rows       int
	Characters []rune
}

/*
TODO if unicode is true i could make a bunch of different ramps for the user to choose from.
	Example:
	█▇▆▅▄▃▂▁ ▁▂▃▄▅▆▇█
	█▓▒░ ░▒▓█
	⣿⣷⣧⣇⣆⣄⣀ ⣀⣄⣆⣇⣧⣷⣿
	●∙•·  ·•∙●
*/

func ConvertImageToString(filePath string) (ConvertedImageArray, error) {
	isDebugEnvAny, _ := Shared().Get("debug")
	isDebugEnv := isDebugEnvAny.(bool)

	f, err := os.Open(filePath)
	if err != nil {
		return ConvertedImageArray{}, err
	}
	defer func() { _ = f.Close() }()

	_ = Logger().Info(fmt.Sprintf("Successfully Loaded: %s", filePath))

	inputImg, format, err := image.Decode(f)
	if err != nil {
		return ConvertedImageArray{}, err
	}
	_ = Logger().Info(fmt.Sprintf("format: %s", format))

	textSizeAny, ok := Shared().Get("textSize")
	if !ok || textSizeAny == nil {
		return ConvertedImageArray{}, errors.New("textSize is nil")
	}
	textSize, ok := textSizeAny.(int)
	if !ok || textSize <= 0 {
		textSize = 8
	}

	fontAspectAny, ok := Shared().Get("fontAspect")
	if !ok || fontAspectAny == nil {
		return ConvertedImageArray{}, errors.New("fontAspect is nil")
	}
	fontAspect, ok := fontAspectAny.(float64)
	if !ok || fontAspect <= 0 {
		fontAspect = 2
	}

	cols, rows := gridFromTextSize(inputImg, textSize, fontAspect)

	//TODO if directionalRender is true apply Sobel per tile or pool full-res edges into the ASCII grid

	splitImages, err := splitImage(textSize, fontAspect, inputImg, isDebugEnv)
	if err != nil {
		return ConvertedImageArray{}, err
	}
	_ = Logger().Info(fmt.Sprintf("Successfully Split Image: %s", filePath))

	_ = Logger().Info(fmt.Sprintf("Beginning image conversion"))
	useUnicodeAny, ok := Shared().Get("useUnicode")
	if !ok || useUnicodeAny == nil {
		return ConvertedImageArray{}, errors.New("useUnicode is nil")
	}
	useUnicode := useUnicodeAny.(bool)

	reverseCharsAny, ok := Shared().Get("useUnicode")
	if !ok || reverseCharsAny == nil {
		return ConvertedImageArray{}, errors.New("useUnicode is nil")
	}
	reverseChars := reverseCharsAny.(bool)

	var outputChars []rune
	for i := 0; i < len(splitImages); i++ {
		m := getMedianColor(splitImages[i])
		outputChars = append(outputChars, getCharForRGBValue(m, useUnicode, reverseChars))
	}
	_ = Logger().Info(fmt.Sprintf("Finished image conversion"))

	return ConvertedImageArray{
		Cols:       cols,
		Rows:       rows,
		Characters: outputChars,
	}, nil

}

func gridFromTextSize(img image.Image, textSize int, fontAspect float64) (cols, rows int) {
	b := img.Bounds()
	imgW, imgH := b.Dx(), b.Dy()

	charW := textSize
	charH := int(float64(textSize) * fontAspect)
	if charW <= 0 {
		charW = 8
	}
	if charH <= 0 {
		charH = 16
	}

	cols = (imgW + charW - 1) / charW
	rows = (imgH + charH - 1) / charH

	if cols < 1 {
		cols = 1
	}
	if rows < 1 {
		rows = 1
	}

	return cols, rows
}

func splitImage(textSize int, fontAspect float64, inputImg image.Image, isDebugEnv bool) ([]image.Image, error) {

	imgBounds := inputImg.Bounds()
	imgWidth, imgHeight := imgBounds.Dx(), imgBounds.Dy()
	_ = Logger().Info(fmt.Sprintf("imgWidth: %d | imgHeight: %d", imgWidth, imgHeight))

	charWidth := textSize
	charHeight := int(float64(textSize) * fontAspect)

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

	//save images to debug folder if flag true
	var tiles []image.Image
	for i, r := range rects {
		var tile image.Image

		if si, ok := inputImg.(interface {
			SubImage(r image.Rectangle) image.Image
		}); ok {
			tile = si.SubImage(r)
		} else {
			// Fallback: copy crop if subimage fails
			dst := image.NewRGBA(image.Rect(0, 0, r.Dx(), r.Dy()))
			draw.Draw(dst, dst.Bounds(), inputImg, r.Min, draw.Src)
			tile = dst
		}
		tiles = append(tiles, tile)

		//If Debug true - save images
		if isDebugEnv {
			err := saveImageToDebugDir(tile, "image_"+strconv.Itoa(i), "cropped_img")
			if err != nil {
				return nil, err
			}
		}
	}

	return tiles, nil
}

func saveImageToDebugDir(img image.Image, outputName string, subFolderName string) error {
	if filepath.Ext(outputName) == "" {
		outputName += ".png"
	}
	if err := os.MkdirAll("debugFolder/"+subFolderName, 0o755); err != nil {
		return err
	}

	outPath := filepath.Join("debugFolder/"+subFolderName, outputName)
	out, err := os.Create(outPath)
	if err != nil {
		return err
	}
	defer func() { _ = out.Close() }()

	return png.Encode(out, img)
}

func getMedianColor(img image.Image) color.NRGBA {
	b := img.Bounds()
	w, h := b.Dx(), b.Dy()
	if w <= 0 || h <= 0 {
		return color.NRGBA{}
	}

	rs := make([]uint8, 0, w*h)
	gs := make([]uint8, 0, w*h)
	bs := make([]uint8, 0, w*h)
	as := make([]uint8, 0, w*h)

	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			c := color.NRGBAModel.Convert(img.At(x, y)).(color.NRGBA)
			if c.A < 10 {
				continue
			}
			rs = append(rs, c.R)
			gs = append(gs, c.G)
			bs = append(bs, c.B)
			as = append(as, c.A)
		}
	}

	return color.NRGBA{
		R: medianUint8(rs),
		G: medianUint8(gs),
		B: medianUint8(bs),
		A: medianUint8(as),
	}
}

func medianUint8(v []uint8) uint8 {
	if len(v) == 0 {
		return 0
	}
	sort.Slice(v, func(i, j int) bool {
		return v[i] < v[j]
	})

	m := len(v) / 2
	if len(v)%2 == 1 {
		return v[m]
	}

	return uint8((uint16(v[m-1]) + uint16(v[m])) / 2)
}

func getCharForRGBValue(color color.NRGBA, useUnicode bool, reverseChars bool) rune {

	var asciiRamp []rune
	var unicodeRamp []rune

	if reverseChars {
		asciiRamp = []rune(asciiRampDarkToBrightStr)
		unicodeRamp = []rune(unicodeRampDarkToBrightStr)
	} else {
		asciiRamp = []rune(asciiRampBrightToDarkStr)
		unicodeRamp = []rune(unicodeRampBrightToDarkStr)
	}

	brightness := (0.2126 * float64(color.R)) + (0.7152 * float64(color.G)) + (0.0722 * float64(color.B))

	var brightnessLevel float64
	var rampLen int
	var returnChar rune

	if useUnicode {
		rampLen = len(unicodeRamp)
		brightnessLevel = 255.0 / float64(rampLen)
	} else {
		rampLen = len(asciiRamp)
		brightnessLevel = 255.0 / float64(rampLen)
	}
	_ = Logger().Info(fmt.Sprintf("Ramp length: %d | Brightness Step: %.2f", rampLen, brightnessLevel))

	charIndex := int(brightness / brightnessLevel)
	if charIndex < 0 {
		charIndex = 0
	}
	if charIndex >= rampLen {
		charIndex = rampLen - 1
	}
	if useUnicode {
		returnChar = unicodeRamp[charIndex]
	} else {
		returnChar = asciiRamp[charIndex]
	}
	_ = Logger().Info(
		fmt.Sprintf(
			"brightness: %.2f | character: %s | median RGBA: %d %d %d %d | character index: %d",
			brightness, string(returnChar), color.R, color.G, color.B, color.A, charIndex,
		),
	)

	return returnChar
}
