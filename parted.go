package main

import (
	"os/exec"
	"strings"
)

func partedCheck(fileName string) string {
	result, _ := exec.Command("parted", "-m", fileName, "print").Output()
	resultText := string(result)

	lines := strings.Split(resultText, "\n")
	fsType := strings.Split(lines[len(lines)-2], ":")
	return fsType[len(fsType)-3]
}
