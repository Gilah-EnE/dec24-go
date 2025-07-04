package test_suite

import (
	"math"
)

func KsTest(totalCounter map[byte]int, readBytesCount int) (float64, int, int, float64, float64) {

	var empiricalCumSum float64
	var theoreticalCumSum float64
	var empiricalCDF []float64
	var theoreticalCDF []float64

	for i := 0; i < 256; i++ {
		empiricalCumSum += float64(totalCounter[byte(i)]) / float64(readBytesCount)
		theoreticalCumSum += float64(readBytesCount) / 256 / float64(readBytesCount)
		empiricalCDF = append(empiricalCDF, empiricalCumSum)
		theoreticalCDF = append(theoreticalCDF, theoreticalCumSum)
	}

	var ksDifferences []float64

	for idx := range empiricalCDF {
		ksDifferences = append(ksDifferences, math.Abs(empiricalCDF[idx]-theoreticalCDF[idx]))
	}
	ksStatistic := ksDifferences[0]
	maxDiffPosition := 0

	for idx, value := range ksDifferences {
		if value > ksStatistic {
			ksStatistic = value
			maxDiffPosition = idx
		} else {
			continue
		}
	}
	criticalValue001 := 1.63 / math.Sqrt(float64(readBytesCount))
	criticalValue005 := 1.36 / math.Sqrt(float64(readBytesCount))
	return ksStatistic, maxDiffPosition, readBytesCount, criticalValue001, criticalValue005
}
