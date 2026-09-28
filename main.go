package main

import (
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/slanalan1203/Cars-sumulation/internal/server"
)

func main() {
	sim := newScenario()
	if err := sim.Validate(); err != nil {
		log.Fatal(err)
	}
	var mu sync.RWMutex
	handler := server.NewHandler(&sim, &mu, newScenario)
	go func() {
		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()
		for range ticker.C {
			mu.Lock()
			sim.Step(0.1)
			mu.Unlock()
		}
	}()
	log.Println("Открой http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", handler))
}
