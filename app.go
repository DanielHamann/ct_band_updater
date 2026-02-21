package main

import (
	"context"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type App struct {
	ctx   context.Context
	store *DataStore
}

func NewApp() *App {
	return &App{store: NewDataStore()}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// --- Settings ---

func (a *App) GetSettings() map[string]interface{} {
	return map[string]interface{}{
		"ctURL":       a.store.GetCtURL(),
		"apiKey":      a.store.GetAPIKey(),
		"storeAPIKey": a.store.GetStoreAPIKey(),
	}
}

func (a *App) SaveSettings(ctURL string, apiKey string, storeKey bool) error {
	if err := a.store.SetCtURL(ctURL); err != nil {
		return err
	}
	if err := a.store.SetStoreAPIKey(storeKey); err != nil {
		return err
	}
	if storeKey {
		return a.store.SetAPIKey(apiKey)
	}
	a.store.DeleteAPIKey()
	return nil
}

// --- Musikteam ---

func (a *App) GetMusikteam() map[string]interface{} {
	return map[string]interface{}{
		"teams":   a.store.GetTeams(),
		"dienste": a.store.GetDienste(),
	}
}

func (a *App) SaveMusikteam(teams []string, dienste []Dienst) error {
	if err := a.store.SetTeams(teams); err != nil {
		return err
	}
	return a.store.SetDienste(dienste)
}

func (a *App) FetchMasterData() (*CTMasterData, error) {
	return fetchMasterData(a.store.GetCtURL(), a.store.GetAPIKey())
}

func (a *App) FetchFacts() ([]CTFactDefinition, error) {
	return fetchFacts(a.store.GetCtURL(), a.store.GetAPIKey())
}

func (a *App) FetchEventFacts(eventID int) ([]EventFactDisplay, error) {
	return fetchEventFacts(a.store.GetCtURL(), a.store.GetAPIKey(), eventID)
}

func (a *App) FetchEvents(from, to string) ([]CTEvent, error) {
	return fetchEvents(a.store.GetCtURL(), a.store.GetAPIKey(), from, to)
}

func (a *App) GetPersons() []Person {
	return a.store.GetPersons()
}

func (a *App) FetchPersons() ([]Person, error) {
	persons, err := FetchPersonsFromAPI(a.store.GetCtURL(), a.store.GetAPIKey())
	if err != nil {
		return nil, err
	}
	a.store.SetPersons(persons)
	return persons, nil
}

// --- Einsaetze ---

func (a *App) GetEinsaetze() []Einsatz {
	return a.store.GetEinsaetze()
}

func (a *App) SetEinsaetze(einsaetze []Einsatz) error {
	return a.store.SetEinsaetze(einsaetze)
}

// --- Teamleiter ---

func (a *App) GetTeamleiter() map[string]TeamleiterEntry {
	return a.store.GetTeamleiter()
}

func (a *App) SetTeamleiter(tl map[string]TeamleiterEntry) error {
	return a.store.SetTeamleiter(tl)
}

// --- Run ---

func (a *App) RunUpdate() {
	go func() {
		result := runUpdateProcess(
			a.store.GetCtURL(),
			a.store.GetAPIKey(),
			a.store.GetEinsaetze(),
			a.store.GetDienste(),
			a.store.GetTeams(),
			a.store.GetTeamleiter(),
			func(msg string) {
				runtime.EventsEmit(a.ctx, "run:log", msg)
			},
		)
		runtime.EventsEmit(a.ctx, "run:done", result)
	}()
}
