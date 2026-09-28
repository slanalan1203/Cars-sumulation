package world

import (
	"math"

	"github.com/slanalan1203/Cars-sumulation/internal/car"
)

func (w *World) findObstacleAhead(carIndex int) car.Obstacle {
	current := w.Cars[carIndex]
	nearest := math.Inf(1)
	offset := -current.Position
	for routeIndex, roadID := range current.Route[current.RoadIndex:] {
		segment, ok := w.Roads[roadID]
		if !ok || segment.Length <= 0 {
			return car.Obstacle{Distance: 0}
		}
		if routeIndex > 0 && segment.Closed {
			nearest = math.Min(nearest, offset-0.5)
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
	nearest = math.Min(nearest, w.roundaboutObstacle(current))
	return car.Obstacle{Distance: math.Max(0, math.Min(nearest, offset))}
}
