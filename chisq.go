package main

import (
	"math"
)

func chiSqTest(totalCounter map[byte]int, readBytesCount int, blockSize int) float64 {
	theoreticalDistribution := map[byte]float64{}
	for i := 0; i < 256; i++ {
		theoreticalDistribution[byte(i)] = float64(readBytesCount) / 256
	}

	var chiSquare, observed, expected float64
	for i := 0; i < 256; i++ {
		observed = float64(totalCounter[byte(i)])
		expected = float64(theoreticalDistribution[byte(i)])
		chiSquare += math.Pow((observed-expected), 2) / expected
	}
	return chiSquare
}
