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
echo "Erstelle DMG..."
STAGING="release/.dmg_staging"
rm -rf "$STAGING"
mkdir -p "$STAGING"
cp -r "release/CT Musikteam.app" "$STAGING/"
ln -s /Applications "$STAGING/Applications"

hdiutil create \
  -volname "CT Musikteam" \
  -srcfolder "$STAGING" \
  -ov \
  -format UDZO \
  -o "release/CT Musikteam.dmg"

rm -rf "$STAGING"

echo ""
echo "Fertig."
echo "  App: release/CT Musikteam.app"
echo "  DMG: release/CT Musikteam.dmg"
