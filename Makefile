.PHONY: dev build run test clean setup generate

# Development - builds assets first, then runs watchers in parallel
# Development - builds assets first, then runs watchers in parallel
dev: build-assets
	@bash -c 'set -a; [ -f .env ] && source .env; set +a; make -j2 dev-go dev-assets'

dev-go:
	@bash -c 'set -a; [ -f .env ] && source .env; set +a; air'

dev-assets:
	cd assets && npm run watch

# Build
build: generate build-assets build-go

build-assets:
	cd assets && npm run build

build-go:
	go build -o tmp/main ./cmd/server

# Setup
setup:
	cd assets && npm install
	go mod download
	go install github.com/a-h/templ/cmd/templ@latest
	go install github.com/air-verse/air@latest

# Templ generation
generate:
	templ generate

# Test
test:
	go test -v ./...

# Clean
clean:
	rm -rf tmp/
	rm -rf assets/dist/
