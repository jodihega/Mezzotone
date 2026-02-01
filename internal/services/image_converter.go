package services

import (
	"errors"
	"fmt"
	"image"
	"image/color"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"os"

	_ "golang.org/x/image/bmp"
	_ "golang.org/x/image/tiff"
	_ "golang.org/x/image/webp"
)

const asciiRampDarkToBrightStr = "$@B%8&WM#*oahkbdpqwmZO0QLCJUYXzcvunxrjft/\\|()1{}[]?-_+~<>i!lI;:,^`. "
const unicodeRampDarkToBrightStr = "█▉▊▋▌▍▎▏▓▒░■□@&%$#*+=-~:;!,\".^`' "

const asciiRampBrightToDarkStr = " .`^,:;Il!i><~+_-?][}{1)(|\\/tfjrxnuvczXYUJCLQ0OZmwqpdbkhao*#MW&8%B@$"
const unicodeRampBrightToDarkStr = " '`^.\",!;:~-=+*#$%&@□■░▒▓▏▎▍▌▋▊▉█"

/*
TODO: If unicode is true, you can offer multiple ramps (style presets) for the user to choose from.
Examples:
	█▇▆▅▄▃▂▁ ▁▂▃▄▅▆▇█
	█▓▒░ ░▒▓█
	⣿⣷⣧⣇⣆⣄⣀ ⣀⣄⣆⣇⣧⣷⣿
	●∙•·  ·•∙●
*/

func ConvertImageToString(filePath string) ([][]rune, error) {
	var outputChars [][]rune

	f, err := os.Open(filePath)
	if err != nil {
		return outputChars, err
	}
	defer func() { _ = f.Close() }()

	_ = Logger().Info(fmt.Sprintf("Successfully Loaded: %s", filePath))

	inputImg, format, err := image.Decode(f)
	if err != nil {
		return outputChars, err
	}
	_ = Logger().Info(fmt.Sprintf("format: %s", format))

	// textSize: roughly controls how many pixels map to one character horizontally.
	// If missing/invalid, fall back to a reasonable default.
	textSizeAny, ok := Shared().Get("textSize")
	if !ok || textSizeAny == nil {
		return outputChars, errors.New("textSize is nil")
	}
	textSize, ok := textSizeAny.(int)
	if !ok || textSize <= 0 {
		textSize = 8
	}

	// fontAspect: terminal characters are typically taller than they are wide,
	// so vertical cell size is textSize * fontAspect.
	fontAspectAny, ok := Shared().Get("fontAspect")
	if !ok || fontAspectAny == nil {
		return outputChars, errors.New("fontAspect is nil")
	}
	fontAspect, ok := fontAspectAny.(float64)
	if !ok || fontAspect <= 0 {
		fontAspect = 2
	}

	// highContrast: optional contrast curve applied after cell luminance averaging.
	// Useful for "washed out" images in ASCII output.
	highContrastAny, ok := Shared().Get("highContrast")
	if !ok || highContrastAny == nil {
		return outputChars, errors.New("highContrast is nil")
	}
	highContrast := highContrastAny.(bool)

	// Compute grid resolution (cols x rows) based on image size + character cell size.
	cols, rows := getColsAndRows(inputImg, textSize, fontAspect)

	outputChars = make([][]rune, rows)
	for r := 0; r < rows; r++ {
		outputChars[r] = make([]rune, cols)
	}

	// Build a luminance grid (rows x cols) where each cell is 0..1.
	// Each cell luminance is computed by averaging pixels in the corresponding image region.
	lumaGrid, err := buildLuminanceGrid(inputImg, cols, rows, highContrast)
	if err != nil {
		return outputChars, err
	}
	_ = Logger().Info(fmt.Sprintf("Successfully Build LumaGrid for %s", filePath))

	// directionalRender: optional Edge Awareness.
	// Derive edge magnitude/orientation from lumaGrid and choose glyphs accordingly.
	directionalRenderAny, ok := Shared().Get("directionalRender")
	if !ok || directionalRenderAny == nil {
		return outputChars, errors.New("directionalRender is nil")
	}
	directionalRender := directionalRenderAny.(bool)
	if directionalRender {
		// TODO: apply Sobel filter on lumaGrid (operate on the *grid*, not original pixels)
		// Steps (high level):
		// 1) compute Gx, Gy per cell (skip borders)
		// 2) magnitude = sqrt(Gx*Gx + Gy*Gy) normalized
		// 3) orientation = atan2(Gy, Gx)
		// 4) map orientation + magnitude to a directional glyph set
	}

	_ = Logger().Info(fmt.Sprintf("Beginning image conversion"))

	// useUnicode: pick between ASCII ramps and Unicode ramps.
	useUnicodeAny, ok := Shared().Get("useUnicode")
	if !ok || useUnicodeAny == nil {
		return outputChars, errors.New("useUnicode is nil")
	}
	useUnicode := useUnicodeAny.(bool)

	// reverseChars: invert ramp direction (useful for dark terminals / preference).
	reverseCharsAny, ok := Shared().Get("reverseChars")
	if !ok || reverseCharsAny == nil {
		return outputChars, errors.New("reverseChars is nil")
	}
	reverseChars := reverseCharsAny.(bool)

	// Convert each luminance cell to a glyph using the chosen ramp.
	// lumaGrid indices are [row][col] matching outputChars.
	for i := 0; i < len(lumaGrid); i++ {
		for j := 0; j < len(lumaGrid[i]); j++ {
			outputChars[i][j] = getCharForLuminanceValue(lumaGrid[i][j], useUnicode, reverseChars)
		}
	}

	_ = Logger().Info(fmt.Sprintf("Finished image conversion"))
	return outputChars, nil
}

//Calculates Columns and Rows for given TextSize and FontAspect
func getColsAndRows(img image.Image, textSize int, fontAspect float64) (cols, rows int) {
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

//Builds a grid of averaged luminance values in [0..1].
func buildLuminanceGrid(inputImg image.Image, cols, rows int, highContrast bool) ([][]float64, error) {

	imgBounds := inputImg.Bounds()
	imgWidth, imgHeight := imgBounds.Dx(), imgBounds.Dy()

	cellWidth := imgWidth / cols
	cellHeight := imgHeight / rows

	// Safety fallback if cols/rows are weird (should be prevented earlier).
	if cellWidth <= 0 {
		cellWidth = 8
	}
	if cellHeight <= 0 {
		cellHeight = 16
	}

	// Allocate luminance grid.
	grid := make([][]float64, rows)
	for gridRow := 0; gridRow < rows; gridRow++ {
		grid[gridRow] = make([]float64, cols)
	}

	for gridRow := 0; gridRow < rows; gridRow++ {
		// Pixel Y-range for this grid row.
		cellRowPixelStartY := gridRow * cellHeight
		cellRowPixelEndY := cellRowPixelStartY + cellHeight
		if cellRowPixelStartY >= imgHeight {
			cellRowPixelStartY = imgHeight
		}
		if cellRowPixelEndY > imgHeight {
			cellRowPixelEndY = imgHeight
		}

		for gridCol := 0; gridCol < cols; gridCol++ {
			// Pixel X-range for this grid column.
			cellColPixelStartX := gridCol * cellWidth
			cellColPixelEndX := cellColPixelStartX + cellWidth
			if cellColPixelStartX >= imgWidth {
				cellColPixelStartX = imgWidth
			}
			if cellColPixelEndX > imgWidth {
				cellColPixelEndX = imgWidth
			}

			// Fallback guard (should not happen if dimensions are sane).
			if cellColPixelEndX <= cellColPixelStartX || cellRowPixelEndY <= cellRowPixelStartY {
				grid[gridRow][gridCol] = 0
				continue
			}

			var lumaSum float64
			var sampleCount float64

			for y := cellRowPixelStartY; y < cellRowPixelEndY; y++ {
				for x := cellColPixelStartX; x < cellColPixelEndX; x++ {
					c := color.NRGBAModel.Convert(
						inputImg.At(imgBounds.Min.X+x, imgBounds.Min.Y+y),
					).(color.NRGBA)

					// Skip mostly transparent pixels to prevent background bleed.
					if c.A < 10 {
						continue
					}

					// Luminance is computed as 0..1.
					pixelLuminance := calculateLuminance(c.R, c.G, c.B)
					lumaSum += pixelLuminance
					sampleCount++
				}
			}

			// Average luminance;
			// if all transparent, treat as black.
			var cellLuma float64
			if sampleCount == 0 {
				cellLuma = 0
			} else {
				cellLuma = lumaSum / sampleCount
			}

			// Optional contrast remap
			if highContrast {
				cellLuma = applyContrast(cellLuma, 1.7)
			}

			grid[gridRow][gridCol] = clamp01(cellLuma)
		}
	}

	return grid, nil
}

/*
Calculates luminance from rgb values and normalizes them from 0..255 into 0..1
	Uses standard relative luminance weights (Rec.709 / sRGB), where green contributes the most to perceived brightness
*/
func calculateLuminance(red uint8, green uint8, blue uint8) float64 {
	luminance := 0.2126*float64(red) + 0.7152*float64(green) + 0.0722*float64(blue)
	return luminance / 255.0
}

// Get the rune correspondent to luminance in selected ramp
func getCharForLuminanceValue(luminance float64, useUnicode bool, reverseChars bool) rune {
	var ramp []rune
	if useUnicode {
		if reverseChars {
			ramp = []rune(unicodeRampBrightToDarkStr)
		} else {
			ramp = []rune(unicodeRampDarkToBrightStr)
		}
	} else {
		if reverseChars {
			ramp = []rune(asciiRampBrightToDarkStr)
		} else {
			ramp = []rune(asciiRampDarkToBrightStr)
		}
	}

	// Map luminance to an index in the ramp:
	index := int(luminance * float64(len(ramp)-1))

	_ = Logger().Info(
		fmt.Sprintf(
			"brightness: %.2f | character: %s | character index: %d",
			luminance, string(ramp[index]), index,
		),
	)

	return ramp[index]
}

// Clamp to [0..1] to keep mapping stable.
func clamp01(x float64) float64 {

	if x < 0 {
		return 0
	}
	if x > 1 {
		return 1
	}
	return x
}

// Applies contrast to lumiance levels with contrast curve at 0.5
func applyContrast(l float64, factor float64) float64 {
	return clamp01((l-0.5)*factor + 0.5)
}
