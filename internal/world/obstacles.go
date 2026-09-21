package world

import (
	"cars-simulation/internal/car"
	"math"
)

// Расстояния измеряются вдоль оставшегося маршрута: до машин,
// красных светофоров и конца маршрута. Маршруты проверяются Validate().
func (w *World) findObstacleAhead(carIndex int) car.Obstacle {
	current := w.Cars[carIndex]
	nearest := math.Inf(1)
	offset := -current.Position
	for _, roadID := range current.Route[current.RoadIndex:] {
		segment, ok := w.Roads[roadID]
		if !ok || segment.Length <= 0 {
			return car.Obstacle{Distance: 0}
		}
		for i, candidate := range w.Cars {
			if i == carIndex || candidate.Finished() || candidate.Route[candidate.RoadIndex] != roadID {
				continue
			}
			frontDistance := offset + candidate.Position
			if frontDistance >= 0 {
				nearest = math.Min(nearest, frontDistance-car.Length-car.MinGap)
			}
		}
		for _, light := range w.TrafficLights {
			if light.RoadID == roadID && !light.Green {
				distance := offset + light.Position
				if distance >= 0 {
					nearest = math.Min(nearest, distance)
				}
			}
		}
		offset += segment.Length
	}
	// Конец маршрута остаётся точкой остановки, как в исходной модели.
	return car.Obstacle{Distance: math.Max(0, math.Min(nearest, offset))}
}
