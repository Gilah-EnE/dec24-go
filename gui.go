package main

import (
	"errors"
	"fmt"
	"github.com/mappu/miqt/qt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
)

const (
	NoEncryption int = iota
	FullDiskEncryption
	FileBasedEncryption
)

func main() {
	qt.NewQApplication(os.Args)
	window := qt.NewQMainWindow(nil)
	window.SetWindowTitle("Код поиска шифрования разделов сырого образа диска, версия 3.1")
	window.SetMinimumSize2(800, 20)

	// Adding menu actions
	menuBar := window.MenuBar()

	fileMenu := qt.NewQMenu3("Файл")
	fileMenu.AddAction2(qt.QIcon_FromTheme("help-about"), "О программе")
	fileMenu.AddAction2(qt.NewQIcon4(":/qt-project.org/qmessagebox/images/qtlogo-64.png"), "О Qt")
	fileMenu.AddSeparator()
	fileMenu.AddAction2(qt.QIcon_FromTheme("application-exit"), "Выход")
	menuBar.AddMenu(fileMenu)

	// Creating window layouts
	widget := qt.NewQWidget(nil)
	mainLayout := qt.NewQVBoxLayout(widget)
	filePickerLayout := qt.NewQGridLayout(widget)
	resultsLayout := qt.NewQGridLayout(widget)

	modeSelectLayout := qt.NewQHBoxLayout(widget)
	oneshotSelector := qt.NewQRadioButton3("Проверка одного файла")
	oneshotSelector.SetChecked(true)
	multipleSelector := qt.NewQRadioButton3("Проверка всех файлов в каталоге")
	// recursiveSelector := qt.NewQCheckBox4("Рекурсивно", widget)
	// recursiveSelector.SetChecked(false)
	// recursiveSelector.SetEnabled(false)
	// oneshotSelector.OnToggled(func(checked bool) {
	// 	if checked == true {
	// 		recursiveSelector.SetChecked(false)
	// 		recursiveSelector.SetEnabled(false)
	// 	}
	// })
	// multipleSelector.OnToggled(func(checked bool) {
	// 	recursiveSelector.SetEnabled(checked)
	// })
	modeSelectLayout.AddWidget(oneshotSelector.QWidget)
	modeSelectLayout.AddWidget(multipleSelector.QWidget)
	// modeSelectLayout.AddWidget(recursiveSelector.QWidget)

	// File picker button
	fileNameTextField := qt.NewQLineEdit(widget)
	if oneshotSelector.IsChecked() {
		fileNameTextField.SetPlaceholderText("Введите путь к исследуемому файлу образа")
	} else if multipleSelector.IsChecked() {
		fileNameTextField.SetPlaceholderText("Введите путь к исследуемому каталогу")
	}
	filePickerButton := qt.NewQPushButton4(qt.QIcon_FromTheme("document-open"), "Выбор файла")

	filePickerButton.OnClicked(func() {
		var caption string
		if oneshotSelector.IsChecked() {
			caption = "Выберите файл для анализа"
		} else {
			caption = "Выберите каталог для анализа"
		}

		fileDialog := qt.NewQFileDialog4(widget, caption)

		if oneshotSelector.IsChecked() {
			fileDialog.SetFileMode(qt.QFileDialog__ExistingFile)
			fileDialog.SetNameFilter("Все файлы (*)")
		} else {
			fileDialog.SetFileMode(qt.QFileDialog__DirectoryOnly)
		}

		if fileDialog.Exec() == int(qt.QDialog__Accepted) {
			selectedFile := fileDialog.SelectedFiles()
			if len(selectedFile) > 0 {
				filePath := selectedFile[0]
				fileNameTextField.SetText(filePath)
			}
		}
	})
	startButton := qt.NewQPushButton4(qt.QIcon_FromTheme("media-playback-start"), "Анализ")

	encryptedFileLocationEdit := qt.NewQLineEdit(widget)
	encryptedFileLocationEdit.SetPlaceholderText("Введите путь, куда будут перемещены зашифрованные файлы")
	cwd, getCwdErr := os.Getwd()
	if getCwdErr != nil {
		log.Fatal(getCwdErr)
	}
	encryptedFileLocationEdit.SetText(cwd)
	encryptedFileLocationPickerButton := qt.NewQPushButton4(qt.QIcon_FromTheme("folder-open"), "Выбор каталога")

	encryptedFileLocationPickerButton.OnClicked(func() {
		caption := "Выберите каталог для сохранения зашифрованных файлов"
		dirDialog := qt.NewQFileDialog4(widget, caption)

		dirDialog.SetFileMode(qt.QFileDialog__DirectoryOnly)

		if dirDialog.Exec() == int(qt.QDialog__Accepted) {
			selectedFile := dirDialog.SelectedFiles()
			if len(selectedFile) > 0 {
				filePath := selectedFile[0]
				encryptedFileLocationEdit.SetText(filePath)
			}
		}
	})

	filePickerLayout.AddWidget2(fileNameTextField.QWidget, 0, 0)
	filePickerLayout.AddWidget2(filePickerButton.QWidget, 0, 1)
	filePickerLayout.AddWidget2(startButton.QWidget, 0, 2)
	filePickerLayout.AddWidget3(encryptedFileLocationEdit.QWidget, 1, 0, 1, 2)
	filePickerLayout.AddWidget2(encryptedFileLocationPickerButton.QWidget, 1, 2)

	// Values display widgets
	autoCorrResultDisplay := qt.NewQLineEdit(widget)
	autoCorrResultDisplay.SetReadOnly(true)

	fsResultDisplay := qt.NewQLineEdit(widget)
	fsResultDisplay.SetReadOnly(true)

	ksResultDisplay := qt.NewQLineEdit(widget)
	ksResultDisplay.SetReadOnly(true)

	compressionStatDisplay := qt.NewQLineEdit(widget)
	compressionStatDisplay.SetReadOnly(true)

	sigResultDisplay := qt.NewQLineEdit(widget)
	sigResultDisplay.SetReadOnly(true)

	entropyStatDisplay := qt.NewQLineEdit(widget)
	entropyStatDisplay.SetReadOnly(true)

	// Placing them in grid with their respecting labels
	resultsLayout.AddWidget2(qt.NewQLabel3("Автокорреляция").QWidget, 1, 0)
	resultsLayout.AddWidget2(autoCorrResultDisplay.QWidget, 1, 1)

	resultsLayout.AddWidget2(qt.NewQLabel3("Поиск ФС").QWidget, 2, 0)
	resultsLayout.AddWidget2(fsResultDisplay.QWidget, 2, 1)

	resultsLayout.AddWidget2(qt.NewQLabel3("Колмогоров-Смирнов").QWidget, 3, 0)
	resultsLayout.AddWidget2(ksResultDisplay.QWidget, 3, 1)

	resultsLayout.AddWidget2(qt.NewQLabel3("Компрессия").QWidget, 4, 0)
	resultsLayout.AddWidget2(compressionStatDisplay.QWidget, 4, 1)

	resultsLayout.AddWidget2(qt.NewQLabel3("Поиск сигнатур").QWidget, 5, 0)
	resultsLayout.AddWidget2(sigResultDisplay.QWidget, 5, 1)

	resultsLayout.AddWidget2(qt.NewQLabel3("Оценка энтропии").QWidget, 6, 0)
	resultsLayout.AddWidget2(entropyStatDisplay.QWidget, 6, 1)

	// Combining sublayouts into the main layout
	mainLayout.AddLayout(filePickerLayout.QLayout)
	mainLayout.AddLayout(modeSelectLayout.QLayout)
	mainLayout.AddLayout(resultsLayout.QLayout)

	// Log window (read-only)
	logWindow := qt.NewQTextEdit4("Здесь будут логи", widget)
	logWindow.SetReadOnly(true)
	logWindow.SetFont(qt.NewQFont2("monospace"))
	mainLayout.AddWidget(logWindow.QWidget)

	startButton.OnClicked(func() {
		logWindow.Clear()
		fileName := fileNameTextField.Text()
		outputDir := encryptedFileLocationEdit.Text()

		if fileName == "" || outputDir == "" {
			errorWindow := qt.NewQErrorMessage(widget)
			errorWindow.ShowMessage("Путь к исследуемому файлу/каталогу или к каталогу перемещения пуст.")
			return
		}

		if oneshotSelector.IsChecked() {
			inputFileStat, inputFileStatErr := os.Stat(fileName)
			outputDirStat, outputDirStatErr := os.Stat(outputDir)
			var inputFileTypeErr, outputDirTypeErr error

			if inputFileStat.IsDir() {
				inputFileTypeErr = errors.New("input file type is invalid")
				errorWindow := qt.NewQErrorMessage(widget)
				errorWindow.ShowMessage("Выбран режим проверки одиночного файла, но путь проверки указывает на каталог. Проверьте правильность введения пути и повторите попытку.")
			} else {
				inputFileTypeErr = nil
			}

			if outputDirStat.IsDir() == false {
				outputDirTypeErr = errors.New("output dir type is invalid")
				errorWindow := qt.NewQErrorMessage(widget)
				errorWindow.ShowMessage("Путь к каталогу для сохранения зашифрованных файлов указывает на файл. Проверьте правильность введения пути и повторите попытку.")
			} else {
				outputDirTypeErr = nil
			}

			if errors.Is(inputFileStatErr, os.ErrNotExist) {
				errorWindow := qt.NewQErrorMessage(widget)
				errorWindow.ShowMessage("Запрашиваемый файл или каталог не найден. Проверьте правильность введения пути и повторите попытку.")
			} else if errors.Is(outputDirStatErr, os.ErrNotExist) {
				errorWindow := qt.NewQErrorMessage(widget)
				errorWindow.ShowMessage("Запрашиваемый каталог для сохранения зашифрованных файлов не найден. Проверьте правильность введения пути и повторите попытку.")
			} else if inputFileStatErr == nil || inputFileTypeErr == nil || outputDirTypeErr == nil {
				var logFile = fmt.Sprintf("%s.enclog", fileName)
				logFileHandle, logOpenErr := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY, 0644)
				if logOpenErr != nil {
					fmt.Printf("Не удалось открыть файл журнала: %s", logOpenErr)
				}
				defer func(logFileHandle *os.File) {
					logCloseErr := logFileHandle.Close()
					if logCloseErr != nil {
						fmt.Printf("Не удалось закрыть файл журнала: %s", logCloseErr)
					}
				}(logFileHandle)

				fileNormalLogger := log.New(logFileHandle, "", log.LstdFlags)
				fileErrorLogger := log.New(logFileHandle, "", log.LstdFlags)

				fileExtension := filepath.Ext(fileName)
				filePath := strings.TrimSuffix(fileName, fileExtension)

				var optimizedfname = fmt.Sprintf("%s_opt%s", filePath, fileExtension)

				if _, optFileOpenErr := os.Stat(optimizedfname); errors.Is(optFileOpenErr, os.ErrNotExist) {
					fileErrorLogger.Printf("Оптимизированный файл %s не найден.", optimizedfname)
					result, fileOptimizationErr := exec.Command("python3", "prepare.py", "optimize", fileName).Output()
					if fileOptimizationErr != nil {
						fmt.Printf("Ошибка оптимизации файлв: %s", result)
					}
				}

				var blockSize = 1048576
				var autocorrThreshold = 0.125
				var ksTestThreshold = 0.1
				var compressionThreshold = 1.1
				var signatureThreshold = 150.0
				var entropyThreshold = 7.95

				var part1Result, part2Result string
				var ksStatistic, compressionStat, signatureStat, entropyStat float64
				var maxDiffPosition, readBytesCount int

				autocorrResult := autoCorrelation(optimizedfname, blockSize)
				partedResult := partedCheck(fileName)
				noFSResults := []string{"", "unknown"}
				contains := slices.Contains(noFSResults, partedResult)
				var encryptionResult = NoEncryption

				if contains {
					part1Result = "Этап 1: Шифрования не обнаружено. Переход на Этап 2."
					counter, total := createFileCounter(optimizedfname, blockSize)
					ksStatistic, maxDiffPosition, readBytesCount, _, _ = ksTest(counter, total)

					compressionStat = compressionTest(optimizedfname)
					signatureStat = signatureAnalysis(optimizedfname, blockSize)
					entropyStat = entropyEstimation(counter, total)

					var autocorrTrue = autocorrResult <= autocorrThreshold
					var ksTrue = ksStatistic <= ksTestThreshold
					var compressionTrue = compressionStat <= compressionThreshold
					var signatureTrue = signatureStat <= signatureThreshold
					var entropyTrue = entropyStat >= entropyThreshold

					var finalResult = countTrueBools(autocorrTrue, ksTrue, compressionTrue, signatureTrue, entropyTrue)

					if finalResult <= 2 {
						part2Result = fmt.Sprintf("Этап 2: Количество положительных результатов %d <= 2, шифрования не обнаружено. Завершение работы программы.", finalResult)
						encryptionResult = NoEncryption
					} else if finalResult > 3 && finalResult <= 5 {
						part2Result = fmt.Sprintf("Этап 2: Количество положительных результатов %d є [3,5], обнаружено шифрование. Завершение работы программы.", finalResult)
						encryptionResult = FullDiskEncryption
					} else {
						part2Result = "Этап 2: Произошла ошибка подсчёта."
						encryptionResult = NoEncryption
					}
				} else {
					if autocorrResult <= autocorrThreshold {
						part1Result = "Этап 1: Файловая система с высокой долей вероятности содержит пофайловое шифрование или сжатые данные. Завершение работы программы."
						encryptionResult = FileBasedEncryption
					} else {
						part1Result = "Этап 1: Шифрования не обнаружено. Файловая система с высокой долей вероятности содержит незашифрованные файлы. Завершение работы программы."
						encryptionResult = NoEncryption
					}
				}

				welcomeText := fmt.Sprintf("Код поиска шифрования разделов сырого образа диска, версия 3.1. Имя файла: %s, размер блока: %d байтов.\n", fileName, blockSize)
				logWindow.Append(welcomeText)
				fileNormalLogger.Println(welcomeText)

				autocorrLogText := fmt.Sprintf("Коэффициент автокорреляции: %f, реф. значение %f\n", autocorrResult, autocorrThreshold)
				logWindow.Append(autocorrLogText)
				fileNormalLogger.Print(autocorrLogText)
				autoCorrResultDisplay.SetText(strconv.FormatFloat(autocorrResult, 'f', -1, 64))

				fsLogText := fmt.Sprintf("Обнаруженная файловая система: %s\n", partedResult)
				logWindow.Append(fsLogText)
				fileNormalLogger.Print(fsLogText)
				fsResultDisplay.SetText(partedResult)

				noFSResults = []string{"", "unknown"}
				contains = slices.Contains(noFSResults, partedResult)

				if contains {
					logWindow.Append(part1Result)
					fileNormalLogger.Println(part1Result)
					ksLogText := fmt.Sprintf("Тест Колмогорова-Смирнова: максимальное отклонение: %f (реф. значение %f) в позиции %d, прочитано %d байтов.\n", ksStatistic, ksTestThreshold, maxDiffPosition, readBytesCount)
					logWindow.Append(ksLogText)
					fileNormalLogger.Print(ksLogText)
					ksResultDisplay.SetText(strconv.FormatFloat(ksStatistic, 'f', -1, 64))

					compLogText := fmt.Sprintf("Средний коэффициент сжатия: %f, реф. значение %f\n", compressionStat, compressionThreshold)
					logWindow.Append(compLogText)
					fileNormalLogger.Print(compLogText)
					compressionStatDisplay.SetText(strconv.FormatFloat(compressionStat, 'f', -1, 64))

					sigLogText := fmt.Sprintf("Удельное количество сигнатур на мегабайт: %f, реф. значение %f\n", signatureStat, signatureThreshold)
					logWindow.Append(sigLogText)
					fileNormalLogger.Print(sigLogText)
					sigResultDisplay.SetText(strconv.FormatFloat(signatureStat, 'f', -1, 64))

					entropyLogText := fmt.Sprintf("Оценочная информационная энтропия файла: %f, реф. значение %f\n", entropyStat, entropyThreshold)
					logWindow.Append(entropyLogText)
					fileNormalLogger.Print(entropyLogText)
					entropyStatDisplay.SetText(strconv.FormatFloat(entropyStat, 'f', -1, 64))
					logWindow.Append(part2Result)
					fileNormalLogger.Print(part2Result)
				} else {
					logWindow.Append(part1Result)
					fileNormalLogger.Print(part1Result)
				}
				fullEncryptedDirPath := filepath.Join(outputDir, "encrypted", "full_encryption")
				fileEncryptedDirPath := filepath.Join(outputDir, "encrypted", "file_encryption")

				fullEncryptedDirPathErr := os.MkdirAll(fullEncryptedDirPath, os.ModePerm)
				if fullEncryptedDirPathErr != nil {
					log.Fatal(fullEncryptedDirPathErr)
				}
				fileEncryptedDirPathErr := os.MkdirAll(fileEncryptedDirPath, os.ModePerm)
				if fileEncryptedDirPathErr != nil {
					log.Fatal(fileEncryptedDirPathErr)
				}

				if encryptionResult == FileBasedEncryption {
					fileMoveErr := os.Rename(fileName, filepath.Join(fileEncryptedDirPath, filepath.Base(fileName)))
					if fileMoveErr != nil {
						log.Fatal(fileMoveErr)
					}
				} else if encryptionResult == FullDiskEncryption {
					fileMoveErr := os.Rename(fileName, filepath.Join(fullEncryptedDirPath, filepath.Base(fileName)))
					if fileMoveErr != nil {
						log.Fatal(fileMoveErr)
					}
				}
			}
		} else if multipleSelector.IsChecked() {
			inputDirStat, inputFileStatErr := os.Stat(fileName)
			outputDirStat, outputDirStatErr := os.Stat(outputDir)
			var inputDirTypeErr, outputDirTypeErr error

			if inputDirStat.IsDir() == false {
				inputDirTypeErr = errors.New("input dir type is invalid")
				errorWindow := qt.NewQErrorMessage(widget)
				errorWindow.ShowMessage("Выбран режим проверки всех файлов в каталоге, но путь проверки указывает на одиночный файл. Проверьте правильность введения пути и повторите попытку.")
			} else {
				inputDirTypeErr = nil
			}

			if outputDirStat.IsDir() == false {
				outputDirTypeErr = errors.New("output dir type is invalid")
				errorWindow := qt.NewQErrorMessage(widget)
				errorWindow.ShowMessage("Путь к каталогу для сохранения зашифрованных файлов указывает на файл. Проверьте правильность введения пути и повторите попытку.")
			} else {
				outputDirTypeErr = nil
			}

			if errors.Is(inputFileStatErr, os.ErrNotExist) {
				errorWindow := qt.NewQErrorMessage(widget)
				errorWindow.ShowMessage("Запрашиваемый файл или каталог не найден. Проверьте правильность введения пути и повторите попытку.")
			} else if errors.Is(outputDirStatErr, os.ErrNotExist) {
				errorWindow := qt.NewQErrorMessage(widget)
				errorWindow.ShowMessage("Запрашиваемый каталог для сохранения зашифрованных файлов не найден. Проверьте правильность введения пути и повторите попытку.")
			} else if inputFileStatErr == nil || inputDirTypeErr == nil || outputDirTypeErr == nil {
				{
					entries, entryReadErr := os.ReadDir(fileName)
					if entryReadErr != nil {
						fmt.Printf("Не удалось прочитать содержимое каталога: %s", entryReadErr)
					}

					for _, entry := range entries {
						if !entry.IsDir() {
							if strings.Contains(entry.Name(), "_opt") || strings.Contains(entry.Name(), "opt") || strings.Contains(entry.Name(), "_opt_") || strings.Contains(entry.Name(), ".enclog") || strings.Contains(entry.Name(), ".enclog") {
								continue
							}
							fname := filepath.Join(fileName, entry.Name())
							log.Printf("Имя файла: %s", fname)
							_, osStatErr := os.Stat(fname)
							if errors.Is(osStatErr, os.ErrNotExist) {
								log.Printf("OS Stat error: %s", osStatErr)
								errorWindow := qt.NewQErrorMessage(widget)
								errorWindow.ShowMessage("Запрашиваемый файл или каталог не найден. Проверьте правильность введения пути и повторите попытку.")
							} else if osStatErr == nil {
								fmt.Printf("/home/gilah/%s.enclog", filepath.Base(fname))
								var logFile = fmt.Sprintf("/home/gilah/%s.enclog", filepath.Base(fname))
								logFileHandle, logOpenErr := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY, 0644)
								if logOpenErr != nil {
									fmt.Printf("Не удалось открыть файл журнала: %s", logOpenErr)
								}

								fileNormalLogger := log.New(logFileHandle, "", log.LstdFlags)
								fileErrorLogger := log.New(logFileHandle, "", log.LstdFlags)

								fileExtension := filepath.Ext(fname)
								filePath := strings.TrimSuffix(fname, fileExtension)

								var optimizedfname = fmt.Sprintf("%s_opt%s", filePath, fileExtension)

								if _, optFileOpenErr := os.Stat(optimizedfname); errors.Is(optFileOpenErr, os.ErrNotExist) {
									fileErrorLogger.Printf("Оптимизированный файл %s не найден.", optimizedfname)
									result, fileOptimizationErr := exec.Command("python3", "prepare.py", "optimize", fname).Output()
									if fileOptimizationErr != nil {
										fmt.Printf("Ошибка оптимизации файла: %s", result)
									}
								}

								var blockSize = 1048576
								var autocorrThreshold = 0.125
								var ksTestThreshold = 0.1
								var compressionThreshold = 1.1
								var signatureThreshold = 150.0
								var entropyThreshold = 7.95

								var part1Result, part2Result string
								var ksStatistic, compressionStat, signatureStat, entropyStat float64
								var maxDiffPosition, readBytesCount int

								autocorrResult := autoCorrelation(optimizedfname, blockSize)
								partedResult := partedCheck(fileName)
								noFSResults := []string{"", "unknown"}
								contains := slices.Contains(noFSResults, partedResult)

								var encryptionResult = NoEncryption

								if contains {
									part1Result = "Этап 1: Шифрования не обнаружено. Переход на Этап 2."
									counter, total := createFileCounter(optimizedfname, blockSize)
									ksStatistic, maxDiffPosition, readBytesCount, _, _ = ksTest(counter, total)

									compressionStat = compressionTest(optimizedfname)
									signatureStat = signatureAnalysis(optimizedfname, blockSize)
									entropyStat = entropyEstimation(counter, total)

									var autocorrTrue = autocorrResult <= autocorrThreshold
									var ksTrue = ksStatistic <= ksTestThreshold
									var compressionTrue = compressionStat <= compressionThreshold
									var signatureTrue = signatureStat <= signatureThreshold
									var entropyTrue = entropyStat >= entropyThreshold

									var finalResult = countTrueBools(autocorrTrue, ksTrue, compressionTrue, signatureTrue, entropyTrue)

									if finalResult <= 2 {
										part2Result = fmt.Sprintf("Этап 2: Количество положительных результатов %d <= 2, шифрования не обнаружено. Завершение работы программы.", finalResult)
										encryptionResult = NoEncryption
									} else if finalResult > 3 && finalResult <= 5 {
										part2Result = fmt.Sprintf("Этап 2: Количество положительных результатов %d є [3,5], обнаружено шифрование. Завершение работы программы.", finalResult)
										encryptionResult = FullDiskEncryption
									} else {
										part2Result = "Этап 2: Произошла ошибка подсчёта."
										encryptionResult = NoEncryption
									}
								} else {
									if autocorrResult <= autocorrThreshold {
										part1Result = "Этап 1: Файловая система с высокой долей вероятности содержит пофайловое шифрование или сжатые данные. Завершение работы программы."
										encryptionResult = FileBasedEncryption
									} else {
										part1Result = "Этап 1: Шифрования не обнаружено. Файловая система с высокой долей вероятности содержит незашифрованные файлы. Завершение работы программы."
										encryptionResult = NoEncryption
									}
								}

								welcomeText := fmt.Sprintf("Код поиска шифрования разделов сырого образа диска, версия 3.1. Имя файла: %s, размер блока: %d байтов.\n", fname, blockSize)
								logWindow.Append(welcomeText)
								fileNormalLogger.Println(welcomeText)

								autocorrLogText := fmt.Sprintf("Коэффициент автокорреляции: %f, реф. значение %f\n", autocorrResult, autocorrThreshold)
								logWindow.Append(autocorrLogText)
								fileNormalLogger.Print(autocorrLogText)
								autoCorrResultDisplay.SetText(strconv.FormatFloat(autocorrResult, 'f', -1, 64))

								fsLogText := fmt.Sprintf("Обнаруженная файловая система: %s\n", partedResult)
								logWindow.Append(fsLogText)
								fileNormalLogger.Print(fsLogText)
								fsResultDisplay.SetText(partedResult)

								noFSResults = []string{"", "unknown"}
								contains = slices.Contains(noFSResults, partedResult)

								if contains {
									logWindow.Append(part1Result)
									fileNormalLogger.Println(part1Result)
									ksLogText := fmt.Sprintf("Тест Колмогорова-Смирнова: максимальное отклонение: %f (реф. значение %f) в позиции %d, прочитано %d байтов.\n", ksStatistic, ksTestThreshold, maxDiffPosition, readBytesCount)
									logWindow.Append(ksLogText)
									fileNormalLogger.Print(ksLogText)
									ksResultDisplay.SetText(strconv.FormatFloat(ksStatistic, 'f', -1, 64))

									compLogText := fmt.Sprintf("Средний коэффициент сжатия: %f, реф. значение %f\n", compressionStat, compressionThreshold)
									logWindow.Append(compLogText)
									fileNormalLogger.Print(compLogText)
									compressionStatDisplay.SetText(strconv.FormatFloat(compressionStat, 'f', -1, 64))

									sigLogText := fmt.Sprintf("Удельное количество сигнатур на мегабайт: %f, реф. значение %f\n", signatureStat, signatureThreshold)
									logWindow.Append(sigLogText)
									fileNormalLogger.Print(sigLogText)
									sigResultDisplay.SetText(strconv.FormatFloat(signatureStat, 'f', -1, 64))

									entropyLogText := fmt.Sprintf("Оценочная информационная энтропия файла: %f, реф. значение %f\n", entropyStat, entropyThreshold)
									logWindow.Append(entropyLogText)
									fileNormalLogger.Print(entropyLogText)
									entropyStatDisplay.SetText(strconv.FormatFloat(entropyStat, 'f', -1, 64))
									logWindow.Append(part2Result)
									fileNormalLogger.Print(part2Result)
								} else {
									logWindow.Append(part1Result)
									fileNormalLogger.Print(part1Result)
								}
								logCloseErr := logFileHandle.Close()
								if logCloseErr != nil {
									fmt.Printf("Не удалось закрыть файл журнала: %s", logCloseErr)
								}

								fullEncryptedDirPath := filepath.Join(outputDir, "encrypted", "full_encryption")
								fileEncryptedDirPath := filepath.Join(outputDir, "encrypted", "file_encryption")

								fullEncryptedDirPathErr := os.MkdirAll(fullEncryptedDirPath, os.ModePerm)
								if fullEncryptedDirPathErr != nil {
									log.Fatal(fullEncryptedDirPathErr)
								}
								fileEncryptedDirPathErr := os.MkdirAll(fileEncryptedDirPath, os.ModePerm)
								if fileEncryptedDirPathErr != nil {
									log.Fatal(fileEncryptedDirPathErr)
								}

								if encryptionResult == FileBasedEncryption {
									fileMoveErr := os.Rename(fileName, filepath.Join(fileEncryptedDirPath, filepath.Base(fileName)))
									if fileMoveErr != nil {
										log.Fatal(fileMoveErr)
									}
								} else if encryptionResult == FullDiskEncryption {
									fileMoveErr := os.Rename(fileName, filepath.Join(fullEncryptedDirPath, filepath.Base(fileName)))
									if fileMoveErr != nil {
										log.Fatal(fileMoveErr)
									}
								}

							}
						}
					}
				}
			}
		}
	})

	// Window deployment
	window.SetCentralWidget(widget)
	window.Show()
	qt.QApplication_Exec()
}
