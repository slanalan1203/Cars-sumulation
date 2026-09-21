package world

import (
	"fmt"
	"math"
)

// Validate проверяет сценарий до запуска. Пока поддерживаются маршруты без петель.
func (w *World) Validate() error {
	for id, segment := range w.Roads {
		if segment.ID != id || !finite(segment.Length) || segment.Length <= 0 {
			return fmt.Errorf("дорога %d: нужен совпадающий ID и положительная конечная длина", id)
		}
	}
	for i, current := range w.Cars {
		if err := w.validateRoute(current.Route); err != nil {
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
	for _, junction := range w.Intersections {
		for _, ids := range [][]int{junction.IncomingRoads, junction.OutgoingRoads} {
			for _, id := range ids {
				if _, ok := w.Roads[id]; !ok {
					return fmt.Errorf("перекрёсток %d: нет дороги %d", junction.ID, id)
				}
			}
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
