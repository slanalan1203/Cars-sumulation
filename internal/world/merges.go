package world

import (
	"math"
	"slices"
	"sort"

	"github.com/slanalan1203/Cars-sumulation/internal/car"
	"github.com/slanalan1203/Cars-sumulation/internal/intersection"
	"github.com/slanalan1203/Cars-sumulation/internal/road"
)

const IntersectionZoneLength = 12.0

func (w *World) resolveIntersectionConflicts(next []car.Car) {
	for _, junction := range w.Intersections {
		active := make([]intersectionMovement, 0)
		for _, current := range w.Cars {
			if w.inIntersectionZone(current, junction) {
				active = append(active, w.movementAtIntersection(current, junction))
			}
		}

		candidates := make([]int, 0)
		for i, before := range w.Cars {
			if before.Finished() || w.inIntersectionZone(before, junction) {
				continue
			}
			if w.inIntersectionZone(next[i], junction) || w.crossedIntersection(before, next[i], junction) {
				candidates = append(candidates, i)
			}
		}
		lastRoad := w.lastIntersectionRoad[junction.ID]
		sort.SliceStable(candidates, func(i, j int) bool {
			first := w.movementAtIntersection(w.Cars[candidates[i]], junction)
			second := w.movementAtIntersection(w.Cars[candidates[j]], junction)
			return incomingPriority(junction, lastRoad, first.incoming) < incomingPriority(junction, lastRoad, second.incoming)
		})
		for _, i := range candidates {
			before := w.Cars[i]
			movement := w.movementAtIntersection(before, junction)
			blocked := false
			for _, other := range active {
				if movementsConflict(movement, other) {
					blocked = true
					break
				}
			}
			if blocked {
				next[i] = before
				next[i].Speed = 0
				continue
			}
			active = append(active, movement)
			if w.lastIntersectionRoad == nil {
				w.lastIntersectionRoad = make(map[int]int)
			}
			w.lastIntersectionRoad[junction.ID] = movement.incoming
		}
	}
}

type intersectionMovement struct {
	incoming int
	outgoing int
	path     []road.Point
	valid    bool
}

func (w *World) movementAtIntersection(current car.Car, junction intersection.Intersection) intersectionMovement {
	if current.Finished() {
		return intersectionMovement{}
	}
	index := current.RoadIndex
	roadID := current.Route[index]
	incoming, outgoing := 0, 0
	if slices.Contains(junction.IncomingRoads, roadID) && index+1 < len(current.Route) {
		incoming, outgoing = roadID, current.Route[index+1]
	} else if slices.Contains(junction.OutgoingRoads, roadID) && index > 0 {
		incoming, outgoing = current.Route[index-1], roadID
	}
	if !junction.Allows(incoming, outgoing) {
		return intersectionMovement{}
	}
	inRoad, inOK := w.Roads[incoming]
	outRoad, outOK := w.Roads[outgoing]
	if !inOK || !outOK || len(inRoad.Points) < 2 || len(outRoad.Points) < 2 {
		return intersectionMovement{incoming: incoming, outgoing: outgoing, valid: true}
	}
	inEnd := inRoad.Points[len(inRoad.Points)-1]
	inBefore := inRoad.Points[len(inRoad.Points)-2]
	outStart := outRoad.Points[0]
	outAfter := outRoad.Points[1]
	inX, inY := inEnd.X-inBefore.X, inEnd.Y-inBefore.Y
	outX, outY := outAfter.X-outStart.X, outAfter.Y-outStart.Y
	inLength, outLength := math.Hypot(inX, inY), math.Hypot(outX, outY)
	movement := intersectionMovement{incoming: incoming, outgoing: outgoing, valid: true}
	if inLength == 0 || outLength == 0 || inRoad.Length < IntersectionZoneLength || outRoad.Length < IntersectionZoneLength {
		return movement
	}
	inX, inY = inX/inLength, inY/inLength
	outX, outY = outX/outLength, outY/outLength
	dot := inX*outX + inY*outY
	cross := inX*outY - inY*outX
	if dot < -0.99 {
		return movement
	}
	const laneOffset = 2.35
	start := road.Point{X: inEnd.X - inX*IntersectionZoneLength - inY*laneOffset, Y: inEnd.Y - inY*IntersectionZoneLength + inX*laneOffset}
	end := road.Point{X: outStart.X + outX*IntersectionZoneLength - outY*laneOffset, Y: outStart.Y + outY*IntersectionZoneLength + outX*laneOffset}
	center := road.Point{X: inEnd.X, Y: inEnd.Y}
	control := road.Point{X: center.X - (inY+outY)*laneOffset, Y: center.Y + (inX+outX)*laneOffset}
	if dot > 0.99 {
		center.X -= inY * laneOffset
		center.Y += inX * laneOffset
	}
	for step := 0; step <= 24; step++ {
		t := float64(step) / 24
		var point road.Point
		if cross > 0.99 {
			point.X = (1-t)*(1-t)*start.X + 2*(1-t)*t*control.X + t*t*end.X
			point.Y = (1-t)*(1-t)*start.Y + 2*(1-t)*t*control.Y + t*t*end.Y
		} else if t <= 0.5 {
			point.X = start.X + (center.X-start.X)*t*2
			point.Y = start.Y + (center.Y-start.Y)*t*2
		} else {
			point.X = center.X + (end.X-center.X)*(t-0.5)*2
			point.Y = center.Y + (end.Y-center.Y)*(t-0.5)*2
		}
		movement.path = append(movement.path, point)
	}
	return movement
}

func movementsConflict(first, second intersectionMovement) bool {
	if !first.valid || !second.valid {
		return true
	}
	if first.incoming == second.incoming && first.outgoing == second.outgoing {
		return false
	}
	if len(first.path) == 0 || len(second.path) == 0 {
		return true
	}
	const clearance = 3.2
	for _, a := range first.path {
		for _, b := range second.path {
			if math.Hypot(a.X-b.X, a.Y-b.Y) < clearance {
				return true
			}
		}
	}
	return false
}

func incomingPriority(junction intersection.Intersection, lastRoad, roadID int) int {
	index := slices.Index(junction.IncomingRoads, roadID)
	lastIndex := slices.Index(junction.IncomingRoads, lastRoad)
	if index < 0 {
		return len(junction.IncomingRoads)
	}
	if lastIndex < 0 {
		return index
	}
	return (index - lastIndex - 1 + len(junction.IncomingRoads)) % len(junction.IncomingRoads)
}

func (w *World) inIntersectionZone(current car.Car, junction intersection.Intersection) bool {
	if current.Finished() {
		return false
	}
	roadID := current.Route[current.RoadIndex]
	if slices.Contains(junction.IncomingRoads, roadID) && w.Roads[roadID].Length-current.Position < IntersectionZoneLength {
		return true
	}
	return slices.Contains(junction.OutgoingRoads, roadID) && current.Position < IntersectionZoneLength
}

func (w *World) crossedIntersection(before, after car.Car, junction intersection.Intersection) bool {
	if before.Finished() || after.Finished() || before.RoadIndex == after.RoadIndex {
		return false
	}
	return junction.Allows(before.Route[before.RoadIndex], after.Route[after.RoadIndex])
}
