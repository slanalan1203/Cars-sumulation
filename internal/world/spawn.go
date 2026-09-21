package world

import "cars-simulation/internal/car"

func (w *World) spawnCars(dt float64) {
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
	// Проверяем свободное место вдоль маршрута: лидер может находиться
	// уже на следующей дороге, но его задний бампер ещё перекрывает въезд.
	offset := 0.0
	for _, roadID := range route {
		for _, existing := range w.Cars {
			if !existing.Finished() && existing.Route[existing.RoadIndex] == roadID && offset+existing.Position < car.Length+car.MinGap {
				return
			}
		}
		offset += w.Roads[roadID].Length
		if offset >= car.Length+car.MinGap {
			break
		}
	}
	w.Cars = append(w.Cars, car.Car{
		DesiredSpeed: 10, Acceleration: 2, Braking: 4,
		Route: append([]int(nil), route...),
	})
	w.spawnElapsed = 0
	w.nextSpawnRoute = (w.nextSpawnRoute + 1) % len(w.SpawnRoutes)
}
