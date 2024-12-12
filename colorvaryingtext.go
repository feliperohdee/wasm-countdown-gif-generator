//go:build js && wasm

package main

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/gif"
	"math"
	"strings"
	"syscall/js"
	"unicode"

	"github.com/fogleman/gg"
	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font"
)

func buildColorVaryingText(this js.Value, args []js.Value) interface{} {
	delay := args[0].Get("delay")
	frames := args[0].Get("frames")
	height := args[0].Get("height")
	text := args[0].Get("text")
	width := args[0].Get("width")
	colorScheme := args[0].Get("colorScheme")
	padding := args[0].Get("padding")

	if delay.IsUndefined() {
		delay = js.ValueOf(100)
	}

	if frames.IsUndefined() {
		frames = js.ValueOf(30)
	}

	if height.IsUndefined() {
		height = js.ValueOf(400)
	}

	if text.IsUndefined() {
		text = js.ValueOf("SALE")
	}

	if width.IsUndefined() {
		width = js.ValueOf(600)
	}

	if colorScheme.IsUndefined() {
		colorScheme = js.ValueOf("complementary")
	}

	if padding.IsUndefined() {
		padding = js.ValueOf(40)
	}

	varying := NewColorVaryingText(ColorVaryingTextOptions{
		Delay:       delay.Float(),
		Frames:      frames.Int(),
		Height:      height.Int(),
		Text:        text.String(),
		Width:       width.Int(),
		ColorScheme: colorScheme.String(),
		Padding:     padding.Int(),
	})

	b := varying.Create()

	return base64.StdEncoding.EncodeToString(b)
}

type ColorVaryingText struct {
	delay       float64
	frames      int
	height      int
	text        string
	width       int
	colorScheme string
	padding     int
}

type ColorVaryingTextOptions struct {
	Delay       float64
	Frames      int
	Height      int
	Text        string
	Width       int
	ColorScheme string
	Padding     int
}

type ColorPair struct {
	background color.Color
	text       color.Color
}

type TextLayout struct {
	lines    []string
	fontSize float64
}

func NewColorVaryingText(opts ColorVaryingTextOptions) *ColorVaryingText {
	if opts.Frames < 1 {
		opts.Frames = 1
	} else if opts.Frames > 60 {
		opts.Frames = 60
	}

	return &ColorVaryingText{
		delay:       opts.Delay,
		frames:      opts.Frames,
		height:      opts.Height,
		text:        strings.TrimSpace(opts.Text),
		width:       opts.Width,
		colorScheme: opts.ColorScheme,
		padding:     opts.Padding,
	}
}

func (cv *ColorVaryingText) Create() []byte {
	var images []*image.Paletted
	var delays []int

	layout := cv.calculateTextLayout()

	fontFace, err := cv.loadFont(layout.fontSize)
	if err != nil {
		panic(err)
	}

	for i := 0; i < cv.frames; i++ {
		frame := cv.createFrame(fontFace, layout, i)
		images = append(images, frame)
		delays = append(delays, int(cv.delay/10))
	}

	b := new(bytes.Buffer)
	gif.EncodeAll(b, &gif.GIF{
		Image: images,
		Delay: delays,
	})
	return b.Bytes()
}

func (cv *ColorVaryingText) calculateTextLayout() TextLayout {
	fontSize := float64(cv.height) * 0.5
	dc := gg.NewContext(cv.width, cv.height)

	for {
		font, err := cv.loadFont(fontSize)
		if err != nil {
			panic(err)
		}
		dc.SetFontFace(font)

		availWidth := float64(cv.width - 2*cv.padding)
		availHeight := float64(cv.height - 2*cv.padding)

		words := cv.splitIntoWords(cv.text)
		lines := cv.arrangeWords(dc, words, availWidth)

		lineHeight := fontSize * 1.2
		totalHeight := lineHeight * float64(len(lines))

		if totalHeight <= availHeight {
			return TextLayout{
				lines:    lines,
				fontSize: fontSize,
			}
		}

		fontSize *= 0.9
		if fontSize < 12 {
			fontSize = 12
			break
		}
	}

	font, _ := cv.loadFont(fontSize)
	dc.SetFontFace(font)
	words := cv.splitIntoWords(cv.text)
	lines := cv.arrangeWords(dc, words, float64(cv.width-2*cv.padding))

	return TextLayout{
		lines:    lines,
		fontSize: fontSize,
	}
}

func (cv *ColorVaryingText) splitIntoWords(text string) []string {
	var words []string
	var currentWord strings.Builder

	for _, r := range text {
		if unicode.IsSpace(r) {
			if currentWord.Len() > 0 {
				words = append(words, currentWord.String())
				currentWord.Reset()
			}
		} else {
			currentWord.WriteRune(r)
		}
	}

	if currentWord.Len() > 0 {
		words = append(words, currentWord.String())
	}

	return words
}

func (cv *ColorVaryingText) arrangeWords(dc *gg.Context, words []string, maxWidth float64) []string {
	var lines []string
	var currentLine strings.Builder

	for _, word := range words {
		testLine := currentLine.String()
		if testLine != "" {
			testLine += " "
		}
		testLine += word

		width, _ := dc.MeasureString(testLine)

		if width <= maxWidth {
			if currentLine.Len() > 0 {
				currentLine.WriteRune(' ')
			}
			currentLine.WriteString(word)
		} else {
			if currentLine.Len() > 0 {
				lines = append(lines, currentLine.String())
				currentLine.Reset()
			}
			currentLine.WriteString(word)
		}
	}

	if currentLine.Len() > 0 {
		lines = append(lines, currentLine.String())
	}

	return lines
}

func (cv *ColorVaryingText) createFrame(fontFace font.Face, layout TextLayout, frameNum int) *image.Paletted {
	dc := gg.NewContext(cv.width, cv.height)

	colors := cv.getColorPair(frameNum)

	dc.SetColor(colors.background)
	dc.Clear()

	dc.SetFontFace(fontFace)
	lineHeight := layout.fontSize * 1.2

	totalTextHeight := lineHeight * float64(len(layout.lines))
	startY := float64(cv.padding) + (float64(cv.height-2*cv.padding)-totalTextHeight)/2 + layout.fontSize

	dc.SetColor(colors.text)
	for i, line := range layout.lines {
		textWidth, _ := dc.MeasureString(line)
		x := float64(cv.padding) + (float64(cv.width-2*cv.padding)-textWidth)/2
		y := startY + float64(i)*lineHeight

		dc.DrawString(line, x, y)
	}

	bounds := dc.Image().Bounds()
	palette := cv.generatePalette(colors)
	palettedImage := image.NewPaletted(bounds, palette)
	draw.Draw(palettedImage, palettedImage.Rect, dc.Image(), bounds.Min, draw.Src)

	return palettedImage
}

func (cv *ColorVaryingText) getColorPair(frameNum int) ColorPair {
	hue := float64(frameNum) * (360.0 / float64(cv.frames))

	switch cv.colorScheme {
	case "complementary":
		return ColorPair{
			background: cv.hslToColor(hue, 1.0, 0.3),
			text:       cv.hslToColor(math.Mod(hue+180, 360), 1.0, 0.8),
		}
	case "monochromatic":
		return ColorPair{
			background: cv.hslToColor(hue, 0.8, 0.2),
			text:       cv.hslToColor(hue, 0.8, 0.8),
		}
	case "triadic":
		return ColorPair{
			background: cv.hslToColor(hue, 1.0, 0.3),
			text:       cv.hslToColor(math.Mod(hue+120, 360), 1.0, 0.8),
		}
	default: // analogous
		return ColorPair{
			background: cv.hslToColor(hue, 1.0, 0.3),
			text:       cv.hslToColor(math.Mod(hue+30, 360), 1.0, 0.8),
		}
	}
}

func (cv *ColorVaryingText) hslToColor(h, s, l float64) color.Color {
	var r, g, b float64

	if s == 0 {
		r, g, b = l, l, l
	} else {
		var q float64
		if l < 0.5 {
			q = l * (1 + s)
		} else {
			q = l + s - l*s
		}
		p := 2*l - q

		r = cv.hueToRGB(p, q, h/360+1.0/3.0)
		g = cv.hueToRGB(p, q, h/360)
		b = cv.hueToRGB(p, q, h/360-1.0/3.0)
	}

	return color.RGBA{
		R: uint8(r * 255),
		G: uint8(g * 255),
		B: uint8(b * 255),
		A: 255,
	}
}

func (cv *ColorVaryingText) hueToRGB(p, q, t float64) float64 {
	if t < 0 {
		t += 1
	}
	if t > 1 {
		t -= 1
	}
	if t < 1.0/6.0 {
		return p + (q-p)*6*t
	}
	if t < 1.0/2.0 {
		return q
	}
	if t < 2.0/3.0 {
		return p + (q-p)*(2.0/3.0-t)*6
	}
	return p
}

func (cv *ColorVaryingText) generatePalette(colors ColorPair) color.Palette {
	palette := make(color.Palette, 0, 256)
	palette = append(palette, colors.background)
	palette = append(palette, colors.text)

	r1, g1, b1, _ := colors.background.RGBA()
	r2, g2, b2, _ := colors.text.RGBA()

	for i := 0; i < 254; i++ {
		t := float64(i) / 254.0
		palette = append(palette, color.RGBA{
			R: uint8((1-t)*float64(r1>>8) + t*float64(r2>>8)),
			G: uint8((1-t)*float64(g1>>8) + t*float64(g2>>8)),
			B: uint8((1-t)*float64(b1>>8) + t*float64(b2>>8)),
			A: 255,
		})
	}

	return palette
}

func (cv *ColorVaryingText) loadFont(size float64) (font.Face, error) {
	fontBytes, err := fontFS.ReadFile("fonts/impact.ttf")
	if err != nil {
		return nil, fmt.Errorf("failed to read embedded font file: %v", err)
	}

	font, err := truetype.Parse(fontBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse font: %v", err)
	}

	return truetype.NewFace(font, &truetype.Options{
		Size: size,
		DPI:  144,
	}), nil
}
