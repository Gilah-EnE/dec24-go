package test_suite

import (
	"errors"
	"fmt"
	"log"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"time"
)

func oneshotMode() {
	if len(os.Args) <= 1 {
		log.Fatalf("Использование: %s <файл>\n", os.Args[0])
	} else {

		var fileName = os.Args[1]
		var blockSize = 1048576
		var autocorrThreshold = 0.125
		var ksTestThreshold = 0.1
		var compressionThreshold = 1.1
		var signatureThreshold = 150.0
		var entropyThreshold = 7.95

		var logFile = fmt.Sprintf("%s.enclog", fileName)
		logFileHandle, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Fatalf("Не удалось открыть файл журнала: %s", err)
		}
		defer func(logFileHandle *os.File) {
			err := logFileHandle.Close()
			if err != nil {
				log.Fatalf("Не удалось закрыть файл журнала: %s", err)
			}
		}(logFileHandle)

		fileNormalLogger := log.New(logFileHandle, "", log.LstdFlags)
		consoleErrorLogger := log.New(os.Stderr, "", log.LstdFlags)
		fileErrorLogger := log.New(logFileHandle, "", log.LstdFlags)

		fileExtension := filepath.Ext(fileName)
		filePath := strings.TrimSuffix(fileName, fileExtension)

		fileStat, fileStatErr := os.Stat(fileName)
		if fileStatErr != nil {
			log.Fatalf("Ошибка открытия файла: %s", fileStatErr)
		}
		fileSize := fileStat.Size()

		var minCounts int64 = 16
		var autocorrBlockSize = int(math.Min(1048576, math.Pow(2, math.Floor(math.Log2(float64(fileSize/minCounts))))))

		var optimizedFileName = fmt.Sprintf("%s_opt%s", filePath, fileExtension)

		if _, optFileOpenErr := os.Stat(optimizedFileName); errors.Is(optFileOpenErr, os.ErrNotExist) {
			consoleErrorLogger.Printf("Оптимизированный файл %s не найден.", optimizedFileName)
			result, fileOptimizationErr := exec.Command("python3", "prepare.py", "optimize", fileName).Output()
			if fileOptimizationErr != nil {
				fmt.Printf("Ошибка оптимизации файла: %s", result)
			}
		}

		fmt.Printf("Код поиска шифрования разделов сырого образа диска, версия 2 (одиночный режим, проверка образов). Имя файла: %s, размер блока: %d байтов, размер блока для теста автокорреляции: %d байтов.\n", fileName, blockSize, autocorrBlockSize)
		autocorrStartTime := time.Now()
		autocorrResult := AutoCorrelation(optimizedFileName, autocorrBlockSize)
		fmt.Printf("Время теста автокорреляции: %.15f \n", time.Since(autocorrStartTime).Seconds())
		fileNormalLogger.Printf("Коэффициент автокорреляции: %f, реф. значение %f\n", autocorrResult, autocorrThreshold)
		partedResult := PartedCheck(fileName)
		fileNormalLogger.Printf("Обнаруженная файловая система: %s\n", partedResult)

		noFSResults := []string{"", "unknown"}

		contains := slices.Contains(noFSResults, partedResult)

		if contains {
			fileNormalLogger.Println("Этап 1: Шифрования не обнаружено. Переход на Этап 2.")
			counterStartTime := time.Now()
			counter, total := CreateFileCounter(optimizedFileName, blockSize)
			fmt.Printf("Время набора стат. модели: %.15f \n", time.Since(counterStartTime).Seconds())
			ksStartTime := time.Now()
			ksStatistic, maxDiffPosition, readBytesCount, _, _ := KsTest(counter, total)
			fmt.Printf("Время теста К-С: %.15f \n", time.Since(ksStartTime).Seconds())
			fileNormalLogger.Printf("Тест Колмогорова-Смирнова: максимальное отклонение: %f (реф. значение %f) в позиции %d, прочитано %d байтов.\n", ksStatistic, ksTestThreshold, maxDiffPosition, readBytesCount)
			compressionStartTime := time.Now()
			compressionStat := CompressionTest(optimizedFileName)
			fmt.Printf("Время теста сжатия: %.15f \n", time.Since(compressionStartTime).Seconds())
			fileNormalLogger.Printf("Средний коэффициент сжатия: %f, реф. значение %f\n", compressionStat, compressionThreshold)
			signatureStartTime := time.Now()
			signatureStat := SignatureAnalysis(optimizedFileName, blockSize)
			fmt.Printf("Время теста сигнатур: %.15f \n", time.Since(signatureStartTime).Seconds())
			fileNormalLogger.Printf("Удельное количество сигнатур на мегабайт: %f, реф. значение %f\n", signatureStat, signatureThreshold)
			entropyStartTime := time.Now()
			entropyStat := EntropyEstimation(counter, total)
			fmt.Printf("Время теста энтропии: %.15f \n", time.Since(entropyStartTime).Seconds())
			fileNormalLogger.Printf("Оценочная информационная энтропия файла: %f, реф. значение %f\n", entropyStat, entropyThreshold)

			var autocorrTrue = autocorrResult <= autocorrThreshold
			var ksTrue = ksStatistic <= ksTestThreshold
			var compressionTrue = compressionStat <= compressionThreshold
			var signatureTrue = signatureStat <= signatureThreshold
			var entropyTrue = entropyStat >= entropyThreshold

			var finalResult = CountTrueBools(autocorrTrue, ksTrue, compressionTrue, signatureTrue, entropyTrue)

			if finalResult <= 2 {
				fileNormalLogger.Printf("Этап 2: Количество положительных результатов %d <= 2, шифрования не обнаружено. Завершение работы программы.", finalResult)
			} else if finalResult > 3 && finalResult <= 5 {
				fileNormalLogger.Printf("Этап 2: Количество положительных результатов %d є [3,5], обнаружено шифрование. Завершение работы программы.", finalResult)
			} else {
				consoleErrorLogger.Println("Этап 2: Произошла ошибка подсчёта.")
				fileErrorLogger.Fatalln("Этап 2: Произошла ошибка подсчёта.")
			}
		} else {
			if autocorrResult <= autocorrThreshold {
				fileNormalLogger.Println("Этап 1: Файловая система с высокой долей вероятности содержит пофайловое шифрование или сжатые данные. Завершение работы программы.")
			} else {
				fileNormalLogger.Println("Этап 1: Шифрования не обнаружено. Файловая система с высокой долей вероятности содержит незашифрованные файлы. Завершение работы программы.")
			}
		}
	}
}
