# Scanner

[![License: MIT](https://img.shields.io/badge/License-MIT-black.svg?style=flat-square)](./LICENSE)
[![Go](https://img.shields.io/badge/Go-1.22+-00ADD8?style=flat-square)](https://go.dev/)
[![Svelte 5](https://img.shields.io/badge/Svelte-5-FF3E00?style=flat-square)](https://svelte.dev/)

Domain inspection, DNS/RDAP lookup, port reconnaissance, and site crawling tool written in Go with an embedded Svelte 5 web interface.

The project compiles into a single binary (`./bin/scanner`) using `go:embed`. Running `./bin/scanner serve` hosts both the web interface and the HTTP API on `http://localhost:8080`.

---

## Quick Start (One Command)

Build and start the server on `http://localhost:8080`:

```bash
./start.sh
```

Or using `make`:

```bash
make run
```

Then open `http://localhost:8080`.

---

## Single-Binary Build

The pre-built web assets in `web/dist` are embedded into the Go binary via [`web/embed.go`](./web/embed.go).

```bash
# Build frontend + single Go binary (requires Go and Node.js)
make single-binary

# Build Go binary only using existing web/dist (requires Go only)
go build -buildvcs=false -o bin/scanner ./cmd/scanner
```

Run the compiled binary directly:

```bash
./bin/scanner serve --port 8080
```

---

## Development Mode

To run the Go API server (`:8080`) alongside the Vite dev server (`:5173`) with hot reload:

```bash
make dev
```

---

## Modules

| Module | Endpoint / Command | Description |
| :--- | :--- | :--- |
| **01. Find & Value** | `POST /api/scan` / `scanner scan` | Checks domain availability across 24 TLDs using DNS (`NS`, `SOA`, `A`, `MX`), RDAP, Port-43 WHOIS, Certificate Transparency (`crt.sh`), and Wayback CDX. Computes heuristic price estimates. |
| **02. RDAP Dossier** | `GET /api/inspect` / `scanner inspect` | Queries ICANN RDAP and Port-43 WHOIS for registration dates, registrar, EPP status flags, DNSSEC, DMARC, and DNS zone records (`NS`, `A`, `AAAA`, `MX`, `TXT`, `CNAME`). |
| **03. Site Tree** | `POST /api/crawl` / `scanner crawl` | Concurrent breadth-first crawler that maps internal routes, status codes, latencies, and page hierarchy. Falls back to `crt.sh` and archive index routes when blocked by WAF rules. |
| **04. Traffic & Rank** | `GET /api/traffic` | Queries Tranco Top-1M rankings and Cloudflare Radar rank buckets to estimate monthly visit ranges and detect frontend/infrastructure stack signatures. |
| **05. Ports & TLS** | `GET /api/recon` | Scans 14 common TCP ports (`21`, `22`, `25`, `53`, `80`, `443`, `3306`, `3389`, `5432`, `6379`, `8080`, `8443`, `9200`, `27017`), inspects TLS handshakes/SANs, and filters wildcard DNS entries. |
| **06. Robots & Sitemap** | `GET /api/robots` | Parses `robots.txt` directives across search/AI user agents and extracts URLs from `sitemap.xml`. |
| **07. Social Meta & Ads** | `GET /api/meta` | Extracts OpenGraph/Twitter meta tags and checks HTML scripts for common analytics and advertising tags. |
| **08. Saved Vault** | `GET/POST /api/watchlist` | Local JSON watchlist (`data/watchlist.json`) for saving domains and re-checking availability. |
| **Report Generator** | `POST /api/parallel-suite` | Accessible from the **Queue** menu (`Make Report`). Runs all inspection modules concurrently for a target domain and formats a printable summary report. |

---

## CLI Usage

```bash
# Check domain availability across specified TLDs
./bin/scanner scan svelte golang --tlds com,dev,org,io --mutations

# Inspect RDAP, WHOIS, and DNS records
./bin/scanner inspect cloudflare.com
./bin/scanner inspect svelte.dev
./bin/scanner inspect golang.org

# Crawl a domain and print its URL tree
./bin/scanner crawl svelte.dev --pages 20 --depth 2

# Start HTTP server + embedded web UI
./bin/scanner serve --port 8080
```

---

## Project Layout

```text
Scanner/
├── cmd/scanner/main.go   # CLI entrypoint
├── internal/
│   ├── cli/              # Cobra CLI commands (scan, inspect, crawl, serve)
│   ├── crawler/          # Concurrent BFS link crawler
│   ├── domain/           # DNS/RDAP/WHOIS availability & heuristic valuation
│   ├── jobs/             # In-memory background job tracker
│   ├── recon/            # TCP port scanner, TLS inspector, subdomain verifier
│   ├── server/           # HTTP handlers and embedded static file server
│   ├── traffic/          # Tranco list & Cloudflare rank lookup
│   ├── watchlist/        # JSON file store for saved domains
│   └── webintel/         # HTML meta, robots.txt, sitemap.xml, and stack parser
├── web/
│   ├── embed.go          # go:embed directive for web/dist
│   ├── dist/             # Compiled static assets
│   └── src/              # Svelte 5 frontend source
├── Makefile              # Build and run targets
├── start.sh              # Single-command build and run script
└── LICENSE               # MIT License
```

---

## License

[MIT License](./LICENSE)
