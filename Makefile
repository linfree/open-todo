.PHONY: all web build run dev clean test

all: web build

# Frontend build (at project root)
web:
	pnpm install && pnpm run build

# PWA build (output stays at root dist/)
web-pwa:
	pnpm install && pnpm run build:pwa

build:
	go build -ldflags="-s -w" -o bin/open-todo ./cmd/open-todo/

build-win:
	GOOS=windows GOARCH=amd64 go build -ldflags="-H windowsgui -s -w" -o bin/open-todo.exe ./cmd/open-todo/

build-linux:
	GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o bin/open-todo-linux ./cmd/open-todo/

build-mac:
	GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o bin/open-todo-mac ./cmd/open-todo/

run:
	go run ./cmd/open-todo/

dev:
	@echo "Starting frontend dev server and Go backend..."
	pnpm run dev &
	go run ./cmd/open-todo/

clean:
	rm -rf bin/ cmd/open-todo/web-dist/

test:
	go test ./... -v -count=1 -timeout 60s
