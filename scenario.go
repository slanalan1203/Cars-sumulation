package main

import (
	"math"

	"github.com/slanalan1203/Cars-sumulation/internal/car"
	"github.com/slanalan1203/Cars-sumulation/internal/intersection"
	"github.com/slanalan1203/Cars-sumulation/internal/road"
	"github.com/slanalan1203/Cars-sumulation/internal/trafficlight"
	"github.com/slanalan1203/Cars-sumulation/internal/world"
)

func newScenario() world.World {
	center := road.Point{X: 100, Y: 0}
	const radius = 24.0
	point := func(angle float64) road.Point {
		return road.Point{X: center.X + radius*math.Cos(angle), Y: center.Y + radius*math.Sin(angle)}
	}
	east, south, diagonal, west := point(0), point(math.Pi/2), point(3*math.Pi/4), point(math.Pi)
	arc := func(start, end float64) road.Road {
		steps := int(math.Ceil((end - start) / 0.12))
		points := make([]road.Point, steps+1)
		for i := range points {
			points[i] = point(start + (end-start)*float64(i)/float64(steps))
		}
		return road.NewRoad(points...)
	}

	roads := make(map[int]road.Road)
	add := func(id int, segment road.Road) {
		segment.ID = id
		if segment.SpeedLimit == 0 {
			segment.SpeedLimit = 10
		}
		roads[id] = segment
	}
	add(1, road.NewRoad(road.Point{X: 0, Y: 0}, west))
	add(2, road.NewRoad(south, road.Point{X: 100, Y: 100}))
	add(3, road.NewRoad(east, road.Point{X: 185, Y: 0}))
	add(4, road.NewRoad(road.Point{X: 0, Y: 100}, diagonal))
	for _, pair := range [][2]int{{1, 5}, {2, 6}, {3, 7}, {4, 8}} {
		add(pair[1], roads[pair[0]].Reverse(pair[1]))
	}
	add(9, arc(0, math.Pi/2))
	add(10, arc(math.Pi/2, 3*math.Pi/4))
	add(11, arc(3*math.Pi/4, math.Pi))
	add(12, arc(math.Pi, 2*math.Pi))
	add(13, road.NewRoad(road.Point{X: 185, Y: 0}, road.Point{X: 280, Y: 0}))
	add(14, roads[13].Reverse(14))
	add(15, road.NewRoad(road.Point{X: 185, Y: -90}, road.Point{X: 185, Y: 0}))
	add(16, roads[15].Reverse(16))
	add(17, road.NewRoad(road.Point{X: 185, Y: 0}, road.Point{X: 185, Y: 90}))
	add(18, roads[17].Reverse(18))
	westRamp := road.NewRoad(road.Point{X: 0, Y: 0}, road.Point{X: 35, Y: -48})
	westRamp.Kind, westRamp.Level, westRamp.SpeedLimit = "ramp", 1, 12
	add(31, westRamp)
	eastbound := road.NewRoad(road.Point{X: 35, Y: -48}, road.Point{X: 245, Y: -48})
	eastbound.Kind, eastbound.Level, eastbound.SpeedLimit = "highway", 1, 80.0/3.6
	add(32, eastbound)
	eastRamp := road.NewRoad(road.Point{X: 245, Y: -48}, road.Point{X: 280, Y: 0})
	eastRamp.Kind, eastRamp.Level, eastRamp.SpeedLimit = "ramp", 1, 12
	add(33, eastRamp)
	add(34, roads[33].Reverse(34))
	add(35, roads[32].Reverse(35))
	add(36, roads[31].Reverse(36))

	const horizontalGreen = 10.0
	const verticalGreen = 10.0
	lights := []trafficlight.TrafficLight{
		{RoadID: 3, Position: roads[3].Length - world.IntersectionZoneLength, RedDuration: verticalGreen, GreenDuration: horizontalGreen},
		{RoadID: 14, Position: roads[14].Length - world.IntersectionZoneLength, RedDuration: verticalGreen, GreenDuration: horizontalGreen},
		{RoadID: 15, Position: roads[15].Length - world.IntersectionZoneLength, Green: true, RedDuration: horizontalGreen, GreenDuration: verticalGreen},
		{RoadID: 18, Position: roads[18].Length - world.IntersectionZoneLength, Green: true, RedDuration: horizontalGreen, GreenDuration: verticalGreen},
	}

	via := func(entry, exit int) []int {
		entryNode := map[int]int{7: 0, 6: 1, 4: 2, 1: 3}[entry]
		exitNode := map[int]int{3: 0, 2: 1, 8: 2, 5: 3}[exit]
		segments := []int{9, 10, 11, 12}
		route := []int{entry}
		for node := entryNode; node != exitNode; node = (node + 1) % len(segments) {
			route = append(route, segments[node])
		}
		return append(route, exit)
	}
	toEast := func(entry, exit int) []int {
		return append(via(entry, 3), exit)
	}
	fromEast := func(exit int) []int {
		return append([]int{14}, via(7, exit)...)
	}
	fromNorth := func(exit int) []int {
		return append([]int{15}, via(7, exit)...)
	}
	fromSoutheast := func(exit int) []int {
		return append([]int{18}, via(7, exit)...)
	}
	buildings, homeAccess, workAccess := addScenarioBuildings(roads)
	homeAccess[31], homeAccess[34] = homeAccess[1], homeAccess[14]
	workAccess[33], workAccess[36] = workAccess[13], workAccess[5]
	homes := make(map[int]int)
	workplaces := make(map[int]int)
	for _, building := range buildings {
		if building.Kind == world.BuildingHome {
			homes[building.RoadID] = building.ID
		} else {
			workplaces[building.RoadID] = building.ID
		}
	}
	trip := func(route []int) []int {
		result := make([]int, 0, len(route)+2)
		result = append(result, homeAccess[route[0]])
		result = append(result, route...)
		return append(result, workAccess[route[len(route)-1]])
	}
	source := func(roadID int, interval float64, coreRoutes [][]int) world.SpawnSource {
		routes := make([][]int, len(coreRoutes))
		destinations := make([]int, len(coreRoutes))
		for i, core := range coreRoutes {
			routes[i] = trip(core)
			destinations[i] = workplaces[routes[i][len(routes[i])-1]]
		}
		accessID := homeAccess[roadID]
		return world.SpawnSource{RoadID: accessID, HomeID: homes[accessID], Interval: interval, Routes: routes, DestinationIDs: destinations}
	}
	initialCar := func(route []int, position, speed float64) car.Car {
		path := trip(route)
		return car.Car{OriginID: homes[path[0]], DestinationID: workplaces[path[len(path)-1]], Position: position, Speed: speed, DesiredSpeed: car.DefaultDesiredSpeed, Acceleration: 2, Braking: 4, Route: path, RoadIndex: 1}
	}

	return world.World{
		Roads: roads,
		Cars: []car.Car{
			initialCar(via(1, 2), 20, 0),
			initialCar(via(1, 8), 10, 0),
			initialCar(toEast(4, 13), 15, 7),
			initialCar(via(6, 5), 25, 7),
			initialCar(fromEast(5), 30, 7),
			initialCar(fromNorth(2), 20, 7),
			initialCar([]int{31, 32, 33}, 20, 8),
			initialCar([]int{34, 35, 36}, 20, 8),
		},
		TrafficLights: lights,
		Buildings:     buildings,
		Arrivals:      make(map[int]int),
		Intersections: []intersection.Intersection{{ID: 1, IncomingRoads: []int{3, 14, 15, 18}, OutgoingRoads: []int{7, 13, 16, 17}}},
		Roundabouts: []world.Roundabout{{
			Center: center, Radius: radius, Segments: []int{9, 10, 11, 12},
			Entries: []world.RoundaboutEntry{
				{ApproachRoadID: 7, PreviousRoadID: 12, NextRoadID: 9, ExitRoadID: 3},
				{ApproachRoadID: 6, PreviousRoadID: 9, NextRoadID: 10, ExitRoadID: 2},
				{ApproachRoadID: 4, PreviousRoadID: 10, NextRoadID: 11, ExitRoadID: 8},
				{ApproachRoadID: 1, PreviousRoadID: 11, NextRoadID: 12, ExitRoadID: 5},
			},
		}},
		SpawnSources: []world.SpawnSource{
			source(1, 12, [][]int{via(1, 2), {31, 32, 33}, via(1, 8), {31, 32, 33, 14, 16}}),
			source(4, 12, [][]int{via(4, 5), toEast(4, 13), via(4, 2), toEast(4, 17)}),
			source(6, 20, [][]int{via(6, 5), toEast(6, 13), via(6, 8)}),
			source(14, 20, [][]int{{34, 35, 36}, fromEast(2), fromEast(8), {14, 17}}),
			source(15, 20, [][]int{fromNorth(5), fromNorth(2), {15, 17}, {15, 13}}),
			source(18, 20, [][]int{fromSoutheast(5), fromSoutheast(8), {18, 16}, {18, 13}}),
		},
	}
}
