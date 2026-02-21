<div align="center">
  <img src="build/appicon.png" width="120" alt="CT Musikteam Icon" />
  <h1>CT Musikteam</h1>
  <p>
    <strong>DE:</strong> Desktop-App zum automatischen Übertragen von Musikteam-Einsätzen in ChurchTools.<br>
    <strong>EN:</strong> Desktop app for automatically assigning music team duties in ChurchTools.
  </p>
  <a href="https://github.com/DanielHamann/ct_band_updater/releases/latest">
    <img src="https://img.shields.io/github/v/release/DanielHamann/ct_band_updater?style=flat-square" alt="Latest Release">
  </a>
  <a href="LICENSE">
    <img src="https://img.shields.io/github/license/DanielHamann/ct_band_updater?style=flat-square" alt="MIT License">
  </a>
  <a href="https://github.com/DanielHamann/ct_band_updater/actions">
    <img src="https://img.shields.io/github/actions/workflow/status/DanielHamann/ct_band_updater/release.yml?style=flat-square" alt="Build Status">
  </a>
</div>

---

## Download

| Plattform / Platform | Datei / File |
|---|---|
| 🍎 macOS (Intel + Apple Silicon) | `CT-Musikteam-vX.X-macOS.dmg` |
| 🪟 Windows 10/11 (64-bit) | `CT-Musikteam-vX.X-Windows.zip` |

→ **[Neueste Version / Latest Release](https://github.com/DanielHamann/ct_band_updater/releases/latest)**

---

## Was macht die App? / What does it do?

**DE:** CT Musikteam liest deine Musikteam-Konfiguration — welche Person welchen Dienst übernimmt — und überträgt diese automatisch per ChurchTools-API auf die jeweiligen Events. Kein manuelles Klicken mehr.

**EN:** CT Musikteam reads your music team configuration — who covers which service role — and automatically assigns them via the ChurchTools API to the respective events. No more manual clicking.

---

## Features

| | DE | EN |
|---|---|---|
| 🎵 | Dienste und Personen pro Team verwalten | Manage service roles and persons per team |
| 📅 | Datum ↔ Musikteam Zuordnungen pflegen | Maintain date ↔ team assignments |
| 🔄 | Alle Zuordnungen automatisch per API übertragen | Transfer all assignments automatically via API |
| 📝 | Teamleiter-Fakten pro Team setzen | Set team leader facts per team |
| 🔍 | Events aus ChurchTools laden und zuordnen | Load and match events from ChurchTools |
| 🔒 | API-Schlüssel im OS-Schlüsselbund | API key stored in OS keychain |
| 🌍 | Eventzeiten in lokale Zeitzone umgerechnet | Event times converted to local timezone |
| 🖥️ | macOS Universal + Windows 64-bit | macOS Universal + Windows 64-bit |

---

## Schnellstart / Quick Start

### 1. Herunterladen / Download

Lade die neueste Version von der [Releases-Seite](https://github.com/DanielHamann/ct_band_updater/releases/latest) herunter.
Download the latest version from the [Releases page](https://github.com/DanielHamann/ct_band_updater/releases/latest).

- **macOS:** DMG öffnen → App in den Programme-Ordner ziehen → starten
- **Windows:** ZIP entpacken → `CT Musikteam.exe` starten

### 2. Einrichten / Setup

1. ⚙️ Zahnrad-Button öffnen → ChurchTools-URL und Login-Token eintragen → **Speichern**
2. 🎼 **Musikteam** → Dienste aus CT laden, Personen pro Team zuordnen
3. 📅 **Einsätze** → Termine und Musikteams eintragen
4. ▶️ **Ausführen** → „Jetzt ausführen"

> **DE:** Der Login-Token wird sicher im OS-Schlüsselbund gespeichert (macOS Keychain / Windows Credential Manager) — nicht in der App selbst.
>
> **EN:** The login token is stored securely in the OS keychain — not inside the app.

---

## Voraussetzungen / Requirements

| Plattform | Anforderung / Requirement |
|---|---|
| macOS | 10.13 High Sierra oder neuer / or newer |
| Windows | Windows 10 / 11 (64-bit), Edge WebView2 (pre-installed) |
| ChurchTools | Instanz mit API-Zugang / instance with API access |

---

## Für Entwickler / For Developers

### Voraussetzungen / Prerequisites

- [Go 1.22+](https://go.dev/dl/)
- [Wails v2](https://wails.io)
- macOS oder Windows

```bash
# Wails installieren / Install Wails (einmalig / once)
go install github.com/wailsapp/wails/v2/cmd/wails@latest

# Lokaler Dev-Server / Local dev server
wails dev

# Build (macOS, erstellt DMG / creates DMG)
./build.sh

# Build (plattformübergreifend / cross-platform)
wails build
```

### Projektstruktur / Project Structure

```text
ct_band_updater/
├── .github/
│   └── workflows/
│       └── release.yml   # CI/CD: macOS + Windows build on tag push
├── docs/
│   └── index.html        # GitHub Pages landing page
├── frontend/
│   ├── index.html        # UI (Tabs + Modals)
│   ├── style.css         # Glassmorphism design
│   └── app.js            # Frontend logic
├── main.go               # Wails entry point
├── app.go                # Backend methods (exposed to frontend)
├── ctapi.go              # ChurchTools REST API + assignment logic
├── datastore.go          # JSON persistence + keychain integration
├── buildinfo.go          # Build ID (injected via ldflags)
├── wails.json            # Wails configuration
└── build.sh              # macOS build + DMG script
```

### Release erstellen / Creating a Release

```bash
git tag v1.0
git push origin v1.0
```

GitHub Actions baut automatisch macOS (Universal DMG) + Windows (ZIP) und veröffentlicht das Release.
GitHub Actions automatically builds macOS (Universal DMG) + Windows (ZIP) and publishes the release.

---

## Lizenz / License

MIT © [Daniel Hamann](https://github.com/DanielHamann) — siehe [LICENSE](LICENSE) / see [LICENSE](LICENSE) for details.
