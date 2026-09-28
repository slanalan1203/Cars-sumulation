package world

import (
	"math"

	"github.com/slanalan1203/Cars-sumulation/internal/car"
)

func (w *World) speedTarget(current car.Car, dt float64) float64 {
	target := current.DesiredSpeed
	distance := 0.0
	for index := current.RoadIndex; index < len(current.Route); index++ {
		segment := w.Roads[current.Route[index]]
		if segment.SpeedLimit > 0 {
			lookahead := distance
			if index > current.RoadIndex {
				lookahead = math.Max(0, distance-current.DesiredSpeed*dt)
			}
			allowed := math.Sqrt(segment.SpeedLimit*segment.SpeedLimit + 2*current.Braking*lookahead)
			target = math.Min(target, allowed)
		}
		if index == current.RoadIndex {
			distance += math.Max(0, segment.Length-current.Position)
		} else {
			distance += segment.Length
		}
	}
	return target
}
