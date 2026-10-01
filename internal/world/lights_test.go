package world

import (
	"math"
	"testing"

	"github.com/slanalan1203/Cars-sumulation/internal/car"
	"github.com/slanalan1203/Cars-sumulation/internal/road"
	"github.com/slanalan1203/Cars-sumulation/internal/trafficlight"
)

func TestTrafficLightRules(t *testing.T) {
	for _, tc := range []struct {
		name      string
		green     bool
		lightRoad int
		want      float64
	}{
		{"red stops car", false, 1, 30},
		{"green lets car pass", true, 1, 40},
		{"red on another road does not block", false, 2, 40},
	} {
		t.Run(tc.name, func(t *testing.T) {
			w := World{
				Roads:         map[int]road.Road{1: {ID: 1, Length: 100}, 2: {ID: 2, Length: 100}},
				Cars:          []car.Car{{Position: 20, Speed: 10, DesiredSpeed: 10, Acceleration: 2, Braking: 4, Route: []int{1}}},
				TrafficLights: []trafficlight.TrafficLight{{RoadID: tc.lightRoad, Position: 30, Green: tc.green, RedDuration: 100, GreenDuration: 100}},
			}
			for i := 0; i < 20; i++ {
				w.Step(0.1)
				if !tc.green && tc.lightRoad == 1 && w.Cars[0].Position > 30 {
					t.Fatal("crossed red light")
				}
			}
			if math.Abs(w.Cars[0].Position-tc.want) > 0.01 {
				t.Fatalf("position=%v, want %v", w.Cars[0].Position, tc.want)
			}
		})
	}
}

func TestLightSwitchesBeforeCarUpdate(t *testing.T) {
	w := World{
		Roads:         map[int]road.Road{1: {ID: 1, Length: 100}},
		Cars:          []car.Car{{Position: 30, DesiredSpeed: 10, Acceleration: 2, Braking: 4, Route: []int{1}}},
		TrafficLights: []trafficlight.TrafficLight{{RoadID: 1, Position: 30, RedDuration: 1, GreenDuration: 3, Elapsed: 0.95}},
	}
	w.Step(0.1)
	if !w.TrafficLights[0].Green || w.Cars[0].Position <= 30 {
		t.Fatal("car must react to new phase on this step")
	}
}

func TestRedLightOnNextRoadStopsCarBeforeIt(t *testing.T) {
	w := World{
		Roads:         map[int]road.Road{1: {ID: 1, Length: 100}, 2: {ID: 2, Length: 100}},
		Cars:          []car.Car{{Position: 99, Speed: 10, DesiredSpeed: 10, Acceleration: 2, Braking: 4, Route: []int{1, 2}}},
		TrafficLights: []trafficlight.TrafficLight{{RoadID: 2, Position: 0, RedDuration: 100, GreenDuration: 100}},
	}
	for i := 0; i < 20; i++ {
		w.Step(0.1)
	}
	if w.Cars[0].RoadIndex != 1 || w.Cars[0].Position != 0 || w.Cars[0].Speed != 0 {
		t.Fatalf("crossed next road red: %+v", w.Cars[0])
	}
}
