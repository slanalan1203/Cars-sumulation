package intersection

// Intersection пока только описывает соединение дорог.
// Управление конфликтующими потоками и резервирование проезда не реализованы.
type Intersection struct {
	ID            int
	IncomingRoads []int
	OutgoingRoads []int
}
