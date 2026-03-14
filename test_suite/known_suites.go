package test_suite

import (
	"encoding/hex"
	"fmt"
	"log"
	"math"
	"os"

	"github.com/BurntSushi/rure-go"
)

type SignatureData struct {
	regex  string
	sector int64
}

type AdvancedSignatureMap struct {
	regex  *rure.Regex
	sector int64
}

func FoundSignaturesTotalToReadable(foundSignaturesTotal map[string]int) string {
	var readable string
	for key, value := range foundSignaturesTotal {
		readable = readable + fmt.Sprintf("%s - %d, ", key, value)
	}
	return readable
}

func SumFoundSignaturesTotal(foundSignaturesTotal map[string]int) int {
	sum := 0
	for _, value := range foundSignaturesTotal {
		sum += value
	}
	return sum
}

func ToolDetection(fileName string, blockSize int, hailMaryMode bool) map[string]int {
	signatures := make(map[string]AdvancedSignatureMap)

	logFileName := fmt.Sprintf("%s_suitedetection.log", fileName)
	logFileHandle, logOpenErr := os.OpenFile(logFileName, os.O_CREATE|os.O_WRONLY, 0644)
	if logOpenErr != nil {
		fmt.Printf("Не удалось открыть файл протокола: %s", logOpenErr)
	}
	defer func(logFileHandle *os.File) {
		logCloseErr := logFileHandle.Close()
		if logCloseErr != nil {
			fmt.Printf("Не удалось закрыть файл протокола: %s", logCloseErr)
		}
	}(logFileHandle)
	fileLogWriter := log.New(logFileHandle, "", log.LstdFlags)

	patterns := map[string]SignatureData{
		"FreeBSD GELI": {"(?i)(47454f4d3a3a454c49)", -1},
		"BitLocker":    {"(?i)(eb58902d4656452d46532d0002080000)", 1},
		"LUKSv1":       {"(?i)4c554b53babe0001", 1},
		"LUKSv2":       {"(?i)4c554b53babe0002", 1},
		"FileVault v2": {"(?i)41505342.{456}0800000000000000", 0},
	}

	foundSignaturesTotal := make(map[string]int)
	fileLogWriter.Println("------Прекомпиляция-------")
	for name, pattern := range patterns {
		regex, err := rure.Compile(pattern.regex)
		if err != nil {
			fileLogWriter.Printf("%s - ОШИБКА: %v", name, err)
			log.Fatalf("Ошибка: не удалось скомпилировать шаблон для %s: %v", name, err)
		}
		fileLogWriter.Printf("%s - OK", name)
		signatures[name] = AdvancedSignatureMap{regex: regex, sector: pattern.sector}
		foundSignaturesTotal[name] = 0
	}

	fileLogWriter.Printf("Было скомпилировано %d сигнатур.\n", len(signatures))

	file, err := os.Open(fileName)
	if err != nil {
		log.Fatal(err)
	}
	defer func(file *os.File) {
		fileCloseErr := file.Close()
		if fileCloseErr != nil {
			log.Fatal(fileCloseErr)
		}
	}(file)

	buffer := make([]byte, blockSize)

	if hailMaryMode {
		buffer := make([]byte, blockSize)
		n := 0
		for {
			bytesRead, err := file.Read(buffer)
			if bytesRead == 0 || err != nil {
				break
			}

			n += bytesRead
			fmt.Printf("%.1f ", float32(n)/1048576)

			hexData := hex.EncodeToString(buffer[:bytesRead])
			for sigType := range signatures {
				if entry, ok := signatures[sigType]; ok {
					foundSignaturesTotal[sigType] += FindBytesPattern(hexData, entry.regex)
				}
			}
			fmt.Print("\r")
		}
	} else {

		for sigType := range signatures {
			if entry, ok := signatures[sigType]; ok {
				skip := entry.sector

				var seekErr error

				if skip == 0 {
					buffer := make([]byte, blockSize)
					n := 0
					for {
						bytesRead, err := file.Read(buffer)
						if bytesRead == 0 || err != nil {
							break
						}

						n += bytesRead
						fmt.Printf("%.1f ", float32(n)/1048576)

						// Convert bytes to hex string
						hexData := hex.EncodeToString(buffer[:bytesRead])
						foundSignaturesTotal[sigType] += FindBytesPattern(hexData, entry.regex)
						fmt.Print("\r")
					}
				} else if skip != 0 {
					if skip < 0 {
						_, seekErr = file.Seek(int64(blockSize*int(math.Abs(float64(skip))-2)), 2)
					} else if skip > 0 {
						skip = skip - 1
						_, seekErr = file.Seek(int64(blockSize-1)*skip, 0)
					}
					if seekErr != nil {
						log.Fatalln("Ошибка поиска: ", seekErr)
					}
					bytesRead, fileReadErr := file.Read(buffer)
					if bytesRead == 0 || fileReadErr != nil {
						break
					}
					hexData := hex.EncodeToString(buffer[:bytesRead])
					foundSignaturesTotal[sigType] += FindBytesPattern(hexData, entry.regex)
					_, returnSeekErr := file.Seek(0, 0)
					if returnSeekErr != nil {
						log.Fatalln("Ошибка возврата поиска: ", returnSeekErr)
					}
				}
			}
		}
	}
	fmt.Print("\r")
	fileLogWriter.Println("----------Итоги----------")
	for key, value := range foundSignaturesTotal {
		fileLogWriter.Printf("Сигнатура: %s - найдено %d вхождений.\n", key, value)
	}

	fmt.Println(foundSignaturesTotalToReadable(foundSignaturesTotal))
	return foundSignaturesTotal
}
