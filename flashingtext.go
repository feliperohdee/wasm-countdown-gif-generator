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

func buildFlashingText(this js.Value, args []js.Value) interface{} {
	background := args[0].Get("background")
	color := args[0].Get("color")
	delay := args[0].Get("delay")
	frames := args[0].Get("frames")
	height := args[0].Get("height")
	text := args[0].Get("text")
	width := args[0].Get("width")
	words := args[0].Get("words")

	if background.IsUndefined() {
		background = js.ValueOf("#000000")
	}

	if color.IsUndefined() {
		color = js.ValueOf("#ffffff")
	}

	if delay.IsUndefined() {
		delay = js.ValueOf(300)
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

	if words.IsUndefined() {
		words = js.ValueOf(10)
	}

	flasher := NewFlashingText(FlashingTextOptions{
		Background: background.String(),
		Color:      color.String(),
		Delay:      delay.Float(),
		Frames:     frames.Int(),
		Height:     height.Int(),
		Text:       text.String(),
		Width:      width.Int(),
		Words:      words.Int(),
	})

	b := flasher.Create()

	return base64.StdEncoding.EncodeToString(b)
}

type FlashingText struct {
	bg     color.Color
	color  color.Color
	delay  float64
	frames int
	height int
	text   string
	width  int
	words  int
}

type FlashingTextOptions struct {
	Background string
	Color      string
	Delay      float64
	Frames     int
	Height     int
	Text       string
	Width      int
	Words      int
}

type WordPosition struct {
	x    float64
	y    float64
	size float64
	word string
}

func NewFlashingText(opts FlashingTextOptions) *FlashingText {
	if opts.Frames < 1 {
		opts.Frames = 1
	} else if opts.Frames > 60 {
		opts.Frames = 60
	}

	if opts.Words < 1 {
		opts.Words = 1
	} else if opts.Words > 20 {
		opts.Words = 20
	}

	return &FlashingText{
		bg:     parseHexString(opts.Background),
		color:  parseHexString(opts.Color),
		delay:  opts.Delay,
		frames: opts.Frames,
		height: opts.Height,
		text:   opts.Text,
		width:  opts.Width,
		words:  opts.Words,
	}
}

func (f *FlashingText) Create() []byte {
	var images []*image.Paletted
	var delays []int

	// Generate random positions for words
	wordPositions := f.generateWordPositions()

	// Generate frames
	for i := 0; i < f.frames; i++ {
		frame := f.createFrame(wordPositions, i)
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

func (f *FlashingText) generateWordPositions() []WordPosition {
	var positions []WordPosition
	text := f.text
	padding := float64(f.height) * 0.1 // 10% padding

	// Use a fixed seed for predictable positions
	rng := rand.New(rand.NewSource(42))

	for i := 0; i < f.words; i++ {
		size := float64(f.height) * 0.15 // Random size between 15%

		// Random position with padding to avoid edge cutting
		x := padding + rng.Float64()*(float64(f.width)-2*padding-size)
		y := padding + rng.Float64()*(float64(f.height)-2*padding-size)

		// Ensure the word does not go outside the frame
		if x+size > float64(f.width)-padding {
			x = float64(f.width) - padding - size
		}

		if y+size > float64(f.height)-padding {
			y = float64(f.height) - padding - size
		}

		positions = append(positions, WordPosition{
			x:    x,
			y:    y,
			size: size,
			word: text,
		})
	}

	return positions
}

func (f *FlashingText) createFrame(positions []WordPosition, frameNum int) *image.Paletted {
	dc := gg.NewContext(f.width, f.height)

	// Set background
	dc.SetColor(f.bg)
	dc.Clear()

	// Select which word will be visible in this frame
	visibleWord := frameNum % len(positions)

	// Draw each word
	for i, pos := range positions {
		// Load font with the random size for this word
		fontFace, err := f.loadFont(pos.size)
		if err != nil {
			continue
		}
		dc.SetFontFace(fontFace)

		// Only show the selected word for this frame
		if i == visibleWord {
			dc.SetColor(f.color)
			dc.DrawStringAnchored(pos.word, pos.x, pos.y, 0.5, 0.5)
		}
	}

	// Convert to paletted image
	bounds := dc.Image().Bounds()
	palette := f.generatePalette()
	palettedImage := image.NewPaletted(bounds, palette)
	draw.Draw(palettedImage, palettedImage.Rect, dc.Image(), bounds.Min, draw.Src)

	return palettedImage
}

func (f *FlashingText) generatePalette() color.Palette {
	var palette color.Palette

	palette = append(palette, f.bg)

	for i := 1; i < 256; i++ {
		t := float64(i) / 255.0
		blendedColor := f.blendColor(t)
		palette = append(palette, blendedColor)
	}

	return palette
}

func (f *FlashingText) blendColor(t float64) color.Color {
	r1, g1, b1, _ := f.color.RGBA()
	r2, g2, b2, _ := f.bg.RGBA()

	r := uint8((1-t)*float64(r1>>8) + t*float64(r2>>8))
	g := uint8((1-t)*float64(g1>>8) + t*float64(g2>>8))
	b := uint8((1-t)*float64(b1>>8) + t*float64(b2>>8))

	return color.RGBA{R: r, G: g, B: b, A: 255}
}

func (f *FlashingText) loadFont(size float64) (font.Face, error) {
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
