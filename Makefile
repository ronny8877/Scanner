.PHONY: all run start single-binary build build-web dev scan inspect crawl serve install-web clean

## Default target: Build the single self-contained binary and start the server on :8080
all: run

## ONE COMMAND TO BUILD & RUN EVERYTHING (Single Binary on http://localhost:8080)
run: single-binary
	./bin/scanner serve --port 8080

## Alias for `make run`
start: run

## Build ONE self-contained binary (`./bin/scanner`) containing BOTH the Go API Engine + Svelte 5 Web UI
single-binary:
	@if command -v npm >/dev/null 2>&1; then \
		if [ ! -d "web/node_modules" ]; then npm --prefix web install; fi; \
		npm --prefix web run build; \
	fi
	@mkdir -p bin
	go build -buildvcs=false -o bin/scanner ./cmd/scanner
	@echo "✓ Single self-contained binary ready at ./bin/scanner"

## Fast Go-only build (uses embedded pre-built web/dist without needing Node.js)
build:
	@mkdir -p bin
	go build -buildvcs=false -o bin/scanner ./cmd/scanner

## Build Svelte 5 production bundle into web/dist
build-web:
	npm --prefix web run build

## Install frontend dependencies in ./web
install-web:
	npm --prefix web install

## Start the Go HTTP Server + Embedded Web Studio on http://localhost:8080
serve: build
	./bin/scanner serve --port 8080

## Start hot-reload development mode (Go API on :8080 + Vite Dev Server on :5173)
dev: build
	@echo "Starting Go API (:8080) and Svelte 5 Vite Dev Server (:5173)..."
	@./bin/scanner serve --port 8080 & \
	SERVER_PID=$$!; \
	trap "kill $$SERVER_PID 2>/dev/null || true" EXIT INT TERM; \
	npm --prefix web run dev

## CLI Mode 1: Scan for available high-value domains
scan: build
	./bin/scanner scan nova pulse --tlds com,ai,io,dev,co,app --mutations

## CLI Mode 2: Inquire about a domain's registration date, age, registrar & DNS
inspect: build
	./bin/scanner inspect svelte.dev

## CLI Mode 3: Crawl a site and render its hierarchical URL structure tree
crawl: build
	./bin/scanner crawl svelte.dev --pages 15 --depth 2

## Clean compiled binary
clean:
	rm -rf bin/scanner
