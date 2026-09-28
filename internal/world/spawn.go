package world

import (
	"fmt"
	"math"

	"github.com/slanalan1203/Cars-sumulation/internal/car"
)

type SpawnSource struct {
	RoadID         int
	HomeID         int
	Interval       float64
	Routes         [][]int
	DestinationIDs []int
	elapsed        float64
	nextRoute      int
}

func (w *World) SetHomeSpawnRate(homeID int, carsPerMinute float64) error {
	if math.IsNaN(carsPerMinute) || math.IsInf(carsPerMinute, 0) || carsPerMinute < 0 || carsPerMinute > 30 {
		return fmt.Errorf("частота появления должна быть от 0 до 30 машин в минуту")
	}
	for i := range w.SpawnSources {
		if w.SpawnSources[i].HomeID == homeID {
			w.SpawnSources[i].Interval = 0
			if carsPerMinute > 0 {
				w.SpawnSources[i].Interval = 60 / carsPerMinute
			}
			w.SpawnSources[i].elapsed = 0
			return nil
		}
	}
	return fmt.Errorf("дом %d не является источником машин", homeID)
}

func (w *World) spawnCars(dt float64) {
	for i := range w.SpawnSources {
		source := &w.SpawnSources[i]
		if source.Interval <= 0 || len(source.Routes) == 0 {
			continue
		}
		source.elapsed += dt
		if source.elapsed < source.Interval {
			continue
		}
		routeIndex := -1
		for step := range source.Routes {
			candidate := (source.nextRoute + step) % len(source.Routes)
			if w.routeOpen(source.Routes[candidate]) || w.canReroute(source.Routes[candidate]) {
				routeIndex = candidate
				break
			}
		}
		if routeIndex < 0 {
			continue
		}
		route := source.Routes[routeIndex]
		if !w.routeOpen(route) {
			route, _ = w.routeAroundClosure(route, 0)
		}
		if !w.canSpawn(route) {
			continue
		}
		destinationID := 0
		if len(source.DestinationIDs) > 0 {
			destinationID = source.DestinationIDs[routeIndex]
		}
		w.addCar(route, source.HomeID, destinationID)
		source.elapsed = 0
		source.nextRoute = (source.nextRoute + 1) % len(source.Routes)
	}

	if w.SpawnInterval <= 0 || len(w.SpawnRoutes) == 0 {
		return
	}
	w.spawnElapsed += dt
	if w.spawnElapsed < w.SpawnInterval {
		return
	}
	route := w.SpawnRoutes[w.nextSpawnRoute%len(w.SpawnRoutes)]
	if len(route) == 0 {
		return
	}
	if !w.routeOpen(route) {
		var ok bool
		route, ok = w.routeAroundClosure(route, 0)
		if !ok {
			return
		}
	}
	if !w.canSpawn(route) {
		return
	}
	w.addCar(route, 0, 0)
	w.spawnElapsed = 0
	w.nextSpawnRoute = (w.nextSpawnRoute + 1) % len(w.SpawnRoutes)
}

func (w *World) canReroute(route []int) bool {
	if len(route) == 0 || w.Roads[route[0]].Closed {
		return false
	}
	_, ok := w.routeAroundClosure(route, 0)
	return ok
}

func (w *World) routeOpen(route []int) bool {
	for _, roadID := range route {
		segment, ok := w.Roads[roadID]
		if !ok || segment.Closed {
			return false
		}
	}
	return true
}

func (w *World) canSpawn(route []int) bool {
	offset := 0.0
	for _, roadID := range route {
		for _, existing := range w.Cars {
			if !existing.Finished() && existing.Route[existing.RoadIndex] == roadID && offset+existing.Position < car.Length+car.MinGap {
				return false
			}
		}
		offset += w.Roads[roadID].Length
		if offset >= car.Length+car.MinGap {
			break
		}
	}
	return true
}

func (w *World) addCar(route []int, originID, destinationID int) {
	w.Cars = append(w.Cars, car.Car{
		DesiredSpeed: car.DefaultDesiredSpeed, Acceleration: 2, Braking: 4,
		Route: append([]int(nil), route...), OriginID: originID, DestinationID: destinationID,
	})
}
