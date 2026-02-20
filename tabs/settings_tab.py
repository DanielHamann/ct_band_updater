from PyQt6.QtWidgets import (
    QWidget, QFormLayout, QLineEdit, QPushButton,
    QHBoxLayout, QVBoxLayout, QLabel
)
from data_store import DataStore


class SettingsTab(QWidget):
    def __init__(self, store: DataStore):
        super().__init__()
        self.store = store
        self._build_ui()

    def _build_ui(self):
        outer = QVBoxLayout(self)
        outer.setContentsMargins(20, 20, 20, 20)

        form = QFormLayout()
        form.setSpacing(12)

        # API Key row
        api_row = QHBoxLayout()
        self.api_key_edit = QLineEdit(self.store.get_api_key())
        self.api_key_edit.setEchoMode(QLineEdit.EchoMode.Password)
        self.api_key_edit.setPlaceholderText("API-Schlüssel eingeben...")
        self.api_key_edit.setMinimumWidth(360)
        self.toggle_btn = QPushButton("Anzeigen")
        self.toggle_btn.setCheckable(True)
        self.toggle_btn.setFixedWidth(90)
        self.toggle_btn.toggled.connect(self._toggle_key_visibility)
        api_row.addWidget(self.api_key_edit)
        api_row.addWidget(self.toggle_btn)
        form.addRow("API-Schlüssel:", api_row)

        # URL row
        self.url_edit = QLineEdit(self.store.get_ct_url())
        self.url_edit.setMinimumWidth(360)
        form.addRow("ChurchTools-URL:", self.url_edit)

        outer.addLayout(form)

        # Save button
        save_btn = QPushButton("Speichern")
        save_btn.setFixedWidth(120)
        save_btn.clicked.connect(self._save)
        outer.addWidget(save_btn)

        outer.addStretch()

    def _toggle_key_visibility(self, checked: bool):
        mode = QLineEdit.EchoMode.Normal if checked else QLineEdit.EchoMode.Password
        self.api_key_edit.setEchoMode(mode)
        self.toggle_btn.setText("Verbergen" if checked else "Anzeigen")

    def _save(self):
        self.store.set_api_key(self.api_key_edit.text().strip())
        self.store.set_ct_url(self.url_edit.text().strip())
