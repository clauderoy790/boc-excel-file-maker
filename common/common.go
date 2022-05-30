package common

import (
	"math"
)

func Average(n1, n2 float64) float64 {
	avg := (n1 + n2) / 2
	return math.Round(avg*100) / 100
}
