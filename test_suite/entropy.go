package test_suite

import (
	"math"
)

func EntropyEstimation(totalCounter map[byte]int, readBytesCount int) float64 {
	var p, entropy float64

	for i := 0; i < 256; i++ {
		p = float64(totalCounter[byte(i)]) / float64(readBytesCount)
		entropy += p * math.Log2(p)
	}
	return -entropy
}
