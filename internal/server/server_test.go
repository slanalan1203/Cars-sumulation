package server

import (
	"cars-simulation/internal/world"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
)

func TestControlPause(t *testing.T) {
	sim := world.World{Paused: false}
	var mu sync.RWMutex

	handler := NewHandler(&sim, &mu)

	request := httptest.NewRequest(
		http.MethodPost,
		"/control",
		strings.NewReader(`{"action":"pause"}`),
	)
	request.Header.Set("Content-Type", "application/json")

	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf(
			"ожидался HTTP 200, получено %d: %s",
			response.Code,
			response.Body.String(),
		)
	}

	if !sim.Paused {
		t.Error("после команды pause симуляция должна быть на паузе")
	}
}

func TestControlResume(t *testing.T) {
	sim := world.World{Paused: true}
	var mu sync.RWMutex

	handler := NewHandler(&sim, &mu)

	request := httptest.NewRequest(
		http.MethodPost,
		"/control",
		strings.NewReader(`{"action":"resume"}`),
	)
	request.Header.Set("Content-Type", "application/json")

	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf(
			"ожидался HTTP 200, получено %d: %s",
			response.Code,
			response.Body.String(),
		)
	}

	if sim.Paused {
		t.Error("после команды resume симуляция должна продолжиться")
	}
}
