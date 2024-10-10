//go:build js && wasm

package main

import (
	"bytes"
	"embed"
	"encoding/base64"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/gif"
	"math"
	"strings"
	"syscall/js"
	"time"

	"github.com/fogleman/gg"
	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font"
)

//go:embed fonts/*.ttf
var fontFS embed.FS

var ALLOWED_FONTS = map[string]bool{
	"impact": true,
}

func main() {
	js.Global().Set("build", js.FuncOf(build))

	select {}
}

func build(this js.Value, args []js.Value) interface{} {
	background := args[0].Get("background")
	color := args[0].Get("color")
	date := args[0].Get("date")
	kind := args[0].Get("kind")

	if background.IsUndefined() {
		background = js.ValueOf("#000000")
	}

	if color.IsUndefined() {
		color = js.ValueOf("#ffffff")
	}

	if date.IsUndefined() {
		var TEN_DAYS = time.Hour * 24 * 10

		date = js.ValueOf(time.Now().Add(TEN_DAYS).Format("2006-01-02T15:04:05.000Z"))
	}

	if kind.IsUndefined() {
		kind = js.ValueOf("rounded")
	}

	countdowm := NewCountdown(CountdownOptions{
		Background: background.String(),
		Color:      color.String(),
		Frames:     10,
		Height:     200,
		Kind:       kind.String(),
		TargetDate: date.String(),
		Width:      700,
	})

	b := countdowm.Create()

	return base64.StdEncoding.EncodeToString(b)
}

func parseHexString(s string) color.Color {
	var r, g, b uint8

	if !strings.HasPrefix(s, "#") {
		s = "#" + s
	}

	if len(s) == 7 {
		fmt.Sscanf(s, "#%02x%02x%02x", &r, &g, &b)
	} else if len(s) == 4 {
		fmt.Sscanf(s, "#%1x%1x%1x", &r, &g, &b)
		r *= 17
		g *= 17
		b *= 17
	} else {
		r, g, b = 255, 255, 255
	}

	return color.RGBA{r, g, b, 255}
}

func parseDateString(s string) (time.Time, error) {
	layout := "2006-01-02T15:04:05.000Z"

	return time.Parse(layout, s)
}

type Countdown struct {
	bg         color.Color
	color      color.Color
	font       string
	frames     int
	kind       string
	targetDate time.Time
	w, h       int
}

type CountdownOptions struct {
	Background string
	Font       string
	Color      string
	Frames     int
	Height     int
	Kind       string
	TargetDate string
	Width      int
}

func NewCountdown(opts CountdownOptions) *Countdown {
	targetDate, err := parseDateString(opts.TargetDate)

	if err != nil {
		panic(err)
	}

	if time.Now().After(targetDate) || time.Now().Equal(targetDate) {
		targetDate = time.Now()
		opts.Frames = 1
	}

	return &Countdown{
		bg:         parseHexString(opts.Background),
		color:      parseHexString(opts.Color),
		frames:     opts.Frames,
		kind:       opts.Kind,
		h:          opts.Height,
		targetDate: targetDate,
		w:          opts.Width,
	}
}

func (c *Countdown) Create() []byte {
	var frame *image.Paletted
	var images []*image.Paletted
	var delays []int

	for i := 0; i < c.frames; i++ {
		now := time.Now().Add(time.Duration(i) * time.Second)
		timeLeft := c.targetDate.Sub(now)

		days := int(timeLeft.Hours() / 24)
		hours := int(timeLeft.Hours()) % 24
		minutes := int(timeLeft.Minutes()) % 60
		seconds := int(timeLeft.Seconds()) % 60

		switch c.kind {
		default:
			frame = c.createFrameBasic(days, hours, minutes, seconds)
		case "rounded", "rounded-ticks", "rounded-dots":
			frame = c.createFrameRounded(days, hours, minutes, seconds)
		}

		images = append(images, frame)
		delays = append(delays, 100)
	}

	b := new(bytes.Buffer)
	gif.EncodeAll(b, &gif.GIF{
		Delay: delays,
		Image: images,
	})
	return b.Bytes()
}

func (c *Countdown) blendColorOverAlpha(alpha uint8) color.Color {
	a := float64(alpha) / 255.0
	r1, g1, b1, _ := c.color.RGBA()
	r2, g2, b2, _ := c.bg.RGBA()

	r := uint8(float64(r1>>8)*a + float64(r2>>8)*(1-a))
	g := uint8(float64(g1>>8)*a + float64(g2>>8)*(1-a))
	b := uint8(float64(b1>>8)*a + float64(b2>>8)*(1-a))
	return color.RGBA{R: r, G: g, B: b, A: 255}
}

func (c *Countdown) blendColor(t float64) color.Color {
	r1, g1, b1, _ := c.color.RGBA()
	r2, g2, b2, _ := c.bg.RGBA()
	r := uint8((1-t)*float64(r1>>8) + t*float64(r2>>8))
	g := uint8((1-t)*float64(g1>>8) + t*float64(g2>>8))
	b := uint8((1-t)*float64(b1>>8) + t*float64(b2>>8))
	return color.RGBA{R: r, G: g, B: b, A: 255}
}

func (c *Countdown) generateColorPalette() color.Palette {
	var palette color.Palette

	palette = append(palette, c.bg)
	for i := 1; i < 256; i++ {
		t := float64(i) / 255.0
		blendedColor := c.blendColor(t)
		palette = append(palette, blendedColor)
	}

	return palette
}

func (c *Countdown) loadFont(size float64) (font.Face, error) {
	if _, ok := ALLOWED_FONTS[c.font]; !ok {
		c.font = "impact"
	}

	fontBytes, err := fontFS.ReadFile("fonts/" + c.font + ".ttf")

	if err != nil {
		return nil, fmt.Errorf("failed to read embedded font file: %v", err)
	}

	f, err := truetype.Parse(fontBytes)

	if err != nil {
		return nil, fmt.Errorf("failed to parse font: %v", err)
	}

	return truetype.NewFace(f, &truetype.Options{Size: size}), nil
}

func (c *Countdown) image(dc *gg.Context) *image.Paletted {
	bounds := dc.Image().Bounds()
	palette := c.generateColorPalette()
	palettedImage := image.NewPaletted(bounds, palette)
	draw.Draw(palettedImage, palettedImage.Rect, dc.Image(), bounds.Min, draw.Src)

	return palettedImage
}

func (c *Countdown) createFrameBasic(days, hours, minutes, seconds int) *image.Paletted {
	dc := gg.NewContext(c.w, c.h)

	dc.SetColor(c.bg)
	dc.Clear()

	if face, err := c.loadFont(60); err == nil {
		dc.SetFontFace(face)
	}

	// Draw countdown text
	dc.SetColor(c.color)
	countdownText := fmt.Sprintf("%dd %dh %dm %ds", days, hours, minutes, seconds)
	dc.DrawStringAnchored(countdownText, float64(c.w)/2, float64(c.h)/2, 0.5, 0.5)

	return c.image(dc)
}

func (c *Countdown) createFrameRounded(days, hours, minutes, seconds int) *image.Paletted {
	dc := gg.NewContext(c.w, c.h)

	dc.SetColor(c.bg)
	dc.Clear()

	circleRadius := 65.0
	spacing := 160.0
	startX := float64(c.w)/2 - 1.5*spacing
	y := float64(c.h) / 2

	if c.kind == "rounded-ticks" || c.kind == "rounded-dots" {
		c.drawDotsOrTicks(dc, startX, y, circleRadius, days, 31, "DAYS")
		c.drawDotsOrTicks(dc, startX+spacing, y, circleRadius, hours, 24, "HOURS")
		c.drawDotsOrTicks(dc, startX+2*spacing, y, circleRadius, minutes, 60, "MINUTES")
		c.drawDotsOrTicks(dc, startX+3*spacing, y, circleRadius, seconds, 60, "SECONDS")
	} else {
		c.drawCircle(dc, startX, y, circleRadius, days, 31, "DAYS")
		c.drawCircle(dc, startX+spacing, y, circleRadius, hours, 24, "HOURS")
		c.drawCircle(dc, startX+2*spacing, y, circleRadius, minutes, 60, "MINUTES")
		c.drawCircle(dc, startX+3*spacing, y, circleRadius, seconds, 60, "SECONDS")
	}

	return c.image(dc)
}

func (c *Countdown) drawCircle(dc *gg.Context, x, y, radius float64, value int, max int, label string) {
	// Draw outer circle
	dc.SetColor(c.blendColorOverAlpha(50))
	dc.SetLineWidth(10)
	dc.DrawArc(x, y, radius, 0, 2*math.Pi)
	dc.Stroke()

	// Draw progress arc
	dc.SetColor(c.color)
	dc.SetLineWidth(10)
	startAngle := -math.Pi / 2
	angle := startAngle + float64(value)/float64(max)*2*math.Pi

	if angle > 2*math.Pi {
		angle = 2 * math.Pi
	}

	dc.DrawArc(x, y, radius, startAngle, angle)
	dc.Stroke()

	// Draw value text
	face, err := c.loadFont(40)
	if err == nil {
		dc.SetFontFace(face)
	}
	dc.SetColor(c.color)
	dc.DrawStringAnchored(fmt.Sprintf("%d", value), x, y-10, 0.5, 0.5)

	// Draw label text
	face, err = c.loadFont(16)
	if err == nil {
		dc.SetFontFace(face)
	}
	dc.DrawStringAnchored(label, x, y+25, 0.5, 0.5)
}

func (c *Countdown) drawDotsOrTicks(dc *gg.Context, x, y, radius float64, value int, max int, label string) {
	// Draw tick marks
	dc.SetColor(c.color)
	progress := int(float64(value) / float64(max) * float64(max))

	if c.kind == "rounded-dots" {
		c.drawDot(dc, x, y, radius, max, progress)
	} else {
		c.drawTick(dc, x, y, radius, max, progress)
	}

	// Draw value text
	face, err := c.loadFont(40)

	if err == nil {
		dc.SetFontFace(face)
	}
	dc.SetColor(c.color)
	dc.DrawStringAnchored(fmt.Sprintf("%d", value), x, y-10, 0.5, 0.5)

	// Draw label text
	face, err = c.loadFont(16)
	if err == nil {
		dc.SetFontFace(face)
	}
	dc.DrawStringAnchored(label, x, y+25, 0.5, 0.5)
}

func (c *Countdown) drawDot(dc *gg.Context, x, y, radius float64, count int, progress int) {
	for i := 0; i < count; i++ {
		startAngle := -math.Pi / 2
		angle := startAngle + float64(i)*2*math.Pi/float64(count)

		startX := x + (radius+5)*math.Cos(angle)
		startY := y + (radius+5)*math.Sin(angle)

		if i <= progress {
			dc.SetColor(c.color)
		} else {
			dc.SetColor(c.blendColorOverAlpha(50))
		}

		dc.DrawCircle(startX, startY, 3)
		dc.Fill()
	}
}

func (c *Countdown) drawTick(dc *gg.Context, x, y, radius float64, count int, progress int) {
	dc.SetLineWidth(3)

	for i := 0; i < count; i++ {
		startAngle := -math.Pi / 2
		angle := startAngle + float64(i)*2*math.Pi/float64(count)

		startX := x + (radius-5)*math.Cos(angle)
		startY := y + (radius-5)*math.Sin(angle)
		endX := x + (radius+5)*math.Cos(angle)
		endY := y + (radius+5)*math.Sin(angle)

		if i <= progress {
			dc.SetColor(c.color)
			// dc.SetLineWidth(3)
		} else {
			dc.SetColor(c.blendColorOverAlpha(50))
			// dc.SetLineWidth(1)
		}

		dc.DrawLine(startX, startY, endX, endY)
		dc.Stroke()
	}
}