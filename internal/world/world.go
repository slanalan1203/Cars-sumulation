package world

import (
	"cars-simulation/internal/car"
	"cars-simulation/internal/intersection"
	"cars-simulation/internal/road"
	"cars-simulation/internal/trafficlight"
	"math"
)

// World — состояние симуляции. Синхронизация HTTP и Step выполняется снаружи.
type World struct {
	Roads          map[int]road.Road
	Cars           []car.Car
	TrafficLights  []trafficlight.TrafficLight
	Intersections  []intersection.Intersection
	Paused         bool
	SpawnInterval  float64
	SpawnRoutes    [][]int
	spawnElapsed   float64
	nextSpawnRoute int
	initial        *State
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
	// Читаем препятствия из w.Cars, изменяем только next.
	// Поэтому результат не зависит от порядка обхода машин.
	next := append([]car.Car(nil), w.Cars...)
	for i := range next {
		current := &next[i]
		if current.Finished() {
			continue
		}
		obstacle := w.findObstacleAhead(i)
		acceleration := current.CalculateAcceleration(obstacle, dt)
		oldSpeed := current.Speed
		current.UpdateSpeed(acceleration, dt)

		// Сохраняем прежнюю интеграцию через среднюю скорость за шаг.
		distance := (oldSpeed + current.Speed) / 2 * dt
		if distance >= obstacle.Distance {
			distance = math.Max(0, obstacle.Distance)
			current.Speed = 0
		}
		current.Advance(distance, w.Roads)
	}
	w.Cars = next
}

func (w *World) removeFinishedCars() {
	remaining := make([]car.Car, 0, len(w.Cars))
	for _, current := range w.Cars {
		if !current.Finished() {
			remaining = append(remaining, current)
		}
	}
	w.Cars = remaining
}
