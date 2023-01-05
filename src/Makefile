.PHONY: dev
dev:
	wrangler dev

.PHONY: build
build:
	mkdir -p dist
	tinygo build -o ./dist/app.wasm -target wasm ./...
	# GOOS=js GOARCH=wasm go build -ldflags "-s -w" -gcflags=all="-l -B" -o ./dist/app.wasm

.PHONY: debug
debug:
	tinygo gdb -target wasm ./...

.PHONY: publish
publish:
	wrangler publish
