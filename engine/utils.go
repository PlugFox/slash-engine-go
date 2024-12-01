package engine

// Clamp a value between a min and max
func clamp(val float64, minValue float64, maxValue float64) float64 {
	if val < minValue {
		return minValue
	}
	if val > maxValue {
		return maxValue
	}
	return val
}
