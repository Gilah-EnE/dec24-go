package test_suite

import (
	"fmt"
	"github.com/montanaflynn/stats"
	"log"
	"math"
	"os"
	"sync"
)

type AutoCorrelationBSTestResult struct {
	Filename  string
	BlockSize int
	Result    float64
}

func autocorrGoroutine(filename string, blockSize int, channel chan AutoCorrelationBSTestResult, wg *sync.WaitGroup) {
	defer wg.Done()
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
			correlation, autocorrErr := stats.Correlation(floatBuffer[lag:], floatBuffer[:len(floatBuffer)-lag])
			if autocorrErr != nil {
				log.Fatalln("Autocorrelation calc error: ", autocorrErr)
			}
			results = append(results, math.Abs(correlation))
		}

		totalAutocorr = append(totalAutocorr, meanFloats(results))
	}
	stdev, err := stats.StandardDeviation(totalAutocorr)
	if err != nil {
		log.Fatalln("Standard deviation calc error: ", err)
	}
	resultStructure := AutoCorrelationBSTestResult{
		Filename:  filename,
		BlockSize: blockSize,
		Result:    stdev,
	}

	channel <- resultStructure
}

func bstest() {
	var wg sync.WaitGroup
	wg.Add(8)
	resultChannel := make(chan AutoCorrelationBSTestResult, 8)
	filename := "/dataset/images/random_32M.img"
	go autocorrGoroutine(filename, 1048576, resultChannel, &wg)
	go autocorrGoroutine(filename, 524288, resultChannel, &wg)
	go autocorrGoroutine(filename, 262144, resultChannel, &wg)
	go autocorrGoroutine(filename, 131072, resultChannel, &wg)
	go autocorrGoroutine(filename, 65536, resultChannel, &wg)
	go autocorrGoroutine(filename, 32768, resultChannel, &wg)
	go autocorrGoroutine(filename, 16384, resultChannel, &wg)
	go autocorrGoroutine(filename, 8192, resultChannel, &wg)
	wg.Wait()
	for job := range resultChannel {
		fmt.Println(job.Filename, job.BlockSize, job.Result)
	}
}
