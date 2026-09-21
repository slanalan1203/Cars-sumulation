package road

import "math"

type Point struct {
	X float64
	Y float64
}

type Road struct {
	ID     int
	Length float64
	Points []Point
}

func NewRoad(points ...Point) Road {
	road := Road{
		Points: points,
	}

	for i := 1; i < len(points); i++ {
		dx := points[i].X - points[i-1].X
		dy := points[i].Y - points[i-1].Y
		road.Length += math.Hypot(dx, dy)
	}
	return road
}
