package gesture

import "math"

func distance(x1 int32, y1 int32, x2 int32, y2 int32) float64 {
	return math.Sqrt(math.Pow(float64(x1-x2), 2) + math.Pow(float64(y1-y2), 2))
}
