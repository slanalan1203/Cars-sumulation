package world

import (
	"github.com/slanalan1203/Cars-sumulation/internal/car"
	"github.com/slanalan1203/Cars-sumulation/internal/intersection"
	"github.com/slanalan1203/Cars-sumulation/internal/road"
	"github.com/slanalan1203/Cars-sumulation/internal/trafficlight"
)

type State struct {
	Roads         map[int]road.Road           `json:"roads"`
	Cars          []car.Car                   `json:"cars"`
	TrafficLights []trafficlight.TrafficLight `json:"trafficLights"`
	Intersections []intersection.Intersection `json:"intersections"`
	Roundabouts   []Roundabout                `json:"roundabouts"`
	Buildings     []Building                  `json:"buildings"`
	SpawnSources  []SpawnSourceSetting        `json:"spawnSources"`
	Arrivals      map[int]int                 `json:"arrivals"`
	Paused        bool                        `json:"paused"`
	CustomMap     bool                        `json:"customMap"`
	Metrics       Metrics                     `json:"metrics"`
}

type RoadLoad struct {
	Cars    int `json:"cars"`
	Waiting int `json:"waiting"`
}

type Metrics struct {
	CompletedTrips    int              `json:"completedTrips"`
	AverageTravelTime float64          `json:"averageTravelTime"`
	AverageWaitTime   float64          `json:"averageWaitTime"`
	RoadLoads         map[int]RoadLoad `json:"roadLoads"`
}

type SpawnSourceSetting struct {
	HomeID   int     `json:"homeID"`
	Interval float64 `json:"interval"`
}

func (w *World) Snapshot() State {
	metrics := Metrics{CompletedTrips: w.CompletedTrips, RoadLoads: make(map[int]RoadLoad)}
	if w.CompletedTrips > 0 {
		metrics.AverageTravelTime = w.TotalTravelTime / float64(w.CompletedTrips)
		metrics.AverageWaitTime = w.TotalWaitTime / float64(w.CompletedTrips)
	}
	for _, current := range w.Cars {
		if current.Finished() {
			continue
		}
		id := current.Route[current.RoadIndex]
		load := metrics.RoadLoads[id]
		load.Cars++
		if current.Speed < 0.5 {
			load.Waiting++
		}
		metrics.RoadLoads[id] = load
	}
	sources := make([]SpawnSourceSetting, len(w.SpawnSources))
	for i, source := range w.SpawnSources {
		sources[i] = SpawnSourceSetting{HomeID: source.HomeID, Interval: source.Interval}
	}
	return cloneState(State{Roads: w.Roads, Cars: w.Cars, TrafficLights: w.TrafficLights, Intersections: w.Intersections, Roundabouts: w.Roundabouts, Buildings: w.Buildings, SpawnSources: sources, Arrivals: w.Arrivals, Paused: w.Paused, CustomMap: w.CustomMap, Metrics: metrics})
}

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
	w.Roundabouts = state.Roundabouts
	w.Buildings, w.Arrivals = state.Buildings, state.Arrivals
	w.CustomMap = state.CustomMap
	w.CompletedTrips = state.Metrics.CompletedTrips
	w.TotalTravelTime = state.Metrics.AverageTravelTime * float64(w.CompletedTrips)
	w.TotalWaitTime = state.Metrics.AverageWaitTime * float64(w.CompletedTrips)
	for i := range w.SpawnSources {
		for _, source := range state.SpawnSources {
			if source.HomeID == w.SpawnSources[i].HomeID {
				w.SpawnSources[i].Interval = source.Interval
				break
			}
		}
	}
	w.Paused = true
	w.spawnElapsed = 0
	w.nextSpawnRoute = 0
	w.lastIntersectionRoad = nil
	for i := range w.SpawnSources {
		w.SpawnSources[i].elapsed = 0
		w.SpawnSources[i].nextRoute = 0
	}
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
	result.Roundabouts = make([]Roundabout, len(source.Roundabouts))
	for i, circle := range source.Roundabouts {
		circle.Segments = append([]int(nil), circle.Segments...)
		circle.Entries = append([]RoundaboutEntry(nil), circle.Entries...)
		result.Roundabouts[i] = circle
	}
	result.Buildings = append([]Building(nil), source.Buildings...)
	result.SpawnSources = append([]SpawnSourceSetting(nil), source.SpawnSources...)
	result.Arrivals = make(map[int]int, len(source.Arrivals))
	for id, count := range source.Arrivals {
		result.Arrivals[id] = count
	}
	result.Metrics.RoadLoads = make(map[int]RoadLoad, len(source.Metrics.RoadLoads))
	for id, load := range source.Metrics.RoadLoads {
		result.Metrics.RoadLoads[id] = load
	}
	return result
}
