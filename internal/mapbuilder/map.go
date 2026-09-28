package mapbuilder

import (
	"fmt"
	"math"

	"github.com/slanalan1203/Cars-sumulation/internal/intersection"
	"github.com/slanalan1203/Cars-sumulation/internal/road"
	"github.com/slanalan1203/Cars-sumulation/internal/trafficlight"
	"github.com/slanalan1203/Cars-sumulation/internal/world"
)

type Node struct {
	ID int     `json:"id"`
	X  float64 `json:"x"`
	Y  float64 `json:"y"`
}

type Street struct {
	ID       int     `json:"id"`
	From     int     `json:"from"`
	To       int     `json:"to"`
	SpeedKmh float64 `json:"speedKmh"`
	OneWay   bool    `json:"oneWay,omitempty"`
}

type LightSetting struct {
	NodeID          int     `json:"nodeID"`
	HorizontalGreen float64 `json:"horizontalGreen"`
	VerticalGreen   float64 `json:"verticalGreen"`
	FirstPhase      string  `json:"firstPhase"`
}

type RoundaboutSpec struct {
	ID        int     `json:"id"`
	CenterX   float64 `json:"centerX"`
	CenterY   float64 `json:"centerY"`
	Radius    float64 `json:"radius"`
	NodeIDs   []int   `json:"nodeIDs"`
	StreetIDs []int   `json:"streetIDs"`
}

type Building struct {
	ID            int     `json:"id"`
	Kind          string  `json:"kind"`
	NodeID        int     `json:"nodeID"`
	X             float64 `json:"x"`
	Y             float64 `json:"y"`
	CarsPerMinute float64 `json:"carsPerMinute"`
}

type Map struct {
	Nodes         []Node           `json:"nodes"`
	Streets       []Street         `json:"streets"`
	Buildings     []Building       `json:"buildings"`
	TrafficLights []int            `json:"trafficLights"`
	LightSettings []LightSetting   `json:"lightSettings,omitempty"`
	Roundabouts   []RoundaboutSpec `json:"roundabouts,omitempty"`
}

func (m Map) Validate() error {
	if len(m.Nodes) > 300 || len(m.Streets) > 500 || len(m.Buildings) > 200 || len(m.TrafficLights) > 300 {
		return fmt.Errorf("карта слишком большая")
	}
	nodes := make(map[int]Node, len(m.Nodes))
	for _, node := range m.Nodes {
		if node.ID <= 0 || node.ID > 10000 || !validCoordinate(node.X, node.Y) {
			return fmt.Errorf("неверная точка %d", node.ID)
		}
		if _, exists := nodes[node.ID]; exists {
			return fmt.Errorf("повтор точки %d", node.ID)
		}
		nodes[node.ID] = node
	}
	streets := make(map[int]bool, len(m.Streets))
	for _, street := range m.Streets {
		from, fromOK := nodes[street.From]
		to, toOK := nodes[street.To]
		if street.ID <= 0 || street.ID > 10000 || streets[street.ID] || !fromOK || !toOK || street.From == street.To || math.Hypot(from.X-to.X, from.Y-to.Y) < 10 || !finite(street.SpeedKmh) || street.SpeedKmh < 5 || street.SpeedKmh > 120 {
			return fmt.Errorf("неверная дорога %d", street.ID)
		}
		streets[street.ID] = true
	}
	buildings := make(map[int]bool, len(m.Buildings))
	for _, building := range m.Buildings {
		node, ok := nodes[building.NodeID]
		if building.ID <= 0 || building.ID > 10000 || buildings[building.ID] || !ok || (building.Kind != world.BuildingHome && building.Kind != world.BuildingWork) || !validCoordinate(building.X, building.Y) || math.Hypot(building.X-node.X, building.Y-node.Y) < 5 || !finite(building.CarsPerMinute) || building.CarsPerMinute < 0 || building.CarsPerMinute > 30 {
			return fmt.Errorf("неверное здание %d", building.ID)
		}
		buildings[building.ID] = true
	}
	lights := make(map[int]bool, len(m.TrafficLights))
	for _, id := range m.TrafficLights {
		if _, ok := nodes[id]; !ok || lights[id] {
			return fmt.Errorf("неверный светофор в точке %d", id)
		}
		lights[id] = true
	}
	settings := make(map[int]bool)
	for _, setting := range m.LightSettings {
		if !lights[setting.NodeID] || settings[setting.NodeID] || !finite(setting.HorizontalGreen) || !finite(setting.VerticalGreen) || setting.HorizontalGreen < 2 || setting.HorizontalGreen > 120 || setting.VerticalGreen < 2 || setting.VerticalGreen > 120 || setting.FirstPhase != "" && setting.FirstPhase != "horizontal" && setting.FirstPhase != "vertical" {
			return fmt.Errorf("неверные фазы светофора в точке %d", setting.NodeID)
		}
		settings[setting.NodeID] = true
	}
	usedNodes := make(map[int]bool)
	usedStreets := make(map[int]bool)
	usedCircles := make(map[int]bool)
	for _, circle := range m.Roundabouts {
		if circle.ID <= 0 || usedCircles[circle.ID] || !validCoordinate(circle.CenterX, circle.CenterY) || !finite(circle.Radius) || circle.Radius < 18 || circle.Radius > 80 || len(circle.NodeIDs) != 4 || len(circle.StreetIDs) != 4 {
			return fmt.Errorf("неверный круг %d", circle.ID)
		}
		usedCircles[circle.ID] = true
		for i := 0; i < 4; i++ {
			n, ok := nodes[circle.NodeIDs[i]]
			if !ok || usedNodes[n.ID] || math.Hypot(n.X-circle.CenterX-circle.Radius*math.Cos(-math.Pi/2+float64(i)*math.Pi/2), n.Y-circle.CenterY-circle.Radius*math.Sin(-math.Pi/2+float64(i)*math.Pi/2)) > 1e-6 {
				return fmt.Errorf("круг %d: неверная точка", circle.ID)
			}
			usedNodes[n.ID] = true
			streetID := circle.StreetIDs[i]
			if usedStreets[streetID] || !streets[streetID] {
				return fmt.Errorf("круг %d: неверная дуга", circle.ID)
			}
			usedStreets[streetID] = true
			found := false
			for _, street := range m.Streets {
				if street.ID == streetID && street.OneWay && street.From == circle.NodeIDs[i] && street.To == circle.NodeIDs[(i+1)%4] {
					found = true
				}
			}
			if !found {
				return fmt.Errorf("круг %d: неверное направление дуги", circle.ID)
			}
		}
	}
	return nil
}

func validCoordinate(x, y float64) bool {
	return finite(x) && finite(y) && x >= 0 && x <= 800 && y >= 0 && y <= 520
}

func finite(value float64) bool { return !math.IsNaN(value) && !math.IsInf(value, 0) }

type link struct {
	to     int
	roadID int
	cost   float64
}

func (m Map) Build() (world.World, error) {
	if err := m.Validate(); err != nil {
		return world.World{}, err
	}
	if len(m.Streets) == 0 {
		return world.World{}, fmt.Errorf("добавь хотя бы одну дорогу")
	}
	nodes := make(map[int]Node, len(m.Nodes))
	for _, node := range m.Nodes {
		nodes[node.ID] = node
	}
	roads := make(map[int]road.Road, len(m.Streets)*2+len(m.Buildings))
	graph := make(map[int][]link)
	incoming := make(map[int][]int)
	outgoing := make(map[int][]int)
	arcByStreet := make(map[int]struct {
		circle RoundaboutSpec
		index  int
	})
	for _, circle := range m.Roundabouts {
		for i, id := range circle.StreetIDs {
			arcByStreet[id] = struct {
				circle RoundaboutSpec
				index  int
			}{circle, i}
		}
	}
	for _, street := range m.Streets {
		from, to := nodes[street.From], nodes[street.To]
		forwardID, reverseID := street.ID*2-1, street.ID*2
		forward := road.NewRoad(road.Point{X: from.X, Y: from.Y}, road.Point{X: to.X, Y: to.Y})
		if arc, ok := arcByStreet[street.ID]; ok {
			points := make([]road.Point, 17)
			for i := range points {
				angle := -math.Pi/2 + float64(arc.index)*math.Pi/2 + float64(i)*math.Pi/32
				points[i] = road.Point{X: arc.circle.CenterX + arc.circle.Radius*math.Cos(angle), Y: arc.circle.CenterY + arc.circle.Radius*math.Sin(angle)}
			}
			points[0], points[len(points)-1] = road.Point{X: from.X, Y: from.Y}, road.Point{X: to.X, Y: to.Y}
			forward = road.NewRoad(points...)
			forward.Kind = "roundabout"
		}
		forward.ID, forward.SpeedLimit = forwardID, street.SpeedKmh/3.6
		forward.OneWay = street.OneWay
		roads[forwardID] = forward
		cost := forward.Length / forward.SpeedLimit
		graph[street.From] = append(graph[street.From], link{street.To, forwardID, cost})
		outgoing[street.From] = append(outgoing[street.From], forwardID)
		incoming[street.To] = append(incoming[street.To], forwardID)
		if !street.OneWay {
			roads[reverseID] = forward.Reverse(reverseID)
			graph[street.To] = append(graph[street.To], link{street.From, reverseID, cost})
			outgoing[street.To] = append(outgoing[street.To], reverseID)
			incoming[street.From] = append(incoming[street.From], reverseID)
		}
	}
	result := world.World{Roads: roads, Arrivals: make(map[int]int), Paused: false, CustomMap: true}
	for _, building := range m.Buildings {
		node := nodes[building.NodeID]
		position, junction := road.Point{X: building.X, Y: building.Y}, road.Point{X: node.X, Y: node.Y}
		id := 20000 + building.ID
		var access road.Road
		if building.Kind == world.BuildingHome {
			access = road.NewRoad(position, junction)
		} else {
			access = road.NewRoad(junction, position)
		}
		access.ID, access.Kind, access.SpeedLimit = id, "access", 5
		roads[id] = access
		name := "Дом"
		if building.Kind == world.BuildingWork {
			name = "Работа"
		}
		result.Buildings = append(result.Buildings, world.Building{ID: building.ID, Kind: building.Kind, Name: fmt.Sprintf("%s %d", name, building.ID), Position: position, RoadID: id})
	}
	for _, home := range m.Buildings {
		if home.Kind != world.BuildingHome {
			continue
		}
		source := world.SpawnSource{HomeID: home.ID, RoadID: 20000 + home.ID}
		if home.CarsPerMinute > 0 {
			source.Interval = 60 / home.CarsPerMinute
		}
		for _, workplace := range m.Buildings {
			if workplace.Kind != world.BuildingWork {
				continue
			}
			path, ok := shortestPath(graph, home.NodeID, workplace.NodeID)
			if !ok {
				continue
			}
			route := append([]int{source.RoadID}, path...)
			route = append(route, 20000+workplace.ID)
			source.Routes = append(source.Routes, route)
			source.DestinationIDs = append(source.DestinationIDs, workplace.ID)
		}
		if len(source.Routes) == 0 {
			return world.World{}, fmt.Errorf("у дома %d нет пути к работе", home.ID)
		}
		result.SpawnSources = append(result.SpawnSources, source)
	}
	if len(result.SpawnSources) == 0 {
		return world.World{}, fmt.Errorf("добавь дом")
	}
	if len(m.TrafficLights) > 0 {
		settings := make(map[int]LightSetting)
		for _, setting := range m.LightSettings {
			settings[setting.NodeID] = setting
		}
		for _, nodeID := range m.TrafficLights {
			if len(incoming[nodeID]) < 2 {
				return world.World{}, fmt.Errorf("светофору в точке %d нужно хотя бы две дороги", nodeID)
			}
			for _, roadID := range incoming[nodeID] {
				segment := roads[roadID]
				start, end := segment.Points[len(segment.Points)-2], segment.Points[len(segment.Points)-1]
				horizontal := math.Abs(end.X-start.X) >= math.Abs(end.Y-start.Y)
				setting := settings[nodeID]
				h, v := setting.HorizontalGreen, setting.VerticalGreen
				if h == 0 {
					h = 10
				}
				if v == 0 {
					v = 10
				}
				green := horizontal == (setting.FirstPhase != "vertical")
				greenDuration, redDuration := h, v
				if !horizontal {
					greenDuration, redDuration = v, h
				}
				result.TrafficLights = append(result.TrafficLights, trafficlight.TrafficLight{RoadID: roadID, Position: math.Max(0, segment.Length-world.IntersectionZoneLength), Green: green, RedDuration: redDuration, GreenDuration: greenDuration})
			}
		}
	}
	circleNodes := make(map[int]bool)
	for _, spec := range m.Roundabouts {
		circle := world.Roundabout{Center: road.Point{X: spec.CenterX, Y: spec.CenterY}, Radius: spec.Radius}
		for _, id := range spec.StreetIDs {
			circle.Segments = append(circle.Segments, id*2-1)
		}
		for i, nodeID := range spec.NodeIDs {
			circleNodes[nodeID] = true
			previous := circle.Segments[(i+3)%4]
			next := circle.Segments[i]
			exit := next
			for _, id := range outgoing[nodeID] {
				if id != next {
					exit = id
					break
				}
			}
			for _, id := range incoming[nodeID] {
				if id != previous {
					circle.Entries = append(circle.Entries, world.RoundaboutEntry{ApproachRoadID: id, PreviousRoadID: previous, NextRoadID: next, ExitRoadID: exit})
				}
			}
		}
		result.Roundabouts = append(result.Roundabouts, circle)
	}
	for _, node := range m.Nodes {
		if len(incoming[node.ID]) > 2 && !circleNodes[node.ID] {
			result.Intersections = append(result.Intersections, intersection.Intersection{ID: node.ID, IncomingRoads: incoming[node.ID], OutgoingRoads: outgoing[node.ID]})
		}
	}
	if err := result.Validate(); err != nil {
		return world.World{}, err
	}
	return result, nil
}

func shortestPath(graph map[int][]link, from, to int) ([]int, bool) {
	if from == to {
		return nil, true
	}
	distance := map[int]float64{from: 0}
	previousNode := make(map[int]int)
	previousRoad := make(map[int]int)
	visited := make(map[int]bool)
	for {
		current, best := 0, math.Inf(1)
		for node, cost := range distance {
			if !visited[node] && cost < best {
				current, best = node, cost
			}
		}
		if current == 0 || current == to {
			break
		}
		visited[current] = true
		for _, next := range graph[current] {
			cost := best + next.cost
			if old, ok := distance[next.to]; !ok || cost < old {
				distance[next.to] = cost
				previousNode[next.to], previousRoad[next.to] = current, next.roadID
			}
		}
	}
	if _, ok := distance[to]; !ok {
		return nil, false
	}
	path := make([]int, 0)
	for node := to; node != from; node = previousNode[node] {
		path = append(path, previousRoad[node])
	}
	for i, j := 0, len(path)-1; i < j; i, j = i+1, j-1 {
		path[i], path[j] = path[j], path[i]
	}
	return path, true
}
