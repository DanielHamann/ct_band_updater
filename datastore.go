package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/zalando/go-keyring"
)

const (
	keyringService  = "ct-musikteam"
	keyringUsername = "api_key"
)

// --- Data types ---

type Dienst struct {
	ID      string            `json:"id"`
	Label   string            `json:"label"`
	Persons map[string]string `json:"persons"`
}

type Musikteam struct {
	Teams   []string `json:"teams"`
	Dienste []Dienst `json:"dienste"`
}

type Einsatz struct {
	Datum     string `json:"datum"`
	Zeit      string `json:"zeit,omitempty"`
	Musikteam string `json:"musikteam"`
	EventID   int    `json:"event_id,omitempty"`
	EventName string `json:"event_name,omitempty"`
}

type Person struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type Settings struct {
	CtURL       string `json:"ct_url"`
	StoreAPIKey bool   `json:"store_api_key"`
}

type TeamleiterEntry struct {
	FactID int    `json:"fact_id"`
	Value  string `json:"value"`
}

type AppData struct {
	BuildID    string                      `json:"build_id,omitempty"`
	Settings   Settings                    `json:"settings"`
	Musikteam  Musikteam                   `json:"musikteam"`
	Einsaetze  []Einsatz                   `json:"einsaetze"`
	Teamleiter map[string]TeamleiterEntry  `json:"teamleiter"`
	Persons    []Person                    `json:"persons"`
}

var defaultData = AppData{
	Settings: Settings{
		CtURL:       "",
		StoreAPIKey: true,
	},
	Musikteam: Musikteam{
		Teams:   []string{},
		Dienste: []Dienst{},
	},
	Einsaetze:  []Einsatz{},
	Teamleiter: map[string]TeamleiterEntry{},
	Persons: []Person{},
}

// --- DataStore ---

type DataStore struct {
	mu   sync.RWMutex
	data AppData
	path string
}

func NewDataStore() *DataStore {
	homeDir, _ := os.UserHomeDir()
	dataDir := filepath.Join(homeDir, ".ct_musikteam")
	os.MkdirAll(dataDir, 0700)

	ds := &DataStore{
		path: filepath.Join(dataDir, "data.json"),
	}
	ds.load()
	return ds
}

func deepCopy(src AppData) AppData {
	b, err := json.Marshal(src)
	if err != nil {
		return src
	}
	var dst AppData
	if err := json.Unmarshal(b, &dst); err != nil {
		return src
	}
	return dst
}

func (ds *DataStore) load() {
	ds.data = deepCopy(defaultData)

	b, err := os.ReadFile(ds.path)
	if err != nil {
		return
	}

	var loaded AppData
	if err := json.Unmarshal(b, &loaded); err != nil {
		return
	}

	// Reset data on new build, preserve settings
	if loaded.BuildID != BuildID && BuildID != "dev" {
		ds.data.Settings = loaded.Settings
		ds.save()
		return
	}

	ds.data = loaded
	// Ensure slices are never nil
	if ds.data.Einsaetze == nil {
		ds.data.Einsaetze = []Einsatz{}
	}
	if ds.data.Persons == nil {
		ds.data.Persons = []Person{}
	}
	if ds.data.Teamleiter == nil {
		ds.data.Teamleiter = map[string]TeamleiterEntry{}
	}
	for i := range ds.data.Musikteam.Dienste {
		if ds.data.Musikteam.Dienste[i].Persons == nil {
			ds.data.Musikteam.Dienste[i].Persons = map[string]string{}
		}
	}
}

func (ds *DataStore) save() error {
	ds.data.BuildID = BuildID
	b, err := json.MarshalIndent(ds.data, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(ds.path, b, 0600)
}

// --- Settings ---

func (ds *DataStore) GetCtURL() string {
	ds.mu.RLock()
	defer ds.mu.RUnlock()
	return ds.data.Settings.CtURL
}

func (ds *DataStore) SetCtURL(url string) error {
	ds.mu.Lock()
	defer ds.mu.Unlock()
	url = strings.TrimSpace(url)
	for len(url) > 0 && url[len(url)-1] == '/' {
		url = url[:len(url)-1]
	}
	if url != "" && !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		url = "https://" + url
	}
	ds.data.Settings.CtURL = url
	return ds.save()
}

func (ds *DataStore) GetAPIKey() string {
	key, _ := keyring.Get(keyringService, keyringUsername)
	return key
}

func (ds *DataStore) SetAPIKey(key string) error {
	return keyring.Set(keyringService, keyringUsername, key)
}

func (ds *DataStore) DeleteAPIKey() {
	keyring.Delete(keyringService, keyringUsername)
}

func (ds *DataStore) GetStoreAPIKey() bool {
	ds.mu.RLock()
	defer ds.mu.RUnlock()
	return ds.data.Settings.StoreAPIKey
}

func (ds *DataStore) SetStoreAPIKey(value bool) error {
	ds.mu.Lock()
	defer ds.mu.Unlock()
	ds.data.Settings.StoreAPIKey = value
	if !value {
		keyring.Delete(keyringService, keyringUsername)
	}
	return ds.save()
}

// --- Musikteam ---

func (ds *DataStore) GetTeams() []string {
	ds.mu.RLock()
	defer ds.mu.RUnlock()
	out := make([]string, len(ds.data.Musikteam.Teams))
	copy(out, ds.data.Musikteam.Teams)
	return out
}

func (ds *DataStore) SetTeams(teams []string) error {
	ds.mu.Lock()
	defer ds.mu.Unlock()
	ds.data.Musikteam.Teams = teams
	return ds.save()
}

func (ds *DataStore) GetDienste() []Dienst {
	ds.mu.RLock()
	defer ds.mu.RUnlock()
	b, _ := json.Marshal(ds.data.Musikteam.Dienste)
	var out []Dienst
	json.Unmarshal(b, &out)
	return out
}

func (ds *DataStore) SetDienste(dienste []Dienst) error {
	ds.mu.Lock()
	defer ds.mu.Unlock()
	ds.data.Musikteam.Dienste = dienste
	return ds.save()
}

// --- Einsaetze ---

func (ds *DataStore) GetEinsaetze() []Einsatz {
	ds.mu.RLock()
	defer ds.mu.RUnlock()
	b, _ := json.Marshal(ds.data.Einsaetze)
	var out []Einsatz
	json.Unmarshal(b, &out)
	return out
}

func (ds *DataStore) SetEinsaetze(einsaetze []Einsatz) error {
	ds.mu.Lock()
	defer ds.mu.Unlock()
	ds.data.Einsaetze = einsaetze
	return ds.save()
}

// --- Teamleiter ---

func (ds *DataStore) GetTeamleiter() map[string]TeamleiterEntry {
	ds.mu.RLock()
	defer ds.mu.RUnlock()
	out := make(map[string]TeamleiterEntry)
	for k, v := range ds.data.Teamleiter {
		out[k] = v
	}
	return out
}

func (ds *DataStore) SetTeamleiter(tl map[string]TeamleiterEntry) error {
	ds.mu.Lock()
	defer ds.mu.Unlock()
	ds.data.Teamleiter = tl
	return ds.save()
}

// --- Persons ---

func (ds *DataStore) GetPersons() []Person {
	ds.mu.RLock()
	defer ds.mu.RUnlock()
	b, _ := json.Marshal(ds.data.Persons)
	var out []Person
	json.Unmarshal(b, &out)
	return out
}

func (ds *DataStore) SetPersons(persons []Person) error {
	ds.mu.Lock()
	defer ds.mu.Unlock()
	ds.data.Persons = persons
	return ds.save()
}
