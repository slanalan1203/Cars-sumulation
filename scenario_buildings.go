package main

import (
	"github.com/slanalan1203/Cars-sumulation/internal/road"
	"github.com/slanalan1203/Cars-sumulation/internal/world"
)

type site struct {
	homeID       int
	workID       int
	homeRoadID   int
	workRoadID   int
	homePosition road.Point
	workPosition road.Point
	name         string
}

func addScenarioBuildings(roads map[int]road.Road) ([]world.Building, map[int]int, map[int]int) {
	sites := []site{
		{101, 201, 1, 5, road.Point{X: -11, Y: -15}, road.Point{X: -11, Y: 15}, "западе"},
		{102, 202, 4, 8, road.Point{X: -14, Y: 90}, road.Point{X: 13, Y: 114}, "юго-западе"},
		{103, 203, 6, 2, road.Point{X: 84, Y: 113}, road.Point{X: 116, Y: 113}, "юге"},
		{104, 204, 14, 13, road.Point{X: 291, Y: -15}, road.Point{X: 291, Y: 15}, "востоке"},
		{105, 205, 15, 16, road.Point{X: 169, Y: -105}, road.Point{X: 201, Y: -105}, "севере"},
		{106, 206, 18, 17, road.Point{X: 169, Y: 106}, road.Point{X: 201, Y: 106}, "юго-востоке"},
	}
	buildings := make([]world.Building, 0, len(sites)*2)
	homeAccess := make(map[int]int, len(sites))
	workAccess := make(map[int]int, len(sites))
	for i, location := range sites {
		homeAccessID := 19 + i*2
		workAccessID := homeAccessID + 1
		homeRoad := roads[location.homeRoadID]
		workRoad := roads[location.workRoadID]
		homeDrive := road.NewRoad(location.homePosition, homeRoad.Points[0])
		homeDrive.ID, homeDrive.Kind, homeDrive.SpeedLimit = homeAccessID, "access", 5
		workDrive := road.NewRoad(workRoad.Points[len(workRoad.Points)-1], location.workPosition)
		workDrive.ID, workDrive.Kind, workDrive.SpeedLimit = workAccessID, "access", 5
		roads[homeAccessID], roads[workAccessID] = homeDrive, workDrive
		buildings = append(buildings,
			world.Building{ID: location.homeID, Kind: world.BuildingHome, Name: "Дом на " + location.name, Position: location.homePosition, RoadID: homeAccessID},
			world.Building{ID: location.workID, Kind: world.BuildingWork, Name: "Работа на " + location.name, Position: location.workPosition, RoadID: workAccessID},
		)
		homeAccess[location.homeRoadID] = homeAccessID
		workAccess[location.workRoadID] = workAccessID
	}
	return buildings, homeAccess, workAccess
}
