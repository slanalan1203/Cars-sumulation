package main

import (
	"cars-simulation/internal/car"
	"cars-simulation/internal/intersection"
	"cars-simulation/internal/road"
	"cars-simulation/internal/server"
	"cars-simulation/internal/trafficlight"
	"cars-simulation/internal/world"
	"log"
	"net/http"
	"sync"
	"time"
)

func main() {
	road1 := road.NewRoad(
		road.Point{X: 0, Y: 0},
		road.Point{X: 100, Y: 0},
	)
	road1.ID = 1
	road2 := road.NewRoad(
		road.Point{X: 100, Y: 0},
		road.Point{X: 100, Y: 100},
	)
	road2.ID = 2

	road3 := road.NewRoad(
		road.Point{X: 100, Y: 0},
		road.Point{X: 200, Y: 0},
	)
	road3.ID = 3

	light := trafficlight.TrafficLight{
		Position:      30.0,
		RedDuration:   3.0,
		GreenDuration: 3.0,
		RoadID:        1,
	}

	cars := []car.Car{
		{
			Position:     20,
			Speed:        0,
			DesiredSpeed: 10,
			Acceleration: 2,
			Braking:      4,
			Route:        []int{1, 2},
			RoadIndex:    0,
		},
		{
			Position:     10,
			Speed:        0,
			DesiredSpeed: 10,
			Acceleration: 2,
			Braking:      4,
			Route:        []int{1, 2},
			RoadIndex:    0,
		},
	}

	sim := world.World{
		Cars: cars,
		Roads: map[int]road.Road{
			road1.ID: road1,
			road2.ID: road2,
			road3.ID: road3,
		},
		TrafficLights: []trafficlight.TrafficLight{light},
		Intersections: []intersection.Intersection{{ID: 1, IncomingRoads: []int{1}, OutgoingRoads: []int{2, 3}}},
		SpawnRoutes:   [][]int{{1, 2}, {1, 3}},
		SpawnInterval: 2.0,
	}

	if err := sim.Validate(); err != nil {
		log.Fatal(err)
	}

	var mu sync.RWMutex
	handler := server.NewHandler(&sim, &mu)

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
