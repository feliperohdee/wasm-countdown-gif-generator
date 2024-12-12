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

	"github.com/fogleman/gg"
	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font"
)

func buildLedBanner(this js.Value, args []js.Value) interface{} {
	background := args[0].Get("background")
	color := args[0].Get("color")
	delay := args[0].Get("delay")
	forward := args[0].Get("forward")
	frames := args[0].Get("frames")
	height := args[0].Get("height")
	spaceSize := args[0].Get("spaceSize")
	text := args[0].Get("text")
	width := args[0].Get("width")

	if background.IsUndefined() {
		background = js.ValueOf("#000000")
	}

	if color.IsUndefined() {
		color = js.ValueOf("#ffffff")
	}

	if delay.IsUndefined() {
		delay = js.ValueOf(50)
	}

	if forward.IsUndefined() {
		forward = js.ValueOf(true)
	}

	if frames.IsUndefined() {
		frames = js.ValueOf(10)
	}

	if height.IsUndefined() {
		height = js.ValueOf(50)
	}

	if spaceSize.IsUndefined() {
		spaceSize = js.ValueOf(4)
	}

	if text.IsUndefined() {
		text = js.ValueOf("Hello World!")
	}

	if width.IsUndefined() {
		width = js.ValueOf(800)
	}

	banner := NewLedBanner(LedBannerOptions{
		Background: background.String(),
		Color:      color.String(),
		Delay:      delay.Float(),
		Forward:    forward.Bool(),
		Frames:     frames.Int(),
		Height:     height.Int(),
		SpaceSize:  spaceSize.Int(),
		Text:       text.String(),
		Width:      width.Int(),
	})

	b := banner.Create()

	return base64.StdEncoding.EncodeToString(b)
}

type LedBanner struct {
	bg        color.Color
	color     color.Color
	delay     float64
	forward   bool
	frames    int
	height    int
	spaceSize int
	text      string
	width     int
}

type LedBannerOptions struct {
	Background string
	Color      string
	Delay      float64
	Forward    bool
	Frames     int
	Height     int
	SpaceSize  int
	Text       string
	Width      int
}

func NewLedBanner(opts LedBannerOptions) *LedBanner {
	if opts.Frames < 1 {
		opts.Frames = 1
	} else if opts.Frames > 30 {
		opts.Frames = 30
	}

	return &LedBanner{
		bg:        parseHexString(opts.Background),
		color:     parseHexString(opts.Color),
		delay:     opts.Delay,
		forward:   opts.Forward,
		frames:    opts.Frames,
		height:    opts.Height,
		spaceSize: opts.SpaceSize,
		text:      opts.Text,
		width:     opts.Width,
	}
}

func (l *LedBanner) Create() []byte {
	var images []*image.Paletted
	var delays []int

	// Create font face
	fontFace, err := l.loadFont(float64(l.height) * 0.8)
	if err != nil {
		panic(err)
	}

	space := strings.Repeat(" ", l.spaceSize)

	// Calculate text width and prepare continuous text
	dc := gg.NewContext(l.width, l.height)
	dc.SetFontFace(fontFace)
	textWidth, _ := dc.MeasureString(l.text + space)

	// Calculate how many copies of the text we need to fill the screen plus one extra
	copies := int(math.Ceil(float64(l.width)/textWidth)) + 2
	continuousText := strings.Repeat(l.text+space, copies)

	// Generate frames
	for i := 0; i < l.frames; i++ {
		frame := l.createFrame(fontFace, i, textWidth, continuousText)
		images = append(images, frame)
		delays = append(delays, int(l.delay/10)) // time in GIF is by 100ths of a second
	}

	// Encode GIF
	b := new(bytes.Buffer)
	gif.EncodeAll(b, &gif.GIF{
		Image: images,
		Delay: delays,
	})
	return b.Bytes()
}

func (l *LedBanner) createFrame(fontFace font.Face, frameNum int, textWidth float64, continuousText string) *image.Paletted {
	dc := gg.NewContext(l.width, l.height)

	// Set background
	dc.SetColor(l.bg)
	dc.Clear()

	// Calculate text position for scrolling effect
	offset := float64(frameNum) * textWidth / float64(l.frames)
	x := -offset
	y := float64(l.height) / 2

	if l.forward {
		x = -textWidth + offset
	}

	// Draw LED effect
	l.drawLedText(dc, fontFace, x, y, continuousText)

	// Convert to paletted image
	bounds := dc.Image().Bounds()
	palette := l.generatePalette()
	palettedImage := image.NewPaletted(bounds, palette)
	draw.Draw(palettedImage, palettedImage.Rect, dc.Image(), bounds.Min, draw.Src)

	return palettedImage
}

func (l *LedBanner) drawLedText(dc *gg.Context, fontFace font.Face, x, y float64, text string) {
	dc.SetFontFace(fontFace)

	// Draw main text
	dc.SetColor(l.color)
	dc.DrawStringAnchored(text, x, y, 0, 0.5)
}

func (l *LedBanner) blendColor(t float64) color.Color {
	r1, g1, b1, _ := l.color.RGBA()
	r2, g2, b2, _ := l.bg.RGBA()

	r := uint8((1-t)*float64(r1>>8) + t*float64(r2>>8))
	g := uint8((1-t)*float64(g1>>8) + t*float64(g2>>8))
	b := uint8((1-t)*float64(b1>>8) + t*float64(b2>>8))

	return color.RGBA{R: r, G: g, B: b, A: 255}
}

func (l *LedBanner) generatePalette() color.Palette {
	var palette color.Palette

	palette = append(palette, l.bg)

	for i := 1; i < 256; i++ {
		t := float64(i) / 255.0
		blendedColor := l.blendColor(t)
		palette = append(palette, blendedColor)
	}

	return palette
}

func (l *LedBanner) loadFont(size float64) (font.Face, error) {
	fontBytes, err := fontFS.ReadFile("fonts/impact.ttf")
	if err != nil {
		return nil, fmt.Errorf("failed to read embedded font file: %v", err)
	}

	f, err := truetype.Parse(fontBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse font: %v", err)
	}

	return truetype.NewFace(f, &truetype.Options{Size: size}), nil
}
