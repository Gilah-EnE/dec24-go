package main

import (
	"github.com/vimeo/go-magic/magic"
)

func libmagicAnalysis(fileName string) string {
	cookie := magic.Open(magic.MAGIC_MIME_TYPE)
	defer magic.Close(cookie)
	defaultDir := magic.GetDefaultDir() + "/magic"
	magic.Load(cookie, defaultDir)
	fileType := magic.File(cookie, fileName)

	return fileType
}
