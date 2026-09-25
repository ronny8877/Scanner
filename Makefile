.PHONY: all run start single-binary build build-web dev scan inspect crawl serve install-web clean

all: run

run: single-binary
	./bin/scanner serve --port 8080

start: run

single-binary:
	@if command -v npm >/dev/null 2>&1; then \
		if [ ! -d "web/node_modules" ]; then npm --prefix web install; fi; \
		npm --prefix web run build; \
	fi
	@mkdir -p bin
	go build -buildvcs=false -o bin/scanner ./cmd/scanner

build:
	@mkdir -p bin
	go build -buildvcs=false -o bin/scanner ./cmd/scanner

build-web:
	npm --prefix web run build

install-web:
	npm --prefix web install

serve: build
	./bin/scanner serve --port 8080

dev: build
	@./bin/scanner serve --port 8080 & \
	SERVER_PID=$$!; \
	trap "kill $$SERVER_PID 2>/dev/null || true" EXIT INT TERM; \
	npm --prefix web run dev

scan: build
	./bin/scanner scan svelte golang --tlds com,dev,org,io --mutations

inspect: build
	./bin/scanner inspect cloudflare.com

crawl: build
	./bin/scanner crawl svelte.dev --pages 15 --depth 2

clean:
	rm -rf bin/scanner
