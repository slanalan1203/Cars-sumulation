package server

import (
	"cars-simulation/internal/world"
	"encoding/json"
	"net/http"
	"sync"
)

func NewHandler(sim *world.World, mu *sync.RWMutex) http.Handler {
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
