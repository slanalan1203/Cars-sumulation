package world

import "github.com/slanalan1203/Cars-sumulation/internal/road"

const (
	BuildingHome = "home"
	BuildingWork = "work"
)

type Building struct {
	ID       int
	Kind     string
	Name     string
	Position road.Point
	RoadID   int
}
