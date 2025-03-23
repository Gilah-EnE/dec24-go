package main

import (
	"fmt"
	"io"
	"log"
	"os"
)

func countTrueBools(bools ...bool) int {
	var trueCount int
	for _, b := range bools {
		if b {
			trueCount++
		}
	}
	return trueCount
}

func countBytes(data []byte) map[byte]int {
	counter := make(map[byte]int)
	for _, b := range data {
		counter[b]++
	}
	return counter
}

func mergeCounterLists(counter1 map[byte]int, counter2 map[byte]int) map[byte]int {
	result := map[byte]int{}
	for k, v := range counter1 {
		result[k] += v
	}
	for k, v := range counter2 {
		result[k] += v
	}
	return result
}

func meanBytes(array []byte) float64 {
	var sum, mean float64

	for _, value := range array {
		sum += float64(value)
	}

	mean = sum / float64(len(array))
	return mean
}

func meanFloats(array []float64) float64 {
	var sum, mean float64

	for _, value := range array {
		sum += value
	}

	mean = sum / float64(len(array))
	return mean
}

func createFileCounter(filename string, blockSize int) (map[byte]int, int) {
	file, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(file)

	buffer := make([]byte, blockSize)
	var readBytesCount int
	totalCounter := map[byte]int{}

	for {
		bytesRead, err := file.Read(buffer)
		if err == io.EOF || bytesRead == 0 {
			break
		} else if err != nil {
			log.Fatal(err)
		}

		readBytesCount += bytesRead

		fmt.Printf("%.1f \r", float32(readBytesCount)/1048576)

		if len(buffer) > bytesRead {
			break
		}

		blockBytesCounter := countBytes(buffer)
		totalCounter = mergeCounterLists(totalCounter, blockBytesCounter)
	}
	return totalCounter, readBytesCount
}
