# 🎬 WASM Countdown GIF Generator

Generate beautiful, animated countdown GIFs for your websites and applications using WebAssembly, Go, and Cloudflare Workers! This powerful tool creates customizable countdown animations in various styles and languages.

## 🎥 Live Demos

### Default Style (Rounded)
![Default Countdown](https://countdown.simpleimg.io/?date=2026-01-01)

### Rounded with Custom Colors
![Green Countdown](https://countdown.simpleimg.io/?date=2026-01-01&color=00ff00&background=333333)

### Rounded Dots Style
![Dots Style](https://countdown.simpleimg.io/?date=2026-01-01&kind=rounded-dots)

### Rounded Ticks Style
![Ticks Style](https://countdown.simpleimg.io/?date=2026-01-01&kind=rounded-ticks)

### Basic Style
![Basic Style](https://countdown.simpleimg.io/?date=2026-01-01&kind=basic)

### Different Languages

### [English](https://countdown.simpleimg.io/?date=2026-01-01&lang=en)
![English](https://countdown.simpleimg.io/?date=2026-01-01&lang=en)


### [German](https://countdown.simpleimg.io/?date=2026-01-01&lang=de)
![German](https://countdown.simpleimg.io/?date=2026-01-01&lang=de)

### [Portuguese](https://countdown.simpleimg.io/?date=2026-01-01&lang=pt)
![Portuguese](https://countdown.simpleimg.io/?date=2026-01-01&lang=pt)

### [Spanish](https://countdown.simpleimg.io/?date=2026-01-01&lang=es)
![Spanish](https://countdown.simpleimg.io/?date=2026-01-01&lang=es)

### [Russian](https://countdown.simpleimg.io/?date=2026-01-01&lang=ru)
![Russian](https://countdown.simpleimg.io/?date=2026-01-01&lang=ru)

### [French](https://countdown.simpleimg.io/?date=2026-01-01&lang=fr)
![French](https://countdown.simpleimg.io/?date=2026-01-01&lang=fr)

## ✨ Features

- 🎨 Multiple countdown styles:
  - `rounded`: Clean circular progress indicators
  - `rounded-ticks`: Modern tick-based circular display
  - `rounded-dots`: Elegant dot-based circular display
  - `basic`: Simple text-based countdown

- 🌈 Customizable colors for background and text
- 🌍 Support for 30+ languages
- ⚡ Lightning-fast performance with WebAssembly
- 🔄 Configurable animation frames
- 🌐 GMT offset support
- 💾 Built-in caching with Cloudflare Workers

## 🚀 Quick Start

Simply make a GET request to your worker's URL with your desired parameters:

```
https://countdown.simpleimg.io/?date=2026-01-01&kind=rounded&color=00ff00
```

## 🎮 URL Parameters

| Parameter    | Description                                | Default     | Example           |
|-------------|--------------------------------------------|-------------|-------------------|
| `date`      | Target date for countdown                  | 2025-01-01  | 2024-12-31       |
| `kind`      | Animation style                            | rounded     | rounded-dots      |
| `color`     | Text/progress color (hex)                  | fff        | 00ff00           |
| `background`| Background color (hex)                     | 000        | 333333           |
| `frames`    | Number of animation frames (1-60)          | 10          | 30               |
| `lang`      | Language code                              | en          | es               |
| `gmt`       | GMT offset in hours                        | 0           | -3               |

## 🖼️ Style Examples

### Rounded (Default)
```html
<img src="https://countdown.simpleimg.io/?date=2026-01-01" alt="Rounded Countdown" />
```

### Rounded Dots
```html
<img src="https://countdown.simpleimg.io/?date=2026-01-01&kind=rounded-dots" alt="Dots Countdown" />
```

### Rounded Ticks
```html
<img src="https://countdown.simpleimg.io/?date=2026-01-01&kind=rounded-ticks" alt="Ticks Countdown" />
```

### Basic
```html
<img src="https://countdown.simpleimg.io/?date=2026-01-01&kind=basic" alt="Basic Countdown" />
```

## 🌍 Supported Languages

- English (en)
- Spanish (es)
- French (fr)
- German (de)
- Portuguese (pt)
- Chinese (zh)
- Japanese (ja)
- Korean (ko)
- Russian (ru)
- And many more!

## 🛠️ Development Setup

1. Clone the repository:
```bash
git clone https://github.com/yourusername/wasm-countdown-gif
cd wasm-countdown-gif
```

2. Install dependencies:
```bash
go mod download
npm install
```

3. Build the WebAssembly:
```bash
GOOS=js GOARCH=wasm go build -o main.wasm
```

4. Deploy to Cloudflare Workers:
```bash
wrangler publish
```

## ⚙️ Architecture

- `main.go`: Core Go code compiled to WebAssembly
- `worker.js`: Cloudflare Worker entry point
- `fonts/`: Embedded font files
- Built with:
  - Go's `image` package for GIF generation
  - `gg` library for graphics
  - WebAssembly for browser execution
  - Cloudflare Workers for hosting and caching

## 📝 Example Usage

### Basic Countdown
```html
<img src="https://countdown.simpleimg.io/?date=2026-01-01" alt="Countdown" />
```

### Custom Styled Countdown
```html
<img src="https://countdown.simpleimg.io/?date=2026-01-01&kind=rounded-dots&color=00ff00&background=333333&frames=30&lang=es" alt="Countdown" />
```

## 🙏 Acknowledgments

- [fogleman/gg](https://github.com/fogleman/gg) for graphics generation
- [golang/freetype](https://github.com/golang/freetype) for font rendering
- Cloudflare Workers for hosting and edge computing

## ⭐ Show Your Support

Give a ⭐️ if this project helped you!

## 👨‍💻 Author

**Felipe Rohde**
* Twitter: [@felipe_rohde](https://twitter.com/felipe_rohde)
* Github: [@feliperohdee](https://github.com/feliperohdee)
* Email: feliperohdee@gmail.com