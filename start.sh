#!/usr/bin/env bash
set -e

echo "================================================================"
echo " ◈ Scanner Studio — One-Command Single-Binary Builder & Launcher"
echo "================================================================"

# 1. If npm is installed and web/src was modified, rebuild web/dist automatically
if command -v npm >/dev/null 2>&1; then
  if [ ! -d "web/node_modules" ]; then
    echo "→ Installing Svelte 5 frontend dependencies (one-time)..."
    npm --prefix web install --silent
  fi
  echo "→ Compiling Svelte 5 Web Studio into web/dist..."
  npm --prefix web run build
else
  echo "→ Using pre-built Svelte 5 Web Studio bundle in web/dist..."
fi

# 2. Compile the single self-contained Go binary (embeds web/dist + Go API + CLI)
echo "→ Building single self-contained binary: ./bin/scanner ..."
mkdir -p bin
go build -buildvcs=false -o bin/scanner ./cmd/scanner

echo ""
echo "✓ Single binary built at: ./bin/scanner"
echo "✓ Starting Scanner Studio on http://localhost:8080 ..."
echo "================================================================"
echo ""

exec ./bin/scanner serve --port 8080
