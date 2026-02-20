#!/usr/bin/env python3
"""
CT Musikteam Updater
Clears stored data so the new version starts fresh.
"""

import sys
import subprocess
from pathlib import Path
from PyQt6.QtWidgets import (
    QApplication, QWidget, QVBoxLayout, QLabel, QPushButton, QMessageBox
)
from PyQt6.QtCore import Qt

DATA_FILE = Path.home() / ".ct_musikteam" / "data.json"
APP_NAME = "CT Musikteam"
APP_PATH = Path(__file__).parent / "dist" / f"{APP_NAME}.app"


class UpdaterWindow(QWidget):
    def __init__(self):
        super().__init__()
        self.setWindowTitle("CT Musikteam Updater")
        self.setFixedSize(400, 220)

        layout = QVBoxLayout(self)
        layout.setSpacing(16)
        layout.setContentsMargins(32, 32, 32, 32)

        title = QLabel("CT Musikteam Update")
        font = title.font()
        font.setPointSize(16)
        font.setBold(True)
        title.setFont(font)
        title.setAlignment(Qt.AlignmentFlag.AlignCenter)
        layout.addWidget(title)

        subtitle = QLabel(
            "Klicke auf 'Jetzt aktualisieren' um die alten Daten zu\n"
            "löschen und die neue Version vorzubereiten."
        )
        subtitle.setAlignment(Qt.AlignmentFlag.AlignCenter)
        subtitle.setWordWrap(True)
        layout.addWidget(subtitle)

        self.update_btn = QPushButton("Jetzt aktualisieren")
        self.update_btn.setFixedHeight(40)
        self.update_btn.clicked.connect(self._do_update)
        layout.addWidget(self.update_btn)

    def _do_update(self):
        try:
            if DATA_FILE.exists():
                DATA_FILE.unlink()
            self.update_btn.setEnabled(False)
            self.update_btn.setText("Erfolgreich aktualisiert!")

            reply = QMessageBox.information(
                self,
                "Update abgeschlossen",
                "Die Daten wurden zurückgesetzt.\n\nMöchtest du CT Musikteam jetzt starten?",
                QMessageBox.StandardButton.Yes | QMessageBox.StandardButton.No
            )
            if reply == QMessageBox.StandardButton.Yes:
                subprocess.Popen(["open", str(APP_PATH)])
            QApplication.quit()
        except Exception as e:
            QMessageBox.critical(self, "Fehler", f"Update fehlgeschlagen:\n{e}")


if __name__ == "__main__":
    app = QApplication(sys.argv)
    app.setApplicationName("CT Musikteam Updater")
    window = UpdaterWindow()
    window.show()
    sys.exit(app.exec())
