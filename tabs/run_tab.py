from datetime import datetime

from PyQt6.QtWidgets import QWidget, QVBoxLayout, QPushButton, QTextEdit
from PyQt6.QtCore import Qt, QThread, pyqtSignal
from PyQt6.QtGui import QFont

import ct_api
from data_store import DataStore


class RunWorker(QThread):
    log = pyqtSignal(str)
    finished = pyqtSignal(int, int, list)  # updated, errors, skipped_dates

    def __init__(self, data_snapshot: dict):
        super().__init__()
        self._data = data_snapshot

    def run(self):
        data = self._data
        settings = data["settings"]
        api_key = settings["api_key"]
        ct_url = settings["ct_url"]
        einsaetze = data["einsaetze"]
        musikteam = data["musikteam"]
        teamleiter = data["teamleiter"]
        teams_list = musikteam["teams"]
        dienste = musikteam["dienste"]

        update_count = 0
        error_count = 0
        skipped_dates = []

        # Validate
        if not api_key:
            self.log.emit("FEHLER: Kein API-Schlüssel gesetzt (Tab: Einstellungen).")
            self.finished.emit(0, 1, [])
            return
        if not einsaetze:
            self.log.emit("FEHLER: Keine Einsätze definiert (Tab: Einsätze).")
            self.finished.emit(0, 1, [])
            return

        # Build team config: {team_name: {service_id_int: [person_id_int, ...]}}
        team_config = {team: {} for team in teams_list}
        for dienst in dienste:
            try:
                sid = int(dienst["id"])
            except (ValueError, KeyError):
                continue
            for team, person_str in dienst.get("persons", {}).items():
                if not person_str:
                    continue
                try:
                    pids = [int(p.strip()) for p in person_str.split(";") if p.strip()]
                    if team not in team_config:
                        team_config[team] = {}
                    team_config[team][sid] = pids
                except ValueError:
                    self.log.emit(
                        f"Warnung: Ungültige Person-ID in Dienst {dienst['id']}, Team {team}"
                    )

        # Compute date range
        try:
            from_date, to_date = ct_api.get_date_range(einsaetze)
        except Exception as e:
            self.log.emit(f"FEHLER beim Berechnen des Datumsbereichs: {e}")
            self.finished.emit(0, 1, [])
            return

        self.log.emit(f"Abrufen von Events: {from_date} bis {to_date}...")

        # Fetch events
        try:
            events = ct_api.fetch_events(ct_url, api_key, from_date, to_date)
            self.log.emit(f"Gefunden: {len(events)} Events")
        except Exception as e:
            self.log.emit(f"FEHLER beim Abrufen der Events: {e}")
            self.finished.emit(0, 1, [])
            return

        # Build assignment lookup: ISO date -> team name
        assignments = {}
        for e in einsaetze:
            try:
                dt = datetime.strptime(e["datum"], "%d.%m.%Y")
                assignments[dt.strftime("%Y-%m-%d")] = e["musikteam"]
            except ValueError:
                self.log.emit(f"Warnung: Ungültiges Datum '{e['datum']}' übersprungen.")

        # Group events by date
        events_by_date: dict = {}
        for event in events:
            start = event.get("startDate", "")
            event_date = start.split("T")[0] if start else None
            if event_date and event_date in assignments:
                events_by_date.setdefault(event_date, []).append(event)

        # Process each date
        for event_date, date_events in events_by_date.items():
            if len(date_events) > 1:
                names = [e.get("name", "Unbekannt") for e in date_events]
                date_display = datetime.strptime(event_date, "%Y-%m-%d").strftime("%d.%m.%Y")
                skip_msg = (
                    f"Es gibt mehrere Gottesdienste an diesem Sonntag {date_display}. "
                    f"Bitte frage hier die Dienste manuel an."
                )

                # 3 or more events: always skip
                if len(date_events) >= 3:
                    skipped_dates.append({"date": event_date, "count": len(date_events), "events": names})
                    self.log.emit(f"ÜBERSPRUNGEN: {skip_msg}")
                    continue

                # Exactly 2 events: check for a Jona event
                non_jona = [e for e in date_events if "Jona" not in e.get("name", "")]

                if len(non_jona) == 1:
                    # One Jona, one regular → process the regular one
                    jona_name = next(e.get("name", "") for e in date_events if "Jona" in e.get("name", ""))
                    self.log.emit(
                        f"Info: {date_display} — Jona-Gottesdienst ignoriert ({jona_name}), "
                        f"verarbeite: {non_jona[0].get('name', 'Unbekannt')}"
                    )
                    date_events = non_jona
                else:
                    # Two events, neither (or both) has Jona → skip
                    skipped_dates.append({"date": event_date, "count": len(date_events), "events": names})
                    self.log.emit(f"ÜBERSPRUNGEN: {skip_msg}")
                    continue

            event = date_events[0]
            team_name = assignments[event_date]
            event_id = event.get("id")
            event_name = event.get("name", "Unbekannt")

            if team_name not in team_config:
                self.log.emit(
                    f"Warnung: Musikteam '{team_name}' nicht konfiguriert — "
                    f"{event_date} übersprungen"
                )
                continue

            self.log.emit(f"\nVerarbeite: {event_name} ({event_date}) — Musikteam: {team_name}")

            # Update team leader fact (fact_id=3, same as original script)
            try:
                leader_text = teamleiter.get(team_name, team_name)
                ct_api.update_event_fact(ct_url, api_key, event_id, 3, leader_text)
                self.log.emit(f"  OK: Event-Fact gesetzt: {leader_text}")
            except Exception as e:
                self.log.emit(f"  FEHLER Event-Fact: {e}")

            # Update service requests
            config = team_config[team_name]
            service_instance_counter: dict = {}

            for service in event.get("eventServices", []):
                service_id = service.get("serviceId")
                request_id = service.get("id")
                service_name = service.get("name", "Unbekannt")

                if service_id not in config:
                    continue

                person_ids = config[service_id]
                idx = service_instance_counter.get(service_id, 0)

                if idx >= len(person_ids):
                    self.log.emit(
                        f"  ÜBERSPRUNGEN: {service_name} (Instanz {idx + 1}) — keine Person"
                    )
                    service_instance_counter[service_id] = idx + 1
                    continue

                person_id = person_ids[idx]
                service_instance_counter[service_id] = idx + 1

                try:
                    ct_api.update_service_request(
                        ct_url, api_key, event_id, request_id, person_id,
                        name=service_name, is_accepted=False, is_valid=True
                    )
                    inst = f" (Instanz {idx + 1})" if len(person_ids) > 1 else ""
                    self.log.emit(
                        f"  OK: {service_name}{inst} → Person {person_id}"
                    )
                    update_count += 1
                except Exception as e:
                    self.log.emit(f"  FEHLER: {service_name}: {e}")
                    error_count += 1

        self.finished.emit(update_count, error_count, skipped_dates)


class RunTab(QWidget):
    def __init__(self, store: DataStore):
        super().__init__()
        self.store = store
        self._worker = None
        self._build_ui()

    def _build_ui(self):
        layout = QVBoxLayout(self)

        self.run_btn = QPushButton("Jetzt ausführen")
        self.run_btn.setFixedHeight(48)
        self.run_btn.clicked.connect(self._run)
        layout.addWidget(self.run_btn)

        self.log_view = QTextEdit()
        self.log_view.setReadOnly(True)
        font = QFont("Courier New", 10)
        font.setStyleHint(QFont.StyleHint.Monospace)
        self.log_view.setFont(font)
        layout.addWidget(self.log_view)

    def _run(self):
        self.run_btn.setEnabled(False)
        self.log_view.clear()
        self.log_view.append("Starte Ausführung...\n")

        snapshot = self.store.snapshot()
        self._worker = RunWorker(snapshot)
        self._worker.log.connect(self._append_log)
        self._worker.finished.connect(self._on_finished)
        self._worker.start()

    def _append_log(self, message: str):
        self.log_view.append(message)

    def _on_finished(self, updated: int, errors: int, skipped: list):
        sep = "=" * 60
        self.log_view.append(f"\n{sep}")
        self.log_view.append("Zusammenfassung:")
        self.log_view.append(f"  Erfolgreich aktualisiert: {updated} Dienstanfragen")
        self.log_view.append(f"  Fehler: {errors}")
        if skipped:
            self.log_view.append("\n  ÜBERSPRUNGEN — Mehrere Events am selben Tag:")
            for s in skipped:
                names = ", ".join(s["events"])
                self.log_view.append(f"    {s['date']}: {names}")
            self.log_view.append(
                "\n  Diese Termine wurden NICHT aktualisiert, da mehrere"
            )
            self.log_view.append(
                "  Events am selben Tag gefunden wurden. Bitte manuell prüfen."
            )
        self.log_view.append(sep)
        self.run_btn.setEnabled(True)
        self._worker = None
