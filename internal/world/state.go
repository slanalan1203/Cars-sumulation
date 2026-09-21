package world

import (
	"cars-simulation/internal/car"
	"cars-simulation/internal/intersection"
	"cars-simulation/internal/road"
	"cars-simulation/internal/trafficlight"
)

// State — независимый снимок для HTTP и восстановления начального состояния.
type State struct {
	Roads         map[int]road.Road           `json:"roads"`
	Cars          []car.Car                   `json:"cars"`
	TrafficLights []trafficlight.TrafficLight `json:"trafficLights"`
	Intersections []intersection.Intersection `json:"intersections"`
	Paused        bool                        `json:"paused"`
}

func (w *World) Snapshot() State {
	return cloneState(State{Roads: w.Roads, Cars: w.Cars, TrafficLights: w.TrafficLights, Intersections: w.Intersections, Paused: w.Paused})
}

// SaveInitialState вызывается один раз до запуска таймера симуляции.
func (w *World) SaveInitialState() {
	state := w.Snapshot()
	w.initial = &state
}

func (w *World) Pause()  { w.Paused = true }
func (w *World) Resume() { w.Paused = false }

func (w *World) Reset() {
	if w.initial == nil {
		return
	}
	state := cloneState(*w.initial)
	w.Cars, w.Roads = state.Cars, state.Roads
	w.TrafficLights, w.Intersections = state.TrafficLights, state.Intersections
	w.Paused = true
	w.spawnElapsed = 0
	w.nextSpawnRoute = 0
}

func cloneState(source State) State {
	result := source
	result.Cars = make([]car.Car, len(source.Cars))
	copy(result.Cars, source.Cars)
	for i := range result.Cars {
		result.Cars[i].Route = append([]int(nil), source.Cars[i].Route...)
	}
	result.TrafficLights = append([]trafficlight.TrafficLight{}, source.TrafficLights...)
	result.Roads = make(map[int]road.Road, len(source.Roads))
	for id, segment := range source.Roads {
		segment.Points = append([]road.Point(nil), segment.Points...)
		result.Roads[id] = segment
	}
	result.Intersections = make([]intersection.Intersection, len(source.Intersections))
	for i, junction := range source.Intersections {
		junction.IncomingRoads = append([]int(nil), junction.IncomingRoads...)
		junction.OutgoingRoads = append([]int(nil), junction.OutgoingRoads...)
		result.Intersections[i] = junction
	}
	return result
}
