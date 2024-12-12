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
	"math/rand"
	"syscall/js"

	"github.com/fogleman/gg"
	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font"
)

func buildFlashingLetters(this js.Value, args []js.Value) interface{} {
	background := args[0].Get("background")
	color := args[0].Get("color")
	delay := args[0].Get("delay")
	frames := args[0].Get("frames")
	height := args[0].Get("height")
	text := args[0].Get("text")
	width := args[0].Get("width")
	flashProbability := args[0].Get("flashProbability")

	if background.IsUndefined() {
		background = js.ValueOf("#000000")
	}

	if color.IsUndefined() {
		color = js.ValueOf("#ffffff")
	}

	if delay.IsUndefined() {
		delay = js.ValueOf(100)
	}

	if frames.IsUndefined() {
		frames = js.ValueOf(20)
	}

	if height.IsUndefined() {
		height = js.ValueOf(200)
	}

	if text.IsUndefined() {
		text = js.ValueOf("SALE")
	}

	if width.IsUndefined() {
		width = js.ValueOf(400)
	}

	if flashProbability.IsUndefined() {
		flashProbability = js.ValueOf(0.3)
	}

	flasher := NewFlashingLetters(FlashingLettersOptions{
		Background:       background.String(),
		Color:            color.String(),
		Delay:            delay.Float(),
		Frames:           frames.Int(),
		Height:           height.Int(),
		Text:             text.String(),
		Width:            width.Int(),
		FlashProbability: flashProbability.Float(),
	})

	b := flasher.Create()

	return base64.StdEncoding.EncodeToString(b)
}

type FlashingLetters struct {
	bg               color.Color
	color            color.Color
	delay            float64
	frames           int
	height           int
	text             string
	width            int
	flashProbability float64
}

type FlashingLettersOptions struct {
	Background       string
	Color            string
	Delay            float64
	Frames           int
	Height           int
	Text             string
	Width            int
	FlashProbability float64
}

func NewFlashingLetters(opts FlashingLettersOptions) *FlashingLetters {
	if opts.Frames < 1 {
		opts.Frames = 1
	} else if opts.Frames > 60 {
		opts.Frames = 60
	}

	if opts.FlashProbability < 0 {
		opts.FlashProbability = 0
	} else if opts.FlashProbability > 1 {
		opts.FlashProbability = 1
	}

	return &FlashingLetters{
		bg:               parseHexString(opts.Background),
		color:            parseHexString(opts.Color),
		delay:            opts.Delay,
		frames:           opts.Frames,
		height:           opts.Height,
		text:             opts.Text,
		width:            opts.Width,
		flashProbability: opts.FlashProbability,
	}
}

func (f *FlashingLetters) Create() []byte {
	var images []*image.Paletted
	var delays []int

	// Create font face
	fontFace, err := f.loadFont(float64(f.height) * 0.6)
	if err != nil {
		panic(err)
	}

	// Generate frames
	for i := 0; i < f.frames; i++ {
		frame := f.createFrame(fontFace)
		images = append(images, frame)
		delays = append(delays, int(f.delay/10))
	}

	// Encode GIF
	b := new(bytes.Buffer)
	gif.EncodeAll(b, &gif.GIF{
		Image: images,
		Delay: delays,
	})
	return b.Bytes()
}

func (f *FlashingLetters) createFrame(fontFace font.Face) *image.Paletted {
	dc := gg.NewContext(f.width, f.height)

	// Set background
	dc.SetColor(f.bg)
	dc.Clear()

	dc.SetFontFace(fontFace)

	// Calculate text position
	textWidth, textHeight := dc.MeasureString(f.text)
	x := (float64(f.width) - textWidth) / 2
	y := (float64(f.height) + textHeight) / 2

	// Draw each character with random flashing
	for _, char := range f.text {
		charWidth, _ := dc.MeasureString(string(char))

		// Randomly decide if this character should flash
		if rand.Float64() < f.flashProbability {
			dc.SetColor(f.bg) // Make character disappear
		} else {
			dc.SetColor(f.color)
		}

		dc.DrawString(string(char), x, y)
		x += charWidth
	}

	// Convert to paletted image
	bounds := dc.Image().Bounds()
	palette := f.generatePalette()
	palettedImage := image.NewPaletted(bounds, palette)
	draw.Draw(palettedImage, palettedImage.Rect, dc.Image(), bounds.Min, draw.Src)

	return palettedImage
}

func (f *FlashingLetters) generatePalette() color.Palette {
	var palette color.Palette

	palette = append(palette, f.bg)

	for i := 1; i < 256; i++ {
		t := float64(i) / 255.0
		blendedColor := f.blendColor(t)
		palette = append(palette, blendedColor)
	}

	return palette
}

func (f *FlashingLetters) blendColor(t float64) color.Color {
	r1, g1, b1, _ := f.color.RGBA()
	r2, g2, b2, _ := f.bg.RGBA()

	r := uint8((1-t)*float64(r1>>8) + t*float64(r2>>8))
	g := uint8((1-t)*float64(g1>>8) + t*float64(g2>>8))
	b := uint8((1-t)*float64(b1>>8) + t*float64(b2>>8))

	return color.RGBA{R: r, G: g, B: b, A: 255}
}

func (f *FlashingLetters) loadFont(size float64) (font.Face, error) {
	fontBytes, err := fontFS.ReadFile("fonts/impact.ttf")
	if err != nil {
		return nil, fmt.Errorf("failed to read embedded font file: %v", err)
	}

	font, err := truetype.Parse(fontBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse font: %v", err)
	}

	return truetype.NewFace(font, &truetype.Options{Size: size}), nil
}
