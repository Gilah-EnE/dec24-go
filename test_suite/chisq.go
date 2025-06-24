package test_suite

import (
	"math"
)

func ChiSqTest(totalCounter map[byte]int, readBytesCount int) float64 {
	theoreticalDistribution := map[byte]float64{}
	for i := 0; i < 256; i++ {
		theoreticalDistribution[byte(i)] = float64(readBytesCount) / 256
	}

	var chiSquare, observed, expected float64
	for i := 0; i < 256; i++ {
		observed = float64(totalCounter[byte(i)])
		expected = theoreticalDistribution[byte(i)]
		chiSquare += math.Pow(observed-expected, 2) / expected
	}
	return chiSquare
}
