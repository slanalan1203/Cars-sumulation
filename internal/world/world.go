package world

import (
	"math"

	"github.com/slanalan1203/Cars-sumulation/internal/car"
	"github.com/slanalan1203/Cars-sumulation/internal/intersection"
	"github.com/slanalan1203/Cars-sumulation/internal/road"
	"github.com/slanalan1203/Cars-sumulation/internal/trafficlight"
)

type World struct {
	Roads                map[int]road.Road
	Cars                 []car.Car
	TrafficLights        []trafficlight.TrafficLight
	Intersections        []intersection.Intersection
	Roundabouts          []Roundabout
	Buildings            []Building
	Arrivals             map[int]int
	CompletedTrips       int
	TotalTravelTime      float64
	TotalWaitTime        float64
	Paused               bool
	CustomMap            bool
	SpawnInterval        float64
	SpawnRoutes          [][]int
	SpawnSources         []SpawnSource
	spawnElapsed         float64
	nextSpawnRoute       int
	lastIntersectionRoad map[int]int
	initial              *State
}

func (w *World) Step(dt float64) {
	if w.Paused || dt <= 0 || math.IsNaN(dt) || math.IsInf(dt, 0) {
		return
	}
	w.updateTrafficLights(dt)
	w.updateCars(dt)
	w.removeFinishedCars()
	w.spawnCars(dt)
}

func (w *World) updateTrafficLights(dt float64) {
	for i := range w.TrafficLights {
		w.TrafficLights[i].Update(dt)
	}
}

func (w *World) updateCars(dt float64) {
	next := append([]car.Car(nil), w.Cars...)
	for i := range next {
		current := &next[i]
		if current.Finished() {
			continue
		}
		obstacle := w.findObstacleAhead(i)
		acceleration := current.CalculateAccelerationForSpeed(obstacle, dt, w.speedTarget(*current, dt))
		oldSpeed := current.Speed
		current.UpdateSpeed(acceleration, dt)

		distance := (oldSpeed + current.Speed) / 2 * dt
		if distance >= obstacle.Distance {
			distance = math.Max(0, obstacle.Distance)
			current.Speed = 0
		}
		current.Advance(distance, w.Roads)
	}
	w.resolveIntersectionConflicts(next)
	for i := range next {
		next[i].TripTime += dt
		if next[i].Speed < 0.5 {
			next[i].WaitTime += dt
		}
	}
	w.Cars = next
}

func (w *World) removeFinishedCars() {
	remaining := make([]car.Car, 0, len(w.Cars))
	for _, current := range w.Cars {
		if current.Finished() {
			w.CompletedTrips++
			w.TotalTravelTime += current.TripTime
			w.TotalWaitTime += current.WaitTime
			if current.DestinationID != 0 {
				if w.Arrivals == nil {
					w.Arrivals = make(map[int]int)
				}
				w.Arrivals[current.DestinationID]++
			}
			continue
		}
		remaining = append(remaining, current)
	}
	w.Cars = remaining
}
