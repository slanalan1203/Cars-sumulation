package world

import (
	"reflect"
	"testing"

	"github.com/slanalan1203/Cars-sumulation/internal/car"
	"github.com/slanalan1203/Cars-sumulation/internal/road"
	"github.com/slanalan1203/Cars-sumulation/internal/trafficlight"
)

func TestSnapshotAndResetOwnTheirSlices(t *testing.T) {
	w := World{
		Roads:         map[int]road.Road{1: {ID: 1, Length: 100, Points: []road.Point{{X: 0}, {X: 100}}}},
		Cars:          []car.Car{{Position: 20, Braking: 4, Route: []int{1}}},
		TrafficLights: []trafficlight.TrafficLight{{RoadID: 1, Position: 30, RedDuration: 3, GreenDuration: 3}},
	}
	w.SaveInitialState()
	expected := w.Snapshot()
	expected.Paused = true
	snapshot := w.Snapshot()
	snapshot.Cars[0].Route[0] = 99
	snapshot.Roads[1].Points[0].X = 99
	if w.Cars[0].Route[0] != 1 || w.Roads[1].Points[0].X != 0 {
		t.Fatal("snapshot shares mutable state")
	}
	for i := 0; i < 2; i++ {
		w.Cars[0].Position = 80
		w.Cars[0].Route[0] = 99
		w.TrafficLights[0].Green = true
		w.spawnElapsed = 1
		w.nextSpawnRoute = 1
		w.Reset()
		if !reflect.DeepEqual(w.Snapshot(), expected) || w.spawnElapsed != 0 || w.nextSpawnRoute != 0 {
			t.Fatal("reset did not restore initial state")
		}
	}
}

func TestPausedWorldDoesNotChange(t *testing.T) {
	w := World{Paused: true, Cars: []car.Car{{Position: 20, Route: []int{1}}}}
	before := w.Snapshot()
	w.Step(1)
	if !reflect.DeepEqual(before, w.Snapshot()) {
		t.Fatal("paused world changed")
	}
}

func TestSpawnChecksAllCarsOnEntryRoad(t *testing.T) {
	w := World{Roads: map[int]road.Road{1: {ID: 1, Length: 100}, 2: {ID: 2, Length: 100}}, SpawnInterval: 1, SpawnRoutes: [][]int{{1, 2}}, Cars: []car.Car{
		{Position: 2, Route: []int{1, 2}},
		{Position: 50, Route: []int{1, 2}, RoadIndex: 1},
	}}
	w.spawnCars(1)
	if len(w.Cars) != 2 {
		t.Fatal("spawned into occupied entry")
	}
	w.Cars[0].Position = 10
	w.spawnCars(0.1)
	if len(w.Cars) != 3 {
		t.Fatal("did not spawn at free entry")
	}
}

func TestSpawnChecksClearanceBeyondShortEntryRoad(t *testing.T) {
	w := World{
		Roads:         map[int]road.Road{1: {ID: 1, Length: 2}, 2: {ID: 2, Length: 100}},
		SpawnInterval: 1,
		SpawnRoutes:   [][]int{{1, 2}},
		Cars:          []car.Car{{Position: 1, Braking: 4, Route: []int{1, 2}, RoadIndex: 1}},
	}
	w.spawnCars(1)
	if len(w.Cars) != 1 {
		t.Fatal("spawn overlaps a leader just beyond the entry road")
	}
	w.Cars[0].Position = 4
	w.spawnCars(0.1)
	if len(w.Cars) != 2 {
		t.Fatal("six meters of front-to-front clearance should allow spawning")
	}
}

func TestValidateRejectsInvalidRoutes(t *testing.T) {
	for _, route := range [][]int{nil, {1, 9}, {1, 1}, {1, 2}} {
		w := World{Roads: map[int]road.Road{
			1: {ID: 1, Length: 10, Points: []road.Point{{X: 0}, {X: 10}}},
			2: {ID: 2, Length: 10, Points: []road.Point{{X: 20}, {X: 30}}},
		}, SpawnRoutes: [][]int{route}}
		if w.Validate() == nil {
			t.Fatalf("accepted invalid route %v", route)
		}
	}
}
