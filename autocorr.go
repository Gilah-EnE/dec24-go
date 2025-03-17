package main

import (
	"fmt"
	"log"
	"math"
	"os"

	"github.com/montanaflynn/stats"
)

func autoCorrelation(filename string, blockSize int) float64 {
	file, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer func(file *os.File) {
		if err := file.Close(); err != nil {
			log.Fatal(err)
		}
	}(file)

	buffer := make([]byte, blockSize)
	var totalAutocorr []float64

	var readBytesCount int
	for {
		bytesRead, err := file.Read(buffer)
		if bytesRead == 0 || err != nil {
			break
		}
		var results []float64
		readBytesCount += bytesRead
		fmt.Printf("%.1f \r", float32(readBytesCount)/1048576)

		if len(buffer) > bytesRead {
			break
		}

		var floatBuffer []float64

		inputMean := meanBytes(buffer)

		for _, val := range buffer {
			floatBuffer = append(floatBuffer, float64(val)-inputMean)
		}

		var maxLag int
		if len(floatBuffer) < 50 {
			maxLag = len(floatBuffer)
		} else {
			maxLag = 50
		}

		for lag := 1; lag < maxLag; lag++ {
			correlation, _ := stats.Correlation(floatBuffer[lag:], floatBuffer[:len(floatBuffer)-lag])
			results = append(results, math.Abs(correlation))
		}

		totalAutocorr = append(totalAutocorr, meanFloats(results))
	}
	std, err := stats.StandardDeviation(totalAutocorr)
	if err != nil {
		log.Fatal(err)
	}
	return std
}
