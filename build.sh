#!/usr/bin/env bash
# Build a standalone executable with PyInstaller.
set -e

VENV_DIR=".venv"

if [ ! -d "$VENV_DIR" ]; then
    echo "Creating virtual environment..."
    python3 -m venv "$VENV_DIR"
fi

source "$VENV_DIR/bin/activate"

echo "Installing dependencies..."
pip install -r requirements.txt
pip install pyinstaller

echo "Writing build info..."
python3 -c "
from datetime import datetime
ts = datetime.now().strftime('%Y%m%d_%H%M%S')
with open('build_info.py', 'w') as f:
    f.write(f'BUILD_ID = \"{ts}\"\n')
print(f'BUILD_ID = {ts}')
"

echo "Building main app..."
pyinstaller --onefile --windowed --name "CT Musikteam" main.py

echo "Building updater..."
pyinstaller --onefile --windowed --name "Updater" updater.py

echo ""
echo "Done. Output:"
ls dist/
