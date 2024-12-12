GOOS=js GOARCH=wasm go build -o main.wasm *.go
cp "$(go env GOROOT)/misc/wasm/wasm_exec.js" .