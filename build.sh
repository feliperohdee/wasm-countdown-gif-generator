GOOS=js GOARCH=wasm go build -o main.wasm gif.go
cp "$(go env GOROOT)/misc/wasm/wasm_exec.js" .