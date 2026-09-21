package server

import (
	"cars-simulation/internal/car"
	"cars-simulation/internal/road"
	"cars-simulation/internal/trafficlight"
	"cars-simulation/internal/world"
	"encoding/json"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
)

func TestResetAndStateAPI(t *testing.T) {
	w := world.World{
		Roads:         map[int]road.Road{1: {ID: 1, Length: 100}},
		Cars:          []car.Car{{Position: 20, Route: []int{1}}},
		TrafficLights: []trafficlight.TrafficLight{{RoadID: 1, Position: 30, RedDuration: 3, GreenDuration: 3}},
	}
	var mu sync.RWMutex
	handler := NewHandler(&w, &mu)
	w.Cars[0].Position = 80
	w.TrafficLights[0].Green = true
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest("POST", "/control", strings.NewReader(`{"action":"reset"}`)))
	if response.Code != 200 {
		t.Fatalf("reset status %d", response.Code)
	}
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest("GET", "/state", nil))
	var state world.State
	if err := json.Unmarshal(response.Body.Bytes(), &state); err != nil {
		t.Fatal(err)
	}
	if !state.Paused || len(state.Cars) != 1 || state.Cars[0].Position != 20 || len(state.TrafficLights) != 1 || state.TrafficLights[0].Green {
		t.Fatalf("bad reset response: %s", response.Body.String())
	}
}

func TestStateRequestsWhileWorldSteps(t *testing.T) {
	w := world.World{
		Roads: map[int]road.Road{1: {ID: 1, Length: 100}},
		Cars:  []car.Car{{Position: 20, DesiredSpeed: 10, Acceleration: 2, Braking: 4, Route: []int{1}}},
	}
	var mu sync.RWMutex
	handler := NewHandler(&w, &mu)
	var done sync.WaitGroup
	done.Add(1)
	go func() {
		defer done.Done()
		for i := 0; i < 100; i++ {
			mu.Lock()
			w.Step(0.1)
			mu.Unlock()
		}
	}()
	for i := 0; i < 100; i++ {
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, httptest.NewRequest("GET", "/state", nil))
		var state world.State
		if response.Code != 200 || json.Unmarshal(response.Body.Bytes(), &state) != nil {
			t.Error("invalid state response")
		}
	}
	done.Wait()
}
