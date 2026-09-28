package road

import "math"

type Point struct {
	X float64
	Y float64
}

type Road struct {
	ID         int
	Kind       string
	Level      int
	SpeedLimit float64
	Closed     bool
	OneWay     bool
	Length     float64
	Points     []Point
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

func (r Road) Reverse(id int) Road {
	points := make([]Point, len(r.Points))
	for i := range r.Points {
		points[i] = r.Points[len(r.Points)-1-i]
	}
	return Road{ID: id, Kind: r.Kind, Level: r.Level, SpeedLimit: r.SpeedLimit, Closed: r.Closed, OneWay: r.OneWay, Length: r.Length, Points: points}
}
