"""
ChurchTools API calls for CT Musikteam app.
Refactored from ct_req_teams.py — ct_url is now a parameter, not hardcoded.
"""

import requests
from datetime import datetime
from typing import Dict, List, Tuple


def _headers(api_key: str) -> dict:
    return {
        "Authorization": f"Login {api_key}",
        "Content-Type": "application/json"
    }


def fetch_events(ct_url: str, api_key: str,
                 from_date: str, to_date: str) -> List[Dict]:
    """GET /api/events with eventServices included."""
    url = f"{ct_url}/api/events"
    params = {
        "from": from_date,
        "to": to_date,
        "direction": "forward",
        "limit": 100,
        "include": "eventServices"
    }
    r = requests.get(url, headers=_headers(api_key), params=params, timeout=30)
    r.raise_for_status()
    return r.json().get("data", [])


def update_service_request(ct_url: str, api_key: str,
                           event_id: int, request_id: int,
                           person_id: int, name: str = None,
                           is_accepted: bool = False,
                           is_valid: bool = True) -> None:
    """PUT /api/events/{event_id}/servicerequests/{request_id}"""
    url = f"{ct_url}/api/events/{event_id}/servicerequests/{request_id}"
    body = {
        "personId": person_id,
        "isAccepted": is_accepted,
        "isValid": is_valid
    }
    if name:
        body["name"] = name
    r = requests.put(url, headers=_headers(api_key), json=body, timeout=30)
    r.raise_for_status()


def update_event_fact(ct_url: str, api_key: str,
                      event_id: int, fact_id: int, value: str) -> None:
    """PUT /api/events/{event_id}/facts/{fact_id}"""
    url = f"{ct_url}/api/events/{event_id}/facts/{fact_id}"
    r = requests.put(url, headers=_headers(api_key),
                     json={"value": value}, timeout=30)
    r.raise_for_status()


def fetch_persons(ct_url: str, api_key: str) -> List[Dict]:
    """
    Fetch all persons from /api/persons (all pages).
    Returns sorted list of {"id": int, "name": str}.
    """
    persons = []
    page = 1
    while True:
        url = f"{ct_url}/api/persons"
        params = {"limit": 100, "page": page}
        r = requests.get(url, headers=_headers(api_key), params=params, timeout=30)
        r.raise_for_status()
        data = r.json()
        for p in data.get("data", []):
            first = p.get("firstName", "") or ""
            last = p.get("lastName", "") or ""
            name = f"{first} {last}".strip()
            if name and p.get("id"):
                persons.append({"id": p["id"], "name": name})
        pagination = data.get("meta", {}).get("pagination", {})
        if page >= pagination.get("lastPage", 1):
            break
        page += 1
    return sorted(persons, key=lambda p: p["name"].lower())


def get_date_range(einsaetze: List[Dict]) -> Tuple[str, str]:
    """Compute ISO date range from a list of {"datum": "DD.MM.YYYY"} dicts."""
    dates = [datetime.strptime(e["datum"], "%d.%m.%Y") for e in einsaetze]
    return (
        min(dates).strftime("%Y-%m-%d"),
        max(dates).strftime("%Y-%m-%d")
    )
