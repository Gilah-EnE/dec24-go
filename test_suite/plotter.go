package test_suite

import (
	"encoding/json"
	"fmt"
	"github.com/batchatco/go-native-netcdf/netcdf/api"
	"github.com/batchatco/go-native-netcdf/netcdf/cdf"
	"github.com/batchatco/go-native-netcdf/netcdf/util"
	"github.com/montanaflynn/stats"
	"log"
	"math"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type AutoCorrelationResult struct {
	Filename string
	Result   map[int]float64
}

func autoCorrGoro(filename string, blockSize int, resultChannel chan AutoCorrelationResult, wg *sync.WaitGroup) {
	defer wg.Done()
	file, err := os.Open(filename)
	if err != nil {
		log.Fatalf("Autocorrelation image opening file error: %s", err)
	}
	defer func(file *os.File) {
		if err := file.Close(); err != nil {
			log.Fatalf("Autocorrelation image closing file error: %s", err)
		}
	}(file)

	buffer := make([]byte, blockSize)
	totalAutocorr := make(map[int]float64)

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

		totalAutocorr[readBytesCount] = meanFloats(results)
	}
	resultStruct := AutoCorrelationResult{
		Filename: filename,
		Result:   totalAutocorr,
	}
	resultChannel <- resultStruct
}

func createAutoCorrNPY(fileName1 string, fileName2 string, autoCorrelation1 map[int]float64, autoCorrelation2 map[int]float64) {
	autoCorrCombined := map[string]map[int]float64{
		fileName1: autoCorrelation1,
		fileName2: autoCorrelation2,
	}

	outputFileName := fmt.Sprintf("%s_%s.json", filepath.Base(fileName1), filepath.Base(fileName2))

	jsonData, jsonErr := json.MarshalIndent(autoCorrCombined, "", "\t")
	if jsonErr != nil {
		log.Fatalf("JSON marshalling error: %s", jsonErr)
	}
	fileWriteErr := os.WriteFile(outputFileName, jsonData, 0644)
	if fileWriteErr != nil {
		log.Fatalf("JSON saving error: %s", fileWriteErr)
	}
}

func createKSGraphNetCDF(fileName1 string, fileName2 string, file1Counter map[byte]int, file2Counter map[byte]int, readBytesCount1 int, readBytesCount2 int) {
	var empiricalCumSum1, empiricalCumSum2, theoreticalCumSum float64
	var empiricalCDF1, empiricalCDF2, theoreticalCDF []float64
	var byteCounter1, byteCounter2 []float64

	for i := 0; i < 256; i++ {
		empiricalCumSum1 += float64(file1Counter[byte(i)]) / float64(readBytesCount1)
		empiricalCumSum2 += float64(file2Counter[byte(i)]) / float64(readBytesCount2)
		theoreticalCumSum += float64(readBytesCount1) / 256 / float64(readBytesCount1)
		empiricalCDF1 = append(empiricalCDF1, empiricalCumSum1)
		empiricalCDF2 = append(empiricalCDF2, empiricalCumSum2)
		theoreticalCDF = append(theoreticalCDF, theoreticalCumSum)
		byteCounter1 = append(byteCounter1, float64(file1Counter[byte(i)]))
		byteCounter2 = append(byteCounter2, float64(file2Counter[byte(i)]))
	}
	cdfFileName := fmt.Sprintf("%s_%s.nc", filepath.Base(fileName1), filepath.Base(fileName2))
	netcdf, openErr := cdf.OpenWriter(cdfFileName)
	// netcdf.dimLengths["units"] = int64(0)
	if openErr != nil {
		log.Fatalf("NetCDF file opening error: %s", openErr)
	}

	createAddVariable("empirical_cdf_1", fileName1, empiricalCDF1, netcdf)
	createAddVariable("empirical_cdf_2", fileName2, empiricalCDF2, netcdf)
	createAddVariable("theoretical_cdf", "Theoretical", empiricalCDF1, netcdf)
	createAddVariable("byte_counter_1", fileName1, byteCounter1, netcdf)
	createAddVariable("byte_counter_2", fileName2, byteCounter2, netcdf)

	closeErr := netcdf.Close()
	if closeErr != nil {
		log.Fatalf("NetCDF file closing error: %s", closeErr)
	}
}

func createAddVariable(varName string, fileName string, values []float64, file *cdf.CDFWriter) {
	attributeMap, err := util.NewOrderedMap(
		[]string{"filename"},
		map[string]interface{}{"filename": fileName},
	)
	if err != nil {
		log.Fatalf("Variable attribute map creation error: %s", err)
	}

	netCDFVar := api.Variable{
		Values:     values,
		Dimensions: []string{"units=0"},
		Attributes: attributeMap,
	}

	addErr := file.AddVar(varName, netCDFVar)
	if addErr != nil {
		log.Fatalf("Variable append error: %s", addErr)
	}
}

func createNetCDFFile(fileName1 string, fileName2 string, wgroup *sync.WaitGroup) {
	defer wgroup.Done()
	start := time.Now()
	var performAutocorr = true
	var performCounting = false
	var wg sync.WaitGroup
	wg.Add(2)

	counterResultChannel := make(chan ByteCounter, 2)
	autoCorrelationResultChannel := make(chan AutoCorrelationResult, 2)

	if performCounting {
		go CreateFileCounterGoro(fileName1, 1048576, counterResultChannel, &wg)
		go CreateFileCounterGoro(fileName2, 1048576, counterResultChannel, &wg)
	}
	if performAutocorr {
		go autoCorrGoro(fileName1, 8192, autoCorrelationResultChannel, &wg)
		go autoCorrGoro(fileName2, 8192, autoCorrelationResultChannel, &wg)
	}

	wg.Wait()

	if performCounting {
		file1ResultStruct := <-counterResultChannel
		fileName1, file1Counter, file1Size := file1ResultStruct.Filename, file1ResultStruct.Counter, file1ResultStruct.BytesRead
		file2ResultStruct := <-counterResultChannel
		fileName2, file2Counter, file2Size := file2ResultStruct.Filename, file2ResultStruct.Counter, file2ResultStruct.BytesRead
		close(counterResultChannel)
		createKSGraphNetCDF(fileName1, fileName2, file1Counter, file2Counter, file1Size, file2Size)
	}

	if performAutocorr {
		AutoCorrelationResultStruct1 := <-autoCorrelationResultChannel
		fileName1, autoCorrelationResult1 := AutoCorrelationResultStruct1.Filename, AutoCorrelationResultStruct1.Result
		AutoCorrelationResultStruct2 := <-autoCorrelationResultChannel
		fileName2, autoCorrelationResult2 := AutoCorrelationResultStruct2.Filename, AutoCorrelationResultStruct2.Result
		close(autoCorrelationResultChannel)
		createAutoCorrNPY(fileName1, fileName2, autoCorrelationResult1, autoCorrelationResult2)
	}

	log.Printf("NetCDF and JSON files creation for %s and %s images has been completed. Total time: %s", fileName1, fileName2, time.Since(start))
}

// func main() {
// 	var wg sync.WaitGroup
// 	wg.Add(1)
// 	go createNetCDFFile("/dataset/images/random_32M.img", "/dataset/images/random_empty_spaces_16M.img", &wg)
// 	// go createNetCDFFile("/dataset/images/wd400.img", "/dataset/images/wd400_opt.img", &wg)
// 	// go createNetCDFFile("/dataset/images/adoptable.img", "/dataset/images/adoptable_opt.img", &wg)
// 	// go createNetCDFFile("/dataset/images/kagura/kagura_data_dec.img", "/dataset/images/kagura/kagura_data_dec_opt.img", &wg)
// 	// go createNetCDFFile("/dataset/images/kagura/kagura_data_enc.img", "/dataset/images/kagura/kagura_data_enc_opt.img", &wg)
// 	// go createNetCDFFile("/dataset/images/vince/vince_data.img", "/dataset/images/vince/vince_data_opt.img", &wg)
// 	// go createNetCDFFile("/dataset/images/miatoll/miatoll_data_fbe.img", "/dataset/images/miatoll/miatoll_data_fbe_opt.img", &wg)
// 	// go createNetCDFFile("/dataset/images/miatoll/miatoll_data_nonfbe.img", "/dataset/images/miatoll/miatoll_data_nonfbe_opt.img", &wg)
// 	wg.Wait()
// }
