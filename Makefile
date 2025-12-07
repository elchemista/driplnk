.PHONY: dev build run test clean setup

# Development (sources .env automatically)
dev:
	@bash -c 'set -a; [ -f .env ] && source .env; set +a; make -j3 dev-go dev-js dev-css'

dev-go:
	@bash -c 'set -a; [ -f .env ] && source .env; set +a; air'

dev-js:
	cd assets && npm run watch:js

dev-css:
	cd assets && npm run watch:css

# Build
build: build-assets build-go

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
