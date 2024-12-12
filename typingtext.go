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
	"strings"
	"syscall/js"

	"github.com/fogleman/gg"
	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font"
)

func buildTypingText(this js.Value, args []js.Value) interface{} {
	background := args[0].Get("background")
	color := args[0].Get("color")
	delay := args[0].Get("delay")
	height := args[0].Get("height")
	text := args[0].Get("text")
	width := args[0].Get("width")
	padding := args[0].Get("padding")

	if background.IsUndefined() {
		background = js.ValueOf("#000000")
	}

	if color.IsUndefined() {
		color = js.ValueOf("#ffffff")
	}

	if delay.IsUndefined() {
		delay = js.ValueOf(100)
	}

	if height.IsUndefined() {
		height = js.ValueOf(200)
	}

	if text.IsUndefined() {
		text = js.ValueOf("BLACK FRIDAY")
	}

	if width.IsUndefined() {
		width = js.ValueOf(800)
	}

	if padding.IsUndefined() {
		padding = js.ValueOf(40)
	}

	typer := NewTypingText(TypingTextOptions{
		Background: background.String(),
		Color:      color.String(),
		Delay:      delay.Float(),
		Height:     height.Int(),
		Text:       text.String(),
		Width:      width.Int(),
		Padding:    padding.Int(),
	})

	b := typer.Create()

	return base64.StdEncoding.EncodeToString(b)
}

type TypingText struct {
	bg      color.Color
	color   color.Color
	delay   float64
	height  int
	text    string
	width   int
	padding int
}

type TypingTextOptions struct {
	Background string
	Color      string
	Delay      float64
	Height     int
	Text       string
	Width      int
	Padding    int
}

func NewTypingText(opts TypingTextOptions) *TypingText {
	return &TypingText{
		bg:      parseHexString(opts.Background),
		color:   parseHexString(opts.Color),
		delay:   opts.Delay,
		height:  opts.Height,
		text:    strings.TrimSpace(opts.Text),
		width:   opts.Width,
		padding: opts.Padding,
	}
}

func (t *TypingText) Create() []byte {
	var images []*image.Paletted
	var delays []int

	// Calculate optimal font size
	fontSize := t.calculateFontSize()
	fontFace, err := t.loadFont(fontSize)
	if err != nil {
		panic(err)
	}

	// Create frames for each letter + blinking cursor
	for i := 0; i <= len(t.text); i++ {
		// Frame with cursor
		// frame := t.createFrame(fontFace, i, true)
		// images = append(images, frame)
		// delays = append(delays, int(t.delay/10))

		// Frame without cursor (blink)
		frame := t.createFrame(fontFace, i, false)
		images = append(images, frame)
		delays = append(delays, int(t.delay/10))
	}

	// Add some extra frames at the end with the cursor blinking
	for i := 0; i < 6; i++ {
		frame := t.createFrame(fontFace, len(t.text), i%2 == 0)
		images = append(images, frame)
		delays = append(delays, int(t.delay/10))
	}

	b := new(bytes.Buffer)
	gif.EncodeAll(b, &gif.GIF{
		Image: images,
		Delay: delays,
	})
	return b.Bytes()
}

func (t *TypingText) calculateFontSize() float64 {
	fontSize := float64(t.height) * 0.5
	dc := gg.NewContext(t.width, t.height)

	maxWidth := float64(t.width - 2*t.padding)
	maxHeight := float64(t.height - 2*t.padding)

	for {
		font, err := t.loadFont(fontSize)
		if err != nil {
			panic(err)
		}
		dc.SetFontFace(font)

		textWidth, textHeight := dc.MeasureString(t.text + "|")

		if textWidth <= maxWidth && textHeight <= maxHeight {
			return fontSize
		}

		fontSize *= 0.9
		if fontSize < 12 {
			return 12
		}
	}
}

func (t *TypingText) createFrame(fontFace font.Face, textLength int, showCursor bool) *image.Paletted {
	dc := gg.NewContext(t.width, t.height)

	// Set background
	dc.SetColor(t.bg)
	dc.Clear()

	dc.SetFontFace(fontFace)

	// Get the visible portion of text
	visibleText := t.text[:textLength]

	// Calculate vertical center
	_, textHeight := dc.MeasureString("M")
	y := (float64(t.height) + textHeight) / 2

	// Draw visible text with padding
	dc.SetColor(t.color)
	dc.DrawString(visibleText, float64(t.padding), y)

	// Draw cursor if needed
	if showCursor {
		cursorX := float64(t.padding)
		if textLength > 0 {
			width, _ := dc.MeasureString(visibleText)
			cursorX += width
		}
		dc.DrawString("|", cursorX, y)
	}

	// Convert to paletted image
	bounds := dc.Image().Bounds()
	palette := t.generatePalette()
	palettedImage := image.NewPaletted(bounds, palette)
	draw.Draw(palettedImage, palettedImage.Rect, dc.Image(), bounds.Min, draw.Src)

	return palettedImage
}

func (t *TypingText) generatePalette() color.Palette {
	palette := make(color.Palette, 0, 256)
	palette = append(palette, t.bg)
	palette = append(palette, t.color)

	// Create gradient between background and text color
	r1, g1, b1, _ := t.bg.RGBA()
	r2, g2, b2, _ := t.color.RGBA()

	for i := 0; i < 254; i++ {
		f := float64(i) / 253.0
		r := uint8((1-f)*float64(r1>>8) + f*float64(r2>>8))
		g := uint8((1-f)*float64(g1>>8) + f*float64(g2>>8))
		b := uint8((1-f)*float64(b1>>8) + f*float64(b2>>8))
		palette = append(palette, color.RGBA{r, g, b, 255})
	}

	return palette
}

func (t *TypingText) loadFont(size float64) (font.Face, error) {
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
