package matrix

import (
	"image"
	"image/color"
	imagedraw "image/draw"
	"math"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"golang.org/x/image/draw"
	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/goregular"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"
)

const (
	imageWidth  = 283
	imageHeight = 273
	supersample = 2

	leftMargin = 11 * supersample
	labelWidth = 64 * supersample
	firstColX  = leftMargin + labelWidth
	columnW    = 67 // 33.5 logical px
	headerH    = 66 * supersample
	rowH       = 33 * supersample
)

var (
	backgroundColor = color.RGBA{R: 255, G: 255, B: 255, A: 255}
	gridColor       = color.RGBA{R: 222, G: 224, B: 223, A: 255}
	textColor       = color.RGBA{R: 23, G: 27, B: 25, A: 255}
	boxGreen        = color.RGBA{R: 53, G: 91, B: 79, A: 255}
)

// NewBooleanMatrixTable renders a fixed-size, cross-platform visual table.
// True cells are filled dark green with a white check; false cells are empty
// outlined squares. Column headings are rotated 90 degrees counterclockwise.
func NewBooleanMatrixTable(matrix [][]bool) fyne.CanvasObject {
	obj := canvas.NewImageFromImage(renderMatrix(matrix))
	obj.FillMode = canvas.ImageFillOriginal
	return obj
}

func renderMatrix(matrix [][]bool) image.Image {
	fontData, err := opentype.Parse(goregular.TTF)
	if err != nil {
		panic("parse embedded Go font: " + err.Error())
	}

	rowFace, err := opentype.NewFace(fontData, &opentype.FaceOptions{
		Size: 17 * supersample, DPI: 72, Hinting: font.HintingFull,
	})
	if err != nil {
		panic("create row-label font: " + err.Error())
	}
	defer rowFace.Close()

	headerFace, err := opentype.NewFace(fontData, &opentype.FaceOptions{
		Size: 19 * supersample, DPI: 72, Hinting: font.HintingFull,
	})
	if err != nil {
		panic("create header font: " + err.Error())
	}
	defer headerFace.Close()
	var matrixCols uint = uint(len(matrix))
	var matrixRows uint = uint(len(matrix[0]))

	big := image.NewRGBA(image.Rect(0, 0, imageWidth*supersample, imageHeight*supersample))
	imagedraw.Draw(big, big.Bounds(), image.NewUniform(backgroundColor), image.Point{}, imagedraw.Src)

	// Draw the light, thin grid first so the text and indicators sit cleanly on top.
	tableRight := int(firstColX + matrixCols*columnW)
	tableBottom := int(headerH + matrixRows*rowH)
	for col := uint(0); col <= matrixCols; col++ {
		x := firstColX + int(col)*columnW
		fillRect(big, x-1, 0, 2, tableBottom, gridColor)
	}
	for row := uint(0); row <= matrixRows; row++ {
		y := headerH + int(row)*rowH
		fillRect(big, leftMargin, y-1, tableRight-leftMargin, 2, gridColor)
	}

	// Row labels are centered in the left-hand label column.
	for row := range matrixRows {
		label := "INP" + string(rune('1'+row))
		centerX := leftMargin + labelWidth/2
		centerY := headerH + int(row)*rowH + rowH/2
		drawCenteredText(big, label, rowFace, centerX, centerY, textColor)
	}

	// Column labels are rasterized, rotated counterclockwise, then centered.
	for col := range matrixCols {
		label := "VER" + string(rune('1'+col))
		rotated := rotatedText(label, headerFace, textColor)
		centerX := firstColX + int(col)*columnW + columnW/2
		centerY := headerH / 2
		x := centerX - rotated.Bounds().Dx()/2
		y := centerY - rotated.Bounds().Dy()/2
		imagedraw.Draw(big, image.Rect(x, y, x+rotated.Bounds().Dx(), y+rotated.Bounds().Dy()), rotated, image.Point{}, imagedraw.Over)
	}

	// Draw the checkbox squares after the grid and labels.
	boxSize := 23 * supersample
	for row := range matrixRows {
		for col := range matrixCols {
			centerX := firstColX + int(col)*columnW + columnW/2
			centerY := headerH + int(row)*rowH + rowH/2
			x := centerX - boxSize/2
			y := centerY - boxSize/2

			fill := backgroundColor
			if matrix[row][col] {
				fill = boxGreen
			}
			drawRoundedRect(big, x, y, boxSize, boxSize, 4*supersample, 3*supersample, fill, boxGreen)

			if matrix[row][col] {
				// Two thick segments form the white check mark.
				drawThickLine(big, x+9*supersample, y+12*supersample,
					x+13*supersample, y+16*supersample, 2.2*supersample, backgroundColor)
				drawThickLine(big, x+12*supersample, y+16*supersample,
					x+19*supersample, y+7*supersample, 2.2*supersample, backgroundColor)
			}
		}
	}

	// Downsample the supersampled drawing for smooth, crisp edges and text.
	final := image.NewRGBA(image.Rect(0, 0, imageWidth, imageHeight))
	draw.CatmullRom.Scale(final, final.Bounds(), big, big.Bounds(), imagedraw.Src, nil)
	return final
}

func drawCenteredText(dst *image.RGBA, text string, face font.Face, centerX, centerY int, col color.Color) {
	width := font.MeasureString(face, text).Ceil()
	metrics := face.Metrics()
	baselineOffset := (metrics.Ascent.Ceil() - metrics.Descent.Ceil()) / 2
	d := font.Drawer{
		Dst:  dst,
		Src:  image.NewUniform(col),
		Face: face,
		Dot:  fixed.P(centerX-width/2, centerY+baselineOffset),
	}
	d.DrawString(text)
}

func rotatedText(text string, face font.Face, col color.Color) *image.RGBA {
	width := font.MeasureString(face, text).Ceil() + 2
	height := face.Metrics().Height.Ceil() + 2
	source := image.NewRGBA(image.Rect(0, 0, width, height))
	d := font.Drawer{
		Dst:  source,
		Src:  image.NewUniform(col),
		Face: face,
		Dot:  fixed.P(1, face.Metrics().Ascent.Ceil()+1),
	}
	d.DrawString(text)

	// Counterclockwise transform: (x, y) -> (y, width-1-x).
	rotated := image.NewRGBA(image.Rect(0, 0, height, width))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			rotated.Set(y, width-1-x, source.At(x, y))
		}
	}
	return rotated
}

func fillRect(dst *image.RGBA, x, y, width, height int, col color.Color) {
	if width <= 0 || height <= 0 {
		return
	}
	r := image.Rect(x, y, x+width, y+height).Intersect(dst.Bounds())
	imagedraw.Draw(dst, r, image.NewUniform(col), image.Point{}, imagedraw.Src)
}

func drawRoundedRect(dst *image.RGBA, x, y, width, height, radius, stroke int, fill, outline color.RGBA) {
	outer := image.Rect(x, y, x+width, y+height).Intersect(dst.Bounds())
	innerX, innerY := x+stroke, y+stroke
	innerW, innerH := width-2*stroke, height-2*stroke
	innerR := max(0, radius-stroke)

	for py := outer.Min.Y; py < outer.Max.Y; py++ {
		for px := outer.Min.X; px < outer.Max.X; px++ {
			cx, cy := float64(px)+0.5, float64(py)+0.5
			if !insideRoundedRect(cx, cy, float64(x), float64(y), float64(width), float64(height), float64(radius)) {
				continue
			}
			col := outline
			if innerW > 0 && innerH > 0 && insideRoundedRect(cx, cy, float64(innerX), float64(innerY), float64(innerW), float64(innerH), float64(innerR)) {
				col = fill
			}
			dst.SetRGBA(px, py, col)
		}
	}
}

func insideRoundedRect(px, py, x, y, width, height, radius float64) bool {
	if px < x || py < y || px >= x+width || py >= y+height {
		return false
	}
	if radius <= 0 {
		return true
	}
	cx := math.Max(x+radius, math.Min(px, x+width-radius))
	cy := math.Max(y+radius, math.Min(py, y+height-radius))
	dx, dy := px-cx, py-cy
	return dx*dx+dy*dy <= radius*radius
}

func drawThickLine(dst *image.RGBA, x1, y1, x2, y2 int, width float64, col color.RGBA) {
	radius := width / 2
	minX := max(0, int(math.Floor(float64(min(x1, x2))-radius-1)))
	maxX := min(dst.Bounds().Dx()-1, int(math.Ceil(float64(max(x1, x2))+radius+1)))
	minY := max(0, int(math.Floor(float64(min(y1, y2))-radius-1)))
	maxY := min(dst.Bounds().Dy()-1, int(math.Ceil(float64(max(y1, y2))+radius+1)))

	dx, dy := float64(x2-x1), float64(y2-y1)
	length2 := dx*dx + dy*dy
	for py := minY; py <= maxY; py++ {
		for px := minX; px <= maxX; px++ {
			pX, pY := float64(px)+0.5, float64(py)+0.5
			t := 0.0
			if length2 > 0 {
				t = ((pX-float64(x1))*dx + (pY-float64(y1))*dy) / length2
				t = math.Max(0, math.Min(1, t))
			}
			nearX, nearY := float64(x1)+t*dx, float64(y1)+t*dy
			d2 := (pX-nearX)*(pX-nearX) + (pY-nearY)*(pY-nearY)
			if d2 <= radius*radius {
				dst.SetRGBA(px, py, col)
			}
		}
	}
}
