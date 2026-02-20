from datetime import datetime

from PyQt6.QtWidgets import (
    QWidget, QVBoxLayout, QHBoxLayout, QPushButton,
    QTableWidget, QTableWidgetItem, QComboBox,
    QStyledItemDelegate, QMessageBox
)
from PyQt6.QtCore import Qt, QDate
from data_store import DataStore


class TeamComboDelegate(QStyledItemDelegate):
    """Renders a QComboBox in the Musikteam column."""

    def __init__(self, store: DataStore, parent=None):
        super().__init__(parent)
        self.store = store

    def createEditor(self, parent, option, index):
        combo = QComboBox(parent)
        combo.addItems(self.store.get_teams())
        return combo

    def setEditorData(self, editor, index):
        val = index.data(Qt.ItemDataRole.EditRole) or ""
        i = editor.findText(val)
        if i >= 0:
            editor.setCurrentIndex(i)

    def setModelData(self, editor, model, index):
        model.setData(index, editor.currentText(), Qt.ItemDataRole.EditRole)

    def updateEditorGeometry(self, editor, option, index):
        editor.setGeometry(option.rect)


class EinsaetzeTab(QWidget):
    def __init__(self, store: DataStore):
        super().__init__()
        self.store = store
        self._suppress_save = False
        self._build_ui()
        self._load_from_store()

    def _build_ui(self):
        layout = QVBoxLayout(self)

        toolbar = QHBoxLayout()
        for label, slot in [
            ("Einsatz hinzufügen", self._add_einsatz),
            ("Einsatz entfernen", self._remove_einsatz),
        ]:
            btn = QPushButton(label)
            btn.clicked.connect(slot)
            toolbar.addWidget(btn)
        toolbar.addStretch()
        layout.addLayout(toolbar)

        self.table = QTableWidget()
        self.table.setColumnCount(2)
        self.table.setHorizontalHeaderLabels(["Datum (TT.MM.JJJJ)", "Musikteam"])
        self.table.setItemDelegateForColumn(1, TeamComboDelegate(self.store, self))
        self.table.horizontalHeader().setStretchLastSection(True)
        self.table.itemChanged.connect(self._on_item_changed)
        layout.addWidget(self.table)

    def _load_from_store(self):
        self._suppress_save = True
        einsaetze = self.store.get_einsaetze()
        self.table.setRowCount(len(einsaetze))
        for row, e in enumerate(einsaetze):
            datum_item = QTableWidgetItem(e.get("datum", ""))
            team_item = QTableWidgetItem(e.get("musikteam", ""))
            self.table.setItem(row, 0, datum_item)
            self.table.setItem(row, 1, team_item)
        self.table.resizeColumnsToContents()
        self._suppress_save = False

    def refresh(self):
        """Called on tab switch — reloads to reflect team changes from Musikteam tab."""
        self._load_from_store()

    def _collect_and_save(self):
        if self._suppress_save:
            return
        einsaetze = []
        for row in range(self.table.rowCount()):
            datum_item = self.table.item(row, 0)
            team_item = self.table.item(row, 1)
            datum = datum_item.text().strip() if datum_item else ""
            team = team_item.text().strip() if team_item else ""
            if datum and self._valid_date(datum):
                einsaetze.append({"datum": datum, "musikteam": team})
        self.store.set_einsaetze(einsaetze)

    @staticmethod
    def _valid_date(s: str) -> bool:
        try:
            datetime.strptime(s, "%d.%m.%Y")
            return True
        except ValueError:
            return False

    def _on_item_changed(self, item):
        if item.column() == 0:
            if not self._valid_date(item.text().strip()):
                item.setBackground(Qt.GlobalColor.red)
            else:
                item.setBackground(Qt.GlobalColor.white)
        self._collect_and_save()

    def _add_einsatz(self):
        self._suppress_save = True
        row = self.table.rowCount()
        self.table.insertRow(row)
        today = QDate.currentDate().toString("dd.MM.yyyy")
        teams = self.store.get_teams()
        self.table.setItem(row, 0, QTableWidgetItem(today))
        self.table.setItem(row, 1, QTableWidgetItem(teams[0] if teams else ""))
        self._suppress_save = False
        self._collect_and_save()

    def _remove_einsatz(self):
        row = self.table.currentRow()
        if row < 0:
            QMessageBox.warning(self, "Fehler", "Bitte eine Zeile auswählen.")
            return
        self._suppress_save = True
        self.table.removeRow(row)
        self._suppress_save = False
        self._collect_and_save()
