package world

import (
	"math"

	"github.com/slanalan1203/Cars-sumulation/internal/road"
)

func (w *World) routeAroundClosure(route []int, currentIndex int) ([]int, bool) {
	if currentIndex < 0 || currentIndex >= len(route) || len(route) < 2 {
		return nil, false
	}
	for _, id := range route[currentIndex+1:] {
		if segment, ok := w.Roads[id]; !ok || segment.Closed {
			goto search
		}
	}
	return route, true
search:
	current := w.Roads[route[currentIndex]]
	destination := w.Roads[route[len(route)-1]]
	if len(current.Points) == 0 || len(destination.Points) == 0 || destination.Closed {
		return nil, false
	}
	start := current.Points[len(current.Points)-1]
	goal := destination.Points[0]
	blocked := make(map[int]bool)
	for _, id := range route[:currentIndex+1] {
		blocked[id] = true
	}
	distance := map[road.Point]float64{start: 0}
	previous := make(map[road.Point]road.Point)
	previousRoad := make(map[road.Point]int)
	visited := make(map[road.Point]bool)
	for {
		best := math.Inf(1)
		var point road.Point
		found := false
		for candidate, cost := range distance {
			if !visited[candidate] && cost < best {
				point, best, found = candidate, cost, true
			}
		}
		if !found || point == goal {
			break
		}
		visited[point] = true
		for id, segment := range w.Roads {
			if blocked[id] || segment.Closed || id == destination.ID || segment.Kind == "access" || len(segment.Points) == 0 || segment.Points[0] != point || segment.SpeedLimit <= 0 {
				continue
			}
			end := segment.Points[len(segment.Points)-1]
			cost := best + segment.Length/segment.SpeedLimit
			if old, ok := distance[end]; !ok || cost < old {
				distance[end], previous[end], previousRoad[end] = cost, point, id
			}
		}
	}
	if _, ok := distance[goal]; !ok {
		return nil, false
	}
	middle := []int{}
	for point := goal; point != start; point = previous[point] {
		middle = append(middle, previousRoad[point])
	}
	result := append([]int(nil), route[:currentIndex+1]...)
	for i := len(middle) - 1; i >= 0; i-- {
		result = append(result, middle[i])
	}
	result = append(result, destination.ID)
	return result, true
}

func (w *World) rerouteCars() {
	for i := range w.Cars {
		current := &w.Cars[i]
		if current.Finished() {
			continue
		}
		if route, ok := w.routeAroundClosure(current.Route, current.RoadIndex); ok {
			current.Route = route
		}
	}
}
