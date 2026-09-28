package world

import (
	"fmt"
	"math"

	"github.com/slanalan1203/Cars-sumulation/internal/road"
)

func (w *World) SetRoadSettings(roadID int, speedKmh *float64, closed *bool) error {
	segment, ok := w.Roads[roadID]
	if !ok || segment.Kind == "access" {
		return fmt.Errorf("участок дороги %d недоступен для настройки", roadID)
	}
	if speedKmh == nil && closed == nil {
		return fmt.Errorf("не указана настройка дороги")
	}
	if speedKmh != nil && (math.IsNaN(*speedKmh) || math.IsInf(*speedKmh, 0) || *speedKmh < 5 || *speedKmh > 120) {
		return fmt.Errorf("скорость должна быть от 5 до 120 км/ч")
	}
	for id, candidate := range w.Roads {
		if id != roadID && !reverseRoads(segment, candidate) {
			continue
		}
		if speedKmh != nil {
			candidate.SpeedLimit = *speedKmh / 3.6
		}
		if closed != nil {
			candidate.Closed = *closed
		}
		w.Roads[id] = candidate
	}
	if closed != nil && *closed {
		w.rerouteCars()
	}
	return nil
}

func reverseRoads(first, second road.Road) bool {
	if first.Kind != second.Kind || first.Level != second.Level || len(first.Points) != len(second.Points) {
		return false
	}
	for i, point := range first.Points {
		if point != second.Points[len(second.Points)-1-i] {
			return false
		}
	}
	return true
}
