# CT Musikteam

Desktop-App zum Verwalten und Uebertragen von Musikteam-Einsaetzen in ChurchTools.

## Fuer Endnutzer

Nach dem ersten Build liegt im Ordner `release/` alles, was du brauchst:

```text
release/
├── CT Musikteam.app   ← Doppelklick zum Starten
└── Updater.command    ← Doppelklick zum Aktualisieren (bei neuem Release)
```

**Erstmalige Einrichtung:**

1. `./build.sh` ausfuehren (einmalig, braucht Go + Wails – siehe unten)
2. `CT Musikteam.app` aus dem `release/`-Ordner starten
3. In der App: Tab **Einstellungen** → ChurchTools-URL + Login-Token eintragen → Speichern

**Update auf neue Version:**

1. Neuen Release-Ordner herunterladen
2. `release/Updater.command` doppelklicken
3. Alte Daten werden geloescht, die App wird neu gebaut
4. Danach: `release/CT Musikteam.app` starten

---

## Was macht die App?

- **Musikteam-Tabelle**: Teams (Spalten) und Dienste/Personen (Zeilen) direkt in der App pflegen
- **Einsaetze-Tabelle**: Zuordnung Datum → Musikteam verwalten
- **Teamleiter-Texte**: Pro Team den Anzeigetext pflegen
- **Ausfuehren**: Uebertraegt alle Zuordnungen automatisch per ChurchTools-API

Alle Daten werden lokal gespeichert (`~/.ct_musikteam/data.json`).
Der API-Schluessel wird sicher im Betriebssystem-Schlusselbund (macOS Keychain / Windows Credential Manager) abgelegt.

## Voraussetzungen (fuer Entwickler / Build)

- [Go 1.22+](https://go.dev/dl/)
- [Wails v2](https://wails.io) (wird beim ersten Build automatisch installiert)
- macOS oder Windows

```bash
./build.sh
```

Beim ersten Aufruf laedt `build.sh` automatisch alle Abhaengigkeiten und installiert
die Wails CLI. Anschliessend liegt die fertige App unter `release/CT Musikteam.app`.

## Einmalige App-Einrichtung

1. Tab **Einstellungen**: ChurchTools-URL und Login-Token eintragen → Speichern
2. Tab **Musikteam**: Dienste aus CT laden, dann Personen pro Team zuordnen
3. Tab **Einsaetze**: Sonntage und zugehoerige Musikteams eintragen
4. Tab **Teamleiter**: Anzeigetexte pro Team pruefen/anpassen
5. Tab **Ausfuehren**: „Jetzt ausfuehren" → Ergebnis im Log

## Verhalten bei mehreren Gottesdiensten an einem Sonntag

| Anzahl Events | Bedingung | Ergebnis |
|---|---|---|
| 2 | Eines enthaelt „Jona" im Titel | Nur den normalen Gottesdienst bearbeiten |
| 2 | Kein „Jona" | Ueberspringen + Hinweis |
| 3+ | – | Immer ueberspringen + Hinweis |

## Projektstruktur (fuer Entwickler)

```text
ct_band_updater/
├── release/             ← Endnutzer-Ordner (nach Build)
│   ├── CT Musikteam.app    (gitignored, wird per build.sh erstellt)
│   └── Updater.command     (doppelklick zum Aktualisieren)
├── frontend/
│   ├── index.html       # Alle 5 Tabs + Modals
│   ├── style.css        # Styling
│   └── app.js           # Frontend-Logik
├── main.go              # Wails-Einstiegspunkt
├── app.go               # Backend-Methoden (exposed to frontend)
├── ctapi.go             # ChurchTools REST API + Run-Logik
├── datastore.go         # JSON-Persistenz + Keychain-Integration
├── buildinfo.go         # Build-ID (wird beim Build per ldflags gesetzt)
├── wails.json           # Wails-Konfiguration
├── go.mod / go.sum      # Go-Abhaengigkeiten
└── build.sh             # Build-Skript
```
