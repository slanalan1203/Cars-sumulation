package car

// Distance — допустимый путь переднего бампера до остановки, в метрах.
// Длина лидера и безопасный зазор уже учтены миром. +Inf — свободный путь.
type Obstacle struct{ Distance float64 }
