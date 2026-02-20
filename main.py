#!/usr/bin/env python3
"""
CT Musikteam — Desktop app for managing ChurchTools music team assignments.
"""

import sys
from PyQt6.QtWidgets import QApplication, QMainWindow, QTabWidget
from PyQt6.QtGui import QIcon
from data_store import DataStore
from tabs.settings_tab import SettingsTab
from tabs.musikteam_tab import MusikteamTab
from tabs.einsaetze_tab import EinsaetzeTab
from tabs.teamleiter_tab import TeamleiterTab
from tabs.run_tab import RunTab


class MainWindow(QMainWindow):
    def __init__(self, store: DataStore):
        super().__init__()
        self.store = store
        self.setWindowTitle("CT Musikteam")
        self.resize(960, 640)

        self.tabs = QTabWidget()

        self.settings_tab = SettingsTab(store)
        self.musikteam_tab = MusikteamTab(store)
        self.einsaetze_tab = EinsaetzeTab(store)
        self.teamleiter_tab = TeamleiterTab(store)
        self.run_tab = RunTab(store)

        self.tabs.addTab(self.settings_tab, "Einstellungen")
        self.tabs.addTab(self.musikteam_tab, "Musikteam")
        self.tabs.addTab(self.einsaetze_tab, "Einsätze")
        self.tabs.addTab(self.teamleiter_tab, "Teamleiter")
        self.tabs.addTab(self.run_tab, "Ausführen")

        self.tabs.currentChanged.connect(self._on_tab_changed)

        self.setCentralWidget(self.tabs)

    def _on_tab_changed(self, index: int):
        widget = self.tabs.widget(index)
        if hasattr(widget, "refresh"):
            widget.refresh()


if __name__ == "__main__":
    app = QApplication(sys.argv)
    app.setApplicationName("CT Musikteam")
    store = DataStore()
    window = MainWindow(store)
    window.show()
    sys.exit(app.exec())
