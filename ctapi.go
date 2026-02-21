package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"
)

// ── HTTP client ─────────────────────────────────────────────────────────────

var httpClient = &http.Client{Timeout: 30 * time.Second}

func doRequest(method, rawURL, apiKey, body string) ([]byte, error) {
	var bodyReader io.Reader
	if body != "" {
		bodyReader = strings.NewReader(body)
	}
	req, err := http.NewRequest(method, rawURL, bodyReader)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Login "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(data))
	}
	return data, nil
}

// decodeData unmarshals the common {"data": T} API envelope.
func decodeData[T any](raw []byte) (T, error) {
	var env struct {
		Data T `json:"data"`
	}
	return env.Data, json.Unmarshal(raw, &env)
}

// ── Types ───────────────────────────────────────────────────────────────────

type CTEvent struct {
	ID            int              `json:"id"`
	Name          string           `json:"name"`
	StartDate     string           `json:"startDate"`
	EventServices []CTEventService `json:"eventServices"`
}

type CTEventService struct {
	ID        int    `json:"id"`
	ServiceID int    `json:"serviceId"`
	Name      string `json:"name"`
}

type EventFactDisplay struct {
	FactID int    `json:"factId"`
	Value  string `json:"value"`
}

type CTFactDefinition struct {
	ID             int      `json:"id"`
	Name           string   `json:"name"`
	NameTranslated string   `json:"nameTranslated"`
	SortKey        int      `json:"sortKey"`
	Type           string   `json:"type"`
	Options        []string `json:"options"`
}

type CTServiceGroup struct {
	ID      int    `json:"id"`
	Name    string `json:"name"`
	SortKey int    `json:"sortKey"`
}

type CTService struct {
	ID             int    `json:"id"`
	Name           string `json:"name"`
	ServiceGroupID int    `json:"serviceGroupId"`
	SortKey        int    `json:"sortKey"`
}

type CTFact struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
}

type CTMasterData struct {
	ServiceGroups []CTServiceGroup `json:"serviceGroups"`
	Services      []CTService      `json:"services"`
	Facts         []CTFact         `json:"facts"`
}

// ── Event facts ─────────────────────────────────────────────────────────────

func fetchEventFacts(ctURL, apiKey string, eventID int) ([]EventFactDisplay, error) {
	data, err := doRequest("GET", fmt.Sprintf("%s/api/events/%d/facts", ctURL, eventID), apiKey, "")
	if err != nil {
		return nil, err
	}

	var result struct {
		Data []struct {
			FactID int             `json:"factId"`
			Value  json.RawMessage `json:"value"`
		} `json:"data"`
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}

	out := make([]EventFactDisplay, 0, len(result.Data))
	for _, f := range result.Data {
		var strVal string
		if err := json.Unmarshal(f.Value, &strVal); err != nil {
			strVal = strings.Trim(string(f.Value), "\"")
		}
		out = append(out, EventFactDisplay{FactID: f.FactID, Value: strVal})
	}
	return out, nil
}

// ── Events ──────────────────────────────────────────────────────────────────

func fetchEvents(ctURL, apiKey, from, to string) ([]CTEvent, error) {
	q := url.Values{
		"from":      {from},
		"to":        {to},
		"direction": {"forward"},
		"limit":     {"100"},
		"include":   {"eventServices"},
	}
	data, err := doRequest("GET", ctURL+"/api/events?"+q.Encode(), apiKey, "")
	if err != nil {
		return nil, err
	}
	return decodeData[[]CTEvent](data)
}

func fetchEvent(ctURL, apiKey string, eventID int) (*CTEvent, error) {
	data, err := doRequest("GET", fmt.Sprintf("%s/api/events/%d?include=eventServices", ctURL, eventID), apiKey, "")
	if err != nil {
		return nil, err
	}
	ev, err := decodeData[CTEvent](data)
	return &ev, err
}

func addServiceSlotsToEvent(ctURL, apiKey string, eventID int, additions map[int]int) error {
	type svcReq struct {
		Count     int `json:"count"`
		ServiceID int `json:"serviceId"`
	}
	services := make([]svcReq, 0, len(additions))
	for sid, count := range additions {
		services = append(services, svcReq{Count: count, ServiceID: sid})
	}
	body, _ := json.Marshal(map[string]any{"eventId": eventID, "services": services})
	_, err := doRequest("PUT", fmt.Sprintf("%s/api/events/%d/servicerequests", ctURL, eventID), apiKey, string(body))
	return err
}

func updateServiceRequest(ctURL, apiKey string, eventID, requestID, personID int) error {
	body, _ := json.Marshal(map[string]any{"personId": personID, "isAccepted": false, "isValid": true})
	_, err := doRequest("PUT", fmt.Sprintf("%s/api/events/%d/servicerequests/%d", ctURL, eventID, requestID), apiKey, string(body))
	return err
}

func updateEventFact(ctURL, apiKey string, eventID, factID int, value string) error {
	body, _ := json.Marshal(map[string]string{"value": value})
	_, err := doRequest("PUT", fmt.Sprintf("%s/api/events/%d/facts/%d", ctURL, eventID, factID), apiKey, string(body))
	return err
}

// ── Facts ───────────────────────────────────────────────────────────────────

func fetchFacts(ctURL, apiKey string) ([]CTFactDefinition, error) {
	data, err := doRequest("GET", ctURL+"/api/facts", apiKey, "")
	if err != nil {
		return nil, err
	}
	facts, err := decodeData[[]CTFactDefinition](data)
	if err != nil {
		return nil, err
	}
	sort.Slice(facts, func(i, j int) bool { return facts[i].SortKey < facts[j].SortKey })
	return facts, nil
}

// ── Masterdata ──────────────────────────────────────────────────────────────

func fetchMasterData(ctURL, apiKey string) (*CTMasterData, error) {
	data, err := doRequest("GET", ctURL+"/api/event/masterdata", apiKey, "")
	if err != nil {
		return nil, err
	}
	md, err := decodeData[CTMasterData](data)
	if err != nil {
		return nil, err
	}
	sort.Slice(md.ServiceGroups, func(i, j int) bool { return md.ServiceGroups[i].SortKey < md.ServiceGroups[j].SortKey })
	sort.Slice(md.Services, func(i, j int) bool { return md.Services[i].SortKey < md.Services[j].SortKey })
	return &md, nil
}

// ── Persons ─────────────────────────────────────────────────────────────────

func FetchPersonsFromAPI(ctURL, apiKey string) ([]Person, error) {
	var persons []Person
	for page := 1; ; page++ {
		q := url.Values{"limit": {"100"}, "page": {fmt.Sprint(page)}}
		data, err := doRequest("GET", ctURL+"/api/persons?"+q.Encode(), apiKey, "")
		if err != nil {
			return nil, err
		}

		var result struct {
			Data []struct {
				ID        int    `json:"id"`
				FirstName string `json:"firstName"`
				LastName  string `json:"lastName"`
			} `json:"data"`
			Meta struct {
				Pagination struct{ LastPage int `json:"lastPage"` } `json:"pagination"`
			} `json:"meta"`
		}
		if err := json.Unmarshal(data, &result); err != nil {
			return nil, err
		}

		for _, p := range result.Data {
			name := strings.TrimSpace(p.FirstName + " " + p.LastName)
			if name != "" && p.ID != 0 {
				persons = append(persons, Person{ID: p.ID, Name: name})
			}
		}
		if page >= result.Meta.Pagination.LastPage {
			break
		}
	}

	sort.Slice(persons, func(i, j int) bool {
		return strings.ToLower(persons[i].Name) < strings.ToLower(persons[j].Name)
	})
	return persons, nil
}

// ── Date helpers ─────────────────────────────────────────────────────────────

func getDateRange(einsaetze []Einsatz) (string, string, error) {
	var min, max time.Time
	first := true
	for _, e := range einsaetze {
		t, err := time.Parse("02.01.2006", e.Datum)
		if err != nil {
			continue
		}
		if first || t.Before(min) {
			min = t
		}
		if first || t.After(max) {
			max = t
		}
		first = false
	}
	if first {
		return "", "", fmt.Errorf("keine gültigen Daten")
	}
	return min.Format("2006-01-02"), max.Format("2006-01-02"), nil
}

// ── Run update ───────────────────────────────────────────────────────────────

type RunResult struct {
	Updated int `json:"updated"`
	Errors  int `json:"errors"`
}

func runUpdateProcess(
	ctURL, apiKey string,
	einsaetze []Einsatz,
	dienste []Dienst,
	teams []string,
	teamleiter map[string]TeamleiterEntry,
	logFn func(string),
) RunResult {
	var result RunResult

	if apiKey == "" {
		logFn("FEHLER: Kein API-Schlüssel gesetzt (Einstellungen).")
		result.Errors++
		return result
	}
	if len(einsaetze) == 0 {
		logFn("FEHLER: Keine Einsätze definiert.")
		result.Errors++
		return result
	}

	// Build: team -> serviceID -> []personID
	teamConfig := make(map[string]map[int][]int)
	for _, team := range teams {
		teamConfig[team] = make(map[int][]int)
	}
	for _, dienst := range dienste {
		var sid int
		fmt.Sscanf(dienst.ID, "%d", &sid)
		if sid == 0 {
			continue
		}
		for team, personStr := range dienst.Persons {
			if personStr == "" {
				continue
			}
			if _, ok := teamConfig[team]; !ok {
				teamConfig[team] = make(map[int][]int)
			}
			for _, p := range strings.Split(personStr, ";") {
				var pid int
				if _, err := fmt.Sscanf(strings.TrimSpace(p), "%d", &pid); err == nil && pid > 0 {
					teamConfig[team][sid] = append(teamConfig[team][sid], pid)
				}
			}
		}
	}

	fromDate, toDate, err := getDateRange(einsaetze)
	if err != nil {
		logFn("FEHLER: " + err.Error())
		result.Errors++
		return result
	}

	logFn(fmt.Sprintf("Abrufe Events: %s bis %s...", fromDate, toDate))
	events, err := fetchEvents(ctURL, apiKey, fromDate, toDate)
	if err != nil {
		logFn("FEHLER beim Abrufen der Events: " + err.Error())
		result.Errors++
		return result
	}
	logFn(fmt.Sprintf("Gefunden: %d Events", len(events)))

	eventsById := make(map[int]CTEvent, len(events))
	eventsByDate := make(map[string][]CTEvent)
	for _, ev := range events {
		eventsById[ev.ID] = ev
		if len(ev.StartDate) >= 10 {
			date := ev.StartDate[:10]
			eventsByDate[date] = append(eventsByDate[date], ev)
		}
	}

	processEvent := func(ev CTEvent, teamName, displayDate string) {
		cfg, ok := teamConfig[teamName]
		if !ok {
			logFn(fmt.Sprintf("WARNUNG: Team '%s' nicht konfiguriert – %s übersprungen.", teamName, displayDate))
			return
		}

		logFn(fmt.Sprintf("\nVerarbeite: %s (%s) – Musikteam: %s", ev.Name, displayDate, teamName))

		if tlEntry := teamleiter[teamName]; tlEntry.FactID > 0 {
			value := tlEntry.Value
			if value == "" {
				value = teamName
			}
			if err := updateEventFact(ctURL, apiKey, ev.ID, tlEntry.FactID, value); err != nil {
				logFn("  FEHLER Event-Fakt: " + err.Error())
			} else {
				logFn("  OK: Event-Fakt gesetzt: " + value)
			}
		}

		// Detect and create missing service slots
		existingCounts := make(map[int]int)
		for _, svc := range ev.EventServices {
			existingCounts[svc.ServiceID]++
		}
		additions := make(map[int]int)
		totalMissing := 0
		for sid, pids := range cfg {
			if missing := len(pids) - existingCounts[sid]; missing > 0 {
				additions[sid] = missing
				totalMissing += missing
			}
		}
		if len(additions) > 0 {
			logFn(fmt.Sprintf("  Erstelle %d fehlende Dienst-Slot(s)...", totalMissing))
			if err := addServiceSlotsToEvent(ctURL, apiKey, ev.ID, additions); err != nil {
				logFn("  FEHLER: Dienst-Slots erstellen: " + err.Error())
				result.Errors++
			} else if updated, err := fetchEvent(ctURL, apiKey, ev.ID); err != nil {
				logFn("  FEHLER: Event neu laden: " + err.Error())
				result.Errors++
			} else {
				ev = *updated
				logFn("  OK: Dienst-Slots erstellt.")
			}
		}

		instanceCounter := make(map[int]int)
		for _, svc := range ev.EventServices {
			pids, ok := cfg[svc.ServiceID]
			if !ok {
				continue
			}
			idx := instanceCounter[svc.ServiceID]
			instanceCounter[svc.ServiceID]++
			if idx >= len(pids) {
				logFn(fmt.Sprintf("  ÜBERSPRUNGEN: %s (Instanz %d) – keine Person", svc.Name, idx+1))
				continue
			}
			if err := updateServiceRequest(ctURL, apiKey, ev.ID, svc.ID, pids[idx]); err != nil {
				logFn(fmt.Sprintf("  FEHLER: %s: %s", svc.Name, err.Error()))
				result.Errors++
			} else {
				suffix := ""
				if len(pids) > 1 {
					suffix = fmt.Sprintf(" (Instanz %d)", idx+1)
				}
				logFn(fmt.Sprintf("  OK: %s%s → Person %d", svc.Name, suffix, pids[idx]))
				result.Updated++
			}
		}
	}

	for _, einsatz := range einsaetze {
		if einsatz.EventID != 0 {
			ev, ok := eventsById[einsatz.EventID]
			if !ok {
				logFn(fmt.Sprintf("WARNUNG: Event-ID %d nicht im abgerufenen Zeitraum – übersprungen.", einsatz.EventID))
				continue
			}
			processEvent(ev, einsatz.Musikteam, einsatz.Datum)
		} else {
			t, err := time.Parse("02.01.2006", einsatz.Datum)
			if err != nil {
				logFn(fmt.Sprintf("WARNUNG: Ungültiges Datum '%s' – übersprungen.", einsatz.Datum))
				continue
			}
			dateEvs := eventsByDate[t.Format("2006-01-02")]
			if len(dateEvs) == 0 {
				logFn(fmt.Sprintf("INFO: Kein Event am %s – übersprungen.", einsatz.Datum))
				continue
			}
			if len(dateEvs) > 1 {
				logFn(fmt.Sprintf("WARNUNG: %d Events am %s – nehme: %s", len(dateEvs), einsatz.Datum, dateEvs[0].Name))
			}
			processEvent(dateEvs[0], einsatz.Musikteam, einsatz.Datum)
		}
	}

	return result
}
