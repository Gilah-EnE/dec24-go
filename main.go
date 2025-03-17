package main

import (
	"fmt"
)

func main() {
	fileName := "/dataset/images/random_32M.img"
	blockSize := 1048576

	fmt.Printf("Go version of statistical analysis suite. Filename: %s, block size: %d bytes.\n", fileName, blockSize)
	counter, total := createFileCounter(fileName, blockSize)

	fmt.Println("Auto-correlation: ", autoCorrelation(fileName, blockSize))
	fmt.Println("Pearson Test: ", chiSqTest(counter, total))
	fmt.Println("Entropy Estimation: ", entropyEstimation(counter, total))
	ksStatistic, maxDiffPosition, readBytesCount, criticalValue001, criticalValue005 := ksTest(counter, total)
	fmt.Println("Kolmogorov-Smirnov test: ", ksStatistic, maxDiffPosition, readBytesCount, criticalValue001, criticalValue005)
	fmt.Println("Signature analysis: ", signatureAnalysis(fileName, blockSize))
}
