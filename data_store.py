"""
Persistence layer for CT Musikteam app.
All data is stored in ~/.ct_musikteam/data.json.
The API key is stored in the OS keychain (macOS Keychain / Windows Credential Manager).
"""

import json
import copy
from pathlib import Path

import keyring

try:
    from build_info import BUILD_ID
except ImportError:
    BUILD_ID = "dev"

DATA_DIR = Path.home() / ".ct_musikteam"
DATA_FILE = DATA_DIR / "data.json"

_KEYRING_SERVICE = "ct-musikteam"
_KEYRING_USERNAME = "api_key"

DEFAULT_DATA = {
    "settings": {
        "ct_url": "https://waldkirche.church.tools"
    },
    "musikteam": {
        "teams": ["A1", "A2", "B1", "B2", "C", "D", "E"],
        "dienste": [
            {"id": "88", "label": "Leiter",      "persons": {}},
            {"id": "11", "label": "Piano",       "persons": {}},
            {"id": "10", "label": "Gitarre",     "persons": {}},
            {"id": "53", "label": "Bass",        "persons": {}},
            {"id": "14", "label": "Drums/Cajon", "persons": {}},
            {"id": "9",  "label": "Sänger",      "persons": {}},
            {"id": "6",  "label": "Mischer",     "persons": {}},
        ]
    },
    "einsaetze": [],
    "persons": [],
    "teamleiter": {
        "A1": "A1 unter Leitung von Stefan",
        "A2": "A2 unter Leitung von Stefan",
        "B1": "B1 unter Leitung von Eugen",
        "B2": "B2 unter Leitung von Torben",
        "C": "C unter Leitung von Ralf Kachel",
        "D": "D unter Leitung von Tim",
        "E": "E unter Leitung von Susanne"
    }
}


class DataStore:
    def __init__(self):
        self._data = copy.deepcopy(DEFAULT_DATA)
        self.load()

    # --- I/O ---

    def load(self) -> None:
        if DATA_FILE.exists():
            try:
                with open(DATA_FILE, "r", encoding="utf-8") as f:
                    loaded = json.load(f)
                stored_build_id = loaded.get("build_id")
                if stored_build_id != BUILD_ID:
                    # New build detected — reset data, preserve settings
                    saved_url = loaded.get("settings", {}).get("ct_url", "")
                    self._data = copy.deepcopy(DEFAULT_DATA)
                    if saved_url:
                        self._data["settings"]["ct_url"] = saved_url
                    self.save()
                else:
                    self._deep_merge(self._data, loaded)
            except (json.JSONDecodeError, OSError):
                pass
        self._data["build_id"] = BUILD_ID
        # Migrate api_key from old plaintext JSON to OS keychain
        old_key = self._data.get("settings", {}).pop("api_key", "")
        if old_key:
            try:
                keyring.set_password(_KEYRING_SERVICE, _KEYRING_USERNAME, old_key)
                self.save()  # rewrite JSON without the api_key field
            except Exception:
                pass

    def save(self) -> None:
        DATA_DIR.mkdir(parents=True, exist_ok=True)
        self._data["build_id"] = BUILD_ID
        with open(DATA_FILE, "w", encoding="utf-8") as f:
            json.dump(self._data, f, ensure_ascii=False, indent=2)

    def _deep_merge(self, base: dict, override: dict) -> None:
        for k, v in override.items():
            if k in base and isinstance(base[k], dict) and isinstance(v, dict):
                self._deep_merge(base[k], v)
            else:
                base[k] = v

    # --- Settings ---

    def get_api_key(self) -> str:
        try:
            return keyring.get_password(_KEYRING_SERVICE, _KEYRING_USERNAME) or ""
        except Exception:
            return ""

    def set_api_key(self, key: str) -> None:
        try:
            keyring.set_password(_KEYRING_SERVICE, _KEYRING_USERNAME, key)
        except Exception:
            pass

    def get_ct_url(self) -> str:
        return self._data["settings"]["ct_url"]

    def set_ct_url(self, url: str) -> None:
        self._data["settings"]["ct_url"] = url.rstrip("/")
        self.save()

    # --- Musikteam ---

    def get_teams(self) -> list:
        return list(self._data["musikteam"]["teams"])

    def set_teams(self, teams: list) -> None:
        self._data["musikteam"]["teams"] = teams
        self.save()

    def get_dienste(self) -> list:
        """Returns list of {"id": str, "label": str, "persons": {team: person_str}}"""
        return copy.deepcopy(self._data["musikteam"]["dienste"])

    def set_dienste(self, dienste: list) -> None:
        self._data["musikteam"]["dienste"] = dienste
        self.save()

    # --- Einsätze ---

    def get_einsaetze(self) -> list:
        """Returns list of {"datum": "DD.MM.YYYY", "musikteam": str}"""
        return copy.deepcopy(self._data["einsaetze"])

    def set_einsaetze(self, einsaetze: list) -> None:
        self._data["einsaetze"] = einsaetze
        self.save()

    # --- Teamleiter ---

    def get_teamleiter(self) -> dict:
        return copy.deepcopy(self._data["teamleiter"])

    def set_teamleiter(self, teamleiter: dict) -> None:
        self._data["teamleiter"] = teamleiter
        self.save()

    # --- Persons cache ---

    def get_persons(self) -> list:
        """Returns cached list of {"id": int, "name": str}, sorted by name."""
        return copy.deepcopy(self._data.get("persons", []))

    def set_persons(self, persons: list) -> None:
        self._data["persons"] = persons
        self.save()

    # --- Thread-safe snapshot for RunWorker ---

    def snapshot(self) -> dict:
        snap = copy.deepcopy(self._data)
        snap["settings"]["api_key"] = self.get_api_key()
        return snap
