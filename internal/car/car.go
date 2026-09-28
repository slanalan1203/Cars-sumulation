package car

import (
	"math"

	"github.com/slanalan1203/Cars-sumulation/internal/road"
)

const Length = 4.0
const MinGap = 2.0
const DefaultDesiredSpeed = 25.0

type Car struct {
	OriginID      int
	DestinationID int
	Position      float64
	Speed         float64
	DesiredSpeed  float64
	Acceleration  float64
	Braking       float64
	Route         []int
	RoadIndex     int
	TripTime      float64
	WaitTime      float64
}

func (c Car) CalculateAcceleration(obstacle Obstacle, dt float64) float64 {
	return c.CalculateAccelerationForSpeed(obstacle, dt, c.DesiredSpeed)
}

func (c Car) CalculateAccelerationForSpeed(obstacle Obstacle, dt, targetSpeed float64) float64 {
	if dt <= 0 {
		return 0
	}
	target := math.Min(c.DesiredSpeed, targetSpeed)
	if !math.IsInf(obstacle.Distance, 1) {
		target = math.Min(target, math.Sqrt(2*c.Braking*math.Max(0, obstacle.Distance)))
	}
	needed := (target - c.Speed) / dt
	return math.Max(-c.Braking, math.Min(c.Acceleration, needed))
}

func (c *Car) UpdateSpeed(acceleration, dt float64) {
	if dt > 0 {
		c.Speed = math.Max(0, math.Min(c.DesiredSpeed, c.Speed+acceleration*dt))
	}
}

func (c *Car) Advance(distance float64, roads map[int]road.Road) {
	if distance < 0 || math.IsNaN(distance) || math.IsInf(distance, 0) {
		return
	}
	c.Position += distance
	for !c.Finished() {
		current, ok := roads[c.Route[c.RoadIndex]]
		if !ok || current.Length <= 0 {
			c.Speed = 0
			return
		}
		if c.Position < current.Length {
			return
		}
		c.Position -= current.Length
		c.RoadIndex++
	}
	c.Speed = 0
}

func (c Car) Finished() bool {
	return c.RoadIndex < 0 || c.RoadIndex >= len(c.Route)
}
