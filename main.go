package main

import (
	"fmt"
	"log"
	"os"
)

func main() {
	if len(os.Args) <= 1 {
		log.Fatalf("Использование: %s <файл>\n", os.Args[0])
	} else {

		var fileName = os.Args[1]
		var blockSize = 1048576
		var autocorrThreshold = 0.1
		var ksTestThreshold = 0.1
		var compressionThreshold = 1.1
		var signatureThreshold = 150.0
		var entropyThreshold = 7.95

		var logFile = fmt.Sprintf("%s.log", fileName)
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
		log.SetOutput(logFileHandle)

		fmt.Printf("Код поиска шифрования разделов сырого образа диска, версия 2. Имя файла: %s, размер блока: %d байтов.\n", fileName, blockSize)

		autocorrResult := autoCorrelation(fileName, blockSize)
		log.Printf("Коэффициент автокорреляции: %f\n", autocorrResult)
		partedResult := partedCheck(fileName)
		log.Printf("Обнаруженная файловая система: %s\n", partedResult)

		if partedResult == "" || partedResult != "unknown" {
			if autocorrResult <= autocorrThreshold {
				log.Println("Этап 1: Файловая система с высокой долей вероятности содержит пофайловое шифрование или сжатые данные. Завершение работы программы.")
			} else {
				log.Println("Этап 1: Шифрования не обнаружено. Файловая система с высокой долей вероятности содержит незашифрованные файлы. Завершение работы программы.")
			}
		} else if partedResult == "unknown" {
			log.Println("Этап 1: Шифрования не обнаружено. Переход на Этап 2.")
			counter, total := createFileCounter(fileName, blockSize)
			ksStatistic, maxDiffPosition, readBytesCount, _, _ := ksTest(counter, total)
			log.Printf("Тест Колмогорова-Смирнова: максимальное отклонение: %f в позиции %d, прочитано %d байтов.\n", ksStatistic, maxDiffPosition, readBytesCount)
			compressionStat := compressionTest(fileName)
			log.Printf("Средний коэффициент сжатия: %f\n", compressionStat)
			signatureStat := signatureAnalysis(fileName, blockSize)
			log.Printf("Удельное количество сигнатур на мегабайт: %f\n", signatureStat)
			entropyStat := entropyEstimation(counter, total)
			log.Printf("Оценочная информационная энтропия файла: %f\n", entropyStat)

			var autocorrTrue = autocorrResult <= autocorrThreshold
			var ksTrue = ksStatistic <= ksTestThreshold
			var compressionTrue = compressionStat <= compressionThreshold
			var signatureTrue = signatureStat <= signatureThreshold
			var entropyTrue = entropyStat >= entropyThreshold

			var finalResult = countTrueBools(autocorrTrue, ksTrue, compressionTrue, signatureTrue, entropyTrue)

			if finalResult <= 2 {
				log.Printf("Этап 2: Количество положительных результатов %d <= 2, шифрования не обнаружено. Завершение работы программы.", finalResult)
			} else if finalResult > 3 && finalResult <= 5 {
				log.Printf("Этап 2: Количество положительных результатов %d є [3,5], обнаружено шифрование. Завершение работы программы.", finalResult)
			} else {
				log.Fatalln("Этап 2: Произошла ошибка подсчёта.")
			}
		}
	}
}
