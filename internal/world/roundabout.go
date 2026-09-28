package world

import (
	"math"

	"github.com/slanalan1203/Cars-sumulation/internal/car"
	"github.com/slanalan1203/Cars-sumulation/internal/road"
)

type RoundaboutEntry struct {
	ApproachRoadID int
	PreviousRoadID int
	NextRoadID     int
	ExitRoadID     int
}

type Roundabout struct {
	Center   road.Point
	Radius   float64
	Segments []int
	Entries  []RoundaboutEntry
}

func (w *World) roundaboutObstacle(current car.Car) float64 {
	roadID := current.Route[current.RoadIndex]
	nearest := math.Inf(1)
	for _, circle := range w.Roundabouts {
		for _, entry := range circle.Entries {
			if entry.ApproachRoadID != roadID {
				continue
			}
			previous := w.Roads[entry.PreviousRoadID]
			for _, other := range w.Cars {
				if other.Finished() || other.Route[other.RoadIndex] != entry.PreviousRoadID {
					continue
				}
				if previous.Length-other.Position <= 14 {
					nearest = math.Min(nearest, w.Roads[roadID].Length-current.Position-0.5)
				}
			}
		}
	}
	return nearest
}
