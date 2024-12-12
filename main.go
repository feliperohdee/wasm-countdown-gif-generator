//go:build js && wasm

package main

import (
	"embed"
	"fmt"
	"image/color"
	"strings"
	"syscall/js"
	"time"
)

//go:embed fonts/*.ttf
var fontFS embed.FS

var ALLOWED_FONTS = map[string]bool{
	"impact": true,
}

func main() {
	js.Global().Set("buildCountdown", js.FuncOf(buildCountdown))
	js.Global().Set("buildLedBanner", js.FuncOf(buildLedBanner))
	js.Global().Set("buildFlashingLetters", js.FuncOf(buildFlashingLetters))
	js.Global().Set("buildFlashingText", js.FuncOf(buildFlashingText))
	js.Global().Set("buildColorVaryingText", js.FuncOf(buildColorVaryingText))
	js.Global().Set("buildTypingText", js.FuncOf(buildTypingText))

	select {}
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
