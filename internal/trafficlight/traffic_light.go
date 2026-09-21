package trafficlight

type TrafficLight struct {
	RoadID   int
	Position float64
	Green    bool

	RedDuration   float64
	GreenDuration float64

	Elapsed float64
}

func (t *TrafficLight) Update(dt float64) {
	if dt <= 0 || t.RedDuration <= 0 || t.GreenDuration <= 0 {
		return
	}

	t.Elapsed += dt

	for {
		duration := t.GreenDuration
		if !t.Green {
			duration = t.RedDuration
		}

		if t.Elapsed < duration {
			return
		}

		t.Elapsed -= duration
		t.Green = !t.Green
	}
}
