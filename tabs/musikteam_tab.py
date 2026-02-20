from PyQt6.QtWidgets import (
    QWidget, QVBoxLayout, QHBoxLayout, QPushButton, QLabel,
    QTableWidget, QTableWidgetItem, QInputDialog, QMessageBox,
    QDialog, QLineEdit, QListWidget, QListWidgetItem, QDialogButtonBox
)
from PyQt6.QtCore import Qt, QThread, pyqtSignal
from data_store import DataStore
import ct_api

# Fixed columns: 0 = Bezeichnung, 1 = Dienst-ID, team columns start at 2
_COL_LABEL = 0
_COL_ID = 1
_TEAM_COL_START = 2


class PersonPickerDialog(QDialog):
    """Popup dialog for picking a person by name search."""

    def __init__(self, persons: list, current_value: str = "", parent=None):
        super().__init__(parent)
        self.setWindowTitle("Person auswählen")
        self.setMinimumSize(400, 350)
        self._persons = persons  # list of {"id": int, "name": str}
        self._selected_id = ""

        layout = QVBoxLayout(self)

        # Search field
        self._search = QLineEdit()
        self._search.setPlaceholderText("Name eingeben...")
        self._search.setFixedHeight(36)
        font = self._search.font()
        font.setPointSize(13)
        self._search.setFont(font)
        layout.addWidget(self._search)

        # Results list
        self._list = QListWidget()
        self._list.setAlternatingRowColors(True)
        layout.addWidget(self._list)

        # Buttons
        btn_box = QDialogButtonBox()
        self._clear_btn = QPushButton("Löschen")
        btn_box.addButton(self._clear_btn, QDialogButtonBox.ButtonRole.ResetRole)
        btn_box.addButton(QDialogButtonBox.StandardButton.Ok)
        btn_box.addButton(QDialogButtonBox.StandardButton.Cancel)
        layout.addWidget(btn_box)

        self._populate_list("")

        # Pre-select current value if it matches a person
        if current_value:
            for i in range(self._list.count()):
                item = self._list.item(i)
                if item.data(Qt.ItemDataRole.UserRole) == current_value:
                    self._list.setCurrentItem(item)
                    break
            # Also pre-fill search with current name for context
            for p in persons:
                if str(p["id"]) == current_value:
                    self._search.setText(p["name"])
                    self._search.selectAll()
                    break

        self._search.textChanged.connect(self._on_search)
        self._list.itemDoubleClicked.connect(self._accept_item)
        btn_box.accepted.connect(self._on_ok)
        btn_box.rejected.connect(self.reject)
        self._clear_btn.clicked.connect(self._on_clear)

        self._search.setFocus()

    def _populate_list(self, query: str):
        self._list.clear()
        query = query.lower()
        for p in self._persons:
            if query in p["name"].lower():
                item = QListWidgetItem(f"{p['name']}  (ID: {p['id']})")
                item.setData(Qt.ItemDataRole.UserRole, str(p["id"]))
                self._list.addItem(item)
        if self._list.count() > 0:
            self._list.setCurrentRow(0)

    def _on_search(self, text: str):
        self._populate_list(text)

    def _accept_item(self, item: QListWidgetItem):
        self._selected_id = item.data(Qt.ItemDataRole.UserRole)
        self.accept()

    def _on_ok(self):
        item = self._list.currentItem()
        if item:
            self._selected_id = item.data(Qt.ItemDataRole.UserRole)
        self.accept()

    def _on_clear(self):
        self._selected_id = ""
        self.accept()

    def selected_id(self) -> str:
        return self._selected_id


class FetchPersonsWorker(QThread):
    finished = pyqtSignal(list)   # list of {"id": int, "name": str}
    error = pyqtSignal(str)

    def __init__(self, ct_url: str, api_key: str):
        super().__init__()
        self._ct_url = ct_url
        self._api_key = api_key

    def run(self):
        try:
            persons = ct_api.fetch_persons(self._ct_url, self._api_key)
            self.finished.emit(persons)
        except Exception as e:
            self.error.emit(str(e))


class MusikteamTab(QWidget):
    def __init__(self, store: DataStore):
        super().__init__()
        self.store = store
        self._suppress_save = False
        self._fetch_worker = None
        self._build_ui()
        self._load_from_store()

    def _build_ui(self):
        layout = QVBoxLayout(self)

        # Row 1: table actions
        toolbar = QHBoxLayout()
        for label, slot in [
            ("Team hinzufügen", self._add_team),
            ("Team entfernen", self._remove_team),
            ("Dienst hinzufügen", self._add_dienst),
            ("Dienst entfernen", self._remove_dienst),
        ]:
            btn = QPushButton(label)
            btn.clicked.connect(slot)
            toolbar.addWidget(btn)
        toolbar.addStretch()
        layout.addLayout(toolbar)

        # Row 2: persons actions
        persons_bar = QHBoxLayout()
        self.fetch_btn = QPushButton("Personen aus ChurchTools laden")
        self.fetch_btn.clicked.connect(self._fetch_persons)
        self.persons_status = QLabel("")
        persons_bar.addWidget(self.fetch_btn)
        persons_bar.addWidget(self.persons_status)
        persons_bar.addStretch()
        layout.addLayout(persons_bar)

        self.table = QTableWidget()
        self.table.itemChanged.connect(self._on_item_changed)
        self.table.cellDoubleClicked.connect(self._on_cell_double_clicked)
        layout.addWidget(self.table)

        # Update status label with cached count
        self._update_persons_status()

    def _update_persons_status(self):
        count = len(self.store.get_persons())
        if count:
            self.persons_status.setText(f"{count} Personen im Cache")
        else:
            self.persons_status.setText("Noch keine Personen geladen")

    def _load_from_store(self):
        self._suppress_save = True
        teams = self.store.get_teams()
        dienste = self.store.get_dienste()

        # Columns: Bezeichnung | Dienst-ID | team1 | team2 | ...
        self.table.setColumnCount(_TEAM_COL_START + len(teams))
        self.table.setHorizontalHeaderLabels(["Bezeichnung", "Dienst-ID"] + teams)

        # Make team columns non-editable via keyboard (only via dialog)
        self.table.setRowCount(len(dienste))
        for row, dienst in enumerate(dienste):
            self.table.setItem(row, _COL_LABEL, QTableWidgetItem(dienst.get("label", "")))
            self.table.setItem(row, _COL_ID, QTableWidgetItem(str(dienst.get("id", ""))))
            for col, team in enumerate(teams, start=_TEAM_COL_START):
                val = dienst.get("persons", {}).get(team, "")
                item = QTableWidgetItem(str(val))
                item.setFlags(item.flags() & ~Qt.ItemFlag.ItemIsEditable)
                self.table.setItem(row, col, item)

        self.table.resizeColumnsToContents()
        self._suppress_save = False

    def _on_cell_double_clicked(self, row: int, col: int):
        if col < _TEAM_COL_START:
            return
        persons = self.store.get_persons()
        if not persons:
            QMessageBox.information(
                self, "Keine Personen",
                "Bitte zuerst Personen aus ChurchTools laden."
            )
            return
        current = self.table.item(row, col)
        current_val = current.text().strip() if current else ""
        dlg = PersonPickerDialog(persons, current_val, parent=self)
        if dlg.exec() == QDialog.DialogCode.Accepted:
            self._suppress_save = True
            item = QTableWidgetItem(dlg.selected_id())
            item.setFlags(item.flags() & ~Qt.ItemFlag.ItemIsEditable)
            self.table.setItem(row, col, item)
            self._suppress_save = False
            self._collect_and_save()

    def _current_teams(self) -> list:
        return [
            self.table.horizontalHeaderItem(c).text()
            for c in range(_TEAM_COL_START, self.table.columnCount())
        ]

    def _collect_and_save(self):
        if self._suppress_save:
            return
        teams = self._current_teams()
        dienste = []
        for row in range(self.table.rowCount()):
            id_item = self.table.item(row, _COL_ID)
            label_item = self.table.item(row, _COL_LABEL)
            if not id_item or not id_item.text().strip():
                continue
            persons = {}
            for col, team in enumerate(teams, start=_TEAM_COL_START):
                cell = self.table.item(row, col)
                val = cell.text().strip() if cell else ""
                if val:
                    persons[team] = val
            dienste.append({
                "id": id_item.text().strip(),
                "label": label_item.text().strip() if label_item else "",
                "persons": persons
            })
        self.store.set_teams(teams)
        self.store.set_dienste(dienste)

    def _on_item_changed(self, _item):
        self._collect_and_save()

    # --- Team actions ---

    def _add_team(self):
        name, ok = QInputDialog.getText(self, "Team hinzufügen", "Team-Name:")
        if not ok or not name.strip():
            return
        name = name.strip()
        if name in self._current_teams():
            QMessageBox.warning(self, "Fehler", f"Team '{name}' existiert bereits.")
            return
        self._suppress_save = True
        col = self.table.columnCount()
        self.table.insertColumn(col)
        self.table.setHorizontalHeaderItem(col, QTableWidgetItem(name))
        self._suppress_save = False
        self._collect_and_save()

    def _remove_team(self):
        col = self.table.currentColumn()
        if col < _TEAM_COL_START:
            QMessageBox.warning(
                self, "Fehler",
                "Bitte eine Team-Spalte auswählen (nicht 'Bezeichnung' oder 'Dienst-ID')."
            )
            return
        team = self._current_teams()[col - _TEAM_COL_START]
        reply = QMessageBox.question(
            self, "Team entfernen",
            f"Team '{team}' und alle Zuordnungen entfernen?",
            QMessageBox.StandardButton.Yes | QMessageBox.StandardButton.No
        )
        if reply != QMessageBox.StandardButton.Yes:
            return
        self._suppress_save = True
        self.table.removeColumn(col)
        self._suppress_save = False
        self._collect_and_save()

    def _add_dienst(self):
        self._suppress_save = True
        row = self.table.rowCount()
        self.table.insertRow(row)
        self.table.setItem(row, _COL_LABEL, QTableWidgetItem(""))
        self.table.setItem(row, _COL_ID, QTableWidgetItem(""))
        self._suppress_save = False

    def _remove_dienst(self):
        row = self.table.currentRow()
        if row < 0:
            QMessageBox.warning(self, "Fehler", "Bitte eine Zeile auswählen.")
            return
        self._suppress_save = True
        self.table.removeRow(row)
        self._suppress_save = False
        self._collect_and_save()

    # --- Persons fetch ---

    def _fetch_persons(self):
        api_key = self.store.get_api_key()
        ct_url = self.store.get_ct_url()
        if not api_key:
            QMessageBox.warning(
                self, "Fehler",
                "Kein API-Schlüssel gesetzt. Bitte zuerst in den Einstellungen speichern."
            )
            return
        self.fetch_btn.setEnabled(False)
        self.persons_status.setText("Lade Personen...")
        self._fetch_worker = FetchPersonsWorker(ct_url, api_key)
        self._fetch_worker.finished.connect(self._on_persons_fetched)
        self._fetch_worker.error.connect(self._on_persons_error)
        self._fetch_worker.start()

    def _on_persons_fetched(self, persons: list):
        self.store.set_persons(persons)
        self.fetch_btn.setEnabled(True)
        self._update_persons_status()
        self._fetch_worker = None

    def _on_persons_error(self, msg: str):
        self.fetch_btn.setEnabled(True)
        self.persons_status.setText("Fehler beim Laden")
        QMessageBox.critical(self, "Fehler beim Laden der Personen", msg)
        self._fetch_worker = None
