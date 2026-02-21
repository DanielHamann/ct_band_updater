package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"time"
)

func apiHeader(apiKey string) http.Header {
	h := http.Header{}
	h.Set("Authorization", "Login "+apiKey)
	h.Set("Content-Type", "application/json")
	return h
}

var httpClient = &http.Client{Timeout: 30 * time.Second}

func doRequest(method, url, apiKey, body string) ([]byte, error) {
	var bodyReader io.Reader
	if body != "" {
		bodyReader = strings.NewReader(body)
	}
	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return nil, err
	}
	req.Header = apiHeader(apiKey)

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(data))
	}
	return data, nil
}

// --- Events ---

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

// EventFactDisplay is returned to the frontend for fact selection.
type EventFactDisplay struct {
	FactID int    `json:"factId"`
	Value  string `json:"value"`
}

func fetchEventFacts(ctURL, apiKey string, eventID int) ([]EventFactDisplay, error) {
	url := fmt.Sprintf("%s/api/events/%d/facts", ctURL, eventID)
	data, err := doRequest("GET", url, apiKey, "")
	if err != nil {
		return nil, err
	}

	// The API value field can be string, number, or null — use RawMessage.
	var result struct {
		Data []struct {
			FactID int             `json:"factId"`
			Value  json.RawMessage `json:"value"`
		} `json:"data"`
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}

	var out []EventFactDisplay
	for _, f := range result.Data {
		var strVal string
		// Try to unmarshal as string first, fall back to raw JSON.
		if err := json.Unmarshal(f.Value, &strVal); err != nil {
			strVal = strings.Trim(string(f.Value), "\"")
		}
		out = append(out, EventFactDisplay{FactID: f.FactID, Value: strVal})
	}
	return out, nil
}

func fetchEvents(ctURL, apiKey, from, to string) ([]CTEvent, error) {
	url := fmt.Sprintf("%s/api/events?from=%s&to=%s&direction=forward&limit=100&include=eventServices", ctURL, from, to)
	data, err := doRequest("GET", url, apiKey, "")
	if err != nil {
		return nil, err
	}
	var result struct {
		Data []CTEvent `json:"data"`
	}
	json.Unmarshal(data, &result)
	return result.Data, nil
}

func fetchEvent(ctURL, apiKey string, eventID int) (*CTEvent, error) {
	url := fmt.Sprintf("%s/api/events/%d?include=eventServices", ctURL, eventID)
	data, err := doRequest("GET", url, apiKey, "")
	if err != nil {
		return nil, err
	}
	var result struct {
		Data CTEvent `json:"data"`
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}
	return &result.Data, nil
}

// addServiceSlotsToEvent creates missing service slots via POST /api/events/{eventId}/servicerequests.
// additions maps serviceID -> number of additional slots to create.
func addServiceSlotsToEvent(ctURL, apiKey string, eventID int, additions map[int]int) error {
	type svcReq struct {
		Count     int `json:"count"`
		ServiceID int `json:"serviceId"`
	}
	services := make([]svcReq, 0, len(additions))
	for sid, count := range additions {
		services = append(services, svcReq{Count: count, ServiceID: sid})
	}
	body, _ := json.Marshal(map[string]interface{}{
		"eventId":  eventID,
		"services": services,
	})
	url := fmt.Sprintf("%s/api/events/%d/servicerequests", ctURL, eventID)
	_, err := doRequest("PUT", url, apiKey, string(body))
	return err
}

func updateServiceRequest(ctURL, apiKey string, eventID, requestID, personID int) error {
	url := fmt.Sprintf("%s/api/events/%d/servicerequests/%d", ctURL, eventID, requestID)
	body, _ := json.Marshal(map[string]interface{}{
		"personId":   personID,
		"isAccepted": false,
		"isValid":    true,
	})
	_, err := doRequest("PUT", url, apiKey, string(body))
	return err
}

func updateEventFact(ctURL, apiKey string, eventID, factID int, value string) error {
	url := fmt.Sprintf("%s/api/events/%d/facts/%d", ctURL, eventID, factID)
	body, _ := json.Marshal(map[string]string{"value": value})
	_, err := doRequest("PUT", url, apiKey, string(body))
	return err
}

// --- Facts ---

type CTFactDefinition struct {
	ID             int      `json:"id"`
	Name           string   `json:"name"`
	NameTranslated string   `json:"nameTranslated"`
	SortKey        int      `json:"sortKey"`
	Type           string   `json:"type"`
	Options        []string `json:"options"`
}

func fetchFacts(ctURL, apiKey string) ([]CTFactDefinition, error) {
	data, err := doRequest("GET", ctURL+"/api/facts", apiKey, "")
	if err != nil {
		return nil, err
	}
	var result struct {
		Data []CTFactDefinition `json:"data"`
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}
	sort.Slice(result.Data, func(i, j int) bool {
		return result.Data[i].SortKey < result.Data[j].SortKey
	})
	return result.Data, nil
}

// --- Masterdata ---

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

func fetchMasterData(ctURL, apiKey string) (*CTMasterData, error) {
	data, err := doRequest("GET", ctURL+"/api/event/masterdata", apiKey, "")
	if err != nil {
		return nil, err
	}
	var result struct {
		Data CTMasterData `json:"data"`
	}
	json.Unmarshal(data, &result)
	sort.Slice(result.Data.ServiceGroups, func(i, j int) bool {
		return result.Data.ServiceGroups[i].SortKey < result.Data.ServiceGroups[j].SortKey
	})
	sort.Slice(result.Data.Services, func(i, j int) bool {
		return result.Data.Services[i].SortKey < result.Data.Services[j].SortKey
	})
	return &result.Data, nil
}

// --- Persons ---

func FetchPersonsFromAPI(ctURL, apiKey string) ([]Person, error) {
	var persons []Person
	for page := 1; ; page++ {
		url := fmt.Sprintf("%s/api/persons?limit=100&page=%d", ctURL, page)
		data, err := doRequest("GET", url, apiKey, "")
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
				Pagination struct {
					LastPage int `json:"lastPage"`
				} `json:"pagination"`
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

// --- Date range ---

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
		return "", "", fmt.Errorf("keine gueltigen Daten")
	}
	return min.Format("2006-01-02"), max.Format("2006-01-02"), nil
}

// --- Run update process ---

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
		logFn("FEHLER: Kein API-Schluessel gesetzt (Einstellungen).")
		result.Errors++
		return result
	}
	if len(einsaetze) == 0 {
		logFn("FEHLER: Keine Einsaetze definiert.")
		result.Errors++
		return result
	}

	// Build team config: team -> serviceID -> []personID
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
			var pids []int
			for _, p := range strings.Split(personStr, ";") {
				p = strings.TrimSpace(p)
				var pid int
				if _, err := fmt.Sscanf(p, "%d", &pid); err == nil && pid > 0 {
					pids = append(pids, pid)
				}
			}
			if _, ok := teamConfig[team]; !ok {
				teamConfig[team] = make(map[int][]int)
			}
			teamConfig[team][sid] = append(teamConfig[team][sid], pids...)
		}
	}

	fromDate, toDate, err := getDateRange(einsaetze)
	if err != nil {
		logFn("FEHLER: " + err.Error())
		result.Errors++
		return result
	}

	logFn(fmt.Sprintf("Abrufen von Events: %s bis %s...", fromDate, toDate))
	events, err := fetchEvents(ctURL, apiKey, fromDate, toDate)
	if err != nil {
		logFn("FEHLER beim Abrufen der Events: " + err.Error())
		result.Errors++
		return result
	}
	logFn(fmt.Sprintf("Gefunden: %d Events", len(events)))

	// Build lookup maps
	eventsById := make(map[int]CTEvent)
	eventsByDate := make(map[string][]CTEvent)
	for _, ev := range events {
		eventsById[ev.ID] = ev
		if len(ev.StartDate) >= 10 {
			date := ev.StartDate[:10]
			eventsByDate[date] = append(eventsByDate[date], ev)
		}
	}

	processEvent := func(ev CTEvent, teamName, displayDate string) {
		if _, ok := teamConfig[teamName]; !ok {
			logFn(fmt.Sprintf("Warnung: Team '%s' nicht konfiguriert - %s uebersprungen", teamName, displayDate))
			return
		}

		logFn(fmt.Sprintf("\nVerarbeite: %s (%s) - Musikteam: %s", ev.Name, displayDate, teamName))

		tlEntry := teamleiter[teamName]
		if tlEntry.FactID > 0 {
			value := tlEntry.Value
			if value == "" {
				value = teamName
			}
			if err := updateEventFact(ctURL, apiKey, ev.ID, tlEntry.FactID, value); err != nil {
				logFn("  FEHLER Event-Fact: " + err.Error())
			} else {
				logFn("  OK: Event-Fact gesetzt: " + value)
			}
		}

		svcConfig := teamConfig[teamName]

		// Detect missing service slots and create them
		existingCounts := make(map[int]int)
		for _, svc := range ev.EventServices {
			existingCounts[svc.ServiceID]++
		}
		additions := make(map[int]int)
		for sid, pids := range svcConfig {
			if len(pids) == 0 {
				continue
			}
			missing := len(pids) - existingCounts[sid]
			if missing > 0 {
				additions[sid] = missing
			}
		}
		if len(additions) > 0 {
			logFn(fmt.Sprintf("  Erstelle %d fehlende Dienst-Slot(s)...", func() int {
				n := 0
				for _, c := range additions {
					n += c
				}
				return n
			}()))
			if err := addServiceSlotsToEvent(ctURL, apiKey, ev.ID, additions); err != nil {
				logFn("  FEHLER: Dienst-Slots erstellen: " + err.Error())
				result.Errors++
			} else {
				// Re-fetch event to get the new slot IDs
				if updated, err := fetchEvent(ctURL, apiKey, ev.ID); err != nil {
					logFn("  FEHLER: Event neu laden: " + err.Error())
					result.Errors++
				} else {
					ev = *updated
					logFn("  OK: Dienst-Slots erstellt.")
				}
			}
		}

		instanceCounter := make(map[int]int)
		for _, svc := range ev.EventServices {
			pids, ok := svcConfig[svc.ServiceID]
			if !ok {
				continue
			}
			idx := instanceCounter[svc.ServiceID]
			instanceCounter[svc.ServiceID]++
			if idx >= len(pids) {
				logFn(fmt.Sprintf("  UEBERSPRUNGEN: %s (Instanz %d) - keine Person", svc.Name, idx+1))
				continue
			}
			personID := pids[idx]
			if err := updateServiceRequest(ctURL, apiKey, ev.ID, svc.ID, personID); err != nil {
				logFn(fmt.Sprintf("  FEHLER: %s: %s", svc.Name, err.Error()))
				result.Errors++
			} else {
				inst := ""
				if len(pids) > 1 {
					inst = fmt.Sprintf(" (Instanz %d)", idx+1)
				}
				logFn(fmt.Sprintf("  OK: %s%s -> Person %d", svc.Name, inst, personID))
				result.Updated++
			}
		}
	}

	for _, einsatz := range einsaetze {
		teamName := einsatz.Musikteam
		displayDate := einsatz.Datum

		if einsatz.EventID != 0 {
			// Direct event ID lookup — no ambiguity
			ev, ok := eventsById[einsatz.EventID]
			if !ok {
				logFn(fmt.Sprintf("WARNUNG: Event ID %d nicht im abgerufenen Zeitraum - uebersprungen", einsatz.EventID))
				continue
			}
			processEvent(ev, teamName, displayDate)
		} else {
			// Fallback: date-based lookup for manually entered einsaetze
			t, err := time.Parse("02.01.2006", einsatz.Datum)
			if err != nil {
				logFn(fmt.Sprintf("WARNUNG: Ungültiges Datum '%s' - uebersprungen", einsatz.Datum))
				continue
			}
			dateEvs := eventsByDate[t.Format("2006-01-02")]
			if len(dateEvs) == 0 {
				logFn(fmt.Sprintf("INFO: Kein Event am %s - uebersprungen", displayDate))
				continue
			}
			if len(dateEvs) > 1 {
				logFn(fmt.Sprintf("WARNUNG: %d Events am %s - nehme: %s", len(dateEvs), displayDate, dateEvs[0].Name))
			}
			processEvent(dateEvs[0], teamName, displayDate)
		}
	}

	return result
}
