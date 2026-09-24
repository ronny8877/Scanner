# ◈ Scanner — Domain Intelligence, Value Finder & Site Structure Suite

An all-in-one monorepo featuring a **Go CLI & API Server** (`Cobra` + `Lipgloss`) paired with a **Svelte 5 + Vite + TypeScript + Tailwind CSS** web studio (`web/`).

## Core Modes

1. **Valuable Available Domain Scanner (`scanner scan` / Default Mode)**
   - Generates brandable root & affix candidates across top TLDs (`.com`, `.ai`, `.io`, `.dev`, `.co`, `.app`, `.net`, `.xyz`).
   - Checks availability in parallel and scores each domain (`0-100`) based on length scarcity, TLD authority, phonetic cadence, and commercial keyword intent.
2. **Deep Domain Inquiry (`scanner inspect <domain>`)**
   - Queries authoritative **ICANN RDAP** endpoints and live **DNS resolvers** (`NS`, `A`, `AAAA`, `MX`, `TXT`).
   - Displays exact registration creation date, calculated domain age, expiration date & remaining days, sponsoring registrar, and EPP status flags.
3. **Site Structure Crawler (`scanner crawl <domain>`)**
   - Crawls a live website up to a configurable page/depth limit and constructs a **hierarchical site URL tree** rendered in both the terminal and the interactive Svelte tree inspector.
4. **API Server (`scanner serve`)**
   - Runs the local HTTP API server on `http://localhost:8080` to power the Svelte 5 UI.

---

## Quick Start

### 1. Run the Go CLI Directly

```bash
# Build the CLI binary
make build

# Default Mode: Scan for available (empty) high-value domains
./bin/scanner nova pulse --available

# Mode 2: Inquire when a domain was registered, its age, registrar & DNS
./bin/scanner inspect svelte.dev

# Mode 3: Crawl a domain and print its hierarchical site structure tree
./bin/scanner crawl svelte.dev --pages 20 --depth 2
```

### 2. Run the Full Stack (Go API Server + Svelte 5 UI)

In **Terminal 1** (Start the Go API Server on `:8080`):
```bash
make serve
```

In **Terminal 2** (Install & run the Svelte 5 Web Studio on `:5173`):
```bash
make install-web
make web
```
Then open **http://localhost:5173** in your browser.
