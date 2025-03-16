package main

import (
	"fmt"
)

func main() {
	var fileName string = "/dataset/images/random_32M.img"
	var blockSize int = 1048576

	fmt.Printf("Go version of statistical analysis suite. Filename: %s, block size: %d bytes.\n", fileName, blockSize)
	counter, total := createFileCounter(fileName, blockSize)

	fmt.Println("Autocorr: ", autoCorrelation(fileName, blockSize))
	fmt.Println("Pearson Test: ", chiSqTest(counter, total, blockSize))
	fmt.Println("Entropy Estimation: ", entropyEstimation(counter, total, blockSize))
	ksStatistic, maxDiffPosition, readBytesCount, criticalValue001, criticalValue005 := ksTest(counter, total, blockSize)
	fmt.Println("Kolmogorov-Smirnov test: ", ksStatistic, maxDiffPosition, readBytesCount, criticalValue001, criticalValue005)
	fmt.Println("Signature analysis: ", signatureAnalysis(fileName, blockSize))
}
