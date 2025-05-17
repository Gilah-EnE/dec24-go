package main

import (
	"os/exec"
	"strings"
)

func partedCheck(fileName string) string {
	result, _ := exec.Command("./parted", "-m", fileName, "print").CombinedOutput()
	resultText := string(result)
	if strings.Contains(resultText, "Error") {
		return ""
	}
	if result != nil {
		lines := strings.Split(resultText, "\n")
		fsType := strings.Split(lines[len(lines)-2], ":")
		return fsType[len(fsType)-3]
	} else {
		return ""
	}
}
