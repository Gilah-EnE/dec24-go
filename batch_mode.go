package main

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
)

func main() {
	if len(os.Args) <= 1 {
		log.Fatalf("Использование: %s <путь к каталогу>\n", os.Args[0])
	} else {

		var path = os.Args[1]
		var blockSize = 1048576
		var autocorrThreshold = 0.125
		var ksTestThreshold = 0.1
		var compressionThreshold = 1.1
		var signatureThreshold = 150.0
		var entropyThreshold = 7.95

		entries, entryReadErr := os.ReadDir(path)
		if entryReadErr != nil {
			fmt.Printf("Не удалось прочитать содержимое каталога: %s", entryReadErr)
		}
		for _, entry := range entries {
			if !entry.IsDir() {
				if strings.Contains(entry.Name(), "_opt") || strings.Contains(entry.Name(), "opt") || strings.Contains(entry.Name(), "_opt_") || strings.Contains(entry.Name(), ".enclog") || strings.Contains(entry.Name(), ".enclog") {
					continue
				}
				fileName := filepath.Join(path, entry.Name())

				fileStat, fileStatErr := os.Stat(fileName)
				if fileStatErr != nil {
					log.Fatalf("Ошибка открытия файла: %s", fileStatErr)
				}
				fileSize := fileStat.Size()

				var minCounts int64 = 16
				var autocorrBlockSize = int(math.Min(1048576, math.Pow(2, math.Floor(math.Log2(float64(fileSize/minCounts))))))

				var logFile = fmt.Sprintf("%s.enclog", fileName)
				logFileHandle, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY, 0644)
				if err != nil {
					log.Fatalf("Не удалось открыть файл журнала: %s", err)
				}

				fileNormalLogger := log.New(logFileHandle, "", log.LstdFlags)
				consoleErrorLogger := log.New(os.Stderr, "", log.LstdFlags)
				fileErrorLogger := log.New(logFileHandle, "", log.LstdFlags)

				fileExtension := filepath.Ext(fileName)
				filePath := strings.TrimSuffix(fileName, fileExtension)

				var optimizedFileName = fmt.Sprintf("%s_opt%s", filePath, fileExtension)

				if _, optFileOpenErr := os.Stat(optimizedFileName); errors.Is(optFileOpenErr, os.ErrNotExist) {
					fileErrorLogger.Printf("Оптимизированный файл %s не найден.", optimizedFileName)
					result, fileOptimizationErr := exec.Command("python3", "prepare.py", "optimize", fileName).Output()
					if fileOptimizationErr != nil {
						fmt.Printf("Ошибка оптимизации файлв: %s", result)
					}
				}

				fmt.Printf("Код поиска шифрования разделов сырого образа диска, версия 2 (пакетный режим, проверка файлов. Имя файла: %s, размер блока: %d байтов, размер блока для теста автокорреляции: %d байтов.\n", fileName, blockSize, autocorrBlockSize)

				autocorrResult := autoCorrelation(optimizedFileName, autocorrBlockSize)
				fileNormalLogger.Printf("Коэффициент автокорреляции: %f, реф. значение %f\n", autocorrResult, autocorrThreshold)
				magicResult := libmagicAnalysis(fileName)
				fileNormalLogger.Printf("Тип файла: %s\n", magicResult)

				// noFSResults := []string{"", "unknown"}
				compressedMIMETypeEntries := []string{
					"application/arj",
					"application/bzip2",
					"application/gzip",
					"application/gzip-compressed",
					"application/gzipped",
					"application/lha",
					"application/lzh",
					"application/maclha",
					"application/x-ace",
					"application/x-arj",
					"application/x-bz2",
					"application/x-bzip",
					"application/x-compress",
					"application/x-compressed",
					"application/x-gzip",
					"application/x-gzip-compressed",
					"application/x-lha",
					"application/x-lzh",
					"application/x-lzh-archive",
					"application/x-stuffit",
					"application/x-winzip",
					"application/automationml-amlx+zip",
					"application/bacnet-xdd+zip",
					"application/epub+zip",
					"application/gzip",
					"application/lpf+zip",
					"application/p21+zip",
					"application/prs.hpub+zip",
					"application/prs.vcfbzip2",
					"application/tlsrpt+gzip",
					"application/urc-targetdesc+xml",
					"application/vnd.airzip.filesecure.azf",
					"application/vnd.airzip.filesecure.azs",
					"application/vnd.avistar+xml",
					"application/vnd.belightsoft.lhzd+zip",
					"application/vnd.belightsoft.lhzl+zip",
					"application/vnd.bzip3",
					"application/vnd.cncf.helm.chart.content.v1.tar+gzip",
					"application/vnd.comicbook+zip",
					"application/vnd.d2l.coursepackage1p0+zip",
					"application/vnd.dataresource+json",
					"application/vnd.dece.zip",
					"application/vnd.eln+zip",
					"application/vnd.espass-espass+zip",
					"application/vnd.etsi.asic-e+zip",
					"application/vnd.etsi.asic-s+zip",
					"application/vnd.exstream-empower+zip",
					"application/vnd.familysearch.gedcom+zip",
					"application/vnd.ficlab.flb+zip",
					"application/vnd.genozip",
					"application/vnd.gov.sk.e-form+zip",
					"application/vnd.imagemeter.folder+zip",
					"application/vnd.imagemeter.image+zip",
					"application/vnd.iso11783-10+zip",
					"application/vnd.keyman.kmp+zip",
					"application/vnd.laszip",
					"application/vnd.logipipe.circuit+zip",
					"application/vnd.maxar.archive.3tz+zip",
					"application/vnd.ms-cab-compressed",
					"application/vnd.nato.openxmlformats-package.iepd+zip",
					"application/vnd.software602.filler.form-xml-zip",
					"application/x-7z-compressed",
					"application/x-ace-compressed",
					"application/x-bzip2",
					"application/x-compress",
					"application/x-cpio",
					"application/x-gzip",
					"application/x-lzma",
					"application/x-rar-compressed",
					"application/x-xz",
					"application/x-zip-compressed",
					"application/x-zstd",
					"application/zip",
					"application/zstd",
					"application/java-archive",
					"application/x-rpm",
					"application/vnd.debian.binary-package",
				}

				contains := slices.Contains(compressedMIMETypeEntries, magicResult)

				if contains {
					fileNormalLogger.Printf("Этап 1: Тестируемый файл %s был предварительно сжат или является сжатым архивом.", fileName)
				} else {
					fileNormalLogger.Println("Этап 1: Предварительного сжатия не обнаружено. Переход на Этап 2.")
					counter, total := createFileCounter(optimizedFileName, blockSize)
					ksStatistic, maxDiffPosition, readBytesCount, _, _ := ksTest(counter, total)
					fileNormalLogger.Printf("Тест Колмогорова-Смирнова: максимальное отклонение: %f (реф. значение %f) в позиции %d, прочитано %d байтов.\n", ksStatistic, ksTestThreshold, maxDiffPosition, readBytesCount)
					compressionStat := compressionTest(optimizedFileName)
					fileNormalLogger.Printf("Средний коэффициент сжатия: %f, реф. значение %f\n", compressionStat, compressionThreshold)
					signatureStat := signatureAnalysis(optimizedFileName, blockSize)
					fileNormalLogger.Printf("Удельное количество сигнатур на мегабайт: %f, реф. значение %f\n", signatureStat, signatureThreshold)
					entropyStat := entropyEstimation(counter, total)
					fileNormalLogger.Printf("Оценочная информационная энтропия файла: %f, реф. значение %f\n", entropyStat, entropyThreshold)

					var autocorrTrue = autocorrResult <= autocorrThreshold
					var ksTrue = ksStatistic <= ksTestThreshold
					var compressionTrue = compressionStat <= compressionThreshold
					var signatureTrue = signatureStat <= signatureThreshold
					var entropyTrue = entropyStat >= entropyThreshold

					var finalResult = countTrueBools(autocorrTrue, ksTrue, compressionTrue, signatureTrue, entropyTrue)

					if finalResult <= 2 {
						fileNormalLogger.Printf("Этап 2: Количество положительных результатов %d <= 2, шифрования не обнаружено. Завершение работы программы.", finalResult)
					} else if finalResult > 3 && finalResult <= 5 {
						fileNormalLogger.Printf("Этап 2: Количество положительных результатов %d є [3,5], обнаружено шифрование. Завершение работы программы.", finalResult)
					} else {
						consoleErrorLogger.Println("Этап 2: Произошла ошибка подсчёта.")
						fileErrorLogger.Fatalln("Этап 2: Произошла ошибка подсчёта.")
					}
				}
				fileCloseErr := logFileHandle.Close()
				if fileCloseErr != nil {
					log.Fatalf("Не удалось закрыть файл журнала: %s", err)
				}
			}
		}
	}
}
