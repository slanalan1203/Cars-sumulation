package world

import (
	"fmt"
	"math"
)

func (w *World) Validate() error {
	for id, segment := range w.Roads {
		if segment.ID != id || !finite(segment.Length) || segment.Length <= 0 {
			return fmt.Errorf("дорога %d: нужен совпадающий ID и положительная конечная длина", id)
		}
		if segment.Level < 0 || !finite(segment.SpeedLimit) || segment.SpeedLimit < 0 {
			return fmt.Errorf("дорога %d: неверный уровень или предел скорости", id)
		}
	}
	buildings := make(map[int]Building, len(w.Buildings))
	for _, building := range w.Buildings {
		if building.ID <= 0 || building.Name == "" || (building.Kind != BuildingHome && building.Kind != BuildingWork) || !finite(building.Position.X) || !finite(building.Position.Y) {
			return fmt.Errorf("здание %d: неверные параметры", building.ID)
		}
		if _, exists := buildings[building.ID]; exists {
			return fmt.Errorf("здание %d: повтор ID", building.ID)
		}
		attachedRoad, exists := w.Roads[building.RoadID]
		if !exists {
			return fmt.Errorf("здание %d: нет дороги %d", building.ID, building.RoadID)
		}
		if attachedRoad.Kind != "access" || len(attachedRoad.Points) < 2 {
			return fmt.Errorf("здание %d: нужна подъездная дорога", building.ID)
		}
		entrance := attachedRoad.Points[0]
		if building.Kind == BuildingWork {
			entrance = attachedRoad.Points[len(attachedRoad.Points)-1]
		}
		if math.Hypot(entrance.X-building.Position.X, entrance.Y-building.Position.Y) > 1e-6 {
			return fmt.Errorf("здание %d: вход не совпадает с подъездной дорогой", building.ID)
		}
		buildings[building.ID] = building
	}
	for id, count := range w.Arrivals {
		if building, ok := buildings[id]; !ok || building.Kind != BuildingWork || count < 0 {
			return fmt.Errorf("прибытия: неверное рабочее место %d", id)
		}
	}
	for i, current := range w.Cars {
		if err := w.validateRoute(current.Route); err != nil {
			return fmt.Errorf("машина %d: %w", i, err)
		}
		if err := validateTrip(current.Route, current.OriginID, current.DestinationID, buildings); err != nil {
			return fmt.Errorf("машина %d: %w", i, err)
		}
		if current.Finished() {
			return fmt.Errorf("машина %d: неверный RoadIndex", i)
		}
		if !finite(current.Position) || current.Position < 0 || current.Position >= w.Roads[current.Route[current.RoadIndex]].Length {
			return fmt.Errorf("машина %d: позиция вне дороги", i)
		}
		if !finite(current.Speed) || !finite(current.DesiredSpeed) || !finite(current.Acceleration) || !finite(current.Braking) || current.Speed < 0 || current.DesiredSpeed < 0 || current.Acceleration < 0 || current.Braking <= 0 {
			return fmt.Errorf("машина %d: неверные параметры движения", i)
		}
	}
	for i, light := range w.TrafficLights {
		segment, ok := w.Roads[light.RoadID]
		if !ok || !finite(light.Position) || light.Position < 0 || light.Position > segment.Length || !finite(light.RedDuration) || !finite(light.GreenDuration) || light.RedDuration <= 0 || light.GreenDuration <= 0 || !finite(light.Elapsed) || light.Elapsed < 0 {
			return fmt.Errorf("светофор %d: проверь дорогу, позицию и длительности фаз", i)
		}
	}
	if !finite(w.SpawnInterval) || w.SpawnInterval < 0 {
		return fmt.Errorf("неверный интервал появления машин")
	}
	for _, route := range w.SpawnRoutes {
		if err := w.validateRoute(route); err != nil {
			return fmt.Errorf("маршрут генератора: %w", err)
		}
	}
	spawnRoads := make(map[int]bool)
	for _, source := range w.SpawnSources {
		if _, ok := w.Roads[source.RoadID]; !ok {
			return fmt.Errorf("источник машин: нет дороги %d", source.RoadID)
		}
		if spawnRoads[source.RoadID] {
			return fmt.Errorf("дорога %d: несколько источников машин", source.RoadID)
		}
		spawnRoads[source.RoadID] = true
		if !finite(source.Interval) || source.Interval < 0 || len(source.Routes) == 0 {
			return fmt.Errorf("дорога %d: неверные настройки появления машин", source.RoadID)
		}
		if len(source.DestinationIDs) != 0 && len(source.DestinationIDs) != len(source.Routes) {
			return fmt.Errorf("дорога %d: число рабочих мест не совпадает с числом маршрутов", source.RoadID)
		}
		for i, route := range source.Routes {
			if err := w.validateRoute(route); err != nil {
				return fmt.Errorf("дорога %d: маршрут генератора: %w", source.RoadID, err)
			}
			if route[0] != source.RoadID {
				return fmt.Errorf("дорога %d: маршрут начинается с дороги %d", source.RoadID, route[0])
			}
			destinationID := 0
			if len(source.DestinationIDs) > 0 {
				destinationID = source.DestinationIDs[i]
			}
			if err := validateTrip(route, source.HomeID, destinationID, buildings); err != nil {
				return fmt.Errorf("дорога %d: %w", source.RoadID, err)
			}
		}
	}
	for _, junction := range w.Intersections {
		for _, ids := range [][]int{junction.IncomingRoads, junction.OutgoingRoads} {
			for _, id := range ids {
				if _, ok := w.Roads[id]; !ok {
					return fmt.Errorf("перекрёсток %d: нет дороги %d", junction.ID, id)
				}
			}
		}
	}
	for i, circle := range w.Roundabouts {
		if !finite(circle.Center.X) || !finite(circle.Center.Y) || !finite(circle.Radius) || circle.Radius <= 0 || len(circle.Segments) < 2 {
			return fmt.Errorf("круг %d: неверная геометрия", i)
		}
		segments := make(map[int]bool)
		for _, id := range circle.Segments {
			if _, ok := w.Roads[id]; !ok || segments[id] {
				return fmt.Errorf("круг %d: неверный участок %d", i, id)
			}
			segments[id] = true
		}
		for _, entry := range circle.Entries {
			if _, ok := w.Roads[entry.ApproachRoadID]; !ok || !segments[entry.PreviousRoadID] || !segments[entry.NextRoadID] {
				return fmt.Errorf("круг %d: неверный въезд", i)
			}
			previous := w.Roads[entry.PreviousRoadID]
			next := w.Roads[entry.NextRoadID]
			approach := w.Roads[entry.ApproachRoadID]
			exit, ok := w.Roads[entry.ExitRoadID]
			if !ok || len(previous.Points) == 0 || len(next.Points) == 0 || len(approach.Points) == 0 || len(exit.Points) == 0 || math.Hypot(previous.Points[len(previous.Points)-1].X-next.Points[0].X, previous.Points[len(previous.Points)-1].Y-next.Points[0].Y) > 1e-6 || math.Hypot(approach.Points[len(approach.Points)-1].X-next.Points[0].X, approach.Points[len(approach.Points)-1].Y-next.Points[0].Y) > 1e-6 || math.Hypot(exit.Points[0].X-next.Points[0].X, exit.Points[0].Y-next.Points[0].Y) > 1e-6 {
				return fmt.Errorf("круг %d: дороги въезда не соединены", i)
			}
		}
	}
	return nil
}

func validateTrip(route []int, originID, destinationID int, buildings map[int]Building) error {
	if originID != 0 {
		home, ok := buildings[originID]
		if !ok || home.Kind != BuildingHome || home.RoadID != route[0] {
			return fmt.Errorf("неверный дом %d для маршрута", originID)
		}
	}
	if destinationID != 0 {
		work, ok := buildings[destinationID]
		if !ok || work.Kind != BuildingWork || work.RoadID != route[len(route)-1] {
			return fmt.Errorf("неверное рабочее место %d для маршрута", destinationID)
		}
	}
	return nil
}

func (w *World) validateRoute(route []int) error {
	if len(route) == 0 {
		return fmt.Errorf("пустой маршрут")
	}
	seen := make(map[int]bool)
	for i, id := range route {
		segment, ok := w.Roads[id]
		if !ok {
			return fmt.Errorf("нет дороги %d", id)
		}
		if seen[id] {
			return fmt.Errorf("повтор дороги %d: маршруты с петлями пока не поддерживаются", id)
		}
		seen[id] = true
		if i > 0 {
			previous := w.Roads[route[i-1]]
			if len(previous.Points) > 0 && len(segment.Points) > 0 {
				end, start := previous.Points[len(previous.Points)-1], segment.Points[0]
				if math.Hypot(end.X-start.X, end.Y-start.Y) > 1e-6 {
					return fmt.Errorf("дороги %d и %d не соединены", previous.ID, id)
				}
			}
		}
	}
	return nil
}

func finite(value float64) bool { return !math.IsNaN(value) && !math.IsInf(value, 0) }
