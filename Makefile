.PHONY: build scan inspect crawl serve web install-web

## Build the Go CLI binary into ./bin/scanner
build:
	go build -buildvcs=false -o bin/scanner ./cmd/scanner

## Default Mode: Scan for available high-value domains
scan: build
	./bin/scanner scan nova pulse --tlds com,ai,io,dev,co,app --mutations

## Mode 2: Inquire about a domain's registration date, age, registrar & DNS
inspect: build
	./bin/scanner inspect svelte.dev

## Mode 3: Crawl a site and render its hierarchical URL structure tree
crawl: build
	./bin/scanner crawl svelte.dev --pages 15 --depth 2

## Start the Go HTTP API Server (:8080) for the Svelte UI
serve: build
	./bin/scanner serve --port 8080

## Install frontend dependencies in ./web
install-web:
	npm --prefix web install

## Start the Svelte 5 Vite dev server (:5173)
web:
	npm --prefix web run dev
