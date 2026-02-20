#!/usr/bin/env bash
# Launch CT Musikteam, rebuilding first if any source file has changed.
set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
cd "$SCRIPT_DIR"

APP="dist/CT Musikteam.app"
SOURCES=(main.py ct_api.py data_store.py requirements.txt tabs/*.py)

needs_rebuild() {
    [ ! -e "$APP" ] && return 0
    for src in "${SOURCES[@]}"; do
        [ "$src" -nt "$APP" ] && return 0
    done
    return 1
}

if needs_rebuild; then
    echo "Source changed — rebuilding..."
    bash build.sh
else
    echo "No changes detected, skipping rebuild."
fi

echo "Launching..."
open "$APP"
