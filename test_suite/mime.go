package test_suite

import (
	"github.com/vimeo/go-magic/magic"
)

func LibmagicAnalysis(fileName string) string {
	cookie := magic.Open(magic.MAGIC_MIME_TYPE)
	defer magic.Close(cookie)
	defaultDir := magic.GetDefaultDir() + "/magic"
	magic.Load(cookie, defaultDir)
	fileType := magic.File(cookie, fileName)

	return fileType
}
