package world

import (
	"testing"

	"github.com/slanalan1203/Cars-sumulation/internal/car"
	"github.com/slanalan1203/Cars-sumulation/internal/road"
)

func TestCarsOnDifferentBranchesDoNotBlockEachOther(t *testing.T) {
	w := World{
		Roads: map[int]road.Road{1: {ID: 1, Length: 100}, 2: {ID: 2, Length: 100}, 3: {ID: 3, Length: 100}},
		Cars: []car.Car{
			{Position: 10, Braking: 4, Route: []int{1, 2}, RoadIndex: 1},
			{Position: 20, Speed: 10, DesiredSpeed: 10, Acceleration: 2, Braking: 4, Route: []int{1, 3}, RoadIndex: 1},
		},
	}
	w.Step(0.1)
	if got := w.Cars[1].Position; got != 21 {
		t.Fatalf("car on another branch must move freely: got %v, want 21", got)
	}
}

func TestLeaderSearchDoesNotDependOnSliceOrder(t *testing.T) {
	w := World{
		Roads: map[int]road.Road{1: {ID: 1, Length: 100}},
		Cars: []car.Car{
			{Position: 20, Speed: 10, DesiredSpeed: 10, Acceleration: 2, Braking: 4, Route: []int{1}},
			{Position: 26, Braking: 4, Route: []int{1}},
		},
	}
	w.Step(0.1)
	if got := w.Cars[0].Position; got != 20 {
		t.Fatalf("follower must stop behind leader even if stored first: got %v", got)
	}
}
