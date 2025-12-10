package vmtest

import (
	"os"

	"github.com/krelinga/go-libs/exam"
)

func FileExists(e exam.E, filePath string) bool {
	_, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		return false
	} else if err != nil {
		e.Fatalf("error checking file %q: %v", filePath, err)
	}
	return true
}
