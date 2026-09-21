package world

import (
	"cars-simulation/internal/car"
	"cars-simulation/internal/road"
	"math"
	"testing"
)

func TestSimulationMovesToNextRoad(t *testing.T) {
	sim := World{
		Roads: map[int]road.Road{
			1: {ID: 1, Length: 100},
			2: {ID: 2, Length: 50},
		},
		Cars: []car.Car{
			{
				Position:     99,
				Speed:        10,
				DesiredSpeed: 10,
				Acceleration: 2,
				Braking:      4,
				Route:        []int{1, 2},
				RoadIndex:    0,
			},
		},
	}

	sim.Step(0.2)

	if len(sim.Cars) != 1 {
		t.Fatalf("после перехода должна остаться одна машина, получено %d", len(sim.Cars))
	}

	car := sim.Cars[0]

	if car.RoadIndex != 1 {
		t.Errorf("индекс дороги: получено %d, ожидается 1", car.RoadIndex)
	}
	if math.Abs(car.Position-1) > 0.01 {
		t.Errorf("позиция на второй дороге: получено %.3f, ожидается 1", car.Position)
	}
	if math.Abs(car.Speed-10) > 0.01 {
		t.Errorf("скорость после перехода: получено %.3f, ожидается 10", car.Speed)
	}
}

func TestSimulationFollowsCarAcrossRoadBoundary(t *testing.T) {
	tests := []struct {
		name             string
		followerPosition float64
		leaderPosition   float64
		wantPosition     float64
	}{
		{
			name:             "далекая машина не мешает движению",
			followerPosition: 20,
			leaderPosition:   10,
			wantPosition:     21,
		},
		{
			name:             "занятый въезд останавливает машину",
			followerPosition: 99,
			leaderPosition:   5,
			wantPosition:     99,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sim := World{
				Roads: map[int]road.Road{
					1: {ID: 1, Length: 100},
					2: {ID: 2, Length: 50},
				},
				Cars: []car.Car{
					{
						Position:     tt.leaderPosition,
						DesiredSpeed: 0,
						Acceleration: 2,
						Braking:      4,
						Route:        []int{1, 2},
						RoadIndex:    1,
					},
					{
						Position:     tt.followerPosition,
						Speed:        10,
						DesiredSpeed: 10,
						Acceleration: 2,
						Braking:      4,
						Route:        []int{1, 2},
						RoadIndex:    0,
					},
				},
			}

			sim.Step(0.1)

			if len(sim.Cars) != 2 {
				t.Fatalf("ожидались две машины, получено %d", len(sim.Cars))
			}

			got := sim.Cars[1].Position
			if math.Abs(got-tt.wantPosition) > 0.01 {
				t.Errorf("позиция: получено %.3f, ожидается %.3f",
					got, tt.wantPosition)
			}
		})
	}
}
