#!/usr/bin/env bash
# Doppelklick auf diese Datei, um CT Musikteam neu zu bauen und alte Daten zu loeschen.

cd "$(dirname "$0")/.."

DATA_FILE="$HOME/.ct_musikteam/data.json"

echo "========================================"
echo "  CT Musikteam Updater"
echo "========================================"
echo ""

# Clear old data
if [ -f "$DATA_FILE" ]; then
  rm "$DATA_FILE"
  echo "Alte Daten geloescht."
else
  echo "Keine alten Daten gefunden."
fi
echo ""

# Run the build
bash build.sh

echo ""
echo "========================================"
echo "  Du kannst jetzt CT Musikteam starten."
echo "========================================"
echo ""
read -r -p "Enter druecken zum Schliessen..."
