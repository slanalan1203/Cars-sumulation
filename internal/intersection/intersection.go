package intersection

import "slices"

type Intersection struct {
	ID            int
	IncomingRoads []int
	OutgoingRoads []int
}

func (i Intersection) Allows(incoming, outgoing int) bool {
	return slices.Contains(i.IncomingRoads, incoming) && slices.Contains(i.OutgoingRoads, outgoing)
}
