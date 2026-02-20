from PyQt6.QtWidgets import (
    QWidget, QVBoxLayout, QHBoxLayout, QPushButton,
    QTableWidget, QTableWidgetItem
)
from PyQt6.QtCore import Qt
from data_store import DataStore


class TeamleiterTab(QWidget):
    def __init__(self, store: DataStore):
        super().__init__()
        self.store = store
        self._suppress_save = False
        self._build_ui()
        self._load_from_store()

    def _build_ui(self):
        layout = QVBoxLayout(self)

        toolbar = QHBoxLayout()
        reload_btn = QPushButton("Neu laden")
        reload_btn.clicked.connect(self.refresh)
        toolbar.addWidget(reload_btn)
        toolbar.addStretch()
        layout.addLayout(toolbar)

        self.table = QTableWidget()
        self.table.setColumnCount(2)
        self.table.setHorizontalHeaderLabels(["Team", "Bezeichnung"])
        self.table.horizontalHeader().setStretchLastSection(True)
        self.table.itemChanged.connect(self._on_item_changed)
        layout.addWidget(self.table)

    def _load_from_store(self):
        self._suppress_save = True
        teams = self.store.get_teams()
        teamleiter = self.store.get_teamleiter()

        self.table.setRowCount(len(teams))
        for row, team in enumerate(teams):
            team_item = QTableWidgetItem(team)
            team_item.setFlags(team_item.flags() & ~Qt.ItemFlag.ItemIsEditable)
            self.table.setItem(row, 0, team_item)

            text_item = QTableWidgetItem(teamleiter.get(team, ""))
            self.table.setItem(row, 1, text_item)

        self.table.resizeColumnsToContents()
        self._suppress_save = False

    def refresh(self):
        """Sync team list, preserving existing Bezeichnung values, called on tab switch."""
        current = self._collect_teamleiter()
        teams = self.store.get_teams()
        merged = {t: current.get(t, "") for t in teams}
        # Preserve default values for new teams that have no text yet
        defaults = self.store.get_teamleiter()
        for t in teams:
            if not merged[t] and t in defaults:
                merged[t] = defaults[t]
        self.store.set_teamleiter(merged)
        self._load_from_store()

    def _collect_teamleiter(self) -> dict:
        result = {}
        for row in range(self.table.rowCount()):
            team_item = self.table.item(row, 0)
            text_item = self.table.item(row, 1)
            if team_item:
                result[team_item.text()] = text_item.text() if text_item else ""
        return result

    def _on_item_changed(self, item):
        if item.column() == 1:
            if not self._suppress_save:
                self.store.set_teamleiter(self._collect_teamleiter())
