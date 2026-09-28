package server

import (
	"encoding/json"
	"net/http"
	"sync"

	"github.com/slanalan1203/Cars-sumulation/internal/mapbuilder"
	"github.com/slanalan1203/Cars-sumulation/internal/world"
)

func NewHandler(sim *world.World, mu *sync.RWMutex, defaultFactory ...func() world.World) http.Handler {
	mux := http.NewServeMux()

	mu.Lock()
	sim.SaveInitialState()
	mu.Unlock()

	mux.HandleFunc("GET /state", func(w http.ResponseWriter, r *http.Request) {
		mu.RLock()
		snapshot := sim.Snapshot()
		mu.RUnlock()

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(snapshot)
	})

	mux.HandleFunc("POST /control", func(w http.ResponseWriter, r *http.Request) {
		var command struct {
			Action string `json:"action"`
		}
		if err := json.NewDecoder(r.Body).Decode(&command); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		mu.Lock()
		switch command.Action {
		case "pause":
			sim.Pause()
		case "resume":
			sim.Resume()
		case "reset":
			sim.Reset()
		default:
			http.Error(w, "unknown action", http.StatusBadRequest)
			mu.Unlock()
			return
		}
		mu.Unlock()

		w.WriteHeader(http.StatusOK)
	})

	mux.HandleFunc("POST /settings/home", func(w http.ResponseWriter, r *http.Request) {
		var command struct {
			HomeID        int      `json:"homeID"`
			CarsPerMinute *float64 `json:"carsPerMinute"`
		}
		decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1024))
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&command); err != nil || command.CarsPerMinute == nil {
			http.Error(w, "неверная настройка дома", http.StatusBadRequest)
			return
		}
		mu.Lock()
		err := sim.SetHomeSpawnRate(command.HomeID, *command.CarsPerMinute)
		mu.Unlock()
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})

	mux.HandleFunc("POST /settings/road", func(w http.ResponseWriter, r *http.Request) {
		var command struct {
			RoadID   int      `json:"roadID"`
			SpeedKmh *float64 `json:"speedKmh"`
			Closed   *bool    `json:"closed"`
		}
		decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1024))
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&command); err != nil {
			http.Error(w, "неверная настройка дороги", http.StatusBadRequest)
			return
		}
		mu.Lock()
		err := sim.SetRoadSettings(command.RoadID, command.SpeedKmh, command.Closed)
		mu.Unlock()
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})

	mux.HandleFunc("GET /map", func(w http.ResponseWriter, r *http.Request) {
		value, err := mapbuilder.Load()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "no-store")
		json.NewEncoder(w).Encode(value)
	})

	mux.HandleFunc("POST /map", func(w http.ResponseWriter, r *http.Request) {
		var value mapbuilder.Map
		decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&value); err != nil {
			http.Error(w, "неверный формат карты", http.StatusBadRequest)
			return
		}
		if err := mapbuilder.Save(value); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})

	mux.HandleFunc("POST /map/run", func(w http.ResponseWriter, r *http.Request) {
		value, err := mapbuilder.Load()
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		created, err := value.Build()
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		mu.Lock()
		*sim = created
		sim.SaveInitialState()
		mu.Unlock()
		w.WriteHeader(http.StatusNoContent)
	})

	mux.HandleFunc("POST /map/default", func(w http.ResponseWriter, r *http.Request) {
		if len(defaultFactory) == 0 {
			http.NotFound(w, r)
			return
		}
		created := defaultFactory[0]()
		mu.Lock()
		*sim = created
		sim.SaveInitialState()
		mu.Unlock()
		w.WriteHeader(http.StatusNoContent)
	})

	mux.HandleFunc("GET /editor", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		http.ServeFile(w, r, "editor.html")
	})

	mux.HandleFunc("GET /editor.js", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		http.ServeFile(w, r, "editor.js")
	})

	mux.HandleFunc("/simulation.js", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		http.ServeFile(w, r, "simulation.js")
	})

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Cache-Control", "no-store")
		http.ServeFile(w, r, "index.html")
	})

	mux.HandleFunc("/Sprites/car.png", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "Sprites/car.png")
	})

	return mux
}
