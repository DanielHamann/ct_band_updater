package main

import (
	"fmt"
	"strings"
	"testing"
)

// fakeCTClient implements ctRunClient with an in-memory "server" state,
// so runUpdateProcess can be exercised without a live ChurchTools instance.
type fakeCTClient struct {
	events []CTEvent

	factUpdates []factUpdateCall
	slotAdds    []slotAddCall
	reqUpdates  []reqUpdateCall

	nextServiceInstanceID int
}

type factUpdateCall struct {
	EventID, FactID int
	Value           string
}

type slotAddCall struct {
	EventID   int
	Additions map[int]int
}

type reqUpdateCall struct {
	EventID, RequestID, PersonID int
}

func newFakeCTClient(events ...CTEvent) *fakeCTClient {
	maxID := 0
	for _, e := range events {
		for _, s := range e.EventServices {
			if s.ID > maxID {
				maxID = s.ID
			}
		}
	}
	return &fakeCTClient{events: events, nextServiceInstanceID: maxID + 1}
}

func (f *fakeCTClient) indexOf(eventID int) int {
	for i := range f.events {
		if f.events[i].ID == eventID {
			return i
		}
	}
	return -1
}

func (f *fakeCTClient) FetchEvents(from, to string) ([]CTEvent, error) {
	out := make([]CTEvent, len(f.events))
	copy(out, f.events)
	return out, nil
}

func (f *fakeCTClient) FetchEvent(eventID int) (*CTEvent, error) {
	i := f.indexOf(eventID)
	if i < 0 {
		return nil, fmt.Errorf("fake: event %d not found", eventID)
	}
	ev := f.events[i]
	return &ev, nil
}

func (f *fakeCTClient) AddServiceSlotsToEvent(eventID int, additions map[int]int) error {
	f.slotAdds = append(f.slotAdds, slotAddCall{eventID, additions})
	i := f.indexOf(eventID)
	if i < 0 {
		return fmt.Errorf("fake: event %d not found", eventID)
	}
	for sid, count := range additions {
		for n := 0; n < count; n++ {
			f.events[i].EventServices = append(f.events[i].EventServices, CTEventService{
				ID:        f.nextServiceInstanceID,
				ServiceID: sid,
				Name:      fmt.Sprintf("Service %d", sid),
			})
			f.nextServiceInstanceID++
		}
	}
	return nil
}

func (f *fakeCTClient) UpdateServiceRequest(eventID, requestID, personID int) error {
	f.reqUpdates = append(f.reqUpdates, reqUpdateCall{eventID, requestID, personID})
	return nil
}

func (f *fakeCTClient) UpdateEventFact(eventID, factID int, value string) error {
	f.factUpdates = append(f.factUpdates, factUpdateCall{eventID, factID, value})
	return nil
}

func collectLogs() (func(string), *[]string) {
	logs := []string{}
	return func(s string) { logs = append(logs, s) }, &logs
}

func containsLog(logs []string, substr string) bool {
	for _, l := range logs {
		if strings.Contains(l, substr) {
			return true
		}
	}
	return false
}

func TestRunUpdateProcess_MatchesByEventID_CreatesSlotsAndAssignsPersons(t *testing.T) {
	fake := newFakeCTClient(CTEvent{ID: 42, Name: "Gottesdienst", StartDate: "2026-01-04T10:00:00Z"})
	dienste := []Dienst{{ID: "1", Label: "Gesang", Persons: map[string]string{"Band A": "100;200"}}}
	logFn, _ := collectLogs()

	result := runUpdateProcess(fake, []Einsatz{{EventID: 42, Datum: "04.01.2026", Musikteam: "Band A"}}, dienste, []string{"Band A"}, map[string]TeamleiterEntry{}, logFn)

	if len(fake.slotAdds) != 1 || fake.slotAdds[0].EventID != 42 || fake.slotAdds[0].Additions[1] != 2 {
		t.Fatalf("expected AddServiceSlotsToEvent(42, {1: 2}), got %+v", fake.slotAdds)
	}
	if len(fake.reqUpdates) != 2 {
		t.Fatalf("expected 2 UpdateServiceRequest calls, got %d: %+v", len(fake.reqUpdates), fake.reqUpdates)
	}
	if fake.reqUpdates[0].PersonID != 100 || fake.reqUpdates[1].PersonID != 200 {
		t.Fatalf("expected persons assigned in order 100, 200, got %+v", fake.reqUpdates)
	}
	if result.Updated != 2 || result.Errors != 0 {
		t.Fatalf("expected RunResult{Updated: 2, Errors: 0}, got %+v", result)
	}
}

func TestRunUpdateProcess_MatchesByDate(t *testing.T) {
	fake := newFakeCTClient(CTEvent{ID: 7, Name: "Gottesdienst", StartDate: "2026-02-15T09:00:00Z"})
	teamleiter := map[string]TeamleiterEntry{"Band A": {FactID: 5, Value: "Max"}}
	logFn, _ := collectLogs()

	runUpdateProcess(fake, []Einsatz{{EventID: 0, Datum: "15.02.2026", Musikteam: "Band A"}}, nil, []string{"Band A"}, teamleiter, logFn)

	if len(fake.factUpdates) != 1 || fake.factUpdates[0].EventID != 7 || fake.factUpdates[0].Value != "Max" {
		t.Fatalf("expected UpdateEventFact(7, 5, \"Max\") via date match, got %+v", fake.factUpdates)
	}
}

func TestRunUpdateProcess_WarnsOnUnconfiguredTeam(t *testing.T) {
	fake := newFakeCTClient(CTEvent{ID: 1, Name: "Gottesdienst", StartDate: "2026-03-01T09:00:00Z"})
	logFn, logs := collectLogs()

	result := runUpdateProcess(fake, []Einsatz{{EventID: 1, Datum: "01.03.2026", Musikteam: "Unbekanntes Team"}}, nil, []string{"Band A"}, nil, logFn)

	if len(fake.factUpdates)+len(fake.slotAdds)+len(fake.reqUpdates) != 0 {
		t.Fatalf("expected no client calls for an unconfigured team, got facts=%+v slots=%+v reqs=%+v", fake.factUpdates, fake.slotAdds, fake.reqUpdates)
	}
	if !containsLog(*logs, "nicht konfiguriert") {
		t.Fatalf("expected a log line about the team not being configured, got %v", *logs)
	}
	if result.Errors != 0 {
		t.Fatalf("expected an unconfigured team to be a skip, not an error, got %+v", result)
	}
}

func TestRunUpdateProcess_SkipsWhenNotEnoughPersons(t *testing.T) {
	fake := newFakeCTClient(CTEvent{
		ID:            9,
		Name:          "Gottesdienst",
		StartDate:     "2026-04-05T09:00:00Z",
		EventServices: []CTEventService{{ID: 501, ServiceID: 1, Name: "Gesang"}, {ID: 502, ServiceID: 1, Name: "Gesang"}},
	})
	dienste := []Dienst{{ID: "1", Label: "Gesang", Persons: map[string]string{"Band A": "100"}}}
	logFn, logs := collectLogs()

	result := runUpdateProcess(fake, []Einsatz{{EventID: 9, Datum: "05.04.2026", Musikteam: "Band A"}}, dienste, []string{"Band A"}, nil, logFn)

	if len(fake.slotAdds) != 0 {
		t.Fatalf("expected no new slots to be created (2 already exist), got %+v", fake.slotAdds)
	}
	if len(fake.reqUpdates) != 1 || fake.reqUpdates[0].PersonID != 100 {
		t.Fatalf("expected exactly 1 UpdateServiceRequest for the configured person, got %+v", fake.reqUpdates)
	}
	if !containsLog(*logs, "ÜBERSPRUNGEN") {
		t.Fatalf("expected a log line about the skipped second slot, got %v", *logs)
	}
	if result.Updated != 1 || result.Errors != 0 {
		t.Fatalf("expected RunResult{Updated: 1, Errors: 0}, got %+v", result)
	}
}
