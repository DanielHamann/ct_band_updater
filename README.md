# CT Musikteam

Desktop-App zum Verwalten und Übertragen von Musikteam-Einsätzen in ChurchTools.

## Was macht die App?

- **Musikteam-Tabelle**: Teams (Spalten) und Dienste/Personen (Zeilen) direkt in der App pflegen – kein Kopieren mehr nötig
- **Einsätze-Tabelle**: Zuordnung Datum → Musikteam verwalten
- **Teamleiter-Texte**: Pro Team den Anzeigetext pflegen
- **Ausführen**: Überträgt alle Zuordnungen automatisch per ChurchTools-API

Alle Daten werden lokal gespeichert (`~/.ct_musikteam/data.json`). Der API-Schlüssel wird sicher im Betriebssystem-Schlüsselbund (macOS Keychain / Windows Credential Manager) abgelegt.

## Voraussetzungen

- Python 3.11 oder neuer
- macOS oder Windows

## Installation

```bash
# Abhängigkeiten installieren
python3 -m venv .venv
source .venv/bin/activate        # macOS
# .venv\Scripts\activate         # Windows

pip install -r requirements.txt
```

## Starten

```bash
python main.py
```

## Einmalige Einrichtung

1. Tab **Einstellungen**: API-Schlüssel und ChurchTools-URL eintragen → Speichern
2. Tab **Musikteam**: Teams als Spalten anlegen, Dienst-IDs als Zeilen, Personen-IDs in die Zellen
3. Tab **Einsätze**: Sonntage und zugehörige Musikteams eintragen
4. Tab **Teamleiter**: Anzeigetexte pro Team prüfen/anpassen
5. Tab **Ausführen**: „Jetzt ausführen" → Ergebnis im Log

## Verhalten bei mehreren Gottesdiensten an einem Sonntag

| Anzahl Events | Bedingung | Ergebnis |
|---|---|---|
| 2 | Eines enthält „Jona" im Titel | Nur den normalen Gottesdienst bearbeiten |
| 2 | Kein „Jona" | Überspringen + Hinweis |
| 3+ | – | Immer überspringen + Hinweis |

## Als Standalone-Executable bauen

```bash
pip install pyinstaller
pyinstaller --onefile --windowed --name "CT Musikteam" main.py
```

- **macOS**: `dist/CT Musikteam.app`
- **Windows**: `dist/CT Musikteam.exe`

> Falls Qt-Plugins fehlen: `--collect-all PyQt6` hinzufügen

## Projektstruktur

```
ct_app/
├── main.py               # Einstiegspunkt, Hauptfenster
├── data_store.py         # JSON-Persistenz + Keychain-Integration
├── ct_api.py             # ChurchTools REST API
├── requirements.txt
└── tabs/
    ├── settings_tab.py   # Einstellungen (API-Key, URL)
    ├── musikteam_tab.py  # Musikteam-Tabelle
    ├── einsaetze_tab.py  # Einsatz-Zuordnungen
    ├── teamleiter_tab.py # Teamleiter-Texte
    └── run_tab.py        # Ausführen + Log
```
