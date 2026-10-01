package car

import (
	"math"
	"testing"

	"github.com/slanalan1203/Cars-sumulation/internal/road"
)

func TestAccelerationDoesNotOvershootTargetSpeed(t *testing.T) {
	c := Car{Speed: 9.9, DesiredSpeed: 10, Acceleration: 2, Braking: 4}
	c.UpdateSpeed(c.CalculateAcceleration(Obstacle{Distance: math.Inf(1)}, 0.1), 0.1)
	if math.Abs(c.Speed-10) > 1e-9 {
		t.Fatalf("speed = %v, want 10", c.Speed)
	}
}

func TestBrakingTowardsObstacle(t *testing.T) {
	c := Car{Speed: 10, DesiredSpeed: 10, Acceleration: 2, Braking: 4}
	acceleration := c.CalculateAcceleration(Obstacle{Distance: 2}, 0.1)
	if acceleration != -4 {
		t.Fatalf("acceleration = %v, want -4", acceleration)
	}
	c.UpdateSpeed(acceleration, 0.1)
	if math.Abs(c.Speed-9.6) > 1e-9 {
		t.Fatalf("speed = %v, want 9.6", c.Speed)
	}
}

func TestAdvanceAcrossSeveralRoads(t *testing.T) {
	c := Car{Position: 9, Speed: 10, Route: []int{1, 2, 3}}
	roads := map[int]road.Road{1: {Length: 10}, 2: {Length: 5}, 3: {Length: 20}}
	c.Advance(8, roads)
	if c.RoadIndex != 2 || c.Position != 2 || c.Speed != 10 {
		t.Fatalf("wrong route transition: %+v", c)
	}
	c.Advance(18, roads)
	if !c.Finished() || c.Speed != 0 {
		t.Fatalf("route not finished: %+v", c)
	}
}
