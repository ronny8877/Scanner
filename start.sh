#!/usr/bin/env bash
set -e

if command -v npm >/dev/null 2>&1; then
  if [ ! -d "web/node_modules" ]; then
    echo "Installing frontend dependencies..."
    npm --prefix web install --silent
  fi
  echo "Building web frontend into web/dist..."
  npm --prefix web run build
fi

echo "Building ./bin/scanner..."
mkdir -p bin
go build -buildvcs=false -o bin/scanner ./cmd/scanner

echo "Starting server on http://localhost:8080"
exec ./bin/scanner serve --port 8080
