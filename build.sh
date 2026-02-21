#!/usr/bin/env bash
# Build CT Musikteam with Go + Wails.
set -e

cd "$(dirname "$0")"

# Check Go
if ! command -v go &>/dev/null; then
  echo "ERROR: Go is not installed. Install from https://go.dev/dl/"
  exit 1
fi

# Install Wails CLI if needed
if ! command -v wails &>/dev/null; then
  echo "Installing Wails CLI..."
  go install github.com/wailsapp/wails/v2/cmd/wails@latest
  export PATH="$PATH:$(go env GOPATH)/bin"
fi

# Download Go dependencies
echo "Fetching dependencies..."
go mod tidy

# Inject build timestamp
BUILD_ID=$(date +%Y%m%d_%H%M%S)
echo "BUILD_ID = $BUILD_ID"

# Build
echo "Building..."
wails build -ldflags "-X 'main.BuildID=${BUILD_ID}'"

echo ""
echo "Kopiere App nach release/..."
mkdir -p release
rm -rf "release/CT Musikteam.app"
cp -r "build/bin/CT Musikteam.app" "release/CT Musikteam.app"

echo ""
echo "Fertig. Die App liegt in: release/CT Musikteam.app"
