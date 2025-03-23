package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

func performCompression(execCommand string) float64 {
	gzipOutput, err := exec.Command("bash", "-c", execCommand).CombinedOutput()
	if err != nil {
		fmt.Println(err)
	}

	gzipCompression, err := strconv.Atoi(strings.Split(string(gzipOutput), "\n")[0])
	if err != nil {
		fmt.Println(err)
	}
	return float64(gzipCompression)
}

func compressionTest(fileName string) float64 {
	var gzipCompression, lz4Compression, bz2Compression, zstdCompression, xzCompression float64
	//var gzipOutput, lz4Output, bz2Output, zstdOutput, xzOutput []byte
	var gzipExec, lz4Exec, bz2Exec, zstdExec, xzExec string

	gzipExec = fmt.Sprintf("pigz < %s | wc -c", fileName)
	lz4Exec = fmt.Sprintf("lz4 < %s | wc -c", fileName)
	bz2Exec = fmt.Sprintf("lbzip2 < %s | wc -c", fileName)
	zstdExec = fmt.Sprintf("zstd < %s | wc -c", fileName)
	xzExec = fmt.Sprintf("pixz < %s | wc -c", fileName)

	stat, err := os.Stat(fileName)
	if err != nil {
		fmt.Println(err)
	}

	fileSize := float64(stat.Size())

	log.Println("Начало сжатия GZip")
	gzipCompression = fileSize / performCompression(gzipExec)
	log.Println("Начало сжатия LZ4")
	lz4Compression = fileSize / performCompression(lz4Exec)
	log.Println("Начало сжатия BZ2")
	bz2Compression = fileSize / performCompression(bz2Exec)
	log.Println("Начало сжатия Zstd")
	zstdCompression = fileSize / performCompression(zstdExec)
	log.Println("Начало сжатия XZ")
	xzCompression = fileSize / performCompression(xzExec)

	avgCompression := (gzipCompression + lz4Compression + bz2Compression + zstdCompression + xzCompression) / 5
	log.Printf("Конец теста компрессии. GZIP: %f, LZ4: %f, BZ2: %f, Zstd: %f, XZ: %f", gzipCompression, lz4Compression, bz2Compression, zstdCompression, xzCompression)

	return avgCompression
}
